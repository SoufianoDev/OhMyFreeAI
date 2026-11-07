# Task Plan: OhMyFreeAI Foundation

## Goal

Establish the verified architectural, repository, developer-tooling, and application foundations for OhMyFreeAI while preserving the free-only Colab + llama.cpp + local Gateway + OpenCode proof path.

## Next Step

Implement the selected Wails v3.0.0 desktop adapter with real application services and lifecycle handling. Implement the selected modernc.org/sqlite v1.62.0 persistence adapter and transactional migration runner. Keep the domain/application layers framework- and driver-independent. Then add actual WebView call tests, abrupt-crash recovery, backup API tests, packaging and Linux distro validation.

## Current Phase

Phase 4

## Phases

### Phase 1: Requirements & Discovery

- [x] Recover the agreed v3.3 architecture and project constraints from conversation context
- [x] Inspect the target repository and confirm its starting state
- [x] Inspect required Material Design and Kotlin Compose Kit skill repositories
- [x] Inspect the supplied prototype archives and Python scripts before adopting any behavior
- [x] Record findings and source boundaries
**Status:** complete

### Phase 2: Planning & Structure

- [x] Establish architecture documentation and repository conventions
- [x] Vendor Planning with Files v3.28.1, Material 3 v1.1.1, and Compose component-design guidance under .agents/skills/
- [x] Define module boundaries, ports/adapters, runtime lifecycle, and state contracts
- [x] Define compatibility layers including Colab resource fit and verification levels
- [x] Define omfaibackup format and backward-compatibility policy
**Status:** complete

### Phase 3: Implementation

- [x] Add the framework-independent Go domain foundation and standard-library developer toolkit
- [x] Add validated versioned configuration defaults, repository port and application service
- [x] Add SQLite initial/up and development-only down migrations with schema integrity tests
- [x] Add Linux/Windows GitHub Actions validation workflow
- [x] Add a pinned Wails v3.0.0 official vanilla-js starter build proof to GitHub Actions; do not lock the app to Wails until native jobs pass
- [x] Extend the Wails proof to generate and assert the starter's Go-to-JavaScript GreetService binding on both native runners
- [x] Select Wails v3.0.0 after the official starter and generated service-binding builds passed on Ubuntu 24.04 and Windows
- [x] Select modernc.org/sqlite v1.62.0 after the Linux/Windows driver matrix and production migration SQL passed
- [x] Add an isolated SQLite driver evaluation module and Ubuntu/Windows matrix for modernc and ncruces; keep experimental dependencies out of the production module
- [x] Establish the initial bilingual Material 3 web shell using the repository's Material Design skill
- [x] Preserve historical Colab scripts and launcher behavior unchanged; keep remote-inference activation policy-gated
**Status:** in_progress

### Phase 4: Testing & Verification

- [x] Run repository Go vet/tests/build, Python tests, frontend syntax and architecture checks on Ubuntu and Windows via GitHub Actions
- [x] Evaluate modernc and ncruces on Ubuntu and Windows, including the production initial migration SQL
- [ ] Run documentation/link validation and review all source changes
- [ ] Verify the original prototype path and identify unverified assumptions
- [ ] Add actual desktop WebView-to-Go invocation, crash-recovery and production backup/restore tests
- [ ] Record exact pass/fail results; do not equate compilation with runtime certification
**Status:** in_progress

### Phase 5: Delivery

- [ ] Review branch diff and ensure no unrelated/destructive changes
- [ ] Summarize implemented foundation, verification, and remaining blockers
- [ ] Provide the branch and next concrete engineering step
**Status:** pending

## Key Questions

1. What files and implementation already exist in the repository beyond the README?
2. Which exact Material Design 3 skill assets and Kotlin Compose Kit guidance are available?
3. Which parts of the historical Colab prototype are reproducible and safe to preserve?
4. Which Wails/Go/Colab combinations can be verified without introducing paid dependencies?

Resolved so far: Wails v3.0.0 official starter + GreetService binding builds passed on Ubuntu 24.04 and Windows. Root Go/Python/frontend/architecture CI passed on both OSes. modernc.org/sqlite v1.62.0 and ncruces v0.37.0 passed the evaluation matrix on both OSes, including applying migrations/001_initial.sql; modernc is selected for production. The actual Wails adapter, production SQLite adapter/migration runner, packaging, crash recovery and broad Linux distro support remain unimplemented. Colab remote-inference activation remains policy-gated.

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Work on foundation/bootstrap | Isolates the initial foundation from the default branch and avoids rewriting main prematurely |
| Free-only remains a hard constraint | Explicit project requirement; no paid services or alternative hosting suggestions |
| Preserve the Colab + llama.cpp + Gateway + OpenCode path until audited | It is the only proven direction in the existing project history |
| Backup extension/name is omfaibackup | Explicit user decision |
| Use Planning with Files structure under .planning | Persistent task state, isolated from future tasks |
| Select Wails v3.0.0 as the desktop shell | Official starter and generated Go service binding built on Ubuntu 24.04 and Windows; app-specific lifecycle and packaging still require implementation |
| Select modernc.org/sqlite v1.62.0 for the SQLite adapter | Both CGO-free candidates passed Linux/Windows smoke tests and the initial migration; modernc avoids the WASM VFS and documented per-connection memory overhead of ncruces |
| Do not claim the skill is installed in the agent runtime | The available environment has no writable agent-home mount; repository vendoring is the durable project-level route |
| Keep Wails version undecided | v3 is current stable but needs Go 1.26+/GTK4/WebKitGTK 6; v2 has a different platform baseline. Verify on Linux and Windows before selecting |

## Errors Encountered

| Error | Attempt | Resolution |
|-------|---------|------------|
| Git clone failed: DNS resolution for github.com unavailable in the container | 1 | Use the connected GitHub integration for repository reads/writes; do not repeat the network clone |
| Code Mode tool-call limit exceeded during bulk skill vendoring | 1 | Reconcile partial writes and split into smaller batches |
| GitHub create-file returned HTTP 422 because a path already existed | 1 | Fetch the existing file SHA and use update_file |
| Policy-gate Go test failed after adding required policy revision | 1 | Add the missing revision to the test fixture and rerun the full Go suite |
| Browser blocked file:// and localhost navigation | 1 | Use Playwright page.set_content() with inline local assets |
| Playwright locator matched both nav and quick-access buttons | 1 | Target unique data-page selectors and rerun successfully |
| Planning phase status markers initially included a list bullet | 1 | Corrected to the upstream `**Status:** value` format and manually revalidated all five phases |

## Notes

- Repository source of truth: https://github.com/SoufianoDev/OhMyFreeAI
- Planning skill source: https://github.com/OthmanAdi/planning-with-files
- The exact Planning with Files initializer could not be executed in this container because GitHub DNS/network access is unavailable and connector files are not mounted into the shell. Preserve the shipped template structure and use the permitted manual fallback; validate the format with the available validator source when possible.
