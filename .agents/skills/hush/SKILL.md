---
name: hush
description: Use on every code write, edit, refactor or review. Keeps useless comments out of code, removes any that slipped in (only on lines you changed), then formats the touched files with the project's own formatter.
---

# Hush

Source: https://github.com/SoufianoDev/hush/blob/dev/skills/hush/SKILL.md

Apply Hush to every code change in this repository. The default is no comment; a comment must explain information that code cannot show.

## Writing rules

- Comment only to explain why an approach exists, a hidden constraint, workaround, unit, invariant, or surprising edge case.
- Never add comments that restate code, narrate edits, guess intent, repeat a signature, or contain commented-out code.
- Prefer a clearer local/private name over a comment when safe. Never rename public APIs just to avoid a comment.
- Explicit user or project instructions about comments take precedence.

## Cleanup scope

- Review only comments in code added or changed in the current task. Never perform a repository-wide cleanup without listing candidates and obtaining approval.
- Leave pre-existing, generated, vendored, and lock-file content alone.
- Preserve directives and tool-consumed comments, including build tags, go: directives, lint suppressions, type-checker directives, formatter controls, coverage markers, shebangs, license headers, required API documentation, Python docstrings, and TODOs with an owner or issue link.
- When uncertain, keep the comment and shorten it rather than deleting potentially useful knowledge.
- Distinguish actual comments from comment-like text inside strings, regular expressions, templates, and trailing code.

## Verification and formatting

- After removal, inspect the diff and confirm it contains only intended comment and whitespace changes from the cleanup pass.
- Run the project's existing fast validation (tests, lint, or typecheck). Restore a comment if its removal breaks tooling.
- Format only files touched in this task, using the formatter declared by project configuration. Do not introduce a formatter or reformat unrelated files.
- If the host already formats the files, do not run a second formatter.

## OhMyFreeAI safety precedence

Hush never authorizes removing comments that explain SQLite/Restore invariants, filesystem race handling, trust boundaries, Colab policy constraints, rollback behavior, platform workarounds, or security assumptions. It never overrides Ponytail Ultra, architecture decisions, explicit requirements, or tests. A smaller diff is not a reason to discard safety-critical knowledge.

## Report

Report comments removed, comments retained and why, and the formatter/check actually run. Do not claim formatting or validation that was not performed.
