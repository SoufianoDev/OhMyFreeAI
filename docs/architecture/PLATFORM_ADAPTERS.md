# Platform Adapters

## Contract

Platform-specific behavior is isolated behind ports. Domain/application packages must not import OS APIs directly.

## Linux

- Process identity: /proc plus pidfd where available; verify executable, start time and ownership before signaling.
- Terminal: native PTY adapter with resize, signal, exit status and stream lifecycle.
- Secrets: Secret Service/libsecret when available; provide a clear fallback that never writes secrets to plaintext logs or backups.
- Desktop integration: XDG paths, desktop entry, icons, native file dialogs and graceful shutdown.
- Detect GTK/WebKit dependencies for the selected Wails version; do not silently install system packages.

## Windows

- Process identity: process handles and Job Objects; do not rely on PID alone.
- Terminal: ConPTY where supported; explicit capability fallback.
- Secrets: Windows Credential Manager/DPAPI-backed storage.
- Desktop integration: WebView2, native dialogs, installer/uninstaller and per-user data paths.

## Capability discovery

Each adapter returns typed capabilities and missing prerequisites. The UI reports exact requirements and remediation commands without running privileged changes automatically.
