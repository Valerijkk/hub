from __future__ import annotations

from dataclasses import dataclass, field
from typing import NoReturn, Protocol, TypeGuard, cast, runtime_checkable

from hub_testkit.domains.files.api import FilesApi
from hub_testkit.domains.files.models import (
    DownloadTicket,
    FileStatus,
    FilesFlow,
    FilesSeed,
    ObjectDigest,
    ProvisionedFiles,
    UploadContent,
    UploadTicket,
    _download_ticket_details,
    _upload_content_details,
    _upload_ticket_details,
)
from hub_testkit.http import HttpExchange, HttpRequest


class FilesRuntimeError(ValueError):
    pass


@runtime_checkable
class FilesHarnessDelegate(Protocol):
    """Launcher-owned delegate for one fresh nonce-bound files scenario."""

    def send(self, request: HttpRequest) -> HttpExchange:
        """Send one admitted product API request through the injected boundary."""
        ...

    def provision(self, seed: FilesSeed) -> ProvisionedFiles:
        """Provision exact owner, foreign principal, project, membership, and ledger.

        The delegate fails before requests unless dependencies are healthy and exclusive.
        """
        ...

    def put(self, ticket: UploadTicket, content: UploadContent) -> None:
        """PUT exact bytes, type, and length to the local S3 URL without Authorization."""
        ...

    def fetch(self, ticket: DownloadTicket) -> ObjectDigest:
        """Fetch one exact local-S3 object and return only its length and digest."""
        ...

    def cleanup(self) -> None:
        """Delete exact recorded object and metadata, then membership, project, and actors.

        Cleanup never scans a prefix, bucket, database, namespace, or foreign ownership.
        """
        ...


@runtime_checkable
class FilesHarnessFactory(Protocol):
    """Future launcher supplies one fresh exclusive delegate for every test."""

    def __call__(self) -> FilesHarnessDelegate:
        """Create one delegate with an empty exact-resource ownership ledger."""
        ...


@dataclass(frozen=True, slots=True, repr=False, eq=False, init=False)
class FilesRuntime:
    _delegate: FilesHarnessDelegate = field(repr=False)
    _api: FilesApi = field(repr=False)
    _runtime_binding: object = field(repr=False)

    def __init__(self) -> None:
        _raise_invalid()

    @classmethod
    def _create(cls, delegate: FilesHarnessDelegate) -> FilesRuntime:
        result = _runtime_create_safe(cls, delegate)
        del delegate
        if type(result) is cls:
            return result
        del result
        _raise_invalid()

    def provision(self, seed: FilesSeed) -> ProvisionedFiles:
        result = _delegate_provision_safe(self._delegate, seed)
        del seed
        if type(result) is ProvisionedFiles:
            return result
        del result
        _raise_invalid()

    def put(self, ticket: UploadTicket, content: UploadContent) -> None:
        result = _runtime_put_safe(
            self._delegate,
            self._runtime_binding,
            ticket,
            content,
        )
        del ticket, content
        if result is True:
            return
        del result
        _raise_invalid()

    def fetch(self, ticket: DownloadTicket) -> ObjectDigest:
        result = _runtime_fetch_safe(
            self._delegate,
            self._runtime_binding,
            ticket,
        )
        del ticket
        if type(result) is ObjectDigest:
            return result
        del result
        _raise_invalid()

    def _cleanup(self) -> None:
        result = _delegate_cleanup_safe(self._delegate)
        if result is True:
            return
        del result
        _raise_invalid()


def assert_presigned_lifecycle_download_and_soft_delete(runtime: FilesRuntime) -> None:
    result = _lifecycle_safe(runtime)
    del runtime
    if result is True:
        return
    del result
    _raise_assertion()


def assert_confirm_missing_object_then_retry_succeeds(runtime: FilesRuntime) -> None:
    result = _missing_object_retry_safe(runtime)
    del runtime
    if result is True:
        return
    del result
    _raise_assertion()


def assert_confirm_and_delete_are_single_transition(runtime: FilesRuntime) -> None:
    result = _single_transition_safe(runtime)
    del runtime
    if result is True:
        return
    del result
    _raise_assertion()


def assert_files_require_authz_and_hide_storage_artifacts(runtime: FilesRuntime) -> None:
    result = _authorization_privacy_safe(runtime)
    del runtime
    if result is True:
        return
    del result
    _raise_assertion()


def assert_upload_rejects_bounded_size_and_unsafe_media(runtime: FilesRuntime) -> None:
    result = _upload_validation_safe(runtime)
    del runtime
    if result is True:
        return
    del result
    _raise_assertion()


