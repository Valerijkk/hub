from __future__ import annotations

import hashlib
import json
from pathlib import Path

import pytest

from observers.backend_unit import container_packet
from observers.backend_unit.container_packet import main, validate_container_packet
from observers.backend_unit.contracts import ReceiptContractError


_REPO_ROOT = Path(__file__).resolve().parents[4]
_CONTAINER_FIXTURE = (
    _REPO_ROOT
    / "testframework/selftests/architecture/observers/fixtures/backend_unit/valid/container-receipt.json"
)
_HOSTNAME = "a1b2c3d4e5f6"
_RUN_DIR = f"backend-unit-{_HOSTNAME}.partial"
_READ_BUDGET_BYTES = 64 * 1024 * 1024
_STREAM_CAPS = {
    "build": (2097152, 2097152),
    "vet": (2097152, 2097152),
    "unit": (16777216, 4194304),
    "race": (25165824, 4194304),
}
_COMPLETE_STAGES = ("build", "vet", "unit", "race")


def _canonical_json(document: dict[str, object]) -> bytes:
    return (
        json.dumps(
            document,
            allow_nan=False,
            ensure_ascii=False,
            separators=(",", ":"),
            sort_keys=True,
        ).encode("utf-8")
        + b"\n"
    )


def _complete_document() -> dict[str, object]:
    document = json.loads(_CONTAINER_FIXTURE.read_bytes())
    assert type(document) is dict
    return document


def _bind_inventory(document: dict[str, object], payloads: dict[str, bytes]) -> None:
    records = document["artifact_inventory"]
    assert type(records) is list
    records.sort(key=lambda record: str(record["path"]))
    lines: list[bytes] = []
    for record in records:
        path = record["path"]
        assert type(path) is str
        payload = payloads[path]
        record["sha256"] = "sha256:" + hashlib.sha256(payload).hexdigest()
        record["size"] = len(payload)
        lines.append(
            f"{path}\t{record['sha256']}\t{record['size']}\n".encode("ascii")
        )
    document["artifact_inventory_sha256"] = "sha256:" + hashlib.sha256(
        b"".join(lines)
    ).hexdigest()


def _complete_payloads() -> dict[str, bytes]:
    payloads = {
        "CONTAINER_COMPLETE": b"",
        "owner.json": b"{}",
    }
    for stage in _COMPLETE_STAGES:
        payloads.update(
            {
                f"package-plan.{stage}.json": f"plan:{stage}".encode("ascii"),
                f"package-results.{stage}.jsonl": f"results:{stage}".encode("ascii"),
                f"stage-{stage}.json": f"stage:{stage}".encode("ascii"),
                f"{stage}.stdout.log": f"stdout:{stage}".encode("ascii"),
                f"{stage}.stderr.log": f"stderr:{stage}".encode("ascii"),
            }
        )
    return payloads


def _bind_complete_inventory(document: dict[str, object], payloads: dict[str, bytes]) -> None:
    document["artifact_inventory"] = [
        {"path": path, "sha256": "sha256:" + "0" * 64, "size": 0}
        for path in sorted(payloads)
    ]
    _bind_inventory(document, payloads)


def _write_complete_packet(tmp_path: Path) -> tuple[Path, Path]:
    document = _complete_document()
    payloads = _complete_payloads()
    _bind_complete_inventory(document, payloads)
    packet = tmp_path / _RUN_DIR
    packet.mkdir()
    for relative, payload in payloads.items():
        (packet / relative).write_bytes(payload)
    (packet / "container-receipt.json").write_bytes(_canonical_json(document))
    return tmp_path, packet


