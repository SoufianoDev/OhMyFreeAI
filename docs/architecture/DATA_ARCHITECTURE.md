# Data Architecture

SQLite is the authoritative store for application metadata, persisted settings and state. The adapter uses `database/sql`; the concrete SQLite driver remains isolated behind the persistence adapter pending platform validation. Schema changes use ordered migrations and transactions. Foreign keys and integrity checks are enabled; WAL is used where appropriate for concurrent readers.

Persist model revisions, artifact manifests, compatibility assessments, runtime sessions, engine installations, client configuration snapshots, download jobs and bounded audit events.

Do not persist model tensors, prompts, complete inference payloads, secrets, transient telemetry streams or process handles. Filesystem artifacts are referenced by stable IDs and verified checksums.
