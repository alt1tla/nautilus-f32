package server

import (
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

var scriptRe = regexp.MustCompile(`(?s)<script[^>]*>(.*?)</script>`)

// TestDashboardScriptParses guards the built-in dashboard's inline JavaScript
// against syntax errors. The page is a single hand-authored HTML file with no
// build step, so a broken script (a duplicate declaration, an unbalanced
// brace) ships silently and the whole dashboard dies at load — the browser
// just sits at "connecting…". This runs `node --check` over the extracted
// script when node is available, and skips otherwise (so it never blocks a
// node-less CI, but catches the regression wherever node exists).
func TestDashboardScriptParses(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed — skipping dashboard JS syntax check")
	}

	var js strings.Builder
	for _, m := range scriptRe.FindAllSubmatch(indexHTML, -1) {
		js.Write(m[1])
		js.WriteByte('\n')
	}
	if js.Len() == 0 {
		t.Fatal("no <script> block found in index.html")
	}

	cmd := exec.Command(node, "--check", "-")
	cmd.Stdin = strings.NewReader(js.String())
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("dashboard inline script has a syntax error:\n%s", out)
	}
}
