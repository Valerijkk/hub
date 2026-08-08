#!/usr/bin/env python3
"""Validate the portable pro-qa-ai toolkit with Python stdlib only.

The validator intentionally reports locations and rule identifiers, not matched
secret values or source lines. Exit codes: 0 = pass, 1 = findings, 2 = usage or
configuration error.
"""

from __future__ import annotations

import argparse
import difflib
import hashlib
import json
import ntpath
import os
import re
import sys
from dataclasses import asdict, dataclass
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, Iterable, List, Optional, Sequence, Set, Tuple
from urllib.parse import unquote, urlsplit

from safe_filesystem import (
    FileSafetyError,
    InspectedText,
    SafeTree,
    TextPolicy,
    lexical_absolute,
    parse_text_policy,
    resolve_directory_root,
)
from secret_rules import secret_rule_matches


DEFAULT_ROOT = Path(__file__).resolve().parents[1]
DEFAULT_CONFIG_NAME = "validation_manifest.json"
CHECK_ORDER = (
    "file_policy",
    "links",
    "inline_paths",
    "frontmatter",
    "skills",
    "prompt_ids",
    "placeholders",
    "cot_disclosure",
    "gsd_routes",
    "secrets",
    "portable_leakage",
    "trailing_whitespace",
    "contracts",
    "provenance",
    "recipes",
    "agent_contract",
    "skill_contract",
    "size_budget",
    "vocabulary",
    "enums",
    "roster",
    "code_spans",
    "duplication",
    "handoff",
    "egress",
)


_BLANK_LINE = re.compile(chr(10) + r"\s*" + chr(10))


@dataclass(frozen=True)
class Issue:
    check: str
    code: str
    severity: str
    path: str
    line: int
    message: str


@dataclass
class Frontmatter:
    present: bool
    fields: Dict[str, str]
    lines: Dict[str, int]
    errors: List[Tuple[int, str]]
    end_line: int = 0


@dataclass
class YamlLike:
    fields: Dict[str, str]
    lines: Dict[str, int]
    errors: List[Tuple[int, str]]
    list_items: Dict[str, List[str]]
    scalar_types: Dict[str, str]


def _matches(path: str, patterns: Sequence[str]) -> bool:
    normalized = path.replace("\\", "/")
    for pattern in patterns:
        candidate = pattern.replace("\\", "/")
        if re.fullmatch(_glob_expression(candidate), normalized):
            return True
    return False


def _glob_expression(pattern: str) -> str:
    """Translate a portable glob where '*' never crosses a path separator."""
    output: List[str] = []
    index = 0
    while index < len(pattern):
        character = pattern[index]
        if pattern[index : index + 3] == "**/":
            output.append("(?:.*/)?")
            index += 3
            continue
        if pattern[index : index + 2] == "**":
            output.append(".*")
            index += 2
            continue
        if character == "*":
            output.append("[^/]*")
        elif character == "?":
            output.append("[^/]")
        elif character == "[":
            closing = pattern.find("]", index + 1)
            if closing == -1:
                output.append(r"\[")
            else:
                content = pattern[index + 1 : closing]
                if content.startswith("!"):
                    content = "^" + content[1:]
                output.append("[" + content.replace("\\", r"\\") + "]")
                index = closing
        else:
            output.append(re.escape(character))
        index += 1
    return "".join(output)


def _line_number(text: str, offset: int) -> int:
    return text.count("\n", 0, offset) + 1


def _strip_inline_comment(value: str) -> str:
    value = value.strip()
    if len(value) >= 2 and value[0] == value[-1] and value[0] in {'"', "'"}:
        return value[1:-1]
    return re.split(r"\s+#", value, maxsplit=1)[0].strip()


def parse_frontmatter(text: str) -> Frontmatter:
    lines = text.splitlines()
    if not lines or lines[0].strip() != "---":
        return Frontmatter(False, {}, {}, [], 0)

    closing = 0
    for index in range(1, len(lines)):
        if lines[index].strip() == "---":
            closing = index + 1
            break
    if not closing:
        return Frontmatter(True, {}, {}, [(1, "frontmatter is not closed")], 0)

    fields: Dict[str, str] = {}
    field_lines: Dict[str, int] = {}
    errors: List[Tuple[int, str]] = []
    active_key: Optional[str] = None
    for index, raw_line in enumerate(lines[1 : closing - 1], start=2):
        if not raw_line.strip() or raw_line.lstrip().startswith("#"):
            continue
        if raw_line[:1].isspace():
            if active_key is None:
                errors.append((index, "indented value has no parent key"))
            continue
        match = re.match(r"^([A-Za-z_][A-Za-z0-9_-]*):(?:\s*(.*))?$", raw_line)
        if not match:
            errors.append((index, "expected a top-level 'key: value' entry"))
            active_key = None
            continue
        key = match.group(1)
        if key in fields:
            errors.append((index, "duplicate frontmatter key: %s" % key))
            continue
        fields[key] = _strip_inline_comment(match.group(2) or "")
        field_lines[key] = index
        active_key = key
    return Frontmatter(True, fields, field_lines, errors, closing)


def _yaml_scalar_type(raw_value: str) -> str:
    value = raw_value.strip()
    if not value:
        return "mapping"
    if len(value) >= 2 and value[0] == value[-1] and value[0] in {'"', "'"}:
        return "string"
    if value in {"true", "false"}:
        return "bool"
    if re.fullmatch(r"-?(?:0|[1-9][0-9]*)", value):
        return "int"
    if value in {"null", "~"}:
        return "null"
    if value.startswith("[") and value.endswith("]"):
        return "sequence"
    if value.startswith("*"):
        return "alias"
    if value.startswith("&"):
        return "mapping"
    return "string"


def parse_yaml_like(text: str) -> YamlLike:
    """Parse context-manifest mappings while indexing every list item."""

    fields: Dict[str, str] = {}
    field_lines: Dict[str, int] = {}
    errors: List[Tuple[int, str]] = []
    list_items: Dict[str, List[str]] = {}
    scalar_types: Dict[str, str] = {}
    stack: List[Tuple[int, str]] = []
    list_counts: Dict[str, int] = {}
    key_pattern = re.compile(
        r"^((?:<<)|[A-Za-z_][A-Za-z0-9_-]*):(?:\s*(.*))?$"
    )

    for line_number, raw_line in enumerate(text.splitlines(), start=1):
        if not raw_line.strip() or raw_line.lstrip().startswith("#"):
            continue
        indentation_text = raw_line[: len(raw_line) - len(raw_line.lstrip())]
        if "\t" in indentation_text:
            errors.append((line_number, "tabs are not allowed for YAML indentation"))
            continue
        indentation = len(indentation_text)
        content = raw_line.strip()
        while stack and indentation <= stack[-1][0]:
            stack.pop()
        parent = stack[-1][1] if stack else ""

        if content.startswith("- "):
            if not parent:
                errors.append((line_number, "top-level list item has no mapping key"))
                continue
            item_content = content[2:].strip()
            item_index = list_counts.get(parent, 0)
            list_counts[parent] = item_index + 1
            item_path = "%s[%d]" % (parent, item_index)
            list_items.setdefault(parent, []).append(item_path)
            fields[item_path] = ""
            field_lines[item_path] = line_number
            scalar_types[item_path] = "mapping"
            item_match = key_pattern.match(item_content)
            if item_match:
                item_key = item_match.group(1)
                raw_value = item_match.group(2) or ""
                item_value = _strip_inline_comment(raw_value)
                item_field = item_path + "." + item_key
                fields[item_field] = item_value
                field_lines[item_field] = line_number
                scalar_types[item_field] = _yaml_scalar_type(raw_value)
                stack.append((indentation, item_path))
                if not item_value:
                    stack.append((indentation + 1, item_field))
            else:
                fields[item_path] = _strip_inline_comment(item_content)
                scalar_types[item_path] = _yaml_scalar_type(item_content)
            continue

        match = key_pattern.match(content)
        if not match:
            errors.append((line_number, "expected a YAML-like 'key: value' entry"))
            continue
        key = match.group(1)
        raw_value = match.group(2) or ""
        value = _strip_inline_comment(raw_value)
        path = parent + "." + key if parent else key
        if path in fields:
            errors.append((line_number, "duplicate YAML-like key path: %s" % path))
        fields[path] = value
        field_lines[path] = line_number
        scalar_types[path] = _yaml_scalar_type(raw_value)
        if not value or raw_value.strip().startswith("&"):
            stack.append((indentation, path))
    return YamlLike(fields, field_lines, errors, list_items, scalar_types)


def without_fenced_blocks(text: str) -> str:
    """Blank fenced Markdown blocks while retaining line numbers."""
    output: List[str] = []
    fence_char: Optional[str] = None
    fence_length = 0
    for line in text.splitlines(keepends=True):
        match = re.match(r"^\s*(`{3,}|~{3,})", line)
        if match:
            marker = match.group(1)
            if fence_char is None:
                fence_char = marker[0]
                fence_length = len(marker)
            elif marker[0] == fence_char and len(marker) >= fence_length:
                fence_char = None
                fence_length = 0
            output.append("\n" if line.endswith("\n") else "")
            continue
        output.append(("\n" if line.endswith("\n") else "") if fence_char else line)
    return "".join(output)


