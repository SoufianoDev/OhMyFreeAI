# ADR 0005: Go toolchain baseline

- Status: Accepted
- Date: 2026-11-07

## Context

OhMyFreeAI is a new cross-platform Go desktop application. Wails v3 requires Go 1.26 or newer. The selected SQLite driver candidate, modernc.org/sqlite v1.62.0, also requires Go 1.26 or newer. The project should not target an end-of-support Go release merely because an older toolchain happens to be installed in a developer environment.

## Decision

Set the module's minimum Go version to **Go 1.27.3**, the latest stable patch listed by the official Go release history on 2026-11-07. CI resolves its Go version from `go.mod`. The developer doctor reports an older local toolchain but never installs or changes it automatically.

## Consequences

- Local development requires Go 1.27.3 or a newer compatible stable patch.
- The current container has Go 1.23.2, so local Go tests were run against a source-identical mirror with the older module directive. The committed module target itself has not been executed in this container.
- CI is configured for Linux and Windows and will validate the committed Go directive when the workflow runs.
- Re-evaluate the baseline only when a newer stable Go release is available; do not move to prereleases without an explicit decision.

## Sources

- https://go.dev/doc/devel/release
- https://v3.wails.io/installation/
- https://pkg.go.dev/modernc.org/sqlite
