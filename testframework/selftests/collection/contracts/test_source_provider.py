from __future__ import annotations

import hashlib
import json
import traceback
from pathlib import Path

import pytest

from hub_testkit.contracts import ContractSourcePath
from hub_testkit.contracts.pytest_source_bundle import (
    create_contract_source_bundle_from_config,
)
from hub_testkit.pytest_plugin import source_provider


COMMIT = "4e9a0dad4f07a5b6123b4cf4e17e380e81c931fb"


class _Config:
    def __init__(self) -> None:
        self.stash: dict[object, object] = {}


def _content(path: ContractSourcePath) -> bytes:
    return f"sealed source for {path.value}\n".encode()


def _write_source_root(root: Path) -> dict[str, bytes]:
    contents: dict[str, bytes] = {}
    for path in ContractSourcePath:
        content = _content(path)
        target = root / path.value
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_bytes(content)
        contents[path.value] = content
    return contents


def _manifest_document(contents: dict[str, bytes]) -> dict[str, object]:
    return {
        "schema": "hub.testframework.contract-source-manifest/v1",
        "byte_format": "git-lf-utf8",
        "source_commit": COMMIT,
        "blobs": [
            {
                "path": path.value,
                "sha256": hashlib.sha256(contents[path.value]).hexdigest(),
            }
            for path in ContractSourcePath
        ],
    }


def _write_manifest(path: Path, document: dict[str, object]) -> None:
    path.write_text(
        json.dumps(document, ensure_ascii=True, separators=(",", ":")) + "\n",
        encoding="utf-8",
        newline="\n",
    )


def _sealed_fixture(tmp_path: Path) -> tuple[Path, Path, dict[str, bytes]]:
    root = tmp_path / "contract-source"
    root.mkdir(parents=True)
    contents = _write_source_root(root)
    manifest = tmp_path / "manifest.json"
    _write_manifest(manifest, _manifest_document(contents))
    return root, manifest, contents


def test_provider_materializes_exact_ordered_snapshot_and_digest(tmp_path: Path) -> None:
    root, manifest, contents = _sealed_fixture(tmp_path)

    snapshot = source_provider._materialize_snapshot(root, manifest)

    assert snapshot.source_commit == COMMIT
    assert tuple(blob.path for blob in snapshot.blobs) == tuple(ContractSourcePath)
    assert tuple(blob.content for blob in snapshot.blobs) == tuple(
        contents[path.value] for path in ContractSourcePath
    )
    assert tuple(blob.sha256 for blob in snapshot.blobs) == tuple(
        hashlib.sha256(contents[path.value]).hexdigest()
        for path in ContractSourcePath
    )


