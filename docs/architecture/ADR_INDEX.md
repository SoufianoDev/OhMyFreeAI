# Architecture Decision Records

| ADR | Decision | Status |
|---|---|---|
| [0001 — Desktop shell evaluation](../adr/0001-desktop-shell-evaluation.md) | Select Wails v3.0.0; native app and binding builds pass on Ubuntu/Windows, while live WebView testing and distro coverage remain open | Accepted |
| [0002 — Colab Free policy gate](../adr/0002-colab-free-policy-gate.md) | No remote inference adapter activation until the exact flow is policy-cleared | Accepted |
| [0003 — Backup artifact naming](../adr/0003-omfaibackup-format.md) | Use `.omfaibackup`; preserve legacy import compatibility | Accepted |
| [0004 — SQLite driver evaluation](../adr/0004-sqlite-driver-evaluation.md) | Select modernc.org/sqlite v1.62.0; production Store/migrations/config repository are implemented, while crash recovery and backup API remain open | Accepted |
| [0005 — Go toolchain baseline](../adr/0005-go-toolchain-baseline.md) | Target latest stable patch, Go 1.27.3 as of 2026-11-07 | Accepted |


## Build and implementation guides

- [Crash-safe SQLite restore lifecycle](RESTORE_LIFECYCLE.md)

- [Native desktop build](DESKTOP_BUILD.md)
- [Historical prototype audit](PROTOTYPE_AUDIT.md)
- [Implementation blueprint](IMPLEMENTATION_BLUEPRINT.md)
