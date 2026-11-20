from __future__ import annotations

import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
RUNBOOK = ROOT / "docs" / "architecture" / "DESKTOP_BUILD.md"


class HostQualificationRunbookTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.runbook = RUNBOOK.read_text(encoding="utf-8")

    def test_linux_cleanup_requires_success_and_explicit_evidence_acknowledgement(self) -> None:
        linux = self.runbook.split("### Linux desktop", 1)[1].split("### Windows desktop", 1)[0]

        self.assertIn("qualification_passed=0", linux)
        self.assertIn(
            'if [ "$status" -eq 0 ] && [ "$qualification_passed" -eq 1 ]; then',
            linux,
        )
        self.assertIn(
            "Qualification failed or was interrupted; preserving test artifacts at %s",
            linux,
        )
        self.assertGreater(linux.index("qualification_passed=1"), linux.index("grep -E"))
        self.assertIn("Do not manually delete a failed qualification directory", linux)

    def test_linux_restore_test_failure_exits_and_preserves_the_test_root(self) -> None:
        linux = self.runbook.split("### Linux desktop", 1)[1].split("### Windows desktop", 1)[0]

        self.assertIn(
            "if ! TMPDIR=\"$test_root\" go test ./internal/infrastructure/sqlite",
            linux,
        )
        self.assertIn(
            "Restore/Journal/NoReplace host tests failed; preserving test artifacts",
            linux,
        )

    def test_windows_native_test_exit_code_is_checked_before_gui_launch(self) -> None:
        windows = self.runbook.split("### Windows desktop", 1)[1].split(
            "Also record the Windows build number", 1
        )[0]

        test_command = windows.index("& go test ./internal/infrastructure/sqlite")
        exit_check = windows.index(
            'if ($LASTEXITCODE -ne 0) { throw "Restore/Journal/NoReplace host tests failed'
        )
        gui_launch = windows.index("& .\\dist\\OhMyFreeAI.exe")

        self.assertLess(test_command, exit_check)
        self.assertLess(exit_check, gui_launch)

    def test_windows_cleanup_preserves_artifacts_after_failure_or_interruption(self) -> None:
        windows = self.runbook.split("### Windows desktop", 1)[1].split(
            "Also record the Windows build number", 1
        )[0]

        self.assertIn("$qualificationPassed = $false", windows)
        self.assertIn("$qualificationPassed = $true", windows)
        self.assertIn("if ($qualificationPassed) {", windows)
        self.assertIn(
            'Write-Warning "Qualification failed or was interrupted; preserving test artifacts at $testRoot"',
            windows,
        )


if __name__ == "__main__":
    unittest.main()
