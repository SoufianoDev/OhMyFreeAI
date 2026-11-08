# Findings & Decisions

## Requirements

- Build OhMyFreeAI as a free-first desktop application; do not propose paid services or alternate paid hosting.
- Go core; Wails remains a candidate subject to proof-of-concept on Linux and Windows.
- Web UI uses HTML/CSS/JavaScript ES Modules and Material Design 3; calm, professional, accessible, bilingual Arabic/English, RTL/LTR.
- SQLite persistence; modular monolith with Clean Architecture and explicit ports/adapters.
- Google Colab Free is the primary remote runtime; llama.cpp is the primary inference engine; OpenCode is the first client.
- Local OpenAI-compatible Gateway defaults to 127.0.0.1:18081; upstream llama-server previously used 8081.
- Google Drive is persistent artifact storage only; stage files to Colab local disk and verify before execution.
- Model compatibility must include architecture/checkpoint/conversion/format/features/backend/runtime and actual inference verification.
- Colab compatibility must include actual GPU/VRAM, RAM, CUDA/driver, disk, quantization, context/KV cache, Flash Attention, GPU layers and overhead.
- UI must distinguish supported, unsupported, and unverified models and explain minimum resources.
- Backup artifact name/extension: omfaibackup; backups contain app config and recoverable metadata/references, not large model/engine artifacts, secrets, or live processes.
- Preserve prior successful Colab Free + llama.cpp + Gateway + OpenCode path until audited.
- Do not claim runtime or model support based on static file presence or compilation alone.

## Research Findings

- Target repository `SoufianoDev/OhMyFreeAI` default branch is `main`.
- The repository's current README.md contains only the Markdown heading `# OhMyFreeAI`.
- GitHub code searches for `task_plan.md`, `.gitignore`, `go.mod`, `package.json`, and `PROJECT_SPEC.md` returned no results. No existing implementation was surfaced by these searches.
- A working branch `foundation/bootstrap` was created from `main`.
- Planning with Files source repository exposes the skill at `skills/planning-with-files/SKILL.md`; fetched metadata reports version 3.28.1.
- Planning with Files requires persistent task_plan.md, findings.md, progress.md; new plans should be initialized from its script/template and validated with check-complete.sh --validate.
- The skill's installer requires sibling scripts including set-active-plan.sh. The container cannot resolve github.com, so cloning and executing the upstream initializer from a local checkout is currently unavailable.
- The supplied archives contain two Wails v2.10.2 desktop-monitor prototypes using Go 1.24 and embedded HTML/CSS/JS. The remote-complete variant includes local process/RAM/file monitoring and a remote Colab probe; the smaller variant is local-only.
- Both prototype frontends are distinct from a production app architecture. The remote-complete CSS contains a broad dark teal/green palette; this is not accepted as the new product design and must not be copied as-is.
- The uploaded Python scripts all parse successfully with Python AST. This is syntax-only validation; none was executed because they mutate a live launcher, stop/restart a Gateway, or depend on the user's home/runtime state.
- The launcher repair script changes context to 262144 and KV cache to q8_0, adds a TCP listener check, validates model metadata and runtime readiness, and patches launcher configuration. It remains a historical patch requiring source-level review before any reuse.
- The SSE repair script backs up and patches the persistent launcher and installed Gateway, terminates matching Gateway processes, restarts the Gateway, and tests SSE framing. It is not safe to run blindly against a live environment.
- The monitor script polls local processes/Gateway/OpenCode and Colab telemetry, defers remote probes while a Colab exec is active, and appends JSONL snapshots. Its hard-coded session is qwen-opencode and its endpoint assumptions require validation.
- The user's Material Design 3 skill repository is confirmed as SoufianoDev/material-3-skill. Its guidance explicitly states web is a limited path: Material Web is in maintenance mode and M3 Expressive is not implemented on Web. Therefore use MD3 design tokens, semantic HTML, custom CSS, responsive layouts and accessibility; do not pretend Wails web UI has full Compose Expressive parity.
- Material skill reference highlights: semantic color roles/tonal surface hierarchy, typography scale, 4dp spacing, shape scale, restrained elevation (tonal surfaces first), reduced-motion-aware transitions, adaptive navigation.
- KotlinComposeKit is confirmed as SoufianoDev/KotlinComposeKit. The compose-component-design skill's transferable principle is caller-composable reusable components: fixed invariant structure, explicit variable content slots, caller-owned placement, no hidden Boolean switches. Compose APIs and Kotlin-specific patterns are not to be copied into the Wails frontend.
- The two zip archives and three Python scripts were inspected; they remain historical inputs, not verified production code.

## Technical Decisions

| Decision | Rationale |
|----------|-----------|
| Repository bootstrap on foundation/bootstrap | Keep work reviewable and avoid direct changes to main |
| Planning artifacts use .planning/2026-11-07-ohmyfreeai-foundation | Isolate this task's durable plan from later tasks |
| Do not force Kotlin Compose into Wails | Use KotlinComposeKit only for transferable design-system guidance |
| Do not mark a model compatible from family name alone | Compatibility is checkpoint and execution-path specific |
| Separate theoretical support from verified support | Avoid misleading model catalog claims |
| Preserve backup legacy import path where safe | Prevent loss of users' older sdevbackup backups |

## Issues Encountered

| Issue | Resolution |
|-------|------------|
| Container cannot resolve github.com for git clone | Use GitHub connector for source control operations; document local execution limitation |

## Resources

- Project: https://github.com/SoufianoDev/OhMyFreeAI
- Planning with Files: https://github.com/OthmanAdi/planning-with-files
- Planning skill file: https://github.com/OthmanAdi/planning-with-files/blob/master/skills/planning-with-files/SKILL.md
- Project architecture baseline: Architecture Foundation v3.3 in project conversation context
- Uploaded historical inputs: /mnt/data/SoufianoDev-Monitor-GUI-Remote-Complete.zip, /mnt/data/SoufianoDev-Monitor-GUI.zip, /mnt/data/fix-soufianodev-free-ai.py, /mnt/data/fix-soufianodev-sse-framing.py, /mnt/data/monitor-soufianodev-free-ai.py

## Toolchain Findings

- Official Go release history lists Go 1.27.3 as the latest stable patch available on 2026-11-07 (released 2026-10-19). The project module and CI target are set to Go 1.27.3. The local container remains on Go 1.23.2, so it cannot validate the declared module toolchain directly.
- Official Wails v3.0.0 docs list v3 as stable and require Go 1.26+; Wails v3's setup wizard is experimental and primarily tested on Linux. The official JavaScript template is `vanilla-js`; exact command: `wails3 init -n myapp -t vanilla-js`.

## Engineering Skills Applied

- Planning with Files v3.28.1 was vendored with its shell/PowerShell helpers, templates, plan inventory and hook dependencies. The active plan pointer is .planning/.active_plan. Native validation scripts are not executable from this connector-only environment; the plan structure was manually checked and the status-marker issue was corrected.
- Read and vendored the relevant ECC MIT-licensed skills: Go patterns, Go testing, database migrations, security review and TDD workflow. The original MIT license is retained under .agents/skills/ecc/LICENSE.
- Applied Go patterns through small application ports, standard-library-only domain logic, typed sentinel errors and wrapped adapter errors.
- Applied Go testing and migration guidance through isolated unit tests, schema constraints, an explicit development-only down migration and immutable forward migration policy.
- Applied security review by rejecting raw Gateway tokens, restricting secret references to OS-store URI schemes, rejecting credential-bearing/query-bearing policy source URLs, and enforcing loopback-only Gateway configuration.
- Applied TDD with an actual RED/GREEN cycle for the security hardening and SHA-256 validation changes. The initial bootstrap code was implemented before its first tests; it is not retroactively claimed as TDD.
- Material Design 3 and Kotlin Compose Kit guidance were read and applied only where appropriate. Compose guidance was limited to component ownership; no Kotlin runtime was added.

## Implementation Findings

- Added a Go 1.23 framework-independent core module. The Wails dependency is intentionally not added to the core while v2/v3 selection remains open.
- Added runtime session states and an explicit transition table, immutable transition results, UTC timestamps and stale-generation rejection.
- Added a compatibility assessment that requires exact model revision, artifact checksum, engine version, backend, runtime fingerprint, integrity verification, successful engine load, health check and real inference before returning Ready.
- Added a fail-closed runtime policy gate. Remote connection is rejected unless the assessment contains an approved decision, provider, explicit scope, HTTPS source, policy revision and review timestamp.
- Added versioned application configuration defaults and validation: Gateway disabled by default, loopback-only bind, fixed default port 18081, engine port 8081, Arabic/system theme defaults, and an OS secret reference required before enabling the Gateway. Added a ConfigRepository port and application service that validates before persistence.
- Added the first application ports for policy assessment and runtime connection. No concrete Colab connector exists or is activated.
- Added `scripts/dev.py` using Python standard library only: doctor, test, build, fmt and arch-check. It does not install dependencies, use a shell, or invoke sudo.
- Added a bilingual Arabic/English static frontend shell with responsive navigation, dashboard, model catalog empty state, runtime policy state, settings, light/dark themes, RTL/LTR, semantic HTML and reduced-motion support. It deliberately displays no fabricated runtime/GPU/model data and is not yet bound to Wails or application services.
- The Material 3 visual baseline uses muted blue semantic primary colors, neutral tonal surfaces, limited outlined surfaces, 4dp-compatible spacing, restrained shape/elevation, and no gradients/glow/glass effects.
- The Compose skill is applied only as component ownership/slot guidance; no Kotlin or Compose dependencies were added.

## Local Verification

- Go 1.23.2: `go test ./...` passed for the local mirror of the committed core foundation, including runtime transition/generation, compatibility readiness and policy-gate tests.
- Go 1.23.2: `go build ./...` passed for the same local module.
- Python 3.13.5: `python3 -m py_compile scripts/dev.py` passed.
- Node 22.16.0: `node --check frontend/app.js` passed.
- Python HTMLParser: `frontend/index.html` parsed successfully; 11 unique IDs, no duplicates.
- Developer command smoke tests: `--help`, `doctor`, Go test/build delegation and architecture-check logic passed against an isolated local fixture. The full script has not been run from a checkout of the remote branch.
- Chromium/Playwright browser smoke test rendered the dashboard and exercised Models navigation, Arabic-to-English RTL/LTR switching (including active page title), Settings navigation and dark theme. Screenshots were created locally for review. The browser's file:// and localhost navigation were blocked by the container policy; embedding the source with page.set_content() worked. The preview used the local UI mirror; remote branch HTML/CSS were separately checked for required IDs and balanced braces.
- Wails proof-of-concept is blocked in this container: Go is 1.23.2 (official Wails v3 docs require Go 1.26+), neither Wails CLI is installed, GTK/WebKit development packages are absent, and Windows is unavailable. No install was attempted.

- Added initial SQLite migration `migrations/001_initial.sql` for settings, model revisions, artifacts, artifact locations, engine installations, compatibility assessments, runtime sessions, downloads, client integrations, operational events and schema history.
- Schema constraints enforce valid SHA-256 hex strings, sanitized source URI columns, positive sizes/generations, allowed state values, foreign keys, model/artifact revision identity, and the rule that a Supported assessment requires verified integrity, successful load, health and inference.
- Added Python standard-library SQLite schema tests. Six tests pass, covering table creation, foreign keys, hash validation, evidence completeness and model/artifact identity.
- Exact Git blob hashes for the committed Go core, Go tests, developer script, SQL migration and Python schema tests match the locally tested mirror. This removes the earlier source-mirror ambiguity for those files.

## Visual/Browser Findings

- The static frontend shell has been authored but not rendered in a real browser/WebView screenshot. Visual acceptance is therefore pending an actual desktop preview.


## External Technical Research (Exa, 2026-11-07)

