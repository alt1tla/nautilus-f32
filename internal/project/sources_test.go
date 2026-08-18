package project

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/joyautomation/nautilus/runtime"
)

// Sources must compose exactly what Load hands runtime.New: libraries
// joined ahead of each task's program, tasks keyed by the names online
// edits address them with.
func TestSources(t *testing.T) {
	fsys := fstest.MapFS{
		"nautilus.yaml": &fstest.MapFile{Data: []byte(
			"name: t\ntasks:\n  - program: main.st\n  - program: totals.st\n  - { program: slow.st, name: slow }\n")},
		"main.st":   &fstest.MapFile{Data: []byte("PROGRAM Main\nEND_PROGRAM\n")},
		"totals.st": &fstest.MapFile{Data: []byte("PROGRAM Totals\nEND_PROGRAM\n")},
		"slow.st":   &fstest.MapFile{Data: []byte("PROGRAM Slow\nEND_PROGRAM\n")},
		"lib.st":    &fstest.MapFile{Data: []byte("FUNCTION F : REAL\nF := 1.0;\nEND_FUNCTION\n")},
	}
	srcs, err := Sources(fsys, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 3 {
		t.Fatalf("want 3 tasks, got %v", len(srcs))
	}
	// First task is the main task; unnamed extra tasks take the file's
	// basename; explicit names win.
	main, ok := srcs[runtime.MainTaskName]
	if !ok || !strings.HasPrefix(main, "FUNCTION F") || !strings.Contains(main, "PROGRAM Main") {
		t.Fatalf("main task must be library-composed: %q", main)
	}
	if _, ok := srcs["totals"]; !ok {
		t.Fatalf("unnamed task should key by basename: %v", keys(srcs))
	}
	if _, ok := srcs["slow"]; !ok {
		t.Fatalf("named task should key by name: %v", keys(srcs))
	}

	// Parity with the boot path: what Sources composes for the main task is
	// byte-for-byte what runtime.New compiles from Load's output.
	proj, err := Load(fsys, "")
	if err != nil {
		t.Fatal(err)
	}
	rt, err := runtime.New(proj.Runtime)
	if err != nil {
		t.Fatal(err)
	}
	if rt.Program().Source() != main {
		t.Fatal("Sources and boot must compose identically")
	}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
