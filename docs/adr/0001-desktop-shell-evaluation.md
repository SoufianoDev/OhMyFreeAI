# ADR 0001: Desktop shell evaluation

- Status: Accepted
- Date: 2026-11-07

## Context

The product requires a Go core and a native desktop shell for Linux and Windows. Wails v3.0.0 is the selected stable baseline for new applications. It requires Go 1.26+ and GTK4/WebKitGTK 6.0 on the default Linux build path. The project toolchain target is Go 1.27.3.

The initial official vanilla-js starter and generated binding proof passed on Ubuntu 24.04 and Windows. The application-specific integration has now also been built on both platforms using the official Wails template.

## Decision

Select **Wails v3.0.0** as the desktop shell for the Linux and Windows application.

The repository keeps the framework boundary narrow:
- Go source templates live in desktop-src/ and are copied into a generated Wails project.
- scripts/build_desktop.py creates the official vanilla-js template in a temporary workspace, overlays the actual app source and frontend, adds a local replace to the root Go module, generates bindings and builds the native executable.
- The root Go module remains free of Wails dependencies; domain/application code does not import Wails.
- The app registers one AppService with GetSnapshot and SaveConfig. The generated binding pass reported 1 service, 2 methods and 7 models.
- The frontend calls GetSnapshot and only displays native connectivity after that call resolves.

Application-specific Wails run 42396866046 passed on ubuntu-24.04 and windows-latest. Both generated bindings, built the Vite production bundle and produced the expected native executable.

## Verified baseline and remaining gates

- Ubuntu 24.04 with GTK4 and WebKitGTK 6.0 development packages is the only Linux build baseline certified so far.
- Windows native executable build passed on the hosted WebView2 environment.
- The CI build verifies service binding generation and native compilation; it does not yet launch the WebView and assert a live call at runtime.
- Process ownership, graceful shutdown/cancellation, terminal integration, packaging, installer/uninstall, and older Linux distribution support remain open.
- macOS remains deferred.
- The Windows Wails build printed a non-fatal "-VThe system cannot find the file specified." line before its Taskfile steps. Binding generation, Vite build and executable creation still succeeded. The origin is not yet established.

## Consequences

- Keep Wails imports at the desktop/platform boundary.
- Keep the official Wails template and version pinned in the build workflow.
- Do not equate a successful native build with a runtime or packaging certificate.
- The old monitor GUI remains an audit reference, not the UI design baseline.

Sources:
- https://wails.io/getting-started/installation/
- https://v3.wails.io/faq/
- https://github.com/wailsapp/wails/tree/v3.0.0/v3/internal/templates