def markdown_inline_targets(text: str) -> Iterable[Tuple[str, int]]:
    """Yield inline Markdown link destinations with balanced parentheses."""
    for marker in re.finditer(r"\]\(", text):
        preceding = 0
        cursor = marker.start()
        while cursor > 0 and text[cursor - 1] == "\\":
            preceding += 1
            cursor -= 1
        if preceding % 2:
            continue
        start = marker.end()
        if start < len(text) and text[start] == "<":
            closing = text.find(">", start + 1)
            if closing != -1:
                yield text[start : closing + 1], start
            continue
        depth = 1
        cursor = start
        while cursor < len(text):
            character = text[cursor]
            if character == "\\":
                cursor += 2
                continue
            if character == "(":
                depth += 1
            elif character == ")":
                depth -= 1
                if depth == 0:
                    yield text[start:cursor], start
                    break
            cursor += 1


def github_heading_slug(heading: str) -> str:
    """Approximate GitHub's Unicode-friendly Markdown heading normalization."""

    value = re.sub(r"<[^>]+>", "", heading)
    value = re.sub(r"!\[([^\]]*)\]\([^)]*\)", r"\1", value)
    value = re.sub(r"\[([^\]]+)\]\([^)]*\)", r"\1", value)
    value = re.sub(r"[`*_~]", "", value).strip().lower()
    kept = "".join(
        character
        for character in value
        if character.isalnum() or character.isspace() or character in {"-", "_"}
    )
    return re.sub(r"\s", "-", kept).strip("-")


def markdown_heading_anchors(text: str) -> Set[str]:
    """Collect GitHub-style anchors, including duplicate-heading suffixes."""

    visible = without_fenced_blocks(text)
    headings: List[str] = []
    lines = visible.splitlines()
    for index, line in enumerate(lines):
        atx = re.match(r"^\s{0,3}#{1,6}\s+(.+?)\s*#*\s*$", line)
        if atx:
            headings.append(atx.group(1))
            continue
        if index + 1 < len(lines) and re.match(
            r"^\s{0,3}(?:=+|-+)\s*$", lines[index + 1]
        ):
            if line.strip():
                headings.append(line.strip())
    counts: Dict[str, int] = {}
    anchors: Set[str] = set()
    for heading in headings:
        base = github_heading_slug(heading)
        if not base:
            continue
        count = counts.get(base, 0)
        anchors.add(base if count == 0 else "%s-%d" % (base, count))
        counts[base] = count + 1
    return anchors


def _json_type_matches(value: Any, expected: str) -> bool:
    """Return JSON type compatibility without treating bool as a number."""

    if expected == "object":
        return isinstance(value, dict)
    if expected == "array":
        return isinstance(value, list)
    if expected == "string":
        return isinstance(value, str)
    if expected == "integer":
        return isinstance(value, int) and not isinstance(value, bool)
    if expected == "number":
        return isinstance(value, (int, float)) and not isinstance(value, bool)
    if expected == "boolean":
        return isinstance(value, bool)
    if expected == "null":
        return value is None
    return False


def _resolve_local_json_ref(root_schema: Dict[str, Any], reference: str) -> Any:
    if not reference.startswith("#/"):
        raise ValueError("only local JSON Schema references are supported")
    current: Any = root_schema
    for raw_token in reference[2:].split("/"):
        token = raw_token.replace("~1", "/").replace("~0", "~")
        if not isinstance(current, dict) or token not in current:
            raise ValueError("unresolved local JSON Schema reference")
        current = current[token]
    return current


def validate_json_instance(
    instance: Any,
    schema: Dict[str, Any],
    *,
    root_schema: Optional[Dict[str, Any]] = None,
    path: str = "$",
    ref_stack: Tuple[str, ...] = (),
) -> List[str]:
    """Validate the Draft 2020-12 subset used by bundled contracts.

    This is not a general JSON Schema implementation. It keeps bundled examples
    and sentinels executable with stdlib only; a controller must use a maintained
    full validator when schema validation is part of policy enforcement.
    """

    root = root_schema or schema
    errors: List[str] = []
    reference = schema.get("$ref")
    if reference is not None:
        if not isinstance(reference, str):
            return ["%s: $ref must be a string" % path]
        if reference in ref_stack:
            return ["%s: recursive $ref is unsupported" % path]
        try:
            target = _resolve_local_json_ref(root, reference)
        except ValueError as exc:
            return ["%s: %s" % (path, exc)]
        if not isinstance(target, dict):
            return ["%s: $ref target must be an object" % path]
        errors.extend(
            validate_json_instance(
                instance,
                target,
                root_schema=root,
                path=path,
                ref_stack=ref_stack + (reference,),
            )
        )

    expected_types = schema.get("type")
    if expected_types is not None:
        choices = expected_types if isinstance(expected_types, list) else [expected_types]
        if not choices or not all(isinstance(item, str) for item in choices):
            errors.append("%s: schema type must be a string or non-empty string list" % path)
            return errors
        if not any(_json_type_matches(instance, item) for item in choices):
            errors.append("%s: expected type %s" % (path, "|".join(choices)))
            return errors

    if "const" in schema and instance != schema["const"]:
        errors.append("%s: value does not match const" % path)
    if "enum" in schema and instance not in schema["enum"]:
        errors.append("%s: value is outside enum" % path)

    for subschema in schema.get("allOf", []):
        errors.extend(validate_json_instance(instance, subschema, root_schema=root, path=path))
    if "anyOf" in schema:
        branches = [
            validate_json_instance(instance, item, root_schema=root, path=path)
            for item in schema["anyOf"]
        ]
        if not any(not result for result in branches):
            errors.append("%s: no anyOf branch matched" % path)
    if "oneOf" in schema:
        matches = sum(
            not validate_json_instance(instance, item, root_schema=root, path=path)
            for item in schema["oneOf"]
        )
        if matches != 1:
            errors.append("%s: expected exactly one oneOf branch" % path)
    if "not" in schema and not validate_json_instance(
        instance, schema["not"], root_schema=root, path=path
    ):
        errors.append("%s: value matches forbidden schema" % path)
    if "if" in schema:
        condition_matches = not validate_json_instance(
            instance, schema["if"], root_schema=root, path=path
        )
        branch = schema.get("then" if condition_matches else "else")
        if branch is not None:
            errors.extend(validate_json_instance(instance, branch, root_schema=root, path=path))

    if isinstance(instance, dict):
        for name in schema.get("required", []):
            if name not in instance:
                errors.append("%s: missing required property %s" % (path, name))
        properties = schema.get("properties", {})
        for name, value in instance.items():
            child_path = "%s.%s" % (path, name)
            if name in properties:
                errors.extend(
                    validate_json_instance(
                        value, properties[name], root_schema=root, path=child_path
                    )
                )
            elif schema.get("additionalProperties") is False:
                errors.append("%s: additional property is forbidden" % child_path)
            elif isinstance(schema.get("additionalProperties"), dict):
                errors.extend(
                    validate_json_instance(
                        value,
                        schema["additionalProperties"],
                        root_schema=root,
                        path=child_path,
                    )
                )

    if isinstance(instance, list):
        if "minItems" in schema and len(instance) < schema["minItems"]:
            errors.append("%s: fewer than minItems" % path)
        if "maxItems" in schema and len(instance) > schema["maxItems"]:
            errors.append("%s: more than maxItems" % path)
        if schema.get("uniqueItems"):
            serialized = [json.dumps(item, sort_keys=True, ensure_ascii=False) for item in instance]
            if len(serialized) != len(set(serialized)):
                errors.append("%s: array items are not unique" % path)
        item_schema = schema.get("items")
        if isinstance(item_schema, dict):
            for index, value in enumerate(instance):
                errors.extend(
                    validate_json_instance(
                        value,
                        item_schema,
                        root_schema=root,
                        path="%s[%d]" % (path, index),
                    )
                )

    if isinstance(instance, str):
        if "minLength" in schema and len(instance) < schema["minLength"]:
            errors.append("%s: shorter than minLength" % path)
        if "maxLength" in schema and len(instance) > schema["maxLength"]:
            errors.append("%s: longer than maxLength" % path)
        if "pattern" in schema and re.search(schema["pattern"], instance) is None:
            errors.append("%s: string does not match pattern" % path)

    if isinstance(instance, (int, float)) and not isinstance(instance, bool):
        if "minimum" in schema and instance < schema["minimum"]:
            errors.append("%s: below minimum" % path)
        if "maximum" in schema and instance > schema["maximum"]:
            errors.append("%s: above maximum" % path)
    return errors


