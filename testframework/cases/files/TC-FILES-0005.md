+++
schema = "hub.testframework.case/v1"
id = "TC-FILES-0005"
title = "Upload admission rejects empty oversized and unsafe media requests"
status = "active"
task = "TASK-214"
requirement = "FILES-UPLOAD-ADMISSION-BOUNDS"
risk = "RISK-FILES-UNSAFE-UPLOAD-ADMISSION"
scenario = "boundary"
subject = "upload metadata validation before presigning"
preconditions = ["The files runtime owns an authorized principal and project but has created no file object"]
oracle = "Sizes zero and 26214401 and content type text/html each return 400 without supplying a large body or receiving an upload ticket"
cleanup = "The files runtime removes only its recorded membership, project, identities, sessions, and any exact metadata recorded by the delegate"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/files/test_files_api.py::test_upload_rejects_bounded_size_and_unsafe_media"
+++
# Upload admission boundaries

The oversized case sends only the declared byte count and never allocates a 25 MiB payload.
