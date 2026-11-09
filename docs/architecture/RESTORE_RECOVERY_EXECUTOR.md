# Restore Recovery Executor — Design Constraints

**Status:** design gate; no filesystem executor is implemented yet.  
**Scope:** Linux and Windows desktop installations, with SQLite state files confined to the canonical application state directory.

## Why the executor is a separate safety boundary

The current journal and PlanRestoreRecovery are metadata persistence and a pure decision layer. They do not perform filesystem operations. os.Rename alone is not an acceptable publication primitive for recovery because replacement behavior differs by platform and can overwrite an unexpected destination. A path-level Lstat followed by a later rename also leaves a time-of-check/time-of-use window.

The executor must never infer ownership merely because a path appears in a journal. It must distinguish transaction artifacts from unexpected files and must preserve ambiguous state rather than deleting, replacing, or silently moving it.

## Required journal/state-machine changes

The current phases describe completed transitions. A crash between a filesystem mutation and the following journal update can leave a state that is intentionally rejected as ambiguous. Before automatic recovery is implemented, introduce durable intent phases so recovery can reason about both sides of each interrupted operation.

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

Implemented:
- Strict durable journal v3 with intent-phase names, canonical path confinement, candidate-quarantine metadata, and paired candidate/previous content identities.
- Identity-aware pure recovery planner for completed phases and all three intent phases: preserving-previous, publishing, and restoring-previous. Contradictory states return an empty plan and an ambiguity error.
- Linux no-replace move primitive: openat2 pins the containing directory while rejecting symlink/magic-link traversal; renameat2(RENAME_NOREPLACE) performs the namespace move.
- Windows no-replace move primitive: opens the state-directory path component-by-component from a volume-root handle, rejects reparse points, opens the source leaf relative to the pinned directory handle, and uses native NtSetInformationFile(FileRenameInformationEx=65) without the replace flag.
- Windows final-directory owner/DACL and fixed-local-NTFS/persistent-ACL qualification gate.
- StoreController leases and lower-level SQLite backup/staging primitives.

Not implemented:
- Filesystem transaction executor, startup recovery coordinator, process-death/failure-injection executor tests, and application-level restore orchestration.
- A Linux source-leaf identity binding across the final name-based rename. The current Linux primitive pins the parent directory, but its source leaf is still selected by name after earlier path validation. The executor must verify the moved artifact's expected size and SHA-256 from a no-follow opened handle immediately after each move and before any database open; on mismatch it must preserve the artifact and stop. It must never use an overwrite-capable fallback.
- Explicit power-loss qualification beyond the documented platform sync contracts.

The latest recovery-planner regression test exposed a lost-candidate state in preserving-previous. The planner now requires a verified candidate at staging for that phase. Root CI [42649675343](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42649675343) completed SUCCESS on commit 7b0ad259b5a741229e0d492b38c2a38b7a1dc3ec: Ubuntu and Windows passed; Linux race detection passed.

Durability constraints: Linux fsync(2) does not by itself persist the containing directory entry; the parent directory must also be synced. On Windows, the journal publisher uses MOVEFILE_WRITE_THROUGH, while journal deletion remains unsupported rather than reporting false success. These are explicit platform boundaries, not a blanket cross-platform power-loss guarantee.

Restore remains disabled. Ambiguous artifacts must remain untouched; no startup recovery, GUI/CLI Restore, or application-level Restore service may be enabled until the executor's identity, no-replace, sync, failure-injection, and restart tests pass on Linux and Windows.


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

Implemented as an isolated primitive, not yet called by a restore executor:

- Linux: golang.org/x/sys/unix.Renameat2 with unix.RENAME_NOREPLACE. Unsupported-kernel/filesystem errors are returned unchanged; there is no replacement-capable fallback.
- Windows: golang.org/x/sys/windows.MoveFileEx with MOVEFILE_WRITE_THROUGH, deliberately omitting MOVEFILE_REPLACE_EXISTING and MOVEFILE_COPY_ALLOWED. Same-directory validation keeps the operation on one volume; existing destinations fail and are mapped to os.ErrExist.
- Shared preflight rejects empty/equal paths, cross-directory moves, and sources that are symlinks or non-regular files. The OS operation remains the authoritative atomic destination-collision check.

