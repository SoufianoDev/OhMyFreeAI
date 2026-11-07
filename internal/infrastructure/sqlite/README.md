# SQLite infrastructure adapter

The package provides the production SQLite implementation of the application ConfigRepository port.

## Startup order

1. Resolve a dedicated per-user application data directory.
2. Call Open(ctx, databasePath). Open creates the file if needed, rejects an in-memory database path, a symlink database file and Windows UNC paths, then verifies WAL mode.
3. Call Store.Migrate(ctx, migrations.Files, ".") before constructing application services.
4. Pass the Store as the ConfigRepository implementation.
5. Close the Store after application shutdown.

## Connection policy

Every connection receives foreign_keys=ON, busy_timeout=5000 and synchronous=FULL. The database is placed in WAL mode. Open limits the pool to one connection while establishing WAL, then allows up to four connections.

The caller must use a dedicated local application-data directory. UNC paths are rejected on Windows, but mapped network drives, NFS and other network filesystem types are not comprehensively detected yet. WAL must not be used on an unverified network filesystem.

## Migration contract

Migrations are embedded in the migrations package and are append-only in production. Each forward migration and its schema_migrations row are committed in the same transaction. The runner rejects gaps and untracked migration history and skips .down.sql files.

The current statement splitter handles quoted values/identifiers and SQL comments. It does not support trigger bodies containing internal semicolons. Keep production migrations free of such triggers until a SQLite-aware parser is adopted.

## Deliberate omissions

This adapter does not yet implement the SQLite online backup API, restore verification, abrupt-process crash tests, long-duration stress tests, filesystem-type detection or schema migration checksums. Production backup/restore must not be implemented as an uncoordinated copy of the live database file.
