# Restore Recovery Executor — Design Constraints

**Status:** design gate; no filesystem executor is implemented yet.  
**Scope:** Linux and Windows desktop installations, with SQLite state files confined to the canonical application state directory.

## Why the executor is a separate safety boundary

The current journal and PlanRestoreRecovery are metadata persistence and a pure decision layer. They do not perform filesystem operations. os.Rename alone is not an acceptable publication primitive for recovery because replacement behavior differs by platform and can overwrite an unexpected destination. A path-level Lstat followed by a later rename also leaves a time-of-check/time-of-use window.

The executor must never infer ownership merely because a path appears in a journal. It must distinguish transaction artifacts from unexpected files and must preserve ambiguous state rather than deleting, replacing, or silently moving it.

## Required journal/state-machine changes

The version-3 state machine already includes durable intent phases for preserving the previous database, publishing the candidate, and restoring the previous database. The pure planner has explicit truth tables for these phases. The future executor must still persist each intent before its mutation, persist completion only after the no-replace move and durability barrier succeed, and exercise both sides of every crash window.

At minimum, model these transitions explicitly:

1. prepared — archive and transaction metadata validated; no target mutation started.
2. staging — candidate database is being built; target remains untouched.
3. preserving-previous — durable intent to move the old target to rollback; either side of the no-replace move may be observed after a crash.
4. previous-preserved — rollback is verified and canonical target is absent.
5. publishing — durable intent to publish the staged candidate; either staging or target may contain the candidate after a crash.
6. published — candidate is at target; rollback remains retained when there was a previous DB.
7. reopening — candidate is being opened and application-level validation is in progress.
8. restoring-previous — durable intent to put rollback back at target; either side of that move may be observed after a crash.
9. committed — candidate is verified and the new Store is live; rollback remains retained.
10. rolled-back — previous database is reopened and verified; any candidate/staging artifact is retained or safely quarantined.

Every intent phase must have an exhaustive truth table for target, staging, rollback and candidate-quarantine presence. Unknown combinations return an ambiguity error and preserve every artifact.

## No-overwrite filesystem contract

- Never use replacement-style rename for target, rollback or quarantine publication.
- Use a platform adapter with atomic no-replace semantics. Linux uses `openat2` to pin the containing directory and `renameat2(RENAME_NOREPLACE)` with validated leaf names. Windows opens every directory component relative to a retained parent handle, opens the source leaf relative to the pinned state-directory handle, then uses native `NtSetInformationFile(FileRenameInformationEx=65)` with the replacement flag omitted. If the OS/filesystem rejects the required operation, fail closed; never fall back to `os.Rename`, `MoveFileExW`, or another overwrite-capable/path-only rename.
- Source and destination must be on the same filesystem. Journal canonical state-directory confinement is a prerequisite, not a substitute for runtime validation.
- Revalidate source and destination immediately before each mutation using no-follow filesystem operations. Reject symlinks, directories, devices, sockets, and other non-regular entries.
- Destination collision is never resolved by deleting or replacing the destination. It is an ambiguity that must be preserved for manual review.
- Sync the moved file where needed and sync the containing directory after each durable namespace transition. Windows limitations must be documented and tested separately.
- A move that succeeds but whose directory sync fails is not reported as success. The journal remains in the corresponding intent phase so restart can inspect and reconcile the actual files.

## Artifact ownership and verification

- Transaction paths must be generated from a cryptographically random transaction identifier, confined to the canonical state directory, and proven absent before the transaction starts.
- The journal must record enough content identity to verify the staged candidate and previous database before moving or restoring them. Archive SHA-256 alone does not identify the migrated SQLite candidate.
- A partial staging file created before its final digest is durable is not automatically deleted unless the executor can prove it was created by this transaction. Otherwise preserve it and require manual review.
- Candidate preservation requires a dedicated, journaled quarantine path. The executor must never invent a replacement name after a collision and then forget that name; the chosen path must be durable before moving the candidate.
- Do not delete rollback as part of recovery. Cleanup is a separate, explicit retention operation after the restored application has been verified.

## Startup ordering

1. Resolve and validate the canonical state directory.
2. Read and strictly validate the journal.
3. Inspect all transaction paths without following symlinks.
4. Build a complete recovery plan from the durable phase and observed filesystem state.
5. Execute one idempotent no-overwrite step at a time, persisting intent/completion phases around each transition.
6. Open the canonical Store only after the recovery result is either verified committed or verified rolled back.
7. If any check, move, sync, open, integrity check, foreign-key check, migration-history check, or config validation fails, leave the controller unavailable and preserve all artifacts.

## Required test matrix before enabling Restore

- Crash before and after every intent journal write.
- Crash before and after every no-replace move.
- Crash after a move but before directory sync; restart must reconcile without overwriting.
- Existing destination collision for target, rollback, staging and quarantine.
- Symlink swap and non-regular file substitution between inspection and mutation.
- Same-file and different-file cases when both source and destination names exist after a crash.
- Missing, truncated, malformed and contradictory journal states.
- Linux filesystems that reject RENAME_NOREPLACE; verify fail-closed behavior.
- Windows destination-exists behavior and write-through semantics.
- Candidate and previous database digest mismatch.
- Rollback reopen, integrity, foreign-key, migration-history and configuration validation failures.
- Repeated recovery execution must be idempotent and must not delete rollback.
- Run the matrix on Linux and Windows; keep Restore disabled until all cases pass.

