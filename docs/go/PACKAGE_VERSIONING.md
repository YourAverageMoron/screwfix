# Go Module Versioning

Rules for versioning Go modules. Follows [Semantic Versioning 2.0.0](https://semver.org/) and the [Go module versioning system](https://go.dev/doc/modules/version-numbers).

## Principles

1. **Public modules are immutable.** Once a version is tagged and published, never move or rewrite the tag. Publish a new version instead.
2. **Versions come from git tags**, not source code. A tag `v1.2.3` on the module root maps to module version `v1.2.3`.
3. **Semver only.** `vMAJOR.MINOR.PATCH` (+ optional prerelease/build metadata). No invented schemes.

## When to Bump

| Change                                   | Bump    |
|------------------------------------------|---------|
| Backwards-incompatible API change        | MAJOR   |
| New feature, backwards-compatible        | MINOR   |
| Bug fix, backwards-compatible            | PATCH   |
| Pre-release / unstable work              | prerelease (e.g. `v1.0.0-alpha.1`) |

- A breaking change commit **must** use the `BREAKING CHANGE:` footer or `!` per [Commit Messages](../git/COMMIT_MESSAGES.md). A MAJOR bump must accompany it.
- For modules still at `v0.x`, any change may be breaking — bump MINOR for features, PATCH for fixes.

## Major Versions (v2+)

Go requires the major version to appear in the module path from `v2` onward.

- `v1` and `v0`: module path unchanged (`github.com/org/name`).
- `v2+`: append `/vN` to the module path (`github.com/org/name/v2`).
- Create a new major version only when you must break the API. Prefer additive change.
- Publish a new major in a new subdirectory or a new branch — never overwrite the old major's tag.

## Tagging Rules

- Tag the commit on the default branch (or the release branch) with `git tag v1.2.3`.
- Use annotated tags: `git tag -a v1.2.3 -m "Release v1.2.3"`.
- Tag the **module root**, the commit whose `go.mod` declares the module being released.
- One tag per release. Never tag the same version twice.
- Prereleases: `v1.2.0-rc.1`, `v1.2.0-beta.2` — semver prerelease suffixes, consumed by `go get` as lower precedence.

## Private / Internal Modules

- Modules under `internal/` are not versioned as public APIs. Do not tag internal packages independently.
- A single module tag versions the whole module; internal packages ride along.

## Checklist Before Tagging a Release

1. `go fmt ./...`, `go vet ./...`, `golangci-lint run ./...` pass — see [Linting](./LINTING.md).
2. `go test ./...` (and `-race` for concurrent paths) passes.
3. `go mod tidy` run; `go.mod`/`go.sum` committed and consistent.
4. Public API changes are reflected in a MINOR or MAJOR bump; breaking changes have a `BREAKING CHANGE:` footer.
5. CHANGELOG updated (where maintained).
6. Tag annotated, on the module root commit, pushed: `git push origin v1.2.3`.

## Don'ts

- Don't force-push or delete published tags.
- Don't tag a dirty or untested tree.
- Don't bump the version for a change with no consumer-visible effect (CI tweaks, internal refactors) — that's a PATCH at most, often no release.
- Don't skip the `/vN` suffix for `v2+` modules; `go get` will not resolve it.

## References

- Go module version numbers: <https://go.dev/doc/modules/version-numbers>
- Go module major version suffixing: <https://go.dev/doc/modules/major-version>
- Semantic Versioning 2.0.0: <https://semver.org/>
- Breaking-change commit format: [Commit Messages](../git/COMMIT_MESSAGES.md)
- Pre-release CI gates: [Linting](./LINTING.md)
