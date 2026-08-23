from __future__ import annotations

import json
from collections.abc import Callable

from hub_testkit.http import HttpExchange

from .models import (
    ActionMessage,
    ErrorShape,
    ManagementSuccess,
    PendingPrincipal,
    ProvisionedPrincipal,
    RegistrationMode,
)


class AuthzOracleError(ValueError):
    pass


_BODY_LIMIT_BYTES = 64 * 1024
_PENDING_LIMIT = 256
_INVALID = object()


def parse_registration_mode(exchange: HttpExchange) -> RegistrationMode:
    try:
        result = _registration_mode_safe(exchange)
    except Exception:
        result = _INVALID
    del exchange
    if type(result) is not RegistrationMode:
        del result
        _raise_invalid()
    return result


def parse_pending_principals(exchange: HttpExchange) -> tuple[PendingPrincipal, ...]:
    try:
        result = _pending_principals_safe(exchange)
    except Exception:
        result = _INVALID
    del exchange
    if type(result) is not tuple:
        del result
        _raise_invalid()
    return result


def parse_success_message(exchange: HttpExchange) -> ActionMessage:
    try:
        result = _success_message_safe(exchange)
    except Exception:
        result = _INVALID
    del exchange
    if type(result) is not ActionMessage:
        del result
        _raise_invalid()
    return result


def parse_management_success(exchange: HttpExchange) -> ManagementSuccess:
    try:
        result = _management_success_safe(exchange)
    except Exception:
        result = _INVALID
    del exchange
    if type(result) is not ManagementSuccess:
        del result
        _raise_invalid()
    return result


def parse_error_shape(exchange: HttpExchange, expected_status: int) -> ErrorShape:
    try:
        result = _error_shape_safe(exchange, expected_status)
    except Exception:
        result = _INVALID
    del exchange, expected_status
    if type(result) is not ErrorShape:
        del result
        _raise_invalid()
    return result


def pending_contains_principals(
    exchange: HttpExchange,
    principals: tuple[ProvisionedPrincipal, ...],
) -> bool:
    try:
        result = _pending_membership_safe(exchange, principals, must_contain=True)
    except Exception:
        result = _INVALID
    del exchange, principals
    if type(result) is not bool:
        del result
        _raise_invalid()
    return result


def pending_excludes_principals(
    exchange: HttpExchange,
    principals: tuple[ProvisionedPrincipal, ...],
) -> bool:
    try:
        result = _pending_membership_safe(exchange, principals, must_contain=False)
    except Exception:
        result = _INVALID
    del exchange, principals
    if type(result) is not bool:
        del result
        _raise_invalid()
    return result


def error_shapes_match(left: ErrorShape, right: ErrorShape) -> bool:
    try:
        result = _error_shapes_match_safe(left, right)
    except Exception:
        result = _INVALID
    del left, right
    if type(result) is not bool:
        del result
        _raise_invalid()
    return result


def principal_repr_is_redacted(principal: ProvisionedPrincipal) -> bool:
    try:
        result = _principal_repr_is_redacted_safe(principal)
    except Exception:
        result = _INVALID
    del principal
    if type(result) is not bool:
        del result
        _raise_invalid()
    return result


def exchanges_hide_principal_canaries(
    exchange: HttpExchange,
    principal: ProvisionedPrincipal,
) -> bool:
    try:
        result = _exchange_hides_principal_canaries_safe(exchange, principal)
    except Exception:
        result = _INVALID
    del exchange, principal
    if type(result) is not bool:
        del result
        _raise_invalid()
    return result


def pending_approval_is_non_enumerating(
    known_exchange: HttpExchange,
    unknown_exchange: HttpExchange,
    principal: ProvisionedPrincipal,
) -> bool:
    try:
        result = _pending_approval_non_enumeration_safe(
            known_exchange,
            unknown_exchange,
            principal,
        )
    except Exception:
        result = _INVALID
    del known_exchange, unknown_exchange, principal
    if result is not True:
        del result
        _raise_invalid()
    return True


def _registration_mode_safe(exchange: object) -> RegistrationMode | object:
    return _decode_with_safe(exchange, 200, _registration_mode_from_document_safe, None)


def _pending_principals_safe(exchange: object) -> tuple[PendingPrincipal, ...] | object:
    return _decode_with_safe(exchange, 200, _pending_principals_from_document_safe, None)


def _success_message_safe(exchange: object) -> ActionMessage | object:
    return _decode_with_safe(exchange, 200, _success_message_from_document_safe, None)


def _management_success_safe(exchange: object) -> ManagementSuccess | object:
    return _decode_with_safe(exchange, 200, _management_success_from_document_safe, None)


def _error_shape_safe(exchange: object, expected_status: object) -> ErrorShape | object:
    if type(expected_status) is not int:
        return _INVALID
    status_code: int = expected_status
    return _decode_with_safe(
        exchange,
        status_code,
        _error_shape_from_document_safe,
        status_code,
    )


def _registration_mode_from_document_safe(
    document: object,
    _: object,
) -> RegistrationMode | object:
    try:
        if type(document) is not dict or set(document) != {"mode"}:
            return _INVALID
        return RegistrationMode(mode=document["mode"])
    except Exception:
        return _INVALID


