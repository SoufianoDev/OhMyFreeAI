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


- Diagnosed a failure in older distro run 42409848051: Ubuntu 22.04 GTK3 built but timed out waiting for GetSnapshot; app output was empty. This same lane passed in run 42405827037, so root cause is not established and the failure appears intermittent/environment-sensitive. No speculative code change made.
- Current CLI-enabled distro rerun 42410214638 is the authoritative retest; its Ubuntu 22.04 result is pending.


- CLI-enabled distro run 42410214638 has passed both native CLI smoke and two-launch WebView/SQLite smoke on Ubuntu 24.04, Debian 13 and Fedora 44 GTK4. Four other lanes are still running and Debian 12 is queued.


- Tooling-only authoring error: one Code Mode attempt to append the backup findings used an unescaped Markdown backtick inside a JavaScript template literal and raised a TypeError before any GitHub write. Retried with a line-array join; the intended findings were then saved successfully. No repository code was affected.


## 2026-11-07 — CLI matrix verified; backup inspect/export integrated

- Verified Linux distro CLI-enabled workflow `42410214638`: all eight jobs completed SUCCESS. Each lane built the actual Wails application, passed native CLI smoke (`--help`, `--version`, `doctor`, unknown command/exit 2 and read-only doctor), and passed two-launch WebView-to-Go plus SQLite reopen.
- Added test-first archive inspection tests in `internal/infrastructure/backup/archive_inspect_test.go` (commit `7de297d`). The expected RED was observed in run `42412080552`: `go vet` failed because `Manager.Inspect` did not exist yet.
- Implemented `Manager.Inspect` (commit `6ae3fc8`) using isolated temporary extraction and existing strict archive validation; added explicit temporary cleanup assertions (commit `2936f86`).
- Wired native Wails `ExportBackup` and `InspectBackup` actions using the pinned v3.0.0 file-dialog API; added bilingual Settings controls and generated bridge exposure for both methods. The build script now requires all four AppService bindings.
- A source inspection caught that Wails v3.0.0 `SaveFileDialogStruct` has no `SetTitle` method. Replaced the attempted fluent call with `SaveFileWithOptions(&SaveFileDialogOptions{Title, Filename, Filters})`. The OpenFile dialog does support `SetTitle`.
- Root CI `42412587313`: Ubuntu SUCCESS, Windows SUCCESS. Includes Go vet/tests/build, Python tests, frontend syntax and architecture checks.
- Wails native run `42412554344`: Windows SUCCESS and Ubuntu SUCCESS. Both actual application builds and native executable checks passed. Ubuntu also passed CLI smoke and live WebView bridge/SQLite reopen smoke. Exact output: `WEBVIEW_GO_BRIDGE_PASS attempt=1 calls=2`, `WEBVIEW_GO_BRIDGE_PASS attempt=2 calls=2`, `WEBVIEW_REOPEN_PASS`.
- Current backup-enabled distro matrix `42412554358` is in progress. At the last check four GTK4 jobs were building and four GTK3/openSUSE jobs were queued. Keep this open until all eight jobs finish.
- Planning docs updated after the verified results. Next: verify all eight lanes in `42412554358`; then add CLI export/inspect commands. Do not expose restore until the Store lifecycle can safely close, restore, reopen, swap repository references and roll back.

### Tooling/authoring notes

- One attempt to append findings used a JavaScript template literal containing unescaped Markdown backticks and failed before a GitHub write. Retried with a line array; findings were saved successfully. No repository code was affected.
- One initial Wails save-dialog implementation used a nonexistent fluent `SetTitle` method on SaveFileDialog. Verified against the pinned upstream source and corrected to `SaveFileWithOptions` before the successful Wails run.


## Backup-enabled Linux matrix: first four lanes green

- Follow-up check of run `42412554358`: Arch Linux rolling / GTK4, Debian 13 / GTK4, Fedora 44 / GTK4 and Ubuntu 24.04 / GTK4 all completed SUCCESS.
- All four passed actual Wails desktop build, native CLI smoke, two-launch WebView-to-Go smoke and SQLite reopen.
- Still queued at this check: openSUSE Tumbleweed / GTK4, Fedora 44 legacy GTK3, Debian 12 legacy GTK3 and Ubuntu 22.04 legacy GTK3. Do not mark the current eight-distro matrix complete yet.


## 2026-11-07T21:xxZ — CLI backup commands and stronger native smoke

- Re-read the active task plan/findings/progress before continuing. Confirmed the prior backup-enabled distro matrix `42412554358`: 7/8 passed; Ubuntu 22.04 GTK3 failed twice at the WebView bridge wait. Its native build and CLI smoke passed; app stdout/stderr was empty. The same lane passed in `42405827037`, so root cause remains unknown.
- Compared `scripts/smoke_desktop.py` at the earlier all-green commit `001533d8` with current: the test was unchanged and used a fixed 500 ms Xvfb delay, fixed display `:97`, and discarded Xvfb output. Improved the harness to wait up to 10 seconds for the X11 Unix socket to accept connections and capture Xvfb output. Do not claim this is a proven fix until the new Ubuntu 22.04 lane passes.
- TDD: added parser tests for `backup export` and `backup inspect`; root CI `42414904161` failed as expected on both platform jobs because `BackupExport` was undefined. Implemented the parser and typed malformed/unknown action behavior.
- Added `internal/application/cli/backup.go` with an infrastructure-independent `BackupOperations` port and `RunBackup` output/dispatch logic; added unit tests for success output, validated manifest metadata, failure propagation and invalid dependencies.
- Added `sqlite.OpenExisting` so backup export can open an existing DB without creating a missing file/directory or changing existing file permissions. Added tests for both guarantees.
- Wired CLI commands in `desktop-src/main.go.tmpl`: export checks the normal application DB path and refuses missing/symlink/special paths, does not run migrations, and uses the shared archive manager; inspect uses the strict read-only `Manager.Inspect` path. Both use a bounded two-minute context and exit 1 on operational failure.
- Expanded `scripts/smoke_cli.py` to test malformed/missing cases and the real success path with a temporary DB seeded from committed migrations: archive structure exactly two entries, manifest, successful inspect, overwrite rejection and preservation of the DB. It uses only Python standard library SQLite/ZIP support.
- Updated `docs/architecture/CLI_ARCHITECTURE.md` to document the implemented commands and current exit codes. Updated plan/findings with current blockers and verification requirements.
- Root CI `42415270003` passed Ubuntu on commit `640fc3f7`; Windows was still queued at the time. This was before the final positive end-to-end CLI fixture and Xvfb harness changes. Final revision CI/Wails/distro runs are pending.
- Current validation run IDs before the latest documentation-only commits: root `42415564618`, Wails `42415564460`, distro matrix `42415564608`. Recheck latest branch runs after documentation updates; GitHub concurrency cancels superseded runs, so use the newest commit's runs as authoritative.

### Decisions and constraints preserved

- CLI export never silently creates a first-run database and never runs migrations.
- CLI inspect is read-only and validates full archive content; restore remains blocked until the GUI Store lifecycle can close, restore, reopen, swap repository references and roll back safely.
- No Colab runtime activation, paid provider, branch merge or main-branch change.


## Verification update — commit 569c0c69

- Root CI run `42415564618` completed SUCCESS on both `ubuntu-latest` and `windows-latest`. This is the first full root CI result covering the CLI backup implementation, `sqlite.OpenExisting`, positive/negative backup CLI smoke helper compilation, and X-server readiness helper syntax/tests.
- Wails native workflow `42415564460` is still building the actual application on Windows and Ubuntu 24.04. Native executable CLI and WebView/SQLite smoke results are not yet confirmed for this revision.
- Eight-distro matrix `42415564608` is queued for the same revision. It includes the updated X11 readiness probe and full native backup CLI success-path smoke. Ubuntu 22.04 remains the key retest lane.
- Root CI is green; do not infer native Wails or distro matrix success from it.


## Native Wails proof GREEN — commit 569c0c69

- Wails workflow `42415564460` completed SUCCESS on both Windows and Ubuntu 24.04.
- Windows native executable smoke passed the full CLI backup workflow: no-database export rejection, missing archive inspect rejection, successful export from a migrated fixture, successful inspect, archive entry/manifest checks, overwrite rejection, and preservation of the source DB.
- Ubuntu native executable smoke passed the same CLI backup workflow and the two-launch WebView-to-Go bridge/SQLite reopen test. Exact markers: `WEBVIEW_GO_BRIDGE_PASS attempt=1 calls=2`, `WEBVIEW_GO_BRIDGE_PASS attempt=2 calls=2`, `WEBVIEW_REOPEN_PASS`.
- Observed Ubuntu WebView smoke timing: each launch took approximately 26 seconds to reach the second GetSnapshot call. The test's previous 30-second deadline had little margin and plausibly explains the Ubuntu 22.04 timeout; this is supported by timing evidence but not yet proven as the only cause.
- Updated the smoke harness to use a 60-second deadline, print a launch/progress marker, wait for the X11 socket to be ready, and retain Xvfb output. The 60-second change is in the newer commit `c4d7bf5c`; its full CI/Wails/distro workflows are queued/pending.
- Root CI `42415564618` is SUCCESS on Ubuntu and Windows for commit `569c0c69`.
- Linux distro matrix `42415564608` for commit `569c0c69` is running; the next authoritative matrix is the current-branch run after the 60-second harness adjustment.


## WebView timing hardening and restore lifecycle design

- Root CI run `42415960280` completed SUCCESS on Ubuntu and Windows for commit `c4d7bf5c`.
- The preceding Wails run `42415564460` completed SUCCESS on both platforms. Ubuntu's live bridge smoke took about 26 seconds per launch; both launches eventually produced the required two Go calls and SQLite reopen passed. This gives direct evidence that the previous 30-second per-launch deadline was too close to the observed startup latency.
- Raised the native WebView smoke deadline to 60 seconds, added explicit launch/first-call progress markers, retained the X11 socket readiness probe and captured Xvfb logs. The final 60-second version is in commit `c4d7bf5c`; current Wails retest `42415960222` is running and current eight-distro matrix `42415960198` is queued.
- Added `docs/architecture/RESTORE_LIFECYCLE.md` and linked it from ADR_INDEX and SECURITY_AND_RECOVERY. It specifies a stable Store controller, shared/exclusive lifecycle leases, durable phase journal, staged restore verification, atomic publish, rollback, startup crash recovery and failure-injection tests.
- Explicitly documented why restore remains disabled: current AppService and appconfig repository retain direct `*sqlite.Store` references; the infrastructure restore primitive has no durable transaction journal or application-level reopen/swap/rollback coordination.


## Continuation session: 2026-11-07T22:04Z — Store lifecycle controller (TDD)

### Work started
- Confirmed the Linux distribution matrix run `42416221475` completed successfully across all eight lanes, including the Ubuntu 22.04 legacy GTK3/WebKit2GTK 4.1 live WebView-to-Go and two-launch SQLite reopen checks.
- Re-read the active plan, restore lifecycle architecture, SQLite Store/config repository, backup manager, and Wails/CLI composition roots before changing implementation.
- Added `internal/infrastructure/sqlite/store_controller_test.go` before the controller implementation. Tests cover:
  - Exclusive lease waits for active shared leases to drain.
  - Waiting exclusive lease blocks new shared leases (writer preference).
  - Canceled exclusive wait releases its queue position and does not poison later operations.
  - Nil Store is rejected by the checked constructor.
  - Store replacement is only published under an exclusive lease; subsequent shared work sees the replacement.
- Added `internal/infrastructure/sqlite/store_controller.go` with context-aware shared/exclusive lease coordination, writer preference, idempotent lease release, guarded Store replacement, and stable ConfigRepository Load/Save delegation.
- Added a compile-time assertion that StoreController implements `ports.ConfigRepository`.
- Kept the implementation isolated so far: Wails AppService, CLI backup composition, and the backup Manager have not yet been rewired to the controller. Restore remains disabled.

### Verification state
- Root CI, Wails, SQLite-driver, and Linux distribution workflows were triggered by implementation commits. Several earlier attempts were superseded/canceled by subsequent commits before their Go test steps ran.
- The latest implementation revision still requires a complete, non-canceled root CI result on both Ubuntu and Windows. Do not mark the controller verified until `go vet ./...`, `go test ./...`, and `go build ./...` finish successfully.
- No local Go execution was claimed. The source was committed through GitHub; CI is the execution environment for the repository's Go 1.27.3 toolchain.
- No Colab session, model runtime, paid service, `main` branch, or user machine state was changed.

### Remaining controller work
- Wait for and inspect the full root CI result on the final controller commit.
- Fix any test, race, vet, or cross-platform failures before proceeding.
- Add controller-mediated backup export and rewire Wails/CLI composition so no long-lived service caches a raw Store.
- Add further tests for repository operations and operation-drain behavior before beginning the durable restore journal.

## Continuation: Store lifecycle integration hardening — 2026-11-07T22:35Z

- Confirmed the Wails integration workflow 42419830292 completed successfully on Windows and Ubuntu 24.04. Windows native build, CLI dispatch and executable verification passed. Ubuntu native build, CLI dispatch, executable verification, live WebView-to-Go invocation and SQLite reopen passed.
- Confirmed Linux distribution workflow 42419830259 completed successfully on all eight lanes, including Ubuntu 22.04 legacy GTK3/WebKit2GTK 4.1. This resolves the earlier in-progress Ubuntu 22.04 lane; the run is green.
- Re-read the current StoreController, Wails composition root, AppService and CLI backup composition. Configuration now uses the controller as its repository; GUI and CLI backup export hold shared leases for the complete Manager.Create call; Wails AppService retains the controller rather than a raw Store; app shutdown drains through StoreController.Close.
- Found a safety gap in WithExclusive: if a destructive callback closed the current Store and then failed, the controller returned the error while retaining the closed Store pointer. Changed the contract so the callback must return the safe Store to expose even on error. A verified rollback Store is published before returning the operation error; returning nil fails the controller closed.
- Added tests for publishing a verified rollback Store after an exclusive-operation error and for fail-closed behavior when no safe Store can be returned.
- Commits: e8806c338590cfd6904d170cd2a4433535508e35 (controller fix), f6cc74d7d1a82e8bc5c684f6d3ab5ea56609c72f (failure-path tests), 41b009e8a29e6c471b6d41bd0ca58e29d22459f4 (plan update).
- New root CI run 42421277886 and Linux matrix run 42421277879 were queued for the latest test commit f6cc74d. These results have not yet been observed; do not claim the new failure-path tests passed until root CI completes.
- Restore remains disabled. Remaining Store work: deterministic concurrent configuration/backup lease-drain tests, race/stress coverage, durable journal, staging verification, atomic publication, rollback and startup crash recovery.

## Continuation: lease cancellation and concurrency stress — 2026-11-07T23:49Z

- Added `TestStoreControllerCanceledCloseLeavesStoreAvailable`: a timed-out shutdown must release writer preference, keep the active Store published, allow new shared operations, and permit a later successful close.
- Added `TestStoreControllerConcurrentLeaseStress`: 12 concurrent shared workers and 4 exclusive workers repeatedly exercise lease acquisition/release while atomically checking for reader/writer overlap and stale Store exposure.
- Commits: `4f5b68667a9901d34207421f7e694ab8dac5d046` (canceled-close regression test), `2a9755622492543252435cf010644a81343854d0` (stress test), `d79c92a9d481e43a137566edc4ac8e27feadd97a` (stress counter correctness).
- Latest root CI run for the current branch tip: `42426993075`; SQLite driver matrix: `42426993082`; Wails: `42426993083`; Linux distribution matrix: `42426993088`. These were queued/pending at the time of this entry; no green result is claimed.
- Previous revision root CI `42426918154`: Ubuntu completed SUCCESS; Windows was canceled when newer commits superseded it. SQLite, Wails and distro jobs were also superseded/canceled or still running. Use the newest runs above as authoritative.
- No race-detector result has been observed yet. The stress test is committed, but it is not a substitute for `go test -race`.
- Next: inspect current CI results; add end-to-end tests proving configuration Save/Load and backup archive creation retain shared leases for their entire operations; then add the durable journal only after lifecycle behavior is fully verified. Restore remains disabled.

## Continuation: race CI and config repository verification — 2026-11-07T23:51Z

- Added a Linux-only `go test -race ./...` step to the main Go CI workflow (`.github/workflows/ci.yml`). This runs after ordinary Go tests and before build; Windows retains its native validation path.
- Added `TestStoreControllerConfigRepositoryRoundTrip`, verifying the controller's real `ports.ConfigRepository` Save/Load delegation persists and returns the validated configuration and closes through the controller.
- Re-inspected composition roots: Wails `SaveConfig` delegates to the configuration service backed by StoreController; Wails and CLI backup export wrap the full `Manager.Create` call in `WithShared`. Restore is not exposed in GUI/CLI.
- Commit `83a0f7c81e2a56facad2f902dc54a30506daa259` contains the config repository round-trip test. Race CI workflow change is commit `185549356aafab17b5a6644a08a2d0177733b2a4`.
- Current revision workflows: root CI `42427135468`, SQLite driver evaluation `42427135523`, Wails `42427135536`, Linux distribution matrix `42427135457`. At the time of this entry, jobs were pending/queued; no success is claimed.
- Older workflow runs for superseded commits were still queued/in progress and are not authoritative. Await the current revision results.
- Next: inspect CI/race output; add a deterministic backup lease-drain integration test if a stable test seam is available; then implement durable restore journal and startup recovery only after the controller suite is green. Restore remains disabled.

