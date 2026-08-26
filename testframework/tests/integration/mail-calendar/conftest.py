from __future__ import annotations

from collections.abc import Iterator

import pytest

from hub_testkit.domains.mail_calendar_qa import MailCalendarQAClient
from hub_testkit.harness.mail_calendar_qa import (
    MailCalendarQAHarnessError,
    close_mail_calendar_qa_client,
    create_mail_calendar_qa_client_from_environment,
)


@pytest.fixture
def mail_calendar_qa_client() -> Iterator[MailCalendarQAClient]:
    try:
        client = create_mail_calendar_qa_client_from_environment()
    except MailCalendarQAHarnessError as error:
        raise pytest.UsageError(str(error)) from None
    try:
        yield client
    finally:
        try:
            close_mail_calendar_qa_client(client)
        except MailCalendarQAHarnessError as error:
            raise pytest.UsageError(str(error)) from None
