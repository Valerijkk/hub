+++
schema = "hub.testframework.case/v1"
id = "TC-AUTHZ-0001"
title = "System administration registration mode enforces the access matrix"
status = "active"
task = "TASK-208"
requirement = "AUTHZ-SYSTEM-ADMIN-REGISTRATION-MODE"
risk = "RISK-ADMIN-ACCESS-MATRIX"
scenario = "authz"
subject = "registration mode administration"
preconditions = ["The authorization harness provisions active founder, admin, and member principals"]
oracle = "Anonymous and member callers are denied while founder and system admin receive only an allowed registration mode"
cleanup = "The authorization harness removes every exact principal and credential from its nonce-bound ledger"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/authz/test_authz_api.py::test_system_admin_read_matrix"
+++
# System administration access matrix

The registration-mode boundary is observable only to the founder and system administrator.