The primitive does not sync the containing directory, persist journal phases, verify database content identities, or coordinate StoreController leases. The executor must perform those operations in the prescribed order. The Lstat preflight is not a standalone defense against a hostile writer racing inside the state directory; callers must supply the canonical private state directory and revalidate transaction paths immediately before mutation. No executor/startup integration is claimed.

### Research sources

- Linux renameat2(RENAME_NOREPLACE): https://www.man7.org/linux/man-pages/man2/rename.2.html
- Microsoft MoveFileExW flags and write-through contract: https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw
- Go x/sys/unix: https://pkg.go.dev/golang.org/x/sys@v0.49.0/unix#Renameat2
- Go x/sys/windows: https://pkg.go.dev/golang.org/x/sys@v0.49.0/windows#MoveFileEx

### Verification boundary

Test-first commit 61a19ea5932bbda9ef722e8789c5af42cabe4c45 produced expected RED in Root Go CI run 42438137306: go vet failed because moveRestoreFileNoReplace did not yet exist. The implementation and platform-specific tests were then added. On code commit d717eac7096d98c3b59b1b9ebd39aa1cff6e2d93, Root Go CI run 42438368020 passed on Ubuntu and Windows, including Linux race tests; SQLite driver matrix run 42438368022 passed 4/4. Wails run 42438368038 and Linux distribution run 42438368003 were still running at the time of this update. A separate retry of the failed Ubuntu 22.04 GTK3 WebView smoke from run 42437541325 was also in progress. Do not infer those pending results.


## Journal namespace durability by platform — 2026-11-08

The previous shared `syncRestoreDirectory` helper silently returned success on Windows. That was not a valid equivalent of POSIX directory `fsync`, and it overstated the journal's durability contract.

The journal publication path is now platform-specific:

- Non-Windows: write and sync the temporary journal, atomically rename it over the current journal in the same directory, then sync the containing directory. A directory-sync error is returned.
- Windows: write and sync the temporary journal, then publish it with `MoveFileExW(MOVEFILE_REPLACE_EXISTING | MOVEFILE_WRITE_THROUGH)`. This requests the documented write-through move behavior; the code no longer pretends that `os.Open(directory).Sync()` is the Windows equivalent.
- Windows journal removal currently fails closed with `errRestoreJournalRemovalUnsupported`; it leaves the journal intact. There is no production caller today, and no destructive cleanup should depend on an unproven namespace-deletion barrier.

### Windows durability boundary

Microsoft documents that `FlushFileBuffers` requires a handle with `GENERIC_WRITE`, and its directory-handle guidance lists supported directory-handle APIs without listing `FlushFileBuffers`. Microsoft also documents `MOVEFILE_WRITE_THROUGH` as not returning until the move is performed on disk, while the explicit flush guarantee in its flag description is phrased for copy-and-delete moves. This implementation uses same-directory/same-volume replacement and requests write-through, but this alone is not an end-to-end power-loss certification for the Restore protocol.

Restore remains disabled. Before Windows Restore can be enabled, the full executor must reconcile source/target identities after every move result, retain rollback artifacts, and pass process-death/restart tests on NTFS. Other Windows filesystem classes remain unverified and must fail closed unless separately qualified. Hardware and storage stacks that falsely acknowledge flush/write-through are outside the application's guarantee.

### Research sources

- Linux `fsync(2)`: https://man7.org/linux/man-pages/man2/fsync.2.html — file sync does not necessarily persist the containing directory entry; the directory itself must be synced.
- Microsoft `FlushFileBuffers`: https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-flushfilebuffers — requires `GENERIC_WRITE`.
- Microsoft directory handles: https://learn.microsoft.com/en-us/windows/win32/fileio/obtaining-a-handle-to-a-directory — lists documented consumers of directory handles.
- Microsoft `MoveFileExW`: https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw — replacement and write-through flags.
- Microsoft file caching: https://learn.microsoft.com/en-us/windows/win32/fileio/file-caching — file metadata is cached and must be flushed or opened write-through.
- SQLite corruption guidance: https://www.sqlite.org/howtocorrupt.html — a hot journal/WAL must remain paired with its database; moving or deleting it can prevent automatic recovery.
- SQLite atomic commit: https://sqlite.org/atomiccommit.html — SQLite's crash tests inject partial writes, garbage sectors, and reordered writes, then reopen and verify transaction consistency.

