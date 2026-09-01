from __future__ import annotations

from sys import exc_info

import pytest

from hub_testkit.domains.mail_calendar_qa import (
    MailCalendarQAClient,
    ThreadSearchDeleteObservation,
    ThreadSearchFixtureSnapshot,
)
from hub_testkit.harness.mail_calendar_qa import (
    close_mail_calendar_qa_client,
    create_mail_calendar_qa_client_from_environment,
)


@pytest.mark.case_id("TC-MAIL-FTS-0003")
@pytest.mark.integration
def test_mail_thread_fixture_delete_guard_stops_atlas_set_null_before_delete() -> None:
    client: MailCalendarQAClient | None = None
    fixture: ThreadSearchFixtureSnapshot | None = None
    armed = False
    observation: ThreadSearchDeleteObservation | None = None
    try:
        client = create_mail_calendar_qa_client_from_environment()
        fixture = client.create_thread_search_fixture()
        client.arm_thread_search_fixture_delete_blocker(
            fixture,
            "atlas-user-set-null-v1",
        )
        armed = True

        observation = client.observe_thread_search_fixture_delete(fixture)

        assert observation is ThreadSearchDeleteObservation.PREDELETE
    finally:
        cleanup_error: Exception | None = None
        if client is not None and fixture is not None:
            if armed:
                try:
                    client.disarm_thread_search_fixture_delete_blocker(
                        fixture,
                        "atlas-user-set-null-v1",
                    )
                except Exception as error:
                    cleanup_error = error
            if observation is not ThreadSearchDeleteObservation.DELETED:
                try:
                    client.delete_thread_search_fixture(fixture)
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


@pytest.mark.case_id("TC-MAIL-FTS-0004")
@pytest.mark.integration
def test_mail_thread_fixture_delete_guard_stops_future_user_cascade_before_delete() -> None:
    client: MailCalendarQAClient | None = None
    fixture: ThreadSearchFixtureSnapshot | None = None
    armed = False
    observation: ThreadSearchDeleteObservation | None = None
    try:
        client = create_mail_calendar_qa_client_from_environment()
        fixture = client.create_thread_search_fixture()
        client.arm_thread_search_fixture_delete_blocker(
            fixture,
            "ent-user-cascade-v1",
        )
        armed = True

        observation = client.observe_thread_search_fixture_delete(fixture)

        assert observation is ThreadSearchDeleteObservation.PREDELETE
    finally:
        cleanup_error: Exception | None = None
        if client is not None and fixture is not None:
            if armed:
                try:
                    client.disarm_thread_search_fixture_delete_blocker(
                        fixture,
                        "ent-user-cascade-v1",
                    )
                except Exception as error:
                    cleanup_error = error
            if observation is not ThreadSearchDeleteObservation.DELETED:
                try:
                    client.delete_thread_search_fixture(fixture)
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


@pytest.mark.case_id("TC-MAIL-FTS-0005")
@pytest.mark.integration
def test_mail_thread_fixture_delete_guard_stops_extra_relation_by_target_message() -> None:
    client: MailCalendarQAClient | None = None
    target_fixture: ThreadSearchFixtureSnapshot | None = None
    peer_fixture: ThreadSearchFixtureSnapshot | None = None
    armed = False
    observation: ThreadSearchDeleteObservation | None = None
    try:
        client = create_mail_calendar_qa_client_from_environment()
        target_fixture = client.create_thread_search_fixture()
        peer_fixture = client.create_thread_search_fixture()
        client.arm_thread_search_fixture_delete_peer_blocker(
            target_fixture,
            peer_fixture,
            "mail-message-folder-by-message-extra-v1",
        )
        armed = True

        observation = client.observe_thread_search_fixture_delete(target_fixture)

        assert observation is ThreadSearchDeleteObservation.PREDELETE
    finally:
        cleanup_error: Exception | None = None
        if client is not None:
            if armed and target_fixture is not None and peer_fixture is not None:
                try:
                    client.disarm_thread_search_fixture_delete_peer_blocker(
                        target_fixture,
                        peer_fixture,
                        "mail-message-folder-by-message-extra-v1",
                    )
                except Exception as error:
                    cleanup_error = error
            if (
                target_fixture is not None
                and observation is not ThreadSearchDeleteObservation.DELETED
            ):
                try:
                    client.delete_thread_search_fixture(target_fixture)
                except Exception as error:
                    if cleanup_error is None:
                        cleanup_error = error
            if peer_fixture is not None:
                try:
                    client.delete_thread_search_fixture(peer_fixture)
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


@pytest.mark.case_id("TC-MAIL-FTS-0006")
@pytest.mark.integration
def test_mail_thread_fixture_delete_guard_stops_extra_relation_by_target_folder() -> None:
    client: MailCalendarQAClient | None = None
    target_fixture: ThreadSearchFixtureSnapshot | None = None
    peer_fixture: ThreadSearchFixtureSnapshot | None = None
    armed = False
    observation: ThreadSearchDeleteObservation | None = None
    try:
        client = create_mail_calendar_qa_client_from_environment()
        target_fixture = client.create_thread_search_fixture()
        peer_fixture = client.create_thread_search_fixture()
        client.arm_thread_search_fixture_delete_peer_blocker(
            target_fixture,
            peer_fixture,
            "mail-message-folder-by-folder-extra-v1",
        )
        armed = True

        observation = client.observe_thread_search_fixture_delete(target_fixture)

        assert observation is ThreadSearchDeleteObservation.PREDELETE
    finally:
        cleanup_error: Exception | None = None
        if client is not None:
            if armed and target_fixture is not None and peer_fixture is not None:
                try:
                    client.disarm_thread_search_fixture_delete_peer_blocker(
                        target_fixture,
                        peer_fixture,
                        "mail-message-folder-by-folder-extra-v1",
                    )
                except Exception as error:
                    cleanup_error = error
            if (
                target_fixture is not None
                and observation is not ThreadSearchDeleteObservation.DELETED
            ):
                try:
                    client.delete_thread_search_fixture(target_fixture)
                except Exception as error:
                    if cleanup_error is None:
                        cleanup_error = error
            if peer_fixture is not None:
                try:
                    client.delete_thread_search_fixture(peer_fixture)
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
