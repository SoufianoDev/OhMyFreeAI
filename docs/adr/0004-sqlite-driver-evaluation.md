# ADR 0004: SQLite driver evaluation

- Status: Proposed
- Date: 2026-11-07

## Context

SQLite is required for local application metadata, persisted configuration, migrations and bounded operational history. Linux and Windows are first-class targets. A CGO dependency can complicate cross-platform builds and packaging.

## Options reviewed

| Driver | License | Build model | Current toolchain implication | Main concern |
|---|---|---|---|---|
| modernc.org/sqlite v1.62.0 | BSD-3-Clause | Pure Go, no CGO | pkg.go.dev metadata requires Go 1.26 | Must validate current Windows/Linux file locking and workload behavior |
| mattn/go-sqlite3 v1.14.x | MIT | C SQLite through CGO | Requires CGO and a C compiler; Windows requires a GCC toolchain | Cross-build and packaging complexity |
| ncruces/go-sqlite3 v0.37.0 | MIT | SQLite WASM translated to Go | pkg.go.dev metadata requires Go 1.26 | Higher per-connection memory and custom VFS/file-locking/WAL behavior require testing |

## Decision

Use `database/sql` at the adapter boundary. Evaluate `modernc.org/sqlite` first because it avoids CGO and fits the Go 1.26 baseline needed by Wails v3. This is a provisional evaluation order, not an approved dependency selection. Do not add or pin a driver until the proof matrix passes.

## Required proof matrix

- Linux amd64: open/create, migrations, foreign keys, WAL, busy timeout, concurrent readers/writers, backup/restore, crash/reopen.
- Windows amd64: same matrix, including WAL and process shutdown/restart.
- Verify SQLite version, driver license and transitive dependency licenses.
- Confirm DB files live on local filesystem paths with supported locking semantics.
- Verify clean uninstall preserves or explicitly offers to export user data.

## Sources

- https://pkg.go.dev/modernc.org/sqlite
- https://github.com/mattn/go-sqlite3
- https://pkg.go.dev/github.com/ncruces/go-sqlite3
