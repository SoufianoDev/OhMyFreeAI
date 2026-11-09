# Restore Lifecycle Architecture

## Status

Archive restore is **disabled and not exposed** in the GUI or CLI.

The repository now contains:
- An application-lifetime `sqlite.StoreController` with shared/exclusive leases. Configuration persistence and GUI/CLI backup export use shared leases; an exclusive operation drains active readers and prevents new ones.
- A strict durable restore journal, schema version 3. It records candidate and previous-database SHA-256/size identities, a fixed candidate-quarantine path, canonical state-directory-confined paths, and explicit intent/completion phases.
- An identity-aware, pure `PlanRestoreRecovery` planner. It receives verified identities for target, staging, rollback and quarantine. Mismatches, unreadable/unverified files, and contradictory states fail closed with an empty plan.
- A platform-specific no-replace move primitive. Linux pins the directory with `openat2` and uses `renameat2(RENAME_NOREPLACE)`. Windows uses retained directory/source handles and native `NtSetInformationFile(FileRenameInformationEx)` semantics without a replace flag. The Windows operation is provider-dependent and currently qualified only by the repository's CI environment; do not infer support for ReFS, SMB, removable media, or other providers.

Still missing:
- A filesystem executor that inspects file identities from safely opened handles and executes the planner.
- Durable journal phase orchestration around every filesystem mutation.
- A startup recovery coordinator that runs before the canonical Store is opened.
- A shared application-level restore transaction used by both CLI and GUI.
- Process-death, race, failure-injection and filesystem-provider qualification for the complete transaction.

The detailed implementation gate is [Restore Recovery Executor](RESTORE_RECOVERY_EXECUTOR.md). The executor design is authoritative for platform-specific no-replace and race constraints. Do not implement a weaker `os.Rename` fallback.

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
5. Create a v3 journal with a cryptographically random transaction ID, fixed staging/rollback/quarantine paths, archive identity, and explicit previous-database metadata.
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
2. Read and strictly validate the bounded v3 journal. Reject v2 and unknown fields.
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

- The directory owner SID must equal the current process token user SID.
- A NULL or missing DACL is rejected.
- Basic allow ACEs are inspected for broad principals (Everyone, Authenticated Users, BUILTIN Users, Guests, Power Users, Pre-Windows 2000 Compatible Access, Interactive, Network and Anonymous). Broad write, add-file, add-subdirectory, delete-child, delete, WRITE_DAC, WRITE_OWNER, generic-write, generic-all, and directory metadata-write rights are rejected.
- ACE layouts that the implementation does not parse (object/callback ACEs) are rejected rather than ignored.
- The opened directory's volume must report a fixed drive, NTFS filesystem, and persistent ACL support. The handle-based remote-protocol query rejects remote-backed directories. UNC paths and mapped remote drive types remain rejected by path/drive validation.

The qualification policy is intentionally NTFS-only. ReFS, SMB, removable media, RAM disks, and other providers are not qualified. This is an implementation policy, not a claim that those filesystems cannot support equivalent semantics.

**Verification is pending.** The code and tests have been pushed to `foundation/bootstrap`, latest code/test commit `8ee245f32fac31ce9dfa1b058163ee90b4f52960`. The GitHub connector returned no combined status checks and does not expose push-run listing without a workflow run ID. No Windows or Linux CI result is claimed for this gate. Restore remains disabled until the gate and the full executor/recovery test matrix pass.
