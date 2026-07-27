package main

import (
	"math"
	"sync"
	"time"

	nio "github.com/joyautomation/nautilus/io"
)

// Plant is a first-order batch-tank simulation, the same shape as
// examples/heated-tank-fbd's Plant: it consumes the controller's field
// outputs (the two valves, the heater, the mixer) and produces the field
// inputs (level, temperature) by integrating simple physics over the
// elapsed wall-clock time each scan. A real deployment swaps this for a
// Modbus/OPC-UA/EtherNet-IP driver — the SFC chart is unchanged.
//
// Plant also plays the auto-operator: TankBatch has no human in the loop,
// so ReadInputs asserts Start whenever it sees the batch is idle (every
// field output off) and lets it drop out on its own the instant the chart
// leaves Idle and energizes FillValve. Abort is left wired (an operator
// could still POST it true over the tag API) but nothing here ever sets it.
type Plant struct {
	mu    sync.Mutex
	level float64 // tank level, %
	tempC float64 // tank temperature, °C

	fillValve  bool
	drainValve bool
	heater     bool
	mixer      bool

	last time.Time
}

const (
	fillPctPerS  = 30.0 // FillValve inflow rate
	drainPctPerS = 25.0 // DrainValve outflow rate
	heatCPerS    = 20.0 // Heater ramp rate
	ambientLoss  = 0.15 // cooling coefficient toward ambient (per °C, per s)
	ambientC     = 20.0 // ambient/room temperature
)

// NewPlant starts with a shallow heel of room-temperature liquid in the
// tank — a believable "just drained" state so the very first batch fills
// from near-empty exactly like every batch after it.
func NewPlant() *Plant {
	return &Plant{level: 10, tempC: ambientC}
}

// WriteOutputs receives the controller's commands.
func (p *Plant) WriteOutputs(v nio.Values) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if b, ok := v["FillValve"].(bool); ok {
		p.fillValve = b
	}
	if b, ok := v["DrainValve"].(bool); ok {
		p.drainValve = b
	}
	if b, ok := v["Heater"].(bool); ok {
		p.heater = b
	}
	if b, ok := v["Mixer"].(bool); ok {
		p.mixer = b
	}
	return nil
}

// ReadInputs steps the physics by the elapsed time and reports the
// transmitters, plus the auto-operator's Start command.
func (p *Plant) ReadInputs() (nio.Values, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	dt := 0.1
	if !p.last.IsZero() {
		dt = math.Min(now.Sub(p.last).Seconds(), 0.5) // cap so a hitch can't blow up Euler
	}
	p.last = now

	if p.fillValve {
		p.level = clamp(p.level+fillPctPerS*dt, 0, 100)
	}
	if p.drainValve {
		p.level = clamp(p.level-drainPctPerS*dt, 0, 100)
	}

	if p.heater {
		p.tempC += heatCPerS * dt
	} else {
		p.tempC -= ambientLoss * (p.tempC - ambientC) * dt
	}
	p.tempC = clamp(p.tempC, ambientC, 120)

	// The batch is away from Idle (or stuck in the dead-end Aborted step)
	// exactly when one of these four is energized — none of them are ever
	// true while Idle/Aborted are active. Assert Start until the chart
	// notices; the moment it does, FillValve comes on next scan and Start
	// clears itself here without the controller needing to tell us.
	batchActive := p.fillValve || p.drainValve || p.heater || p.mixer
	start := !batchActive

	return nio.Values{
		"Level": p.level,
		"TempC": p.tempC,
		"Start": start,
	}, nil
}

func clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, v))
}