class ToolkitValidator:
    def __init__(self, root: Path, config: Dict[str, Any]) -> None:
        self.config = config
        self.issues: List[Issue] = []
        self._tree_error: Optional[FileSafetyError] = None
        try:
            self.root = resolve_directory_root(root, "validation root")
            self.text_policy = parse_text_policy(
                config.get("scan", {}), field="file_policy"
            )
        except FileSafetyError as exc:
            raise ValueError(str(exc)) from exc
        global_excludes = self.config.get("scan", {}).get(
            "global_exclude_globs", []
        )
        try:
            self.tree: Optional[SafeTree] = SafeTree(
                self.root,
                self.text_policy,
                exclude=lambda relative, _is_directory: _matches(
                    relative, global_excludes
                ),
            )
        except FileSafetyError as exc:
            self.tree = None
            self._tree_error = exc
        self._text_cache: Dict[Path, str] = {}
        self._inspection_cache: Dict[Path, InspectedText] = {}
        self._read_failures: Set[Path] = set()
        self._paths_cache: Optional[Tuple[Path, ...]] = None

    def add(
        self,
        check: str,
        code: str,
        path: Path,
        line: int,
        message: str,
        severity: str = "ERROR",
    ) -> None:
        try:
            relative = path.relative_to(self.root).as_posix()
        except ValueError:
            relative = path.as_posix()
        self.issues.append(
            Issue(check, code, severity, relative, max(1, line), message)
        )

    def read_text(self, path: Path, check: str) -> Optional[str]:
        if path in self._text_cache:
            return self._text_cache[path]
        if path in self._read_failures:
            return None
        if self.tree is None:
            return None
        try:
            inspected = self.tree.inspect_text(path)
        except FileSafetyError as exc:
            message = str(exc)
            if "unapproved text file type" in message:
                code = "FILE001"
            elif "byte limit" in message:
                code = "FILE003"
            elif "binary/" in message or "control content" in message or "UTF-8" in message:
                code = "FILE002"
            else:
                code = "SAFEIO001"
            self.add(
                "file_policy",
                code,
                path,
                1,
                message,
            )
            self._read_failures.add(path)
            return None
        self._inspection_cache[path] = inspected
        self._text_cache[path] = inspected.text
        return inspected.text

    def _all_paths(self) -> Tuple[Path, ...]:
        if self._paths_cache is not None:
            return self._paths_cache
        self._paths_cache = self.tree.files() if self.tree is not None else ()
        return self._paths_cache

    def files(
        self, includes: Sequence[str], excludes: Sequence[str] = ()
    ) -> Iterable[Path]:
        global_excludes = self.config.get("scan", {}).get("global_exclude_globs", [])
        for path in self._all_paths():
            relative = path.relative_to(self.root).as_posix()
            if _matches(relative, global_excludes) or _matches(relative, excludes):
                continue
            if _matches(relative, includes):
                yield path

    def _snapshot_exists(self, path: Path) -> bool:
        if self.tree is None:
            return False
        try:
            return self.tree.exists(path)
        except FileSafetyError:
            return False

    def _snapshot_is_file(self, path: Path) -> bool:
        if self.tree is None:
            return False
        return self.tree.is_file(path)

    def _snapshot_is_dir(self, path: Path) -> bool:
        if self.tree is None:
            return False
        return self.tree.is_dir(path)

    def _snapshot_files_in(
        self, directory: Path, *, suffix: Optional[str] = None
    ) -> Tuple[Path, ...]:
        if self.tree is None:
            return ()
        return self.tree.files_in(directory, suffix=suffix)

    def _snapshot_directories_in(self, directory: Path) -> Tuple[Path, ...]:
        if self.tree is None:
            return ()
        return self.tree.directories_in(directory)

    def check_file_policy(self) -> None:
        paths = list(self._all_paths())
        if len(paths) > self.text_policy.max_file_count:
            self.add(
                "file_policy",
                "FILE005",
                self.root,
                1,
                "validation tree exceeds max file count",
            )
        total = 0
        total_reported = False
        for path in paths:
            if self.read_text(path, "file_policy") is None:
                continue
            total += self._inspection_cache[path].size
            if total > self.text_policy.max_total_bytes and not total_reported:
                self.add(
                    "file_policy",
                    "FILE004",
                    path,
                    1,
                    "validation tree exceeds max total bytes",
                )
                total_reported = True


    def check_inline_paths(self) -> None:
        """Пути, записанные в бэктиках, а не markdown-ссылкой.

        Именно здесь копятся следы переездов: скрипты пересчёта правят
        ссылки и не трогают текст, а check_links смотрит только ссылки.
        """
        settings = self.config.get("inline_paths")
        if not settings:
            return
        roots = tuple(settings.get("package_roots", []))
        pattern = re.compile(r"`([^`\n]+?\.(?:md|py|json|ya?ml))`")
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            text = self.read_text(path, "inline_paths")
            if text is None:
                continue
            scan_text = without_fenced_blocks(text)
            for match in pattern.finditer(scan_text):
                raw = match.group(1).strip()
                if raw.startswith(("http", "mailto", "<")) or "{{" in raw or "*" in raw:
                    continue
                if "<" in raw or ">" in raw or " " in raw:
                    continue
                if "/" not in raw:
                    continue
                candidates = [path.parent / raw]
                if roots and raw.startswith(roots):
                    candidates.append(self.root.parent / raw)
                else:
                    candidates.append(self.root / raw)
                if any(self._snapshot_exists(candidate) for candidate in candidates):
                    continue
                if True:
                    self.add(
                        "inline_paths",
                        "PATH001",
                        path,
                        _line_number(scan_text, match.start(1)),
                        f"inline path does not resolve: {raw}",
                    )

    def check_links(self) -> None:
        settings = self.config["links"]
        reference = re.compile(
            r"(?m)^\s*\[[^\]\n]+\]:\s*(<[^>\n]+>|\S+)"
        )
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            text = self.read_text(path, "links")
            if text is None:
                continue
            scan_text = without_fenced_blocks(text)
            targets = list(markdown_inline_targets(scan_text)) + [
                (match.group(1), match.start(1))
                for match in reference.finditer(scan_text)
            ]
            for raw_target, offset in targets:
                raw_target = raw_target.strip()
                target = self._link_target(raw_target)
                if target is None:
                    continue
                line = _line_number(scan_text, offset)
                if "{{" in target or "}}" in target:
                    self.add(
                        "links",
                        "LINK002",
                        path,
                        line,
                        "relative link target contains an unresolved placeholder",
                    )
                    continue
                line = _line_number(scan_text, offset)
                markdown_unescaped = re.sub(r"\\([() ])", r"\1", target)
                drive, _ = ntpath.splitdrive(markdown_unescaped)
                if (
                    drive
                    or markdown_unescaped.startswith(("/", "\\", "//", "\\\\"))
                ):
                    self.add(
                        "links",
                        "LINK003",
                        path,
                        line,
                        "absolute, drive-qualified, and UNC link targets are forbidden",
                    )
                    continue
                parsed = urlsplit(markdown_unescaped)
                if parsed.scheme or parsed.netloc:
                    continue
                decoded_path = unquote(parsed.path).replace("\\", "/")
                fragment = unquote(parsed.fragment)
                if "\x00" in decoded_path or "\x00" in fragment:
                    self.add(
                        "links",
                        "LINK003",
                        path,
                        line,
                        "link target contains an invalid character",
                    )
                    continue
                destination_candidate = lexical_absolute(
                    path if not decoded_path else path.parent / decoded_path
                )
                try:
                    common = os.path.commonpath(
                        [
                            os.path.normcase(os.fspath(destination_candidate)),
                            os.path.normcase(os.fspath(self.root)),
                        ]
                    )
                except ValueError:
                    common = ""
                if common != os.path.normcase(os.fspath(self.root)):
                    self.add(
                        "links",
                        "LINK003",
                        path,
                        line,
                        "relative link target escapes the validation root",
                    )
                    continue
                if not self._snapshot_exists(destination_candidate):
                    digest = hashlib.sha256(
                        decoded_path.encode("utf-8")
                    ).hexdigest()[:10]
                    self.add(
                        "links",
                        "LINK001",
                        path,
                        line,
                        "relative link target does not exist (target fingerprint %s)"
                        % digest,
                    )
                    continue
                destination = destination_candidate
                if fragment and destination.suffix.lower() in {".md", ".markdown"}:
                    target_text = self.read_text(destination, "links")
                    if target_text is None:
                        continue
                    if fragment.lower() not in markdown_heading_anchors(target_text):
                        self.add(
                            "links",
                            "LINK004",
                            path,
                            line,
                            "Markdown link fragment does not match a target heading",
                        )

    @staticmethod
    def _link_target(raw_target: str) -> Optional[str]:
        if raw_target.startswith("<") and ">" in raw_target:
            target = raw_target[1 : raw_target.index(">")]
        else:
            target = raw_target.split(maxsplit=1)[0]
        if not target:
            return None
        return target

    def check_frontmatter(self) -> None:
        settings = self.config["frontmatter"]
        required = settings.get("required_globs", [])
        markdown = list(
            self.files(settings["include_globs"], settings.get("exclude_globs", []))
        )
        for path in markdown:
            text = self.read_text(path, "frontmatter")
            if text is None:
                continue
            relative = path.relative_to(self.root).as_posix()
            frontmatter = parse_frontmatter(text)
            if _matches(relative, required) and not frontmatter.present:
                self.add(
                    "frontmatter",
                    "FM001",
                    path,
                    1,
                    "required YAML-like frontmatter is missing",
                )
                continue
            for line, error in frontmatter.errors:
                self.add("frontmatter", "FM002", path, line, error)

    def check_skills(self) -> None:
        settings = self.config["skills"]
        name_pattern = re.compile(settings["name_pattern"])
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            text = self.read_text(path, "skills")
            if text is None:
                continue
            frontmatter = parse_frontmatter(text)
            if not frontmatter.present:
                self.add("skills", "SKILL001", path, 1, "skill frontmatter is missing")
                continue
            for field in settings["required_fields"]:
                if not frontmatter.fields.get(field):
                    self.add(
                        "skills",
                        "SKILL002",
                        path,
                        1,
                        "required skill field is missing or empty: %s" % field,
                    )
            name = frontmatter.fields.get("name", "")
            if name and name != path.stem:
                self.add(
                    "skills",
                    "SKILL003",
                    path,
                    frontmatter.lines.get("name", 1),
                    "skill name must equal its file name",
                )
            if name and not name_pattern.fullmatch(name):
                self.add(
                    "skills",
                    "SKILL004",
                    path,
                    frontmatter.lines.get("name", 1),
                    "skill name violates the configured Agent Skills pattern",
                )
            name_maximum = int(settings.get("name_max_length", 64))
            if len(name) > name_maximum:
                self.add(
                    "skills",
                    "SKILL006",
                    path,
                    frontmatter.lines.get("name", 1),
                    "skill name exceeds %d characters" % name_maximum,
                )
            description = frontmatter.fields.get("description", "")
            maximum = int(settings.get("description_max_length", 1024))
            if len(description) > maximum:
                self.add(
                    "skills",
                    "SKILL005",
                    path,
                    frontmatter.lines.get("description", 1),
                    "skill description exceeds %d characters" % maximum,
                )

    def check_prompt_ids(self) -> None:
        settings = self.config["prompt_ids"]
        pattern = re.compile(settings["line_pattern"], re.MULTILINE)
        expected_start = int(settings.get("expected_start", 1))
        global_ids: Dict[str, Tuple[Path, int]] = {}
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            text = self.read_text(path, "prompt_ids")
            if text is None:
                continue
            scan_text = without_fenced_blocks(text)
            found: Dict[int, List[int]] = {}
            category = path.stem.split("-", 1)[0]
            for match in pattern.finditer(scan_text):
                number = int(match.group("id"))
                found.setdefault(number, []).append(_line_number(scan_text, match.start()))
                canonical = "%s-%03d" % (category, number)
                if canonical in global_ids:
                    previous_path, previous_line = global_ids[canonical]
                    self.add(
                        "prompt_ids",
                        "PROMPT001",
                        path,
                        found[number][-1],
                        "duplicate prompt ID %s; first seen at %s:%d"
                        % (
                            canonical,
                            previous_path.relative_to(self.root).as_posix(),
                            previous_line,
                        ),
                    )
                else:
                    global_ids[canonical] = (path, found[number][-1])
            if not found:
                self.add(
                    "prompt_ids",
                    "PROMPT002",
                    path,
                    1,
                    "no numbered prompt IDs were detected",
                )
                continue
            for number, lines in found.items():
                if len(lines) > 1:
                    self.add(
                        "prompt_ids",
                        "PROMPT001",
                        path,
                        lines[1],
                        "prompt ID %s-%03d appears %d times"
                        % (category, number, len(lines)),
                    )
            maximum = max(found)
            missing = [number for number in range(expected_start, maximum + 1) if number not in found]
            if missing:
                preview = ", ".join(
                    "%s-%03d" % (category, number) for number in missing[:12]
                )
                suffix = " ..." if len(missing) > 12 else ""
                self.add(
                    "prompt_ids",
                    "PROMPT003",
                    path,
                    1,
                    "missing sequential prompt IDs: %s%s" % (preview, suffix),
                )

    def check_placeholders(self) -> None:
        settings = self.config["placeholders"]
        placeholder = re.compile(settings["placeholder_pattern"])
        generated = set(settings["generated_statuses"])
        templates = set(settings["template_statuses"])
        generated_profiles: List[Path] = []
        for path in self.files(settings["profile_globs"], settings.get("exclude_globs", [])):
            text = self.read_text(path, "placeholders")
            if text is None:
                continue
            frontmatter = parse_frontmatter(text)
            status = frontmatter.fields.get(settings["status_field"], "")
            matches = list(placeholder.finditer(text))
            if status in generated:
                generated_profiles.append(path)
                if matches:
                    self.add(
                        "placeholders",
                        "PLACE001",
                        path,
                        _line_number(text, matches[0].start()),
                        "generated profile contains %d unresolved placeholder(s)"
                        % len(matches),
                    )
                for field in settings["generated_required_fields"]:
                    if not frontmatter.fields.get(field):
                        self.add(
                            "placeholders",
                            "PLACE002",
                            path,
                            1,
                            "generated profile lacks provenance field: %s" % field,
                        )
            elif status in templates:
                if settings.get("template_requires_placeholder", False) and not matches:
                    self.add(
                        "placeholders",
                        "PLACE003",
                        path,
                        1,
                        "template profile contains no placeholder slots",
                        "WARNING",
                    )
            else:
                self.add(
                    "placeholders",
                    "PLACE004",
                    path,
                    frontmatter.lines.get(settings["status_field"], 1),
                    "profile status is missing or not configured",
                )
        self._check_context_manifest(generated_profiles, settings["context_manifest"])
        if settings.get("structured_profiles"):
            self._check_structured_context_profiles(
                settings["structured_profiles"], placeholder
            )

    def _check_context_manifest(
        self, generated_profiles: Sequence[Path], settings: Dict[str, Any]
    ) -> None:
        manifest_path = self.root / settings["path"]
        must_exist = bool(generated_profiles) and settings.get(
            "required_when_profiles_generated", True
        )
        if not self._snapshot_is_file(manifest_path):
            if must_exist:
                self.add(
                    "placeholders",
                    "PLACE005",
                    manifest_path,
                    1,
                    "generated context requires a context manifest",
                )
            return
        text = self.read_text(manifest_path, "placeholders")
        if text is None:
            return
        frontmatter = parse_frontmatter(text)
        if not frontmatter.present:
            self.add(
                "placeholders",
                "PLACE006",
                manifest_path,
                1,
                "context manifest frontmatter is missing",
            )
            return
        for field in settings["required_fields"]:
            if not frontmatter.fields.get(field):
                self.add(
                    "placeholders",
                    "PLACE007",
                    manifest_path,
                    1,
                    "context manifest lacks required field: %s" % field,
                )

    def _check_structured_context_profiles(
        self, settings: Dict[str, Any], placeholder: re.Pattern[str]
    ) -> None:
        for relative in settings["template_directories"]:
            directory = self.root / relative
            if not self._snapshot_is_dir(directory):
                self.add(
                    "placeholders",
                    "PLACE010",
                    directory,
                    1,
                    "required context template directory is missing",
                )
                continue
            self._check_profile_directory(
                directory, settings, placeholder, is_template=True
            )

        projects_root = self.root / settings["projects_root"]
        if not self._snapshot_is_dir(projects_root):
            self.add(
                "placeholders",
                "PLACE011",
                projects_root,
                1,
                "context projects root is missing",
            )
            return
        excluded = set(settings.get("excluded_project_directories", []))
        for directory in self._snapshot_directories_in(projects_root):
            if directory.name in excluded:
                continue
            self._check_profile_directory(
                directory, settings, placeholder, is_template=False
            )

    def _check_profile_directory(
        self,
        directory: Path,
        settings: Dict[str, Any],
        placeholder: re.Pattern[str],
        is_template: bool,
    ) -> None:
        manifest_path = directory / settings["manifest_name"]
        if not self._snapshot_is_file(manifest_path):
            self.add(
                "placeholders",
                "PLACE012",
                manifest_path,
                1,
                "context profile manifest is missing",
            )
        else:
            text = self.read_text(manifest_path, "placeholders")
            if text is not None:
                manifest = parse_yaml_like(text)
                for line, error in manifest.errors:
                    self.add(
                        "placeholders", "PLACE013", manifest_path, line, error
                    )
                for field in settings["manifest_required_paths"]:
                    if "[]" not in field:
                        missing = [field] if field not in manifest.fields else []
                    else:
                        list_path, suffix = field.split("[]", 1)
                        suffix = suffix.lstrip(".")
                        items = manifest.list_items.get(list_path, [])
                        missing = (
                            [field]
                            if not items
                            else [
                                item + ("." + suffix if suffix else "")
                                for item in items
                                if (
                                    item + ("." + suffix if suffix else "")
                                    not in manifest.fields
                                    or not manifest.fields.get(
                                        item + ("." + suffix if suffix else ""),
                                        "",
                                    )
                                )
                            ]
                        )
                    for missing_path in missing:
                        self.add(
                            "placeholders",
                            "PLACE014",
                            manifest_path,
                            manifest.lines.get(missing_path.rsplit(".", 1)[0], 1),
                            "context manifest lacks required path: %s"
                            % missing_path,
                        )
                status_path = settings["manifest_status_path"]
                status = manifest.fields.get(status_path, "")
                if is_template:
                    if status != settings["template_status"]:
                        self.add(
                            "placeholders",
                            "PLACE015",
                            manifest_path,
                            manifest.lines.get(status_path, 1),
                            "template context manifest has a non-template status",
                        )
                else:
                    if status not in settings["generated_statuses"]:
                        self.add(
                            "placeholders",
                            "PLACE016",
                            manifest_path,
                            manifest.lines.get(status_path, 1),
                            "generated context manifest status is not configured",
                        )
                    matches = list(placeholder.finditer(text))
                    if matches:
                        self.add(
                            "placeholders",
                            "PLACE017",
                            manifest_path,
                            _line_number(text, matches[0].start()),
                            "generated context manifest contains unresolved placeholders",
                        )
                    slug_path = settings["manifest_slug_path"]
                    if manifest.fields.get(slug_path) != directory.name:
                        self.add(
                            "placeholders",
                            "PLACE018",
                            manifest_path,
                            manifest.lines.get(slug_path, 1),
                            "context manifest slug must equal its profile directory",
                        )
                    self._validate_generated_manifest(
                        manifest_path, manifest, settings
                    )

        expected = set(settings["expected_markdown_files"])
        ignored = set(settings.get("ignored_markdown_files", []))
        actual = {
            path.name
            for path in self._snapshot_files_in(directory, suffix=".md")
            if path.name not in ignored
        }
        for filename in sorted(expected - actual):
            self.add(
                "placeholders",
                "PLACE019",
                directory / filename,
                1,
                "required context profile artifact is missing",
            )
        for filename in sorted(expected | actual):
            path = directory / filename
            if not self._snapshot_is_file(path):
                continue
            text = self.read_text(path, "placeholders")
            if text is None:
                continue
            frontmatter = parse_frontmatter(text)
            if not frontmatter.present:
                self.add(
                    "placeholders",
                    "PLACE020",
                    path,
                    1,
                    "context profile artifact frontmatter is missing",
                )
                continue
            for field in settings["markdown_required_fields"]:
                if not frontmatter.fields.get(field):
                    self.add(
                        "placeholders",
                        "PLACE021",
                        path,
                        1,
                        "context profile artifact lacks field: %s" % field,
                    )
            status = frontmatter.fields.get(settings["markdown_status_field"], "")
            if is_template:
                if status != settings["template_status"]:
                    self.add(
                        "placeholders",
                        "PLACE022",
                        path,
                        frontmatter.lines.get(settings["markdown_status_field"], 1),
                        "template context artifact has a non-template status",
                    )
            else:
                if status not in settings["generated_statuses"]:
                    self.add(
                        "placeholders",
                        "PLACE023",
                        path,
                        frontmatter.lines.get(settings["markdown_status_field"], 1),
                        "generated context artifact status is not configured",
                    )
                matches = list(placeholder.finditer(text))
                if matches:
                    self.add(
                        "placeholders",
                        "PLACE024",
                        path,
                        _line_number(text, matches[0].start()),
                        "generated context artifact contains unresolved placeholders",
                    )
                slug = frontmatter.fields.get(settings["markdown_slug_field"], "")
                if slug != directory.name:
                    self.add(
                        "placeholders",
                        "PLACE025",
                        path,
                        frontmatter.lines.get(settings["markdown_slug_field"], 1),
                        "context artifact slug must equal its profile directory",
                    )

    def _validate_generated_manifest(
        self,
        path: Path,
        manifest: YamlLike,
        settings: Dict[str, Any],
    ) -> None:
        legacy_path = "scan.commands_executed"
        if legacy_path in manifest.fields:
            self.add(
                "placeholders",
                "PLACE026",
                path,
                manifest.lines.get(legacy_path, 1),
                "legacy aggregate commands_executed is forbidden",
            )
        for command_path in (
            "scan.metadata_commands_executed",
            "scan.project_commands_executed",
        ):
            if (
                manifest.fields.get(command_path) not in {"true", "false"}
                or manifest.scalar_types.get(command_path) != "bool"
            ):
                self.add(
                    "placeholders",
                    "PLACE027",
                    path,
                    manifest.lines.get(command_path, 1),
                    "%s must be an explicit YAML boolean" % command_path,
                )

        if manifest.fields.get("scan.status") != "complete":
            return

        hash_pattern = re.compile(r"^[0-9a-f]{64}$")
        timestamp_pattern = re.compile(
            r"^[0-9]{4}-[0-9]{2}-[0-9]{2}T"
            r"[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]+)?"
            r"(?:Z|[+-][0-9]{2}:[0-9]{2})$"
        )
        na_pattern = re.compile(r"^N/A:\s+\S")

        def valid_timestamp(value: str) -> bool:
            if not timestamp_pattern.fullmatch(value):
                return False
            try:
                parsed = datetime.fromisoformat(value.replace("Z", "+00:00"))
            except ValueError:
                return False
            return parsed.tzinfo is not None

        source_items = manifest.list_items.get("sources", [])
        seen_ids: Set[str] = set()
        for item in source_items:
            source_id = manifest.fields.get(item + ".id", "")
            if not source_id or source_id in seen_ids:
                self.add(
                    "placeholders",
                    "PLACE028",
                    path,
                    manifest.lines.get(item + ".id", 1),
                    "each source item requires a unique nonempty id",
                )
            seen_ids.add(source_id)
            kind = manifest.fields.get(item + ".kind", "")
            root = manifest.fields.get(item + ".root", "")
            source_path = manifest.fields.get(item + ".path", "")
            if not root or not source_path or not kind:
                self.add(
                    "placeholders",
                    "PLACE032",
                    path,
                    manifest.lines.get(item, 1),
                    "each source item requires nonempty root, path, and kind",
                )
            documented_non_file = (
                kind == "invocation_metadata" and root == "invocation"
            ) or kind == "filesystem_inventory"
            if kind == "filesystem_inventory" and not any(
                key.startswith(item + ".inventory.") for key in manifest.fields
            ):
                documented_non_file = False
            for field_name, validator in (
                ("sha256", lambda value: bool(hash_pattern.fullmatch(value))),
                (
                    "size_bytes",
                    lambda value: bool(re.fullmatch(r"(?:0|[1-9][0-9]*)", value)),
                ),
                (
                    "modified_at",
                    valid_timestamp,
                ),
                (
                    "observed_at",
                    valid_timestamp,
                ),
            ):
                field_path = item + "." + field_name
                value = manifest.fields.get(field_path, "")
                is_na = bool(na_pattern.match(value))
                valid = validator(value)
                if field_name == "size_bytes" and valid:
                    valid = manifest.scalar_types.get(field_path) == "int"
                if not valid and not (is_na and documented_non_file):
                    self.add(
                        "placeholders",
                        "PLACE029",
                        path,
                        manifest.lines.get(field_path, manifest.lines.get(item, 1)),
                        "%s requires typed file metadata or documented non-file N/A"
                        % field_path,
                    )

        output_items = manifest.list_items.get("outputs", [])
        seen_outputs: Set[str] = set()
        for item in output_items:
            output_path = manifest.fields.get(item + ".path", "")
            if not output_path or output_path in seen_outputs:
                self.add(
                    "placeholders",
                    "PLACE030",
                    path,
                    manifest.lines.get(item + ".path", 1),
                    "each output item requires a unique nonempty path",
                )
            seen_outputs.add(output_path)
            if not manifest.fields.get(item + ".status", ""):
                self.add(
                    "placeholders",
                    "PLACE033",
                    path,
                    manifest.lines.get(item + ".status", manifest.lines.get(item, 1)),
                    "each output item requires a nonempty status",
                )
            checks = (
                (
                    "sha256",
                    hash_pattern.fullmatch(
                        manifest.fields.get(item + ".sha256", "")
                    )
                    is not None,
                ),
                (
                    "size_bytes",
                    re.fullmatch(
                        r"(?:0|[1-9][0-9]*)",
                        manifest.fields.get(item + ".size_bytes", ""),
                    )
                    is not None
                    and manifest.scalar_types.get(item + ".size_bytes") == "int",
                ),
                (
                    "modified_at",
                    valid_timestamp(
                        manifest.fields.get(item + ".modified_at", "")
                    ),
                ),
            )
            for field_name, valid in checks:
                if not valid:
                    field_path = item + "." + field_name
                    self.add(
                        "placeholders",
                        "PLACE031",
                        path,
                        manifest.lines.get(field_path, manifest.lines.get(item, 1)),
                        "%s requires nonempty typed output metadata" % field_path,
                    )

    def _pattern_scan(
        self,
        check: str,
        settings: Dict[str, Any],
        code: str,
        redacted_message: str,
    ) -> None:
        compiled = [
            (entry["id"], re.compile(entry["regex"]))
            for entry in settings["patterns"]
        ]
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            text = self.read_text(path, check)
            if text is None:
                continue
            seen: Set[Tuple[str, int]] = set()
            for pattern_id, pattern in compiled:
                for match in pattern.finditer(text):
                    line = _line_number(text, match.start())
                    marker = (pattern_id, line)
                    if marker in seen:
                        continue
                    seen.add(marker)
                    self.add(
                        check,
                        code,
                        path,
                        line,
                        "%s '%s'" % (redacted_message, pattern_id),
                    )

    def check_cot_disclosure(self) -> None:
        settings = self.config["cot_disclosure"]
        compiled = [
            (entry["id"], re.compile(entry["regex"]))
            for entry in settings["patterns"]
        ]
        negations = [
            re.compile(pattern)
            for pattern in settings.get("negation_prefix_patterns", [])
        ]
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            text = self.read_text(path, "cot_disclosure")
            if text is None:
                continue
            seen: Set[Tuple[str, int]] = set()
            for pattern_id, pattern in compiled:
                for match in pattern.finditer(text):
                    line_start = text.rfind("\n", 0, match.start()) + 1
                    prefix = text[line_start : match.start()]
                    if any(negation.search(prefix) for negation in negations):
                        continue
                    line = _line_number(text, match.start())
                    marker = (pattern_id, line)
                    if marker in seen:
                        continue
                    seen.add(marker)
                    self.add(
                        "cot_disclosure",
                        "COT001",
                        path,
                        line,
                        "forbidden explicit reasoning-disclosure instruction matched rule '%s'"
                        % pattern_id,
                    )

    def check_gsd_routes(self) -> None:
        settings = self.config["gsd_routes"]
        known = set(settings["known_routes"])
        patterns = [re.compile(item) for item in settings["route_patterns"]]
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            text = self.read_text(path, "gsd_routes")
            if text is None:
                continue
            seen: Set[Tuple[str, int]] = set()
            for pattern in patterns:
                for match in pattern.finditer(text):
                    route = match.groupdict().get("route") or match.group(1)
                    route = route.lstrip("/")
                    line = _line_number(text, match.start())
                    if (route, line) in seen:
                        continue
                    seen.add((route, line))
                    if route not in known:
                        suggestion = difflib.get_close_matches(route, sorted(known), n=1)
                        suffix = (
                            "; nearest configured route: %s" % suggestion[0]
                            if suggestion
                            else ""
                        )
                        self.add(
                            "gsd_routes",
                            "ROUTE001",
                            path,
                            line,
                            "unknown GSD route: %s%s" % (route, suffix),
                        )

    def check_secrets(self) -> None:
        settings = self.config["secrets"]
        for path in self.files(
            settings["include_globs"], settings.get("exclude_globs", [])
        ):
            text = self.read_text(path, "secrets")
            if text is None:
                continue
            seen: Set[Tuple[str, int]] = set()
            for pattern_id, offset in secret_rule_matches(text):
                line = _line_number(text, offset)
                marker = (pattern_id, line)
                if marker in seen:
                    continue
                seen.add(marker)
                self.add(
                    "secrets",
                    "SECRET001",
                    path,
                    line,
                    "secret-like material matched redacted rule '%s'"
                    % pattern_id,
                )

    def check_portable_leakage(self) -> None:
        settings = self.config["portable_leakage"]
        patterns = [
            (entry["id"], re.compile(entry["regex"]))
            for entry in settings["deny_patterns"]
        ]
        allows = settings.get("allow_rules", [])
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            text = self.read_text(path, "portable_leakage")
            if text is None:
                continue
            relative = path.relative_to(self.root).as_posix()
            seen: Set[Tuple[str, int]] = set()
            for pattern_id, pattern in patterns:
                if any(
                    pattern_id in allow.get("pattern_ids", [])
                    and _matches(relative, allow.get("path_globs", []))
                    for allow in allows
                ):
                    continue
                for match in pattern.finditer(text):
                    line = _line_number(text, match.start())
                    marker = (pattern_id, line)
                    if marker in seen:
                        continue
                    seen.add(marker)
                    self.add(
                        "portable_leakage",
                        "LEAK001",
                        path,
                        line,
                        "portable-core deny rule matched: %s" % pattern_id,
                    )

    def check_trailing_whitespace(self) -> None:
        settings = self.config["trailing_whitespace"]
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            text = self.read_text(path, "trailing_whitespace")
            if text is None:
                continue
            affected: List[int] = []
            for line_number, line in enumerate(text.splitlines(), start=1):
                if line.endswith(" ") or line.endswith("\t"):
                    affected.append(line_number)
            if affected:
                preview = ", ".join(str(number) for number in affected[:12])
                suffix = " ..." if len(affected) > 12 else ""
                self.add(
                    "trailing_whitespace",
                    "WS001",
                    path,
                    affected[0],
                    "%d line(s) have trailing whitespace: %s%s"
                    % (len(affected), preview, suffix),
                )

    def check_cursor_adapter(self) -> None:
        settings = self.config["cursor_adapter"]
        mapping_path = self.root / settings["mapping_path"]
        wrapper_path = self.root / settings["wrapper_path"]
        mapping_text = self.read_text(mapping_path, "cursor_adapter")
        wrapper_text = self.read_text(wrapper_path, "cursor_adapter")
        if mapping_text is None or wrapper_text is None:
            return

        mapping = parse_yaml_like(mapping_text)
        for line, error in mapping.errors:
            self.add("cursor_adapter", "ADAPT001", mapping_path, line, error)
        if mapping.fields.get("default_model") != "inherit":
            self.add(
                "cursor_adapter",
                "ADAPT002",
                mapping_path,
                mapping.lines.get("default_model", 1),
                "Cursor adapter default_model must be inherit",
            )

        items = mapping.list_items.get("agents", [])
        if not items:
            self.add(
                "cursor_adapter",
                "ADAPT003",
                mapping_path,
                1,
                "Cursor adapter must configure at least one agent",
            )
        names: Set[str] = set()
        name_pattern = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")
        configured_readonly = set(settings.get("required_readonly_roles", []))
        for item in items:
            name_path = item + ".name"
            name = mapping.fields.get(name_path, "")
            if not name or not name_pattern.fullmatch(name) or name in names:
                self.add(
                    "cursor_adapter",
                    "ADAPT004",
                    mapping_path,
                    mapping.lines.get(name_path, mapping.lines.get(item, 1)),
                    "agent names must be unique nonempty kebab-case values",
                )
            names.add(name)

            description_path = item + ".description"
            if not mapping.fields.get(description_path, "").strip():
                self.add(
                    "cursor_adapter",
                    "ADAPT005",
                    mapping_path,
                    mapping.lines.get(description_path, mapping.lines.get(item, 1)),
                    "agent description must be nonempty",
                )

            readonly_path = item + ".readonly"
            readonly = mapping.fields.get(readonly_path, "")
            if (
                readonly not in {"true", "false"}
                or mapping.scalar_types.get(readonly_path) != "bool"
            ):
                self.add(
                    "cursor_adapter",
                    "ADAPT006",
                    mapping_path,
                    mapping.lines.get(readonly_path, mapping.lines.get(item, 1)),
                    "agent readonly must be an explicit YAML boolean",
                )
            if (
                name in configured_readonly
                or name.endswith("-reviewer")
                or name.endswith("-auditor")
            ) and readonly != "true":
                self.add(
                    "cursor_adapter",
                    "ADAPT007",
                    mapping_path,
                    mapping.lines.get(readonly_path, mapping.lines.get(item, 1)),
                    "configured reviewer/auditor roles must be readonly",
                )

            core_path = item + ".core_file"
            core_file = mapping.fields.get(core_path, "")
            if not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*\.md", core_file):
                self.add(
                    "cursor_adapter",
                    "ADAPT008",
                    mapping_path,
                    mapping.lines.get(core_path, mapping.lines.get(item, 1)),
                    "agent core_file must be a safe Markdown basename",
                )
                continue
            candidate = self.root / settings["core_directory"] / core_file
            if not self._snapshot_is_file(candidate):
                self.add(
                    "cursor_adapter",
                    "ADAPT009",
                    mapping_path,
                    mapping.lines.get(core_path, mapping.lines.get(item, 1)),
                    "agent core_file does not resolve to a safe existing file",
                )

        for token in settings["required_wrapper_tokens"]:
            if token not in wrapper_text:
                self.add(
                    "cursor_adapter",
                    "ADAPT010",
                    wrapper_path,
                    1,
                    "wrapper template lacks required metadata/policy reference",
                )
        lowered = wrapper_text.lower()
        for claim in settings.get("forbidden_offline_claims", []):
            if claim.lower() in lowered:
                self.add(
                    "cursor_adapter",
                    "ADAPT011",
                    wrapper_path,
                    1,
                    "offline wrapper must not claim runtime catalog/data-policy verification",
                )

    def check_contracts(self) -> None:
        """Validate the bundled record-schema registry and its examples."""

        settings = self.config.get("contract_schemas") or {}
        dialect = settings.get("dialect")
        configured = settings.get("schemas") or {}
        seen_ids: Dict[str, Path] = {}
        for title, relative in configured.items():
            path = self.root / relative
            if not self._snapshot_is_file(path):
                self.add("contracts", "CONTRACT001", path, 1, "required contract schema is missing")
                continue
            text = self.read_text(path, "contracts")
            if text is None:
                continue
            try:
                schema = json.loads(text)
            except json.JSONDecodeError as exc:
                self.add("contracts", "CONTRACT002", path, exc.lineno, "contract schema is not valid JSON")
                continue
            if not isinstance(schema, dict):
                self.add("contracts", "CONTRACT003", path, 1, "contract schema root must be an object")
                continue
            properties = schema.get("properties")
            required = schema.get("required")
            identity_valid = (
                schema.get("$schema") == dialect
                and schema.get("title") == title
                and schema.get("type") == "object"
                and schema.get("additionalProperties") is False
                and isinstance(schema.get("$id"), str)
                and bool(schema.get("$id"))
                and isinstance(properties, dict)
                and isinstance(required, list)
                and properties.get("schema_version", {}).get("const") == "1.0"
                and properties.get("record_type", {}).get("const") == title
                and {"schema_version", "record_type"}.issubset(set(required))
            )
            if not identity_valid:
                self.add(
                    "contracts",
                    "CONTRACT003",
                    path,
                    1,
                    "contract schema has an invalid dialect/root/record identity",
                )
            schema_id = schema.get("$id")
            if isinstance(schema_id, str) and schema_id:
                if schema_id in seen_ids:
                    self.add("contracts", "CONTRACT004", path, 1, "contract schema $id is duplicated")
                else:
                    seen_ids[schema_id] = path

            pending: List[Any] = [schema]
            while pending:
                node = pending.pop()
                if isinstance(node, dict):
                    reference = node.get("$ref")
                    if reference is not None:
                        try:
                            if not isinstance(reference, str):
                                raise ValueError("$ref must be a string")
                            _resolve_local_json_ref(schema, reference)
                        except ValueError:
                            self.add(
                                "contracts",
                                "CONTRACT005",
                                path,
                                1,
                                "contract schema contains unsupported or unresolved $ref",
                            )
                    pending.extend(node.values())
                elif isinstance(node, list):
                    pending.extend(node)

            examples = schema.get("examples", [])
            if not isinstance(examples, list) or not examples:
                self.add("contracts", "CONTRACT006", path, 1, "contract schema needs a valid example")
                continue
            for index, example in enumerate(examples):
                errors = validate_json_instance(example, schema)
                if errors:
                    self.add(
                        "contracts",
                        "CONTRACT006",
                        path,
                        1,
                        "contract example %d is invalid: %s" % (index, errors[0]),
                    )

    def check_provenance(self) -> None:
        settings = self.config["toolkit_manifest"]
        path = self.root / settings["path"]
        if not self._snapshot_is_file(path):
            self.add(
                "provenance",
                "PROV001",
                path,
                1,
                "toolkit manifest is missing",
            )
            return
        text = self.read_text(path, "provenance")
        if text is None:
            return
        frontmatter = parse_frontmatter(text)
        if not frontmatter.present:
            self.add(
                "provenance",
                "PROV002",
                path,
                1,
                "toolkit manifest frontmatter is missing",
            )
            return
        for field in settings["required_fields"]:
            if not frontmatter.fields.get(field):
                self.add(
                    "provenance",
                    "PROV003",
                    path,
                    1,
                    "toolkit manifest lacks required field: %s" % field,
                )
        source_field = settings.get("source_registry_field")
        if source_field and frontmatter.fields.get(source_field):
            source = self.root / frontmatter.fields[source_field]
            if not self._snapshot_exists(source):
                self.add(
                    "provenance",
                    "PROV004",
                    path,
                    frontmatter.lines.get(source_field, 1),
                    "configured source registry does not exist",
                )

    def check_recipes(self) -> None:
        """Структура рецептов против recipes/_SCHEMA.md.

        До появления этой проверки соответствие рецептов схеме не
        проверялось ничем: секции разъезжались, а неверное имя секции
        жило с момента создания файла и валидатор молчал.
        """
        config = self.config.get("recipes") or {}
        required = list(config.get("required_sections") or [])
        optional = set(config.get("optional_sections") or [])
        if not required:
            return
        recipes_dir = self.root / "workflows" / "recipes"
        if not self._snapshot_is_dir(recipes_dir):
            return
        for path in self._snapshot_files_in(recipes_dir, suffix=".md"):
            name = path.name
            if name.startswith("_") or name == "README.md":
                continue
            text = self.read_text(path, "recipes")
            if text is None:
                continue
            found = [
                line[3:].strip()
                for line in text.splitlines()
                if line.startswith("## ")
            ]
            for section in required:
                if section not in found:
                    self.add(
                        "recipes",
                        "RCP001",
                        path,
                        1,
                        "missing required section: %s" % section,
                    )
            allowed = set(required) | optional
            for section in found:
                if section not in allowed:
                    self.add(
                        "recipes",
                        "RCP002",
                        path,
                        1,
                        "unknown section not in schema: %s" % section,
                    )
            order = [s for s in found if s in required]
            if order != [s for s in required if s in order]:
                self.add(
                    "recipes",
                    "RCP003",
                    path,
                    1,
                    "required sections out of schema order",
                )

    def check_size_budget(self) -> None:
        """Бюджет размера файлов инструкций.

        Пакет про контекстную инженерию не имеет права грузить в агента
        файл, который расходует его внимание до начала работы. Правило
        без машинной проверки протухает: размер растёт по строке за
        правку, и порог переходится незаметно.
        """
        settings = self.config.get("size_budget")
        if not settings:
            return
        tiers = settings.get("tiers") or {}
        default_kb = int(tiers.get("default_kb", settings.get("default_kb", 38)))
        ceiling_kb = int(tiers.get("reference_kb", 90))
        overrides = settings.get("overrides") or {}
        reasons = settings.get("override_reasons") or {}
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            if self.tree is None:
                continue
            size_kb = self.tree.metadata(path).st_size / 1024
            rel = path.relative_to(self.root).as_posix()
            limit = int(overrides.get(rel, default_kb))
            if rel in overrides and not reasons.get(rel):
                self.add(
                    "size_budget",
                    "SIZE002",
                    path,
                    1,
                    "size budget override needs a recorded reason",
                )
            if limit > ceiling_kb:
                self.add(
                    "size_budget",
                    "SIZE003",
                    path,
                    1,
                    "budget override of %dKB exceeds the %dKB ceiling" % (limit, ceiling_kb),
                )
            if size_kb > limit:
                self.add(
                    "size_budget",
                    "SIZE001",
                    path,
                    1,
                    "instruction file is %.1fKB, over its %dKB budget" % (size_kb, limit),
                )

    def _text_lines(self, path):
        """Строки вне кодовых блоков — с номерами, начиная с 1."""
        raw = self.read_text(path, "text_lines")
        if raw is None:
            return []
        out = []
        fenced = False
        for number, line in enumerate(raw.splitlines(), 1):
            if line.lstrip().startswith("```"):
                fenced = not fenced
                continue
            if not fenced:
                out.append((number, line))
        return out

    def check_vocabulary(self) -> None:
        """Понятия, удалённые из пакета, но оставшиеся в тексте.

        Классы дефектов, которые не ловит ни одна структурная проверка:
        ссылка на удалённое издание, на убранную подсистему, на скрипт,
        которого больше нет. Файл при этом валиден и связен — он просто
        обещает то, чего нет.
        """
        settings = self.config.get("vocabulary")
        if not settings:
            return
        allow = settings.get("allow_globs", [])
        for entry in settings.get("forbidden", []):
            pattern = re.compile(entry["pattern"], re.IGNORECASE)
            hint = entry.get("hint", "")
            excludes = list(entry.get("exclude_globs", [])) + list(allow)
            for path in self.files(settings["include_globs"], excludes):
                for number, line in self._text_lines(path):
                    if pattern.search(line):
                        self.add(
                            "vocabulary",
                            "VOCAB001",
                            path,
                            number,
                            "retired concept in text: %s%s" % (entry["pattern"], hint and " — " + hint),
                        )

    def check_enums(self) -> None:
        """Разнобой машинных значений.

        Одно значение, записанное двумя способами, — расхождение контракта:
        принимающая сторона сверяет строку, а не смысл.
        """
        settings = self.config.get("enums")
        if not settings:
            return
        for group in settings.get("groups", []):
            canonical = set(group["values"])
            pattern = re.compile(group["detect"])
            for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
                for number, line in self._text_lines(path):
                    for found in pattern.findall(line):
                        if found not in canonical:
                            self.add(
                                "enums",
                                "ENUM001",
                                path,
                                number,
                                "value %s is not in the %s enum" % (found, group["name"]),
                            )

    def check_roster(self) -> None:
        """Ссылки на роли, которых нет в ростере."""
        settings = self.config.get("roster")
        if not settings:
            return
        agents_dir = self.root / "agents"
        if not self._snapshot_is_dir(agents_dir):
            return
        known = {
            path.stem
            for path in self._snapshot_files_in(agents_dir, suffix=".md")
            if path.name != "README.md"
        }
        pattern = re.compile(settings["detect"])
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            for number, line in self._text_lines(path):
                for found in pattern.findall(line):
                    if found not in known:
                        self.add(
                            "roster",
                            "ROLE001",
                            path,
                            number,
                            "role %s is not in agents/" % found,
                        )

    def check_code_spans(self) -> None:
        """Пустые и разорванные вставки кода."""
        settings = self.config.get("code_spans")
        if not settings:
            return
        empty = re.compile(r"(?<!`)``(?!`)")
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            for number, line in self._text_lines(path):
                if empty.search(line):
                    self.add("code_spans", "SPAN001", path, number, "empty code span")
                elif line.count("`") % 2:
                    self.add("code_spans", "SPAN002", path, number, "unbalanced code span")

    def check_duplication(self) -> None:
        """Повтор текста между файлами.

        Копия абзаца стоит токенов дважды: у отправителя и у получателя
        контекста. Хуже цены — расхождение: правят одну копию, вторая
        продолжает учить старому. Правило пакета — ссылка вместо
        пересказа; здесь оно проверяется.
        """
        settings = self.config.get("duplication")
        if not settings:
            return
        threshold = float(settings.get("threshold", 0.5))
        minimum = int(settings.get("min_chars", 150))
        shingle = int(settings.get("shingle", 5))
        allowed = [tuple(pair) for pair in settings.get("allow_pairs", [])]
        word = re.compile(r"[а-яёa-z0-9]+", re.IGNORECASE)
        fence = re.compile(r"```.*?```", re.S)

        blocks = []
        for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
            text = self.read_text(path, "duplication")
            if text is None:
                continue
            rel = path.relative_to(self.root).as_posix()
            for chunk in re.split(_BLANK_LINE, fence.sub("", text)):
                flat = " ".join(chunk.split())
                if len(flat) < minimum or flat[:1] in ("|", "#", ">"):
                    continue
                words = [m.group(0).lower() for m in word.finditer(flat)]
                grams = {
                    tuple(words[i:i + shingle])
                    for i in range(max(0, len(words) - shingle + 1))
                }
                if len(grams) >= 12:
                    blocks.append((rel, path, flat, grams))

        reported = set()
        for i in range(len(blocks)):
            for j in range(i + 1, len(blocks)):
                left, path, _, grams = blocks[i]
                right, _, _, other = blocks[j]
                if left == right:
                    continue
                pair = tuple(sorted((left, right)))
                if pair in reported or pair in allowed:
                    continue
                union = len(grams | other)
                if not union:
                    continue
                overlap = len(grams & other) / union
                if overlap >= threshold:
                    reported.add(pair)
                    self.add(
                        "duplication",
                        "DUP001",
                        path,
                        1,
                        "text repeats %s at %d%% overlap; link instead of restating"
                        % (right, round(overlap * 100)),
                    )

    def check_handoff(self) -> None:
        """Переход между ролями описан с обеих сторон.

        Роль A пишет, что отдаёт результат роли B; у B этого во входах
        нет. Односторонне описанный переход разъезжается при первом же
        изменении: A правит формат, B продолжает ждать прежний.
        """
        settings = self.config.get("handoff")
        if not settings:
            return
        agents_dir = self.root / "agents"
        if not self._snapshot_is_dir(agents_dir):
            return
        token = re.compile(settings.get("token", r"`(test-[a-z-]+)`"))
        inbound_title = settings["inbound_section"]
        outbound_title = settings["outbound_section"]

        roles = {}
        for path in self._snapshot_files_in(agents_dir, suffix=".md"):
            if path.name.startswith("_") or path.name == "README.md":
                continue
            text = self.read_text(path, "handoff")
            if text is None:
                continue

            def section(title):
                found = re.search(
                    "^## " + re.escape(title) + chr(10) + "(.*?)(?=^## )",
                    text,
                    re.S | re.M,
                )
                return found.group(1) if found else ""

            roles[path.stem] = (
                path,
                set(token.findall(section(inbound_title))),
                set(token.findall(section(outbound_title))),
            )

        for name, (path, inbound, outbound) in roles.items():
            for target in sorted(outbound):
                if target in roles and name not in roles[target][1]:
                    self.add(
                        "handoff",
                        "HAND001",
                        path,
                        1,
                        "hands off to %s, but %s does not list it as an input"
                        % (target, target),
                    )
            for source in sorted(inbound):
                if source in roles and name not in roles[source][2]:
                    self.add(
                        "handoff",
                        "HAND002",
                        path,
                        1,
                        "expects input from %s, but %s does not list it as a consumer"
                        % (source, source),
                    )

    def check_egress(self) -> None:
        """Следы самовольного выхода наружу в текстах пакета.

        Пакет учит не ходить в сеть без разрешения — и не имеет права сам
        приводить примеры, где агент это делает. Ловятся два класса:
        адрес, собранный из содержимого репозитория, и указание сходить
        наружу без упоминания разрешения.
        """
        settings = self.config.get("egress")
        if not settings:
            return
        near = int(settings.get("permission_window", 3))
        permission = re.compile(settings["permission_pattern"], re.IGNORECASE)
        for entry in settings.get("patterns", []):
            trigger = re.compile(entry["pattern"], re.IGNORECASE)
            for path in self.files(settings["include_globs"], settings.get("exclude_globs", [])):
                lines = self._text_lines(path)
                index = {number: text for number, text in lines}
                for number, line in lines:
                    if not trigger.search(line):
                        continue
                    window = " ".join(
                        index.get(number + shift, "")
                        for shift in range(-near, near + 1)
                    )
                    if permission.search(window):
                        continue
                    self.add(
                        "egress",
                        entry.get("code", "EGRESS001"),
                        path,
                        number,
                        "%s without a nearby permission or prohibition marker"
                        % entry["hint"],
                    )

    def _contract_scan(self, check: str, config: Dict[str, Any], folder: str,
                       codes: Sequence[str]) -> None:
        """Общая проверка формы файла: состав, порядок, чеклист.

        Одна реализация на роли и скилы: два независимых прохода
        разъезжаются, и второй начинает пропускать то, что ловит первый.
        """
        required = list(config.get("required_sections") or [])
        if not required:
            return
        folder_path = self.root / folder
        if not self._snapshot_is_dir(folder_path):
            return
        checklist_section = config.get("checklist_section")
        missing_code, order_code, checklist_code = codes
        for path in self._snapshot_files_in(folder_path, suffix=".md"):
            if path.name.startswith("_") or path.name == "README.md":
                continue
            text = self.read_text(path, check)
            if text is None:
                continue
            found = []
            fenced = False
            for line in text.splitlines():
                if line.lstrip().startswith("```"):
                    fenced = not fenced
                    continue
                if not fenced and line.startswith("## "):
                    found.append(line[3:].strip())
            for section in required:
                if section not in found:
                    self.add(check, missing_code, path, 1,
                             "file is missing the required section: %s" % section)
            ordered = [item for item in found if item in required]
            expected = [item for item in required if item in ordered]
            if ordered != expected:
                self.add(check, order_code, path, 1,
                         "required sections are out of the contracted order")
            if checklist_section and checklist_section in found:
                body = text.split("## %s" % checklist_section, 1)[1]
                body = body.split(chr(10) + "## ", 1)[0]
                if "- [ ]" not in body:
                    self.add(check, checklist_code, path, 1,
                             "section %s must contain a checklist" % checklist_section)

    def check_skill_contract(self) -> None:
        """Единая форма файла скила против skills/README.md."""
        self._contract_scan("skill_contract", self.config.get("skill_contract") or {},
                            "skills", ("SKILL010", "SKILL011", "SKILL012"))

    def check_agent_contract(self) -> None:
        """Единая форма файла роли против agents/README.md."""
        self._contract_scan("agent_contract", self.config.get("agent_contract") or {},
                            "agents", ("AGENT001", "AGENT002", "AGENT003"))

    def run(self, checks: Sequence[str]) -> List[Issue]:
        if self._tree_error is not None:
            message = str(self._tree_error)
            self.add(
                "file_policy",
                "SAFEIO001",
                self.root,
                1,
                message,
            )
            if "max file count" in message:
                self.add(
                    "file_policy",
                    "FILE005",
                    self.root,
                    1,
                    "validation tree exceeds max file count",
                )
            if "max total bytes" in message:
                self.add(
                    "file_policy",
                    "FILE004",
                    self.root,
                    1,
                    "validation tree exceeds max total bytes",
                )
            return sorted(
                self.issues,
                key=lambda item: (item.path, item.line, item.check, item.code),
            )
        methods = {
            "file_policy": self.check_file_policy,
            "links": self.check_links,
            "inline_paths": self.check_inline_paths,
            "frontmatter": self.check_frontmatter,
            "skills": self.check_skills,
            "prompt_ids": self.check_prompt_ids,
            "placeholders": self.check_placeholders,
            "cot_disclosure": self.check_cot_disclosure,
            "gsd_routes": self.check_gsd_routes,
            "secrets": self.check_secrets,
            "portable_leakage": self.check_portable_leakage,
            "trailing_whitespace": self.check_trailing_whitespace,
            "contracts": self.check_contracts,
            "provenance": self.check_provenance,
            "recipes": self.check_recipes,
            "agent_contract": self.check_agent_contract,
            "skill_contract": self.check_skill_contract,
            "size_budget": self.check_size_budget,
            "vocabulary": self.check_vocabulary,
            "enums": self.check_enums,
            "roster": self.check_roster,
            "code_spans": self.check_code_spans,
            "duplication": self.check_duplication,
            "handoff": self.check_handoff,
            "egress": self.check_egress,
        }
        for check in checks:
            try:
                methods[check]()
            except FileSafetyError as exc:
                self.add("file_policy", "SAFEIO001", self.root, 1, str(exc))
                break
        return sorted(
            self.issues,
            key=lambda item: (item.path, item.line, item.check, item.code),
        )


