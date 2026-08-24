from __future__ import annotations

import json
import math
import os
import stat
import traceback
from collections.abc import Mapping
from pathlib import Path
from types import MappingProxyType
from typing import NoReturn

import pytest
import yaml
from yaml.nodes import MappingNode, ScalarNode, SequenceNode
from yaml.resolver import BaseResolver


_REPO_ROOT = Path(__file__).resolve().parents[4]
_POLICY_RELATIVE_PATH = Path(
    "testframework/policies/root-compose-testframework.v1.json"
)
_MAX_FILE_BYTES = 1_048_576
_INVALID = object()
_CANARY = "root-compose-contract-canary-secret"
_POLICY_KEYS = (
    "schema",
    "root_compose",
    "public_services",
    "retained_docker_paths",
    "retired_paths",
    "forbidden_tokens",
)
_PUBLIC_SERVICES = (
    "backend-unit",
    "frontend-unit",
    "python-contract",
    "python-api",
    "python-integration",
    "python-e2e-smoke",
    "python-e2e-full",
    "python-load",
)
_QA_INFRASTRUCTURE = (
    "qa-postgres",
    "qa-mongo",
    "qa-redis",
    "qa-nats",
    "qa-seaweedfs",
)
_RETAINED_DOCKER_PATHS = (
    "testframework/docker/Dockerfile.qa",
    "testframework/docker/README.md",
    "testframework/docker/entrypoints/backend-unit.sh",
    "testframework/docker/entrypoints/frontend-unit.sh",
    "testframework/docker/entrypoints/python-qa.sh",
    "testframework/docker/vitest.frontend-unit.config.ts",
)
_RETIRED_PATHS = (
    ".cursor/skills/qa-gate/SKILL.md",
    "deployments/docker-compose.test.yml",
    "testframework/runner/contracts/hub.qa-bootstrap.v1.json",
    "testframework/runner/invoke-qa-bootstrap-v1.ps1",
    "testframework/runner/invoke-qa-gate-v2.ps1",
    "testframework/runner/new-repository-snapshot.ps1",
    "testframework/runner/new-run-envelope.ps1",
    "testframework/runner/README.md",
    "testframework/runner/run-task.ps1",
    "testframework/runner/run-task-v2.ps1",
    "testframework/runner/write-g0-activation.ps1",
    "testframework/selftests/runner/fixtures.ps1",
    "testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1",
    "testframework/selftests/runner/g0-image-contract.ps1",
    "testframework/selftests/runner/g0-runner-v2-contract.ps1",
    "testframework/selftests/runner/private-adapter.ps1",
    "testframework/selftests/runner/runner-contract.ps1",
    "testframework/docker/admission/admit_local_dependency_assets.py",
    "testframework/docker/admission/verify_local_dependency_bundle.py",
    "testframework/docker/admit-local-dependency-assets.ps1",
    "testframework/docker/source-image-attestor.ps1",
    "testframework/docker/bootstrap-control.v1.json",
    "testframework/docker/build-manifest.v2.json",
    "testframework/docker/dependency-bundle.schema.v1.json",
    "testframework/docker/source-policy.v1.json",
    "testframework/docker/compose.g0-bootstrap.yaml",
    "testframework/docker/compose.unit.yaml",
    "testframework/docker/Dockerfile.backend-openapi",
    "testframework/docker/Dockerfile.backend-unit",
    "testframework/docker/Dockerfile.frontend-unit",
    "testframework/docker/Dockerfile.g0-admission-helper",
    "testframework/docker/Dockerfile.g0-planning-selftest",
    "testframework/docker/entrypoints/backend-openapi.sh",
    "testframework/docker/entrypoints/g0-admission-helper.sh",
    "testframework/docker/entrypoints/g0-planning-selftest.sh",
    "testframework/docker/runner-fixtures/g0-planning-selftest.py",
)
_FORBIDDEN_TOKENS = (
    "/qa-gate",
    "invoke-qa-bootstrap",
    "invoke-qa-gate",
    "compose.unit.yaml",
    "compose.g0-bootstrap.yaml",
    "docker.sock",
)
_FIXED_COMMANDS = MappingProxyType(
    {
        "backend-unit": (),
        "frontend-unit": (),
        "python-contract": ("contract",),
        "python-api": ("api",),
        "python-integration": ("integration",),
        "python-e2e-smoke": ("e2e-smoke",),
        "python-e2e-full": ("e2e-full",),
        "python-load": ("load",),
    }
)
_BUILD_TARGETS = MappingProxyType(
    {
        "backend-unit": "backend-unit",
        "frontend-unit": "frontend-unit",
        "python-contract": "python-tests",
        "python-api": "python-tests",
        "python-integration": "python-tests",
        "python-e2e-smoke": "python-tests",
        "python-e2e-full": "python-tests",
        "python-load": "python-tests",
    }
)
_INFRASTRUCTURE_IMAGES = MappingProxyType(
    {
        "qa-postgres": "postgres:17",
        "qa-mongo": "mongo:8.0",
        "qa-redis": "redis:7.4-alpine",
        "qa-nats": "nats:2.11",
        "qa-seaweedfs": "chrislusf/seaweedfs:3.97",
    }
)
_INFRASTRUCTURE_DEPENDENTS = frozenset(
    {"python-integration", "python-e2e-smoke", "python-e2e-full"}
)
_QA_NETWORK = "qa-internal"
_MAINTENANCE_SERVICE = "qa-lock"
_MAINTENANCE_PROFILE = "qa-maintenance"
_QA_LOCK_USER = "1000:1000"
_QA_LOCK_UID_OPTION = "uid=1000"
_QA_LOCK_GID_OPTION = "gid=1000"
_SEAWEEDFS_S3_CONFIG = (
    "{\n"
    '  "identities": [\n'
    "    {\n"
    '      "name": "hubqa-s3",\n'
    '      "credentials": [\n'
    "        {\n"
    '          "accessKey": "hubqa-access",\n'
    '          "secretKey": "hubqa-secret-local-only"\n'
    "        }\n"
    "      ],\n"
    '      "actions": ["Admin", "Read", "List", "Tagging", "Write"]\n'
    "    }\n"
    "  ]\n"
    "}\n"
)
_NATS_READINESS_TEST = (
    "CMD-SHELL",
    "wget -qO- "
    "'http://127.0.0.1:8222/healthz?js-enabled-only=true' >/dev/null",
)
_SEAWEEDFS_READINESS_TEST = (
    "CMD-SHELL",
    "test -r /etc/seaweedfs/s3.json && "
    "grep -q '\"accessKey\": \"hubqa-access\"' "
    "/etc/seaweedfs/s3.json && "
    "wget -qO- http://127.0.0.1:9333/cluster/status >/dev/null && "
    "wget -S -O /dev/null http://127.0.0.1:8333/ 2>&1 | "
    "grep -q ' 403 '",
)
_GO_BASE = (
    "golang@sha256:bbb255b0e131db500cf0520adc97441d2260cf629c7fa7e39e025ddf53995a24"
)
_FRONTEND_BASE = (
    "mcr.microsoft.com/playwright@sha256:"
    "57b65fdc9ceabe0ef613124c7bbe2babcf9362c4d85e382fe3b03604e84b428a"
)
_PYTHON_BASE = (
    "python@sha256:229a2c5bfa27522db7815ea81f9bed70af17ccb9de9fc7ad142b1877b5830d36"
)
_NON_ROOT_SETUP = (
    "RUN groupadd --gid 10001 hubqa && useradd --uid 10001 --gid 10001 "
    "--create-home --home-dir /home/hubqa hubqa"
)
_MINIMAL_VALID_COMPOSE = "services:\n  sample:\n    profiles: [qa]\n"


class RootComposeContractError(ValueError):
    pass


class _DuplicateRejectingSafeLoader(yaml.SafeLoader):
    pass


def test_root_compose_exposes_exact_fixed_qa_services() -> None:
    contract = load_root_compose_contract(_REPO_ROOT)
    policy = _mapping_value(contract, "policy")
    compose = _mapping_value(contract, "compose")
    services = _mapping_value(compose, "services")

    assert tuple(contract) == ("policy", "compose")
    assert tuple(policy) == _POLICY_KEYS
    assert policy["schema"] == "hub.qa.root-compose-testframework/v1"
    assert policy["root_compose"] == "docker-compose.local.yml"
    assert policy["public_services"] == _PUBLIC_SERVICES
    assert policy["retained_docker_paths"] == _RETAINED_DOCKER_PATHS
    assert policy["retired_paths"] == _RETIRED_PATHS
    assert policy["forbidden_tokens"] == _FORBIDDEN_TOKENS
    assert _policy_text(_REPO_ROOT) == _canonical_policy_text()
    assert tuple(name for name in services if name in _PUBLIC_SERVICES) == (
        _PUBLIC_SERVICES
    )
    assert _qa_service_profiles_are_closed(services)
    assert _maintenance_service_profiles_are_closed(services)
    assert _qa_lock_service_is_closed(services)
    assert _qa_extensions_are_closed(compose)
    assert _qa_network_is_closed(compose)
    assert _qa_seaweedfs_config_is_closed(compose)
    assert _qa_merge_provenance_is_closed(
        _required_asset_text(_REPO_ROOT / "docker-compose.local.yml")
    )


def test_qa_services_have_fixed_commands_and_security_options() -> None:
    services = _compose_services(load_root_compose_contract(_REPO_ROOT))

    assert _qa_public_services_are_closed(services)
    assert _qa_lock_service_is_closed(services)
    for name in _PUBLIC_SERVICES:
        service = _mapping_value(services, name)
        assert service["profiles"] == ("qa",)
        assert service["command"] == _FIXED_COMMANDS[name]
        assert service["build"] == MappingProxyType(
            {
                "context": ".",
                "dockerfile": "testframework/docker/Dockerfile.qa",
                "target": _BUILD_TARGETS[name],
            }
        )
        assert service["read_only"] is True
        assert service["security_opt"] == ("no-new-privileges:true",)
        assert service["cap_drop"] == ("ALL",)
        assert service["init"] is True
        assert _resource_limits_are_bounded(service)
        assert _has_reports_mount(service)
        assert "ports" not in service
        assert not _contains_token(service, "docker.sock")
        assert _dependencies_are_closed(name, service)


