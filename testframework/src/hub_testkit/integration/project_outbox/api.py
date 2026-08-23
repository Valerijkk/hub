from __future__ import annotations

from dataclasses import dataclass, field
from typing import NoReturn, Protocol, runtime_checkable

from hub_testkit.http import HttpExchange, HttpRequest

from .builders import (
    build_anonymous_project_create_request,
    build_authenticated_project_create_request,
    build_malformed_project_create_request,
)
from .models import ProjectOutboxSeed, ProvisionedProjectOutbox


class ProjectOutboxApiError(ValueError):
    pass


@runtime_checkable
class ProjectOutboxSender(Protocol):
    def send(self, request: HttpRequest) -> HttpExchange:
        """Send one already-admitted request through the injected boundary."""
        ...


@dataclass(frozen=True, slots=True)
class ProjectOutboxApi:
    sender: ProjectOutboxSender = field(repr=False)

    def __post_init__(self) -> None:
        result = _sender_is_valid_safe(self.sender)
        if result is True:
            return
        del result
        _raise_invalid()

    def create(
        self,
        seed: ProjectOutboxSeed,
        provisioned: ProvisionedProjectOutbox,
        correlation_id: str,
    ) -> HttpExchange:
        result = _operation_safe(
            self.sender,
            seed,
            provisioned,
            correlation_id,
            "authenticated",
        )
        del seed, provisioned, correlation_id
        if type(result) is HttpExchange:
            return result
        del result
        _raise_invalid()

    def create_anonymous(
        self,
        seed: ProjectOutboxSeed,
        provisioned: ProvisionedProjectOutbox,
        correlation_id: str,
    ) -> HttpExchange:
        result = _operation_safe(
            self.sender,
            seed,
            provisioned,
            correlation_id,
            "anonymous",
        )
        del seed, provisioned, correlation_id
        if type(result) is HttpExchange:
            return result
        del result
        _raise_invalid()

    def create_malformed(
        self,
        seed: ProjectOutboxSeed,
        provisioned: ProvisionedProjectOutbox,
        correlation_id: str,
    ) -> HttpExchange:
        result = _operation_safe(
            self.sender,
            seed,
            provisioned,
            correlation_id,
            "malformed",
        )
        del seed, provisioned, correlation_id
        if type(result) is HttpExchange:
            return result
        del result
        _raise_invalid()


def _operation_safe(
    sender: object,
    seed: object,
    provisioned: object,
    correlation_id: object,
    mode: object,
) -> HttpExchange | object:
    try:
        if (
            type(seed) is not ProjectOutboxSeed
            or type(provisioned) is not ProvisionedProjectOutbox
            or type(correlation_id) is not str
        ):
            return object()
        if mode == "authenticated":
            request = build_authenticated_project_create_request(
                seed,
                provisioned,
                correlation_id,
            )
        elif mode == "anonymous":
            request = build_anonymous_project_create_request(
                seed,
                provisioned,
                correlation_id,
            )
        elif mode == "malformed":
            request = build_malformed_project_create_request(
                seed,
                provisioned,
                correlation_id,
            )
        else:
            return object()
        return _send_safe(sender, request)
    except Exception:
        return object()


def _send_safe(sender: object, request: object) -> HttpExchange | object:
    try:
        if not isinstance(sender, ProjectOutboxSender) or type(request) is not HttpRequest:
            return object()
        exchange = sender.send(request)
        if type(exchange) is not HttpExchange:
            return object()
        normalized = HttpExchange(
            response=exchange.response,
            diagnostic=exchange.diagnostic,
        )
        if (
            normalized.diagnostic.method != "POST"
            or normalized.diagnostic.path != "/api/projects"
            or normalized.diagnostic.correlation_id != request.correlation_id
            or normalized.diagnostic.status_code != normalized.response.status_code
        ):
            return object()
        return normalized
    except Exception:
        return object()


def _sender_is_valid_safe(sender: object) -> bool:
    try:
        return isinstance(sender, ProjectOutboxSender)
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise ProjectOutboxApiError("project outbox API operation failed") from None
