+++
schema = "hub.testframework.case/v1"
id = "TC-OPENAPI-0002"
title = "Magic verify documents the implemented error contract"
status = "active"
task = "TASK-227"
requirement = "MAGIC-VERIFY-ERROR-CONTRACT"
risk = "RISK-MAGIC-VERIFY-ERROR-CONTRACT"
scenario = "negative"
subject = "generated magic verify operation and auth handler annotation"
preconditions = ["The launcher injects one sealed contract source bundle"]
oracle = "The generated operation matches only the contiguous VerifyMagicLink annotation block with its request, token response, and four named denial statuses and has no secret examples"
cleanup = "The injected source bundle remains immutable and requires no cleanup"

[automation]
kind = "pytest"
layer = "contract"
node_id = "testframework/tests/contract/openapi/test_openapi_security_contract.py::test_magic_verify_operation_lists_implemented_error_contract"
+++
# Magic verify error contract

The document must derive its public success and denial outcomes only from the contiguous annotation block attached to VerifyMagicLink and must not expose token examples.
