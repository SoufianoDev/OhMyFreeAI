# Implementation Blueprint

## Phase 0 — Architecture and policy

- Maintain the v3.3 project spec and this architecture set.
- Close the Colab Free policy/feasibility gate.
- Record Wails major-version decision only after PoC.
- Define licenses, privacy and third-party service disclosures.

## Phase 1 — Proven-path audit

- Inspect historical launcher, Gateway, SSE and monitor sources.
- Reproduce only allowed behavior in a controlled test environment.
- Record exact commands, versions, configs, failure modes and resource snapshots.
- Do not execute scripts that mutate the user's live system as part of the audit.

## Phase 2 — Feasibility proofs

- Wails v3.0.0 official starter and complete application-specific build passed on Ubuntu 24.04 and Windows; live WebView runtime call, packaging and distro coverage remain open.
- Go core compile/test.
- Official llama.cpp prebuilt distribution and backend compatibility.
- modernc.org/sqlite v1.62.0 selected and implemented in the production Store, ConfigRepository and transactional migration runner; crash recovery and backup API remain open.
- Colab policy-compatible use case or explicit no-go.

## Phase 3 — Bootstrap

- Go module and clean package layout.
- Developer toolkit and CI.
- Wails v3.0.0 desktop adapter with AppService GetSnapshot/SaveConfig bindings and generated native builds; live WebView invocation and full shutdown lifecycle still require tests.
- Material 3 token system and minimal application shell.

## Phase 4 — Domain and persistence

- Domain entities, typed errors and state machines.
- SQLite schema, embedded forward migrations and transactional migration runner.
- Config validation, persistence and secret references.

## Phase 5 — Process and security

- Linux/Windows process adapters.
- Ownership and shutdown.
- Local auth and privacy-safe diagnostics.

## Phase 6 — Runtime and engine

- Runtime capability contract.
- llama.cpp pinned distribution.
- Model loading, health and actual inference evidence.
- Runtime provider adapter only after policy clearance.

## Phase 7 — Storage and downloads

- Persistent artifact store, local staging, checksums.
- Download queue and resume/cancel.

## Phase 8 — Gateway and client

- OpenAI-compatible endpoints, SSE, cancellation and timeouts.
- OpenCode discovery/config backup/atomic update/rollback.

## Phase 9 — Monitoring and recovery

- Independent telemetry, freshness and non-blocking control.
- Bounded self-healing with explicit error classes.

## Phase 10 — Terminal, backup and UI

- Native local PTY.
- `.omfaibackup` export/import.
- Complete only UI flows backed by implemented services.

## Phase 11 — Hardening and release

- Linux/Windows test matrix, security/privacy review, packaging and uninstall.
- Verify all acceptance criteria and document unsupported states.
