"""Pure semantic validation for the backend-unit evidence packet.

The Draft 2020-12 schemas close the document shapes.  This module closes the
relations which cannot be represented by those schemas, without reading a
file, starting a process, or calling Docker.
"""

from __future__ import annotations

import hashlib
import re
from collections.abc import Mapping, Sequence
from functools import wraps
from typing import Callable, Final, ParamSpec, TypeVar


class ReceiptContractError(ValueError):
    """Raised when a receipt is structurally plausible but semantically unsafe."""


__all__ = (
    "ReceiptContractError",
    "parse_seal",
    "validate_package_plan",
    "validate_container_receipt",
    "validate_observer_receipt",
    "validate_final_receipt",
)


_P = ParamSpec("_P")
_T = TypeVar("_T")


def _contract_boundary(function: Callable[_P, _T]) -> Callable[_P, _T]:
    """Fail closed if hostile mapping behavior escapes semantic helpers."""

    @wraps(function)
    def wrapped(*args: _P.args, **kwargs: _P.kwargs) -> _T:
        try:
            return function(*args, **kwargs)
        except ReceiptContractError:
            raise
        except Exception as error:
            raise ReceiptContractError("receipt contract evaluation failed") from error

    return wrapped


_SHA256_RE: Final = re.compile(r"sha256:[0-9a-f]{64}\Z")
_RUN_ID_RE: Final = re.compile(
    r"backend-unit-[0-9]{8}T[0-9]{6}Z-[0-9a-f]{12,64}-[0-9a-f]{8}\Z"
)
_NONCE_RE: Final = re.compile(r"[0-9a-f]{32}\Z")
_ENGINE_ID_RE: Final = re.compile(r"[0-9a-f]{64}\Z")
_HOSTNAME_RE: Final = re.compile(r"[0-9a-f]{12,64}\Z")
_GIT_SHA_RE: Final = re.compile(r"[0-9a-f]{40}\Z")

_DISCOVERY_FORMAT: Final = '{{printf "%s\\t%s" .ImportPath .Name}}'
_DISCOVERY_ARGV: Final = {
    "build": ("go", "list", "-f", _DISCOVERY_FORMAT, "./..."),
    "vet": ("go", "list", "-tags", "hubtest", "-f", _DISCOVERY_FORMAT, "./..."),
    "unit": (
        "go",
        "list",
        "-tags",
        "hubtest",
        "-f",
        _DISCOVERY_FORMAT,
        "./tests/...",
    ),
    "race": (
        "go",
        "list",
        "-tags",
        "hubtest",
        "-f",
        _DISCOVERY_FORMAT,
        "./tests/...",
    ),
}

_SEALED_CLASSIFICATIONS: Final = frozenset(
    {
        "SEALED_EXECUTABLE_PASS",
        "SEALED_EXECUTABLE_FAIL_DISCOVERY",
        "SEALED_EXECUTABLE_FAIL_STAGE",
        "SEALED_TIMEOUT_UNTRIAGED",
        "SEALED_INFRA_ABORTED_OOM",
        "SEALED_INFRA_ABORTED_RESOURCE_CONTRACT",
        "SEALED_INFRA_ABORTED_REPORT_BUDGET",
        "SEALED_INFRA_ABORTED_TMPFS",
        "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR",
        "SEALED_INFRA_ABORTED_DISK_PRESSURE",
        "SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT",
        "SEALED_INTERRUPTED",
        "SEALED_INVALID_QA_OVERLAP",
        "SEALED_INVALID_SOURCE_MUTATED",
        "SEALED_INVALID_PROVENANCE",
    }
)
_STOP_TIE_ORDER: Final = ("interrupt", "qa", "disk", "ram", "deadline")
_STOP_CLASSIFICATIONS: Final = {
    "interrupt": "SEALED_INTERRUPTED",
    "qa": "SEALED_INVALID_QA_OVERLAP",
    "disk": "SEALED_INFRA_ABORTED_DISK_PRESSURE",
    "ram": "SEALED_INFRA_ABORTED_HOST_RAM_FLOOR",
    "deadline": "SEALED_INFRA_ABORTED_WHOLE_RUN_TIMEOUT",
}
_STOP_PREDICATES: Final = {
    "interrupt": "interrupt",
    "qa": "qa_overlap",
    "disk": "disk_pressure",
    "ram": "host_ram_floor",
    "deadline": "whole_run_timeout",
}

