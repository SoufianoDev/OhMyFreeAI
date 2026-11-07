# Progress Log

## Session: 2026-11-07

### Phase 1: Requirements & Discovery

- **Status:** complete
- **Started:** 2026-11-07T15:51Z
- Actions taken:
  - Recovered the full OhMyFreeAI v3.3 architecture commitments from project conversation context.
  - Read Planning with Files SKILL.md; fetched version 3.28.1 source.
  - Confirmed repository README is only `# OhMyFreeAI`.
  - Created branch `foundation/bootstrap` from `main`.
  - Attempted local git clone; failed because the container could not resolve github.com. Switched to GitHub connector for repository operations.
  - Created task planning files using the upstream template structure as the documented fallback because the initializer cannot be executed in the shell environment.
  - Inspected both supplied Wails prototypes: Go 1.24, Wails v2.10.2, embedded frontend; the remote-complete archive includes remote Colab polling and local monitoring. Did not adopt the rejected visual design.
  - Parsed all three uploaded Python scripts with ast.parse successfully; did not execute any script because they modify or interact with the user's live runtime.
  - Read the user's Material Design 3 skill and key references; confirmed web-specific limitations (Material Web maintenance mode, no M3 Expressive parity).
  - Read KotlinComposeKit's reusable component design skill; extracted only transferable component API principles, not Kotlin/Compose implementation details.
  - Vendored Planning with Files v3.28.1, Material 3 v1.1.1, and the KotlinComposeKit compose-component-design skill plus their source references/licenses into .agents/skills/ on the project branch. These copies are project-scoped; they do not install user-wide host hooks in this ChatGPT runtime.
  - Confirmed local toolchain: Go 1.23.2, Node 22.16.0, npm 10.9.2, Python 3.13.5. Wails v3's official docs require Go 1.26+, so this container cannot certify a Wails v3 build.
- Files created/modified:
  - .planning/2026-11-07-ohmyfreeai-foundation/task_plan.md
  - .planning/2026-11-07-ohmyfreeai-foundation/findings.md
  - .planning/2026-11-07-ohmyfreeai-foundation/progress.md

### Phase 2: Planning & Structure

- **Status:** complete
- Actions taken:
  - Established the project specification, system architecture, module boundaries, domain model, state machines, compatibility, runtime/engine, Gateway/client, storage/backup, security/recovery, UI/UX, developer toolkit and implementation blueprint documents.
  - Added platform, process, monitoring, downloads, configuration, CLI, persistent storage, terminal, client integration, privacy, event and data architecture contracts.
  - Added ADRs for Wails evaluation, Colab policy gate and omfaibackup format.
  - Added draft terms, privacy and third-party service disclosure documents.
  - Recorded the Colab Free policy restrictions as a feasibility gate; no prohibited remote-control or service-tunnel workaround will be implemented.

### Phase 3: Implementation

- **Status:** in_progress
- Actions taken:
  - Added a framework-independent Go module and domain state machine for runtime sessions.
  - Added evidence-backed model compatibility status and strict Ready predicate.
  - Added standard-library Python developer commands: doctor, test, build, fmt and arch-check.
  - Added repository-level agent instructions and .gitignore rules.
  - Ran Go tests and build successfully in an isolated local module matching the committed domain foundation.
  - Ran Python syntax validation, developer command help, doctor, test/build delegation and architecture-check logic on local test fixtures.
  - Preserved historical launcher/monitor scripts unchanged; did not run them against the user's live machine.
  - Added a bilingual responsive Material 3 web shell with Arabic RTL default, English LTR, light/dark theme controls, Dashboard/Models/Runtime/Settings pages, policy warning, and honest empty states.
  - Fixed language switching so the active page title is refreshed after changing locale.
  - Added a fail-closed runtime policy assessment port and connection use case; tests prove the connector is never called before a scoped, revisioned HTTPS approval.
  - Added versioned config defaults/validation, a ConfigRepository port and an application config service that rejects invalid settings before persistence. Gateway defaults to disabled and loopback-only.
  - Added the shared Planning with Files pointer at .planning/.active_plan for the single named plan.
  - Manually validated the plan structure: active pointer resolves to the plan directory, all three required files exist, five phases have exactly five valid status markers. The upstream check-complete.sh --validate was not run because the repository checkout and helper scripts are not mounted in the shell environment.
  - Verified the static UI with Node syntax check and Python HTML parsing; no duplicate IDs were found.
  - Verified the committed Go core against an exact local mirror (Git blob SHA match before the latest TDD hardening): go test ./..., go vet ./..., go test -race ./... and go build ./... passed. The latest security/compatibility hardening was also retested locally; exact SHA reconciliation is being refreshed.
  - Added initial SQLite schema with foreign keys, state/checksum constraints, redacted URI fields and cross-table model/artifact identity; six Python sqlite3/unittest cases passed against the exact committed SQL and test file.
  - Matched committed Go source blobs against the tested local mirror after applying gofmt formatting to the remaining files.
  - Verified Python syntax and developer command smoke tests. The architecture checker was exercised against a local fixture, not a full remote checkout.
  - Rendered the local UI mirror in Chromium via Playwright; tested Models navigation, Arabic/English direction and active page title, Settings navigation and dark theme. UI preview screenshots were captured locally.
  - Evaluated SQLite drivers: modernc.org/sqlite v1.62.0 is CGO-free/BSD-3-Clause but requires Go 1.26; mattn/go-sqlite3 is MIT but requires CGO/GCC; ncruces/go-sqlite3 is CGO-free but uses a WASM-derived SQLite VFS with additional memory and WAL/locking validation needs. Added a proposed ADR; no driver is pinned.
  - Set the repository toolchain target to Go 1.27.3, the latest stable patch shown by official Go release history on 2026-11-07. scripts/dev.py doctor now flags older local Go toolchains without installing or changing anything.
  - Added and retained the relevant ECC skills (Go patterns/testing, database migrations, security review, TDD workflow) under .agents/skills/ecc/ with the MIT license. Updated AGENTS.md to require them for matching work.
  - Used a real TDD RED/GREEN cycle for policy-source URL hardening, OS secret-reference validation and duplicate-port rejection. Added strict SHA-256 format validation to model readiness; malformed digests no longer qualify as ready.
  - Added the development-only down migration and test proving it removes the initial schema. Production migration policy remains forward-only.
  - Added a Linux/Windows GitHub Actions validation workflow using the go.mod toolchain target and expanded architecture checks. No Actions run/status is available yet on this feature branch.
  - Wails v3 PoC remains blocked in this container by local Go 1.23.2, missing Wails CLI/GTK4/WebKitGTK 6 dependencies and no Windows host. Official v3 docs identify v3.0.0 as stable and provide the vanilla-js template; no Wails dependency or CLI was installed.

