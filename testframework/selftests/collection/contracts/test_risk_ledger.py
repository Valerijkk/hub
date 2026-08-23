from __future__ import annotations

import hashlib
import traceback
from dataclasses import FrozenInstanceError

import pytest

from hub_testkit.contracts import (
    ContractSourceBundle,
    ContractSourceError,
    ContractSourcePath,
    SourceBlob,
    SourceSnapshot,
)
from hub_testkit.contracts.risk_ledger import RiskLedgerError, load_f1_04_risk_ledger


POLICY = """schema = "hub.testframework.f1-security-risk-ledger/v1"

[[edge]]
risk_id = "RISK-OPENAPI-CANONICAL-DOCUMENT"
severity = "critical"
case_id = "TC-OPENAPI-0001"
pytest_node_id = "testframework/tests/contract/openapi/test_openapi_security_contract.py::test_swagger_document_is_valid_with_single_api_prefix"
profile = "openapi-contract"
case_root = "testframework/cases/openapi"

[[edge]]
risk_id = "RISK-MAGIC-VERIFY-ERROR-CONTRACT"
severity = "high"
case_id = "TC-OPENAPI-0002"
pytest_node_id = "testframework/tests/contract/openapi/test_openapi_security_contract.py::test_magic_verify_operation_lists_implemented_error_contract"
profile = "openapi-contract"
case_root = "testframework/cases/openapi"

[[edge]]
risk_id = "RISK-TMS-WEBHOOK-AUTH-DRIFT"
severity = "critical"
case_id = "TC-OPENAPI-0003"
pytest_node_id = "testframework/tests/contract/openapi/test_openapi_security_contract.py::test_tms_webhooks_document_real_auth_not_legacy_token"
profile = "openapi-contract"
case_root = "testframework/cases/openapi"

[[edge]]
risk_id = "RISK-PUBLIC-ERROR-ARTIFACT"
severity = "high"
case_id = "TC-PRIVACY-0001"
pytest_node_id = "testframework/tests/contract/privacy/test_public_error_contract.py::test_public_error_artifact_is_redacted_to_wire_schema"
profile = "privacy-security"
case_root = "testframework/cases/privacy"

[[edge]]
risk_id = "RISK-AUTH-SESSION-EVENT-LINK"
severity = "high"
case_id = "TC-SECURITY-0001"
pytest_node_id = "testframework/tests/contract/security/test_auth_session_event_contract.py::test_auth_session_created_event_has_exact_producer_consumer_link"
profile = "security-risk"
case_root = "testframework/cases/security"
""".encode("utf-8")


def _blob(path: ContractSourcePath, content: bytes) -> SourceBlob:
    return SourceBlob(path=path, content=content, sha256=hashlib.sha256(content).hexdigest())


def _bundle(policy: bytes = POLICY) -> ContractSourceBundle:
    blobs = tuple(
        _blob(
            path,
            policy
            if path is ContractSourcePath.F1_SECURITY_RISK_LEDGER
            else f"source for {path.value}\n".encode("utf-8"),
        )
        for path in ContractSourcePath
    )
    snapshot = SourceSnapshot(
        source_commit="0123456789abcdef0123456789abcdef01234567",
        blobs=blobs,
    )
    return ContractSourceBundle(snapshot)


def _ledger_error(bundle: object) -> RiskLedgerError:
    try:
        load_f1_04_risk_ledger(bundle)  # type: ignore[arg-type]
    except RiskLedgerError as error:
        return error
    raise AssertionError("invalid ledger input was accepted")


def test_risk_ledger_accepts_only_five_ordered_literal_edges_and_lookup() -> None:
    ledger = load_f1_04_risk_ledger(_bundle())

    assert tuple(edge.risk_id for edge in ledger.edges) == (
        "RISK-OPENAPI-CANONICAL-DOCUMENT",
        "RISK-MAGIC-VERIFY-ERROR-CONTRACT",
        "RISK-TMS-WEBHOOK-AUTH-DRIFT",
        "RISK-PUBLIC-ERROR-ARTIFACT",
        "RISK-AUTH-SESSION-EVENT-LINK",
    )
    assert tuple(edge.severity for edge in ledger.edges) == (
        "critical",
        "high",
        "critical",
        "high",
        "high",
    )
    edge = ledger.edge_for_node_id(
        "testframework/tests/contract/privacy/test_public_error_contract.py::test_public_error_artifact_is_redacted_to_wire_schema"
    )
    assert edge.case_id == "TC-PRIVACY-0001"
    assert edge.profile == "privacy-security"
    assert edge.case_root == "testframework/cases/privacy"
    with pytest.raises(RiskLedgerError):
        ledger.edge_for_node_id("testframework/tests/contract/openapi/test_unknown.py::test_unknown")


def test_risk_ledger_rejects_root_edge_order_count_key_and_value_deviations() -> None:
    first_edge = POLICY.split(b"[[edge]]", 2)[1]
    reordered = (
        b"[[edge]]"
        + first_edge
        + b"\nschema = \"hub.testframework.f1-security-risk-ledger/v1\"\n"
    )
    invalid_policies = (
        POLICY.replace(b"hub.testframework.f1-security-risk-ledger/v1", b"other/v1"),
        POLICY.replace(b"schema = \"hub.testframework.f1-security-risk-ledger/v1\"\n", b""),
        POLICY.replace(b"[[edge]]", b"[[edge]]\nunexpected = \"forbidden\"", 1),
        POLICY.replace(b"severity = \"critical\"", b"severity = \"high\"", 1),
        POLICY.replace(b"case_root = \"testframework/cases/openapi\"", b"case_root = \"cases/openapi\"", 1),
        POLICY.replace(b"[[edge]]", b"", 1),
        reordered,
    )

    for policy in invalid_policies:
        with pytest.raises(RiskLedgerError) as error:
            load_f1_04_risk_ledger(_bundle(policy))
        assert str(error.value) == "F1-04 risk ledger is invalid"
        assert error.value.__cause__ is None
        assert error.value.__suppress_context__ is True


def test_risk_ledger_is_immutable_and_never_exposes_policy_bytes() -> None:
    canary = b"risk-ledger-policy-canary"
    ledger = load_f1_04_risk_ledger(_bundle(POLICY + b"# risk-ledger-policy-canary\n"))

    with pytest.raises(FrozenInstanceError):
        ledger.edges = ()  # type: ignore[misc]
    assert canary.decode("ascii") not in repr(ledger)
    assert canary.decode("ascii") not in str(ledger)


class _HostileBundle:
    canary = "risk-ledger-hostile-canary"

    def select(self, selector: object) -> bytes:
        raise RuntimeError(self.canary)


def test_risk_ledger_suppresses_hostile_bundle_exception_and_canary() -> None:
    error = _ledger_error(_HostileBundle())
    formatted = "".join(traceback.format_exception(error))

    assert str(error) == "F1-04 risk ledger is invalid"
    assert error.__cause__ is None
    assert error.__suppress_context__ is True
    assert _HostileBundle.canary not in formatted


def test_source_bundle_rejects_invalid_policy_bytes_before_ledger_loading() -> None:
    with pytest.raises(ContractSourceError):
        _bundle(b"\xef\xbb\xbfschema = \"hub.testframework.f1-security-risk-ledger/v1\"\n")