- Current official Wails documentation presents v3.0.0 as stable for desktop (Windows/macOS/Linux), and says new projects should use v3 while v2 continues receiving critical updates. Wails v3 requires Go 1.26+, Node/npm for most frontend templates, Windows WebView2, and Linux GTK4 + WebKitGTK 6.0 by default. Ubuntu 24.04+ is the documented Linux baseline; legacy GTK3/WebKit2GTK 4.1 is temporary through v3.1.
- Official Wails v2 documentation reports v2.17.0 docs, Go 1.21+ and Linux GTK3/WebKit dependencies; the supplied prototype's v2.10.2 dependency is older than the current v2 docs and must not be treated as current baseline.
- This creates a concrete tradeoff for the architecture's Linux-first/distro-agnostic requirement: Wails v3 is the current new-project recommendation but its GTK4/WebKitGTK 6 baseline narrows older distro compatibility; Wails v2 has broader legacy stack coverage but is no longer the recommended start for new projects. A small Linux/Windows PoC is required before choosing.
- Google Colab official FAQ describes free compute including GPUs/TPUs but does not guarantee a particular GPU or stable resource allotment. Usage limits, availability and runtime duration can vary; product architecture must treat hardware/session continuity as dynamic and not promise a GPU.
- **Policy blocker:** Google's official Colab FAQ explicitly disallows from managed runtimes remote control such as SSH shells/remote desktops, bypassing the notebook UI to interact primarily via a web UI, and file hosting/media serving/other web-service offerings unrelated to interactive compute. It says these activities may be terminated at any time on the free tier without positive compute-unit balance. The planned local OpenCode → Gateway → remote llama-server serving flow may fall into these prohibited categories depending on its exact interaction pattern; no compliance claim is justified without a policy-grounded design review. The architecture must include an explicit Colab policy gate and must not build tunneling, remote-control or UI-bypass workarounds.
- The free-only constraint therefore remains a product requirement but is not proof that the intended always-on inference service can be operated on Colab Free within Google's terms. This is an unresolved feasibility risk, not something to conceal or work around.
- Upstream llama.cpp documentation advertises official prebuilt release binaries, source builds, multiple backends (CUDA, HIP, MUSA, Vulkan, SYCL), quantization formats, CPU/GPU hybrid inference, and OpenAI-compatible llama-server endpoints. These broad capabilities do not prove every model/checkpoint/backend combination; compatibility still requires exact artifact and real runtime verification.
- Google's Material Design developer site states Material Web and Android Views are in maintenance mode, while Jetpack Compose is the primary Material Android path. Material 3 design guidance can still inform web tokens/layout/accessible interaction, but @material/web must not be treated as an actively evolving Expressive component system.

- SQLite driver evaluation: `modernc.org/sqlite v1.62.0` is a pure-Go, CGO-free `database/sql` driver under BSD-3-Clause, but its current pkg.go.dev metadata requires Go 1.26. This aligns with Wails v3's Go 1.26 baseline, but cannot be built in the current Go 1.23.2 container.
- `github.com/mattn/go-sqlite3` is MIT-licensed and the long-standing C SQLite wrapper, but requires CGO_ENABLED=1 and a C compiler; its Windows build also requires a GCC toolchain. This increases build/release complexity across the target platforms.
- `github.com/ncruces/go-sqlite3 v0.37.0` is MIT-licensed and CGO-free but wraps a WASM build of SQLite via wasm2go; its documentation warns of higher per-connection memory and its custom VFS has platform-specific file-locking/WAL behavior. Do not select it without dedicated Windows concurrency tests.
- Provisional direction: isolate the driver behind `database/sql` in a SQLite adapter and evaluate `modernc.org/sqlite` first after the Go 1.26 toolchain is available. Do not add the dependency or pin a version until Linux/Windows open/migrate/concurrency/crash-recovery smoke tests pass.

### External source URLs

- Wails v3 install: https://wails.io/getting-started/installation/
- Wails v3 FAQ: https://v3.wails.io/faq/
- Wails v2 install: https://v2.wails.io/docs/gettingstarted/installation/
- Google Colab FAQ: https://research.google.com/colaboratory/faq.html
- llama.cpp upstream: https://github.com/ggml-org/llama.cpp
- Material Design developer platform overview: https://m3.material.io/develop
- Material Web roadmap: https://github.com/material-components/material-web/blob/main/docs/roadmap.md


## Wails v3 proof-of-concept research (2026-11-07)

- Re-read the official Wails v3 installation, first-application, and build documentation.
- The documented stable CLI installation is pinned to `github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0`; do not install the moving development branch.
- The official JavaScript-only starter command is `wails3 init -n <name> -t vanilla-js`; the official production build command is `wails3 build`.
- Wails v3 requires Go 1.26 or later. This repository's current Go target 1.27.3 satisfies that documented minimum, subject to the selected toolchain being available on CI runners.
- Official Linux prerequisites are GCC build tools, GTK4 and WebKitGTK 6.0. Ubuntu CI packages selected for the proof are `build-essential`, `pkg-config`, `libgtk-4-dev`, and `libwebkitgtk-6.0-dev`.
- Official Windows prerequisite is the WebView2 Runtime; the standard GitHub-hosted Windows runner is expected to provide it, but the workflow must prove this rather than infer end-user availability.
- Wails documents Ubuntu 24.04 as its supported Linux baseline and Windows AMD64/ARM64. The CI proof is therefore pinned to `ubuntu-24.04` and `windows-latest`.
- The proof deliberately generates the official starter template in a temporary CI workspace and builds it on each native OS. This verifies the pinned Wails toolchain and platform dependencies, but does **not** yet prove OhMyFreeAI service bindings, process lifecycle, packaging, or end-user distro coverage.
- Official references:
  - https://wails.io/getting-started/installation/
  - https://wails.io/getting-started/your-first-app/
  - https://wails.io/guides/build/building/


## Wails service-binding smoke test research (2026-11-07)

- Official Wails v3 method-binding documentation describes the supported flow: register an exported Go service with `application.NewService(...)`, generate bindings using `wails3 generate bindings`, then import the generated JavaScript service module and call its exported methods asynchronously.
- The official vanilla project structure includes a `GreetService` example and generated frontend bindings. The Wails proof workflow now explicitly runs the binding generator on both Ubuntu and Windows, locates the generated `*greetservice*.js` file, and asserts that the exported `Greet` method is present before the native build.
- This is a generator + generated-artifact + native-build smoke test. It is not yet an end-to-end GUI runtime test that launches the built application and executes the JavaScript-to-Go call in a live WebView. Do not describe it as such.
- The workflow update has been committed to `foundation/bootstrap`; both native jobs still require a real GitHub Actions result before this proof can be marked passed.
- Official references:
  - https://v3.wails.io/features/bindings/methods/
  - https://v3.wails.io/guides/dev/project-structure/
  - https://v3.wails.io/quick-start/first-app/


## SQLite driver re-evaluation (2026-11-07)

- Rechecked current package metadata and official project documentation rather than relying only on the earlier candidate summary.
- `modernc.org/sqlite` v1.62.0 currently declares Go 1.26.0 and BSD-3-Clause, and describes itself as a pure-Go/no-CGO SQLite driver compatible with `database/sql`. Its canonical development repository is GitLab; the GitHub repository is an official mirror. The repository publishes third-party license/SBOM materials, so release packaging still needs a dependency-license review.
- `github.com/ncruces/go-sqlite3` v0.37.0 currently declares Go 1.26.0 and MIT. It wraps a WASM SQLite build translated to Go and provides a `database/sql` driver. The project explicitly warns that per-connection memory is higher because each connection runs in a WASM sandbox.
- ncruces' VFS documentation reports tested full file-lock and shared-memory WAL support for Linux amd64 and Windows amd64 in its default configuration. It also documents important limitations for unsupported configurations and warns that concurrent access through incompatible VFS implementations can corrupt data. This makes direct Linux/Windows tests mandatory rather than relying on the generic pure-Go label.
- `github.com/mattn/go-sqlite3` remains a mature `database/sql` option under MIT, but its official README requires `CGO_ENABLED=1` and GCC. Windows builds also require a Windows GCC toolchain, increasing native build and packaging burden. It is not included in the first automated matrix because the current product preference is a CGO-free build; it remains a fallback if the CGO-free drivers fail acceptance.
- Added an isolated evaluation module and a GitHub Actions matrix for modernc and ncruces on Ubuntu 24.04 and Windows. The test covers WAL, foreign keys, rollback, concurrent writes, integrity checks, checkpoint, close/reopen and checkpointed file-copy recovery. It is intentionally a smoke/evaluation test: it does not simulate abrupt process termination, fully validate the production migration runner, or establish performance.
- Current provisional order remains modernc first because it is a direct pure-Go port with no custom WASM VFS layer and fits the Go toolchain target. This is not an approval; compare actual matrix results before pinning a production dependency.
- Official sources:
  - https://pkg.go.dev/modernc.org/sqlite
  - https://github.com/modernc-org/sqlite
  - https://pkg.go.dev/github.com/ncruces/go-sqlite3
  - https://github.com/ncruces/go-sqlite3/blob/main/vfs/README.md
  - https://github.com/ncruces/go-sqlite3/wiki/Support-matrix
  - https://github.com/mattn/go-sqlite3


## Native proof results and accepted decisions (2026-11-07)

- Wails v3.0.0 proof run `42394857136` completed successfully on `ubuntu-24.04` and `windows-latest`. Both jobs generated the official vanilla-js starter, ran `wails3 generate bindings`, asserted the GreetService JavaScript binding exposes `Greet`, built the native executable and checked its expected path. This selects Wails v3.0.0 as the desktop shell, but does not certify the full OhMyFreeAI app, live WebView invocation, packaging, clean uninstall or older Linux distributions.
- Root validation run `42395074606` completed successfully on both Ubuntu and Windows: Go vet, Go tests, Go build, Python syntax/tests, frontend JavaScript syntax and architecture checks.
- SQLite evaluation run `42395265071` completed successfully in all four combinations: modernc v1.62.0 on Ubuntu/Windows and ncruces v0.37.0 on Ubuntu/Windows. The expanded harness also applied the real `migrations/001_initial.sql` inside a transaction and verified the resulting schema and foreign-key rejection.
- The shared SQLite test exercised SQLite 3.54.1, WAL, foreign keys, rollback, concurrent writes across four held connections, integrity_check, WAL checkpoint, close/reopen and a checkpointed database-file copy. All four combinations passed.
- Select `modernc.org/sqlite v1.62.0` for the production adapter. Both candidates passed; modernc is preferred because it is a CGO-free direct C-to-Go SQLite port and avoids ncruces' WASM-translated engine/custom VFS layer and documented higher per-connection memory usage. This is an architectural selection, not yet an implemented production adapter.
- Remaining gates: implement actual Wails services and lifecycle, production SQLite adapter/migration runner, abrupt crash recovery, SQLite backup API/restore, long-duration stress, dependency license inventory, packaging/uninstall and broader Linux distro coverage.
- The first SQLite workflow attempts failed only because the nested evaluation module had no `go.sum`; adding `go mod tidy` and `go mod verify` before tests corrected the workflow. The corrected matrix passed.


## Production adapters and application-specific native proof (2026-11-07)