### Errors and corrections in this session
  - Bulk skill vendoring exceeded the Code Mode tool-call limit; some initial files were created before interruption. Reconciled by fetching existing files and creating only the missing files in smaller batches.
  - GitHub create-file returned HTTP 422 because a path already existed from the partial batch; fetched its SHA and updated it instead of retrying create.
  - A policy-gate test initially failed after adding a required policy revision; updated the test fixture and reran all Go tests successfully.
  - Browser file:// and localhost navigation were blocked by the container policy; switched to Playwright page.set_content() with inline local assets.
  - A checksum-readiness test failed as expected in the TDD RED stage; implemented strict 64-character hexadecimal SHA-256 validation and the suite passed. An initial broad fixture replacement also changed the negative test vector; restored the malformed value and reran successfully.
  - First UI interaction test used an ambiguous accessible button name; switched to unique data-page selectors and reran successfully.
  - Manual Planning with Files validation found phase status markers had a leading list bullet; corrected them to the exact upstream `**Status:** value` format and reran structural validation successfully.

## Test Results

| Test | Input | Expected | Actual | Status |
|------|-------|----------|--------|--------|
| Repository starting state | README.md | Inspect default branch entry point | README contains only `# OhMyFreeAI` | PASS |
| GitHub code search | go.mod, package.json, .gitignore, PROJECT_SPEC.md | Detect existing project files | No results returned for searched paths | PARTIAL |
| Local git clone | HTTPS GitHub URL | Checkout repository and skill source | DNS resolution failed | BLOCKED |
| Planning skill source | GitHub fetch_file | Read authoritative SKILL.md | Fetched v3.28.1 | PASS |
| Planning initializer | Local execution | Initialize plan using upstream helper | Cannot run: upstream files not mounted and network unavailable | BLOCKED |

## Error Log

| Timestamp | Error | Attempt | Resolution |
|-----------|-------|---------|------------|
| 2026-11-07T15:52Z | Could not resolve github.com during git clone | 1 | Use GitHub connector; do not retry same clone |

## 5-Question Reboot Check

| Question | Answer |
|----------|--------|
| Where am I? | Phase 1 - Requirements & Discovery |
| Where am I going? | Phase 2 - Planning & Structure, then implementation and verification |
| What's the goal? | Establish a verified foundation for OhMyFreeAI without violating free-only constraints |
| What have I learned? | See findings.md |
| What have I done? | Branch created, repository inspected, planning artifacts established |


## Continuation session: 2026-11-07T16:52Z

### Resumed state
- Read the selected plan, findings, progress, and vendored Planning with Files v3.28.1 skill before implementation.
- Reconfirmed the active task is the existing `.planning/2026-11-07-ohmyfreeai-foundation` plan; no competing plan was created.
- Re-read the official Wails v3 installation, first-app, and build documentation using Exa. Research notes were recorded in findings.md before implementation.

### Work completed
- Added `.github/workflows/wails-poc.yml`.
- The workflow pins Wails CLI to v3.0.0 and Go to 1.27.3.
- Added native build jobs for Ubuntu 24.04 and Windows latest.
- Linux job installs GCC/build tools, GTK4 and WebKitGTK 6.0 development packages.
- Each runner generates the official `vanilla-js` starter using `wails3 init`, runs `wails3 build`, and checks for the expected native executable.
- Kept this as a generated temporary CI proof rather than committing generated starter files or adding Wails dependencies to the framework-independent core.
- Corrected Windows workflow steps to use PowerShell and a Windows-native temporary path rather than relying on Bash path conversion.
- Updated task_plan.md: the workflow addition is complete, but Wails selection remains gated on successful native CI results.

### Verification and limitations
- GitHub connector accepted the workflow and planning updates on `foundation/bootstrap`.
- No GitHub Actions result has yet been observed for the new Wails workflow. The workflow is **added, not yet passed**.
- The proof covers official template generation and native compilation only. It does not yet verify OhMyFreeAI Go service bindings, app lifecycle integration, packaging, or supported Linux distributions beyond Ubuntu 24.04.
- The local execution environment still cannot run the current Go 1.27.3 module or Wails v3 because its Go toolchain is 1.23.2 and it lacks the required Linux GUI development stack and a Windows host.
- No SQLite driver was selected; the cross-platform driver test matrix remains open.
- No changes were made to the user's historical Colab launcher/scripts or live runtime.

### Files changed in this continuation
- `.github/workflows/wails-poc.yml` (new)
- `.planning/2026-11-07-ohmyfreeai-foundation/findings.md` (Wails research)
- `.planning/2026-11-07-ohmyfreeai-foundation/task_plan.md` (next step and checklist)
- `.planning/2026-11-07-ohmyfreeai-foundation/progress.md` (this log)


## Continuation session: 2026-11-07T17:05Z

### Resumed state
- Re-read the active foundation task plan, progress log, findings, and current Wails workflow before continuing.
- Confirmed the active branch head was `caa004cd39d3f073e477da65926d63dbb088d5b4`; no newer branch commits existed at session start.
- Queried commit status and workflow-run lookup for the Wails workflow commits. Both returned no status checks and no workflow runs, so there is still no observed CI result.

### Work completed
- Reviewed official Wails v3 method-binding documentation and the vanilla project structure.
- Extended `.github/workflows/wails-poc.yml` on both Ubuntu 24.04 and Windows runners to run `wails3 generate bindings`, locate the generated GreetService JavaScript module, and assert that the `Greet` method is exposed before building the native executable.
- Kept the Wails template generated in a temporary runner directory; no generated demo app or Wails dependency was added to the production Go module.
- Updated the active task plan and findings with the exact scope and limitations of this binding smoke test.

### Verification and limitations
- GitHub accepted the workflow and planning updates on `foundation/bootstrap`.
- Workflow source was fetched back and inspected after the edit; both OS branches contain the binding-generation and assertion steps.
- Commit status lookup for the prior Wails commits returned empty status arrays; workflow-run lookup returned empty lists. The new commits have not yet been observed running. CI is **not verified green**.
- This test checks binding generation and exported method presence, then builds the native app. It does not launch a GUI or execute the JS-to-Go call in a live WebView.
- Wails remains an unselected candidate pending successful Ubuntu and Windows CI jobs.
- No production runtime, Colab launcher, historical Python scripts, or user live session was modified.

### Files changed in this continuation
- `.github/workflows/wails-poc.yml`
- `.planning/2026-11-07-ohmyfreeai-foundation/task_plan.md`
- `.planning/2026-11-07-ohmyfreeai-foundation/findings.md`
- `.planning/2026-11-07-ohmyfreeai-foundation/progress.md`


