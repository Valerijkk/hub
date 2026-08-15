+++
schema = "hub.testframework.case/v1"
id = "TC-OUTBOX-0001"
title = "Project creation yields one consistent published final state"
status = "active"
task = "TASK-211"
requirement = "PROJECT-CREATE-FINAL-OUTBOX-STATE"
risk = "RISK-PROJECT-FINAL-STATE-DIVERGENCE"
scenario = "happy"
subject = "authenticated project creation and durable final observation"
preconditions = ["The project outbox runtime owns an authenticated owner and a pre-created durable observer on the InterestPolicy EVENTS stream"]
oracle = "Creation returns 201; its projection, one durable typed project.created receipt, and final counts for one project, owner membership, outbox row, and published outbox row agree without inferring write atomicity or acknowledgement order"
cleanup = "The runtime quiesces its exact server and removes only its observer, JetStream, MongoDB, Redis, and PostgreSQL resources in the specified reverse order"

[automation]
kind = "pytest"
layer = "integration"
node_id = "testframework/tests/integration/redis-nats/project_outbox/test_project_outbox.py::test_project_create_returns_consistent_published_final_state"
+++
# Consistent published final state

The case observes only the public response, one durable typed receipt, and closed final counts.
