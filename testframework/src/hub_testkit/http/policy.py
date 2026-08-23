from __future__ import annotations

import math
from dataclasses import dataclass, field

import httpx

from .errors import HttpPolicyError


@dataclass(frozen=True, slots=True)
class TimeoutPolicy:
    connect: float = 2.0
    read: float = 5.0
    write: float = 5.0
    pool: float = 2.0

    def __post_init__(self) -> None:
        try:
            values = (self.connect, self.read, self.write, self.pool)
            for value in values:
                if type(value) is not float or not math.isfinite(value) or value <= 0.0:
                    raise HttpPolicyError("timeout policy is invalid")
            if values != (2.0, 5.0, 5.0, 2.0):
                raise HttpPolicyError("timeout policy is invalid")
        except Exception:
            raise HttpPolicyError("timeout policy is invalid") from None

    def to_httpx_timeout(self) -> httpx.Timeout:
        return httpx.Timeout(
            connect=self.connect,
            read=self.read,
            write=self.write,
            pool=self.pool,
        )


@dataclass(frozen=True, slots=True)
class RetryPolicy:
    max_attempts: int = 2

    def __post_init__(self) -> None:
        try:
            if type(self.max_attempts) is not int or self.max_attempts != 2:
                raise HttpPolicyError("retry policy is invalid")
        except Exception:
            raise HttpPolicyError("retry policy is invalid") from None


@dataclass(frozen=True, slots=True)
class HttpPolicy:
    timeout: TimeoutPolicy = field(default_factory=TimeoutPolicy)
    retry: RetryPolicy = field(default_factory=RetryPolicy)

    def __post_init__(self) -> None:
        try:
            if type(self.timeout) is not TimeoutPolicy or type(self.retry) is not RetryPolicy:
                raise HttpPolicyError("HTTP policy is invalid")
            object.__setattr__(
                self,
                "timeout",
                TimeoutPolicy(
                    connect=self.timeout.connect,
                    read=self.timeout.read,
                    write=self.timeout.write,
                    pool=self.timeout.pool,
                ),
            )
            object.__setattr__(
                self,
                "retry",
                RetryPolicy(max_attempts=self.retry.max_attempts),
            )
        except Exception:
            raise HttpPolicyError("HTTP policy is invalid") from None
