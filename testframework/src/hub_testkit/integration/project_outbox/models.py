from __future__ import annotations

import re
from dataclasses import dataclass, field
from datetime import datetime
from typing import Literal, NoReturn
from uuid import UUID


class ProjectOutboxModelError(ValueError):
    pass


ProjectOutboxFlow = Literal[
    "create",
    "duplicate",
    "anonymous",
    "malformed",
    "observer-replay",
]

_FLOWS = frozenset(
    {"create", "duplicate", "anonymous", "malformed", "observer-replay"}
)
_LABEL_RE = re.compile(r"^[a-z][a-z0-9-]{2,31}$")
_NONCE_RE = re.compile(r"^[0-9a-f]{32}$")
_KEY_RE = re.compile(r"^[A-Z0-9]{2,10}$")
_INVALID = object()


@dataclass(frozen=True, slots=True, repr=False, eq=False, init=False)
class ProjectOutboxSeed:
    _label: str = field(repr=False)
    _flow: ProjectOutboxFlow = field(repr=False)
    _nonce: str = field(repr=False)
    _project_name: str = field(repr=False)
    _project_key: str = field(repr=False)
    _description: str = field(repr=False)

    def __init__(self) -> None:
        _raise_invalid()

    @classmethod
    def _create(
        cls,
        *,
        label: str,
        flow: ProjectOutboxFlow,
        nonce: str,
        project_name: str,
        project_key: str,
        description: str,
    ) -> ProjectOutboxSeed:
        result = _seed_create_safe(
            cls,
            label,
            flow,
            nonce,
            project_name,
            project_key,
            description,
        )
        del label, flow, nonce, project_name, project_key, description
        if type(result) is cls:
            return result
        del result
        _raise_invalid()

    def __copy__(self) -> NoReturn:
        _raise_invalid()

    def __deepcopy__(self, memo: object) -> NoReturn:
        del memo
        _raise_invalid()

    def __repr__(self) -> str:
        return "ProjectOutboxSeed()"


@dataclass(frozen=True, slots=True, repr=False, eq=False, init=False)
class ProvisionedProjectOutbox:
    _owner_user_id: str = field(repr=False)
    _owner_subject: str = field(repr=False)
    _owner_token: str = field(repr=False)
    _nonce: str = field(repr=False)
    _observer_precreated: bool = field(repr=False)
    _retention_policy: Literal["interest"] = field(repr=False)
    _runtime_binding: object = field(repr=False)

    def __init__(self) -> None:
        _raise_invalid()

    @classmethod
    def _create(
        cls,
        *,
        owner_user_id: str,
        owner_subject: str,
        owner_token: str,
        nonce: str,
        observer_precreated: bool,
        retention_policy: Literal["interest"],
        runtime_binding: object,
    ) -> ProvisionedProjectOutbox:
        result = _provisioned_create_safe(
            cls,
            owner_user_id,
            owner_subject,
            owner_token,
            nonce,
            observer_precreated,
            retention_policy,
            runtime_binding,
        )
        del (
            owner_user_id,
            owner_subject,
            owner_token,
            nonce,
            observer_precreated,
            retention_policy,
            runtime_binding,
        )
        if type(result) is cls:
            return result
        del result
        _raise_invalid()

    def __copy__(self) -> NoReturn:
        _raise_invalid()

    def __deepcopy__(self, memo: object) -> NoReturn:
        del memo
        _raise_invalid()

    def __repr__(self) -> str:
        return "ProvisionedProjectOutbox()"


@dataclass(frozen=True, slots=True, repr=False, eq=False, init=False)
class ProjectProjection:
    _project_id: str = field(repr=False)
    _name: str = field(repr=False)
    _key: str = field(repr=False)
    _description: str = field(repr=False)
    _status: Literal["active"] = field(repr=False)
    _member_count: int = field(repr=False)
    _role: Literal["owner"] = field(repr=False)
    _version: int = field(repr=False)
    _created_at: str = field(repr=False)
    _updated_at: str = field(repr=False)

    def __init__(self) -> None:
        _raise_invalid()

    @classmethod
    def _create(
        cls,
        *,
        project_id: str,
        name: str,
        key: str,
        description: str,
        status: Literal["active"],
        member_count: int,
        role: Literal["owner"],
        version: int,
        created_at: str,
        updated_at: str,
    ) -> ProjectProjection:
        result = _projection_create_safe(
            cls,
            project_id,
            name,
            key,
            description,
            status,
            member_count,
            role,
            version,
            created_at,
            updated_at,
        )
        del (
            project_id,
            name,
            key,
            description,
            status,
            member_count,
            role,
            version,
            created_at,
            updated_at,
        )
        if type(result) is cls:
            return result
        del result
        _raise_invalid()

    def __copy__(self) -> NoReturn:
        _raise_invalid()

    def __deepcopy__(self, memo: object) -> NoReturn:
        del memo
        _raise_invalid()

    def __repr__(self) -> str:
        return "ProjectProjection()"


