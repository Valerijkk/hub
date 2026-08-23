from __future__ import annotations

import traceback
from pathlib import Path

import pytest

from hub_testkit.traceability import (
    CaseIndex,
    CaseRecord,
    LocustAutomation,
    ManualAutomation,
    PytestAutomation,
    TraceabilityError,
    load_case,
    load_marker_policy,
)


PROJECT_ROOT = Path(__file__).resolve().parents[3]
SCHEMA_PATH = PROJECT_ROOT / "schemas" / "test-case.schema.json"
MARKERS_PATH = PROJECT_ROOT / "policies" / "markers.toml"
PYTEST_NODE = "testframework/tests/api/health/test_live.py::test_live"


def _policy():
    return load_marker_policy(MARKERS_PATH)


def _pytest_automation(
    node_id: str = PYTEST_NODE,
    layer: str = "api",
    extra: str = "",
) -> str:
    return "\n".join(
        (
            "[automation]",
            'kind = "pytest"',
            f'layer = "{layer}"',
            f'node_id = "{node_id}"',
            extra,
        )
    ).rstrip()


def _locust_automation() -> str:
    return "\n".join(
        (
            "[automation]",
            'kind = "locust"',
            'entrypoint = "testframework/load/health.py"',
        )
    )


def _manual_automation() -> str:
    return "\n".join(
        (
            "[automation]",
            'kind = "manual"',
            'evidence_owner = "qa-owner"',
        )
    )


def _case_text(
    case_id: str,
    *,
    status: str = "active",
    scenario: str = "happy",
    automation: str | None = None,
) -> str:
    return "\n".join(
        (
            "+++",
            'schema = "hub.testframework.case/v1"',
            f'id = "{case_id}"',
            'title = "Traceability case"',
            f'status = "{status}"',
            'task = "TASK-206"',
            'requirement = "TF-TRACEABILITY"',
            'risk = "RISK-HIGH"',
            f'scenario = "{scenario}"',
            'subject = "case loader"',
            'preconditions = ["isolated fixture"]',
            'oracle = "record is loaded"',
            'cleanup = "temporary file removed"',
            automation or _pytest_automation(),
            "+++",
            "# Case",
            "",
            "The loader accepts this bounded record.",
        )
    )


def _write_case(tmp_path: Path, case_id: str, contents: str) -> Path:
    path = tmp_path / f"{case_id}.md"
    path.write_text(contents, encoding="utf-8", newline="\n")
    return path


def _load(tmp_path: Path, case_id: str, contents: str | None = None):
    return load_case(
        _write_case(tmp_path, case_id, contents or _case_text(case_id)),
        SCHEMA_PATH,
        _policy(),
    )


class _HostileRecordIterable:
    canary = "hostile-record-iteration-canary"

    def __iter__(self):
        raise RuntimeError(self.canary)


class _HostileCollectionIterable:
    canary = "hostile-collection-iteration-canary"

    def __iter__(self):
        raise RuntimeError(self.canary)


