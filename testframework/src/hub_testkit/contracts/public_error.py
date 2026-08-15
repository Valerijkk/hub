from __future__ import annotations

from typing import NoReturn

from .openapi import _contains_example, _parse_openapi_json_safe
from .source_bundle import ContractSourceBundle, ContractSourcePath


_INVALID = object()


class PublicErrorContractError(ValueError):
    pass


def assert_public_error_artifact_contract(bundle: object) -> None:
    result = _public_error_artifact_valid_safe(bundle)
    del bundle
    if result is True:
        return
    del result
    _raise_invalid()


def _public_error_artifact_valid_safe(bundle: object) -> bool:
    try:
        if type(bundle) is not ContractSourceBundle:
            return False
        middleware_source = bundle.select(ContractSourcePath.PLATFORM_MIDDLEWARE)
        swagger_source = bundle.select(ContractSourcePath.SWAGGER_JSON)
        middleware_text = middleware_source.decode("utf-8", errors="strict")
        document = _parse_openapi_json_safe(swagger_source)
        if type(document) is not dict:
            return False
        if not _source_wire_relation_valid(middleware_text):
            return False
        definitions = document.get("definitions")
        if type(definitions) is not dict:
            return False
        public_names = tuple(
            name
            for name in definitions
            if type(name) is str
            and (name == "tzErrorBody" or name.endswith(".tzErrorBody"))
        )
        if len(public_names) != 1:
            return False
        public_name = public_names[0]
        public_definition = definitions.get(public_name, _INVALID)
        if public_definition != {
            "type": "object",
            "properties": {
                "error": {"type": "string"},
                "request_id": {"type": "string"},
            },
        }:
            return False
        if _contains_example(public_definition):
            return False
        if any(_is_app_error_name(name) for name in definitions):
            return False
        references = _reference_values(document)
        public_reference = f"#/definitions/{public_name}"
        if _public_error_response_reference_count(document, public_reference) < 1:
            return False
        if not _public_reference_contexts_valid(document, public_reference):
            return False
        return not any(_is_app_error_reference(reference) for reference in references)
    except Exception:
        return False


def _source_wire_relation_valid(source: str) -> bool:
    lines = tuple(line.strip() for line in source.splitlines())
    struct = (
        "type tzErrorBody struct {",
        'Error     string `json:"error"`',
        'RequestID string `json:"request_id"`',
        "}",
    )
    writers = (
        (
            "return c.Status(appErr.Code).JSON(tzErrorBody{",
            "Error:     appErr.Message,",
            "RequestID: requestID,",
            "})",
        ),
        (
            "return c.Status(fiberErr.Code).JSON(tzErrorBody{",
            "Error:     fiberErr.Message,",
            "RequestID: requestID,",
            "})",
        ),
        (
            "return c.Status(fiber.StatusInternalServerError).JSON(tzErrorBody{",
            'Error:     "internal server error",',
            "RequestID: requestID,",
            "})",
        ),
    )
    if _contiguous_count(lines, struct) != 1:
        return False
    if any(_contiguous_count(lines, writer) != 1 for writer in writers):
        return False
    return sum(".JSON(tzErrorBody{" in line for line in lines) == len(writers)


def _contiguous_count(lines: tuple[str, ...], expected: tuple[str, ...]) -> int:
    width = len(expected)
    return sum(
        lines[index : index + width] == expected
        for index in range(len(lines) - width + 1)
    )


def _reference_values(value: object) -> tuple[str, ...]:
    if type(value) is dict:
        raw_reference = value.get("$ref")
        direct = (raw_reference,) if type(raw_reference) is str else ()
        nested = tuple(
            reference
            for item in value.values()
            for reference in _reference_values(item)
        )
        return direct + nested
    if type(value) is list:
        return tuple(
            reference
            for item in value
            for reference in _reference_values(item)
        )
    return ()


def _public_reference_contexts_valid(value: object, public_reference: str) -> bool:
    if type(value) is dict:
        if value.get("$ref") == public_reference and value != {"$ref": public_reference}:
            return False
        if value.get("schema") == {"$ref": public_reference} and _contains_example(value):
            return False
        return all(
            _public_reference_contexts_valid(item, public_reference)
            for item in value.values()
        )
    if type(value) is list:
        return all(
            _public_reference_contexts_valid(item, public_reference) for item in value
        )
    return True


def _public_error_response_reference_count(
    document: dict[str, object],
    public_reference: str,
) -> int:
    paths = document.get("paths")
    if type(paths) is not dict:
        return -1
    count = 0
    for path_item in paths.values():
        if type(path_item) is not dict:
            return -1
        for operation in path_item.values():
            if type(operation) is not dict or "responses" not in operation:
                continue
            responses = operation.get("responses")
            if type(responses) is not dict:
                return -1
            for status, response in responses.items():
                if not _error_status(status):
                    continue
                if type(response) is dict and response.get("schema") == {
                    "$ref": public_reference
                }:
                    count += 1
    return count


def _error_status(value: object) -> bool:
    return (
        type(value) is str
        and len(value) == 3
        and value.isascii()
        and value.isdigit()
        and 400 <= int(value) <= 599
    )


def _is_app_error_name(value: object) -> bool:
    return type(value) is str and (value == "AppError" or value.endswith(".AppError"))


def _is_app_error_reference(value: str) -> bool:
    return value.startswith("#/definitions/") and _is_app_error_name(
        value.removeprefix("#/definitions/")
    )


def _raise_invalid() -> NoReturn:
    raise PublicErrorContractError("public error source contract is invalid") from None
