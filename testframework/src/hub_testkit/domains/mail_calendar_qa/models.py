"""Strict, immutable values for the Mail and Calendar QA runtime boundary."""

from __future__ import annotations

from collections.abc import Mapping
from dataclasses import dataclass, field
from math import isfinite
from re import fullmatch
from types import MappingProxyType
from typing import Final, Self, TypeAlias
from uuid import UUID


class MailCalendarQAError(ValueError):
    """Safe public failure for the Mail and Calendar QA boundary."""


_INVALID_TEXT: Final = "mail calendar QA value is invalid"
_API_BASE_URL: Final = "http://qa-backend:3000/"
_FIXTURE_SCENARIO: Final = "inbound-request-v1"
_FIXTURE_STATES: Final = frozenset({"created", "published"})
_ACCOUNT_MATERIALIZATION_FIXTURE_SCENARIO: Final = "account-materialization-v1"
_ACCOUNT_MATERIALIZATION_FIXTURE_STATES: Final = frozenset({"created"})
_ACCOUNT_MATERIALIZATION_STATES: Final = frozenset(
    {"not_materialized", "materialized_pending", "finalized"}
)
_ACCOUNT_MATERIALIZATION_PRIMARY_IDENTITY_STATUS: Final = "pending"
_RSVP_STATUSES: Final = frozenset({"needs_action", "accepted", "declined", "tentative"})
_MAX_TOKEN_LENGTH: Final = 8_192
_MAX_JSON_DEPTH: Final = 12
_MAX_JSON_STRING_LENGTH: Final = 16_384
_MAX_JSON_MAPPING_ITEMS: Final = 256
_MAX_JSON_SEQUENCE_ITEMS: Final = 256

_FIXTURE_KEYS: Final = frozenset(
    {
        "nonce",
        "scenario",
        "state",
        "owner_email",
        "organization_id",
        "owner_user_id",
        "mail_account_id",
        "mail_folder_id",
        "mail_thread_id",
        "mail_message_id",
        "envelope_event_id",
    }
)
_ACCOUNT_MATERIALIZATION_FIXTURE_KEYS: Final = frozenset(
    {
        "nonce",
        "scenario",
        "state",
        "owner_email",
        "organization_id",
        "owner_user_id",
        "mail_domain_id",
        "expected_mail_account_id",
        "materialization_state",
        "mail_account_id",
        "primary_identity_id",
        "primary_identity_status",
    }
)
_DEV_SESSION_KEYS: Final = frozenset({"access_token", "refresh_token", "expires_in", "user"})
_MAIL_DETAIL_KEYS: Final = frozenset(
    {
        "id",
        "thread_id",
        "from",
        "to",
        "cc",
        "subject",
        "snippet",
        "sent_at",
        "received_at",
        "priority",
        "size_bytes",
        "has_attachments",
        "is_draft",
        "folder_ids",
        "seen",
        "flagged",
        "answered",
        "junk",
        "calendar_method",
        "calendar_event_id",
        "plain_body",
        "sanitized_html",
        "remote_images_blocked",
        "attachments",
        "calendar",
    }
)
_CALENDAR_DETAIL_KEYS: Final = frozenset(
    {
        "id",
        "title",
        "description",
        "location",
        "starts_at",
        "ends_at",
        "start_date",
        "end_date_exclusive",
        "all_day",
        "timezone",
        "rrule",
        "exdates",
        "sequence",
        "status",
        "version",
        "organizer_id",
        "organizer_email",
        "attendees",
        "viewer_attendee",
        "can_rsvp",
        "mail_message_id",
    }
)
_RSVP_RESULT_KEYS: Final = frozenset(
    {
        "id",
        "user_id",
        "email",
        "display_name",
        "role",
        "rsvp_status",
        "recurrence_id",
        "responded_at",
        "version",
        "reply_delivery_status",
    }
)

FrozenJSON: TypeAlias = None | bool | int | float | str | tuple["FrozenJSON", ...] | Mapping[str, "FrozenJSON"]


