# Restore Lifecycle Architecture

## Status

Archive restore is **not exposed** in the GUI or CLI. `backup.Manager.Restore` and `sqlite.RestoreDatabase` remain infrastructure primitives only. The Wails composition root now owns a `sqlite.StoreController`; configuration persistence and GUI/CLI backup export use its shared leases. This removes the former long-lived raw-Store references from those paths, but does not yet make restore safe: the journal storage primitive and pure recovery planner exist, but there is still no filesystem recovery executor, startup recovery coordinator, or application-level restore transaction.

## Required ownership boundary

- The application-lifetime `sqlite.StoreController` now owns the current `*sqlite.Store` and implements the stable configuration repository port.
- Configuration `Load`/`Save` and GUI/CLI backup export now acquire shared lifecycle leases and hold them for the complete SQLite/archive operation.
- Restore acquires an exclusive lifecycle lease, preventing new work and waiting for all in-flight DB operations to finish.
- No application service may retain or cache a raw Store pointer across lifecycle transitions. The controller is the only owner allowed to close or replace it. Current Wails and CLI composition routes follow this boundary.
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

Journal updates must use write-to-temp, fsync, atomic rename and directory sync where supported. Journal paths must be canonicalized and constrained to the application state directory; reject target/staging/rollback path collisions and require rollback metadata to agree with whether a previous database existed; do not trust arbitrary paths from a damaged journal.

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

The stable Store controller and shared-lease wiring are implemented. Its tests cover shared/exclusive exclusion, writer preference, cancellation, shutdown draining, failed replacement, panic fail-closed behavior, concurrent lease stress, and configuration repository round-trip. The Linux `go test -race ./...` step passed in CI run `42427211242`.

The durable journal storage primitive is implemented in `sqlite/restore_journal.go`: strict bounded JSON decoding, explicit phase values, content-based archive identity, state-directory path confinement, symlink rejection, same-directory temporary writes, file sync, atomic rename and directory sync where supported. Journal schema version 2 requires an explicit `previous_database_existed` value; missing metadata is rejected rather than interpreted as `false`. Journal validation also requires a rollback path exactly when a previous database existed, rejects unexpected rollback metadata, and rejects case-aware target/staging/rollback path collisions. Windows path spellings are normalized through resolved absolute paths so 8.3 aliases do not break read/write round trips.

A pure `PlanRestoreRecovery` decision layer now maps validated journal phases plus observed target/staging/rollback presence into conservative actions. It performs no filesystem mutations. It preserves the original target before publication, restores a preserved database after an interrupted publication, preserves a first-ever published candidate for manual review, and requires verification for committed state. Contradictory states return a typed ambiguity error. Tests cover normal phase decisions and fail-closed combinations. Neither the planner nor the journal is connected to `RestoreDatabase` or application startup. The newest rollback-metadata and path-collision hardening changes have not yet passed a fresh CI run.

The remaining blockers are a collision-safe filesystem recovery executor, startup recovery before canonical Store open, a coordinated application-level restore/rollback/reopen service, durable-phase process-death simulation, and deterministic failure-injection tests across every phase. `sqlite.RestoreDatabase` currently performs its own staging and target publication, but it does not use the journal. Restore therefore remains disabled until the full transaction and recovery coordinator are implemented and verified.


## CI evidence — 2026-11-08

The latest journal path-hardening commit `31b9a963c1403bee223f0d9da632f8e570d9ffa6` passed all four required GitHub Actions workflows:

- Root Go CI: Ubuntu and Windows passed; run [42431721560](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42431721560).
- SQLite driver evaluation: modernc and ncruces on Ubuntu and Windows passed (4/4); run [42431721488](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42431721488).
- Wails v3.0.0 platform proof: Ubuntu and Windows passed (2/2); run [42431721597](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42431721597).
- Linux distribution compatibility: all eight distributions passed (8/8), including Ubuntu 22.04 GTK3/WebKit2GTK 4.1, Debian 12 and Fedora GTK3; run [42431721639](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42431721639), attempt 2.

Ubuntu 22.04's first attempt timed out waiting for the frontend-to-Go WebView bridge. The targeted rerun passed, with the bridge smoke test succeeding twice and the application reopen check passing. No code change was made in response to that transient failure.

These results verify the current foundation and journal/planner tests. They do not constitute verification of a filesystem recovery executor or a complete restore transaction. Restore remains disabled.
