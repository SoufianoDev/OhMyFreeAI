# ADR 0006: Re-evaluate Wails v3.1.0 against the Linux distribution matrix

- Status: Accepted
- Date: 2026-11-08

## Context

ADR 0001 selected Wails v3.0.0 after the application-specific native build and generated-binding proof passed on Ubuntu 24.04 and Windows. The project also maintains a wider Linux validation matrix, including distributions whose supported system repositories provide GTK3 and WebKit2GTK 4.1 but not GTK4 and WebKitGTK 6.0.

Wails v3.1.0 was released on 2026-11-01. Its release notes introduce experimental CGO-free Linux/macOS backends, but do not establish that the legacy GTK3 backend remains supported.

The upstream Wails migration record and Linux packaging documentation state that:
- GTK4 + WebKitGTK 6.0 is the default Linux stack from v3.0.0.
- GTK3 + WebKit2GTK 4.1 is an opt-in legacy stack for the v3.0.x line.
- The GTK3 backend is removed in v3.1.0.
- Wails v3.1.0 requires Go 1.26 or newer.
- The current Wails installation guide identifies Ubuntu 24.04 as its supported Linux baseline; other distributions may work but are not thereby certified.

The project's existing Linux matrix deliberately includes Ubuntu 22.04 and Debian 12 legacy GTK3 lanes. Removing those lanes or silently dropping their coverage would conflict with the project's first-class Linux support goal and the explicit distinction between supported, tested, and certified distributions.

## Decision

Keep **Wails v3.0.0** as the pinned desktop-shell baseline for now.

Do not upgrade the production build or CI pin to Wails v3.1.0 until the project has an approved migration plan for distributions that currently rely on GTK3/WebKit2GTK 4.1.

This is a compatibility decision, not a claim that v3.0.0 is newer or generally preferable. Wails v3.1.0 remains a candidate for a future migration after:
1. Every required Linux distribution in the intended support matrix has a viable GTK4/WebKitGTK 6.0 runtime and build environment, or an explicitly accepted alternative is proven.
2. The native application builds and launches on Linux and Windows with the candidate release.
3. Generated bindings, CLI/GUI dispatch, database lifecycle, shutdown behavior, and the real WebView-to-Go bridge pass.
4. Packaging and runtime dependency checks are completed for the target distributions.
5. The project records any support-matrix changes explicitly rather than treating a successful Ubuntu build as universal Linux support.

## Consequences

- Keep the Wails v3.0.0 version pinned in the Wails proof workflow and generated-template build process.
- Keep the GTK3/WebKit2GTK 4.1 distro lanes while they are supported by the pinned framework version.
- Keep GTK4/WebKitGTK 6.0 lanes for modern distributions.
- Do not use the experimental CGO-free backend as a shortcut around missing native WebKit runtime libraries; Wails documents that users still need the native runtime libraries.
- Revisit this ADR when the Linux support matrix changes or a later Wails release restores a suitable legacy path.
- A passing native build remains distinct from a successful runtime launch, package installation, and certification on a distribution.

## Sources

- Wails v3.1.0 release notes: https://wails.io/blog/wails-v3-1-released/
- Wails changelog: https://v3.wails.io/changelog/
- Wails Linux packaging and legacy GTK3 removal note: https://v3.wails.io/guides/build/linux/
- Upstream GTK4 default / GTK3 removal plan: https://github.com/wailsapp/wails/pull/5463
- Wails supported platforms and dependencies: https://wails.io/getting-started/installation/
