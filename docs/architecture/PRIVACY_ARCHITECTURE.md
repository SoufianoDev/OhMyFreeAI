# Privacy Architecture

- Local-first by default; no analytics/telemetry upload by default.
- Do not persist prompts or full request/response bodies by default.
- Redact API keys, authorization headers, cookies, environment variables and secret-bearing paths.
- SQLite stores metadata and bounded operational history, not secrets.
- Backup archives exclude credentials, model weights, engine binaries and live process state.
- User-visible diagnostics explain which data is collected and where it is stored.
- External service use is disclosed in `docs/legal/THIRD_PARTY_SERVICES.md`.
