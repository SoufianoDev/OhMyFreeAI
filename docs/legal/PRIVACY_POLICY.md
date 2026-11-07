# OhMyFreeAI — Privacy Policy (Draft)

**Draft for review. Not legal advice. Must be aligned with actual implementation before release.**

## Data handled locally

The planned application stores user preferences, model/engine metadata, runtime configuration, client references, operational state and bounded diagnostics in a local SQLite database and application files.

## Sensitive data

- Credentials and tokens must be stored only through OS-backed secret storage where available.
- Secret values, prompts, full request bodies and authorization headers must not be written to logs, SQLite, backups or telemetry by default.
- Backup files exclude credentials and model weights.

## External services

When enabled, the application may communicate with user-selected model artifact sources, Google services, and OpenCode-related local configuration. The exact destinations and data flows must be disclosed in the final service inventory.

## Telemetry

No analytics or remote telemetry should be enabled by default. Any future telemetry requires explicit disclosure and opt-in.

## User control

Users must be able to inspect, export and delete local application metadata without deleting their model artifacts. The implementation must document exactly what is and is not removed.

This draft must be updated after data-flow mapping and legal review. It is not a final privacy commitment.
