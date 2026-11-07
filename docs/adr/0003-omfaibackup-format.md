# ADR 0003: omfaibackup format and naming

- Status: Accepted
- Date: 2026-11-07

## Context

The user explicitly selected `omfaibackup` as the application backup name instead of `sdevbackup`.

## Decision

Use `.omfaibackup` as the file suffix and `omfaibackup` as the format/product identifier. The format is versioned and manifest-driven with per-entry integrity checks.

Include app configuration, preferences, model/engine metadata and recoverable references. Exclude model weights, engine binaries, secrets, tokens, live processes, sockets and transient runtime state.

Keep a read-only/validated legacy import path for existing `sdevbackup` files where safely identifiable. Never delete or silently rewrite the original backup.

## Consequences

- Restore must validate archive structure, schema, checksums, compatibility and references.
- Full restore success requires healthy services and a real inference test, not successful extraction alone.
