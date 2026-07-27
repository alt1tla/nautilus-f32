# Releasing nautilus

Three artifacts, three version lines, one rule: **the repo is the source of
truth, and a registry may never run ahead of it.** CI's `version-sync` job
enforces the rule on every push; `publish.yml` makes registries catch up.

| Artifact | Version lives in | Ships when | Where |
|---|---|---|---|
| CLI + Go libraries | git tag `v*` | you push the tag | GitHub Release: GoReleaser binaries + a VSIX |
| VS Code extension | `tools/vscode-iec/package.json` | a version bump lands on main | VS Code Marketplace + Open VSX |
| `@joyautomation/nautilus-hmi` | `hmi/package.json` | a version bump lands on main | npm |

There are no manual `vsce publish` or `npm publish` runs anymore — publishing
by hand puts the registry ahead of the repo, which is exactly the drift the
guard rails exist to prevent.

## Ship the extension or the HMI kit

Bump the `version` in the package's `package.json` (extension: update its
CHANGELOG too) in the same PR as the change, and merge. On the push to main,
`publish.yml` compares each package's repo version against its registry:

- versions equal → nothing to do (the workflow is idempotent on every push),
- repo ahead → publish,
- registry ahead → fail loudly: someone published out of band.

**Extension channel convention** (VS Code Marketplace): odd minor =
pre-release channel, even minor = stable. `0.9.x` publishes as pre-release;
the first `0.10.x` bump would be the first stable release.

## Cut a CLI release

```sh
git tag v0.3.2
git push origin v0.3.2
```

That triggers `release.yml`: GoReleaser builds cross-platform CLI binaries
onto a GitHub Release (the tag versions the CLI via ldflags), and a VSIX —
built at the extension's own package.json version — is attached for offline
installs. The tag versions the **Go module and CLI only**; it does not touch
the extension or HMI versions.

## Guard rails

- `version-sync` (in `ci.yml`) fails any push/PR where npm, the Marketplace,
  or Open VSX has a higher version than the corresponding package.json.
- `publish.yml` re-checks the same invariant before publishing.
- Recovering from out-of-band drift: bump the repo package.json past the
  registry version and merge.

## Registry credentials (Settings → Secrets and variables → Actions)

- **`VSCE_PAT`** (secret) — VS Code Marketplace. Azure DevOps org, publisher
  `joyauto` (already in `package.json`), Personal Access Token with
  Marketplace → Manage scope.
- **`OVSX_PAT`** (secret) — Open VSX. Eclipse Foundation account + signed
  Publisher Agreement, `joyauto` namespace, access token.
- **npm — OIDC, no secret.** On npmjs.com, the package's Settings → Trusted
  Publisher must point at repo `joyautomation/nautilus`, workflow
  **`publish.yml`** (npm verifies the workflow *filename*; it was previously
  configured for `release.yml` — update it or CI publishes will be rejected).
  Repository **variable** `PUBLISH_HMI` = `true` arms the publish step.

## Validating changes to the pipeline

`ci.yml` runs `goreleaser check` on every push/PR, so a broken
`.goreleaser.yaml` fails a normal CI run rather than a tagged release. To dry
-run a full build locally without publishing: `goreleaser release --snapshot
--clean`.
