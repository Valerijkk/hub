+++
schema = "hub.testframework.case/v1"
id = "TC-IDENTITY-0001"
title = "Refresh rejects malformed, invalid, and deactivated credentials"
status = "active"
task = "TASK-207"
requirement = "IDENTITY-REFRESH-REJECTION"
risk = "RISK-REFRESH-REJECTION"
scenario = "negative"
subject = "anonymous refresh credential validation"
preconditions = ["The identity harness provisions one deactivated credential"]
oracle = "Malformed, unknown, and deactivated refresh credentials receive only their documented denial status"
cleanup = "The identity harness removes every provisioned identity and session"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/identity/test_identity_api.py::test_refresh_rejects_invalid_and_deactivated_credentials"
+++
# Refresh rejection

The refresh boundary rejects malformed and invalid credentials without creating a new session.