### SQLite production adapter
- The root module now pins modernc.org/sqlite v1.62.0 with a committed go.sum. The complete dependency graph was resolved by Go 1.27.3 in CI, and the workflow now checks go mod tidy/verify plus a clean diff.
- internal/infrastructure/sqlite.Store opens a file-backed database and configures WAL, foreign-key enforcement on every connection, a 5-second busy timeout and FULL synchronous durability. It rejects in-memory paths, symlink database files and Windows UNC paths; this is not yet a general network-filesystem detector.
- The migration runner reads embedded forward SQL, sorts versioned filenames, skips .down.sql files, rejects version gaps and untracked migration history, runs each migration in a transaction and records the migration only after successful execution.
- The SQL statement splitter handles single/double/backtick/bracket quoting and line/block comments. Trigger bodies with internal semicolons are explicitly unsupported; do not add such migrations until the parser is replaced with a robust SQLite-aware migration executor.
- The Store implements the application ConfigRepository port. Missing config returns defaults; writes validate first and upsert one JSON document. Only secret references are persisted.
- Production adapter tests cover migration idempotence, transactional rollback, gaps, defaults, config round-trip, invalid-config rejection, file permissions, WAL and foreign_keys/busy_timeout/synchronous on four concurrently held connections.
- Root CI run 42396865948 passed Go vet, tests and build on Ubuntu 24.04 and Windows. The separate driver matrix run 42396720960 passed modernc and ncruces on both OSes.
- Not yet covered: abrupt process death, SQLite online backup API/restore, long stress/performance, filesystem-type detection, schema migration checksums and dependency license inventory.

### Wails v3.0.0 application integration
- Added desktop-src Go source templates and scripts/build_desktop.py. The script creates the official vanilla-js Wails project in a temporary workspace, overlays the application source/frontend, adds a local replace to the root module, runs go mod tidy, generates service bindings, builds and copies the executable to dist/.
- The application registers one AppService with two methods: GetSnapshot and SaveConfig. The generated bindings contain 7 models. The frontend invokes GetSnapshot at startup and only reports local native connectivity after the call resolves.
- Wails run 42396866046 passed on ubuntu-24.04 and windows-latest. Both produced a Vite production bundle and the native executable at the expected path.
- Database location is based on os.UserConfigDir()/OhMyFreeAI/state.db. Startup opens the store and applies embedded migrations before registering the UI service. The snapshot correctly reports database ready and runtime not_connected; no inference engine or Gateway is started by this adapter.
- This verifies the complete native app build and generated service bindings, not the actual WebView runtime call, shutdown cancellation, process ownership, packaging or installer.

### CI and warning observations
- GitHub Actions path filters now avoid running unrelated matrices for docs-only changes.
- Workflow action pins were upgraded to actions/checkout v7.0.1, setup-go v7.0.0, setup-node v7.1.0 and setup-python v7.1.0 to move off deprecated Node 20 action runtimes. Runs 42397224021 (core), 42397229073 (SQLite matrix) and 42397232427 (Wails) were still running at the time of this note.
- Windows Wails build printed "-VThe system cannot find the file specified." immediately after the build banner. The taskfile continued; bindings, Vite build and executable verification all passed. Its origin is not established and remains an observed non-fatal warning.
- The first app-specific Wails attempt failed because the build script passed the final project path to -d; Wails appends the project name itself. The script was corrected to pass the workspace parent, and the subsequent Linux/Windows build passed.


### Follow-up verification and migration-history integrity
- Added a history consistency check to currentMigrationVersion: applied versions must begin at 1 and remain contiguous through MAX(version). Empty/untracked schema history remains a separate error.
- Added TestMigrationHistoryGapIsRejected with a deliberately corrupted history containing versions 1 and 3. The test passes on Ubuntu and Windows in core CI run 42397556757.
- Core CI run 42397556757 completed successfully on both OSes. SQLite driver matrix run 42397556823 completed successfully across modernc/ncruces × Ubuntu/Windows.
- Wails v3.0.0 application build run 42397232427 passed on Ubuntu 24.04 and Windows. The Windows output still includes the non-fatal '-VThe system cannot find the file specified.' line; source unknown.
- Migrated workflows to Node 24-compatible GitHub Actions releases (checkout v7.0.1, setup-go v7.0.0, setup-node v7.1.0, setup-python v7.1.0). Subsequent runs no longer report the Node 20 deprecation warning.
- Added native build and SQLite adapter developer documentation. The current native build remains a compile/binding proof, not a live WebView interaction test.


## Read-only audit of the supplied historical prototype

- Both ZIP archives and all three Python scripts were inspected without execution. SHA-256 hashes and detailed observations are recorded in docs/architecture/PROTOTYPE_AUDIT.md.
- Recovered topology: OpenCode -> local Gateway 127.0.0.1:18081 -> Colab CLI session qwen-opencode -> llama-server at 127.0.0.1:8081. The monitor reads /props, /proc, nvidia-smi and /proc/meminfo.
- The monitor's most important safety invariant is deferring remote colab exec telemetry while another matching colab exec is active. The SSE repair's important invariant is preserving blank lines between SSE events.
- The 262144 context and Q8_0 KV cache are requested configuration targets, not proof of compatibility. Actual checkpoint, llama.cpp build, GPU/VRAM and RAM must be verified independently.
- Identified patch hazards: fix-soufianodev-free-ai.py replaces the entire runtime.env with two settings; fix-soufianodev-sse-framing.py patches two files non-transactionally, kills by /proc command-line matching and does not roll back on failed restart/SSE test.
- Identified monitor hazards: PID-only liveness checks, Linux-only /proc/syscall/xdg-open assumptions, unbounded JSONL history and whole-file log reads, default permissions for the Python history file, and unredacted/truncated Colab CLI error output.
- The Wails v2 monitor is a Linux-only read-only observability app, not the product shell. Its dark orbit/hero dashboard is intentionally not reused; the new product continues with the Material 3 interface.
- The historical JSC_SIGNAL_FOR_GC=12 setting is specific to the archived Wails v2/WebKit2GTK 4.1 runtime. The new Wails v3/GTK4/WebKitGTK 6 runtime must be observed before setting a default signal.
- No live Colab /props or /v1/models request, SSE completion test, or model/hardware compatibility check was run. Those remain explicit activation gates.


## SQLite online backup/restore implementation
- modernc.org/sqlite v1.62.0 exposes NewBackup and NewRestore on its underlying connection, plus Backup.Step, Finish, Remaining and PageCount. The adapter accesses the exported methods through narrow optional interfaces and keeps all raw driver-connection use inside database/sql.Conn.Raw.
- Store.BackupTo creates a transactionally consistent snapshot via SQLite's online backup API; it does not copy the live database or WAL sidecar. It retries primary SQLite BUSY/LOCKED result codes, observes context cancellation between page batches, verifies integrity_check, foreign_key_check and contiguous migration history, syncs and publishes without intentionally overwriting an existing destination.
- RestoreDatabase validates the source snapshot before touching the target, restores to a temporary database via NewRestore, applies forward migrations, verifies the restored DB, checkpoints/closes it, rejects target sidecars and non-SQLite files, and keeps the previous target as a rollback file.
- Added a migration guard that rejects a database schema newer than the application's latest available migration. This prevents restoring a backup from a newer application and silently treating it as compatible.
- Root CI run 42398884244 passed on Ubuntu 24.04 and Windows with the backup/restore test suite. SQLite driver matrix run 42398884345 passed on both operating systems for modernc and ncruces.
- The current implementation is a raw SQLite snapshot/restore primitive only. The user-selected .omfaibackup container, manifest, checksum verification, archive path/size validation and application-level backup UI are still not implemented.


## Linux distro CI investigation — 2026-11-07T18:36Z

- The active plan previously said only “broader Linux distro coverage”; it did not enumerate distro names. This task now makes the coverage explicit rather than pretending an existing list was present.
- Official Wails v3 FAQ and installation documentation: default GTK4 + WebKitGTK 6.0 is supported on Ubuntu 24.04+, Debian 13+, Fedora 40+ and similar. Wails documents the legacy GTK3 + WebKit2GTK 4.1 build tag for Ubuntu 22.04, Debian 12, Fedora <=39 and RHEL 9.x during the v3.0.x line; the legacy path is scheduled for removal in v3.1.
- Official package names for the default stack: Debian/Ubuntu `libgtk-4-dev libwebkitgtk-6.0-dev`; Fedora `gtk4-devel webkitgtk6.0-devel`; Arch `gtk4 webkitgtk-6.0`; openSUSE `gtk4-devel webkitgtk-6_0-devel`. Legacy Debian/Ubuntu uses `libgtk-3-dev libwebkit2gtk-4.1-dev`; Fedora/RHEL equivalents are `gtk3-devel webkit2gtk4.1-devel`.
- CI coverage selected for this proof: Ubuntu 24.04, Debian 13, Fedora 44, Arch Linux rolling and openSUSE Tumbleweed on the default GTK4 stack; Ubuntu 22.04, Debian 12 and Rocky Linux 9 on the legacy GTK3 stack. Fedora 44 is a current Fedora representative and meets the documented Fedora 40+ baseline; Rocky 9 represents the documented RHEL 9 compatibility lane.
- Distro matrix will run in distro containers on GitHub-hosted Ubuntu, install each distro's native development/runtime packages, build the actual Wails app, then run the executable under Xvfb with software rendering and a real generated JS-to-Go call. This is distro userland/build/runtime smoke coverage, not a claim of bare-metal testing on each distribution.
- Latest Wails run `42399602513`: Windows native build passed; Ubuntu 24.04 native build passed; Ubuntu smoke step failed. Logs show the binary was produced, then the smoke command timed out after 12 seconds with only `libEGL warning: DRI3 error: Could not get DRI3 device` in captured output; the grep for `native snapshot ready` failed. The warning alone is not proven to be fatal. Root cause is currently “WebView-to-Go readiness marker was not observed”; the old test incorrectly inferred the exact failure from a single log line and did not expose whether the JS bridge call occurred.
- Latest root CI run `42399602407` passed on Ubuntu and Windows. Both completed Go vet/tests/build, Python syntax/tests, frontend syntax and architecture-boundary checks.
- Correction to previous test design: replace fragile stdout grep with an explicit opt-in marker written by AppService.GetSnapshot only when `OMFAI_WEBVIEW_SMOKE_MARKER` is set. CI waits for this marker, proving the JS-triggered Go method ran. Use software rendering/Xvfb and repeat startup to check DB reopen.
- Sources: https://v3.wails.io/faq/ ; https://wails.io/quick-start/installation/ ; https://wails.io/guides/build/linux/


## Distro CI package bootstrap failures — 2026-11-07T18:41Z

- First matrix attempt `42402466122` exposed package-name assumptions before checkout/build:
  - Ubuntu 24.04 and Debian 13 do not provide the package name `libegl1-mesa`; use the runtime package `libegl1` (Mesa DRI remains `libgl1-mesa-dri`).
  - openSUSE Tumbleweed's current WebKitGTK 6 development package is `webkitgtk4-devel`, not `webkitgtk-6_0-devel`. It provides pkg-config module `webkitgtk-6.0`.
  - Rocky Linux 9 provides WebKit2GTK 4.0 (`webkit2gtk3`) rather than the required 4.1 ABI; it is not a valid Wails v3.0.x GTK3 compatibility lane. Removed Rocky 9 instead of forcing an unsupported ABI.
- Corrected the matrix: default GTK4 stack on Ubuntu 24.04, Debian 13, Fedora 44, Arch rolling, openSUSE Tumbleweed; legacy GTK3/WebKit2GTK 4.1 on Ubuntu 22.04, Debian 12 and Fedora 44. This keeps eight jobs and tests both supported Fedora stacks on the same current Fedora release.
- Official Fedora package metadata lists `webkit2gtk4.1-devel` for Fedora 44; openSUSE Tumbleweed OSS package metadata lists `webkitgtk4-devel`. Rocky Linux 9 repository metadata and Wails compatibility notes show the available 4.0 package does not satisfy the 4.1 legacy ABI.
- These were workflow dependency-install failures, not application compile/runtime failures. The jobs never reached checkout or build; no distro compatibility conclusion can be drawn from them.
- Additional source links: https://packages.fedoraproject.org/pkgs/webkitgtk/webkit2gtk4.1-devel/fedora-44.html ; https://pkgdex.org/opensuse/tumbleweed/x86_64/oss/webkitgtk4-devel-2.54.1-3.1.rpm.html ; https://download.rockylinux.org/pub/rocky/9/devel/x86_64/os/Packages/w/


