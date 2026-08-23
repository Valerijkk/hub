+++
schema = "hub.testframework.case/v1"
id = "TC-IDENTITY-0002"
title = "Refresh rotation replaces the old credential and rejects replay"
status = "active"
task = "TASK-207"
requirement = "IDENTITY-REFRESH-ROTATION"
risk = "RISK-REFRESH-REPLAY"
scenario = "security"
subject = "active refresh credential rotation"
preconditions = ["The identity harness provisions one active credential"]
oracle = "A valid refresh returns a distinct token pair and the original refresh credential is rejected on replay"
cleanup = "The identity harness removes every provisioned identity and session"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/identity/test_identity_api.py::test_refresh_rotates_and_replay_is_rejected"
+++
# Refresh replay protection

The refresh flow must rotate credentials and prevent replay of the old refresh token.
