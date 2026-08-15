from __future__ import annotations

import hashlib
import re
from dataclasses import dataclass, field
from enum import Enum
from types import MappingProxyType
from typing import Mapping, NoReturn


_SOURCE_COMMIT_RE = re.compile(r"^[0-9a-f]{40}$")
_SHA256_RE = re.compile(r"^[0-9a-f]{64}$")
_MAX_BLOB_BYTES = 4 * 1024 * 1024
_MAX_TOTAL_BYTES = 8 * 1024 * 1024
_BUNDLE_PREFIX = b"hub.qa.contract-source-bundle/v1\0"
_INVALID = object()


class ContractSourceError(ValueError):
    pass


class ContractSourcePath(str, Enum):
    SERVER_MAIN = "backend/cmd/server/main.go"
    SWAGGER_JSON = "backend/docs/swagger.json"
    AUTH_EVENTS = "backend/internal/auth/events.go"
    AUTH_HANDLER = "backend/internal/auth/handler.go"
    AUTH_SERVICE_SESSION = "backend/internal/auth/service_session.go"
    NOTIFICATION_CONSUMER = "backend/internal/notification/consumer.go"
    NOTIFICATION_EVENTS = "backend/internal/notification/events.go"
    PLATFORM_MIDDLEWARE = "backend/internal/platform/middleware/middleware.go"
    TMS_HANDLER = "backend/internal/tms/handler.go"
    F1_SECURITY_RISK_LEDGER = "testframework/policies/f1-security-risk-ledger.toml"


SOURCE_MANIFEST: tuple[ContractSourcePath, ...] = tuple(ContractSourcePath)


@dataclass(frozen=True, slots=True)
class SourceBlob:
    path: ContractSourcePath
    content: bytes = field(repr=False)
    sha256: str

    def __post_init__(self) -> None:
        valid = _source_blob_valid(self.path, self.content, self.sha256)
        if valid:
            return
        object.__setattr__(self, "path", ContractSourcePath.SERVER_MAIN)
        object.__setattr__(self, "content", b"")
        object.__setattr__(self, "sha256", "0" * 64)
        _raise_invalid()


@dataclass(frozen=True, slots=True)
class SourceSnapshot:
    source_commit: str
    blobs: tuple[SourceBlob, ...] = field(repr=False)

    def __post_init__(self) -> None:
        valid = _snapshot_shape_valid(self.source_commit, self.blobs)
        if valid:
            return
        object.__setattr__(self, "source_commit", "0" * 40)
        object.__setattr__(self, "blobs", ())
        _raise_invalid()


@dataclass(frozen=True, slots=True, repr=False, init=False)
class ContractSourceBundle:
    _source_commit: str = field(repr=False)
    _bundle_digest: str = field(repr=False)
    _content_by_path: Mapping[ContractSourcePath, bytes] = field(repr=False)

    def __init__(self, snapshot: SourceSnapshot) -> None:
        state = _materialize_bundle_safe(snapshot)
        del snapshot
        if type(state) is _BundleState:
            object.__setattr__(self, "_source_commit", state.source_commit)
            object.__setattr__(self, "_bundle_digest", state.bundle_digest)
            object.__setattr__(self, "_content_by_path", state.content_by_path)
            return
        del state
        _raise_invalid()

    @property
    def source_commit(self) -> str:
        return self._source_commit

    @property
    def bundle_digest(self) -> str:
        return self._bundle_digest

    def select(self, selector: ContractSourcePath) -> bytes:
        content = _select_safe(self._content_by_path, selector)
        if type(content) is bytes:
            return content
        del selector, content
        _raise_invalid()

    def __repr__(self) -> str:
        return "ContractSourceBundle()"

    def __str__(self) -> str:
        return "ContractSourceBundle()"


@dataclass(frozen=True, slots=True)
class _BundleState:
    source_commit: str
    bundle_digest: str
    content_by_path: Mapping[ContractSourcePath, bytes] = field(repr=False)


def _source_blob_valid(path: object, content: object, sha256: object) -> bool:
    try:
        if type(path) is not ContractSourcePath or type(content) is not bytes:
            return False
        if type(sha256) is not str or _SHA256_RE.fullmatch(sha256) is None:
            return False
        if len(content) > _MAX_BLOB_BYTES or not _canonical_utf8_lf(content):
            return False
        return hashlib.sha256(content).hexdigest() == sha256
    except Exception:
        return False


def _snapshot_shape_valid(source_commit: object, blobs: object) -> bool:
    try:
        if type(source_commit) is not str or _SOURCE_COMMIT_RE.fullmatch(source_commit) is None:
            return False
        return type(blobs) is tuple and all(type(blob) is SourceBlob for blob in blobs)
    except Exception:
        return False


def _canonical_utf8_lf(content: bytes) -> bool:
    try:
        if content.startswith(b"\xef\xbb\xbf") or b"\x00" in content or b"\r" in content:
            return False
        decoded = content.decode("utf-8", errors="strict")
        del decoded
        return True
    except Exception:
        return False


def _materialize_bundle_safe(snapshot: object) -> _BundleState | object:
    try:
        if type(snapshot) is not SourceSnapshot:
            return _INVALID
        source_commit = snapshot.source_commit
        blobs = snapshot.blobs
        if not _snapshot_shape_valid(source_commit, blobs):
            return _INVALID
        if tuple(blob.path for blob in blobs) != SOURCE_MANIFEST:
            return _INVALID
        if sum(len(blob.content) for blob in blobs) > _MAX_TOTAL_BYTES:
            return _INVALID
        if not all(_source_blob_valid(blob.path, blob.content, blob.sha256) for blob in blobs):
            return _INVALID
        contents = MappingProxyType({blob.path: blob.content for blob in blobs})
        digest = _bundle_digest(source_commit, blobs)
        return _BundleState(source_commit, digest, contents)
    except Exception:
        return _INVALID


def _bundle_digest(source_commit: str, blobs: tuple[SourceBlob, ...]) -> str:
    digest = hashlib.sha256()
    digest.update(_BUNDLE_PREFIX)
    digest.update(source_commit.encode("ascii"))
    digest.update(b"\0")
    for blob in blobs:
        digest.update(blob.path.value.encode("ascii"))
        digest.update(b"\0")
        digest.update(blob.sha256.encode("ascii"))
        digest.update(b"\0")
    return digest.hexdigest()


def _select_safe(
    content_by_path: Mapping[ContractSourcePath, bytes],
    selector: object,
) -> bytes | object:
    try:
        if type(selector) is not ContractSourcePath:
            return _INVALID
        content = content_by_path.get(selector, _INVALID)
        if type(content) is not bytes:
            return _INVALID
        return content
    except Exception:
        return _INVALID


def _raise_invalid() -> NoReturn:
    raise ContractSourceError("contract source bundle is invalid") from None
