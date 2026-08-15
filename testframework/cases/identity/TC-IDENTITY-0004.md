+++
schema = "hub.testframework.case/v1"
id = "TC-IDENTITY-0004"
title = "Sessions remain owner scoped and revocation invalidates its target"
status = "active"
task = "TASK-207"
requirement = "IDENTITY-SESSION-OWNERSHIP"
risk = "RISK-SESSION-OWNERSHIP"
scenario = "authz"
subject = "owner and foreign session lifecycle"
preconditions = ["The identity harness provisions an owner with two sessions and a foreign session"]
oracle = "Only owner sessions are listed, foreign revocation is absent, and revoking an owned session invalidates its refresh"
cleanup = "The identity harness removes every provisioned identity and session"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/identity/test_identity_api.py::test_sessions_are_owner_scoped_and_revoke_invalidates_target_refresh"
+++
# Session ownership isolation

The session boundary lists only the owner set and revokes only an owned session.
