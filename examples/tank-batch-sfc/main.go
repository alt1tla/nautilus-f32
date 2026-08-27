// Command tank-batch-sfc is a batch tank controller with its logic written
// as an IEC 61131-3 Sequential Function Chart (.sfc) — nautilus's fourth
// and last IEC language, alongside ST, FBD, and LD. The chart transpiles to
// ST through lang/sfc at startup and runs on the same runtime as every
// other language: one compiler, four source languages.
//
// It exists to exercise the whole SFC toolchain against a live controller:
// open program.sfc in VS Code for syntax highlighting and diagnostics as
// you type, the live diagram preview, and inline live tag values streaming
// from this process — the same experience examples/heated-tank-fbd gives
// FBD.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/alt1tla/nautilus-f32/runtime"
	"github.com/alt1tla/nautilus-f32/server"
)

//go:embed program.sfc
var program string

func main() {
	// The runtime takes the .sfc source directly — it detects the SFC
	// block, transpiles through lang/sfc, and keeps the ORIGINAL source as
	// the program of record, so online edits and diffs speak .sfc end to
	// end (step activity, timers, and the S/R abort latch all survive a
	// swap by name — docs/design/sfc.md §2.7).
	rt, err := runtime.New(runtime.Options{
		Program: program,
		Driver:  NewPlant(),
		Scan:    100 * time.Millisecond, // 10 Hz
		// One entry per tag: its ROLE in the scan data path, its seed, and
		// its HMI documentation together (GET /api/meta, dashboard table).
		Tags: []runtime.TagDef{
			runtime.Input("Level", runtime.Desc("Tank level"), runtime.Unit("%")),
			runtime.Input("TempC", runtime.Desc("Tank temperature"), runtime.Unit("°C")),
			// The auto-operator: plant.go asserts this whenever the chart
			// is idle so the batch cycles forever without a human.
			runtime.Input("Start", runtime.Desc("Auto-operator start command (plant.go)")),
			runtime.Setpoint("Abort", false, runtime.Desc("Abort the batch (wired, never asserted by the auto-operator)")),
			runtime.Setpoint("FillSP", 80.0, runtime.Desc("Fill target level"), runtime.Unit("%")),
			runtime.Setpoint("EmptySP", 10.0, runtime.Desc("Drain target level"), runtime.Unit("%")),
			runtime.Setpoint("HeatSP", 70.0, runtime.Desc("Heat target temperature"), runtime.Unit("°C")),
			runtime.Output("FillValve", runtime.Desc("Fill valve command")),
			runtime.Output("DrainValve", runtime.Desc("Drain valve command")),
			runtime.Output("Heater", runtime.Desc("Heater command")),
			runtime.Output("Mixer", runtime.Desc("Mixer command")),
			runtime.Output("RunLamp", runtime.Desc("Chart running indicator")),
			runtime.Output("AbortLamp", runtime.Desc("Latched abort indicator (S/R demo)")),
			// BatchCount is pure controller state: the P1 CountBatch action
			// reads-then-writes it every batch, so it needs to exist (and
			// survive) from scan one — a State tag, not an Output.
			runtime.State("BatchCount", int64(0), runtime.Desc("Completed batch count")),
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "compile:", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go rt.Run(ctx)

	// Tag API for the HMI kit and the VS Code extension's inline live
	// values. NAUTILUS_API overrides the bind address (e.g.
	// localhost:8081 when another controller already owns 8080 — match
	// nautilus.runtimeUrl). OnlineEdits on: this is a dev playground, so
	// PLC-style online edits of the running .sfc program (VS Code
	// "Download Program to Controller") are allowed — the live token
	// (step activity) survives the swap.
	srv := server.New(rt, server.Options{
		AuthToken:   os.Getenv("NAUTILUS_TOKEN"),
		OnlineEdits: true,
	})
	go srv.Run(ctx)
	apiAddr := os.Getenv("NAUTILUS_API")
	if apiAddr == "" {
		apiAddr = "localhost:8080"
	}
	apiUp := false
	if ln, err := net.Listen("tcp", apiAddr); err != nil {
		fmt.Fprintf(os.Stderr, "tag api: %v (continuing without it)\n", err)
	} else {
		apiUp = true
		go func() {
			if err := http.Serve(ln, srv.Handler()); err != nil && ctx.Err() == nil {
				fmt.Fprintln(os.Stderr, "tag api:", err)
			}
		}()
	}

	banner := "nautilus · tank-batch (SFC) — Ctrl+C to stop"
	if apiUp {
		banner = "nautilus · tank-batch (SFC) — tag API on http://" + apiAddr + " — Ctrl+C to stop"
	}
	fmt.Println(banner)
	t := rt.Tags()
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nstopped.")
			return
		case <-tick.C:
			fmt.Printf("step %-9s level %5.1f%%  temp %5.1f°C  fill %-3v  heater %-3v  mixer %-3v  drain %-3v  batches %d\n",
				activeSteps(rt), t.Real("Level"), t.Real("TempC"),
				onOff(t.Bool("FillValve")), onOff(t.Bool("Heater")), onOff(t.Bool("Mixer")), onOff(t.Bool("DrainValve")),
				batchCount(rt))
		}
	}
}

// activeSteps reads the chart's retained step-activity slots
// (_S_<Step>_X — docs/design/sfc.md §2.2) straight out of AllLocals, the
// same watch surface the LSP/editor use for inline live values, and joins
// whichever are currently TRUE ("Heat+Mix" during the simultaneous branch).
func activeSteps(rt *runtime.Runtime) string {
	steps := []string{"Idle", "Fill", "Heat", "Mix", "Drain", "Aborted"}
	locals := rt.AllLocals()
	var active []string
	for _, s := range steps {
		if b, _ := locals["_S_"+s+"_X"].(bool); b {
			active = append(active, s)
		}
	}
	if len(active) == 0 {
		return "?"
	}
	return strings.Join(active, "+")
}

// batchCount reads BatchCount as an int64 regardless of the exact numeric
// representation All() hands back.
func batchCount(rt *runtime.Runtime) int64 {
	switch v := rt.Tags().All()["BatchCount"].(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func onOff(b bool) string {
	if b {
		return "ON"
	}
	return "off"
}
