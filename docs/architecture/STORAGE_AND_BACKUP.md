# Storage and Backup

## SQLite

SQLite is the application-owned database. Use schema migrations, foreign keys, transactions, integrity checks and WAL where appropriate. Persist metadata, state transitions, bounded event history and artifact manifests. Do not store prompts/full request bodies or secrets by default.

## Artifact storage

Persistent store contains verified model artifacts, engine distributions, manifests, checksums and download metadata. Runtime staging copies to local ephemeral disk and re-verifies before load. Never execute directly from Google Drive.

## Downloads

Support queue, progress, pause/cancel, resumable transfers where the source supports it, atomic finalization and integrity verification. aria2 may be used if installed/approved; Go-native fallback is mandatory. Do not silently install aria2.

## omfaibackup

- User-facing name: `omfaibackup`.
- File suffix: `.omfaibackup`.
- Versioned archive with a manifest and checksums.
- Includes app configuration, preferences, model/engine metadata, runtime/client references, and recoverable state.
- Excludes model weights, engine binaries, credentials, tokens, live process state, active sockets and transient telemetry buffers.
- Restore validates schema/version/checksums and all referenced paths; then migrates config, re-discovers services, checks health, loads a model and performs real inference before reporting full recovery.
- Preserve a safe legacy import path for prior `sdevbackup` backups; do not silently rewrite or delete originals.
- Never claim restore success merely because archive extraction succeeded.
