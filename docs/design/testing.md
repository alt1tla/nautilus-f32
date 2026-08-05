# Design: acceptance testing for manifest projects

Status: **brief** — grounding and constraints only. The interface is not designed yet; that is the next session's job.
Author: feasibility spike, 2026-08-05
Scope: gives a `--no-go` project a way to assert on its control logic, and gives the runtime the deterministic clock that any such mechanism needs.

---

## 0. Why this exists

The product direction is that the **manifest form is nautilus**, and Go becomes the extension tier for custom field buses and richer simulation. Everything else already supports that: the manifest is close to a strict subset of `runtime.Options` in expressive power (multi-task, per-task scan rates, `dt-tag`, tag roles, units/descriptions, library composition, EtherNet/IP), and what is genuinely manifest-only is *lifecycle*, not capability — `nautilus run` / `nautilus build` need no toolchain, and the online-edit → `nautilus pull` → commit loop closes without a rebuild because the files are what the runtime loads.

Two things stand between the manifest tier and being the whole product: **custom drivers** and **tests**. Custom drivers are a legitimate SDK concern where Go belongs. Tests are not.

The stake is concrete. The pitch against vendor tooling is that PLC code has no tests and no pipeline that runs one. If the default nautilus project answers that with a compile check and a `git diff`, the argument is materially weaker than it should be. Today:

- `nautilus test` does not exist (`cmd/nautilus/main.go`, command switch).
- The scaffolded CI for a manifest project runs `nautilus check` + `nautilus build` and nothing else; the Go branch is the only one with `go test ./...` (`cmd/nautilus/templates/ci.yml.tmpl`).
- Acceptance testing exists only as `cmd/nautilus/templates/program_test.go.tmpl`, in the Go tier.

## 1. The blocking finding: the runtime has no virtual time

**This is the reason a test format alone would not solve the problem.** It is not specific to the manifest tier — the existing Go tests have the same hole.

There are two independent clocks, both wall-clock:

1. **Scan `dt`** — `runtime/runtime.go:358` `Scan()` computes
   `dt := t0.Sub(r.lastScan).Seconds()`, falling back to the configured target only on the first scan (`runtime.go:359-365`). `scanTask` does the same per task (`runtime.go:328-336`). This feeds `DtTag`, so it drives every PI loop, integrator, and the ST plant simulation.
2. **IEC timers** — `TON`/`TOF`/`TP` read `ctx.NowMs` (`lang/ir/builtins_fb.go:80,125,163`), supplied via the `Host` interface (`lang/ir/vm.go:10,202`) and implemented as `func (t *Tags) NowMs() int64 { return time.Now().UnixMilli() }` (`runtime/tags.go:44`).

Consequence: a test that loops `rt.Scan()` 100 times simulates a few hundred microseconds of process time, not 10 seconds. So none of the following can be asserted today, in either tier:

- the 10-second `TempLowAlm` delay in `examples/heated-tank-nogo`
- PI settling (measured: a `TempSP` 65 → 72 °C step settles in ~32 s, heater saturating at 100 % for ~26 s before backing off)
- any `TON`/`TOF`/`TP`, i.e. most real interlock logic

This is why `program_test.go.tmpl` only ever asserts *direction* ("heater should drive on cold error, got > 0") and never a delay, a timer, or a convergence. The tests were written to fit what is testable.

**Both clocks already have a seam**, which is the good news:

- `NowMs()` is already an interface method on `ir.Host`. Making the implementation swappable is small and touches no VM code.
- Scan `dt` is computed inline and needs a real (small) change to accept an injected step.

Get both and `advance: 12s` becomes deterministic *and* instant.

## 2. Second constraint: multi-task scheduling is non-deterministic

`Run` starts one goroutine per additional task, each with its own `time.Ticker`, serialized only by the per-scan lock (`runtime/runtime.go:286-311`). `heated-tank-nogo` has four tasks at 100 ms / 100 ms / 1 s / 200 ms.

A test harness therefore **must not use `Run`**. It needs its own virtual scheduler that replays tick order deterministically (next-due-time, ties broken by declared order). The primitives already exist: `Scan()` for the main task and `ScanTask(name)` (`runtime.go:315`), whose doc comment already says "for tests and custom schedulers."

