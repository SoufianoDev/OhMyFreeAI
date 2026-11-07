# Native desktop build

## Toolchain

- Go 1.27.3
- Wails CLI v3.0.0
- Node.js 22 and npm
- Python 3.13
- Linux baseline: Ubuntu 24.04 with build-essential, pkg-config, libgtk-4-dev and libwebkitgtk-6.0-dev
- Windows baseline: Windows WebView2 runtime and the native build prerequisites documented by Wails

Install the pinned Wails CLI:

    go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0

Build the application from the repository root:

    python scripts/build_desktop.py

The script generates the official Wails vanilla-js template in a temporary directory, overlays the application Go source templates from desktop-src and the current frontend, adds a local replace to the root Go module, resolves dependencies, generates JavaScript bindings, runs the Vite production build and builds the native executable.

Output:
- Linux: dist/OhMyFreeAI
- Windows: dist/OhMyFreeAI.exe

The generated workspace lives under .build and is removed after a successful build. dist and .build are ignored by Git. The root Go module does not import Wails; the Wails dependency exists only in the generated desktop module.

## Native service boundary

The desktop app registers one AppService:
- GetSnapshot loads validated persisted configuration and reports database readiness. It currently reports runtime not_connected; no Gateway, Colab runtime or model is started by this adapter.
- SaveConfig delegates to the application configuration service, which validates settings before persistence.

The frontend invokes the generated GetSnapshot binding and only marks the local UI as connected after the native call resolves.

## Verification limits

The GitHub Actions Wails v3.0.0 application build passed on Ubuntu 24.04 and Windows. It generated one service, two methods and seven models, built the Vite production bundle and verified the native executable.

This is a native compile/build and binding-generation proof. It does not yet launch the built WebView and assert a live call, certify graceful shutdown or process ownership, package installers, validate clean uninstall, or establish compatibility with older Linux distributions. macOS remains deferred.
