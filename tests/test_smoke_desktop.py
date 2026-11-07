from __future__ import annotations

import unittest

from scripts.smoke_desktop import parse_snapshot_marker


class SnapshotMarkerTests(unittest.TestCase):
    def test_accepts_one_successful_bridge_call(self) -> None:
        self.assertEqual(parse_snapshot_marker("GetSnapshot success count=1\n"), 1)

    def test_accepts_multiple_successful_bridge_calls(self) -> None:
        self.assertEqual(parse_snapshot_marker("GetSnapshot success count=3\n"), 3)

    def test_rejects_legacy_unstructured_marker(self) -> None:
        with self.assertRaises(ValueError):
            parse_snapshot_marker("GetSnapshot success\n")

    def test_rejects_zero_or_malformed_counts(self) -> None:
        invalid_values = (
            "GetSnapshot success count=0\n",
            "GetSnapshot success count=x\n",
            "GetSnapshot success count=1\nextra",
        )
        for value in invalid_values:
            with self.subTest(value=value), self.assertRaises(ValueError):
                parse_snapshot_marker(value)


if __name__ == "__main__":
    unittest.main()
