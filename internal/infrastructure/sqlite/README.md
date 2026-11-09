# SQLite infrastructure adapter

The package provides the production SQLite implementation of the application ConfigRepository port.

## Startup order

1. Resolve a dedicated per-user application data directory.
2. Call Open(ctx, databasePath) for normal first-run startup. It creates the file if needed, rejects an in-memory database path, a symlink database file and Windows UNC paths, then verifies WAL mode. Explicit backup export uses OpenExisting(ctx, databasePath), which requires a regular existing file and does not create the file/directory or change their permissions.
3. Call Store.Migrate(ctx, migrations.Files, ".") before constructing application services.
4. Wrap the migrated Store in one application-lifetime `StoreController` and pass the controller as the ConfigRepository implementation. Do not retain or distribute raw Store pointers across lifecycle transitions.
5. Close the controller after application shutdown; it acquires an exclusive lease, drains shared operations, closes the active Store and rejects later operations.

## Connection policy

Every connection receives foreign_keys=ON, busy_timeout=5000 and synchronous=FULL. The database is placed in WAL mode. Both open paths limit the pool to one connection while establishing WAL, then allow up to four connections. Open applies private directory/file permissions; OpenExisting preserves existing filesystem permissions.

The caller must use a dedicated local application-data directory. UNC paths are rejected on Windows, but mapped network drives, NFS and other network filesystem types are not comprehensively detected yet. WAL must not be used on an unverified network filesystem.

## Migration contract

Migrations are embedded in the migrations package and are append-only in production. Each forward migration and its schema_migrations row are committed in the same transaction. The runner rejects gaps, non-contiguous applied history and untracked migration history and skips .down.sql files.

The current statement splitter handles quoted values/identifiers and SQL comments. It does not support trigger bodies containing internal semicolons. Keep production migrations free of such triggers until a SQLite-aware parser is adopted.

## Backup and restore

- Store.BackupTo creates a consistent snapshot through modernc's SQLite online backup API. It does not copy the live database file or WAL sidecar.
- The snapshot is verified with PRAGMA integrity_check, foreign_key_check and contiguous schema_migrations history before atomic publication. Existing destinations are not overwritten; the snapshot file is mode 0600 on Unix.
- RestoreDatabase verifies the source before touching the target, restores into a temporary database, applies forward migrations, verifies the result, checkpoints/closes it, and atomically replaces the target. The previous database is retained as a rollback file.
- The caller must close all target database connections before restore. Active SQLite sidecars are treated as a busy target. The archive manager now validates and extracts `.omfaibackup` archives before calling the SQLite restore primitive.
- The application does not expose restore in the GUI or CLI. Wails AppService and configuration use the application-lifetime StoreController; GUI and CLI backup export hold a shared lease through complete archive publication. The repository has a strict v3 journal, identity-aware recovery planner, Linux/Windows no-replace move primitives, and a Windows owner/DACL plus fixed-local-NTFS qualification gate. Root CI 42643886765 passed on Ubuntu and Windows for the TokenUser/TokenOwner correction. The collision-safe filesystem executor, durable mutation orchestration, startup recovery, process-death/failure-injection matrix and shared app-level Restore transaction are still missing. Restore remains disabled. See `docs/architecture/RESTORE_LIFECYCLE.md` and `docs/architecture/RESTORE_RECOVERY_EXECUTOR.md`.

## Deliberate omissions

Abrupt-process crash tests, long-duration stress tests, filesystem-type detection, schema migration checksums, application-level restore lifecycle/crash recovery and clean-uninstall data handling remain open. Production backup/restore must not be implemented as an uncoordinated copy of the live database file.
