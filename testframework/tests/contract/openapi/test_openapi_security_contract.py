from __future__ import annotations

import pytest

from hub_testkit.contracts.openapi import (
    assert_magic_verify_error_contract,
    assert_openapi_canonical_document,
    assert_tms_webhook_auth_contract,
)


@pytest.mark.case_id("TC-OPENAPI-0001")
@pytest.mark.contract
def test_swagger_document_is_valid_with_single_api_prefix(contract_source_bundle: object) -> None:
    assert_openapi_canonical_document(contract_source_bundle)


@pytest.mark.case_id("TC-OPENAPI-0002")
@pytest.mark.contract
def test_magic_verify_operation_lists_implemented_error_contract(
    contract_source_bundle: object,
) -> None:
    assert_magic_verify_error_contract(contract_source_bundle)


@pytest.mark.case_id("TC-OPENAPI-0003")
@pytest.mark.contract
def test_tms_webhooks_document_real_auth_not_legacy_token(
    contract_source_bundle: object,
) -> None:
    assert_tms_webhook_auth_contract(contract_source_bundle)
