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
- Use a platform adapter with atomic no-replace semantics. Linux should use renameat2 with RENAME_NOREPLACE; Windows should use a move operation that fails when the destination exists and requests write-through behavior. If the filesystem or OS cannot provide the required semantics, fail closed rather than falling back to overwrite-capable rename.
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

Implemented: durable journal v2, canonical path confinement, non-regular path rejection, pure conservative recovery planner, StoreController leases, and lower-level SQLite staging primitive.

Not implemented: intent phases, candidate quarantine metadata, database content identities, platform-specific no-replace move adapter, filesystem executor, startup coordinator, process-death executor tests, or application-level restore orchestration.

This design deliberately blocks a quick implementation based on os.Rename or check-then-rename. Safety takes priority over automatic cleanup: ambiguous artifacts are preserved and Restore remains disabled.
