"""Closed scoped fault capability for Mail and Calendar QA recovery evidence."""

from __future__ import annotations

from collections.abc import Iterator
from contextlib import contextmanager
from enum import Enum
import sys
from typing import TYPE_CHECKING

from .models import FixtureSnapshot, _nonce

if TYPE_CHECKING:
    from .client import MailCalendarQAClient


class PublishedCleanupFaultKind(str, Enum):
    """Finite QA-only faults; values contain no database, broker, or entity IDs."""

    ORIGIN_P3_DUPLICATE = "origin-p3-duplicate-v1"
    ORIGIN_P4_MALFORMED_ENVELOPE = "origin-p4-malformed-envelope-v1"
    EXTERNAL_USER_AI_PROVIDER = "external-user-ai-provider-v1"
    CLEANUP_COMMIT_UNKNOWN_PRESENT = "cleanup-commit-unknown-present-v1"
    CLEANUP_COMMIT_UNKNOWN_ABSENT = "cleanup-commit-unknown-absent-v1"
    CLEANUP_SERIALIZATION_ROLLBACK = "cleanup-serialization-rollback-v1"
    CLEANUP_ENGINE_AFTER_DB_DELETE_FAILURE = "cleanup-engine-after-db-delete-failure-v1"
    CLEANUP_FINAL_CAS_LOST = "cleanup-final-cas-lost-v1"
    CLEANUP_MAPLESS_PARTIAL = "cleanup-mapless-partial-v1"
    CLEANUP_MAPLESS_AMBIGUOUS = "cleanup-mapless-ambiguous-v1"


def _published_cleanup_fault_kind(value: object) -> PublishedCleanupFaultKind:
    if type(value) is not PublishedCleanupFaultKind:
        raise ValueError
    return value


def _published_cleanup_fault_fixture(
    value: object,
    kind: PublishedCleanupFaultKind,
) -> FixtureSnapshot:
    _published_cleanup_fault_kind(kind)
    if type(value) is not FixtureSnapshot or value.state != "published":
        raise ValueError
    _nonce(value.nonce)
    return value


@contextmanager
def scoped_published_cleanup_fault(
    client: MailCalendarQAClient,
    fixture: FixtureSnapshot,
    kind: PublishedCleanupFaultKind,
) -> Iterator[None]:
    """Arm one fault and clear it unless the exercised operation consumed it."""

    from .client import MailCalendarQAClient

    if type(client) is not MailCalendarQAClient:
        raise ValueError
    fault_kind = _published_cleanup_fault_kind(kind)
    fault_fixture = _published_cleanup_fault_fixture(fixture, fault_kind)
    client.arm_published_cleanup_fault(fault_fixture, fault_kind)
    try:
        yield
    finally:
        if sys.exc_info()[0] is None:
            client.disarm_published_cleanup_fault(fault_fixture, fault_kind)
        else:
            try:
                client.disarm_published_cleanup_fault(fault_fixture, fault_kind)
            except BaseException:
                pass
