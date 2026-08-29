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


@pytest.mark.case_id("TC-MAIL-FTS-0002")
@pytest.mark.api
def test_mail_thread_search_scopes_results_and_rejects_legacy_search_route() -> None:
    client: MailCalendarQAClient | None = None
    owner_fixture: ThreadSearchFixtureSnapshot | None = None
    foreign_fixture: ThreadSearchFixtureSnapshot | None = None
    try:
        client = create_mail_calendar_qa_client_from_environment()
        owner_fixture = client.create_thread_search_fixture()
        foreign_fixture = client.create_thread_search_fixture()
        session = client.dev_login_thread_search_fixture(owner_fixture)

        projection = client.list_mail_threads(session, search="qa calendar")
        returned_thread_ids = tuple(item.thread_id for item in projection.items)
        foreign_thread_ids = (
            *foreign_fixture.matching_thread_ids,
            *foreign_fixture.decoy_thread_ids,
        )

        assert returned_thread_ids == owner_fixture.matching_thread_ids
        assert not any(
            thread_id in owner_fixture.decoy_thread_ids for thread_id in returned_thread_ids
        )
        assert not any(thread_id in foreign_thread_ids for thread_id in returned_thread_ids)
        assert tuple(item.is_pinned for item in projection.items) == (True, True, False, False)
        assert projection.next_cursor is None

        starred_projection = client.list_mail_threads(
            session,
            search="qa calendar",
            starred=True,
        )
        starred_thread_ids = tuple(item.thread_id for item in starred_projection.items)

        assert starred_thread_ids == owner_fixture.matching_thread_ids[:2]
        assert all(item.is_pinned for item in starred_projection.items)
        assert starred_projection.next_cursor is None
        assert client.legacy_mail_search_status(session) == 404
    finally:
        cleanup_error: Exception | None = None
        if client is not None and owner_fixture is not None:
            try:
                client.delete_thread_search_fixture(owner_fixture)
            except Exception as error:
                cleanup_error = error
        if client is not None and foreign_fixture is not None:
            try:
                client.delete_thread_search_fixture(foreign_fixture)
            except Exception as error:
                if cleanup_error is None:
                    cleanup_error = error
        if client is not None:
            try:
                close_mail_calendar_qa_client(client)
            except Exception as error:
                if cleanup_error is None:
                    cleanup_error = error
        if cleanup_error is not None and exc_info()[0] is None:
            raise cleanup_error
