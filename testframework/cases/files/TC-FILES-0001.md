+++
schema = "hub.testframework.case/v1"
id = "TC-FILES-0001"
title = "Presigned upload completes a downloadable soft-delete lifecycle"
status = "active"
task = "TASK-214"
requirement = "FILES-PRESIGNED-LIFECYCLE"
risk = "RISK-FILES-OBJECT-METADATA-DIVERGENCE"
scenario = "happy"
subject = "owner-scoped file lifecycle"
preconditions = ["The files runtime owns one active principal, one project membership with upload, view, and delete permissions, and an exact task-owned object key"]
oracle = "The owner uploads exact bytes, confirms with the empty entity binding, downloads the same digest, soft-deletes metadata, and then receives 404"
cleanup = "The files runtime removes only its recorded object, metadata, membership, project, identities, and sessions"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/files/test_files_api.py::test_presigned_lifecycle_download_and_soft_delete"
+++
# Presigned file lifecycle

The scenario observes public metadata transitions and an owned object through opaque tickets.
