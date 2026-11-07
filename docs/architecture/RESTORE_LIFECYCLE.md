# Restore Lifecycle Architecture

## Status

Archive restore is **not exposed** in the GUI or CLI. `backup.Manager.Restore` and `sqlite.RestoreDatabase` are infrastructure primitives only. The current Wails AppService owns one Store for the full application lifetime, and configuration/backup services retain references to it. Calling the primitive directly from the current service would be unsafe.

## Required ownership boundary

- Introduce one application-lifetime Store controller that owns the current `*sqlite.Store` and implements the stable configuration repository port.
- Normal repository operations and backup export acquire a shared lifecycle lease and hold it until the operation completes.
- Restore acquires an exclusive lifecycle lease, preventing new work and waiting for all in-flight DB operations to finish.
- No application service may retain or cache a raw Store pointer across lifecycle transitions. The controller is the only owner allowed to close or replace it.
- The desktop adapter and CLI must call the same restore application service; they must not implement separate restore logic.

## Required restore transaction

1. Validate the selected `.omfaibackup` with the complete read-only archive inspector. No live database file is touched at this stage.
2. Acquire the exclusive Store lifecycle lease; stop accepting DB operations and wait for existing leases to drain.
3. Checkpoint WAL, close the current Store, and confirm that no target `-wal`, `-shm`, or `-journal` sidecars remain.
4. Write and fsync a durable restore journal in the application state directory. The journal records the target, archive identity, staging path, rollback path, transaction ID and phase; never include secrets.
5. Restore into a same-filesystem staging database. Verify SQLite integrity, foreign keys, contiguous migration history and schema compatibility before publishing.
6. Atomically preserve the previous database as a rollback file and publish the staged database at the canonical target path. Sync the containing directory where supported.
7. Open the new target through `sqlite.OpenExisting`; run integrity and foreign-key checks, confirm migration history, load and validate application configuration, and verify all required repositories can operate.
8. Only after verification succeeds, atomically mark the journal committed and publish the new Store through the controller. Keep the rollback file until explicit cleanup policy permits removal.
9. Release the exclusive lease and report success only after the new Store is live and all post-restore checks pass.

## Failure and rollback rules

- Before publication, cancellation or validation failure leaves the original database untouched; clean only transaction-owned staging files.
- After publication but before commit, any open, migration, integrity or configuration failure must close the candidate Store, atomically restore the rollback database, reopen it and verify it before releasing the lifecycle lease.
- If rollback or reopening the previous database fails, fail closed: do not accept normal DB operations, preserve every recovery file, and report exact paths and phases for manual recovery.
- Never delete or overwrite the rollback file as part of an unverified retry.
- Once publication begins, cancellation must not interrupt the rollback/commit decision. Complete the transaction to a verified committed or verified rolled-back state.
- Never replay inference, downloads or other uncertain external work as part of database restore.

## Crash recovery

The existing `sqlite.RestoreDatabase` atomically stages a snapshot and preserves the previous database, but it does not persist a transaction journal. That primitive alone is not sufficient for crash-safe application restore.

On startup, before opening the canonical Store, a recovery coordinator must inspect any durable restore journal:

- `prepared` or `staging`: retain the original target; remove only validated transaction-owned staging artifacts.
- `previous-preserved` with no published target: put the rollback file back at the canonical path and verify it.
- `published` or `reopening` without a committed marker: restore the previous database and verify it before normal startup.
- `committed`: open and verify the new database; keep the rollback file until the configured retention/cleanup policy runs.
- Unknown, malformed or contradictory journal/filesystem states: fail closed and preserve all files.

Journal updates must use write-to-temp, fsync, atomic rename and directory sync where supported. Journal paths must be canonicalized and constrained to the application state directory; do not trust arbitrary paths from a damaged journal.

## Required tests before enabling restore

- Shared operations block correctly while exclusive restore waits; no operation can use a closed or stale Store.
- Concurrent `Load`, `Save` and backup export are drained before close; operations arriving during restore receive a typed unavailable/busy state.
- Cancellation before publication preserves the original DB and cleans only owned staging files.
- Inject failures at checkpoint, close, extraction, migration, publication, reopen, integrity check, config validation, journal commit and rollback.
- Simulate process termination after every durable journal phase and verify startup recovery chooses the correct database.
- Verify rollback DB contents, WAL sidecars, file permissions, fsync/atomic rename behavior and no-overwrite guarantees on Linux and Windows.
- Verify the Wails AppService and CLI both observe the new repository only after commit; no cached pointer may refer to the closed Store.
- Retain rollback until the user has had a clear opportunity to confirm the restored application state; never silently delete it.

## Current implementation gap

`AppService` currently retains a direct `*sqlite.Store`; `appconfig.Service.Repository` also points directly to that Store. There is no stable Store controller, exclusive lifecycle lease, durable restore journal, startup journal recovery or coordinated rollback/reopen service. Restore therefore remains disabled.