def test_current_source_overlays_reject_writable_stale_and_undersized_mutations() -> None:
    services = _compose_services(load_root_compose_contract(_REPO_ROOT))

    assert _qa_public_services_are_closed(services)

    frontend = _mapping_value(services, "frontend-unit")
    frontend_volumes = frontend["volumes"]
    assert type(frontend_volumes) is tuple and len(frontend_volumes) == 3
    frontend_src = frontend_volumes[1]
    assert isinstance(frontend_src, Mapping)
    writable_frontend_src = dict(frontend_src)
    writable_frontend_src["read_only"] = False
    assert not _qa_public_services_are_closed(
        _mutate_service(
            services,
            "frontend-unit",
            {
                "volumes": (
                    frontend_volumes[0],
                    MappingProxyType(writable_frontend_src),
                    *frontend_volumes[2:],
                )
            },
        )
    )

    frontend_tests = frontend_volumes[2]
    assert isinstance(frontend_tests, Mapping)
    stale_frontend_tests = dict(frontend_tests)
    stale_frontend_tests["source"] = "./frontend/tests-stale"
    assert not _qa_public_services_are_closed(
        _mutate_service(
            services,
            "frontend-unit",
            {
                "volumes": (
                    frontend_volumes[0],
                    frontend_volumes[1],
                    MappingProxyType(stale_frontend_tests),
                )
            },
        )
    )

    frontend_tmpfs = frontend["tmpfs"]
    assert type(frontend_tmpfs) is tuple and len(frontend_tmpfs) == 3
    frontend_tmp = frontend_tmpfs[2]
    assert type(frontend_tmp) is str
    assert not _qa_public_services_are_closed(
        _mutate_service(
            services,
            "frontend-unit",
            {
                "tmpfs": (
                    frontend_tmpfs[0],
                    frontend_tmpfs[1],
                    frontend_tmp.replace("size=4g", "size=3g", 1),
                )
            },
        )
    )

    python_contract = _mapping_value(services, "python-contract")
    python_contract_volumes = python_contract["volumes"]
    assert (
        type(python_contract_volumes) is tuple
        and len(python_contract_volumes) == 5
    )
    selftests_bind = python_contract_volumes[2]
    assert isinstance(selftests_bind, Mapping)
    writable_selftests_bind = dict(selftests_bind)
    writable_selftests_bind["read_only"] = False
    assert not _qa_public_services_are_closed(
        _mutate_service(
            services,
            "python-contract",
            {
                "volumes": (
                    python_contract_volumes[0],
                    python_contract_volumes[1],
                    MappingProxyType(writable_selftests_bind),
                    *python_contract_volumes[3:],
                )
            },
        )
    )

    assert not _qa_public_services_are_closed(
        _mutate_service(
            services,
            "python-contract",
            {
                "volumes": (
                    python_contract_volumes[0],
                    python_contract_volumes[1],
                    python_contract_volumes[2],
                    python_contract_volumes[4],
                )
            },
        )
    )

    phase_plan = python_contract_volumes[4]
    assert isinstance(phase_plan, Mapping)
    writable_phase_plan = dict(phase_plan)
    writable_phase_plan["read_only"] = False
    assert not _qa_public_services_are_closed(
        _mutate_service(
            services,
            "python-contract",
            {
                "volumes": (
                    *python_contract_volumes[:4],
                    MappingProxyType(writable_phase_plan),
                )
            },
        )
    )

    backend = _mapping_value(services, "backend-unit")
    backend_tmpfs = backend["tmpfs"]
    assert type(backend_tmpfs) is tuple and len(backend_tmpfs) == 3
    backend_cache = backend_tmpfs[1]
    assert type(backend_cache) is str
    assert not _qa_public_services_are_closed(
        _mutate_service(
            services,
            "backend-unit",
            {
                "tmpfs": (
                    backend_tmpfs[0],
                    backend_cache.replace("size=3g", "size=2g", 1),
                    backend_tmpfs[2],
                )
            },
        )
    )

def test_qa_infrastructure_is_unpublished_and_disposable() -> None:
    services = _compose_services(load_root_compose_contract(_REPO_ROOT))

    assert _qa_infrastructure_services_are_closed(services)
    for name in _QA_INFRASTRUCTURE:
        service = _mapping_value(services, name)
        assert service["profiles"] == ("qa",)
        assert service["image"] == _INFRASTRUCTURE_IMAGES[name]
        assert service["networks"] == (_QA_NETWORK,)
        assert "ports" not in service
        assert _mapping_value(service, "healthcheck")
        assert _resource_limits_are_bounded(service)
        assert _storage_is_disposable(service)
        assert not _contains_token(service, "docker.sock")
        assert "privileged" not in service
        assert "network_mode" not in service


def test_docker_directory_contains_only_manifested_assets() -> None:
    contract = load_root_compose_contract(_REPO_ROOT)
    policy = _mapping_value(contract, "policy")

    dockerfile = _required_asset_text(
        _REPO_ROOT / "testframework" / "docker" / "Dockerfile.qa"
    )
    backend_entrypoint = _required_asset_text(
        _REPO_ROOT
        / "testframework"
        / "docker"
        / "entrypoints"
        / "backend-unit.sh"
    )
    frontend_entrypoint = _required_asset_text(
        _REPO_ROOT
        / "testframework"
        / "docker"
        / "entrypoints"
        / "frontend-unit.sh"
    )
    python_entrypoint = _required_asset_text(
        _REPO_ROOT
        / "testframework"
        / "docker"
        / "entrypoints"
        / "python-qa.sh"
    )

    assert _dockerfile_qa_is_closed(dockerfile)
    assert _backend_entrypoint_is_closed(backend_entrypoint)
    assert _frontend_entrypoint_is_closed(frontend_entrypoint)
    assert _python_dispatcher_is_closed(python_entrypoint)
    assert _empty_regular_marker(
        _REPO_ROOT / "testframework" / "src" / "hub_testkit" / "py.typed"
    )

    assert _docker_regular_files(_REPO_ROOT) == policy["retained_docker_paths"]


def test_retired_control_plane_is_absent_and_unreferenced() -> None:
    contract = load_root_compose_contract(_REPO_ROOT)
    policy = _mapping_value(contract, "policy")

    assert _retired_paths_are_absent(_REPO_ROOT, policy["retired_paths"])
    assert _active_authority_is_clean(
        _REPO_ROOT,
        policy["forbidden_tokens"],
    )