@dataclass(frozen=True, slots=True, repr=False, eq=False, init=False)
class ProjectCreatedReceipt:
    """One durable typed receipt, without acknowledgement-order provenance."""

    _event_id: str = field(repr=False)
    _subject: Literal["project.created"] = field(repr=False)
    _project_id: str = field(repr=False)
    _name: str = field(repr=False)
    _key: str = field(repr=False)
    _created_by_id: str = field(repr=False)
    _runtime_binding: object = field(repr=False)

    def __init__(self) -> None:
        _raise_invalid()

    @classmethod
    def _create(
        cls,
        *,
        event_id: str,
        subject: Literal["project.created"],
        project_id: str,
        name: str,
        key: str,
        created_by_id: str,
        runtime_binding: object,
    ) -> ProjectCreatedReceipt:
        result = _receipt_create_safe(
            cls,
            event_id,
            subject,
            project_id,
            name,
            key,
            created_by_id,
            runtime_binding,
        )
        del (
            event_id,
            subject,
            project_id,
            name,
            key,
            created_by_id,
            runtime_binding,
        )
        if type(result) is cls:
            return result
        del result
        _raise_invalid()

    def __copy__(self) -> NoReturn:
        _raise_invalid()

    def __deepcopy__(self, memo: object) -> NoReturn:
        del memo
        _raise_invalid()

    def __repr__(self) -> str:
        return "ProjectCreatedReceipt()"


@dataclass(frozen=True, slots=True, repr=False, eq=False, init=False)
class OwnedProjectOutboxState:
    """Closed final counts, without transaction or transition-order provenance."""

    _project_count: int = field(repr=False)
    _owner_membership_count: int = field(repr=False)
    _outbox_row_count: int = field(repr=False)
    _published_outbox_row_count: int = field(repr=False)
    _runtime_binding: object = field(repr=False)

    def __init__(self) -> None:
        _raise_invalid()

    @classmethod
    def _create(
        cls,
        *,
        project_count: int,
        owner_membership_count: int,
        outbox_row_count: int,
        published_outbox_row_count: int,
        runtime_binding: object,
    ) -> OwnedProjectOutboxState:
        result = _state_create_safe(
            cls,
            project_count,
            owner_membership_count,
            outbox_row_count,
            published_outbox_row_count,
            runtime_binding,
        )
        del (
            project_count,
            owner_membership_count,
            outbox_row_count,
            published_outbox_row_count,
            runtime_binding,
        )
        if type(result) is cls:
            return result
        del result
        _raise_invalid()

    def __copy__(self) -> NoReturn:
        _raise_invalid()

    def __deepcopy__(self, memo: object) -> NoReturn:
        del memo
        _raise_invalid()

    def __repr__(self) -> str:
        return "OwnedProjectOutboxState()"


def provisioned_project_outbox(
    *,
    owner_user_id: str,
    owner_subject: str,
    owner_token: str,
    nonce: str,
    observer_precreated: bool,
    retention_policy: Literal["interest"],
    runtime_binding: object,
) -> ProvisionedProjectOutbox:
    result = _provisioned_create_safe(
        ProvisionedProjectOutbox,
        owner_user_id,
        owner_subject,
        owner_token,
        nonce,
        observer_precreated,
        retention_policy,
        runtime_binding,
    )
    del (
        owner_user_id,
        owner_subject,
        owner_token,
        nonce,
        observer_precreated,
        retention_policy,
        runtime_binding,
    )
    if type(result) is ProvisionedProjectOutbox:
        return result
    del result
    _raise_invalid()


def _seed_create_safe(
    cls: type[ProjectOutboxSeed],
    label: object,
    flow: object,
    nonce: object,
    project_name: object,
    project_key: object,
    description: object,
) -> ProjectOutboxSeed | object:
    try:
        if (
            type(label) is not str
            or _LABEL_RE.fullmatch(label) is None
            or type(flow) is not str
            or flow not in _FLOWS
            or not _nonce_safe(nonce)
            or not _text_safe(project_name, 100, allow_empty=False)
            or type(project_key) is not str
            or _KEY_RE.fullmatch(project_key) is None
            or not _text_safe(description, 2000, allow_empty=True)
        ):
            return _INVALID
        result = object.__new__(cls)
        object.__setattr__(result, "_label", label)
        object.__setattr__(result, "_flow", flow)
        object.__setattr__(result, "_nonce", nonce)
        object.__setattr__(result, "_project_name", project_name)
        object.__setattr__(result, "_project_key", project_key)
        object.__setattr__(result, "_description", description)
        return result
    except Exception:
        return _INVALID