_FROZEN_SUBJECTS: Final = {
    "backend/internal/calls/call_durable_credential_allocation.go": (
        "sha256:be6ff058468c4e288e0b0d7060fd17cf3118229d009f28dc056d89f281e99d4d"
    ),
    "backend/tests/calls/call_durable_credential_allocation_test.go": (
        "sha256:0ae525c86a75369bdb5790f26b1308f42574fb545883328da24011d4150b03a9"
    ),
}
_STREAM_CAPS: Final = {
    "build": (2097152, 2097152),
    "vet": (2097152, 2097152),
    "unit": (16777216, 4194304),
    "race": (25165824, 4194304),
}
_FIXED_HOST_CONFIG: Final = {
    "memory": 2147483648,
    "memory_swap": 2147483648,
    "nano_cpus": 6000000000,
    "read_only_rootfs": True,
    "pids_limit": 256,
    "init": True,
    "network_mode": "worktree-master_qa-internal",
}
_FIXED_HOST_MOUNTS: Final = (
    ("bind", "./backend/cmd", "/opt/hub/source/backend/cmd", True, None),
    ("bind", "./backend/internal", "/opt/hub/source/backend/internal", True, None),
    ("bind", "./backend/tests", "/opt/hub/source/backend/tests", True, None),
    ("bind", "./backend/ent/schema", "/opt/hub/source/backend/ent/schema", True, None),
    ("bind", "./backend/migrations", "/opt/hub/source/backend/migrations", True, None),
    ("bind", "./backend/testframework", "/opt/hub/source/backend/testframework", True, None),
    ("bind", "./testframework/reports", "/reports", False, None),
    ("tmpfs", None, "/tmp", False, 67108864),
    ("tmpfs", None, "/var/cache/hub-go", False, 805306368),
    ("tmpfs", None, "/var/tmp/hub-go", False, 402653184),
)

# This is the frozen policy order after wildcard expansion.  It is a protocol
# input, not a filesystem walk: a pure validator must reject a substituted
# path even if its self-consistent digest was recomputed.
_CANONICAL_BUILD_INPUT_PATHS: Final = (
    "docker-compose.local.yml",
    "testframework/docker/Dockerfile.qa",
    "testframework/docker/entrypoints/backend-unit.sh",
    "testframework/docker/backend_unit/runtime.sh",
    "testframework/docker/backend_unit/plan.sh",
    "testframework/docker/backend_unit/execute.sh",
    "testframework/docker/backend_unit/receipt.sh",
    "testframework/observers/__init__.py",
    "testframework/observers/backend_unit/__init__.py",
    "testframework/observers/backend_unit/contracts.py",
    "testframework/observers/backend_unit/safe_paths.py",
    "testframework/observers/backend_unit/lease.py",
    "testframework/observers/backend_unit/resources.py",
    "testframework/observers/backend_unit/docker_adapter.py",
    "testframework/observers/backend_unit/publication.py",
    "testframework/observers/backend_unit/observe.py",
    "testframework/schemas/backend-unit-package-plan.schema.v1.json",
    "testframework/schemas/backend-unit-container-receipt.schema.v1.json",
    "testframework/schemas/backend-unit-observer-receipt.schema.v1.json",
    "testframework/policies/backend-unit-observer.v1.json",
    "testframework/policies/backend-unit-phase150-08c2-subject.v1.json",
    "testframework/policies/root-compose-testframework.v1.json",
    "testframework/selftests/architecture/docker/test_root_compose_contract.py",
    "testframework/selftests/architecture/observers/test_backend_unit_receipts.py",
    "testframework/selftests/architecture/observers/test_backend_unit_receipts_review_contract.py",
    "testframework/selftests/architecture/observers/test_backend_unit_observer.py",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/valid/container-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/valid/final-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/valid/observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/valid/package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/valid/SEAL",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/bound-backend-image-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/build-input-hash-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/build-input-manifest-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/count-package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/discovery-argv-package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/duplicate-import-path-package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/engine-id-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/fixed-image-id-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/foreign-run-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/inventory-container-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/inventory-hash-container-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/non-bijection-package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/nonce-observer-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/row-set-hash-package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/SEAL.extra-field",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/SEAL.missing-field",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/SEAL.reordered",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/timeout-outcome-final-receipt.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/unknown-key-package-plan.json",
    "testframework/selftests/architecture/observers/fixtures/backend_unit/invalid/uppercase-digest-package-plan.json",
)


@_contract_boundary
def parse_seal(payload: bytes) -> dict[str, str]:
    """Parse only the exact four-line, UTF-8/LF detached seal representation."""

    if type(payload) is not bytes:
        _fail("SEAL must be bytes")
    if b"\r" in payload or not payload.endswith(b"\n") or payload.count(b"\n") != 4:
        _fail("SEAL must contain exactly four LF-finalized lines")

    lines = payload.split(b"\n")
    if lines != [
        b"hub.qa.backend-unit-seal/v1",
        lines[1],
        lines[2],
        lines[3],
        b"",
    ]:
        _fail("SEAL header or line count is invalid")

    fields: dict[str, str] = {}
    for expected_key, line in zip(
        ("run_id", "owner_nonce", "observer_receipt_sha256"),
        lines[1:4],
        strict=True,
    ):
        key, separator, raw_value = line.partition(b"=")
        if separator != b"=" or key.decode("ascii", "ignore") != expected_key:
            _fail("SEAL fields must be present once and in canonical order")
        try:
            fields[expected_key] = raw_value.decode("ascii")
        except UnicodeDecodeError:
            _fail("SEAL fields must be ASCII")

    _require_run_id(fields["run_id"], "SEAL.run_id")
    _require_nonce(fields["owner_nonce"], "SEAL.owner_nonce")
    _require_sha256(fields["observer_receipt_sha256"], "SEAL.observer_receipt_sha256")
    return fields