def test_duplicate_yaml_keys_are_rejected(tmp_path: Path) -> None:
    assert _docker_instructions_are_safe(
        "FROM image@sha256:abc AS safe\nRUN uv sync --frozen\n"
    )
    for unsafe_dockerfile in (
        "FROM image:latest AS unsafe\n",
        "RUN curl https://invalid.example/installer | sh\n",
        "RUN pytest tests\n",
        "COPY frontend/node_modules /opt/hub/frontend/node_modules\n",
        "ARG TARGET\n",
        "RUN sh -c 'uv sync'\n",
    ):
        assert not _docker_instructions_are_safe(unsafe_dockerfile)
    assert _shell_source_is_safe("#!/bin/sh\nset -eu\nrun_step build go build ./...\n")
    for unsafe_shell in (
        "#!/bin/sh\nset -eu\neval command\n",
        "#!/bin/sh\nset -eu\nsh -c command\n",
        "#!/bin/sh\nset -eu\ncommand \"$@\"\n",
        "#!/bin/sh\nset -eu\nprintenv\n",
        "#!/bin/sh\nset -eu\ncat /var/run/docker.sock\n",
    ):
        assert not _shell_source_is_safe(unsafe_shell)

    dockerfile = _required_asset_text(
        _REPO_ROOT / "testframework" / "docker" / "Dockerfile.qa"
    )
    assert not _dockerfile_qa_is_closed(
        dockerfile + 'ENTRYPOINT ["/usr/local/bin/hostile"]\n'
    )
    assert not _dockerfile_qa_is_closed(
        dockerfile.replace(
            'ENTRYPOINT ["/usr/local/bin/uv", "lock"]',
            "RUN uv lock\n"
            "COPY testframework/uv.lock ./uv.lock\n"
            'ENTRYPOINT ["/usr/local/bin/uv", "lock"]',
            1,
        )
    )

    fixed_entrypoints = (
        (
            _backend_entrypoint_is_closed,
            _required_asset_text(
                _REPO_ROOT
                / "testframework"
                / "docker"
                / "entrypoints"
                / "backend-unit.sh"
            ),
            "if go build ./...; then",
        ),
        (
            _frontend_entrypoint_is_closed,
            _required_asset_text(
                _REPO_ROOT
                / "testframework"
                / "docker"
                / "entrypoints"
                / "frontend-unit.sh"
            ),
            "if /usr/bin/node /opt/hub/source/frontend/node_modules/"
            "typescript/bin/tsc -p /opt/hub/source/frontend/"
            "tsconfig.json --noEmit; then",
        ),
    )
    for validator, source, first_command in fixed_entrypoints:
        assert not validator(
            source.replace(first_command, f"exit 0\n{first_command}", 1)
        )
        assert not validator(
            source.replace(
                first_command,
                f'command "$HUB_QA_COMMAND"\n{first_command}',
                1,
            )
        )
    frontend_source = fixed_entrypoints[1][1]
    assert frontend_source.count("/usr/bin/node ") == 2
    assert not _frontend_entrypoint_is_closed(
        frontend_source.replace("/usr/bin/node", "node")
    )

    services = _compose_services(load_root_compose_contract(_REPO_ROOT))
    assert _qa_public_services_are_closed(services)
    assert _qa_infrastructure_services_are_closed(services)
    assert _maintenance_service_profiles_are_closed(services)
    assert _qa_lock_service_is_closed(services)

    assert not _maintenance_service_profiles_are_closed(
        _mutate_service(
            services,
            _MAINTENANCE_SERVICE,
            {"profiles": ("qa",)},
        )
    )
    assert not _qa_lock_service_is_closed(
        _mutate_nested_mapping(
            services,
            _MAINTENANCE_SERVICE,
            "build",
            "target",
            "python-tests",
        )
    )
    assert not _qa_lock_service_is_closed(
        _mutate_service(
            services,
            _MAINTENANCE_SERVICE,
            {"command": ("--project", "/workspace/hostile")},
        )
    )
    assert not _qa_lock_service_is_closed(
        _mutate_service(
            services,
            _MAINTENANCE_SERVICE,
            {
                "command": (
                    "--project",
                    "/workspace/testframework",
                    "--upgrade",
                )
            },
        )
    )
    qa_lock = _mapping_value(services, _MAINTENANCE_SERVICE)
    project_binds = qa_lock["volumes"]
    assert type(project_binds) is tuple and len(project_binds) == 1
    project_bind = project_binds[0]
    assert isinstance(project_bind, Mapping)
    assert not _qa_lock_service_is_closed(
        _mutate_service(
            services,
            _MAINTENANCE_SERVICE,
            {
                "volumes": project_binds
                + (
                    MappingProxyType(
                        {
                            "type": "bind",
                            "source": "./testframework/.cache",
                            "target": "/cache",
                            "read_only": False,
                        }
                    ),
                )
            },
        )
    )
    read_only_bind = dict(project_bind)
    read_only_bind["read_only"] = True
    assert not _qa_lock_service_is_closed(
        _mutate_service(
            services,
            _MAINTENANCE_SERVICE,
            {"volumes": (MappingProxyType(read_only_bind),)},
        )
    )
    socket_bind = dict(project_bind)
    socket_bind["source"] = "/var/run/docker.sock"
    socket_bind["target"] = "/var/run/docker.sock"
    assert not _qa_lock_service_is_closed(
        _mutate_service(
            services,
            _MAINTENANCE_SERVICE,
            {"volumes": (MappingProxyType(socket_bind),)},
        )
    )
    assert not _qa_lock_service_is_closed(
        _remove_service_field(
            services,
            _MAINTENANCE_SERVICE,
            "security_opt",
        )
    )
    assert not _qa_lock_service_is_closed(
        _remove_service_field(
            services,
            _MAINTENANCE_SERVICE,
            "user",
        )
    )
    assert not _qa_lock_service_is_closed(
        _mutate_service(
            services,
            _MAINTENANCE_SERVICE,
            {"user": "${QA_HOST_UID:-1000}:1000"},
        )
    )
    assert not _qa_lock_service_is_closed(
        _mutate_service(
            services,
            _MAINTENANCE_SERVICE,
            {"user": "0:0"},
        )
    )
    assert not _qa_lock_service_is_closed(
        _mutate_service(
            services,
            _MAINTENANCE_SERVICE,
            {"user": "1001:1001"},
        )
    )
    assert not _qa_lock_service_is_closed(
        _mutate_service(
            services,
            _MAINTENANCE_SERVICE,
            {"user": "1000:1000:extra"},
        )
    )
    assert not _qa_lock_service_is_closed(
        _mutate_service(
            services,
            _MAINTENANCE_SERVICE,
            {"group_add": ("1000",)},
        )
    )
    qa_lock_tmpfs = qa_lock["tmpfs"]
    assert type(qa_lock_tmpfs) is tuple and len(qa_lock_tmpfs) == 3
    first_tmpfs = qa_lock_tmpfs[0]
    assert type(first_tmpfs) is str
    mismatched_tmpfs = (
        first_tmpfs.replace(_QA_LOCK_UID_OPTION, "uid=1001", 1),
        qa_lock_tmpfs[1],
        qa_lock_tmpfs[2],
    )
    assert not _qa_lock_service_is_closed(
        _mutate_service(
            services,
            _MAINTENANCE_SERVICE,
            {"tmpfs": mismatched_tmpfs},
        )
    )
    second_maintenance = dict(services)
    second_maintenance["qa-lock-second"] = MappingProxyType(
        {"profiles": (_MAINTENANCE_PROFILE,)}
    )
    assert not _maintenance_service_profiles_are_closed(
        MappingProxyType(second_maintenance)
    )

    extra_qa = dict(services)
    extra_qa["qa-hostile"] = MappingProxyType({"profiles": ("qa",)})
    assert not _qa_service_profiles_are_closed(MappingProxyType(extra_qa))

    assert not _qa_public_services_are_closed(
        _mutate_service(services, "backend-unit", {"ports": ("3000:3000",)})
    )
    assert not _qa_public_services_are_closed(
        _mutate_service(
            services,
            "backend-unit",
            {"volumes": ("/var/run/docker.sock:/var/run/docker.sock",)},
        )
    )
    assert not _qa_public_services_are_closed(
        _mutate_nested_mapping(
            services,
            "backend-unit",
            "build",
            "target",
            "hostile-target",
        )
    )
    assert not _qa_public_services_are_closed(
        _mutate_service(services, "python-api", {"command": ("contract",)})
    )
    assert not _qa_infrastructure_services_are_closed(
        _mutate_service(
            services,
            "qa-postgres",
            {"volumes": ("qa-persistent:/var/lib/postgresql/data",)},
        )
    )
    for infrastructure_name in ("qa-nats", "qa-seaweedfs"):
        assert not _qa_infrastructure_services_are_closed(
            _mutate_nested_mapping(
                services,
                infrastructure_name,
                "healthcheck",
                "test",
                ("CMD-SHELL", "kill -0 1"),
            )
        )
    integration = _mapping_value(services, "python-integration")
    dependencies = dict(_mapping_value(integration, "depends_on"))
    dependencies.pop("qa-nats")
    assert not _qa_public_services_are_closed(
        _mutate_service(
            services,
            "python-integration",
            {"depends_on": MappingProxyType(dependencies)},
        )
    )
    assert not _qa_public_services_are_closed(
        _mutate_service(services, "frontend-unit", {"privileged": True})
    )
    compose = _mapping_value(load_root_compose_contract(_REPO_ROOT), "compose")
    hostile_config = dict(compose)
    hostile_config["configs"] = MappingProxyType(
        {
            "qa-seaweedfs-s3": MappingProxyType(
                {"content": '{"identities":[{"name":"anonymous"}]}\n'}
            )
        }
    )
    assert not _qa_seaweedfs_config_is_closed(
        MappingProxyType(hostile_config)
    )

    compose_text = _required_asset_text(
        _REPO_ROOT / "docker-compose.local.yml"
    )
    copied_python_api = (
        "    profiles: [qa]\n"
        "    read_only: true\n"
        '    security_opt: ["no-new-privileges:true"]\n'
        "    cap_drop: [ALL]\n"
        "    init: true\n"
        "    pids_limit: 256\n"
        "    cpus: 2.0\n"
        "    mem_limit: 2g\n"
        "    volumes:\n"
        "      - ./testframework/reports:/reports\n"
        "    networks: [qa-internal]\n"
        "    stop_grace_period: 30s\n"
        "    tmpfs:\n"
        '      - "/tmp:rw,noexec,nosuid,nodev,size=64m"\n'
        "    environment:\n"
        '      PYTHONDONTWRITEBYTECODE: "1"\n'
        '      PYTEST_ADDOPTS: "-p no:cacheprovider"\n'
        '      PYTHONSAFEPATH: "1"\n'
    )
    assert "\r" not in compose_text.replace("\r\n", "")
    normalized_compose_text = compose_text.replace("\r\n", "\n")
    python_api_marker = "  python-api:\n    <<: [*qa-common, *qa-python]\n"
    assert normalized_compose_text.count(python_api_marker) == 1
    copy_expanded = normalized_compose_text.replace(
        python_api_marker,
        "  python-api:\n" + copied_python_api,
        1,
    )
    assert copy_expanded != normalized_compose_text
    expanded_document = _parse_yaml_safe(copy_expanded)
    assert isinstance(expanded_document, Mapping)
    assert _qa_public_services_are_closed(
        _mapping_value(expanded_document, "services")
    )
    assert not _qa_merge_provenance_is_closed(copy_expanded)

    materialized_merge = _parse_yaml_safe(
        "x-hostile: &hostile\n"
        "  privileged: true\n"
        "services:\n"
        "  frontend-unit:\n"
        "    <<: *hostile\n"
        "    profiles: [qa]\n"
    )
    assert isinstance(materialized_merge, Mapping)
    assert not _qa_public_services_are_closed(
        _mapping_value(materialized_merge, "services")
    )
    merged_profile_override = _parse_yaml_safe(
        "x-common: &common\n"
        "  profiles: [qa]\n"
        "services:\n"
        "  backend-unit:\n"
        "    <<: *common\n"
        "    profiles: [qa, hostile]\n"
    )
    assert isinstance(merged_profile_override, Mapping)
    assert not _qa_public_services_are_closed(
        _mapping_value(merged_profile_override, "services")
    )
    assert (
        _parse_yaml_safe(
            "services:\n"
            "  backend-unit:\n"
            "    profiles: [qa]\n"
            "    profiles: [qa, hostile]\n"
        )
        is _INVALID
    )

    admitted_qa_services = MappingProxyType(
        {
            "backend-unit": MappingProxyType({"profiles": ("qa",)}),
            "frontend-unit": MappingProxyType({"profiles": ("qa",)}),
            "python-contract": MappingProxyType({"profiles": ("qa",)}),
            "python-api": MappingProxyType({"profiles": ("qa",)}),
            "python-integration": MappingProxyType({"profiles": ("qa",)}),
            "python-e2e-smoke": MappingProxyType({"profiles": ("qa",)}),
            "python-e2e-full": MappingProxyType({"profiles": ("qa",)}),
            "python-load": MappingProxyType({"profiles": ("qa",)}),
            "qa-postgres": MappingProxyType({"profiles": ("qa",)}),
            "qa-mongo": MappingProxyType({"profiles": ("qa",)}),
            "qa-redis": MappingProxyType({"profiles": ("qa",)}),
            "qa-nats": MappingProxyType({"profiles": ("qa",)}),
            "qa-seaweedfs": MappingProxyType({"profiles": ("qa",)}),
        }
    )
    assert _qa_service_profiles_are_closed(admitted_qa_services)
    escaped_qa_services = dict(admitted_qa_services)
    escaped_qa_services["unbounded-extra"] = MappingProxyType(
        {"profiles": ("qa", "another-profile")}
    )
    assert not _qa_service_profiles_are_closed(
        MappingProxyType(escaped_qa_services)
    )

    stable_identity = (11, 22, 33, 44, 55)
    assert _descriptor_metadata_is_stable(stable_identity, stable_identity)
    assert not _descriptor_metadata_is_stable(
        stable_identity,
        (11, 22, 33, 45, 55),
    )
    assert not _descriptor_metadata_is_stable(
        stable_identity,
        (11, 22, 33, 44, 56),
    )

    nested_duplicate = (
        '{"outer":[{"admitted":"value",'
        f'"admitted":"{_CANARY}"}}]}}'
    )
    assert _load_json_document_safe(nested_duplicate) is _INVALID

    _write_valid_layout(tmp_path)
    contract = load_root_compose_contract(tmp_path)
    policy = _mapping_value(contract, "policy")
    compose = _mapping_value(contract, "compose")
    assert tuple(contract) == ("policy", "compose")
    assert policy["public_services"] == _PUBLIC_SERVICES
    services = _mapping_value(compose, "services")
    sample = _mapping_value(services, "sample")
    assert sample["profiles"] == ("qa",)
    with pytest.raises(TypeError):
        contract["policy"] = MappingProxyType({})  # type: ignore[index]
    with pytest.raises(TypeError):
        policy["schema"] = "changed"  # type: ignore[index]
    with pytest.raises(TypeError):
        sample["profiles"] = ("changed",)  # type: ignore[index]

    compose_path = tmp_path / "docker-compose.local.yml"
    compose_path.write_text(
        "services: {}\nx-unrelated-authority: docker.sock\n",
        encoding="utf-8",
        newline="\n",
    )
    assert not _active_authority_is_clean(tmp_path, _FORBIDDEN_TOKENS)

    _write_valid_layout(tmp_path)
    nested_doc = tmp_path / "testframework" / "guides" / "active.md"
    nested_doc.parent.mkdir(parents=True, exist_ok=True)
    nested_doc.write_text(
        "Use /qa-gate as an active authority.\n",
        encoding="utf-8",
        newline="\n",
    )
    assert not _active_authority_is_clean(tmp_path, _FORBIDDEN_TOKENS)
    nested_doc.unlink()

    nested_doc.write_bytes(b"")
    assert not _active_authority_is_clean(tmp_path, _FORBIDDEN_TOKENS)
    nested_doc.unlink()

    nested_target = tmp_path / "nested-authority-target.md"
    nested_target.write_text("active authority\n", encoding="utf-8", newline="\n")
    try:
        os.symlink(nested_target, nested_doc)
    except OSError:
        pass
    else:
        assert not _active_authority_is_clean(tmp_path, _FORBIDDEN_TOKENS)
        nested_doc.unlink()

    compose_path.write_text(
        "services:\n  qa-postgres:\n    image: postgres:17\n"
        "    healthcheck:\n      test: first\n      test: second\n",
        encoding="utf-8",
        newline="\n",
    )
    _assert_generic_failure(tmp_path)

    _write_valid_layout(tmp_path)
    policy_path = tmp_path / _POLICY_RELATIVE_PATH
    policy_path.write_text(
        '{"schema":"hub.qa.root-compose-testframework/v1",'
        f'"schema":"{_CANARY}"}}\n',
        encoding="utf-8",
        newline="\n",
    )
    _assert_generic_failure(tmp_path)

    _write_valid_layout(tmp_path)
    nested_policy = _canonical_policy_text().replace(
        '    "backend-unit",',
        f'    {{"name": "backend-unit", "name": "{_CANARY}"}},',
        1,
    )
    policy_path.write_text(
        nested_policy,
        encoding="utf-8",
        newline="\n",
    )
    _assert_generic_failure(tmp_path)

    _write_valid_layout(tmp_path)
    unknown = _policy_document()
    unknown["unexpected"] = _CANARY
    _write_json(policy_path, unknown)
    _assert_generic_failure(tmp_path)

    _write_valid_layout(tmp_path)
    document = _policy_document()
    reordered = {"root_compose": document["root_compose"]}
    reordered.update(document)
    _write_json(policy_path, reordered)
    _assert_generic_failure(tmp_path)

    _write_valid_layout(tmp_path)
    compose_path.write_bytes(b"\xff" + _CANARY.encode("utf-8"))
    _assert_generic_failure(tmp_path)

    canary_root = tmp_path / _CANARY
    _write_valid_layout(canary_root)
    (canary_root / "docker-compose.local.yml").write_bytes(b"\xff")
    _assert_generic_failure(canary_root)

    _write_valid_layout(tmp_path)
    compose_path.write_bytes(b"x" * 1_048_577)
    _assert_generic_failure(tmp_path)

    _write_valid_layout(tmp_path)
    compose_path.write_bytes(b"")
    _assert_generic_failure(tmp_path)

    _write_valid_layout(tmp_path)
    compose_path.unlink()
    compose_path.mkdir()
    _assert_generic_failure(tmp_path)
    compose_path.rmdir()

    _write_valid_layout(tmp_path)
    target = tmp_path / "compose-target.yml"
    target.write_text(_MINIMAL_VALID_COMPOSE, encoding="utf-8", newline="\n")
    compose_path.unlink()
    try:
        os.symlink(target, compose_path)
    except OSError:
        pass
    else:
        _assert_generic_failure(tmp_path)


