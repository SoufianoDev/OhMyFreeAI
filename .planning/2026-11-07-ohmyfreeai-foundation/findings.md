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
