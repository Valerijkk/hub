#!/usr/bin/env python3
"""Build a bounded, no-command inventory of explicit project/test roots."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import stat
import sys
from datetime import datetime, timezone
from functools import lru_cache
from pathlib import Path
from typing import Any, Dict, List, Mapping, Optional, Sequence, Tuple

from safe_filesystem import (
    FileSafetyError,
    TextPolicy,
    assert_no_link_components,
    enforce_containment,
    inspect_text_file,
    is_link_like,
    lexical_absolute,
    parse_text_policy,
    resolve_directory_root,
    safe_open_read,
    safe_open_write,
    text_path_allowed,
)


DEFAULT_CONFIG_PATH = Path(__file__).resolve().with_name(
    "context_inventory_config.json"
)
CONFIG_LIMIT = 1024 * 1024


class InventoryError(ValueError):
    """A safe inventory configuration or boundary error."""


def _glob_expression(pattern: str) -> str:
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


@lru_cache(maxsize=None)
def _compiled_glob(pattern: str) -> re.Pattern[str]:
    return re.compile(r"^(?:%s)$" % _glob_expression(pattern.replace("\\", "/")))


def _matches(path: str, patterns: Sequence[str]) -> bool:
    normalized = path.replace("\\", "/")
    return any(_compiled_glob(pattern).fullmatch(normalized) for pattern in patterns)


def _load_json_bounded(path: Path) -> Tuple[Dict[str, Any], bytes]:
    try:
        parent = resolve_directory_root(path.parent, "inventory config parent")
        with safe_open_read(parent, path) as source:
            data = source.read(CONFIG_LIMIT + 1)
        if len(data) > CONFIG_LIMIT or b"\x00" in data:
            raise InventoryError("inventory config is not bounded UTF-8 text")
        decoded = json.loads(data.decode("utf-8"))
    except (
        OSError,
        UnicodeDecodeError,
        json.JSONDecodeError,
        FileSafetyError,
    ) as exc:
        raise InventoryError(
            "cannot load inventory config (%s)" % type(exc).__name__
        ) from exc
    if not isinstance(decoded, dict) or decoded.get("schema_version") != 1:
        raise InventoryError("inventory config schema_version must be 1")
    return decoded, data


def load_config(path: Path = DEFAULT_CONFIG_PATH) -> Dict[str, Any]:
    config, _ = _load_json_bounded(path)
    parse_inventory_config(config)
    return config


def _string_list(config: Mapping[str, Any], field: str) -> Tuple[str, ...]:
    values = config.get(field)
    if not isinstance(values, list):
        raise InventoryError("%s must be an array" % field)
    output: List[str] = []
    for index, value in enumerate(values):
        if not isinstance(value, str) or not value or "\x00" in value:
            raise InventoryError("%s[%d] must be a nonempty string" % (field, index))
        normalized = value.replace("\\", "/")
        if normalized in output:
            raise InventoryError("%s[%d] duplicates an earlier value" % (field, index))
        output.append(normalized)
    return tuple(output)


def parse_inventory_config(
    config: Mapping[str, Any],
) -> Tuple[TextPolicy, Tuple[str, ...], Tuple[str, ...]]:
    try:
        policy = parse_text_policy(config)
    except FileSafetyError as exc:
        raise InventoryError(str(exc)) from exc
    excluded_names = _string_list(config, "exclude_directory_names")
    excluded_globs = _string_list(config, "exclude_globs")
    for pattern in excluded_globs:
        try:
            _compiled_glob(pattern)
        except re.error as exc:
            raise InventoryError("exclude_globs contains invalid syntax") from exc
    return policy, excluded_names, excluded_globs


def _iso_timestamp(nanoseconds: int) -> str:
    return (
        datetime.fromtimestamp(nanoseconds / 1_000_000_000, tz=timezone.utc)
        .isoformat(timespec="microseconds")
        .replace("+00:00", "Z")
    )


def _canonical_json(value: object) -> bytes:
    return json.dumps(
        value,
        ensure_ascii=False,
        sort_keys=True,
        separators=(",", ":"),
    ).encode("utf-8")


def build_inventory(
    project_root: Path,
    test_roots: Sequence[Path],
    config: Mapping[str, Any],
) -> Dict[str, Any]:
    """Inventory only the explicitly supplied roots; never execute anything."""

    if not test_roots:
        raise InventoryError("at least one explicit test root is required")
    policy, excluded_names, excluded_globs = parse_inventory_config(config)
    try:
        roots = [("project_root", resolve_directory_root(project_root, "project root"))]
        roots.extend(
            (
                "test_root_%02d" % index,
                resolve_directory_root(root, "test root %02d" % index),
            )
            for index, root in enumerate(test_roots, start=1)
        )
    except FileSafetyError as exc:
        raise InventoryError(str(exc)) from exc
    normalized_roots = [os.path.normcase(os.fspath(root)) for _, root in roots]
    if len(set(normalized_roots)) != len(normalized_roots):
        raise InventoryError("declared roots must be distinct")

    files: List[Dict[str, Any]] = []
    exclusions: List[Dict[str, str]] = []
    total_bytes = 0

    def onerror(exc: OSError) -> None:
        raise InventoryError("cannot enumerate declared root (%s)" % type(exc).__name__)

    for root_id, root in roots:
        for directory, directory_names, file_names in os.walk(
            str(root), topdown=True, onerror=onerror, followlinks=False
        ):
            directory_names.sort()
            file_names.sort()
            base = Path(directory)
            try:
                checked = enforce_containment(
                    root, base, must_exist=True, allow_directory=True
                )
                if not checked.is_dir():
                    raise FileSafetyError("inventory component is not a directory")
            except FileSafetyError as exc:
                raise InventoryError(str(exc)) from exc

            retained: List[str] = []
            for name in directory_names:
                candidate = base / name
                relative = candidate.relative_to(root).as_posix()
                try:
                    if is_link_like(candidate):
                        raise InventoryError(
                            "declared root contains a symbolic link or reparse point: %s"
                            % relative
                        )
                except FileSafetyError as exc:
                    raise InventoryError(str(exc)) from exc
                if name in excluded_names or _matches(relative, excluded_globs):
                    exclusions.append(
                        {
                            "root_id": root_id,
                            "path": relative,
                            "reason": "configured-directory"
                            if name in excluded_names
                            else "configured-path",
                        }
                    )
                    continue
                retained.append(name)
            directory_names[:] = retained

            for name in file_names:
                candidate = base / name
                relative = candidate.relative_to(root).as_posix()
                try:
                    if is_link_like(candidate):
                        raise InventoryError(
                            "declared root contains a symbolic link or reparse point: %s"
                            % relative
                        )
                    metadata = os.lstat(candidate)
                    enforce_containment(root, candidate, must_exist=True)
                except (OSError, FileSafetyError) as exc:
                    raise InventoryError(
                        "cannot safely inspect inventory file %s (%s)"
                        % (relative, type(exc).__name__)
                    ) from exc
                if not stat.S_ISREG(metadata.st_mode):
                    raise InventoryError(
                        "declared root contains a non-regular file: %s" % relative
                    )
                if _matches(relative, excluded_globs):
                    exclusions.append(
                        {
                            "root_id": root_id,
                            "path": relative,
                            "reason": "configured-path",
                        }
                    )
                    continue
                if not text_path_allowed(Path(relative), policy):
                    exclusions.append(
                        {
                            "root_id": root_id,
                            "path": relative,
                            "reason": "unapproved-file-type",
                        }
                    )
                    continue
                try:
                    inspected = inspect_text_file(root, candidate, relative, policy)
                except FileSafetyError as exc:
                    message = str(exc)
                    if "byte limit" in message:
                        reason = "oversized-file"
                    elif "binary/NUL" in message or "UTF-8" in message:
                        reason = "binary-or-non-utf8"
                    else:
                        raise InventoryError(message) from exc
                    exclusions.append(
                        {"root_id": root_id, "path": relative, "reason": reason}
                    )
                    continue
                if len(files) >= policy.max_file_count:
                    raise InventoryError("inventory exceeds max file count")
                total_bytes += inspected.size
                if total_bytes > policy.max_total_bytes:
                    raise InventoryError("inventory exceeds max total bytes")
                files.append(
                    {
                        "root_id": root_id,
                        "path": relative,
                        "sha256": inspected.sha256,
                        "size_bytes": inspected.size,
                        "modified_at": _iso_timestamp(inspected.modified_at_ns),
                    }
                )

    files.sort(key=lambda item: (item["root_id"], item["path"]))
    exclusions.sort(key=lambda item: (item["root_id"], item["path"], item["reason"]))
    roots_payload = [{"id": root_id, "path": os.fspath(root)} for root_id, root in roots]
    policy_sha256 = hashlib.sha256(_canonical_json(config)).hexdigest()
    integrity_payload = {
        "schema_version": 1,
        "policy_sha256": policy_sha256,
        "roots": roots_payload,
        "files": files,
        "exclusions": exclusions,
        "file_count": len(files),
        "total_bytes": total_bytes,
    }
    inventory_sha256 = hashlib.sha256(_canonical_json(integrity_payload)).hexdigest()
    return {
        "schema_version": 1,
        "generated_at": datetime.now(timezone.utc)
        .isoformat(timespec="microseconds")
        .replace("+00:00", "Z"),
        **integrity_payload,
        "inventory_sha256": inventory_sha256,
    }


def inventory_bytes(inventory: Mapping[str, Any]) -> bytes:
    return (
        json.dumps(inventory, ensure_ascii=False, indent=2, sort_keys=True) + "\n"
    ).encode("utf-8")


def write_inventory(path: Path, inventory: Mapping[str, Any], roots: Sequence[Path]) -> None:
    output = lexical_absolute(path)
    try:
        assert_no_link_components(output, allow_missing_tail=True)
        if output.exists():
            raise InventoryError("inventory output already exists; choose a new path")
        for root in roots:
            safe_root = resolve_directory_root(root, "declared root")
            try:
                output.relative_to(safe_root)
            except ValueError:
                continue
            raise InventoryError("inventory output must be outside declared roots")
        parent = resolve_directory_root(output.parent, "inventory output parent")
        with safe_open_write(parent, output) as target:
            target.write(inventory_bytes(inventory))
    except FileSafetyError as exc:
        raise InventoryError(str(exc)) from exc


def render_summary(inventory: Mapping[str, Any]) -> str:
    return "\n".join(
        [
            "Context inventory: PASS",
            "File count: %d" % inventory["file_count"],
            "Total bytes: %d" % inventory["total_bytes"],
            "Excluded: %d" % len(inventory["exclusions"]),
            "Inventory SHA-256: %s" % inventory["inventory_sha256"],
        ]
    )


def main(argv: Optional[Sequence[str]] = None) -> int:
    parser = argparse.ArgumentParser(
        description="Inventory explicit project/test roots without commands or network."
    )
    parser.add_argument("--project-root", required=True, type=Path)
    parser.add_argument("--test-root", required=True, action="append", type=Path)
    parser.add_argument("--config", type=Path, default=DEFAULT_CONFIG_PATH)
    parser.add_argument("--output", type=Path)
    parser.add_argument(
        "--print-inventory",
        action="store_true",
        help="print full JSON even when --output is supplied",
    )
    args = parser.parse_args(argv)
    try:
        config = load_config(args.config)
        inventory = build_inventory(args.project_root, args.test_root, config)
        if args.output:
            write_inventory(
                args.output,
                inventory,
                [args.project_root, *args.test_root],
            )
        if args.print_inventory or not args.output:
            sys.stdout.buffer.write(inventory_bytes(inventory))
        else:
            print(render_summary(inventory))
    except InventoryError as exc:
        print("ERROR: %s" % exc, file=sys.stderr)
        return 2
    except Exception:
        print(
            "ERROR: inventory failed safely; no file contents were printed",
            file=sys.stderr,
        )
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