def load_root_compose_contract(repo_root: Path) -> Mapping[str, object]:
    result = _load_contract_safe(repo_root)
    del repo_root
    if isinstance(result, Mapping):
        return result
    del result
    _raise_invalid()


def _load_contract_safe(repo_root: object) -> Mapping[str, object] | object:
    try:
        if not isinstance(repo_root, Path):
            return _INVALID
        policy_text = _read_bounded_regular(repo_root / _POLICY_RELATIVE_PATH)
        if type(policy_text) is not str:
            return _INVALID
        policy = _parse_policy_safe(policy_text)
        if not isinstance(policy, Mapping):
            return _INVALID
        compose_text = _read_bounded_regular(repo_root / "docker-compose.local.yml")
        if type(compose_text) is not str:
            return _INVALID
        compose = _parse_yaml_safe(compose_text)
        if not isinstance(compose, Mapping):
            return _INVALID
        return MappingProxyType({"policy": policy, "compose": compose})
    except Exception:
        return _INVALID


def _read_bounded_regular(path: Path) -> str | object:
    descriptor = -1
    try:
        before = path.lstat()
        if not _stat_is_admitted(before):
            return _INVALID
        flags = os.O_RDONLY
        flags |= getattr(os, "O_BINARY", 0)
        flags |= getattr(os, "O_NOFOLLOW", 0)
        descriptor = os.open(path, flags)
        opened = os.fstat(descriptor)
        before_identity = _descriptor_metadata(before)
        opened_identity = _descriptor_metadata(opened)
        if (
            not _descriptor_metadata_is_stable(before_identity, opened_identity)
            or not _stat_is_admitted(opened)
        ):
            return _INVALID
        chunks: list[bytes] = []
        remaining = _MAX_FILE_BYTES + 1
        while remaining:
            chunk = os.read(descriptor, min(65_536, remaining))
            if not chunk:
                break
            chunks.append(chunk)
            remaining -= len(chunk)
        after = os.fstat(descriptor)
        content = b"".join(chunks)
        if (
            not _descriptor_metadata_is_stable(
                opened_identity,
                _descriptor_metadata(after),
            )
            or not 1 <= len(content) <= _MAX_FILE_BYTES
            or content.startswith(b"\xef\xbb\xbf")
            or b"\x00" in content
        ):
            return _INVALID
        return content.decode("utf-8", errors="strict")
    except Exception:
        return _INVALID
    finally:
        if descriptor >= 0:
            try:
                os.close(descriptor)
            except OSError:
                pass


def _stat_is_admitted(value: os.stat_result) -> bool:
    reparse_flag = getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", 0)
    attributes = getattr(value, "st_file_attributes", 0)
    return (
        stat.S_ISREG(value.st_mode)
        and not attributes & reparse_flag
        and 1 <= value.st_size <= _MAX_FILE_BYTES
    )


def _descriptor_metadata(value: os.stat_result) -> tuple[int, int, int, int, int]:
    return (
        value.st_dev,
        value.st_ino,
        value.st_size,
        value.st_mtime_ns,
        value.st_ctime_ns,
    )


def _descriptor_metadata_is_stable(
    before: tuple[int, int, int, int, int],
    after: tuple[int, int, int, int, int],
) -> bool:
    return before == after


def _parse_policy_safe(text: object) -> Mapping[str, object] | object:
    try:
        if type(text) is not str:
            return _INVALID
        document = _load_json_document_safe(text)
        if type(document) is not dict or tuple(document) != _POLICY_KEYS:
            return _INVALID
        if document["schema"] != "hub.qa.root-compose-testframework/v1":
            return _INVALID
        if document["root_compose"] != "docker-compose.local.yml":
            return _INVALID
        expected_lists = (
            ("public_services", _PUBLIC_SERVICES),
            ("retained_docker_paths", _RETAINED_DOCKER_PATHS),
            ("retired_paths", _RETIRED_PATHS),
            ("forbidden_tokens", _FORBIDDEN_TOKENS),
        )
        for key, expected in expected_lists:
            value = document[key]
            if (
                type(value) is not list
                or any(type(item) is not str for item in value)
                or len(value) != len(set(value))
                or tuple(value) != expected
            ):
                return _INVALID
        return MappingProxyType(
            {
                "schema": document["schema"],
                "root_compose": document["root_compose"],
                "public_services": tuple(document["public_services"]),
                "retained_docker_paths": tuple(document["retained_docker_paths"]),
                "retired_paths": tuple(document["retired_paths"]),
                "forbidden_tokens": tuple(document["forbidden_tokens"]),
            }
        )
    except Exception:
        return _INVALID


def _load_json_document_safe(text: object) -> object:
    try:
        if type(text) is not str:
            return _INVALID
        return json.loads(text, object_pairs_hook=_json_pairs_without_duplicates)
    except Exception:
        return _INVALID


def _json_pairs_without_duplicates(
    pairs: list[tuple[str, object]],
) -> dict[str, object]:
    result: dict[str, object] = {}
    for key, value in pairs:
        if type(key) is not str or key in result:
            raise ValueError("invalid policy")
        result[key] = value
    return result


def _construct_unique_mapping(
    loader: _DuplicateRejectingSafeLoader,
    node: MappingNode,
    deep: bool = False,
) -> dict[str, object]:
    if not isinstance(node, MappingNode):
        raise ValueError("invalid compose")
    explicit_keys: set[str] = set()
    for key_node, _ in node.value:
        if key_node.tag == "tag:yaml.org,2002:merge":
            continue
        key = loader.construct_object(key_node, deep=deep)
        if type(key) is not str or key in explicit_keys:
            raise ValueError("invalid compose")
        explicit_keys.add(key)
    loader.flatten_mapping(node)
    result: dict[str, object] = {}
    for key_node, value_node in node.value:
        key = loader.construct_object(key_node, deep=deep)
        if type(key) is not str:
            raise ValueError("invalid compose")
        result[key] = loader.construct_object(value_node, deep=deep)
    return result


_DuplicateRejectingSafeLoader.add_constructor(
    BaseResolver.DEFAULT_MAPPING_TAG,
    _construct_unique_mapping,
)


def _parse_yaml_safe(text: object) -> Mapping[str, object] | object:
    try:
        if type(text) is not str:
            return _INVALID
        document = yaml.load(text, Loader=_DuplicateRejectingSafeLoader)
        if type(document) is not dict or type(document.get("services")) is not dict:
            return _INVALID
        frozen = _freeze_safe(document, set(), 0, [0])
        if isinstance(frozen, Mapping):
            return frozen
        return _INVALID
    except Exception:
        return _INVALID


def _freeze_safe(
    value: object,
    active_ids: set[int],
    depth: int,
    count: list[int],
) -> object:
    if depth > 64:
        return _INVALID
    count[0] += 1
    if count[0] > 32_768:
        return _INVALID
    if type(value) is dict:
        identity = id(value)
        if identity in active_ids:
            return _INVALID
        active_ids.add(identity)
        result: dict[str, object] = {}
        for key, child in value.items():
            if type(key) is not str:
                active_ids.remove(identity)
                return _INVALID
            frozen = _freeze_safe(child, active_ids, depth + 1, count)
            if frozen is _INVALID:
                active_ids.remove(identity)
                return _INVALID
            result[key] = frozen
        active_ids.remove(identity)
        return MappingProxyType(result)
    if type(value) is list:
        identity = id(value)
        if identity in active_ids:
            return _INVALID
        active_ids.add(identity)
        children = tuple(
            _freeze_safe(child, active_ids, depth + 1, count) for child in value
        )
        active_ids.remove(identity)
        if any(child is _INVALID for child in children):
            return _INVALID
        return children
    if value is None or type(value) in {str, int, float, bool}:
        return value
    return _INVALID


