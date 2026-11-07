# Developer Toolkit

## Goal

A predictable Python standard-library-only entry point for developer tasks. No shell interpolation of user input, no automatic sudo, no destructive Git behavior, no silent installs and no network downloads without explicit user action.

## Planned commands

- `python3 scripts/dev.py doctor` — inspect Go, Node, npm, Python, Wails, platform dependencies and project prerequisites.
- `setup` — explain/install only explicitly approved dependencies.
- `dev` — launch development workflow.
- `test`, `lint`, `fmt`, `build` — deterministic checks.
- `deps` — report dependency versions and provenance.
- `diagnose` — collect redacted diagnostics.
- `clean` — preview first; never delete user artifacts or model data.
- `arch-check` — validate import boundaries and required architecture docs.

## Testing

Use Go unit tests, Python unittest, adapter contract tests, Gateway/SSE integration tests, Linux/Windows CI, and end-to-end tests of only policy-compliant runtime flows. A successful compile is not a runtime certification.