## Current implementation status

Implemented as low-level primitives only:

- Strict durable journal v3 transaction records in a bounded retained ledger v1, with intent phases, canonical path confinement, candidate-quarantine metadata, and paired candidate/previous content identities. The ledger allows one active transaction, retains terminal history, migrates legacy single-record v3 files on the next successful write, and is capped at 256 records / 1 MiB.
- Identity-aware pure recovery planner for completed phases and all three intent phases: preserving-previous, publishing, and restoring-previous. Contradictory states return an empty plan and an ambiguity error.
- Linux no-replace move: `openat2` pins the private containing directory, `renameat2(RENAME_NOREPLACE)` performs the move, and the same pinned parent directory FD is synced with `fsync` before success is reported.
- Windows no-replace move: the state-directory path is opened component-by-component from the volume root; the source leaf is opened relative to the retained directory handle with native `FILE_WRITE_THROUGH`; `FlushFileBuffers` is called before the same source handle is renamed using `NtSetInformationFile(FileRenameInformationEx=65)` with the replace flag omitted.
- Windows owner/DACL and volume qualification: fixed local NTFS with persistent ACLs only. ReFS, SMB, removable and unknown volumes fail closed.
- Linux and Windows no-follow content identity readers, plus `moveRestoreFileNoReplaceVerified`, which checks the moved destination's size and SHA-256 through a no-follow handle and preserves it on mismatch.
- Platform-specific journal publication: Linux syncs the containing directory after same-directory ledger publication; Windows uses the handle-bound native write-through replacement rename path for the ledger. Journal deletion is intentionally unsupported on both platforms; terminal records are retained instead.
- StoreController shared/exclusive leases and lower-level SQLite backup/staging primitives.

Not implemented:

- Filesystem transaction executor, startup recovery coordinator, process-death/failure-injection executor tests, and application-level Restore orchestration.
- Integration of the post-move identity verifier into every transaction move. The future executor must call it immediately after each successful no-replace move and before any SQLite database open.
- Automatic reconciliation of journal intent phases after process death.
- Hardware-independent power-loss qualification.

Threat and durability boundaries:

- Linux still selects the source leaf by name at the final `renameat2` call. The private state directory blocks untrusted cross-user namespace writers, while post-move no-follow size/SHA-256 verification detects a substituted artifact before SQLite is opened. This is not a claim that every same-user concurrent mutation race is eliminated.
- Windows is qualified only for fixed local NTFS with persistent ACLs. The documented write-through metadata-flush request is used on the exact source handle processed by the handle-bound rename. This is not a hardware-independent power-loss guarantee; storage devices/controllers may not honor flush requests correctly.
- Automatic journal deletion is unsupported on both platforms. Do not delete, prune or replace retained terminal records in a way that loses rollback/quarantine evidence.
- The exported legacy `RestoreDatabase` helper remains a separate path-based implementation. Do not call it from the journal executor or expose it through GUI/CLI.

Restore remains disabled. Do not add startup recovery, GUI/CLI Restore, or application-level Restore service until the executor's identity, no-replace, journal, rollback, process-death and restart tests pass on Linux and Windows.

## Retained journal ledger lifecycle — implemented 2026-11-10

`restore-journal.json` now contains a bounded ledger envelope (ledger version 1) of version-3 transaction records:

- The serialized ledger is limited to 1 MiB and 256 transaction records. The writer fails closed when either limit is reached; it does not prune old records.
- Exactly one non-terminal transaction is permitted. A new transaction must start at `prepared`; a second active transaction is rejected.
- Transaction IDs are unique. All records must target the same canonical database path, and staging/rollback/quarantine artifact paths cannot be reused across retained transactions.
- `committed` and `rolled-back` records are retained. A later transaction is appended after all earlier records are terminal. `ReadRestoreJournal` returns the sole active record, or the latest terminal record when idle.
- The previous single-record v3 format remains readable and is migrated to the ledger envelope on its next successful write.
- `RemoveRestoreJournal` validates terminal state but deliberately does not delete the ledger; it returns the retained-history error. There is no automatic journal pruning or deletion on either platform.
- A temporary ledger file left by a crash before publication is not treated as committed state and is not deleted by the reader. The last successfully published ledger remains authoritative.

Regression tests now cover two sequential terminal transactions, active-record selection, retention after a removal attempt, legacy v3 migration, path reuse rejection, and an orphan temporary file. These tests do not simulate every process-death point or prove the full transaction recovery protocol.

The 256-record / 1 MiB cap is an explicit fail-closed limit, not a completed retention policy. A safe manual export/archive or future ledger compaction protocol must be designed and verified before enabling Restore for long-term use. Never bypass the cap by deleting evidence.

## Version-3 journal contract (implementation gate)

