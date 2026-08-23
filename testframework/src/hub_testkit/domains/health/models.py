from __future__ import annotations

from collections.abc import Mapping
from dataclasses import dataclass
from typing import Literal, Protocol


@dataclass(frozen=True, slots=True)
class HealthResponse:
    status_code: object
    body: object


@dataclass(frozen=True, slots=True)
class LiveHealth:
    status: Literal["ok"]


class HealthTransport(Protocol):
    def request(
        self,
        method: str,
        path: str,
        *,
        headers: Mapping[str, str],
    ) -> HealthResponse:
        ...
