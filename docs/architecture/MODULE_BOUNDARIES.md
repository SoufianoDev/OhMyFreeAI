# Module Boundaries

| Module | Owns | Does not own |
|---|---|---|
| configuration | schema, precedence, validation, migrations | secrets in plain text |
| persistence | SQLite schema, migrations, repositories | domain decisions |
| models | model identity, artifact metadata, manifests | claiming runtime support |
| compatibility | architecture/checkpoint/format/backend/resource evaluation | downloading or starting engines |
| downloads | queues, progress, resume, cancellation, integrity | model readiness |
| storage | persistent artifact store and local staging | executing from mounted Drive |
| runtime | session lifecycle, capabilities, health and policy gate | engine-specific flags |
| engine | llama.cpp distribution, configuration, process lifecycle | runtime allocation policy |
| gateway | OpenAI-compatible HTTP/SSE, auth, limits, cancellation | client-specific settings |
| monitoring | telemetry snapshots, freshness, event history | control-plane decisions |
| terminal | local PTY and explicit remote capability labels | pretending command session is a PTY |
| clients | OpenCode discovery/config backup/atomic update/rollback | modifying unrelated user settings |
| recovery | failure classification, reconnect, bounded retry | blind request replay |
| artifacts | omfaibackup schema, export/import validation | model weights or secrets |
| process | process identity, ownership, signals, handles | killing by PID alone |
| security | local auth, secret references, consent and audit | logging credentials |
| platform | Linux/Windows OS adapters | domain state |

Modules communicate through typed application interfaces and domain events. Avoid circular imports and direct adapter-to-adapter calls. Cross-module workflows belong in application services.
