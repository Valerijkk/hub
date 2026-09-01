+++
schema = "hub.testframework.case/v1"
id = "TC-MAIL-FTS-0004"
title = "Private thread-search fixture delete guard stops declared Ent user cascade before database enforcement"
status = "active"
task = "TASK-237"
requirement = "MAIL-THREAD-FTS-FIXTURE-SAFE-DELETE-GUARD"
risk = "RISK-MAIL-THREAD-FTS-FUTURE-USER-CASCADE-BYPASSES-GUARD"
scenario = "failure-recovery"
subject = "Guard ordering for CalendarMutationReceipt actor cascade"
preconditions = ["Docker python-integration has the private QA fixture seam and a valid QA token."]
oracle = "After arming `ent-user-cascade-v1`, DELETE must be classified as literal `PREDELETE`: generic 409 body and `X-Hub-QA-Fixture-Delete-Guard: predelete` before any owner deletion."
cleanup = "Disarm the exact blocker, then delete the fixture normally. A stored `DELETED` observation alone permits already-absent target cleanup; cleanup errors never mask the primary failure."

[automation]
kind = "pytest"
layer = "integration"
node_id = "testframework/tests/integration/mail/test_mail_thread_fixture_delete_guard.py::test_mail_thread_fixture_delete_guard_stops_future_user_cascade_before_delete"
+++

Setup: create one `thread-search-v1` fixture and arm the closed single-fixture blocker endpoint `POST /api/_qa/mail-calendar/fixtures/:nonce/thread-search-delete-blocker/ent-user-cascade-v1`.

Action: observe the guarded fixture DELETE through the typed QA client.

Current semantic RED: current Atlas `NO ACTION` enforcement rejects the parent delete but provides no private guard header, so the client classifies it as literal `CONFLICT_UNMARKED`. This demonstrates that database enforcement, not the future-Ent guard, stopped deletion.

Desired protection: literal `PREDELETE` for the generic 409 plus exact private marker before database enforcement. The guard must cover `CalendarMutationReceipt.actor_user_id -> User CASCADE` when that Ent intent becomes active.

Cleanup: disarm through the matching closed DELETE blocker route, then perform ordinary fixture cleanup. The source-first test preserves the primary assertion over every cleanup failure.

Evidence: `NOT_RUN`. This case is intentionally RED until the typed pre-delete guard and private blocker routes exist.
