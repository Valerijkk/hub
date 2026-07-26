#!/usr/bin/env python3
"""Shared redacted secret rules for validation and exact-package scanning."""

from __future__ import annotations

import re
from functools import lru_cache
from typing import Iterable, Tuple


SECRET_RULES: Tuple[Tuple[str, str], ...] = (
    (
        "private-key-header",
        r"-----BEGIN (?:RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----",
    ),
    ("aws-access-key", r"\b(?:AKIA|ASIA)[0-9A-Z]{16}\b"),
    (
        "github-token",
        r"\b(?:gh[pousr]_[A-Za-z0-9]{30,}|github_pat_[A-Za-z0-9_]{40,})\b",
    ),
    (
        "vendor-token",
        r"\b(?:sk-[A-Za-z0-9]{20,}|glpat-[A-Za-z0-9_-]{20,}|"
        r"xox[baprs]-[A-Za-z0-9-]{20,})\b",
    ),
    (
        "jwt-like",
        r"\beyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}"
        r"\.[A-Za-z0-9_-]{10,}\b",
    ),
    (
        "credential-assignment",
        r"(?i)\b(?:api[_-]?key|client[_-]?secret|password|access[_-]?token)"
        r"\s*[:=]\s*[\"']?[A-Za-z0-9+/=_-]{20,}",
    ),
)


@lru_cache(maxsize=1)
def compiled_secret_rules() -> Tuple[Tuple[str, re.Pattern[str]], ...]:
    return tuple((rule_id, re.compile(expression)) for rule_id, expression in SECRET_RULES)


def secret_rule_matches(text: str) -> Iterable[Tuple[str, int]]:
    """Yield rule ID and character offset without ever returning matched text."""

    for rule_id, pattern in compiled_secret_rules():
        for match in pattern.finditer(text):
            yield rule_id, match.start()
