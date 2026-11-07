# ADR 0004: SQLite driver evaluation

- Status: Accepted
- Date: 2026-11-07

## Context

SQLite is required for local application metadata, persisted configuration, migrations and bounded operational history. Linux and Windows are first-class targets. A CGO dependency complicates cross-platform builds and packaging.

## Options reviewed

| Driver | License | Build model | Current toolchain implication | Main concern |
|---|---|---|---|---|
| modernc.org/sqlite v1.62.0 | BSD-3-Clause | Pure Go, no CGO | pkg.go.dev metadata requires Go 1.26 | Must validate file locking and workload behavior |
| mattn/go-sqlite3 v1.14.x | MIT | C SQLite through CGO | Requires CGO and a C compiler; Windows requires a GCC toolchain | Cross-build and packaging complexity |
| ncruces/go-sqlite3 v0.37.0 | MIT | SQLite WASM translated to Go | pkg.go.dev metadata requires Go 1.26 | Higher per-connection memory and custom VFS/file-locking/WAL behavior |

## Decision

Use database/sql at the persistence adapter boundary and select **modernc.org/sqlite v1.62.0** as the production SQLite driver.

Both modernc v1.62.0 and ncruces v0.37.0 passed the evaluation tests on Ubuntu 24.04 and Windows. The tests applied the actual migrations/001_initial.sql in a transaction and exercised WAL, foreign keys, rollback, concurrent writes, integrity checks, WAL checkpoint, close/reopen and a checkpointed database-file copy. All four driver/OS combinations passed.

Modernc is selected because it is a CGO-free direct C-to-Go SQLite port and avoids ncruces' WASM-translated engine/custom VFS layer and its documented higher per-connection memory use. ncruces remains a validated alternative, not the selected production driver. mattn/go-sqlite3 remains excluded from the primary build because its CGO/GCC requirement increases Linux/Windows packaging complexity.

## Production implementation

The root Go module pins modernc.org/sqlite v1.62.0 and commits go.mod/go.sum. CI runs go mod tidy, go mod verify and verifies that tidy does not modify the committed lockfiles.

The production adapter lives in internal/infrastructure/sqlite:
- Store.Open creates/opens a file-backed database, rejects in-memory paths, database-file symlinks and Windows UNC paths, and configures WAL, foreign_keys=ON on every connection, busy_timeout=5000 and synchronous=FULL.
- Store.Migrate applies embedded forward migrations in order, skips .down.sql files, rejects version gaps, non-contiguous applied history and untracked history and commits each migration with its schema_migrations record in one transaction.
- The ConfigRepository adapter returns defaults when no settings row exists and validates configuration before an atomic upsert. Persisted data contains secret references, not secret values.
- Tests cover migration idempotence, rollback, version gaps, config round-trip and rejection, file permissions, WAL and connection-scoped PRAGMAs on four simultaneously held connections.

Root CI run 42396865948 passed Go vet, tests and build on Ubuntu 24.04 and Windows. The separate evaluation matrix run 42396720960 passed modernc and ncruces on both operating systems.

## Remaining acceptance gates

- [x] Linux amd64 and Windows amd64 smoke matrix for both CGO-free candidates.
- [x] Production migration runner tests for ordering, idempotence, gap rejection and transactional rollback.
- [x] Config repository tests, file permission check and per-connection PRAGMA verification.
- [ ] Abrupt process termination and clean reopen.
- [ ] SQLite online backup API export/restore. Raw file copy is not the production backup mechanism.
- [ ] Long-duration reader/writer stress and performance baselines.
- [ ] Detect or reject all unsupported network filesystems; UNC rejection alone does not detect NFS or mapped network drives.
- [ ] Verify SQLite and transitive dependency licenses/SBOM.
- [ ] Verify clean uninstall preserves or explicitly offers to export user data.

The migration SQL splitter handles quotes and comments but deliberately does not support trigger bodies containing internal semicolons. Do not add such triggers until the migration executor is replaced with a SQLite-aware parser.

## Sources

- https://pkg.go.dev/modernc.org/sqlite
- https://github.com/modernc-org/sqlite
- https://github.com/mattn/go-sqlite3
- https://pkg.go.dev/github.com/ncruces/go-sqlite3
