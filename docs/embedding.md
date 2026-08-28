# Embedding the Structured Text frontend

This guide covers embedding `github.com/alt1tla/nautilus-f32` in another Go
application. The library is transport-independent: WebSocket, HTTP, JSON-RPC,
and LSP adapters belong in the consuming project.

## Install

```sh
go get github.com/alt1tla/nautilus-f32@latest
```

For a local checkout:

```go
require github.com/alt1tla/nautilus-f32 v0.0.0

replace github.com/alt1tla/nautilus-f32 => ../nautilus-f32
```

## Common analysis function

```go
package compilerapi

import "github.com/alt1tla/nautilus-f32/lang/stanalysis"

func AnalyzeST(source, librarySource string) stanalysis.Result {
    return stanalysis.Analyze(source, stanalysis.Options{
        ScalarMode: stanalysis.Float32Scalar,
        Prelude:    librarySource,
    })
}
```

Always check `Valid()` before giving IR to a backend:

```go
result := AnalyzeST(source, library)
if !result.Valid() {
    return result.Diagnostics
}
compile(result.IR)
```

| Field | Available | Meaning |
| --- | --- | --- |
| `AST` | after parsing | submitted document AST |
| `ContextAST` | after parsing | AST including optional prelude |
| `IR` | only when valid | typed compiler input |
| `Diagnostics` | on errors | errors with 1-based positions |
| `Symbols` | after parsing | declarations from the submitted document |

Syntax failure returns neither AST nor IR. Semantic failure retains AST and
symbols but returns no IR.

## WebSocket methods

Use separate protocol methods but the same analysis function.

Live diagnostics request:

```json
{
  "method": "textDocument/analyze",
  "documentId": "main.st",
  "version": 17,
  "source": "PROGRAM Main ..."
}
```

Response:

```json
{
  "method": "textDocument/diagnostics",
  "documentId": "main.st",
  "version": 17,
  "valid": false,
  "diagnostics": []
}
```

Manual compiler check:

```json
{
  "method": "compiler/check",
  "requestId": "check-42",
  "source": "PROGRAM Main ..."
}
```

The live path should debounce and discard stale document versions. The manual
check should run immediately and may continue into backend-specific validation.
Do not implement two ST validators: both paths call `stanalysis.Analyze`.

## Diagnostic positions

Library positions are 1-based. LSP positions are 0-based, so an LSP adapter
subtracts one and expands the position to an identifier or line range using the
original source.

Stages are `stanalysis.StageSyntax` and `stanalysis.StageSemantic`.

## Language catalog endpoint

Expose the registry-derived catalog instead of duplicating names:

```go
func CatalogJSON() ([]byte, error) {
    return json.Marshal(stanalysis.BuildCatalog())
}
```

Suggested request:

```json
{ "method": "language/catalog" }
```

The response contains sorted `operators`, `functions`, and `functionBlocks`
with public FB pins.

## Backend guidance

In `Float32Scalar` mode:

- boolean context compares a scalar with zero;
- `TIME` is measured in milliseconds;
- real target arithmetic uses float32;
- each FB instance requires independent state;
- arrays and structures preserve their declared shape.

`FTB(value, index)` and `BTF(bits...)` use IEEE-754 float32 bit order. Index 0
is least significant; valid indexes are `0..31`.

`VAL(reference)` is a target intrinsic with signature `STRING -> REAL`. A
backend should recognize the IR call named `VAL` and translate its string
literal, for example `VAL('Controller1.Temperature')`, into its controller
access mechanism. The inherited VM intentionally refuses to execute it.

## Concurrency

`Analyze` creates per-call parser/lowerer state. Complete mutations through
`ir.RegisterBuiltin` or `ir.RegisterFB` during application initialization,
before concurrent analysis/catalog requests.

## Fork status

`nautilus-f32` is maintained independently from
[`joyautomation/nautilus`](https://github.com/joyautomation/nautilus). The
public embedding API and float32 behavior are fork-specific guarantees.
