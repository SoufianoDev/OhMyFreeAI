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

Migrations are embedded in the migrations package and are append-only in production. Each forward migration and its schema_migrations row are committed in the same transaction. The runner rejects gaps, non-contiguous applied history and untracked migration history and skips .down.sql files.

The current statement splitter handles quoted values/identifiers and SQL comments. It does not support trigger bodies containing internal semicolons. Keep production migrations free of such triggers until a SQLite-aware parser is adopted.

## Backup and restore

- Store.BackupTo creates a consistent snapshot through modernc's SQLite online backup API. It does not copy the live database file or WAL sidecar.
- The snapshot is verified with PRAGMA integrity_check, foreign_key_check and contiguous schema_migrations history before atomic publication. Existing destinations are not overwritten; the snapshot file is mode 0600 on Unix.
- RestoreDatabase verifies the source before touching the target, restores into a temporary database, applies forward migrations, verifies the result, checkpoints/closes it, and atomically replaces the target. The previous database is retained as a rollback file.
- The caller must close all target database connections before restore. Active SQLite sidecars are treated as a busy target. The API accepts a raw SQLite snapshot; the complete .omfaibackup archive/manifest wrapper is not implemented yet.

## Deliberate omissions

Abrupt-process crash tests, long-duration stress tests, filesystem-type detection, schema migration checksums, full .omfaibackup archive validation and clean-uninstall data handling remain open. Production backup/restore must not be implemented as an uncoordinated copy of the live database file.