def _provisioned_create_safe(
    cls: type[ProvisionedProjectOutbox],
    owner_user_id: object,
    owner_subject: object,
    owner_token: object,
    nonce: object,
    observer_precreated: object,
    retention_policy: object,
    runtime_binding: object,
) -> ProvisionedProjectOutbox | object:
    try:
        if (
            not _uuid_safe(owner_user_id)
            or not _text_safe(owner_subject, 128, allow_empty=False)
            or owner_subject != owner_user_id
            or not _text_safe(owner_token, 4096, allow_empty=False)
            or not _nonce_safe(nonce)
            or type(observer_precreated) is not bool
            or observer_precreated is not True
            or type(retention_policy) is not str
            or retention_policy != "interest"
            or runtime_binding is None
        ):
            return _INVALID
        result = object.__new__(cls)
        object.__setattr__(result, "_owner_user_id", owner_user_id)
        object.__setattr__(result, "_owner_subject", owner_subject)
        object.__setattr__(result, "_owner_token", owner_token)
        object.__setattr__(result, "_nonce", nonce)
        object.__setattr__(result, "_observer_precreated", observer_precreated)
        object.__setattr__(result, "_retention_policy", retention_policy)
        object.__setattr__(result, "_runtime_binding", runtime_binding)
        return result
    except Exception:
        return _INVALID


def _projection_create_safe(
    cls: type[ProjectProjection],
    project_id: object,
    name: object,
    key: object,
    description: object,
    status: object,
    member_count: object,
    role: object,
    version: object,
    created_at: object,
    updated_at: object,
) -> ProjectProjection | object:
    try:
        if (
            not _uuid_safe(project_id)
            or not _text_safe(name, 100, allow_empty=False)
            or type(key) is not str
            or _KEY_RE.fullmatch(key) is None
            or not _text_safe(description, 2000, allow_empty=True)
            or type(status) is not str
            or status != "active"
            or type(member_count) is not int
            or member_count < 1
            or type(role) is not str
            or role != "owner"
            or type(version) is not int
            or version < 1
            or not _timestamp_safe(created_at)
            or not _timestamp_safe(updated_at)
        ):
            return _INVALID
        result = object.__new__(cls)
        object.__setattr__(result, "_project_id", project_id)
        object.__setattr__(result, "_name", name)
        object.__setattr__(result, "_key", key)
        object.__setattr__(result, "_description", description)
        object.__setattr__(result, "_status", status)
        object.__setattr__(result, "_member_count", member_count)
        object.__setattr__(result, "_role", role)
        object.__setattr__(result, "_version", version)
        object.__setattr__(result, "_created_at", created_at)
        object.__setattr__(result, "_updated_at", updated_at)
        return result
    except Exception:
        return _INVALID


def _receipt_create_safe(
    cls: type[ProjectCreatedReceipt],
    event_id: object,
    subject: object,
    project_id: object,
    name: object,
    key: object,
    created_by_id: object,
    runtime_binding: object,
) -> ProjectCreatedReceipt | object:
    try:
        if (
            not _uuid_safe(event_id)
            or type(subject) is not str
            or subject != "project.created"
            or not _uuid_safe(project_id)
            or not _text_safe(name, 100, allow_empty=False)
            or type(key) is not str
            or _KEY_RE.fullmatch(key) is None
            or not _uuid_safe(created_by_id)
            or runtime_binding is None
        ):
            return _INVALID
        result = object.__new__(cls)
        object.__setattr__(result, "_event_id", event_id)
        object.__setattr__(result, "_subject", subject)
        object.__setattr__(result, "_project_id", project_id)
        object.__setattr__(result, "_name", name)
        object.__setattr__(result, "_key", key)
        object.__setattr__(result, "_created_by_id", created_by_id)
        object.__setattr__(result, "_runtime_binding", runtime_binding)
        return result
    except Exception:
        return _INVALID


def _state_create_safe(
    cls: type[OwnedProjectOutboxState],
    project_count: object,
    owner_membership_count: object,
    outbox_row_count: object,
    published_outbox_row_count: object,
    runtime_binding: object,
) -> OwnedProjectOutboxState | object:
    try:
        counts = (
            project_count,
            owner_membership_count,
            outbox_row_count,
            published_outbox_row_count,
        )
        if (
            any(type(count) is not int or count not in {0, 1} for count in counts)
            or runtime_binding is None
        ):
            return _INVALID
        result = object.__new__(cls)
        object.__setattr__(result, "_project_count", project_count)
        object.__setattr__(result, "_owner_membership_count", owner_membership_count)
        object.__setattr__(result, "_outbox_row_count", outbox_row_count)
        object.__setattr__(
            result,
            "_published_outbox_row_count",
            published_outbox_row_count,
        )
        object.__setattr__(result, "_runtime_binding", runtime_binding)
        return result
    except Exception:
        return _INVALID


