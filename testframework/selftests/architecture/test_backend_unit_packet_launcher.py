from __future__ import annotations

from pathlib import Path

import yaml


_REPO_ROOT = Path(__file__).resolve().parents[3]
_ENTRYPOINT_PATH = "testframework/docker/entrypoints/python-qa.sh"
_ENTRYPOINT = ["/bin/sh", f"/opt/hub/{_ENTRYPOINT_PATH}"]
_SOURCE_MOUNTS = {
    "./testframework/observers": "/opt/hub/testframework/observers",
    "./testframework/schemas": "/opt/hub/testframework/schemas",
    "./testframework/policies": "/opt/hub/testframework/policies",
    "./testframework/docker": "/opt/hub/testframework/docker",
}


def test_python_contract_has_hardened_source_bound_packet_launcher() -> None:
    compose_path = _REPO_ROOT / "docker-compose.local.yml"
    compose = yaml.safe_load(compose_path.read_text(encoding="utf-8"))
    service = compose["services"]["python-contract"]

    assert service["command"] == ["contract"]
    assert service["entrypoint"] == _ENTRYPOINT
    assert service["mem_limit"] == "2g"
    assert service["memswap_limit"] == "2g"
    assert service["read_only"] is True
    assert service.get("privileged") is not True

    mounts = {
        volume["source"]: volume
        for volume in service["volumes"]
        if volume["source"] in _SOURCE_MOUNTS
    }
    assert set(mounts) == set(_SOURCE_MOUNTS)
    assert {
        source: volume["target"] for source, volume in mounts.items()
    } == _SOURCE_MOUNTS
    assert all(volume["read_only"] is True for volume in mounts.values())
    assert all(
        "docker.sock" not in str(volume["source"])
        and volume["target"] != "/var/run/docker.sock"
        for volume in service["volumes"]
    )

    entrypoint_bytes = (_REPO_ROOT / _ENTRYPOINT_PATH).read_bytes()
    assert b"export PYTHONPATH=/opt/hub/testframework/src:/opt/hub/testframework\n" in entrypoint_bytes
    assert b"\r\n" not in entrypoint_bytes
    assert b'[ "$#" -eq 1 ] || exit 64\n' in entrypoint_bytes
    assert (
        b"  backend-unit-receipt)\n"
        b"    exec /opt/hub/.venv/bin/python -m "
        b"observers.backend_unit.container_packet\n"
        b"    ;;\n"
    ) in entrypoint_bytes