def _invalid() -> MailCalendarQAError:
    return MailCalendarQAError(_INVALID_TEXT)


def _exact_object(value: object, keys: frozenset[str]) -> dict[str, object]:
    if type(value) is not dict or frozenset(value) != keys:
        raise _invalid()
    return value


def _safe_text(value: object, *, maximum: int) -> str:
    if type(value) is not str or not value or len(value) > maximum:
        raise _invalid()
    if value != value.strip() or any(ord(character) < 32 or ord(character) == 127 for character in value):
        raise _invalid()
    return value


def _canonical_uuid(value: object) -> str:
    if type(value) is not str:
        raise _invalid()
    try:
        parsed = UUID(value)
    except (AttributeError, TypeError, ValueError):
        raise _invalid() from None
    if parsed.int == 0 or str(parsed) != value:
        raise _invalid()
    return value


def _optional_uuid(value: object) -> str | None:
    if value is None:
        return None
    return _canonical_uuid(value)


def _email(value: object) -> str:
    address = _safe_text(value, maximum=320)
    local, separator, domain = address.rpartition("@")
    if not separator or not local or not domain or "@" in local or "@" in domain:
        raise _invalid()
    return address


def _positive_int(value: object) -> int:
    if type(value) is not int or value <= 0 or value > 2_147_483_647:
        raise _invalid()
    return value


def _rsvp_status(value: object) -> str:
    if type(value) is not str or value not in _RSVP_STATUSES:
        raise _invalid()
    return value


def _freeze_json(value: object, *, depth: int) -> FrozenJSON:
    if depth > _MAX_JSON_DEPTH:
        raise _invalid()
    if value is None or type(value) is bool:
        return value
    if type(value) is int:
        if abs(value) > 9_007_199_254_740_991:
            raise _invalid()
        return value
    if type(value) is float:
        if not isfinite(value):
            raise _invalid()
        return value
    if type(value) is str:
        if len(value) > _MAX_JSON_STRING_LENGTH:
            raise _invalid()
        return value
    if isinstance(value, Mapping):
        if len(value) > _MAX_JSON_MAPPING_ITEMS:
            raise _invalid()
        copied: dict[str, FrozenJSON] = {}
        for key, item in value.items():
            if type(key) is not str or not key or len(key) > 256:
                raise _invalid()
            copied[key] = _freeze_json(item, depth=depth + 1)
        return MappingProxyType(copied)
    if type(value) is list:
        if len(value) > _MAX_JSON_SEQUENCE_ITEMS:
            raise _invalid()
        return tuple(_freeze_json(item, depth=depth + 1) for item in value)
    raise _invalid()


def _user_payload(value: object) -> Mapping[str, FrozenJSON]:
    frozen = _freeze_json(value, depth=0)
    if not isinstance(frozen, Mapping):
        raise _invalid()
    for field_name in ("id", "email", "timezone", "language", "system_role", "status"):
        _safe_text(frozen.get(field_name), maximum=4_096)
    _canonical_uuid(frozen["id"])
    _email(frozen["email"])
    if type(frozen.get("onboarding_completed")) is not bool:
        raise _invalid()
    return frozen


@dataclass(frozen=True, slots=True)
class MailCalendarQAConfig:
    """Fixed QA-backend location and non-displayable QA capability token."""

    api_base_url: str = field(repr=False)
    qa_token: str = field(repr=False)

    def __post_init__(self) -> None:
        if type(self.api_base_url) is not str or self.api_base_url != _API_BASE_URL:
            raise _invalid()
        if type(self.qa_token) is not str or fullmatch(r"[0-9a-f]{64}", self.qa_token) is None:
            raise _invalid()


