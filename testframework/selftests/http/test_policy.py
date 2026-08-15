from __future__ import annotations

import ast
import dataclasses
import traceback
from collections.abc import Mapping
from pathlib import Path

import httpx
import pytest

from hub_testkit.domains.health.api import HealthTransportError, get_live_health_via_http
from hub_testkit.domains.health.models import LiveHealth
from hub_testkit.domains.health.oracles import HealthOracleError
from hub_testkit.http.auth import Actor
from hub_testkit.http.errors import (
    HttpAuthError,
    HttpPolicyError,
    HttpRequestError,
    HttpTransportError,
)
from hub_testkit.http.diagnostics import diagnostic_fields
from hub_testkit.http.policy import HttpPolicy, RetryPolicy, TimeoutPolicy
from hub_testkit.http.transport import (
    HttpDiagnostic,
    HttpExchange,
    HttpRequest,
    HttpResponse,
    HttpxTransport,
)
from hub_testkit.runtime.redaction import Redactor


REDACTION_POLICY_PATH = Path(__file__).resolve().parents[2] / "policies" / "redaction.toml"


class _RecordingClient:
    def __init__(self, outcomes: list[object]) -> None:
        self.calls: list[tuple[str, str, dict[str, str], bytes | None, httpx.Timeout]] = []
        self.outcomes = outcomes

    def request(
        self,
        method: str,
        path: str,
        *,
        headers: Mapping[str, str],
        content: bytes | None,
        timeout: httpx.Timeout,
    ) -> httpx.Response:
        self.calls.append((method, path, dict(headers), content, timeout))
        outcome = self.outcomes.pop(0)
        if isinstance(outcome, BaseException):
            raise outcome
        if not isinstance(outcome, httpx.Response):
            raise AssertionError("fake outcome is invalid")
        return outcome


def _response(status_code: int, body: bytes = b"{}") -> httpx.Response:
    return httpx.Response(status_code=status_code, content=body)


def _redactor() -> Redactor:
    return Redactor.from_policy(REDACTION_POLICY_PATH)


def _transport(client: _RecordingClient, policy: HttpPolicy) -> HttpxTransport:
    return HttpxTransport(client=client, policy=policy, redactor=_redactor())


def _request(
    method: str = "GET",
    *,
    actor: Actor | None = None,
    idempotency_key: str | None = None,
    body: bytes | None = None,
) -> HttpRequest:
    return HttpRequest(
        method=method,
        path="/api/resource",
        actor=actor,
        correlation_id="trace-001",
        idempotency_key=idempotency_key,
        body=body,
    )


def _assert_generic_without_canary(error: BaseException, canary: str, message: str) -> None:
    assert str(error) == message
    assert error.__cause__ is None
    assert error.__suppress_context__ is True
    assert canary not in "".join(traceback.format_exception(error))


