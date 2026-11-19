from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

from scripts.docs_check import find_broken_local_links, markdown_files


class MarkdownLinkCheckTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary = tempfile.TemporaryDirectory()
        self.root = Path(self.temporary.name).resolve()

    def tearDown(self) -> None:
        self.temporary.cleanup()

    def write(self, relative: str, content: str = "") -> Path:
        path = self.root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")
        return path

    def test_accepts_relative_root_image_and_percent_encoded_targets(self) -> None:
        self.write("README.md", "# Project\n")
        self.write("assets/logo image.svg", "<svg/>")
        self.write(
            "docs/guide.md",
            "[root](/README.md) [readme](../README.md) "
            "![logo](../assets/logo%20image.svg) [section](#local-section)\n",
        )

        self.assertEqual(find_broken_local_links(self.root), [])

    def test_resolves_links_from_nested_planning_documents(self) -> None:
        self.write("docs/architecture/DESKTOP_BUILD.md", "# Desktop build" + chr(10))
        self.write(
            ".planning/2026-11-07-ohmyfreeai-foundation/task_plan.md",
            "[valid](../../docs/architecture/DESKTOP_BUILD.md)"
            + chr(10)
            + "[too-shallow](../docs/architecture/DESKTOP_BUILD.md)"
            + chr(10),
        )

        errors = find_broken_local_links(self.root)

        self.assertEqual(len(errors), 1)
        self.assertIn(".planning/2026-11-07-ohmyfreeai-foundation/task_plan.md:2", errors[0])
        self.assertIn("../docs/architecture/DESKTOP_BUILD.md", errors[0])

    def test_reports_missing_local_target_with_source_line(self) -> None:
        self.write("docs/guide.md", "# Guide\n\n[missing](not-here.md)\n")

        errors = find_broken_local_links(self.root)

        self.assertEqual(len(errors), 1)
        self.assertIn("docs/guide.md:3", errors[0])
        self.assertIn("not-here.md", errors[0])

    def test_rejects_links_that_escape_repository_root(self) -> None:
        self.write("docs/guide.md", "[outside](../../outside.md)\n")

        errors = find_broken_local_links(self.root)

        self.assertEqual(len(errors), 1)
        self.assertIn("escapes the repository root", errors[0])

    def test_checks_reference_definitions_and_html_attributes(self) -> None:
        self.write(
            "docs/guide.md",
            "[reference]: missing-reference.md\n"
            '<a href="missing-html-target.md">link</a>\n',
        )

        errors = find_broken_local_links(self.root)

        self.assertEqual(len(errors), 2)
        self.assertTrue(any("missing-reference.md" in error for error in errors))
        self.assertTrue(any("missing-html-target.md" in error for error in errors))

    def test_ignores_external_links_and_code_examples(self) -> None:
        inline_code = chr(96) + "[inline](missing-inline.md)" + chr(96)
        self.write(
            "docs/guide.md",
            "[external](https://example.com/path) [email](mailto:test@example.com)\n"
            + inline_code
            + "\n~~~markdown\n[block](missing-block.md)\n~~~\n",
        )

        self.assertEqual(find_broken_local_links(self.root), [])

    def test_excludes_vendored_agent_skill_markdown(self) -> None:
        self.write("README.md", "# Project\n")
        self.write(".agents/skills/vendor/README.md", "[upstream](missing-upstream-file.md)\n")

        files = markdown_files(self.root)

        self.assertEqual([path.relative_to(self.root).as_posix() for path in files], ["README.md"])
        self.assertEqual(find_broken_local_links(self.root, files), [])


if __name__ == "__main__":
    unittest.main()
