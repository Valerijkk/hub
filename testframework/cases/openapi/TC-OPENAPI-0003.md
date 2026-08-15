+++
schema = "hub.testframework.case/v1"
id = "TC-OPENAPI-0003"
title = "TMS webhooks document real bearer and API key protection"
status = "active"
task = "TASK-227"
requirement = "TMS-WEBHOOK-AUTH-DRIFT"
risk = "RISK-TMS-WEBHOOK-AUTH-DRIFT"
scenario = "authz"
subject = "generated TMS webhook operations and sealed middleware execution chain"
preconditions = ["The launcher injects the v3 sealed contract source bundle with server wiring, API key middleware, auth middleware, TMS routes, and Swagger witnesses"]
oracle = "Server wiring orders auth before TMS, X-API-Key verification sets userID before Next, requireAuth admits that identity, and both webhook operations expose the exact BearerAuth or APIKeyAuth alternatives plus PermTestRunCreate without legacy token authentication"
cleanup = "The injected source bundle remains immutable and requires no cleanup"

[automation]
kind = "pytest"
layer = "contract"
node_id = "testframework/tests/contract/openapi/test_openapi_security_contract.py::test_tms_webhooks_document_real_auth_not_legacy_token"
+++
# TMS webhook authorization

The document must match the sealed server and middleware execution chain for both webhook operations, including exact BearerAuth or APIKeyAuth alternatives and the named PermTestRunCreate authorization boundary.