def _aborted_document() -> tuple[dict[str, object], dict[str, bytes]]:
    complete = _complete_document()
    prefix = (
        "schema",
        "target",
        "run_id",
        "owner_nonce",
        "container_hostname",
        "launcher_sha256",
        "command_profile_sha256",
        "source_manifest_before_sha256",
        "source_manifest_after_sha256",
        "cgroup",
        "go_environment",
    )
    document = {key: complete[key] for key in prefix}
    planned = ("build", "vet", "unit", "race")
    completed = ("build",)
    document["abort"] = {
        "signal": "TERM",
        "phase": "execution",
        "active_stage": "vet",
        "planned_stages": list(planned),
        "completed_stages": list(completed),
        "child_group_reaped": True,
        "stdout_reader_reaped": True,
        "stderr_reader_reaped": True,
        "unpublished_stage_rolled_back": True,
    }
    paths = {"CONTAINER_ABORTED", "owner.json"}
    paths.update(f"package-plan.{stage}.json" for stage in planned)
    for stage in completed:
        paths.update(
            {
                f"{stage}.stderr.log",
                f"{stage}.stdout.log",
                f"package-results.{stage}.jsonl",
                f"stage-{stage}.json",
            }
        )
    payloads = {
        path: b"" if path == "CONTAINER_ABORTED" else path.encode("utf-8")
        for path in paths
    }
    document["artifact_inventory"] = [
        {"path": path, "sha256": "sha256:" + "0" * 64, "size": 0}
        for path in sorted(paths)
    ]
    _bind_inventory(document, payloads)
    document["terminal"] = {"signal": "TERM", "exit_code": 143}
    document["state"] = "CONTAINER_ABORTED"
    document["report_budget"] = {
        "total_bytes": 67108864,
        "reserved_metadata_bytes": 8388608,
        "metadata_used_bytes": 0,
        "streams": {
            stage: {
                stream: {
                    "cap_bytes": cap,
                    "used_bytes": 0,
                    "truncated": False,
                }
                for stream, cap in zip(("stdout", "stderr"), caps, strict=True)
            }
            for stage, caps in _STREAM_CAPS.items()
        },
        "total_used_bytes": 0,
        "truncated_due_to_budget": False,
        "raw_output_partial": False,
    }
    budget = document["report_budget"]
    assert type(budget) is dict
    inventory = document["artifact_inventory"]
    assert type(inventory) is list
    sizes = {record["path"]: record["size"] for record in inventory}
    assert all(type(path) is str and type(size) is int for path, size in sizes.items())
    streams = budget["streams"]
    assert type(streams) is dict
    for stream in ("stdout", "stderr"):
        stream_usage = streams["build"]
        assert type(stream_usage) is dict
        usage = stream_usage[stream]
        assert type(usage) is dict
        usage["used_bytes"] = sizes[f"build.{stream}.log"]
    metadata_inventory = sum(
        size
        for path, size in sizes.items()
        if not (path.endswith(".stdout.log") or path.endswith(".stderr.log"))
    )
    inventory_total = sum(sizes.values())
    for _ in range(16):
        receipt_size = len(_canonical_json(document))
        metadata_used = metadata_inventory + receipt_size
        total_used = inventory_total + receipt_size
        if (
            budget["metadata_used_bytes"] == metadata_used
            and budget["total_used_bytes"] == total_used
        ):
            return document, payloads
        budget["metadata_used_bytes"] = metadata_used
        budget["total_used_bytes"] = total_used
    raise AssertionError("aborted packet budget fixed point did not converge")


def _write_aborted_packet(tmp_path: Path) -> tuple[Path, Path]:
    document, payloads = _aborted_document()
    packet = tmp_path / _RUN_DIR
    packet.mkdir()
    for relative, payload in payloads.items():
        (packet / relative).write_bytes(payload)
    (packet / "container-receipt.json").write_bytes(_canonical_json(document))
    return tmp_path, packet


def test_accepts_complete_packet_with_actual_inventory_bytes(tmp_path: Path) -> None:
    reports_root, _packet = _write_complete_packet(tmp_path)

    assert validate_container_packet(reports_root, _RUN_DIR) is None


