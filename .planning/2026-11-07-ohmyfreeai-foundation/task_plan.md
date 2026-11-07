# Task Plan: OhMyFreeAI Foundation

## Goal

Establish the verified architectural, repository, developer-tooling, and application foundations for OhMyFreeAI while preserving the free-only Colab + llama.cpp + local Gateway + OpenCode proof path.

## Next Step

CLI smoke is verified on all eight Linux distro lanes. Read-only archive inspection, native export/inspect dialogs and bilingual settings UI are implemented; verify the current eight-distro retest for this backup-enabled revision. Next: add safe CLI export/inspect entry points, then design coordinated SQLite close/reopen/rollback before exposing restore. Keep Colab remote inference policy-gated.

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
- [x] Add SQLite online backup/restore primitives with integrity, foreign-key, migration-history, cancellation, no-overwrite, target-safety and rollback tests on Ubuntu and Windows
- [x] Implement strict .omfaibackup snapshot archive with manifest, SHA-256, fixed entries, size limits and path-traversal protections; round-trip, rollback, checksum, overwrite and cancellation tests pass on Linux/Windows
- [x] Implement read-only Manager.Inspect with isolated temporary extraction, full manifest/checksum/schema validation and cleanup tests\n- [x] Wire backup export/inspect into the Wails service using native file dialogs and bilingual settings UI\n- [ ] Add backup export/inspect CLI commands and a coordinated Store lifecycle before enabling restore
- [x] Previous native Ubuntu 24.04 WebView-to-Go and two-launch DB reopen smoke passed (Wails run 42404058636)
- [x] Revalidate the refactored two-stage bridge bootstrap on native Ubuntu; run 42405729902 passed with two GetSnapshot calls per launch and DB reopen
- [x] Complete the same WebView-to-Go and two-launch DB reopen smoke on all eight Linux distro jobs (run 42405827037)
- [x] Diagnose GitHub runner WebKit bubblewrap namespace restriction and keep WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS=1 confined to the smoke-test subprocess only
- [ ] Validate WebKit's production sandbox on a real Linux host; container CI explicitly does not certify sandbox behavior
- [x] Define explicit Linux distribution targets and add GTK4/WebKitGTK 6.0 plus legacy GTK3/WebKit2GTK 4.1 build/runtime smoke matrix (Ubuntu 24.04, Debian 13, Fedora 44, Arch, openSUSE Tumbleweed, Ubuntu 22.04, Debian 12, Fedora 44 legacy)
- [x] Verify all eight Linux distro jobs pass in run 42405827037; distinguish container userland coverage from bare-metal certification
- [x] Implement framework-independent GUI-default CLI parser with typed unknown-command/extra-argument errors and unit tests
- [x] Verify the built native executable handles `--help`, `--version`, `doctor`, unknown commands and read-only doctor behavior on Ubuntu 24.04 and Windows (Wails run 42409974970)
- [x] Verify native CLI smoke on all eight Linux distro lanes (run 42410214638)
- [ ] Add abrupt-process crash/reopen recovery tests
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

Resolved so far: Wails v3.0.0 native builds pass on Ubuntu 24.04 and Windows. CLI dispatch and its native executable smoke pass on both platforms and all eight Linux distro lanes (run 42410214638). The backup-enabled revision builds on Ubuntu/Windows; Windows native executable smoke passes, and Ubuntu passes CLI plus two-stage WebView-to-Go calls and SQLite reopen (run 42412554344). Root CI passes on Ubuntu/Windows (run 42412587313). The read-only `Manager.Inspect` validates the complete archive in temporary storage, removes it afterwards, and does not modify the live DB; export/inspect are wired to native Wails dialogs and bilingual Settings controls. Current eight-distro retest for this revision is run 42412554358; at the latest check four GTK4 jobs were building and four GTK3/openSUSE jobs were queued. The CI-only WebKit sandbox bypass remains confined to smoke-test subprocesses; real-host production sandbox validation remains open. Still open: CLI backup commands, coordinated Store lifecycle and restore integration, abrupt-crash recovery, long-duration stress/performance, dependency license inventory, current Colab/model/SSE revalidation, packaging/uninstall and full source review. Colab remote-inference activation remains policy-gated.

## Decisions Made

| Decision | Rationale |
|----------|-----------|
| Work on foundation/bootstrap | Isolates the initial foundation from the default branch and avoids rewriting main prematurely |
| Free-only remains a hard constraint | Explicit project requirement; no paid services or alternative hosting suggestions |
| Preserve the Colab + llama.cpp + Gateway + OpenCode path until audited | It is the only proven direction in the existing project history |
| Backup extension/name is omfaibackup | Explicit user decision |
| Use Planning with Files structure under .planning | Persistent task state, isolated from future tasks |
| Use native Ubuntu/Windows CI plus distro-container compatibility lanes | Ubuntu 24.04, Debian 13, Fedora 44, Arch rolling and openSUSE Tumbleweed exercise GTK4/WebKitGTK 6.0; Ubuntu 22.04, Debian 12 and Fedora 44 legacy exercise the Wails v3.0.x GTK3/WebKit2GTK 4.1 path. Container results are not bare-metal certification |
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
| Wails smoke run 42399602513 failed its Ubuntu WebView readiness assertion after 12 seconds; output only showed a DRI3/EGL warning | 1 | Do not assume the warning is causal. Replace stdout grep with an explicit opt-in marker written by GetSnapshot, force software rendering under Xvfb, and wait for the actual JS-to-Go invocation |
| Windows Wails build printed a non-fatal "-VThe system cannot find the file specified." line | 1 | Native executable and binding generation still succeeded; source of the message is not yet identified and remains a tracked warning |

## Notes

- Repository source of truth: https://github.com/SoufianoDev/OhMyFreeAI
- Planning skill source: https://github.com/OthmanAdi/planning-with-files
- The exact Planning with Files initializer could not be executed in this container because GitHub DNS/network access is unavailable and connector files are not mounted into the shell. Preserve the shipped template structure and use the permitted manual fallback; validate the format with the available validator source when possible.