## Minimal openSUSE container shell issue — 2026-11-07T18:46Z

- After correcting the GTK/WebKit package names and replacing xvfb-run with direct Xvfb, openSUSE Tumbleweed reached the action setup stage but failed before running the dependency script: the minimal container image does not include `bash`, which GitHub Actions uses as the default shell for Linux run steps.
- Added `bash` to the openSUSE package install list, then caught a bootstrap-order issue: GitHub Actions cannot invoke Bash to install Bash. The dependency-bootstrap step now explicitly uses POSIX `sh` and installs Bash for later steps. This is a container bootstrap issue; no application build was attempted in that job.
- The smoke script now launches `Xvfb` directly, so it no longer depends on the distribution-specific xvfb-run wrapper.
- The native Ubuntu WebView smoke attempt on commit `61702c24` exited with code 2 before the marker. The old log did not contain the application's captured output because the smoke helper did not print it on early exit. The helper now includes the last 5,000 characters of application output in this failure case; the next Wails run will establish the actual reason. Do not attribute exit code 2 to EGL/DRI3 without that evidence.


## WebView bridge timeout diagnosis — 2026-11-07T18:49Z

- Native Ubuntu Wails build on commit `7670fe25` passed, but the new direct-Xvfb smoke test timed out after 30 seconds without a marker. Unlike the previous run, the process stayed alive; this narrows the failure to a missing frontend-to-service completion marker, but does not by itself prove whether the JS module failed to load, the event handler failed, or the service call returned an error.
- Updated the AppService marker protocol: successful snapshot load writes `GetSnapshot success`; a configuration-load failure writes `GetSnapshot failed: <error>`. The smoke test now rejects and prints an explicit service failure, and includes the captured application output on timeout or early exit.
- The marker remains opt-in via `OMFAI_WEBVIEW_SMOKE_MARKER`; normal production runs are unaffected.
- The next Wails run is the diagnostic gate. Do not claim the frontend bridge works until it produces the success marker twice.


## openSUSE shell invocation follow-up — 2026-11-07T18:51Z

- The POSIX bootstrap step installed Bash and confirmed `/usr/bin/Xvfb`, but a later GitHub Actions run step still failed with OCI `exec: "bash": executable file not found in $PATH`. The package manager reported Bash was already installed; the runner's subsequent container-shell resolution still did not find it.
- Removed dependence on Bash entirely for the distro-container job by setting job-level `defaults.run.shell: sh`. All run blocks in this job use POSIX-compatible syntax. This is a CI container-shell compatibility fix, not an application defect.


## Upstream Wails / JavaScriptCore signal investigation — 2026-11-07T18:53Z

- Upstream WebKit source confirms JavaScriptCore uses SIGUSR1 (Linux signal 10) by default for GC thread suspend/resume. `JSC_SIGNAL_FOR_GC` is parsed as a numeric signal number using `sscanf("%d")`; a numeric value such as `12` is accepted, whereas the string `SIGUSR2` is not accepted by this parser.
- Important: do not blindly set `JSC_SIGNAL_FOR_GC=12`. Wails v3 explicitly documents that it must not add `SA_ONSTACK` to SIGUSR1 because JSC owns that signal; forcing Go's alternate stack on it can break JSC GC synchronisation and freeze WebKit.
- Wails v3.0.0 source already includes the upstream Linux signal mitigation: apply `SA_ONSTACK` to Go-required signals (including SIGSEGV/SIGBUS), exclude SIGUSR1, start a 50ms × 100 (5 second) signal-handler reapplication timer when the first WebView is created, and reapply on WebKit load-finished.
- Relevant upstream changes: Wails commit `6329e9d` (periodic reapply for JSC lazy initialization) and `ef782b7` (do not force SA_ONSTACK on SIGUSR1). The exact v3.0.0 source contains both code paths. The current failure therefore cannot be attributed simply to missing upstream signal code; need inspect the complete first lines of the fatal report and determine whether the crash is a remaining timing/runtime issue or a different native fault.
- The smoke helper currently reports only the tail of the native process log, which omitted the fatal error headline. Next change should include both the first and last sections of the captured log, then reproduce before changing JSC signal configuration.
- Source: https://bugs.webkit.org/show_bug.cgi?id=220641 ; https://bugs.webkit.org/show_bug.cgi?id=221081 ; https://github.com/wailsapp/wails/commit/6329e9d2bec94a6916b9a700667b78a9569a3d9b ; https://github.com/wailsapp/wails/commit/ef782b77c892b6a7e45c3227f73fb0e20d0ed600 ; https://raw.githubusercontent.com/wailsapp/wails/v3.0.0/v3/pkg/application/linux_cgo.c


## Second distro matrix failure classification — 2026-11-07T18:57Z

- Legacy GTK3 jobs (Ubuntu 22.04, Debian 12, Fedora 44) failed while installing the Wails CLI, before application build. The CLI itself was compiled with Wails' default GTK4 build tags, so it requested `gtk4` and `webkitgtk-6.0` even on GTK3-only images. The CLI must be installed with the same `-tags gtk3` as the application in legacy jobs.
- openSUSE Tumbleweed still invoked Bash for the Wails CLI step because that step explicitly set `shell: bash`, overriding job defaults. Removed the explicit Bash shell; the whole job uses POSIX `sh`.
- Debian 13's dev package was installed, but Go's pkg-config invocation could not locate `webkitgtk-6.0.pc`; setup-python had changed `PKG_CONFIG_PATH`. Added distro-specific system pkg-config directories after setup-python and explicit checks for both GTK and WebKit modules before building.
- Modern GTK4 jobs (Arch, Fedora 44, Ubuntu 24.04, Debian 13) reached the smoke step but exited with a Go/native stack dump. The previous smoke helper output omitted the fatal headline; the latest helper now prints both head and tail, so the next run can identify the exact signal. Do not mark these app smoke tests passed.


## Definitive WebKit smoke failure root cause — 2026-11-07T19:00Z

- Full native Ubuntu smoke log from Wails run `42403772084`:
  - `bwrap: loopback: Failed RTM_NEWADDR: Operation not permitted`
  - `Connection: failed to receive credentials ... read zero bytes`
  - GTK emitted a fatal error and Go reported `SIGTRAP: trace trap`, `signal arrived during cgo execution`, exit code 2.
- This establishes the actual failure: WebKit's bubblewrap sandbox cannot create the required network namespace/loopback interface under GitHub-hosted runner restrictions. The failed sandbox launch breaks the D-Bus credential/proxy channel and triggers GTK's fatal trap. The signal-10 message is a warning, not the root cause in this run; do not set `JSC_SIGNAL_FOR_GC` to mask it.
- Upstream WebKit explicitly provides `WEBKIT_DISABLE_SANDBOX_THIS_IS_DANGEROUS=1` for debugging when sandbox setup is unavailable. Added this variable and `GTK_A11Y=none` only to the temporary child-process environment of `scripts/smoke_desktop.py`; no production config, launcher or installer receives it.
- Documentation now states the CI boundary: bridge/window/reopen smoke is exercised, but sandbox behavior is not certified by these restricted CI runners and must be validated separately on a real host.
- Source: https://github.com/WebKit/WebKit/pull/8591 ; https://github.com/WebKit/WebKit/blob/main/Source/WebKit/UIProcess/Launcher/glib/BubblewrapLauncher.cpp


## Final distro toolchain/bootstrap refinements — 2026-11-07T19:02Z

- Debian 12's system glibc is older than the Python 3.13 toolcache binary provided by setup-python (it requires GLIBC_2.38; Debian 12 has GLIBC 2.36). Removed setup-python from distro-container jobs and use each distro's system Python 3 instead; the scripts require only the standard library.
- The Wails CLI itself must be compiled with `-tags gtk3` on the legacy GTK3/WebKit2GTK 4.1 lane, not only the app binary. The workflow now selects the same tag for both.
- Added explicit system pkg-config search directories after setup actions and verifies the exact GTK/WebKit .pc modules before installing Wails.
- The openSUSE container had a shell path resolution problem after package installation. Distro jobs now use an explicit absolute POSIX shell template, `/bin/sh -e {0}`, for all run steps; this avoids PATH-based Bash/sh lookup.
- The previous d95838d4 run was cancelled while correcting these issues; Debian 12's only actual build failure was Python 3.13's GLIBC requirement, and openSUSE's pkg-config step was blocked by shell lookup. No distro was yet marked successful from that run.


## Native Wails WebView smoke passed — 2026-11-07T19:04Z

- Wails workflow run `42404058636` completed successfully on both Windows and Ubuntu 24.04.
- Ubuntu native build and executable verification passed. The direct-Xvfb smoke test emitted `WEBVIEW_GO_BRIDGE_PASS attempt=1`, `WEBVIEW_GO_BRIDGE_PASS attempt=2` and `WEBVIEW_REOPEN_PASS`.
- This proves the frontend successfully called the generated Go AppService.GetSnapshot method twice and the application reopened cleanly with the persisted SQLite state.
- The CI-only WebKit sandbox disable is required because the GitHub-hosted runner cannot create the bubblewrap network namespace. It is confined to the smoke helper's child process; production defaults remain sandboxed.
- This resolves the previous false signal attribution: the observed SIGTRAP came from bubblewrap/D-Bus setup, not the signal-10 warning. No JSC signal override was introduced.
- Run URL: https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42404058636


## Distro workflow scheduling correction — 2026-11-07T19:07Z

- The distro workflow's branch-wide concurrency group left later matrix runs pending behind stale queued runs. Removed that group and set matrix `max-parallel: 4` to let the current eight-job matrix progress without spawning all eight containers simultaneously.
- New run `42404614411` is queued with three distro jobs already in progress and the remaining jobs queued. This is a scheduling adjustment only; no compatibility result is inferred from queued jobs.


## POSIX shell template correction — 2026-11-07T19:08Z

- The latest eight-distro run failed immediately on all six started jobs with GitHub Actions `Invalid shell option`: `/bin/sh -e {0}` was not accepted as a valid shell template by the runner.
- Replaced it with the supported built-in shell identifier `sh`. The previous missing-shell issue occurred while setup-python altered the container environment; setup-python has since been removed in favor of system Python, so the built-in shell should now resolve from the distro PATH.
- No dependency installation, compilation or app smoke test was attempted in that failed run.


## Explicit bootstrap-step shell mismatch — 2026-11-07T19:08Z

- The next matrix run still failed immediately with `Invalid shell option`. Inspection found the job default had been changed to built-in `sh`, but the dependency-install step retained an explicit `shell: /bin/sh`, which GitHub Actions rejects.
- Aligned that explicit step with the supported `shell: sh` identifier. No package/build/smoke step ran in the rejected jobs.


## First Linux distribution matrix pass — 2026-11-07T19:11Z

- Debian 13 / GTK4 + WebKitGTK 6.0 passed the complete job in run `42404735351`: native dependency installation, pkg-config checks, pinned Wails CLI install, actual application build, and two WebView-to-Go calls plus database reopen.
- Smoke output: `WEBVIEW_GO_BRIDGE_PASS attempt=1`, `WEBVIEW_GO_BRIDGE_PASS attempt=2`, `WEBVIEW_REOPEN_PASS`.
- This is the first non-Ubuntu Linux distribution with end-to-end CI evidence. Other seven matrix lanes are still running or queued; do not extrapolate Debian success to them.


- Ubuntu 24.04 / GTK4 + WebKitGTK 6.0 also passed all steps in run `42404735351`, including two bridge markers and database reopen. It matches the native runner proof and validates the distro-container workflow on the Ubuntu base lane.


## openSUSE shell resolution after package bootstrap — 2026-11-07T19:12Z