@_contract_boundary
def validate_package_plan(document: Mapping[str, object]) -> None:
    """Validate discovery, deterministic plan rows, and result partition semantics."""

    document = _require_mapping(document, "package plan")
    _require_exact(document, "schema", "hub.qa.backend-unit-package-plan/v1")
    _require_exact(document, "target", "backend-unit")
    _require_run_id(_require_string(_field(document, "run_id"), "run_id"), "run_id")
    _require_nonce(_require_string(_field(document, "owner_nonce"), "owner_nonce"), "owner_nonce")

    stage = _require_string(_field(document, "stage"), "stage")
    expected_argv = _DISCOVERY_ARGV.get(stage)
    if expected_argv is None:
        _fail("package plan stage is not recognized")
    if _string_tuple(_field(document, "discovery_argv"), "discovery_argv") != expected_argv:
        _fail("discovery_argv does not match the frozen stage command")
    _require_sha256(
        _field(document, "go_environment_sha256"), "go_environment_sha256"
    )

    rows = _require_sequence(_field(document, "rows"), "rows")
    if not rows:
        _fail("package plan rows must be nonempty")

    previous_path: str | None = None
    row_lines: list[bytes] = []
    row_paths: list[str] = []
    for expected_ordinal, raw_row in enumerate(rows, start=1):
        row = _require_mapping(raw_row, f"rows[{expected_ordinal - 1}]")
        ordinal = _require_integer(_field(row, "ordinal"), f"rows[{expected_ordinal - 1}].ordinal")
        import_path = _require_string(
            _field(row, "import_path"), f"rows[{expected_ordinal - 1}].import_path"
        )
        package_name = _require_string(
            _field(row, "package_name"), f"rows[{expected_ordinal - 1}].package_name"
        )
        if ordinal != expected_ordinal:
            _fail("package plan row ordinals must be contiguous and one-based")
        if not import_path or not package_name or _contains_record_separator(import_path) or _contains_record_separator(package_name):
            _fail("package plan rows must contain nonempty tab- and LF-free strings")
        if previous_path is not None and import_path <= previous_path:
            _fail("package plan import paths must be strictly lexical and unique")
        previous_path = import_path
        row_paths.append(import_path)
        row_lines.append(_utf8(import_path, "import_path") + b"\t" + _utf8(package_name, "package_name") + b"\n")

    if _require_integer(_field(document, "row_count"), "row_count") != len(rows):
        _fail("row_count does not equal the ordered row count")
    expected_row_set_sha256 = _sha256(b"".join(row_lines))
    if _require_sha256(_field(document, "row_set_sha256"), "row_set_sha256") != expected_row_set_sha256:
        _fail("row_set_sha256 does not bind canonical ordered rows")

    partition = _require_sequence(_field(document, "result_partition"), "result_partition")
    if len(partition) != len(rows):
        _fail("result_partition must contain one result for every plan row")
    for expected_ordinal, (raw_result, import_path) in enumerate(
        zip(partition, row_paths, strict=True), start=1
    ):
        result = _require_mapping(raw_result, f"result_partition[{expected_ordinal - 1}]")
        if _require_integer(_field(result, "ordinal"), f"result_partition[{expected_ordinal - 1}].ordinal") != expected_ordinal:
            _fail("result_partition ordinals must exactly match plan order")
        if _require_string(_field(result, "import_path"), f"result_partition[{expected_ordinal - 1}].import_path") != import_path:
            _fail("result_partition import paths must exactly match plan rows")
        if not _require_string(_field(result, "result"), f"result_partition[{expected_ordinal - 1}].result"):
            _fail("result_partition result must be nonempty")


@_contract_boundary
def validate_container_receipt(document: Mapping[str, object]) -> None:
    """Validate the container-owned artifact inventory and immutable identity fields."""

    document = _require_mapping(document, "container receipt")
    _require_exact(document, "schema", "hub.qa.backend-unit-container-receipt/v1")
    _require_exact(document, "target", "backend-unit")
    _require_run_id(_require_string(_field(document, "run_id"), "run_id"), "run_id")
    _require_nonce(_require_string(_field(document, "owner_nonce"), "owner_nonce"), "owner_nonce")
    _require_hostname(_require_string(_field(document, "container_hostname"), "container_hostname"), "container_hostname")
    _validate_artifact_inventory(document, "container receipt")
    _validate_report_budget(document)


