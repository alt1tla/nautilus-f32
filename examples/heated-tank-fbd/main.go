// Command heated-tank-fbd is the heated surge tank controller with its logic
// written as an IEC 61131-3 Function Block Diagram (.fbd) instead of ST. The
// netlist transpiles through lang/fbd at startup and runs on the same runtime
// — one compiler, two source languages.
//
// It exists to exercise the whole FBD toolchain against a live controller:
// open program.fbd in VS Code for syntax highlighting, diagnostics as you
// type, the live diagram preview, the visual diff vs git HEAD, and inline
// live tag values streaming from this process.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/alt1tla/nautilus-f32/runtime"
	"github.com/alt1tla/nautilus-f32/server"
)

//go:embed program.fbd
var program string

//go:embed reports.fbd
var reports string

//go:embed interlocks.ld
var interlocks string

//go:embed blocks.st
var blocks string

func main() {
	// The runtime takes the .fbd source directly — it detects the FBD block,
	// transpiles through lang/fbd, and keeps the ORIGINAL source as the
	// program of record, so online edits and diffs speak .fbd end to end.
	rt, err := runtime.New(runtime.Options{
		Program: program,
		// blocks.st holds ST-authored FUNCTION_BLOCKs callable from either
		// FBD program — composed exactly as the editor/CLI compose the
		// project directory, so online edits round-trip.
		Libraries: []string{blocks},
		Driver:    NewPlant(),
		Scan:      100 * time.Millisecond, // 10 Hz
		DtTag:     "ScanDtS",
		// A second program on the same tag store, at its own rate — the
		// IEC resource/task model. Main owns field I/O; Reports computes
		// operator-facing values at 1 Hz. Open reports.fbd in VS Code:
		// online edits route to THIS task by its PROGRAM name.
		Tasks: []runtime.Task{{
			Name:      "reports",
			Program:   reports,
			Libraries: []string{blocks},
			Scan:      time.Second,
			DtTag:     "RepDtS",
		}, {
			// The annunciator, written as LADDER (interlocks.ld) — three
			// languages, one tag store, each program online-editable.
			Name:    "interlocks",
			Program: interlocks,
			Scan:    200 * time.Millisecond,
		}},
		// One entry per tag: its ROLE in the scan data path, its seed, and
		// its HMI documentation together (GET /api/meta, dashboard table).
		// Inputs are driver-fed (list them in plant.go's ReadInputs too);
		// Setpoints exist from scan one and take HMI/API writes; Outputs go
		// to the driver — Init() the ones the logic reads back.
		Tags: []runtime.TagDef{
			runtime.Input("LevelPct", runtime.Desc("Tank level"), runtime.Unit("%")),
			runtime.Input("TempC", runtime.Desc("Tank temperature"), runtime.Unit("°C")),
			runtime.Input("testExt", runtime.Desc("Wiring demo: plant-produced constant")),
			runtime.Setpoint("TempSP", 65.0, runtime.Desc("Temperature setpoint"), runtime.Unit("°C")),
			runtime.Setpoint("TempSPEco", 55.0, runtime.Desc("Eco-mode temperature setpoint"), runtime.Unit("°C")),
			runtime.Setpoint("EcoMode", false, runtime.Desc("Run at the eco setpoint (SEL in the diagram)")),
			runtime.Setpoint("Kp", 12.0, runtime.Desc("PI proportional gain")),
			runtime.Setpoint("Ki", 0.15, runtime.Desc("PI integral gain"), runtime.Unit("1/s")),
			runtime.Setpoint("PumpStartLevel", 40.0, runtime.Desc("Pump seal-in level"), runtime.Unit("%")),
			runtime.Setpoint("PumpStopLevel", 75.0, runtime.Desc("Pump drop-out level"), runtime.Unit("%")),
			// The seal-in latch READS PumpRun before its first write — Init
			// makes the tag exist on scan one instead of faulting.
			runtime.Output("PumpRun", runtime.Init(false), runtime.Desc("Pump run command")),
			runtime.Output("Heater", runtime.Desc("Heater output command"), runtime.Unit("%")),
			// The Reports task owns these — State seeds them so the HMI
			// sees them before the task's first 1 Hz scan lands.
			runtime.State("TempAvg", 0.0, runtime.Desc("Temperature, 4 s rolling average"), runtime.Unit("°C")),
			runtime.State("TempRate", 0.0, runtime.Desc("Temperature rate of change"), runtime.Unit("°C/min")),
			runtime.State("Status", "", runtime.Desc("Operator status line (built in reports.fbd)")),
			// The ladder task's tags: the horn latch and its operator ack.
			runtime.State("HiTempAlm", false, runtime.Desc("High temperature alarm (interlocks.ld)")),
			runtime.State("Horn", false, runtime.Desc("Annunciator horn (interlocks.ld)")),
			runtime.Setpoint("HornAck", false, runtime.Desc("Horn acknowledge — set from the HMI, releases when alarms clear")),
		},
		Meta: map[string]runtime.TagMeta{
			"ScanDtS":    {Desc: "Measured scan interval", Unit: "s"},
			"RepDtS":     {Desc: "Reports task measured interval", Unit: "s"},
			"TempLowAlm": {Desc: "Low temperature alarm (10 s delay)"},
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "compile:", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go rt.Run(ctx)

	// Tag API for the HMI kit and the VS Code extension's inline live values.
	// NAUTILUS_API overrides the bind address (e.g. localhost:8081 when
	// another controller already owns 8080 — match nautilus.runtimeUrl).
	// OnlineEdits on: this is a dev playground, so PLC-style online edits of
	// the running .fbd program (VS Code "Download Program to Controller",
	// text + graphical controller diffs) are allowed.
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

	banner := "nautilus · heated-tank (FBD) — Ctrl+C to stop"
	if apiUp {
		banner = "nautilus · heated-tank (FBD) — tag API on http://" + apiAddr + " — Ctrl+C to stop"
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
			fmt.Printf("level %5.1f%%  temp %5.1f°C  pump %-3v  heater %3.0f%%  tempLowAlm %-3v  scans %d\n",
				t.Real("LevelPct"), t.Real("TempC"), onOff(t.Bool("PumpRun")),
				t.Real("Heater"), onOff(t.Bool("TempLowAlm")), rt.Stats().Count)
		}
	}
}

func onOff(b bool) string {
	if b {
		return "ON"
	}
	return "off"
}
