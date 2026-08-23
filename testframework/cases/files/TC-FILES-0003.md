+++
schema = "hub.testframework.case/v1"
id = "TC-FILES-0003"
title = "Confirm and delete each admit only one state transition"
status = "active"
task = "TASK-214"
requirement = "FILES-SINGLE-STATE-TRANSITION"
risk = "RISK-FILES-DUPLICATE-TRANSITION"
scenario = "boundary"
subject = "confirmed and deleted file state boundaries"
preconditions = ["The files runtime owns one uploaded pending file and an authorized owner session"]
oracle = "First confirmation succeeds, repeated confirmation returns 409, first deletion succeeds, and repeated deletion returns 404"
cleanup = "The files runtime removes only its recorded object, metadata, membership, project, identities, and sessions"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/files/test_files_api.py::test_confirm_and_delete_are_single_transition"
+++
# Single transitions

No client idempotency key is supplied for either transition.