@_contract_boundary
def validate_observer_receipt(document: Mapping[str, object]) -> None:
    """Validate the observer's source, image, identity, inventory, and final bindings."""

    document = _require_mapping(document, "observer receipt")
    _require_exact(document, "schema", "hub.qa.backend-unit-observer-receipt/v1")
    _require_exact(document, "target", "backend-unit")

    run_id = _require_string(_field(document, "run_id"), "run_id")
    nonce = _require_string(_field(document, "owner_nonce"), "owner_nonce")
    container_id = _require_string(_field(document, "container_id"), "container_id")
    hostname = _require_string(_field(document, "container_hostname"), "container_hostname")
    _require_run_id(run_id, "run_id")
    _require_nonce(nonce, "owner_nonce")
    _require_engine_id(container_id, "container_id")
    _require_hostname(hostname, "container_hostname")
    if not container_id.startswith(hostname):
        _fail("container_hostname must be the full engine ID prefix")
    if not run_id.endswith(f"-{hostname}-{nonce[:8]}"):
        _fail("run_id must bind the owner hostname and nonce prefix")

    container_image = _require_sha256(_field(document, "container_image"), "container_image")
    bound_backend_image_id = _require_sha256(
        _field(document, "bound_backend_image_id"), "bound_backend_image_id"
    )
    fixed_image_ids = _require_mapping(_field(document, "fixed_image_ids"), "fixed_image_ids")
    _require_sha256(_field(fixed_image_ids, "python_contract"), "fixed_image_ids.python_contract")
    fixed_backend_image_id = _require_sha256(
        _field(fixed_image_ids, "backend_unit"), "fixed_image_ids.backend_unit"
    )
    if container_image != fixed_backend_image_id or bound_backend_image_id != fixed_backend_image_id:
        _fail("container image and bound backend image must equal the fixed backend image ID")

    manifest_sha256, manifest_digests = _validate_canonical_build_input_manifest(document)
    if _require_sha256(_field(document, "compose_sha256"), "compose_sha256") != manifest_digests[
        "docker-compose.local.yml"
    ]:
        _fail("compose_sha256 must bind the docker-compose.local.yml manifest row")
    _validate_fixed_host_config(document)
    facts = _validate_classification_facts(document, container_id)
    _validate_source_snapshots(document, manifest_sha256, fixed_image_ids, facts)

    records = _validate_artifact_inventory(document, "observer receipt")
    paths = tuple(_require_string(_field(record, "path"), "artifact_inventory.path") for record in records)
    if paths != ("container-receipt.json", "final-receipt.json"):
        _fail("observer artifact inventory must contain only canonical payload artifacts")
    if _require_sha256(_field(records[0], "sha256"), "container-receipt artifact sha256") != _require_sha256(
        _field(document, "container_receipt_sha256"), "container_receipt_sha256"
    ):
        _fail("container-receipt artifact hash does not bind container_receipt_sha256")

    final_receipt = _require_mapping(_field(document, "final_receipt"), "final_receipt")
    validate_final_receipt(final_receipt)
    _require_final_bindings(
        final_receipt,
        run_id=run_id,
        nonce=nonce,
        container_id=container_id,
        container_receipt_sha256=_require_sha256(
            _field(document, "container_receipt_sha256"), "container_receipt_sha256"
        ),
        backend_image_id=fixed_backend_image_id,
        manifest_sha256=manifest_sha256,
        classification=_require_string(_field(document, "classification"), "classification"),
        immutable_outcome=_field(document, "immutable_outcome"),
        seal_timestamp=_field(document, "seal_timestamp"),
    )


@_contract_boundary
def validate_final_receipt(document: Mapping[str, object]) -> None:
    """Validate final-receipt fields independent of an observer receipt wrapper."""

    document = _require_mapping(document, "final receipt")
    if "observer_receipt_sha256" in document:
        _fail("final receipt must not contain an observer receipt hash")
    _require_exact(document, "schema", "hub.qa.backend-unit-final-receipt/v1")
    _require_exact(document, "target", "backend-unit")
    _require_run_id(_require_string(_field(document, "run_id"), "run_id"), "run_id")
    _require_nonce(_require_string(_field(document, "owner_nonce"), "owner_nonce"), "owner_nonce")
    _require_engine_id(_require_string(_field(document, "container_id"), "container_id"), "container_id")
    _require_sha256(_field(document, "container_receipt_sha256"), "container_receipt_sha256")
    _require_sha256(_field(document, "backend_image_id"), "backend_image_id")
    _require_sha256(
        _field(document, "canonical_build_input_manifest_sha256"),
        "canonical_build_input_manifest_sha256",
    )
    _validate_classification_pair(
        _require_string(_field(document, "classification"), "classification"),
        _field(document, "immutable_outcome"),
    )
    _require_string(_field(document, "seal_timestamp"), "seal_timestamp")


def _validate_canonical_build_input_manifest(
    document: Mapping[str, object],
) -> tuple[str, dict[str, str]]:
    manifest = _require_string(
        _field(document, "canonical_build_input_manifest_bytes"),
        "canonical_build_input_manifest_bytes",
    )
    payload = _utf8(manifest, "canonical_build_input_manifest_bytes")
    if not payload or b"\r" in payload or not payload.endswith(b"\n"):
        _fail("canonical build-input manifest must be nonempty LF-finalized UTF-8")

    paths: list[str] = []
    manifest_digests: dict[str, str] = {}
    for raw_line in payload[:-1].split(b"\n"):
        if not raw_line:
            _fail("canonical build-input manifest contains an empty record")
        fields = raw_line.split(b"\t")
        if len(fields) != 2:
            _fail("canonical build-input manifest records require one tab")
        try:
            path = fields[0].decode("utf-8")
            digest = fields[1].decode("ascii")
        except UnicodeDecodeError:
            _fail("canonical build-input manifest record encoding is invalid")
        if not path or _contains_record_separator(path) or not _SHA256_RE.fullmatch(digest):
            _fail("canonical build-input manifest path or digest is invalid")
        paths.append(path)
        manifest_digests[path] = digest

    if tuple(paths) != _CANONICAL_BUILD_INPUT_PATHS:
        _fail("canonical build-input manifest paths or order differ from frozen policy")
    digest = _sha256(payload)
    if _require_sha256(
        _field(document, "canonical_build_input_manifest_sha256"),
        "canonical_build_input_manifest_sha256",
    ) != digest:
        _fail("canonical build-input manifest hash does not bind its exact bytes")
    return digest, manifest_digests


