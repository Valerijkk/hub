"""Typed runtime boundary for the QA Mail and Calendar fixture seam."""

from .client import MailCalendarQAClient
from .models import (
    AccountMaterializationFixtureSnapshot,
    CalendarCorrelation,
    CalendarViewerAttendee,
    DevSession,
    FixtureSnapshot,
    MailCalendarQAConfig,
    MailCalendarQAError,
    MailCorrelation,
    RSVPResult,
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
    "RSVPResult",
]
