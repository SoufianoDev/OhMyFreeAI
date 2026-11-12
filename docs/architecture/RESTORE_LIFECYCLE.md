# Restore Lifecycle Architecture

## Status

Archive restore is **disabled and not exposed** in the GUI or CLI.

The repository now contains:
- An application-lifetime `sqlite.StoreController` with shared/exclusive leases. Configuration persistence and GUI/CLI backup export use shared leases; an exclusive operation drains active readers and prevents new ones.
- Strict durable restore journal v3 transaction records in a retained ledger envelope v1. It records candidate and previous-database SHA-256/size identities, a fixed candidate-quarantine path, canonical state-directory-confined paths, and explicit intent/completion phases. The ledger permits one active transaction, retains terminal records, migrates legacy single-record v3 journals on write, and is bounded to 256 records / 1 MiB.
- An identity-aware, pure `PlanRestoreRecovery` planner. It receives verified identities for target, staging, rollback and quarantine. Mismatches, unreadable/unverified files, and contradictory states fail closed with an empty plan.
- A platform-specific no-replace move primitive. Linux pins the directory with `openat2` and uses `renameat2(RENAME_NOREPLACE)`. Windows uses retained directory/source handles and native `NtSetInformationFile(FileRenameInformationEx)` semantics without a replace flag. The Windows operation is provider-dependent and currently qualified only by the repository's CI environment; do not infer support for ReFS, SMB, removable media, or other providers.
- A staging-only candidate builder: `backup.Manager.BuildRestoreCandidate` validates and extracts the archive into an isolated temporary source directory; `sqlite.BuildRestoreCandidate` creates a private same-filesystem candidate, applies migrations, verifies SQLite integrity and configuration, checkpoints/closes it, syncs the file, and returns its measured size/SHA-256 identity. It does not mutate the live Store or execute journaled moves. Failed candidate artifacts are returned and preserved for inspection.

Still missing:
- A filesystem transaction executor that consumes the verified candidate, inspects file identities from safely opened handles, acquires the StoreController exclusive lease, and executes the planner.
- Durable journal phase orchestration around every filesystem mutation.
- A startup recovery coordinator that runs before the canonical Store is opened.
- A shared application-level restore transaction used by both CLI and GUI.
- Process-death, race, failure-injection and filesystem-provider qualification for the complete transaction.

The detailed implementation gate is [Restore Recovery Executor](RESTORE_RECOVERY_EXECUTOR.md). The executor design is authoritative for platform-specific no-replace and race constraints. Do not implement a weaker `os.Rename` fallback.

## Retained journal storage lifecycle

The journal file is a bounded ledger, not a single replaceable transaction record. Terminal records remain available as evidence; starting a later restore appends a new `prepared` record after all previous records are terminal. The writer rejects a second active transaction, duplicate transaction IDs, reused artifact paths, malformed existing ledgers, and history beyond 256 records or 1 MiB. A legacy single-record v3 journal is accepted and migrated on the next successful write.

Journal deletion and automatic pruning are intentionally unsupported. The cap is a fail-closed safety limit; a verified manual export/archive or compaction design is still required before enabling Restore for indefinite long-term use. A temporary file left before atomic publication is not authoritative and is preserved for review. This ledger storage behavior does not itself execute recovery or prove the full process-death matrix.

## Ownership boundary

- The application-lifetime `sqlite.StoreController` owns the current `*sqlite.Store` and is the stable configuration repository.
- Configuration `Load`/`Save` and backup export hold a shared lease for the complete operation.
- Restore must acquire an exclusive lease, stop new database work, and wait for all shared leases to drain.
- No application service may retain a raw Store pointer across a lifecycle transition. Only the controller may publish or close the active Store.
- CLI and Wails GUI must invoke the same application restore service. They must not implement separate restore logic.
- If restore or rollback cannot establish a verified usable Store, the controller must remain unavailable (fail closed).

