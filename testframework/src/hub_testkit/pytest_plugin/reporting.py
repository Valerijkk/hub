from __future__ import annotations

from collections.abc import Iterable
from dataclasses import dataclass
from typing import Any

from hub_testkit.runtime.redaction import Redactor
from hub_testkit.traceability import CaseRecord

from .markers import CollectionError


REPORT_PROPERTY_KEYS = (
    "hub.case_id",
    "hub.layer",
    "hub.task",
    "hub.requirement",
    "hub.risk",
)


@dataclass(frozen=True, slots=True)
class _PropertyPlan:
    item: Any
    container: list[tuple[str, object]]
    original: tuple[tuple[str, object], ...]
    final: list[tuple[str, object]]


def apply_report_properties(
    bound: Iterable[tuple[Any, CaseRecord, str]],
    redactor: Redactor,
) -> None:
    try:
        plans = tuple(
            _property_plan(item, record, layer, redactor)
            for item, record, layer in bound
        )
    except Exception:
        raise CollectionError("report properties are invalid") from None

    mutated: list[_PropertyPlan] = []
    try:
        for plan in plans:
            if plan.item.user_properties is not plan.container or type(plan.container) is not list:
                raise CollectionError("report properties are invalid")
            mutated.append(plan)
            plan.container[:] = plan.final
    except Exception:
        for plan in reversed(mutated):
            try:
                plan.container[:] = plan.original
            except Exception:
                pass
        raise CollectionError("report properties are invalid") from None


def properties_for_record(
    record: CaseRecord,
    layer: str,
    redactor: Redactor,
) -> tuple[tuple[str, str], ...]:
    try:
        values = (
            record.metadata.id,
            layer,
            record.metadata.task,
            record.metadata.requirement,
            record.metadata.risk,
        )
        redacted = tuple(redactor.redact_text(value) for value in values)
    except Exception:
        raise CollectionError("report properties are invalid") from None
    if any(type(value) is not str for value in redacted):
        raise CollectionError("report properties are invalid")
    return tuple(zip(REPORT_PROPERTY_KEYS, redacted, strict=True))


def _property_plan(
    item: Any,
    record: CaseRecord,
    layer: str,
    redactor: Redactor,
) -> _PropertyPlan:
    try:
        properties = item.user_properties
        if type(properties) is not list:
            raise CollectionError("report properties are invalid")
        original = tuple(
            _valid_property(property_item, redactor) for property_item in properties
        )
        final = [*original, *properties_for_record(record, layer, redactor)]
    except Exception:
        raise CollectionError("report properties are invalid") from None
    return _PropertyPlan(
        item=item,
        container=properties,
        original=original,
        final=final,
    )


def _valid_property(property_item: object, redactor: Redactor) -> tuple[str, object]:
    if type(property_item) is not tuple or len(property_item) != 2:
        raise CollectionError("report properties are invalid")
    name, value = property_item
    if type(name) is not str or type(value) is not str or name.startswith("hub."):
        raise CollectionError("report properties are invalid")
    if redactor.redact_text(name) != name or redactor.redact_text(value) != value:
        raise CollectionError("report properties are invalid")
    return name, value
