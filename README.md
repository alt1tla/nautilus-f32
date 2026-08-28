# nautilus-f32

`nautilus-f32` is a Go library for parsing, validating, inspecting, and
lowering IEC 61131-3 Structured Text (ST) into typed intermediate
representation (IR). It is intended for compiler backends, WebSocket services,
editors, and language servers.

This repository is an independent fork of
[`joyautomation/nautilus`](https://github.com/joyautomation/nautilus), a
broader SCADA/runtime toolkit. This fork is maintained by `alt1tla` and
focuses on the ST frontend and a float32-oriented target model. It is not an
official upstream release and is not affiliated with the upstream Marketplace
or npm publishing accounts.

The upstream Apache 2.0 license and attribution are preserved in
[`LICENSE`](LICENSE).

## Public scope

| Package | Purpose |
| --- | --- |
| `lang/stanalysis` | Public diagnostics, AST, IR, symbols, and language catalog API |
| `lang/st` | ST lexer, parser, semantic validation, and lowering |
| `lang/ir` | Typed IR, builtins, and function-block definitions |
| `lang/stgen` | Programmatic generation of ST type declarations |

Inherited runtime, FBD, LD, SFC, HMI, server, and driver code remains in the
repository but is outside this fork's current development scope.

## Install

```sh
go get github.com/alt1tla/nautilus-f32@latest
```

For local development, add to the consuming project's `go.mod`:

```go
require github.com/alt1tla/nautilus-f32 v0.0.0

replace github.com/alt1tla/nautilus-f32 => ../nautilus-f32
```

Then run `go mod tidy`. Go 1.24 or newer is required.

## Analyze ST

```go
import "github.com/alt1tla/nautilus-f32/lang/stanalysis"

result := stanalysis.Analyze(source, stanalysis.Options{
    ScalarMode: stanalysis.Float32Scalar,
    Prelude:    librarySource,
})

if !result.Valid() {
    for _, diagnostic := range result.Diagnostics {
        // diagnostic.Stage, Message, Pos.Line, Pos.Col
    }
    return
}

ast := result.AST
contextAST := result.ContextAST
programIR := result.IR
symbols := result.Symbols
```

- `AST` describes the submitted document.
- `ContextAST` also includes optional prelude/library source.
- `IR` is present only after successful syntax and semantic analysis.
- diagnostics use 1-based positions.
- symbols belong to the submitted document, not its prelude.

Both live editor diagnostics and a manual compiler-readiness check should call
this same API. WebSocket handlers add only transport concerns such as debounce,
document versions, cancellation, and JSON conversion.

## Scalar modes

`IECStrict` preserves upstream IEC type boundaries.

`Float32Scalar` models a target where `BOOL`, integer, `REAL`, and `TIME`
values share a float32-oriented representation at assignment and call
boundaries:

- zero is false and any nonzero scalar is true;
- numeric `TIME` is measured in milliseconds;
- explicit conversion functions are unnecessary between these scalar kinds;
- semantic IR types remain available to the backend;
- strings, arrays, structures, and FB instances remain distinct.

The inherited VM still carries `REAL` in its existing representation. This
mode is primarily a frontend/target contract for an embedding compiler.

## Language catalog

Frontends should request the supported surface instead of maintaining another
hard-coded list:

```go
catalog := stanalysis.BuildCatalog()
payload, err := json.Marshal(catalog)
```

The deterministic catalog contains ST operators, stateless functions with
arity/type metadata, and function blocks with public input/output pins. It
includes `TON`, `TOF`, `TP`, `SR`, `RS`, `R_TRIG`, and `F_TRIG`. Functions and
FBs registered by an embedding application appear automatically.

## Supported ST surface

The parser/lowerer supports program declarations, variables, types, arrays,
structures, functions, function blocks, expressions, assignments, `IF`,
`CASE`, `FOR`, `WHILE`, and `REPEAT`.

`AND`, `OR`, `XOR`, and `MOD` retain operator forms and can also be called:

```st
Logic := AND(A, OR(B, C));
Remainder := MOD(A, B);
```

The float32 bit bridge is:

```st
Bit := FTB(Value, BitIndex);
Value := BTF(Bit0, Bit1, Bit2, Bit3);
```

`FTB` reads IEEE-754 float32 bit `0..31` (bit 0 is least significant).
`BTF` rebuilds a float32 from 1–32 low-to-high bits; omitted high bits are zero.

See [`docs/embedding.md`](docs/embedding.md) for WebSocket integration.

## Development

```sh
go test ./lang/st ./lang/ir ./lang/stanalysis ./lang/stgen
```

Run `go test ./...` when changing shared inherited code. Some inherited tests
require external tools, generated files, platform-specific timing, or specific
line endings; report those failures separately.

## Versioning

```sh
git tag -a v0.1.0 -m "nautilus-f32 v0.1.0"
git push origin v0.1.0
```

See [`RELEASING.md`](RELEASING.md).

## Upstream and license

This fork derives from `joyautomation/nautilus`. Inherited names, assets,
examples, extension metadata, and HMI metadata may remain in their original
directories; their presence does not mean this fork controls those artifacts.

Fork repository: <https://github.com/alt1tla/nautilus-f32>

Licensed under Apache License 2.0. See [`LICENSE`](LICENSE).