The next journal revision is intentionally incompatible with v2. Restore is not enabled and there is no supported live transaction to migrate, so the reader must reject v2 rather than guess missing ownership or content metadata.

The v3 record must contain these fields in addition to the existing transaction metadata:

| Field | Required invariant |
|---|---|
| `candidate_quarantine_path` | Always present, confined to the canonical state directory, distinct from target, staging and rollback. The path is selected once and persisted before any candidate-preservation move. |
| `candidate_sha256`, `candidate_size_bytes` | Both absent/zero only while the candidate is not yet complete. Both required before the first publication intent and retained through published, reopening, restoring-previous, committed and rolled-back. |
| `previous_sha256`, `previous_size_bytes` | Required if and only if `previous_database_existed=true). Capture and verify before moving the previous database. Must be absent/zero for a first-ever database. |

The implementation should represent optional content identities explicitly (digest plus size) and validate them as a pair. SHA-256 must be exactly 32 decoded bytes in lowercase hexadecimal; size must be non-zero. A digest without a size, size without a digest, or identity for a non-existent previous database is invalid.

### v3 phase transitions

```text
prepared
  -> staging
  -> preserving-previous (only when a previous DB exists)
  -> previous-preserved
  -> publishing
  -> published
  -> reopening
  -> committed

published/reopening
  -> restoring-previous
  -> rolled-back