## Test correction: initialize schema in controller repository test — 2026-11-07T23:52Z

- Root CI run `42427135468` exposed a test-fixture error in `TestStoreControllerConfigRepositoryRoundTrip`: the test opened a fresh SQLite file but attempted `Save` before applying migrations, so SQLite correctly reported `no such table: app_settings`.
- This was a test setup defect, not a StoreController implementation failure. Updated the fixture to apply the committed production migrations before exercising controller `Save`/`Load`.
- Fix commit: `5405998c0dbee5b7c184925c49031bd7650b6738`. The corrected test still requires a fresh root CI run; no green claim yet.
- Race detector did not run in the failed Ubuntu job because ordinary `go test ./...` failed first. Await the new revision's Ubuntu run for both normal and race tests.

## Verified CI and first durable journal primitive — 2026-11-08T00:03Z

- All four workflows passed on commit `5405998c0dbee5b7c184925c49031bd7650b6738`: Root Go CI (Ubuntu + Windows), Linux `go test -race ./...`, SQLite driver matrix (4/4), Wails v3 (2/2), and Linux distribution matrix (8/8). Run IDs: `42427211242`, `42427211222`, `42427211260`, `42427211256`.
- Added `internal/infrastructure/sqlite/restore_journal.go` as a storage primitive, not yet integrated into restore execution or startup. It writes bounded strict JSON metadata with phase, transaction ID, target/staging/rollback paths and archive SHA-256/size; constrains transaction paths to the canonical state directory; rejects symlinked paths; writes a same-directory temp file, syncs it, atomically renames, and syncs the directory where supported (Windows skips directory sync through the standard library limitation).
- Added journal tests for round-trip, atomic phase update, mode 0600, generated timestamp, path confinement, invalid phase/digest, unknown fields, trailing data, and symlink rejection.
- Journal implementation commits: `2a54883d5c5b3b3639736d9a6c34f52be32c1a54`, `cacd03b3632d204fdd6a332c05467ab0e3c0c622`, `485b2c60cd110fbba45725de01b0915fb663f058`, `991ac99eecfa29039e6a400a7b8dfa3c0775b263`, `b62124c3f6d7699d0762b81f0c3b5affc65fb87f`.
- Current journal tests are awaiting CI on the newest revision; run `42428160345` started for the earlier journal-test commit `485b2c6...` and has not yet returned jobs. Subsequent commits superseded that run. Do not claim journal CI green yet.
- Architecture documentation now records the implemented journal primitive and explicitly distinguishes it from integrated durable restore. Restore remains disabled.
- Next: get clean CI on current tip; add recovery-state validation/decision logic that operates only on validated journal paths; then implement startup recovery before opening the canonical Store. Keep the application-level exclusive lease and rollback/verification protocol as the only route to restore.


## Continuation: cross-platform journal correction and pure recovery planner — 2026-11-08

### CI failure diagnosis
- Inspected the completed jobs from the journal validation runs instead of relying on their queued status.
- Root CI run `42428193863`: Ubuntu passed all steps including `go test -race ./...`; Windows failed `TestRestoreJournalRoundTripAndAtomicPhaseUpdate`.
- Exact Windows failure: journal read-back returned the long temporary directory spelling while the expected struct retained the 8.3 short-path spelling. This was a path canonicalization mismatch, not a database or journal-content corruption.
- Linux distro run `42428193878`: 7/8 jobs passed. Ubuntu 22.04 failed the WebView smoke because the helper used fixed X display `:97`; the desktop process then reported `cannot open display: :97`. Native build and CLI smoke had passed.
- Wails run `42428193865`: Windows passed; Ubuntu native build and CLI passed, but the live WebView smoke was cancelled. SQLite driver evaluation `42428193862` passed 4/4.

### Fixes committed
- `18260e34ce1bf9f3c961690008b30d8bb985aded`: normalize restore journal state and transaction paths to resolved absolute paths consistently, avoiding mixed 8.3/long-path spellings on Windows.
- `f79a554555bd4e8d514478024bef4df7020fe21c` and `6a1a19523c08f65e74daabe2dc37d0fc1587fefd`: compare the journal round-trip against canonicalized expected paths and populate the newly required previous-database metadata in the fixture.
- `cbd549dad5714603f33cf74ba0cc04a20b455743`: replace fixed Xvfb display `:97` with Xvfb `-displayfd` allocation, preventing stale display/socket collisions in containerized distro lanes.

### Recovery design implemented (planner only)
- Journal schema advanced to version 2 and now requires explicit `previous_database_existed` metadata. A pointer is used so a missing field is distinguishable from an explicit `false`; old/incomplete journal records fail validation rather than silently guessing.
- `85df4cd7c09601b9c4101b2bdc35f6900149230c`: add required prior-database metadata and schema version.
- `4620e08253897cc770dc68a13ccdbde61563d08f`: add pure `PlanRestoreRecovery` with no filesystem mutations.
- `e617a8a31510f3deb9a7e5cf7bddbc3498724dec`: add phase/file-state decision tests and fail-closed ambiguity tests.
- Planner policy: pre-publication phases preserve the original target; previous-preserved restores the rollback; published/reopening with a prior DB preserve the candidate and restore rollback; first-ever publication is retained for manual review; committed state requires target verification and retains rollback. Conflicting file/phase combinations return `ErrRestoreRecoveryAmbiguous`.
- This is not a recovery executor, is not wired into startup, and does not enable Restore.

### Current verification status
- The above fixes and planner commits are on `foundation/bootstrap`.
- A complete CI run for the newest commits has not yet been observed. The earlier Windows and Ubuntu 22.04 failures remain open until the corrected revision passes.
- No local Go tests are claimed. Do not treat source acceptance or commit creation as test success.
- Restore remains disabled. No changes were made to `main`, live Colab, paid services, or the user's local machine.

### Next
1. Obtain a fresh root CI result for Windows + Ubuntu and verify the Linux race step.
2. Obtain fresh Wails and all-eight-distro results; inspect any remaining WebView failures.
3. Review and test the version-2 journal schema and pure planner, then implement a conservative filesystem recovery executor with collision-safe candidate preservation and crash-injection tests.
4. Wire recovery before canonical Store open only after the executor is verified.


### Follow-up hardening
- `04214370dcefb4299a5a38f6bbb16ed851be5c52`: make the pure planner reject unsupported journal versions, invalid phases, empty transaction IDs and empty required paths before returning any action.
- `950bca04627d46d8103ddfbfe846732d8fe14e8f`: add fail-closed tests for unsupported versions and missing target paths.
- `947124d4e0abeabcb280f0c0ae38ec5ec79c530a`: update RESTORE_LIFECYCLE.md to distinguish the version-2 journal and pure planner from a filesystem executor or integrated restore.
- `e8769e665fb316714f89fcbbeba3b42e5c5ccf85`: update the active task plan and next-step ordering.
- No fresh workflow run/status for these commits is currently exposed by the available GitHub Actions lookup. The fixes are implemented but unverified; do not mark them passed.


### Final test coverage additions in this continuation
- `32f21f59481e5e2eee880a96cce942250c6f6b72`: added a journal round-trip test proving explicit `previous_database_existed=false` remains distinguishable from missing metadata.
- `e46da7305213bba156098987671336d37d6a4870`: removed an unused import from the dynamic Xvfb smoke helper.
- These source changes also require a fresh CI run. The available Actions connector only exposes pull-request workflow runs for commit lookup, and no push-run IDs/check results were returned for the new commits. Verification is therefore still pending.


## Continuation: restore journal rollback invariants — 2026-11-08T00:36Z

- Reviewed the durable journal and recovery planner against the required rollback protocol.
- Found a validation gap: the journal parser accepted previous_database_existed=true with no rollback path, and accepted false with an unexpected rollback path. The planner rejected some of these combinations later, but the durable journal boundary itself should reject inconsistent metadata before any executor can consume it.
- Updated canonicalizeRestoreJournal to require a rollback path if and only if a previous database existed. Path-collision checks now use the platform-aware samePath helper for target/staging/rollback comparisons.
- Updated the test fixture to include the required rollback path and adjusted the explicit-false case to omit it.
- Added write-time and read-time tests for both inconsistent combinations.
- Commits: 62068a8d7c17a8988d571aab37ee379e392f2401 (journal validation), 1548e9d62cd1a6e10a51337ed304f8c61940691f (tests), followed by this documentation correction.
- CI status for the latest commits is not available from the current GitHub status lookup; no test success is claimed.
- Restore remains disabled. Next: obtain full CI evidence, then implement the recovery executor only with collision-safe no-overwrite semantics and process-death tests.


## Continuation: 2026-11-08T01:04Z — Restore journal hardening CI closure

### CI verification

The latest code commit under test was `31b9a963c1403bee223f0d9da632f8e570d9ffa6` (non-regular restore transaction path rejection). All four workflows completed successfully:

- Root Go CI `42431721560`: Ubuntu and Windows passed.
- SQLite driver evaluation `42431721488`: modernc and ncruces on Ubuntu and Windows passed (4/4).
- Wails v3.0.0 platform proof `42431721597`: Ubuntu and Windows passed (2/2).
- Linux distribution compatibility `42431721639`, attempt 2: all eight jobs passed (8/8), including Ubuntu 22.04 GTK3/WebKit2GTK 4.1, Debian 12, Fedora GTK3, Ubuntu 24.04, openSUSE Tumbleweed, Debian 13, Fedora GTK4 and Arch Linux.

The first Ubuntu 22.04 attempt timed out waiting for the frontend-to-Go WebView call. The targeted rerun passed; its bridge smoke test succeeded twice and the WebView reopen check passed. No code change was made for this transient failure. This is a successful rerun, not evidence of a code fix.

### Documentation

- Updated `docs/architecture/RESTORE_LIFECYCLE.md` with exact workflow/run evidence and a clear boundary on what CI does and does not prove.
- Restore remains disabled. CI success for the foundation does not certify a recovery executor or the complete restore transaction.

### Next engineering step

Proceed with the collision-safe filesystem recovery executor only after designing its ownership and no-overwrite semantics. It must revalidate paths immediately before each mutation, preserve unexpected candidates at collision-free transaction-owned locations, sync directory metadata where supported, and fail closed on any race or contradictory state. Add process-death tests at every journal phase before startup recovery is wired into Store opening.


### Recovery executor design gate — 2026-11-08T01:06Z

- Added `docs/architecture/RESTORE_RECOVERY_EXECUTOR.md` before implementation.
- The design identifies an important gap in the current completed-transition journal: a process can die after a filesystem move but before the following journal phase is persisted. The current planner correctly fails closed on some such states, but cannot automatically reconcile them.
- The executor therefore requires explicit durable intent phases (`preserving-previous`, `publishing`, `restoring-previous`) and exhaustive truth tables for the filesystem states observable after each interrupted move.
- Prohibited a check-then-`os.Rename` implementation. The planned adapter must provide atomic no-replace moves (Linux `renameat2(RENAME_NOREPLACE)`, Windows destination-exists-safe move with write-through); unsupported filesystems fail closed.
- Candidate quarantine path and database content identities must be journaled before mutation. Unknown or partial staging artifacts are preserved rather than deleted without ownership proof.
- No runtime code changed in this design step. Restore remains disabled. Next implementation begins with journal schema/state-machine TDD and platform no-replace primitives, followed by executor crash-injection tests.


## Continuation: 2026-11-08T01:12Z — v3 journal contract

- Re-read repository instructions, active task plan, findings/progress, and the restore recovery design before continuing.
- Confirmed current branch head before edits: `7054668df048aa6ad931219f801c7a320e68aec4`; all writes remain on `foundation/bootstrap`.
- Expanded `docs/architecture/RESTORE_RECOVERY_EXECUTOR.md` with a concrete v3 journal contract: persisted quarantine path, paired candidate and previous-database SHA-256/size identities, incompatible-v2 policy, phase transitions, and validation/planner requirements.
- Documentation commit: `84aa731726ba9cff72cfafdbed72ec54a5cec22e`.
- No runtime code changed and no CI result is claimed for this documentation-only commit.
- Next: implement v3 journal fields and strict validation with RED/GREEN tests; update the pure planner to require identity verification inputs; then implement Linux/Windows atomic no-replace primitives. Restore stays disabled.

## Continuation: 2026-11-08T01:43Z — Ponytail Ultra project policy

- Read Exa Search skill guidance and researched the official Ponytail repository with Exa; fetched the upstream `SKILL.md`, compact `AGENTS.md`, and MIT license through the GitHub integration.
- Vendored the upstream skill and license at `.agents/skills/ponytail/`.
- Updated root `AGENTS.md`: Ponytail Ultra is mandatory for coding/debugging/refactoring/review/dependency work; requires scope challenge, caller tracing, minimal complete diffs, tests, and explicit risk reporting. Explicitly preserves security, accessibility, validation, data-loss protections and user constraints.
- Updated `.agents/skills/planning-with-files/SKILL.md` with project-specific Ultra rules: research goes in findings only, external content remains untrusted, active plan ownership is preserved, tests cannot be weakened, and progress must record exact evidence.
- Registered Ponytail in `.agents/skills/README.md`.
- CI recheck on commit `b554116f91a9978f576e686b2aad4ded8cca43d2`: root Go CI passed Ubuntu/Windows; Wails passed Ubuntu/Windows; SQLite matrix passed 4/4. Linux distro matrix had 6/8 passed, with Fedora 44 legacy GTK3 and Debian 12 legacy GTK3 still running at this check.
- No application runtime code changed in this policy adoption. Restore remains disabled.
- Next: finish the Linux matrix check, then continue the already-planned identity-aware pure recovery planner. Do not add a generic workflow engine or enable Restore prematurely.


## Continuation: Hush + Deslop Code adoption — 2026-11-08T02:00Z

- Confirmed through the connected GitHub account that `SoufianoDev/hush` exists, is public, and uses `dev` as its default branch. Read the canonical `skills/hush/SKILL.md` and MIT license.
- Researched two distinct Deslop variants with Exa: Stephen Turner's prose-focused skill and Brian Lovin's code-focused `skills/deslop/SKILL.md`. Selected the code-focused variant for this repository; the prose-only rules were not misapplied to code.
- Added `.agents/skills/hush/SKILL.md` and retained its MIT license at `.agents/skills/hush/LICENSE`. Added project-scoped `.agents/skills/deslop-code/SKILL.md` with explicit safety constraints.
- Updated root `AGENTS.md`, `.agents/skills/planning-with-files/SKILL.md`, and `.agents/skills/README.md` to require Hush on every code change and Deslop Code as a scoped post-implementation pass. Both policies preserve trust-boundary validation, Restore/SQLite invariants, error handling, tests, and user-owned pre-existing changes.
- Updated the active task plan and findings with the research decision and next Restore executor gate.
- Commits: `97c2fb56290791236bc56b74a6f2e578c201b65c` (Hush skill), `72b78f986fa3f086ea0b3f3ab3bed7323c52dbd6` (Hush MIT license), `5cd2f3026f1ea7c93c76c266751e8c9cc2c2f489` (Deslop Code), `dd1f4606a464fef13e550eb6af0f7a8f58b863c7` (AGENTS), `cea6e174a97308a5bd475da6f46caae44e844656` (Planning with Files), `3b7d0c6628bc40d0301a0cb510e5d742b3cacf64` (skills index), followed by plan/findings/progress documentation commits.
- CI recheck for planner commit `a4218970b8045b8cce03355373bfb6b2ac5586d3`: Root Go CI run `42435761597` passed (Ubuntu + Windows); Wails run `42435761521` passed (2/2); SQLite run `42435761588` passed (4/4). Linux distribution run `42435761590` passed 7/8; Ubuntu 22.04 GTK3 built the native executable and passed CLI smoke, but its WebView-to-Go `GetSnapshot` smoke timed out. The other seven distro lanes passed. This is a remaining CI failure, not a green matrix.
- No application runtime code changed in this skill adoption. No code tests were run specifically for the documentation/skill changes. Restore remains disabled; remote Colab inference remains policy-gated.


### Linux matrix retry — 2026-11-08T02:03Z

- Re-ran only the failed Ubuntu 22.04 job from workflow run `42435761590` (job `128597325747`), leaving the seven successful distro jobs untouched.
- At the latest check, Ubuntu 22.04 was still in progress; no retry result is claimed. The original attempt's failure was the frontend-to-Go `GetSnapshot` WebView smoke timeout after native build and CLI smoke passed.


## Continuation: atomic no-replace Restore move primitive — 2026-11-08

