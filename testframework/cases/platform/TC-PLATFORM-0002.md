+++
schema = "hub.testframework.case/v1"
id = "TC-PLATFORM-0002"
title = "Magic verification rejects one-time code replay"
status = "active"
task = "TASK-209"
requirement = "PLATFORM-MAGIC-VERIFY-REPLAY"
risk = "RISK-MAGIC-VERIFY-REPLAY"
scenario = "security"
subject = "public one-time magic-link verification"
preconditions = ["The platform abuse harness provisions one owned active identity and one one-time code without delivery or development login"]
oracle = "The first public verification returns a strict token pair and an identical replay returns 401 without exposing credentials or ownership state"
cleanup = "The platform abuse harness removes only the four exact nonce-bound state entries it recorded"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/platform/test_platform_abuse.py::test_magic_verify_rejects_replay_without_secret_leak"
+++
# One-time verification replay

The test uses the public verification route only; provisioning and cleanup remain launcher-owned.
