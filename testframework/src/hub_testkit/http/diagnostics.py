from __future__ import annotations

from collections.abc import Mapping
from types import MappingProxyType
from typing import Literal, cast

from hub_testkit.runtime.redaction import Redactor

from .errors import HttpDiagnosticError
from .transport import HttpDiagnostic


_DIAGNOSTIC_KEYS = (
    "method",
    "path",
    "status_code",
    "attempts",
    "retry_class",
    "correlation_id",
)


def diagnostic_fields(
    diagnostic: HttpDiagnostic,
    redactor: Redactor,
) -> Mapping[str, str | int | None]:
    try:
        if type(diagnostic) is not HttpDiagnostic or type(redactor) is not Redactor:
            raise HttpDiagnosticError("HTTP diagnostic is invalid")
        source = {
            "method": diagnostic.method,
            "path": diagnostic.path,
            "status_code": diagnostic.status_code,
            "attempts": diagnostic.attempts,
            "retry_class": diagnostic.retry_class,
            "correlation_id": diagnostic.correlation_id,
        }
        redacted = redactor.redact_mapping(source)
        if type(redacted) is not dict or tuple(redacted) != _DIAGNOSTIC_KEYS:
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
        normalized = HttpDiagnostic(
            method=method,
            path=path,
            status_code=status_code,
            attempts=attempts,
            retry_class=cast(Literal["none", "status", "transport"], retry_class),
            correlation_id=correlation_id,
        )
        return MappingProxyType(
            {
                "method": normalized.method,
                "path": normalized.path,
                "status_code": normalized.status_code,
                "attempts": normalized.attempts,
                "retry_class": normalized.retry_class,
                "correlation_id": normalized.correlation_id,
            }
        )
    except Exception:
        raise HttpDiagnosticError("HTTP diagnostic is invalid") from None
