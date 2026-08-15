from __future__ import annotations

import re
from collections.abc import Mapping
from dataclasses import dataclass, field
from typing import Literal, Protocol, cast

import httpx

from hub_testkit.runtime.redaction import Redactor

from .auth import Actor
from .errors import HttpDiagnosticError, HttpRequestError, HttpTransportError
from .policy import HttpPolicy


_METHODS = frozenset({"GET", "HEAD", "OPTIONS", "POST", "PUT", "PATCH", "DELETE"})
_IDEMPOTENT_METHODS = frozenset({"GET", "HEAD", "OPTIONS", "PUT", "DELETE"})
_KEYED_METHODS = frozenset({"POST", "PATCH"})
_RETRY_STATUSES = frozenset({502, 503, 504})
_RETRY_EXCEPTIONS = (
    httpx.ConnectTimeout,
    httpx.ReadTimeout,
    httpx.WriteTimeout,
    httpx.PoolTimeout,
    httpx.ConnectError,
    httpx.ReadError,
    httpx.WriteError,
    httpx.RemoteProtocolError,
)
_CORRELATION_RE = re.compile(r"^[A-Za-z0-9._-]{1,128}$")
_IDEMPOTENCY_RE = re.compile(r"^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$")
_RETRY_CLASSES = frozenset({"none", "status", "transport"})


@dataclass(frozen=True, slots=True)
class HttpRequest:
    method: str
    path: str
    actor: Actor | None = field(repr=False)
    correlation_id: str
    idempotency_key: str | None = field(repr=False)
    body: bytes | None = field(repr=False)
    content_type: Literal["application/json"] | None = None

    def __post_init__(self) -> None:
        try:
            _require_method(self.method)
            _require_path(self.path)
            if self.actor is not None:
                if type(self.actor) is not Actor:
                    raise HttpRequestError("HTTP request is invalid")
                object.__setattr__(
                    self,
                    "actor",
                    Actor(subject=self.actor.subject, token=self.actor.token),
                )
            _require_correlation_id(self.correlation_id)
            _require_idempotency_key(self.method, self.idempotency_key)
            if self.body is not None and type(self.body) is not bytes:
                raise HttpRequestError("HTTP request is invalid")
            if self.content_type is not None and (
                type(self.content_type) is not str
                or self.content_type != "application/json"
                or type(self.body) is not bytes
            ):
                raise HttpRequestError("HTTP request is invalid")
        except Exception:
            raise HttpRequestError("HTTP request is invalid") from None


@dataclass(frozen=True, slots=True)
class HttpResponse:
    status_code: int
    body: bytes

    def __post_init__(self) -> None:
        try:
            if type(self.status_code) is not int or not 100 <= self.status_code <= 599:
                raise HttpRequestError("HTTP response is invalid")
            if type(self.body) is not bytes:
                raise HttpRequestError("HTTP response is invalid")
        except Exception:
            raise HttpRequestError("HTTP response is invalid") from None


@dataclass(frozen=True, slots=True)
class HttpDiagnostic:
    method: str
    path: str
    status_code: int | None
    attempts: int
    retry_class: Literal["none", "status", "transport"]
    correlation_id: str

    def __post_init__(self) -> None:
        try:
            _require_method(self.method)
            _require_path(self.path)
            if self.status_code is not None and (
                type(self.status_code) is not int or not 100 <= self.status_code <= 599
            ):
                raise HttpDiagnosticError("HTTP diagnostic is invalid")
            if type(self.attempts) is not int or self.attempts not in {1, 2}:
                raise HttpDiagnosticError("HTTP diagnostic is invalid")
            if type(self.retry_class) is not str or self.retry_class not in _RETRY_CLASSES:
                raise HttpDiagnosticError("HTTP diagnostic is invalid")
            if self.attempts == 1 and self.retry_class != "none":
                raise HttpDiagnosticError("HTTP diagnostic is invalid")
            if self.attempts == 2 and self.retry_class == "none":
                raise HttpDiagnosticError("HTTP diagnostic is invalid")
            _require_correlation_id(self.correlation_id)
        except Exception:
            raise HttpDiagnosticError("HTTP diagnostic is invalid") from None