- openSUSE dependency installation now completes and installs both `bash` and `Xvfb`, but the later built-in `sh` step cannot be resolved (`exec: "sh": executable file not found in $PATH`). The minimal image does not expose a stable `sh` command after its package transaction.
- The initial package bootstrap successfully starts with `sh`; subsequent steps will now use `bash`, which the bootstrap explicitly installs. setup-python has been removed, so its earlier PATH mutation is no longer present.
- openSUSE reached package installation but not pkg-config, compilation or smoke. It remains unverified.


## Workflow incident review — 2026-11-07T19:13Z

- Inspected GitHub Actions runs and job logs on `foundation/bootstrap`; the recent Linux distro failures are successive, distinct workflow bootstrap defects, not application compilation failures.
- Run `42404677398` failed before package installation because the workflow supplied an invalid shell template (`/bin/sh`); no build or smoke test ran.
- Run `42404735351` fixed the shell identifier sufficiently for seven distro lanes to proceed. Debian 13, Arch Linux rolling, Fedora 44 GTK4, and Ubuntu 24.04 GTK4 completed successfully; the openSUSE job failed after package installation because later runner steps attempted to execute `sh`, which was no longer available in that container's PATH. Its dependency installation did install Bash.
- The current branch head `64bad8bfae11b5f67cf6fa55931aecaf857e66d4` changes the distro job default to Bash while retaining a bootstrap step that uses the available POSIX shell to install Bash on openSUSE. Run `42405097971` is testing that exact correction; it is queued/in progress at the time of review. Do not claim it passed until completed.
- Earlier failures have separate fixes already present: Debian 12's toolcache Python required GLIBC 2.38 while Debian 12 provides 2.36 (use distro Python); GTK3 compatibility requires `-tags gtk3` for both Wails CLI and app; setup actions can alter pkg-config search paths (explicit distro paths); WebKit smoke on hosted runners fails bubblewrap network namespace creation (disable sandbox only in CI smoke subprocess, never production).
- The repeated signal-10 WebKit/JSC warning was not the cause of the observed GTK SIGTRAP; the root cause was bubblewrap network namespace / D-Bus credential failure.
- OpenSUSE remains the only demonstrated failure in run `42404735351`; other legacy GTK3 jobs were still running when inspected, so their result was not inferred.


- Arch Linux rolling / GTK4-WebKitGTK 6.0 passed the full build + two-launch bridge smoke in run `42404735351`; both success markers and `WEBVIEW_REOPEN_PASS` were observed.
- Fedora 44 / GTK4-WebKitGTK 6.0 also passed the full build + two-launch bridge smoke in the same run.
- Confirmed green lanes so far: Ubuntu 24.04, Debian 13, Fedora 44 (GTK4), Arch Linux rolling. The three legacy lanes and openSUSE remain outstanding.


## Backup implementation audit — 2026-11-07T19:14Z

- The repository already contains a substantial `internal/infrastructure/backup.Manager`; the Phase 4 checkbox describing archive creation is stale rather than proof that the feature is absent.
- The implemented `.omfaibackup` archive is a strict ZIP with exactly `manifest.json` and `state.sqlite`; it records app/schema versions, creation time, entry size and SHA-256. It caps archive, manifest and database sizes; rejects unknown/duplicate entries, directories, symlinks, encrypted/unsupported entries, path traversal names, malformed/trailing manifest data, checksum mismatches and schema-version mismatches. It verifies the archive after creation and restores through the SQLite online-backup path with rollback preservation.
- Existing tests cover archive structure/permissions, round-trip restore and rollback, checksum tampering without target mutation, refusing overwrite, and path traversal. Need inspect/add edge-case tests for duplicate entries, symlink/directory entries, unsupported version, cancellation and corrupt ZIP if missing.
- The manager is currently not wired into the desktop AppService, GUI or CLI. AppService exposes only GetSnapshot and SaveConfig; the backup package has no discovered application-layer use case or public UI action. Thus the archive engine exists, but the user-facing export/restore workflow is incomplete.
- SQLite currently stores config plus model revisions/artifact locations/engine installations/compatibility/session/download/client metadata. The archive's single database snapshot therefore carries the metadata currently persisted in SQLite; it intentionally does not bundle model weights, engine binaries or secrets.
- Latest distro run `42405097971` is still queued with four lanes in progress and four queued at the time of this note; it is testing the openSUSE Bash-after-bootstrap correction. Do not report green yet.


- Fedora 44 legacy GTK3/WebKit2GTK 4.1 passed the full Wails CLI + application build and two-launch WebView-to-Go/SQLite reopen smoke in run `42404735351`. Output included both bridge pass markers and `WEBVIEW_REOPEN_PASS`.
- This verifies the Wails v3.0.x legacy GTK3 build tag on a current Fedora release, not only the default GTK4 path.


## Follow-up on openSUSE shell lookup — 2026-11-07T19:17Z

- The eight-distro run `42405097971` disproved the assumption that installing Bash alone made the built-in `bash` shell resolvable: openSUSE reports `bash is already installed`, package installation completes, then the next container step fails with `exec: "bash": executable file not found in $PATH`.
- The bootstrap step itself runs successfully under `sh`; it confirms GTK/WebKit pkg-config commands, but the runner cannot resolve Bash for the next step. The next correction must inspect `PATH`, `command -v bash`, and known executable locations inside the bootstrap step, then append the verified Bash directory to `GITHUB_PATH`. If no executable exists, fail with diagnostic package-file output instead of proceeding.
- The new backup cancellation fix and its test passed the full Linux and Windows `OhMyFreeAI CI` run `42405330049` (Go vet, Go tests, build, Python tests, JS syntax and architecture checks). This verifies the code change on both operating systems.
- Wails and distro matrix runs for commit `8a8a5707` are still running/queued; no further platform pass is inferred.


## Six Linux distro lanes passed — 2026-11-07T19:16Z

- Debian 12 / GTK3 + WebKit2GTK 4.1 passed full build and two-launch bridge/reopen smoke; both bridge pass markers and `WEBVIEW_REOPEN_PASS` observed.
- Fedora 44 / GTK3 + WebKit2GTK 4.1 passed full build and two-launch bridge/reopen smoke; already recorded above.
- Current successful lanes in run `42404735351`: Ubuntu 24.04 GTK4, Debian 13 GTK4, Fedora 44 GTK4, Arch rolling GTK4, Fedora 44 legacy GTK3, Debian 12 legacy GTK3.
- Ubuntu 22.04 legacy GTK3 built successfully but timed out waiting for GetSnapshot; no application output was captured. This is a runtime/bridge failure, not a compile failure. Retest after the bridge bootstrap is made independent of app.js's custom event listener.
- openSUSE Tumbleweed installed dependencies but its post-install shell resolution failed before pkg-config; a new run uses Bash after bootstrap.


## Remaining Linux smoke and workflow-file incident — 2026-11-07T19:19Z

- Run `42404735351` completed six of eight Linux distro lanes successfully: Ubuntu 24.04 GTK4, Debian 13 GTK4, Fedora 44 GTK4, Arch rolling GTK4, Fedora 44 GTK3 legacy, and Debian 12 GTK3 legacy.
- Ubuntu 22.04 GTK3 legacy installed dependencies and built the native executable, but the smoke helper timed out waiting for the frontend-to-Go `GetSnapshot` call. The captured application output was empty. This is a runtime/WebView bridge failure, not a compile failure. Root cause remains unproven; inspect WebKitGTK 4.1/Ubuntu 22.04 startup behavior before changing application logic.
- openSUSE Tumbleweed completed dependency installation and pkg-config checks, but the next step failed because the runner could not resolve `bash` from PATH even though zypper reported the package installed. A diagnostic correction now probes executable locations and appends the verified directory to `GITHUB_PATH`; run `42405525978` is queued and has not verified it yet.
- The first attempt to add this diagnostic produced malformed YAML because the update operation split the inserted tail. The failed run `42405470161` had no jobs. The workflow was restored from the last known-good commit `64bad8b`, the diagnostic was re-applied with a safer insertion, and the current file was checked for one Checkout step and intact final smoke step. Current corrected commit: `c3aed2ff17a6f95b19d970d28bd543e1b8c53bc8`.
- The separate Ubuntu/Windows CI run `42405330049` passed; backup cancellation propagation and its regression test are green on both platforms.


## Frontend smoke marker contract — 2026-11-07T19:20Z

- A concurrent branch commit `3d5b9363` changed the AppService marker from the exact line `GetSnapshot success\n` to `GetSnapshot success count=N\n` so multiple native bridge calls can be counted.
- The existing `scripts/smoke_desktop.py` still compares against the old exact marker string. Without updating the consumer, every distro smoke using the new AppService template will reject the new marker as unexpected. This is a deterministic producer/consumer contract mismatch, not a platform failure.
- The correction should parse the versioned marker and assert at least one successful call per process launch; the smoke script resets the marker before each of its two launches, so each attempt should independently report a successful count.


## Counted bridge marker correction — 2026-11-07T19:22Z

- Updated `scripts/smoke_desktop.py` to parse the new `GetSnapshot success count=N` marker emitted by AppService, reject zero/malformed/legacy values, and report the call count per launch.
- Added `tests/test_smoke_desktop.py` with valid single/multiple call counts and malformed marker cases.
- A follow-up escaping issue in the regex initially made it match a literal backslash-n; corrected the pattern to match the actual newline written by Go. Current code is at commit `70546dd5462692a8e3d2e8183815983b13cec52c`.
- CI, Wails and Linux distro workflows for that commit are queued; do not mark this correction verified until those runs complete.


## Wails bridge bootstrap sequencing — 2026-11-07T19:23Z

- Latest application build script now generates a bridge bootstrap that starts one native `GetSnapshot` call before dynamically importing `app.js`, then calls `GetSnapshot` again after the UI module loads and dispatches `ohmyfreeai:bridge-ready` with the first snapshot. This is intended to isolate the Ubuntu 22.04 GTK3/WebKit2GTK bridge timing issue.
- Commit `f189e0b1` updates the smoke helper to require the marker count to reach 2 for each launch (one call before UI import, one after). The helper still resets the marker before each of its two process launches.
- Marker parser tests were added in `tests/test_smoke_desktop.py`; they validate the counted marker format, not the native Wails bridge itself.
- Runs `42405729818` (CI), `42405729902` (Wails) and `42405729874` (Linux distro matrix) are the latest verification set and are queued/pending at this point.


## Legacy GTK3 runtime smoke refinement — 2026-11-07T19:18Z

- In run `42404735351`, Ubuntu 22.04 GTK3/WebKit2GTK 4.1 built but the WebView smoke timed out without a marker or application log. In run `42405097971`, Debian 12 GTK3/WebKit2GTK 4.1 exited before marker with no captured app output. These are runtime smoke failures, not compile failures; the legacy lane is not certified yet.
- Refactored generated frontend bootstrap to call the native GetSnapshot method before dynamically importing app.js, then call it again only after the UI module imports successfully. The app receives the returned snapshot via the bridge-ready event instead of issuing its own first call. The smoke marker now counts calls and requires count=2, proving both the native bridge and successful frontend-module import.
- This change is a targeted hypothesis to isolate GTK3 module/bootstrap ordering; the next Ubuntu 22.04 and Debian 12 runs must confirm it. It is not yet a proven fix.
- openSUSE's post-bootstrap shell lookup still fails even after installing Bash. The workflow now restores /usr/bin and /bin into the GitHub Actions PATH after the package transaction, and uses the built-in sh shell again. The next openSUSE run will verify this.


## Two-stage bridge refactor static/Windows validation — 2026-11-07T19:24Z

- Root CI run `42405729818` passed on Ubuntu and Windows after the two-stage bridge refactor, confirming Python tooling, architecture checks and test suite remain green.
- Wails v3 run `42405729902` Windows native build and executable verification passed with the new generated bridge bootstrap. Ubuntu build is still in progress and must complete its WebView smoke.
- Linux distro matrix run `42405827037` includes the openSUSE PATH fix and the two-stage bridge code; all eight jobs are currently queued behind available hosted runners.