def load_config(path: Path) -> Dict[str, Any]:
    try:
        config_root = resolve_directory_root(path.parent, "validation manifest parent")
        config_tree = SafeTree(
            config_root,
            TextPolicy(
                allowed_text_extensions=(".json",),
                allowed_text_names=(),
                max_file_bytes=2 * 1024 * 1024,
                max_total_bytes=64 * 1024 * 1024,
                max_file_count=5000,
            ),
            exclude=lambda relative, is_directory: is_directory
            and Path(relative).name in {"__pycache__", ".pytest_cache"},
        )
        config = json.loads(config_tree.inspect_text(path).text)
    except (
        json.JSONDecodeError,
        FileSafetyError,
    ) as exc:
        raise ValueError("cannot load validation manifest: %s" % type(exc).__name__) from exc
    if not isinstance(config, dict):
        raise ValueError("validation manifest must contain a JSON object")
    if config.get("schema_version") != 1:
        raise ValueError("validation manifest schema_version must be 1")
    return config


def parse_checks(raw: str) -> List[str]:
    if raw == "all":
        return list(CHECK_ORDER)
    checks = [item.strip() for item in raw.split(",") if item.strip()]
    unknown = sorted(set(checks) - set(CHECK_ORDER))
    if unknown:
        raise ValueError("unknown checks: %s" % ", ".join(unknown))
    return checks


