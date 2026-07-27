package sfc

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/joyautomation/nautilus/lang/st"
)

// Parse parses a full .sfc source file into an AST: the ST-standard header
// (PROGRAM name + VAR blocks) plus the SFC ... END_SFC body.
//
// The header is parsed by feeding it to st.Parse with a synthesized
// trailing END_PROGRAM — the real END_PROGRAM sits after END_SFC in the
// actual file, out of reach of a header-only parse. This reuses lang/st's
// complete declaration grammar (multi-name decls, arrays, structs, RETAIN/
// CONSTANT, initializers — everything st.Parse already knows) without
// forking it or modifying lang/st; see the package doc in sfc.go for why
// this is the chosen reuse seam. Because the header is a byte-for-byte
// prefix of the original file, st.Parse's line numbers land on the exact
// original file lines with no remapping needed.
func Parse(source string) (*Program, error) {
	header, body, _, bodyLine, err := splitSFC(source)
	if err != nil {
		return nil, err
	}
	headerProg, err := st.Parse(header + "\nEND_PROGRAM\n")
	if err != nil {
		return nil, fmt.Errorf("sfc: header: %w", err)
	}
	prog := &Program{
		Name:      headerProg.Name,
		VarBlocks: headerProg.VarBlocks,
		Pos:       headerPos(header),
	}
	p := &bodyParser{tokens: Lex(body, bodyLine), src: body}
	if err := p.parseBody(prog); err != nil {
		return nil, err
	}
	return prog, nil
}

var programLineRe = regexp.MustCompile(`(?im)^\s*PROGRAM\s+`)

// headerPos finds the PROGRAM keyword's line for a friendlier anchor on
// program-wide diagnostics (e.g. "no INITIAL_STEP"); Pos{1,1} if absent.
func headerPos(header string) Pos {
	lines := strings.Split(header, "\n")
	for i, l := range lines {
		if programLineRe.MatchString(l) {
			return Pos{Line: i + 1, Col: 1}
		}
	}
	return Pos{Line: 1, Col: 1}
}

// splitSFC separates source into the ST header (everything before the SFC
// line), the SFC body (between the SFC and END_SFC marker lines), and the
// footer (END_SFC onward). bodyLine is the 1-based file line of the SFC
// marker itself; adding it to a 1-based line computed relative to body
// recovers the original file line — the same convention lang/fbd's
// splitFBD uses for bodyLine.
func splitSFC(source string) (header, body, footer string, bodyLine int, err error) {
	lines := strings.Split(source, "\n")
	start, end := -1, -1
	for i, l := range lines {
		if start == -1 && sfcStartRe.MatchString(l) {
			start = i
		}
		if sfcEndRe.MatchString(l) {
			end = i
		}
	}
	if start == -1 || end == -1 || end < start {
		return "", "", "", 0, fmt.Errorf("sfc: source must contain an SFC ... END_SFC body")
	}
	header = strings.Join(lines[:start], "\n")
	body = strings.Join(lines[start+1:end], "\n")
	footer = strings.Join(lines[end+1:], "\n")
	return header, body, footer, start + 1, nil
}

// ─── body parser ────────────────────────────────────────────────────────────

// bodyParser parses the token stream produced by Lex over one .sfc file's
// SFC body into steps, transitions, and action blocks.
type bodyParser struct {
	tokens []Token
	pos    int
	src    string // the body text Lex tokenized; spans slice into this
}

func (p *bodyParser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *bodyParser) advance() Token {
	tok := p.peek()
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
	return tok
}

func (p *bodyParser) expect(tt TokenType) (Token, error) {
	tok := p.peek()
	if tok.Type != tt {
		return tok, fmt.Errorf("sfc: line %d: unexpected %q", tok.Line, tok.Literal)
	}
	return p.advance(), nil
}

// spanBetween slices the verbatim body text from start's offset up to (not
// including) end's offset, trimmed of trailing whitespace.
func (p *bodyParser) spanBetween(start, end Token) Span {
	text := p.src[start.Offset:end.Offset]
	text = strings.TrimRight(text, " \t\r\n")
	return Span{Text: text, Line: start.Line, Col: start.Col}
}

