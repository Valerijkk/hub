from __future__ import annotations

from sys import exc_info

import pytest

from hub_testkit.domains.mail_calendar_qa import (
    AccountMaterializationFixtureSnapshot,
    MailCalendarQAClient,
    MailCalendarQAConfig,
    MailCalendarQAError,
)
from hub_testkit.harness.mail_calendar_qa import load_mail_calendar_qa_config_from_environment


_DENIED_QA_TOKEN = "f" * 64
_GENERIC_FAILURE = "mail calendar QA request failed"


@pytest.mark.case_id("TC-MAIL-CALENDAR-0001")
@pytest.mark.api
def test_mail_calendar_qa_account_materialization_requires_capability_and_cleans_fixture(
    mail_calendar_qa_client: MailCalendarQAClient,
) -> None:
    config = load_mail_calendar_qa_config_from_environment()
    if config.qa_token == _DENIED_QA_TOKEN:
        raise AssertionError("fixed QA denial token is configured")

    denied_client = MailCalendarQAClient(
        MailCalendarQAConfig(
            api_base_url=config.api_base_url,
            qa_token=_DENIED_QA_TOKEN,
        )
    )
    try:
        with pytest.raises(MailCalendarQAError) as denied:
            denied_client.create_account_materialization_fixture()

        assert str(denied.value) == _GENERIC_FAILURE
        assert denied.value.__cause__ is None
        assert denied.value.__suppress_context__ is True
    finally:
        denied_client.close()

    created: AccountMaterializationFixtureSnapshot | None = None
    deleted = False
    try:
        created = mail_calendar_qa_client.create_account_materialization_fixture()

        assert type(created) is AccountMaterializationFixtureSnapshot
        assert (created.scenario, created.state) == ("account-materialization-v1", "created")
        assert created.owner_email == f"owner@qa-materialization-{created.nonce}.example.test"
        assert created.materialization_state in {
            "not_materialized",
            "materialized_pending",
            "finalized",
        }
        if created.materialization_state == "not_materialized":
            assert (
                created.mail_account_id,
                created.primary_identity_id,
                created.primary_identity_status,
            ) == (None, None, None)
        else:
            assert created.mail_account_id == created.expected_mail_account_id
            assert created.primary_identity_id is not None
            assert created.primary_identity_status == "pending"

        mail_calendar_qa_client.quiesce_account_materialization_fixture(created.nonce)
        mail_calendar_qa_client.delete_account_materialization_fixture(created.nonce)
        deleted = True
    finally:
        if created is not None and not deleted:
            cleanup_error: Exception | None = None
            try:
                mail_calendar_qa_client.quiesce_account_materialization_fixture(created.nonce)
            except Exception as error:
                cleanup_error = error
            else:
                try:
                    mail_calendar_qa_client.delete_account_materialization_fixture(created.nonce)
                except Exception as error:
                    cleanup_error = error
            if cleanup_error is not None and exc_info()[0] is None:
                raise cleanup_error
