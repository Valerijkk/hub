+++
schema = "hub.testframework.case/v1"
id = "TC-OUTBOX-0004"
title = "Malformed project JSON leaves zero owned final counts"
status = "active"
task = "TASK-211"
requirement = "PROJECT-CREATE-MALFORMED-JSON"
risk = "RISK-MALFORMED-PROJECT-OWNED-STATE"
scenario = "negative"
subject = "malformed authenticated project request"
preconditions = ["The project outbox runtime owns an authenticated owner, a clean nonce-bound ledger, and a pre-created durable observer"]
oracle = "The syntactically malformed JSON request returns 400 and final owned counts contain no project, membership, outbox row, or published outbox row"
cleanup = "The runtime quiesces its exact server and removes only its observer, JetStream, MongoDB, Redis, and PostgreSQL resources in the specified reverse order"

[automation]
kind = "pytest"
layer = "integration"
node_id = "testframework/tests/integration/redis-nats/project_outbox/test_project_outbox.py::test_malformed_project_json_is_rejected_with_zero_owned_final_counts"
+++
# Malformed project JSON

The case covers exact final owned counts after syntactic JSON rejection, without write atomicity or acknowledgement-order inference.
