from __future__ import annotations

import pytest

from hub_testkit.domains.mail_calendar_qa import FixtureSnapshot, MailCalendarQAClient
from hub_testkit.harness.mail_calendar_qa import wait_for_mail_calendar_correlation


@pytest.mark.case_id("TC-MAIL-CALENDAR-0002")
@pytest.mark.integration
def test_published_invite_projects_bidirectional_mail_calendar_correlation_and_rsvp(
    mail_calendar_qa_client: MailCalendarQAClient,
) -> None:
    created = mail_calendar_qa_client.create_fixture()
    assert type(created) is FixtureSnapshot
    assert created.state == "created"

    session = mail_calendar_qa_client.dev_login(created)
    published = mail_calendar_qa_client.publish_fixture(created)
    assert published.state == "published"

    mail_correlation = wait_for_mail_calendar_correlation(
        mail_calendar_qa_client,
        session,
        published.mail_message_id,
    )
    assert mail_correlation.message_id == published.mail_message_id
    assert mail_correlation.thread_id == published.mail_thread_id
    assert mail_correlation.calendar_event_id is not None

    calendar_correlation = mail_calendar_qa_client.calendar_event_detail(
        session,
        mail_correlation.calendar_event_id,
    )
    assert calendar_correlation.mail_message_id == published.mail_message_id
    assert calendar_correlation.can_rsvp is True
    assert calendar_correlation.viewer_attendee is not None
    assert calendar_correlation.viewer_attendee.rsvp_status == "needs_action"

    rsvp = mail_calendar_qa_client.respond_to_rsvp(
        session,
        calendar_correlation.event_id,
        "accepted",
        calendar_correlation.viewer_attendee.version,
        f"mail-calendar-rsvp-{published.nonce}",
    )
    assert rsvp.attendee_id == calendar_correlation.viewer_attendee.attendee_id
    assert rsvp.rsvp_status == "accepted"

    updated_calendar_correlation = mail_calendar_qa_client.calendar_event_detail(
        session,
        calendar_correlation.event_id,
    )
    assert updated_calendar_correlation.viewer_attendee is not None
    assert updated_calendar_correlation.viewer_attendee.rsvp_status == "accepted"
    assert updated_calendar_correlation.viewer_attendee.version == rsvp.version
