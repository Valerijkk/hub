#!/usr/bin/env python3
"""Create a project-neutral portable toolkit package with Python stdlib only.

The package policy is read from ``package_manifest.json`` beside this script.
Diagnostics report rule and path locations only; file contents and matched
secret values are never printed.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import shutil
import sys
import tempfile
from dataclasses import dataclass
from functools import lru_cache
from pathlib import Path
from typing import Any, List, Mapping, Optional, Sequence, Tuple

from safe_filesystem import (
    FileSafetyError,
    SafeTree,
    TextPolicy,
    enforce_containment,
    inspect_text_file,
    is_link_like,
    parse_text_policy,
    resolve_directory_root,
    resolve_new_or_empty_directory,
    safe_open_read,
    safe_open_write,
)
from secret_rules import secret_rule_matches


DEFAULT_CONFIG_PATH = Path(__file__).resolve().with_name("package_manifest.json")
BUFFER_SIZE = 1024 * 1024


class PackageError(ValueError):
    """A safe, user-facing packaging or policy error."""


@dataclass(frozen=True)
class PackagePolicy:
    include_globs: Tuple[str, ...]
    exclude_globs: Tuple[str, ...]
    exclude_with_allow_globs: Tuple[str, ...]
    exclude_allow_globs: Tuple[str, ...]
    prune_globs: Tuple[str, ...]
    required_paths: Tuple[str, ...]
    inventory_path: Optional[str]
    file_policy: TextPolicy
    provenance_gate: "ProvenanceGate"


@dataclass(frozen=True)
class ProvenanceGate:
    gsd_exact_revision: str
    contains_substantial_copied_gsd_material: bool
    attribution_paths: Tuple[str, ...]


@dataclass(frozen=True)
class PackageEntry:
    relative_path: str
    source_path: Path
    size: int
    sha256: str


@dataclass(frozen=True)
class PackagePlan:
    root: Path
    entries: Tuple[PackageEntry, ...]
    manifest_sha256: str
    inventory_path: Optional[str]


def _glob_expression(pattern: str) -> str:
    """Translate a portable glob where ``*`` never crosses ``/``."""

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
    return re.compile(r"^(?:%s)$" % _glob_expression(pattern))


def _matches(relative_path: str, patterns: Sequence[str]) -> bool:
    normalized = relative_path.replace("\\", "/")
    return any(_compiled_glob(pattern).fullmatch(normalized) for pattern in patterns)


def _validate_relative(value: Any, field: str, *, literal: bool) -> str:
    if not isinstance(value, str) or not value.strip():
        raise PackageError("%s must be a non-empty relative path" % field)
    normalized = value.replace("\\", "/")
    if "\x00" in normalized:
        raise PackageError("%s contains an invalid character" % field)
    if normalized.startswith(("/", "//")) or re.match(r"^[A-Za-z]:", normalized):
        raise PackageError("%s must be relative" % field)
    parts = normalized.split("/")
    if any(part in {"", ".", ".."} for part in parts):
        raise PackageError("%s must not contain empty or traversal segments" % field)
    if literal and any(character in normalized for character in "*?[]"):
        raise PackageError("%s must not contain glob syntax" % field)
    return normalized


def _string_list(config: Mapping[str, Any], field: str, *, literal: bool) -> Tuple[str, ...]:
    values = config.get(field)
    if not isinstance(values, list):
        raise PackageError("%s must be an array" % field)
    normalized: List[str] = []
    for index, value in enumerate(values):
        item_field = "%s[%d]" % (field, index)
        item = _validate_relative(value, item_field, literal=literal)
        if item in normalized:
            raise PackageError("%s duplicates an earlier entry" % item_field)
        if not literal:
            try:
                _compiled_glob(item)
            except re.error as exc:
                raise PackageError("%s contains invalid glob syntax" % item_field) from exc
        normalized.append(item)
    return tuple(normalized)


def parse_policy(config: Mapping[str, Any]) -> PackagePolicy:
    """Validate and normalize a decoded package manifest."""

    if not isinstance(config, Mapping):
        raise PackageError("package manifest must contain a JSON object")
    if type(config.get("schema_version")) is not int or config["schema_version"] != 1:
        raise PackageError("package manifest schema_version must be 1")

    include_globs = _string_list(config, "include_globs", literal=False)
    if not include_globs:
        raise PackageError("include_globs must not be empty")

    inventory_value = config.get("inventory_path")
    inventory_path: Optional[str]
    if inventory_value is None:
        inventory_path = None
    else:
        inventory_path = _validate_relative(
            inventory_value, "inventory_path", literal=True
        )

    provenance_raw = config.get("provenance_gate")
    if not isinstance(provenance_raw, Mapping):
        raise PackageError("provenance_gate must be an object")
    revision = provenance_raw.get("gsd_exact_revision")
    if (
        not isinstance(revision, str)
        or not revision.strip()
        or "\x00" in revision
        or "\n" in revision
        or "\r" in revision
    ):
        raise PackageError("provenance_gate.gsd_exact_revision must be a value")
    copied = provenance_raw.get("contains_substantial_copied_gsd_material")
    if type(copied) is not bool:
        raise PackageError(
            "provenance_gate.contains_substantial_copied_gsd_material "
            "must be a boolean"
        )
    if copied and revision.strip().upper() == "UNKNOWN":
        raise PackageError(
            "substantial copied GSD material requires an exact upstream revision"
        )
    if copied and not re.fullmatch(r"[0-9a-fA-F]{40,64}", revision.strip()):
        raise PackageError(
            "substantial copied GSD material requires an immutable commit SHA"
        )

    return PackagePolicy(
        include_globs=include_globs,
        exclude_globs=_string_list(config, "exclude_globs", literal=False),
        exclude_with_allow_globs=_string_list(
            config, "exclude_with_allow_globs", literal=False
        ),
        exclude_allow_globs=_string_list(
            config, "exclude_allow_globs", literal=False
        ),
        prune_globs=_string_list(config, "prune_globs", literal=False),
        required_paths=_string_list(config, "required_paths", literal=True),
        inventory_path=inventory_path,
        file_policy=parse_text_policy(config),
        provenance_gate=ProvenanceGate(
            gsd_exact_revision=revision.strip(),
            contains_substantial_copied_gsd_material=copied,
            attribution_paths=_string_list(
                provenance_raw, "attribution_paths", literal=True
            ),
        ),
    )


def load_policy(path: Path = DEFAULT_CONFIG_PATH) -> PackagePolicy:
    """Load a package policy without echoing malformed manifest contents."""

    try:
        config_root = resolve_directory_root(path.parent, "package manifest parent")
        with safe_open_read(config_root, path) as source:
            data = source.read(1024 * 1024 + 1)
        if len(data) > 1024 * 1024 or b"\x00" in data:
            raise PackageError("package manifest is not bounded UTF-8 text")
        text = data.decode("utf-8")
    except (OSError, UnicodeDecodeError, FileSafetyError) as exc:
        raise PackageError(
            "cannot read package manifest (%s)" % type(exc).__name__
        ) from exc
    try:
        decoded = json.loads(text)
    except json.JSONDecodeError as exc:
        raise PackageError(
            "cannot parse package manifest at line %d column %d"
            % (exc.lineno, exc.colno)
        ) from exc
    return parse_policy(decoded)


def _is_relative_to(path: Path, parent: Path) -> bool:
    try:
        path.relative_to(parent)
    except ValueError:
        return False
    return True


def _is_link_like(path: Path) -> bool:
    try:
        return is_link_like(path)
    except FileSafetyError as exc:
        raise PackageError(
            "cannot inspect filesystem link metadata (%s)" % type(exc).__name__
        ) from exc


def _resolved_root(root: Path) -> Path:
    try:
        return resolve_directory_root(root, "source root")
    except FileSafetyError as exc:
        raise PackageError(str(exc)) from exc


def _directory_nonempty(path: Path) -> bool:
    try:
        enforce_containment(path, path, must_exist=True, allow_directory=True)
        return next(path.iterdir(), None) is not None
    except (OSError, FileSafetyError) as exc:
        raise PackageError(
            "cannot inspect output directory (%s)" % type(exc).__name__
        ) from exc


def validate_request(root: Path, output: Path) -> Tuple[Path, Path]:
    """Resolve roots and require a disjoint new or empty output directory."""

    resolved_root = _resolved_root(root)
    try:
        resolved_output = resolve_new_or_empty_directory(output, "output")
    except FileSafetyError as exc:
        raise PackageError(str(exc)) from exc

    if resolved_output.parent == resolved_output:
        raise PackageError("output must not be a filesystem root")
    if resolved_output == resolved_root or _is_relative_to(
        resolved_output, resolved_root
    ):
        raise PackageError("output must be outside the source root")
    if _is_relative_to(resolved_root, resolved_output):
        raise PackageError("source and output trees must be disjoint")

    if resolved_output.exists():
        if _directory_nonempty(resolved_output):
            raise PackageError(
                "existing output is non-empty; choose a new empty output path"
            )
    return resolved_root, resolved_output


def _assert_no_link_components(root: Path, relative_path: str) -> None:
    candidate = root.joinpath(*relative_path.split("/"))
    try:
        enforce_containment(root, candidate, must_exist=True, allow_directory=True)
    except FileSafetyError as exc:
        raise PackageError(
            "source path is unsafe: %s (%s)"
            % (relative_path, type(exc).__name__)
        ) from exc


def _hash_source_file(
    root: Path,
    path: Path,
    relative_path: str,
    policy: TextPolicy,
) -> Tuple[int, str]:
    """Boundedly inspect, secret-scan, and hash an exact package file."""

    try:
        inspected = inspect_text_file(root, path, relative_path, policy)
    except FileSafetyError as exc:
        raise PackageError(str(exc)) from exc
    finding = next(secret_rule_matches(inspected.text), None)
    if finding is not None:
        rule_id, _ = finding
        raise PackageError(
            "secret-like material blocks package file %s (redacted rule %s)"
            % (relative_path, rule_id)
        )
    return inspected.size, inspected.sha256


def _selected(relative_path: str, policy: PackagePolicy) -> bool:
    if not _matches(relative_path, policy.include_globs):
        return False
    if _matches(relative_path, policy.exclude_globs):
        return False
    conditionally_excluded = _matches(
        relative_path, policy.exclude_with_allow_globs
    )
    allowed = _matches(relative_path, policy.exclude_allow_globs)
    return not conditionally_excluded or allowed


def _manifest_digest(entries: Sequence[PackageEntry]) -> str:
    digest = hashlib.sha256()
    for entry in entries:
        canonical = json.dumps(
            [entry.relative_path, entry.size, entry.sha256],
            ensure_ascii=False,
            separators=(",", ":"),
        ).encode("utf-8")
        digest.update(canonical)
        digest.update(b"\n")
    return digest.hexdigest()


def _validate_required_paths(
    root: Path, entries: Sequence[PackageEntry], required_paths: Sequence[str]
) -> None:
    selected = {entry.relative_path for entry in entries}
    for relative_path in required_paths:
        candidate = root.joinpath(*relative_path.split("/"))
        _assert_no_link_components(root, relative_path)
        try:
            exists = candidate.exists()
            is_directory = candidate.is_dir() if exists else False
        except OSError as exc:
            raise PackageError(
                "cannot inspect required path %s (%s)"
                % (relative_path, type(exc).__name__)
            ) from exc
        if not exists:
            raise PackageError("required package path is missing: %s" % relative_path)
        if is_directory:
            prefix = relative_path.rstrip("/") + "/"
            if not any(path.startswith(prefix) for path in selected):
                raise PackageError(
                    "required package directory has no selected files: %s"
                    % relative_path
                )
        elif relative_path not in selected:
            raise PackageError(
                "required package file is excluded by policy: %s" % relative_path
            )


def _validate_provenance_gate(
    entries: Sequence[PackageEntry], gate: ProvenanceGate
) -> None:
    """Require attribution and immutable provenance only when copying upstream."""

    selected = {entry.relative_path for entry in entries}
    missing = sorted(set(gate.attribution_paths) - selected)
    if missing:
        raise PackageError(
            "provenance attribution path is not packaged: %s" % missing[0]
        )
    if gate.contains_substantial_copied_gsd_material:
        if gate.gsd_exact_revision.upper() == "UNKNOWN":
            raise PackageError(
                "substantial copied GSD material requires an exact upstream revision"
            )
        if not re.fullmatch(r"[0-9a-fA-F]{40,64}", gate.gsd_exact_revision):
            raise PackageError(
                "substantial copied GSD material requires an immutable commit SHA"
            )


def build_package_plan(root: Path, policy: PackagePolicy) -> PackagePlan:
    """Select and hash package payload files without writing."""

    resolved_root = _resolved_root(root)
    entries: List[PackageEntry] = []
    total_bytes = 0

    def onerror(exc: OSError) -> None:
        raise PackageError("cannot scan source tree (%s)" % type(exc).__name__)

    for directory, directory_names, file_names in os.walk(
        str(resolved_root),
        topdown=True,
        onerror=onerror,
        followlinks=False,
    ):
        directory_names.sort()
        file_names.sort()
        base = Path(directory)

        retained_directories: List[str] = []
        for name in directory_names:
            candidate = base / name
            relative = candidate.relative_to(resolved_root).as_posix()
            if _is_link_like(candidate):
                raise PackageError("source tree contains a link: %s" % relative)
            if not _matches(relative, policy.prune_globs):
                retained_directories.append(name)
        directory_names[:] = retained_directories

        for name in file_names:
            source_path = base / name
            relative_path = source_path.relative_to(resolved_root).as_posix()
            if _is_link_like(source_path):
                raise PackageError("source tree contains a link: %s" % relative_path)
            if not _selected(relative_path, policy):
                continue
            if len(entries) >= policy.file_policy.max_file_count:
                raise PackageError("selected package exceeds max file count")
            size, sha256 = _hash_source_file(
                resolved_root,
                source_path,
                relative_path,
                policy.file_policy,
            )
            total_bytes += size
            if total_bytes > policy.file_policy.max_total_bytes:
                raise PackageError("selected package exceeds max total bytes")
            entries.append(
                PackageEntry(
                    relative_path=relative_path,
                    source_path=source_path,
                    size=size,
                    sha256=sha256,
                )
            )

    entries.sort(key=lambda entry: entry.relative_path)
    _validate_required_paths(resolved_root, entries, policy.required_paths)
    _validate_provenance_gate(entries, policy.provenance_gate)
    if policy.inventory_path and any(
        entry.relative_path == policy.inventory_path for entry in entries
    ):
        raise PackageError("inventory path collides with a selected source file")
    return PackagePlan(
        root=resolved_root,
        entries=tuple(entries),
        manifest_sha256=_manifest_digest(entries),
        inventory_path=policy.inventory_path,
    )


def inventory_bytes(plan: PackagePlan) -> bytes:
    """Return deterministic UTF-8 inventory bytes for a package plan."""

    inventory = {
        "algorithm": "sha256",
        "file_count": len(plan.entries),
        "files": [
            {
                "path": entry.relative_path,
                "sha256": entry.sha256,
                "size": entry.size,
            }
            for entry in plan.entries
        ],
        "manifest_sha256": plan.manifest_sha256,
        "schema_version": 1,
    }
    rendered = json.dumps(
        inventory,
        ensure_ascii=False,
        indent=2,
        sort_keys=True,
    )
    return (rendered + "\n").encode("utf-8")


def _copy_verified(
    entry: PackageEntry,
    destination: Path,
    source_root: Path,
    output_root: Path,
) -> None:
    """Copy through verified descriptors and recheck the selected fingerprint."""

    digest = hashlib.sha256()
    total = 0
    try:
        enforce_containment(source_root, entry.source_path, must_exist=True)
        with safe_open_read(source_root, entry.source_path) as source:
            with safe_open_write(output_root, destination) as target:
                while True:
                    chunk = source.read(BUFFER_SIZE)
                    if not chunk:
                        break
                    target.write(chunk)
                    digest.update(chunk)
                    total += len(chunk)
    except (OSError, FileSafetyError) as exc:
        raise PackageError(
            "cannot copy selected file %s (%s)"
            % (entry.relative_path, type(exc).__name__)
        ) from exc
    if total != entry.size or digest.hexdigest() != entry.sha256:
        raise PackageError("source file changed during packaging: %s" % entry.relative_path)


def _staging_prefix(output: Path) -> str:
    return ".%s.staging-" % output.name


def _create_staging(output: Path) -> Path:
    """Create a task-owned sibling directory on the destination filesystem."""

    try:
        parent = resolve_directory_root(output.parent, "output parent")
        created = Path(
            tempfile.mkdtemp(prefix=_staging_prefix(output), dir=os.fspath(parent))
        )
        staging = resolve_directory_root(created, "package staging")
    except (OSError, FileSafetyError) as exc:
        raise PackageError(
            "cannot create package staging safely (%s)" % type(exc).__name__
        ) from exc
    if staging.parent != parent:
        raise PackageError("package staging is not a sibling of output")
    return staging


def _cleanup_staging(staging: Path, output: Path) -> None:
    """Remove only the exact task-owned sibling staging directory."""

    prefix = _staging_prefix(output)
    try:
        parent = resolve_directory_root(output.parent, "output parent")
        if staging.parent != parent or not staging.name.startswith(prefix):
            raise PackageError("refusing to clean an unexpected staging path")
        if is_link_like(staging):
            raise PackageError("refusing to clean link-like package staging")
        shutil.rmtree(staging)
    except (OSError, FileSafetyError) as exc:
        raise PackageError(
            "cannot clean package staging safely (%s)" % type(exc).__name__
        ) from exc


def _verify_staged_package(
    staging: Path, plan: PackagePlan, policy: PackagePolicy
) -> None:
    """Verify the complete staged tree before it becomes externally visible."""

    inventory = inventory_bytes(plan) if plan.inventory_path else b""
    extra_count = 1 if plan.inventory_path else 0
    verification_policy = TextPolicy(
        allowed_text_extensions=policy.file_policy.allowed_text_extensions,
        allowed_text_names=policy.file_policy.allowed_text_names,
        max_file_bytes=max(policy.file_policy.max_file_bytes, len(inventory) or 1),
        max_total_bytes=policy.file_policy.max_total_bytes + len(inventory),
        max_file_count=policy.file_policy.max_file_count + extra_count,
    )
    try:
        tree = SafeTree(staging, verification_policy)
        expected = {entry.relative_path for entry in plan.entries}
        if plan.inventory_path:
            expected.add(plan.inventory_path)
        actual = {tree.relative_path(path) for path in tree.files()}
        if actual != expected:
            raise PackageError("staged package file set does not match the plan")
        expected_directories = {""}
        for relative_path in expected:
            parts = relative_path.split("/")[:-1]
            for index in range(1, len(parts) + 1):
                expected_directories.add("/".join(parts[:index]))
        actual_directories = {
            tree.relative_path(path) for path in tree.directories()
        }
        if actual_directories != expected_directories:
            raise PackageError("staged package directory set does not match the plan")
        for entry in plan.entries:
            inspected = tree.inspect_text(entry.relative_path)
            if inspected.size != entry.size or inspected.sha256 != entry.sha256:
                raise PackageError(
                    "staged package fingerprint mismatch: %s" % entry.relative_path
                )
        if plan.inventory_path:
            inspected_inventory = tree.inspect_text(plan.inventory_path)
            if inspected_inventory.data != inventory:
                raise PackageError("staged package inventory mismatch")
    except FileSafetyError as exc:
        raise PackageError(
            "cannot verify package staging safely (%s)" % type(exc).__name__
        ) from exc


def _publish_staging(source_root: Path, staging: Path, output: Path) -> None:
    """Atomically rename a verified sibling staging tree into place."""

    _, rechecked_output = validate_request(source_root, output)
    if rechecked_output != output:
        raise PackageError("output identity changed before publish")
    try:
        os.replace(staging, output)
    except OSError as exc:
        raise PackageError(
            "cannot atomically publish package output (%s)" % type(exc).__name__
        ) from exc


def package_directory(
    root: Path,
    output: Path,
    policy: PackagePolicy,
    *,
    dry_run: bool = False,
) -> PackagePlan:
    """Validate, select, and optionally write a portable package."""

    resolved_root, resolved_output = validate_request(root, output)
    plan = build_package_plan(resolved_root, policy)
    if dry_run:
        return plan

    staging: Optional[Path] = None
    try:
        staging = _create_staging(resolved_output)
        output_root = resolve_directory_root(staging, "package staging")
        for entry in plan.entries:
            destination = output_root.joinpath(*entry.relative_path.split("/"))
            _copy_verified(entry, destination, plan.root, output_root)
        if plan.inventory_path:
            inventory = output_root.joinpath(*plan.inventory_path.split("/"))
            with safe_open_write(output_root, inventory) as target:
                target.write(inventory_bytes(plan))
        _verify_staged_package(output_root, plan, policy)
        _publish_staging(plan.root, output_root, resolved_output)
        staging = None
    except (OSError, FileSafetyError, PackageError) as exc:
        cleanup_error: Optional[PackageError] = None
        if staging is not None and os.path.lexists(staging):
            try:
                _cleanup_staging(staging, resolved_output)
            except PackageError as cleanup_exc:
                cleanup_error = cleanup_exc
        if cleanup_error is not None:
            raise PackageError(
                "packaging failed and staging cleanup also failed (%s)"
                % type(cleanup_error).__name__
            ) from exc
        if isinstance(exc, PackageError):
            raise
        raise PackageError(
            "cannot write package output safely (%s)" % type(exc).__name__
        ) from exc
    return plan


def render_summary(
    plan: PackagePlan, *, dry_run: bool, verbose: bool = False
) -> str:
    lines = [
        "Mode: %s" % ("dry-run" if dry_run else "package"),
        "File count: %d" % len(plan.entries),
        "Manifest SHA-256: %s" % plan.manifest_sha256,
    ]
    if verbose:
        lines.append("Files:")
        for entry in plan.entries:
            encoded_path = json.dumps(entry.relative_path, ensure_ascii=False)
            lines.append("%s  %s" % (entry.sha256, encoded_path))
    if plan.inventory_path:
        suffix = " (not written)" if dry_run else ""
        lines.append("Inventory: %s%s" % (plan.inventory_path, suffix))
    return "\n".join(lines)


def main(
    argv: Optional[Sequence[str]] = None,
    *,
    config_path: Path = DEFAULT_CONFIG_PATH,
) -> int:
    parser = argparse.ArgumentParser(
        description="Build a project-neutral portable test-automation AI package."
    )
    parser.add_argument(
        "--root",
        required=True,
        type=Path,
        help="source pro-qa-ai root",
    )
    parser.add_argument(
        "--output",
        required=True,
        type=Path,
        help="destination directory outside the source root",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="validate, select, and hash without writing",
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="print every selected path and SHA-256; default output is compact",
    )
    args = parser.parse_args(argv)

    try:
        policy = load_policy(config_path)
        plan = package_directory(
            args.root,
            args.output,
            policy,
            dry_run=args.dry_run,
        )
    except PackageError as exc:
        print("ERROR: %s" % exc, file=sys.stderr)
        return 2
    except Exception:
        print(
            "ERROR: packaging failed safely; no file contents were printed",
            file=sys.stderr,
        )
        return 2

    print(render_summary(plan, dry_run=args.dry_run, verbose=args.verbose))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