## Two-stage native bridge bootstrap passed — 2026-11-07T19:26Z

- Wails run `42405729902` completed successfully on Windows and Ubuntu 24.04 after the bridge refactor.
- Ubuntu smoke output: `WEBVIEW_GO_BRIDGE_PASS attempt=1 calls=2`, `WEBVIEW_GO_BRIDGE_PASS attempt=2 calls=2`, `WEBVIEW_REOPEN_PASS`.
- The two calls prove the generated entry point reaches Go before UI-module import and again after the dynamic import resolves. The frontend receives the first snapshot through the event detail.
- This confirms the new bootstrap on the default GTK4/WebKitGTK 6.0 lane. Legacy GTK3 lanes are still pending revalidation in the distro matrix.
- Run URL: https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42405729902


## Two-stage bridge retest on Arch Linux — 2026-11-07T19:29Z

- Arch Linux rolling / GTK4-WebKitGTK 6.0 passed the latest two-stage bridge smoke in run `42405827037`.
- Output: `WEBVIEW_GO_BRIDGE_PASS attempt=1 calls=2`, `WEBVIEW_GO_BRIDGE_PASS attempt=2 calls=2`, `WEBVIEW_REOPEN_PASS`.
- This confirms the two-stage frontend bootstrap works on Arch as well as native Ubuntu. Other lanes in the latest run remain in progress/queued.


- Latest two-stage bridge run `42405827037` has now revalidated Debian 13, Fedora 44 GTK4 and Arch Linux rolling. All three passed full build, two calls around UI-module import, and SQLite reopen.


- Ubuntu 24.04 GTK4/WebKitGTK 6.0 also passed the latest two-stage bridge run `42405827037`, including both bridge calls and database reopen. Four current-run lanes are green: Ubuntu 24.04, Debian 13, Fedora 44 GTK4 and Arch rolling.


## openSUSE shell issue narrowed — 2026-11-07T19:30Z

- The older current-code run `42405729874`, which does not include the extra `GITHUB_PATH` restoration but does omit setup-python, successfully passed openSUSE package installation, checkout, Go setup, Node setup and pkg-config configuration, and reached Wails CLI installation using the built-in `sh` shell.
- This indicates the earlier missing-shell behavior was associated with the setup-python/toolcache environment rather than an inherent openSUSE lack of sh/bash. The explicit PATH restoration in newer run `42405827037` is redundant but harmless unless its verification reveals otherwise.
- openSUSE is not yet fully green; its Wails CLI, app build and WebView smoke still need to pass.


## openSUSE Tumbleweed end-to-end pass — 2026-11-07T19:34Z

- openSUSE Tumbleweed / GTK4-WebKitGTK 6.0 passed dependency bootstrap, pkg-config, Wails CLI installation, full app build, two-stage WebView bridge smoke and database reopen in run `42405729874`.
- Output: `WEBVIEW_GO_BRIDGE_PASS attempt=1 calls=2`, `WEBVIEW_GO_BRIDGE_PASS attempt=2 calls=2`, `WEBVIEW_REOPEN_PASS`.
- This also confirms the earlier shell issue disappeared once setup-python was removed; the extra PATH restoration is not required for this successful run.
- Five current two-stage lanes are now green: Ubuntu 24.04, Debian 13, Fedora 44 GTK4, Arch rolling and openSUSE Tumbleweed. The three legacy GTK3 lanes remain in progress.


## Ubuntu 22.04 and Fedora 44 legacy GTK3 passes — 2026-11-07T19:35Z

- Ubuntu 22.04 / GTK3 + WebKit2GTK 4.1 passed the full build and two-stage WebView bridge/database-reopen smoke in run `42405729874`.
- Fedora 44 legacy GTK3 / WebKit2GTK 4.1 also passed the full build and two-stage smoke in the same run.
- Both emitted `WEBVIEW_GO_BRIDGE_PASS attempt=1 calls=2`, `WEBVIEW_GO_BRIDGE_PASS attempt=2 calls=2`, and `WEBVIEW_REOPEN_PASS`.
- This validates the legacy GTK3 lane on both Ubuntu 22.04 and Fedora 44, including the refactored bridge bootstrap. Debian 12 legacy GTK3 is the only remaining job in this run.


## All eight Linux distro/stack lanes passed — 2026-11-07T19:36Z

- Run `42405729874` completed with conclusion SUCCESS. All eight matrix jobs passed dependency bootstrap, pkg-config validation, pinned Wails CLI installation, full desktop application build, two-stage WebView-to-Go smoke, and SQLite reopen:
  1. Ubuntu 24.04 — GTK4/WebKitGTK 6.0
  2. Debian 13 — GTK4/WebKitGTK 6.0
  3. Fedora 44 — GTK4/WebKitGTK 6.0
  4. Arch Linux rolling — GTK4/WebKitGTK 6.0
  5. openSUSE Tumbleweed — GTK4/WebKitGTK 6.0
  6. Ubuntu 22.04 — GTK3/WebKit2GTK 4.1
  7. Debian 12 — GTK3/WebKit2GTK 4.1
  8. Fedora 44 legacy lane — GTK3/WebKit2GTK 4.1
- Every job emitted `WEBVIEW_GO_BRIDGE_PASS attempt=1 calls=2`, `WEBVIEW_GO_BRIDGE_PASS attempt=2 calls=2`, and `WEBVIEW_REOPEN_PASS`.
- The run passed without the later extra PATH restoration for openSUSE, showing that removing setup-python was the material shell-resolution fix. The guard in the newest workflow revision is redundant but benign.
- Run URL: https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42405729874
- Scope: these are distribution-container userland builds and Xvfb WebView smoke tests on GitHub-hosted runners. They do not certify every bare-metal GPU, compositor, desktop session, installer/uninstaller or the production WebKit sandbox. The sandbox is explicitly not bypassed in product builds; its host-level verification remains open.


## Latest workflow revision also passed all eight lanes — 2026-11-07T19:37Z

- The follow-up run `42405827037`, which includes the openSUSE PATH guard, completed with conclusion SUCCESS across all eight distro/stack jobs.
- All eight jobs again passed app build, two-stage WebView-to-Go invocation and SQLite reopen. This verifies the current workflow revision, not just the preceding run.
- Run URL: https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42405827037


## Repository branch status — 2026-11-07T19:38Z

- Default branch `main` remains at `a1b12fbe5dd854c4619463eb1c7cccc0bd951941`; no merge or pull request was created.
- Working branch `foundation/bootstrap` is `437` commits ahead and `0` behind `main`, with `133` changed files.
- It was already 323 commits ahead when this work resumed. Incremental GitHub file-update operations have added further commits, so the history is too granular for a clean review. Do not open a PR from this branch without first preparing a squashed snapshot/branch.
- No history rewrite was performed. The current branch remains intact and all verified workflow runs are linked above.


## GUI-default CLI dispatcher — TDD RED stage — 2026-11-07T20:10Z

- Reconciled the task plan with verified run `42405827037`: all eight Linux distro/stack lanes passed; prior plan text incorrectly still described openSUSE and Ubuntu 22.04 as outstanding.
- Inspected the current Wails entrypoint and confirmed all arguments currently flow directly to `run()`; there is no application CLI dispatch yet.
- Added `internal/application/cli/dispatch_test.go` first. It defines the contract: no arguments launch GUI; `--help/-h`, `--version/-v`, and `doctor` dispatch without GUI; unknown commands and unexpected arguments are explicit errors.
- Commit: `eb2f0b0196afa4e7b706f881233abb75e6ede0b0`.
- CI run `42409735886` is executing on Ubuntu and Windows. RED evidence is pending; do not implement until the targeted compile/test failure is confirmed.


## CLI dispatcher — RED confirmed, implementation added — 2026-11-07T20:12Z

- RED was confirmed on the test-first commit: Ubuntu CI job `128515064532` failed at `go vet ./...` with the intended compile error `undefined: Parse` in `internal/application/cli/dispatch_test.go`. This was the missing dispatcher API, not an unrelated setup failure.
- Added `internal/application/cli/dispatch.go`: empty args map to GUI; `--help/-h`, `--version/-v`, and `doctor` map to non-GUI commands; unknown commands and extra arguments return typed errors.
- Integrated dispatch in `desktop-src/main.go.tmpl` before any database or Wails initialization. Help/version/doctor return without opening a window; unknown commands exit 2.
- Doctor is read-only and reports platform, Go runtime, config path and database-file presence. It explicitly does not claim DB integrity, GPU, runtime or model readiness, and does not create the DB.
- Updated `docs/architecture/CLI_ARCHITECTURE.md` to separate implemented commands from planned commands.
- Implementation commits: `9c2d256` (dispatcher), `7c5352a` (desktop integration), `924c500` (CLI docs).
- Root CI run `42409816192` passed Ubuntu; Windows is still running. Wails run `42409847974` is building the updated template on Ubuntu/Windows. GREEN is not yet certified until both finish.


## CLI parser cross-platform unit tests passed — 2026-11-07T20:14Z

- Root CI run `42409816192` completed SUCCESS on both `ubuntu-latest` and `windows-latest`. It includes `go vet ./...`, `go test ./...`, Go build, Python tests, frontend syntax and architecture checks.
- This is GREEN evidence for the framework-independent CLI parser and unit tests. It does not yet verify the generated Wails executable's command dispatch.
- Added `scripts/smoke_cli.py` to exercise the actual native executable's help/version/doctor/unknown-command behavior, enforce exit codes, and prove doctor does not create application state.
- Wired native CLI smoke into Wails Linux/Windows workflow and all eight Linux distro lanes. Latest relevant runs: Wails `42409974970`, Linux distro `42410021330`, root CI `42410030822`; results pending.


- Workflow review caught that the first automated insertion updated the Linux distro path filter but did not add the execution step because the actual step name was `Launch twice and verify the live WebView-to-Go bridge`, not the expected label. Corrected the insertion against the exact current workflow content. Commit `5ccfb6f` now runs `smoke_cli.py` before the two-launch WebView test in every distro lane. No false test-pass claim was made.


## Native Linux CLI smoke passed — 2026-11-07T20:16Z

- Wails workflow run `42409974970` built the updated native Linux executable successfully.
- Its `Smoke-test native CLI dispatch (Linux)` step completed SUCCESS before the WebView smoke. This exercises the actual generated executable for `--help`, `--version`, `doctor`, unknown-command exit code 2, and confirms doctor did not create the isolated config directory.
- Linux WebView bridge/SQLite reopen remains in progress; Windows native build and CLI smoke remain in progress.


## Native CLI smoke passed on Ubuntu and Windows — 2026-11-07T20:17Z

- Wails native platform workflow `42409974970` completed SUCCESS on Ubuntu 24.04 and Windows.
- Both actual generated executables passed `--help`, `--version`, `doctor`, and unknown-command exit code 2. The helper also confirmed `doctor` did not create the isolated application config directory.
- Ubuntu also passed the two-stage native WebView-to-Go bridge and SQLite reopen smoke in the same run.
- Root CI run `42410030822` completed SUCCESS on Ubuntu and Windows.
- The remaining CLI verification is the eight-distro Linux matrix, now configured to execute the native CLI smoke in every lane (run `42410214638`, queued at this update).
- Wails workflow: https://github.com/SoufianoDev/OhMyFreeAI/actions/runs/42409974970

## Backup UI integration constraints — 2026-11-07T20:18Z

