# Restore Lifecycle Architecture

## Status

Archive Restore remains **disabled and not exposed** in the GUI, CLI, or startup recovery.

The repository now contains:
- An application-lifetime `sqlite.StoreController` with shared/exclusive leases. Configuration persistence and backup export use shared leases; exclusive operations drain active readers and prevent new ones.
- Strict durable Restore journal v3 records in a retained JSON ledger v1, with candidate/previous SHA-256 and size identities, canonical state-directory confinement, explicit intent/completion phases, one-active-transaction enforcement, terminal-history retention, and a 256-record / 1 MiB fail-closed cap.
- An identity-aware `PlanRestoreRecovery` planner that fails closed on mismatched, unreadable, unverified, or contradictory artifact states.
- Linux and Windows no-follow identity readers and verified no-replace moves. Linux uses pinned-directory `openat2` plus `renameat2(RENAME_NOREPLACE)`; Windows uses retained handles and native `NtSetInformationFile(FileRenameInformationEx)` without a replacement flag. Windows remains qualified only for fixed local NTFS with persistent ACLs.
- A staging-only candidate builder that validates and extracts `.omfaibackup`, builds/migrates a same-filesystem candidate, verifies SQLite integrity and configuration, and returns a measured size/SHA-256 identity without touching the live Store.
- An isolated in-process `RestoreTransactionExecutor` and restart-reconciliation `RestoreRecoveryExecutor`. They retain the journal and StoreController leases, persist intent before mutation, verify moved identities before SQLite open, use only no-replace moves, and preserve ambiguous artifacts.
- An isolated application-level `restore.Service` and infrastructure adapter that connect the real archive candidate builder to the transaction executor. A real `.omfaibackup` integration test verifies the replacement Store, committed journal, and retained rollback database.

The transaction executor has a 19-point process-death matrix. The recovery executor has restart-boundary, failure-injection, source-substitution, destination-collision, and first-database manual-review tests. These tests run in Ubuntu/Windows CI; they do not certify arbitrary filesystem providers or hardware-independent power-loss behavior.

Still missing:
- A startup recovery coordinator that runs before the canonical Store opens. It remains intentionally unwired until host/provider qualification is complete.
- Real-host WebKit sandbox validation. The current CI smoke disables WebKit's sandbox only inside its test subprocess because GitHub-hosted/container runners cannot create the required network namespace.
- Qualification on the actual Linux application-state filesystem and real Windows host. Linux CI covers the runner filesystem and a required tmpfs case; Windows Restore remains limited to fixed local NTFS. ReFS, SMB, removable media, and unknown providers must fail closed.
- Installer/package, upgrade, and clean-uninstall verification. The current build script produces a native executable, not an installer.
- A safe manual export/retention protocol before the bounded ledger reaches 256 records or 1 MiB. Automatic pruning/deletion remains forbidden.
- Hardware/storage-controller power-loss qualification. Process-death tests and write-through requests are not power-loss certification.

The detailed implementation constraints and remaining gates are in [Restore Recovery Executor](RESTORE_RECOVERY_EXECUTOR.md). Never add an overwrite-capable `os.Rename`, `MoveFileExW`, or fallback path to the transaction executor.

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

## Current verification evidence — 2026-11-16

All four workflow families passed on exact branch head `f0090357767bfad20ffa452ba1289e9d47556dfa` (16/16 jobs):
- Root CI [43975095320](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43975095320): Ubuntu and Windows 2/2. Both passed Go vet/tests/build, Python compile/tests, frontend syntax and architecture checks; Ubuntu also passed the race detector.
- SQLite driver evaluation [43975095316](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43975095316): modernc and ncruces on Ubuntu and Windows, 4/4.
- Wails platform proof [43975095477](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43975095477): Ubuntu and Windows 2/2. Ubuntu built and launched the native app, passed CLI dispatch, WebView-to-Go bridge, and database reopen; Windows passed native build and CLI smoke.
- Linux distribution compatibility [43975095377](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43975095377): 8/8. All configured distro lanes built the actual app and passed CLI and two-launch WebView bridge/reopen smoke.

The Linux no-replace test now requires the CI `/dev/shm` fixture to exist and be tmpfs; it may skip only in non-CI local environments. This prevents the tmpfs case from silently disappearing from CI coverage. The CI WebView smoke still disables WebKit sandboxing in its subprocess and is not real-host sandbox evidence. These results do not certify arbitrary filesystems, ReFS/SMB, power-loss behavior, packaging, or clean uninstall.

## Historical verification evidence — 2026-11-08

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


## Production Restore journal selection — 2026-11-14

**Decision: retain the bounded JSON ledger; reject SQLite PERSIST for production.**

The isolated PERSIST symlink capability probe showed a cross-platform safety mismatch:
- modernc / Ubuntu: PERSIST activation failed closed; the symlink and external target were unchanged.
- ncruces / Ubuntu: same fail-closed result.
- ncruces / Windows: PERSIST activation and write succeeded, but the sidecar link entry was replaced; the external target was unchanged.
- modernc / Windows: PERSIST activation and write succeeded while the sidecar remained a symlink; the external target changed. This violates the required no-follow contract for the pinned production driver.

Do not switch the production SQLite driver to ncruces based on this diagnostic probe. Its behavior differs by OS, and the production driver decision remains modernc. The JSON ledger is the selected journal protocol.

### Qualified JSON-ledger boundaries
- Linux cross-process serialization uses `flock`; Windows uses `LockFileEx`. The transaction executor must retain one lease for the entire restore, not lock each journal write independently.
- Linux and Windows ledger readers use no-follow handle-based reads and validate the opened file identity/type before decoding.
- Windows journal lock files must be regular, non-reparse, empty, private and single-link. A hard-link regression test now runs in the cross-platform Go suite.
- Linux publication syncs the pinned parent-directory FD after the namespace update.
- Windows publication and artifact moves use the same source handle opened with native `FILE_WRITE_THROUGH`, flush that handle, then perform the handle-bound rename under the fixed-local-NTFS/persistent-ACL volume policy. See [Restore Recovery Executor](RESTORE_RECOVERY_EXECUTOR.md) for the documented API boundary and its hardware caveat.
- The ledger retains terminal transaction history and is capped at 256 records / 1 MiB. It fails closed at the cap; safe manual export/archive or compaction is still required before enabling Restore for indefinite long-term use.

### Verification
- Root CI on [865c745](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43564870024) passed 2/2; SQLite driver evaluation [43564870094](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43564870094) passed 4/4; Wails [43564870013](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43564870013) passed 2/2; Linux distributions [43564870022](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43564870022) passed 8/8.
- After the WebView smoke-marker race fix, Root CI [43565534118](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43565534118) passed 2/2, Wails [43565533969](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43565533969) passed 2/2, and Linux distributions [43565533971](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43565533971) passed 8/8. The SQLite matrix was not triggered by the Python-only follow-up.

The journal protocol is qualified sufficiently to begin the isolated filesystem transaction executor. This does **not** mean Restore is enabled: executor, rollback/reopen orchestration, startup recovery, and the full failure-injection/process-death/restart matrix are still missing. Restore remains disabled and must not be wired to GUI, CLI or startup recovery.
