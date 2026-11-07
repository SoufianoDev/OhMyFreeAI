-- Initial OhMyFreeAI metadata schema. The migration runner records the applied version.
-- Secrets, prompts, full inference payloads and model weights never belong in this DB.

CREATE TABLE schema_migrations (
    version       INTEGER PRIMARY KEY,
    description   TEXT NOT NULL,
    applied_at    TEXT NOT NULL
);

CREATE TABLE app_settings (
    singleton_id  INTEGER PRIMARY KEY CHECK (singleton_id = 1),
    schema_version INTEGER NOT NULL CHECK (schema_version > 0),
    config_json   TEXT NOT NULL,
    updated_at    TEXT NOT NULL
);

CREATE TABLE model_revisions (
    id                  TEXT PRIMARY KEY,
    family              TEXT NOT NULL,
    source_repository   TEXT NOT NULL,
    revision            TEXT NOT NULL,
    architecture        TEXT NOT NULL,
    tokenizer_revision  TEXT NOT NULL DEFAULT '',
    license_identifier  TEXT NOT NULL DEFAULT '',
    created_at          TEXT NOT NULL,
    UNIQUE (source_repository, revision)
);

CREATE TABLE artifacts (
    id              TEXT PRIMARY KEY,
    model_revision_id TEXT NOT NULL REFERENCES model_revisions(id) ON DELETE RESTRICT,
    format          TEXT NOT NULL,
    quantization    TEXT NOT NULL DEFAULT '',
    file_name       TEXT NOT NULL,
    size_bytes      INTEGER NOT NULL CHECK (size_bytes >= 0),
    sha256          TEXT NOT NULL CHECK (length(sha256) = 64 AND sha256 NOT GLOB '*[^0-9a-fA-F]*'),
    source_uri_sanitized TEXT NOT NULL DEFAULT '',
    integrity_state TEXT NOT NULL CHECK (integrity_state IN ('unknown', 'pending', 'verified', 'failed')),
    created_at      TEXT NOT NULL,
    UNIQUE (sha256),
    UNIQUE (id, model_revision_id)
);

CREATE TABLE artifact_locations (
    artifact_id     TEXT NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    store_kind      TEXT NOT NULL CHECK (store_kind IN ('local', 'drive', 'other')),
    location_ref    TEXT NOT NULL,
    verified_at     TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (artifact_id, store_kind, location_ref)
);

CREATE TABLE engine_installations (
    id              TEXT PRIMARY KEY,
    version         TEXT NOT NULL,
    target_os       TEXT NOT NULL,
    target_arch     TEXT NOT NULL,
    backend         TEXT NOT NULL,
    asset_sha256    TEXT NOT NULL CHECK (length(asset_sha256) = 64 AND asset_sha256 NOT GLOB '*[^0-9a-fA-F]*'),
    install_path    TEXT NOT NULL,
    state           TEXT NOT NULL CHECK (state IN ('verified', 'installed', 'unhealthy', 'removed')),
    created_at      TEXT NOT NULL,
    UNIQUE (version, target_os, target_arch, backend)
);

CREATE TABLE compatibility_assessments (
    id                          TEXT PRIMARY KEY,
    model_revision_id           TEXT NOT NULL REFERENCES model_revisions(id) ON DELETE RESTRICT,
    artifact_id                 TEXT NOT NULL,
    engine_installation_id      TEXT NOT NULL REFERENCES engine_installations(id) ON DELETE RESTRICT,
    backend                     TEXT NOT NULL,
    runtime_fingerprint         TEXT NOT NULL,
    status                      TEXT NOT NULL CHECK (status IN ('supported', 'supported_with_limits', 'unsupported', 'unverified', 'resource_ineligible', 'policy_blocked')),
    artifact_integrity_verified INTEGER NOT NULL CHECK (artifact_integrity_verified IN (0, 1)),
    engine_load_succeeded       INTEGER NOT NULL CHECK (engine_load_succeeded IN (0, 1)),
    health_check_passed         INTEGER NOT NULL CHECK (health_check_passed IN (0, 1)),
    inference_succeeded         INTEGER NOT NULL CHECK (inference_succeeded IN (0, 1)),
    reason_code                 TEXT NOT NULL DEFAULT '',
    minimum_requirements_json   TEXT NOT NULL DEFAULT '{}',
    tested_at                   TEXT NOT NULL,
    expires_at                  TEXT NOT NULL DEFAULT '',
    CHECK (status NOT IN ('supported', 'supported_with_limits') OR
        (artifact_integrity_verified = 1 AND engine_load_succeeded = 1 AND
         health_check_passed = 1 AND inference_succeeded = 1)),
    FOREIGN KEY (artifact_id, model_revision_id) REFERENCES artifacts(id, model_revision_id) ON DELETE RESTRICT
);

CREATE TABLE runtime_sessions (
    id              TEXT NOT NULL,
    generation      INTEGER NOT NULL CHECK (generation > 0),
    provider        TEXT NOT NULL,
    state           TEXT NOT NULL,
    created_at      TEXT NOT NULL,
    updated_at      TEXT NOT NULL,
    PRIMARY KEY (id, generation)
);

CREATE TABLE download_jobs (
    id                  TEXT PRIMARY KEY,
    artifact_id         TEXT REFERENCES artifacts(id) ON DELETE SET NULL,
    source_uri_sanitized TEXT NOT NULL,
    destination_ref     TEXT NOT NULL,
    expected_size       INTEGER NOT NULL DEFAULT 0 CHECK (expected_size >= 0),
    expected_sha256     TEXT NOT NULL DEFAULT '' CHECK (expected_sha256 = '' OR (length(expected_sha256) = 64 AND expected_sha256 NOT GLOB '*[^0-9a-fA-F]*')),
    bytes_received      INTEGER NOT NULL DEFAULT 0 CHECK (bytes_received >= 0),
    state               TEXT NOT NULL CHECK (state IN ('queued', 'preparing', 'downloading', 'verifying', 'completed', 'paused', 'cancelling', 'cancelled', 'failed', 'integrity_error')),
    created_at          TEXT NOT NULL,
    updated_at          TEXT NOT NULL
);

CREATE TABLE client_integrations (
    id                  TEXT PRIMARY KEY,
    client_name         TEXT NOT NULL,
    config_path_ref     TEXT NOT NULL,
    backup_reference    TEXT NOT NULL DEFAULT '',
    state               TEXT NOT NULL,
    updated_at          TEXT NOT NULL,
    UNIQUE (client_name, config_path_ref)
);

CREATE TABLE operational_events (
    id              TEXT PRIMARY KEY,
    entity_type     TEXT NOT NULL,
    entity_id       TEXT NOT NULL,
    generation      INTEGER NOT NULL DEFAULT 0,
    event_type      TEXT NOT NULL,
    occurred_at     TEXT NOT NULL,
    details_redacted_json TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_artifacts_model_revision ON artifacts(model_revision_id);
CREATE INDEX idx_assessments_model_artifact ON compatibility_assessments(model_revision_id, artifact_id, tested_at DESC);
CREATE INDEX idx_sessions_provider_state ON runtime_sessions(provider, state, updated_at DESC);
CREATE INDEX idx_download_jobs_state ON download_jobs(state, updated_at DESC);
CREATE INDEX idx_events_entity_time ON operational_events(entity_type, entity_id, occurred_at DESC);
