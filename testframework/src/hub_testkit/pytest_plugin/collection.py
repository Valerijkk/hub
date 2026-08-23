from __future__ import annotations

import os
import stat
from collections.abc import Iterable
from pathlib import Path
from typing import Any

from hub_testkit.runtime.redaction import Redactor
from hub_testkit.traceability import (
    CaseIndex,
    CaseRecord,
    MarkerPolicy,
    PytestAutomation,
    load_case,
)

from .markers import CollectionError, case_id_from_item, layer_from_item, reject_blocked_markers
from .reporting import apply_report_properties


MAX_CASES = 512
MAX_ENTRIES = 2048
MAX_DEPTH = 16
MAX_RELATIVE_PATH_BYTES = 512
MAX_RELATIVE_PATH_CHARS = 512


def discover_case_paths(root: str | Path) -> tuple[Path, ...]:
    try:
        case_root = Path(root)
        root_stat = os.lstat(case_root)
    except (OSError, TypeError, ValueError):
        raise CollectionError("case root is invalid") from None
    if stat.S_ISLNK(root_stat.st_mode) or not stat.S_ISDIR(root_stat.st_mode):
        raise CollectionError("case root is invalid")

    paths: list[Path] = []
    entries = 0
    try:
        for current_root, directory_names, file_names in os.walk(
            case_root,
            topdown=True,
            followlinks=False,
        ):
            current = Path(current_root)
            _require_directory(current)
            ordered_directories = sorted(directory_names)
            ordered_files = sorted(file_names)
            entries += len(ordered_directories) + len(ordered_files)
            if entries > MAX_ENTRIES:
                raise CollectionError("case root exceeds discovery limits")
            directory_names[:] = []
            for name in ordered_directories:
                child = current / name
                _require_relative_path(case_root, child, is_directory=True)
                _require_directory(child)
                if not name.startswith("."):
                    directory_names.append(name)
            for name in ordered_files:
                child = current / name
                _require_relative_path(case_root, child, is_directory=False)
                child_stat = os.lstat(child)
                if stat.S_ISLNK(child_stat.st_mode):
                    raise CollectionError("case source is invalid")
                if name == "README.md" or name.startswith(".") or child.suffix != ".md":
                    continue
                if not stat.S_ISREG(child_stat.st_mode):
                    raise CollectionError("case source is invalid")
                paths.append(child)
                if len(paths) > MAX_CASES:
                    raise CollectionError("case root exceeds discovery limits")
    except CollectionError:
        raise
    except (OSError, TypeError, ValueError):
        raise CollectionError("case root is invalid") from None
    return tuple(sorted(paths, key=lambda path: path.as_posix()))


def load_case_index(
    case_root: str | Path,
    schema_path: str | Path,
    marker_policy: MarkerPolicy,
) -> CaseIndex:
    paths = discover_case_paths(case_root)
    if not paths:
        raise CollectionError("case sources are empty")
    try:
        records = tuple(load_case(path, schema_path, marker_policy) for path in paths)
        return CaseIndex.build(records)
    except Exception:
        raise CollectionError("case sources are invalid") from None


def validate_collection(
    items: Iterable[Any],
    case_index: CaseIndex,
    redactor: Redactor,
) -> None:
    try:
        collected = tuple(items)
    except Exception:
        raise CollectionError("pytest collection is invalid") from None
    if not collected:
        raise CollectionError("pytest collection is empty")

    node_ids: list[str] = []
    bound: list[tuple[Any, CaseRecord, str]] = []
    for item in collected:
        reject_blocked_markers(item)
        case_id = case_id_from_item(item)
        layer = layer_from_item(item)
        node_id = _node_id(item)
        record = case_index.lookup(case_id)
        automation = record.metadata.automation
        if record.metadata.status != "active" or not isinstance(automation, PytestAutomation):
            raise CollectionError("case is not active pytest automation")
        if automation.node_id != node_id or automation.layer != layer:
            raise CollectionError("case automation does not match item")
        if not node_id.startswith(f"testframework/tests/{layer}/"):
            raise CollectionError("item node ID is invalid")
        node_ids.append(node_id)
        bound.append((item, record, layer))

    try:
        case_index.validate_pytest_collection(node_ids)
        apply_report_properties(bound, redactor)
    except Exception:
        raise CollectionError("pytest collection is invalid") from None


def _require_directory(path: Path) -> None:
    try:
        information = os.lstat(path)
    except (OSError, TypeError, ValueError):
        raise CollectionError("case source is invalid") from None
    if stat.S_ISLNK(information.st_mode) or not stat.S_ISDIR(information.st_mode):
        raise CollectionError("case source is invalid")


def _require_relative_path(root: Path, path: Path, *, is_directory: bool) -> None:
    try:
        relative = path.relative_to(root)
        relative_text = relative.as_posix()
        depth = len(relative.parts) if is_directory else len(relative.parts) - 1
        relative_bytes = len(relative_text.encode("utf-8"))
    except (TypeError, ValueError, UnicodeError):
        raise CollectionError("case source is invalid") from None
    if (
        depth > MAX_DEPTH
        or len(relative_text) > MAX_RELATIVE_PATH_CHARS
        or relative_bytes > MAX_RELATIVE_PATH_BYTES
    ):
        raise CollectionError("case root exceeds discovery limits")


def _node_id(item: Any) -> str:
    try:
        node_id = item.nodeid
    except Exception:
        raise CollectionError("item node ID is invalid") from None
    if type(node_id) is not str:
        raise CollectionError("item node ID is invalid")
    return node_id
