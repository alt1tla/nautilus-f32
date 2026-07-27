package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/joyautomation/nautilus/internal/stproject"
	"github.com/joyautomation/nautilus/lang/fbd"
	"github.com/joyautomation/nautilus/lang/ld"
	"github.com/joyautomation/nautilus/lang/sfc"
	"github.com/joyautomation/nautilus/lang/st"
)

// runCheck compiles every .st file under the given paths (files or
// directories; default ".") and prints gcc-style diagnostics:
//
//	path/to/program.st:12:5: undeclared identifier "y" ...
//
// Exit code 0 = clean, 1 = diagnostics found, 2 = usage/IO error.
func runCheck(args []string) int {
	paths := args
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			fmt.Fprintln(os.Stderr, "nautilus check:", err)
			return 2
		}
		if !info.IsDir() {
			files = append(files, p)
			continue
		}
		err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				// Skip the usual dependency/VCS trees.
				switch d.Name() {
				case ".git", "node_modules", "vendor":
					return filepath.SkipDir
				}
				return nil
			}
			if ext := strings.ToLower(filepath.Ext(path)); ext == ".st" || ext == ".fbd" || ext == ".ld" || ext == ".sfc" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "nautilus check:", err)
			return 2
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "nautilus check: no .st, .fbd, .ld, or .sfc files found")
		return 0
	}

	bad := 0
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, "nautilus check:", err)
			return 2
		}
		source := string(src)
		// Graphical languages compile by transpiling toward ST — LD to the
		// FBD netlist, FBD to ST — then check exactly like an .st file; the
		// composed line maps project diagnostic positions back onto the
		// original source.
		var lineMap []int
		if strings.EqualFold(filepath.Ext(f), ".sfc") {
			// SFC transpiles directly to ST (a sibling of the LD/FBD hops,
			// not a stage in their chain — docs/design/sfc.md §3). The
			// structural checks of §5.1 run for real here; the ST-level
			// hop (sfc.TranspileWithLines) is a follow-on slice's
			// deliverable and errors cleanly until it lands.
			prog, perr := sfc.Parse(source)
			if perr != nil {
				bad++
				fmt.Printf("%s: %s\n", f, perr.Error())
				continue
			}
			hasErr := false
			for _, d := range sfc.Check(prog) {
				fmt.Printf("%s:%d:%d: %s: %s\n", f, d.Pos.Line, d.Pos.Col, d.Severity, d.Message)
				if d.Severity == sfc.SeverityError {
					hasErr = true
				}
			}
			if hasErr {
				bad++
				continue
			}
			stSrc, lm, terr := sfc.TranspileWithLines(source)
			if terr != nil {
				bad++
				fmt.Printf("%s: %s\n", f, terr.Error())
				continue
			}
			source, lineMap = stSrc, lm
		}
		if strings.EqualFold(filepath.Ext(f), ".ld") {
			fbdSrc, lm, terr := ld.TranspileWithLines(source)
			if terr != nil {
				bad++
				fmt.Printf("%s: %s\n", f, terr.Error())
				continue
			}
			source, lineMap = fbdSrc, lm
		}
		// After the LD hop the source always carries an FBD block. SFC is
		// not in this chain — it transpiled directly to ST above.
		if strings.EqualFold(filepath.Ext(f), ".fbd") || strings.EqualFold(filepath.Ext(f), ".ld") {
			stSrc, lm, terr := fbd.TranspileWithLines(source)
			if terr != nil {
				bad++
				fmt.Printf("%s: %s\n", f, terr.Error())
				continue
			}
			// Compose: ST line → FBD line → (for .ld) LD line.
			if lineMap != nil {
				composed := make([]int, len(lm))
				for i, fbdLine := range lm {
					if fbdLine >= 1 && fbdLine <= len(lineMap) {
						composed[i] = lineMap[fbdLine-1]
					} else {
						composed[i] = 1
					}
				}
				lm = composed
			}
			source, lineMap = stSrc, lm
		}
		// Sibling library files (TYPE/FB/FUNCTION-only .st in the same
		// directory) are in scope, exactly as the LSP and a runtime that
		// concatenates sources see it.
		prelude, preludeLines := stproject.Prelude(f, nil)
		if msg, pos, failed := compileErr(source, prelude, preludeLines); failed {
			bad++
			if lineMap != nil {
				if pos.Line >= 1 && pos.Line <= len(lineMap) {
					pos = st.Pos{Line: lineMap[pos.Line-1], Col: 1}
				} else {
					pos = st.Pos{Line: 1, Col: 1}
				}
			}
			fmt.Printf("%s:%d:%d: %s\n", f, pos.Line, pos.Col, msg)
		}
	}

	fmt.Printf("nautilus check: %d file(s), %d with errors\n", len(files), bad)
	if bad > 0 {
		return 1
	}
	return 0
}

// compileErr runs the same parse+lower pipeline as the LSP and returns the
// first diagnostic. Positions default to 1:1 when the compiler couldn't
// attach one (e.g. some parse errors). The prelude participates in lowering
// only; positions are remapped back into the checked file.
func compileErr(src, prelude string, preludeLines int) (string, st.Pos, bool) {
	prog, err := st.Parse(src)
	if err != nil {
		// Anchor on the parser-reported position (shared with the LSP via
		// st.ParseErrorPos) instead of always 1:1.
		pos := st.Pos{Line: 1, Col: 1}
		if p, ok := st.ParseErrorPos(err); ok {
			pos = p
		}
		return err.Error(), pos, true
	}
	lowerProg := prog
	if prelude != "" {
		if combined, cerr := st.Parse(prelude + src); cerr == nil {
			lowerProg = combined
		} else {
			preludeLines = 0
		}
	} else {
		preludeLines = 0
	}
	if _, err := st.Lower(lowerProg); err != nil {
		pos := st.Pos{Line: 1, Col: 1}
		msg := err.Error()
		if le, ok := st.AsLowerError(err); ok && le.Pos.Line > 0 {
			// Print the unwrapped message: the position prefix that
			// LowerError.Error() adds is already in the path:line:col.
			pos, msg = le.Pos, le.Err.Error()
		}
		if pos.Line > preludeLines {
			pos.Line -= preludeLines
		} else if preludeLines > 0 {
			pos = st.Pos{Line: 1, Col: 1}
			msg = "in project library files: " + msg
		}
		return msg, pos, true
	}
	return "", st.Pos{}, false
}
