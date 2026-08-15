+++
schema = "hub.testframework.case/v1"
id = "TC-PRIVACY-0001"
title = "Public error artifact matches the redacted wire shape"
status = "active"
task = "TASK-227"
requirement = "PUBLIC-ERROR-ARTIFACT"
risk = "RISK-PUBLIC-ERROR-ARTIFACT"
scenario = "security"
subject = "generated public error definition and centralized error writer"
preconditions = ["The launcher injects one sealed contract source bundle"]
oracle = "The generated public error reference exposes only error and request_id from the centralized writer without internal AppError fields or secret examples"
cleanup = "The injected source bundle remains immutable and requires no cleanup"

[automation]
kind = "pytest"
layer = "contract"
node_id = "testframework/tests/contract/privacy/test_public_error_contract.py::test_public_error_artifact_is_redacted_to_wire_schema"
+++
# Public error artifact

The generated contract must expose only the canonical public error wire shape and must not publish internal error state.
