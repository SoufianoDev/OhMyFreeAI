# OhMyFreeAI

OhMyFreeAI is a free-first, cross-platform desktop application for managing AI models, inference engines, runtimes, gateways, coding clients, monitoring, downloads, and recovery.

The initial runtime target is Google Colab Free. The initial coding-client integration is OpenCode. Linux and Windows are first-class targets.

## Project status

The repository is at its initial foundation stage. The architecture is defined in **Architecture Foundation v3.3 (7 November 2026)**; implementation and platform validation are still pending. Do not interpret this repository state as a working application.

## Non-negotiable principles

- One shared Go application core with GUI and CLI using the same application services.
- Wails is the desktop framework candidate, subject to a real Linux/Windows feasibility check.
- HTML, CSS, and JavaScript ES modules for the frontend.
- Material Design 3 semantic tokens and adaptive layouts.
- Google Colab Free as the initial remote runtime; no paid service dependency.
- Preserve and audit the previously proven Colab + llama.cpp + local Gateway + OpenCode inference path before changing it.
- No false Ready, Healthy, Compatible, or Verified states.
- Keep the MVP practical for a solo maintainer.

## Documentation

- [UI foundation and Material 3 design contract](docs/UI-FOUNDATION.md)

The full Architecture Foundation v3.3 document is the authoritative product/architecture baseline. Keep implementation decisions aligned with it and record intentional deviations.
