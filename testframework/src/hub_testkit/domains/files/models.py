from __future__ import annotations

import hashlib
import re
from dataclasses import dataclass, field
from typing import Literal, NoReturn
from uuid import UUID


class FilesModelError(ValueError):
    pass


FilesFlow = Literal[
    "lifecycle",
    "missing-object-retry",
    "single-transition",
    "authorization",
    "upload-validation",
]

_FLOWS = frozenset(
    {
        "lifecycle",
        "missing-object-retry",
        "single-transition",
        "authorization",
        "upload-validation",
    }
)
_LABEL_RE = re.compile(r"^[a-z][a-z0-9-]{2,31}$")
_NONCE_RE = re.compile(r"^[0-9a-f]{32}$")
_SHA256_RE = re.compile(r"^[0-9a-f]{64}$")
_LOCAL_S3_PREFIX = "http://seaweedfs:8333/"
_INVALID = object()


@dataclass(frozen=True, slots=True, repr=False)
class FilesSeed:
    label: str
    flow: FilesFlow

    def __post_init__(self) -> None:
        if _seed_is_valid_safe(self.label, self.flow):
            return
        _raise_invalid()


@dataclass(frozen=True, slots=True, repr=False)
class FilesPrincipal:
    user_id: str = field(repr=False)
    session_id: str = field(repr=False)
    access_token: str = field(repr=False)
    refresh_token: str = field(repr=False)

    def __post_init__(self) -> None:
        if _principal_is_valid_safe(
            self.user_id,
            self.session_id,
            self.access_token,
            self.refresh_token,
        ):
            return
        _raise_invalid()


@dataclass(frozen=True, slots=True, repr=False, eq=False)
class ProvisionedFiles:
    owner: FilesPrincipal = field(repr=False)
    foreign: FilesPrincipal = field(repr=False)
    project_id: str = field(repr=False)
    nonce: str = field(repr=False)

    def __post_init__(self) -> None:
        if _provisioned_is_valid_safe(
            self.owner,
            self.foreign,
            self.project_id,
            self.nonce,
        ):
            return
        _raise_invalid()


@dataclass(frozen=True, slots=True)
class FileStatus:
    status_code: int

    def __post_init__(self) -> None:
        if type(self.status_code) is int and self.status_code in {
            200,
            400,
            401,
            403,
            404,
            409,
        }:
            return
        _raise_invalid()


@dataclass(frozen=True, slots=True, repr=False, eq=False, init=False)
class UploadTicket:
    file_id: str
    _runtime_binding: object = field(repr=False)
    _upload_url: str = field(repr=False)
    _object_key: str = field(repr=False)
    _content_type: str = field(repr=False)
    _expected_length: int = field(repr=False)

    def __init__(self) -> None:
        _raise_invalid()

    @classmethod
    def _create(
        cls,
        *,
        file_id: str,
        runtime_binding: object,
        upload_url: str,
        object_key: str,
        content_type: str,
        expected_length: int,
    ) -> UploadTicket:
        result = _upload_ticket_safe(
            cls,
            file_id,
            runtime_binding,
            upload_url,
            object_key,
            content_type,
            expected_length,
        )
        del file_id, runtime_binding, upload_url, object_key, content_type, expected_length
        if type(result) is cls:
            return result
        del result
        _raise_invalid()

    def __repr__(self) -> str:
        return "UploadTicket()"


@dataclass(frozen=True, slots=True, repr=False, eq=False, init=False)
class DownloadTicket:
    file_id: str
    _runtime_binding: object = field(repr=False)
    _download_url: str = field(repr=False)

    def __init__(self) -> None:
        _raise_invalid()

    @classmethod
    def _create(
        cls,
        *,
        file_id: str,
        runtime_binding: object,
        download_url: str,
    ) -> DownloadTicket:
        result = _download_ticket_safe(cls, file_id, runtime_binding, download_url)
        del file_id, runtime_binding, download_url
        if type(result) is cls:
            return result
        del result
        _raise_invalid()

    def __repr__(self) -> str:
        return "DownloadTicket()"


@dataclass(frozen=True, slots=True, repr=False, eq=False, init=False)
class UploadContent:
    length: int
    sha256: str
    _body: bytes = field(repr=False)
    _content_type: str = field(repr=False)

    def __init__(self) -> None:
        _raise_invalid()

    @classmethod
    def from_bytes(cls, body: bytes, content_type: str) -> UploadContent:
        result = _upload_content_safe(cls, body, content_type)
        del body, content_type
        if type(result) is cls:
            return result
        del result
        _raise_invalid()

    def __repr__(self) -> str:
        return "UploadContent()"


@dataclass(frozen=True, slots=True)
class ObjectDigest:
    length: int
    sha256: str

    def __post_init__(self) -> None:
        if _digest_is_valid_safe(self.length, self.sha256):
            return
        _raise_invalid()


def _upload_ticket_safe(
    cls: type[UploadTicket],
    file_id: object,
    runtime_binding: object,
    upload_url: object,
    object_key: object,
    content_type: object,
    expected_length: object,
) -> UploadTicket | object:
    try:
        if not _canonical_uuid_safe(file_id) or runtime_binding is None:
            return _INVALID
        if not _local_s3_url_safe(upload_url) or not _text_safe(object_key, 1024):
            return _INVALID
        if not _content_type_safe(content_type):
            return _INVALID
        if type(expected_length) is not int or not 1 <= expected_length <= 25 * 1024 * 1024:
            return _INVALID
        result = object.__new__(cls)
        object.__setattr__(result, "file_id", file_id)
        object.__setattr__(result, "_runtime_binding", runtime_binding)
        object.__setattr__(result, "_upload_url", upload_url)
        object.__setattr__(result, "_object_key", object_key)
        object.__setattr__(result, "_content_type", content_type)
        object.__setattr__(result, "_expected_length", expected_length)
        return result
    except Exception:
        return _INVALID


