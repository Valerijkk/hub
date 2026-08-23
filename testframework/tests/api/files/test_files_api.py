import pytest
from hub_testkit.harness.files import (
    assert_presigned_lifecycle_download_and_soft_delete,
    assert_confirm_missing_object_then_retry_succeeds,
    assert_confirm_and_delete_are_single_transition,
    assert_files_require_authz_and_hide_storage_artifacts,
    assert_upload_rejects_bounded_size_and_unsafe_media,
)


@pytest.mark.case_id("TC-FILES-0001")
@pytest.mark.api
def test_presigned_lifecycle_download_and_soft_delete(files_runtime):
    assert_presigned_lifecycle_download_and_soft_delete(files_runtime)


@pytest.mark.case_id("TC-FILES-0002")
@pytest.mark.api
def test_confirm_missing_object_then_retry_succeeds(files_runtime):
    assert_confirm_missing_object_then_retry_succeeds(files_runtime)


@pytest.mark.case_id("TC-FILES-0003")
@pytest.mark.api
def test_confirm_and_delete_are_single_transition(files_runtime):
    assert_confirm_and_delete_are_single_transition(files_runtime)


@pytest.mark.case_id("TC-FILES-0004")
@pytest.mark.api
def test_files_require_authz_and_hide_storage_artifacts(files_runtime):
    assert_files_require_authz_and_hide_storage_artifacts(files_runtime)


@pytest.mark.case_id("TC-FILES-0005")
@pytest.mark.api
def test_upload_rejects_bounded_size_and_unsafe_media(files_runtime):
    assert_upload_rejects_bounded_size_and_unsafe_media(files_runtime)
