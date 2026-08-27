from __future__ import annotations

from sys import exc_info

import pytest

from hub_testkit.domains.mail_calendar_qa import (
    AccountMaterializationFixtureSnapshot,
    MailCalendarQAClient,
)
from hub_testkit.harness.mail_calendar_qa import wait_for_mail_account_materialization


@pytest.mark.case_id("TC-MAIL-CALENDAR-0004")
@pytest.mark.integration
def test_account_materialization_eventually_finalizes_with_stable_public_roots(
    mail_calendar_qa_client: MailCalendarQAClient,
) -> None:
    created: AccountMaterializationFixtureSnapshot | None = None
    try:
        created = mail_calendar_qa_client.create_account_materialization_fixture()
        finalized = wait_for_mail_account_materialization(mail_calendar_qa_client, created.nonce)
        observed = mail_calendar_qa_client.get_account_materialization_fixture(created.nonce)

        assert type(finalized) is AccountMaterializationFixtureSnapshot
        assert type(observed) is AccountMaterializationFixtureSnapshot
        assert finalized.immutable_roots == created.immutable_roots
        assert observed.immutable_roots == created.immutable_roots
        assert finalized.materialization_state == "finalized"
        assert observed.materialization_state == "finalized"
        assert finalized.mail_account_id == created.expected_mail_account_id
        assert observed.mail_account_id == created.expected_mail_account_id
        assert finalized.primary_identity_id is not None
        assert observed.primary_identity_id == finalized.primary_identity_id
        assert finalized.primary_identity_status == "pending"
        assert observed.primary_identity_status == "pending"
    finally:
        if created is not None:
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
