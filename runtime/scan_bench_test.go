package runtime_test

import (
	"testing"

	nio "github.com/joyautomation/nautilus/io"
	"github.com/joyautomation/nautilus/runtime"
)

// The heated-tank program: hysteresis latch + PI loop with LIMIT calls —
// the representative "plain ST" scan workload.
const benchST = `PROGRAM Main
VAR_EXTERNAL
    LevelPct       : REAL;
    TempC          : REAL;
    ScanDtS        : REAL;
    TempSP         : REAL;
    Kp             : REAL;
    Ki             : REAL;
    PumpStartLevel : REAL;
    PumpStopLevel  : REAL;
    PumpRun        : BOOL;
    Heater         : REAL;
END_VAR
VAR
    integral : REAL;
    err      : REAL;
END_VAR
IF LevelPct <= PumpStartLevel THEN
    PumpRun := TRUE;
ELSIF LevelPct >= PumpStopLevel THEN
    PumpRun := FALSE;
END_IF;
err := TempSP - TempC;
integral := integral + Ki * err * ScanDtS;
integral := LIMIT(0.0, integral, 100.0);
Heater := LIMIT(0.0, Kp * err + integral, 100.0);
END_PROGRAM`

// A workload heavy on the FB-call and user-FUNCTION paths: builtin timers
// and counters plus a user function, the constructs whose call plumbing
// (frames, arg slices, slot lookups) is where per-scan allocation hides.
const benchFB = `FUNCTION Scale : REAL
VAR_INPUT
    x   : REAL;
    lo  : REAL;
    hi  : REAL;
END_VAR
Scale := lo + x * (hi - lo) / 100.0;
END_FUNCTION

PROGRAM Main
VAR_EXTERNAL
    RunCmd  : BOOL;
    LevelPct : REAL;
    Heater  : REAL;
    Done    : BOOL;
END_VAR
VAR
    t1 : TON;
    t2 : TOF;
    c1 : CTU;
    edge : R_TRIG;
END_VAR
t1(IN := RunCmd, PT := T#500ms);
t2(IN := t1.Q, PT := T#200ms);
edge(CLK := t2.Q);
c1(CU := edge.Q, R := FALSE, PV := 10);
Heater := Scale(x := LevelPct, lo := 0.0, hi := 100.0);
Done := c1.Q;
END_PROGRAM`

func benchSeed() nio.Values {
	return nio.Values{
		"LevelPct": 42.0, "TempC": 61.0, "ScanDtS": 0.1,
		"TempSP": 65.0, "Kp": 8.0, "Ki": 0.5,
		"PumpStartLevel": 30.0, "PumpStopLevel": 80.0,
		"PumpRun": false, "Heater": 0.0,
		"RunCmd": true, "Done": false,
	}
}

// BenchmarkScan measures the full cycle — driver read, VM, driver write,
// stats — the number that decides whether the GC ever has work to do
// between scans.
func BenchmarkScan(b *testing.B) {
	drv := nio.NewMemory()
	_ = drv.WriteOutputs(benchSeed())
	r, err := runtime.New(runtime.Options{
		Program: benchST,
		Driver:  drv,
		Inputs:  []string{"LevelPct", "TempC"},
		Outputs: []string{"PumpRun", "Heater"},
		Seed:    benchSeed(),
		DtTag:   "ScanDtS",
	})
	if err != nil {
		b.Fatal(err)
	}
	r.Scan() // warm: first scan pays one-time costs
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Scan()
	}
}

// BenchmarkProgramRun isolates the VM: one scan of the ST program against
// the tag store, no driver I/O, no stats.
func BenchmarkProgramRun(b *testing.B) {
	r, err := runtime.New(runtime.Options{Program: benchST, Seed: benchSeed()})
	if err != nil {
		b.Fatal(err)
	}
	prog, tags := r.Program(), r.Tags()
	if err := prog.Run(tags); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = prog.Run(tags)
	}
}

// BenchmarkProgramRunFB isolates the VM on the FB/function-call workload.
func BenchmarkProgramRunFB(b *testing.B) {
	r, err := runtime.New(runtime.Options{Program: benchFB, Seed: benchSeed()})
	if err != nil {
		b.Fatal(err)
	}
	prog, tags := r.Program(), r.Tags()
	if err := prog.Run(tags); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = prog.Run(tags)
	}
}
