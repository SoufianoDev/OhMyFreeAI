# ADR 0001: Desktop shell evaluation

- Status: Proposed
- Date: 2026-11-07

## Context

The product requires a Go core and a native desktop shell for Linux and Windows. Wails is a candidate, but the previous prototype uses Wails v2.10.2, which is not a current baseline. Official documentation now presents Wails v3.0.0 as stable and recommends it for new projects. Wails v3 requires Go 1.26+ and GTK4/WebKitGTK 6.0 on Linux by default; older distro stacks require a temporary legacy build tag. Wails v2 continues to receive critical updates and documents Go 1.21+ and GTK3/WebKit dependencies.

The current tool container has Go 1.23.2 and cannot certify a Wails v3 build.

## Decision

Do not lock the application to Wails v2 or v3 yet. Build a minimal Linux/Windows proof-of-concept against the current supported toolchains and document:
- supported Linux distro baseline and dependencies;
- Windows WebView2 behavior;
- packaging and clean uninstall;
- native process/terminal integration;
- event bridge stability and lifecycle behavior;
- accessibility, RTL, dark/light and frontend performance.

## Consequences

- No Wails dependency is added to the core module until the PoC is accepted.
- The domain/application layers remain framework-independent.
- The old monitor GUI is an audit reference, not the UI design baseline.

Sources:
- https://wails.io/getting-started/installation/
- https://v3.wails.io/faq/
- https://v2.wails.io/docs/gettingstarted/installation/