@dataclass(frozen=True, slots=True)
class FixtureSnapshot:
    """One immutable QA fixture identity and its lifecycle state."""

    nonce: str
    scenario: str
    state: str
    owner_email: str = field(repr=False)
    organization_id: str
    owner_user_id: str
    mail_account_id: str
    mail_folder_id: str
    mail_thread_id: str
    mail_message_id: str
    envelope_event_id: str

    def __post_init__(self) -> None:
        _nonce(self.nonce)
        if type(self.scenario) is not str or self.scenario != _FIXTURE_SCENARIO:
            raise _invalid()
        if type(self.state) is not str or self.state not in _FIXTURE_STATES:
            raise _invalid()
        _email(self.owner_email)
        for identifier in (
            self.organization_id,
            self.owner_user_id,
            self.mail_account_id,
            self.mail_folder_id,
            self.mail_thread_id,
            self.mail_message_id,
            self.envelope_event_id,
        ):
            _canonical_uuid(identifier)

    @classmethod
    def from_payload(cls, payload: object) -> Self:
        value = _exact_object(payload, _FIXTURE_KEYS)
        return cls(**value)  # type: ignore[arg-type]

    @property
    def immutable_identity(self) -> tuple[str, ...]:
        return (
            self.nonce,
            self.scenario,
            self.owner_email,
            self.organization_id,
            self.owner_user_id,
            self.mail_account_id,
            self.mail_folder_id,
            self.mail_thread_id,
            self.mail_message_id,
            self.envelope_event_id,
        )


@dataclass(frozen=True, slots=True)
class AccountMaterializationFixtureSnapshot:
    """One immutable public account-materialization fixture observation."""

    nonce: str
    scenario: str
    state: str
    owner_email: str = field(repr=False)
    organization_id: str
    owner_user_id: str
    mail_domain_id: str
    expected_mail_account_id: str
    materialization_state: str
    mail_account_id: str | None
    primary_identity_id: str | None
    primary_identity_status: str | None

    def __post_init__(self) -> None:
        _nonce(self.nonce)
        if type(self.scenario) is not str or self.scenario != _ACCOUNT_MATERIALIZATION_FIXTURE_SCENARIO:
            raise _invalid()
        if type(self.state) is not str or self.state not in _ACCOUNT_MATERIALIZATION_FIXTURE_STATES:
            raise _invalid()
        _email(self.owner_email)
        for identifier in (
            self.organization_id,
            self.owner_user_id,
            self.mail_domain_id,
            self.expected_mail_account_id,
        ):
            _canonical_uuid(identifier)
        if type(self.materialization_state) is not str or self.materialization_state not in _ACCOUNT_MATERIALIZATION_STATES:
            raise _invalid()
        if self.materialization_state == "not_materialized":
            if self.mail_account_id is not None or self.primary_identity_id is not None or self.primary_identity_status is not None:
                raise _invalid()
            return
        _canonical_uuid(self.mail_account_id)
        _canonical_uuid(self.primary_identity_id)
        if self.primary_identity_status != _ACCOUNT_MATERIALIZATION_PRIMARY_IDENTITY_STATUS:
            raise _invalid()

    @classmethod
    def from_payload(cls, payload: object) -> Self:
        value = _exact_object(payload, _ACCOUNT_MATERIALIZATION_FIXTURE_KEYS)
        return cls(**value)  # type: ignore[arg-type]

    @property
    def immutable_roots(self) -> tuple[str, ...]:
        return (
            self.nonce,
            self.scenario,
            self.owner_email,
            self.organization_id,
            self.owner_user_id,
            self.mail_domain_id,
            self.expected_mail_account_id,
        )


def _nonce(value: object) -> str:
    if type(value) is not str or fullmatch(r"[0-9a-f]{32}", value) is None:
        raise _invalid()
    return value


