from __future__ import annotations

import pytest

from hub_testkit.contracts import contract_source_bundle, load_f1_04_risk_ledger

__all__ = ("contract_source_bundle",)


@pytest.fixture(autouse=True)
def verify_openapi_contract_traceability(
    request: pytest.FixtureRequest,
    contract_source_bundle: object,
) -> None:
    result = _traceability_valid_safe(request, contract_source_bundle)
    del request, contract_source_bundle
    if result is True:
        return
    del result
    raise pytest.UsageError("OpenAPI contract traceability is invalid") from None


def _traceability_valid_safe(
    request: pytest.FixtureRequest,
    bundle: object,
) -> bool:
    try:
        ledger = load_f1_04_risk_ledger(bundle)  # type: ignore[arg-type]
        node = request.node
        edge = ledger.edge_for_node_id(node.nodeid)
        properties = node.user_properties
        expected = (("hub.case_id", edge.case_id), ("hub.risk", edge.risk_id))
        actual = tuple(
            property_value
            for property_value in properties
            if type(property_value) is tuple
            and len(property_value) == 2
            and property_value[0] in {"hub.case_id", "hub.risk"}
        )
        return actual == expected
    except Exception:
        return False