- Researched Linux renameat2(RENAME_NOREPLACE) and Windows MoveFileExW from the Linux man-pages and Microsoft Learn using Exa. Reused the existing golang.org/x/sys v0.49.0 dependency; no new dependency version.
- Test-first commit 61a19ea5932bbda9ef722e8789c5af42cabe4c45 produced expected RED in Root Go CI run 42438137306: Ubuntu job 128601331157 failed at go vet with undefined: moveRestoreFileNoReplace.
- Added common path/source validation, Linux Renameat2(RENAME_NOREPLACE), Windows MoveFileEx(MOVEFILE_WRITE_THROUGH) without replacement/copy flags, and an unsupported-platform implementation that fails closed.
- Added tests for successful move, preserving source and destination on collision, symbolic-link and non-regular source rejection, cross-directory rejection, and a concurrent two-source/one-destination collision.
- Hush/Deslop review: no explanatory comments were added; required build constraints remain. No unrelated files were changed.
- The first implementation CI run exposed only a go mod tidy grouping mismatch in go.mod; corrected the direct dependency grouping in commit d717eac7096d98c3b59b1b9ebd39aa1cff6e2d93. No dependency version or go.sum change.
- Verified on d717eac: Root Go CI run 42438368020 passed Ubuntu + Windows (including Linux go test -race ./...); SQLite driver matrix run 42438368022 passed 4/4. Wails run 42438368038 and eight-distro matrix run 42438368003 were still in progress at the time of writing.
- The preceding distro run 42437541325 finished with 7/8; Ubuntu 22.04 built and passed CLI smoke but timed out waiting for the WebView-to-Go call. A targeted retry of that lane is in progress. This remains an environment-sensitive failure until a successful retry is observed.
- Restore remains disabled. The primitive is not called by an executor. Not yet implemented: journaled move orchestration, content identity revalidation around each move, directory sync and restart reconciliation, crash-injection tests, startup recovery and shared application Restore service.


## CI closure update — 2026-11-08T02:27Z

- On code commit d717eac7096d98c3b59b1b9ebd39aa1cff6e2d93, Root Go CI run 42438368020 completed SUCCESS on Ubuntu and Windows. Ubuntu passed go vet, go test, go test -race, build, Python tests, frontend syntax and architecture checks. Windows passed the native equivalent except the Linux-only race step.
- SQLite driver evaluation run 42438368022 completed SUCCESS, 4/4 (modernc and ncruces on Ubuntu and Windows).
- Wails native platform run 42438368038 completed SUCCESS, 2/2. Ubuntu built the actual application, passed native CLI smoke, executable verification, WebView-to-Go bridge and database reopen. Windows built and verified the native executable.
- The previous eight-distro matrix run 42437541325 initially had Ubuntu 22.04 GTK3 fail at the WebView-to-Go smoke timeout. Re-running only that failed job succeeded; the latest attempt is 8/8 successful. No code change was made to claim a fix; this is a successful retry of an intermittent smoke failure.
- The new full distro matrix for d717eac, run 42438368003, was still running when this note was written. Keep its status separate from the successful prior-run retry.
- The no-replace primitive is verified by Linux and Windows unit tests, but is still not integrated into a durable executor. Restore remains disabled.


## Journal namespace barrier correction — 2026-11-08

- Research: Exa reviewed Linux fsync(2), Microsoft FlushFileBuffers, directory-handle guidance, MoveFileExW and file-caching docs, plus SQLite atomic-commit and corruption guidance.
- Found and corrected a real false-success path: the former syncRestoreDirectory returned nil on Windows without syncing anything.
- Changed journal publication to platform-specific helpers. Non-Windows performs same-directory os.Rename followed by directory Sync. Windows performs MoveFileExW with MOVEFILE_REPLACE_EXISTING | MOVEFILE_WRITE_THROUGH after the temporary file's Sync.
- Windows RemoveRestoreJournal now fails closed with errRestoreJournalRemovalUnsupported and leaves the journal untouched. This is deliberate; no production caller currently depends on deletion.
- Added common publication replacement test and Windows deletion fail-closed test. Existing round-trip test now expects Windows removal to be unsupported and confirms the journal remains readable.
- TDD caveat: test commits f876c42 and db2eddd were created first, but GitHub Actions cancelled their attempts before tests ran due subsequent branch updates. Therefore no observed RED result is claimed. Do not present this as a completed RED/GREEN cycle.
- Code commits: a51e9589, 45c5329e, cbccd640, a68bba15, 1cfd0805, 673782ca, 48bc9401.
- Latest CI for code revision 48bc9401 has not completed: Ubuntu root job queued, Windows root job cancelled; Wails and distro jobs queued. Pending is not pass.
- Remaining limitation: Windows MoveFileExW write-through is used as the documented platform barrier for publication, but its explicit flag guarantee is not a complete proof of power-loss recovery for every same-volume filesystem/storage stack. Windows Restore remains disabled until the full NTFS executor/restart matrix is tested; other filesystem classes remain unqualified.


## CI closure — platform-specific journal namespace barriers — 2026-11-08

- Code commit 6b582b3f1ee5537fb4f5c15c209e8fe02a75bc68 passed Root Go CI run 42439943545: Ubuntu and Windows both SUCCESS.
- Ubuntu: go vet, go test, go test -race, build, Python compile/tests, frontend JavaScript syntax and architecture boundary checks all passed.
- Windows: go vet, go test, build, Python compile/tests, frontend JavaScript syntax and architecture boundary checks all passed; Linux-only race step skipped.
- Current eight-distro matrix run 42439943571 remained queued. This is pending, not a pass.
- The new tests verify journal publication replacement and Windows fail-closed removal. Test-only workflows were cancelled before execution, so no RED evidence is claimed.
- Restore remains disabled. Windows publication uses MoveFileExW write-through; full Windows power-loss guarantees and non-NTFS support remain unqualified. Windows journal deletion remains unsupported and preserves the journal.


## StoreController integration gate and cross-platform move coverage — 2026-11-08

### Required starting run

- Requested run 42419106009 for commit f5396ee completed successfully on Ubuntu and Windows. This was the required baseline check, not the latest branch validation.

### StoreController integration review

- Reviewed `internal/infrastructure/sqlite/store_controller.go` and its tests. The controller uses shared/exclusive leases, writer preference, context cancellation, idempotent releases, and fail-closed handling when an exclusive callback panics or returns no safe Store.
- The Wails composition root constructs the controller after startup migration, uses it as the configuration repository, passes it to AppService, and closes through the controller after the app exits.
- AppService exposes backup export/inspect, dashboard snapshot and configuration save; it retains no raw Store. GUI and CLI export call `CreateWithStoreController`, which holds a shared lease through archive publication. CLI shutdown also closes through the controller.
- Config Load/Save drain tests and the backup-export drain test hold the lease while the underlying SQLite operation is blocked. The controller suite also covers close draining, closed-controller rejection, replacement, cancellation, panic and concurrent lease stress.
- No Restore method is exposed by AppService, the composition root does not run startup recovery, and the frontend has no Restore action. Existing low-level journal v3, planner and no-replace primitives remain disconnected; Restore stays disabled.

### CI closure before the next phase

- Branch commit `baca870a72dccf067fe3ec032e47425351ec71af`: Root CI run 42473729669 passed Ubuntu and Windows; Linux `go test -race ./...` passed. Wails run 42473729641 passed both platforms and logged two successful WebView-to-Go calls on each Linux launch plus `WEBVIEW_REOPEN_PASS`. SQLite driver evaluation 42473729677 passed 4/4.
- Linux distribution run 42473729674 initially failed only Debian 12 GTK3/WebKit2GTK smoke with `Gtk-WARNING: cannot open display: :0`. The targeted retry (attempt 2, job 128713500788) passed; its bridge smoke succeeded twice and the reopen check passed. The completed run is 8/8. No code change was made for that transient failure.
- Test-coverage commit `3a814ff960e57da2e60bb7fc7dd697461af8bda1` changes `restore_move_test.go` to `linux || windows` and moves the Linux-only symlinked-parent test into `restore_move_linux_test.go`. This is test coverage only; no production move behavior changed.
- On `3a814ff`, Root CI run 42475338115 passed Ubuntu and Windows; the Linux race step passed and the Windows race step was skipped by workflow design. SQLite driver evaluation 42475338112 passed 4/4. Wails run 42475338129 passed 2/2; Linux logs show both WebView bridge launches succeeded and database reopen passed, while Windows executable verification passed. Linux distribution run 42475338092 passed 8/8.

### Restore planner gap found during executor review

- `PlanRestoreRecovery` returns `PreserveStaging=true` from the `previous-preserved` phase, because the staged candidate must be retained before restoring the old database.
- The `restoring-previous` intent phase currently rejects every non-missing staging file and only accepts the candidate at target or quarantine. That leaves no recoverable state for a crash after the journal records the rollback intent but before it has moved the staged candidate to quarantine.
- The next change is to add a focused truth-table test for candidate-at-staging under `restoring-previous`, then extend the planner without weakening its ambiguity checks. The filesystem executor remains unimplemented; no Restore UI/startup recovery was added.


## Windows handle-bound no-replace rename — 2026-11-08

- Added Windows source-handle identity and destination-collision tests first.
- Observed the expected RED on Windows Root CI: run 42496137044, job 128778788124, go vet failed because renameRestoreHandleNoReplace was not yet implemented.
- Replaced the Windows MoveFileExW path-based primitive with SetFileInformationByHandle(FileRenameInfoEx), using a retained source file handle and destination directory handle; no replacement flag or fallback is used.
- Implementation commit: 18aa3a3455857d5a68cd80afeb92a674ca6d1eca.
- Current verification runs: Root CI 42496232166; Wails 42496232229; SQLite driver matrix 42496232217; Linux distribution matrix 42496232194. All were pending at this update. Do not claim green until checked.
- Known boundaries: parent-component reparse protection and source identity hashing/revalidation belong to the future executor; no actual-volume capability probe is implemented. Restore remains disabled.


## Correction — Windows rooted rename requires the native NT API — 2026-11-08

- The first handle-bound implementation used SetFileInformationByHandle(FileRenameInfoEx=22) with a non-null RootDirectory. Windows Root CI run 42496232166 passed Ubuntu but failed Windows: the regular move, collision and concurrent-collision tests all returned ERROR_INVALID_PARAMETER ("The parameter is incorrect").
- This was not a buffer-size or compilation failure. Microsoft's Win32 SetFileInformationByHandle documentation's supported-class table lists FileRenameInfo (3), not FileRenameInfoEx. A source-handle rename using a rooted relative destination is supported at the native NtSetInformationFile layer.
- Exa research verified Microsoft's FILE_INFORMATION_CLASS table: FileRenameInformationEx is native class 65, available starting with Windows 10 version 1709. Microsoft's NtSetInformationFile documentation describes the native operation and its information-class buffer. A practitioner source-code case also documents that passing non-null RootDirectory through SetFileInformationByHandle is rejected with ERROR_INVALID_PARAMETER.
- Corrected the adapter to call x/sys/windows.NtSetInformationFile with native class 65, the existing source handle, retained destination directory handle, and the MS-FSCC 24-byte-minimum FILE_RENAME_INFORMATION_EX payload. No replacement flag is set. STATUS_OBJECT_NAME_COLLISION maps to os.ErrExist.
- Corrected implementation commit: d6f6bb54c9f72cfcd118087e4468e7d425e83251.
- Current validation runs for this corrected commit: Root CI 42496496153, Wails 42496496135, SQLite driver evaluation 42496496199, and Linux distro matrix 42496496136. They were pending at the time of this note; no GREEN claim.
- No legacy FileRenameInformation fallback is added. NTFS is the initial target; filesystems that reject class 65 remain unsupported. The source path is still opened by pathname and parent-component reparse traversal is not pinned; the future executor must enforce a private canonical state directory. Restore remains disabled.


## CI closure — Windows handle-bound move — 2026-11-08

- Root CI 42496659593: Ubuntu SUCCESS (including go test -race); Windows SUCCESS (go vet/tests/build/Python/frontend/architecture).
- SQLite driver evaluation 42496659645: SUCCESS 4/4 across modernc/ncruces on Ubuntu/Windows.
- Wails platform proof 42496659732: SUCCESS 2/2. Ubuntu native app, CLI, WebView-to-Go bridge and DB reopen passed; Windows native app/executable passed.
- Linux distro matrix 42496659764: 4/8 passed (Arch, Fedora 44 GTK4, Ubuntu 24.04 GTK4, Debian 13 GTK4); openSUSE and three GTK3 legacy lanes still running. No 8/8 claim.
- Next security task: open the source leaf relative to the pinned Windows directory handle using NtCreateFile; preserve fail-closed behavior. Restore remains disabled.


## Final validation closure — 2026-11-08

- Root CI 42496659593 passed Ubuntu and Windows; Ubuntu race detector passed.
- Wails v3 platform proof 42496659732 passed 2/2, including the live Linux WebView-to-Go bridge and database reopen.
- SQLite driver evaluation 42496659645 passed 4/4.
- Linux distribution compatibility 42496659764 passed 8/8 across modern GTK4/WebKit6 and legacy GTK3/WebKit2GTK 4.1 lanes.
- Current Restore code remains low-level only. No durable executor, startup recovery, app-level Restore service, or Restore UI is enabled.


## Current code revision verification — c89d532a — 2026-11-08

- Root CI 42497620421 passed Ubuntu and Windows; Ubuntu race detector passed. The new Windows relative-source-open regression test ran as part of the Windows Go package suite.
- SQLite driver evaluation 42497620420 passed 4/4.
- Wails platform proof 42497620429 passed 2/2, including Linux WebView-to-Go/database reopen and Windows native executable verification.
- Linux distribution matrix 42497620422 was in progress with 4/8 jobs passed at this update. Do not claim full matrix green for c89d532a yet.
- Code commit: c89d532a3ba04d43c7d2ccbed1675319c6f90ee8. Current branch also includes docs-only follow-up commits.
- Remaining Windows path boundary: destination directory path is opened by CreateFile; OPEN_REPARSE_POINT protects its final component only. Next implement component-by-component directory opening from a trusted volume root and verify private ACL/permissions before the executor.


## Continuation — historical prototype provenance and Colab MCP policy review — 2026-11-08

- Reconfirmed that the uploaded Python patch/monitor scripts and archived Wails monitor applications are historical proof-of-concept artifacts, not production source files in OhMyFreeAI.
- Recorded the user's confirmation that the initial experiment successfully started the model on free hosting, connected it to OpenCode, and completed a real conversation from inside OpenCode after approximately four hours of work. This is historical technical evidence only; it does not establish policy approval or production readiness. Additional scripts were not uploaded and remain unknown.
- Reviewed Google's official Colab FAQ and official Colab MCP `create_runtime` reference. The MCP API documents runtime creation and authenticated Jupyter access, but does not expressly authorize a general-purpose inference server or OpenCode-consumed inference endpoint on a free managed runtime.
- Updated `docs/legal/COLAB_RISK_DISCLOSURE.md`, `docs/adr/0002-colab-free-policy-gate.md`, and `docs/architecture/PROTOTYPE_AUDIT.md`; recorded research in `findings.md` and added the open policy reassessment item to `task_plan.md`.
- Documentation commit: `45b6a9b998eef1e5657b359f971e99dc8313001b` on `foundation/bootstrap`.
- No application code, live runtime, launcher, or user process was modified. No automated tests were run because this change is documentation-only. The Colab adapter remains policy-gated and disabled.
- Exa's follow-up search endpoint returned HTTP 500 twice; conclusions above rely on the successfully retrieved official Google documentation, not third-party tutorials.


## StoreController gate verification and Windows directory pinning — 2026-11-09

### Requested baseline

- Workflow run [42419106009](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42419106009), commit `f5396ee592b9b9756779f67aeabd6d7167520c40`, completed successfully on `foundation/bootstrap`.
- Both jobs passed: Ubuntu `128543483467` and Windows `128543483323`. No failed steps.

### StoreController integration review

- The integration was already present; I did not duplicate it. The Wails composition root opens and migrates the initial Store before constructing the controller, passes the controller to the configuration service and AppService, and closes through the controller after the app exits.
- AppService retains `*sqlite.StoreController`, not a raw `*sqlite.Store`. Config Load/Save use the controller's repository implementation. GUI and CLI backup export call `CreateWithStoreController`, which holds the shared lease until archive creation and publication finish.
- Controller tests cover shared/exclusive ordering, writer preference, cancellation, panic/failure fail-closed behavior, Store replacement, shutdown drain, and rejection after close. Backup tests cover an export holding its lease while shutdown waits. The composition root's native Wails smoke also exercises the live service bridge and database reopen.
- The latest pre-path-hardening integration revision `adde70da56b7af4e71c68796a2e20dbe017d72d6` passed Root CI [42525602855](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42525602855), Wails [42525602932](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42525602932), and Linux distro matrix [42525602852](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42525602852), including all eight distro lanes.

### Windows restore-directory path hardening

- RED checkpoint: `6a93f2e0d76aac977a15e736052e9d2a23e6ed49`. Root CI [42599254395](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42599254395) passed Ubuntu and failed Windows `go vet` because the new tests referenced the not-yet-implemented `openRestoreDirectoryNoReparse`; this was the intended compile-time RED.
- Implementation: `1c03e8ff76933734c5d08620ebb93e2b376c278c`. The Windows no-replace move now opens the drive root and each directory component in sequence, opening children relative to the retained parent handle with `NtCreateFile`, `FILE_DIRECTORY_FILE`, and `FILE_OPEN_REPARSE_POINT`. Each opened component is checked for directory, reparse-point, and device attributes. UNC/device namespace paths are rejected.
- The first Windows GREEN attempt exposed a test-harness issue, not a production failure: the test retained an open handle to the nested child while renaming its parent, and Windows returned `Access is denied`. Test correction: `b9f1d579a9587c1df451ca3bed5e8feece045589` captures the child identity and closes that child handle before the rename while retaining only the parent handle under test.
- Final verification on `b9f1d579a9587c1df451ca3bed5e8feece045589`:
  - Root CI [42599641312](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42599641312): Ubuntu and Windows passed; Linux race detector passed; vet, Go tests/build, Python tests, frontend syntax and architecture checks passed.
  - Wails [42599641344](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42599641344): 2/2 passed, including Linux WebView bridge/database reopen and Windows native executable/CLI checks.
  - SQLite driver evaluation [42599641296](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42599641296): 4/4 passed.
  - Linux distro matrix [42599641297](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42599641297): 8/8 passed, including Ubuntu 22.04 legacy GTK3/WebKit2GTK 4.1.