_INVALID = object()
_CONTENT_BYTES = b"hub-files-owned-object-v1"
_CONTENT_TYPE = "application/octet-stream"
_FILENAME = "owned-private-evidence.bin"


def _runtime_create_safe(
    cls: type[FilesRuntime],
    delegate: object,
) -> FilesRuntime | object:
    try:
        if not _delegate_is_valid_safe(delegate):
            return _INVALID
        binding = object()
        api = FilesApi(sender=delegate, _runtime_binding=binding)
        result = object.__new__(cls)
        object.__setattr__(result, "_delegate", delegate)
        object.__setattr__(result, "_api", api)
        object.__setattr__(result, "_runtime_binding", binding)
        del api, binding
        return result
    except Exception:
        return _INVALID


def _runtime_put_safe(
    delegate: object,
    runtime_binding: object,
    ticket: object,
    content: object,
) -> bool:
    try:
        if not _delegate_is_valid_safe(delegate):
            return False
        ticket_details = _upload_ticket_details(ticket)
        content_details = _upload_content_details(content)
        if type(ticket_details) is not tuple or type(content_details) is not tuple:
            return False
        binding, _url, _key, expected_type, expected_length = ticket_details
        body, actual_type = content_details
        if (
            binding is not runtime_binding
            or expected_type != actual_type
            or expected_length != len(body)
            or type(ticket) is not UploadTicket
            or type(content) is not UploadContent
        ):
            return False
        result: object = delegate.put(ticket, content)
        del body, ticket_details, content_details
        return result is None
    except Exception:
        return False


def _runtime_fetch_safe(
    delegate: object,
    runtime_binding: object,
    ticket: object,
) -> ObjectDigest | object:
    try:
        if not _delegate_is_valid_safe(delegate):
            return _INVALID
        details = _download_ticket_details(ticket)
        if type(details) is not tuple or details[0] is not runtime_binding:
            return _INVALID
        if type(ticket) is not DownloadTicket:
            return _INVALID
        result = delegate.fetch(ticket)
        if type(result) is not ObjectDigest:
            return _INVALID
        return ObjectDigest(length=result.length, sha256=result.sha256)
    except Exception:
        return _INVALID


def _delegate_provision_safe(delegate: object, seed: object) -> ProvisionedFiles | object:
    try:
        if not _delegate_is_valid_safe(delegate) or type(seed) is not FilesSeed:
            return _INVALID
        result = delegate.provision(seed)
        if type(result) is not ProvisionedFiles:
            return _INVALID
        return ProvisionedFiles(
            owner=result.owner,
            foreign=result.foreign,
            project_id=result.project_id,
            nonce=result.nonce,
        )
    except Exception:
        return _INVALID


def _delegate_cleanup_safe(delegate: object) -> bool:
    try:
        if not _delegate_is_valid_safe(delegate):
            return False
        result: object = delegate.cleanup()
        return result is None
    except Exception:
        return False


def _delegate_is_valid_safe(value: object) -> TypeGuard[FilesHarnessDelegate]:
    try:
        return isinstance(value, FilesHarnessDelegate)
    except Exception:
        return False


def _lifecycle_safe(runtime: object) -> bool:
    try:
        prepared = _uploaded_file_safe(runtime, "lifecycle", "files-life")
        if type(prepared) is not tuple:
            return False
        scenario, content, ticket = prepared
        if type(runtime) is not FilesRuntime:
            return False
        confirmed = runtime._api.confirm(
            scenario.owner,
            ticket.file_id,
            "files-life-confirm",
        )
        download = runtime._api.download(
            scenario.owner,
            ticket.file_id,
            "files-life-download",
        )
        if type(confirmed) is not FileStatus or confirmed.status_code != 200:
            return False
        if type(download) is not DownloadTicket or download.file_id != ticket.file_id:
            return False
        digest = runtime.fetch(download)
        if digest.length != content.length or digest.sha256 != content.sha256:
            return False
        deleted = runtime._api.delete(
            scenario.owner,
            ticket.file_id,
            "files-life-delete",
        )
        absent = runtime._api.download(
            scenario.owner,
            ticket.file_id,
            "files-life-after-delete",
        )
        return (
            type(deleted) is FileStatus
            and deleted.status_code == 200
            and type(absent) is FileStatus
            and absent.status_code == 404
        )
    except Exception:
        return False


