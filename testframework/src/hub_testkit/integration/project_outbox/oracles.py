from __future__ import annotations

import json
from collections.abc import Mapping
from typing import Literal, NoReturn, cast

from hub_testkit.http import HttpExchange

from .models import (
    OwnedProjectOutboxState,
    ProjectCreatedReceipt,
    ProjectOutboxSeed,
    ProjectProjection,
    ProvisionedProjectOutbox,
    _projection_details,
    _provisioned_details,
    _receipt_details,
    _seed_details,
    _state_details,
)


class ProjectOutboxOracleError(ValueError):
    pass


_INVALID = object()
_MAX_DOCUMENT_BYTES = 256 * 1024
_MAX_NODES = 256
_MAX_DEPTH = 16
_PROJECT_KEYS = frozenset(
    {
        "id",
        "name",
        "key",
        "description",
        "status",
        "member_count",
        "my_role",
        "version",
        "created_at",
        "updated_at",
    }
)
_EVENT_KEYS = frozenset({"project_id", "name", "key", "created_by_id"})


def parse_created_project(exchange: HttpExchange) -> ProjectProjection:
    result = _project_safe(exchange)
    del exchange
    if type(result) is ProjectProjection:
        return result
    del result
    _raise_invalid()


def require_project_outbox_status(exchange: HttpExchange, status_code: int) -> None:
    result = _status_safe(exchange, status_code)
    del exchange, status_code
    if result is True:
        return
    del result
    _raise_invalid()


def parse_project_created_receipt(
    *,
    event_id: str,
    subject: str,
    body: bytes,
    runtime_binding: object,
) -> ProjectCreatedReceipt:
    """Project one durable receipt without inferring publication ordering."""

    result = _receipt_safe(event_id, subject, body, runtime_binding)
    del event_id, subject, body, runtime_binding
    if type(result) is ProjectCreatedReceipt:
        return result
    del result
    _raise_invalid()


