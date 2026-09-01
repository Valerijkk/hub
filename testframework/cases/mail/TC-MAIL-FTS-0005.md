+++
schema = "hub.testframework.case/v1"
id = "TC-MAIL-FTS-0005"
title = "Private thread-search fixture delete guard stops a cross-fixture mail-message-folder relation by target message"
status = "active"
task = "TASK-237"
requirement = "MAIL-THREAD-FTS-FIXTURE-SAFE-DELETE-GUARD"
risk = "RISK-MAIL-THREAD-FTS-CROSS-FIXTURE-MESSAGE-RELATION"
scenario = "failure-recovery"
subject = "Independent MailMessageFolder by-message exclusivity direction"
preconditions = ["Docker python-integration has the private QA fixture seam and a valid QA token."]
oracle = "With two distinct fixtures and `mail-message-folder-by-message-extra-v1` armed, target DELETE must be literal `PREDELETE`; only the target row-0 message points to the peer folder."
cleanup = "Disarm the paired blocker before ordinary target then peer deletion. A stored target `DELETED` alone permits target absence; peer cleanup remains mandatory; cleanup errors never mask the primary failure."

[automation]
kind = "pytest"
layer = "integration"
node_id = "testframework/tests/integration/mail/test_mail_thread_fixture_delete_guard.py::test_mail_thread_fixture_delete_guard_stops_extra_relation_by_target_message"
+++

Setup: create distinct target and peer `thread-search-v1` fixtures and arm the closed peer route `POST /api/_qa/mail-calendar/fixtures/:nonce/thread-search-delete-blocker/mail-message-folder-by-message-extra-v1/peer/:peer_nonce`.

Action: observe target fixture DELETE through the typed QA client. The fixture seam owns exactly one extra relation from target row-0 message to peer folder.

Current semantic RED: accepted Atlas reaches `NO ACTION` while the cross-fixture relation remains, returning generic 409 without the guard marker. The client classifies this as literal `CONFLICT_UNMARKED`.

Desired protection: literal `PREDELETE` before deleting target children, proving the independent by-message direction is observed and rejected by the typed guard.

Cleanup: disarm the exact paired relation first, delete the target unless the stored observation was `DELETED`, then delete the peer unconditionally. The source-first test preserves the primary assertion over every cleanup failure.

Evidence: `NOT_RUN`. This case is intentionally RED until the typed pre-delete guard and private blocker routes exist.
