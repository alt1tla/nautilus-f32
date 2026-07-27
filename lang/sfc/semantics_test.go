package sfc

import (
	"strings"
	"testing"

	"github.com/joyautomation/nautilus/lang/ir"
)

// runScans compiles src, runs len(times) scans setting inputs[i] and now=times[i]
// before each, and returns the program+frame for assertions.
func runScans(t *testing.T, src string, times []int64, inputs []map[string]ir.Value) (*ir.Program, *ir.Frame, *clockHost) {
	t.Helper()
	prog := mustCompile(t, src)
	h := &clockHost{vals: map[string]ir.Value{}}
	frame := ir.NewFrame(prog)
	for i, now := range times {
		if i < len(inputs) {
			for k, v := range inputs[i] {
				h.vals[k] = v
			}
		}
		h.now = now
		if err := ir.Run(prog, frame, h); err != nil {
			t.Fatalf("scan %d: %v", i, err)
		}
	}
	return prog, frame, h
}

// TestAltPriorityMutualExclusion: two transitions from the same step both
// enabled → only the first (declared) one fires.
func TestAltPriorityMutualExclusion(t *testing.T) {
	src := `PROGRAM P
VAR_EXTERNAL Go : BOOL; END_VAR
SFC
  INITIAL_STEP S:
  END_STEP
  STEP A:
  END_STEP
  STEP B:
  END_STEP
  TRANSITION t1 FROM S TO A := Go;
  END_TRANSITION
  TRANSITION t2 FROM S TO B := Go;
  END_TRANSITION
END_SFC
END_PROGRAM`
	prog, frame, _ := runScans(t, src, []int64{1000}, []map[string]ir.Value{
		{"Go": ir.BoolVal(true)},
	})
	if !stepActive(t, prog, frame, "A") {
		t.Error("higher-priority t1 should fire → A active")
	}
	if stepActive(t, prog, frame, "B") {
		t.Error("t2 must be suppressed → B inactive")
	}
	if stepActive(t, prog, frame, "S") {
		t.Error("S should have cleared")
	}
}

// TestEnabledNotCondGuard is the regression of §2.3: a higher-priority
// convergence whose OTHER source is inactive (so cond can be true but it is not
// enabled) must NOT suppress a lower-priority branch sharing the step.
func TestEnabledNotCondGuard(t *testing.T) {
	src := `PROGRAM P
VAR_EXTERNAL Go : BOOL; END_VAR
SFC
  INITIAL_STEP S:
  END_STEP
  STEP Other:
  END_STEP
  STEP Z:
  END_STEP
  STEP W:
  END_STEP
  TRANSITION t_hi FROM (S, Other) TO Z := Go;
  END_TRANSITION
  TRANSITION t_lo FROM S TO W := Go;
  END_TRANSITION
  TRANSITION seed FROM Z TO Other := FALSE;
  END_TRANSITION
END_SFC
END_PROGRAM`
	// S is active (initial), Other is NOT. Go=TRUE so cond(t_hi) is true, but
	// enabled(t_hi) is false (Other inactive) → t_lo must fire.
	prog, frame, _ := runScans(t, src, []int64{1000}, []map[string]ir.Value{
		{"Go": ir.BoolVal(true)},
	})
	if !stepActive(t, prog, frame, "W") {
		t.Error("lower branch must fire when higher convergence is not enabled")
	}
	if stepActive(t, prog, frame, "Z") {
		t.Error("t_hi cannot fire with Other inactive")
	}
}

// TestSetDominatesClear: a self-loop FROM S TO S leaves S active (§2.4).
func TestSetDominatesClear(t *testing.T) {
	src := `PROGRAM P
VAR_EXTERNAL Go : BOOL; END_VAR
SFC
  INITIAL_STEP S:
  END_STEP
  TRANSITION t FROM S TO S := Go;
  END_TRANSITION
END_SFC
END_PROGRAM`
	prog, frame, _ := runScans(t, src, []int64{1000}, []map[string]ir.Value{
		{"Go": ir.BoolVal(true)},
	})
	if !stepActive(t, prog, frame, "S") {
		t.Error("self-loop must leave S active (set dominates clear)")
	}
}

