from __future__ import annotations


class HttpError(ValueError):
    pass


class HttpAuthError(HttpError):
    pass


class HttpPolicyError(HttpError):
    pass


class HttpRequestError(HttpError):
    pass


class HttpTransportError(HttpError):
    pass


class HttpDiagnosticError(HttpError):
    pass