- No comments were added or removed. The new symlink-component test skips when the Windows runner cannot create a directory symlink; the non-verbose CI output does not establish whether that branch skipped. Coverage percentage was not measured. No standalone formatter invocation was performed in this connector-only edit session; Go vet/tests/build passed.

### Remaining boundary

- This change only hardens directory traversal for the low-level Windows move primitive. It does not verify directory owner/DACL privacy, distinguish a local fixed disk from a mapped network drive, or qualify the actual filesystem's support for the no-replace operation.
- No filesystem recovery executor, startup recovery, app-level Restore service, CLI Restore command, or Restore UI was added. Restore remains disabled.
- Next: verify the canonical state directory's private owner/DACL and qualify the volume/filesystem before implementing the collision-safe recovery executor. Then add publication/reopen/rollback failure injection and process-death recovery tests before any app-level Restore wiring.


## Continuation — Windows Restore state-directory security gate — 2026-11-09

- Confirmed the requested baseline workflow [42419106009](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42419106009) is completed successfully on commit `f5396ee592b9b9756779f67aeabd6d7167520c40`. Both Ubuntu and Windows jobs passed Go vet/tests/build, Python compile/tests, frontend syntax and architecture-boundary checks.
- Confirmed branch `foundation/bootstrap` and the last previously recorded planning commit `d61baffee574816bbfad3f31285d935f7c285ad1`. No writes were made to `main`.
- Added Windows-only tests for volume qualification (fixed local NTFS with persistent ACLs only), broad-writable ACE detection, private directory owner/DACL verification, and actual handle-based volume inspection.
- Added a Windows security adapter that reads owner/DACL from the opened directory handle via GetSecurityInfo, requires the current process user to own the canonical state directory, rejects NULL DACLs, rejects broad write/delete/metadata rights for Everyone, Authenticated Users, BUILTIN Users and other broad well-known principals, and fails closed on ACE layouts it does not parse.
- Added volume qualification using GetVolumeInformationByHandleW, GetDriveTypeW, and FileRemoteProtocolInfo. The initial qualification policy is deliberately narrow: fixed local NTFS with persistent ACL support; remote, mapped, removable, unknown and unqualified filesystems are rejected. No ReFS/SMB/removable provider support is claimed.
- Updated Windows directory traversal to retain READ_CONTROL on each opened component, enforce the final state-directory owner/DACL gate, enforce the volume gate, and reject a volume root as the state directory.
- Relevant commits: test setup `fb747cff`, import correction `df441e69`, security/volume adapter `ff9b643e`, traversal integration `c551179b`, volume-root rejection `93bc1236`, READ_CONTROL retention `2e6ebe5d`, metadata-write ACL hardening `02504155`, additional test coverage `971a3294`.
- CI status at 2026-11-09T15:20Z: Root CI [42642834055](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42642834055) is queued; both Ubuntu and Windows jobs are queued. Wails [42642834036](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42642834036) is in progress on both platforms; Linux desktop dependencies passed and both jobs are installing the pinned Wails CLI. No conclusion is available yet. Combined commit-status arrays remain empty; the workflow run/job endpoints provide the current state.
- The new tests have not yet produced an observed CI RED/GREEN cycle. No RED evidence is claimed.
- Restore remains disabled. No filesystem executor, startup recovery, app-level Restore service, CLI Restore command, or UI was added.
- Next: obtain and inspect the Root CI run for the latest code commit, fix any Windows API/test failures, then verify Ubuntu/Windows and relevant Wails/SQLite/Linux matrix workflows. Only after this gate is green should the executor's failure-injection and process-death recovery work proceed.


## CI investigation — Windows restore directory owner — 2026-11-09

- Root CI run [42642834055](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42642834055), commit `8ee245f32fac31ce9dfa1b058163ee90b4f52960`: Ubuntu passed all checks, including race tests and build. Windows passed vet but failed Go tests in `internal/infrastructure/sqlite`; later Windows validation steps were skipped.
- The failing cases all reported `restore state directory owner does not match the current user`: regular no-replace move, destination collision, concurrent collision, pinned-parent traversal, and private temp-directory security test. The failure was in the new owner identity assumption, not compilation.
- Wails run [42642834036](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42642834036) passed both Ubuntu and Windows native builds, CLI smoke tests and executable verification; Ubuntu WebView bridge/database reopen also passed.
- Microsoft TOKEN_OWNER documentation confirms that the token's default owner SID is the identity Windows assigns to newly created objects; this can differ from TokenUser (notably for elevated tokens). The previous exact TokenUser-only comparison rejected legitimate objects created under the process token's default owner.
- Updated `restore_security_windows.go` to retrieve TokenOwner with `GetTokenInformation(TokenOwner)` and accept the directory owner only when it equals either TokenUser or the token's default owner. DACL broad-write checks remain unchanged and fail-closed; no arbitrary group membership is accepted.
- Fix commit: `8cb7b7fd3d987fbb2e2fb0484ad67e2e3733baf5` on `foundation/bootstrap`. This is a targeted code correction; no new CI result exists yet. Re-run Root CI and inspect Windows test output before proceeding.
- Restore remains disabled. No executor, journal, startup recovery or UI work started.


## CI closure and lifecycle review — 2026-11-09 15:30Z

### Windows security gate result

- Root CI [42643886765](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42643886765) completed SUCCESS on commit `8cb7b7fd3d987fbb2e2fb0484ad67e2e3733baf5`.
- Ubuntu job `129251819154`: SUCCESS. Go vet, Go tests, Linux race detector, Go build, Python compile/tests, frontend syntax and architecture boundaries all passed.
- Windows job `129251819474`: SUCCESS. Go vet/tests/build, Python compile/tests, frontend syntax and architecture boundaries passed; Linux-only race step skipped as expected.
- Job logs were retrieved for both platforms; no failure markers or failed steps.
- Wails run [42642834036](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42642834036) previously completed successfully on Ubuntu and Windows.

### StoreController and composition-root review

- Re-read `internal/infrastructure/sqlite/store_controller.go` and its test suite. The controller is already integrated; no duplicate implementation was added.
- Verified Wails composition in `desktop-src/main.go.tmpl`: migrate before controller creation; create one application-lifetime controller; configure `appconfig.Service` with the controller; pass the same controller to AppService; defer bounded controller.Close after `app.Run`.
- Verified `desktop-src/app_service.go.tmpl`: AppService retains the controller, not a raw Store; backup export calls `CreateWithStoreController`; config load/save use the controller-backed repository.
- Verified `internal/infrastructure/backup/archive.go` and `archive_controller_test.go`: archive creation holds a shared lease for the complete archive operation/publication; test pauses at publication and proves Close cannot drain until export releases its lease.
- StoreController tests cover shared/exclusive ordering, writer preference, cancellation, stale-store replacement, close drain, closed-controller rejection, panic fail-closed, stress, and config load/save drain. No missing integration work was found in these areas.

### Documentation correction

- Updated task plan to close the Windows owner/DACL and volume qualification gates with the actual CI evidence and mark executor/startup recovery as the next implementation phase.
- The Windows policy is deliberately limited to fixed local NTFS with persistent ACLs. Token owner validation accepts only TokenUser or the process token's explicit TokenOwner; broad-write DACL rejection remains fail-closed.
- Restore remains disabled. No executor, startup recovery, app-level Restore service or UI was introduced in this review.

### Next engineering step

Proceed test-first with the collision-safe filesystem recovery executor, following `docs/architecture/RESTORE_RECOVERY_EXECUTOR.md`. It must use the pinned directory/source handles, verify content identities, persist intent/completion around each no-replace mutation, sync namespace changes where supported, and preserve all ambiguous artifacts. Add failure-injection and abrupt-process-restart tests on Linux and Windows before any Restore exposure.


## Windows restore ancestor ACL hardening — 2026-11-09

- During post-green StoreController/Restore boundary review, identified that the Windows path walk pinned every directory component but only checked the final state directory's owner/DACL. A writable ancestor could permit namespace replacement at the canonical path across process restarts.
- Added Windows tests for DACL inspection on a private temporary directory and the local volume root in commit `cc5b770926cf308ab1d255cfa8eef4596ff82e76`. Its workflow run [42645878928](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42645878928) was cancelled before test execution when the implementation commit superseded it; no RED is claimed.
- Added `verifyRestoreDirectoryACL` and shared DACL validation in `restore_security_windows.go`; it rejects NULL DACLs, unsupported ACE layouts and broad write-capable ACEs without imposing a current-user owner requirement on system-owned ancestors.
- Updated Windows path traversal to apply the ACL check to the volume root and every opened child component before proceeding. The final state directory still additionally requires owner == TokenUser or explicit TokenOwner and passes the fixed-local-NTFS/persistent-ACL gate.
- Implementation commit `7f7d7ee773f5656ebbd414a1c9276446e1b48d34`; traversal integration commit `781d44be75346ce7a5506fc1bc6b2e7b12ff3e96`.
- Root CI [42645966152](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42645966152) is running on the integrated revision. Both Ubuntu and Windows jobs have started; no test result is claimed yet.
- Restore remains disabled. This closes an additional Windows namespace-permission gap; it does not replace executor identity checks, durable intent/move orchestration, or process-death recovery.

## Restore recovery planner correction — 2026-11-09 16:30Z

- Confirmed Root CI [42649675343](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42649675343) completed SUCCESS on `7b0ad259b5a741229e0d492b38c2a38b7a1dc3ec`; Ubuntu and Windows passed, and Ubuntu race detection passed.
- Test-first regression: `cd14e2e6a415b6decd77ac7a0b8036c8a6a0d6e3` added cases for candidate loss during `preserving-previous`. The test-only commit failed on Ubuntu as expected because the planner still accepted both states; no Windows RED claim is made.
- Fix `7b0ad259b5a741229e0d492b38c2a38b7a1dc3ec` now requires the staged candidate identity in `preserving-previous`. The planner returns an empty plan and `ErrRestoreRecoveryAmbiguous` when it is missing.
- Reconciled executor documentation with actual code: all three intent-phase planner truth tables and both platform no-replace primitives exist; the filesystem executor and process-death tests do not.
- Identified a remaining Linux boundary: the parent directory is pinned with openat2, but the source leaf is still selected by name at rename time. The executor must hash/size-verify the moved file through a no-follow opened handle after each move and before opening SQLite; mismatch means preserve and stop.
- Primary durability reference: Linux fsync(2) explicitly says syncing a file does not necessarily persist its containing directory entry; sync the parent directory as a separate step. Windows journal publication uses MOVEFILE_WRITE_THROUGH; durable journal removal intentionally remains unsupported.
- Next: implement cross-platform no-follow handle-based content identity verification, then build the executor one journaled no-replace move at a time with injected failures and abrupt-process restart tests. Keep Restore disabled and do not add GUI/CLI/startup wiring.

## Continuation — Restore identity verification and Windows durability gate — 2026-11-09

### Repository and CI state

- Re-read the active foundation plan, progress log, Restore executor design, journal, recovery planner, and platform move/identity implementations on foundation/bootstrap.
- Reconfirmed Root CI 42643886765 on 8cb7b7fd3d987fbb2e2fb0484ad67e2e3733baf5: Ubuntu and Windows both SUCCESS; Ubuntu race detector SUCCESS. Wails 42642834036 passed both platforms.
- Checked Root CI 42645966152 after Windows ancestor ACL hardening: Ubuntu SUCCESS, Windows FAILURE in Go tests. The later planner revision 7b0ad259b5a741229e0d492b38c2a38b7a1dc3ec passed Root CI 42649675343 on Ubuntu and Windows; Ubuntu race detector passed. This closes the observed ACL/planner failure on that tested revision.
- Current code head observed before documentation updates: 38ffa21f8b2fef05bc98b80e4ed5c8c1836bd0ff, commit message “test(sqlite): cover post-move restore identity verification”. Combined commit status returned an empty status list, and the PR-filtered workflow lookup returned no runs. Do not claim the later identity-reader and verified-move tests passed CI.
- Documentation-only commits from this review: d0b9e77d3b1f7e68c2e10f37f1ec7ea8a0cb247e (executor durability gate) and b41255b4e00c5b637e8e4137729bef567e0da687 (task plan). The root CI workflow path filter does not include these Markdown-only changes.

### Code review findings

- Cross-platform no-follow content identity readers are now present:
  - Linux pins the parent with openat2, opens the leaf with O_NOFOLLOW, hashes from the descriptor, and compares file identity/size/timestamps before and after reading.
  - Windows opens the leaf relative to the retained, security-checked directory handle with NtCreateFile and FILE_OPEN_REPARSE_POINT, hashes from that handle, and compares volume/file ID/size/last-write metadata before and after reading.
- moveRestoreFileNoReplaceVerified calls the platform no-replace move, then verifies the destination's size and SHA-256 through the no-follow identity reader. If verification fails or mismatches, it leaves the destination untouched and returns an error. It does not open SQLite.
- The verified-move helper is not a complete executor and currently does not sync the changed directory namespace. The Linux no-replace primitive uses renameat2(RENAME_NOREPLACE) but does not yet call parent-directory fsync. The Windows handle-bound rename also has no integrated, documented namespace durability barrier.
- The latest test commit adds success, mismatch-preservation, and invalid-expected-identity cases for the verified move. These tests are present in source but remain CI-unverified at this review point.

### Exa research — primary API contracts

- Linux fsync(2) explicitly states that syncing a file does not necessarily persist its directory entry; the containing directory must be synced separately: https://man7.org/linux/man-pages/man2/fsync.2.html
- Linux rename(2) documents RENAME_NOREPLACE as failing when the destination exists and notes that filesystem support is required: https://man7.org/linux/man-pages/man2/rename.2.html
- Microsoft MS-FSCC FileRenameInformationEx defines the no-replace behavior when FILE_RENAME_REPLACE_IF_EXISTS is unset. Its published flags do not include a write-through/durability flag: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/4217551b-d2c0-42cb-9dc1-69a716cf6d0c
- Microsoft NtSetInformationFile documents the operation/status contract, not a durable namespace barrier: https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntsetinformationfile
- Microsoft FlushFileBuffers requires GENERIC_WRITE; its documented volume-wide form requires administrative privileges. The reviewed documentation did not establish a supported non-administrative user-mode sequence to flush the parent directory entry changed by the handle-bound rename: https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-flushfilebuffers
- Microsoft MoveFileExW describes MOVEFILE_WRITE_THROUGH, but the stated flush guarantee specifically covers a move performed as a copy-and-delete operation. It does not establish the required durability contract for the same-volume rename path, and replacing the retained-handle native rename with path-based MoveFileExW would reintroduce the source-substitution race: https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw

### Decision and next step

- Updated docs/architecture/RESTORE_RECOVERY_EXECUTOR.md with the explicit Windows namespace durability blocker and the current identity-helper implementation status.
- Updated task_plan.md: identity readers/post-move verification are implemented but still need CI verification; Linux parent-directory fsync and a documented Windows namespace durability barrier are explicit gates.
- Do not start the filesystem executor yet. Resolve the Windows durability contract using primary documentation and a regression test, or keep Windows Restore unavailable. Do not add a path-only or overwrite-capable fallback. Implement Linux directory fsync as a separate, testable barrier.
- Restore remains disabled. No startup recovery, application-level Restore service, GUI/CLI Restore entry point, or main merge was added.


## CI observation — verified-move change — 2026-11-09 18:15Z

- Code under validation: `38ffa21f8b2fef05bc98b80e4ed5c8c1836bd0ff` (includes `moveRestoreFileNoReplaceVerified` and its Linux/Windows regression tests).
- Root CI run [42671657858](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42671657858) reports FAILURE, but the GitHub jobs endpoint returned zero jobs, no job logs were available, and the Actions page displayed “Waiting for pending jobs”. No Go test failure or code-level error is evidenced. Treat this as an unresolved workflow/job scheduling result, not a proven code regression.
- SQLite driver evaluation [42671658166](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42671658166) reports FAILURE while its only created job, `modernc / windows-latest`, completed SUCCESS. Its logs show `go mod verify`, `TestProductionMigration`, and `TestSQLiteDriverEvaluation` passing. The other expected matrix cells were not created in the job list; the Actions page displayed “Waiting for pending jobs”. This workflow does not exercise the main Restore package tests.
- Wails platform proof [42671658474](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42671658474) reports FAILURE while its only created job, `Wails v3.0.0 / windows-latest`, completed SUCCESS, including native build, Windows CLI smoke test, and executable verification. The Ubuntu matrix lane was not created; no failing step is evidenced.
- Linux distribution compatibility [42671658275](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42671658275) remained IN_PROGRESS. At the last inspection, only `Debian 12 / legacy-gtk3-webkit41` was active and building the native application; no conclusion is claimed.
- The earlier runs for intermediate commit `ee223221` were superseded/cancelled by the later test commit. No test result is claimed for those cancelled runs.
- Do not add a code commit merely to force another workflow while the distro matrix remains active. Recheck GitHub Actions after the active lane completes; if root CI still has no jobs, use an actual workflow-level retry path if available or record the infrastructure limitation. Do not mark identity tests verified until Ubuntu and Windows Go test jobs actually execute.
- Restore remains disabled. No executor, startup recovery, CLI/GUI Restore, or main merge.


