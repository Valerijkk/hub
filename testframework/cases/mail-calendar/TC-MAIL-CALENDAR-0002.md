+++
schema = "hub.testframework.case/v1"
id = "TC-MAIL-CALENDAR-0002"
title = "Published inbound invite projects bidirectional Mail and Calendar correlation and accepts RSVP"
status = "active"
task = "TASK-237"
requirement = "MAIL-CALENDAR-INBOUND-RFC5545-CORRELATION"
risk = "RISK-MAIL-CALENDAR-CORRELATION-DIVERGENCE"
scenario = "happy"
subject = "Published inbound Mail fixture correlated with a canonical Calendar event"
preconditions = ["The Docker-only python-integration profile provides isolated QA infrastructure, a QA backend URL and the Mail and Calendar control-plane capability token"]
oracle = "The published Mail message eventually exposes one canonical Calendar event, the event reciprocally exposes the Mail message, and the active viewer's accepted RSVP is returned at the response version"
cleanup = "The published fixture is not individually deleted because the QA endpoint accepts deletion only while created; the disposable python-integration QA profile teardown removes its temporary state, while the test closes only its HTTP client"

[automation]
kind = "pytest"
layer = "integration"
node_id = "testframework/tests/integration/mail-calendar/test_mail_calendar_sync.py::test_published_invite_projects_bidirectional_mail_calendar_correlation_and_rsvp"
+++
# Published Mail and Calendar correlation

The scenario waits with a bounded 30-second eventual-consistency poll for the
asynchronous inbound invite projection. It uses only public Mail and Calendar
HTTP boundaries: no database, broker or storage access is used. The fixture is
published once and is deliberately left to the disposable Docker
`python-integration` profile teardown, because a published fixture cannot be
deleted through the narrow QA control-plane endpoint.