def test_actor_policy_and_request_models_are_frozen_and_strict() -> None:
    actor = Actor(subject="qa-user", token="actor-token")
    assert dict(actor.headers()) == {"Authorization": "Bearer actor-token"}
    request_with_secrets = _request(
        "POST",
        actor=actor,
        idempotency_key="idempotency-repr-canary",
        body=b"body-repr-canary",
    )
    assert "actor-token" not in repr(actor)
    assert "actor-token" not in repr(request_with_secrets)
    assert "idempotency-repr-canary" not in repr(request_with_secrets)
    assert "body-repr-canary" not in repr(request_with_secrets)
    with pytest.raises(dataclasses.FrozenInstanceError):
        actor.token = "changed"
    with pytest.raises(HttpAuthError) as auth_error:
        Actor(subject="qa\nuser", token="TOKEN-CANARY")
    _assert_generic_without_canary(auth_error.value, "TOKEN-CANARY", "actor is invalid")
    bypass_actor = object.__new__(Actor)
    object.__setattr__(bypass_actor, "subject", "qa-user")
    object.__setattr__(bypass_actor, "token", "actor\x00BYPASS-ACTOR-CANARY")
    with pytest.raises(HttpRequestError) as actor_bypass_error:
        _request(actor=bypass_actor)
    _assert_generic_without_canary(
        actor_bypass_error.value,
        "BYPASS-ACTOR-CANARY",
        "HTTP request is invalid",
    )

    timeout_policy = TimeoutPolicy()
    timeout = timeout_policy.to_httpx_timeout()
    assert (timeout.connect, timeout.read, timeout.write, timeout.pool) == (2.0, 5.0, 5.0, 2.0)
    assert RetryPolicy().max_attempts == 2
    with pytest.raises(HttpPolicyError):
        TimeoutPolicy(connect=2, read=5.0, write=5.0, pool=2.0)
    with pytest.raises(HttpPolicyError):
        TimeoutPolicy(connect=float("nan"), read=5.0, write=5.0, pool=2.0)
    with pytest.raises(HttpPolicyError):
        TimeoutPolicy(connect=3.0, read=5.0, write=5.0, pool=2.0)
    with pytest.raises(HttpPolicyError):
        RetryPolicy(max_attempts=3)
    with pytest.raises(HttpPolicyError):
        RetryPolicy(max_attempts=True)
    with pytest.raises(dataclasses.FrozenInstanceError):
        HttpPolicy().retry.max_attempts = 1
    bypass_timeout = object.__new__(TimeoutPolicy)
    object.__setattr__(bypass_timeout, "connect", "POLICY-BYPASS-CANARY")
    object.__setattr__(bypass_timeout, "read", 5.0)
    object.__setattr__(bypass_timeout, "write", 5.0)
    object.__setattr__(bypass_timeout, "pool", 2.0)
    with pytest.raises(HttpPolicyError) as policy_bypass_error:
        HttpPolicy(timeout=bypass_timeout, retry=RetryPolicy())
    _assert_generic_without_canary(
        policy_bypass_error.value,
        "POLICY-BYPASS-CANARY",
        "HTTP policy is invalid",
    )
    bypass_policy = object.__new__(HttpPolicy)
    object.__setattr__(bypass_policy, "timeout", bypass_timeout)
    object.__setattr__(bypass_policy, "retry", RetryPolicy())
    no_io_client = _RecordingClient([_response(200)])
    with pytest.raises(HttpTransportError) as transport_policy_error:
        HttpxTransport(no_io_client, bypass_policy, _redactor())
    _assert_generic_without_canary(
        transport_policy_error.value,
        "POLICY-BYPASS-CANARY",
        "HTTP transport is invalid",
    )
    assert no_io_client.calls == []
    bypass_transport = object.__new__(HttpxTransport)
    bypass_transport_client = _RecordingClient([_response(200)])
    object.__setattr__(bypass_transport, "client", bypass_transport_client)
    object.__setattr__(bypass_transport, "policy", bypass_policy)
    object.__setattr__(bypass_transport, "redactor", _redactor())
    with pytest.raises(HttpTransportError) as bypass_transport_error:
        bypass_transport.send(_request())
    _assert_generic_without_canary(
        bypass_transport_error.value,
        "POLICY-BYPASS-CANARY",
        "HTTP transport failed",
    )
    assert bypass_transport_client.calls == []
    with pytest.raises(TypeError):
        HttpxTransport(no_io_client, HttpPolicy())
    with pytest.raises(HttpTransportError) as redactor_error:
        HttpxTransport(no_io_client, HttpPolicy(), object())
    _assert_generic_without_canary(
        redactor_error.value,
        "REDACTOR-CANARY",
        "HTTP transport is invalid",
    )

    request = _request(actor=actor, body=b"payload")
    assert tuple(field.name for field in dataclasses.fields(HttpRequest)) == (
        "method",
        "path",
        "actor",
        "correlation_id",
        "idempotency_key",
        "body",
    )
    assert request.body == b"payload"
    for invalid in (
        {"method": "get"},
        {"path": "https://outside"},
        {"path": "//outside"},
        {"path": "/api/live?query"},
        {"correlation_id": "trace\ncanary"},
        {"body": bytearray(b"body")},
    ):
        with pytest.raises(HttpRequestError):
            HttpRequest(
                method=invalid.get("method", "GET"),
                path=invalid.get("path", "/api/resource"),
                actor=None,
                correlation_id=invalid.get("correlation_id", "trace-001"),
                idempotency_key=None,
                body=invalid.get("body", None),
            )
    with pytest.raises(HttpRequestError):
        _request(method="GET", idempotency_key="idempotency-001")
    for method in ("GET", "HEAD", "OPTIONS", "PUT", "DELETE"):
        with pytest.raises(HttpRequestError):
            _request(method=method, idempotency_key="idempotency-001")
    with pytest.raises(HttpRequestError):
        _request(method="POST", idempotency_key="key\nCANARY")
    with pytest.raises(TypeError):
        HttpRequest(
            method="GET",
            path="/api/resource",
            actor=None,
            correlation_id="trace-001",
            idempotency_key=None,
            body=None,
            headers={},
        )
    for response in (
        {"status_code": 99, "body": b""},
        {"status_code": True, "body": b""},
        {"status_code": 200, "body": bytearray()},
    ):
        with pytest.raises(HttpRequestError):
            HttpResponse(**response)
    bypass_diagnostic = object.__new__(HttpDiagnostic)
    object.__setattr__(bypass_diagnostic, "method", "GET")
    object.__setattr__(bypass_diagnostic, "path", "/api/resource")
    object.__setattr__(bypass_diagnostic, "status_code", 200)
    object.__setattr__(bypass_diagnostic, "attempts", 1)
    object.__setattr__(bypass_diagnostic, "retry_class", "none")
    object.__setattr__(bypass_diagnostic, "correlation_id", "trace\nBYPASS-EXCHANGE-CANARY")
    with pytest.raises(HttpTransportError) as exchange_bypass_error:
        HttpExchange(HttpResponse(status_code=200, body=b"{}"), bypass_diagnostic)
    _assert_generic_without_canary(
        exchange_bypass_error.value,
        "BYPASS-EXCHANGE-CANARY",
        "HTTP exchange is invalid",
    )
    bypass_response = object.__new__(HttpResponse)
    object.__setattr__(bypass_response, "status_code", True)
    object.__setattr__(bypass_response, "body", b"BYPASS-RESPONSE-CANARY")
    with pytest.raises(HttpTransportError) as response_bypass_error:
        HttpExchange(
            bypass_response,
            HttpDiagnostic(
                method="GET",
                path="/api/resource",
                status_code=200,
                attempts=1,
                retry_class="none",
                correlation_id="trace-001",
            ),
        )
    _assert_generic_without_canary(
        response_bypass_error.value,
        "BYPASS-RESPONSE-CANARY",
        "HTTP exchange is invalid",
    )