## Required transaction

1. Inspect and validate the selected `.omfaibackup` through the complete read-only archive inspector. Do not touch the live database.
2. Resolve the canonical private application state directory and validate all transaction paths.
3. Acquire the exclusive Store lease; drain existing operations and reject new ones.
4. Checkpoint WAL, close the active Store, and verify that target `-wal`, `-shm`, and `-journal` sidecars are absent.
5. Append a `prepared` v3 transaction record to the retained ledger with a cryptographically random transaction ID, unique staging/rollback/quarantine paths, archive identity, and explicit previous-database metadata. Preserve all terminal records and allow only one active transaction.
6. Build the candidate in same-filesystem staging. Verify SQLite integrity, foreign keys, migration history, schema compatibility, configuration validity, and candidate content identity before publication.
7. Persist each intent phase before its filesystem mutation. Use the qualified atomic no-replace adapter; verify the moved file identity and required durability barrier before recording the completion phase.
8. Open the published candidate with `sqlite.OpenExisting`, repeat integrity/foreign-key/migration checks, load and validate configuration, and verify required repositories.
9. Persist `committed` only after all checks pass. Publish the new Store through the controller only after the durable commit decision. Retain rollback and quarantine artifacts.
10. Release the exclusive lease and report success only when the new Store is live and verified.

## Journal phases

The v3 state machine records both intent and completed transitions:

```text
prepared -> staging
staging -> preserving-previous -> previous-preserved
previous-preserved -> publishing -> published -> reopening -> committed

published/reopening -> restoring-previous -> rolled-back
```

For a first-ever database, staging may proceed directly to publishing. A failed first-ever publication has no previous database to restore; preserve the candidate and require manual review. Never manufacture a rollback state that did not exist.

A crash between an intent write and its filesystem operation, or between the operation and its completion write, must be reconciled by inspecting content identities. Presence or filename alone never establishes ownership.

## Failure and rollback rules

- Before publication, validation failure or cancellation leaves the original target untouched. Remove only transaction-owned artifacts whose identity has been verified.
- Once publication begins, cancellation must not interrupt the commit-versus-rollback decision.
- After publication but before commit, failure must quarantine the verified candidate and restore the verified previous database using no-replace operations.
- Never overwrite an unexpected destination or invent a new quarantine path after a collision.
- Never delete rollback or quarantine as part of recovery. Cleanup is a separate explicit retention operation after the user verifies the restored application state.
- If rollback, reopen, integrity, foreign-key, migration-history or configuration verification fails, preserve all artifacts and keep the StoreController unavailable.
- Never replay inference, downloads, or other uncertain external work as part of database restore.

## Startup recovery ordering

Recovery must run before opening the canonical Store:

1. Resolve and validate the canonical state directory.
2. Read and strictly validate the bounded ledger v1 and its v3 records. Select the sole active transaction; if all records are terminal, no transaction recovery is pending. Reject v2, unknown fields, duplicate IDs, path reuse and multiple active records.
3. Inspect target, staging, rollback and quarantine without following symlinks/reparse points. Verify SHA-256 and size from the opened file objects.
4. Pass verified identities to `PlanRestoreRecovery`.
5. Execute one idempotent no-replace step at a time, persisting intent/completion phases around each namespace transition.
6. Verify the final database and configuration.
7. Open/publish the Store only after recovery is verified committed or verified rolled back.

Unknown, malformed, contradictory, unreadable or identity-mismatched states must fail closed and preserve every artifact.

## Required verification before enabling Restore