## Continuation session: 2026-11-07T17:05Z — SQLite evaluation

### Work completed
- Re-evaluated current official driver metadata and documentation for modernc, ncruces and mattn.
- Kept production dependency selection unapproved. modernc remains first evaluation candidate; ncruces is the comparison candidate; mattn remains a CGO fallback due GCC/Windows packaging overhead.
- Added nested module `tests/sqlite-driver-eval` with pinned modernc v1.62.0 and ncruces v0.37.0 test dependencies. These do not enter the production `go.mod`.
- Added shared `database/sql` smoke coverage for WAL, per-connection foreign keys, rollback, concurrent writes, integrity checks, WAL checkpoint, close/reopen and a checkpointed database-file copy.
- Added `.github/workflows/sqlite-driver-eval.yml` with a 2×2 matrix: modernc/ncruces × Ubuntu 24.04/Windows latest. It tidies/verifies the nested module and runs the same test against each driver.
- Updated ADR 0004 with the test harness scope and its explicit limitations; updated findings and the active plan.

### Verification and limitations
- Fetched the committed workflow, module and test source back from GitHub and reviewed the contents.
- The test harness and workflow are committed, but no matrix result has been observed yet. Do not mark either driver as passing or select it for production.
- The smoke test does not simulate abrupt process termination, does not run the production migration runner, and is not a performance benchmark.
- No local execution was claimed: the available shell toolchain is Go 1.23.2, while both selected candidate versions require Go 1.26.0. Native Linux/Windows GitHub Actions is the intended verification environment.
- The current Wails and SQLite workflows are separate, both use free GitHub-hosted runners, and neither changes the historical Colab runtime or adds paid infrastructure.

### Files changed in this continuation
- `.github/workflows/sqlite-driver-eval.yml` (new)
- `tests/sqlite-driver-eval/go.mod` (new isolated module)
- `tests/sqlite-driver-eval/driver_test.go` (new)
- `docs/adr/0004-sqlite-driver-evaluation.md`
- `.planning/2026-11-07-ohmyfreeai-foundation/findings.md`
- `.planning/2026-11-07-ohmyfreeai-foundation/task_plan.md`
- `.planning/2026-11-07-ohmyfreeai-foundation/progress.md`


## CI execution follow-up: 2026-11-07T17:11Z

- Inspected GitHub Actions run and job logs instead of treating queued/failed states as generic.
- The repository validation workflow completed successfully on Ubuntu 24.04: Go vet, Go tests, Go build, Python compile/tests, frontend JavaScript syntax and architecture-boundary check all passed on commit `47d0696`. Windows validation was still running at the last check.
- The first SQLite evaluation attempts failed before compiling the test: Go reported a missing `go.sum` entry for the imported driver. This was a workflow bootstrap defect, not evidence of a SQLite driver runtime failure. The evaluation workflow was updated to run `go mod tidy` and `go mod verify` before `go test`; the new run must confirm this fix.
- Wails native jobs reached toolchain/dependency installation on Ubuntu and Windows; native build and generated-binding steps had not yet completed at the last check.
- Multiple push commits had queued redundant copies of the same workflows. Added per-workflow/per-branch GitHub Actions concurrency groups with cancel-in-progress to ensure the newest branch revision supersedes stale queued runs.
- No driver or desktop framework has been selected. Continue from the latest run logs and preserve exact platform outcomes.


## Continuation session: 2026-11-07T17:15Z — native proofs completed

### Verified CI results
- Wails v3.0.0 native proof run `42394857136`: **success** on `ubuntu-24.04` and `windows-latest`. The official starter was generated, the GreetService binding was generated and checked, and each native executable was built and verified.
- Root validation run `42395074606`: **success** on Ubuntu and Windows. Go vet/tests/build, Python compile/tests, frontend JavaScript syntax and architecture-boundary checks all passed.
- SQLite driver matrix run `42395265071`: **success** in all four combinations (modernc/ncruces × Ubuntu/Windows). The expanded test applied the actual `migrations/001_initial.sql` in a transaction, checked schema creation and foreign-key rejection, then ran the WAL/concurrency/integrity/checkpoint/close-reopen/file-copy checks. Both drivers reported SQLite 3.54.1.

### Decisions
- Accepted Wails v3.0.0 as the desktop shell, with Ubuntu 24.04 as the only Linux baseline certified so far. The official template proof does not yet represent the complete OhMyFreeAI app or its runtime lifecycle.
- Selected modernc.org/sqlite v1.62.0 for the production SQLite adapter. Both modernc and ncruces passed the current matrix; modernc is preferred to avoid the WASM-translated engine/custom VFS and documented per-connection memory overhead in ncruces.
- Kept experimental driver dependencies in the nested evaluation module. The production module still has no SQLite driver import; add modernc when implementing the concrete persistence adapter.

### Remaining work
- Implement actual Wails service registration, frontend calls, lifecycle/shutdown and native adapters.
- Implement the production SQLite adapter and transactional migration runner using modernc.
- Add abrupt process crash/reopen, SQLite backup API/restore, long-duration stress and performance tests.
- Verify packaging, clean uninstall, license inventory and Linux distributions beyond Ubuntu 24.04.
- Run documentation/link validation and review the complete branch diff.
- No changes to the user's historical Colab launcher/scripts or live runtime were made.

### Files changed in this continuation
- `docs/adr/0001-desktop-shell-evaluation.md`
- `docs/adr/0004-sqlite-driver-evaluation.md`
- `docs/architecture/ADR_INDEX.md`
- `docs/architecture/IMPLEMENTATION_BLUEPRINT.md`
- `migrations/README.md`
- `.planning/2026-11-07-ohmyfreeai-foundation/task_plan.md`
- `.planning/2026-11-07-ohmyfreeai-foundation/findings.md`
- `.planning/2026-11-07-ohmyfreeai-foundation/progress.md`
- `tests/sqlite-driver-eval/driver_test.go` (import formatting)


## Continuation: 2026-11-07T17:05Z onward — production adapters and real desktop build

### Production SQLite implementation
- Added the production dependency modernc.org/sqlite v1.62.0 to the root Go module and committed the complete go.mod/go.sum graph generated by Go 1.27.3.
- Added internal/infrastructure/sqlite Store.Open with a private local database file, symlink/UNC rejection, WAL, per-connection foreign_keys=ON, busy_timeout=5000 and synchronous=FULL.
- Added an embedded forward-migration runner. It orders numbered migrations, skips development-only .down.sql files, rejects version gaps/untracked history, executes each migration transactionally and records its version only after the SQL succeeds.
- Added a validated ConfigRepository adapter. Missing settings load defaults; saves validate and atomically upsert JSON containing secret references only.
- Added tests for migration idempotence, transactional rollback, version gaps, config round-trip/rejection, file permissions, WAL and all four pooled connection PRAGMAs.
- Root CI run 42396865948 passed on Ubuntu 24.04 and Windows. Go vet, all Go tests and Go build passed on both OSes.
- SQLite evaluation run 42396720960 passed all four driver/OS combinations (modernc and ncruces on Ubuntu and Windows).

