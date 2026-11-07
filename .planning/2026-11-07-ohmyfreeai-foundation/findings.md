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

## Visual/Browser Findings

- No visual design references have been accepted or applied yet. Inspect the user's Material Design 3 skill before making UI decisions.
