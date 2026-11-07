# SQLite migrations

- 001_initial.sql is the initial forward migration.
- 001_initial.down.sql is for disposable development databases only, before production adoption.
- The production runner is internal/infrastructure/sqlite.Store.Migrate and reads the embedded filesystem from the migrations package.
- Each forward migration runs in its own transaction. The schema_migrations record is inserted before that transaction commits; a failed migration rolls back both DDL and its version record.
- The runner orders numbered migrations, rejects version gaps and untracked migration history, skips .down.sql files and never automatically runs a down migration.
- Applied production migrations are immutable. Production evolution is forward-only: add a new migration or restore a verified backup.
- The current SQL splitter handles quoted strings/identifiers and line/block comments. Trigger bodies with internal semicolons are not supported; do not add such triggers until a SQLite-aware parser replaces the splitter.
- Enable and verify PRAGMA foreign_keys=ON for every database connection. This is connection-scoped.
- Store UTC timestamps. Sanitize source URIs before persistence; never store bearer tokens, signed query parameters, prompts or full inference payloads.
- Persist OS secret references only. The secret values themselves belong in the platform secret store.
- Store.Open configures WAL, foreign_keys, a 5-second busy timeout and synchronous=FULL. It rejects in-memory paths, database-file symlinks and Windows UNC paths. It does not yet identify every network filesystem type.
- Selected production driver: modernc.org/sqlite v1.62.0 (ADR 0004). Import it only inside the concrete SQLite persistence adapter; keep application/domain code on database/sql-backed ports.
- Production tests cover idempotence, migration gap rejection, rollback, configuration round-trip, invalid-config rejection, file permissions, WAL and per-connection PRAGMAs on four simultaneously held connections.
- Abrupt process termination, SQLite backup API export/restore, long-duration stress/performance and clean-uninstall data handling remain open.