### Actual Wails application integration
- Added desktop-src/main.go.tmpl and app_service.go.tmpl. The templates are intentionally not compiled as part of the core Go module; scripts/build_desktop.py generates the official Wails v3.0.0 vanilla-js project in a temporary workspace, overlays the real app sources and current frontend, adds a local replace to the root module, generates bindings and builds the native executable.
- Registered AppService with GetSnapshot and SaveConfig. GetSnapshot reports database ready and runtime not_connected; it does not claim that a Gateway, Colab runtime or model is running.
- Connected the frontend to the generated GetSnapshot binding. The UI only changes its local status label after the native call resolves.
- Wails run 42396866046 passed on Ubuntu 24.04 and Windows. Each build generated 1 service, 2 methods and 7 models, built the Vite production bundle and verified dist/OhMyFreeAI or dist/OhMyFreeAI.exe.
- This is a successful native build and binding-generation proof, not yet an automated live WebView launch/invocation test.

### CI hardening and remaining observations
- Fixed the Wails init destination handling after the first application-specific build exposed that the CLI appends the project name to the -d parent directory.
- Fixed malformed Go test string/rune literals found by Go vet; the corrected Go suite passed on Ubuntu and Windows.
- Committed go.mod/go.sum from the actual Go 1.27.3 CI environment. CI now runs go mod tidy, go mod verify and git diff --exit-code on the lockfiles.
- Scoped push path filters to relevant source/workflow changes to avoid rerunning every matrix for documentation-only commits.
- Upgraded checkout/setup-go/setup-node/setup-python actions to Node 24-compatible releases after the runner warned that the previous major versions targeted deprecated Node 20. Verification runs 42397224021, 42397229073 and 42397232427 are in progress.
- The Windows Wails build emitted a non-fatal "-VThe system cannot find the file specified." line before its Taskfile steps. Binding generation, Vite build and executable creation all succeeded. The origin is not established; do not describe it as fixed.
- No historical Colab launcher, monitor or user-provided Python patch scripts were changed or executed.

### Remaining acceptance gates
- Add a real WebView-to-Go invocation test (not only generated binding and compile).
- Test abrupt process termination and clean database reopen.
- Implement and test SQLite backup API export/restore; do not use raw file copy as the production backup method.
- Add long-duration reader/writer stress and performance baselines.
- Review third-party licenses/SBOM, clean uninstall and data-preservation behavior.
- Validate additional Linux distributions; Ubuntu 24.04 is the only certified Linux build baseline.
- Audit the original Colab + llama.cpp + Gateway + OpenCode proof path and its current policy constraints.
- Review documentation, full branch diff and remaining acceptance gaps before any merge.


## Verification follow-up: 2026-11-07

- Core CI run 42397556757 completed successfully on Ubuntu 24.04 and Windows after the migration-history integrity test was added. Go vet, all Go tests, Go build, Python checks, frontend syntax and architecture checks passed.
- SQLite evaluation run 42397556823 completed successfully on all four matrix combinations: modernc/Ubuntu, modernc/Windows, ncruces/Ubuntu and ncruces/Windows.
- The production migration runner now rejects non-contiguous applied history (for example versions 1 and 3 with version 2 missing). The new corruption test passed in the full Ubuntu/Windows CI.
- Wails run 42397232427 completed successfully on Ubuntu 24.04 and Windows. Both generated one service, two methods and seven models, produced the Vite bundle and verified the native executable.
- Updated GitHub Actions pins to checkout v7.0.1, setup-go v7.0.0, setup-node v7.1.0 and setup-python v7.1.0. The Node 20 deprecation warning is gone in the subsequent runs.
- The Windows Wails build still prints a non-fatal '-VThe system cannot find the file specified.' line before Taskfile tasks. The executable and bindings succeed; origin remains unknown and is not marked resolved.
- Added docs/architecture/DESKTOP_BUILD.md and internal/infrastructure/sqlite/README.md; refreshed ADR 0001/0004, the ADR index, implementation blueprint and migrations README.


## Historical prototype audit: 2026-11-07

- Read-only inspection completed for both supplied monitor GUI archives and all three Python scripts. Nothing was executed; no launcher, runtime.env, Gateway, OpenCode process or Colab session was modified.
- Recorded SHA-256 hashes for both archives and the three scripts in docs/architecture/PROTOTYPE_AUDIT.md.
- Recovered the actual historical topology: OpenCode -> local Gateway 127.0.0.1:18081 -> Colab CLI session qwen-opencode -> remote llama-server at 127.0.0.1:8081. Telemetry uses /props, /proc, nvidia-smi and /proc/meminfo.
- Confirmed the core safety invariant: remote telemetry polling defers while a matching colab exec is active. The SSE patch's key correction is preserving blank lines, because they delimit SSE events.
- Found concrete risks: the Q8_0 patch overwrites runtime.env with only two keys; the SSE patch is not transactional and has no rollback after restart/test failure; PID-only process checks are identity-unsafe; monitor history/log reads are unbounded; the Python history file does not enforce private permissions; the Wails v2 monitor is Linux-only and uses a UI style that should not be reused.
- Documented all risks and unverified runtime assumptions in docs/architecture/PROTOTYPE_AUDIT.md. The old Wails v2 JSC_SIGNAL_FOR_GC=12 workaround is not carried into Wails v3 until tested against the actual GTK4/WebKitGTK 6 runtime.


## SQLite online backup and restore primitives: 2026-11-07

- Added Store.BackupTo using modernc.org/sqlite v1.62.0's NewBackup/Step/Finish API through database/sql.Conn.Raw. All driver-connection use remains inside the Raw callback.
- Snapshot publication uses a temporary file in the destination directory, mode 0600, integrity_check, foreign_key_check, contiguous migration-history validation, file sync and rename. Existing destinations are refused; BUSY/LOCKED are retried with context cancellation.
- Added RestoreDatabase using NewRestore/Step/Finish. It verifies the source before touching the target, restores into a temporary database, applies forward migrations, verifies integrity, checkpoints and closes it, rejects active sidecars/non-SQLite targets, and preserves an existing database as a rollback file.
- Added tests for snapshot round-trip, file permissions, no-overwrite, cancellation, corrupted migration history, schema newer than the application, restore rollback preservation, active-target refusal, non-SQLite-target refusal and corrupt-source preservation.
- Root CI run 42398884244 passed on Ubuntu 24.04 and Windows with the backup/restore tests. SQLite driver evaluation run 42398884345 passed all modernc/ncruces × Ubuntu/Windows combinations.
- The raw SQLite snapshot primitive is not yet the complete user-facing .omfaibackup format. Archive manifest/checksum, strict ZIP entry validation and archive restore orchestration remain open.


