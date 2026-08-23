+++
schema = "hub.testframework.case/v1"
id = "TC-FILES-0002"
title = "Missing uploaded object can be confirmed only after the same ticket is used"
status = "active"
task = "TASK-214"
requirement = "FILES-CONFIRM-OBJECT-PRESENCE"
risk = "RISK-FILES-PREMATURE-CONFIRMATION"
scenario = "failure-recovery"
subject = "pending file confirmation"
preconditions = ["The files runtime owns one pending metadata record and its opaque upload ticket but no object bytes"]
oracle = "Confirmation returns 404 before upload, then the same ticket stores exact bytes and confirmation returns the exact confirmed response"
cleanup = "The files runtime removes only its recorded object, metadata, membership, project, identities, and sessions"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/files/test_files_api.py::test_confirm_missing_object_then_retry_succeeds"
+++
# Missing object recovery

The retry is an explicit second public request after object placement; it uses no polling or delay.