def test_rejects_truncated_self_described_complete_inventory_and_tree(tmp_path: Path) -> None:
    reports_root = tmp_path
    packet = reports_root / _RUN_DIR
    packet.mkdir()
    document = _complete_document()
    payloads = {
        path: payload
        for path, payload in _complete_payloads().items()
        if "race" not in path
    }
    _bind_complete_inventory(document, payloads)
    for relative, payload in payloads.items():
        (packet / relative).write_bytes(payload)
    (packet / "container-receipt.json").write_bytes(_canonical_json(document))

    with pytest.raises(ReceiptContractError):
        validate_container_packet(reports_root, _RUN_DIR)


def test_accepts_packet_at_exact_unique_byte_read_budget(
    tmp_path: Path, monkeypatch: pytest.MonkeyPatch
) -> None:
    reports_root, packet = _write_complete_packet(tmp_path)
    unique_packet_bytes = sum(path.stat().st_size for path in packet.iterdir())
    monkeypatch.setattr(container_packet, "_READ_BUDGET_BYTES", unique_packet_bytes)

    assert validate_container_packet(reports_root, _RUN_DIR) is None


def test_accepts_aborted_packet_without_translating_it_to_test_success(tmp_path: Path) -> None:
    reports_root, _packet = _write_aborted_packet(tmp_path)

    assert validate_container_packet(reports_root, _RUN_DIR) is None


def test_cli_rejects_unsupported_arguments() -> None:
    assert main(("unexpected",)) == 1


@pytest.mark.parametrize("mutation", ("tamper", "missing", "extra"))
def test_rejects_tampered_missing_or_extra_actual_artifact(
    tmp_path: Path, mutation: str
) -> None:
    reports_root, packet = _write_complete_packet(tmp_path)
    if mutation == "tamper":
        (packet / "stage-build.json").write_bytes(b"tampered")
    elif mutation == "missing":
        (packet / "owner.json").unlink()
    else:
        (packet / "unexpected.txt").write_bytes(b"unexpected")

    with pytest.raises(ReceiptContractError):
        validate_container_packet(reports_root, _RUN_DIR)


@pytest.mark.parametrize(
    ("mutation", "run_dir"),
    (
        ("nonempty_marker", _RUN_DIR),
        ("wrong_hostname", _RUN_DIR),
        ("duplicate_json", _RUN_DIR),
        ("traversal", "../backend-unit-a1b2c3d4e5f6.partial"),
        ("wrong_run_dir", "backend-unit-not-a-container.partial"),
    ),
)
def test_rejects_unusable_marker_identity_or_json_packet(
    tmp_path: Path, mutation: str, run_dir: str
) -> None:
    reports_root, packet = _write_complete_packet(tmp_path)
    receipt_path = packet / "container-receipt.json"
    if mutation == "nonempty_marker":
        (packet / "CONTAINER_COMPLETE").write_bytes(b"\n")
    elif mutation == "wrong_hostname":
        document = _complete_document()
        document["container_hostname"] = "f" * 12
        _bind_complete_inventory(document, _complete_payloads())
        receipt_path.write_bytes(_canonical_json(document))
    elif mutation == "duplicate_json":
        payload = receipt_path.read_bytes()
        receipt_path.write_bytes(
            payload.replace(
                b'"target":"backend-unit",',
                b'"target":"backend-unit","target":"backend-unit",',
                1,
            )
        )

    with pytest.raises(ReceiptContractError):
        validate_container_packet(reports_root, run_dir)


def test_rejects_symlinked_artifact_and_receipt_larger_than_read_budget(tmp_path: Path) -> None:
    reports_root, packet = _write_complete_packet(tmp_path)
    owner = packet / "owner.json"
    owner.unlink()
    owner.symlink_to(packet / "stage-build.json")

    with pytest.raises(ReceiptContractError):
        validate_container_packet(reports_root, _RUN_DIR)

    owner.unlink()
    owner.write_bytes(b"{}")
    receipt = packet / "container-receipt.json"
    with receipt.open("wb") as handle:
        handle.seek(_READ_BUDGET_BYTES)
        handle.write(b"x")

    with pytest.raises(ReceiptContractError):
        validate_container_packet(reports_root, _RUN_DIR)