def _validate_artifact_inventory(
    document: Mapping[str, object], label: str
) -> tuple[Mapping[str, object], ...]:
    inventory = _require_sequence(_field(document, "artifact_inventory"), f"{label}.artifact_inventory")
    if not inventory:
        _fail(f"{label} artifact inventory must be nonempty")

    previous_path: str | None = None
    records: list[Mapping[str, object]] = []
    lines: list[bytes] = []
    for index, raw_record in enumerate(inventory):
        record = _require_mapping(raw_record, f"{label}.artifact_inventory[{index}]")
        path = _require_string(_field(record, "path"), f"{label}.artifact_inventory[{index}].path")
        digest = _require_sha256(_field(record, "sha256"), f"{label}.artifact_inventory[{index}].sha256")
        size = _require_integer(_field(record, "size"), f"{label}.artifact_inventory[{index}].size")
        if not path or size < 0 or _contains_record_separator(path) or path.startswith(("/", "\\")) or "/../" in f"/{path}/":
            _fail(f"{label} artifact path or size is invalid")
        if previous_path is not None and path <= previous_path:
            _fail(f"{label} artifact paths must be strictly canonical and unique")
        previous_path = path
        records.append(record)
        lines.append(_utf8(path, "artifact path") + b"\t" + digest.encode("ascii") + b"\t" + str(size).encode("ascii") + b"\n")

    if _require_sha256(_field(document, "artifact_inventory_sha256"), f"{label}.artifact_inventory_sha256") != _sha256(b"".join(lines)):
        _fail(f"{label} artifact_inventory_sha256 does not bind canonical inventory")
    return tuple(records)


def _validate_report_budget(document: Mapping[str, object]) -> None:
    budget = _require_mapping(_field(document, "report_budget"), "report_budget")
    if _require_integer(_field(budget, "total_bytes"), "report_budget.total_bytes") != 67108864:
        _fail("report_budget.total_bytes must equal the frozen 64 MiB total")
    if _require_integer(
        _field(budget, "reserved_metadata_bytes"), "report_budget.reserved_metadata_bytes"
    ) != 8388608:
        _fail("report_budget.reserved_metadata_bytes must equal the frozen 8 MiB reserve")
    metadata_used = _require_integer(
        _field(budget, "metadata_used_bytes"), "report_budget.metadata_used_bytes"
    )
    if not 0 <= metadata_used <= 8388608:
        _fail("report_budget.metadata_used_bytes exceeds its reserved bound")

    streams = _require_mapping(_field(budget, "streams"), "report_budget.streams")
    if set(streams) != set(_STREAM_CAPS):
        _fail("report_budget streams must be exactly build, vet, unit, and race")
    stream_used_total = 0
    stream_truncated = False
    for stage, (stdout_cap, stderr_cap) in _STREAM_CAPS.items():
        stage_streams = _require_mapping(_field(streams, stage), f"report_budget.streams.{stage}")
        if set(stage_streams) != {"stdout", "stderr"}:
            _fail(f"report_budget.streams.{stage} must contain stdout and stderr only")
        for name, cap in (("stdout", stdout_cap), ("stderr", stderr_cap)):
            usage = _require_mapping(
                _field(stage_streams, name), f"report_budget.streams.{stage}.{name}"
            )
            if _require_integer(
                _field(usage, "cap_bytes"), f"report_budget.streams.{stage}.{name}.cap_bytes"
            ) != cap:
                _fail(f"report_budget.streams.{stage}.{name} cap is not frozen")
            used = _require_integer(
                _field(usage, "used_bytes"), f"report_budget.streams.{stage}.{name}.used_bytes"
            )
            truncated = _field(usage, "truncated")
            if not 0 <= used <= cap or type(truncated) is not bool:
                _fail(f"report_budget.streams.{stage}.{name} use or truncation is invalid")
            if truncated and used != cap:
                _fail(f"report_budget.streams.{stage}.{name} truncation must reach its cap")
            stream_used_total += used
            stream_truncated = stream_truncated or truncated

    total_used = _require_integer(_field(budget, "total_used_bytes"), "report_budget.total_used_bytes")
    if total_used != metadata_used + stream_used_total:
        _fail("report_budget.total_used_bytes must equal metadata plus all stream use")
    if not 0 <= total_used <= 67108864 or stream_used_total > 67108864 - 8388608:
        _fail("report_budget use exceeds the frozen total or stream allocation")
    truncated_due_to_budget = _field(budget, "truncated_due_to_budget")
    raw_output_partial = _field(budget, "raw_output_partial")
    if type(truncated_due_to_budget) is not bool or type(raw_output_partial) is not bool:
        _fail("report_budget truncation flags must be boolean")
    if truncated_due_to_budget != stream_truncated or raw_output_partial != stream_truncated:
        _fail("report_budget truncation and raw-output-partial flags must bind stream truncation")