### Verification boundary

Test-first commits `f876c42f215a80340245fb4334af6ab2ba9caa97` and `db2eddd33ff7c76a216c8dfce67565c0506841c3` added the journal publication and Windows fail-closed deletion tests before implementation. Their GitHub Actions attempts were cancelled before test execution because later commits superseded them; therefore no RED result is claimed. Implementation commits: `a51e9589c77795441a56de56670db17ec8ed1756`, `45c5329e0524c736b8c31eda1c833a0326a18ea3`, `cbccd640b8aa154b58abc293ea054613fafa36f5`, `a68bba150a627b143c9cf95ca6a1ed918de742d8`, `1cfd08056d2ecfa1afe31910c38cfde0886b415b`, `673782ca5aa49c0c32bfadb17234a20bdfdb4924`, and `48bc9401ea070bea00cb155fe7ab91c4b72b2f6e`. CI for the latest code revision is not yet verified; the Ubuntu root job is queued and the Windows root job was cancelled in the latest attempt. Do not infer success.
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

## Linux directory-handle move hardening — 2026-11-08

The Linux no-replace primitive now opens the common containing directory with `openat2(2)`, requesting `RESOLVE_NO_SYMLINKS | RESOLVE_NO_MAGICLINKS`, then calls `renameat2(2)` with that pinned directory FD and validated leaf names. It no longer resolves the containing directory again through `AT_FDCWD` at the actual rename. Kernels that do not support `openat2` fail closed; there is no path-based fallback.

This closes parent-directory symlink traversal and pins the directory object across the rename. It does **not** bind the source leaf to a previously hashed inode: Linux `renameat2` is name-based. The executor must still enforce a private state directory against untrusted concurrent writers and verify candidate/previous identity immediately before mutation. The move primitive itself is not a complete executor and does not perform directory sync or journal transitions.

The regression test for a symlinked containing directory was committed first in `b529fae0b3f92ad0e6f3a04f9150df9ca78c3895`. Its CI outcome must be checked separately; do not claim a RED result until the run has actually executed the relevant test.


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


## Windows handle-bound no-replace move implementation — 2026-11-08

The Windows move uses NtSetInformationFile with native FileRenameInformationEx class 65. The source file is now opened relative to the retained destination directory handle using NtCreateFile and OBJECT_ATTRIBUTES.RootDirectory, rather than by re-resolving the source's full path.

- Opens the destination directory handle and retains it through source open and rename.
- Opens the source leaf relative to that handle using NtCreateFile, FILE_OPEN_REPARSE_POINT, FILE_NON_DIRECTORY_FILE and FILE_SYNCHRONOUS_IO_NONALERT, requesting SYNCHRONIZE, DELETE and FILE_READ_ATTRIBUTES.
- Inspects source and directory handles; rejects reparse points, directories and devices as applicable.
- Uses native FileRenameInformationEx class 65 with zero flags and the same retained directory handle as RootDirectory. Destination collision maps to os.ErrExist; no replacement or legacy fallback is used.

### Verification

- The source-directory substitution test was added first in commit 3f1110029423cdf5893732f086af8d385dd02b8d and produced observed RED in Root CI 42497376480 / Windows job 128782981057.
- The first implementation attempt failed vet because NtCreateFile's EA-buffer parameter is uintptr, not a pointer; corrected to zero in commit c89d532a3ba04d43c7d2ccbed1675319c6f90ee8.
- Root CI 42497620421 passed Ubuntu and Windows; Ubuntu race detector passed.
- SQLite driver evaluation 42497620420 passed 4/4.
- Wails run 42497620429 was in progress and Linux distro matrix 42497620422 was queued at this update.

### Explicit limitations and next step

The destination directory is still initially opened by full path through CreateFile. FILE_FLAG_OPEN_REPARSE_POINT protects only the final component, not all ancestors. Next, open the canonical state-directory path component-by-component from a trusted volume root, reject every reparse component, and verify private directory permissions/ACL. The future executor must hash and revalidate source identity from the opened handle immediately before mutation, persist intent before every move, sync the required namespace state, reconcile ambiguous results and recover process death.

The native rename class is not assumed to work on every filesystem/provider. Unsupported providers fail closed. No full executor, startup recovery or app-level Restore service exists. Restore remains disabled.


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
