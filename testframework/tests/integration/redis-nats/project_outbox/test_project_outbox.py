import pytest
from hub_testkit.harness.project_outbox import (
    assert_project_create_returns_consistent_published_final_state,
    assert_duplicate_project_key_conflict_leaves_one_consistent_final_state,
    assert_anonymous_project_create_is_rejected_with_zero_owned_final_counts,
    assert_malformed_project_json_is_rejected_with_zero_owned_final_counts,
    assert_durable_project_created_observer_replays_after_observer_reconnect,
)


@pytest.mark.case_id("TC-OUTBOX-0001")
@pytest.mark.integration
def test_project_create_returns_consistent_published_final_state(project_outbox_runtime):
    assert_project_create_returns_consistent_published_final_state(
        project_outbox_runtime
    )


@pytest.mark.case_id("TC-OUTBOX-0002")
@pytest.mark.integration
def test_duplicate_project_key_conflict_leaves_one_consistent_final_state(
    project_outbox_runtime,
):
    assert_duplicate_project_key_conflict_leaves_one_consistent_final_state(
        project_outbox_runtime
    )


@pytest.mark.case_id("TC-OUTBOX-0003")
@pytest.mark.integration
def test_anonymous_project_create_is_rejected_with_zero_owned_final_counts(
    project_outbox_runtime,
):
    assert_anonymous_project_create_is_rejected_with_zero_owned_final_counts(
        project_outbox_runtime
    )


@pytest.mark.case_id("TC-OUTBOX-0004")
@pytest.mark.integration
def test_malformed_project_json_is_rejected_with_zero_owned_final_counts(
    project_outbox_runtime,
):
    assert_malformed_project_json_is_rejected_with_zero_owned_final_counts(
        project_outbox_runtime
    )


@pytest.mark.case_id("TC-OUTBOX-0005")
@pytest.mark.integration
def test_durable_project_created_observer_replays_after_observer_reconnect(
    project_outbox_runtime,
):
    assert_durable_project_created_observer_replays_after_observer_reconnect(
        project_outbox_runtime
    )
