+++
schema = "hub.testframework.case/v1"
id = "TC-OUTBOX-0005"
title = "Durable project observer replays after reconnect"
status = "active"
task = "TASK-211"
requirement = "PROJECT-CREATED-DURABLE-OBSERVER"
risk = "RISK-PROJECT-EVENT-LOST-DURING-OBSERVER-DISCONNECT"
scenario = "failure-recovery"
subject = "pre-created durable observer reconnect"
preconditions = ["The project outbox runtime owns an authenticated owner and a durable observer created before publication on the InterestPolicy EVENTS stream"]
oracle = "A project created while the observer is disconnected returns 201; after reconnect, the exact observer receives the same durable typed receipt and final state contains one project, owner membership, outbox row, and published outbox row"
cleanup = "The runtime quiesces its exact server and removes only its observer, JetStream, MongoDB, Redis, and PostgreSQL resources in the specified reverse order"

[automation]
kind = "pytest"
layer = "integration"
node_id = "testframework/tests/integration/redis-nats/project_outbox/test_project_outbox.py::test_durable_project_created_observer_replays_after_observer_reconnect"
+++
# Durable observer replay

The case proves observer reconnect replay and consistent final counts only; it does not infer acknowledgement order, write atomicity, or worker process recovery.
