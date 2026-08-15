from __future__ import annotations

import pytest

from hub_testkit.contracts.public_error import (
    assert_public_error_artifact_contract,
)


@pytest.mark.case_id("TC-PRIVACY-0001")
@pytest.mark.contract
def test_public_error_artifact_is_redacted_to_wire_schema(
    contract_source_bundle: object,
) -> None:
    assert_public_error_artifact_contract(contract_source_bundle)
