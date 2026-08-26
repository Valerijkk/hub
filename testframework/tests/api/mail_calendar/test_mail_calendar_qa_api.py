from __future__ import annotations

import pytest

from hub_testkit.domains.mail_calendar_qa import (
    FixtureSnapshot,
    MailCalendarQAClient,
    MailCalendarQAConfig,
    MailCalendarQAError,
)
from hub_testkit.harness.mail_calendar_qa import load_mail_calendar_qa_config_from_environment


_DENIED_QA_TOKEN = "f" * 64
_GENERIC_FAILURE = "mail calendar QA request failed"


@pytest.mark.case_id("TC-MAIL-CALENDAR-0001")
@pytest.mark.api
def test_mail_calendar_qa_fixture_control_plane_requires_capability_and_cleans_created_fixture(
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
            denied_client.create_fixture()

        assert str(denied.value) == _GENERIC_FAILURE
        assert denied.value.__cause__ is None
        assert denied.value.__suppress_context__ is True
    finally:
        denied_client.close()

    created: FixtureSnapshot | None = None
    try:
        created = mail_calendar_qa_client.create_fixture()

        assert type(created) is FixtureSnapshot
        assert (created.scenario, created.state) == ("inbound-request-v1", "created")
        assert created.owner_email == f"owner@qa-{created.nonce}.example.test"
    finally:
        if created is not None:
            mail_calendar_qa_client.delete_created_fixture(created)
