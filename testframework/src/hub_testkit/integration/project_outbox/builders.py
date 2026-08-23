from __future__ import annotations

import json
from typing import NoReturn, cast

from hub_testkit.http import Actor, HttpRequest

from .models import (
    ProjectOutboxFlow,
    ProjectOutboxSeed,
    ProvisionedProjectOutbox,
    _provisioned_details,
    _seed_details,
)


class ProjectOutboxBuilderError(ValueError):
    pass


_INVALID = object()
_FLOW_LABELS: dict[ProjectOutboxFlow, str] = {
    "create": "create",
    "duplicate": "duplicate",
    "anonymous": "anonymous",
    "malformed": "malformed",
    "observer-replay": "observer-replay",
}


def build_project_outbox_seed(
    flow: ProjectOutboxFlow,
    nonce: str,
) -> ProjectOutboxSeed:
    result = _seed_safe(flow, nonce)
    del flow, nonce
    if type(result) is ProjectOutboxSeed:
        return result
    del result
    _raise_invalid()


def build_authenticated_project_create_request(
    seed: ProjectOutboxSeed,
    provisioned: ProvisionedProjectOutbox,
    correlation_id: str,
) -> HttpRequest:
    result = _request_safe(seed, provisioned, correlation_id, "authenticated")
    del seed, provisioned, correlation_id
    if type(result) is HttpRequest:
        return result
    del result
    _raise_invalid()


def build_anonymous_project_create_request(
    seed: ProjectOutboxSeed,
    provisioned: ProvisionedProjectOutbox,
    correlation_id: str,
) -> HttpRequest:
    result = _request_safe(seed, provisioned, correlation_id, "anonymous")
    del seed, provisioned, correlation_id
    if type(result) is HttpRequest:
        return result
    del result
    _raise_invalid()


def build_malformed_project_create_request(
    seed: ProjectOutboxSeed,
    provisioned: ProvisionedProjectOutbox,
    correlation_id: str,
) -> HttpRequest:
    result = _request_safe(seed, provisioned, correlation_id, "malformed")
    del seed, provisioned, correlation_id
    if type(result) is HttpRequest:
        return result
    del result
    _raise_invalid()


def _seed_safe(flow: object, nonce: object) -> ProjectOutboxSeed | object:
    try:
        if type(flow) is not str or flow not in _FLOW_LABELS:
            return _INVALID
        if type(nonce) is not str or len(nonce) != 32:
            return _INVALID
        typed_flow = cast(ProjectOutboxFlow, flow)
        label = _FLOW_LABELS[typed_flow]
        project_name = f"Hub QA {label} {nonce[:8]}"
        project_key = f"Q{nonce[:9].upper()}"
        return ProjectOutboxSeed._create(
            label=label,
            flow=typed_flow,
            nonce=nonce,
            project_name=project_name,
            project_key=project_key,
            description=f"Owned {label} project",
        )
    except Exception:
        return _INVALID


def _request_safe(
    seed: object,
    provisioned: object,
    correlation_id: object,
    mode: object,
) -> HttpRequest | object:
    try:
        seed_details = _seed_details(seed)
        provisioned_details = _provisioned_details(provisioned)
        if not isinstance(seed_details, tuple) or not isinstance(
            provisioned_details,
            tuple,
        ):
            return _INVALID
        if type(correlation_id) is not str or mode not in {
            "authenticated",
            "anonymous",
            "malformed",
        }:
            return _INVALID
        _, _, seed_nonce, name, key, description = seed_details
        owner_id, owner_subject, owner_token, provisioned_nonce, _ = (
            provisioned_details
        )
        if seed_nonce != provisioned_nonce or owner_id != owner_subject:
            return _INVALID
        actor = None
        if mode != "anonymous":
            actor = Actor(subject=owner_subject, token=owner_token)
        body = b'{"name":'
        if mode != "malformed":
            document = {
                "name": name,
                "key": key,
                "description": description,
            }
            body = json.dumps(
                document,
                ensure_ascii=True,
                separators=(",", ":"),
            ).encode("utf-8")
        return HttpRequest(
            method="POST",
            path="/api/projects",
            actor=actor,
            correlation_id=correlation_id,
            idempotency_key=None,
            body=body,
            content_type="application/json",
        )
    except Exception:
        return _INVALID


def _raise_invalid() -> NoReturn:
    raise ProjectOutboxBuilderError("project outbox request is invalid") from None