- Shared operations drain before exclusive restore; waiting restore prevents reader starvation.
- Concurrent `Load`, `Save`, and backup export cannot use a closed or stale Store.
- Cancellation before publication leaves the original database intact.
- Inject failures at checkpoint, close, extraction, migration, file sync, journal write, each move, directory sync, reopen, integrity, foreign-key, migration-history, config validation, commit and rollback.
- Simulate process termination before and after every durable journal write and no-replace move.
- Test destination collisions, source replacement races, symlink/reparse-point substitution, non-regular files, digest mismatches and unsupported filesystem operations.
- Verify WAL/SHM/journal handling, file permissions, rollback retention and idempotent repeated recovery.
- Run the transaction matrix on Linux and Windows. Qualify Windows filesystem providers separately; passing NTFS CI does not certify other providers.
- Verify CLI and GUI see the new repository only after commit.
- Keep Restore disabled until every required case passes.

## Current verification evidence — 2026-11-08

The corrected handle-bound no-replace implementation is covered by:
- Root CI run [42496659593](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42496659593): Ubuntu and Windows passed; Ubuntu race detector passed.
- SQLite driver evaluation [42496659645](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42496659645): modernc and ncruces passed on Ubuntu and Windows (4/4).
- Wails platform proof [42496659732](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42496659732): Ubuntu and Windows passed (2/2), including Linux WebView-to-Go and database reopen checks.
- Linux distribution matrix [42496659764](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42496659764): all eight configured distro lanes passed.

These results verify the current foundation, journal/planner tests, and the no-replace move primitive. They do **not** verify the missing filesystem executor, startup recovery coordinator, complete restore transaction, every Windows filesystem provider, or power-loss durability. Restore remains disabled.


## Current Windows source-open revision — 2026-11-08

The Windows source is now opened relative to the retained destination directory handle using NtCreateFile and OBJECT_ATTRIBUTES.RootDirectory. The source-open regression test changes the directory's path after opening the handle and verifies that the operation still targets the original directory.

For code commit c89d532a:
- Root CI 42497620421 passed Ubuntu and Windows; Ubuntu race detector passed.
- SQLite driver evaluation 42497620420 passed 4/4.
- Wails platform proof 42497620429 passed 2/2.
- Linux distribution matrix 42497620422 was still running with 4/8 lanes passed at the time of this update.

The remaining Windows boundary is the initial open of the destination directory by full path: only the final component is checked for reparse-point status. Before executor work, open each directory component relative to a trusted parent handle, reject reparse points at every level, and verify the state directory's private ACL. Restore remains disabled.


## Windows Restore state-directory gate — implementation pending CI

The Windows no-replace move path now performs these checks on the final opened state-directory handle before returning it to the caller:

- The directory owner SID must equal either the current process TokenUser SID or the explicit TokenOwner SID returned by the current process token. This accounts for Windows' default object-owner selection without accepting arbitrary group membership.
- A NULL or missing DACL is rejected.
- Basic allow ACEs are inspected for broad principals (Everyone, Authenticated Users, BUILTIN Users, Guests, Power Users, Pre-Windows 2000 Compatible Access, Interactive, Network and Anonymous). Broad write, add-file, add-subdirectory, delete-child, delete, WRITE_DAC, WRITE_OWNER, generic-write, generic-all, and directory metadata-write rights are rejected.
- ACE layouts that the implementation does not parse (object/callback ACEs) are rejected rather than ignored.
- The opened directory's volume must report a fixed drive, NTFS filesystem, and persistent ACL support. The handle-based remote-protocol query rejects remote-backed directories. UNC paths and mapped remote drive types remain rejected by path/drive validation.

The qualification policy is intentionally NTFS-only. ReFS, SMB, removable media, RAM disks, and other providers are not qualified. This is an implementation policy, not a claim that those filesystems cannot support equivalent semantics.

**Verification is pending.** The code and tests have been pushed to `foundation/bootstrap`, latest code/test commit `8ee245f32fac31ce9dfa1b058163ee90b4f52960`. Root CI [42642834055](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42642834055) is queued on Ubuntu and Windows. Wails [42642834036](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42642834036) is running on both platforms; Linux desktop dependencies passed and Wails CLI installation is in progress. No Windows or Linux test result is claimed for this gate. Restore remains disabled until the gate and the full executor/recovery test matrix pass.


