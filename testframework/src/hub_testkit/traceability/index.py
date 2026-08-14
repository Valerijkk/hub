from __future__ import annotations

import os
import re
import warnings
from collections.abc import Iterable, Mapping
from dataclasses import dataclass
from types import MappingProxyType

from .loader import TraceabilityError
from .model import CASE_ID_PATTERN, PYTEST_NODE_PATTERN, CaseRecord, PytestAutomation


_CASE_ID_RE = re.compile(CASE_ID_PATTERN)
_PYTEST_NODE_RE = re.compile(PYTEST_NODE_PATTERN)


@dataclass(frozen=True, slots=True, init=False)
class CaseIndex:
    records: tuple[CaseRecord, ...]
    _by_id: Mapping[str, CaseRecord]
    _active_pytest_nodes: frozenset[str]

    @classmethod
    def build(cls, records: Iterable[CaseRecord]) -> CaseIndex:
        try:
            ordered = tuple(records)
        except Exception:
            raise TraceabilityError("case records are invalid") from None
        validated = tuple(_normalize_record(record) for record in ordered)
        by_id: dict[str, CaseRecord] = {}
        source_paths: set[str] = set()
        active_nodes: set[str] = set()
        for record in validated:
            case_id = record.metadata.id
            source_path = _source_path_key(record)
            if case_id in by_id or source_path in source_paths:
                raise TraceabilityError("case records are not unique")
            by_id[case_id] = record
            source_paths.add(source_path)
            automation = record.metadata.automation
            if record.metadata.status == "active" and isinstance(automation, PytestAutomation):
                if automation.node_id in active_nodes:
                    raise TraceabilityError("active pytest node IDs are not unique")
                active_nodes.add(automation.node_id)
        return cls._create(
            records=validated,
            by_id=MappingProxyType(by_id),
            active_pytest_nodes=frozenset(active_nodes),
        )

    @classmethod
    def _create(
        cls,
        *,
        records: tuple[CaseRecord, ...],
        by_id: Mapping[str, CaseRecord],
        active_pytest_nodes: frozenset[str],
    ) -> CaseIndex:
        instance: CaseIndex = object.__new__(cls)
        object.__setattr__(instance, "records", records)
        object.__setattr__(instance, "_by_id", by_id)
        object.__setattr__(instance, "_active_pytest_nodes", active_pytest_nodes)
        return instance

    def lookup(self, case_id: str) -> CaseRecord:
        if type(case_id) is not str or _CASE_ID_RE.fullmatch(case_id) is None:
            raise TraceabilityError("case ID is invalid")
        record = self._by_id.get(case_id)
        if record is None:
            raise TraceabilityError("case ID is unknown")
        return record

    def validate_pytest_collection(self, collected_node_ids: Iterable[str]) -> None:
        try:
            collected = tuple(collected_node_ids)
        except Exception:
            raise TraceabilityError("pytest collection is invalid") from None
        if any(
            type(node_id) is not str or _PYTEST_NODE_RE.fullmatch(node_id) is None
            for node_id in collected
        ):
            raise TraceabilityError("pytest collection is invalid")
        if len(set(collected)) != len(collected):
            raise TraceabilityError("pytest collection is not unique")
        if set(collected) != self._active_pytest_nodes:
            raise TraceabilityError("pytest collection does not match active cases")


def _source_path_key(record: CaseRecord) -> str:
    return os.path.normcase(os.path.abspath(os.fspath(record.source_path)))


def _normalize_record(record: object) -> CaseRecord:
    try:
        if not isinstance(record, CaseRecord):
            raise TypeError("record type is invalid")
        with warnings.catch_warnings():
            warnings.simplefilter("ignore")
            document = record.model_dump(mode="python")
        return CaseRecord.model_validate(document)
    except Exception:
        raise TraceabilityError("case record is invalid") from None
