"""Environment-bound lifecycle for the narrow Mail and Calendar QA client."""

from __future__ import annotations

from os import environ

from hub_testkit.domains.mail_calendar_qa import (
    MailCalendarQAClient,
    MailCalendarQAConfig,
    MailCalendarQAError,
)


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
