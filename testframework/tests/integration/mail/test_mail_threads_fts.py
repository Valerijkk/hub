from __future__ import annotations

from sys import exc_info

import pytest

from hub_testkit.domains.mail_calendar_qa import (
    MailCalendarQAClient,
    ThreadSearchFixtureSnapshot,
)
from hub_testkit.harness.mail_calendar_qa import (
    close_mail_calendar_qa_client,
    create_mail_calendar_qa_client_from_environment,
)


@pytest.mark.case_id("TC-MAIL-FTS-0001")
@pytest.mark.integration
def test_mail_thread_search_returns_only_ordered_matching_threads() -> None:
    client: MailCalendarQAClient | None = None
    fixture: ThreadSearchFixtureSnapshot | None = None
    try:
        client = create_mail_calendar_qa_client_from_environment()
        fixture = client.create_thread_search_fixture()
        session = client.dev_login_thread_search_fixture(fixture)

        projection = client.list_mail_threads(session, search="qa calendar")
        returned_thread_ids = tuple(item.thread_id for item in projection.items)

        assert returned_thread_ids == fixture.matching_thread_ids
        assert not any(
            thread_id in fixture.decoy_thread_ids for thread_id in returned_thread_ids
        )
        assert projection.next_cursor is None
    finally:
        cleanup_error: Exception | None = None
        if client is not None and fixture is not None:
            try:
                client.delete_thread_search_fixture(fixture)
            except Exception as error:
                cleanup_error = error
        if client is not None:
            try:
                close_mail_calendar_qa_client(client)
            except Exception as error:
                if cleanup_error is None:
                    cleanup_error = error
        if cleanup_error is not None and exc_info()[0] is None:
            raise cleanup_error
