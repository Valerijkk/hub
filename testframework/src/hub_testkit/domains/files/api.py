from __future__ import annotations

from dataclasses import dataclass, field
from typing import NoReturn, Protocol

from hub_testkit.http import (
    Actor,
    HttpDiagnostic,
    HttpExchange,
    HttpRequest,
    HttpResponse,
)

from .builders import (
    actor_from_principal,
    confirm_request,
    delete_request,
    download_request,
    upload_request,
)
from .models import DownloadTicket, FileStatus, FilesPrincipal, UploadTicket
from .oracles import project_confirm, project_delete, project_download, project_upload


class FilesApiError(ValueError):
    pass


class FilesSender(Protocol):
    def send(self, request: HttpRequest) -> HttpExchange:
        """Send one admitted files request through the injected runtime boundary."""
        ...


_INVALID = object()


@dataclass(frozen=True, slots=True, repr=False, eq=False)
class FilesApi:
    sender: FilesSender = field(repr=False)
    _runtime_binding: object = field(repr=False)

    def __post_init__(self) -> None:
        valid = _sender_is_valid_safe(self.sender) and self._runtime_binding is not None
        if valid is True:
            return
        del valid
        _raise_invalid()

    def upload(
        self,
        principal: FilesPrincipal,
        project_id: str,
        filename: str,
        content_type: str,
        size: int,
        correlation_id: str,
        forbidden_values: tuple[str, ...] = (),
    ) -> UploadTicket | FileStatus:
        result = _upload_safe(
            self,
            principal,
            project_id,
            filename,
            content_type,
            size,
            correlation_id,
            forbidden_values,
        )
        del principal, project_id, filename, content_type, size, correlation_id
        del forbidden_values
        if type(result) in {UploadTicket, FileStatus}:
            return result
        del result
        _raise_invalid()

    def confirm(
        self,
        principal: FilesPrincipal | None,
        file_id: str,
        correlation_id: str,
        forbidden_values: tuple[str, ...] = (),
    ) -> FileStatus:
        result = _confirm_safe(
            self,
            principal,
            file_id,
            correlation_id,
            forbidden_values,
        )
        del principal, file_id, correlation_id, forbidden_values
        if type(result) is FileStatus:
            return result
        del result
        _raise_invalid()

    def download(
        self,
        principal: FilesPrincipal | None,
        file_id: str,
        correlation_id: str,
        forbidden_values: tuple[str, ...] = (),
    ) -> DownloadTicket | FileStatus:
        result = _download_safe(
            self,
            principal,
            file_id,
            correlation_id,
            forbidden_values,
        )
        del principal, file_id, correlation_id, forbidden_values
        if type(result) in {DownloadTicket, FileStatus}:
            return result
        del result
        _raise_invalid()

    def delete(
        self,
        principal: FilesPrincipal | None,
        file_id: str,
        correlation_id: str,
        forbidden_values: tuple[str, ...] = (),
    ) -> FileStatus:
        result = _delete_safe(
            self,
            principal,
            file_id,
            correlation_id,
            forbidden_values,
        )
        del principal, file_id, correlation_id, forbidden_values
        if type(result) is FileStatus:
            return result
        del result
        _raise_invalid()


def _upload_safe(
    api: FilesApi,
    principal: object,
    project_id: object,
    filename: object,
    content_type: object,
    size: object,
    correlation_id: object,
    forbidden_values: object,
) -> UploadTicket | FileStatus | object:
    try:
        if type(principal) is not FilesPrincipal:
            return _INVALID
        if not _arguments_are_typed_safe(
            project_id,
            filename,
            content_type,
            size,
            correlation_id,
            forbidden_values,
        ):
            return _INVALID
        actor = actor_from_principal(principal)
        request = upload_request(
            actor,
            project_id,
            filename,
            content_type,
            size,
            correlation_id,
        )
        exchange = _send_safe(api.sender, request)
        del actor, request
        if type(exchange) is not HttpExchange:
            return _INVALID
        result = project_upload(
            exchange,
            api._runtime_binding,
            project_id,
            content_type,
            size,
            correlation_id,
            forbidden_values,
        )
        del exchange
        return result
    except Exception:
        return _INVALID


