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
