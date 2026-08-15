+++
schema = "hub.testframework.case/v1"
id = "TC-SECURITY-0001"
title = "Session created event keeps its producer consumer link"
status = "active"
task = "TASK-227"
requirement = "AUTH-SESSION-EVENT-LINK"
risk = "RISK-AUTH-SESSION-EVENT-LINK"
scenario = "security"
subject = "session created event source contract"
preconditions = ["The launcher injects one sealed contract source bundle"]
oracle = "The exact SessionCreatedEvent JSON fields, auth.session.created publisher, notification mapping, and typed consumer remain linked"
cleanup = "The injected source bundle remains immutable and requires no cleanup"

[automation]
kind = "pytest"
layer = "contract"
node_id = "testframework/tests/contract/security/test_auth_session_event_contract.py::test_auth_session_created_event_has_exact_producer_consumer_link"
+++
# Session created event link

The source contract must preserve the exact typed event from session creation through notification dispatch without executing a broker.