## CI closure — verified-move revision — 2026-11-09 18:16Z

- Final inspected code-test SHA: `38ffa21f8b2fef05bc98b80e4ed5c8c1836bd0ff`. Documentation-only follow-ups moved the branch head afterward; they do not change the tested code.
- Root CI [42671657858](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42671657858): FAILURE with zero jobs returned by the GitHub jobs endpoint and no job logs. The Actions page displayed “Waiting for pending jobs”. There is no platform test output for this run.
- SQLite driver evaluation [42671658166](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42671658166): workflow FAILURE; its only created Windows modernc job passed. The log confirms module verification, production migration test, and SQLite driver evaluation passed. Expected matrix lanes were absent; this workflow does not execute the main Restore package.
- Wails platform proof [42671658474](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42671658474): workflow FAILURE; its only created Windows job passed native build, Windows CLI smoke, and executable verification. The Ubuntu lane was absent.
- Linux distribution matrix [42671658275](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42671658275): workflow FAILURE; the only created job, Debian 12 / legacy GTK3-WebKit2GTK 4.1, completed SUCCESS. Its log confirms native build, CLI smoke, two WebView-to-Go bridge launches, and database reopen passed. Seven other matrix lanes were not created. Wails binding generation emitted warnings about an initial untagged WebKitGTK 6.0 probe in this GTK3 lane, but the tagged GTK3 build and all smoke steps passed.
- No observed job failure identifies a regression in the verified-move code. However, the main Ubuntu/Windows Root CI test jobs never ran, so the new identity and verified-move tests remain unverified.
- The run-level failures with missing matrix jobs remain an unresolved GitHub Actions scheduling/runner issue. Do not mask this by calling the workflows green or by claiming all matrix cells passed.
- Restore remains disabled. Next: obtain an actual Root CI run with Ubuntu and Windows jobs, then verify the identity tests. Continue to defer executor work until the Linux parent-directory fsync and Windows namespace durability gates are resolved.


## Continuation — Linux no-replace directory durability barrier — 2026-11-09

- Updated the Linux no-replace move adapter to call `fsync(2)` on the same parent-directory descriptor already pinned with `openat2`, immediately after a successful `renameat2(RENAME_NOREPLACE)`. It does not reopen the directory by path.
- A directory-sync error is returned to the caller and is not reported as a successful durable move. The namespace operation may already have happened; the destination is deliberately left in place for journal-driven inspection/reconciliation. No rollback, deletion, or overwrite fallback is attempted.
- Added Linux regression coverage for syncing an opened directory descriptor and rejecting an invalid descriptor. Existing verified-move tests also exercise the integrated rename-plus-directory-sync path when run on Linux.
- Tests have not been executed locally in this connector-only repository session. The next code commit must be verified by actual Ubuntu and Windows Root CI jobs; the previous Root CI run had no jobs and did not execute tests.
- Windows handle-bound rename still has no established non-admin namespace durability barrier. Keep Windows Restore unavailable and Restore globally disabled; no executor, startup recovery, UI/CLI entry point, or main merge.


## Continuation — CI audit and Linux regression-test restoration — 2026-11-09

- Re-audited the completed workflows for commit `6bb79ce73ea558bb07b2d24e889ea0c85f2013fc`.
- Root CI run #42673055491 reports run-level failure even though its only materialized job, `Validate (ubuntu-latest)`, completed every step successfully, including `go vet ./...`, `go test ./...`, `go test -race ./...`, `go build ./...`, Python tests, frontend syntax, and architecture checks. The configured Windows matrix job was not materialized. This is not Windows Restore test evidence.
- SQLite driver evaluation run #42673055205 reports run-level failure although all three materialized jobs succeeded: ncruces on Ubuntu, ncruces on Windows, and modernc on Windows. Those jobs test the driver evaluation package, not the Restore filesystem package.
- Linux distribution run #42673055109 completed all seven materialized distro/stack jobs successfully. This proves those desktop builds/smoke checks, not Windows Restore correctness.
- Wails run #42673055064 failed with zero jobs. Its workflow-level failure remains unexplained; no job logs exist to diagnose.
- Restored `TestMoveRestoreFileNoReplaceRejectsSymlinkedContainingDirectory`, which had been accidentally dropped when the new Linux directory-sync tests replaced the file. Kept both directory-sync tests.
- The restored test has been committed but has not yet been executed by CI. Do not claim this regression is green until a new Linux test job passes.
- Windows namespace durability remains unresolved. Restore stays disabled; no executor, startup recovery, UI/CLI entry point, or main merge.


## CI verification — restored Linux regression and cross-platform Restore tests — 2026-11-09

- Root CI #42682590849 completed successfully for commit `f1eb8179a3a95e5cabfa19baebbade03e97b8bf5`: both `Validate (ubuntu-latest)` and `Validate (windows-latest)` succeeded. Logs show `go test ./...` passed on both operating systems; Ubuntu also passed `go test -race ./...`. The SQLite package passed on both platforms. On Linux this includes the restored symlinked-parent rejection test and both directory-sync tests.
- SQLite driver evaluation #42682590841 completed successfully across all four materialized jobs: modernc Ubuntu, modernc Windows, ncruces Ubuntu, and ncruces Windows. This is driver-evaluation coverage, separate from the Restore package tests above.
- Wails v3 platform proof #42682591267 is still running on Ubuntu and Windows; both jobs are at the native application build step. No conclusion recorded yet.
- Linux distribution compatibility #42682591212 is still running. Arch, Fedora 44, Debian 13, and Ubuntu 24.04 were building; Ubuntu 22.04, openSUSE, Fedora legacy GTK3, and Debian 12 remained queued at last check. No workflow conclusion recorded yet.
- This validates the current Linux directory-sync implementation and the existing Windows identity/verified-move tests under Root CI, but it does not resolve Windows namespace durability. Do not interpret Windows test success as proof of a durable parent-directory namespace flush.
- Restore remains disabled. No filesystem executor, startup recovery, GUI/CLI Restore, or main merge.


## Follow-up — Wails Windows native build proof — 2026-11-09

- Wails v3 platform proof run #42682591267: the Windows job completed successfully. Its native application build, Windows CLI dispatch smoke test, and executable verification all passed. The Ubuntu job was still running its native application build at last check; the workflow is not yet green.
- Linux distribution run #42682591212: Debian 13 completed successfully. Arch, Fedora 44, openSUSE, and Ubuntu 24.04 were still active; Ubuntu 22.04, Debian 12, and Fedora legacy GTK3 remained queued at last check.
- Root CI remains green on Ubuntu and Windows, and SQLite driver evaluation remains green across its four jobs for commit `f1eb8179a3a95e5cabfa19baebbade03e97b8bf5`.
- No additional code changes in this update. Windows namespace durability remains unresolved; Restore remains disabled.


## Follow-up — Wails platform proof completed — 2026-11-09

- Wails v3 platform proof run #42682591267 completed successfully on both Windows and Ubuntu. Windows: native application build, CLI dispatch smoke test, and executable verification passed. Ubuntu: native build, CLI dispatch smoke test, executable verification, and live WebView bridge/database reopen smoke test all passed.
- Linux distribution run #42682591212: Arch Linux, Fedora 44 GTK4, and Debian 13 completed successfully. Ubuntu 22.04, openSUSE Tumbleweed, Fedora 44 legacy GTK3, and Ubuntu 24.04 were still active; Debian 12 remained queued at last check. Workflow still in progress.
- Root CI #42682590849 and SQLite driver evaluation #42682590841 are both green. These runs tested code commit `f1eb8179a3a95e5cabfa19baebbade03e97b8bf5`; subsequent commits were documentation-only.
- Windows namespace durability remains unresolved despite successful Windows compilation/tests and Wails build. Restore remains disabled.


## Continuation — Linux matrix completion and Windows durability contract review — 2026-11-09T19:47Z

### CI verification
- Linux distribution compatibility run #42682591212 completed successfully: all 8/8 jobs passed (Arch Linux rolling / GTK4-WebKit6; Fedora 44 / GTK4-WebKit6; Ubuntu 22.04 / legacy GTK3-WebKit41; openSUSE Tumbleweed / GTK4-WebKit6; Fedora 44 legacy GTK3 / WebKit41; Debian 13 / GTK4-WebKit6; Ubuntu 24.04 / GTK4-WebKit6; Debian 12 / legacy GTK3-WebKit41).
- Each of the eight jobs completed the native Wails application build, CLI dispatch smoke test, and live WebView-to-Go bridge/database reopen smoke test. The run-level conclusion is success.
- The run tested code commit `f1eb8179a3a95e5cabfa19baebbade03e97b8bf5`. Branch HEAD was rechecked and remains documentation-only descendant `10b301a32477a53a70f9284e74786c7a6e611ab5`.
- Linux matrix run: https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42682591212

### Windows durability research
- Re-read the Microsoft contracts for `MoveFileExW`, `FILE_RENAME_INFORMATION_EX`, `NtFlushBuffersFileEx`, `NtCreateFile`, file access rights, and `ReplaceFileW`.
- `FILE_RENAME_INFORMATION_EX` documents no-replace behavior when `FILE_RENAME_REPLACE_IF_EXISTS` is omitted, but its published flags contain no write-through or namespace-durability flag. It remains the handle-relative no-replace primitive, not a documented durability barrier.
- `MoveFileExW(MOVEFILE_WRITE_THROUGH)` says the call does not return until the file is moved on disk, but its explicit flush guarantee is scoped to copy-and-delete moves. The public contract does not clearly extend that guarantee to a same-volume rename's destination directory entry. It is also path-based, so substituting it would abandon the current retained-source-handle/rooted-destination semantics and reintroduce path substitution concerns.
- `NtFlushBuffersFileEx(Flags=0)` documents flushing file data/metadata and synchronizing storage, but its error contract requires write or append access. The `NtCreateFile` documentation says not to request `FILE_WRITE_DATA` or `FILE_APPEND_DATA` for directories and says generic access rights cannot be specified for a directory. No documented, reconciled user-mode directory-handle access contract was found.
- `FLUSH_FLAGS_FILE_DATA_SYNC_ONLY` is expressly invalid for directory handles. `ReplaceFileW` documents `REPLACEFILE_WRITE_THROUGH` as unsupported.
- The two Microsoft Q&A pages reviewed are technical clarification questions, not an authoritative Microsoft engineering answer. Their existence is not treated as Microsoft confirmation. The blocker is the gap between the published API contracts, not a claim that NTFS cannot persist directory metadata.
- Decision unchanged: no documented non-admin Windows barrier currently satisfies both handle-bound no-replace semantics and the required namespace durability guarantee. Do not add a path-based or overwrite-capable fallback. Windows Restore remains unavailable; Restore stays globally disabled.

### Scope and next work
- No application code changed in this continuation. No Restore executor, journal integration, startup recovery, GUI/CLI Restore, or merge to main was started.
- Continue evaluating only documented Windows APIs or an alternative storage design that can prove equivalent safety. Otherwise retain the explicit Windows gate and record the limitation.


## Continuation — in-place SQLite restore feasibility probe — 2026-11-09

### Work performed
- Added test-only coverage in `internal/infrastructure/sqlite/store_test.go`; no production Restore path was enabled or changed.
- Verified that `Store.restoreFromPath` can apply a verified snapshot into an existing database file without changing its OS file identity.
- Verified that an invalid source does not change the existing target.
- Added subprocess termination tests: termination after one copied page but before `SQLITE_DONE` recovers the previous database; termination after the backup API completes but before a clean Store close exposes the restored database after reopen.
- Tested a 1024-byte-page source against a 4096-byte-page target. A WAL destination rejects the mismatch and retains its old data. With PERSIST rollback journaling, the in-place backup completes; after closing and reopening, the target has the source page size, passes database verification, and contains the expected configuration.
- The connection that performed the backup may report its cached pre-restore `PRAGMA page_size`; verification must close and reopen the target before inspecting or using the restored database.

### CI evidence for code commit `733623d438b81bcf4792098d5ebc330583cfdfd8`
- Root CI #42695205394: Ubuntu and Windows succeeded; Go tests and Linux race tests passed.
- SQLite driver evaluation #42695205322: modernc and ncruces succeeded on Ubuntu and Windows (4/4).
- Wails platform proof #42695205294: Ubuntu and Windows succeeded.
- Linux distribution compatibility #42695205101: all 8/8 distributions/GTK-WebKit combinations succeeded.

### Safety scope and decision
- This is a lower-level SQLite primitive probe, not a replacement Restore architecture. `RestoreDatabase` still uses its existing staged-file publication path.
- PERSIST journal mode retains its rollback-journal file and commits by invalidating the journal header instead of deleting the filename. It is not persistent across database reopen, so a restore bootstrap must explicitly set it. The initial creation of the journal file still requires a durable file-creation guarantee; process-death tests do not establish power-loss durability.
- The Windows JSON journal publisher remains a separate blocker: `restore_journal_publish_windows.go` still uses path-based `MoveFileExW(REPLACE_EXISTING|WRITE_THROUGH)`. The Microsoft contract reviewed earlier does not clearly guarantee same-volume rename namespace persistence.
- No Restore executor, durable journal integration, startup recovery, UI/CLI Restore, or merge to main was started. Restore remains disabled pending complete Linux and Windows safety gates.


## Continuation — Windows pinned-directory relative rename — 2026-11-09 23:55Z

### CI audit on base commit `0e3e2868e6a74c9fdae698b314c987fe39391ddd`
- Root CI [42706477500](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42706477500): Ubuntu and Windows jobs completed successfully. Both ran Go vet/tests/build, Python compile/tests, frontend syntax and architecture checks; Ubuntu race detection passed. Windows correctly skipped the Linux-only race step.
- SQLite driver evaluation [42706477451](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42706477451): all four modernc/ncruces × Ubuntu/Windows jobs completed successfully.
- Wails platform proof [42706477450](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42706477450): Ubuntu and Windows jobs completed successfully; Linux WebView bridge/database reopen and Windows native build/CLI/executable checks passed.
- Linux distribution compatibility [42706477397](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42706477397): all eight jobs completed successfully, including both legacy GTK3/WebKit and GTK4/WebKit families.

### Windows design correction prepared in this revision
- Removed the full-path destination rename path and its second directory guard.
- Source is opened using a path obtained from the pinned directory handle, then its own final path and the directory handle's final path are re-read and checked for exact parent + expected leaf before proceeding.
- The exact source handle retains `FILE_FLAG_WRITE_THROUGH`; source contents are flushed before the namespace operation.
- Rename is now issued with native `NtSetInformationFile`, information class 65 (`FileRenameInformationEx`), `RootDirectory` set to the pinned destination directory handle, and a single relative destination leaf. No path-based or overwrite-capable fallback is introduced. No-replace remains the default; journal replacement explicitly opts into replace flags.
- Added Windows regression coverage for ancestor rename + replacement-tree decoy, rejection of a source handle from the decoy tree, and relative journal-leaf replacement.

### Safety boundary
- Microsoft documents `RootDirectory` as the handle-relative target root and requires a simple leaf name for that form. The MS-FSCC layout specifies a 20-byte fixed header and minimum 24-byte information buffer; network operations require a zero RootDirectory, so this code remains limited to the already-qualified local NTFS path.
- Microsoft documents that a `FILE_FLAG_WRITE_THROUGH` request causes NTFS to flush metadata changes, including rename operations, that result from processing the request. The implementation sends the rename request through the same write-through source file object. This is a documented NTFS contract, not a claim of hardware-independent power-loss guarantees.
- The new Windows implementation is **not yet CI-verified**. Keep Restore disabled until the actual Windows Root CI and all other required platform gates pass. No executor, startup recovery, UI/CLI Restore, or merge to main.


## Windows Root CI regression correction — 2026-11-10 00:12Z

- Root CI run [42725119111](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42725119111) on `79c3e6f3e5b8e6420bb965a112aba8f4e6664be4`: Ubuntu passed; Windows failed in exactly two new test setups. Logs show both failures occurred before the rename assertions because Windows returned `Access is denied` when the tests attempted to rename an ancestor directory while a descendant handle was open.
- SQLite driver evaluation [42725119131](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42725119131) passed all four jobs. Wails proof [42725119125](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42725119125) passed Ubuntu and Windows. Linux distribution matrix [42725119228](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42725119228) completed 8/8 successfully.
- Corrected the two regression tests to rename the pinned directory object itself, then create a replacement directory at its old path. This exercises the security property without relying on Windows allowing an ancestor rename with an open descendant. The test requires the native relative rename to publish into the original directory object, not the replacement tree, and rejects a decoy source handle from the replacement directory.
- Also corrected accidental extra indentation in the Windows source-open declarations in the move path, journal publisher, and tests.
- The implementation remains native `NtSetInformationFile(FileRenameInformationEx)` class 65 with the pinned directory handle in `RootDirectory`, a single relative leaf, and the same `FILE_FLAG_WRITE_THROUGH` source handle. No full-path rename or overwrite-capable fallback.
- This revision is pending fresh Windows Root CI. Do not mark the Windows safety gate passed until the Go package tests succeed on Windows. Restore remains disabled; no executor/recovery/UI/CLI integration and no merge to main.


