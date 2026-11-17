from __future__ import annotations

import re
from pathlib import Path
from urllib.parse import unquote, urlsplit

BACKTICK = chr(96)
EXCLUDED_DIRECTORIES = {".agents", ".build", ".git", "dist", "node_modules", ".venv"}
INLINE_LINK_START = re.compile(r"!?(?:\[[^\]\n]*\])\(")
REFERENCE_DEFINITION = re.compile(
    r"(?m)^[ \t]{0,3}\[[^\]\n]+\]:[ \t]*(<[^>\n]+>|[^\s]+)"
)
HTML_LINK_ATTRIBUTE = re.compile(r"""\b(?:href|src)\s*=\s*(['"])(.*?)\1""", re.IGNORECASE | re.DOTALL)


def markdown_files(root: Path) -> list[Path]:
    root = root.resolve()
    return sorted(
        path
        for path in root.rglob("*.md")
        if path.is_file()
        and not any(part in EXCLUDED_DIRECTORIES for part in path.relative_to(root).parts)
    )


def _without_code(text: str) -> str:
    output: list[str] = []
    fence_character: str | None = None
    fence_length = 0

    for line in text.splitlines(keepends=True):
        marker = re.match(r"^[ \t]{0,3}(" + BACKTICK + r"{3,}|~{3,})", line)
        if fence_character is None:
            if marker:
                fence_character = marker.group(1)[0]
                fence_length = len(marker.group(1))
                output.append("\n" if line.endswith("\n") else "")
                continue
            output.append(line)
            continue

        if (
            marker
            and marker.group(1)[0] == fence_character
            and len(marker.group(1)) >= fence_length
        ):
            fence_character = None
            fence_length = 0
        output.append("\n" if line.endswith("\n") else "")

    cleaned = "".join(output)
    cleaned = re.sub(
        r"<!--.*?-->",
        lambda match: "\n" * match.group(0).count("\n"),
        cleaned,
        flags=re.DOTALL,
    )
    return re.sub(r"(" + BACKTICK + r"+).*?\1", "", cleaned)


def _inline_destinations(text: str):
    for match in INLINE_LINK_START.finditer(text):
        position = match.end()
        while position < len(text) and text[position].isspace():
            position += 1
        if position >= len(text):
            continue

        if text[position] == "<":
            end = position + 1
            escaped = False
            while end < len(text):
                character = text[end]
                if character == ">" and not escaped:
                    break
                if character == "\\" and not escaped:
                    escaped = True
                else:
                    escaped = False
                end += 1
            if end < len(text):
                yield text[position + 1 : end], match.start()
            continue

        destination: list[str] = []
        depth = 0
        while position < len(text):
            character = text[position]
            if character == "\\" and position + 1 < len(text):
                destination.append(text[position + 1])
                position += 2
                continue
            if character == "(":
                depth += 1
            elif character == ")":
                if depth == 0:
                    break
                depth -= 1
            elif character.isspace() and depth == 0:
                break
            destination.append(character)
            position += 1

        if destination:
            yield "".join(destination), match.start()


def _destinations(text: str):
    yield from _inline_destinations(text)
    for match in REFERENCE_DEFINITION.finditer(text):
        destination = match.group(1)
        if destination.startswith("<") and destination.endswith(">"):
            destination = destination[1:-1]
        yield destination, match.start()
    for match in HTML_LINK_ATTRIBUTE.finditer(text):
        yield match.group(2), match.start()


def find_broken_local_links(root: Path, files: list[Path] | None = None) -> list[str]:
    root = root.resolve()
    selected_files = files if files is not None else markdown_files(root)
    errors: list[str] = []

    for source in selected_files:
        try:
            text = source.read_text(encoding="utf-8")
        except (OSError, UnicodeError) as error:
            errors.append(f"{source.relative_to(root).as_posix()}: cannot read Markdown file: {error}")
            continue

        cleaned = _without_code(text)
        for destination, offset in _destinations(cleaned):
            destination = destination.strip()
            if not destination or destination.startswith("//"):
                continue

            parsed = urlsplit(destination)
            if parsed.scheme:
                continue

            relative_path = unquote(parsed.path)
            if not relative_path:
                continue

            target = (
                root / relative_path.lstrip("/")
                if relative_path.startswith("/")
                else source.parent / relative_path
            ).resolve(strict=False)

            try:
                target.relative_to(root)
            except ValueError:
                reason = "link escapes the repository root"
            else:
                if target.exists():
                    continue
                reason = "local link target does not exist"

            line = cleaned.count("\n", 0, offset) + 1
            errors.append(
                f"{source.relative_to(root).as_posix()}:{line}: {reason}: {destination!r}"
            )

    return errors
