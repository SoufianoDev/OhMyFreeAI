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

    def test_supported_assessment_requires_all_execution_evidence(self):
        self.db.execute(
            "INSERT INTO model_revisions(id,family,source_repository,revision,architecture,created_at) "
            "VALUES('m','Qwen','org/model','rev1','qwen','now')"
        )
        self.db.execute(
            "INSERT INTO artifacts(id,model_revision_id,format,file_name,size_bytes,sha256,integrity_state,created_at) "
            "VALUES('a','m','gguf','x.gguf',1,?,'verified','now')", ("a" * 64,)
        )
        self.db.execute(
            "INSERT INTO engine_installations(id,version,target_os,target_arch,backend,asset_sha256,install_path,state,created_at) "
            "VALUES('e','v1','linux','amd64','cpu',?,'/tmp/engine','installed','now')", ("b" * 64,)
        )
        with self.assertRaises(sqlite3.IntegrityError):
            self.db.execute(
                "INSERT INTO compatibility_assessments(id,model_revision_id,artifact_id,engine_installation_id,backend,runtime_fingerprint,status,artifact_integrity_verified,engine_load_succeeded,health_check_passed,inference_succeeded,tested_at) "
                "VALUES('c','m','a','e','cpu','cpu-test','supported',1,1,1,0,'now')"
            )

    def test_hash_columns_reject_non_hex_values(self):
        self.db.execute(
            "INSERT INTO model_revisions(id,family,source_repository,revision,architecture,created_at) "
            "VALUES('m','Qwen','org/model','rev1','qwen','now')"
        )
        with self.assertRaises(sqlite3.IntegrityError):
            self.db.execute(
                "INSERT INTO artifacts(id,model_revision_id,format,file_name,size_bytes,sha256,integrity_state,created_at) "
                "VALUES('a','m','gguf','x.gguf',1,?,'verified','now')", ("z" * 64,)
            )

    def test_assessment_artifact_must_match_model_revision(self):
        self.db.execute(
            "INSERT INTO model_revisions(id,family,source_repository,revision,architecture,created_at) VALUES('m1','Qwen','org/model','rev1','qwen','now')"
        )
        self.db.execute(
            "INSERT INTO model_revisions(id,family,source_repository,revision,architecture,created_at) VALUES('m2','Llama','org/model2','rev1','llama','now')"
        )
        self.db.execute(
            "INSERT INTO artifacts(id,model_revision_id,format,file_name,size_bytes,sha256,integrity_state,created_at) VALUES('a1','m1','gguf','x.gguf',1,?,'verified','now')",
            ("a" * 64,),
        )
        self.db.execute(
            "INSERT INTO engine_installations(id,version,target_os,target_arch,backend,asset_sha256,install_path,state,created_at) VALUES('e1','v1','linux','amd64','cpu',?,'/tmp/engine','installed','now')",
            ("b" * 64,),
        )
        with self.assertRaises(sqlite3.IntegrityError):
            self.db.execute(
                "INSERT INTO compatibility_assessments(id,model_revision_id,artifact_id,engine_installation_id,backend,runtime_fingerprint,status,artifact_integrity_verified,engine_load_succeeded,health_check_passed,inference_succeeded,tested_at) VALUES('c1','m2','a1','e1','cpu','cpu-test','unsupported',0,0,0,0,'now')"
            )


if __name__ == "__main__":
    unittest.main()