def test_transport_emits_exact_headers_and_retry_matrix() -> None:
    actor = Actor(subject="qa-user", token="actor-token")
    client = _RecordingClient([_response(200, b"first")])
    exchange = _transport(client, HttpPolicy()).send(
        _request(method="POST", actor=actor, idempotency_key="post-001", body=b"payload")
    )
    assert exchange.response == HttpResponse(status_code=200, body=b"first")
    assert exchange.diagnostic.attempts == 1
    assert exchange.diagnostic.retry_class == "none"
    assert client.calls[0][0:4] == (
        "POST",
        "/api/resource",
        {
            "X-Request-ID": "trace-001",
            "Authorization": "Bearer actor-token",
            "Idempotency-Key": "post-001",
        },
        b"payload",
    )
    assert (client.calls[0][4].connect, client.calls[0][4].read) == (2.0, 5.0)

    for status_code in (502, 503, 504):
        retry_client = _RecordingClient([_response(status_code), _response(200, b"final")])
        retried = _transport(retry_client, HttpPolicy()).send(_request("GET"))
        assert retried.response == HttpResponse(status_code=200, body=b"final")
        assert (retried.diagnostic.attempts, retried.diagnostic.retry_class) == (2, "status")
        assert len(retry_client.calls) == 2

    final_retry_client = _RecordingClient([_response(502), _response(504, b"final")])
    final_retry = _transport(final_retry_client, HttpPolicy()).send(_request("GET"))
    assert final_retry.response == HttpResponse(status_code=504, body=b"final")
    assert (final_retry.diagnostic.attempts, final_retry.diagnostic.retry_class) == (2, "status")

    for method in ("GET", "HEAD", "OPTIONS", "PUT", "DELETE"):
        retry_client = _RecordingClient([_response(503), _response(200)])
        retried = _transport(retry_client, HttpPolicy()).send(_request(method))
        assert (retried.diagnostic.attempts, retried.diagnostic.retry_class) == (2, "status")

    for exception_type in (
        httpx.ConnectTimeout,
        httpx.ReadTimeout,
        httpx.WriteTimeout,
        httpx.PoolTimeout,
        httpx.ConnectError,
        httpx.ReadError,
        httpx.WriteError,
        httpx.RemoteProtocolError,
    ):
        retry_client = _RecordingClient([exception_type("retryable"), _response(200)])
        retried = _transport(retry_client, HttpPolicy()).send(_request("GET"))
        assert (retried.diagnostic.attempts, retried.diagnostic.retry_class) == (2, "transport")

    for status_code in (429, 400, 401, 404, 418, 500, 501, 599):
        final_client = _RecordingClient([_response(status_code)])
        final = _transport(final_client, HttpPolicy()).send(_request("GET"))
        assert (final.diagnostic.attempts, final.diagnostic.retry_class) == (1, "none")
        assert len(final_client.calls) == 1

    exception_client = _RecordingClient([httpx.ProtocolError("not retryable")])
    with pytest.raises(HttpTransportError):
        _transport(exception_client, HttpPolicy()).send(_request("GET"))
    assert len(exception_client.calls) == 1

    for method in ("POST", "PATCH"):
        final_client = _RecordingClient([_response(503)])
        final = _transport(final_client, HttpPolicy()).send(_request(method))
        assert (final.diagnostic.attempts, final.diagnostic.retry_class) == (1, "none")
    for method in ("POST", "PATCH"):
        keyed_client = _RecordingClient([_response(503), _response(200)])
        keyed = _transport(keyed_client, HttpPolicy()).send(
            _request(method, idempotency_key="post-001")
        )
        assert (keyed.diagnostic.attempts, keyed.diagnostic.retry_class) == (2, "status")