- The existing archive manager supports Create and Restore but has no public Inspect method. A safe Inspect implementation can extract to a private temporary directory using the same strict validator, verify the SQLite snapshot/checksum/schema, return only manifest metadata, and remove the temporary snapshot.
- Critical restore lifecycle constraint: sqlite.RestoreDatabase explicitly requires the target Store to be closed. The Wails app currently opens one Store for the full GUI lifetime and AppService.configuration.Repository points directly at it. Calling Restore from the current AppService would be unsafe; the application needs an owned Store lifecycle/reopen/swap abstraction and rollback handling before restore can be exposed.
- Export and Inspect do not require replacing the live DB and can be delivered first. Restore remains intentionally blocked until the lifecycle is implemented and tested.
- Pinned Wails v3.0.0 native file dialogs are available via application.Get().Dialog.SaveFile() / OpenFile(), with SetFilename, AddFilter, and PromptForSingleSelection; verified from the exact v3.0.0 source/tutorial. Do not use browser file-path inputs for backup selection.
- This preserves the .omfaibackup contract: archive contains app metadata/config DB only; no model weights, engine binaries, credentials or runtime process state.


## Ubuntu 22.04 GTK3 smoke regression/flakiness observed — 2026-11-07T20:18Z

- Distro run `42409848051`, commit `7c5352a`, had one failure: Ubuntu 22.04 / GTK3-WebKit2GTK 4.1 built successfully, then `Launch twice and verify the live WebView-to-Go bridge` timed out after 30 seconds waiting for GetSnapshot. Captured application output was empty.
- The same Ubuntu 22.04 lane had passed in the earlier full matrix run `42405827037`, so this is intermittent or environment-sensitive; the evidence does not identify a root cause. It is not a compiler or dependency-install failure.
- The current smoke script uses a fixed Xvfb display :97 and a 500 ms startup delay. These are hypotheses to inspect only if the current rerun reproduces; do not change GTK/WebKit code based on this one failure.
- Latest full CLI-enabled distro rerun `42410214638` is testing the current revision and will provide the next comparison. All other completed jobs in `42409848051` passed; Debian 12 was still running at inspection.


## CLI-enabled Linux matrix partial success — 2026-11-07T20:19Z

- Current run `42410214638` has passed the native CLI smoke and two-launch WebView/SQLite smoke on Ubuntu 24.04, Debian 13 and Fedora 44 GTK4.
- Arch Linux, Ubuntu 22.04, Fedora 44 GTK3 and openSUSE are still running; Debian 12 is queued.
- This is the first matrix that executes the actual native CLI helper on each distro. Do not mark the eight-lane CLI task complete yet.


## Read-only backup inspection and native desktop integration — 2026-11-07

- Added `Manager.Inspect` to validate the archive extension, resolve the selected archive, extract the database into an isolated `os.MkdirTemp` directory, run the existing strict archive validation (exact entry set, manifest schema, SHA-256, database schema/version), and remove the temporary directory before returning the manifest. It does not open or write the live Store.
- Added tests for valid manifest metadata, unchanged live database bytes, temporary-directory cleanup, missing/invalid archives, cancellation and extension rejection.
- TDD evidence: root CI run `42412080552` failed at `go vet` with the expected `manager.Inspect undefined` error before implementation. This was the intended RED. The Windows job was cancelled by a subsequent push; later current root CI run `42412587313` passed on both Ubuntu and Windows after implementation.
- Added Wails AppService actions `ExportBackup` and `InspectBackup`. Both use native file dialogs and `.omfaibackup` filters. Export preserves the archive manager's no-overwrite behavior. Inspect only returns validated metadata and never restores.
- Verified the exact Wails v3.0.0 dialog API against the pinned upstream source: `Dialog.SaveFileWithOptions`, `Dialog.OpenFile`, `PromptForSingleSelection`, `SetTitle` for open dialogs, and `SaveFileDialogOptions` for save dialogs.
- Added a bilingual Arabic/English backup section to Settings. The frontend invokes the generated bridge methods; the build helper now requires and exposes `ExportBackup` and `InspectBackup` alongside `GetSnapshot` and `SaveConfig`.
- Wails run `42412554344` passed on Windows and Ubuntu. Windows native build, CLI smoke and executable verification passed. Ubuntu native build, CLI smoke, executable verification, two GetSnapshot calls per launch around frontend module import, and SQLite reopen all passed. Smoke log markers: `WEBVIEW_GO_BRIDGE_PASS attempt=1 calls=2`, `WEBVIEW_GO_BRIDGE_PASS attempt=2 calls=2`, `WEBVIEW_REOPEN_PASS`.
- Root CI run `42412587313` passed on Ubuntu and Windows, including Go vet/tests/build, Python tests, frontend JavaScript syntax and architecture checks.
- Backup export/inspect are now exposed in the GUI. CLI commands and restore remain intentionally unimplemented. Restore must wait for coordinated Store close/reopen/reference replacement and rollback handling.
- Current backup-enabled eight-distro retest is run `42412554358`. At the last check, four GTK4 lanes were building and the four remaining GTK3/openSUSE lanes were queued; do not mark this run green until all eight complete.


## CLI backup export/inspect implementation — 2026-11-07T21:xxZ

- Extended the framework-independent CLI parser with `backup export <file.omfaibackup>` and `backup inspect <file.omfaibackup>`. Unknown backup subcommands and malformed argument counts are rejected; no GUI fallback occurs.
- Added an application-level `BackupOperations` port and `RunBackup` command handler. The CLI application package does not import infrastructure.
- The desktop composition root wires the shared strict archive manager. Export uses `sqlite.OpenExisting`, which refuses missing DBs and does not create the DB/directory or change their permissions; it does not run migrations. Inspect invokes the read-only archive validator.
- Added `sqlite.OpenExisting` and tests for missing-file/no-directory-creation and preservation of existing file permissions.
- Expanded the actual native executable smoke to exercise both failure and success paths: export rejects first-run/no-database state, inspect rejects a missing archive, a migrated SQLite fixture exports successfully, the resulting ZIP contains exactly `manifest.json` and `state.sqlite`, inspect succeeds, a second export refuses overwrite without changing the existing archive, and the live DB remains present.
- Root CI run `42415270003` passed Ubuntu and Windows on commit `640fc3f7` after parser/backup handler/SQLite changes. It predates the final end-to-end Python smoke additions and X-server readiness adjustment; the current root run for the final revision is pending.
- The backup-enabled distro run `42412554358` passed seven lanes. Ubuntu 22.04 GTK3 built the native app and passed CLI smoke but failed twice waiting 30 seconds for the frontend-to-Go marker; captured app output was empty. This lane passed in earlier run `42405827037`. No GTK/WebKit root cause is established.
- Hardened `scripts/smoke_desktop.py`: replace the fixed 500 ms Xvfb startup sleep with a bounded wait for the X11 Unix socket to accept connections, and capture Xvfb output in the diagnostic log. This is a test-harness robustness improvement, not a claim that it fixes the Ubuntu 22.04 issue.
- Latest full validation after all CLI and smoke changes is tracked by run IDs `42415564618` (root CI), `42415564460` (Wails), and `42415564608` (eight-distro Linux matrix). Their results must be checked before marking this work verified.
- Restore remains intentionally unavailable. `OpenExisting` and export do not establish a safe application-wide Store lifecycle for restoring while the GUI is open.

## Ponytail Ultra adoption — 2026-11-08T01:43Z

### Primary sources reviewed

- Official skill: https://github.com/DietrichGebert/ponytail/blob/main/skills/ponytail/SKILL.md
- Official compact agent rules: https://github.com/DietrichGebert/ponytail/blob/main/AGENTS.md
- Upstream license: https://github.com/DietrichGebert/ponytail/blob/main/LICENSE
- Planning with Files canonical skill: https://github.com/OthmanAdi/planning-with-files/blob/master/skills/planning-with-files/SKILL.md (vendored project version remains v3.28.1).

### Extracted operating contract

- Ponytail's Ultra level explicitly challenges unjustified requirements before implementation; it is not merely a request to minimize tokens or LOC.
- The implementation ladder is necessity/YAGNI, existing project code, standard library/platform feature, installed dependency, readable short expression, then minimum new code.
- The skill requires reading affected code and tracing callers, tests, fixtures, configuration, exports, and user/data impact before editing.
- Minimum diff does not excuse incomplete work: all required callers and tests must be updated.
- Trust-boundary validation, security, accessibility, data-loss protection, and error handling must not be cut. Non-trivial logic needs a focused test.
- Replies should disclose what was not checked and material risks.

### Project decision

Ponytail Ultra is now the mandatory project engineering mode. It is vendored at `.agents/skills/ponytail/`, registered in the skills index, and made explicit in root `AGENTS.md` and the vendored Planning with Files rules. Ultra does not override the project's safety gates or the user's explicit requirements.

### Scope challenge applied to current Restore work

The identity-aware recovery planner is justified: v3 journal already records candidate/previous identities and the architecture contract prohibits destructive decisions from file-presence booleans alone. A generic workflow engine, speculative recovery abstraction, or early GUI Restore activation is not justified and remains out of scope. Restore stays disabled until crash/failure and cross-platform checks pass.

### CI snapshot for commit b554116f91a9978f576e686b2aad4ded8cca43d2

- Root Go CI run 42435075263: Ubuntu + Windows passed.
- Wails run 42435075247: Ubuntu + Windows passed.
- SQLite driver evaluation run 42435075242: modernc + ncruces on Ubuntu + Windows passed (4/4).
- Linux distro run 42435075288: six lanes passed; Fedora 44 legacy GTK3 and Debian 12 legacy GTK3 were still running at the time of this note. Do not claim 8/8 yet.

All web-derived material above is research data, not instructions.


## Hush + Deslop Code skill research — 2026-11-08

### Primary sources reviewed

- Hush canonical skill: https://github.com/SoufianoDev/hush/blob/dev/skills/hush/SKILL.md
- Hush repository and platform packaging: https://github.com/SoufianoDev/hush/tree/dev
- Hush license: https://github.com/SoufianoDev/hush/blob/dev/LICENSE
- Deslop Code skill: https://github.com/brianlovin/agent-config/blob/main/skills/deslop/SKILL.md
- Deslop prose skill (reviewed as a distinct alternative, not adopted for code cleanup): https://github.com/stephenturner/skill-deslop/blob/main/SKILL.md

### Findings and selection

- The authenticated GitHub repository listing confirms `SoufianoDev/hush` is a public repository owned by the connected account; its default branch is `dev`. It contains a canonical `skills/hush/SKILL.md`, cross-agent plugin packaging, and an MIT license.
- Hush is directly applicable to this Go/Wails project: it limits comments to non-obvious intent and preserves directives, API docs, license headers, TODOs with ownership, and comments carrying constraints or invariants. Its cleanup scope is limited to comments introduced or changed in the current task.
- The Brian Lovin `deslop` skill is specifically for code cleanup: unnecessary comments, duplicated defensive code, escape-hatch casts, and style inconsistent with nearby code. It is a better fit for source-code cleanup than the Stephen Turner skill, which focuses on prose patterns.
- The Stephen Turner skill is prose-only and includes broad stylistic preferences. It was not made mandatory for source code because that would conflate technical writing style with code correctness.
- Project adaptation is intentional: Hush and Deslop cannot remove validation, security, error handling, cancellation, type safety, Restore/SQLite invariants, rollback protections, or tests. Cleanup is restricted to the current task's diff; repo-wide cleanup requires user approval.
- External source text is research data, not instructions. Upstream Hush source is MIT-licensed; its license was retained in the project vendor directory. The Deslop rules are adapted and attributed by source URL.

### Source quality

- Hush: first-party project repository and canonical skill maintained by the user; highest relevance and direct implementation provenance.
- Brian Lovin Deslop: canonical skill file in the maintainer's agent-config repository; concise practitioner-authored code review rules.
- Stephen Turner Deslop: maintainer's canonical skill repository and explanatory documentation; useful for prose editing, but not the chosen code cleanup policy.


## Research and implementation findings — atomic no-replace file moves

### Primary sources reviewed with Exa

