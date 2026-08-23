from __future__ import annotations

import json
from typing import NoReturn

from hub_testkit.http import HttpExchange

from .models import (
    DownloadTicket,
    FileStatus,
    UploadTicket,
    _text_safe,
)


class FilesOracleError(ValueError):
    pass


_BODY_LIMIT_BYTES = 64 * 1024
_INVALID = object()


def project_upload(
    exchange: HttpExchange,
    runtime_binding: object,
    project_id: str,
    content_type: str,
    expected_length: int,
    correlation_id: str,
    forbidden_values: tuple[str, ...] = (),
) -> UploadTicket | FileStatus:
    result = _project_upload_safe(
        exchange,
        runtime_binding,
        project_id,
        content_type,
        expected_length,
        correlation_id,
        forbidden_values,
    )
    del exchange, runtime_binding, project_id, content_type, expected_length
    del correlation_id
    del forbidden_values
    if type(result) in {UploadTicket, FileStatus}:
        return result
    del result
    _raise_invalid()


def project_confirm(
    exchange: HttpExchange,
    file_id: str,
    correlation_id: str,
    forbidden_values: tuple[str, ...] = (),
) -> FileStatus:
    result = _project_status_safe(
        exchange,
        "POST",
        file_id,
        "confirm",
        correlation_id,
        {200, 401, 403, 404, 409},
        "confirmed",
        forbidden_values,
    )
    del exchange, file_id, correlation_id, forbidden_values
    if type(result) is FileStatus:
        return result
    del result
    _raise_invalid()


def project_download(
    exchange: HttpExchange,
    runtime_binding: object,
    file_id: str,
    correlation_id: str,
    forbidden_values: tuple[str, ...] = (),
) -> DownloadTicket | FileStatus:
    result = _project_download_safe(
        exchange,
        runtime_binding,
        file_id,
        correlation_id,
        forbidden_values,
    )
    del exchange, runtime_binding, file_id, correlation_id, forbidden_values
    if type(result) in {DownloadTicket, FileStatus}:
        return result
    del result
    _raise_invalid()


def project_delete(
    exchange: HttpExchange,
    file_id: str,
    correlation_id: str,
    forbidden_values: tuple[str, ...] = (),
) -> FileStatus:
    result = _project_status_safe(
        exchange,
        "DELETE",
        file_id,
        "delete",
        correlation_id,
        {200, 401, 403, 404},
        "deleted",
        forbidden_values,
    )
    del exchange, file_id, correlation_id, forbidden_values
    if type(result) is FileStatus:
        return result
    del result
    _raise_invalid()


def _project_upload_safe(
    exchange: object,
    runtime_binding: object,
    project_id: object,
    content_type: object,
    expected_length: object,
    correlation_id: object,
    forbidden_values: object,
) -> UploadTicket | FileStatus | object:
    try:
        if not _exchange_boundary_safe(
            exchange,
            "POST",
            "/api/files/upload-url",
            correlation_id,
            forbidden_values,
        ):
            return _INVALID
        status = exchange.response.status_code
        if status == 200:
            document = _document_safe(exchange)
            if type(document) is not dict or set(document) != {
                "fileId",
                "uploadUrl",
                "s3Key",
            }:
                return _INVALID
            file_id = document["fileId"]
            upload_url = document["uploadUrl"]
            object_key = document["s3Key"]
            if (
                type(file_id) is not str
                or type(upload_url) is not str
                or type(object_key) is not str
                or type(project_id) is not str
                or type(content_type) is not str
                or type(expected_length) is not int
            ):
                return _INVALID
            expected_prefix = f"{project_id}/files/{file_id}/"
            if (
                not object_key.startswith(expected_prefix)
                or ".." in object_key
                or "\\" in object_key
            ):
                return _INVALID
            return UploadTicket._create(
                file_id=file_id,
                runtime_binding=runtime_binding,
                upload_url=upload_url,
                object_key=object_key,
                content_type=content_type,
                expected_length=expected_length,
            )
        if status not in {400, 401, 403} or not _error_document_safe(
            exchange,
            correlation_id,
        ):
            return _INVALID
        return FileStatus(status)
    except Exception:
        return _INVALID


