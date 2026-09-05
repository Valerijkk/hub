from __future__ import annotations

from pathlib import Path

import yaml


_REPO_ROOT = Path(__file__).resolve().parents[3]
_EXPECTED_VOLUMES = [
    {
        "type": "bind",
        "source": "./testframework/reports",
        "target": "/reports",
    },
    {
        "type": "bind",
        "source": "./backend/cmd",
        "target": "/opt/hub/source/backend/cmd",
        "read_only": True,
    },
    {
        "type": "bind",
        "source": "./backend/internal",
        "target": "/opt/hub/source/backend/internal",
        "read_only": True,
    },
    {
        "type": "bind",
        "source": "./backend/tests",
        "target": "/opt/hub/source/backend/tests",
        "read_only": True,
    },
    {
        "type": "bind",
        "source": "./backend/ent/schema",
        "target": "/opt/hub/source/backend/ent/schema",
        "read_only": True,
    },
    {
        "type": "bind",
        "source": "./backend/migrations",
        "target": "/opt/hub/source/backend/migrations",
        "read_only": True,
    },
    {
        "type": "bind",
        "source": "./backend/testframework",
        "target": "/opt/hub/source/backend/testframework",
        "read_only": True,
    },
    {
        "type": "bind",
        "source": "./testframework/docker/backend_unit",
        "target": "/usr/local/lib/hub/backend-unit",
        "read_only": True,
    },
    {
        "type": "bind",
        "source": "./testframework/docker/entrypoints/backend-unit.sh",
        "target": "/usr/local/bin/backend-unit",
        "read_only": True,
    },
]
_SOURCE_SHELL_ASSETS = (
    "testframework/docker/entrypoints/backend-unit.sh",
    "testframework/docker/backend_unit/runtime.sh",
    "testframework/docker/backend_unit/receipt.sh",
    "testframework/docker/backend_unit/plan.sh",
    "testframework/docker/backend_unit/execute.sh",
)


def test_backend_unit_mounts_reviewed_runner_from_source_without_rebuilding_image() -> None:
    compose = yaml.safe_load(
        (_REPO_ROOT / "docker-compose.local.yml").read_text(encoding="utf-8")
    )
    service = compose["services"]["backend-unit"]

    assert service["build"] == {
        "context": ".",
        "dockerfile": "testframework/docker/Dockerfile.qa",
        "target": "backend-unit",
    }
    assert service["volumes"] == _EXPECTED_VOLUMES
    assert service["entrypoint"] == ["/bin/sh", "/usr/local/bin/backend-unit"]
    assert service["command"] == []

    assert service["profiles"] == ["qa"]
    assert service["read_only"] is True
    assert service["security_opt"] == ["no-new-privileges:true"]
    assert service["cap_drop"] == ["ALL"]
    assert service["init"] is True
    assert service["pids_limit"] == 256
    assert service["cpus"] == 6.0
    assert service["mem_limit"] == "2g"
    assert service["memswap_limit"] == "2g"
    assert service["tmpfs"] == [
        "/tmp:rw,noexec,nosuid,nodev,size=64m",
        "/var/cache/hub-go:rw,noexec,nosuid,nodev,size=768m,uid=10001,gid=10001,mode=1770",
        "/var/tmp/hub-go:rw,nosuid,nodev,exec,size=384m,uid=10001,gid=10001,mode=1770",
    ]
    assert service["networks"] == ["qa-internal"]
    assert service["stop_grace_period"] == "30s"
    assert service["environment"] == {
        "GOCACHE": "/var/cache/hub-go",
        "GOENV": "off",
        "GOFLAGS": "-p=1",
        "GOPROXY": "off",
        "GOTOOLCHAIN": "local",
        "GOTMPDIR": "/var/tmp/hub-go",
        "GOWORK": "off",
    }
    assert "depends_on" not in service
    assert "ports" not in service

    for asset in _SOURCE_SHELL_ASSETS:
        assert b"\r\n" not in (_REPO_ROOT / asset).read_bytes()
