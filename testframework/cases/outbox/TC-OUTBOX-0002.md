+++
schema = "hub.testframework.case/v1"
id = "TC-OUTBOX-0002"
title = "Duplicate project key leaves one consistent final state"
status = "active"
task = "TASK-211"
requirement = "PROJECT-KEY-UNIQUE-OUTBOX"
risk = "RISK-DUPLICATE-PROJECT-FINAL-STATE"
scenario = "negative"
subject = "duplicate project key conflict"
preconditions = ["The project outbox runtime owns an authenticated owner and a pre-created durable observer on the InterestPolicy EVENTS stream"]
oracle = "The first create returns 201 with one durable typed receipt, the duplicate returns 409, and final state remains exactly one project, owner membership, outbox row, published outbox row, and receipt"
cleanup = "The runtime quiesces its exact server and removes only its observer, JetStream, MongoDB, Redis, and PostgreSQL resources in the specified reverse order"

[automation]
kind = "pytest"
layer = "integration"
node_id = "testframework/tests/integration/redis-nats/project_outbox/test_project_outbox.py::test_duplicate_project_key_conflict_leaves_one_consistent_final_state"
+++
# Duplicate project key

The conflict boundary observes only the one-receipt final state and excludes write atomicity or acknowledgement-order inference.