def test_transport_sanitizes_every_public_diagnostic_before_exchange() -> None:
    canary = "TRANSPORT-PATH-TOKEN-CANARY"
    request = HttpRequest(
        method="GET",
        path=f"/api/Bearer {canary}",
        actor=None,
        correlation_id="trace-001",
        idempotency_key=None,
        body=None,
    )
    redactor = _redactor()

    exchange = HttpxTransport(
        client=_RecordingClient([_response(200)]),
        policy=HttpPolicy(),
        redactor=redactor,
    ).send(request)
    public_fields = diagnostic_fields(exchange.diagnostic, redactor)

    assert exchange.diagnostic.path == "/api/[REDACTED]"
    assert public_fields["path"] == "/api/[REDACTED]"
    assert canary not in str(exchange.diagnostic)
    assert canary not in repr(exchange.diagnostic)
    assert canary not in repr(public_fields)


def test_transport_hides_body_token_idempotency_and_exception_canaries() -> None:
    canary = "HTTP-TRANSPORT-CANARY"
    bad_client = _RecordingClient([RuntimeError(canary)])
    with pytest.raises(HttpTransportError) as transport_error:
        _transport(bad_client, HttpPolicy()).send(
            _request(
                "POST",
                actor=Actor(subject="qa-user", token="token-canary"),
                idempotency_key="idempotency-canary",
                body=b"body-canary",
            )
        )
    _assert_generic_without_canary(transport_error.value, canary, "HTTP transport failed")
    formatted = "".join(traceback.format_exception(transport_error.value))
    for secret in ("token-canary", "body-canary", "idempotency-canary"):
        assert secret not in formatted

    second_canary = "SECOND-TRANSPORT-CANARY"
    retry_client = _RecordingClient([httpx.ConnectTimeout("retry"), RuntimeError(second_canary)])
    with pytest.raises(HttpTransportError) as retry_error:
        _transport(retry_client, HttpPolicy()).send(_request())
    _assert_generic_without_canary(retry_error.value, second_canary, "HTTP transport failed")

    for factory, value in (
        (lambda: Actor(subject="qa-user", token="token\x00canary"), "token"),
        (lambda: _request("POST", idempotency_key="key\x00canary"), "key"),
    ):
        with pytest.raises((HttpAuthError, HttpRequestError)) as raised:
            factory()
        assert value not in str(raised.value)

    bypass_canary = "BYPASS-REQUEST-CANARY"
    bypass = object.__new__(HttpRequest)
    object.__setattr__(bypass, "method", "GET")
    object.__setattr__(bypass, "path", "/api/resource")
    object.__setattr__(bypass, "actor", None)
    object.__setattr__(bypass, "correlation_id", f"trace\n{bypass_canary}")
    object.__setattr__(bypass, "idempotency_key", None)
    object.__setattr__(bypass, "body", None)
    with pytest.raises(HttpTransportError) as bypass_error:
        _transport(_RecordingClient([_response(200)]), HttpPolicy()).send(bypass)
    _assert_generic_without_canary(bypass_error.value, bypass_canary, "HTTP transport failed")


