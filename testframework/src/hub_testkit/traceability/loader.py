from __future__ import annotations

import json
import os
import stat
import tomllib
from collections.abc import Mapping
from pathlib import Path

from jsonschema import Draft202012Validator
from jsonschema.exceptions import SchemaError, ValidationError as JsonSchemaValidationError
from pydantic import ValidationError

from .model import (
    CASE_ID_PATTERN,
    Automation,
    CaseMetadata,
    CaseRecord,
    LocustAutomation,
    ManualAutomation,
    MarkerPolicy,
    PytestAutomation,
)


class TraceabilityError(ValueError):
    pass


_CASE_LIMIT_BYTES = 256 * 1024
_BODY_LIMIT_BYTES = 64 * 1024
_POLICY_LIMIT_BYTES = 32 * 1024
_SCHEMA_LIMIT_BYTES = 256 * 1024
_CASE_FIELDS = frozenset(
    {
        "schema",
        "id",
        "title",
        "status",
        "task",
        "requirement",
        "risk",
        "scenario",
        "subject",
        "preconditions",
        "oracle",
        "cleanup",
        "automation",
    }
)
_POLICY_FIELDS = frozenset(
    {"schema", "case_id_pattern", "statuses", "scenarios", "pytest_layers", "automation_kinds"}
)
_PYTEST_FIELDS = frozenset({"kind", "layer", "node_id"})
_LOCUST_FIELDS = frozenset({"kind", "entrypoint"})
_MANUAL_FIELDS = frozenset({"kind", "evidence_owner"})
_EXPECTED_STATUSES = ("active", "planned", "retired")
_EXPECTED_SCENARIOS = (
    "happy",
    "negative",
    "authn",
    "authz",
    "boundary",
    "edge",
    "corner",
    "failure-recovery",
    "security",
    "accessibility",
    "performance",
)
_EXPECTED_PYTEST_LAYERS = ("contract", "api", "integration", "e2e")
_EXPECTED_AUTOMATION_KINDS = ("pytest", "locust", "manual")


def load_marker_policy(path: str | Path) -> MarkerPolicy:
    document = _load_toml_document(path, _POLICY_LIMIT_BYTES, "marker policy")
    try:
        _require_exact_keys(document, _POLICY_FIELDS)
        schema = _require_string(document["schema"])
        case_id_pattern = _require_string(document["case_id_pattern"])
        statuses = _require_string_tuple(document["statuses"])
        scenarios = _require_string_tuple(document["scenarios"])
        pytest_layers = _require_string_tuple(document["pytest_layers"])
        automation_kinds = _require_string_tuple(document["automation_kinds"])
        if schema != "hub.testframework.markers/v1" or case_id_pattern != CASE_ID_PATTERN:
            raise TraceabilityError("marker policy is invalid")
        if statuses != _EXPECTED_STATUSES or scenarios != _EXPECTED_SCENARIOS:
            raise TraceabilityError("marker policy is invalid")
        if (
            pytest_layers != _EXPECTED_PYTEST_LAYERS
            or automation_kinds != _EXPECTED_AUTOMATION_KINDS
        ):
            raise TraceabilityError("marker policy is invalid")
        return MarkerPolicy(
            schema=schema,
            case_id_pattern=case_id_pattern,
            statuses=statuses,
            scenarios=scenarios,
            pytest_layers=pytest_layers,
            automation_kinds=automation_kinds,
        )
    except (KeyError, ValidationError):
        raise TraceabilityError("marker policy is invalid") from None


def load_case(
    path: str | Path,
    schema_path: str | Path,
    marker_policy: MarkerPolicy,
) -> CaseRecord:
    if not isinstance(marker_policy, MarkerPolicy):
        raise TraceabilityError("marker policy is invalid")
    source_path = _as_path(path, "case")
    if source_path.suffix != ".md":
        raise TraceabilityError("case path is invalid")
    text = _decode_case(_read_regular_file(source_path, _CASE_LIMIT_BYTES, "case"))
    metadata_document, markdown_body = _split_case_document(text)
    _validate_against_schema(metadata_document, schema_path)
    metadata = _load_metadata(metadata_document)
    _validate_metadata_against_policy(metadata, marker_policy)
    if source_path.stem != metadata.id:
        raise TraceabilityError("case filename is invalid")
    try:
        return CaseRecord(
            metadata=metadata,
            source_path=source_path,
            markdown_body=markdown_body,
        )
    except ValidationError:
        raise TraceabilityError("case record is invalid") from None


