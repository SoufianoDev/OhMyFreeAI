# OhMyFreeAI

**Free-first desktop control plane for local and permitted remote AI inference.**

OhMyFreeAI is a cross-platform desktop application focused on managing model artifacts, inference engines, runtime sessions, a local OpenAI-compatible Gateway, client integration, monitoring, recovery, and backups.

> **Important Colab Free limitation:** Google does not guarantee free GPU availability or runtime duration. Its current managed-runtime policy prohibits certain remote-control, web-service, and notebook-UI-bypass patterns. OhMyFreeAI will not implement tunnels or workarounds to evade those restrictions. The Colab adapter remains subject to an explicit policy-compliance and feasibility gate. See [Colab policy and risk](docs/legal/COLAB_RISK_DISCLOSURE.md).

## Product constraints

- Free-only. No paid hosting, model APIs, or mandatory paid dependencies.
- Linux and Windows are first-class targets; macOS is deferred.
- Go 1.27.3 is the target toolchain; Go owns application logic and lifecycle.
- Wails is a desktop-shell candidate; its major version is not locked until the Linux/Windows proof-of-concept passes.
- Frontend: semantic HTML, CSS, and JavaScript ES modules, using Material Design 3 tokens and accessible interaction patterns.
- SQLite is the planned application store.
- llama.cpp is the primary inference engine, subject to exact artifact/backend/runtime compatibility checks.
- OpenCode is the first supported client.
- Default local Gateway address: `127.0.0.1:18081`.
- Backup artifact: `.omfaibackup` (product name: `omfaibackup`).

## Status

This repository is in the architecture/bootstrap phase. The current work is on `foundation/bootstrap`. The earlier Wails monitor archives and Python patch/monitor scripts are audit inputs, not production code.

## Documentation

- [Project specification](docs/architecture/PROJECT_SPEC.md)
- [System architecture](docs/architecture/SYSTEM_ARCHITECTURE.md)
- [Module boundaries](docs/architecture/MODULE_BOUNDARIES.md)
- [Domain model](docs/architecture/DOMAIN_MODEL.md)
- [State machines](docs/architecture/STATE_MACHINES.md)
- [Model and runtime compatibility](docs/architecture/COMPATIBILITY_ARCHITECTURE.md)
- [Runtime and engine distribution](docs/architecture/RUNTIME_AND_ENGINE.md)
- [Gateway and client integration](docs/architecture/GATEWAY_AND_CLIENT.md)
- [Storage and backup](docs/architecture/STORAGE_AND_BACKUP.md)
- [Security and recovery](docs/architecture/SECURITY_AND_RECOVERY.md)
- [UI/UX architecture](docs/architecture/UI_UX_ARCHITECTURE.md)
- [Developer toolkit](docs/architecture/DEVELOPER_TOOLKIT.md)
- [Implementation blueprint](docs/architecture/IMPLEMENTATION_BLUEPRINT.md)
- [Colab Free policy disclosure](docs/legal/COLAB_RISK_DISCLOSURE.md)
- [Architecture decisions](docs/adr/)

## Repository workflow

Project-scoped agent skills live in [`.agents/skills/`](.agents/skills/README.md). Planning records live under [`.planning/`](.planning/). Read the active task plan before making architectural changes.

## Current engineering gate

Do not start broad model-catalog implementation or enable remote inference until:
1. the Colab Free usage pattern is policy-reviewed and explicitly permitted;
2. the Wails major version passes a Linux and Windows proof-of-concept;
3. llama.cpp binary distribution is pinned to a tested stable release/backend combination;
4. the exact model/checkpoint/quantization has passed a real load and inference test.

## License

Project licensing will be selected and documented before the first distributable release. Third-party skill materials retain their original licenses and attribution in `.agents/skills/`.