def _pending_principals_from_document_safe(
    document: object,
    _: object,
) -> tuple[PendingPrincipal, ...] | object:
    try:
        if type(document) is not list or len(document) > _PENDING_LIMIT:
            return _INVALID
        principals: list[PendingPrincipal] = []
        for item in document:
            if type(item) is not dict or set(item) != {"id", "email", "pending_since"}:
                return _INVALID
            principals.append(
                PendingPrincipal(
                    user_id=item["id"],
                    email=item["email"],
                    pending_since=item["pending_since"],
                )
            )
        if len({principal.user_id for principal in principals}) != len(principals):
            return _INVALID
        return tuple(principals)
    except Exception:
        return _INVALID


def _success_message_from_document_safe(document: object, _: object) -> ActionMessage | object:
    try:
        if type(document) is not dict or set(document) != {"message"}:
            return _INVALID
        return ActionMessage(message=document["message"])
    except Exception:
        return _INVALID


def _management_success_from_document_safe(
    document: object,
    _: object,
) -> ManagementSuccess | object:
    try:
        if type(document) is not dict or set(document) != {"success", "message"}:
            return _INVALID
        if type(document["success"]) is not bool or type(document["message"]) is not str:
            return _INVALID
        return ManagementSuccess(success=document["success"])
    except Exception:
        return _INVALID


def _error_shape_from_document_safe(document: object, expected_status: object) -> ErrorShape | object:
    try:
        if type(expected_status) is not int:
            return _INVALID
        status_code: int = expected_status
        if type(document) is not dict or set(document) != {"error", "request_id"}:
            return _INVALID
        if type(document["error"]) is not str or type(document["request_id"]) is not str:
            return _INVALID
        return ErrorShape(status_code=status_code, canonical_message=document["error"])
    except Exception:
        return _INVALID


def _pending_membership_safe(
    exchange: object,
    principals: object,
    *,
    must_contain: object,
) -> bool | object:
    try:
        if type(principals) is not tuple or not principals or type(must_contain) is not bool:
            return _INVALID
        if any(type(principal) is not ProvisionedPrincipal for principal in principals):
            return _INVALID
        requested_ids = frozenset(principal.user_id for principal in principals)
        if len(requested_ids) != len(principals):
            return _INVALID
        observed = _pending_principals_safe(exchange)
        if type(observed) is not tuple:
            return _INVALID
        observed_ids = frozenset(principal.user_id for principal in observed)
        return requested_ids <= observed_ids if must_contain else requested_ids.isdisjoint(observed_ids)
    except Exception:
        return _INVALID


def _error_shapes_match_safe(left: object, right: object) -> bool | object:
    try:
        if type(left) is not ErrorShape or type(right) is not ErrorShape:
            return _INVALID
        return left.status_code == right.status_code and left.canonical_message == right.canonical_message
    except Exception:
        return _INVALID


def _principal_repr_is_redacted_safe(principal: object) -> bool | object:
    try:
        if type(principal) is not ProvisionedPrincipal or principal.state != "active":
            return _INVALID
        if len(principal.credentials) != 1:
            return _INVALID
        credential = principal.credentials[0]
        raw_values = (
            principal.email,
            principal.user_id,
            credential.access_token,
            credential.refresh_token,
            principal.nonce,
        )
        rendered = repr(principal)
        return all(value not in rendered for value in raw_values)
    except Exception:
        return _INVALID


def _exchange_hides_principal_canaries_safe(
    exchange: object,
    principal: object,
) -> bool | object:
    try:
        if type(exchange) is not HttpExchange or type(principal) is not ProvisionedPrincipal:
            return _INVALID
        if principal.state != "active" or len(principal.credentials) != 1:
            return _INVALID
        credential = principal.credentials[0]
        raw_values = (
            principal.email,
            principal.user_id,
            credential.access_token,
            credential.refresh_token,
            principal.nonce,
        )
        response_body = exchange.response.body
        diagnostic = exchange.diagnostic
        return all(value.encode("utf-8") not in response_body for value in raw_values) and all(
            value not in repr(diagnostic) and value not in str(diagnostic) for value in raw_values
        )
    except Exception:
        return _INVALID


def _pending_approval_non_enumeration_safe(
    known_exchange: object,
    unknown_exchange: object,
    principal: object,
) -> bool | object:
    try:
        known_shape = _error_shape_safe(known_exchange, 404)
        unknown_shape = _error_shape_safe(unknown_exchange, 404)
        if type(known_shape) is not ErrorShape or type(unknown_shape) is not ErrorShape:
            return _INVALID
        if _error_shapes_match_safe(known_shape, unknown_shape) is not True:
            return False
        if _principal_repr_is_redacted_safe(principal) is not True:
            return False
        if _exchange_hides_principal_canaries_safe(known_exchange, principal) is not True:
            return False
        return _exchange_hides_principal_canaries_safe(unknown_exchange, principal)
    except Exception:
        return _INVALID


def _decode_with_safe(
    exchange: object,
    expected_status: object,
    projector: Callable[[object, object], object],
    context: object,
) -> object:
    try:
        if type(exchange) is not HttpExchange or type(expected_status) is not int:
            return _INVALID
        if exchange.response.status_code != expected_status:
            return _INVALID
        body = exchange.response.body
        if len(body) > _BODY_LIMIT_BYTES or body.startswith(b"\xef\xbb\xbf"):
            return _INVALID
        document = json.loads(
            body.decode("utf-8", errors="strict"),
            object_pairs_hook=_unique_object_safe,
        )
        result = projector(document, context)
        del body, document
        return result
    except Exception:
        return _INVALID


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


def _raise_invalid() -> None:
    raise AuthzOracleError("authorization response is invalid") from None