def _validate_fixed_host_config(document: Mapping[str, object]) -> None:
    host_config = _require_mapping(_field(document, "host_config"), "host_config")
    expected_fields = set(_FIXED_HOST_CONFIG) | {
        "security_opt",
        "cap_drop",
        "mounts",
    }
    if set(host_config) != expected_fields:
        _fail("host_config fields differ from the sealed observer policy")
    for field, expected in _FIXED_HOST_CONFIG.items():
        if _field(host_config, field) != expected:
            _fail(f"host_config.{field} differs from the sealed observer policy")
    if _string_tuple(_field(host_config, "security_opt"), "host_config.security_opt") != (
        "no-new-privileges:true",
    ):
        _fail("host_config.security_opt differs from the sealed observer policy")
    if _string_tuple(_field(host_config, "cap_drop"), "host_config.cap_drop") != ("ALL",):
        _fail("host_config.cap_drop differs from the sealed observer policy")

    mounts = _require_sequence(_field(host_config, "mounts"), "host_config.mounts")
    if len(mounts) != len(_FIXED_HOST_MOUNTS):
        _fail("host_config.mounts count differs from the sealed observer policy")
    for index, (raw_mount, expected) in enumerate(zip(mounts, _FIXED_HOST_MOUNTS, strict=True)):
        mount = _require_mapping(raw_mount, f"host_config.mounts[{index}]")
        if set(mount) != {"type", "source", "destination", "read_only", "size_bytes"}:
            _fail(f"host_config.mounts[{index}] fields differ from the sealed observer policy")
        actual = tuple(_field(mount, field) for field in ("type", "source", "destination", "read_only", "size_bytes"))
        if actual != expected:
            _fail(f"host_config.mounts[{index}] differs from the sealed observer policy")


def _require_frozen_subjects(value: object, label: str) -> Mapping[str, object]:
    subjects = _require_mapping(value, label)
    if set(subjects) != set(_FROZEN_SUBJECTS):
        _fail(f"{label} paths differ from the frozen 08C2 subject policy")
    for path, expected_digest in _FROZEN_SUBJECTS.items():
        if _require_sha256(_field(subjects, path), f"{label}.{path}") != expected_digest:
            _fail(f"{label}.{path} differs from the frozen 08C2 subject policy")
    return subjects


def _validate_source_snapshots(
    document: Mapping[str, object],
    manifest_sha256: str,
    fixed_image_ids: Mapping[str, object],
    facts: Mapping[str, object],
) -> None:
    git = _require_mapping(_field(document, "git"), "git")
    head_before = _require_string(_field(git, "head_before"), "git.head_before")
    head_after = _require_string(_field(git, "head_after"), "git.head_after")
    _require_git_sha(head_before, "git.head_before")
    _require_git_sha(head_after, "git.head_after")
    status_before = _require_string(_field(git, "status_before"), "git.status_before")
    status_after = _require_string(_field(git, "status_after"), "git.status_after")

    frozen_subjects = _require_frozen_subjects(_field(document, "frozen_subjects"), "frozen_subjects")
    snapshots = _require_mapping(_field(document, "source_snapshots"), "source_snapshots")
    ordered_snapshots = tuple(
        _require_mapping(_field(snapshots, stage), f"source_snapshots.{stage}")
        for stage in ("prebuild", "postgreen", "observer_snapshot")
    )

    snapshot_values: list[tuple[object, ...]] = []
    for stage, snapshot in zip(
        ("prebuild", "postgreen", "observer_snapshot"), ordered_snapshots, strict=True
    ):
        head = _require_string(_field(snapshot, "head"), f"source_snapshots.{stage}.head")
        _require_git_sha(head, f"source_snapshots.{stage}.head")
        status = _require_string(_field(snapshot, "status"), f"source_snapshots.{stage}.status")
        source_roots = _require_mapping(
            _field(snapshot, "source_root_sha256"), f"source_snapshots.{stage}.source_root_sha256"
        )
        _validate_digest_mapping(source_roots, f"source_snapshots.{stage}.source_root_sha256")
        snapshot_subjects = _require_frozen_subjects(
            _field(snapshot, "frozen_subjects"), f"source_snapshots.{stage}.frozen_subjects"
        )
        snapshot_manifest = _require_sha256(
            _field(snapshot, "canonical_build_input_manifest_sha256"),
            f"source_snapshots.{stage}.canonical_build_input_manifest_sha256",
        )
        snapshot_images = _require_mapping(
            _field(snapshot, "fixed_image_ids"), f"source_snapshots.{stage}.fixed_image_ids"
        )
        _require_sha256(
            _field(snapshot_images, "python_contract"),
            f"source_snapshots.{stage}.fixed_image_ids.python_contract",
        )
        _require_sha256(
            _field(snapshot_images, "backend_unit"),
            f"source_snapshots.{stage}.fixed_image_ids.backend_unit",
        )
        snapshot_values.append(
            (head, status, source_roots, snapshot_subjects, snapshot_manifest, snapshot_images)
        )

    prebuild, postgreen, observer_snapshot = ordered_snapshots
    if _field(prebuild, "head") != head_before or _field(prebuild, "status") != status_before:
        _fail("prebuild snapshot must bind Git before state")
    if _field(observer_snapshot, "head") != head_after or _field(observer_snapshot, "status") != status_after:
        _fail("observer snapshot must bind Git after state")
    if _field(observer_snapshot, "frozen_subjects") != frozen_subjects:
        _fail("observer snapshot must bind top-level frozen subjects")
    if _field(observer_snapshot, "canonical_build_input_manifest_sha256") != manifest_sha256:
        _fail("observer snapshot must bind top-level canonical manifest")
    if _field(observer_snapshot, "fixed_image_ids") != fixed_image_ids:
        _fail("observer snapshot must bind top-level fixed image IDs")

    source_changed = head_before != head_after or status_before != status_after or bool(status_before) or bool(status_after)
    source_changed = source_changed or any(
        snapshot_values[index] != snapshot_values[0]
        for index in range(1, len(snapshot_values))
    )
    if source_changed and _field(facts, "source_mutated") is not True:
        _fail("source snapshot drift must be classified as source_mutated")