def _mapping_value(value: Mapping[str, object], key: str) -> Mapping[str, object]:
    result = value[key]
    assert isinstance(result, Mapping)
    return result


def _compose_services(contract: Mapping[str, object]) -> Mapping[str, object]:
    return _mapping_value(_mapping_value(contract, "compose"), "services")


def _qa_service_profiles_are_closed(services: object) -> bool:
    try:
        if not isinstance(services, Mapping):
            return False
        allowed = frozenset(_PUBLIC_SERVICES + _QA_INFRASTRUCTURE)
        selected: set[str] = set()
        for name, value in services.items():
            if type(name) is not str or not isinstance(value, Mapping):
                return False
            profiles = value.get("profiles")
            includes_qa = profiles == "qa" or (
                type(profiles) is tuple and "qa" in profiles
            )
            if includes_qa:
                if name not in allowed or profiles != ("qa",):
                    return False
                selected.add(name)
        return selected == allowed
    except Exception:
        return False


def _maintenance_service_profiles_are_closed(services: object) -> bool:
    try:
        if not isinstance(services, Mapping):
            return False
        selected: set[str] = set()
        for name, value in services.items():
            if type(name) is not str or not isinstance(value, Mapping):
                return False
            profiles = value.get("profiles")
            includes_maintenance = profiles == _MAINTENANCE_PROFILE or (
                type(profiles) is tuple
                and _MAINTENANCE_PROFILE in profiles
            )
            if includes_maintenance:
                if (
                    name != _MAINTENANCE_SERVICE
                    or profiles != (_MAINTENANCE_PROFILE,)
                ):
                    return False
                selected.add(name)
        maintenance = services.get(_MAINTENANCE_SERVICE)
        return (
            selected == {_MAINTENANCE_SERVICE}
            and isinstance(maintenance, Mapping)
            and maintenance.get("profiles") == (_MAINTENANCE_PROFILE,)
        )
    except Exception:
        return False


def _qa_lock_service_is_closed(services: object) -> bool:
    try:
        if not isinstance(services, Mapping):
            return False
        return services.get(_MAINTENANCE_SERVICE) == MappingProxyType(
            {
                "profiles": (_MAINTENANCE_PROFILE,),
                "build": MappingProxyType(
                    {
                        "context": ".",
                        "dockerfile": "testframework/docker/Dockerfile.qa",
                        "target": "python-lock",
                    }
                ),
                "command": ("--project", "/workspace/testframework"),
                "volumes": (
                    MappingProxyType(
                        {
                            "type": "bind",
                            "source": "./testframework",
                            "target": "/workspace/testframework",
                            "read_only": False,
                        }
                    ),
                ),
                "read_only": True,
                "user": _QA_LOCK_USER,
                "tmpfs": (
                    "/tmp:rw,noexec,nosuid,nodev,size=64m,"
                    f"{_QA_LOCK_UID_OPTION},{_QA_LOCK_GID_OPTION},mode=1770",
                    "/home/hubqa:rw,noexec,nosuid,nodev,size=32m,"
                    f"{_QA_LOCK_UID_OPTION},{_QA_LOCK_GID_OPTION},mode=0700",
                    "/opt/hub/uv-cache:rw,noexec,nosuid,nodev,size=512m,"
                    f"{_QA_LOCK_UID_OPTION},{_QA_LOCK_GID_OPTION},mode=0700",
                ),
                "environment": MappingProxyType(
                    {
                        "HOME": "/home/hubqa",
                        "UV_CACHE_DIR": "/opt/hub/uv-cache",
                    }
                ),
                "security_opt": ("no-new-privileges:true",),
                "cap_drop": ("ALL",),
                "init": True,
                "pids_limit": 128,
                "cpus": 1.0,
                "mem_limit": "1g",
                "stop_grace_period": "30s",
            }
        )
    except Exception:
        return False


def _qa_extensions_are_closed(compose: object) -> bool:
    try:
        if not isinstance(compose, Mapping):
            return False
        return (
            compose.get("x-qa-common") == _qa_common_service_fields()
            and compose.get("x-qa-python")
            == MappingProxyType({"environment": _qa_python_environment()})
            and compose.get("x-qa-infrastructure")
            == _qa_infrastructure_common_fields()
        )
    except Exception:
        return False


def _qa_network_is_closed(compose: object) -> bool:
    try:
        if not isinstance(compose, Mapping):
            return False
        return compose.get("networks") == MappingProxyType(
            {_QA_NETWORK: MappingProxyType({"internal": True})}
        )
    except Exception:
        return False


def _qa_seaweedfs_config_is_closed(compose: object) -> bool:
    try:
        if not isinstance(compose, Mapping):
            return False
        return compose.get("configs") == MappingProxyType(
            {
                "qa-seaweedfs-s3": MappingProxyType(
                    {"content": _SEAWEEDFS_S3_CONFIG}
                )
            }
        )
    except Exception:
        return False


def _qa_merge_provenance_is_closed(text: object) -> bool:
    try:
        if type(text) is not str:
            return False
        root = yaml.compose(text, Loader=_DuplicateRejectingSafeLoader)
        if not isinstance(root, MappingNode):
            return False
        qa_common = _yaml_mapping_child(root, "x-qa-common")
        qa_python = _yaml_mapping_child(root, "x-qa-python")
        qa_infrastructure = _yaml_mapping_child(
            root,
            "x-qa-infrastructure",
        )
        services = _yaml_mapping_child(root, "services")
        if not all(
            isinstance(value, MappingNode)
            for value in (
                qa_common,
                qa_python,
                qa_infrastructure,
                services,
            )
        ):
            return False
        for name in _PUBLIC_SERVICES + _QA_INFRASTRUCTURE:
            service = _yaml_mapping_child(services, name)
            if not isinstance(service, MappingNode):
                return False
            if name in _QA_INFRASTRUCTURE:
                expected_sources = (qa_infrastructure,)
            elif name.startswith("python-"):
                expected_sources = (qa_common, qa_python)
            else:
                expected_sources = (qa_common,)
            if not _yaml_merge_sources_are_exact(service, expected_sources):
                return False
        return True
    except Exception:
        return False


def _yaml_mapping_child(node: object, key: str) -> object:
    if not isinstance(node, MappingNode):
        return _INVALID
    matches = [
        value_node
        for key_node, value_node in node.value
        if isinstance(key_node, ScalarNode) and key_node.value == key
    ]
    if len(matches) != 1:
        return _INVALID
    return matches[0]


def _yaml_merge_sources_are_exact(
    service: MappingNode,
    expected_sources: tuple[object, ...],
) -> bool:
    merge_values = [
        value_node
        for key_node, value_node in service.value
        if isinstance(key_node, ScalarNode) and key_node.value == "<<"
    ]
    if len(merge_values) != 1:
        return False
    merge_value = merge_values[0]
    if len(expected_sources) == 1:
        return merge_value is expected_sources[0]
    return (
        isinstance(merge_value, SequenceNode)
        and len(merge_value.value) == len(expected_sources)
        and all(
            actual is expected
            for actual, expected in zip(
                merge_value.value,
                expected_sources,
                strict=True,
            )
        )
    )


def _qa_public_services_are_closed(services: object) -> bool:
    try:
        if not isinstance(services, Mapping):
            return False
        return all(
            services.get(name) == _expected_public_service(name)
            for name in _PUBLIC_SERVICES
        )
    except Exception:
        return False


def _expected_public_service(name: str) -> Mapping[str, object]:
    expected = dict(_qa_common_service_fields())
    expected["build"] = MappingProxyType(
        {
            "context": ".",
            "dockerfile": "testframework/docker/Dockerfile.qa",
            "target": _BUILD_TARGETS[name],
        }
    )
    expected["command"] = _FIXED_COMMANDS[name]
    if name.startswith("python-"):
        expected["environment"] = _qa_python_environment()
    if name in _INFRASTRUCTURE_DEPENDENTS:
        expected["depends_on"] = _healthy_infrastructure_dependencies()
    expected.update(_public_service_overrides(name))
    return MappingProxyType(expected)


def _public_service_overrides(name: str) -> Mapping[str, object]:
    overrides: dict[str, Mapping[str, object]] = {
        "backend-unit": MappingProxyType(
            {
                "volumes": (
                    _qa_reports_bind(),
                    _read_only_bind(
                        "./backend/cmd",
                        "/opt/hub/source/backend/cmd",
                    ),
                    _read_only_bind(
                        "./backend/internal",
                        "/opt/hub/source/backend/internal",
                    ),
                    _read_only_bind(
                        "./backend/tests",
                        "/opt/hub/source/backend/tests",
                    ),
                    _read_only_bind(
                        "./backend/ent/schema",
                        "/opt/hub/source/backend/ent/schema",
                    ),
                    _read_only_bind(
                        "./backend/migrations",
                        "/opt/hub/source/backend/migrations",
                    ),
                    _read_only_bind(
                        "./backend/testframework",
                        "/opt/hub/source/backend/testframework",
                    ),
                ),
                "environment": MappingProxyType(
                    {
                        "GOCACHE": "/var/cache/hub-go",
                        "GOTMPDIR": "/var/tmp/hub-go",
                    }
                ),
                "cpus": 6.0,
                "mem_limit": "10g",
                "tmpfs": (
                    "/tmp:rw,noexec,nosuid,nodev,size=64m",
                    "/var/cache/hub-go:rw,noexec,nosuid,nodev,"
                    "size=3g,uid=10001,gid=10001,mode=1770",
                    "/var/tmp/hub-go:rw,nosuid,nodev,exec,"
                    "size=4g,uid=10001,gid=10001,mode=1770",
                ),
            }
        ),
        "frontend-unit": MappingProxyType(
            {
                "volumes": (
                    _qa_reports_bind(),
                    _read_only_bind(
                        "./frontend/src",
                        "/opt/hub/source/frontend/src",
                    ),
                    _read_only_bind(
                        "./frontend/tests",
                        "/opt/hub/source/frontend/tests",
                    ),
                ),
                "cpus": 6.0,
                "mem_limit": "10g",
                "tmpfs": (
                    "/tmp:rw,noexec,nosuid,nodev,size=64m",
                    "/reports/frontend-unit/home:rw,noexec,nosuid,nodev,"
                    "size=512m,uid=10001,gid=10001,mode=0700",
                    "/reports/frontend-unit/tmp:rw,noexec,nosuid,nodev,"
                    "size=4g,uid=10001,gid=10001,mode=0700",
                ),
            }
        ),
        "python-contract": MappingProxyType(
            {
                "volumes": (
                    _qa_reports_bind(),
                    _read_only_bind(
                        "./testframework/src",
                        "/opt/hub/testframework/src",
                    ),
                    _read_only_bind(
                        "./testframework/selftests",
                        "/opt/hub/testframework/selftests",
                    ),
                    _read_only_bind(
                        "./docker-compose.local.yml",
                        "/opt/hub/docker-compose.local.yml",
                    ),
                    _read_only_bind(
                        "./.planning/phases/130-tf-traceability-gates/PLAN.md",
                        "/opt/hub/.planning/phases/"
                        "130-tf-traceability-gates/PLAN.md",
                    ),
                ),
            }
        ),
    }
    return overrides.get(name, MappingProxyType({}))


