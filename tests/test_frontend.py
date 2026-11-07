import re
import unittest
from html.parser import HTMLParser
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


class IDCollector(HTMLParser):
    def __init__(self):
        super().__init__()
        self.ids = []

    def handle_starttag(self, tag, attrs):
        value = dict(attrs).get("id")
        if value:
            self.ids.append(value)


class FrontendFoundationTests(unittest.TestCase):
    def test_html_has_unique_ids_and_required_pages(self):
        html = (ROOT / "frontend" / "index.html").read_text(encoding="utf-8")
        parser = IDCollector()
        parser.feed(html)
        self.assertEqual(len(parser.ids), len(set(parser.ids)))
        for required in ("page-dashboard", "page-models", "page-runtime", "page-settings"):
            self.assertIn("backup-export", parser.ids)
            self.assertIn("backup-inspect", parser.ids)
            self.assertIn("backup-status", parser.ids)
            self.assertIn(required, parser.ids)
        self.assertRegex(html, r'<html[^>]*lang="ar"[^>]*dir="rtl"')

    def test_stylesheets_have_balanced_rule_blocks(self):
        for relative in ("frontend/styles/tokens.css", "frontend/styles/app.css"):
            css = (ROOT / relative).read_text(encoding="utf-8")
            self.assertEqual(css.count("{"), css.count("}"), relative)

    def test_javascript_is_module_source_without_external_framework(self):
        source = (ROOT / "frontend" / "app.js").read_text(encoding="utf-8")
        self.assertIn("function applyLanguage", source)
        self.assertIn("function applyTheme", source)
        self.assertNotRegex(source, re.compile(r"from\s+['\"](react|vue|svelte)['\"]"))


if __name__ == "__main__":
    unittest.main()
