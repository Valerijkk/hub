from __future__ import annotations

import pytest

from hub_testkit.contracts.auth_events import (
    assert_auth_session_created_event_contract,
)


@pytest.mark.case_id("TC-SECURITY-0001")
@pytest.mark.contract
def test_auth_session_created_event_has_exact_producer_consumer_link(
    contract_source_bundle: object,
) -> None:
    assert_auth_session_created_event_contract(contract_source_bundle)