def test_marker_policy_is_closed_and_literal() -> None:
    policy = _policy()

    assert policy.schema_ == "hub.testframework.markers/v1"
    assert policy.statuses == ("active", "planned", "retired")
    assert policy.scenarios == (
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
    assert policy.pytest_layers == ("contract", "api", "integration", "e2e")
    assert policy.automation_kinds == ("pytest", "locust", "manual")
    assert policy.case_id_pattern == r"^TC-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*-[0-9]{4}$"


@pytest.mark.parametrize(
    "contents",
    (
        MARKERS_PATH.read_text(encoding="utf-8") + "\nunexpected = true\n",
        MARKERS_PATH.read_text(encoding="utf-8").replace(
            'statuses = ["active", "planned", "retired"]',
            'statuses = ["active", "active", "retired"]',
        ),
        MARKERS_PATH.read_text(encoding="utf-8").replace(
            'pytest_layers = ["contract", "api", "integration", "e2e"]',
            "pytest_layers = true",
        ),
        MARKERS_PATH.read_text(encoding="utf-8").replace(
            'case_id_pattern = "^TC-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*-[0-9]{4}$"\n',
            "",
        ),
    ),
)
def test_marker_policy_rejects_unknown_missing_duplicate_or_wrong_typed_values(
    tmp_path: Path,
    contents: str,
) -> None:
    path = tmp_path / "markers.toml"
    path.write_text(contents, encoding="utf-8", newline="\n")

    with pytest.raises(TraceabilityError):
        load_marker_policy(path)


def test_loader_accepts_valid_pytest_locust_and_manual_records(tmp_path: Path) -> None:
    pytest_record = _load(tmp_path, "TC-TRACE-0001")
    locust_record = _load(
        tmp_path,
        "TC-TRACE-0002",
        _case_text(
            "TC-TRACE-0002",
            status="planned",
            automation=_locust_automation(),
        ),
    )
    manual_record = _load(
        tmp_path,
        "TC-TRACE-0003",
        _case_text(
            "TC-TRACE-0003",
            status="retired",
            automation=_manual_automation(),
        ),
    )

    assert isinstance(pytest_record.metadata.automation, PytestAutomation)
    assert isinstance(locust_record.metadata.automation, LocustAutomation)
    assert isinstance(manual_record.metadata.automation, ManualAutomation)
    assert pytest_record.markdown_body.startswith("# Case\n")
    assert pytest_record.metadata.preconditions == ("isolated fixture",)


def test_planned_manual_case_requires_manual_policy_admission(tmp_path: Path) -> None:
    case_id = "TC-TRACE-0004"
    contents = _case_text(
        case_id,
        status="planned",
        automation=_manual_automation(),
    )
    admitted = _policy()
    manual_is_disabled = admitted.model_copy(
        update={"automation_kinds": ("pytest", "locust")}
    )
    disabled_path = _write_case(
        tmp_path,
        "TC-TRACE-0005",
        _case_text(
            "TC-TRACE-0005",
            status="planned",
            automation=_manual_automation(),
        ),
    )

    assert isinstance(_load(tmp_path, case_id, contents).metadata.automation, ManualAutomation)
    with pytest.raises(TraceabilityError):
        load_case(disabled_path, SCHEMA_PATH, manual_is_disabled)


@pytest.mark.parametrize(
    ("case_id", "status", "automation"),
    (
        ("TC-TRACE-0006", "planned", _pytest_automation()),
        ("TC-TRACE-0007", "retired", _pytest_automation()),
        ("TC-TRACE-0008", "active", _locust_automation()),
        ("TC-TRACE-0009", "active", _manual_automation()),
    ),
)
def test_loader_rejects_invalid_status_to_automation_combinations(
    tmp_path: Path,
    case_id: str,
    status: str,
    automation: str,
) -> None:
    with pytest.raises(TraceabilityError):
        _load(
            tmp_path,
            case_id,
            _case_text(case_id, status=status, automation=automation),
        )


@pytest.mark.parametrize(
    "case_id",
    (
        "TC-AUTH-03",
        "HUB-TC-AUTH-0003",
        "TC-auth-0003",
        "TC-AUTH-000",
        "TC-AUTH-00030",
    ),
)
def test_loader_rejects_noncanonical_case_identifier(tmp_path: Path, case_id: str) -> None:
    with pytest.raises(TraceabilityError):
        _load(tmp_path, case_id, _case_text(case_id))


@pytest.mark.parametrize(
    "contents",
    (
        _case_text("TC-TRACE-0010").replace(
            'preconditions = ["isolated fixture"]',
            'preconditions = ["isolated fixture"]\nunexpected = "forbidden"',
        ),
        _case_text("TC-TRACE-0011").replace('cleanup = "temporary file removed"\n', ""),
        _case_text("TC-TRACE-0012").replace('title = "Traceability case"', "title = true"),
        _case_text("TC-TRACE-0013").replace(
            'preconditions = ["isolated fixture"]',
            'preconditions = ["same", "same"]',
        ),
    ),
)
def test_loader_rejects_open_missing_or_wrong_typed_metadata(
    tmp_path: Path,
    contents: str,
) -> None:
    case_id = contents.split('id = "', 1)[1].split('"', 1)[0]

    with pytest.raises(TraceabilityError):
        _load(tmp_path, case_id, contents)


@pytest.mark.parametrize(
    "contents",
    (
        _case_text(
            "TC-TRACE-0020",
            automation=_pytest_automation(layer="contract"),
        ),
        _case_text(
            "TC-TRACE-0021",
            automation=_pytest_automation(
                "testframework/tests/api/../test_escape.py::test_escape"
            ),
        ),
        _case_text(
            "TC-TRACE-0022",
            automation=_pytest_automation(extra='entrypoint = "testframework/load/health.py"'),
        ),
        _case_text(
            "TC-TRACE-0023",
            status="active",
            automation=_manual_automation(),
        ),
    ),
)
def test_loader_rejects_wrong_or_cross_kind_automation(tmp_path: Path, contents: str) -> None:
    case_id = contents.split('id = "', 1)[1].split('"', 1)[0]

    with pytest.raises(TraceabilityError):
        _load(tmp_path, case_id, contents)


@pytest.mark.parametrize(
    "contents",
    (
        "---\nschema: hub.testframework.case/v1\n---\n# Case\n",
        _case_text("TC-TRACE-0031") + "+++\n",
        _case_text("TC-TRACE-0032").replace("\n", "\r\n"),
    ),
)
def test_loader_rejects_noncanonical_frontmatter(tmp_path: Path, contents: str) -> None:
    case_id = "TC-TRACE-0031" if "0031" in contents else "TC-TRACE-0032"
    if contents.startswith("---"):
        case_id = "TC-TRACE-0030"

    with pytest.raises(TraceabilityError):
        _load(tmp_path, case_id, contents)


def test_loader_rejects_invalid_utf8_and_hides_decoder_context(tmp_path: Path) -> None:
    case_id = "TC-TRACE-0033"
    path = tmp_path / f"{case_id}.md"
    canary = b"traceability-decoder-canary"
    path.write_bytes(_case_text(case_id).encode("utf-8") + canary + b"\xff")

    with pytest.raises(TraceabilityError) as error:
        load_case(path, SCHEMA_PATH, _policy())

    formatted = "".join(traceback.format_exception(error.value))
    assert error.value.__cause__ is None
    assert canary.decode("ascii") not in formatted


def test_loader_rejects_nul_and_oversized_body(tmp_path: Path) -> None:
    nul_id = "TC-TRACE-0034"
    nul_path = tmp_path / f"{nul_id}.md"
    nul_path.write_bytes(_case_text(nul_id).encode("utf-8") + b"\x00")
    body_id = "TC-TRACE-0035"
    body_path = tmp_path / f"{body_id}.md"
    body_path.write_text(
        _case_text(body_id).replace(
            "The loader accepts this bounded record.",
            "x" * (64 * 1024 + 1),
        ),
        encoding="utf-8",
        newline="\n",
    )

    with pytest.raises(TraceabilityError):
        load_case(nul_path, SCHEMA_PATH, _policy())
    with pytest.raises(TraceabilityError):
        load_case(body_path, SCHEMA_PATH, _policy())


def test_loader_rejects_oversized_case_and_mismatched_filename(tmp_path: Path) -> None:
    oversized_id = "TC-TRACE-0036"
    oversized_path = tmp_path / f"{oversized_id}.md"
    oversized_path.write_bytes(b"x" * (256 * 1024 + 1))
    mismatched_path = _write_case(
        tmp_path,
        "TC-TRACE-0037",
        _case_text("TC-TRACE-0038"),
    )

    with pytest.raises(TraceabilityError):
        load_case(oversized_path, SCHEMA_PATH, _policy())
    with pytest.raises(TraceabilityError):
        load_case(mismatched_path, SCHEMA_PATH, _policy())


def test_loader_rejects_symlink_when_the_platform_supports_it(tmp_path: Path) -> None:
    case_id = "TC-TRACE-0039"
    target = _write_case(tmp_path, case_id, _case_text(case_id))
    link = tmp_path / "TC-TRACE-0040.md"
    try:
        link.symlink_to(target)
    except OSError:
        pytest.skip("symlinks are unavailable on this platform")

    with pytest.raises(TraceabilityError):
        load_case(link, SCHEMA_PATH, _policy())


def test_case_index_enforces_duplicates_and_exact_active_pytest_collection(tmp_path: Path) -> None:
    first = _load(tmp_path, "TC-TRACE-0040")
    planned_manual = _load(
        tmp_path,
        "TC-TRACE-0041",
        _case_text(
            "TC-TRACE-0041",
            status="planned",
            automation=_manual_automation(),
        ),
    )
    active_orphan = _load(
        tmp_path,
        "TC-TRACE-0042",
        _case_text(
            "TC-TRACE-0042",
            automation=_pytest_automation(
                "testframework/tests/api/health/test_orphan.py::test_orphan"
            ),
        ),
    )
    duplicate_node = _load(
        tmp_path,
        "TC-TRACE-0043",
        _case_text("TC-TRACE-0043"),
    )

    index = CaseIndex.build((first, planned_manual))
    assert index.records == (first, planned_manual)
    assert index.lookup("TC-TRACE-0040") == first
    index.validate_pytest_collection((PYTEST_NODE,))

    with pytest.raises(TraceabilityError):
        index.validate_pytest_collection(())
    with pytest.raises(TraceabilityError):
        index.validate_pytest_collection(
            (
                PYTEST_NODE,
                active_orphan.metadata.automation.node_id,
            )
        )
    with pytest.raises(TraceabilityError):
        index.validate_pytest_collection((active_orphan.metadata.automation.node_id,))
    with pytest.raises(TraceabilityError):
        index.validate_pytest_collection(("not-a-node",))
    with pytest.raises(TraceabilityError):
        CaseIndex.build(
            (first, first.model_copy(update={"source_path": tmp_path / "other.md"}))
        )
    with pytest.raises(TraceabilityError):
        CaseIndex.build(
            (
                first,
                planned_manual.model_copy(update={"source_path": first.source_path}),
            )
        )
    with pytest.raises(TraceabilityError):
        CaseIndex.build((first, duplicate_node))


def test_case_index_suppresses_hostile_iteration_context(tmp_path: Path) -> None:
    index = CaseIndex.build((_load(tmp_path, "TC-TRACE-0044"),))

    with pytest.raises(TraceabilityError) as record_error:
        CaseIndex.build(_HostileRecordIterable())
    with pytest.raises(TraceabilityError) as collection_error:
        index.validate_pytest_collection(_HostileCollectionIterable())

    for error, canary in (
        (record_error.value, _HostileRecordIterable.canary),
        (collection_error.value, _HostileCollectionIterable.canary),
    ):
        formatted = "".join(traceback.format_exception(error))
        assert error.__cause__ is None
        assert error.__suppress_context__ is True
        assert canary not in formatted


def test_case_index_revalidates_model_construct_and_model_copy_bypasses(tmp_path: Path) -> None:
    record = _load(tmp_path, "TC-TRACE-0045")
    constructed = CaseRecord.model_construct(
        metadata="not-case-metadata",
        source_path=record.source_path,
        markdown_body=record.markdown_body,
    )
    copied = record.model_copy(update={"source_path": "bypass-source-path-canary"})

    for invalid_record, canary in (
        (constructed, "not-case-metadata"),
        (copied, "bypass-source-path-canary"),
    ):
        with pytest.raises(TraceabilityError) as error:
            CaseIndex.build((invalid_record,))
        formatted = "".join(traceback.format_exception(error.value))
        assert error.value.__cause__ is None
        assert error.value.__suppress_context__ is True
        assert canary not in formatted


def test_case_index_can_only_be_constructed_by_build_and_exposes_immutable_state(
    tmp_path: Path,
) -> None:
    record = _load(tmp_path, "TC-TRACE-0046")
    index = CaseIndex.build((record,))

    with pytest.raises(TypeError):
        CaseIndex((record,), {}, frozenset())
    with pytest.raises(TypeError):
        index._by_id[record.metadata.id] = record
    assert isinstance(index._active_pytest_nodes, frozenset)
