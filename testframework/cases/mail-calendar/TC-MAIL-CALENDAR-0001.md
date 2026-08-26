+++
schema = "hub.testframework.case/v1"
id = "TC-MAIL-CALENDAR-0001"
title = "QA fixture control plane rejects an invalid capability and cleans an exact created fixture"
status = "active"
task = "TASK-237"
requirement = "MAIL-QA-FIXTURE-LIFECYCLE"
risk = "RISK-MAIL-QA-CONTROL-PLANE-LEAK"
scenario = "authn"
subject = "QA-only Mail and Calendar fixture lifecycle"
preconditions = ["The Docker-only python-api profile provides the fixed internal QA backend URL and control-plane capability token"]
oracle = "A wrong fixed-length token receives only the generic client failure; the authorized client receives one typed created fixture with its nonce-bound owner identity"
cleanup = "The test deletes only its one un-published created fixture in finally; no published fixture is created"

[automation]
kind = "pytest"
layer = "api"
node_id = "testframework/tests/api/mail_calendar/test_mail_calendar_qa_api.py::test_mail_calendar_qa_fixture_control_plane_requires_capability_and_cleans_created_fixture"
+++
# QA fixture lifecycle

The API scenario couples an authentication-negative control-plane probe with the happy
create-and-delete lifecycle. It executes only through the disposable Docker
`python-api` profile and never publishes an inbound fixture.
