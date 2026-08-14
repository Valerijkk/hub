from __future__ import annotations

import re
import tomllib
from collections.abc import Mapping
from dataclasses import dataclass
from pathlib import Path


class RedactionError(ValueError):
    pass


_POLICY_KEYS = frozenset(
    {"schema", "replacement", "sensitive_key_fragments", "value_patterns"}
)
_POLICY_SCHEMA = "hub.testframework.redaction/v1"


@dataclass(frozen=True, slots=True)
class Redactor:
    replacement: str
    _sensitive_key_fragments: tuple[str, ...]
    _value_patterns: tuple[re.Pattern[str], ...]

    @classmethod
    def from_policy(cls, path: str | Path) -> Redactor:
        try:
            with Path(path).open("rb") as stream:
                document = tomllib.load(stream)
        except (OSError, UnicodeDecodeError, tomllib.TOMLDecodeError):
            raise RedactionError("redaction policy could not be loaded") from None

        try:
            if set(document) != _POLICY_KEYS:
                raise RedactionError("redaction policy has unsupported keys")
            if _require_string(document["schema"]) != _POLICY_SCHEMA:
                raise RedactionError("redaction policy schema is invalid")
            replacement = _require_string(document["replacement"])
            fragments = _require_string_tuple(document["sensitive_key_fragments"])
            patterns = _require_string_tuple(document["value_patterns"])
            if not fragments or not patterns:
                raise RedactionError("redaction policy is incomplete")
            if any(fragment != fragment.casefold() for fragment in fragments):
                raise RedactionError("redaction policy key fragments are invalid")
            return cls(
                replacement=replacement,
                _sensitive_key_fragments=fragments,
                _value_patterns=tuple(re.compile(pattern) for pattern in patterns),
            )
        except (KeyError, re.error):
            raise RedactionError("redaction policy is invalid") from None

    def redact_text(self, value: str) -> str:
        if type(value) is not str:
            raise RedactionError("redaction text is invalid")
        redacted = value
        for pattern in self._value_patterns:
            redacted = pattern.sub(self.replacement, redacted)
        return redacted

    def redact_mapping(self, value: Mapping[str, object]) -> dict[str, object]:
        if not isinstance(value, Mapping):
            raise RedactionError("redaction mapping is invalid")
        redacted = self._redact(value, set())
        if not isinstance(redacted, dict):
            raise RedactionError("redaction mapping is invalid")
        return redacted

    def _redact(self, value: object, active_containers: set[int]) -> object:
        if type(value) is str:
            return self.redact_text(value)
        if type(value) in {bool, int, float} or value is None:
            return value
        if isinstance(value, Mapping):
            return self._redact_mapping(value, active_containers)
        if type(value) is list:
            return self._redact_sequence(value, active_containers, list)
        if type(value) is tuple:
            return self._redact_sequence(value, active_containers, tuple)
        raise RedactionError("unsupported redaction value")

    def _redact_mapping(
        self,
        value: Mapping[object, object],
        active_containers: set[int],
    ) -> dict[str, object]:
        identity = id(value)
        if identity in active_containers:
            raise RedactionError("cyclic redaction value")
        active_containers.add(identity)
        try:
            redacted: dict[str, object] = {}
            for key, nested in value.items():
                if type(key) is not str:
                    raise RedactionError("redaction mapping keys must be strings")
                if self._is_sensitive_key(key):
                    redacted[key] = self.replacement
                else:
                    redacted[key] = self._redact(nested, active_containers)
            return redacted
        finally:
            active_containers.remove(identity)

    def _redact_sequence(
        self,
        value: list[object] | tuple[object, ...],
        active_containers: set[int],
        constructor: type[list[object]] | type[tuple[object, ...]],
    ) -> list[object] | tuple[object, ...]:
        identity = id(value)
        if identity in active_containers:
            raise RedactionError("cyclic redaction value")
        active_containers.add(identity)
        try:
            items = [self._redact(item, active_containers) for item in value]
            return constructor(items)
        finally:
            active_containers.remove(identity)

    def _is_sensitive_key(self, key: str) -> bool:
        normalized_key = key.casefold()
        return any(fragment in normalized_key for fragment in self._sensitive_key_fragments)


def _require_string(value: object) -> str:
    if type(value) is not str or not value:
        raise RedactionError("redaction policy values must be non-empty strings")
    return value


def _require_string_tuple(value: object) -> tuple[str, ...]:
    if type(value) is not list:
        raise RedactionError("redaction policy values must be lists")
    items = tuple(_require_string(item) for item in value)
    if len(set(items)) != len(items):
        raise RedactionError("redaction policy values must be unique")
    return items