// TestP1FiresOncePerActivation: a P1 body runs exactly once per activation,
// including on re-entry.
func TestP1FiresOncePerActivation(t *testing.T) {
	src := `PROGRAM P
VAR_EXTERNAL Fwd : BOOL; Back : BOOL; Count : INT; END_VAR
SFC
  INITIAL_STEP Home:
  END_STEP
  STEP Work:
    P1 Bump;
  END_STEP
  TRANSITION t1 FROM Home TO Work := Fwd;
  END_TRANSITION
  TRANSITION t2 FROM Work TO Home := Back;
  END_TRANSITION
  ACTION Bump:
    Count := Count + 1;
  END_ACTION
END_SFC
END_PROGRAM`
	prog := mustCompile(t, src)
	h := &clockHost{vals: map[string]ir.Value{}}
	frame := ir.NewFrame(prog)
	run := func(in map[string]ir.Value) {
		for k, v := range in {
			h.vals[k] = v
		}
		h.now += 1000
		if err := ir.Run(prog, frame, h); err != nil {
			t.Fatal(err)
		}
	}
	run(map[string]ir.Value{"Fwd": ir.BoolVal(true)})               // Home→Work, pulse #1
	run(map[string]ir.Value{"Fwd": ir.BoolVal(false)})              // Work, no pulse
	run(nil)                                                        // Work, no pulse
	if h.vals["Count"].I != 1 {
		t.Fatalf("after first activation Count=%d, want 1", h.vals["Count"].I)
	}
	run(map[string]ir.Value{"Back": ir.BoolVal(true)})              // Work→Home
	run(map[string]ir.Value{"Back": ir.BoolVal(false), "Fwd": ir.BoolVal(true)}) // Home→Work, pulse #2
	if h.vals["Count"].I != 2 {
		t.Fatalf("after re-entry Count=%d, want 2 (one pulse per activation)", h.vals["Count"].I)
	}
}

// TestStoredLatchAcrossSteps: S on one step sets a stored boolean action; R on
// a later step clears it; the value latches across the steps between.
func TestStoredLatchAcrossSteps(t *testing.T) {
	src := `PROGRAM P
VAR_EXTERNAL Go : BOOL; Next : BOOL; Clear : BOOL; Lamp : BOOL; END_VAR
SFC
  INITIAL_STEP S0:
  END_STEP
  STEP S1:
    S Lamp;
  END_STEP
  STEP S2:
  END_STEP
  STEP S3:
    R Lamp;
  END_STEP
  TRANSITION a FROM S0 TO S1 := Go;
  END_TRANSITION
  TRANSITION b FROM S1 TO S2 := Next;
  END_TRANSITION
  TRANSITION c FROM S2 TO S3 := Clear;
  END_TRANSITION
END_SFC
END_PROGRAM`
	prog := mustCompile(t, src)
	h := &clockHost{vals: map[string]ir.Value{}}
	frame := ir.NewFrame(prog)
	run := func(in map[string]ir.Value) {
		for k, v := range in {
			h.vals[k] = v
		}
		h.now += 1000
		if err := ir.Run(prog, frame, h); err != nil {
			t.Fatal(err)
		}
	}
	run(map[string]ir.Value{"Go": ir.BoolVal(true)}) // S0→S1: sets Lamp
	if !h.vals["Lamp"].B {
		t.Fatal("S Lamp should set on S1 activation")
	}
	run(map[string]ir.Value{"Go": ir.BoolVal(false), "Next": ir.BoolVal(true)}) // S1→S2
	if !h.vals["Lamp"].B {
		t.Fatal("Lamp should latch across S2 (no R yet)")
	}
	run(map[string]ir.Value{"Next": ir.BoolVal(false), "Clear": ir.BoolVal(true)}) // S2→S3: clears
	if h.vals["Lamp"].B {
		t.Fatal("R Lamp on S3 should clear the latch")
	}
}

// TestFinalScanResetsTON is the §2.5.1 hazard in isolation: a body action that
// drives a TON from the step's .X must, on the falling edge, get one final scan
// that calls the TON with IN:=FALSE, resetting it for a clean (fresh-timing)
// re-entry.
func TestFinalScanResetsTON(t *testing.T) {
	src := `PROGRAM P
VAR_EXTERNAL Go : BOOL; Reset : BOOL; Q : BOOL; END_VAR
VAR mt : TON; END_VAR
SFC
  INITIAL_STEP Home:
  END_STEP
  STEP Run:
    N Timing;
  END_STEP
  TRANSITION t1 FROM Home TO Run := Go;
  END_TRANSITION
  TRANSITION t2 FROM Run TO Home := Reset;
  END_TRANSITION
  ACTION Timing:
    mt(IN := Run.X, PT := T#10S);
    Q := mt.Q;
  END_ACTION
END_SFC
END_PROGRAM`
	prog := mustCompile(t, src)
	h := &clockHost{vals: map[string]ir.Value{}}
	frame := ir.NewFrame(prog)
	mtIdx := slotIdx(prog, "mt")
	et := func() int64 { return frame.Slots[mtIdx].FB.Slots[3].I }
	run := func(in map[string]ir.Value, now int64) {
		for k, v := range in {
			h.vals[k] = v
		}
		h.now = now
		if err := ir.Run(prog, frame, h); err != nil {
			t.Fatal(err)
		}
	}
	// mt is a global (VAR_EXTERNAL) here so ET is observable via the host.
	run(map[string]ir.Value{"Go": ir.BoolVal(true)}, 1000)  // Home→Run, mt starts
	run(map[string]ir.Value{"Go": ir.BoolVal(false)}, 4000) // Run, mt at 3s
	if et() == 0 {
		t.Fatal("mt should be counting mid-run")
	}
	run(map[string]ir.Value{"Reset": ir.BoolVal(true)}, 6000) // Run→Home: final scan resets mt
	if et() != 0 {
		t.Fatalf("final scan must reset the TON, ET=%d", et())
	}
	// Re-enter: timing must start fresh, not from a stale value.
	run(map[string]ir.Value{"Reset": ir.BoolVal(false), "Go": ir.BoolVal(true)}, 8000) // Home→Run
	if et() != 0 {
		t.Fatalf("fresh re-entry should start timing at 0, ET=%d", et())
	}
	run(nil, 9000) // Run, 1s
	if got := et(); got < 900 || got > 1100 {
		t.Errorf("re-entry timing should be ~1s, ET=%d", got)
	}
}

