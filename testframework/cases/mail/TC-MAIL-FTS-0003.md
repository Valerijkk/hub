+++
schema = "hub.testframework.case/v1"
id = "TC-MAIL-FTS-0003"
title = "Private thread-search fixture delete guard stops accepted Atlas SetNull before deleting the owner"
status = "active"
task = "TASK-237"
requirement = "MAIL-THREAD-FTS-FIXTURE-SAFE-DELETE-GUARD"
risk = "RISK-MAIL-THREAD-FTS-UNGUARDED-ATLAS-SETNULL"
scenario = "failure-recovery"
subject = "Guard ordering for accepted AIProvider owner SetNull"
preconditions = ["Docker python-integration has the private QA fixture seam and a valid QA token."]
oracle = "After arming `atlas-user-set-null-v1`, DELETE must be classified as literal `PREDELETE`: generic 409 body and `X-Hub-QA-Fixture-Delete-Guard: predelete` before any owner deletion."
cleanup = "Disarm the exact blocker, then delete the fixture normally. A stored `DELETED` observation alone permits already-absent target cleanup; cleanup errors never mask the primary failure."

[automation]
kind = "pytest"
layer = "integration"
node_id = "testframework/tests/integration/mail/test_mail_thread_fixture_delete_guard.py::test_mail_thread_fixture_delete_guard_stops_atlas_set_null_before_delete"
+++

Setup: create one `thread-search-v1` fixture and arm the closed single-fixture blocker endpoint `POST /api/_qa/mail-calendar/fixtures/:nonce/thread-search-delete-blocker/atlas-user-set-null-v1`.

Action: observe the guarded fixture DELETE through the typed QA client; it owns the exact private route and recognizes only the closed delete observations.

Current semantic RED: current unguarded Atlas behavior deletes the target and is classified as literal `DELETED`, proving that `AIProvider.owner_user_id` could be SetNull before the fixture guard runs.

Desired protection: literal `PREDELETE` for the exact generic 409 response plus guard marker, before any target parent is deleted. No dependent type or private row identifier is exposed.

Cleanup: disarm via the matching closed DELETE blocker route, then perform ordinary fixture cleanup. The source-first test preserves the primary assertion over every cleanup failure.

Evidence: `NOT_RUN`. This case is intentionally RED until the typed pre-delete guard and private blocker routes exist.
