+++
schema = "hub.testframework.case/v1"
id = "TC-HEALTH-0001"
title = "Live health endpoint returns the canonical availability payload"
status = "active"
task = "TASK-218"
requirement = "HEALTH-LIVE"
risk = "RISK-HIGH"
scenario = "happy"
subject = "anonymous live health probe"
preconditions = ["A fake-only health transport returns the canonical payload"]
oracle = "The probe uses GET /api/live without authorization and parses only status ok"
cleanup = "No external resources are created"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/health/test_health_api.py::test_live_health_uses_fixed_anonymous_probe_and_returns_parsed_result"
+++
# Live health canonical response

The API health adapter admits only the fixed liveness response required by this case.
