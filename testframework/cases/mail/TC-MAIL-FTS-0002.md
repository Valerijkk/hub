+++
schema = "hub.testframework.case/v1"
id = "TC-MAIL-FTS-0002"
title = "Owner scope, starred filtering, stable order, and legacy search route 404 are preserved"
status = "active"
task = "TASK-237"
requirement = "MAIL-THREAD-SEARCH-PUBLIC-CONTRACT"
risk = "RISK-MAIL-THREAD-SEARCH-SCOPE-OR-ORDER-REGRESSION"
scenario = "authz"
subject = "Authenticated thread search boundary"
preconditions = ["Docker python-api has two isolated QA fixtures."]
oracle = "Results have the exact owner order; foreign threads and decoys are absent; starred search returns the two pinned IDs; the legacy route returns 404."
cleanup = "Delete exactly the two created fixtures; cleanup errors never mask the primary failure."

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/mail/test_list_mail_threads_search.py::test_mail_thread_search_scopes_results_and_rejects_legacy_search_route"
+++

Exercise only the public authenticated contract. Do not inspect source code or access the database.
