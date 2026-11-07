# ADR 0001: Desktop shell evaluation

- Status: Accepted
- Date: 2026-11-07

## Context

The product requires a Go core and a native desktop shell for Linux and Windows. Wails is a candidate, but the previous prototype uses Wails v2.10.2, which is not a current baseline. Official documentation now presents Wails v3.0.0 as stable and recommends it for new projects. Wails v3 requires Go 1.26+ and GTK4/WebKitGTK 6.0 on Linux by default; older distro stacks require a temporary legacy build tag. Wails v2 continues to receive critical updates and documents Go 1.21+ and GTK3/WebKit dependencies.

The project toolchain target is Go 1.27.3, which exceeds the documented Wails v3 Go 1.26+ requirement. GitHub Actions verified the pinned Wails v3.0.0 official vanilla-js starter on Ubuntu 24.04 and Windows using Go 1.27.3 and Node 22. The test generated the official GreetService JavaScript binding, asserted the exported Greet method, built the native executable and checked the output file on each OS.

## Decision

Select **Wails v3.0.0** as the desktop shell for the Linux and Windows application.

The native proof passed on both Ubuntu 24.04 and Windows. The Linux baseline is Ubuntu 24.04 with GTK4 and WebKitGTK 6.0 development/runtime dependencies; the Windows proof uses the hosted WebView2 environment. This does not establish support for older Linux distributions or macOS.

The Go domain and application layers remain independent of Wails. Wails belongs only in the desktop/platform adapter. The actual application service registration, WebView invocation, lifecycle/shutdown integration, native process/terminal integration, packaging and uninstall still require implementation and dedicated tests. The proof built the official starter, not the complete OhMyFreeAI application.

## Consequences

- Keep Wails imports at the desktop/platform boundary; do not leak framework types into domain/application packages.
- Add the Wails Go dependency as part of the actual desktop adapter implementation, not to domain packages.
- Ubuntu 24.04 is the only Linux baseline certified so far; older distro compatibility and packaging remain open.
- The old monitor GUI is an audit reference, not the UI design baseline.

Sources:
- https://wails.io/getting-started/installation/
- https://v3.wails.io/faq/
- https://v2.wails.io/docs/gettingstarted/installation/