## Windows owner/DACL and volume gate — CI verified — 2026-11-09

Root CI [42643886765](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42643886765) passed on commit `8cb7b7fd3d987fbb2e2fb0484ad67e2e3733baf5` on both Ubuntu and Windows. The Windows Go test suite now verifies the owner check against TokenUser or the token's explicit TokenOwner; it also exercises broad-writable ACE rejection, private directory validation, and fixed-local-NTFS/persistent-ACL qualification. Ubuntu Go tests and race detector also passed. Wails run [42642834036](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42642834036) passed on Ubuntu and Windows.

This is CI-runner qualification only. It does not certify ReFS, SMB, mapped drives, removable media, other filesystem providers, arbitrary Windows installations, or power-loss durability. Restore remains disabled because the filesystem executor, durable mutation orchestration, startup recovery, complete failure-injection matrix, and abrupt-process-restart tests are still missing.


## Latest retained-ledger verification — 2026-11-11

The retained-ledger sequential rollover tests are in commit [c464d7b](https://github.com/SoufianoDev/OhMyFreeAI/commit/c464d7b767db049de15c5724b13bcb3dc5685bad). All workflows for that exact head passed:

- [Root CI](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42975743863): 2/2 Ubuntu/Windows.
- [SQLite driver evaluation](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42975743844): 4/4.
- [Wails platform proof](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42975743834): 2/2.
- [Linux distribution compatibility](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42975743861): 8/8.

This verifies the journal-ledger tests and current build matrix, not the absent transaction executor, power-loss behavior, or full process-death/restart matrix. Restore remains disabled.


## SQLite PERSIST restore-control journal candidate evaluation — 2026-11-12

**Status: evaluation evidence only; not approved for production.** The retained JSON ledger remains the current production journal.

The isolated driver-evaluation module tested modernc v1.62.0 and ncruces v0.37.0 on Ubuntu 24.04 and Windows:

- Both drivers accepted rollback-journal PERSIST mode and synchronous FULL.
- After a committed transaction, the PERSIST sidecar remained as a regular file and its first eight header bytes were zeroed.
- The sidecar survived database close/reopen. Journal mode reverted to DELETE on a fresh connection, so every new connection must explicitly set and verify PERSIST before any journal write. After reapplying PERSIST, both drivers reused the same sidecar file identity for a subsequent transaction.
- A subprocess test forced process termination while a large uncommitted transaction had spilled pages and left a non-zero rollback-journal header. Reopening recovered the committed baseline, removed all uncommitted rows, and passed SQLite integrity_check.
- These tests passed for all four driver/OS combinations in [SQLite evaluation run 43170647220](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43170647220).

A critical fail-closed limitation was also confirmed: normal SQLite open/use silently creates a missing database file and recreates a missing PERSIST sidecar. SQLite does not enforce the required pre-provisioning contract. Any production implementation must inspect both artifacts without following symlinks/reparse points, reject missing/non-regular/unverified artifacts, and open existing files only before recording a restore phase or touching the live Store.

This evaluation does not prove durable initial database/sidecar provisioning, every journal transaction boundary, or hardware-independent power-loss behavior. In particular, Windows still requires a separately verified fixed-local-NTFS write-through creation/metadata barrier for both pre-provisioned artifacts. Do not infer that the existing SQLite VFS file flush provides a parent-directory durability barrier.

The PERSIST candidate must be compared against the retained JSON ledger, including concurrent-process serialization, migration, crash recovery, and operational complexity. Adopt it only if the complete evidence is stronger and existing transaction history can be migrated without loss or guesswork. Otherwise retain JSON and finish qualifying its publication/locking contract.

Restore remains disabled. No executor, startup recovery, GUI/CLI Restore, or application-level restore orchestration is authorized by this evaluation.