def _load_metadata(document: Mapping[str, object]) -> CaseMetadata:
    try:
        _require_exact_keys(document, _CASE_FIELDS)
        normalized = {
            "schema": _require_string(document["schema"]),
            "id": _require_string(document["id"]),
            "title": _require_string(document["title"]),
            "status": _require_string(document["status"]),
            "task": _require_string(document["task"]),
            "requirement": _require_string(document["requirement"]),
            "risk": _require_string(document["risk"]),
            "scenario": _require_string(document["scenario"]),
            "subject": _require_string(document["subject"]),
            "preconditions": _require_string_tuple(document["preconditions"]),
            "oracle": _require_string(document["oracle"]),
            "cleanup": _require_string(document["cleanup"]),
            "automation": _normalize_automation(document["automation"]),
        }
        return CaseMetadata.model_validate(normalized)
    except (KeyError, ValidationError):
        raise TraceabilityError("case metadata is invalid") from None


def _normalize_automation(value: object) -> dict[str, str]:
    if not isinstance(value, Mapping):
        raise TraceabilityError("case automation is invalid")
    try:
        kind = _require_string(value["kind"])
    except KeyError:
        raise TraceabilityError("case automation is invalid") from None
    expected_fields = {
        "pytest": _PYTEST_FIELDS,
        "locust": _LOCUST_FIELDS,
        "manual": _MANUAL_FIELDS,
    }.get(kind)
    if expected_fields is None:
        raise TraceabilityError("case automation is invalid")
    _require_exact_keys(value, expected_fields)
    try:
        return {field: _require_string(value[field]) for field in expected_fields}
    except KeyError:
        raise TraceabilityError("case automation is invalid") from None


def _validate_metadata_against_policy(metadata: CaseMetadata, policy: MarkerPolicy) -> None:
    if metadata.status not in policy.statuses or metadata.scenario not in policy.scenarios:
        raise TraceabilityError("case metadata is not admitted")
    automation: Automation = metadata.automation
    if isinstance(automation, PytestAutomation):
        if "pytest" not in policy.automation_kinds or automation.layer not in policy.pytest_layers:
            raise TraceabilityError("case automation is not admitted")
    elif isinstance(automation, LocustAutomation):
        if "locust" not in policy.automation_kinds:
            raise TraceabilityError("case automation is not admitted")
    elif isinstance(automation, ManualAutomation):
        if "manual" not in policy.automation_kinds:
            raise TraceabilityError("case automation is not admitted")
    else:
        raise TraceabilityError("case automation is not admitted")


def _split_case_document(text: str) -> tuple[dict[str, object], str]:
    if not text.startswith("+++\n"):
        raise TraceabilityError("case frontmatter is invalid")
    closing = text.find("\n+++\n", len("+++\n"))
    if closing < 0:
        raise TraceabilityError("case frontmatter is invalid")
    frontmatter = text[len("+++\n") : closing]
    markdown_body = text[closing + len("\n+++\n") :]
    if "+++\n" in markdown_body or not markdown_body.strip():
        raise TraceabilityError("case frontmatter is invalid")
    if len(markdown_body.encode("utf-8")) > _BODY_LIMIT_BYTES:
        raise TraceabilityError("case Markdown body is invalid")
    try:
        document = tomllib.loads(frontmatter)
    except tomllib.TOMLDecodeError:
        raise TraceabilityError("case frontmatter is invalid") from None
    if not isinstance(document, dict):
        raise TraceabilityError("case frontmatter is invalid")
    return document, markdown_body


def _validate_against_schema(document: Mapping[str, object], schema_path: str | Path) -> None:
    schema_bytes = _read_regular_file(
        _as_path(schema_path, "case schema"),
        _SCHEMA_LIMIT_BYTES,
        "case schema",
    )
    try:
        schema = json.loads(_decode_utf8(schema_bytes, "case schema"))
        if not isinstance(schema, dict) or _contains_reference(schema):
            raise TraceabilityError("case schema is invalid")
        Draft202012Validator.check_schema(schema)
        validator = Draft202012Validator(schema)
        if next(validator.iter_errors(document), None) is not None:
            raise TraceabilityError("case metadata is invalid")
    except (json.JSONDecodeError, SchemaError, JsonSchemaValidationError):
        raise TraceabilityError("case schema is invalid") from None