## CI failure investigation and Linux distribution expansion — 2026-11-07T18:36Z

### Repository and workflow state inspected
- Confirmed repository branch `foundation/bootstrap`; default `main` remains untouched.
- Latest root CI run `42399602407` passed on Ubuntu and Windows.
- Latest Wails run `42399602513`: Windows native build passed. Ubuntu native build and executable check passed; Ubuntu WebView smoke step failed. The earlier cancelled run is tracked separately and is not counted as a failure.
- Failure log: the old smoke step ran the app for 12 seconds, then `grep -q "native snapshot ready"` failed. Captured output contained only a Mesa/libEGL DRI3 warning. The log does not establish that DRI3 was fatal; the actual fact is that the expected service-call log line was not observed.

### Root cause correction implemented
- Added an opt-in marker to `AppService.GetSnapshot`, activated only by `OMFAI_WEBVIEW_SMOKE_MARKER`. This marker is written only after the frontend invokes the generated Go method.
- Added `scripts/smoke_desktop.py`: launches the actual desktop binary under Xvfb, requests software rendering, waits for the marker, terminates cleanly, then launches again and requires a second marker. This tests both JS-to-Go invocation and database reopen.
- Replaced the fragile stdout grep in `.github/workflows/wails-poc.yml` with the new smoke script.
- Extended `scripts/build_desktop.py` with validated optional build tags so older distributions can use Wails v3.0.x's GTK3 compatibility path without changing the default GTK4 build.

### Explicit Linux distribution matrix added
- Default GTK4/WebKitGTK 6.0: Ubuntu 24.04, Debian 13, Fedora 44, Arch Linux rolling, openSUSE Tumbleweed.
- Legacy GTK3/WebKit2GTK 4.1 for Wails v3.0.x: Ubuntu 22.04, Debian 12, Rocky Linux 9.
- Added `.github/workflows/linux-distro.yml` to install native distro packages, build the actual application, and run the WebView-to-Go/reopen smoke test inside each distro container on a free GitHub-hosted Ubuntu runner.
- Updated `docs/architecture/DESKTOP_BUILD.md` and the active task plan with the exact targets and the limitation that container userland coverage is not bare-metal certification.
- Official Wails docs identify GTK4/WebKitGTK 6.0 as the default on Ubuntu 24.04+, Debian 13+, Fedora 40+ and similar; the GTK3/WebKit2GTK 4.1 compatibility path is scheduled for removal in v3.1. Package names and version lanes were recorded in findings.md.

### Verification status
- Changes have been committed to `foundation/bootstrap`; the new CI runs are the next verification gate.
- Do not mark the WebView bridge smoke test or any distro as passing until the new Actions jobs complete and their logs are inspected.
- No direct writes to `main`; no paid service, Colab tunnel, live user runtime, or historical launcher was changed.


## Distro package bootstrap correction — 2026-11-07T18:41Z

- First Linux distro matrix run failed during package installation on Ubuntu 24.04, Debian 13, openSUSE Tumbleweed and Rocky Linux 9. None of these jobs reached checkout, Wails compilation or application smoke testing.
- Diagnosed exact causes from job logs: `libegl1-mesa` package name is unavailable on Ubuntu/Debian; openSUSE package is `webkitgtk4-devel`; Rocky Linux 9 only offers WebKit2GTK 4.0 and therefore cannot satisfy the required 4.1 compatibility lane.
- Corrected Ubuntu/Debian dependency to `libegl1`; corrected openSUSE package to `webkitgtk4-devel`; removed Rocky 9 from the legacy lane and replaced it with a Fedora 44 GTK3/WebKit2GTK 4.1 job (official Fedora package metadata confirms availability).
- The eight-job matrix is now five default GTK4 distributions plus Ubuntu 22.04, Debian 12 and Fedora 44 legacy GTK3.
- These initial failures are recorded as package-bootstrap errors, not app build failures. The corrected workflow must run to completion before marking any distro passed.


## Xvfb portability and diagnostic refinement — 2026-11-07T18:46Z

- Removed reliance on `xvfb-run`; the smoke helper now starts the `Xvfb` server directly, sets `DISPLAY`, and tears down both the app and virtual display after each of two launches.
- Corrected the Go and Python smoke marker literals to use an actual newline rather than a literal backslash-n sequence.
- The first direct-Xvfb native Ubuntu attempt exited with status 2 before the marker. The prior helper discarded the useful application log on this early-exit path. Updated it to flush and include the last 5,000 log characters in the failure report; the next Wails run is required to diagnose the exact startup error.
- openSUSE Tumbleweed failed because its minimal container did not include Bash, the default shell GitHub Actions requested. Added Bash to its dependency installation.
- All of these are test-harness/CI bootstrap corrections. Do not mark distro or WebView coverage passed until the latest matrix and native Wails jobs finish.


- Follow-up correction: the openSUSE dependency step itself cannot use Bash before Bash is installed. Changed that bootstrap step to POSIX `sh` and retained Bash installation for subsequent GitHub Actions steps. The next matrix run will validate this ordering.


## WebView bridge timeout follow-up — 2026-11-07T18:49Z

- The direct-Xvfb Wails build and executable verification passed on Ubuntu 24.04. The WebView smoke process stayed alive but did not produce the marker within 30 seconds; the test failed, so live JS-to-Go invocation remains unverified.
- Added distinct AppService marker values for success and config-load failure. The Python smoke helper now prints captured application output on timeout and process exit, and reports an explicit GetSnapshot failure separately.
- New Wails workflow run 42403217881 and Linux distro workflow run 42403217777 are pending/queued on the diagnostic commit. The root CI run is also queued.
- No conclusion about the cause is drawn until the new smoke output is available.


- openSUSE Tumbleweed package installation now succeeds and confirms Xvfb/WebKitGTK packages, but subsequent actions still failed to resolve Bash despite zypper reporting it installed. Removed the Bash dependency from workflow execution entirely: distro job run steps now use POSIX sh globally. The next matrix run will verify.


## Upstream WebKit/Wails signal research — 2026-11-07T18:53Z

