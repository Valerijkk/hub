from __future__ import annotations

import json
from decimal import Decimal
from typing import NoReturn

from .source_bundle import ContractSourceBundle, ContractSourcePath


_MAX_JSON_DEPTH = 32
_MAX_JSON_ITEMS = 4_096
_MAX_JSON_STRING = 65_536
_INVALID = object()


class OpenApiContractError(ValueError):
    pass


class _JsonRejected(ValueError):
    pass


def assert_openapi_canonical_document(bundle: object) -> None:
    result = _canonical_document_valid_safe(bundle)
    del bundle
    if result is True:
        return
    del result
    _raise_invalid()


def assert_magic_verify_error_contract(bundle: object) -> None:
    result = _magic_verify_valid_safe(bundle)
    del bundle
    if result is True:
        return
    del result
    _raise_invalid()


def assert_tms_webhook_auth_contract(bundle: object) -> None:
    result = _tms_webhook_valid_safe(bundle)
    del bundle
    if result is True:
        return
    del result
    _raise_invalid()


def _canonical_document_valid_safe(bundle: object) -> bool:
    try:
        if type(bundle) is not ContractSourceBundle:
            return False
        main_source = bundle.select(ContractSourcePath.SERVER_MAIN)
        swagger_source = bundle.select(ContractSourcePath.SWAGGER_JSON)
        main_text = main_source.decode("utf-8", errors="strict")
        document = _parse_openapi_json_safe(swagger_source)
        if type(document) is not dict:
            return False
        source_lines = tuple(main_text.splitlines())
        annotation_block = _contiguous_annotation_block(source_lines, "func main()")
        if type(annotation_block) is not tuple or not annotation_block:
            return False
        base_paths = tuple(
            line for line in annotation_block if line.startswith("// @BasePath ")
        )
        if base_paths != ("// @BasePath /api",):
            return False
        if annotation_block.count("// @securityDefinitions.apikey BearerAuth") != 1:
            return False
        if annotation_block.count("// @securityDefinitions.apikey APIKeyAuth") != 1:
            return False
        if not _contains_contiguous_lines(
            annotation_block,
            (
                "// @securityDefinitions.apikey BearerAuth",
                "// @in header",
                "// @name Authorization",
            ),
        ):
            return False
        if not _contains_contiguous_lines(
            annotation_block,
            (
                "// @securityDefinitions.apikey APIKeyAuth",
                "// @in header",
                "// @name X-API-Key",
            ),
        ):
            return False
        security_definitions = document.get("securityDefinitions")
        paths = document.get("paths")
        if document.get("swagger") != "2.0" or document.get("basePath") != "/api":
            return False
        if type(security_definitions) is not dict or type(paths) is not dict or not paths:
            return False
        if security_definitions.get("BearerAuth") != {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
        }:
            return False
        if security_definitions.get("APIKeyAuth") != {
            "type": "apiKey",
            "name": "X-API-Key",
            "in": "header",
        }:
            return False
        return all(
            type(path) is str
            and path.startswith("/")
            and path != "/api"
            and not path.startswith("/api/")
            for path in paths
        )
    except Exception:
        return False


