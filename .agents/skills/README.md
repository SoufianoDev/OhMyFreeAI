# Project-scoped agent skills

These skills are vendored so compatible coding agents can discover them after cloning the repository. Planning with Files includes its shell and PowerShell lifecycle helpers, hook dependencies, templates and plan inventory reference.

| Skill | Source | Version/scope | Application |
|---|---|---|---|
| planning-with-files | https://github.com/OthmanAdi/planning-with-files | v3.28.1 | Persistent plans, findings, progress, validation |
| material-3 | https://github.com/SoufianoDev/material-3-skill | v1.1.1 | Material Design 3 tokens, component semantics, layout, accessibility |
| compose-component-design | https://github.com/SoufianoDev/KotlinComposeKit | Current source skill | Reusable component ownership and caller-controlled content/placement principles only |
| ECC Go patterns | https://github.com/affaan-m/ECC/tree/main/skills/golang-patterns | Current source skill | Idiomatic Go, small interfaces, error handling and package boundaries |
| ECC Go testing | https://github.com/affaan-m/ECC/tree/main/skills/golang-testing | Current source skill | Table-driven tests, subtests, race, coverage and fuzzing |
| ECC database migrations | https://github.com/affaan-m/ECC/tree/main/skills/database-migrations | Current source skill | Immutable, versioned migration and rollback principles; adapt to SQLite |
| ECC security review | https://github.com/affaan-m/ECC/tree/main/skills/security-review | Current source skill | Secrets, validation, SQL, XSS, auth, logging and dependency review |
| ECC TDD workflow | https://github.com/affaan-m/ECC/tree/main/skills/tdd-workflow | Current source skill | RED/GREEN/refactor and evidence reporting for future changes |

## Runtime note

This directory makes skill instructions available to project-aware agents after checkout. It does not itself register user-wide plugins or host lifecycle hooks in ChatGPT. Planning state is therefore also stored explicitly in `.planning/`.

ECC skill materials are MIT-licensed; the original license is retained at `ecc/LICENSE`.

## Third-party licenses

- Material 3 skill: MIT. The original license is retained at `material-3/LICENSE`.
- KotlinComposeKit skill: Apache-2.0. The original license is retained at `compose-component-design/LICENSE`.
- Planning with Files: use the upstream repository's license and attribution terms; source is linked above.