@dataclass(frozen=True, slots=True)
class DevSession:
    """Copied dev-login result with secrets and raw user data hidden from repr."""

    access_token: str = field(repr=False)
    refresh_token: str = field(repr=False)
    expires_in: int
    user: Mapping[str, FrozenJSON] = field(repr=False)

    def __post_init__(self) -> None:
        _session_token(self.access_token)
        _session_token(self.refresh_token)
        _positive_int(self.expires_in)
        object.__setattr__(self, "user", _user_payload(self.user))

    @classmethod
    def from_payload(cls, payload: object, fixture: FixtureSnapshot) -> Self:
        value = _exact_object(payload, _DEV_SESSION_KEYS)
        if type(fixture) is not FixtureSnapshot:
            raise _invalid()
        session = cls(**value)  # type: ignore[arg-type]
        if session.user["id"] != fixture.owner_user_id or session.user["email"] != fixture.owner_email:
            raise _invalid()
        return session


def _session_token(value: object) -> str:
    token = _safe_text(value, maximum=_MAX_TOKEN_LENGTH)
    if any(character.isspace() for character in token):
        raise _invalid()
    return token


@dataclass(frozen=True, slots=True)
class MailCorrelation:
    """Only Mail-to-Calendar identifiers needed by QA scenarios."""

    message_id: str
    thread_id: str
    calendar_event_id: str | None

    def __post_init__(self) -> None:
        _canonical_uuid(self.message_id)
        _canonical_uuid(self.thread_id)
        _optional_uuid(self.calendar_event_id)

    @classmethod
    def from_payload(cls, payload: object) -> Self:
        value = _exact_object(payload, _MAIL_DETAIL_KEYS)
        return cls(
            message_id=value["id"],
            thread_id=value["thread_id"],
            calendar_event_id=value["calendar_event_id"],
        )


@dataclass(frozen=True, slots=True)
class CalendarViewerAttendee:
    """The current viewer's constrained RSVP concurrency state."""

    attendee_id: str
    version: int
    rsvp_status: str

    def __post_init__(self) -> None:
        _canonical_uuid(self.attendee_id)
        _positive_int(self.version)
        _rsvp_status(self.rsvp_status)

    @classmethod
    def from_payload(cls, payload: object) -> Self:
        if type(payload) is not dict:
            raise _invalid()
        if frozenset(payload) != frozenset(
            {"id", "user_id", "email", "role", "rsvp_status", "version", "responded_at"}
        ):
            raise _invalid()
        return cls(
            attendee_id=payload["id"],
            version=payload["version"],
            rsvp_status=payload["rsvp_status"],
        )


@dataclass(frozen=True, slots=True)
class CalendarCorrelation:
    """Only Calendar-to-Mail identifiers and viewer RSVP state used by QA."""

    event_id: str
    mail_message_id: str | None
    can_rsvp: bool
    viewer_attendee: CalendarViewerAttendee | None

    def __post_init__(self) -> None:
        _canonical_uuid(self.event_id)
        _optional_uuid(self.mail_message_id)
        if type(self.can_rsvp) is not bool:
            raise _invalid()
        if self.viewer_attendee is not None and type(self.viewer_attendee) is not CalendarViewerAttendee:
            raise _invalid()
        if self.can_rsvp and self.viewer_attendee is None:
            raise _invalid()

    @classmethod
    def from_payload(cls, payload: object) -> Self:
        value = _exact_object(payload, _CALENDAR_DETAIL_KEYS)
        raw_attendee = value["viewer_attendee"]
        attendee = None if raw_attendee is None else CalendarViewerAttendee.from_payload(raw_attendee)
        return cls(
            event_id=value["id"],
            mail_message_id=value["mail_message_id"],
            can_rsvp=value["can_rsvp"],
            viewer_attendee=attendee,
        )


@dataclass(frozen=True, slots=True)
class RSVPResult:
    """Constrained attendee receipt returned from a canonical series RSVP."""

    attendee_id: str
    version: int
    rsvp_status: str

    def __post_init__(self) -> None:
        _canonical_uuid(self.attendee_id)
        _positive_int(self.version)
        _rsvp_status(self.rsvp_status)

    @classmethod
    def from_payload(cls, payload: object) -> Self:
        value = _exact_object(payload, _RSVP_RESULT_KEYS)
        return cls(
            attendee_id=value["id"],
            version=value["version"],
            rsvp_status=value["rsvp_status"],
        )
