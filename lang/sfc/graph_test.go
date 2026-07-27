package sfc

import (
	"encoding/json"
	"testing"
)

// mustGraph builds the render model or fails the test immediately.
func mustGraph(t *testing.T, src string) *Model {
	t.Helper()
	m, err := Graph(src)
	if err != nil {
		t.Fatalf("Graph: %v\n%s", err, src)
	}
	return m
}

func TestGraphWorkedExample(t *testing.T) {
	m := mustGraph(t, workedExample)

	if m.Name != "TankBatch" {
		t.Errorf("Name = %q, want TankBatch", m.Name)
	}
	if len(m.Vars) != 15 {
		t.Errorf("len(Vars) = %d, want 15", len(m.Vars))
	}

	if len(m.Steps) != 5 {
		t.Fatalf("len(Steps) = %d, want 5", len(m.Steps))
	}
	wantSteps := []struct {
		id      string
		initial bool
		nAssoc  int
	}{
		{"st:Idle", true, 2},
		{"st:Fill", false, 2},
		{"st:Heat", false, 2},
		{"st:Mix", false, 2},
		{"st:Drain", false, 3},
	}
	for i, want := range wantSteps {
		s := m.Steps[i]
		if s.ID != want.id || s.Initial != want.initial || len(s.Actions) != want.nAssoc {
			t.Errorf("Steps[%d] = %+v, want id=%s initial=%v nAssoc=%d", i, s, want.id, want.initial, want.nAssoc)
		}
		if s.Line == 0 || s.EndLine <= s.Line {
			t.Errorf("Steps[%d] (%s) has bad line span %d..%d", i, s.Name, s.Line, s.EndLine)
		}
	}

	if len(m.Trans) != 5 {
		t.Fatalf("len(Trans) = %d, want 5", len(m.Trans))
	}
	wantTrans := []struct {
		id   string
		from []string
		to   []string
		kind string
	}{
		{"tr:t_start", []string{"Idle"}, []string{"Fill"}, "normal"},
		{"tr:t_abort", []string{"Fill"}, []string{"Idle"}, "alt"},
		{"tr:t_full", []string{"Fill"}, []string{"Heat", "Mix"}, "simDiverge"},
		{"tr:t_done", []string{"Heat", "Mix"}, []string{"Drain"}, "simConverge"},
		{"tr:t_empty", []string{"Drain"}, []string{"Idle"}, "normal"},
	}
	for i, want := range wantTrans {
		tr := m.Trans[i]
		if tr.ID != want.id || tr.Kind != want.kind {
			t.Errorf("Trans[%d] = id=%s kind=%s, want id=%s kind=%s", i, tr.ID, tr.Kind, want.id, want.kind)
		}
		if !equalStrings(tr.From, want.from) {
			t.Errorf("Trans[%d] (%s) From = %v, want %v", i, tr.ID, tr.From, want.from)
		}
		if !equalStrings(tr.To, want.to) {
			t.Errorf("Trans[%d] (%s) To = %v, want %v", i, tr.ID, tr.To, want.to)
		}
	}

	if len(m.Actions) != 3 {
		t.Fatalf("len(Actions) = %d, want 3", len(m.Actions))
	}
	if m.Actions[0].ID != "ac:Stir" || m.Actions[0].Body == "" {
		t.Errorf("Actions[0] = %+v, want id=ac:Stir with a body", m.Actions[0])
	}
	if m.Actions[2].ID != "ac:HoldAbort" || m.Actions[2].Body != "" {
		t.Errorf("Actions[2] (HoldAbort) = %+v, want an empty body", m.Actions[2])
	}

	// The worked example's comments are all (* ... *) block form — no bare
	// `//` runs — so the diagram's comment-note list is empty, and there's
	// no @layout block.
	if len(m.Comments) != 0 {
		t.Errorf("Comments = %v, want none (workedExample only uses (* *) comments)", m.Comments)
	}
	if m.Layout != nil {
		t.Errorf("Layout = %v, want nil (no @layout block)", m.Layout)
	}
}

// TestGraphDeterministic asserts Graph(src) always produces byte-identical
// JSON for the same input (design doc §4.1).
func TestGraphDeterministic(t *testing.T) {
	a := mustGraph(t, workedExample)
	b := mustGraph(t, workedExample)
	ja, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	jb, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(ja) != string(jb) {
		t.Errorf("Graph(src) is not deterministic:\n%s\nvs\n%s", ja, jb)
	}
}

// TestGraphCommentsAndLayout exercises the `//` comment scan and the
// (* @layout *) pin block on a small custom chart (workedExample uses
// neither).
func TestGraphCommentsAndLayout(t *testing.T) {
	src := `PROGRAM P
VAR
  X : BOOL;
  Lamp : BOOL;
END_VAR
SFC
// a note about A
// spanning two lines
INITIAL_STEP A:
  N Lamp;
END_STEP
STEP B:
  N Lamp;
END_STEP
TRANSITION FROM A TO B := X;
END_TRANSITION
TRANSITION FROM B TO A := X;
END_TRANSITION
(* @layout
  st:A 10,20
  st:B 100,20
*)
END_SFC
END_PROGRAM
`
	m := mustGraph(t, src)
	if len(m.Comments) != 1 {
		t.Fatalf("len(Comments) = %d, want 1", len(m.Comments))
	}
	if m.Comments[0].Text != "a note about A\nspanning two lines" {
		t.Errorf("Comments[0].Text = %q", m.Comments[0].Text)
	}
	if m.Layout == nil {
		t.Fatal("Layout = nil, want the two pinned entries")
	}
	if p := m.Layout["st:A"]; p != (Point{X: 10, Y: 20}) {
		t.Errorf("Layout[st:A] = %+v, want {10 20}", p)
	}
	if p := m.Layout["st:B"]; p != (Point{X: 100, Y: 20}) {
		t.Errorf("Layout[st:B] = %+v, want {100 20}", p)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