def _validate_classification_facts(
    document: Mapping[str, object], container_id: str
) -> Mapping[str, object]:
    facts = _require_mapping(_field(document, "classification_facts"), "classification_facts")
    for name in (
        "source_mutated",
        "invalid_provenance",
        "oom",
        "resource",
        "budget",
        "tmpfs",
        "timeout",
        "discovery_failure",
        "stage_failure",
    ):
        if type(_field(facts, name)) is not bool:
            _fail(f"classification_facts.{name} must be boolean")
    predicates = _require_mapping(_field(facts, "predicates"), "classification_facts.predicates")
    for name in (
        "preflight_ram_floor_met",
        "preflight_c_free_floor_met",
        "preflight_disk_pressure_2_of_3",
        "two_consecutive_ram_below_2_gib",
        "rolling_disk_pressure_2_of_3",
        "prohibited_backend_unit_overlap",
        "prohibited_mail_calendar_qa_test_build_overlap",
        "buildkit_idle",
        "whole_run_deadline_expired",
    ):
        if type(_field(predicates, name)) is not bool:
            _fail(f"classification_facts.predicates.{name} must be boolean")

    stops = _require_sequence(_field(facts, "explicit_stops"), "classification_facts.explicit_stops")
    parsed_stops: list[tuple[str, int]] = []
    for index, raw_stop in enumerate(stops):
        stop = _require_sequence(raw_stop, f"classification_facts.explicit_stops[{index}]")
        if len(stop) != 2:
            _fail("explicit stop records require kind and timestamp")
        kind = _require_string(stop[0], f"classification_facts.explicit_stops[{index}][0]")
        timestamp = _require_integer(stop[1], f"classification_facts.explicit_stops[{index}][1]")
        if kind not in _STOP_TIE_ORDER or timestamp < 0:
            _fail("explicit stop kind or timestamp is invalid")
        parsed_stops.append((kind, timestamp))

    selected = None
    if parsed_stops:
        selected = min(parsed_stops, key=lambda stop: (stop[1], _STOP_TIE_ORDER.index(stop[0])))
    _require_selected_stop(_field(facts, "selected_stop"), selected, "classification_facts.selected_stop")
    _require_selected_stop(_field(document, "selected_stop"), selected, "selected_stop")

    _validate_guarded_stop(document, selected, container_id)
    expected_classification = _classification_from_facts(facts, selected)
    expected_outcome: object = (
        "TIMEOUT_UNTRIAGED"
        if expected_classification == "SEALED_TIMEOUT_UNTRIAGED"
        else None
    )
    for container, label in ((facts, "classification_facts"), (document, "observer receipt")):
        classification = _require_string(_field(container, "classification"), f"{label}.classification")
        immutable_outcome = _field(container, "immutable_outcome")
        if classification != expected_classification or immutable_outcome != expected_outcome:
            _fail(f"{label} classification does not match frozen priority relation")
        _validate_classification_pair(classification, immutable_outcome)
    return facts


def _classification_from_facts(
    facts: Mapping[str, object], selected_stop: tuple[str, int] | None
) -> str:
    if _field(facts, "source_mutated") is True:
        return "SEALED_INVALID_SOURCE_MUTATED"
    if _field(facts, "invalid_provenance") is True:
        return "SEALED_INVALID_PROVENANCE"
    if selected_stop is not None:
        return _STOP_CLASSIFICATIONS[selected_stop[0]]
    for field, classification in (
        ("oom", "SEALED_INFRA_ABORTED_OOM"),
        ("resource", "SEALED_INFRA_ABORTED_RESOURCE_CONTRACT"),
        ("budget", "SEALED_INFRA_ABORTED_REPORT_BUDGET"),
        ("tmpfs", "SEALED_INFRA_ABORTED_TMPFS"),
        ("timeout", "SEALED_TIMEOUT_UNTRIAGED"),
        ("discovery_failure", "SEALED_EXECUTABLE_FAIL_DISCOVERY"),
        ("stage_failure", "SEALED_EXECUTABLE_FAIL_STAGE"),
    ):
        if _field(facts, field) is True:
            return classification
    return "SEALED_EXECUTABLE_PASS"


def _validate_guarded_stop(
    document: Mapping[str, object], selected_stop: tuple[str, int] | None, container_id: str
) -> None:
    stop_argv = _field(document, "stop_argv")
    guarded_stop = _require_mapping(_field(document, "guarded_stop"), "guarded_stop")
    if selected_stop is None:
        if stop_argv is not None or any(_field(guarded_stop, name) is not None for name in ("predicate", "argv", "result")):
            _fail("unselected stops must have null stop evidence")
        return

    expected_argv = ("docker", "stop", "--timeout", "5", container_id)
    if _string_tuple(stop_argv, "stop_argv") != expected_argv:
        _fail("selected stop must use the one frozen Docker stop argv")
    if _require_string(_field(guarded_stop, "predicate"), "guarded_stop.predicate") != _STOP_PREDICATES[selected_stop[0]]:
        _fail("guarded stop predicate does not match the selected stop")
    if _string_tuple(_field(guarded_stop, "argv"), "guarded_stop.argv") != expected_argv:
        _fail("guarded stop argv does not bind the selected container")
    if _field(guarded_stop, "result") != "stopped":
        _fail("guarded stop result must record the one authorized stop")


