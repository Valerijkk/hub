from __future__ import annotations

import hashlib
import traceback
from dataclasses import FrozenInstanceError

import pytest

from hub_testkit.contracts import (
    ContractSourceBundle,
    ContractSourceError,
    ContractSourcePath,
    SourceBlob,
    SourceSnapshot,
)
from hub_testkit.contracts.pytest_source_bundle import (
    create_contract_source_bundle_from_config,
    provide_contract_source_bundle_factory,
    require_contract_source_bundle_factory,
)


COMMIT = "0123456789abcdef0123456789abcdef01234567"


def _contents(path: ContractSourcePath) -> bytes:
    return f"source for {path.value}\n".encode("utf-8")


def _blob(path: ContractSourcePath, contents: bytes | None = None) -> SourceBlob:
    raw = _contents(path) if contents is None else contents
    return SourceBlob(
        path=path,
        content=raw,
        sha256=hashlib.sha256(raw).hexdigest(),
    )


def _snapshot(
    blobs: tuple[SourceBlob, ...] | None = None,
    commit: str = COMMIT,
) -> SourceSnapshot:
    return SourceSnapshot(
        source_commit=commit,
        blobs=tuple(_blob(path) for path in ContractSourcePath) if blobs is None else blobs,
    )


def _bundle(
    blobs: tuple[SourceBlob, ...] | None = None,
    commit: str = COMMIT,
) -> ContractSourceBundle:
    return ContractSourceBundle(_snapshot(blobs, commit))


def _bundle_error(snapshot: object) -> ContractSourceError:
    try:
        ContractSourceBundle(snapshot)  # type: ignore[arg-type]
    except ContractSourceError as error:
        return error
    raise AssertionError("invalid snapshot was accepted")


def test_bundle_admits_only_literal_manifest_and_publishes_safe_selector() -> None:
    bundle = _bundle()

    assert tuple(path.value for path in ContractSourcePath) == (
        "backend/cmd/server/main.go",
        "backend/cmd/server/wiring_auth.go",
        "backend/cmd/server/wiring_messenger.go",
        "backend/docs/swagger.json",
        "backend/internal/apigateway/apikey.go",
        "backend/internal/auth/events.go",
        "backend/internal/auth/handler.go",
        "backend/internal/auth/middleware.go",
        "backend/internal/auth/service_session.go",
        "backend/internal/notification/consumer.go",
        "backend/internal/notification/events.go",
        "backend/internal/platform/middleware/middleware.go",
        "backend/internal/tms/handler.go",
        "testframework/policies/f1-security-risk-ledger.toml",
    )
    assert bundle.source_commit == COMMIT
    expected = hashlib.sha256()
    expected.update(b"hub.qa.contract-source-bundle/v3\0")
    expected.update(COMMIT.encode("ascii"))
    expected.update(b"\0")
    for path in ContractSourcePath:
        expected.update(path.value.encode("ascii"))
        expected.update(b"\0")
        expected.update(hashlib.sha256(_contents(path)).hexdigest().encode("ascii"))
        expected.update(b"\0")
    assert bundle.bundle_digest == expected.hexdigest()
    assert bundle.select(ContractSourcePath.AUTH_HANDLER) == _contents(
        ContractSourcePath.AUTH_HANDLER
    )
    with pytest.raises(ContractSourceError):
        bundle.select("backend/internal/auth/handler.go")  # type: ignore[arg-type]
    with pytest.raises(AttributeError):
        getattr(bundle, "blobs")


def test_bundle_rejects_manifest_digest_commit_and_size_deviations() -> None:
    canonical = tuple(_blob(path) for path in ContractSourcePath)
    bad_digest = object.__new__(SourceBlob)
    object.__setattr__(bad_digest, "path", ContractSourcePath.AUTH_HANDLER)
    object.__setattr__(bad_digest, "content", b"altered\n")
    object.__setattr__(bad_digest, "sha256", "0" * 64)
    bad_path = object.__new__(SourceBlob)
    object.__setattr__(bad_path, "path", "backend/escape.go")
    object.__setattr__(bad_path, "content", b"source\n")
    object.__setattr__(bad_path, "sha256", hashlib.sha256(b"source\n").hexdigest())
    oversized_content = b"x" * (4 * 1024 * 1024 + 1)
    oversized = object.__new__(SourceBlob)
    object.__setattr__(oversized, "path", ContractSourcePath.SERVER_MAIN)
    object.__setattr__(oversized, "content", oversized_content)
    object.__setattr__(oversized, "sha256", hashlib.sha256(oversized_content).hexdigest())
    del oversized_content
    total = tuple(_blob(path, b"x" * 900_000) for path in ContractSourcePath)

    invalid_snapshots = (
        _snapshot(canonical[:-1]),
        _snapshot((*canonical[:1], *canonical[2:])),
        _snapshot((*canonical[:4], *canonical[5:])),
        _snapshot((*canonical, _blob(ContractSourcePath.APIGATEWAY_APIKEY))),
        _snapshot((canonical[0], canonical[2], canonical[1], *canonical[3:])),
        _snapshot((*canonical[:3], canonical[4], canonical[3], *canonical[5:])),
        _snapshot((*canonical[:5], canonical[4], *canonical[6:])),
        _snapshot((*canonical[:6], bad_digest, *canonical[7:])),
        _snapshot((*canonical[:6], bad_path, *canonical[7:])),
        _snapshot((oversized, *canonical[1:])),
        _snapshot(total),
    )

    for snapshot in invalid_snapshots:
        with pytest.raises(ContractSourceError) as error:
            ContractSourceBundle(snapshot)
        assert error.value.__cause__ is None
        assert error.value.__suppress_context__ is True

    for commit in (
        "0123456789abcdef0123456789abcdef0123456G",
        "0123456789abcdef0123456789abcdef0123456",
    ):
        with pytest.raises(ContractSourceError):
            SourceSnapshot(source_commit=commit, blobs=canonical)


