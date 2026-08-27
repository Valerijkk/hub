+++
schema = "hub.testframework.case/v1"
id = "TC-MAIL-CALENDAR-0004"
title = "Account materialization finalizes with stable public roots and pending primary identity"
status = "active"
task = "TASK-237"
requirement = "MAIL-ACCOUNT-MATERIALIZATION-PROJECTION"
risk = "RISK-MAIL-ACCOUNT-MATERIALIZATION-DIVERGENCE"
scenario = "happy"
subject = "Asynchronous Mail account materialization fixture"
preconditions = ["The Docker-only python-integration profile provides isolated QA infrastructure, a QA backend URL, the Mail and Calendar control-plane capability token, and the Mail account materialization worker"]
oracle = "The nonce-bound materialization finalizes within the bounded public poll, keeps immutable owner and root identifiers, and exposes the expected account plus one pending primary identity across a second public observation"
cleanup = "The test quiesces and deletes only its one created account-materialization fixture in finally; cleanup does not mask a primary assertion or polling failure"

[automation]
kind = "pytest"
layer = "integration"
node_id = "testframework/tests/integration/mail-calendar/test_mail_account_materialization.py::test_account_materialization_eventually_finalizes_with_stable_public_roots"
+++
# Account materialization projection

The scenario creates exactly one QA-only account-materialization fixture and waits through
the public control-plane observation until the asynchronous worker finalizes it. A second
public observation before cleanup proves the immutable roots and resulting account and
primary identity stay consistent. It uses no database, broker, raw HTTP, private controller,
external mail server, or injected worker fault. The test explicitly quiesces before deletion
because the fixture lifecycle requires the worker to stop touching the owned rows first.