def _contains_reference(value: object) -> bool:
    if isinstance(value, Mapping):
        return any(
            key in {"$ref", "$dynamicRef", "$recursiveRef"} or _contains_reference(item)
            for key, item in value.items()
        )
    if isinstance(value, list):
        return any(_contains_reference(item) for item in value)
    return False


def _load_toml_document(path: str | Path, max_bytes: int, label: str) -> dict[str, object]:
    data = _read_regular_file(_as_path(path, label), max_bytes, label)
    try:
        document = tomllib.loads(_decode_utf8(data, label))
    except tomllib.TOMLDecodeError:
        raise TraceabilityError(f"{label} is invalid") from None
    if not isinstance(document, dict):
        raise TraceabilityError(f"{label} is invalid")
    return document


def _decode_case(data: bytes) -> str:
    return _decode_utf8(data, "case")


def _as_path(value: str | Path, label: str) -> Path:
    try:
        return Path(value)
    except (TypeError, ValueError):
        raise TraceabilityError(f"{label} path is invalid") from None


def _decode_utf8(data: bytes, label: str) -> str:
    if data.startswith(b"\xef\xbb\xbf") or b"\x00" in data or b"\r" in data:
        raise TraceabilityError(f"{label} contains unsupported bytes")
    try:
        return data.decode("utf-8", errors="strict")
    except UnicodeDecodeError:
        raise TraceabilityError(f"{label} is not UTF-8") from None


def _read_regular_file(path: Path, max_bytes: int, label: str) -> bytes:
    try:
        before = os.lstat(path)
    except (OSError, ValueError):
        raise TraceabilityError(f"{label} could not be read") from None
    if stat.S_ISLNK(before.st_mode) or not stat.S_ISREG(before.st_mode):
        raise TraceabilityError(f"{label} is not a regular file")
    flags = os.O_RDONLY | getattr(os, "O_BINARY", 0) | getattr(os, "O_NOFOLLOW", 0)
    try:
        descriptor = os.open(path, flags)
    except (OSError, ValueError):
        raise TraceabilityError(f"{label} could not be read") from None
    try:
        opened = os.fstat(descriptor)
        if not _same_identity(before, opened) or not stat.S_ISREG(opened.st_mode):
            raise TraceabilityError(f"{label} changed while opening")
        data = _read_at_most(descriptor, max_bytes)
        after = os.fstat(descriptor)
        path_after = os.lstat(path)
        if not _same_identity(opened, after) or not _same_identity(opened, path_after):
            raise TraceabilityError(f"{label} changed while reading")
        return data
    except (OSError, ValueError):
        raise TraceabilityError(f"{label} could not be read") from None
    finally:
        try:
            os.close(descriptor)
        except OSError:
            pass


def _read_at_most(descriptor: int, max_bytes: int) -> bytes:
    data = bytearray()
    limit = max_bytes + 1
    while len(data) < limit:
        chunk = os.read(descriptor, min(64 * 1024, limit - len(data)))
        if not chunk:
            break
        data.extend(chunk)
    if len(data) > max_bytes:
        raise TraceabilityError("file exceeds its byte limit")
    return bytes(data)


def _same_identity(left: os.stat_result, right: os.stat_result) -> bool:
    return (
        left.st_dev == right.st_dev
        and left.st_ino == right.st_ino
        and left.st_mode == right.st_mode
    )


def _require_exact_keys(value: Mapping[str, object], expected: frozenset[str]) -> None:
    if set(value) != expected:
        raise TraceabilityError("mapping has unsupported keys")


def _require_string(value: object) -> str:
    if type(value) is not str or not value:
        raise TraceabilityError("value must be a non-empty string")
    return value


def _require_string_tuple(value: object) -> tuple[str, ...]:
    if type(value) is not list:
        raise TraceabilityError("value must be a list")
    items = tuple(_require_string(item) for item in value)
    if not items or len(set(items)) != len(items):
        raise TraceabilityError("list must be non-empty and unique")
    return items