@dataclass(frozen=True, slots=True)
class HttpExchange:
    response: HttpResponse
    diagnostic: HttpDiagnostic

    def __post_init__(self) -> None:
        try:
            if (
                type(self.response) is not HttpResponse
                or type(self.diagnostic) is not HttpDiagnostic
            ):
                raise HttpTransportError("HTTP exchange is invalid")
            object.__setattr__(
                self,
                "response",
                HttpResponse(
                    status_code=self.response.status_code,
                    body=self.response.body,
                ),
            )
            object.__setattr__(
                self,
                "diagnostic",
                HttpDiagnostic(
                    method=self.diagnostic.method,
                    path=self.diagnostic.path,
                    status_code=self.diagnostic.status_code,
                    attempts=self.diagnostic.attempts,
                    retry_class=self.diagnostic.retry_class,
                    correlation_id=self.diagnostic.correlation_id,
                ),
            )
        except Exception:
            raise HttpTransportError("HTTP exchange is invalid") from None


class HttpClientProtocol(Protocol):
    def request(
        self,
        method: str,
        path: str,
        *,
        headers: Mapping[str, str],
        content: bytes | None,
        timeout: httpx.Timeout,
    ) -> httpx.Response:
        ...


@dataclass(frozen=True, slots=True)
class HttpxTransport:
    client: HttpClientProtocol
    policy: HttpPolicy
    redactor: Redactor

    def __post_init__(self) -> None:
        try:
            if (
                self.client is None
                or type(self.policy) is not HttpPolicy
                or type(self.redactor) is not Redactor
            ):
                raise HttpTransportError("HTTP transport is invalid")
            object.__setattr__(
                self,
                "policy",
                HttpPolicy(
                    timeout=self.policy.timeout,
                    retry=self.policy.retry,
                ),
            )
        except Exception:
            raise HttpTransportError("HTTP transport is invalid") from None

    def send(self, request: HttpRequest) -> HttpExchange:
        try:
            if type(request) is not HttpRequest:
                raise HttpTransportError("HTTP transport is invalid")
            if (
                self.client is None
                or type(self.policy) is not HttpPolicy
                or type(self.redactor) is not Redactor
            ):
                raise HttpTransportError("HTTP transport is invalid")
            policy = HttpPolicy(timeout=self.policy.timeout, retry=self.policy.retry)
            client = self.client
            redactor = self.redactor
            request = HttpRequest(
                method=request.method,
                path=request.path,
                actor=request.actor,
                correlation_id=request.correlation_id,
                idempotency_key=request.idempotency_key,
                body=request.body,
                content_type=request.content_type,
            )
            headers = _headers_for(request)
            may_retry = _may_retry(request)
            retry_class: Literal["none", "status", "transport"] = "none"
            for attempt in range(1, policy.retry.max_attempts + 1):
                try:
                    raw_response = client.request(
                        request.method,
                        request.path,
                        headers=headers,
                        content=request.body,
                        timeout=policy.timeout.to_httpx_timeout(),
                    )
                    response = _response_from_httpx(raw_response)
                except Exception as error:
                    if attempt == 1 and may_retry and _is_retryable_exception(error):
                        retry_class = "transport"
                        continue
                    raise HttpTransportError("HTTP transport failed") from None
                if attempt == 1 and may_retry and response.status_code in _RETRY_STATUSES:
                    retry_class = "status"
                    continue
                raw_diagnostic = HttpDiagnostic(
                    method=request.method,
                    path=request.path,
                    status_code=response.status_code,
                    attempts=attempt,
                    retry_class=retry_class,
                    correlation_id=request.correlation_id,
                )
                return HttpExchange(
                    response=response,
                    diagnostic=_sanitize_diagnostic(raw_diagnostic, redactor),
                )
        except HttpTransportError:
            raise
        except Exception:
            raise HttpTransportError("HTTP transport failed") from None
        raise HttpTransportError("HTTP transport failed") from None


