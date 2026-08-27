"""Environment-bound lifecycle for the narrow Mail and Calendar QA client."""

from __future__ import annotations

from os import environ
from time import monotonic, sleep
from typing import Final

from hub_testkit.domains.mail_calendar_qa import (
    AccountMaterializationFixtureSnapshot,
    DevSession,
    MailCalendarQAClient,
    MailCalendarQAConfig,
    MailCalendarQAError,
    MailCorrelation,
)


_CORRELATION_TIMEOUT_SECONDS: Final = 30.0
_CORRELATION_POLL_INTERVAL_SECONDS: Final = 0.25
_CORRELATION_TIMEOUT_FAILURE: Final = "mail calendar QA request failed"


class MailCalendarQAHarnessError(ValueError):
    """Safe setup or cleanup failure for the Mail and Calendar QA fixture."""


def load_mail_calendar_qa_config_from_environment() -> MailCalendarQAConfig:
    """Read the two launcher-owned values that bind the closed QA control plane."""
    try:
        return MailCalendarQAConfig(
            api_base_url=environ["HUB_QA_API_BASE_URL"],
            qa_token=environ["HUB_QA_MAIL_CALENDAR_TOKEN"],
        )
    except Exception:
        raise MailCalendarQAHarnessError("mail calendar QA harness is unavailable") from None


def create_mail_calendar_qa_client_from_environment() -> MailCalendarQAClient:
    """Construct one typed client from the closed launcher-owned environment."""
    try:
        return MailCalendarQAClient(load_mail_calendar_qa_config_from_environment())
    except MailCalendarQAError:
        raise MailCalendarQAHarnessError("mail calendar QA harness is unavailable") from None


def close_mail_calendar_qa_client(client: MailCalendarQAClient) -> None:
    """Close one client without allowing cleanup internals to escape the fixture."""
    try:
        if type(client) is not MailCalendarQAClient:
            raise ValueError
        client.close()
    except Exception:
        raise MailCalendarQAHarnessError("mail calendar QA harness cleanup failed") from None


def wait_for_mail_calendar_correlation(
    client: MailCalendarQAClient,
    session: DevSession,
    message_id: str,
) -> MailCorrelation:
    """Boundedly wait until a published fixture's mail message projects an event."""
    deadline = monotonic() + _CORRELATION_TIMEOUT_SECONDS
    while True:
        correlation = client.mail_message_detail(session, message_id)
        if correlation.calendar_event_id is not None:
            return correlation

        remaining_seconds = deadline - monotonic()
        if remaining_seconds <= 0:
            raise MailCalendarQAError(_CORRELATION_TIMEOUT_FAILURE)
        sleep(min(_CORRELATION_POLL_INTERVAL_SECONDS, remaining_seconds))


def wait_for_mail_account_materialization(
    client: MailCalendarQAClient,
    nonce: str,
) -> AccountMaterializationFixtureSnapshot:
    """Wait for a final fixture observation within the monotonic polling window."""
    deadline = monotonic() + _CORRELATION_TIMEOUT_SECONDS
    while True:
        if monotonic() >= deadline:
            raise MailCalendarQAError(_CORRELATION_TIMEOUT_FAILURE)
        fixture = client.get_account_materialization_fixture(nonce)
        remaining_seconds = deadline - monotonic()
        if remaining_seconds <= 0:
            raise MailCalendarQAError(_CORRELATION_TIMEOUT_FAILURE)
        if fixture.materialization_state == "finalized":
            return fixture

        sleep(min(_CORRELATION_POLL_INTERVAL_SECONDS, remaining_seconds))
