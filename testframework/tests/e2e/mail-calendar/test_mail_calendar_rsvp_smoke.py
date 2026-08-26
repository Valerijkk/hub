from __future__ import annotations

import json
from os import environ

import pytest
from playwright.sync_api import Page, expect

from hub_testkit.harness.mail_calendar_qa import (
    MailCalendarQAHarnessError,
    close_mail_calendar_qa_client,
    create_mail_calendar_qa_client_from_environment,
    wait_for_mail_calendar_correlation,
)


_QA_FRONTEND_BASE_URL = "http://qa-frontend:8080"


def _qa_frontend_base_url_from_environment() -> str:
    try:
        value = environ["HUB_QA_FRONTEND_BASE_URL"]
    except Exception:
        raise pytest.UsageError("mail calendar E2E harness is unavailable") from None
    if type(value) is not str or value != _QA_FRONTEND_BASE_URL:
        raise pytest.UsageError("mail calendar E2E harness is unavailable")
    return value


@pytest.mark.case_id("TC-MAIL-CALENDAR-0003")
@pytest.mark.e2e
@pytest.mark.smoke
def test_mail_origin_invite_opens_event_updates_rsvp_and_returns_to_mail(page: Page) -> None:
    frontend_base_url = _qa_frontend_base_url_from_environment()
    try:
        client = create_mail_calendar_qa_client_from_environment()
    except MailCalendarQAHarnessError as error:
        raise pytest.UsageError(str(error)) from None

    try:
        created = client.create_fixture()
        session = client.dev_login(created)
        published = client.publish_fixture(created)
        mail_correlation = wait_for_mail_calendar_correlation(
            client,
            session,
            published.mail_message_id,
        )
        event_id = mail_correlation.calendar_event_id
        assert event_id is not None

        auth_bootstrap = {
            "accessToken": session.access_token,
            "refreshToken": session.refresh_token,
            "user": dict(session.user),
            "isAuthenticated": True,
        }
        auth_bootstrap_json = json.dumps(
            auth_bootstrap,
            ensure_ascii=True,
            separators=(",", ":"),
        )
        page.add_init_script(
            f"window.__HUB_AUTH_BOOTSTRAP__ = {auth_bootstrap_json};",
        )
        mail_thread_url = (
            f"{frontend_base_url}/mail/all/thread/{published.mail_thread_id}"
        )
        page.goto(f"{frontend_base_url}/mail/message/{published.mail_message_id}")
        expect(page).to_have_url(mail_thread_url)

        calendar_link = page.locator(f'a[href="/calendar/events/{event_id}"]')
        expect(calendar_link).to_be_visible()
        calendar_link.click()
        expect(page).to_have_url(f"{frontend_base_url}/calendar/events/{event_id}")

        accept_rsvp = page.get_by_test_id("calendar-rsvp-accepted")
        expect(accept_rsvp).to_be_visible()
        accept_rsvp.click()
        expect(accept_rsvp).to_have_attribute("aria-pressed", "true")

        related_mail_link = page.locator(
            f'a[href="/mail/message/{published.mail_message_id}"]'
        )
        expect(related_mail_link).to_be_visible()
        related_mail_link.click()
        expect(page).to_have_url(mail_thread_url)
    finally:
        try:
            close_mail_calendar_qa_client(client)
        except MailCalendarQAHarnessError as error:
            raise pytest.UsageError(str(error)) from None
