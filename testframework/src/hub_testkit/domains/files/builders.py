from __future__ import annotations

import json
from typing import NoReturn

from hub_testkit.http import Actor, HttpRequest

from .models import FilesPrincipal, _canonical_uuid_safe, _text_safe


class FilesBuilderError(ValueError):
    pass


_INVALID = object()


def actor_from_principal(principal: FilesPrincipal) -> Actor:
    result = _actor_safe(principal)
    del principal
    if type(result) is Actor:
        return result
    del result
    _raise_invalid()


def upload_request(
    actor: Actor,
    project_id: str,
    filename: str,
    content_type: str,
    size: int,
    correlation_id: str,
) -> HttpRequest:
    result = _upload_request_safe(
        actor,
        project_id,
        filename,
        content_type,
        size,
        correlation_id,
    )
    del actor, project_id, filename, content_type, size, correlation_id
    if type(result) is HttpRequest:
        return result
    del result
    _raise_invalid()


def confirm_request(
    actor: Actor | None,
    file_id: str,
    correlation_id: str,
) -> HttpRequest:
    result = _file_request_safe(
        "POST",
        "confirm",
        actor,
        file_id,
        correlation_id,
        b'{"entityType":"","entityId":""}',
    )
    del actor, file_id, correlation_id
    if type(result) is HttpRequest:
        return result
    del result
    _raise_invalid()


def download_request(
    actor: Actor | None,
    file_id: str,
    correlation_id: str,
) -> HttpRequest:
    result = _file_request_safe(
        "GET",
        "url",
        actor,
        file_id,
        correlation_id,
        None,
    )
    del actor, file_id, correlation_id
    if type(result) is HttpRequest:
        return result
    del result
    _raise_invalid()


def delete_request(
    actor: Actor | None,
    file_id: str,
    correlation_id: str,
) -> HttpRequest:
    result = _file_request_safe(
        "DELETE",
        "delete",
        actor,
        file_id,
        correlation_id,
        None,
    )
    del actor, file_id, correlation_id
    if type(result) is HttpRequest:
        return result
    del result
    _raise_invalid()


def _actor_safe(principal: object) -> Actor | object:
    try:
        if type(principal) is not FilesPrincipal:
            return _INVALID
        return Actor(subject=principal.user_id, token=principal.access_token)
    except Exception:
        return _INVALID


def _upload_request_safe(
    actor: object,
    project_id: object,
    filename: object,
    content_type: object,
    size: object,
    correlation_id: object,
) -> HttpRequest | object:
    try:
        if type(actor) is not Actor or not _canonical_uuid_safe(project_id):
            return _INVALID
        if not _text_safe(filename, 255) or not _text_safe(content_type, 128):
            return _INVALID
        if type(size) is not int or not 0 <= size <= 25 * 1024 * 1024 + 1:
            return _INVALID
        if type(correlation_id) is not str:
            return _INVALID
        body = json.dumps(
            {
                "filename": filename,
                "contentType": content_type,
                "projectId": project_id,
                "size": size,
            },
            ensure_ascii=True,
            separators=(",", ":"),
        ).encode("utf-8")
        result = HttpRequest(
            method="POST",
            path="/api/files/upload-url",
            actor=actor,
            correlation_id=correlation_id,
            idempotency_key=None,
            body=body,
            content_type="application/json",
        )
        del body
        return result
    except Exception:
        return _INVALID


def _file_request_safe(
    method: object,
    operation: object,
    actor: object,
    file_id: object,
    correlation_id: object,
    body: object,
) -> HttpRequest | object:
    try:
        if type(method) is not str or type(operation) is not str:
            return _INVALID
        if not _canonical_uuid_safe(file_id):
            return _INVALID
        if actor is not None and type(actor) is not Actor:
            return _INVALID
        if type(correlation_id) is not str:
            return _INVALID
        if body is not None and type(body) is not bytes:
            return _INVALID
        if operation == "confirm":
            path = f"/api/files/{file_id}/confirm"
        elif operation == "url":
            path = f"/api/files/{file_id}/url"
        elif operation == "delete":
            path = f"/api/files/{file_id}"
        else:
            return _INVALID
        return HttpRequest(
            method=method,
            path=path,
            actor=actor,
            correlation_id=correlation_id,
            idempotency_key=None,
            body=body,
            content_type="application/json" if body is not None else None,
        )
    except Exception:
        return _INVALID


def _raise_invalid() -> NoReturn:
    raise FilesBuilderError("files request is invalid") from None