def _confirm_safe(
    api: FilesApi,
    principal: object,
    file_id: object,
    correlation_id: object,
    forbidden_values: object,
) -> FileStatus | object:
    try:
        actor = _optional_actor_safe(principal)
        if actor is _INVALID or type(file_id) is not str or type(correlation_id) is not str:
            return _INVALID
        if not _forbidden_values_safe(forbidden_values):
            return _INVALID
        request = confirm_request(actor, file_id, correlation_id)
        exchange = _send_safe(api.sender, request)
        del actor, request
        if type(exchange) is not HttpExchange:
            return _INVALID
        result = project_confirm(exchange, file_id, correlation_id, forbidden_values)
        del exchange
        return result
    except Exception:
        return _INVALID


def _download_safe(
    api: FilesApi,
    principal: object,
    file_id: object,
    correlation_id: object,
    forbidden_values: object,
) -> DownloadTicket | FileStatus | object:
    try:
        actor = _optional_actor_safe(principal)
        if actor is _INVALID or type(file_id) is not str or type(correlation_id) is not str:
            return _INVALID
        if not _forbidden_values_safe(forbidden_values):
            return _INVALID
        request = download_request(actor, file_id, correlation_id)
        exchange = _send_safe(api.sender, request)
        del actor, request
        if type(exchange) is not HttpExchange:
            return _INVALID
        result = project_download(
            exchange,
            api._runtime_binding,
            file_id,
            correlation_id,
            forbidden_values,
        )
        del exchange
        return result
    except Exception:
        return _INVALID


def _delete_safe(
    api: FilesApi,
    principal: object,
    file_id: object,
    correlation_id: object,
    forbidden_values: object,
) -> FileStatus | object:
    try:
        actor = _optional_actor_safe(principal)
        if actor is _INVALID or type(file_id) is not str or type(correlation_id) is not str:
            return _INVALID
        if not _forbidden_values_safe(forbidden_values):
            return _INVALID
        request = delete_request(actor, file_id, correlation_id)
        exchange = _send_safe(api.sender, request)
        del actor, request
        if type(exchange) is not HttpExchange:
            return _INVALID
        result = project_delete(exchange, file_id, correlation_id, forbidden_values)
        del exchange
        return result
    except Exception:
        return _INVALID


def _optional_actor_safe(principal: object) -> Actor | None | object:
    try:
        if principal is None:
            return None
        if type(principal) is not FilesPrincipal:
            return _INVALID
        return actor_from_principal(principal)
    except Exception:
        return _INVALID


def _send_safe(sender: object, request: object) -> HttpExchange | object:
    try:
        if type(request) is not HttpRequest:
            return _INVALID
        send = getattr(sender, "send", None)
        if not callable(send):
            return _INVALID
        exchange = send(request)
        if type(exchange) is not HttpExchange:
            return _INVALID
        response = exchange.response
        diagnostic = exchange.diagnostic
        result = HttpExchange(
            response=HttpResponse(response.status_code, response.body),
            diagnostic=HttpDiagnostic(
                method=diagnostic.method,
                path=diagnostic.path,
                status_code=diagnostic.status_code,
                attempts=diagnostic.attempts,
                retry_class=diagnostic.retry_class,
                correlation_id=diagnostic.correlation_id,
            ),
        )
        del exchange, response, diagnostic
        return result
    except Exception:
        return _INVALID


def _sender_is_valid_safe(sender: object) -> bool:
    try:
        return sender is not None and callable(getattr(sender, "send", None))
    except Exception:
        return False


def _arguments_are_typed_safe(
    project_id: object,
    filename: object,
    content_type: object,
    size: object,
    correlation_id: object,
    forbidden_values: object,
) -> bool:
    return (
        type(project_id) is str
        and type(filename) is str
        and type(content_type) is str
        and type(size) is int
        and type(correlation_id) is str
        and _forbidden_values_safe(forbidden_values)
    )


def _forbidden_values_safe(values: object) -> bool:
    try:
        return type(values) is tuple and all(type(value) is str and value for value in values)
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise FilesApiError("files API boundary is invalid") from None
