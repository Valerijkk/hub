+++
schema = "hub.testframework.case/v1"
id = "TC-MAIL-CALENDAR-0001"
title = "QA account-materialization control plane rejects an invalid capability and cleans its exact fixture"
status = "active"
task = "TASK-237"
requirement = "MAIL-ACCOUNT-MATERIALIZATION-LIFECYCLE"
risk = "RISK-MAIL-ACCOUNT-MATERIALIZATION-CONTROL-PLANE-LEAK"
scenario = "authn"
subject = "QA-only Mail account-materialization fixture lifecycle"
preconditions = ["The Docker-only python-api profile provides the fixed internal QA backend URL and control-plane capability token"]
oracle = "An invalid fixed-length token receives only the generic client failure; the authorized client creates a nonce-bound owner-root account-materialization-v1 fixture with closed initial state and accepts quiesce and delete acknowledgements"
cleanup = "After a successful create, finally quiesces then deletes only that fixture nonce; a cleanup failure never masks a primary test failure and otherwise fails the test"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/mail_calendar/test_mail_calendar_qa_api.py::test_mail_calendar_qa_account_materialization_requires_capability_and_cleans_fixture"
+++
# QA account-materialization lifecycle

The API scenario couples an authentication-negative control-plane probe with the public
`create -> quiesce -> delete` lifecycle of one account-materialization fixture. It executes
only through the disposable Docker `python-api` profile. It does not poll, await a worker,
inspect persistence, inject faults, contact an external provider, or exercise UI behavior.