def _qa_reports_bind() -> Mapping[str, object]:
    return MappingProxyType(
        {
            "type": "bind",
            "source": "./testframework/reports",
            "target": "/reports",
        }
    )


def _read_only_bind(source: str, target: str) -> Mapping[str, object]:
    return MappingProxyType(
        {
            "type": "bind",
            "source": source,
            "target": target,
            "read_only": True,
        }
    )


def _qa_common_service_fields() -> Mapping[str, object]:
    return MappingProxyType(
        {
            "profiles": ("qa",),
            "read_only": True,
            "security_opt": ("no-new-privileges:true",),
            "cap_drop": ("ALL",),
            "init": True,
            "pids_limit": 256,
            "cpus": 2.0,
            "mem_limit": "2g",
            "volumes": ("./testframework/reports:/reports",),
            "networks": (_QA_NETWORK,),
            "stop_grace_period": "30s",
            "tmpfs": ("/tmp:rw,noexec,nosuid,nodev,size=64m",),
        }
    )


def _qa_python_environment() -> Mapping[str, object]:
    return MappingProxyType(
        {
            "PYTHONDONTWRITEBYTECODE": "1",
            "PYTEST_ADDOPTS": "-p no:cacheprovider",
            "PYTHONSAFEPATH": "1",
        }
    )


def _healthy_infrastructure_dependencies() -> Mapping[str, object]:
    return MappingProxyType(
        {
            name: MappingProxyType({"condition": "service_healthy"})
            for name in _QA_INFRASTRUCTURE
        }
    )


def _dependencies_are_closed(
    name: str,
    service: Mapping[str, object],
) -> bool:
    if name in _INFRASTRUCTURE_DEPENDENTS:
        return service.get("depends_on") == _healthy_infrastructure_dependencies()
    return "depends_on" not in service


def _qa_infrastructure_services_are_closed(services: object) -> bool:
    try:
        if not isinstance(services, Mapping):
            return False
        return all(
            services.get(name) == _expected_infrastructure_service(name)
            for name in _QA_INFRASTRUCTURE
        )
    except Exception:
        return False


def _expected_infrastructure_service(name: str) -> Mapping[str, object]:
    expected = dict(_qa_infrastructure_common_fields())
    expected.update(_infrastructure_specific_fields(name))
    return MappingProxyType(expected)


def _qa_infrastructure_common_fields() -> Mapping[str, object]:
    return MappingProxyType(
        {
            "profiles": ("qa",),
            "networks": (_QA_NETWORK,),
            "pids_limit": 128,
            "cpus": 1.0,
            "mem_limit": "1g",
            "stop_grace_period": "30s",
        }
    )


def _infrastructure_specific_fields(name: str) -> Mapping[str, object]:
    specifics: dict[str, Mapping[str, object]] = {
        "qa-postgres": MappingProxyType(
            {
                "image": "postgres:17",
                "environment": MappingProxyType(
                    {
                        "POSTGRES_USER": "hubqa",
                        "POSTGRES_PASSWORD": "hubqa-local-only",
                        "POSTGRES_DB": "hubqa",
                    }
                ),
                "healthcheck": _healthcheck(
                    ("CMD-SHELL", "pg_isready -U hubqa -d hubqa")
                ),
                "tmpfs": (
                    "/var/lib/postgresql/data:"
                    "rw,noexec,nosuid,nodev,size=512m",
                ),
            }
        ),
        "qa-mongo": MappingProxyType(
            {
                "image": "mongo:8.0",
                "environment": MappingProxyType(
                    {
                        "MONGO_INITDB_ROOT_USERNAME": "hubqa",
                        "MONGO_INITDB_ROOT_PASSWORD": "hubqa-local-only",
                    }
                ),
                "healthcheck": _healthcheck(
                    (
                        "CMD",
                        "mongosh",
                        "--quiet",
                        "--eval",
                        "db.adminCommand('ping').ok",
                    )
                ),
                "tmpfs": ("/data/db:rw,noexec,nosuid,nodev,size=512m",),
            }
        ),
        "qa-redis": MappingProxyType(
            {
                "image": "redis:7.4-alpine",
                "healthcheck": _healthcheck(("CMD", "redis-cli", "ping")),
                "tmpfs": ("/data:rw,noexec,nosuid,nodev,size=256m",),
            }
        ),
        "qa-nats": MappingProxyType(
            {
                "image": "nats:2.11",
                "command": ("-js", "--store_dir", "/data", "-m", "8222"),
                "healthcheck": _healthcheck(_NATS_READINESS_TEST),
                "tmpfs": ("/data:rw,noexec,nosuid,nodev,size=256m",),
            }
        ),
        "qa-seaweedfs": MappingProxyType(
            {
                "image": "chrislusf/seaweedfs:3.97",
                "command": (
                    "server",
                    "-s3",
                    "-dir=/data",
                    "-s3.config=/etc/seaweedfs/s3.json",
                ),
                "configs": (
                    MappingProxyType(
                        {
                            "source": "qa-seaweedfs-s3",
                            "target": "/etc/seaweedfs/s3.json",
                            "mode": "0444",
                        }
                    ),
                ),
                "healthcheck": _healthcheck(_SEAWEEDFS_READINESS_TEST),
                "tmpfs": ("/data:rw,noexec,nosuid,nodev,size=512m",),
            }
        ),
    }
    return specifics[name]


def _healthcheck(test: tuple[str, ...]) -> Mapping[str, object]:
    return MappingProxyType(
        {
            "test": test,
            "interval": "5s",
            "timeout": "3s",
            "retries": 20,
            "start_period": "5s",
        }
    )


def _mutate_service(
    services: Mapping[str, object],
    name: str,
    updates: Mapping[str, object],
) -> Mapping[str, object]:
    mutated_services = dict(services)
    service = dict(_mapping_value(services, name))
    service.update(updates)
    mutated_services[name] = MappingProxyType(service)
    return MappingProxyType(mutated_services)


def _mutate_nested_mapping(
    services: Mapping[str, object],
    name: str,
    outer_key: str,
    inner_key: str,
    value: object,
) -> Mapping[str, object]:
    service = _mapping_value(services, name)
    nested = dict(_mapping_value(service, outer_key))
    nested[inner_key] = value
    return _mutate_service(
        services,
        name,
        {outer_key: MappingProxyType(nested)},
    )


def _remove_service_field(
    services: Mapping[str, object],
    name: str,
    key: str,
) -> Mapping[str, object]:
    mutated_services = dict(services)
    service = dict(_mapping_value(services, name))
    service.pop(key, None)
    mutated_services[name] = MappingProxyType(service)
    return MappingProxyType(mutated_services)


def _resource_limits_are_bounded(service: Mapping[str, object]) -> bool:
    pids_limit = service.get("pids_limit")
    cpus = service.get("cpus")
    mem_limit = service.get("mem_limit")
    return (
        type(pids_limit) is int
        and pids_limit > 0
        and type(cpus) in {int, float}
        and math.isfinite(cpus)
        and cpus > 0
        and type(mem_limit) is str
        and bool(mem_limit.strip())
    )


def _has_reports_mount(service: Mapping[str, object]) -> bool:
    volumes = service.get("volumes")
    if type(volumes) is not tuple:
        return False
    matches = 0
    for volume in volumes:
        if isinstance(volume, Mapping):
            if volume.get("target") == "/reports" and volume.get("type") == "bind":
                matches += 1
        elif type(volume) is str and volume.rsplit(":", 1)[-1] == "/reports":
            matches += 1
    return matches == 1


def _storage_is_disposable(service: Mapping[str, object]) -> bool:
    tmpfs = service.get("tmpfs", ())
    volumes = service.get("volumes", ())
    if type(tmpfs) is not tuple or type(volumes) is not tuple:
        return False
    if not tmpfs and not volumes:
        return False
    for volume in volumes:
        if isinstance(volume, Mapping):
            if volume.get("type") == "tmpfs":
                continue
            if "source" in volume:
                return False
        elif type(volume) is str:
            if ":" in volume:
                return False
        else:
            return False
    return True


def _contains_token(value: object, token: str) -> bool:
    if type(value) is str:
        return token in value
    if isinstance(value, Mapping):
        return any(
            _contains_token(key, token) or _contains_token(child, token)
            for key, child in value.items()
        )
    if type(value) is tuple:
        return any(_contains_token(child, token) for child in value)
    return False


def _required_asset_text(path: Path) -> str:
    result = _read_bounded_regular(path)
    assert type(result) is str
    return result