def test_health_http_adapter_uses_anonymous_fixed_json_probe() -> None:
    client = _RecordingClient([_response(200, b'{"status":"ok"}')])
    redactor = _redactor()

    result = get_live_health_via_http(
        client,
        correlation_id="health-001",
        redactor=redactor,
    )

    assert result == LiveHealth(status="ok")
    assert client.calls[0][0:4] == ("GET", "/api/live", {"X-Request-ID": "health-001"}, None)
    for body in (b"\xef\xbb\xbf{}", b"[]", b"{"):
        with pytest.raises(HealthTransportError):
            get_live_health_via_http(
                _RecordingClient([_response(200, body)]),
                correlation_id="health-001",
                redactor=redactor,
            )
    with pytest.raises(HealthOracleError):
        get_live_health_via_http(
            _RecordingClient([_response(200, b'{"status": "ok", "version": "1"}')]),
            correlation_id="health-001",
            redactor=redactor,
        )
    transport_canary = "HEALTH-TRANSPORT-CANARY"
    with pytest.raises(HealthTransportError) as transport_error:
        get_live_health_via_http(
            _RecordingClient([RuntimeError(transport_canary)]),
            correlation_id="health-001",
            redactor=redactor,
        )
    _assert_generic_without_canary(
        transport_error.value,
        transport_canary,
        "live health response is invalid",
    )
    with pytest.raises(TypeError):
        get_live_health_via_http(client, correlation_id="health-001")


