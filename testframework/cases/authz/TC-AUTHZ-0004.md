+++
schema = "hub.testframework.case/v1"
id = "TC-AUTHZ-0004"
title = "Pending administration does not enumerate active principals"
status = "active"
task = "TASK-208"
requirement = "AUTHZ-PENDING-NONENUMERATION"
risk = "RISK-ADMIN-PRINCIPAL-ENUMERATION"
scenario = "security"
subject = "pending administration lookup boundary"
preconditions = ["The authorization harness provisions an active admin and a distinct active principal"]
oracle = "Approving an active principal and an unknown UUID produce equivalent generic denials without exposing owned canaries"
cleanup = "The authorization harness removes every exact principal and credential from its nonce-bound ledger"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/authz/test_authz_api.py::test_pending_admin_endpoint_does_not_enumerate_active_users"
+++
# Pending non-enumeration

The pending approval endpoint must not distinguish an active principal from an unknown identifier.