- Confirmed from the exact Wails v3.0.0 source that it already includes the periodic SA_ONSTACK repair and the WebKit load-finished repair, and intentionally excludes SIGUSR1 from SA_ONSTACK.
- Confirmed from upstream WebKit source that JSC_SIGNAL_FOR_GC accepts a numeric signal ID (not the symbolic name SIGUSR2) and that JSC uses SIGUSR1 by default. Do not add a signal override until the fatal error's exact signal/headline is known; it may be harmful or irrelevant.
- The current smoke output includes a native Go stack and register dump, but the diagnostic excerpt starts too late to show the fatal headline. Next step is to emit both head and tail of the captured application log, then diagnose from the exact signal message.


## GTK stack selection and pkg-config fixes — 2026-11-07T18:57Z

- The 0ea0680e distro run showed that legacy GTK3 images were trying to compile the Wails CLI itself with the default GTK4 stack. Updated CLI installation to use `go install -tags gtk3` for legacy lanes, matching the application build tag.
- openSUSE's later CLI step explicitly overrode the job's POSIX shell with Bash. Removed the override; all distro job commands now use POSIX sh.
- Added explicit system pkg-config paths after setup-python and validation of the exact GTK/WebKit pkg-config modules before installing the Wails CLI. This catches missing headers at the dependency stage rather than later.
- Modern GTK4 jobs reached the actual executable smoke test but crashed with a native Go stack; the diagnostic headline is not yet available from that run. Latest smoke helper emits head+tail and will be inspected on the next run.


## Definitive WebKit sandbox diagnosis and CI-only workaround — 2026-11-07T19:00Z

- Retrieved the full fatal headline from Wails run 42403772084. Root cause is WebKit bubblewrap sandbox setup inside the GitHub-hosted runner: `bwrap: loopback: Failed RTM_NEWADDR: Operation not permitted`; its D-Bus credential/proxy connection then fails and GTK terminates with SIGTRAP/Go exit 2.
- The earlier signal-10 warning is not causal for this failure. No JSC signal override was added.
- Added `WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS=1` and `GTK_A11Y=none` strictly to the smoke subprocess environment. This bypasses a sandbox requirement unavailable in GitHub's restricted namespace, only for CI. It is not stored in product configuration, launch scripts or release builds.
- Updated desktop build documentation to explicitly state that this CI smoke validates WebView startup, JS-to-Go, and database reopen but does not validate the WebKit sandbox itself. A separate host-level security check remains required.


## Final distro bootstrap refinements — 2026-11-07T19:02Z

- Removed setup-python from distro containers; its Python 3.13 toolcache requires GLIBC_2.38 and fails on Debian 12's older glibc. All matrix jobs now install/use distro-provided Python 3, which is sufficient for the standard-library-only scripts.
- The Wails CLI itself now uses the GTK3 build tag on legacy lanes; previously only the app build was tagged.
- The workflow now verifies exact pkg-config modules and prepends distro-native pkg-config directories after setup actions.
- Replaced shell lookup with the explicit /bin/sh command template across the distro matrix to avoid openSUSE container PATH resolution problems.
- Next verification run is 42404224038 (Linux distribution compatibility); no matrix lane is marked green until build + two-launch WebView bridge smoke pass.


## Native Wails WebView smoke passed — 2026-11-07T19:04Z

- Wails run 42404058636 is fully green: Windows native build/executable verification passed; Ubuntu 24.04 build and executable verification passed.
- The Ubuntu WebView smoke produced `WEBVIEW_GO_BRIDGE_PASS attempt=1`, `WEBVIEW_GO_BRIDGE_PASS attempt=2`, and `WEBVIEW_REOPEN_PASS`. This verifies the actual generated JS-to-Go call and a second application startup against the persisted SQLite state.
- The smoke helper sets WebKit sandbox-disable and GTK accessibility variables only for CI subprocesses due to runner namespace restrictions. No production environment/config was changed.
- The Linux distro matrix remains the outstanding verification gate.


## Distro workflow scheduling correction — 2026-11-07T19:07Z

- Removed the branch-wide concurrency lock that kept newer matrix runs pending behind stale queued runs.
- Added `max-parallel: 4` to limit concurrent container jobs while allowing the current matrix to proceed.
- New run 42404614411 is active: Arch Linux, Fedora 44 GTK4 and Ubuntu 24.04 are in progress; remaining distro jobs are queued.


- The first matrix after the shell-template change failed immediately: GitHub Actions rejected the custom `/bin/sh -e {0}` value as an invalid shell option. Switched back to the supported built-in `sh` shell; setup-python has already been removed, so the prior PATH regression should no longer apply. No distro build ran in the rejected attempt.


- A follow-up run showed the dependency-install step still explicitly used the unsupported `/bin/sh` shell even though job defaults were corrected. Aligned the step to built-in `sh`; no distro setup/build was attempted in that run.


## First Linux distro end-to-end pass — 2026-11-07T19:11Z

- Debian 13 / GTK4-WebKitGTK 6.0 passed every stage in run 42404735351: dependencies, pkg-config, Wails CLI, real application build, WebView-to-Go twice, and database reopen.
- Output included WEBVIEW_GO_BRIDGE_PASS attempt=1, WEBVIEW_GO_BRIDGE_PASS attempt=2 and WEBVIEW_REOPEN_PASS.
- This confirms the new matrix and CI-only sandbox workaround work on Debian 13. Other seven distro lanes remain in progress/queued.


- Ubuntu 24.04 / GTK4-WebKitGTK 6.0 has also passed the distro-container job in run 42404735351, including both WebView-to-Go markers and DB reopen. Debian 13 and Ubuntu 24.04 are now green; six other distro lanes remain.


- openSUSE Tumbleweed now completes package installation, but subsequent steps cannot resolve the built-in `sh` command after the package transaction. The bootstrap itself runs under sh and installs Bash; changed post-bootstrap job steps to use Bash (setup-python is no longer present). openSUSE did not reach pkg-config/build/smoke in this attempt.


## 2026-11-07T19:13Z — CI incident review

- Resumed the existing named plan `.planning/2026-11-07-ohmyfreeai-foundation`; no competing plan was created.
- Read the vendored Planning with Files v3.28.1 skill and the existing plan/findings/progress ledgers before acting.
- Confirmed repository state: active work is on `foundation/bootstrap`, not `main`; main still contains only the initial commit, while the foundation branch contains the implemented architecture, app shell, skills, and CI workflows.
- Reviewed GitHub Actions logs. The newest failed distro run before the current head was a runner shell resolution failure in openSUSE after package bootstrap. Current head switches subsequent steps to Bash after installing Bash; run `42405097971` is the verification run and was not yet complete at review time.
- Historical causes/fixes are recorded in findings.md. No claim is made that the current eight-distro matrix is green.
- Next: monitor run `42405097971`; inspect and complete the existing backup archive/manifest/checksum implementation and recovery test gaps, then update this ledger with actual test evidence.