def _dockerfile_qa_is_closed(text: object) -> bool:
    try:
        if type(text) is not str or not _docker_instructions_are_safe(text):
            return False
        stages = _docker_stages_safe(text)
        if not isinstance(stages, Mapping):
            return False
        expected_stages = (
            (
                "backend-unit",
                (
                    _GO_BASE,
                    (
                        _NON_ROOT_SETUP,
                        "WORKDIR /opt/hub/source/backend",
                        "COPY backend/go.mod backend/go.sum ./",
                        "RUN go mod download",
                        "COPY backend/ ./",
                        "RUN go generate ./ent",
                        "RUN test -f ent/ent.go -a -f ent/client.go -a -f "
                        "ent/migrate/schema.go",
                        "COPY testframework/docker/entrypoints/backend-unit.sh "
                        "/usr/local/bin/backend-unit",
                        "RUN sed -i 's/\\r$//' /usr/local/bin/backend-unit && "
                        "chmod 0555 /usr/local/bin/backend-unit",
                        "ENV GOCACHE=/reports/backend-unit/go-cache",
                        "ENV GOTMPDIR=/reports/backend-unit/tmp",
                        "ENV HOME=/tmp",
                        "USER 10001:10001",
                        'ENTRYPOINT ["/usr/local/bin/backend-unit"]',
                    ),
                ),
            ),
            (
                "frontend-unit",
                (
                    _FRONTEND_BASE,
                    (
                        _NON_ROOT_SETUP,
                        "WORKDIR /opt/hub/source/frontend",
                        "COPY frontend/package.json frontend/package-lock.json ./",
                        "RUN npm ci",
                        "COPY frontend/src/ ./src/",
                        "COPY frontend/tests/ ./tests/",
                        "RUN find /opt/hub/source/frontend/src "
                        "/opt/hub/source/frontend/tests -type f -exec sed -i 's/\\r$//' {} +",
                        "COPY frontend/public/ ./public/",
                        "COPY frontend/index.html frontend/biome.json "
                        "frontend/components.json frontend/tsconfig.json "
                        "frontend/vite.config.ts ./",
                        "COPY testframework/docker/vitest.frontend-unit.config.ts "
                        "/opt/hub/source/frontend/.qa/vitest.frontend-unit.config.ts",
                        "COPY testframework/docker/entrypoints/frontend-unit.sh "
                        "/usr/local/bin/frontend-unit",
                        "RUN sed -i 's/\\r$//' /usr/local/bin/frontend-unit && "
                        "chmod 0555 /usr/local/bin/frontend-unit",
                        "ENV HOME=/reports/frontend-unit/home",
                        "ENV TMPDIR=/reports/frontend-unit/tmp",
                        "USER 10001:10001",
                        'ENTRYPOINT ["/usr/local/bin/frontend-unit"]',
                    ),
                ),
            ),
            (
                "python-base",
                (
                    _PYTHON_BASE,
                    (
                        _NON_ROOT_SETUP,
                        "RUN python -m pip install --no-cache-dir uv==0.12.0",
                        "RUN mkdir -p /opt/hub/testframework /opt/hub/.venv && "
                        "chown -R 10001:10001 /opt/hub/testframework "
                        "/opt/hub/.venv",
                    ),
                ),
            ),
            (
                "python-tests",
                (
                    "python-base",
                    (
                        "WORKDIR /opt/hub/testframework",
                        "COPY testframework/pyproject.toml testframework/uv.lock "
                        "testframework/.python-version testframework/README.md ./",
                        "RUN UV_PROJECT_ENVIRONMENT=/opt/hub/.venv "
                        "uv sync --frozen --all-groups --no-install-project",
                        "COPY testframework/src/ ./src/",
                        "RUN UV_PROJECT_ENVIRONMENT=/opt/hub/.venv "
                        "uv sync --frozen --all-groups",
                        "COPY docker-compose.local.yml "
                        "/opt/hub/docker-compose.local.yml",
                        "COPY .planning/phases/130-tf-traceability-gates/PLAN.md "
                        "/opt/hub/.planning/phases/130-tf-traceability-gates/PLAN.md",
                        "COPY testframework/docker/ ./docker/",
                        "RUN find /opt/hub/docker-compose.local.yml "
                        "/opt/hub/.planning/phases/130-tf-traceability-gates/PLAN.md "
                        "/opt/hub/testframework/docker -type f -exec "
                        "sed -i 's/\\r$//' {} +",
                        "COPY testframework/selftests/ ./selftests/",
                        "COPY testframework/tests/ ./tests/",
                        "COPY testframework/cases/ ./cases/",
                        "COPY testframework/policies/ ./policies/",
                        "COPY testframework/schemas/ ./schemas/",
                        "COPY testframework/load/ ./load/",
                        "RUN find /opt/hub/testframework/cases "
                        "/opt/hub/testframework/policies "
                        "/opt/hub/testframework/schemas -type f -exec "
                        "sed -i 's/\\r$//' {} +",
                        "COPY testframework/docker/entrypoints/python-qa.sh "
                        "/usr/local/bin/python-qa",
                        "RUN sed -i 's/\\r$//' /usr/local/bin/python-qa && "
                        "chmod 0555 /usr/local/bin/python-qa && "
                        "chown -R 10001:10001 /opt/hub/testframework "
                        "/opt/hub/.venv",
                        "ENV PATH=/opt/hub/.venv/bin:$PATH",
                        "ENV PYTHONPATH=/opt/hub/testframework/src",
                        "ENV PYTHONSAFEPATH=1",
                        "USER 10001:10001",
                        'ENTRYPOINT ["/usr/local/bin/python-qa"]',
                    ),
                ),
            ),
            (
                "python-lock",
                (
                    "python-base",
                    (
                        "RUN mkdir -p /workspace/testframework "
                        "/opt/hub/uv-cache /opt/hub/lock-input "
                        "/home/hubqa && chown -R 10001:10001 "
                        "/workspace/testframework /opt/hub/uv-cache "
                        "/opt/hub/lock-input /home/hubqa",
                        "WORKDIR /workspace/testframework",
                        "COPY --chown=10001:10001 "
                        "testframework/pyproject.toml "
                        "/opt/hub/lock-input/pyproject.toml",
                        "ENV HOME=/home/hubqa",
                        "ENV UV_CACHE_DIR=/opt/hub/uv-cache",
                        "USER 10001:10001",
                        'ENTRYPOINT ["/usr/local/bin/uv", "lock"]',
                    ),
                ),
            ),
        )
        return tuple(stages.items()) == expected_stages
    except Exception:
        return False


def _docker_stages_safe(text: str) -> Mapping[str, object] | object:
    try:
        stages: dict[str, tuple[str, tuple[str, ...]]] = {}
        current_name: str | None = None
        current_base = ""
        current_instructions: list[str] = []
        for raw_line in text.splitlines():
            line = raw_line.strip()
            if not line or line.startswith("#"):
                continue
            if line.endswith("\\"):
                return _INVALID
            if line.upper().startswith("FROM "):
                if current_name is not None:
                    stages[current_name] = (
                        current_base,
                        tuple(current_instructions),
                    )
                parts = line.split()
                if len(parts) != 4 or parts[2].upper() != "AS":
                    return _INVALID
                current_base = parts[1]
                current_name = parts[3]
                if current_name in stages:
                    return _INVALID
                current_instructions = []
            elif current_name is None:
                return _INVALID
            else:
                current_instructions.append(line)
        if current_name is None:
            return _INVALID
        stages[current_name] = (current_base, tuple(current_instructions))
        return MappingProxyType(stages)
    except Exception:
        return _INVALID


def _docker_instructions_are_safe(text: object) -> bool:
    try:
        if type(text) is not str:
            return False
        lowered = text.casefold()
        if (
            "latest" in lowered
            or "docker.sock" in lowered
            or "curl" in lowered and "| sh" in lowered
            or "eval" in lowered
            or "sh -c" in lowered
        ):
            return False
        run_bans = (
            " go build ",
            " go test ",
            " go vet ",
            " npm test ",
            " npm run ",
            " pytest ",
            " vitest ",
            " vite ",
            " locust ",
            " biome ",
            " tsc ",
            " ruff ",
            " mypy ",
        )
        for raw_line in text.splitlines():
            line = " ".join(raw_line.casefold().split())
            if line.startswith("arg "):
                return False
            if line.startswith("copy "):
                parts = line.split()
                source_index = 1
                while source_index < len(parts) and parts[source_index].startswith("--"):
                    source_index += 1
                sources = parts[source_index:-1]
                if not sources or any(
                    source.startswith("/")
                    or ".." in Path(source).parts
                    or source.startswith(("http://", "https://"))
                    or "node_modules" in source
                    or ".venv" in source
                    or source.startswith("vendor")
                    for source in sources
                ):
                    return False
            if line.startswith("run ") and any(
                fragment in f" {line[4:]} " for fragment in run_bans
            ):
                return False
        return True
    except Exception:
        return False


def _backend_entrypoint_is_closed(text: object) -> bool:
    try:
        if type(text) is not str or not _shell_source_is_safe(text):
            return False
        return text == _canonical_fixed_unit_entrypoint(
            target="backend-unit",
            report_setup='mkdir -p "$report_dir" "$report_dir/go-cache" "$report_dir/tmp"',
            workdir="/opt/hub/source/backend",
            steps=(
                ("build", "go build ./..."),
                ("vet", "go vet -tags hubtest ./..."),
                (
                    "unit",
                    "go test -tags hubtest -json -count=1 -timeout=150s "
                    "./tests/...",
                ),
                (
                    "race",
                    "go test -tags hubtest -json -race -count=1 "
                    "-timeout=150s ./tests/...",
                ),
            ),
        )
    except Exception:
        return False


def _frontend_entrypoint_is_closed(text: object) -> bool:
    try:
        if type(text) is not str or not _shell_source_is_safe(text):
            return False
        return text == _canonical_fixed_unit_entrypoint(
            target="frontend-unit",
            report_setup=(
                'mkdir -p "$report_dir" "$report_dir/home" "$report_dir/tmp"'
            ),
            workdir="/opt/hub/source/frontend",
            steps=(
                (
                    "typescript",
                    "/usr/bin/node /opt/hub/source/frontend/node_modules/"
                    "typescript/bin/tsc -p /opt/hub/source/frontend/"
                    "tsconfig.json --noEmit",
                ),
                (
                    "vitest",
                    "/usr/bin/node /opt/hub/source/frontend/node_modules/"
                    "vitest/"
                    "vitest.mjs run --config /opt/hub/source/frontend/.qa/"
                    "vitest.frontend-unit.config.ts --configLoader runner",
                ),
            ),
        )
    except Exception:
        return False


def _canonical_fixed_unit_entrypoint(
    *,
    target: str,
    report_setup: str,
    workdir: str,
    steps: tuple[tuple[str, str], ...],
) -> str:
    lines = [
        "#!/bin/sh",
        "set -eu",
        "",
        '[ "$#" -eq 0 ] || exit 64',
        "",
        f'report_dir="/reports/{target}"',
        'result_path="$report_dir/result.json"',
        "first_exit=0",
        "",
        "record_failure() {",
        '  code="$1"',
        '  if [ "$first_exit" -eq 0 ]; then',
        '    first_exit="$code"',
        "  fi",
        "}",
        "",
        "write_step() {",
        '  step="$1"',
        '  status="$2"',
        '  code="$3"',
        "  printf '{\"step\":\"%s\",\"status\":\"%s\",\"exit\":%s}\\n' \\",
        '    "$step" "$status" "$code" > "$report_dir/$step.result.json"',
        "}",
        "",
        report_setup,
        f"cd {workdir}",
        "",
    ]
    for step, command in steps:
        lines.extend(
            (
                f"if {command}; then",
                f"  write_step {step} PASS 0",
                "else",
                "  code=$?",
                '  record_failure "$code"',
                f'  write_step {step} FAIL "$code"',
                f'fi > "$report_dir/{step}.stdout.log" '
                f'2> "$report_dir/{step}.stderr.log"',
                "",
            )
        )
    lines.extend(
        (
            'if [ "$first_exit" -eq 0 ]; then',
            "  status=PASS",
            "else",
            "  status=FAIL",
            "fi",
            f"printf '{{\"target\":\"{target}\",\"status\":\"%s\","
            "\"first_exit\":%s}\\n' \\",
            '  "$status" "$first_exit" > "$result_path"',
            'exit "$first_exit"',
        )
    )
    return "\n".join(lines) + "\n"


