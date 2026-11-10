# OhMyFreeAI

**Free-first desktop control plane for local and permitted remote AI inference.**

OhMyFreeAI is a cross-platform desktop application focused on managing model artifacts, inference engines, runtime sessions, a local OpenAI-compatible Gateway, client integration, monitoring, recovery, and backups.

> **Important Colab Free limitation:** Google does not guarantee free GPU availability or runtime duration. Its current managed-runtime policy prohibits certain remote-control, web-service, and notebook-UI-bypass patterns. OhMyFreeAI will not implement tunnels or workarounds to evade those restrictions. The Colab adapter remains subject to an explicit policy-compliance and feasibility gate. See [Colab policy and risk](docs/legal/COLAB_RISK_DISCLOSURE.md).

## Product constraints

- Free-only. No paid hosting, model APIs, or mandatory paid dependencies.
- Linux and Windows are first-class targets; macOS is deferred.
- Go 1.27.3 is the target toolchain; Go owns application logic and lifecycle.
- Wails v3.0.0 is pinned as the desktop shell; Linux defaults to GTK4/WebKitGTK 6.0, with a tested legacy GTK3/WebKitGTK 4.1 compatibility path.
- Frontend: semantic HTML, CSS, and JavaScript ES modules, using Material Design 3 tokens and accessible interaction patterns.
- SQLite is the planned application store.
- llama.cpp is the primary inference engine, subject to exact artifact/backend/runtime compatibility checks.
- OpenCode is the first supported client.
- Default local Gateway address: `127.0.0.1:18081`.
- Backup artifact: `.omfaibackup` (product name: `omfaibackup`).

## Status

This repository is in the architecture/bootstrap phase. The current work is on `foundation/bootstrap`. The native Wails v3.0.0 build passes on Linux and Windows; Linux live WebView-to-Go bridge/database-reopen smoke tests pass. The earlier Wails monitor archives and Python patch/monitor scripts are audit inputs, not production code.

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

## Linux WebKitGTK signal diagnostic

JavaScriptCore uses `SIGUSR1` (signal 10 on Linux) for GC thread suspend/resume. Depending on the host's existing signal handlers, WebKitGTK may print `Overriding existing handler for signal 10. Set JSC_SIGNAL_FOR_GC if you want WebKit to use a different signal`. Wails v3.0.0 contains the upstream Linux signal-stack compatibility fix: it leaves `SIGUSR1` to JavaScriptCore, reapplies Go's `SA_ONSTACK` requirement to the other relevant signals during initial WebView creation, and reapplies it after `WEBKIT_LOAD_FINISHED`. Do not silence the message by assigning an arbitrary `JSC_SIGNAL_FOR_GC` value; that signal participates in JavaScriptCore's GC coordination. The warning was not reproduced in the current Ubuntu 24.04 CI smoke test, which successfully exercised the bridge twice and reopened the database.

## Current engineering gate

Do not start broad model-catalog implementation or enable remote inference until:
1. the Colab Free usage pattern is policy-reviewed and explicitly permitted;
2. llama.cpp binary distribution is pinned to a tested stable release/backend combination;
3. the exact model/checkpoint/quantization has passed a real load and inference test.

## License

Project licensing will be selected and documented before the first distributable release. Third-party skill materials retain their original licenses and attribution in `.agents/skills/`.