- Arch Linux rolling / GTK4-WebKitGTK 6.0 and Fedora 44 / GTK4-WebKitGTK 6.0 both passed full build and two-launch WebView-to-Go/database-reopen smoke in run 42404735351. Both emitted the two bridge pass markers and WEBVIEW_REOPEN_PASS.
- Four of eight lanes are now green: Ubuntu 24.04, Debian 13, Fedora 44 GTK4, and Arch rolling. Fedora 44 GTK3, Debian 12 GTK3, Ubuntu 22.04 GTK3 and openSUSE Tumbleweed remain outstanding.


- Fedora 44 legacy GTK3/WebKit2GTK 4.1 passed Wails CLI install, full application build, and the two-launch WebView bridge/SQLite reopen smoke in run 42404735351. Both bridge markers and WEBVIEW_REOPEN_PASS were observed.


## 2026-11-07T19:17Z — backup cancellation fix and CI result

- Fixed `extractArchive` to return an already-cancelled context immediately and preserve context cancellation during database-entry extraction rather than misclassifying it as an invalid archive.
- Added `TestExtractArchiveHonorsCancellation`.
- GitHub Actions run `42405330049` passed on Ubuntu and Windows, including vet, Go tests/build, Python tests, frontend syntax and architecture checks.
- Distro workflow run `42405097971` still exposes an openSUSE follow-up: Bash is reported installed but cannot be resolved by the next container step. The upcoming fix will inspect and export the verified executable directory via `GITHUB_PATH`.


## Six Linux distro lanes passed — 2026-11-07T19:16Z

- Debian 12 legacy GTK3/WebKit2GTK 4.1 passed the actual build, two WebView-to-Go calls and SQLite reopen. Both success markers and WEBVIEW_REOPEN_PASS were observed.
- Six lanes are green in run 42404735351: Ubuntu 24.04, Debian 13, Fedora 44 GTK4, Arch rolling, Fedora 44 legacy GTK3, Debian 12 legacy GTK3.
- Ubuntu 22.04 legacy GTK3 built but the WebView smoke timed out without app output; the generated Go bridge was not observed. openSUSE installed dependencies but failed at the post-install shell lookup. Both are being retested with the current workflow.


## 2026-11-07T19:19Z — Linux matrix follow-up

- Recorded the six confirmed distro passes from run `42404735351`.
- Ubuntu 22.04 legacy GTK3 is still failing only at the WebView-to-Go smoke: build succeeds, but the frontend never calls GetSnapshot within the timeout and the app log is empty. Root cause is not yet known; no unsupported diagnosis is recorded.
- openSUSE's current correction probes the Bash executable and adds its containing directory to `GITHUB_PATH`. Verification run `42405525978` is queued.
- One malformed workflow edit generated run `42405470161` with no jobs. Restored the file from the known-good `64bad8b` version and re-applied the change; current workflow content was checked for intact structure.
- Next: wait for the corrected distro run and inspect the Ubuntu 22.04 WebKit startup path; then continue with CLI dispatch and backup/recovery user-facing integration without enabling policy-blocked Colab remote inference.


## 2026-11-07T19:22Z — smoke marker contract fix

- Aligned the smoke-test consumer with AppService's new counted marker format.
- Added Python unit tests for marker parsing and malformed values.
- Corrected the regex newline escape after source inspection.
- Verification pending in runs `42405708011` (CI), `42405708063` (Wails) and `42405708191` (Linux distro matrix), all queued at the time of this update.


## 2026-11-07T19:23Z — bridge bootstrap sequencing

- Reviewed the latest bridge bootstrap changes: native snapshot call before UI import, a second call after import, and explicit bridge-ready event carrying the snapshot.
- Smoke helper now requires two successful native calls per process launch, matching the producer's count marker.
- Awaiting CI/Wails/distro runs `42405729818`, `42405729902`, `42405729874`; no result inferred yet.


## GTK3 bridge bootstrap and openSUSE PATH refinement — 2026-11-07T19:18Z

- Ubuntu 22.04 GTK3 built but its smoke timed out without marker; Debian 12 GTK3 built but its process exited before marker with no captured output. These remain runtime failures.
- Changed the generated bridge to call GetSnapshot before dynamically importing app.js, then call it again after the module import resolves. The frontend now consumes the returned snapshot from the bridge-ready event. Smoke requires marker count=2, so it verifies both the initial native call and successful UI-module import.
- This is an experiment to isolate a GTK3 module/bootstrap issue; no claim of success until CI.
- openSUSE's post-install shell still could not be resolved. Added /usr/bin and /bin to GITHUB_PATH after package installation and restored the built-in sh shell for later steps.


## Bridge refactor CI checks — 2026-11-07T19:24Z

- Root CI run 42405729818 passed on Ubuntu and Windows after the two-stage bridge bootstrap refactor.
- Wails Windows native build/executable verification passed in run 42405729902. Wails Ubuntu build and live smoke are still in progress.
- The updated eight-distro matrix run 42405827037 includes the openSUSE PATH fix and two-stage bridge code; its jobs are queued pending hosted runner capacity.


## Two-stage native bridge smoke passed — 2026-11-07T19:26Z

- Wails run 42405729902 is fully green on Windows and Ubuntu 24.04.
- The new two-stage bridge test passed twice on Ubuntu: WEBVIEW_GO_BRIDGE_PASS attempt=1 calls=2; WEBVIEW_GO_BRIDGE_PASS attempt=2 calls=2; WEBVIEW_REOPEN_PASS.
- This confirms the initial Go call, successful dynamic import of app.js, second Go call and SQLite reopen on GTK4/WebKitGTK 6.0.
- The same code is now being tested across all eight distro lanes in run 42405827037, including the openSUSE PATH correction.


- Arch Linux rolling passed the latest two-stage bridge retest in run 42405827037: both GetSnapshot calls succeeded before/after UI-module import and database reopen passed. Output: WEBVIEW_GO_BRIDGE_PASS attempt=1 calls=2, attempt=2 calls=2, WEBVIEW_REOPEN_PASS.


- The latest two-stage bridge matrix has revalidated Debian 13, Fedora 44 GTK4 and Arch rolling end-to-end. All three pass application build, two GetSnapshot calls around UI-module import, and SQLite reopen.


- Ubuntu 24.04 GTK4/WebKitGTK 6.0 passed the latest two-stage bridge run as well. Four current-run lanes are green: Ubuntu 24.04, Debian 13, Fedora 44 GTK4 and Arch rolling.


