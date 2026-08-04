---
title: Getting started
description: Install the nautilus CLI, scaffold a project, and see a live scan loop with a dashboard in a few minutes.
---

**Prerequisites:** Go 1.24+ with `$(go env GOPATH)/bin` on your `PATH`, and
VS Code for the editor experience.

## 1. Install the CLI

```sh
go install github.com/joyautomation/nautilus/cmd/nautilus@latest
```

This gives you `nautilus new` (scaffold a project), `nautilus check`
(headless Structured Text compile for CI), and `nautilus lsp` (the language
server the VS Code extension uses).

## 2. Scaffold a project

```sh
nautilus new my-plant --no-go     # manifest project: IEC logic + nautilus.yaml, no Go
nautilus new my-plant             # or the Go library form (simulated plant, Go tests)
```

A `--no-go` project is just your logic and a manifest — run and ship it
with the CLI alone:

```sh
cd my-plant
nautilus run        # scan loop + dashboard + tag API on http://localhost:8080
nautilus build      # emit ./my-plant — a self-contained controller binary
```

`nautilus.yaml` declares the tasks (one program file each, any language,
own scan rates), the tags by role, the server, and the field driver.
`nautilus build` emits one deployable binary — no Go toolchain anywhere.

Go is the *extension tier*: custom field buses, simulation physics, and Go
acceptance tests live in the library form, which is the same runtime with
the manifest written as code:

```sh
cd my-plant
go mod tidy      # resolves github.com/joyautomation/nautilus from the proxy
go run .         # scan loop + tag API on http://localhost:8080
go test ./...    # the program's acceptance tests
```

Open **http://localhost:8080** for the built-in live dashboard, or
`GET /api/state` for the raw tag snapshot. Setpoints are click-to-set right in
the tag table — click a value and type, or flip a BOOL with its toggle — so you
can drive the loop before there is an HMI. Inputs and outputs stay read-only:
the driver rewrites an input before every scan and the logic rewrites an output
after, so an edit there would be discarded within a scan.

## 3. Develop in VS Code

Install **nautilus IEC 61131-3** from the
[VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=joyauto.vscode-iec)
or [Open VSX](https://open-vsx.org/extension/joyauto/vscode-iec) —
currently on the **pre-release** channel, so use *Install Pre-Release
Version*. With your project open and the controller running you get compile
diagnostics as you type, go-to-definition, hover, completion, and live tag
values next to identifiers in your program.

## 4. Make it yours

- Write control logic in `program.st` — or `.ld` / `.fbd`; the graphical
  languages open in full diagram editors in VS Code.
- Swap the simulated plant for a real `io.Driver` — EtherNet/IP, Modbus,
  your bus — when you have hardware. The control logic doesn't change.
- Add an HMI with the SvelteKit component kit: faceplates, trends, and an
  SSE realtime client.
- Ship it like any Go binary. The scaffolded CI gates on `go test` and
  `nautilus check`.

## Next

- [The tag model](/guides/tag-model/) — how tags come to exist, which role
  fits which job, and the one rule that bites.
- [Language reference](/reference/functions/) — evaluation semantics and
  every built-in operator, function, and function block.
