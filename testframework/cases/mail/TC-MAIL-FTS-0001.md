+++
schema = "hub.testframework.case/v1"
id = "TC-MAIL-FTS-0001"
title = "Non-contiguous subject and snippet tokens return exact ordered live owner threads through the public PostgreSQL-backed route"
status = "active"
task = "TASK-237"
requirement = "MAIL-THREAD-FTS-TOKEN-MEMBERSHIP"
risk = "RISK-MAIL-THREAD-FTS-PHRASE-FALLBACK"
scenario = "happy"
subject = "Public authenticated list search by subject"
preconditions = ["Docker python-integration isolated qa-backend and QA fixture token are available."]
oracle = "Searching `qa calendar` returns exactly the four matching thread IDs in the prescribed order, excludes both one-token decoys, and returns no cursor."
cleanup = "Delete exactly the created fixtures; cleanup errors never mask the primary failure."

[automation]
kind = "pytest"
layer = "integration"
node_id = "testframework/tests/integration/mail/test_mail_threads_fts.py::test_mail_thread_search_returns_only_ordered_matching_threads"
+++

Exercise the public HTTP route only. Do not use a database driver or assert planner behavior. PostgreSQL index and catalog details are owned by schema-maintenance.
