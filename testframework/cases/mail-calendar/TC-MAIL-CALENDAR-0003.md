+++
schema = "hub.testframework.case/v1"
id = "TC-MAIL-CALENDAR-0003"
title = "Mail-origin invite opens its Calendar event, accepts RSVP and returns to Mail"
status = "active"
task = "TASK-237"
requirement = "MAIL-CALENDAR-MAIL-ORIGIN-RSVP-JOURNEY"
risk = "RISK-MAIL-CALENDAR-UI-CORRELATION-BREAK"
scenario = "happy"
subject = "Mail and Calendar reciprocal invitation journey"
preconditions = ["The Docker-only python-e2e-smoke profile provides isolated QA infrastructure, fixed internal backend and frontend URLs, and the Mail and Calendar control-plane capability token"]
oracle = "A dev-session browser opened from the Mail message reaches its exact canonical event, persists accepted RSVP through the visible control, and follows the exact reciprocal Mail link back to the canonical thread"
cleanup = "The published fixture is not individually deleted because the QA endpoint accepts deletion only while created; the disposable python-e2e-smoke QA profile teardown removes temporary state, while the test closes only its HTTP client"

[automation]
kind = "pytest"
layer = "e2e"
node_id = "testframework/tests/e2e/smoke/mail-calendar/test_mail_calendar_rsvp_smoke.py::test_mail_origin_invite_opens_event_updates_rsvp_and_returns_to_mail"
+++
# Mail-origin Calendar RSVP journey

The test injects only the authenticated dev-session bootstrap before the first
application module evaluation. It follows the reciprocal visible links from a
published inbound Mail invite to the canonical Calendar event and back, while
asserting the visible accepted RSVP state. It uses no database, broker, direct
backend RSVP call, random test identity or explicit post-publication cleanup;
the disposable Docker profile owns published-fixture teardown.
