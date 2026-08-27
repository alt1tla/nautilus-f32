// Package stanalysis provides a transport-independent Structured Text
// compiler front end. It is suitable for editors, language servers, HTTP or
// WebSocket handlers, and compiler integrations that need one stable entry
// point for parsing, validation, and IR generation.
package stanalysis

import (
	"github.com/alt1tla/nautilus-f32/lang/ir"
	"github.com/alt1tla/nautilus-f32/lang/st"
)

// Stage identifies the compiler phase that produced a diagnostic.
type Stage string

const (
	StageSyntax   Stage = "syntax"
	StageSemantic Stage = "semantic"
)

// Diagnostic is a source-positioned problem found during analysis. Positions
// are 1-based, matching st.Pos. End positions are intentionally omitted: a
// transport adapter can expand Pos to an LSP range using the source text.
type Diagnostic struct {
	Stage   Stage  `json:"stage"`
	Message string `json:"message"`
	Pos     st.Pos `json:"pos"`
}

// Symbol describes a declaration that an editor or compiler integration may
// expose for navigation and introspection.
type Symbol struct {
	Name      string `json:"name"`
	Datatype  string `json:"datatype,omitempty"`
	BlockKind string `json:"blockKind"`
	Container string `json:"container,omitempty"`
	Pos       st.Pos `json:"pos"`
}

// Options supplies declarations that live outside the analyzed source. It
// mirrors the lowerer's public context and leaves room for target-specific
// validation modes without coupling analysis to a network protocol.
type Options struct {
	UserFBs         map[string]*ir.FBDef
	UserFuncs       map[string]*ir.FuncDef
	ImplicitGlobals map[string]*ir.Type
	ScalarMode      ScalarMode
}

// ScalarMode selects the elementary-value compatibility rules used during
// semantic validation.
type ScalarMode uint8

const (
	IECStrict ScalarMode = iota
	// Float32Scalar accepts BOOL, integer, REAL, and TIME values at each
	// other's assignment and call boundaries. The target represents them as
	// float32, interprets zero as false, and interprets TIME as milliseconds.
	Float32Scalar
)

// Result contains every artifact produced by Analyze. AST remains available
// after semantic failure; IR is available only when parsing and lowering both
// succeed.
type Result struct {
	AST         *st.Program  `json:"-"`
	IR          *ir.Program  `json:"-"`
	Diagnostics []Diagnostic `json:"diagnostics,omitempty"`
	Symbols     []Symbol     `json:"symbols,omitempty"`
}

// Valid reports whether the source parsed and passed semantic validation.
func (r Result) Valid() bool { return len(r.Diagnostics) == 0 && r.IR != nil }

// Analyze parses source, collects its declarations, and lowers it to typed IR.
// It returns diagnostics as data rather than returning compiler errors, making
// it safe to call directly from long-lived editor and WebSocket handlers.
func Analyze(source string, opts Options) Result {
	ast, err := st.Parse(source)
	if err != nil {
		pos := st.Pos{Line: 1, Col: 1}
		if parsedPos, ok := st.ParseErrorPos(err); ok {
			pos = parsedPos
		}
		return Result{Diagnostics: []Diagnostic{{Stage: StageSyntax, Message: err.Error(), Pos: pos}}}
	}

	result := Result{AST: ast, Symbols: collectSymbols(ast)}
	program, err := st.LowerWithOpts(ast, st.LowerOpts{
		UserFBs:         opts.UserFBs,
		UserFuncs:       opts.UserFuncs,
		ImplicitGlobals: opts.ImplicitGlobals,
		ScalarMode:      st.ScalarMode(opts.ScalarMode),
	})
	if err != nil {
		pos := st.Pos{Line: 1, Col: 1}
		message := err.Error()
		if lowerErr, ok := st.AsLowerError(err); ok {
			if lowerErr.Pos.Line > 0 {
				pos = lowerErr.Pos
			}
			message = lowerErr.Err.Error()
		}
		result.Diagnostics = []Diagnostic{{Stage: StageSemantic, Message: message, Pos: pos}}
		return result
	}

	result.IR = program
	return result
}

func collectSymbols(program *st.Program) []Symbol {
	var symbols []Symbol
	addVariables := func(container string, blocks []st.VarBlock) {
		for _, block := range blocks {
			for _, variable := range block.Variables {
				symbols = append(symbols, Symbol{
					Name: variable.Name, Datatype: variable.Datatype,
					BlockKind: block.Kind, Container: container, Pos: variable.Pos,
				})
			}
		}
	}

	addVariables("", program.VarBlocks)
	for i := range program.TypeDecls {
		decl := &program.TypeDecls[i]
		symbols = append(symbols, Symbol{
			Name: decl.Name, Datatype: decl.Type.String(), BlockKind: "TYPE", Pos: decl.Pos,
		})
	}
	for _, decl := range program.FBDecls {
		symbols = append(symbols, Symbol{
			Name: decl.Name, Datatype: decl.Name, BlockKind: "FUNCTION_BLOCK", Pos: decl.Pos,
		})
		addVariables(decl.Name, decl.VarBlocks)
	}
	for _, decl := range program.FuncDecls {
		returnType := ""
		if decl.ReturnType != nil {
			returnType = decl.ReturnType.String()
		}
		symbols = append(symbols, Symbol{
			Name: decl.Name, Datatype: returnType, BlockKind: "FUNCTION", Pos: decl.Pos,
		})
		addVariables(decl.Name, decl.VarBlocks)
	}
	return symbols
}