def _shell_source_is_safe(text: object) -> bool:
    try:
        if type(text) is not str:
            return False
        lowered = text.casefold()
        return not any(
            forbidden in lowered
            for forbidden in (
                "eval",
                "sh -c",
                "$@",
                "printenv",
                "docker.sock",
            )
        )
    except Exception:
        return False


def _python_dispatcher_is_closed(text: object) -> bool:
    try:
        if type(text) is not str or not _shell_source_is_safe(text):
            return False
        return tuple(line.strip() for line in text.splitlines() if line.strip()) == (
            "#!/bin/sh",
            "set -eu",
            '[ "$#" -eq 1 ] || exit 64',
            "cd /opt/hub/testframework",
            "export PYTHONSAFEPATH=1",
            'case "$1" in',
            "contract)",
            "exec /opt/hub/.venv/bin/pytest selftests --strict-markers",
            ";;",
            "api)",
            "exec /opt/hub/.venv/bin/pytest -p "
            "hub_testkit.pytest_plugin.plugin tests/api "
            "--hub-case-root cases "
            "--hub-case-schema schemas/test-case.schema.json "
            "--hub-marker-policy policies/markers.toml "
            "--hub-redaction-policy policies/redaction.toml --strict-markers",
            ";;",
            "integration)",
            "exec /opt/hub/.venv/bin/pytest -p "
            "hub_testkit.pytest_plugin.plugin tests/integration "
            "--hub-case-root cases "
            "--hub-case-schema schemas/test-case.schema.json "
            "--hub-marker-policy policies/markers.toml "
            "--hub-redaction-policy policies/redaction.toml --strict-markers",
            ";;",
            "e2e-smoke)",
            "exec /opt/hub/.venv/bin/pytest -p "
            "hub_testkit.pytest_plugin.plugin tests/e2e -m smoke "
            "--hub-case-root cases "
            "--hub-case-schema schemas/test-case.schema.json "
            "--hub-marker-policy policies/markers.toml "
            "--hub-redaction-policy policies/redaction.toml --strict-markers",
            ";;",
            "e2e-full)",
            "exec /opt/hub/.venv/bin/pytest -p "
            "hub_testkit.pytest_plugin.plugin tests/e2e "
            "--hub-case-root cases "
            "--hub-case-schema schemas/test-case.schema.json "
            "--hub-marker-policy policies/markers.toml "
            "--hub-redaction-policy policies/redaction.toml --strict-markers",
            ";;",
            "load)",
            "exec /opt/hub/.venv/bin/locust -f load/locustfile.py --headless",
            ";;",
            "*)",
            "exit 64",
            ";;",
            "esac",
        )
    except Exception:
        return False


def _empty_regular_marker(path: Path) -> bool:
    descriptor = -1
    try:
        before = path.lstat()
        reparse_flag = getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", 0)
        attributes = getattr(before, "st_file_attributes", 0)
        if (
            not stat.S_ISREG(before.st_mode)
            or attributes & reparse_flag
            or before.st_size != 0
        ):
            return False
        flags = os.O_RDONLY | getattr(os, "O_BINARY", 0)
        flags |= getattr(os, "O_NOFOLLOW", 0)
        descriptor = os.open(path, flags)
        opened = os.fstat(descriptor)
        if not _descriptor_metadata_is_stable(
            _descriptor_metadata(before),
            _descriptor_metadata(opened),
        ):
            return False
        if os.read(descriptor, 1) != b"":
            return False
        return _descriptor_metadata_is_stable(
            _descriptor_metadata(opened),
            _descriptor_metadata(os.fstat(descriptor)),
        )
    except Exception:
        return False
    finally:
        if descriptor >= 0:
            try:
                os.close(descriptor)
            except OSError:
                pass


def _docker_regular_files(repo_root: Path) -> tuple[str, ...] | object:
    root = repo_root / "testframework" / "docker"
    try:
        root_stat = root.lstat()
        if not stat.S_ISDIR(root_stat.st_mode) or root.is_symlink():
            return _INVALID
        found: list[str] = []
        for directory, directory_names, file_names in os.walk(root, followlinks=False):
            directory_path = Path(directory)
            for name in tuple(directory_names):
                child = directory_path / name
                child_stat = child.lstat()
                if not stat.S_ISDIR(child_stat.st_mode) or child.is_symlink():
                    return _INVALID
            for name in file_names:
                child = directory_path / name
                child_stat = child.lstat()
                if not stat.S_ISREG(child_stat.st_mode) or child.is_symlink():
                    return _INVALID
                found.append(child.relative_to(repo_root).as_posix())
        return tuple(sorted(found))
    except Exception:
        return _INVALID


def _retired_paths_are_absent(repo_root: Path, retired_paths: object) -> bool:
    try:
        if type(retired_paths) is not tuple:
            return False
        return all(
            type(path) is str and not os.path.lexists(repo_root / path)
            for path in retired_paths
        )
    except Exception:
        return False


def _active_authority_is_clean(repo_root: Path, tokens: object) -> bool:
    try:
        if type(tokens) is not tuple or any(type(token) is not str for token in tokens):
            return False
        candidates = _active_authority_paths(repo_root)
        if not isinstance(candidates, tuple):
            return False
        for path in candidates:
            text = _read_bounded_regular(path)
            if type(text) is not str:
                return False
            if any(token in text for token in tokens):
                return False
        return True
    except Exception:
        return False


def _active_authority_paths(repo_root: Path) -> tuple[Path, ...] | object:
    try:
        candidates: set[Path] = set()
        for relative in (
            Path("AGENTS.md"),
            Path("README.md"),
            Path("docker-compose.local.yml"),
        ):
            path = repo_root / relative
            if os.path.lexists(path):
                candidates.add(path)
        scan_roots = (
            (repo_root / ".github" / "workflows", None, ()),
            (repo_root / ".cursor" / "skills", None, ()),
            (
                repo_root / ".planning",
                ".md",
                (repo_root / ".planning" / "archive",),
            ),
            (
                repo_root / "docs",
                ".md",
                (repo_root / "docs" / "superpowers" / "plans",),
            ),
            (repo_root / "testframework" / "docker", None, ()),
            (repo_root / "testframework" / "runner", None, ()),
        )
        for root, suffix, excluded_directories in scan_roots:
            discovered = _regular_files_below(root, suffix, excluded_directories)
            if not isinstance(discovered, tuple):
                return _INVALID
            candidates.update(discovered)
        testframework_docs = _regular_files_below(
            repo_root / "testframework",
            ".md",
            (repo_root / "testframework" / "reports",),
        )
        if not isinstance(testframework_docs, tuple):
            return _INVALID
        candidates.update(testframework_docs)
        filtered = tuple(
            sorted(
                path
                for path in candidates
                if not _authority_path_is_excluded(path, repo_root)
            )
        )
        return filtered
    except Exception:
        return _INVALID


def _regular_files_below(
    root: Path,
    suffix: str | None,
    excluded_directories: tuple[Path, ...],
) -> tuple[Path, ...] | object:
    if not os.path.lexists(root):
        return ()
    try:
        root_stat = root.lstat()
        if not stat.S_ISDIR(root_stat.st_mode) or root.is_symlink():
            return _INVALID
        found: list[Path] = []
        for directory, directory_names, file_names in os.walk(root, followlinks=False):
            directory_path = Path(directory)
            for name in tuple(directory_names):
                child = directory_path / name
                if child in excluded_directories:
                    directory_names.remove(name)
                    continue
                child_stat = child.lstat()
                if not stat.S_ISDIR(child_stat.st_mode) or child.is_symlink():
                    return _INVALID
            for name in file_names:
                child = directory_path / name
                child_stat = child.lstat()
                if not stat.S_ISREG(child_stat.st_mode) or child.is_symlink():
                    return _INVALID
                if suffix is None or child.suffix == suffix:
                    found.append(child)
        return tuple(found)
    except Exception:
        return _INVALID


def _authority_path_is_excluded(path: Path, repo_root: Path) -> bool:
    relative = path.relative_to(repo_root).as_posix()
    return (
        relative.startswith(".planning/archive/")
        or relative.startswith("docs/superpowers/plans/")
        or relative
        == "docs/superpowers/specs/2026-08-15-docker-compose-"
        "testframework-consolidation-design.md"
        or relative == _POLICY_RELATIVE_PATH.as_posix()
        or relative
        == "testframework/selftests/architecture/docker/"
        "test_root_compose_contract.py"
    )


def _policy_text(repo_root: Path) -> str:
    result = _read_bounded_regular(repo_root / _POLICY_RELATIVE_PATH)
    assert type(result) is str
    return result


def _policy_document() -> dict[str, object]:
    return {
        "schema": "hub.qa.root-compose-testframework/v1",
        "root_compose": "docker-compose.local.yml",
        "public_services": list(_PUBLIC_SERVICES),
        "retained_docker_paths": list(_RETAINED_DOCKER_PATHS),
        "retired_paths": list(_RETIRED_PATHS),
        "forbidden_tokens": list(_FORBIDDEN_TOKENS),
    }


def _canonical_policy_text() -> str:
    return json.dumps(
        _policy_document(),
        ensure_ascii=False,
        indent=2,
    ) + "\n"


def _write_valid_layout(repo_root: Path) -> None:
    policy_path = repo_root / _POLICY_RELATIVE_PATH
    policy_path.parent.mkdir(parents=True, exist_ok=True)
    _write_json(policy_path, _policy_document())
    (repo_root / "docker-compose.local.yml").write_text(
        _MINIMAL_VALID_COMPOSE,
        encoding="utf-8",
        newline="\n",
    )


def _write_json(path: Path, document: Mapping[str, object]) -> None:
    path.write_text(
        json.dumps(document, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
        newline="\n",
    )


def _assert_generic_failure(repo_root: Path) -> None:
    error = _capture_failure_safe(repo_root)
    del repo_root
    assert type(error) is RootComposeContractError
    rendered = "".join(
        traceback.TracebackException.from_exception(
            error,
            capture_locals=True,
        ).format()
    )
    assert str(error) == "root compose contract is invalid"
    assert repr(error) == "RootComposeContractError('root compose contract is invalid')"
    assert error.__cause__ is None
    assert error.__context__ is None
    assert error.__suppress_context__ is True
    assert _CANARY not in rendered
    assert _CANARY not in str(error)
    assert _CANARY not in repr(error)


def _capture_failure_safe(repo_root: Path) -> RootComposeContractError | object:
    try:
        load_root_compose_contract(repo_root)
    except RootComposeContractError as error:
        del repo_root
        return error
    del repo_root
    return _INVALID


def _raise_invalid() -> NoReturn:
    raise RootComposeContractError("root compose contract is invalid") from None