- openSUSE Tumbleweed in run 42405729874 (without the later GITHUB_PATH restoration, but with setup-python removed) successfully passed the dependency bootstrap, checkout, Go/Node setup and pkg-config configuration, and reached Wails CLI install under built-in sh. This narrows the earlier shell lookup issue to the removed setup-python/toolcache environment. The newer PATH restoration remains an extra guard, not a proven requirement. openSUSE's build and smoke are still pending.


- openSUSE Tumbleweed / GTK4-WebKitGTK 6.0 passed the full current two-stage bridge job in run 42405729874. Output confirmed two successful GetSnapshot calls and database reopen. This run used built-in sh successfully after setup-python had been removed; the additional PATH restoration was not needed for this pass.
- Five of eight current lanes are green: Ubuntu 24.04, Debian 13, Fedora 44 GTK4, Arch rolling and openSUSE. Three legacy GTK3 lanes remain in progress.


- Ubuntu 22.04 legacy GTK3/WebKit2GTK 4.1 and Fedora 44 legacy GTK3/WebKit2GTK 4.1 both passed full build and two-stage WebView bridge/SQLite reopen in run 42405729874. Each emitted two bridge pass markers and WEBVIEW_REOPEN_PASS.
- Seven of eight lanes are now green. Debian 12 legacy GTK3 is the only remaining job in this run.


## All eight Linux distro/stack lanes passed — 2026-11-07T19:36Z

- Linux distribution workflow run 42405729874 completed SUCCESS.
- All eight jobs passed: Ubuntu 24.04 GTK4, Debian 13 GTK4, Fedora 44 GTK4, Arch rolling GTK4, openSUSE Tumbleweed GTK4, Ubuntu 22.04 GTK3, Debian 12 GTK3, Fedora 44 legacy GTK3.
- Each job built the actual desktop app, launched it twice under Xvfb, recorded two successful GetSnapshot bridge calls, and verified SQLite reopen.
- Run URL: https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42405729874
- Production WebKit sandbox remains enabled. The CI-only sandbox-disable variable is restricted to smoke-test subprocesses because GitHub's runner lacks bubblewrap network-namespace permissions. Bare-metal sandbox verification remains open.


## Latest Linux matrix revision fully green — 2026-11-07T19:37Z

- Follow-up run 42405827037 (including the openSUSE PATH guard) also completed SUCCESS for all eight distro/stack jobs.
- Each job passed native package bootstrap, pkg-config verification, Wails CLI installation, actual app build, two-stage JS-to-Go smoke and SQLite reopen.
- This verifies the latest workflow revision. Run: https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42405827037


## Repository status and review hygiene — 2026-11-07T19:38Z

- main remains unchanged at a1b12fbe5dd854c4619463eb1c7cccc0bd951941. No PR or merge was created.
- foundation/bootstrap is 437 commits ahead / 0 behind main and changes 133 files. It was already 323 commits ahead at the start of this work; repeated per-file GitHub updates have made the history too granular for a review-ready PR.
- No history rewrite was performed. Before opening a PR, prepare a squashed snapshot branch or obtain approval to rewrite the working branch. Current branch and tested code remain intact.


## 2026-11-07T20:10Z — resumed foundation and started CLI TDD

- Read the vendored Planning with Files v3.28.1 skill and restored the selected plan/findings/progress ledgers.
- Verified latest Linux distro run 42405827037: all eight matrix jobs completed SUCCESS. Also verified the Wails and root CI workflows from the previous run were green.
- Corrected the stale Phase 4 checklist and summary: all eight Linux distro smoke lanes are complete; bare-metal WebKit sandbox remains open.
- Inspected current desktop entrypoint, AppService, application config service, backup archive manager, and developer CLI. The desktop entrypoint still unconditionally starts GUI; there is no application CLI dispatcher.
- Applied ECC TDD workflow: added dispatch tests before implementation (commit eb2f0b0). Expected contract covers GUI default, help/version/doctor, unknown commands and unexpected args. CI run 42409735886 is pending/in progress; wait for actual RED evidence before implementation.


## 2026-11-07T20:12Z — CLI dispatcher implementation

- Confirmed the intended TDD RED: test-first CI job 128515064532 failed because Parse was undefined in the new dispatch tests.
- Implemented the command parser and integrated it before GUI/database startup in the Wails entrypoint. Added read-only doctor diagnostics and updated CLI architecture documentation.
- Verification in progress: root CI run 42409816192 (Ubuntu passed; Windows running); Wails native build run 42409847974 (Ubuntu and Windows building).
- Next: inspect both platform results and any failures. Do not mark CLI complete on the parser unit tests alone; the generated Wails application must compile with the new entrypoint on both platforms.


## 2026-11-07T20:14Z — parser GREEN, native CLI smoke added

- Root CI run 42409816192 passed on Ubuntu and Windows, including Go vet/tests/build and existing validation suites. This validates the new parser and its test cases cross-platform.
- Added a standard-library native executable smoke helper covering --help, --version, doctor, unknown command exit code, and a check that doctor creates no application config directory.
- Added the CLI smoke to the native Wails Linux/Windows workflow and all eight Linux distro lanes; added py_compile coverage for the helper.
- Native Wails run 42409974970 and Linux distro run 42410021330 are pending. Do not close CLI task until those executable-level tests pass.


- Latest native Wails run 42409974970 has completed Linux dependency bootstrap and pinned Wails installation and is building the updated application; the CLI smoke steps are queued immediately after build.
- Linux distro run 42410021330 includes the new native CLI smoke in all eight lanes. Four GTK4 lanes are installing/building while legacy GTK3/openSUSE lanes are queued.
- Root CI run 42410030822 passed Ubuntu; Windows has completed its validation steps and is finalizing. This run also compiles scripts/smoke_cli.py.


- Corrected a workflow wiring omission found during source review: the Linux distro workflow path filter included smoke_cli.py, but the execution step had not been inserted because the actual WebView step label differed from the replacement target. The exact insertion is now present in commit 5ccfb6f and runs before WebView smoke in all eight lanes.


- Wails run 42409974970: Linux native app build passed and the native CLI smoke step passed all four commands plus the doctor read-only assertion. Linux WebView/SQLite smoke is still running; Windows app build is still running.


## 2026-11-07T20:17Z — native CLI GREEN on Linux and Windows

- Wails run 42409974970 completed SUCCESS on Ubuntu 24.04 and Windows.
- The native executable smoke passed --help, --version, doctor, unknown-command exit 2, and read-only doctor behavior on both platforms.
- Ubuntu also passed the live two-stage WebView-to-Go bridge and SQLite reopen smoke.
- Root CI run 42410030822 passed Ubuntu and Windows.
- All eight Linux distro lanes now have the CLI smoke wired in the workflow, but run 42410214638 is queued; do not mark that matrix complete yet.
