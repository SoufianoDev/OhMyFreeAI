import sqlite3
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SCHEMA = ROOT / "migrations" / "001_initial.sql"


class InitialSchemaTests(unittest.TestCase):
    def setUp(self):
        self.db = sqlite3.connect(":memory:")
        self.db.execute("PRAGMA foreign_keys=ON")
        self.db.executescript(SCHEMA.read_text(encoding="utf-8"))

    def tearDown(self):
        self.db.close()

    def test_expected_tables_exist(self):
        tables = {row[0] for row in self.db.execute(
            "SELECT name FROM sqlite_master WHERE type='table'"
        )}
        self.assertTrue({
            "schema_migrations", "app_settings", "model_revisions", "artifacts",
            "artifact_locations", "engine_installations", "compatibility_assessments",
            "runtime_sessions", "download_jobs", "client_integrations", "operational_events",
        }.issubset(tables))

    def test_foreign_keys_are_enforced_by_connection(self):
        with self.assertRaises(sqlite3.IntegrityError):
            self.db.execute(
                "INSERT INTO artifacts(id,model_revision_id,format,file_name,size_bytes,sha256,integrity_state,created_at) "
                "VALUES('a','missing','gguf','x.gguf',1,?,'verified','now')",
                ("0" * 64,),
            )

    def test_invalid_artifact_checksum_is_rejected(self):
        self.db.execute(
            "INSERT INTO model_revisions(id,family,source_repository,revision,architecture,created_at) "
            "VALUES('m','Qwen','org/model','rev1','qwen','now')"
        )
        with self.assertRaises(sqlite3.IntegrityError):
            self.db.execute(
                "INSERT INTO artifacts(id,model_revision_id,format,file_name,size_bytes,sha256,integrity_state,created_at) "
                "VALUES('a','m','gguf','x.gguf',1,'short','verified','now')"
            )


if __name__ == "__main__":
    unittest.main()
