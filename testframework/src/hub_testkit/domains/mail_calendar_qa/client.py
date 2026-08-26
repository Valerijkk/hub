"""No-leak HTTP client for the narrow Mail and Calendar QA runtime seam."""

from __future__ import annotations

import json
from typing import Final, NoReturn

import httpx

from .models import (
    CalendarCorrelation,
    DevSession,
    FixtureSnapshot,
    MailCalendarQAConfig,
    MailCalendarQAError,
    MailCorrelation,
    RSVPResult,
    _canonical_uuid,
    _positive_int,
)


_FAILURE_TEXT: Final = "mail calendar QA request failed"
_CREATE_FIXTURE_BODY: Final = b'{"scenario":"inbound-request-v1"}'
_RSVP_INPUT_STATUSES: Final = frozenset({"accepted", "declined", "tentative"})
_MAX_RESPONSE_BYTES: Final = 262_144


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


def _valid_session(value: DevSession) -> None:
    if type(value) is not DevSession:
        raise ValueError


def _authorization_headers(session: DevSession) -> dict[str, str]:
    return {"Authorization": f"Bearer {session.access_token}"}


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