def _project_download_safe(
    exchange: object,
    runtime_binding: object,
    file_id: object,
    correlation_id: object,
    forbidden_values: object,
) -> DownloadTicket | FileStatus | object:
    try:
        if type(file_id) is not str:
            return _INVALID
        path = f"/api/files/{file_id}/url"
        if not _exchange_boundary_safe(
            exchange,
            "GET",
            path,
            correlation_id,
            forbidden_values,
        ):
            return _INVALID
        status = exchange.response.status_code
        if status == 200:
            document = _document_safe(exchange)
            if type(document) is not dict or set(document) != {"downloadUrl"}:
                return _INVALID
            download_url = document["downloadUrl"]
            if type(download_url) is not str:
                return _INVALID
            return DownloadTicket._create(
                file_id=file_id,
                runtime_binding=runtime_binding,
                download_url=download_url,
            )
        if status not in {401, 403, 404} or not _error_document_safe(
            exchange,
            correlation_id,
        ):
            return _INVALID
        return FileStatus(status)
    except Exception:
        return _INVALID


def _project_status_safe(
    exchange: object,
    method: object,
    file_id: object,
    operation: object,
    correlation_id: object,
    allowed_statuses: object,
    success_value: object,
    forbidden_values: object,
) -> FileStatus | object:
    try:
        if type(file_id) is not str or type(operation) is not str:
            return _INVALID
        if operation == "confirm":
            path = f"/api/files/{file_id}/confirm"
        elif operation == "delete":
            path = f"/api/files/{file_id}"
        else:
            return _INVALID
        if not _exchange_boundary_safe(
            exchange,
            method,
            path,
            correlation_id,
            forbidden_values,
        ):
            return _INVALID
        status = exchange.response.status_code
        if type(allowed_statuses) is not set or status not in allowed_statuses:
            return _INVALID
        if status == 200:
            document = _document_safe(exchange)
            if document != {"status": success_value}:
                return _INVALID
        elif not _error_document_safe(exchange, correlation_id):
            return _INVALID
        return FileStatus(status)
    except Exception:
        return _INVALID


def _exchange_boundary_safe(
    exchange: object,
    method: object,
    path: object,
    correlation_id: object,
    forbidden_values: object,
) -> bool:
    try:
        if type(exchange) is not HttpExchange:
            return False
        diagnostic = exchange.diagnostic
        if (
            diagnostic.method != method
            or diagnostic.path != path
            or diagnostic.status_code != exchange.response.status_code
            or diagnostic.correlation_id != correlation_id
        ):
            return False
        if type(forbidden_values) is not tuple or not all(
            type(value) is str and value for value in forbidden_values
        ):
            return False
        rendered = f"{exchange!r}{exchange!s}{diagnostic!r}{diagnostic!s}"
        return all(value not in rendered for value in forbidden_values)
    except Exception:
        return False


def _document_safe(exchange: object) -> dict[str, object] | object:
    try:
        if type(exchange) is not HttpExchange:
            return _INVALID
        body = exchange.response.body
        if len(body) > _BODY_LIMIT_BYTES or body.startswith(b"\xef\xbb\xbf"):
            return _INVALID
        document = json.loads(
            body.decode("utf-8", errors="strict"),
            object_pairs_hook=_unique_object_safe,
        )
        del body
        if type(document) is not dict or _contains_invalid_safe(document):
            return _INVALID
        return document
    except Exception:
        return _INVALID


def _error_document_safe(exchange: object, correlation_id: object) -> bool:
    try:
        document = _document_safe(exchange)
        if type(document) is not dict or set(document) != {"error", "request_id"}:
            return False
        return (
            _text_safe(document["error"], 1024)
            and type(correlation_id) is str
            and document["request_id"] == correlation_id
        )
    except Exception:
        return False


def _unique_object_safe(pairs: object) -> dict[str, object] | object:
    try:
        if type(pairs) is not list:
            return _INVALID
        document: dict[str, object] = {}
        for pair in pairs:
            if type(pair) is not tuple or len(pair) != 2:
                return _INVALID
            key, value = pair
            if type(key) is not str or key in document:
                return _INVALID
            document[key] = value
        return document
    except Exception:
        return _INVALID


def _contains_invalid_safe(value: object) -> bool:
    try:
        if value is _INVALID:
            return True
        if type(value) is dict:
            return any(_contains_invalid_safe(item) for item in value.values())
        if type(value) is list:
            return any(_contains_invalid_safe(item) for item in value)
        return False
    except Exception:
        return True


def _raise_invalid() -> NoReturn:
    raise FilesOracleError("files response is invalid") from None