## Windows directory-rename regression sequencing — 2026-11-10 00:15Z

- Root CI [42726626677](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42726626677) on `ec10edfbc91e3edd937abd71fad94e8a5fa2ca3f`: Ubuntu passed; Windows Go tests failed in the new directory-rename regression. The Windows log confirmed `Access is denied` when renaming the directory while the source file handle was open.
- Microsoft documents that a directory cannot be renamed while files in its tree have open handles. The test now renames the pinned directory before opening the source file, then creates a replacement directory and decoy at the original path. It opens the source through the retained directory handle and checks that the native relative rename lands only in the renamed original directory.
- The separate decoy-handle validation test likewise creates the replacement tree after renaming the pinned directory and rejects a handle opened from that replacement.
- This changes test ordering only; production rename implementation is unchanged. Fresh Windows Root CI is required.
- Restore remains disabled; no executor/recovery/UI/CLI integration and no merge to main.


## Verification closure — Windows relative rename regression and WebKit signal audit — 2026-11-10 00:20Z

### Current code commit
- Branch `foundation/bootstrap`, code HEAD `410dbf88d091e012f360fe5f3fe409da05811615`: [commit](https://github.com/SoufianoDev/OhMyFreeAI/commit/410dbf88d091e012f360fe5f3fe409da05811615).

### CI on `410dbf8`
- Root CI [42726913181](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42726913181): Ubuntu and Windows both SUCCESS. Windows Go vet/tests/build, Python compile/tests, frontend syntax and architecture checks passed; Linux Go race detector passed. Windows correctly skipped Linux-only race detection.
- SQLite driver evaluation [42726913184](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42726913184): all four jobs SUCCESS (modernc and ncruces on Ubuntu and Windows).
- Wails platform proof [42726913002](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42726913002): both jobs SUCCESS. Linux native build, CLI smoke, executable verification, two live frontend-to-Go bridge calls and database reopen smoke passed. Windows native build, CLI smoke and executable verification passed.
- Linux distro compatibility [42725119228](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42725119228) on ancestor commit `79c3e6f`: 8/8 jobs SUCCESS. No production code changed after that matrix; later commits changed only Windows tests and planning records.
- Root CI run [42726626677] on intermediate commit `ec10edf` passed Ubuntu and failed Windows only in a test setup that attempted to rename a directory while a file in it remained open. The test was reordered to rename before opening the source. The final Windows Root CI above passes.

### Windows handle-bound rename
- The final regression now proves: rename the retained directory object before opening its source; create a decoy replacement tree at the old path; open the source using the retained directory handle's resolved path; then perform `NtSetInformationFile(FileRenameInformationEx)` with the destination directory handle in `RootDirectory`. The original content must publish only into the retained directory, and the decoy remains untouched.
- The independent decoy-handle validation test and journal relative-replace test also pass in Windows Root CI.
- This verifies the implementation and regressions on the Windows CI runner. It is not a hardware-independent power-loss proof and does not activate Restore.

### JSC signal 10 investigation
- Exa research reviewed WebKit's configurable GC-signal work and upstream Wails changes [#5507](https://github.com/wailsapp/wails/commit/6329e9d2bec94a6916b9a700667b78a9569a3d9b) and [#5530](https://github.com/wailsapp/wails/commit/ef782b77c892b6a7e45c3227f73fb0e20d0ed600).
- Inspected the actual Wails `v3.0.0` tag source. GTK4 and GTK3 paths both leave `SIGUSR1` to JavaScriptCore (do not add Go's `SA_ONSTACK` flag to JSC's handler), schedule the signal-stack compatibility repair for the first five seconds after WebView creation, and reapply it at `WEBKIT_LOAD_FINISHED`.
- The warning is a known JSC signal-handler ownership diagnostic, not proof by itself of a failed WebView. Do not set `JSC_SIGNAL_FOR_GC` to an arbitrary signal as a cosmetic workaround. The current Linux CI smoke log did not reproduce the warning; it did pass both bridge calls and database reopen.
- No local signal-handler override was added; the project uses the pinned Wails v3.0.0 upstream implementation.

### Remaining safety boundary
- Restore remains disabled. Passing Go tests and CI does not prove crash/power-loss durability across hardware, storage controllers, or filesystems.
- Do not start Restore Executor, durable journal integration, startup recovery, UI/CLI Restore, or merge to `main` until all required safety gates are separately demonstrated.


## 2026-11-10 07:14Z — planning reconciliation follow-up

- Rechecked the docs-only commit `dd3a3185ef4b2b9cb1a5ea9654a7a85551f6aed3`. It changed only `task_plan.md` and `RESTORE_RECOVERY_EXECUTOR.md`; the workflow API returned zero runs for that SHA. This is expected: the core CI path filters do not include `docs/**` or `.planning/**`. No code changed in that commit.
- Updated the plan's remaining Restore work to distinguish the isolated transaction executor, failure/process-death tests, and the later startup coordinator. Startup recovery must not be wired before the executor matrix is green.
- Confirmed no open pull requests. The branch remains 804 commits ahead of `main` and 0 behind; do not merge it wholesale.
- Code verification remains tied to `9ca1f7e6de8ecca00970b3201601ca70f31ddb95`, not the documentation-only head. Its Root CI, SQLite matrix, Wails proof and Linux distro matrix are all green as recorded above.


## Follow-up session: 2026-11-10T14:08Z — reconcile distro CI and planning state

### Initial repository and workflow check
- Rechecked the latest branch-triggered Actions runs for `foundation/bootstrap`; the latest code head remained `beab78a04f4ef9ae1421247983d25da66ffc3986`.
- The latest Root CI run `42775712875` completed successfully: Ubuntu and Windows jobs passed; the Linux race-detector step passed.
- The latest SQLite driver evaluation run `42775712758` completed successfully: all four modernc/ncruces × Ubuntu/Windows jobs passed.
- The latest Wails platform proof run `42775712787` completed successfully: both Ubuntu and Windows jobs passed. Ubuntu included the native WebView-to-Go bridge and database-reopen smoke.
- The latest Linux distribution compatibility run `42775712919` completed successfully: all eight jobs passed. In particular, `Ubuntu 22.04 / legacy-gtk3-webkit41` (job `129693622514`) completed every required build, CLI, and WebView/database-reopen step successfully.
- The prior report's pending-job/log-fetch issue is resolved by querying the current workflow and job state. The earlier 404 did not establish a failure in the distro workflow; the latest run is 8/8 successful.
- Also inspected the failed Root CI run `42775485018` (commit `cab00b8`) and fetched both platform job logs. Both failed in `go test ./...` with the same regression: journal writes accepted illegal phase transitions, changed immutable transaction metadata, allowed candidate identity in invalid phases, and permitted replacing malformed/terminal journals. This was a real production validation gap, not runner flakiness.
- The immediate follow-up code commit `beab78a` fixed the root cause by validating existing journal contents and enforcing immutable metadata, candidate identity timing, allowed phase transitions, and terminal-only journal removal. The exact code commit passed Root CI `42775712875` on Ubuntu and Windows; the other three workflow families also passed on that commit.
- No workflow was active at the time of this reconciliation.

### Changes made
- Updated the active task plan's top-level Next Step evidence from stale commit `9ca1f7e` to verified head `beab78a`.
- Replaced the stale 7/8/pending Ubuntu 22.04 statement with the actual 8/8 successful job results and exact run links.
- Kept the executor as the next engineering unit. No Restore implementation was started in this reconciliation because the requested issue was the unresolved CI status/reporting discrepancy; the executor remains a separate security-sensitive implementation with substantial cross-platform test requirements.
- No source code, workflow definitions, Restore safety gates, or `main` branch files were changed.

### Verification boundary
- Verified workflow/job totals: Root CI 2/2, SQLite 4/4, Wails 2/2, Linux distro 8/8.
- Verified the Ubuntu 22.04 job's individual steps: setup, container initialization, native dependency installation, checkout, Go/Node setup, pkg-config configuration, pinned Wails CLI installation, actual desktop build, native CLI smoke, and live WebView-to-Go/database-reopen smoke all succeeded.
- No local test execution is claimed; this session performed a read-only GitHub Actions reconciliation and planning-document update.
- Restore remains disabled and is not connected to startup recovery, GUI, or CLI.


## 2026-11-10 14:19Z — StoreController gate re-verification and next Restore prerequisite

### Requested workflow verification
- Inspected Root CI run [42419106009](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42419106009) for commit `f5396ee592b9b9756779f67aeabd6d7167520c40`.
- Ubuntu and Windows jobs both completed successfully (2/2). Both completed module verification, Go vet/tests/build, Python compile/tests, frontend JavaScript syntax and architecture checks. No failed steps were present in either job. This run predates the current code head and is not used as the sole proof for later changes.

### Current branch and code verification
- Current branch: `foundation/bootstrap`; branch head is documentation-only commit `e086fa76895c2d46ab4bfeee7bd5d104e4bd5e3f`, which changes only this task plan. Latest verified code commit remains `beab78a04f4ef9ae1421247983d25da66ffc3986`.
- Root CI [42775712875](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712875): Ubuntu and Windows 2/2 SUCCESS.
- SQLite driver evaluation [42775712758](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712758): modernc/ncruces × Ubuntu/Windows 4/4 SUCCESS.
- Wails platform proof [42775712787](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712787): Ubuntu and Windows 2/2 SUCCESS.
- Linux distribution compatibility [42775712919](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712919): 8/8 SUCCESS, including Ubuntu 22.04 legacy GTK3/WebKit41.
- The Root, SQLite, Wails and distro job lists were fetched and inspected; no failed or pending jobs were returned.

### StoreController review — integration already complete
- Reviewed `internal/infrastructure/sqlite/store_controller.go` and all 16 lifecycle tests in `store_controller_test.go`: shared/exclusive drain, writer preference, cancellation, panic/failure fail-closed behavior, replacement, close/rejection, concurrent stress, configuration round-trip and Load/Save drain.
- Re-verified `desktop-src/main.go.tmpl`: startup opens and migrates the Store before creating one application-lifetime StoreController; the configuration service and AppService receive that controller; application shutdown uses a bounded controller Close after `app.Run`.
- Re-verified `desktop-src/app_service.go.tmpl`: AppService retains the controller, not a raw Store. Config Load/Save go through the controller-backed repository. GUI backup export calls `CreateWithStoreController`.
- Re-verified CLI backup export creates a controller for the existing database and closes it through the controller. Backup inspection does not access the live Store.
- Re-verified `Manager.CreateWithStoreController`: it wraps the entire snapshot/archive/verification/publication operation in `WithShared`. `TestManagerCreateWithStoreControllerDrainsBeforeClose` blocks at the publication barrier and proves Close waits for the export lease. No duplicate integration changes were needed.

### New executor prerequisite found
- `RemoveRestoreJournal` validates a terminal phase, then delegates to the platform adapter. The Windows adapter currently returns `errRestoreJournalRemovalUnsupported` unconditionally.
- The journal uses one fixed filename and transition validation keeps transaction ID and transaction metadata immutable. Therefore, after a terminal journal is written, a later transaction cannot safely start on Windows: removal is unsupported and a new transaction cannot replace the old transaction's metadata through the transition validator.
- Do not ignore the removal error, delete the journal through a weaker path API, or overwrite terminal evidence. Before implementing the transaction executor, design a durable multi-transaction journal lifecycle that preserves old evidence and supports crash-safe discovery/recovery, then add Linux/Windows failure-injection and process-death tests for rollover.
- No source code changed in this review. Restore remains disabled; no startup recovery, GUI/CLI Restore or main merge was added.


## 2026-11-10 15:45Z — Windows namespace durability re-review and journal rollover blocker

### Repository and CI reconciliation
- Current branch: `foundation/bootstrap`; current branch head before this progress entry: `11ccba2f2b125775b991744642bf087d4171652f`.
- Latest code-bearing commit remains `beab78a04f4ef9ae1421247983d25da66ffc3986`. The later commits through `11ccba2` are planning/documentation updates; no production source changed after `beab78a`.
- Root CI [42775712875](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712875) was checked at job, step and log level. Both Ubuntu and Windows passed. Logs confirm both checked out exactly `beab78a04f4ef9ae1421247983d25da66ffc3986`. Ubuntu Go race detector passed; Windows correctly skipped the Linux-only race step. Go vet/tests/build, Python compile/tests, frontend syntax and architecture checks passed on both platforms.
- SQLite driver evaluation [42775712758](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712758): all 4 modernc/ncruces × Ubuntu/Windows jobs passed; job steps and checkout SHA were inspected.
- Wails proof [42775712787](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712787): both jobs passed. Linux native build, CLI smoke, executable verification, live WebView bridge and database reopen passed; Windows native build, CLI dispatch and executable verification passed. Platform-specific skipped steps were expected.
- Linux distribution matrix [42775712919](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42775712919): all 8 distro jobs passed, including Ubuntu 22.04 legacy GTK3/WebKit41. The Ubuntu 22.04 job's full build, CLI and WebView/database-reopen steps were inspected.
- These workflows verify `beab78a`; they do not verify power-loss durability. The current branch head is documentation-only beyond that code commit.
- Current compare against `main`: 818 commits ahead, 0 behind. Do not merge the branch wholesale.

### Windows namespace durability documentation correction
- Re-reviewed primary Microsoft sources through Exa: [File Caching](https://learn.microsoft.com/en-us/windows/win32/fileio/file-caching), [NtCreateFile](https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntcreatefile), [FlushFileBuffers](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-flushfilebuffers), and [MS-FSCC FileRenameInformationEx](https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/4217551b-d2c0-42cb-9dc1-69a716cf6d0c).
- The current Windows code opens the source relative to a pinned directory handle with native `FILE_WRITE_THROUGH`, flushes the source handle, and calls handle-bound `NtSetInformationFile(FileRenameInformationEx=65)`. This is useful handle/identity protection and a write-through request.
- The earlier documentation statement that Microsoft explicitly guarantees NTFS rename metadata durability was too strong. `NtCreateFile` describes `FILE_WRITE_THROUGH` in terms of file data writes; File Caching discusses metadata caching generally but does not explicitly promise that a parent directory's namespace entry changed by this handle-based rename is durable; MS-FSCC defines rename/no-replace semantics but no durability flag. `FlushFileBuffers` documentation describes flushing the specified file handle, not its parent directory entry.
- Therefore the Windows namespace-durability gate is **unresolved**, not closed. No undocumented directory flush, path-based fallback, `MoveFileExW`, or overwrite-capable fallback is acceptable. The previous claim is explicitly withdrawn in `docs/architecture/RESTORE_RECOVERY_EXECUTOR.md`; the task plan was corrected accordingly.
- Linux's documented boundary remains clearer: [fsync(2)](https://man7.org/linux/man-pages/man2/fsync.2.html) says syncing a file does not necessarily sync its directory entry; the Linux adapter syncs the pinned parent directory FD after the namespace mutation.

### Journal rollover blocker
- `restore_journal.go` uses one fixed `restore-journal.json`; `validateRestoreJournalTransition` keeps transaction ID and transaction metadata immutable.
- `RemoveRestoreJournal` permits only `committed` or `rolled-back` records, then calls the platform adapter. The Windows adapter intentionally returns `errRestoreJournalRemovalUnsupported`; its test asserts the terminal journal remains present.
- A second transaction therefore cannot start safely on Windows with the current single-file lifecycle. Do not work around this by calling `os.Remove`, weakening publication, or overwriting terminal/ambiguous evidence.
- Exa review included SQLite atomic-commit guidance, Linux `fsync(2)`, ext4 journal documentation and Windows primary API docs. No journal rollover design was approved in this session. An append-only or archived multi-transaction format still needs a complete crash-discovery protocol and platform-specific durable publication proof before implementation.

### Changes made
- Updated `docs/architecture/RESTORE_RECOVERY_EXECUTOR.md` in commit [032d55c](https://github.com/SoufianoDev/OhMyFreeAI/commit/032d55c94ef39bc2792cd43bae1e13aa19d26b6e) to withdraw the unsupported Windows rename-durability claim and document the exact source contracts and unresolved boundary.
- Updated `.planning/2026-11-07-ohmyfreeai-foundation/task_plan.md` in commit [11ccba2](https://github.com/SoufianoDev/OhMyFreeAI/commit/11ccba2f2b125775b991744642bf087d4171652f) to make both Windows namespace durability and journal rollover explicit prerequisites before the executor.
- No production code changed; no Restore executor, startup recovery, UI/CLI Restore or main merge was started.
- Restore remains disabled. Next engineering work is to design and prove a journal lifecycle that does not depend on unsupported Windows directory-entry durability, or to establish a documented supported Windows namespace-durability contract. Only then proceed to executor design and its Linux/Windows process-death/failure-injection matrix.


### Windows directory-flush follow-up — 2026-11-10 15:45Z
- Exa review also covered Microsoft's [NtFlushBuffersFileEx](https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntflushbuffersfileex) and [File Access Rights Constants](https://learn.microsoft.com/en-us/windows/win32/fileio/file-access-rights-constants).
- The native flush API documents `STATUS_ACCESS_DENIED` when the handle has neither write nor append access; `FILE_DATA_SYNC_ONLY` is explicitly invalid for directory handles. `NtCreateFile` prohibits file-data/write/append and generic file rights on directory opens. Directory-specific `FILE_ADD_FILE` / `FILE_ADD_SUBDIRECTORY` share the same numeric bits as file write/append rights, but the reviewed documentation does not define an exact supported directory-handle access mask for the flush API.
- This is a possible research lead, not an approved API contract. No directory-flush code or experiment was added. Do not infer support from a successful one-run probe, and do not use privileged volume flush as a normal-user workaround.
- Updated `RESTORE_RECOVERY_EXECUTOR.md` in commit [7346f34](https://github.com/SoufianoDev/OhMyFreeAI/commit/7346f3438887797f0286a2f0a12578d833ea518c) to record the access-contract gap.
- Clarified that Windows journal phase publication uses the same handle-bound replacement rename as database-artifact publication. Thus the unresolved namespace-durability gate blocks every intent/completion journal update, not merely artifact moves. This clarification is in [18a8f83](https://github.com/SoufianoDev/OhMyFreeAI/commit/18a8f8371b6441af6763b647d9079c1709f844df) and the task-plan correction [a05da29](https://github.com/SoufianoDev/OhMyFreeAI/commit/a05da293e60db419c6c012b89367b4a7884ca0b0).
- No production source changed. Restore remains disabled.


## 2026-11-10 — FIFO identity hardening, CI closure and Windows durability re-review

### Repository and code change
- Reconciled the remote branch before editing: `foundation/bootstrap` was at documentation-only head `a5dd3e60ec3806bc6f1416513d0ec81d6ebf0029`; the last behavior-changing code commit was `beab78a`. The 19 commits between them changed only planning/progress/architecture documentation.
- Added `O_NONBLOCK` to the Linux no-follow identity leaf open. Without it, a source leaf swapped to a FIFO after the earlier move preflight could block in `openat2` before the regular-file check. The flag has no effect on normal regular-file reads.
- Added `TestReadRestoreContentIdentityRejectsFIFOWithoutBlocking`, which creates a FIFO and asserts the identity reader returns an error promptly rather than hanging.
- Commit: [45b9ac80d701c888a2c1eb07f8ed0f297aa7307b](https://github.com/SoufianoDev/OhMyFreeAI/commit/45b9ac80d701c888a2c1eb07f8ed0f297aa7307b). Only the Linux identity reader and its Linux-specific regression test changed in that code commit.

### CI verification for the exact code commit
- Root CI [42913865169](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42913865169): Ubuntu and Windows 2/2 SUCCESS. Both job logs show checkout of `45b9ac80d701c888a2c1eb07f8ed0f297aa7307b`. Ubuntu passed `go vet ./...`, `go test ./...`, `go test -race ./...`, build, Python compile/tests, frontend syntax and architecture checks. Windows passed vet, Go tests/build, Python compile/tests, frontend syntax and architecture checks; the Linux-only race step was skipped as designed.
- SQLite driver evaluation [42913865376](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42913865376): all four modernc/ncruces × Ubuntu/Windows jobs SUCCESS.
- Wails proof [42913865308](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42913865308): attempt 1 had one Ubuntu-only environment failure at the WebView smoke step because Xvfb did not allocate a display within 10 seconds. The failed Ubuntu job was rerun without a code change; attempt 2 passed both Ubuntu and Windows. The Ubuntu logs show CLI smoke success, two GetSnapshot bridge calls on each of two application launches, and `WEBVIEW_REOPEN_PASS`. Windows native build, CLI smoke and executable verification passed.
- Linux distribution compatibility [42913865277](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42913865277): 8/8 jobs SUCCESS, including the Ubuntu 22.04 legacy GTK3/WebKit41 job.

### Windows namespace durability contract correction
- Re-reviewed Microsoft's [CreateFileW write-through documentation](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-createfilew). It explicitly states that a write-through request causes NTFS to flush metadata changes, including rename operations.
- The native [NtCreateFile documentation](https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntcreatefile) defines native `FILE_WRITE_THROUGH` as CreateOptions bit `0x00000002`. The current Windows adapter opens the source relative to the retained directory handle with that option and uses the same handle for handle-bound `NtSetInformationFile(FileRenameInformationEx=65)`.
- This resolves the documented NTFS namespace-metadata barrier under the current fixed-local-NTFS policy. It is not hardware-independent power-loss certification and does not qualify other providers. The architecture document now records the superseding evidence.
- The independent Windows journal rollover blocker remains: terminal journals cannot be durably removed, and the single fixed filename prevents another transaction. Next design a retained multi-transaction journal with crash-safe discovery and no deletion of evidence.

### Remaining safety state
- The no-follow identity reader and verified no-replace move helper are tested, but no transaction executor invokes them yet.
- The journal lifecycle, filesystem executor, durable intent/mutation orchestration, rollback/quarantine execution, startup recovery coordinator and process-death/restart matrix remain incomplete.
- Restore remains disabled. No GUI/CLI Restore, startup recovery wiring or merge to `main` was added.


## 2026-11-11 06:55Z — retained Restore journal rollover regression coverage

### Repository and plan review
- Rechecked `foundation/bootstrap` at `c464d7b767db049de15c5724b13bcb3dc5685bad`; its parent production-code head is `430196b80b3342b7e9d300d6894759e853d5e97a`.
- Reviewed the active task plan, progress history, Restore lifecycle, executor constraints, StoreController, journal v3 writer/reader, ledger validation, recovery planner, no-replace move adapters and identity readers.
- Confirmed there were no active workflows before the test change. The four workflow families for production commit `430196b` had all passed 2/2, 4/4, 2/2 and 8/8.
- Found that the new ledger implementation had tests for rejecting a second active transaction and malformed multi-active records, but lacked a positive sequential rollover test proving that a second transaction can start after the first reaches a terminal phase.

### Change
- Added `TestRestoreJournalLedgerRetainsSequentialTerminalTransactions`: complete one transaction, append a second, verify the second becomes active, complete it, verify both terminal records remain ordered and retained, and verify the removal API cannot delete them.
- Added `TestRestoreJournalLedgerIgnoresAndPreservesUnpublishedTemporaryFile`: simulate an orphan temporary journal from a crash before publication; the reader continues to use the last published ledger and leaves the orphan evidence untouched.
- Commit: [c464d7b767db049de15c5724b13bcb3dc5685bad](https://github.com/SoufianoDev/OhMyFreeAI/commit/c464d7b767db049de15c5724b13bcb3dc5685bad). Compare against `430196b` confirms only `internal/infrastructure/sqlite/restore_journal_ledger_test.go` changed (+96 lines); no production source or `main` files changed.

### Exact-head CI verification
- Root CI [42975743863](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42975743863): Ubuntu and Windows 2/2 SUCCESS. Both job logs confirm checkout of `c464d7b`. `go test ./...` passed on both platforms; Ubuntu `go test -race ./...` passed. Go vet/build, Python compile/tests, frontend syntax and architecture checks passed. Windows skipped only the Linux-only race step.
- SQLite driver evaluation [42975743844](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42975743844): 4/4 SUCCESS (modernc/ncruces × Ubuntu/Windows).
- Wails platform proof [42975743834](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42975743834): 2/2 SUCCESS. Ubuntu native build, CLI smoke, live WebView-to-Go bridge and database reopen passed; Windows native build, CLI dispatch and executable verification passed.
- Linux distribution compatibility [42975743861](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42975743861): 8/8 SUCCESS, including Ubuntu 22.04 and Debian 12 legacy GTK3/WebKit41.
- All 16 jobs completed successfully. No local test execution is claimed; verification is from the exact-head GitHub Actions runs.

### Safety and next step
- The retained ledger resolves the previous single-terminal-journal rollover blocker without deleting history. It is bounded to 256 records / 1 MiB and intentionally fails closed at the limit; no automatic pruning/archive path exists.
- The orphan-temp fixture covers the pre-publication crash shape only. It does not establish the full journal publication, process-death, move, rollback or restart matrix.
- Restore remains disabled and is not wired to startup, GUI or CLI. The legacy `backup.Manager.Restore` still routes through path-based `sqlite.RestoreDatabase`; the new executor must not call or expose that path.
- Branch comparison against `main`: 832 ahead, 0 behind. No open PR; no merge attempted.
- Next: add a staging-only candidate builder that validates and closes the candidate without touching the live Store, then implement the isolated StoreController-backed executor and its deterministic Linux/Windows failure-injection/process-death matrix.


## 2026-11-11 07:20Z — isolated Restore candidate builder and Root CI syntax failure

### Repository and architecture review
- Rechecked `foundation/bootstrap` at `b4c845e23f6864e712572c101f4b21dd9221fcec`; the branch ref still pointed to this exact head before the repair.
- Reviewed the current task plan, progress history, Restore lifecycle and executor design, StoreController lease semantics, retained journal ledger, recovery planner, no-replace adapters and identity readers.
- Confirmed the new candidate-builder change is limited to `internal/infrastructure/backup/archive.go`, `archive_test.go`, `internal/infrastructure/sqlite/restore.go` and `restore_candidate_test.go`. It adds no GUI/CLI/startup wiring and does not enable Restore.

### Exact-head CI diagnosis
- Root CI [42978165926](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42978165926) failed on both Ubuntu and Windows in `go vet ./...`. The job logs identify the same two Go parser errors: an unterminated string at `archive_test.go:486` and another at `restore_candidate_test.go:86`. Subsequent test/build steps were skipped. This is a real source syntax defect in the newly added tests.
- SQLite driver evaluation [42978165940](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42978165940) passed 4/4.
- Wails platform proof [42978165989](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42978165989) passed 2/2.
- Linux distribution compatibility [42978166212](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42978166212) passed 8/8 after all eight jobs completed. No workflow remained active before the repair commit.

### Repair and verification boundary
- Corrected the two affected test files by expressing multiline diagnostic messages with escaped newline sequences. The archive test contains two affected messages; the SQLite candidate test contains one.
- Updated the active task plan and Restore lifecycle architecture to record the implemented staging-only candidate builder and to keep the filesystem transaction executor as the next engineering unit.
- This commit contains the source-test repair and documentation updates. Exact-head CI for this repair is pending; no test or workflow success is claimed here.
- Restore remains disabled and unexposed. No transaction executor, rollback orchestration, startup recovery, GUI/CLI Restore, or merge to `main` was added.


## 2026-11-11 14:39Z — requested CI audit and StoreController gate reconciliation

### Requested workflow

- Inspected Root CI run [42419106009](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42419106009) for commit `f5396ee592b9b9756779f67aeabd6d7167520c40`.
- The run was already completed, not queued or running. Its two jobs passed: `Validate (ubuntu-latest)` and `Validate (windows-latest)`. Each completed Go vet, Go tests/build, Python compile/tests, frontend JavaScript syntax and architecture-boundary checks. No failure logs required repair for this requested run.

### Current branch and subsequent CI

- Current branch head before planning updates: `eedeac445cc9259b743c3d87c5d1cce7f2ead1aa`.
- Audited the intervening Root CI failure [42984745263](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42984745263) on commit `4075baf`. Ubuntu passed. Windows failed three candidate-path test assertions because Windows returned a long path while the expected value used the equivalent 8.3 short-path alias. The logs locate the failures in backup and SQLite candidate tests; production behavior was not shown to fail.
- Commit [eedeac4](https://github.com/SoufianoDev/OhMyFreeAI/commit/eedeac445cc9259b743c3d87c5d1cce7f2ead1aa) changed those assertions to compare canonical path equivalence with `sameArchivePath` / `samePath`. This is a test-only correction.
- Verified all workflows on exact head `eedeac4`: Root CI [42985836167] 2/2; SQLite driver matrix [42985836284] 4/4; Wails platform proof [42985836260] 2/2; Linux distribution matrix [42985836201] 8/8. All 16 jobs succeeded. Root Ubuntu passed the race detector; Windows skipped only the Linux-only race step.
- Wails jobs verified the native application build and platform CLI/executable checks; Ubuntu also passed the live WebView-to-Go bridge and database-reopen smoke test. The Linux distro matrix passed all eight configured distributions.

### StoreController review

- Re-read `store_controller.go`, all 16 StoreController tests, Wails `main.go.tmpl` and `app_service.go.tmpl`, the configuration service and backup controller tests.
- Confirmed the Wails composition root owns one StoreController for the application lifetime and closes it after `app.Run` using a bounded context.
- Confirmed AppService and config use the controller; config Load/Save run under shared leases. Wails and CLI backup export use `CreateWithStoreController`, whose shared lease remains held through archive verification and publication.
- Confirmed exclusive leases drain readers, waiting writers prevent reader starvation, shutdown waits for active operations, cancellation does not invalidate a live Store, stale/closed Stores are not re-exposed, and failed destructive replacement fails closed unless a verified replacement Store is supplied.
- The backup-specific test holds the publication barrier and proves controller Close times out while export is active, then succeeds after publication completes. Current-head Linux and Windows CI passed these tests.

### Next gate and safety

- The StoreController integration gate is already complete; it was not reimplemented in this run.
- The branch already contains the low-level retained Restore journal ledger, recovery planner, platform no-replace move and no-follow identity primitives, plus the staging-only candidate builder. The candidate builder uses an exclusive destination create, returns the measured content identity, and preserves partial artifacts on failure.
- The next engineering unit remains the isolated StoreController-backed transaction executor with intent-before-mutation journal updates, immediate post-move no-follow identity checks, verified reopen, rollback/quarantine retention and deterministic Linux/Windows process-death tests.
- Restore remains disabled and unexposed. No executor, startup recovery, GUI/CLI Restore, or merge to `main` was started. No production code changed in this run.
- Compare against `main`: 838 commits ahead, 0 behind. No merge or force-push was attempted.
- No local tests are claimed; all verification cited above is from inspected GitHub Actions jobs and logs.



## 2026-11-11 21:35Z — SQLite PERSIST journal alternative research

### Repository and CI audit
- Confirmed the current `foundation/bootstrap` source file `internal/infrastructure/sqlite/restore.go` has the same blob as commit [eedeac4](https://github.com/SoufianoDev/OhMyFreeAI/commit/eedeac445cc9259b743c3d87c5d1cce7f2ead1aa). The last verified behavior-changing code commit remains `eedeac4`; the current Restore executor is still absent.
- Re-inspected the four workflow families associated with `eedeac4`: Root CI [42985836167](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42985836167) 2/2; SQLite drivers [42985836284](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42985836284) 4/4; Wails [42985836260](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42985836260) 2/2; Linux distributions [42985836201](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42985836201) 8/8. All 16 jobs completed successfully.
- Root job steps/logs confirm vet, Go tests/build, Linux race detector, Python compile/tests, frontend syntax and architecture checks. Wails logs confirm native builds and CLI smoke on both platforms plus live WebView bridge/database reopen on Linux. Windows skipped only the Linux-only race step and Linux-only WebView smoke.

### New journal-storage candidate — research only, not approved
- Reviewed SQLite's primary [journal_mode documentation](https://sqlite.org/pragma.html#pragma_journal_mode), [atomic commit design](https://sqlite.org/atomiccommit.html), [WAL documentation](https://sqlite.org/wal.html), [corruption guidance](https://www.sqlite.org/howtocorrupt.html), [Online Backup API](https://www.sqlite.org/c3ref/backup_finish.html), and the current SQLite Windows VFS source `src/os_win.c`.
- SQLite `PERSIST` mode does not delete the rollback journal after each transaction; it overwrites the journal header with zeroes. This could avoid per-phase JSON-file replacement for an independent restore-state database.
- Important constraint: `PERSIST` is connection-scoped and is not retained after close/reopen. Only WAL mode is documented as persistent across reopen. A dedicated restore-journal connection must explicitly set and verify `PRAGMA journal_mode=PERSIST` before every transaction on every new connection; do not assume it remains set.
- The Windows SQLite VFS `winSync` flushes the specified file handle with `FlushFileBuffers`. The inspected `winOpen` path does not request `FILE_FLAG_WRITE_THROUGH`, and the VFS source does not provide a parent-directory sync equivalent to Linux directory-FD `fsync`. Therefore PERSIST alone does not solve durable creation of the database or its initial `-journal` directory entries.
- Candidate design to test: a separate, pre-provisioned `restore-journal.sqlite` containing retained transaction/event records, using rollback-journal PERSIST and `synchronous=FULL`; create and durably publish the DB and its persistent rollback-journal sidecar before any restore transaction can begin. Linux must sync the containing directory. Windows provisioning must use the already-qualified handle-bound NTFS write-through creation path, then SQLite can reuse the existing sidecar for subsequent commits. If either file is missing or fails validation when a restore is requested, fail closed before touching the live Store.
- Keep this separate from the application database being restored. The main database and its WAL/SHM lifecycle must not be confused with the restore-control database.
- A local exploratory Python 3.13 / SQLite 3.46.1 probe confirmed that the PERSIST journal file remains after commits while PERSIST is active, and that a reopened connection reports DELETE until PERSIST is explicitly re-enabled. This is a behavior probe only; it is not a Go-driver, Windows, process-death, or power-loss test.
- This direction is not approved for implementation yet. Required next evidence: test both modernc and ncruces, prove journal-sidecar pre-provisioning and missing-sidecar fail-closed behavior on Linux/Windows, exercise abrupt process termination at each journal transaction boundary, and inspect SQLite recovery before the next restore phase. No inference from process-death tests to power-loss certification.

### Decision and safety
- Do not replace the current JSON ledger yet. Do not begin the filesystem executor until the journal publication contract is settled.
- No production code changed in this review. Restore remains disabled and unexposed; no startup recovery, GUI/CLI Restore, or merge to `main`.


## 2026-11-12 06:50Z — PERSIST journal lifecycle and abrupt-process-death evaluation

### Repository and workflow audit
- Rechecked foundation/bootstrap, main, recent commits, the active plan, progress history, Restore lifecycle, executor constraints, journal writer/ledger, no-replace move adapters, no-follow identity readers, candidate builder and StoreController.
- Branch head before the documentation updates: 01913ec7fdbf63134739a21136f82540f1e66493; latest production-code head: 1cde3c316fbf11271fe427c9c03e6ea67c380f47.
- Compared against main: 851 commits ahead, 0 behind. No merge or force-push was attempted.
- Audited the failed Root CI run 43167168518 on 590a67d. Ubuntu and Windows both failed at go vet because the new test passed an expected identity to a helper that still accepted only two arguments. Commit 1cde3c corrected the helper to call moveRestoreFileNoReplaceVerified. The exact code head passed all four workflow families: Root 2/2, SQLite 4/4, Wails 2/2 and Linux distributions 8/8.
- The current test-only head 01913ec passed Root CI [43170647066](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43170647066) 2/2 and SQLite evaluation [43170647220](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43170647220) 4/4. Wails and distro workflows were not triggered by changes confined to tests/sqlite-driver-eval; their latest exact production-code runs remain the successful 1cde3c runs.

### Changes made
- Added tests/sqlite-driver-eval/restore_journal_persist_test.go to evaluate PERSIST mode, synchronous FULL, sidecar header zeroing, persistence across close/reopen, explicit mode reapplication, sidecar identity reuse, and the driver's behavior when the DB or sidecar is missing.
- Added tests/sqlite-driver-eval/restore_journal_crash_test.go. A child process starts a large uncommitted transaction, confirms a live rollback-journal header, then exits with os.Exit without SQLite cleanup. The parent reopens the database and verifies that the committed baseline remains, uncommitted rows are absent, and PRAGMA integrity_check returns ok.
- Test commits: [3142c8f](https://github.com/SoufianoDev/OhMyFreeAI/commit/3142c8fce61ee05e073e5bedc2567087aa1c8817) and [01913ec](https://github.com/SoufianoDev/OhMyFreeAI/commit/01913ec7fdbf63134739a21136f82540f1e66493). No production Restore code was changed in these commits.
- Updated the active task plan to replace stale CI/commit information and record the new evidence and remaining gate.

### PERSIST findings — all four driver/OS combinations passed
- modernc v1.62.0 and ncruces v0.37.0 both accept PRAGMA journal_mode=PERSIST and PRAGMA synchronous=FULL on Ubuntu 24.04 and Windows.
- After a committed transaction, both drivers retain a regular rollback-journal sidecar with a zeroed header. The sidecar survives close/reopen. Journal mode resets to DELETE on a fresh connection, so PERSIST must be explicitly set and verified again before writes. Reapplying PERSIST and committing another transaction reused the same sidecar file identity in all four combinations.
- The forced process-death test passed on modernc/Ubuntu, ncruces/Ubuntu, modernc/Windows and ncruces/Windows. SQLite recovered the uncommitted transaction and passed integrity_check.
- Critical limitation: normal SQLite open/use silently creates a missing database and silently recreates a missing PERSIST sidecar. The driver does not fail closed. A production design must validate both pre-provisioned artifacts through no-follow handles and open existing files only before touching the live Store.
- These are driver and process-death results only. They do not prove durable initial DB/sidecar provisioning or hardware-independent power-loss safety. PERSIST remains a candidate, not an approved replacement for the retained JSON ledger.

### Safety and next step
- The JSON ledger remains the production journal. It is retained, bounded to 256 records / 1 MiB, and uses platform-specific publication. Initial publication now uses no-replace plus post-move no-follow identity verification. The PERSIST candidate has not yet been shown strictly stronger.
- Next: prototype the no-follow fail-closed preflight/open-existing boundary and durable DB/sidecar pre-provisioning, then compare PERSIST against JSON including concurrent-process serialization and recovery. Do not start the transaction executor until the journal protocol is explicitly selected and qualified.
- Restore remains disabled and is not connected to startup recovery, GUI or CLI. No main changes, merge, startup recovery or executor implementation occurred.
- No local test execution is claimed; the reported results are from the exact-head GitHub Actions job logs.

## 2026-11-12 14:31Z — StoreController gate recheck and existing-only journal DB probe

### Requested CI and StoreController gate
- Rechecked the specifically requested Root CI [42419106009](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42419106009) on `f5396ee` through GitHub job details. Ubuntu and Windows both completed successfully (2/2).
- Re-reviewed the current `StoreController`, its 16 tests, Wails composition root, AppService, configuration adapter and backup integration. The gate is already complete: Wails creates one application-lifetime controller and closes it with a bounded shutdown context; AppService and config use the controller; controller Load/Save hold shared leases; Wails/CLI backup export uses `CreateWithStoreController`; `archive_controller_test.go` verifies Close waits for archive publication and that a timed-out close does not discard the active Store. No StoreController production code needed to be repeated or changed.
- The 16 controller tests cover reader/writer drain and priority, cancellation, replacement publication, stale/unavailable Store behavior, panic cleanup, concurrent stress, config round-trip, and config Load/Save drain. The separate backup test covers the publication barrier.

### Isolated SQLite existing-only open probe
- Added `TestSQLiteExistingOnlyOpenRequiresExistingDatabase` in `tests/sqlite-driver-eval/restore_journal_persist_test.go`. It verifies a missing control DB is rejected without being created and an existing DB can be opened without CREATE.
- The probe passes for modernc v1.62.0 on Ubuntu/Windows and ncruces v0.37.0 on Ubuntu. On Windows, the pinned ncruces VFS mishandles an absolute `file:///C:/...` URI, so the evaluation uses its direct `OpenFlags(OPEN_READWRITE)` API for the same existing-only guarantee. This is a driver-specific Windows URI-path limitation; do not claim that ncruces `database/sql` URI mode=rw was verified on Windows.
- The existing `TestSQLiteSilentlyRecreatesMissingRestoreControlArtifacts` also passes in all four driver/OS combinations. It demonstrates that existing-only opening of the main database does not protect a missing PERSIST sidecar: SQLite can recreate that sidecar on a later write. The DB and sidecar therefore need separate no-follow preflight and identity revalidation before any restore journal write.
- This probe does not establish race-free preflight, durable initial DB/sidecar provisioning, cross-process serialization, or power-loss safety. PERSIST remains an evaluation candidate; the retained JSON ledger remains production.

### CI and corrections
- Latest test-only commit: [5f679e5](https://github.com/SoufianoDev/OhMyFreeAI/commit/5f679e5645a4893e22b5fc6ff1e2e6e9900e70cc).
- Root CI [43237308026](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43237308026): Ubuntu and Windows SUCCESS (2/2). Logs confirm Go vet/tests/build and the Ubuntu race detector.
- SQLite driver evaluation [43237307833](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43237307833): modernc/ncruces × Ubuntu/Windows SUCCESS (4/4). Job logs explicitly show the new existing-only probe and the PERSIST/process-death tests passing.
- Earlier evaluation attempts failed and were corrected rather than hidden: the first test commit missed imports; the next exposed the ncruces Windows absolute-URI path issue; the relative-URI attempt had a fixture-scope error and was rejected by the pinned SQLite URI/VFS behavior. The final test uses the supported native existing-only API on ncruces Windows and passed the full matrix.
- The task also rechecked Root CI [42419106009](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42419106009): 2/2 success on the requested historical commit.
- Latest production-code head remains `1cde3c316fbf11271fe427c9c03e6ea67c380f47`; no production Restore code changed in this run. Wails and Linux-distro workflows were not triggered by evaluation-only changes; their latest exact production-code runs remain the successful `1cde3c` evidence recorded above.
- Current code/test head `5f679e5` compared 861 commits ahead of `main`, 0 behind. The only later commit in this run is the planning update. No `main` changes, merge or force-push.

### Next gate
- Continue the journal qualification, not the filesystem executor: design and test no-follow preflight/open-existing for both the control DB and its PERSIST sidecar, verify identities around SQLite open before any write, and prove durable pre-provisioning on Linux and fixed-local NTFS.
- The current JSON ledger read-modify-publish path has no explicit cross-process lock acquisition in `WriteRestoreJournal`. Resolve and test this serialization contract as part of the PERSIST-versus-JSON comparison before selecting the journal protocol.
- Do not start the transaction executor, startup recovery, GUI/CLI Restore or any Restore wiring until the journal protocol and Linux/Windows safety gates are explicitly qualified. Restore remains disabled.

## 2026-11-12 14:36Z — historical CI, StoreController gate, and journal boundary review

### CI verification
- Rechecked the explicitly requested historical Root CI [42419106009](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42419106009) on `f5396ee`: Ubuntu and Windows jobs completed successfully (2/2); all setup, vet, Go test/build, Python, frontend syntax, and architecture steps succeeded.
- Rechecked Root CI [43237308026](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43237308026) on test-only head `5f679e5`: Ubuntu and Windows succeeded (2/2). Ubuntu logs show vet, Go tests, race detector and build; Windows logs show vet, Go tests and build.
- Rechecked SQLite driver evaluation [43237307833](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43237307833): modernc/ncruces × Ubuntu/Windows succeeded (4/4). Retrieved logs explicitly show the PERSIST lifecycle, abrupt-process-death recovery, missing-sidecar recreation and existing-only main DB probe passing in every combination.
- Rechecked the production-code head `1cde3c` separately: Wails [43168230458](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43168230458) succeeded 2/2; Linux distribution compatibility [43168230439](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43168230439) succeeded 8/8. These Wails/distro runs are not evidence for test-only head `5f679e5`.
- The branch has two commits after `5f679e5` according to compare; both changed only `.planning/.../progress.md` and `.planning/.../task_plan.md`. No production/test changes were observed after the verified test head.

### StoreController gate re-review
- Counted and inspected all 16 `TestStoreController*` tests. They cover exclusive drain, writer preference, canceled waits, nil Store rejection, replacement publication, close/drain, fail-closed replacement, panic handling, concurrent lease stress, config round-trip and config Load/Save drain.
- Rechecked Wails composition root and AppService: one application-lifetime controller is created and closed through the controller; AppService retains the controller; configuration uses the controller-backed `ports.ConfigRepository`.
- Rechecked backup `CreateWithStoreController`: the shared lease encloses the complete archive creation/publication path. This integration stage is complete; no code was duplicated or changed.

### Journal qualification finding
- SQLite's official `sqlite3_open_v2` documentation describes `SQLITE_OPEN_NOFOLLOW` as a restriction on the database filename and `mode=rw` as existing-database read/write without CREATE: https://sqlite.org/c3ref/open.html. Neither contract documents a no-follow guarantee for the separate PERSIST rollback-journal sidecar.
- The exact-head evaluation logs confirm `TestSQLiteSilentlyRecreatesMissingRestoreControlArtifacts` passes on modernc and ncruces on both platforms: SQLite recreates a missing PERSIST sidecar on a later write. Main database existing-only open is therefore not enough.
- The retained JSON ledger has a separate unresolved issue: `WriteRestoreJournal` performs a read-modify-publish sequence without explicit cross-process serialization. This must be fixed and tested before choosing JSON as the final protocol.
- Windows journal namespace durability remains unresolved; do not infer power-loss safety from CI or process-death tests.

### Outcome and next gate
- No production code changed and no local test execution is claimed. Evidence above comes from GitHub Actions job status and retrieved job logs.
- Continue with an isolated, test-first comparison: (1) establish whether either pinned SQLite driver can enforce no-follow for both main DB and journal sidecar through a supported VFS/open boundary; (2) design cross-process serialization for the JSON ledger and test it with separate processes on Linux and Windows; (3) settle the Windows namespace-durability contract. Select the journal protocol only after these gates pass.
- Do not start the filesystem transaction executor, startup recovery, GUI/CLI Restore, or main merge. Restore remains disabled.

## 2026-11-13 14:04Z — requested CI, StoreController recheck, and journal lease review

### Requested historical workflow
- Rechecked Root CI [42419106009](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42419106009) on commit `f5396ee592b9b9756779f67aeabd6d7167520c40` through GitHub job details.
- Ubuntu and Windows jobs both completed successfully (2/2). Both jobs completed module verification, Go vet/tests/build, Python compile/tests, frontend syntax and architecture checks. No failure logs were applicable.

### Current branch head and CI
- Resolved `foundation/bootstrap` through the GitHub commit endpoint: `21fdfb44bfc9da20bb94777092dcb78ee93b0e7e`, message `fix(sqlite): declare restore journal write error`.
- The latest commit fixes the short variable declaration in `writeRestoreJournalLocked` and adds a lease read/write/closed-state lifecycle test.
- All four workflows for this exact SHA completed successfully:
  - Root CI [43377131882](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43377131882): Ubuntu + Windows, 2/2. Job logs show `internal/infrastructure/sqlite` passed on both platforms; Ubuntu also passed `go test -race ./...`.
  - SQLite driver evaluation [43377131804](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43377131804): modernc/ncruces × Ubuntu/Windows, 4/4. Logs show production migration, PERSIST lifecycle, abrupt-process-death recovery, missing-control-artifact recreation and existing-only-open probes passed.
  - Wails platform proof [43377131897](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43377131897): Ubuntu + Windows, 2/2. Logs include CLI smoke checks; Ubuntu also passed the WebView-to-Go bridge and database reopen checks.
  - Linux distribution compatibility [43377131790](https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/43377131790): 8/8 across Debian 12/13, Ubuntu 22.04/24.04, Fedora 44 GTK3/GTK4, Arch rolling and openSUSE Tumbleweed.
- These results are attached to the exact current branch SHA, not inferred from older runs.

### StoreController gate re-review
- Re-fetched the current Wails composition root, AppService, controller and tests from `foundation/bootstrap`.
- The integration remains complete: one application-lifetime controller is retained by the composition root; AppService uses it; configuration Load/Save use its shared leases; Wails and CLI backup export use `CreateWithStoreController` through archive publication; shutdown drains through controller Close.
- Counted 16 `TestStoreController*` tests covering exclusive drain, writer preference, cancellation, replacement publication, stale/unavailable Store rejection, panic cleanup, concurrent leases and config Load/Save drain. `TestManagerCreateWithStoreControllerDrainsBeforeClose` covers the backup publication barrier.
- No StoreController code was duplicated or changed.

### Cross-process JSON ledger serialization — implementation now present
- Current source contains `restoreJournalLease` and the persistent `.restore-journal.lock` artifact.
- Linux acquires the lease with `flock` on a no-follow lock file opened relative to a pinned private directory handle.
- Windows acquires the lease with `LockFileEx` on a handle opened relative to the validated directory handle; reparse points and unexpected lock-file contents are rejected.
- `TestWriteRestoreJournalSerializesAcrossProcesses` starts a child process that holds the lease, verifies another process and `WriteRestoreJournal` wait, then verifies the write proceeds after release. Additional tests reject symlink/non-empty lock files and use-after-close.
- Root CI on the current SHA confirms the SQLite package tests passed on Ubuntu and Windows; Ubuntu race detection also passed.
- This supersedes the 2026-11-12 note that the JSON ledger had no explicit cross-process lock. The executor is not implemented yet: when added, it must hold one `restoreJournalLease` for the entire transaction and use that lease's Read/Write methods. The convenience `WriteRestoreJournal` acquires a lease per call and is not a substitute for a transaction-long lease.

### Remaining blockers and scope
- The PERSIST evaluation still proves SQLite can recreate a missing rollback-journal sidecar. Existing-only opening of the main DB does not qualify the sidecar; no-follow preflight/open and identity revalidation for both artifacts remain unresolved.
- Windows namespace-entry power-loss durability is still not established by the available rename/write-through contract. Process-death tests are not power-loss tests.
- No filesystem transaction executor, startup recovery, GUI/CLI Restore or Restore enablement was added.
- No production code changed and no local test execution is claimed in this review. Evidence is from GitHub Actions jobs/logs and current branch source.
- Branch comparison: `foundation/bootstrap` is 867 commits ahead and 0 behind `main`. No main changes, merge or force-push.

### Next gate
- Continue journal qualification: establish a supported no-follow boundary for both the control DB and PERSIST sidecar, and settle Windows journal namespace durability. Only after those platform gates pass may the transaction executor be started. Restore remains disabled.