def render_text(root: Path, checks: Sequence[str], issues: Sequence[Issue]) -> str:
    errors = sum(issue.severity == "ERROR" for issue in issues)
    warnings = sum(issue.severity == "WARNING" for issue in issues)
    status = "PASS" if not errors else "FAIL"
    output = ["Toolkit validation: %s" % status, "Root: %s" % root]
    for issue in issues:
        output.append(
            "[%s] %s %s:%d %s"
            % (issue.severity, issue.code, issue.path, issue.line, issue.message)
        )
    output.append(
        "Summary: %d error(s), %d warning(s); checks=%s"
        % (errors, warnings, ",".join(checks))
    )
    return "\n".join(output)


def main(argv: Optional[Sequence[str]] = None) -> int:
    parser = argparse.ArgumentParser(
        description="Validate the portable test-automation AI toolkit."
    )
    parser.add_argument(
        "--root",
        type=Path,
        default=DEFAULT_ROOT,
        help="pro-qa-ai root (default: script parent)",
    )
    parser.add_argument(
        "--manifest",
        type=Path,
        help="validation JSON manifest (default: <root>/scripts/%s)"
        % DEFAULT_CONFIG_NAME,
    )
    parser.add_argument(
        "--checks",
        default="all",
        help="comma-separated check names or 'all'",
    )
    parser.add_argument(
        "--format",
        choices=("text", "json"),
        default="text",
        help="output format",
    )
    parser.add_argument(
        "--fail-on",
        choices=("error", "warning", "never"),
        default="error",
        help="exit-code threshold",
    )
    args = parser.parse_args(argv)

    root = lexical_absolute(args.root)
    manifest = (
        lexical_absolute(args.manifest)
        if args.manifest
        else root / "scripts" / DEFAULT_CONFIG_NAME
    )
    try:
        config = load_config(manifest)
        checks = parse_checks(args.checks)
        issues = ToolkitValidator(root, config).run(checks)
    except (KeyError, TypeError, ValueError, re.error) as exc:
        print("validation configuration error: %s" % exc, file=sys.stderr)
        return 2

    if args.format == "json":
        payload = {
            "root": str(root),
            "status": "FAIL"
            if any(issue.severity == "ERROR" for issue in issues)
            else "PASS",
            "checks": checks,
            "issues": [asdict(issue) for issue in issues],
        }
        print(json.dumps(payload, ensure_ascii=False, indent=2))
    else:
        print(render_text(root, checks, issues))

    if args.fail_on == "never":
        return 0
    if args.fail_on == "warning":
        return 1 if issues else 0
    return 1 if any(issue.severity == "ERROR" for issue in issues) else 0


if __name__ == "__main__":
    raise SystemExit(main())