Worth stating explicitly in the design: a deterministic interleaving is what makes tests reproducible, and it also means **tests cannot prove the absence of races between tasks**. That is a deliberate trade, not an oversight, and it should be written down.

## 3. Recommendation to argue with, not to assume

**A declarative test format, not a DSL.**

The entire Go test reduces to three primitives — set inputs, scan, assert tags — and virtual time adds a fourth. That is a data shape, not a language:

```yaml
tests:
  - name: pump latches on below start level
    given:   { LevelPct: 35.0 }
    scans:   1
    expect:  { PumpRun: true }

  - name: low-temp alarm after 10 s
    given:   { TempC: 45.0 }
    advance: 12s
    expect:  { TempLowAlm: true }
```

Reasons:

- The assertion surface is tiny and regular; a DSL buys expressiveness that is not needed.
- A DSL is a grammar, a parser, diagnostics, **and LSP support**. The extension already carries four IEC languages; a fifth bespoke one is permanent cost.
- YAML tests diff in git, which is the entire argument the product makes.
- It fits the existing frame: `nautilus.yaml` declares the system, so tests are more declaration rather than a new tier.

**Rejected for now, with the door left open:** a bespoke DSL for sequencing ("run the batch sequence, force a fault at step 3, assert it aborts"). If that becomes necessary, the better answer is probably **test POUs written in ST**, reusing the existing compiler, LSP, and editors, rather than a new grammar.

## 4. Open questions for the design session

1. **Where do tests live?** A `tests:` key in `nautilus.yaml`, or separate `*_test.yaml` files? Separate files scale and keep the manifest readable; a manifest key keeps one source of truth.
2. **What is the time model?** Fixed-step (`scans: N` at the configured rate) only, or wall-clock-shaped (`advance: 12s`) mapped onto ticks? The second is what people want to write; the first is what the runtime does.
3. **`given` — what does it actually set?** Driver inputs, tag writes, or both? The Go test uses `drv.WriteOutputs` for inputs and `Seed` for setpoints; those are different seams and the format should not blur them.
4. **How do assertions handle REAL?** A tolerance is mandatory. Default epsilon, or required explicit tolerance?
5. **Eventually-style assertions.** `advance: 12s; expect: {...}` covers "after"; is there a need for "within 12 s, at some point"?
6. **Does the sim task run during tests?** `heated-tank-nogo` simulates the plant in ST. A test of the control logic may want it running (closed loop) or frozen (open loop, inject `TempC` directly). Probably both, selectable.
7. **What does `nautilus test` output?** TAP, Go-style, or its own. It has to be readable in CI logs and in the VS Code extension.
8. **Does the Go tier converge on this?** A Go project could run the same YAML through the same harness, which would give one test story instead of two.
9. **Injected-clock blast radius.** Sparkplug timestamps, scan diagnostics (`periods`, `jitterMs`), and the SSE `ts` field all read real time. Decide which follow virtual time and which do not.

## 5. Non-goals

- Replacing the Go acceptance tests. They keep working; this is about the manifest tier having an answer at all.
- Testing the *physics*. `sim.st` and `plant.go` are fixtures, not subjects.
- Hardware-in-the-loop or driver conformance testing.

## 6. Verification fixture

`examples/heated-tank-nogo` is the right target: four tasks, three languages, a plant simulated in ST, a 10-second alarm delay, and a PI loop with a measured 32-second step response. If the design can express and deterministically pass tests for that alarm delay and that step response, it is sufficient. If it cannot, it has reproduced the current hole in a nicer syntax.

## 7. Reference points

| What | Where |
|---|---|
| Reference test shape | `cmd/nautilus/templates/program_test.go.tmpl` |
| Scan loop + `dt` | `runtime/runtime.go:358` (`Scan`), `:328` (`scanTask`) |
| Task scheduling | `runtime/runtime.go:286-311` (`Run`), `:315` (`ScanTask`) |
| Timer clock | `runtime/tags.go:44`, `lang/ir/vm.go:10,202`, `lang/ir/builtins_fb.go` |
| Manifest schema + load | `internal/project/project.go` (`Project`, `Load`) |
| Scaffolded CI | `cmd/nautilus/templates/ci.yml.tmpl` |
| CLI command switch | `cmd/nautilus/main.go` |
