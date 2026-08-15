+++
schema = "hub.testframework.case/v1"
id = "TC-OPENAPI-0001"
title = "OpenAPI has canonical header security and one API prefix"
status = "active"
task = "TASK-227"
requirement = "OPENAPI-CANONICAL-DOCUMENT"
risk = "RISK-OPENAPI-CANONICAL-DOCUMENT"
scenario = "security"
subject = "generated OpenAPI document and server route prefix"
preconditions = ["The launcher injects one sealed contract source bundle"]
oracle = "Header security definitions and the generated path prefix match the named server source relation"
cleanup = "The injected source bundle remains immutable and requires no cleanup"

[automation]
kind = "pytest"
layer = "contract"
node_id = "testframework/tests/contract/openapi/test_openapi_security_contract.py::test_swagger_document_is_valid_with_single_api_prefix"
+++
# Canonical OpenAPI document

The generated document must preserve the named header security and one composed API prefix.
