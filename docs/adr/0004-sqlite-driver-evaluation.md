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


## Evaluation harness

A separate nested Go module at `tests/sqlite-driver-eval` keeps experimental driver dependencies out of the production module. The GitHub Actions matrix runs it on Ubuntu 24.04 and Windows for:

- `modernc.org/sqlite` v1.62.0 (driver name `sqlite`)
- `github.com/ncruces/go-sqlite3` v0.37.0 (driver name `sqlite3`)

The shared smoke test checks open/ping, SQLite version reporting, WAL activation, per-connection foreign-key enforcement, transaction rollback, concurrent inserts across four held connections with a busy timeout, integrity_check, WAL checkpoint, close/reopen and a checkpointed database-file copy. It does not yet simulate abrupt process termination, test the complete production migration runner, or establish long-duration stress/performance results. The test is evidence for narrowing the driver choice, not a complete acceptance certificate.

The workflow is added but must not be treated as passed until both OS jobs complete successfully. Driver selection remains Proposed until results and a follow-up review of locking, dependency/license inventory, packaging and recovery are recorded.


## CI bootstrap note

The first evaluation run stopped at module setup because the nested module had no committed `go.sum`. This was corrected in the workflow by running `go mod tidy` and `go mod verify` before the test command. Until a run using that corrected workflow passes, the earlier failure is classified as test-workflow setup failure, not a driver compatibility result.