def _magic_verify_valid_safe(bundle: object) -> bool:
    try:
        if type(bundle) is not ContractSourceBundle:
            return False
        handler_source = bundle.select(ContractSourcePath.AUTH_HANDLER)
        swagger_source = bundle.select(ContractSourcePath.SWAGGER_JSON)
        handler_text = handler_source.decode("utf-8", errors="strict")
        document = _parse_openapi_json_safe(swagger_source)
        if type(document) is not dict:
            return False
        source_lines = tuple(handler_text.splitlines())
        lines = tuple(line.strip() for line in source_lines)
        if lines.count('auth.Post("/magic-link/verify", h.VerifyMagicLink)') != 1:
            return False
        annotation_block = _contiguous_annotation_block(
            source_lines,
            "func (h *Handler) VerifyMagicLink(",
        )
        if type(annotation_block) is not tuple or not annotation_block:
            return False
        if not annotation_block[0].startswith("// VerifyMagicLink "):
            return False
        if annotation_block[-1] != "// @Router /auth/magic-link/verify [post]":
            return False
        if sum(
            line.startswith("// @Param body body VerifyMagicLinkRequest true ")
            for line in annotation_block
        ) != 1:
            return False
        if sum(
            line.startswith("// @Success 200 {object} TokenPairResponse ")
            for line in annotation_block
        ) != 1:
            return False
        failures = tuple(
            line for line in annotation_block if line.startswith("// @Failure ")
        )
        if len(failures) != 4:
            return False
        for status in (400, 401, 403, 429):
            if sum(
                line.startswith(f"// @Failure {status} ") for line in failures
            ) != 1:
                return False
        operation = _operation(document, "/auth/magic-link/verify", "post")
        if type(operation) is not dict:
            return False
        parameters = operation.get("parameters")
        responses = operation.get("responses")
        if type(parameters) is not list or type(responses) is not dict:
            return False
        body_parameters = tuple(
            parameter
            for parameter in parameters
            if type(parameter) is dict and parameter.get("in") == "body"
        )
        if len(body_parameters) != 1:
            return False
        body_parameter = body_parameters[0]
        if body_parameter.get("name") != "body" or body_parameter.get("required") is not True:
            return False
        if body_parameter.get("schema") != {
            "$ref": "#/definitions/internal_auth.VerifyMagicLinkRequest"
        }:
            return False
        if set(responses) != {"200", "400", "401", "403", "429"}:
            return False
        success = responses.get("200")
        if type(success) is not dict or success.get("schema") != {
            "$ref": "#/definitions/internal_auth.TokenPairResponse"
        }:
            return False
        request_schema = _definition(document, "internal_auth.VerifyMagicLinkRequest")
        token_schema = _definition(document, "internal_auth.TokenPairResponse")
        if type(request_schema) is not dict or type(token_schema) is not dict:
            return False
        return not any(
            _contains_example(value)
            for value in (operation, request_schema, token_schema)
        )
    except Exception:
        return False


