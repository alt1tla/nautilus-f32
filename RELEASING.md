# Releasing nautilus-f32

This document covers releases of the independent
`github.com/alt1tla/nautilus-f32` fork. The repository originated from
[`joyautomation/nautilus`](https://github.com/joyautomation/nautilus), but
fork releases are owned and versioned independently.

Do not use upstream npm, VS Code Marketplace, Open VSX, GoReleaser, or GitHub
credentials to release this module.

## Supported artifact

The supported release artifact is the Go module:

```text
github.com/alt1tla/nautilus-f32
```

Primary public packages:

```text
github.com/alt1tla/nautilus-f32/lang/stanalysis
github.com/alt1tla/nautilus-f32/lang/st
github.com/alt1tla/nautilus-f32/lang/ir
github.com/alt1tla/nautilus-f32/lang/stgen
```

Inherited CLI, extension, HMI, and runtime directories are not automatically
published as fork artifacts. Publishing them requires a separate ownership,
renaming, credential, and compatibility audit.

## Pre-release checklist

1. Confirm `go.mod` declares:

   ```go
   module github.com/alt1tla/nautilus-f32
   ```

2. Confirm Go source no longer imports the upstream module path:

   ```sh
   rg "github.com/joyautomation/nautilus" --glob "*.go"
   ```

3. Keep the upstream `LICENSE` and attribution.
4. Update `README.md` and `docs/embedding.md` for public API changes.
5. Review the working tree for unrelated or generated changes.

## Validate

Minimum supported suite:

```sh
go mod tidy
go test ./lang/st ./lang/ir ./lang/stanalysis ./lang/stgen
```

When the environment supports inherited tools and fixtures, also run:

```sh
go test ./...
```

Record environment-specific failures separately. Never claim that the full
suite passed when only selected packages ran.

## Version policy

Use semantic versions:

- patch: compatible fixes and diagnostic improvements;
- minor: compatible public APIs or ST features;
- major: incompatible exported APIs, IR contracts, or target semantics.

Use `v0.x.y` while the API is stabilizing and document breaking changes even
during `v0`.

## Create a release

Commit and push the intended state, then create an annotated tag:

```sh
git status
git add .
git commit -m "prepare v0.1.0"
git push origin main

git tag -a v0.1.0 -m "nautilus-f32 v0.1.0"
git push origin v0.1.0
```

No separate Go registry is required. A public GitHub repository and semantic
Git tag are sufficient.

Verify from a clean consumer:

```sh
go get github.com/alt1tla/nautilus-f32@v0.1.0
go list -m github.com/alt1tla/nautilus-f32
```

## Test before tagging

Remote commit:

```sh
go get github.com/alt1tla/nautilus-f32@<commit-hash>
```

Go records a pseudo-version. For local development, prefer:

```go
require github.com/alt1tla/nautilus-f32 v0.0.0

replace github.com/alt1tla/nautilus-f32 => ../nautilus-f32
```

## Release notes

Include public API changes, `Float32Scalar` changes, recognized ST features,
IR migrations, known limitations, tests executed, and a statement that this
is an independent fork.

## Attribution

Do not remove the Apache 2.0 license or imply that the fork was created from
scratch. Retain applicable upstream notices when copying further work. Do not
imply that upstream maintainers support, publish, or endorse a fork release.
