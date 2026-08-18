package vcs

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joyautomation/nautilus/internal/project"
	"github.com/joyautomation/nautilus/runtime"
)

// gitRepo builds a throwaway repo with the project in a subdirectory —
// the monorepo shape, which exercises the --show-prefix subtree path —
// and returns the project dir plus a commit helper.
func gitRepo(t *testing.T) (string, func(msg string) string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	proj := filepath.Join(root, "plant", "line1")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatal(err)
	}
	git := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
		// Isolate from the machine's git config; identity comes from env.
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
			"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=t@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q", "-b", "main")
	commit := func(msg string) string {
		git("add", "-A")
		git("commit", "-q", "-m", msg)
		return git("rev-parse", "HEAD")
	}
	return proj, commit
}

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCapture(t *testing.T) {
	proj, commit := gitRepo(t)
	write(t, proj, "nautilus.yaml", "name: line1\ntasks:\n  - program: program.st\n")
	write(t, proj, "program.st", "PROGRAM P\nEND_PROGRAM\n")
	write(t, proj, "program_test.yaml", "suite: {}\n")
	sha1 := commit("first cut")
	write(t, proj, "program.st", "PROGRAM P\n(* v2 *)\nEND_PROGRAM\n")
	sha2 := commit("tune the loop")

	h, err := Capture(proj, 0)
	if err != nil {
		t.Fatal(err)
	}
	if h == nil || h.Built != sha2 || h.BuiltDirty {
		t.Fatalf("built = %+v", h)
	}
	if len(h.Commits) != 2 || h.Commits[0].Sha != sha2 || h.Commits[1].Sha != sha1 {
		t.Fatalf("commits: %+v", h.Commits)
	}
	c := h.Commits[0]
	if c.Subject != "tune the loop" || c.Author != "Test" || !strings.Contains(c.Diff, "(* v2 *)") {
		t.Fatalf("head commit: %+v", c)
	}
	// Snapshots are project-relative and skip acceptance suites; the
	// unchanged manifest shares one blob across both commits.
	if _, ok := c.Files["program.st"]; !ok {
		t.Fatalf("snapshot paths must be project-relative: %v", c.Files)
	}
	if _, ok := c.Files["program_test.yaml"]; ok {
		t.Fatal("acceptance suites must not ride in snapshots")
	}
	if h.Commits[0].Files["nautilus.yaml"] != h.Commits[1].Files["nautilus.yaml"] {
		t.Fatal("an unchanged file must dedupe to one blob")
	}

	// SnapshotFS reconstructs the old revision, by short sha too.
	fsys, ok := SnapshotFS(h, h.Commits[1].Short)
	if !ok {
		t.Fatal("snapshot for sha1 should exist")
	}
	raw, err := fsys.ReadFile("program.st")
	if err != nil || strings.Contains(string(raw), "(* v2 *)") {
		t.Fatalf("old snapshot content wrong: %q %v", raw, err)
	}
	if _, ok := SnapshotFS(h, "nope"); ok {
		t.Fatal("unknown sha must not resolve")
	}

	// A dirty tree is flagged — the artifact's provenance is approximate.
	write(t, proj, "program.st", "PROGRAM P\n(* uncommitted *)\nEND_PROGRAM\n")
	h, err = Capture(proj, 0)
	if err != nil || !h.BuiltDirty {
		t.Fatalf("dirty tree not flagged: %v %v", h.BuiltDirty, err)
	}
}

// The whole SourcesAt chain the runner wires: capture a repo, reconstruct
// an old commit's tree, and compose its program sources exactly as boot
// would — library joined ahead, main task keyed by name.
func TestSnapshotComposesSources(t *testing.T) {
	proj, commit := gitRepo(t)
	write(t, proj, "nautilus.yaml", "name: line1\ntasks:\n  - program: program.st\n")
	write(t, proj, "lib.st", "FUNCTION F : REAL\nF := 1.0;\nEND_FUNCTION\n")
	write(t, proj, "program.st", "PROGRAM P\nEND_PROGRAM\n")
	old := commit("v1")
	write(t, proj, "program.st", "PROGRAM P\n(* v2 *)\nEND_PROGRAM\n")
	commit("v2")

	h, err := Capture(proj, 0)
	if err != nil {
		t.Fatal(err)
	}
	snap, ok := SnapshotFS(h, old)
	if !ok {
		t.Fatal("old snapshot should exist")
	}
	srcs, err := project.Sources(snap, "")
	if err != nil {
		t.Fatal(err)
	}
	main := srcs[runtime.MainTaskName]
	if !strings.HasPrefix(main, "FUNCTION F") || !strings.Contains(main, "PROGRAM P") ||
		strings.Contains(main, "(* v2 *)") {
		t.Fatalf("old composed source wrong: %q", main)
	}
}

// Outside a repo the answer is "no history", not an error.
func TestCaptureNoRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	h, err := Capture(t.TempDir(), 0)
	if h != nil || err != nil {
		t.Fatalf("want (nil, nil), got (%v, %v)", h, err)
	}
}