def _tms_webhook_valid_safe(bundle: object) -> bool:
    try:
        if type(bundle) is not ContractSourceBundle:
            return False
        server_main_source = bundle.select(ContractSourcePath.SERVER_MAIN)
        wiring_auth_source = bundle.select(ContractSourcePath.WIRING_AUTH)
        wiring_messenger_source = bundle.select(ContractSourcePath.WIRING_MESSENGER)
        apikey_source = bundle.select(ContractSourcePath.APIGATEWAY_APIKEY)
        auth_middleware_source = bundle.select(ContractSourcePath.AUTH_MIDDLEWARE)
        handler_source = bundle.select(ContractSourcePath.TMS_HANDLER)
        swagger_source = bundle.select(ContractSourcePath.SWAGGER_JSON)
        server_main_text = server_main_source.decode("utf-8", errors="strict")
        wiring_auth_text = wiring_auth_source.decode("utf-8", errors="strict")
        wiring_messenger_text = wiring_messenger_source.decode("utf-8", errors="strict")
        apikey_text = apikey_source.decode("utf-8", errors="strict")
        auth_middleware_text = auth_middleware_source.decode("utf-8", errors="strict")
        handler_text = handler_source.decode("utf-8", errors="strict")
        document = _parse_openapi_json_safe(swagger_source)
        if type(document) is not dict:
            return False
        server_main_lines = tuple(line.strip() for line in server_main_text.splitlines())
        server_chain = ("wireAuth(d)", "wireMessenger(d)")
        if any(server_main_lines.count(line) != 1 for line in server_chain):
            return False
        server_positions = tuple(server_main_lines.index(line) for line in server_chain)
        if server_positions != tuple(sorted(server_positions)):
            return False
        wiring_auth_lines = tuple(line.strip() for line in wiring_auth_text.splitlines())
        auth_chain = (
            "requireAuth := auth.RequireAuth(authSvc)",
            "d.requireAuth = requireAuth",
            "api.Use(apigatewaymod.APIKeyAuth(apiKeySvc))",
            "authHandler.RegisterRoutes(api, requireAuth)",
        )
        if any(wiring_auth_lines.count(line) != 1 for line in auth_chain):
            return False
        auth_positions = tuple(wiring_auth_lines.index(line) for line in auth_chain)
        if auth_positions != tuple(sorted(auth_positions)):
            return False
        wiring_messenger_lines = tuple(
            line.strip() for line in wiring_messenger_text.splitlines()
        )
        messenger_chain = (
            "requireAuth := d.requireAuth",
            "tmsHandler := tmsmod.NewHandler(tmsSvc)",
            "tmsHandler.RegisterRoutes(api, requireAuth, rbacSvc)",
        )
        if any(wiring_messenger_lines.count(line) != 1 for line in messenger_chain):
            return False
        messenger_positions = tuple(
            wiring_messenger_lines.index(line) for line in messenger_chain
        )
        if messenger_positions != tuple(sorted(messenger_positions)):
            return False
        apikey_lines = tuple(line.strip() for line in apikey_text.splitlines())
        apikey_chain = (
            "func APIKeyAuth(svc *APIKeyService) fiber.Handler {",
            'key := c.Get("X-API-Key")',
            "userID, err := svc.VerifyAPIKey(c.Context(), key)",
            'c.Locals("userID", userID.String())',
        )
        if any(apikey_lines.count(line) != 1 for line in apikey_chain):
            return False
        apikey_positions = tuple(apikey_lines.index(line) for line in apikey_chain)
        if apikey_positions != tuple(sorted(apikey_positions)):
            return False
        if not _contains_contiguous_lines(
            apikey_lines,
            (
                'c.Locals("userID", userID.String())',
                "return c.Next()",
            ),
        ):
            return False
        auth_middleware_lines = tuple(
            line.strip() for line in auth_middleware_text.splitlines()
        )
        if auth_middleware_lines.count(
            "func RequireAuth(svc *AuthService) fiber.Handler {"
        ) != 1:
            return False
        if not _contains_contiguous_lines(
            auth_middleware_lines,
            (
                'if uid, ok := c.Locals("userID").(string); ok && uid != "" {',
                "return c.Next()",
                "}",
            ),
        ):
            return False
        handler_lines = tuple(line.strip() for line in handler_text.splitlines())
        required_source = (
            "func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, "
            "rbacSvc *rbac.RBACService) {",
            'tms := api.Group("/projects/:projectId/tms", authMw)',
            'webhook := tms.Group("/webhook")',
            'webhook.Post("/", h.WebhookAllure, '
            "rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc))",
            'webhook.Post("/junit", h.WebhookJUnit, '
            "rbac.RequirePermission(rbac.PermTestRunCreate, rbacSvc))",
        )
        if any(handler_lines.count(line) != 1 for line in required_source):
            return False
        operations = (
            _operation(document, "/projects/{projectId}/tms/webhook", "post"),
            _operation(document, "/projects/{projectId}/tms/webhook/junit", "post"),
        )
        for operation in operations:
            if type(operation) is not dict:
                return False
            if operation.get("security") != [
                {"BearerAuth": []},
                {"APIKeyAuth": []},
            ]:
                return False
            parameters = operation.get("parameters")
            if type(parameters) is not list:
                return False
            for parameter in parameters:
                if type(parameter) is not dict:
                    return False
                name = parameter.get("name")
                location = parameter.get("in")
                if name == "X-Hub-Token" or (name == "token" and location == "query"):
                    return False
            if _contains_legacy_token_text(operation):
                return False
        return True
    except Exception:
        return False


def _parse_openapi_json_safe(raw: object) -> dict[str, object] | object:
    try:
        if type(raw) is not bytes or len(raw) > 4 * 1024 * 1024:
            return _INVALID
        text = raw.decode("utf-8", errors="strict")
        document = json.loads(
            text,
            object_pairs_hook=_object_pairs,
            parse_constant=_reject_constant,
            parse_int=_parse_bounded_int_lexeme,
            parse_float=_parse_bounded_finite_decimal_lexeme,
        )
        if type(document) is not dict or not _json_value_valid(document, 0):
            return _INVALID
        del raw, text
        return document
    except Exception:
        return _INVALID


