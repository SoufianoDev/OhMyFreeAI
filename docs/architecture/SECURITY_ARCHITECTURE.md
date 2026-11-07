# Security Architecture

## Trust boundaries

- The desktop UI is not a trust boundary. All commands are validated by Go application services.
- Domain and application packages do not depend on Wails, database drivers, HTTP handlers or platform APIs.
- Remote runtime providers are untrusted until their policy, identity and capabilities are explicitly assessed.
- The local Gateway binds to loopback only and is disabled by default until authentication is configured.

## Process and network safety

- Never signal a process by PID alone. Verify process identity, executable, start time and ownership.
- Bind local Gateway and control interfaces to parsed loopback addresses. Reject 0.0.0.0 and non-loopback IPv6/IPv4.
- Do not implement tunnels, SSH relays, notebook UI bypasses or remote-control workarounds for Colab Free.
- Enforce separate Gateway and engine ports.

## Secrets

- Store secret values only in the platform secret store.
- Persist OS secret references, never secret values.
- Reject raw tokens and secret references containing userinfo, query parameters or fragments.
- Redact authorization headers, cookies, tokens, signed URLs and secret-bearing environment variables from logs and diagnostics.

## SQLite and artifacts

- SQLite stores metadata, settings with secret references, compatibility evidence and bounded operational events.
- Do not store prompts, full inference payloads, model weights, engine binaries or credentials.
- Sanitize source URIs before persistence. Store only redacted event detail JSON.
- Enable and verify `PRAGMA foreign_keys=ON` for every database connection.
- Keep the live SQLite database on a supported local filesystem; do not place it on a Google Drive mount or an unverified network filesystem.
- Verify artifact and engine checksums before use. Never execute from persistent Drive storage.

## Policy evidence

A remote inference connection fails closed unless a policy assessment contains provider, approved decision, explicit scope, policy revision, review timestamp and an HTTPS source URL without user information, query parameters or fragments. This data model does not itself constitute legal approval; a provider-specific review must be performed before recording approval.

## Backups and diagnostics

- `.omfaibackup` contains recoverable settings and metadata references only, not secrets, model weights, engine binaries or process state.
- Diagnostics are opt-in where they could expose local paths or provider details; redact by default.
- No prompts or complete request/response bodies are logged by default.

## Bootstrap controls implemented

- Gateway disabled by default and loopback-only.
- Typed OS secret-reference validation and distinct Gateway/engine ports.
- Fail-closed runtime policy gate with strict source URL parsing.
- SQLite schema constraints for checksum format, model/artifact identity and evidence completeness.
