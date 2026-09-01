"""Typed runtime boundary for the QA Mail and Calendar fixture seam."""

from .client import MailCalendarQAClient, ThreadSearchDeleteObservation
from .mail_calendar_faults import PublishedCleanupFaultKind, scoped_published_cleanup_fault
from .models import (
    AccountMaterializationFixtureSnapshot,
    CalendarCorrelation,
    CalendarViewerAttendee,
    DevSession,
    FixtureSnapshot,
    MailCalendarQAConfig,
    MailCalendarQAError,
    MailCorrelation,
    MailThreadProjectionItem,
    MailThreadProjectionPage,
    RSVPResult,
    ThreadSearchFixtureSnapshot,
)

__all__ = [
    "AccountMaterializationFixtureSnapshot",
    "CalendarCorrelation",
    "CalendarViewerAttendee",
    "DevSession",
    "FixtureSnapshot",
    "MailCalendarQAClient",
    "MailCalendarQAConfig",
    "MailCalendarQAError",
    "MailCorrelation",
    "MailThreadProjectionItem",
    "MailThreadProjectionPage",
    "PublishedCleanupFaultKind",
    "RSVPResult",
    "scoped_published_cleanup_fault",
    "ThreadSearchDeleteObservation",
    "ThreadSearchFixtureSnapshot",
]
