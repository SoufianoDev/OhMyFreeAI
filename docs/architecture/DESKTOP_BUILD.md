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


## Linux distribution verification matrix

The pinned Wails v3.0.0 proof now has a separate distro-container matrix. The matrix builds the actual application and runs the executable twice under Xvfb; the frontend must call the generated `GetSnapshot` method and produce an explicit smoke marker on each launch.

| Distribution | Web stack | Build mode | CI scope |
|---|---|---|---|
| Ubuntu 24.04 | GTK4 + WebKitGTK 6.0 | Default | Build + WebView bridge + reopen |
| Debian 13 | GTK4 + WebKitGTK 6.0 | Default | Build + WebView bridge + reopen |
| Fedora 44 | GTK4 + WebKitGTK 6.0 | Default | Build + WebView bridge + reopen |
| Arch Linux rolling | GTK4 + WebKitGTK 6.0 | Default | Build + WebView bridge + reopen |
| openSUSE Tumbleweed | GTK4 + WebKitGTK 6.0 | Default | Build + WebView bridge + reopen |
| Ubuntu 22.04 | GTK3 + WebKit2GTK 4.1 | `gtk3` tag | Build + WebView bridge + reopen |
| Debian 12 | GTK3 + WebKit2GTK 4.1 | `gtk3` tag | Build + WebView bridge + reopen |
| Rocky Linux 9 | GTK3 + WebKit2GTK 4.1 | `gtk3` tag | Build + WebView bridge + reopen |

The GTK3 compatibility lane is specific to Wails v3.0.x; upstream schedules removal of the legacy stack in v3.1. These jobs run in distribution containers on GitHub-hosted Linux runners. They validate distro userland package availability, compilation, dynamic linking at launch, the virtual-display WebView and the Go service bridge; they are not bare-metal tests of every desktop environment, GPU driver, Wayland compositor, installer or update path.

The native Ubuntu 24.04 and Windows jobs remain separate. Passing a distro container job does not replace native Windows verification or certify all hardware configurations.
