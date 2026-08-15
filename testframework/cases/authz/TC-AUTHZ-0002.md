+++
schema = "hub.testframework.case/v1"
id = "TC-AUTHZ-0002"
title = "Pending approvals and rejections require a system administrator"
status = "active"
task = "TASK-208"
requirement = "AUTHZ-PENDING-ADMINISTRATION"
risk = "RISK-PENDING-PRIVILEGE-ESCALATION"
scenario = "authz"
subject = "pending principal administration"
preconditions = ["The authorization harness provisions an active admin, active member, and two pending principals"]
oracle = "A member is denied every pending administration action while an admin observes, approves, and rejects the exact owned pending principals"
cleanup = "The authorization harness removes every exact principal and credential from its nonce-bound ledger"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/authz/test_authz_api.py::test_pending_approval_and_rejection_are_admin_only"
+++
# Pending administration

Only the system administrator may resolve a pending principal lifecycle.