def project_outbox_state_projection(
    *,
    project_count: int,
    owner_membership_count: int,
    outbox_row_count: int,
    published_outbox_row_count: int,
    runtime_binding: object,
) -> OwnedProjectOutboxState:
    """Project final counts without inferring a shared transaction boundary."""

    result = _state_safe(
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
    if type(result) is OwnedProjectOutboxState:
        return result
    del result
    _raise_invalid()


def project_matches_seed(
    project: ProjectProjection,
    seed: ProjectOutboxSeed,
) -> bool:
    return _project_matches_seed_safe(project, seed)


def receipt_matches_project_and_owner(
    receipt: ProjectCreatedReceipt,
    project: ProjectProjection,
    seed: ProjectOutboxSeed,
    provisioned: ProvisionedProjectOutbox,
    runtime_binding: object,
) -> bool:
    return _receipt_matches_safe(
        receipt,
        project,
        seed,
        provisioned,
        runtime_binding,
    )


def state_matches_counts(
    state: OwnedProjectOutboxState,
    expected: tuple[int, int, int, int],
    runtime_binding: object,
) -> bool:
    """Compare only closed final counts and their runtime binding."""

    return _state_matches_safe(state, expected, runtime_binding)


def _project_safe(exchange: object) -> ProjectProjection | object:
    try:
        parts = _exchange_parts_safe(exchange)
        if not isinstance(parts, tuple):
            return _INVALID
        status_code, body = parts
        if status_code != 201:
            return _INVALID
        document = _load_document_safe(body)
        if type(document) is not dict or frozenset(document) != _PROJECT_KEYS:
            return _INVALID
        project_id = document["id"]
        name = document["name"]
        key = document["key"]
        description = document["description"]
        status = document["status"]
        member_count = document["member_count"]
        role = document["my_role"]
        version = document["version"]
        created_at = document["created_at"]
        updated_at = document["updated_at"]
        if (
            type(project_id) is not str
            or type(name) is not str
            or type(key) is not str
            or type(description) is not str
            or type(status) is not str
            or type(member_count) is not int
            or type(role) is not str
            or type(version) is not int
            or type(created_at) is not str
            or type(updated_at) is not str
        ):
            return _INVALID
        return ProjectProjection._create(
            project_id=project_id,
            name=name,
            key=key,
            description=description,
            status=cast(Literal["active"], status),
            member_count=member_count,
            role=cast(Literal["owner"], role),
            version=version,
            created_at=created_at,
            updated_at=updated_at,
        )
    except Exception:
        return _INVALID


def _status_safe(exchange: object, status_code: object) -> bool:
    try:
        if type(status_code) is not int or status_code not in {201, 400, 401, 409}:
            return False
        parts = _exchange_parts_safe(exchange)
        return isinstance(parts, tuple) and parts[0] == status_code
    except Exception:
        return False


def _receipt_safe(
    event_id: object,
    subject: object,
    body: object,
    runtime_binding: object,
) -> ProjectCreatedReceipt | object:
    try:
        if type(event_id) is not str or type(subject) is not str or type(body) is not bytes:
            return _INVALID
        document = _load_document_safe(body)
        if type(document) is not dict or frozenset(document) != _EVENT_KEYS:
            return _INVALID
        project_id = document["project_id"]
        name = document["name"]
        key = document["key"]
        created_by_id = document["created_by_id"]
        if (
            type(project_id) is not str
            or type(name) is not str
            or type(key) is not str
            or type(created_by_id) is not str
        ):
            return _INVALID
        return ProjectCreatedReceipt._create(
            event_id=event_id,
            subject=cast(Literal["project.created"], subject),
            project_id=project_id,
            name=name,
            key=key,
            created_by_id=created_by_id,
            runtime_binding=runtime_binding,
        )
    except Exception:
        return _INVALID


def _state_safe(
    project_count: object,
    owner_membership_count: object,
    outbox_row_count: object,
    published_outbox_row_count: object,
    runtime_binding: object,
) -> OwnedProjectOutboxState | object:
    try:
        if (
            type(project_count) is not int
            or type(owner_membership_count) is not int
            or type(outbox_row_count) is not int
            or type(published_outbox_row_count) is not int
        ):
            return _INVALID
        return OwnedProjectOutboxState._create(
            project_count=project_count,
            owner_membership_count=owner_membership_count,
            outbox_row_count=outbox_row_count,
            published_outbox_row_count=published_outbox_row_count,
            runtime_binding=runtime_binding,
        )
    except Exception:
        return _INVALID


def _project_matches_seed_safe(project: object, seed: object) -> bool:
    try:
        project_details = _projection_details(project)
        seed_details = _seed_details(seed)
        if not isinstance(project_details, tuple) or not isinstance(seed_details, tuple):
            return False
        project_id, name, key, description, status, member_count, role, version = (
            project_details
        )
        _, _, _, expected_name, expected_key, expected_description = seed_details
        return (
            bool(project_id)
            and name == expected_name
            and key == expected_key
            and description == expected_description
            and status == "active"
            and member_count == 1
            and role == "owner"
            and version == 1
        )
    except Exception:
        return False


def _receipt_matches_safe(
    receipt: object,
    project: object,
    seed: object,
    provisioned: object,
    runtime_binding: object,
) -> bool:
    try:
        receipt_details = _receipt_details(receipt)
        project_details = _projection_details(project)
        seed_details = _seed_details(seed)
        provisioned_details = _provisioned_details(provisioned)
        if (
            not isinstance(receipt_details, tuple)
            or not isinstance(project_details, tuple)
            or not isinstance(seed_details, tuple)
            or not isinstance(provisioned_details, tuple)
        ):
            return False
        event_id, subject, project_id, name, key, created_by_id, binding = (
            receipt_details
        )
        projected_id, projected_name, projected_key = project_details[:3]
        _, _, _, seed_name, seed_key, _ = seed_details
        owner_id, _, _, _, provisioned_binding = provisioned_details
        return (
            bool(event_id)
            and subject == "project.created"
            and project_id == projected_id
            and name == projected_name == seed_name
            and key == projected_key == seed_key
            and created_by_id == owner_id
            and binding is runtime_binding
            and provisioned_binding is runtime_binding
        )
    except Exception:
        return False


def _state_matches_safe(
    state: object,
    expected: object,
    runtime_binding: object,
) -> bool:
    try:
        details = _state_details(state)
        if not isinstance(details, tuple):
            return False
        if (
            type(expected) is not tuple
            or len(expected) != 4
            or any(type(value) is not int for value in expected)
        ):
            return False
        return details[:4] == expected and details[4] is runtime_binding
    except Exception:
        return False


def _exchange_parts_safe(exchange: object) -> tuple[int, bytes] | object:
    try:
        if type(exchange) is not HttpExchange:
            return _INVALID
        normalized = HttpExchange(
            response=exchange.response,
            diagnostic=exchange.diagnostic,
        )
        return normalized.response.status_code, normalized.response.body
    except Exception:
        return _INVALID


def _load_document_safe(body: object) -> dict[str, object] | object:
    try:
        if type(body) is not bytes or not 1 <= len(body) <= _MAX_DOCUMENT_BYTES:
            return _INVALID
        if body.startswith(b"\xef\xbb\xbf") or b"\x00" in body:
            return _INVALID
        text = body.decode("utf-8", errors="strict")
        document = json.loads(
            text,
            object_pairs_hook=_pairs_without_duplicates,
            parse_constant=_reject_constant,
        )
        if type(document) is not dict or not _shape_is_bounded_safe(document):
            return _INVALID
        return document
    except Exception:
        return _INVALID


def _pairs_without_duplicates(pairs: list[tuple[str, object]]) -> dict[str, object]:
    result: dict[str, object] = {}
    for key, value in pairs:
        if type(key) is not str or key in result:
            raise ValueError("invalid document")
        result[key] = value
    return result


def _reject_constant(value: str) -> NoReturn:
    del value
    raise ValueError("invalid document")


def _shape_is_bounded_safe(value: object) -> bool:
    try:
        stack: list[tuple[object, int]] = [(value, 1)]
        seen = 0
        while stack:
            current, depth = stack.pop()
            seen += 1
            if seen > _MAX_NODES or depth > _MAX_DEPTH:
                return False
            if isinstance(current, Mapping):
                if type(current) is not dict:
                    return False
                for key, child in current.items():
                    if type(key) is not str or len(key) > 128:
                        return False
                    stack.append((child, depth + 1))
            elif type(current) is list:
                if len(current) > _MAX_NODES:
                    return False
                stack.extend((child, depth + 1) for child in current)
            elif current is None or type(current) in {str, int, float, bool}:
                if type(current) is str and len(current) > 8192:
                    return False
            else:
                return False
        return True
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise ProjectOutboxOracleError("project outbox observation is invalid") from None