def test_blob_rejects_noncanonical_text_and_type_boundaries() -> None:
    path = ContractSourcePath.SERVER_MAIN
    invalid_contents = (
        b"\xef\xbb\xbftext\n",
        b"text\x00\n",
        b"text\r\n",
        "text\n",  # type: ignore[arg-type]
    )

    for contents in invalid_contents:
        with pytest.raises(ContractSourceError):
            SourceBlob(
                path=path,
                content=contents,
                sha256=hashlib.sha256(b"text\n").hexdigest(),
            )
    with pytest.raises(ContractSourceError):
        SourceBlob(path=path, content=b"text\n", sha256="A" * 64)
    with pytest.raises(ContractSourceError):
        SourceSnapshot(source_commit=True, blobs=())  # type: ignore[arg-type]


def test_bundle_is_immutable_and_repr_does_not_reveal_source_canary() -> None:
    canary = b"contract-source-repr-canary"
    blobs = tuple(
        _blob(path, canary if path is ContractSourcePath.AUTH_EVENTS else _contents(path))
        for path in ContractSourcePath
    )
    bundle = _bundle(blobs)

    with pytest.raises(FrozenInstanceError):
        bundle.source_commit = "f" * 40  # type: ignore[misc]
    assert canary.decode("ascii") not in repr(bundle)
    assert canary.decode("ascii") not in str(bundle)


class _HostileSnapshot:
    canary = "contract-source-hostile-canary"

    @property
    def source_commit(self) -> str:
        raise RuntimeError(self.canary)


def test_bundle_suppresses_hostile_snapshot_context_and_canary() -> None:
    error = _bundle_error(_HostileSnapshot())
    formatted = "".join(traceback.format_exception(error))

    assert str(error) == "contract source bundle is invalid"
    assert error.__cause__ is None
    assert error.__suppress_context__ is True
    assert _HostileSnapshot.canary not in formatted


class _Delegate:
    def __init__(self, snapshot: SourceSnapshot) -> None:
        self._snapshot = snapshot

    def materialize(self) -> SourceSnapshot:
        return self._snapshot


class _Factory:
    def __init__(self, snapshot: SourceSnapshot) -> None:
        self._snapshot = snapshot
        self.calls = 0

    def __call__(self) -> _Delegate:
        self.calls += 1
        return _Delegate(self._snapshot)


class _Config:
    def __init__(self) -> None:
        self.stash: dict[object, object] = {}


def test_factory_stash_is_singleton_and_materializes_once_per_bundle_request() -> None:
    config = _Config()
    factory = _Factory(_snapshot())

    provide_contract_source_bundle_factory(config, factory)  # type: ignore[arg-type]
    assert require_contract_source_bundle_factory(config) is factory  # type: ignore[arg-type]
    bundle = create_contract_source_bundle_from_config(config)  # type: ignore[arg-type]
    assert bundle.source_commit == COMMIT
    assert factory.calls == 1
    with pytest.raises(pytest.UsageError) as duplicate:
        provide_contract_source_bundle_factory(config, factory)  # type: ignore[arg-type]
    assert str(duplicate.value) == "contract source bundle factory is invalid"


class _HostileFactory:
    canary = "contract-source-factory-canary"

    def __call__(self) -> _Delegate:
        raise RuntimeError(self.canary)


def test_factory_suppresses_hostile_materialization_context_and_canary() -> None:
    config = _Config()
    provide_contract_source_bundle_factory(config, _HostileFactory())  # type: ignore[arg-type]

    with pytest.raises(pytest.UsageError) as captured:
        create_contract_source_bundle_from_config(config)  # type: ignore[arg-type]
    formatted = "".join(traceback.format_exception(captured.value))
    assert str(captured.value) == "contract source bundle factory is unavailable"
    assert captured.value.__cause__ is None
    assert captured.value.__suppress_context__ is True
    assert _HostileFactory.canary not in formatted
