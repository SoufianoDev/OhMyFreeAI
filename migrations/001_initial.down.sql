-- Development-only rollback for the initial schema before production adoption.
-- Production schema evolution is forward-only; restore a verified backup or add a new migration.

DROP TABLE IF EXISTS operational_events;
DROP TABLE IF EXISTS client_integrations;
DROP TABLE IF EXISTS download_jobs;
DROP TABLE IF EXISTS runtime_sessions;
DROP TABLE IF EXISTS compatibility_assessments;
DROP TABLE IF EXISTS engine_installations;
DROP TABLE IF EXISTS artifact_locations;
DROP TABLE IF EXISTS artifacts;
DROP TABLE IF EXISTS model_revisions;
DROP TABLE IF EXISTS app_settings;
DROP TABLE IF EXISTS schema_migrations;