def _missing_object_retry_safe(runtime: object) -> bool:
    try:
        if type(runtime) is not FilesRuntime:
            return False
        scenario = runtime.provision(
            FilesSeed(label="files-missing", flow="missing-object-retry")
        )
        content = UploadContent.from_bytes(_CONTENT_BYTES, _CONTENT_TYPE)
        ticket = runtime._api.upload(
            scenario.owner,
            scenario.project_id,
            _FILENAME,
            _CONTENT_TYPE,
            content.length,
            "files-missing-upload",
        )
        if type(ticket) is not UploadTicket:
            return False
        missing = runtime._api.confirm(
            scenario.owner,
            ticket.file_id,
            "files-missing-confirm",
        )
        if type(missing) is not FileStatus or missing.status_code != 404:
            return False
        runtime.put(ticket, content)
        confirmed = runtime._api.confirm(
            scenario.owner,
            ticket.file_id,
            "files-missing-retry",
        )
        return type(confirmed) is FileStatus and confirmed.status_code == 200
    except Exception:
        return False


def _single_transition_safe(runtime: object) -> bool:
    try:
        prepared = _uploaded_file_safe(runtime, "single-transition", "files-once")
        if type(prepared) is not tuple or type(runtime) is not FilesRuntime:
            return False
        scenario, _content, ticket = prepared
        first_confirm = runtime._api.confirm(
            scenario.owner,
            ticket.file_id,
            "files-once-confirm-one",
        )
        second_confirm = runtime._api.confirm(
            scenario.owner,
            ticket.file_id,
            "files-once-confirm-two",
        )
        first_delete = runtime._api.delete(
            scenario.owner,
            ticket.file_id,
            "files-once-delete-one",
        )
        second_delete = runtime._api.delete(
            scenario.owner,
            ticket.file_id,
            "files-once-delete-two",
        )
        return _statuses_are_safe(
            (first_confirm, second_confirm, first_delete, second_delete),
            (200, 409, 200, 404),
        )
    except Exception:
        return False


def _authorization_privacy_safe(runtime: object) -> bool:
    try:
        prepared = _uploaded_file_safe(runtime, "authorization", "files-authz")
        if type(prepared) is not tuple or type(runtime) is not FilesRuntime:
            return False
        scenario, _content, ticket = prepared
        confirmed = runtime._api.confirm(
            scenario.owner,
            ticket.file_id,
            "files-authz-owner-confirm",
        )
        if type(confirmed) is not FileStatus or confirmed.status_code != 200:
            return False
        forbidden = _scenario_secrets_safe(scenario, ticket)
        if type(forbidden) is not tuple:
            return False
        anonymous = (
            runtime._api.confirm(
                None,
                ticket.file_id,
                "files-authz-anon-confirm",
                forbidden,
            ),
            runtime._api.download(
                None,
                ticket.file_id,
                "files-authz-anon-download",
                forbidden,
            ),
            runtime._api.delete(
                None,
                ticket.file_id,
                "files-authz-anon-delete",
                forbidden,
            ),
        )
        foreign = (
            runtime._api.confirm(
                scenario.foreign,
                ticket.file_id,
                "files-authz-foreign-confirm",
                forbidden,
            ),
            runtime._api.download(
                scenario.foreign,
                ticket.file_id,
                "files-authz-foreign-download",
                forbidden,
            ),
            runtime._api.delete(
                scenario.foreign,
                ticket.file_id,
                "files-authz-foreign-delete",
                forbidden,
            ),
        )
        return (
            _statuses_are_safe(anonymous, (401, 401, 401))
            and _statuses_are_safe(foreign, (403, 403, 403))
            and _representations_hide_safe(runtime, scenario, ticket, forbidden)
            and _generic_failure_is_safe(runtime, ticket, forbidden)
        )
    except Exception:
        return False


def _upload_validation_safe(runtime: object) -> bool:
    try:
        if type(runtime) is not FilesRuntime:
            return False
        scenario = runtime.provision(
            FilesSeed(label="files-bounds", flow="upload-validation")
        )
        results = (
            runtime._api.upload(
                scenario.owner,
                scenario.project_id,
                _FILENAME,
                _CONTENT_TYPE,
                0,
                "files-bounds-zero",
            ),
            runtime._api.upload(
                scenario.owner,
                scenario.project_id,
                _FILENAME,
                _CONTENT_TYPE,
                25 * 1024 * 1024 + 1,
                "files-bounds-large",
            ),
            runtime._api.upload(
                scenario.owner,
                scenario.project_id,
                _FILENAME,
                "text/html",
                1,
                "files-bounds-media",
            ),
        )
        return _statuses_are_safe(results, (400, 400, 400))
    except Exception:
        return False