def _download_ticket_safe(
    cls: type[DownloadTicket],
    file_id: object,
    runtime_binding: object,
    download_url: object,
) -> DownloadTicket | object:
    try:
        if not _canonical_uuid_safe(file_id) or runtime_binding is None:
            return _INVALID
        if not _local_s3_url_safe(download_url):
            return _INVALID
        result = object.__new__(cls)
        object.__setattr__(result, "file_id", file_id)
        object.__setattr__(result, "_runtime_binding", runtime_binding)
        object.__setattr__(result, "_download_url", download_url)
        return result
    except Exception:
        return _INVALID


def _upload_content_safe(
    cls: type[UploadContent],
    body: object,
    content_type: object,
) -> UploadContent | object:
    try:
        if (
            type(body) is not bytes
            or not 1 <= len(body) <= 25 * 1024 * 1024
            or not _content_type_safe(content_type)
        ):
            return _INVALID
        result = object.__new__(cls)
        object.__setattr__(result, "length", len(body))
        object.__setattr__(result, "sha256", hashlib.sha256(body).hexdigest())
        object.__setattr__(result, "_body", body)
        object.__setattr__(result, "_content_type", content_type)
        return result
    except Exception:
        return _INVALID


def _upload_ticket_details(
    ticket: object,
) -> tuple[object, str, str, str, int] | object:
    try:
        if type(ticket) is not UploadTicket:
            return _INVALID
        if not _canonical_uuid_safe(ticket.file_id):
            return _INVALID
        if not _local_s3_url_safe(ticket._upload_url):
            return _INVALID
        if not _text_safe(ticket._object_key, 1024):
            return _INVALID
        if not _content_type_safe(ticket._content_type):
            return _INVALID
        if type(ticket._expected_length) is not int:
            return _INVALID
        return (
            ticket._runtime_binding,
            ticket._upload_url,
            ticket._object_key,
            ticket._content_type,
            ticket._expected_length,
        )
    except Exception:
        return _INVALID


def _download_ticket_details(ticket: object) -> tuple[object, str] | object:
    try:
        if type(ticket) is not DownloadTicket:
            return _INVALID
        if not _canonical_uuid_safe(ticket.file_id):
            return _INVALID
        if not _local_s3_url_safe(ticket._download_url):
            return _INVALID
        return ticket._runtime_binding, ticket._download_url
    except Exception:
        return _INVALID


def _upload_content_details(content: object) -> tuple[bytes, str] | object:
    try:
        if type(content) is not UploadContent:
            return _INVALID
        if not _digest_is_valid_safe(content.length, content.sha256):
            return _INVALID
        if type(content._body) is not bytes or len(content._body) != content.length:
            return _INVALID
        if not _content_type_safe(content._content_type):
            return _INVALID
        if hashlib.sha256(content._body).hexdigest() != content.sha256:
            return _INVALID
        return content._body, content._content_type
    except Exception:
        return _INVALID


def _seed_is_valid_safe(label: object, flow: object) -> bool:
    try:
        return (
            type(label) is str
            and _LABEL_RE.fullmatch(label) is not None
            and type(flow) is str
            and flow in _FLOWS
        )
    except Exception:
        return False


def _principal_is_valid_safe(
    user_id: object,
    session_id: object,
    access_token: object,
    refresh_token: object,
) -> bool:
    try:
        return (
            _canonical_uuid_safe(user_id)
            and _canonical_uuid_safe(session_id)
            and _text_safe(access_token, 4096)
            and _text_safe(refresh_token, 4096)
        )
    except Exception:
        return False


def _provisioned_is_valid_safe(
    owner: object,
    foreign: object,
    project_id: object,
    nonce: object,
) -> bool:
    try:
        return (
            type(owner) is FilesPrincipal
            and type(foreign) is FilesPrincipal
            and _principal_is_valid_safe(
                owner.user_id,
                owner.session_id,
                owner.access_token,
                owner.refresh_token,
            )
            and _principal_is_valid_safe(
                foreign.user_id,
                foreign.session_id,
                foreign.access_token,
                foreign.refresh_token,
            )
            and owner.user_id != foreign.user_id
            and owner.session_id != foreign.session_id
            and _canonical_uuid_safe(project_id)
            and type(nonce) is str
            and _NONCE_RE.fullmatch(nonce) is not None
        )
    except Exception:
        return False


def _canonical_uuid_safe(value: object) -> bool:
    try:
        return type(value) is str and str(UUID(value)) == value
    except Exception:
        return False


def _text_safe(value: object, maximum: int) -> bool:
    try:
        return (
            type(value) is str
            and bool(value.strip())
            and len(value) <= maximum
            and all(32 <= ord(character) != 127 for character in value)
        )
    except Exception:
        return False


def _content_type_safe(value: object) -> bool:
    return (
        _text_safe(value, 128)
        and type(value) is str
        and value == value.strip().lower()
        and ";" not in value
    )


def _local_s3_url_safe(value: object) -> bool:
    return (
        _text_safe(value, 8192)
        and type(value) is str
        and value.startswith(_LOCAL_S3_PREFIX)
    )


def _digest_is_valid_safe(length: object, sha256: object) -> bool:
    try:
        return (
            type(length) is int
            and length >= 0
            and type(sha256) is str
            and _SHA256_RE.fullmatch(sha256) is not None
        )
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise FilesModelError("files model is invalid") from None
