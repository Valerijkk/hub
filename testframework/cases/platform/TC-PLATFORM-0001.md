+++
schema = "hub.testframework.case/v1"
id = "TC-PLATFORM-0001"
title = "Magic verification enforces the production request bound"
status = "active"
task = "TASK-209"
requirement = "PLATFORM-MAGIC-VERIFY-RATE-LIMIT"
risk = "RISK-MAGIC-VERIFY-ABUSE"
scenario = "security"
subject = "public magic-link verification rate admission"
preconditions = ["The platform abuse harness binds DevMode=false, healthy exclusive Redis and database state, and one server-observed client identity"]
oracle = "Four identical valid-shape unknown-code requests return 401, 401, 401, and 429; only the fourth response has Retry-After 600 without sensitive diagnostics"
cleanup = "The platform abuse harness removes only the four exact nonce-bound state entries it recorded"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/platform/test_platform_abuse.py::test_magic_verify_has_bounded_production_rate_limit"
+++
# Bounded magic verification

The test observes the immediate public abuse boundary; it does not infer expiry or recovery.
