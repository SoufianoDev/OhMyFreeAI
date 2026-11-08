---
name: deslop-code
description: Review the current change for unnecessary AI-generated code patterns after implementation. Remove only demonstrably redundant comments, abnormal defensive code, workaround casts, and style inconsistent with nearby code.
---

# Deslop Code

Source: https://github.com/brianlovin/agent-config/blob/main/skills/deslop/SKILL.md

Run this as a focused cleanup pass after implementation. Compare the current task's diff with the repository baseline and inspect nearby code before deciding that a pattern is unnecessary.

Check for:
- Comments that merely narrate obvious code or repeat nearby documentation.
- Defensive branches, try/catch blocks, or validation duplicated by a trusted and already-validated caller.
- Type casts or escape hatches added only to silence a type error instead of fixing the type contract.
- Naming, formatting, abstractions, and control flow inconsistent with the surrounding module.
- Dead code or speculative wrappers introduced by the current change.

Rules:
- Remove only code introduced by the current task and only when its redundancy is proven by caller tracing and tests.
- Do not change behavior during a comment-only cleanup. For broader simplification, explain the behavior change and test it as a normal implementation change.
- Do not remove input validation at trust boundaries, authorization checks, error handling, cancellation, observability required for diagnosis, or defensive checks that protect public interfaces.
- Never remove checks protecting filesystem paths, symlink boundaries, atomic no-overwrite operations, SQLite integrity, rollback, data loss, secrets, or Colab policy compliance.
- Do not replace precise types with `any`, unsafe casts, ignored errors, or weaker contracts.
- Follow Ponytail Ultra: deletion is preferred only when the result remains complete, correct, and safe. Existing code style and architecture take precedence over generic preferences.
- Do not edit generated, vendored, lock, or user-owned pre-existing changes.
- Run focused tests after any behavioral cleanup; preserve and extend tests rather than weakening them.

Finish with a concise report of what was removed, what was intentionally retained, and which checks ran. Never claim a cleanup is safe merely because the diff is shorter.
