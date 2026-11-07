# Project-scoped agent skills

These skills are vendored so compatible coding agents can discover them after cloning the repository. Planning with Files includes its shell and PowerShell lifecycle helpers, hook dependencies, templates and plan inventory reference.

| Skill | Source | Version/scope | Application |
|---|---|---|---|
| planning-with-files | https://github.com/OthmanAdi/planning-with-files | v3.28.1 | Persistent plans, findings, progress, validation |
| material-3 | https://github.com/SoufianoDev/material-3-skill | v1.1.1 | Material Design 3 tokens, component semantics, layout, accessibility |
| compose-component-design | https://github.com/SoufianoDev/KotlinComposeKit | Current source skill | Reusable component ownership and caller-controlled content/placement principles only |

## Runtime note

This directory makes skill instructions available to project-aware agents after checkout. It does not itself register user-wide plugins or host lifecycle hooks in ChatGPT. Planning state is therefore also stored explicitly in `.planning/`.

## Third-party licenses

- Material 3 skill: MIT. The original license is retained at `material-3/LICENSE`.
- KotlinComposeKit skill: Apache-2.0. The original license is retained at `compose-component-design/LICENSE`.
- Planning with Files: use the upstream repository's license and attribution terms; source is linked above.
