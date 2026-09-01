"""No-leak HTTP client for the narrow Mail and Calendar QA runtime seam."""

from __future__ import annotations

from enum import Enum
import json
import time
from typing import Final, NoReturn
from unicodedata import category, normalize

import httpx

from .models import (
    AccountMaterializationFixtureSnapshot,
    CalendarCorrelation,
    DevSession,
    FixtureSnapshot,
    MailCalendarQAConfig,
    MailCalendarQAError,
    MailCorrelation,
    MailThreadProjectionPage,
    RSVPResult,
    ThreadSearchFixtureSnapshot,
    _canonical_uuid,
    _nonce,
    _positive_int,
)
from .mail_calendar_faults import (
    PublishedCleanupFaultKind,
    _published_cleanup_fault_fixture,
    _published_cleanup_fault_kind,
)


_FAILURE_TEXT: Final = "mail calendar QA request failed"
_CREATE_FIXTURE_BODY: Final = b'{"scenario":"inbound-request-v1"}'
_CREATE_ACCOUNT_MATERIALIZATION_FIXTURE_BODY: Final = b'{"scenario":"account-materialization-v1"}'
_CREATE_THREAD_SEARCH_FIXTURE_BODY: Final = b'{"scenario":"thread-search-v1"}'
_RSVP_INPUT_STATUSES: Final = frozenset({"accepted", "declined", "tentative"})
_MAX_RESPONSE_BYTES: Final = 262_144
_MAX_THREAD_SEARCH_BYTES: Final = 128
_THREAD_SEARCH_DELETE_GUARD_HEADER: Final = "X-Hub-QA-Fixture-Delete-Guard"
_THREAD_SEARCH_DELETE_GUARD_PREDELETE: Final = "predelete"
_THREAD_SEARCH_DELETE_CONFLICT_BODY: Final = {"error": "fixture lifecycle conflict"}
_THREAD_SEARCH_DELETE_SINGLE_BLOCKER_KINDS: Final = frozenset(
    {"atlas-user-set-null-v1", "ent-user-cascade-v1"}
)
_THREAD_SEARCH_DELETE_PEER_BLOCKER_KINDS: Final = frozenset(
    {
        "mail-message-folder-by-message-extra-v1",
        "mail-message-folder-by-folder-extra-v1",
    }
)


class ThreadSearchDeleteObservation(str, Enum):
    """Closed safe observations for a thread-search fixture delete request."""

    DELETED = "deleted"
    PREDELETE = "predelete"
    CONFLICT_UNMARKED = "conflict_unmarked"


