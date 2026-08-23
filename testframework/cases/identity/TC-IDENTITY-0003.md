+++
schema = "hub.testframework.case/v1"
id = "TC-IDENTITY-0003"
title = "Active bearer identity is projected and logout revokes refresh"
status = "active"
task = "TASK-207"
requirement = "IDENTITY-SESSION-REVOCATION"
risk = "RISK-LOGOUT-REVOCATION"
scenario = "authn"
subject = "active and deactivated bearer session access"
preconditions = ["The identity harness provisions active and deactivated credentials"]
oracle = "Only the active bearer can read its identity, and logout invalidates that session refresh credential"
cleanup = "The identity harness removes every provisioned identity and session"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/identity/test_identity_api.py::test_me_requires_active_bearer_and_logout_revokes_refresh"
+++
# Identity session revocation

The identity API projects only an active bearer and logout invalidates its refresh credential.
