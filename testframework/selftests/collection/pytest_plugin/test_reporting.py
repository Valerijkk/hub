from __future__ import annotations

import traceback
from pathlib import Path

import pytest

from hub_testkit.runtime.redaction import Redactor
from hub_testkit.traceability import CaseMetadata, CaseRecord, PytestAutomation

from hub_testkit.pytest_plugin.markers import CollectionError
from hub_testkit.pytest_plugin.reporting import apply_report_properties


REDACTION_POLICY_PATH = Path(__file__).resolve().parents[3] / "policies" / "redaction.toml"


class _Item:
    def __init__(self, properties: object) -> None:
        self.user_properties = properties


class _NonStringCanary:
    def __repr__(self) -> str:
        return "NONSTRING-SECRET-CANARY"


def _record() -> CaseRecord:
    return CaseRecord(
        metadata=CaseMetadata(
            schema="hub.testframework.case/v1",
            id="TC-PLUGIN-0001",
            title="Properties case",
            status="active",
            task="TASK-206",
            requirement="TF-PLUGIN",
            risk="RISK-HIGH",
            scenario="happy",
            subject="report properties",
            preconditions=("isolated fixture",),
            oracle="properties are bound",
            cleanup="fixture removed",
            automation=PytestAutomation(
                kind="pytest",
                layer="api",
                node_id="testframework/tests/api/health/test_live.py::test_live",
            ),
        ),
        source_path=Path("testframework/cases/TC-PLUGIN-0001.md"),
        markdown_body="# Properties case\n",
    )


def test_property_batch_preflights_all_items_before_mutating_any_list() -> None:
    first = _Item([("unrelated", "preserved")])
    canary = "SECOND-SECRET-CANARY"
    second = _Item(canary)

    with pytest.raises(CollectionError) as raised:
        apply_report_properties(
            ((first, _record(), "api"), (second, _record(), "api")),
            Redactor.from_policy(REDACTION_POLICY_PATH),
        )

    assert first.user_properties == [("unrelated", "preserved")]
    assert str(raised.value) == "report properties are invalid"
    assert raised.value.__cause__ is None
    assert raised.value.__suppress_context__ is True
    assert canary not in "".join(traceback.format_exception(raised.value))


def test_property_batch_preserves_safe_unrelated_string_properties() -> None:
    item = _Item([("unrelated", "preserved")])

    apply_report_properties(
        ((item, _record(), "api"),),
        Redactor.from_policy(REDACTION_POLICY_PATH),
    )

    assert item.user_properties == [
        ("unrelated", "preserved"),
        ("hub.case_id", "TC-PLUGIN-0001"),
        ("hub.layer", "api"),
        ("hub.task", "TASK-206"),
        ("hub.requirement", "TF-PLUGIN"),
        ("hub.risk", "RISK-HIGH"),
    ]


@pytest.mark.parametrize(
    ("value", "canary"),
    (
        ("Bearer UNRELATED-SECRET-CANARY", "UNRELATED-SECRET-CANARY"),
        (_NonStringCanary(), "NONSTRING-SECRET-CANARY"),
    ),
)
def test_property_batch_rejects_unsafe_unrelated_values_without_partial_mutation(
    value: object,
    canary: str,
) -> None:
    first = _Item([("unrelated", "preserved")])
    second = _Item([("unrelated", value)])

    with pytest.raises(CollectionError) as raised:
        apply_report_properties(
            ((first, _record(), "api"), (second, _record(), "api")),
            Redactor.from_policy(REDACTION_POLICY_PATH),
        )

    assert first.user_properties == [("unrelated", "preserved")]
    assert str(raised.value) == "report properties are invalid"
    assert raised.value.__cause__ is None
    assert raised.value.__suppress_context__ is True
    assert canary not in "".join(traceback.format_exception(raised.value))