def _object_pairs(pairs: list[tuple[str, object]]) -> dict[str, object]:
    result: dict[str, object] = {}
    for key, value in pairs:
        if type(key) is not str or key in result:
            raise _JsonRejected from None
        result[key] = value
    return result


def _reject_constant(raw: str) -> NoReturn:
    del raw
    raise _JsonRejected from None


def _parse_bounded_int_lexeme(raw: str) -> int:
    if type(raw) is not str or len(raw) > 20:
        del raw
        raise _JsonRejected from None
    try:
        value = int(raw)
    except Exception:
        del raw
        raise _JsonRejected from None
    del raw
    return value


def _parse_bounded_finite_decimal_lexeme(raw: str) -> Decimal:
    if type(raw) is not str or len(raw) > 64:
        del raw
        raise _JsonRejected from None
    try:
        value = Decimal(raw)
        if not value.is_finite():
            raise _JsonRejected
    except Exception:
        del raw
        raise _JsonRejected from None
    del raw
    return value


def _json_value_valid(value: object, depth: int) -> bool:
    if depth > _MAX_JSON_DEPTH:
        return False
    if value is None or type(value) in {bool, int, Decimal}:
        return True
    if type(value) is str:
        return len(value) <= _MAX_JSON_STRING
    if type(value) is list:
        return len(value) <= _MAX_JSON_ITEMS and all(
            _json_value_valid(item, depth + 1) for item in value
        )
    if type(value) is dict:
        return len(value) <= _MAX_JSON_ITEMS and all(
            type(key) is str
            and len(key) <= _MAX_JSON_STRING
            and _json_value_valid(item, depth + 1)
            for key, item in value.items()
        )
    return False


def _operation(document: dict[str, object], path: str, method: str) -> object:
    paths = document.get("paths")
    if type(paths) is not dict:
        return _INVALID
    path_item = paths.get(path)
    if type(path_item) is not dict:
        return _INVALID
    return path_item.get(method, _INVALID)


def _definition(document: dict[str, object], name: str) -> object:
    definitions = document.get("definitions")
    if type(definitions) is not dict:
        return _INVALID
    return definitions.get(name, _INVALID)


def _contiguous_annotation_block(source_lines: tuple[str, ...], signature: str) -> object:
    matches = tuple(
        index
        for index, line in enumerate(source_lines)
        if line.strip().startswith(signature)
    )
    if len(matches) != 1:
        return _INVALID
    start = matches[0]
    cursor = start - 1
    while cursor >= 0 and source_lines[cursor].strip().startswith("//"):
        cursor -= 1
    return tuple(line.strip() for line in source_lines[cursor + 1 : start])


def _contains_contiguous_lines(
    lines: tuple[str, ...],
    expected: tuple[str, ...],
) -> bool:
    width = len(expected)
    return sum(
        lines[index : index + width] == expected
        for index in range(len(lines) - width + 1)
    ) == 1


def _contains_example(value: object) -> bool:
    if type(value) is dict:
        if any(key in {"example", "examples"} for key in value):
            return True
        return any(_contains_example(item) for item in value.values())
    if type(value) is list:
        return any(_contains_example(item) for item in value)
    return False


def _contains_legacy_token_text(value: object) -> bool:
    if type(value) is str:
        folded = value.casefold()
        return "x-hub-token" in folded or ("query" in folded and "token" in folded)
    if type(value) is dict:
        return any(
            _contains_legacy_token_text(key) or _contains_legacy_token_text(item)
            for key, item in value.items()
        )
    if type(value) is list:
        return any(_contains_legacy_token_text(item) for item in value)
    return False


def _raise_invalid() -> NoReturn:
    raise OpenApiContractError("OpenAPI source contract is invalid") from None
