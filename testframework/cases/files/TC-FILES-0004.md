+++
schema = "hub.testframework.case/v1"
id = "TC-FILES-0004"
title = "File operations require authorization without exposing storage artifacts"
status = "active"
task = "TASK-214"
requirement = "FILES-AUTHORIZATION-BOUNDARY"
risk = "RISK-FILES-CROSS-PRINCIPAL-ACCESS"
scenario = "authz"
subject = "known owner file accessed anonymously and by a foreign principal"
preconditions = ["The files runtime owns one confirmed file, one authorized owner, and one authenticated foreign nonmember"]
oracle = "Anonymous operations return 401, foreign confirm, download, and delete return 403, and public artifacts disclose no owned filename, project, key, URL, credentials, or nonce"
cleanup = "The files runtime removes only its recorded object, metadata, membership, project, identities, and sessions while preserving foreign state outside its ledger"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/files/test_files_api.py::test_files_require_authz_and_hide_storage_artifacts"
+++
# File authorization and privacy

The case claims only behavior for a known owned file and makes no unknown-file enumeration claim.
