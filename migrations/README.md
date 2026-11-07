# SQLite migrations

- `001_initial.sql` is the initial forward migration.
- `001_initial.down.sql` is for disposable development databases only, before production adoption.
- The migration runner must apply each migration in a transaction and insert its version into `schema_migrations` only after successful execution.
- Applied production migrations are immutable. Production evolution is forward-only: add a new migration or restore a verified backup; do not automatically run down migrations.
- Enable and verify `PRAGMA foreign_keys=ON` for every database connection. This is connection-scoped.
- Use WAL only on supported local filesystems. Do not put the live SQLite database on Google Drive or an unverified network filesystem.
- Store UTC timestamps. Sanitize source URIs before persistence; never store bearer tokens, signed query parameters, prompts or full inference payloads.
- Persist OS secret references only. The secret values themselves belong in the platform secret store.
- Selected driver: `modernc.org/sqlite v1.62.0` (ADR 0004). Import it only inside the concrete SQLite persistence adapter; keep application/domain code on `database/sql`-backed ports.