def test_provider_registers_one_factory_and_materializes_per_request(
    tmp_path: Path,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    root, manifest, _ = _sealed_fixture(tmp_path)
    monkeypatch.setattr(source_provider, "_SOURCE_ROOT", root)
    monkeypatch.setattr(source_provider, "_MANIFEST_PATH", manifest)
    config = _Config()

    source_provider.pytest_configure(config)  # type: ignore[arg-type]
    first = create_contract_source_bundle_from_config(config)  # type: ignore[arg-type]
    second = create_contract_source_bundle_from_config(config)  # type: ignore[arg-type]

    assert first.source_commit == COMMIT
    assert first.bundle_digest == second.bundle_digest
    with pytest.raises(pytest.UsageError):
        source_provider.pytest_configure(config)  # type: ignore[arg-type]


def test_sealed_image_snapshot_matches_real_manifest_and_all_sources() -> None:
    snapshot = source_provider._materialize_snapshot(
        source_provider._SOURCE_ROOT,
        source_provider._MANIFEST_PATH,
    )

    assert snapshot.source_commit == COMMIT
    assert tuple(blob.path for blob in snapshot.blobs) == tuple(ContractSourcePath)
    assert all(hashlib.sha256(blob.content).hexdigest() == blob.sha256 for blob in snapshot.blobs)


def test_provider_rejects_manifest_shape_order_commit_and_digest(
    tmp_path: Path,
) -> None:
    root, manifest, contents = _sealed_fixture(tmp_path)
    canonical = _manifest_document(contents)
    invalid_documents = (
        {**canonical, "unknown": True},
        {**canonical, "byte_format": "worktree-native"},
        {**canonical, "source_commit": "F" * 40},
        {**canonical, "blobs": canonical["blobs"][1:]},  # type: ignore[index]
        {
            **canonical,
            "blobs": [
                canonical["blobs"][1],  # type: ignore[index]
                canonical["blobs"][0],  # type: ignore[index]
                *canonical["blobs"][2:],  # type: ignore[index]
            ],
        },
        {
            **canonical,
            "blobs": [
                {**canonical["blobs"][0], "sha256": "0" * 64},  # type: ignore[index]
                *canonical["blobs"][1:],  # type: ignore[index]
            ],
        },
    )

    for document in invalid_documents:
        _write_manifest(manifest, document)
        with pytest.raises(source_provider.ContractSourceProviderError):
            source_provider._materialize_snapshot(root, manifest)

    manifest.write_text(
        '{"schema":"hub.testframework.contract-source-manifest/v1",'
        '"schema":"duplicate","byte_format":"git-lf-utf8",'
        '"source_commit":"' + COMMIT + '","blobs":[]}\n',
        encoding="utf-8",
        newline="\n",
    )
    with pytest.raises(source_provider.ContractSourceProviderError):
        source_provider._materialize_snapshot(root, manifest)


def test_provider_rejects_missing_extra_and_nonregular_sources(tmp_path: Path) -> None:
    root, manifest, _ = _sealed_fixture(tmp_path)
    missing = root / ContractSourcePath.AUTH_EVENTS.value
    missing.unlink()
    with pytest.raises(source_provider.ContractSourceProviderError):
        source_provider._materialize_snapshot(root, manifest)

    root, manifest, _ = _sealed_fixture(tmp_path / "extra")
    (root / "extra.go").write_text("extra\n", encoding="utf-8", newline="\n")
    with pytest.raises(source_provider.ContractSourceProviderError):
        source_provider._materialize_snapshot(root, manifest)

    root, manifest, _ = _sealed_fixture(tmp_path / "directory")
    target = root / ContractSourcePath.AUTH_EVENTS.value
    target.unlink()
    target.mkdir()
    with pytest.raises(source_provider.ContractSourceProviderError):
        source_provider._materialize_snapshot(root, manifest)


def test_provider_rejects_symlink_and_noncanonical_source_bytes(tmp_path: Path) -> None:
    root, manifest, contents = _sealed_fixture(tmp_path)
    target = root / ContractSourcePath.AUTH_EVENTS.value
    target.unlink()
    target.symlink_to(root / ContractSourcePath.AUTH_HANDLER.value)
    with pytest.raises(source_provider.ContractSourceProviderError):
        source_provider._materialize_snapshot(root, manifest)

    canonical = _content(ContractSourcePath.AUTH_EVENTS)
    invalid_contents = (
        canonical + b"\r\n",
        canonical.replace(b"source", b"source\r", 1),
        canonical + b"\x00",
        b"\xef\xbb\xbf" + canonical,
    )
    for changed in invalid_contents:
        root, manifest, contents = _sealed_fixture(
            tmp_path / hashlib.sha256(changed).hexdigest()
        )
        target = root / ContractSourcePath.AUTH_EVENTS.value
        target.write_bytes(changed)
        document = _manifest_document({**contents, ContractSourcePath.AUTH_EVENTS.value: changed})
        _write_manifest(manifest, document)
        with pytest.raises(source_provider.ContractSourceProviderError):
            source_provider._materialize_snapshot(root, manifest)


def test_provider_rejects_oversized_source_before_bundle_creation(tmp_path: Path) -> None:
    root, manifest, contents = _sealed_fixture(tmp_path)
    target = root / ContractSourcePath.SERVER_MAIN.value
    changed = b"x" * (4 * 1024 * 1024 + 1)
    target.write_bytes(changed)
    document = _manifest_document({**contents, ContractSourcePath.SERVER_MAIN.value: changed})
    _write_manifest(manifest, document)

    with pytest.raises(source_provider.ContractSourceProviderError):
        source_provider._materialize_snapshot(root, manifest)


def test_provider_suppresses_source_and_manifest_canaries(tmp_path: Path) -> None:
    root, manifest, _ = _sealed_fixture(tmp_path)
    canary = "provider-secret-canary"
    manifest.write_text(canary, encoding="utf-8", newline="\n")

    try:
        source_provider._materialize_snapshot(root, manifest)
    except source_provider.ContractSourceProviderError as error:
        formatted = "".join(traceback.format_exception(error))
        assert str(error) == "contract source snapshot is invalid"
        assert error.__cause__ is None
        assert error.__suppress_context__ is True
        assert canary not in formatted
    else:
        raise AssertionError("hostile manifest was accepted")