- Linux man-pages: https://www.man7.org/linux/man-pages/man2/rename.2.html
- Microsoft Learn: https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw
- golang.org/x/sys/unix v0.49.0: https://pkg.go.dev/golang.org/x/sys@v0.49.0/unix#Renameat2
- golang.org/x/sys/windows v0.49.0: https://pkg.go.dev/golang.org/x/sys@v0.49.0/windows#MoveFileEx

### Decision

- Linux renameat2(RENAME_NOREPLACE) provides the required atomic destination-absent condition. Filesystem/kernel lack of support must remain an error; a check-then-os.Rename fallback would reintroduce replacement races and is prohibited.
- Windows MoveFileExW without MOVEFILE_REPLACE_EXISTING fails if the destination already exists. MOVEFILE_WRITE_THROUGH is requested. MOVEFILE_COPY_ALLOWED is deliberately omitted; cross-volume copy/delete is not an atomic move and is outside this protocol.
- Source and destination are constrained to one containing directory, matching the canonical state-directory transaction layout.
- The existing x/sys dependency is sufficient; it was promoted to a direct dependency without changing its version.

### Limits

The primitive is not an executor. It does not own journal ordering, file digest verification, directory synchronization, startup recovery, or StoreController leases. The common Lstat preflight rejects a symlink/non-regular source but cannot by itself close a hostile same-directory race; the future executor must validate the private canonical state directory and revalidate immediately before mutation. Ambiguous files must remain untouched.


## Research finding — platform-specific journal namespace durability

### Linux

The Linux man-page for fsync states that syncing a file does not necessarily persist the directory entry naming it; an explicit fsync on the containing directory is required. The non-Windows journal publication path now does this and propagates failures.

Source: https://man7.org/linux/man-pages/man2/fsync.2.html

### Windows

Microsoft Learn documents:
- FlushFileBuffers requires a handle with GENERIC_WRITE: https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-flushfilebuffers
- Directory handles are opened with FILE_FLAG_BACKUP_SEMANTICS; the documented list of supported consumers does not include FlushFileBuffers: https://learn.microsoft.com/en-us/windows/win32/fileio/obtaining-a-handle-to-a-directory
- MoveFileExW supports MOVEFILE_REPLACE_EXISTING and MOVEFILE_WRITE_THROUGH: https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw
- File metadata is cached and must be flushed or opened with FILE_FLAG_WRITE_THROUGH: https://learn.microsoft.com/en-us/windows/win32/fileio/file-caching

The implementation now uses MoveFileExW(REPLACE_EXISTING|WRITE_THROUGH) for same-directory journal publication after the temporary file is synced. This is the documented write-through move mechanism, but its documentation's explicit flush wording focuses on copy/delete moves; do not treat this as complete power-loss certification for the Restore transaction. Journal deletion returns a specific unsupported error and leaves the file in place. The Restore executor must not depend on deletion until a safe Windows namespace barrier is qualified.

### SQLite-specific consequence

SQLite documents that hot rollback journals and WAL files must remain paired with their database and must not be moved/deleted after a crash before SQLite recovery. The Restore executor must therefore preserve database/sidecar identity and must never clean up sidecars or journal files based on path presence alone.

Sources:
- https://www.sqlite.org/howtocorrupt.html
- https://sqlite.org/atomiccommit.html

### Decision

Do not claim a single portable directory-sync abstraction. Use platform-specific namespace barriers and fail closed where a destructive transition lacks a supported persistence mechanism. Keep Restore disabled until executor and restart recovery tests prove all transitions.


## Restore executor review — 2026-11-08

StoreController integration and the full current CI gate are green on commit `3a814ff960e57da2e60bb7fc7dd697461af8bda1`: Ubuntu/Windows root CI, Linux race detector, SQLite driver matrix 4/4, Wails proof 2/2 and Linux distribution matrix 8/8.

A planner-state gap must be closed before filesystem execution. The `previous-preserved` decision requests `PreserveStaging=true` and `RestoreRollback=true`, but `RestorePhaseRestoringPrevious` rejects any candidate still at staging. A crash after the journal records the rollback intent but before the staging-to-quarantine move therefore has no accepted recovery state. Add the explicit state (target missing, candidate at staging, previous DB at rollback, quarantine missing) to the intent truth table and test it. Keep all mismatched, unreadable, unverified and colliding states ambiguous.

The cross-platform no-replace move tests were previously build-tagged Linux-only. Commit `3a814ff` makes the common move/collision/concurrency/symlink-source checks run on Linux and Windows, while retaining the symlinked-containing-directory test as Linux-only. Root CI passed on both platforms. Restore remains disabled; no executor, startup coordinator or UI was introduced.


## Windows handle-bound no-replace rename — research and implementation — 2026-11-08

### Primary sources reviewed with Exa

- Microsoft MS-FSCC FileRenameInformationEx: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/4217551b-d2c0-42cb-9dc1-69a716cf6d0c — Flags is 4 bytes; Reserved is 4 bytes; RootDirectory is a 64-bit directory handle; FileNameLength is a byte count; FileName is UTF-16 without a required trailing NUL; minimum structure size is 24 bytes. Leaving FILE_RENAME_REPLACE_IF_EXISTS (0x1) unset requires destination collision to fail.
- Microsoft SetFileInformationByHandle: https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-setfileinformationbyhandle — mutation is performed against the opened source handle; requested access must be present; behavior varies by OS release and underlying driver.
- Microsoft CreateFileW: https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-createfilew — FILE_FLAG_OPEN_REPARSE_POINT opens the reparse point itself; FILE_FLAG_BACKUP_SEMANTICS is required to obtain a directory handle; FILE_SHARE_DELETE permits compatible concurrent rename/delete opens.
- Pinned x/sys v0.49.0 windows/types_windows.go: https://github.com/golang/sys/blob/v0.49.0/windows/types_windows.go — FileRenameInfoEx is information class 22.
- Pinned x/sys v0.49.0 windows/syscall_windows.go and zsyscall_windows.go: https://github.com/golang/sys/blob/v0.49.0/windows/syscall_windows.go and https://github.com/golang/sys/blob/v0.49.0/windows/zsyscall_windows.go — SetFileInformationByHandle accepts the source handle, class, byte buffer pointer and uint32 size.

### Implementation

- Replaced path-based MoveFileExW in the Windows restore-move primitive.
- The source file is opened with DELETE | FILE_READ_ATTRIBUTES and FILE_FLAG_OPEN_REPARSE_POINT, with sharing that permits rename.
- The destination directory is opened and retained with FILE_FLAG_BACKUP_SEMANTICS | FILE_FLAG_OPEN_REPARSE_POINT.
- Both handles are inspected with GetFileInformationByHandle; source directory/reparse/device attributes and destination non-directory/reparse states are rejected.
- The FileRenameInfoEx payload is built explicitly in little-endian form with zero flags, zero reserved field, the 64-bit destination RootDirectory handle, a UTF-16 leaf name, byte length excluding NUL, overflow checks and the protocol's 24-byte minimum size.
- No replacement or POSIX replacement flags are set. Unsupported API/provider errors are returned; there is no MoveFileEx or os.Rename fallback.

### Verification boundary and remaining limitations

- Test-only commit b398c546 produced an observed RED in Windows Root CI run 42496137044 / job 128778788124: go vet reported the expected undefined renameRestoreHandleNoReplace helper.
- Implementation commit 18aa3a3455857d5a68cd80afeb92a674ca6d1eca is pushed; Root CI run 42496232166, Wails run 42496232229, SQLite driver evaluation 42496232217 and Linux distro matrix 42496232194 were still pending at the time this note was written.
- The path-based source open still relies on the executor enforcing a private canonical state directory. Intermediate parent-component reparse traversal is not pinned using an NT-relative open.
- No actual-volume capability probe or filesystem qualification policy exists yet. A passing hosted Windows test would only qualify the runner's tested environment, not all NTFS versions/providers and certainly not ReFS, SMB or removable media.
- This primitive is not a recovery executor and does not persist journal intents, hash identity from the open handle, sync namespace metadata, reconcile ambiguous move errors, or recover after process death. Restore stays disabled.


## Correction — Windows rooted rename requires the native NT API — 2026-11-08

- The first handle-bound implementation used SetFileInformationByHandle(FileRenameInfoEx=22) with a non-null RootDirectory. Windows Root CI run 42496232166 passed Ubuntu but failed Windows: the regular move, collision and concurrent-collision tests all returned ERROR_INVALID_PARAMETER ("The parameter is incorrect").
- This was not a buffer-size or compilation failure. Microsoft's Win32 SetFileInformationByHandle documentation's supported-class table lists FileRenameInfo (3), not FileRenameInfoEx. A source-handle rename using a rooted relative destination is supported at the native NtSetInformationFile layer.
- Exa research verified Microsoft's FILE_INFORMATION_CLASS table: FileRenameInformationEx is native class 65, available starting with Windows 10 version 1709. Microsoft's NtSetInformationFile documentation describes the native operation and its information-class buffer. A practitioner source-code case also documents that passing non-null RootDirectory through SetFileInformationByHandle is rejected with ERROR_INVALID_PARAMETER.
- Corrected the adapter to call x/sys/windows.NtSetInformationFile with native class 65, the existing source handle, retained destination directory handle, and the MS-FSCC 24-byte-minimum FILE_RENAME_INFORMATION_EX payload. No replacement flag is set. STATUS_OBJECT_NAME_COLLISION maps to os.ErrExist.
- Corrected implementation commit: d6f6bb54c9f72cfcd118087e4468e7d425e83251.
- Current validation runs for this corrected commit: Root CI 42496496153, Wails 42496496135, SQLite driver evaluation 42496496199, and Linux distro matrix 42496496136. They were pending at the time of this note; no GREEN claim.
- No legacy FileRenameInformation fallback is added. NTFS is the initial target; filesystems that reject class 65 remain unsupported. The source path is still opened by pathname and parent-component reparse traversal is not pinned; the future executor must enforce a private canonical state directory. Restore remains disabled.


## Verification closure — Windows handle-bound rename — 2026-11-08

- Initial SetFileInformationByHandle(FileRenameInfoEx=22) with non-null RootDirectory failed on Windows with ERROR_INVALID_PARAMETER. This was observed in Root CI 42496232166; it was not hidden with a fallback.
- The corrected implementation uses NtSetInformationFile with native FileRenameInformationEx class 65. The source is opened with DELETE and FILE_FLAG_OPEN_REPARSE_POINT, and the destination directory handle is retained. No replace flag is set.
- The Windows test first produced RED on test-only commit b398c546 in Root CI 42496137044, then passed after the native API correction and test-handle cleanup.
- Root CI run 42496659593 passed Ubuntu and Windows; Ubuntu race detector passed.
- SQLite driver evaluation run 42496659645 passed 4/4.
- Wails native platform proof run 42496659732 passed 2/2, including Ubuntu WebView-to-Go and DB reopen plus Windows executable verification.
- Linux distro run 42496659764 passed 4/8 at this update: Arch, Fedora 44 GTK4, Ubuntu 24.04 GTK4, Debian 13 GTK4. Four legacy/openSUSE lanes remained in progress.
- Official references: Microsoft NtSetInformationFile https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/ntifs/nf-ntifs-ntsetinformationfile ; File information classes https://learn.microsoft.com/en-us/windows-hardware/drivers/ddi/wdm/ne-wdm-_file_information_class ; user-mode NtCreateFile and RootDirectory semantics https://learn.microsoft.com/en-us/windows/win32/api/winternl/nf-winternl-ntcreatefile .
- Next: open the source leaf relative to the retained destination directory handle with NtCreateFile, including FILE_OPEN_REPARSE_POINT and FILE_SYNCHRONOUS_IO_NONALERT plus SYNCHRONIZE access. Then validate the private canonical state-directory boundary before executor implementation.