def _require_final_bindings(
    final_receipt: Mapping[str, object],
    *,
    run_id: str,
    nonce: str,
    container_id: str,
    container_receipt_sha256: str,
    backend_image_id: str,
    manifest_sha256: str,
    classification: str,
    immutable_outcome: object,
    seal_timestamp: object,
) -> None:
    expected = {
        "run_id": run_id,
        "owner_nonce": nonce,
        "container_id": container_id,
        "container_receipt_sha256": container_receipt_sha256,
        "backend_image_id": backend_image_id,
        "canonical_build_input_manifest_sha256": manifest_sha256,
        "classification": classification,
        "immutable_outcome": immutable_outcome,
        "seal_timestamp": seal_timestamp,
    }
    for field, value in expected.items():
        if _field(final_receipt, field) != value:
            _fail(f"final receipt {field} does not exactly bind observer receipt")


def _validate_classification_pair(classification: str, immutable_outcome: object) -> None:
    if classification not in _SEALED_CLASSIFICATIONS:
        _fail("classification is not a frozen sealed classification")
    expected = "TIMEOUT_UNTRIAGED" if classification == "SEALED_TIMEOUT_UNTRIAGED" else None
    if immutable_outcome != expected:
        _fail("immutable timeout outcome must pair only with sealed timeout classification")


def _require_selected_stop(value: object, expected: tuple[str, int] | None, label: str) -> None:
    if expected is None:
        if value is not None:
            _fail(f"{label} must be null when no explicit stop exists")
        return
    if tuple(_require_sequence(value, label)) != expected:
        _fail(f"{label} must select the earliest frozen-priority explicit stop")


def _validate_digest_mapping(mapping: Mapping[str, object], label: str) -> None:
    if not mapping:
        _fail(f"{label} must be nonempty")
    for key, value in mapping.items():
        if not isinstance(key, str) or not key:
            _fail(f"{label} contains an invalid key")
        _require_sha256(value, f"{label}.{key}")


def _require_mapping(value: object, label: str) -> Mapping[str, object]:
    if not isinstance(value, Mapping):
        _fail(f"{label} must be a mapping")
    return value


def _require_sequence(value: object, label: str) -> Sequence[object]:
    if isinstance(value, (str, bytes, bytearray)) or not isinstance(value, Sequence):
        _fail(f"{label} must be an array")
    return value


def _string_tuple(value: object, label: str) -> tuple[str, ...]:
    return tuple(_require_string(item, label) for item in _require_sequence(value, label))


def _field(document: Mapping[str, object], field: str) -> object:
    try:
        return document[field]
    except ReceiptContractError:
        raise
    except Exception as error:
        raise ReceiptContractError(f"cannot read required field {field!r}") from error


def _require_exact(document: Mapping[str, object], field: str, expected: str) -> None:
    if _field(document, field) != expected:
        _fail(f"{field} must equal {expected!r}")


def _require_string(value: object, label: str) -> str:
    if not isinstance(value, str):
        _fail(f"{label} must be a string")
    return value


def _require_integer(value: object, label: str) -> int:
    if type(value) is not int:
        _fail(f"{label} must be an integer")
    return value


def _require_sha256(value: object, label: str) -> str:
    digest = _require_string(value, label)
    if not _SHA256_RE.fullmatch(digest):
        _fail(f"{label} must be a lowercase sha256 digest")
    return digest


def _require_run_id(value: str, label: str) -> None:
    if not _RUN_ID_RE.fullmatch(value):
        _fail(f"{label} is not a canonical backend-unit run ID")


def _require_nonce(value: str, label: str) -> None:
    if not _NONCE_RE.fullmatch(value):
        _fail(f"{label} is not a lowercase 32-hex nonce")


def _require_engine_id(value: str, label: str) -> None:
    if not _ENGINE_ID_RE.fullmatch(value):
        _fail(f"{label} is not a lowercase full engine ID")


def _require_hostname(value: str, label: str) -> None:
    if not _HOSTNAME_RE.fullmatch(value):
        _fail(f"{label} is not a lowercase engine-hostname prefix")


def _require_git_sha(value: str, label: str) -> None:
    if not _GIT_SHA_RE.fullmatch(value):
        _fail(f"{label} is not a lowercase Git SHA")


def _contains_record_separator(value: str) -> bool:
    return "\t" in value or "\r" in value or "\n" in value


def _utf8(value: str, label: str) -> bytes:
    try:
        return value.encode("utf-8")
    except UnicodeEncodeError:
        _fail(f"{label} is not valid UTF-8 text")


def _sha256(payload: bytes) -> str:
    return "sha256:" + hashlib.sha256(payload).hexdigest()


def _fail(message: str) -> None:
    raise ReceiptContractError(message)
