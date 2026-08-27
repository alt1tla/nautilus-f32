package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/alt1tla/nautilus-f32/lang/ld"
)

const ldUsage = `nautilus ld — Ladder Diagram tools

Usage:
  nautilus ld graph <file>   Emit the ladder render model for a .ld file as
                             JSON on stdout: rungs of contacts, branches,
                             coils, and in-rung function blocks, in canonical
                             ladder order. Used by the VS Code ladder
                             preview. "-" reads source from stdin. On a
                             parse error, emits {"error": "..."} and exits 1.
  nautilus ld edit           Apply a structural edit op to .ld source. Reads
                             {"source": "...", "op": {...}} JSON on stdin and
                             writes {"edits": [...]} — the rung-level text
                             edits realizing the op (1-based, end-exclusive).
                             Ops address the render model: setRef, toggleNeg,
                             setCoilMode, setArgs, insert, delete, addLeg,
                             wrapBranch, addRung, renameRung, deleteRung,
                             move, setComment, addComment, setRungComment,
                             declareVar, deleteVar. On a rejected op, emits
                             {"error": "..."} and exits 1.
`

func runLD(args []string) int {
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, ldUsage)
		return 2
	}
	switch args[0] {
	case "graph":
		return runLDGraph(args[1:])
	case "edit":
		return runLDEdit()
	default:
		fmt.Fprintf(os.Stderr, "nautilus ld: unknown subcommand %q\n\n%s", args[0], ldUsage)
		return 2
	}
}

func runLDEdit() int {
	var req struct {
		Source string    `json:"source"`
		Op     ld.EditOp `json:"op"`
	}
	enc := json.NewEncoder(os.Stdout)
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil || req.Source == "" {
		_ = enc.Encode(map[string]string{"error": "expected {\"source\": ..., \"op\": {...}} on stdin"})
		return 2
	}
	edits, err := ld.ApplyEdit(req.Source, req.Op)
	if err != nil {
		_ = enc.Encode(map[string]string{"error": err.Error()})
		return 1
	}
	_ = enc.Encode(map[string]any{"edits": edits})
	return 0
}

func runLDGraph(args []string) int {
	enc := json.NewEncoder(os.Stdout)
	if len(args) != 1 {
		fmt.Fprint(os.Stderr, ldUsage)
		return 2
	}
	var src []byte
	var err error
	if args[0] == "-" {
		src, err = io.ReadAll(os.Stdin)
	} else {
		src, err = os.ReadFile(args[0])
	}
	if err != nil {
		_ = enc.Encode(map[string]string{"error": err.Error()})
		return 2
	}
	model, err := ld.Graph(string(src))
	if err != nil {
		_ = enc.Encode(map[string]string{"error": err.Error()})
		return 1
	}
	_ = enc.Encode(model)
	return 0
}
