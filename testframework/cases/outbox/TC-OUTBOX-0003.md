+++
schema = "hub.testframework.case/v1"
id = "TC-OUTBOX-0003"
title = "Anonymous project creation leaves zero owned final counts"
status = "active"
task = "TASK-211"
requirement = "PROJECT-CREATE-AUTHENTICATION"
risk = "RISK-ANONYMOUS-PROJECT-OWNED-STATE"
scenario = "authn"
subject = "anonymous project creation rejection"
preconditions = ["The project outbox runtime owns a clean nonce-bound ledger and a pre-created durable observer on the InterestPolicy EVENTS stream"]
oracle = "The anonymous request returns 401 and final owned counts contain no project, membership, outbox row, or published outbox row"
cleanup = "The runtime quiesces its exact server and removes only its observer, JetStream, MongoDB, Redis, and PostgreSQL resources in the specified reverse order"

[automation]
kind = "pytest"
layer = "integration"
node_id = "testframework/tests/integration/redis-nats/project_outbox/test_project_outbox.py::test_anonymous_project_create_is_rejected_with_zero_owned_final_counts"
+++
# Anonymous create rejection

The case observes only the public authentication boundary and exact final owned counts, without write atomicity or acknowledgement-order inference.
