from __future__ import annotations

import tomllib
from dataclasses import dataclass, field
from types import MappingProxyType
from typing import Mapping, NoReturn

from .source_bundle import ContractSourceBundle, ContractSourcePath


_SCHEMA = "hub.testframework.f1-security-risk-ledger/v1"
_ROOT_KEYS = ("schema", "edge")
_EDGE_KEYS = (
    "risk_id",
    "severity",
    "case_id",
    "pytest_node_id",
    "profile",
    "case_root",
)
_EXPECTED_EDGES = (
    (
        "RISK-OPENAPI-CANONICAL-DOCUMENT",
        "critical",
        "TC-OPENAPI-0001",
        "testframework/tests/contract/openapi/test_openapi_security_contract.py::test_swagger_document_is_valid_with_single_api_prefix",
        "openapi-contract",
        "testframework/cases/openapi",
    ),
    (
        "RISK-MAGIC-VERIFY-ERROR-CONTRACT",
        "high",
        "TC-OPENAPI-0002",
        "testframework/tests/contract/openapi/test_openapi_security_contract.py::test_magic_verify_operation_lists_implemented_error_contract",
        "openapi-contract",
        "testframework/cases/openapi",
    ),
    (
        "RISK-TMS-WEBHOOK-AUTH-DRIFT",
        "critical",
        "TC-OPENAPI-0003",
        "testframework/tests/contract/openapi/test_openapi_security_contract.py::test_tms_webhooks_document_real_auth_not_legacy_token",
        "openapi-contract",
        "testframework/cases/openapi",
    ),
    (
        "RISK-PUBLIC-ERROR-ARTIFACT",
        "high",
        "TC-PRIVACY-0001",
        "testframework/tests/contract/privacy/test_public_error_contract.py::test_public_error_artifact_is_redacted_to_wire_schema",
        "privacy-security",
        "testframework/cases/privacy",
    ),
    (
        "RISK-AUTH-SESSION-EVENT-LINK",
        "high",
        "TC-SECURITY-0001",
        "testframework/tests/contract/security/test_auth_session_event_contract.py::test_auth_session_created_event_has_exact_producer_consumer_link",
        "security-risk",
        "testframework/cases/security",
    ),
)
_INVALID = object()


class RiskLedgerError(ValueError):
    pass


@dataclass(frozen=True, slots=True)
class RiskLedgerEdge:
    risk_id: str
    severity: str
    case_id: str
    pytest_node_id: str
    profile: str
    case_root: str


@dataclass(frozen=True, slots=True, init=False)
class F1SecurityRiskLedger:
    edges: tuple[RiskLedgerEdge, ...]
    _by_node_id: Mapping[str, RiskLedgerEdge] = field(repr=False)

    @classmethod
    def _from_edges(cls, edges: tuple[RiskLedgerEdge, ...]) -> F1SecurityRiskLedger:
        instance = object.__new__(cls)
        object.__setattr__(instance, "edges", edges)
        object.__setattr__(
            instance,
            "_by_node_id",
            MappingProxyType({edge.pytest_node_id: edge for edge in edges}),
        )
        return instance

    def edge_for_node_id(self, node_id: str) -> RiskLedgerEdge:
        edge = _lookup_safe(self._by_node_id, node_id)
        del node_id
        if type(edge) is RiskLedgerEdge:
            return edge
        del edge
        _raise_invalid()


def load_f1_04_risk_ledger(bundle: ContractSourceBundle) -> F1SecurityRiskLedger:
    edges = _load_edges_safe(bundle)
    del bundle
    if type(edges) is tuple and all(type(edge) is RiskLedgerEdge for edge in edges):
        return F1SecurityRiskLedger._from_edges(edges)
    del edges
    _raise_invalid()


def _load_edges_safe(bundle: object) -> tuple[RiskLedgerEdge, ...] | object:
    try:
        if type(bundle) is not ContractSourceBundle:
            return _INVALID
        raw_policy = bundle.select(ContractSourcePath.F1_SECURITY_RISK_LEDGER)
        text = raw_policy.decode("utf-8", errors="strict")
        document = tomllib.loads(text)
        edges = _edges_from_document(document)
        del raw_policy, text, document
        return edges
    except Exception:
        return _INVALID


def _edges_from_document(document: object) -> tuple[RiskLedgerEdge, ...] | object:
    try:
        if type(document) is not dict or tuple(document.keys()) != _ROOT_KEYS:
            return _INVALID
        if document.get("schema") != _SCHEMA or type(document["schema"]) is not str:
            return _INVALID
        raw_edges = document.get("edge")
        if type(raw_edges) is not list or len(raw_edges) != len(_EXPECTED_EDGES):
            return _INVALID
        result: list[RiskLedgerEdge] = []
        for raw_edge, expected in zip(raw_edges, _EXPECTED_EDGES, strict=True):
            if type(raw_edge) is not dict or tuple(raw_edge.keys()) != _EDGE_KEYS:
                return _INVALID
            values = tuple(raw_edge.get(key) for key in _EDGE_KEYS)
            if any(type(value) is not str for value in values) or values != expected:
                return _INVALID
            result.append(RiskLedgerEdge(*values))
        return tuple(result)
    except Exception:
        return _INVALID


def _lookup_safe(
    by_node_id: Mapping[str, RiskLedgerEdge],
    node_id: object,
) -> RiskLedgerEdge | object:
    try:
        if type(node_id) is not str:
            return _INVALID
        edge = by_node_id.get(node_id, _INVALID)
        if type(edge) is not RiskLedgerEdge:
            return _INVALID
        return edge
    except Exception:
        return _INVALID


def _raise_invalid() -> NoReturn:
    raise RiskLedgerError("F1-04 risk ledger is invalid") from None