```

For a first-ever database, `staging -> publishing` is allowed and rollback phases are not. A failed first-ever publication cannot be called rolled-back because no previous database exists; preserve the candidate and require manual review. The executor must not manufacture a prior state.

Intent phases are durable *before* their corresponding filesystem mutation. Completion phases are durable only after the move and required directory sync have succeeded. If sync fails after a move, leave the intent phase intact; recovery inspects the actual file identities and reconciles without replacement.

### Validation and planner implications

- All four transaction paths (target, staging, rollback when present, quarantine) must be pairwise distinct using platform-aware path comparison.
- The candidate identity must match the bytes at whichever candidate-bearing path is observed; previous identity must match rollback before restoration and target after rollback.
- File presence alone is insufficient. An unexpected file with the right name but wrong digest is ambiguous and must be preserved.
- A quarantine destination collision is never resolved by generating a new path during recovery. It is a manual-review condition because the chosen path is already part of the durable transaction record.
- `rolled-back` requires the verified previous database at target, no candidate at target, and retained rollback/quarantine artifacts consistent with the last durable intent. It does not authorize deleting rollback.
- The pure planner must model candidate/quarantine presence and identity verification results explicitly; it must never emit a destructive action from booleans that omit content identity.
- Tests must reject v2, missing paired identity fields, malformed digests, zero sizes, inconsistent previous-database identity, and every path collision before any executor operation can run.


## Identity-aware planner implementation status

The pure planner now receives explicit per-path identity outcomes for target, staging, rollback and quarantine: missing, candidate match, previous-database match, mismatch, unreadable, or unverified. Mismatch, unreadable and unverified states return ambiguity with an empty plan. A partial staging artifact is never selected for cleanup unless its candidate identity matches the journal.

Completed-phase decisions require the expected content identity at each relevant path. Candidate publication rollback requires a verified previous database at rollback and a verified candidate either at target or already in quarantine. Committed verification requires the candidate at target and retains the previous rollback. Rolled-back verification requires the previous database at target and the candidate retained in quarantine.

The planner now accepts the staged-candidate state under `restoring-previous`, so an executor can persist rollback intent before moving staging to quarantine and safely resume after a crash. All other staging combinations remain ambiguous. This is a planner decision only, not automatic crash recovery.


## Intent-phase planner truth tables and no-replace move primitive — 2026-11-08

The pure planner now has explicit decisions for all three durable intent phases. These are planner decisions only; the filesystem executor is still absent.

| Intent phase | Verified filesystem state | Plan |
|---|---|---|
| preserving-previous | Previous DB at target, rollback missing, quarantine missing; staging missing or verified candidate | Resume with PreservePrevious=true |
| preserving-previous | Target missing, previous DB at rollback, quarantine missing; staging missing or verified candidate | Resume without repeating the move |
| publishing | Candidate at staging, target missing, rollback matches prior-DB expectation, quarantine missing | PublishCandidate=true |
| publishing | Candidate already at target, staging missing, rollback matches expectation, quarantine missing | Resume without repeating the move |
| restoring-previous | Candidate at target, previous DB at rollback, quarantine missing, staging missing | Preserve target candidate, then restore rollback |
| restoring-previous | Candidate at staging, target missing, previous DB at rollback, quarantine missing | Preserve staged candidate in quarantine, then restore rollback |
| restoring-previous | Candidate at quarantine, previous DB at rollback, target and staging missing | Restore rollback |
| restoring-previous | Candidate at quarantine, previous DB already at target, rollback and staging missing | Resume without repeating the move |

Any other state—including identity mismatch, unreadable/unverified files, unexpected quarantine presence, contradictory prior-DB metadata, or ambiguous multiple copies—returns ErrRestoreRecoveryAmbiguous with an empty plan. The planner never authorizes cleanup from file presence alone.

### Platform no-replace move primitive

Implemented and verified as isolated primitives; neither is yet called by a transaction executor:

- Linux: `openat2` pins the common parent directory and rejects symlink/magic-link traversal; `renameat2(RENAME_NOREPLACE)` performs the atomic no-replace move; `fsync` is called on the already-pinned parent directory FD before the operation reports success. Unsupported kernel/filesystem errors fail closed.
- Windows: opens the canonical state directory component-by-component from a trusted volume-root handle; opens the source leaf relative to that retained directory handle using `NtCreateFile` with native `FILE_WRITE_THROUGH` (CreateOptions 0x2); flushes the source handle; then uses `NtSetInformationFile(FileRenameInformationEx=65)` on that same handle with the replacement flag unset. Existing destinations fail; unsupported providers fail closed.
- Shared preflight rejects empty/equal paths, cross-directory moves, invalid leaves and non-regular sources. The OS operation remains authoritative for destination collision.
- `moveRestoreFileNoReplaceVerified` wraps the move and verifies the moved destination's expected size and SHA-256 through a no-follow handle. On mismatch or read failure, the destination is preserved and the call fails.
- The Linux source leaf remains name-selected at mutation time; the post-move identity check is mandatory and must occur before SQLite opens the destination. No overwrite-capable fallback is allowed.
- Journal publication is separate from artifact movement: journal updates intentionally replace the prior journal atomically, while candidate/previous database artifacts must always use the no-replace move adapter.

### Research sources

- Linux renameat2(RENAME_NOREPLACE): https://www.man7.org/linux/man-pages/man2/rename.2.html
- Microsoft MoveFileExW flags and write-through contract: https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw
- Go x/sys/unix: https://pkg.go.dev/golang.org/x/sys@v0.49.0/unix#Renameat2
- Go x/sys/windows: https://pkg.go.dev/golang.org/x/sys@v0.49.0/windows#MoveFileEx

### Verification boundary

Test-first commit 61a19ea5932bbda9ef722e8789c5af42cabe4c45 produced expected RED in Root Go CI run 42438137306: go vet failed because moveRestoreFileNoReplace did not yet exist. The implementation and platform-specific tests were then added. On code commit d717eac7096d98c3b59b1b9ebd39aa1cff6e2d93, Root Go CI run 42438368020 passed on Ubuntu and Windows, including Linux race tests; SQLite driver matrix run 42438368022 passed 4/4. Wails run 42438368038 and Linux distribution run 42438368003 were still running at the time of this update. A separate retry of the failed Ubuntu 22.04 GTK3 WebView smoke from run 42437541325 was also in progress. Do not infer those pending results.


## Journal namespace durability by platform — verified implementation

The journal publisher uses platform-specific namespace persistence rather than pretending that directory `fsync` is portable:

- Linux: write the temporary journal, sync the file, close it, atomically rename it in the same directory, then sync the containing directory. A directory-sync error is returned; callers must leave the durable intent/artifacts for inspection and must not continue to the next destructive move.
- Windows: write and sync the temporary journal, open it relative to the retained state-directory handle with native `FILE_WRITE_THROUGH`, call `FlushFileBuffers`, then publish it with the handle-bound `NtSetInformationFile(FileRenameInformationEx=65)` path. Replacement is permitted only for this journal update; it is not the artifact move primitive. The source handle is the same write-through object processed by the rename.
- Windows journal deletion still returns `errRestoreJournalRemovalUnsupported`. Leave the journal intact; do not report a deletion as durable.

### Windows durability boundary

Microsoft's `NtCreateFile` documentation defines `FILE_WRITE_THROUGH` as CreateOptions 0x00000002. The Win32 `CreateFile` documentation states that a write-through request also causes NTFS to flush metadata changes—including a rename operation—that result from processing the request. The Windows adapter applies the native write-through option to the source handle and performs the handle-bound rename through that same handle. This establishes the documented NTFS write-through request for the rename metadata; it does not certify every filesystem provider or storage stack.

### Research sources

- Microsoft `NtCreateFile`: https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntcreatefile
- Microsoft `CreateFileA` and `FILE_FLAG_WRITE_THROUGH`: https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-createfilea
- Microsoft file caching: https://learn.microsoft.com/en-us/windows/win32/fileio/file-caching
- Linux `fsync(2)`: https://man7.org/linux/man-pages/man2/fsync.2.html
- Linux `renameat2(2)`: https://man7.org/linux/man-pages/man2/renameat2.2.html

### Verification boundary

The latest code revision and all four CI workflow families passed on commit `9ca1f7e6de8ecca00970b3201601ca70f31ddb95`:

- Root CI [42772101619](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42772101619): Ubuntu and Windows passed.
- SQLite driver evaluation [42772101510](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42772101510): 4/4 Ubuntu/Windows × modernc/ncruces jobs passed.
- Wails platform proof [42772101659](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42772101659): 2/2 jobs passed, including Linux CLI and WebView/database-reopen smoke plus Windows native CLI/executable smoke.
- Linux distribution compatibility [42772101557](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42772101557): 8/8 distro jobs passed.

These results verify the current code/tests on the GitHub runners. They are not process-death recovery tests and are not a hardware-independent power-loss guarantee. Restore remains disabled.

## Handle-relative filesystem boundary research — 2026-11-08

A second review found that the current move primitive's Lstat preflight followed by a path-based rename is not, by itself, a defense against a hostile concurrent writer replacing the source entry between inspection and mutation. The journal's canonical path validation also does not close that race. The executor must state and enforce its filesystem threat model instead of describing the preflight as no-follow atomicity.

### Linux

- renameat2(old_dirfd, old_leaf, new_dirfd, new_leaf, RENAME_NOREPLACE) accepts directory file descriptors and resolves relative names from those pinned directory handles. It provides atomic destination no-replace behavior where the filesystem supports the flag.
- openat2 can constrain opening paths beneath a pinned directory with RESOLVE_BENEATH | RESOLVE_NO_MAGICLINKS | RESOLVE_NO_SYMLINKS; O_NOFOLLOW alone only affects the final component.
- renameat2 does not accept an already-open source file descriptor and has no openat2-style resolution flags. A pre-opened/hash-verified source FD therefore does not, by itself, bind the later rename to that exact inode.
- Consequently, Linux execution must use pinned directory handles, validated single-component leaf names, a state directory whose permissions prevent untrusted concurrent namespace mutation, and identity revalidation immediately before each rename. Unsupported RENAME_NOREPLACE filesystems remain fail-closed. If the state directory cannot be proven private, automatic execution is unsupported.

### Windows

- SetFileInformationByHandle operates on an already-open source file handle. FILE_RENAME_INFO documents a RootDirectory handle for resolving a relative destination name and specifies that a destination collision fails when replacement is false.
- The newer FileRenameInfoEx protocol structure defines a flags field; with FILE_RENAME_REPLACE_IF_EXISTS omitted, the rename must fail if the destination exists. Its RootDirectory is a directory handle for a relative destination. This is a better fit than a path-based MoveFileExW for binding the source object and destination directory.
- Microsoft documents that file-information classes may vary across OS releases and underlying drivers. The Win32 SetFileInformationByHandle documentation's published class table does not list FileRenameInfoEx, despite the protocol structure documentation. Treat the Ex class as an explicitly qualified capability, not a universal Windows promise.
- Windows implementation must open source with reparse-point-aware semantics, reject reparse points, retain source and destination-directory handles through the rename, omit replacement flags, and fail closed when the class/provider is unsupported. Qualification is NTFS-specific until separately tested on another filesystem. This does not remove the separate journal namespace durability and deletion limitations already recorded above.

### Required implementation order

1. Add platform capability probes and tests for the exact rename information class/flags before wiring any executor.
2. Replace path-only source handling with pinned directory handles and validated leaf names on Linux; use a retained source handle plus destination directory handle on qualified Windows.
3. Verify file identity from an opened handle, not by hashing a pathname and assuming it still names the same object.
4. Keep the state directory private against untrusted writers for the entire transaction. A check-then-rename race cannot be solved by a comment or another Lstat.
5. Add race-oriented tests that replace source/destination entries between inspection and mutation, plus process-death tests around each namespace operation.
6. Keep unsupported filesystem/provider combinations disabled. Do not fall back to os.Rename, MoveFileExW, or a weaker path-based operation to make a test pass.

### Sources reviewed

- Linux [renameat2(2)](https://man7.org/linux/man-pages/man2/renameat2.2.html): directory-FD-relative names and RENAME_NOREPLACE.
- Linux [openat2(2)](https://man7.org/linux/man-pages/man2/openat2.2.html): beneath/no-symlink resolution controls.
- Go [traversal-resistant file APIs](https://go.dev/blog/osroot): os.Root protections and its current operation boundaries.
- Microsoft [SetFileInformationByHandle](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-setfileinformationbyhandle): source-handle-based file metadata operations and provider variability.
- Microsoft [FILE_RENAME_INFO](https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_rename_info): no-replace collision behavior and destination RootDirectory.
- Microsoft [MS-FSCC FileRenameInformationEx](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/4217551b-d2c0-42cb-9dc1-69a716cf6d0c): Ex flags and mandatory no-replace behavior when the replacement flag is absent.

This research changes the executor implementation gate; it does not implement an executor and does not authorize Restore activation.

## Linux directory-handle move hardening — current verified boundary

The Linux no-replace primitive opens the common containing directory with `openat2(2)`, requesting `RESOLVE_NO_SYMLINKS | RESOLVE_NO_MAGICLINKS`, then calls `renameat2(2)` with that pinned directory FD and validated leaf names. It now also calls `fsync(2)` on the same parent-directory FD after the namespace mutation. Unsupported `openat2`, `RENAME_NOREPLACE`, or directory-sync behavior fails closed.

This pins the parent object and persists the changed directory entry according to the filesystem's documented sync contract. It does **not** bind the source leaf to a previously hashed inode: Linux `renameat2` is name-based. The executor must enforce the private state-directory boundary and verify the moved artifact's expected size/SHA-256 through a no-follow handle immediately after every move and before opening SQLite. A mismatch or verification failure preserves the destination and stops. This does not claim to eliminate all same-user concurrent mutation races.

The current implementation and regression tests are covered by Root CI [42772101619](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42772101619) on Ubuntu and Windows.

## Windows handle-bound rename API verification — 2026-11-08

The pinned production dependency was checked directly at the `golang.org/x/sys` tag `v0.49.0`, not inferred from current package documentation.

Verified in that exact tag:

- `windows.SetFileInformationByHandle(handle, class, buffer, size)` is available.
- `windows.FileRenameInfoEx` is defined as information class value `22`.
- The package does not expose a Go structure for the variable-length `FILE_RENAME_INFORMATION_EX` payload. The adapter must define a narrowly scoped local ABI structure/buffer and assert its layout rather than assume a convenient exported type exists.
- The call accepts a source file handle and an information-class buffer; this is the intended handle-bound source operation.
- The no-replace contract is explicit: leave `FILE_RENAME_REPLACE_IF_EXISTS` (0x1) unset. Do not set POSIX replacement semantics or unrelated flags.
- A rooted relative destination uses a retained destination directory handle in `RootDirectory` and a single validated UTF-16 leaf name. Network/redirector operations are out of scope because the protocol requires RootDirectory to be zero for network operations.

The Microsoft protocol structure describes a 32-bit Flags field, a reserved 32-bit alignment field, a 64-bit RootDirectory handle, a 32-bit FileNameLength, and a variable UTF-16 name. The minimum structure size is 24 bytes. The buffer builder must calculate the name byte length with overflow checks, omit the trailing NUL from FileNameLength, and allocate enough aligned storage for the fixed header plus the name and protocol minimum size.

Important limitation: Microsoft's Win32 SetFileInformationByHandle documentation currently lists FileRenameInfo but does not list FileRenameInfoEx in its published class table. The protocol specification defines FileRenameInformationEx, and x/sys exposes class value 22, but neither fact certifies support by every Windows version or filesystem provider. Therefore the implementation must capability-test the exact operation on the actual state volume, treat unsupported class/provider errors as unsupported, and keep automatic Restore disabled unless the filesystem is explicitly qualified. NTFS is the initial qualification target; ReFS, SMB, removable media, and other providers remain unsupported until separately tested.

This is API-surface verification and ABI design only. No Windows handle-bound rename implementation or NTFS runtime qualification is claimed by this entry.

### Sources

- Pinned [x/sys v0.49.0 windows/types_windows.go](https://github.com/golang/sys/blob/v0.49.0/windows/types_windows.go) — FileRenameInfoEx class value 22.
- Pinned [x/sys v0.49.0 windows/syscall_windows.go](https://github.com/golang/sys/blob/v0.49.0/windows/syscall_windows.go) and [generated syscall wrapper](https://github.com/golang/sys/blob/v0.49.0/windows/zsyscall_windows.go) — SetFileInformationByHandle declaration and implementation.
- Microsoft [MS-FSCC FileRenameInformationEx](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/4217551b-d2c0-42cb-9dc1-69a716cf6d0c) — field layout, RootDirectory semantics, and no-replace behavior when the replace flag is absent.
- Microsoft [FILE_RENAME_INFO](https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_rename_info) — Win32 union semantics and relative destination handle.
- Microsoft [SetFileInformationByHandle](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-setfileinformationbyhandle) — source handle access requirements and provider-dependent support caveat.


## Windows handle-bound no-replace move implementation — current verified boundary

The Windows move uses native `NtSetInformationFile` with `FileRenameInformationEx` class 65. Both source and destination are bound to retained handles:

- Opens the qualified state directory component-by-component from the volume root; rejects reparse/device/non-directory components and checks ancestor DACLs.
- Opens the source leaf relative to the retained directory handle using `NtCreateFile`, `FILE_OPEN_REPARSE_POINT`, `FILE_NON_DIRECTORY_FILE`, `FILE_SYNCHRONOUS_IO_NONALERT`, and native `FILE_WRITE_THROUGH` (CreateOptions 0x2).
- Requests the required source access, rejects reparse points and non-regular files, and verifies the opened source's location against the retained directory handle.
- Calls `FlushFileBuffers` on the source handle before rename, then issues `NtSetInformationFile(FileRenameInformationEx=65)` on that same handle with the destination directory handle in `RootDirectory` and the replace flag unset.
- Destination collisions fail; no `os.Rename`, `MoveFileExW`, replacement fallback or legacy path-based fallback is used for database artifact moves.

The Microsoft write-through contract covers NTFS metadata changes such as a rename resulting from processing the request. This is the documented NTFS boundary used by the implementation, not an all-provider or hardware-independent power-loss certification. The volume gate currently permits fixed local NTFS with persistent ACLs only.

### Verification

Root CI [42772101619](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42772101619) passed both platforms. Windows tests cover retained-directory substitution, decoy-tree rejection, no-replace collision, journal relative replacement and the native rename path. The CI result verifies behavior on the Windows runner; it does not simulate sudden power loss.

### Remaining boundary

The move helper is still isolated. It is not yet called by a transaction executor, and process-death/restart tests have not been implemented. Linux source-leaf identity remains name-based and requires the post-move verification step described above. Restore remains disabled.

## Current code revision verification — c89d532a — 2026-11-08

- Root CI 42497620421 passed Ubuntu and Windows; Ubuntu race detector passed. The new Windows relative-source-open regression test ran as part of the Windows Go package suite.
- SQLite driver evaluation 42497620420 passed 4/4.
- Wails platform proof 42497620429 passed 2/2, including Linux WebView-to-Go/database reopen and Windows native executable verification.
- Linux distribution matrix 42497620422 was in progress with 4/8 jobs passed at this update. Do not claim full matrix green for c89d532a yet.
- Code commit: c89d532a3ba04d43c7d2ccbed1675319c6f90ee8. Current branch also includes docs-only follow-up commits.
- Remaining Windows path boundary: destination directory path is opened by CreateFile; OPEN_REPARSE_POINT protects its final component only. Next implement component-by-component directory opening from a trusted volume root and verify private ACL/permissions before the executor.


## Current verified boundary — 2026-11-09

- Windows directory traversal opens each path component relative to a retained parent handle, rejects reparse/device/non-directory components, and validates the final directory owner/DACL and volume from handles.
- The final directory owner must equal the process TokenUser or explicit TokenOwner SID. Broad write-capable ACEs are rejected. The accepted volume policy is fixed local NTFS with persistent ACLs; other providers are unsupported.
- Root CI [42643886765](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42643886765) passed Ubuntu and Windows on `8cb7b7fd3d987fbb2e2fb0484ad67e2e3733baf5`; Ubuntu race tests passed. Wails [42642834036](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42642834036) passed both platforms.
- The exported legacy `RestoreDatabase` helper still performs its own standalone path-based publication and is not the transaction executor. Do not call it from the future journal executor or expose it through GUI/CLI. The executor must stage/verify a candidate separately, then publish only through the qualified no-replace adapter while holding the StoreController exclusive lease.
- Restore remains disabled. No complete executor, startup coordinator, shared GUI/CLI Restore application service, or abrupt-process-restart verification is claimed.


## Windows ancestor ACL closure — 2026-11-09

The Windows directory walk now checks the DACL of the trusted volume root and every opened ancestor component, not only the final state directory. Ancestor checks reject NULL DACLs, unsupported ACE layouts and broad write-capable ACEs; they do not require system-owned ancestors to be owned by the current user. The final canonical state directory additionally requires its owner to match TokenUser or the process token's explicit TokenOwner. The implementation is in commits `7f7d7ee773f5656ebbd414a1c9276446e1b48d34` and `781d44be75346ce7a5506fc1bc6b2e7b12ff3e96`; Root CI [42645966152](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42645966152) is pending. Do not treat this additional boundary as verified until both platform jobs complete.


## Windows namespace durability gate — initial unresolved review (superseded by the 2026-11-10 write-through documentation re-review)

The implementation is handle-bound, but the available Microsoft contract does not conclusively establish durable persistence of the parent-directory namespace update performed by `NtSetInformationFile(FileRenameInformationEx=65)`. The previous statement that the documentation explicitly guaranteed rename metadata durability was too strong and is withdrawn.

### What the implementation currently does

- Opens the source relative to the retained state-directory handle with native `NtCreateFile`, `FILE_OPEN_REPARSE_POINT`, and native `FILE_WRITE_THROUGH` CreateOptions 0x00000002.
- Calls `FlushFileBuffers` on the source handle before rename.
- Performs the no-replace rename through that same source handle using `NtSetInformationFile(FileRenameInformationEx=65)`; the replacement flag is omitted and the destination is rooted at the retained directory handle.
- Restricts the supported Windows volume to fixed local NTFS with persistent ACLs. Other filesystem/provider combinations remain fail-closed.

### What the primary documentation actually establishes

- Microsoft [File Caching](https://learn.microsoft.com/en-us/windows/win32/fileio/file-caching) says file-system metadata is always cached and that metadata changes must be flushed or the file opened with `FILE_FLAG_WRITE_THROUGH`. It does not explicitly describe the durability of the parent-directory entry changed by a handle-based rename.
- Microsoft [NtCreateFile](https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntcreatefile) describes native `FILE_WRITE_THROUGH` as requiring data written to the file to be transferred before a write request completes. That wording does not explicitly specify the crash-durability contract for `FileRenameInformationEx`.
- Microsoft [MS-FSCC FileRenameInformationEx](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/4217551b-d2c0-42cb-9dc1-69a716cf6d0c) specifies rename layout, relative `RootDirectory`, and no-replace behavior when the replace flag is absent. It defines no write-through or durable-directory-entry flag.
- Microsoft [FlushFileBuffers](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-flushfilebuffers) flushes buffers for the specified file handle. The reviewed contract does not say that flushing the renamed file handle also durably flushes its containing directory's namespace entry.
- Linux has a clearer documented boundary: [fsync(2)](https://man7.org/linux/man-pages/man2/fsync.2.html) explicitly says syncing a file does not necessarily sync its directory entry and requires a separate directory-FD `fsync`. The current Linux adapter performs that parent-directory sync after namespace mutation.

### Gate decision

The current Windows implementation is valuable identity protection and requests write-through behavior, but this review cannot certify the rename's directory-entry durability after power loss from the cited public contract alone. Do not describe the Windows namespace durability gate as resolved. Do not infer a guarantee from NTFS CI success or process-crash tests.

This is not limited to database-artifact moves. `WriteRestoreJournal` writes and flushes a temporary JSON file, then publishes/replaces `restore-journal.json`. The Windows publisher performs that replacement through `NtSetInformationFile(FileRenameInformationEx=65)` using the write-through source handle. Therefore every intent/completion journal phase currently depends on the same unproven namespace-durability boundary. A successful `WriteRestoreJournal` return must not be treated as a proven durable intent for a future executor until this gate is closed.

No ordinary-user directory-handle flush recipe with a documented, supported contract has been established in this review. Do not add an undocumented directory flush, a path-based fallback, `MoveFileExW`, or an overwrite-capable fallback to make the test pass. Keep the gate unresolved until a supported contract or a different journal/database protocol removes the dependency.

A targeted follow-up reviewed Microsoft's [NtFlushBuffersFileEx](https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntflushbuffersfileex) and [File Access Rights Constants](https://learn.microsoft.com/en-us/windows/win32/fileio/file-access-rights-constants). The flush API documents `STATUS_ACCESS_DENIED` when a handle has neither write nor append access, and its `FILE_DATA_SYNC_ONLY` mode is invalid for directory handles. The directory access-right table assigns `FILE_ADD_FILE` and `FILE_ADD_SUBDIRECTORY` the same bit values used for file write/append rights, but the reviewed `NtCreateFile` documentation does not define a supported directory-handle access mask for this flush operation. This leaves a plausible implementation question, not an approved contract. Do not infer support from a successful experiment on one Windows build or use elevated volume flush as a substitute; resolve the exact API/access/version contract first.

Linux continues to use `fsync` on the pinned parent-directory FD after successful no-replace namespace mutations. Its source-leaf selection race remains separate: the executor must verify size/SHA-256 from a no-follow handle after every move and before SQLite is opened.

The retained ledger lifecycle now permits sequential transactions without deleting terminal evidence. The filesystem transaction executor, startup recovery, rollback orchestration and process-death/failure-injection tests remain unimplemented.

### Required next implementation step

Implement the transaction executor as an isolated infrastructure component that consumes a fully staged candidate inside the private state directory. It must:

1. Acquire the StoreController exclusive lease and drain all shared operations.
2. Verify the staged candidate and previous database identities from no-follow handles.
3. Persist each intent phase before its corresponding filesystem mutation.
4. Use only the no-replace artifact move primitive; verify the moved identity immediately after each move and before any SQLite open.
5. Retain rollback and quarantine artifacts; never overwrite, guess, or delete ambiguous files.
6. Reopen and validate the published database; if validation fails, persist rollback intent, quarantine the candidate, restore the verified previous database, and verify the reopened rollback Store.
7. Fail closed when journal publication, directory sync, move verification, reopen or rollback is uncertain.
8. Add Linux/Windows failure-injection and abrupt-process-death/restart tests before any startup, GUI or CLI wiring.

Do not use the legacy path-based `RestoreDatabase` helper from this executor. Do not enable Restore until the complete test matrix passes on both platforms.



## Windows NTFS write-through rename contract — re-reviewed 2026-11-10

The earlier conclusion that the public Microsoft contract did not cover rename metadata was too conservative and is superseded for the project's explicitly qualified fixed-local-NTFS path.

Microsoft Learn's [CreateFileW documentation](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-createfilew), under Caching Behavior, explicitly states that a write-through request causes NTFS to flush metadata changes such as timestamp updates or a rename operation. The native [NtCreateFile documentation](https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntcreatefile) defines the native `FILE_WRITE_THROUGH` CreateOptions bit as `0x00000002`. The current adapter uses that native option in `openRestoreSourceWriteThrough`, calls `FlushFileBuffers` on the opened source, then passes the exact same source handle to `NtSetInformationFile(FileRenameInformationEx=65)`. The rename destination is anchored to the retained state-directory handle through `RootDirectory`; artifact moves omit the replace flag.

Together these documented contracts support the NTFS write-through metadata barrier for the handle-bound rename used by this implementation. The existing volume policy remains fixed local NTFS with persistent ACLs only. This does not establish hardware-independent power-loss safety: storage devices/controllers must honor flush requests. It does not qualify ReFS, SMB, removable media or other filesystem providers.

This closes the documented Windows namespace-metadata prerequisite for both artifact moves and the current journal-ledger replacement operation under the fixed-local-NTFS policy. The retained ledger removes the need to delete a terminal journal before starting another transaction. It does **not** complete process-death recovery or the filesystem executor; the 256-record / 1 MiB cap also remains a fail-closed retention limit.

