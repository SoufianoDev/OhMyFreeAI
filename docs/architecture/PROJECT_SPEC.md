# OhMyFreeAI — Project Specification

## Purpose

Provide a desktop control plane that helps a user manage AI model artifacts, inference engines, runtime sessions, a local OpenAI-compatible Gateway, client configuration, observability, recovery and portable application configuration.

## Target platforms

- Linux: first-class, with distro-specific dependency detection and truthful compatibility reporting.
- Windows: first-class, using native process and packaging APIs.
- macOS: deferred; architecture must not make it impossible, but it is not an MVP acceptance target.

## Product constraints

- Free-only. No paid hosting, inference APIs, or required paid service.
- Google Colab Free is a proposed runtime, not a guaranteed or policy-cleared inference host. Its adapter stays gated until the exact interaction model is confirmed compatible with Google's current terms.
- llama.cpp is the primary engine. Alternative engines are not MVP scope.
- OpenCode is the first client. Other coding clients are deferred.
- Go core, modular monolith, Clean Architecture, ports and adapters.
- Wails is a candidate shell. Select v2/v3 only after Linux and Windows proof-of-concept and compatibility review.
- HTML/CSS/JavaScript ES modules for UI; Material 3 design tokens and accessible semantics, not a forced Material Web dependency.
- SQLite is the application database.
- Model downloads should support resume, progress, cancellation and integrity verification. aria2 is preferred only when available and approved; Go-native fallback is required.
- Backup artifact suffix is `.omfaibackup`.

## Readiness semantics

A model is **Ready** only when:
1. the exact artifact is present and integrity-verified;
2. the model/checkpoint and conversion are compatible with the selected engine/backend;
3. measured runtime resources meet the selected operating profile;
4. the engine has loaded the artifact successfully;
5. health checks pass; and
6. a real inference request completes successfully.

A filename, model-family label, successful download, or successful conversion is not sufficient.

## MVP-0 proof gates

- Reproduce the allowed portion of the existing free Colab experiment without policy evasion.
- Prove the Wails shell on supported Linux and Windows environments.
- Prove official pinned llama.cpp distribution on target CPU/GPU environments.
- Prove Gateway OpenAI-compatible behavior, streaming, cancellation and timeouts.
- Prove monitoring cannot block inference or control operations.
- Prove persistent artifact reuse, staging, checksum verification and recovery semantics.
- Prove security boundaries and explicit consent.
- If Colab policy does not permit the required remote inference interaction, do not ship that adapter as a remote inference host.

## MVP-1

Deliver a stable Linux/Windows desktop application only after MVP-0 gates pass. Include GUI/CLI, SQLite, model catalog and compatibility, runtime/engine lifecycle, local Gateway, OpenCode integration, monitoring, local terminal, downloads, artifact storage, recovery and `omfaibackup` export/import.

## Non-goals

Paid providers, account rotation, policy bypasses, arbitrary plugin execution, auto-deleting valid model artifacts, nightly engine builds as production defaults, unverified broad model support, and macOS release in the first milestone.