class _HttpSourceGuard(ast.NodeVisitor):
    _FORBIDDEN_ROOTS = frozenset({"os", "pytest", "requests", "socket", "urllib"})
    _SUITE_PARTS = frozenset({"tests", "selftests", "cases"})
    _HTTPX_NETWORK_CALLS = frozenset(
        {
            "httpx.Client",
            "httpx.AsyncClient",
            "httpx.delete",
            "httpx.get",
            "httpx.head",
            "httpx.options",
            "httpx.patch",
            "httpx.post",
            "httpx.put",
            "httpx.request",
            "httpx.stream",
        }
    )

    def __init__(self) -> None:
        self.aliases: dict[str, str] = {}
        self.violations: set[str] = set()
        self._scope_depth = 0

    def visit_Import(self, node: ast.Import) -> None:
        for alias in node.names:
            self.aliases[alias.asname or alias.name.split(".", 1)[0]] = alias.name
            self._check_module(alias.name)

    def visit_ImportFrom(self, node: ast.ImportFrom) -> None:
        module = node.module or ""
        self._check_module(module)
        for alias in node.names:
            target = f"{module}.{alias.name}" if module else alias.name
            self.aliases[alias.asname or alias.name] = target
            self._check_module(target)

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        self._scope_depth += 1
        self.generic_visit(node)
        self._scope_depth -= 1

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef) -> None:
        self._scope_depth += 1
        self.generic_visit(node)
        self._scope_depth -= 1

    def visit_Assign(self, node: ast.Assign) -> None:
        if self._scope_depth == 0 and self._is_mutable(node.value):
            self.violations.add("global-mutable")
        self.generic_visit(node)

    def visit_AnnAssign(self, node: ast.AnnAssign) -> None:
        if self._scope_depth == 0 and node.value is not None and self._is_mutable(node.value):
            self.violations.add("global-mutable")
        self.generic_visit(node)

    def visit_Call(self, node: ast.Call) -> None:
        target = self._resolve(node.func)
        if target in self._HTTPX_NETWORK_CALLS:
            self.violations.add("default-client")
        if target == "time.sleep":
            self.violations.add("sleep")
        if target in {"__import__", "builtins.__import__", "importlib.import_module"}:
            self.violations.add("dynamic-import")
        if target in {"os.getenv", "os.environ"} or target.startswith("os.environ."):
            self.violations.add("environment")
        if target.endswith(".Redactor.from_policy"):
            self.violations.add("policy-load")
        self.generic_visit(node)

    def visit_Attribute(self, node: ast.Attribute) -> None:
        target = self._resolve(node)
        if target == "os.environ" or target.startswith("os.environ."):
            self.violations.add("environment")
        self.generic_visit(node)

    def visit_Subscript(self, node: ast.Subscript) -> None:
        if self._resolve(node.value) == "os.environ":
            self.violations.add("environment")
        self.generic_visit(node)

    def _check_module(self, module: str) -> None:
        parts = tuple(part for part in module.split(".") if part)
        if parts and parts[0] in self._FORBIDDEN_ROOTS:
            self.violations.add("forbidden-import")
        if any(part in self._SUITE_PARTS for part in parts):
            self.violations.add("suite-import")

    def _is_mutable(self, value: ast.expr) -> bool:
        if isinstance(value, (ast.Dict, ast.List, ast.Set)):
            return True
        if not isinstance(value, ast.Call):
            return False
        return self._resolve(value.func) in {
            "builtins.dict",
            "builtins.list",
            "builtins.set",
            "dict",
            "list",
            "set",
        }

    def _resolve(self, node: ast.expr) -> str:
        if isinstance(node, ast.Name):
            return self.aliases.get(node.id, node.id)
        if isinstance(node, ast.Attribute):
            return f"{self._resolve(node.value)}.{node.attr}"
        return ""


def _guard_violations(source: str) -> set[str]:
    guard = _HttpSourceGuard()
    guard.visit(ast.parse(source))
    return guard.violations


def test_http_source_has_no_ambient_client_or_suite_escape() -> None:
    adversarial_sources = {
        "import httpx as h\nh.Client()": "default-client",
        "from httpx import Client as C\nC()": "default-client",
        "import time as clock\nclock.sleep(1)": "sleep",
        "from time import sleep as pause\npause(1)": "sleep",
        "import importlib as loader\nloader.import_module(name)": "dynamic-import",
        "__import__('os')": "dynamic-import",
        "from builtins import __import__ as load\nload(name)": "dynamic-import",
        "import os as system\nsystem.environ['HTTP_TOKEN']": "environment",
        "from os import getenv as read_env\nread_env('HTTP_TOKEN')": "environment",
        "from builtins import list as MutableList\nCACHE = MutableList()": "global-mutable",
        "from testframework.tests.api import health": "suite-import",
    }
    for source, expected in adversarial_sources.items():
        assert expected in _guard_violations(source)
    assert _guard_violations(
        "from collections.abc import Mapping\n"
        "class Client:\n"
        "    def timeout(self) -> str:\n"
        "        return 'sleep'\n"
    ) == set()

    source_root = Path(__file__).resolve().parents[2] / "src" / "hub_testkit"
    source_paths = [
        *sorted((source_root / "http").rglob("*.py")),
        source_root / "domains" / "health" / "api.py",
    ]
    for path in source_paths:
        assert _guard_violations(path.read_text(encoding="utf-8")) == set()
