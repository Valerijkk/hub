from __future__ import annotations

from dataclasses import dataclass, field
from types import MappingProxyType

from .errors import HttpAuthError


@dataclass(frozen=True, slots=True)
class Actor:
    subject: str
    token: str = field(repr=False)

    def __post_init__(self) -> None:
        try:
            _require_actor_text(self.subject, 128)
            _require_actor_text(self.token, 4096)
        except Exception:
            raise HttpAuthError("actor is invalid") from None

    def headers(self) -> MappingProxyType[str, str]:
        return MappingProxyType({"Authorization": f"Bearer {self.token}"})


def _require_actor_text(value: object, maximum: int) -> None:
    if type(value) is not str or not value.strip() or len(value) > maximum:
        raise HttpAuthError("actor is invalid")
    if any(ord(character) < 32 or ord(character) == 127 for character in value):
        raise HttpAuthError("actor is invalid")