// TestStepTimerGatesTransition: Step.T via a hidden TON gates a transition once
// the clock advances past the compared bound.
func TestStepTimerGatesTransition(t *testing.T) {
	src := `PROGRAM P
VAR_EXTERNAL Go : BOOL; END_VAR
SFC
  INITIAL_STEP Idle:
  END_STEP
  STEP Wait:
  END_STEP
  STEP Done:
  END_STEP
  TRANSITION t1 FROM Idle TO Wait := Go;
  END_TRANSITION
  TRANSITION t2 FROM Wait TO Done := Wait.T >= T#5S;
  END_TRANSITION
END_SFC
END_PROGRAM`
	prog := mustCompile(t, src)
	h := &clockHost{vals: map[string]ir.Value{}}
	frame := ir.NewFrame(prog)
	run := func(in map[string]ir.Value, now int64) {
		for k, v := range in {
			h.vals[k] = v
		}
		h.now = now
		if err := ir.Run(prog, frame, h); err != nil {
			t.Fatal(err)
		}
	}
	run(map[string]ir.Value{"Go": ir.BoolVal(true)}, 1000) // Idle→Wait; timer starts
	if !stepActive(t, prog, frame, "Wait") {
		t.Fatal("expected Wait active")
	}
	run(map[string]ir.Value{"Go": ir.BoolVal(false)}, 3000) // Wait, .T=2s < 5s
	if stepActive(t, prog, frame, "Done") {
		t.Fatal("must not advance before 5s")
	}
	run(nil, 7000) // .T=6s ≥ 5s (read next scan via pipeline)
	run(nil, 8000)
	if !stepActive(t, prog, frame, "Done") {
		t.Fatal("Step.T should gate the transition once past 5s")
	}
}

// TestWarmMigration: transpile+lower, run 3 scans to move the token, then
// lower a trivially-edited variant and MigrateFrame — the live token position
// must survive (§2.7 warm restart).
func TestWarmMigration(t *testing.T) {
	prog1 := mustCompile(t, tankBatchSrc)
	h := &clockHost{vals: map[string]ir.Value{
		"FillSP": ir.RealVal(100), "HeatSP": ir.RealVal(80), "EmptySP": ir.RealVal(5),
	}}
	frame1 := ir.NewFrame(prog1)
	run := func(in map[string]ir.Value, now int64) {
		for k, v := range in {
			h.vals[k] = v
		}
		h.now = now
		if err := ir.Run(prog1, frame1, h); err != nil {
			t.Fatal(err)
		}
	}
	run(map[string]ir.Value{"Start": ir.BoolVal(true), "Abort": ir.BoolVal(false)}, 1000) // →Fill
	run(map[string]ir.Value{"Start": ir.BoolVal(false), "Level": ir.RealVal(40)}, 2000)   // Fill
	if !stepActive(t, prog1, frame1, "Fill") {
		t.Fatal("token should be at Fill before edit")
	}

	// A trivially-edited variant (an added comment line changes nothing
	// structural but recompiles a fresh program).
	edited := strings.Replace(tankBatchSrc, "  STEP Fill:", "  (* edited *)\n  STEP Fill:", 1)
	prog2 := mustCompile(t, edited)
	frame2, resets := ir.MigrateFrame(prog2, prog1, frame1)
	if !frame2.Slots[slotIdx(prog2, "_S_Fill_X")].B {
		t.Fatalf("warm migration lost the token at Fill (resets=%v)", resets)
	}
	if frame2.Slots[slotIdx(prog2, "_S_Idle_X")].B {
		t.Error("Idle should not be active after migration")
	}
}
