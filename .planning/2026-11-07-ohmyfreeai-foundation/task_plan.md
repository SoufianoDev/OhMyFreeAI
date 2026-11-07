# Task Plan: OhMyFreeAI Foundation

## Goal

Establish the verified architectural, repository, developer-tooling, and application foundations for OhMyFreeAI while preserving the free-only Colab + llama.cpp + local Gateway + OpenCode proof path.

## Next Step

Continue from the implemented Wails v3.0.0 shell and modernc.org/sqlite v1.62.0 persistence adapter. Next add a live WebView-to-Go invocation test, validate the JavaScriptCore signal setting on Wails v3, add abrupt-crash recovery and SQLite backup API/restore, then finish packaging, Linux distro coverage and full diff review.

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
- [x] Add a pinned Wails v3.0.0 official vanilla-js starter build proof to GitHub Actions
- [x] Extend the Wails proof to generate and assert the starter's Go-to-JavaScript GreetService binding on both native runners
- [x] Select Wails v3.0.0 after official and application-specific builds passed on Ubuntu 24.04 and Windows
- [x] Add the actual Wails AppService with GetSnapshot and SaveConfig methods; generate and verify the JavaScript binding
- [x] Add a reproducible Python build script that generates the official Wails template in a temporary workspace and builds the real application
- [x] Select modernc.org/sqlite v1.62.0 after the Linux/Windows driver matrix and production migration SQL passed
- [x] Implement the production SQLite Store, per-connection PRAGMAs, transactional migration runner and validated config repository
- [x] Pin modernc dependency graph in go.mod/go.sum and verify tidy does not change committed lockfiles
- [x] Add an isolated SQLite driver evaluation module and Ubuntu/Windows matrix for modernc and ncruces; keep experimental dependencies out of the production module
- [x] Establish the initial bilingual Material 3 web shell using the repository's Material Design skill
- [x] Preserve historical Colab scripts and launcher behavior unchanged; keep remote-inference activation policy-gated
**Status:** complete

### Phase 4: Testing & Verification

- [x] Run repository Go vet/tests/build, Python tests, frontend syntax and architecture checks on Ubuntu and Windows via GitHub Actions
- [x] Evaluate modernc and ncruces on Ubuntu and Windows, including the production initial migration SQL
- [x] Test the production SQLite adapter, migration rollback/idempotence, version-gap and corrupt-history rejection, config round-trip, file permissions and per-connection PRAGMAs on Ubuntu and Windows
- [x] Build the complete Wails application on Ubuntu and Windows; generate 1 service, 2 methods and 7 models, produce the Vite production bundle and verify the native executable
- [x] Pin and verify Go dependency lockfiles in CI
- [x] Upgrade GitHub Actions to Node 24-compatible releases; latest core, SQLite and Wails workflows pass without the Node 20 deprecation warning
- [ ] Run documentation/link validation and review all source changes
- [x] Perform a read-only static audit of the supplied prototype archives and Python patch/monitor scripts; record hashes, preserved invariants and risks in docs/architecture/PROTOTYPE_AUDIT.md
- [ ] Revalidate current Colab /props and /v1/models behavior, model-specific Q8_0/262144 compatibility and the live SSE path before enabling remote inference
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

Resolved so far: the complete Wails v3.0.0 application now builds on Ubuntu 24.04 and Windows, with one registered AppService, two generated methods (GetSnapshot/SaveConfig), seven generated models, a Vite production bundle and verified native executables. The production modernc.org/sqlite v1.62.0 Store, transactional migration runner and config repository are implemented and tested on both OSes. Root Go/Python/frontend/architecture CI and the separate modernc/ncruces evaluation matrix passed. Still open: live WebView invocation testing, Wails v3 JavaScriptCore signal validation, abrupt-crash recovery, SQLite backup/restore API, long-duration stress/performance, dependency license inventory, current Colab/model/SSE revalidation, broader Linux distro coverage and packaging/uninstall. Colab remote-inference activation remains policy-gated.

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Work on foundation/bootstrap | Isolates the initial foundation from the default branch and avoids rewriting main prematurely |
| Free-only remains a hard constraint | Explicit project requirement; no paid services or alternative hosting suggestions |
| Preserve the Colab + llama.cpp + Gateway + OpenCode path until audited | It is the only proven direction in the existing project history |
| Backup extension/name is omfaibackup | Explicit user decision |
| Use Planning with Files structure under .planning | Persistent task state, isolated from future tasks |
| Select Wails v3.0.0 as the desktop shell | The actual OhMyFreeAI app, AppService binding and Vite production bundle build on Ubuntu 24.04 and Windows; live WebView invocation and packaging remain unverified |
| Select modernc.org/sqlite v1.62.0 for the SQLite adapter | Both CGO-free candidates passed Linux/Windows smoke tests and the initial migration; modernc avoids the WASM VFS and documented per-connection memory overhead of ncruces |
| Do not claim the skill is installed in the agent runtime | The available environment has no writable agent-home mount; repository vendoring is the durable project-level route |
| Keep Wails confined to the desktop adapter | Core domain/application packages remain Wails-independent; the desktop build is generated from the pinned official vanilla-js template in a temporary workspace |
| Keep the historical Colab monitor read-only | Preserve the busy-session guard and SSE framing invariant, but do not port Linux-only Wails v2 UI/process code into the cross-platform product |

## Errors Encountered

| Error | Attempt | Resolution |
|-------|---------|------------|
| Git clone failed: DNS resolution for github.com unavailable in the container | 1 | Use the connected GitHub integration for repository reads/writes; do not repeat the network clone |
| Code Mode tool-call limit exceeded during bulk skill vendoring | 1 | Reconcile partial writes and split into smaller batches |
| GitHub create-file returned HTTP 422 because a path already existed | 1 | Fetch the existing file SHA and use update_file |
| Policy-gate Go test failed after adding required policy revision | 1 | Add the missing revision to the test fixture and rerun the full Go suite |
| Browser blocked file:// and localhost navigation | 1 | Use Playwright page.set_content() with inline local assets |
| Playwright locator matched both nav and quick-access buttons | 1 | Target unique data-page selectors and rerun successfully |
| Planning phase status markers initially included a list bullet | 1 | Corrected to the upstream **Status:** value format and manually revalidated all five phases |
| Initial Wails build script passed the project folder as the CLI destination parent | 1 | Read the Wails init output, pass a workspace parent to -d, and rebuild the real app successfully on both OSes |
| Go test source contained a literal newline inside a quoted format string | 1 | Escaped the newline and reran Go vet/tests/build on Ubuntu and Windows |
| SQL splitter source template emitted malformed quote/newline rune literals | 1 | Replaced ambiguous escaped byte literals with numeric byte constants and reran the full Go matrix |
| Windows Wails build printed a non-fatal "-VThe system cannot find the file specified." line | 1 | Native executable and binding generation still succeeded; source of the message is not yet identified and remains a tracked warning |

## Notes

- Repository source of truth: https://github.com/SoufianoDev/OhMyFreeAI
- Planning skill source: https://github.com/OthmanAdi/planning-with-files
- The exact Planning with Files initializer could not be executed in this container because GitHub DNS/network access is unavailable and connector files are not mounted into the shell. Preserve the shipped template structure and use the permitted manual fallback; validate the format with the available validator source when possible.