class MailCalendarQAClient:
    """Owns one fixed-base client and exposes the frozen QA workflow only."""

    __slots__ = ("_client", "_closed", "_config")

    def __init__(self, config: MailCalendarQAConfig) -> None:
        if type(config) is not MailCalendarQAConfig:
            raise _failure()
        try:
            self._config = config
            self._client = httpx.Client(
                base_url=config.api_base_url,
                follow_redirects=False,
                timeout=httpx.Timeout(connect=2.0, read=5.0, write=5.0, pool=2.0),
                trust_env=False,
            )
            self._closed = False
        except Exception:
            raise _failure() from None

    def create_fixture(self) -> FixtureSnapshot:
        try:
            self._require_open()
            response = self._client.request(
                "POST",
                "/api/_qa/mail-calendar/fixtures",
                content=_CREATE_FIXTURE_BODY,
                headers={"Content-Type": "application/json", "X-Hub-QA-Token": self._config.qa_token},
            )
            fixture = FixtureSnapshot.from_payload(_json_response(response, status_code=201))
            if fixture.state != "created":
                raise ValueError
            return fixture
        except Exception:
            raise _failure() from None

    def create_account_materialization_fixture(self) -> AccountMaterializationFixtureSnapshot:
        try:
            self._require_open()
            response = self._client.request(
                "POST",
                "/api/_qa/mail-calendar/fixtures",
                content=_CREATE_ACCOUNT_MATERIALIZATION_FIXTURE_BODY,
                headers={"Content-Type": "application/json", "X-Hub-QA-Token": self._config.qa_token},
            )
            return AccountMaterializationFixtureSnapshot.from_payload(_json_response(response, status_code=201))
        except Exception:
            raise _failure() from None

    def create_thread_search_fixture(self) -> ThreadSearchFixtureSnapshot:
        try:
            self._require_open()
            response = self._client.request(
                "POST",
                "/api/_qa/mail-calendar/fixtures",
                content=_CREATE_THREAD_SEARCH_FIXTURE_BODY,
                headers={"Content-Type": "application/json", "X-Hub-QA-Token": self._config.qa_token},
            )
            fixture = ThreadSearchFixtureSnapshot.from_payload(_json_response(response, status_code=201))
            if fixture.state != "created":
                raise ValueError
            return fixture
        except Exception:
            raise _failure() from None

    def get_account_materialization_fixture(self, nonce: str) -> AccountMaterializationFixtureSnapshot:
        try:
            fixture_nonce = _nonce(nonce)
            self._require_open()
            response = self._client.request(
                "GET",
                f"/api/_qa/mail-calendar/fixtures/{fixture_nonce}/account-materialization",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            fixture = AccountMaterializationFixtureSnapshot.from_payload(_json_response(response, status_code=200))
            if fixture.nonce != fixture_nonce:
                raise ValueError
            return fixture
        except Exception:
            raise _failure() from None

    def quiesce_account_materialization_fixture(self, nonce: str) -> None:
        try:
            fixture_nonce = _nonce(nonce)
            self._require_open()
            response = self._client.request(
                "POST",
                f"/api/_qa/mail-calendar/fixtures/{fixture_nonce}/account-materialization/quiesce",
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            if response.status_code != 204 or response.content:
                raise ValueError
        except Exception:
            raise _failure() from None

    def delete_account_materialization_fixture(self, nonce: str) -> None:
        try:
            fixture_nonce = _nonce(nonce)
            self._require_open()
            response = self._client.request(
                "DELETE",
                f"/api/_qa/mail-calendar/fixtures/{fixture_nonce}/account-materialization",
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            if response.status_code != 204 or response.content:
                raise ValueError
        except Exception:
            raise _failure() from None

    def publish_fixture(self, created_fixture: FixtureSnapshot) -> FixtureSnapshot:
        try:
            fixture = _created_fixture(created_fixture)
            self._require_open()
            response = self._client.request(
                "POST",
                f"/api/_qa/mail-calendar/fixtures/{fixture.nonce}/publish",
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            published = FixtureSnapshot.from_payload(_json_response(response, status_code=200))
            if published.state != "published" or published.immutable_identity != fixture.immutable_identity:
                raise ValueError
            return published
        except Exception:
            raise _failure() from None

    def delete_created_fixture(self, created_fixture: FixtureSnapshot) -> None:
        try:
            fixture = _created_fixture(created_fixture)
            self._require_open()
            response = self._client.request(
                "DELETE",
                f"/api/_qa/mail-calendar/fixtures/{fixture.nonce}",
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            if response.status_code != 204 or response.content:
                raise ValueError
        except Exception:
            raise _failure() from None

    def delete_published_fixture(self, published_fixture: FixtureSnapshot) -> None:
        try:
            fixture = _published_fixture(published_fixture)
            fixture_nonce = _nonce(fixture.nonce)
            self._require_open()
            response = self._client.request(
                "DELETE",
                f"/api/_qa/mail-calendar/fixtures/{fixture_nonce}",
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            if response.status_code != 204 or response.content:
                raise ValueError
        except Exception:
            raise _failure() from None

    def delete_published_fixture_when_quiescent(self, published_fixture: FixtureSnapshot) -> None:
        try:
            fixture = _published_fixture(published_fixture)
            fixture_nonce = _nonce(fixture.nonce)
            self._require_open()
            for attempt in range(20):
                response = self._client.request(
                    "DELETE",
                    f"/api/_qa/mail-calendar/fixtures/{fixture_nonce}",
                    content=b"",
                    headers={"X-Hub-QA-Token": self._config.qa_token},
                )
                if response.status_code == 204 and not response.content:
                    return
                if response.status_code != 409 or attempt == 19:
                    raise ValueError
                time.sleep(0.025)
        except Exception:
            raise _failure() from None

    def arm_published_cleanup_fault(
        self,
        fixture: FixtureSnapshot,
        kind: PublishedCleanupFaultKind,
    ) -> None:
        try:
            fault_kind = _published_cleanup_fault_kind(kind)
            fault_fixture = _published_cleanup_fault_fixture(fixture, fault_kind)
            self._require_open()
            response = self._client.request(
                "POST",
                _published_cleanup_fault_path(fault_fixture.nonce, fault_kind),
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            if not _published_cleanup_fault_empty_response(response, allow_missing=False):
                raise ValueError
        except Exception:
            raise _failure() from None

    def disarm_published_cleanup_fault(
        self,
        fixture: FixtureSnapshot,
        kind: PublishedCleanupFaultKind,
    ) -> bool:
        try:
            fault_kind = _published_cleanup_fault_kind(kind)
            fault_fixture = _published_cleanup_fault_fixture(fixture, fault_kind)
            self._require_open()
            response = self._client.request(
                "DELETE",
                _published_cleanup_fault_path(fault_fixture.nonce, fault_kind),
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            return _published_cleanup_fault_empty_response(response, allow_missing=True)
        except Exception:
            raise _failure() from None

    def delete_thread_search_fixture(self, fixture: ThreadSearchFixtureSnapshot) -> None:
        try:
            search_fixture = _created_thread_search_fixture(fixture)
            self._require_open()
            response = self._client.request(
                "DELETE",
                f"/api/_qa/mail-calendar/fixtures/{search_fixture.nonce}",
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            if response.status_code != 204 or response.content:
                raise ValueError
        except Exception:
            raise _failure() from None

    def arm_thread_search_fixture_delete_blocker(
        self,
        fixture: ThreadSearchFixtureSnapshot,
        kind: str,
    ) -> None:
        try:
            search_fixture = _created_thread_search_fixture(fixture)
            blocker_kind = _thread_search_delete_single_blocker_kind(kind)
            self._require_open()
            response = self._client.request(
                "POST",
                _thread_search_delete_single_blocker_path(search_fixture.nonce, blocker_kind),
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            _thread_search_delete_empty_response(response)
        except Exception:
            raise _failure() from None

    def disarm_thread_search_fixture_delete_blocker(
        self,
        fixture: ThreadSearchFixtureSnapshot,
        kind: str,
    ) -> None:
        try:
            search_fixture = _created_thread_search_fixture(fixture)
            blocker_kind = _thread_search_delete_single_blocker_kind(kind)
            self._require_open()
            response = self._client.request(
                "DELETE",
                _thread_search_delete_single_blocker_path(search_fixture.nonce, blocker_kind),
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            _thread_search_delete_empty_response(response)
        except Exception:
            raise _failure() from None

    def arm_thread_search_fixture_delete_peer_blocker(
        self,
        target_fixture: ThreadSearchFixtureSnapshot,
        peer_fixture: ThreadSearchFixtureSnapshot,
        kind: str,
    ) -> None:
        try:
            target = _created_thread_search_fixture(target_fixture)
            peer = _created_thread_search_fixture(peer_fixture)
            blocker_kind = _thread_search_delete_peer_blocker_kind(kind)
            _thread_search_delete_distinct_peer(target, peer)
            self._require_open()
            response = self._client.request(
                "POST",
                _thread_search_delete_peer_blocker_path(target.nonce, blocker_kind, peer.nonce),
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            _thread_search_delete_empty_response(response)
        except Exception:
            raise _failure() from None

    def disarm_thread_search_fixture_delete_peer_blocker(
        self,
        target_fixture: ThreadSearchFixtureSnapshot,
        peer_fixture: ThreadSearchFixtureSnapshot,
        kind: str,
    ) -> None:
        try:
            target = _created_thread_search_fixture(target_fixture)
            peer = _created_thread_search_fixture(peer_fixture)
            blocker_kind = _thread_search_delete_peer_blocker_kind(kind)
            _thread_search_delete_distinct_peer(target, peer)
            self._require_open()
            response = self._client.request(
                "DELETE",
                _thread_search_delete_peer_blocker_path(target.nonce, blocker_kind, peer.nonce),
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            _thread_search_delete_empty_response(response)
        except Exception:
            raise _failure() from None

    def observe_thread_search_fixture_delete(
        self,
        fixture: ThreadSearchFixtureSnapshot,
    ) -> ThreadSearchDeleteObservation:
        try:
            search_fixture = _created_thread_search_fixture(fixture)
            self._require_open()
            response = self._client.request(
                "DELETE",
                f"/api/_qa/mail-calendar/fixtures/{search_fixture.nonce}",
                content=b"",
                headers={"X-Hub-QA-Token": self._config.qa_token},
            )
            return _thread_search_delete_observation(response)
        except Exception:
            raise _failure() from None

    def dev_login(self, fixture: FixtureSnapshot) -> DevSession:
        try:
            if type(fixture) is not FixtureSnapshot:
                raise ValueError
            body = json.dumps({"email": fixture.owner_email}, ensure_ascii=True, separators=(",", ":")).encode("ascii")
            self._require_open()
            response = self._client.request(
                "POST",
                "/api/dev/login",
                content=body,
                headers={"Content-Type": "application/json"},
            )
            return DevSession.from_payload(_json_response(response, status_code=200), fixture)
        except Exception:
            raise _failure() from None

    def dev_login_thread_search_fixture(self, fixture: ThreadSearchFixtureSnapshot) -> DevSession:
        try:
            search_fixture = _created_thread_search_fixture(fixture)
            body = json.dumps({"email": search_fixture.owner_email}, ensure_ascii=True, separators=(",", ":")).encode("ascii")
            self._require_open()
            response = self._client.request(
                "POST",
                "/api/dev/login",
                content=body,
                headers={"Content-Type": "application/json"},
            )
            return DevSession.from_payload(_json_response(response, status_code=200), search_fixture)
        except Exception:
            raise _failure() from None

    def list_mail_threads(
        self,
        session: DevSession,
        *,
        search: str,
        starred: bool | None = None,
    ) -> MailThreadProjectionPage:
        try:
            _valid_session(session)
            normalized_search = _normalize_mail_thread_search(search)
            if starred is not None and type(starred) is not bool:
                raise ValueError
            params = {"search": normalized_search}
            if starred is not None:
                params["starred"] = "true" if starred else "false"
            self._require_open()
            response = self._client.request(
                "GET",
                "/api/mail/threads",
                params=params,
                headers=_authorization_headers(session),
            )
            return MailThreadProjectionPage.from_payload(_json_response(response, status_code=200))
        except Exception:
            raise _failure() from None

    def legacy_mail_search_status(self, session: DevSession) -> int:
        try:
            _valid_session(session)
            self._require_open()
            response = self._client.request(
                "GET",
                "/api/mail/search",
                headers=_authorization_headers(session),
            )
            return response.status_code
        except Exception:
            raise _failure() from None

    def mail_message_detail(self, session: DevSession, message_id: str) -> MailCorrelation:
        try:
            _valid_session(session)
            message_identifier = _canonical_uuid(message_id)
            self._require_open()
            response = self._client.request(
                "GET",
                f"/api/mail/messages/{message_identifier}",
                headers=_authorization_headers(session),
            )
            correlation = MailCorrelation.from_payload(_json_response(response, status_code=200))
            if correlation.message_id != message_identifier:
                raise ValueError
            return correlation
        except Exception:
            raise _failure() from None

    def calendar_event_detail(self, session: DevSession, event_id: str) -> CalendarCorrelation:
        try:
            _valid_session(session)
            event_identifier = _canonical_uuid(event_id)
            self._require_open()
            response = self._client.request(
                "GET",
                f"/api/calendar/v2/events/{event_identifier}",
                headers=_authorization_headers(session),
            )
            correlation = CalendarCorrelation.from_payload(_json_response(response, status_code=200))
            if correlation.event_id != event_identifier:
                raise ValueError
            return correlation
        except Exception:
            raise _failure() from None

    def respond_to_rsvp(
        self,
        session: DevSession,
        event_id: str,
        status: str,
        attendee_version: int,
        idempotency_key: str,
    ) -> RSVPResult:
        try:
            _valid_session(session)
            event_identifier = _canonical_uuid(event_id)
            if type(status) is not str or status not in _RSVP_INPUT_STATUSES:
                raise ValueError
            version = _positive_int(attendee_version)
            key = _idempotency_key(idempotency_key)
            body = json.dumps(
                {"status": status, "attendee_version": version}, separators=(",", ":"), ensure_ascii=True
            ).encode("ascii")
            self._require_open()
            response = self._client.request(
                "POST",
                f"/api/calendar/v2/events/{event_identifier}/rsvp",
                content=body,
                headers={
                    "Authorization": f"Bearer {session.access_token}",
                    "Content-Type": "application/json",
                    "Idempotency-Key": key,
                },
            )
            result = RSVPResult.from_payload(_json_response(response, status_code=200))
            if result.rsvp_status != status:
                raise ValueError
            return result
        except Exception:
            raise _failure() from None

    def close(self) -> None:
        if self._closed:
            return
        try:
            self._client.close()
            self._closed = True
        except Exception:
            raise _failure() from None

    def _require_open(self) -> None:
        if self._closed:
            raise ValueError


def _created_fixture(value: FixtureSnapshot) -> FixtureSnapshot:
    if type(value) is not FixtureSnapshot or value.state != "created":
        raise ValueError
    return value


def _published_fixture(value: FixtureSnapshot) -> FixtureSnapshot:
    if type(value) is not FixtureSnapshot or value.state != "published":
        raise ValueError
    _nonce(value.nonce)
    return value


def _created_thread_search_fixture(value: ThreadSearchFixtureSnapshot) -> ThreadSearchFixtureSnapshot:
    if type(value) is not ThreadSearchFixtureSnapshot or value.state != "created":
        raise ValueError
    return value


def _thread_search_delete_single_blocker_kind(value: object) -> str:
    if type(value) is not str or value not in _THREAD_SEARCH_DELETE_SINGLE_BLOCKER_KINDS:
        raise ValueError
    return value


def _thread_search_delete_peer_blocker_kind(value: object) -> str:
    if type(value) is not str or value not in _THREAD_SEARCH_DELETE_PEER_BLOCKER_KINDS:
        raise ValueError
    return value


def _thread_search_delete_distinct_peer(
    target: ThreadSearchFixtureSnapshot,
    peer: ThreadSearchFixtureSnapshot,
) -> None:
    if target.nonce == peer.nonce:
        raise ValueError


def _thread_search_delete_single_blocker_path(nonce: str, kind: str) -> str:
    fixture_nonce = _nonce(nonce)
    blocker_kind = _thread_search_delete_single_blocker_kind(kind)
    return f"/api/_qa/mail-calendar/fixtures/{fixture_nonce}/thread-search-delete-blocker/{blocker_kind}"


def _published_cleanup_fault_path(nonce: str, kind: PublishedCleanupFaultKind) -> str:
    fixture_nonce = _nonce(nonce)
    fault_kind = _published_cleanup_fault_kind(kind)
    return f"/api/_qa/mail-calendar/fixtures/{fixture_nonce}/published-cleanup-fault/{fault_kind.value}"


def _published_cleanup_fault_empty_response(
    response: httpx.Response,
    *,
    allow_missing: bool,
) -> bool:
    if (
        response.status_code == 204
        and not response.content
        and not response.headers.get_list("content-type")
        and not response.headers.get_list(_THREAD_SEARCH_DELETE_GUARD_HEADER)
    ):
        return True
    if (
        allow_missing
        and response.status_code == 409
        and not response.headers.get_list(_THREAD_SEARCH_DELETE_GUARD_HEADER)
        and _thread_search_delete_conflict_body(response) == _THREAD_SEARCH_DELETE_CONFLICT_BODY
    ):
        return False
    raise ValueError


def _thread_search_delete_peer_blocker_path(nonce: str, kind: str, peer_nonce: str) -> str:
    fixture_nonce = _nonce(nonce)
    blocker_kind = _thread_search_delete_peer_blocker_kind(kind)
    canonical_peer_nonce = _nonce(peer_nonce)
    if fixture_nonce == canonical_peer_nonce:
        raise ValueError
    return (
        f"/api/_qa/mail-calendar/fixtures/{fixture_nonce}/thread-search-delete-blocker/"
        f"{blocker_kind}/peer/{canonical_peer_nonce}"
    )


def _thread_search_delete_empty_response(response: httpx.Response) -> None:
    if (
        response.status_code != 204
        or response.content
        or response.headers.get_list("content-type")
        or response.headers.get_list(_THREAD_SEARCH_DELETE_GUARD_HEADER)
    ):
        raise ValueError


def _thread_search_delete_observation(response: httpx.Response) -> ThreadSearchDeleteObservation:
    guard_values = response.headers.get_list(_THREAD_SEARCH_DELETE_GUARD_HEADER)
    if response.status_code == 204:
        if response.content or response.headers.get_list("content-type") or guard_values:
            raise ValueError
        return ThreadSearchDeleteObservation.DELETED
    if response.status_code != 409 or _thread_search_delete_conflict_body(response) != _THREAD_SEARCH_DELETE_CONFLICT_BODY:
        raise ValueError
    if guard_values == [_THREAD_SEARCH_DELETE_GUARD_PREDELETE]:
        return ThreadSearchDeleteObservation.PREDELETE
    if not guard_values:
        return ThreadSearchDeleteObservation.CONFLICT_UNMARKED
    raise ValueError


def _thread_search_delete_conflict_body(response: httpx.Response) -> object:
    content_types = response.headers.get_list("content-type")
    if (
        len(content_types) != 1
        or content_types[0].split(";", 1)[0].strip().lower() != "application/json"
        or not response.content
        or len(response.content) > _MAX_RESPONSE_BYTES
    ):
        raise ValueError
    try:
        return json.loads(
            response.content.decode("utf-8"),
            object_pairs_hook=_unique_object,
            parse_constant=_reject_json_constant,
        )
    except (TypeError, UnicodeDecodeError, ValueError, json.JSONDecodeError):
        raise ValueError from None


def _valid_session(value: DevSession) -> None:
    if type(value) is not DevSession:
        raise ValueError


def _authorization_headers(session: DevSession) -> dict[str, str]:
    return {"Authorization": f"Bearer {session.access_token}"}


def _normalize_mail_thread_search(value: object) -> str:
    if type(value) is not str or len(value.encode("utf-8")) > _MAX_THREAD_SEARCH_BYTES:
        raise ValueError
    if any(category(character) == "Cc" for character in value):
        raise ValueError
    normalized = normalize("NFC", normalize("NFC", value).casefold())
    normalized = " ".join(normalized.split())
    if (
        not normalized
        or len(normalized.encode("utf-8")) > _MAX_THREAD_SEARCH_BYTES
        or any(category(character) == "Cc" for character in normalized)
        or any(character in normalized for character in "%_\\")
    ):
        raise ValueError
    return normalized


def _idempotency_key(value: object) -> str:
    if type(value) is not str or not 16 <= len(value) <= 128:
        raise ValueError
    if value != value.strip() or any(character.isspace() or ord(character) < 33 or ord(character) > 126 for character in value):
        raise ValueError
    return value


def _json_response(response: httpx.Response, *, status_code: int) -> object:
    if response.status_code != status_code:
        raise ValueError
    content_type = response.headers.get("content-type", "").lower()
    if not content_type.startswith("application/json"):
        raise ValueError
    body = response.content
    if not body or len(body) > _MAX_RESPONSE_BYTES:
        raise ValueError
    try:
        return json.loads(
            body.decode("utf-8"),
            object_pairs_hook=_unique_object,
            parse_constant=_reject_json_constant,
        )
    except (TypeError, UnicodeDecodeError, ValueError, json.JSONDecodeError):
        raise ValueError from None


def _unique_object(pairs: list[tuple[str, object]]) -> dict[str, object]:
    result: dict[str, object] = {}
    for key, value in pairs:
        if key in result:
            raise ValueError
        result[key] = value
    return result


def _reject_json_constant(_: str) -> NoReturn:
    raise ValueError


def _failure() -> MailCalendarQAError:
    return MailCalendarQAError(_FAILURE_TEXT)
