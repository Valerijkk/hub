+++
schema = "hub.testframework.case/v1"
id = "TC-AUTHZ-0003"
title = "Deactivation revokes ordinary credentials without disabling the founder"
status = "active"
task = "TASK-208"
requirement = "AUTHZ-DEACTIVATION-REVOCATION"
risk = "RISK-REVOKED-CREDENTIAL-SURVIVAL"
scenario = "security"
subject = "ordinary principal deactivation"
preconditions = ["The authorization harness provisions active founder, admin, and ordinary member credentials"]
oracle = "Admin deactivation succeeds for the ordinary principal, revokes its access and refresh, and preserves founder access"
cleanup = "The authorization harness removes every exact principal and credential from its nonce-bound ledger"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/authz/test_authz_api.py::test_deactivation_revokes_credentials_and_preserves_founder"
+++
# Deactivation revocation

An ordinary account loses both access and refresh credentials after administrative deactivation.