func (p *bodyParser) parseBody(prog *Program) error {
	for p.peek().Type != TokenEOF {
		switch p.peek().Type {
		case TokenInitialStep, TokenStep:
			step, err := p.parseStep()
			if err != nil {
				return err
			}
			prog.Steps = append(prog.Steps, step)
		case TokenTransition:
			tr, err := p.parseTransition()
			if err != nil {
				return err
			}
			prog.Transitions = append(prog.Transitions, tr)
		case TokenAction:
			ab, err := p.parseAction()
			if err != nil {
				return err
			}
			prog.Actions = append(prog.Actions, ab)
		default:
			tok := p.peek()
			return fmt.Errorf("sfc: line %d: expected STEP, INITIAL_STEP, TRANSITION, or ACTION, got %q", tok.Line, tok.Literal)
		}
	}
	return nil
}

// parseStep parses `(INITIAL_STEP|STEP) name: { assoc } END_STEP`.
func (p *bodyParser) parseStep() (*Step, error) {
	kwTok := p.advance()
	initial := kwTok.Type == TokenInitialStep
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, fmt.Errorf("sfc: line %d: expected a step name after %s", kwTok.Line, kwTok.Literal)
	}
	if _, err := p.expect(TokenColon); err != nil {
		return nil, fmt.Errorf("sfc: line %d: step %q: expected ':'", kwTok.Line, nameTok.Literal)
	}
	step := &Step{Name: nameTok.Literal, Initial: initial, Pos: Pos{Line: kwTok.Line, Col: kwTok.Col}}
	for p.peek().Type != TokenEndStep {
		if p.peek().Type == TokenEOF {
			return nil, fmt.Errorf("sfc: line %d: step %q: missing END_STEP", kwTok.Line, step.Name)
		}
		assoc, err := p.parseAssoc()
		if err != nil {
			return nil, err
		}
		step.Actions = append(step.Actions, assoc)
	}
	endTok := p.advance()
	step.EndPos = Pos{Line: endTok.Line, Col: endTok.Col}
	return step, nil
}

// parseAssoc parses `qualifier target [(time-literal)] ;`.
func (p *bodyParser) parseAssoc() (Assoc, error) {
	qualTok, err := p.expect(TokenIdent)
	if err != nil {
		return Assoc{}, fmt.Errorf("sfc: line %d: expected an action qualifier (N, S, R, P, P0, P1, ...)", p.peek().Line)
	}
	targetTok, err := p.expect(TokenIdent)
	if err != nil {
		return Assoc{}, fmt.Errorf("sfc: line %d: expected an action or variable name after qualifier %q", qualTok.Line, qualTok.Literal)
	}
	a := Assoc{Qualifier: strings.ToUpper(qualTok.Literal), Target: targetTok.Literal, Pos: Pos{Line: qualTok.Line, Col: qualTok.Col}}
	if p.peek().Type == TokenLParen {
		p.advance()
		startOff := p.peek().Offset
		depth := 1
		for depth > 0 {
			t := p.peek()
			if t.Type == TokenEOF {
				return Assoc{}, fmt.Errorf("sfc: line %d: association %s %s: unclosed '('", a.Pos.Line, a.Qualifier, a.Target)
			}
			if t.Type == TokenLParen {
				depth++
			}
			if t.Type == TokenRParen {
				depth--
				if depth == 0 {
					break
				}
			}
			p.advance()
		}
		endOff := p.peek().Offset
		a.Time = strings.TrimSpace(p.src[startOff:endOff])
		p.advance() // consume the closing ')'
	}
	if _, err := p.expect(TokenSemicolon); err != nil {
		return Assoc{}, fmt.Errorf("sfc: line %d: association %s %s: expected ';'", a.Pos.Line, a.Qualifier, a.Target)
	}
	return a, nil
}