def _uploaded_file_safe(
    runtime: object,
    flow: object,
    label: object,
) -> tuple[ProvisionedFiles, UploadContent, UploadTicket] | object:
    try:
        if type(runtime) is not FilesRuntime or type(flow) is not str or type(label) is not str:
            return _INVALID
        if flow not in {
            "lifecycle",
            "missing-object-retry",
            "single-transition",
            "authorization",
            "upload-validation",
        }:
            return _INVALID
        scenario = runtime.provision(
            FilesSeed(label=label, flow=cast(FilesFlow, flow))
        )
        content = UploadContent.from_bytes(_CONTENT_BYTES, _CONTENT_TYPE)
        ticket = runtime._api.upload(
            scenario.owner,
            scenario.project_id,
            _FILENAME,
            _CONTENT_TYPE,
            content.length,
            f"{label}-upload",
        )
        if type(ticket) is not UploadTicket:
            return _INVALID
        runtime.put(ticket, content)
        return scenario, content, ticket
    except Exception:
        return _INVALID


def _scenario_secrets_safe(
    scenario: object,
    ticket: object,
) -> tuple[str, ...] | object:
    try:
        if type(scenario) is not ProvisionedFiles or type(ticket) is not UploadTicket:
            return _INVALID
        details = _upload_ticket_details(ticket)
        if type(details) is not tuple:
            return _INVALID
        _binding, upload_url, object_key, _content_type, _length = details
        return (
            _FILENAME,
            scenario.project_id,
            scenario.nonce,
            scenario.owner.access_token,
            scenario.owner.refresh_token,
            scenario.foreign.access_token,
            scenario.foreign.refresh_token,
            upload_url,
            object_key,
        )
    except Exception:
        return _INVALID


def _statuses_are_safe(values: object, expected: object) -> bool:
    try:
        if type(values) is not tuple or type(expected) is not tuple:
            return False
        if len(values) != len(expected):
            return False
        return all(
            type(value) is FileStatus
            and type(status) is int
            and value.status_code == status
            for value, status in zip(values, expected, strict=True)
        )
    except Exception:
        return False


def _representations_hide_safe(
    runtime: object,
    scenario: object,
    ticket: object,
    forbidden: object,
) -> bool:
    try:
        if type(forbidden) is not tuple:
            return False
        rendered = f"{runtime!r}{runtime!s}{scenario!r}{scenario!s}{ticket!r}{ticket!s}"
        return all(value not in rendered for value in forbidden)
    except Exception:
        return False


def _generic_failure_is_safe(
    runtime: object,
    ticket: object,
    forbidden: object,
) -> bool:
    try:
        if (
            type(runtime) is not FilesRuntime
            or type(ticket) is not UploadTicket
            or type(forbidden) is not tuple
        ):
            return False
        details = _upload_ticket_details(ticket)
        if type(details) is not tuple:
            return False
        _binding, url, key, content_type, expected_length = details
        cross_runtime = UploadTicket._create(
            file_id=ticket.file_id,
            runtime_binding=object(),
            upload_url=url,
            object_key=key,
            content_type=content_type,
            expected_length=expected_length,
        )
        wrong_content = UploadContent.from_bytes(b"wrong", "text/plain")
        return _one_put_failure_is_generic_safe(
            runtime,
            cross_runtime,
            UploadContent.from_bytes(_CONTENT_BYTES, _CONTENT_TYPE),
            forbidden,
        ) and _one_put_failure_is_generic_safe(
            runtime,
            ticket,
            wrong_content,
            forbidden,
        )
    except Exception:
        return False


def _one_put_failure_is_generic_safe(
    runtime: FilesRuntime,
    ticket: UploadTicket,
    content: UploadContent,
    forbidden: tuple[str, ...],
) -> bool:
    try:
        runtime.put(ticket, content)
    except Exception as error:
        rendered = f"{error!r}{error!s}"
        return (
            type(error) is FilesRuntimeError
            and str(error) == "files runtime is invalid"
            and error.__cause__ is None
            and error.__suppress_context__ is True
            and all(value not in rendered for value in forbidden)
        )
    return False


def _raise_invalid() -> NoReturn:
    raise FilesRuntimeError("files runtime is invalid") from None


def _raise_assertion() -> NoReturn:
    raise FilesRuntimeError("files semantic contract failed") from None
