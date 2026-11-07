# ADR 0004: SQLite driver evaluation

- Status: Accepted
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

Use `database/sql` at the persistence adapter boundary and select **`modernc.org/sqlite v1.62.0`** as the production SQLite driver.

Both modernc v1.62.0 and ncruces v0.37.0 passed the evaluation tests on Ubuntu 24.04 and Windows. The tests applied the actual `migrations/001_initial.sql` in a transaction and exercised WAL, foreign keys, rollback, concurrent writes, integrity checks, WAL checkpoint, close/reopen and a checkpointed database-file copy. All four driver/OS combinations passed.

Modernc is selected because it is a CGO-free direct C-to-Go SQLite port and does not add ncruces' WASM-translated engine/custom VFS layer or its documented higher per-connection memory use. ncruces remains a validated alternative, not the selected production driver. mattn/go-sqlite3 remains excluded from the primary build because its CGO/GCC requirement increases Linux/Windows packaging complexity.

The production adapter and migration runner are not implemented yet. Add the pinned dependency only when that adapter is introduced. The current tests do not simulate abrupt process termination, test the production migration runner implementation, exercise the SQLite backup API, or establish long-duration performance/stress behavior. Those remain acceptance work.

## Required proof matrix

- [x] Linux amd64: open/create, initial migration SQL, foreign keys, WAL, busy timeout, concurrent writers, integrity check, checkpoint, close/reopen and checkpointed file copy.
- [x] Windows amd64: same smoke matrix, including WAL.
- [ ] Abrupt process termination, production migration runner, SQLite backup API and restore, long-duration reader/writer stress and performance.
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

The GitHub Actions matrix passed all four combinations (modernc/ncruces × Ubuntu/Windows), including the actual initial migration SQL. Driver selection is Accepted as modernc v1.62.0; the remaining acceptance items above still gate production release.


## CI bootstrap note

The first evaluation run stopped at module setup because the nested module had no committed `go.sum`. This was corrected in the workflow by running `go mod tidy` and `go mod verify` before the test command. Until a run using that corrected workflow passes, the earlier failure is classified as test-workflow setup failure, not a driver compatibility result.


## CI bootstrap note

The first evaluation run stopped at module setup because the nested module had no committed `go.sum`. This was corrected in the workflow by running `go mod tidy` and `go mod verify` before the test command. The corrected matrix passed on both operating systems.