// parseTransition parses `TRANSITION [name] FROM step-set TO step-set :=
// <raw condition text> ; END_TRANSITION`.
func (p *bodyParser) parseTransition() (*Transition, error) {
	kwTok := p.advance()
	tr := &Transition{Pos: Pos{Line: kwTok.Line, Col: kwTok.Col}}
	if p.peek().Type == TokenIdent {
		tr.Name = p.advance().Literal
	}
	if _, err := p.expect(TokenFrom); err != nil {
		return nil, fmt.Errorf("sfc: line %d: transition %s: expected FROM", kwTok.Line, trName(tr))
	}
	from, err := p.parseStepSet()
	if err != nil {
		return nil, err
	}
	tr.From = from
	if _, err := p.expect(TokenTo); err != nil {
		return nil, fmt.Errorf("sfc: line %d: transition %s: expected TO", kwTok.Line, trName(tr))
	}
	to, err := p.parseStepSet()
	if err != nil {
		return nil, err
	}
	tr.To = to
	if _, err := p.expect(TokenAssign); err != nil {
		return nil, fmt.Errorf("sfc: line %d: transition %s: expected ':='", kwTok.Line, trName(tr))
	}
	startTok := p.peek()
	for p.peek().Type != TokenSemicolon {
		if p.peek().Type == TokenEOF {
			return nil, fmt.Errorf("sfc: line %d: transition %s: missing terminating ';'", kwTok.Line, trName(tr))
		}
		p.advance()
	}
	semiTok := p.peek()
	tr.Cond = p.spanBetween(startTok, semiTok)
	p.advance() // consume ';'
	endTok, err := p.expect(TokenEndTransition)
	if err != nil {
		return nil, fmt.Errorf("sfc: line %d: transition %s: expected END_TRANSITION", semiTok.Line, trName(tr))
	}
	tr.EndPos = Pos{Line: endTok.Line, Col: endTok.Col}
	return tr, nil
}

// parseStepSet parses `ident | "(" ident { "," ident } ")"`.
func (p *bodyParser) parseStepSet() ([]string, error) {
	if p.peek().Type == TokenLParen {
		p.advance()
		var names []string
		for {
			idTok, err := p.expect(TokenIdent)
			if err != nil {
				return nil, fmt.Errorf("sfc: line %d: expected a step name in a step-set", idTok.Line)
			}
			names = append(names, idTok.Literal)
			if p.peek().Type == TokenComma {
				p.advance()
				continue
			}
			break
		}
		if _, err := p.expect(TokenRParen); err != nil {
			return nil, fmt.Errorf("sfc: line %d: step-set: expected ')'", p.peek().Line)
		}
		return names, nil
	}
	idTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, fmt.Errorf("sfc: line %d: expected a step name or '(' to start a step-set", idTok.Line)
	}
	return []string{idTok.Literal}, nil
}

// parseAction parses `ACTION name: <raw statement-list text> END_ACTION`.
func (p *bodyParser) parseAction() (*ActionBlock, error) {
	kwTok := p.advance()
	nameTok, err := p.expect(TokenIdent)
	if err != nil {
		return nil, fmt.Errorf("sfc: line %d: expected an action name after ACTION", kwTok.Line)
	}
	if _, err := p.expect(TokenColon); err != nil {
		return nil, fmt.Errorf("sfc: line %d: action %q: expected ':'", kwTok.Line, nameTok.Literal)
	}
	ab := &ActionBlock{Name: nameTok.Literal, Pos: Pos{Line: kwTok.Line, Col: kwTok.Col}}
	startTok := p.peek()
	for p.peek().Type != TokenEndAction {
		if p.peek().Type == TokenEOF {
			return nil, fmt.Errorf("sfc: line %d: action %q: missing END_ACTION", kwTok.Line, ab.Name)
		}
		p.advance()
	}
	endTok := p.peek()
	ab.Body = p.spanBetween(startTok, endTok)
	p.advance() // consume END_ACTION
	ab.EndPos = Pos{Line: endTok.Line, Col: endTok.Col}
	return ab, nil
}

// trName renders a transition's diagnostic label: its declared name, or a
// line-anchored placeholder for the (legal) unnamed case.
func trName(t *Transition) string {
	if t.Name != "" {
		return t.Name
	}
	return fmt.Sprintf("(unnamed, line %d)", t.Pos.Line)
}