def _require_method(value: object) -> None:
    if type(value) is not str or value not in _METHODS:
        raise HttpRequestError("HTTP request is invalid")


def _require_path(value: object) -> None:
    if type(value) is not str or not value.startswith("/") or value.startswith("//"):
        raise HttpRequestError("HTTP request is invalid")
    if "//" in value or "://" in value or "?" in value or "#" in value:
        raise HttpRequestError("HTTP request is invalid")
    if any(ord(character) < 32 or ord(character) == 127 for character in value):
        raise HttpRequestError("HTTP request is invalid")


def _require_correlation_id(value: object) -> None:
    if type(value) is not str or _CORRELATION_RE.fullmatch(value) is None:
        raise HttpRequestError("HTTP request is invalid")


def _require_idempotency_key(method: str, value: object) -> None:
    if value is None:
        return
    if method not in _KEYED_METHODS:
        raise HttpRequestError("HTTP request is invalid")
    if type(value) is not str or _IDEMPOTENCY_RE.fullmatch(value) is None:
        raise HttpRequestError("HTTP request is invalid")


def _headers_for(request: HttpRequest) -> dict[str, str]:
    headers = {"X-Request-ID": request.correlation_id}
    if request.actor is not None:
        headers.update(request.actor.headers())
    if request.idempotency_key is not None:
        headers["Idempotency-Key"] = request.idempotency_key
    if request.content_type is not None:
        headers["Content-Type"] = request.content_type
    return headers


def _may_retry(request: HttpRequest) -> bool:
    return request.method in _IDEMPOTENT_METHODS or request.idempotency_key is not None


def _is_retryable_exception(error: Exception) -> bool:
    return type(error) in _RETRY_EXCEPTIONS


def _response_from_httpx(response: object) -> HttpResponse:
    try:
        if type(response) is not httpx.Response:
            raise HttpTransportError("HTTP transport failed")
        return HttpResponse(status_code=response.status_code, body=response.content)
    except Exception:
        raise HttpTransportError("HTTP transport failed") from None


def _sanitize_diagnostic(diagnostic: HttpDiagnostic, redactor: Redactor) -> HttpDiagnostic:
    try:
        if type(diagnostic) is not HttpDiagnostic or type(redactor) is not Redactor:
            raise HttpDiagnosticError("HTTP diagnostic is invalid")
        redacted = redactor.redact_mapping(
            {
                "method": diagnostic.method,
                "path": diagnostic.path,
                "status_code": diagnostic.status_code,
                "attempts": diagnostic.attempts,
                "retry_class": diagnostic.retry_class,
                "correlation_id": diagnostic.correlation_id,
            }
        )
        if type(redacted) is not dict or tuple(redacted) != (
            "method",
            "path",
            "status_code",
            "attempts",
            "retry_class",
            "correlation_id",
        ):
            raise HttpDiagnosticError("HTTP diagnostic is invalid")
        method = redacted["method"]
        path = redacted["path"]
        status_code = redacted["status_code"]
        attempts = redacted["attempts"]
        retry_class = redacted["retry_class"]
        correlation_id = redacted["correlation_id"]
        if type(method) is not str or type(path) is not str or type(attempts) is not int:
            raise HttpDiagnosticError("HTTP diagnostic is invalid")
        if status_code is not None and type(status_code) is not int:
            raise HttpDiagnosticError("HTTP diagnostic is invalid")
        if type(retry_class) is not str or type(correlation_id) is not str:
            raise HttpDiagnosticError("HTTP diagnostic is invalid")
        return HttpDiagnostic(
            method=method,
            path=path,
            status_code=status_code,
            attempts=attempts,
            retry_class=cast(Literal["none", "status", "transport"], retry_class),
            correlation_id=correlation_id,
        )
    except Exception:
        raise HttpDiagnosticError("HTTP diagnostic is invalid") from None