def _seed_details(
    seed: object,
) -> tuple[str, ProjectOutboxFlow, str, str, str, str] | object:
    try:
        if type(seed) is not ProjectOutboxSeed:
            return _INVALID
        values = (
            seed._label,
            seed._flow,
            seed._nonce,
            seed._project_name,
            seed._project_key,
            seed._description,
        )
        clone = _seed_create_safe(ProjectOutboxSeed, *values)
        if type(clone) is not ProjectOutboxSeed:
            return _INVALID
        return values
    except Exception:
        return _INVALID


def _provisioned_details(
    value: object,
) -> tuple[str, str, str, str, object] | object:
    try:
        if type(value) is not ProvisionedProjectOutbox:
            return _INVALID
        clone = _provisioned_create_safe(
            ProvisionedProjectOutbox,
            value._owner_user_id,
            value._owner_subject,
            value._owner_token,
            value._nonce,
            value._observer_precreated,
            value._retention_policy,
            value._runtime_binding,
        )
        if type(clone) is not ProvisionedProjectOutbox:
            return _INVALID
        return (
            value._owner_user_id,
            value._owner_subject,
            value._owner_token,
            value._nonce,
            value._runtime_binding,
        )
    except Exception:
        return _INVALID


def _projection_details(
    value: object,
) -> tuple[str, str, str, str, str, int, str, int] | object:
    try:
        if type(value) is not ProjectProjection:
            return _INVALID
        clone = _projection_create_safe(
            ProjectProjection,
            value._project_id,
            value._name,
            value._key,
            value._description,
            value._status,
            value._member_count,
            value._role,
            value._version,
            value._created_at,
            value._updated_at,
        )
        if type(clone) is not ProjectProjection:
            return _INVALID
        return (
            value._project_id,
            value._name,
            value._key,
            value._description,
            value._status,
            value._member_count,
            value._role,
            value._version,
        )
    except Exception:
        return _INVALID


def _receipt_details(
    value: object,
) -> tuple[str, str, str, str, str, str, object] | object:
    try:
        if type(value) is not ProjectCreatedReceipt:
            return _INVALID
        clone = _receipt_create_safe(
            ProjectCreatedReceipt,
            value._event_id,
            value._subject,
            value._project_id,
            value._name,
            value._key,
            value._created_by_id,
            value._runtime_binding,
        )
        if type(clone) is not ProjectCreatedReceipt:
            return _INVALID
        return (
            value._event_id,
            value._subject,
            value._project_id,
            value._name,
            value._key,
            value._created_by_id,
            value._runtime_binding,
        )
    except Exception:
        return _INVALID


def _state_details(
    value: object,
) -> tuple[int, int, int, int, object] | object:
    try:
        if type(value) is not OwnedProjectOutboxState:
            return _INVALID
        clone = _state_create_safe(
            OwnedProjectOutboxState,
            value._project_count,
            value._owner_membership_count,
            value._outbox_row_count,
            value._published_outbox_row_count,
            value._runtime_binding,
        )
        if type(clone) is not OwnedProjectOutboxState:
            return _INVALID
        return (
            value._project_count,
            value._owner_membership_count,
            value._outbox_row_count,
            value._published_outbox_row_count,
            value._runtime_binding,
        )
    except Exception:
        return _INVALID


def _uuid_safe(value: object) -> bool:
    try:
        return type(value) is str and str(UUID(value)) == value
    except Exception:
        return False


def _nonce_safe(value: object) -> bool:
    try:
        return type(value) is str and _NONCE_RE.fullmatch(value) is not None
    except Exception:
        return False


def _text_safe(value: object, maximum: int, *, allow_empty: bool) -> bool:
    try:
        if type(value) is not str or len(value) > maximum:
            return False
        if not allow_empty and not value.strip():
            return False
        return all(ord(character) >= 32 and ord(character) != 127 for character in value)
    except Exception:
        return False


def _timestamp_safe(value: object) -> bool:
    try:
        if not _text_safe(value, 128, allow_empty=False) or type(value) is not str:
            return False
        parsed = datetime.fromisoformat(value.replace("Z", "+00:00"))
        return parsed.tzinfo is not None
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise ProjectOutboxModelError("project outbox model is invalid") from None
