from __future__ import annotations

import ast
import glob
import hashlib
import json
import math
import os
import re
import shlex
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
    "hard2g_observer",
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
    "testframework/docker/backend_unit/execute.sh",
    "testframework/docker/backend_unit/plan.sh",
    "testframework/docker/backend_unit/receipt.sh",
    "testframework/docker/backend_unit/runtime.sh",
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
_HARD2G_IMAGE_INSPECT_ARGVS = (
    (
        "docker",
        "image",
        "inspect",
        "--format",
        "{{json .}}",
        "worktree-master-python-contract",
    ),
    (
        "docker",
        "image",
        "inspect",
        "--format",
        "{{json .}}",
        "worktree-master-backend-unit",
    ),
)
_HARD2G_BUILD_INPUT_PATHS = (
    "docker-compose.local.yml",
    "testframework/docker/Dockerfile.qa",
    "testframework/docker/entrypoints/backend-unit.sh",
    "testframework/docker/backend_unit/runtime.sh",
    "testframework/docker/backend_unit/plan.sh",
    "testframework/docker/backend_unit/execute.sh",
    "testframework/docker/backend_unit/receipt.sh",
    "testframework/observers/**",
    "testframework/schemas/backend-unit-package-plan.schema.v1.json",
    "testframework/schemas/backend-unit-container-receipt.schema.v1.json",
    "testframework/schemas/backend-unit-observer-receipt.schema.v1.json",
    "testframework/policies/backend-unit-observer.v1.json",
    "testframework/policies/backend-unit-phase150-08c2-subject.v1.json",
    "testframework/policies/root-compose-testframework.v1.json",
    "testframework/selftests/architecture/docker/**",
    "testframework/selftests/architecture/observers/**",
)
_HARD2G_OBSERVER_CONFIG_KEYS = (
    "module",
    "python_stage",
    "source_paths",
    "selftest_paths",
    "policy_path",
    "image_inspect_argvs",
    "canonical_build_input_paths",
    "snapshot_assertion_source",
    "fixture_path",
)
_HARD2G_OBSERVER_MODULE = "observers.backend_unit"
_HARD2G_OBSERVER_POLICY_PATH = (
    "testframework/policies/backend-unit-observer.v1.json"
)
_HARD2G_OBSERVER_SOURCE_PATHS = (
    "testframework/observers/__init__.py",
    "testframework/observers/backend_unit/__init__.py",
    "testframework/observers/backend_unit/contracts.py",
    "testframework/observers/backend_unit/safe_paths.py",
    "testframework/observers/backend_unit/lease.py",
    "testframework/observers/backend_unit/resources.py",
    "testframework/observers/backend_unit/docker_adapter.py",
    "testframework/observers/backend_unit/publication.py",
    "testframework/observers/backend_unit/observe.py",
)
_HARD2G_OBSERVER_SAFE_PATHS_SOURCE_PATH = (
    "testframework/observers/backend_unit/safe_paths.py"
)
_HARD2G_OBSERVER_SELFTEST_PATHS = (
    "testframework/selftests/architecture/observers/test_backend_unit_receipts.py",
    "testframework/selftests/architecture/observers/"
    "test_backend_unit_receipts_review_contract.py",
    "testframework/selftests/architecture/observers/test_backend_unit_observer.py",
)
_HARD2G_OBSERVER_SNAPSHOT_ASSERTION_SOURCE = (
    _HARD2G_OBSERVER_SELFTEST_PATHS[-1]
)
_HARD2G_OBSERVER_FIXTURE_PATH = (
    "testframework/selftests/architecture/observers/fixtures/backend_unit/valid/"
    "observer-receipt.json"
)
_HARD2G_METADATA_README_PATHS = frozenset(
    {
        "testframework/README.md",
        "testframework/schemas/README.md",
        "testframework/docker/README.md",
    }
)
_HARD2G_REPORT_BUDGET = MappingProxyType(
    {
        "total_bytes": 64 * 1024 * 1024,
        "reserved_metadata_bytes": 8 * 1024 * 1024,
        "streams": MappingProxyType(
            {
                "build": MappingProxyType(
                    {"stdout": 2 * 1024 * 1024, "stderr": 2 * 1024 * 1024}
                ),
                "vet": MappingProxyType(
                    {"stdout": 2 * 1024 * 1024, "stderr": 2 * 1024 * 1024}
                ),
                "unit": MappingProxyType(
                    {"stdout": 16 * 1024 * 1024, "stderr": 4 * 1024 * 1024}
                ),
                "race": MappingProxyType(
                    {"stdout": 24 * 1024 * 1024, "stderr": 4 * 1024 * 1024}
                ),
            }
        ),
    }
)
_HARD2G_C02_RUNTIME_OWNED_HELPERS = (
    "start_owned_go_child_group",
    "owned_go_child_is_live",
    "monitor_owned_budget_stop",
    "terminate_owned_go_child_group",
    "reap_owned_go_child_group",
)
_HARD2G_C02_EXECUTE_EXTENSION_ALLOWLIST: tuple[str, ...] = ()
_HARD2G_RECEIPT_BUDGET_BUILDER_BODY = (
    'case "${report_budget_truncated_due_to_budget-}" in',
    "false)",
    (
        "report_budget_json='"
        '{"limit_bytes":67108864,"metadata_reserve_bytes":8388608,'
        '"truncated_due_to_budget":false}'
        "'"
    ),
    ";;",
    "true)",
    (
        "report_budget_json='"
        '{"limit_bytes":67108864,"metadata_reserve_bytes":8388608,'
        '"truncated_due_to_budget":true}'
        "'"
    ),
    ";;",
    "*)",
    "return 64",
    ";;",
    "esac",
    'case "${receipt-}" in',
    "*'\"report_budget\"'*)",
    "return 64",
    ";;",
    "\\{*\\})",
    ";;",
    "*)",
    "return 64",
    ";;",
    "esac",
    r'receipt="${receipt%?},\"report_budget\":$report_budget_json}"',
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
_COMPOSE_SERVICE_NAMES = (
    "postgres",
    "mongo",
    "redis",
    "nats",
    "seaweedfs",
    "livekit",
    "meilisearch",
    "hub-mail",
    "backend",
    "frontend",
    *_PUBLIC_SERVICES,
    _MAINTENANCE_SERVICE,
    *_QA_INFRASTRUCTURE,
)
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
    assert _hard2g_observer_config_is_closed(policy["hard2g_observer"])
    assert _policy_text(_REPO_ROOT) == _canonical_policy_text()
    assert tuple(name for name in services if name in _PUBLIC_SERVICES) == (
        _PUBLIC_SERVICES
    )
    assert _compose_service_inventory_is_closed(services)
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


def test_root_compose_rejects_every_additional_service() -> None:
    services = _compose_services(load_root_compose_contract(_REPO_ROOT))
    assert _compose_service_inventory_is_closed(services)
    for value in (
        MappingProxyType({"image": "hostile:latest"}),
        MappingProxyType({"profiles": ("aux",), "privileged": True}),
    ):
        hostile = dict(services)
        hostile["unbounded-extra"] = value
        assert not _compose_service_inventory_is_closed(
            MappingProxyType(hostile)
        )


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
                    backend_cache.replace("size=768m", "size=767m", 1),
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


def test_hard2g_backend_unit_compose_docker_and_runner_contract() -> None:
    services = _compose_services(load_root_compose_contract(_REPO_ROOT))
    backend = _mapping_value(services, "backend-unit")
    dockerfile = _required_asset_text(
        _REPO_ROOT / "testframework" / "docker" / "Dockerfile.qa"
    )
    entrypoint = _required_asset_text(
        _REPO_ROOT
        / "testframework"
        / "docker"
        / "entrypoints"
        / "backend-unit.sh"
    )

    assert _hard2g_backend_unit_service_is_closed(backend)
    assert _dockerfile_qa_is_closed(dockerfile)
    assert _backend_entrypoint_is_closed(entrypoint)
    assert _hard2g_runner_helpers_are_closed(_REPO_ROOT)
    assert _hard2g_readme_metadata_copies_are_closed(dockerfile)


def test_hard2g_backend_unit_compose_contract_rejects_authority_mutations() -> None:
    backend = _expected_public_service("backend-unit")
    volumes = backend["volumes"]
    environment = backend["environment"]
    tmpfs = backend["tmpfs"]
    assert type(volumes) is tuple and len(volumes) == 7
    assert isinstance(environment, Mapping)
    assert type(tmpfs) is tuple and len(tmpfs) == 3
    cache_tmpfs = tmpfs[1]
    assert type(cache_tmpfs) is str
    assert _hard2g_backend_unit_service_is_closed(backend)

    def mutated(**updates: object) -> Mapping[str, object]:
        candidate = dict(backend)
        candidate.update(updates)
        return MappingProxyType(candidate)

    writable_source = dict(volumes[1])
    writable_source["read_only"] = False
    hostile_environment = dict(environment)
    hostile_environment["GOPROXY"] = "https://proxy.invalid"
    broad_root = _read_only_bind("./backend", "/opt/hub/source/backend")
    docs_root = _read_only_bind(
        "./backend/docs",
        "/opt/hub/source/backend/docs",
    )
    observer_root = _read_only_bind(
        "./testframework/observers",
        "/opt/hub/testframework/observers",
    )
    host_launcher = _read_only_bind(
        "./testframework/docker/entrypoints/backend-unit.sh",
        "/usr/local/bin/backend-unit",
    )
    socket_bind = MappingProxyType(
        {
            "type": "bind",
            "source": "/var/run/docker.sock",
            "target": "/var/run/docker.sock",
        }
    )
    extra_writer = MappingProxyType(
        {
            "type": "bind",
            "source": "./testframework/reports-extra",
            "target": "/reports-extra",
        }
    )
    mutations = (
        {"command": ("--build",)},
        {"cpus": 2.0},
        {"mem_limit": "10g"},
        {"memswap_limit": "10g"},
        {"environment": MappingProxyType(hostile_environment)},
        {"tmpfs": (tmpfs[0], cache_tmpfs.replace("768m", "767m"), tmpfs[2])},
        {"volumes": (volumes[0], broad_root)},
        {"volumes": (*volumes, docs_root)},
        {"volumes": (*volumes, observer_root)},
        {"volumes": (*volumes, host_launcher)},
        {"volumes": (*volumes, socket_bind)},
        {"volumes": (*volumes, extra_writer)},
        {
            "volumes": (
                volumes[0],
                MappingProxyType(writable_source),
                *volumes[2:],
            )
        },
        {"privileged": True},
    )
    for updates in mutations:
        assert not _hard2g_backend_unit_service_is_closed(mutated(**updates))

    admitted = MappingProxyType(
        {
            name: MappingProxyType({"profiles": ("qa",)})
            for name in _PUBLIC_SERVICES + _QA_INFRASTRUCTURE
        }
    )
    assert _qa_service_profiles_are_closed(admitted)
    added_service = dict(admitted)
    added_service["observer"] = MappingProxyType({"profiles": ("qa",)})
    assert not _qa_service_profiles_are_closed(MappingProxyType(added_service))


def test_hard2g_task2_provenance_oracle_rejects_role_blind_binding() -> None:
    source = _required_asset_text(
        _REPO_ROOT / _HARD2G_OBSERVER_SNAPSHOT_ASSERTION_SOURCE
    )
    assert _hard2g_task2_provenance_oracle_chain_is_closed(source)
    mutated = source.replace(
        '        == observer["container_image"]\n',
        '        == observer["fixed_image_ids"]["backend_unit"]\n',
        1,
    )
    assert mutated != source
    assert not _hard2g_task2_provenance_oracle_chain_is_closed(mutated)


def test_hard2g_task2_provenance_oracle_requires_live_publish_binding() -> None:
    source = _required_asset_text(
        _REPO_ROOT / _HARD2G_OBSERVER_SNAPSHOT_ASSERTION_SOURCE
    )
    assert _hard2g_task2_provenance_oracle_chain_is_closed(source)
    hostile_sources = (
        source.replace(
            "    runtime = _RawObserverRuntime("
            "raw_observations=raw_observations)\n",
            "    runtime = _RawObserverRuntime()\n",
            1,
        ),
        source.replace(
            "    runtime = _RawObserverRuntime("
            "raw_observations=raw_observations)\n",
            "    runtime = _RawObserverRuntime("
            "raw_observations=raw_observations)\n"
            "    runtime = fake\n",
            1,
        ),
        source.replace(
            "    runtime = _RawObserverRuntime("
            "raw_observations=raw_observations)\n",
            "    runtime = _RawObserverRuntime("
            "raw_observations=raw_observations)\n"
            "    runtime.published = _publication_payloads()\n",
            1,
        ),
        source.replace(
            "    final, observer = runtime.published\n",
            "    runtime.__dict__.update({\"published\": _publication_payloads()})\n"
            "    final, observer = runtime.published\n",
            1,
        ),
        source.replace(
            "    final, observer = runtime.published\n",
            "    final, observer = _publication_payloads()\n",
            1,
        ),
        source.replace(
            "    final, observer = runtime.published\n",
            "    runtime = fake\n"
            "    final, observer = runtime.published\n",
            1,
        ),
        source.replace(
            "    final, observer = runtime.published\n",
            "    runtime.published = _publication_payloads()\n"
            "    final, observer = runtime.published\n",
            1,
        ),
        source.replace(
            '    assert final["classification"] == expected\n',
            "    return\n"
            '    assert final["classification"] == expected\n',
            1,
        ),
        source.replace(
            '    assert final["classification"] == expected\n',
            '    final = {"backend_image_id": _BACKEND_IMAGE_ID}\n'
            '    assert final["classification"] == expected\n',
            1,
        ),
        source.replace(
            '    assert final["classification"] == expected\n',
            '    observer["container_image"] = _BACKEND_IMAGE_ID\n'
            '    assert final["classification"] == expected\n',
            1,
        ),
        source.replace(
            '    assert final["classification"] == expected\n',
            '    observer.update({"container_image": _BACKEND_IMAGE_ID})\n'
            '    assert final["classification"] == expected\n',
            1,
        ),
    )
    assert all(hostile != source for hostile in hostile_sources)
    assert all(
        not _hard2g_task2_provenance_oracle_chain_is_closed(hostile)
        for hostile in hostile_sources
    )


def test_hard2g_runner_rejects_permissive_authority() -> None:
    canonical_entrypoint = "\n".join(
        (
            "#!/bin/sh",
            "set -eu",
            '[ "$#" -eq 0 ] || exit 64',
            "unset GOMAXPROCS",
            ". /usr/local/lib/hub/backend-unit/runtime.sh",
            ". /usr/local/lib/hub/backend-unit/plan.sh",
            ". /usr/local/lib/hub/backend-unit/execute.sh",
            ". /usr/local/lib/hub/backend-unit/receipt.sh",
            "backend_unit_main",
            "",
        )
    )
    assert _hard2g_backend_entrypoint_is_closed(canonical_entrypoint)
    for hostile in (
        canonical_entrypoint.replace("#!/bin/sh", "#!/bin/sh\n", 1),
        canonical_entrypoint.replace("set -eu", "set -e", 1),
        canonical_entrypoint.replace(
            '[ "$#" -eq 0 ] || exit 64',
            '[ "$#" -le 0 ] || exit 64',
            1,
        ),
        canonical_entrypoint.replace("unset GOMAXPROCS", "GOMAXPROCS=1", 1),
        canonical_entrypoint.replace(
            ". /usr/local/lib/hub/backend-unit/plan.sh",
            ". /usr/local/lib/hub/backend-unit/receipt.sh",
            1,
        ),
        canonical_entrypoint.replace(
            "backend_unit_main", "backend_unit_main --help", 1
        ),
        canonical_entrypoint.replace(
            "backend_unit_main", "exec backend_unit_main", 1
        ),
        canonical_entrypoint.replace(
            "backend_unit_main", "eval backend_unit_main", 1
        ),
        canonical_entrypoint.replace(
            "backend_unit_main", 'command "$HUB_QA_COMMAND"', 1
        ),
        canonical_entrypoint.replace(
            "backend_unit_main", "docker compose run backend-unit", 1
        ),
        canonical_entrypoint.replace(
            "backend_unit_main", "touch /reports/SEAL", 1
        ),
        canonical_entrypoint.replace(
            "backend_unit_main\n",
            "true\nbackend_unit_main\n",
            1,
        ),
    ):
        assert not _hard2g_backend_entrypoint_is_closed(hostile)


def test_hard2g_runtime_budget_contract_is_closed() -> None:
    assignments = _hard2g_runtime_budget_assignments()
    assert type(assignments) is tuple
    canonical_runtime = _hard2g_runtime_budget_fixture(assignments)
    assert _hard2g_runtime_budget_contract_is_closed(canonical_runtime)

    report_symbol, _, _ = assignments[0].partition("=")
    metadata_assignment = assignments[1]
    cap_symbol, _, cap_value = assignments[2].partition("=")
    report_check = '  [ "$report_bytes" -le "$REPORT_BUDGET_BYTES" ] || return 75\n'
    metadata_check = (
        '  [ "$metadata_bytes" -le "$REPORT_METADATA_BUDGET_BYTES" ] '
        "|| return 75\n"
    )
    for hostile in (
        canonical_runtime.replace(
            assignments[0], f"{report_symbol}=1", 1
        ),
        canonical_runtime.replace(f"{metadata_assignment}\n", "", 1),
        canonical_runtime.replace(f'"${cap_symbol}"', f'"{cap_value}"', 1),
        canonical_runtime.replace(report_check, "", 1),
        canonical_runtime.replace(
            metadata_check,
            metadata_check + "  return 0\n",
            1,
        ),
    ):
        assert not _hard2g_runtime_budget_contract_is_closed(hostile)


def test_hard2g_execute_budget_bridge_is_closed() -> None:
    canonical_execute = _hard2g_c02_durable_budget_execute_fixture()
    assert _hard2g_execute_budget_bridge_is_closed(canonical_execute)

    for hostile in (
        canonical_execute.replace(
            'stream_tmp="$GOTMPDIR/stream-capture/$stage.$stream.$package_ordinal.tmp"',
            'stream_tmp="$stream_path.raw"',
            1,
        ),
        canonical_execute.replace(
            'read_limit=$((remaining + 1))',
            'read_limit=$((stream_cap + 1))',
            1,
        ),
        canonical_execute.replace(
            'head -c "$read_limit" < "$fifo_path" > "$stream_tmp"',
            'while IFS= read -r chunk; do printf "%s\n" "$chunk"; done',
            1,
        ),
        canonical_execute.replace(
            '  if check_budget; then\n'
            '    cat -- "$stream_tmp" >> "$stream_path"',
            '  cat -- "$stream_tmp" >> "$stream_path"\n'
            '  if check_budget; then\n    :',
            1,
        ),
        canonical_execute.replace(
            '    read_bounded_stage_stream "$stage" stdout '
            '"$stdout_fifo" "$stdout_path" "$stdout_used_path" '
            '"$package_ordinal" &',
            '    :',
            1,
        ),
        canonical_execute.replace(
            '    if monitor_owned_budget_stop "$go_child_pid" '
            '"$budget_stop_marker"; then',
            '    if :; then',
            1,
        ),
        canonical_execute.replace(
            '      terminate_owned_go_child_group\n'
            '      reap_owned_go_child_group\n'
            '      wait "$stdout_reader_pid"',
            '      wait "$go_child_pid" || child_exit=$?\n'
            '      terminate_owned_go_child_group\n'
            '      reap_owned_go_child_group\n'
            '      wait "$stdout_reader_pid"',
            1,
        ),
        canonical_execute.replace(
            '    if check_budget; then\n      :\n    else\n'
            '      report_budget_truncated_due_to_budget=true\n'
            '      terminate_owned_go_child_group\n'
            '      reap_owned_go_child_group\n'
            '      return 75\n    fi\n'
            '    wait "$go_child_pid"',
            '    wait "$go_child_pid"',
            1,
        ),
        canonical_execute.replace(
            "report_budget_truncated_due_to_budget=true",
            "truncation_evidence=budget-stop",
            1,
        ),
    ):
        assert hostile != canonical_execute
        assert not _hard2g_execute_budget_bridge_is_closed(hostile)


def test_hard2g_read_bounded_stage_stream_automaton_is_closed() -> None:
    execute = _hard2g_c02_durable_budget_execute_fixture()
    assert _hard2g_read_bounded_stage_stream_automaton_is_closed(execute)

    append_line = '    cat -- "$stream_tmp" >> "$stream_path"'
    overflow_block = (
        '  if [ "$stream_bytes" -gt "$remaining" ]; then\n'
        '    : > "$budget_stop_marker"\n'
        '    return 75\n'
        '  fi\n'
    )
    account_line = "  stream_used=$((stream_used + stream_bytes))\n"
    for hostile in (
        execute.replace(
            'stream_tmp="$GOTMPDIR/stream-capture/$stage.$stream.$package_ordinal.tmp"',
            'stream_tmp="$stream_path.raw"',
            1,
        ),
        execute.replace(
            'head -c "$read_limit" < "$fifo_path" > "$stream_tmp"',
            'cat -- "$fifo_path" > "$stream_tmp"',
            1,
        ),
        execute.replace(append_line, append_line + "\n" + append_line, 1),
        execute.replace(
            overflow_block + account_line,
            account_line + overflow_block,
            1,
        ),
        execute.replace(
            append_line,
            '    cat -- "$stream_tmp" >> "$target"',
            1,
        ),
        execute.replace(
            '    printf \'%s\\n\' "$stream_used" > "$stream_used_path"',
            '    printf \'%s\\n\' "$stream_used" > "$used_target"',
            1,
        ),
        execute.replace(
            '  return 0\n}\n\nrun_stage() {',
            '  cat -- "$fifo_path" >> "$stream_path"\n'
            '  return 0\n}\n\nrun_stage() {',
            1,
        ),
        execute.replace(
            "read_bounded_stage_stream() {",
            "removed_reader() {",
            1,
        ),
        execute + "\nread_bounded_stage_stream () {\n  :\n}\n",
    ):
        assert hostile != execute
        assert not _hard2g_read_bounded_stage_stream_automaton_is_closed(hostile)

    assert not _hard2g_read_bounded_stage_stream_automaton_is_closed(None)
    assert not _hard2g_read_bounded_stage_stream_automaton_is_closed(
        execute.replace("\n", "\r\n", 1)
    )
    assert not _hard2g_read_bounded_stage_stream_automaton_is_closed(
        execute.removesuffix("\n")
    )

    run_stage_only = execute.replace(
        "run_stage() {\n",
        "run_stage() {\n  :\n",
        1,
    )
    assert run_stage_only != execute
    assert _hard2g_read_bounded_stage_stream_automaton_is_closed(run_stage_only)
    assert not _hard2g_execute_budget_bridge_is_closed(run_stage_only)


def test_hard2g_run_stage_automaton_is_closed() -> None:
    execute = _hard2g_c02_durable_budget_execute_fixture()
    assert _hard2g_run_stage_automaton_is_closed(execute)

    stdout_wait = '    wait "$stdout_reader_pid" || stdout_reader_exit=$?'
    stderr_wait = '    wait "$stderr_reader_pid" || stderr_reader_exit=$?'
    go_wait = '    wait "$go_child_pid" || child_exit=$?'
    for hostile in (
        execute.replace("  stage_exit=0", "  stage_exit=1", 1),
        execute.replace(
            '    start_owned_go_child_group "$stage" "$package"',
            "    :",
            1,
        ),
        execute.replace(
            '    read_bounded_stage_stream "$stage" stdout '
            '"$stdout_fifo" "$stdout_path" "$stdout_used_path" '
            '"$package_ordinal" &',
            "    :",
            1,
        ),
        execute.replace('    stdout_reader_pid="$!"', "    stdout_reader_pid=0", 1),
        execute.replace(
            '    if monitor_owned_budget_stop "$go_child_pid" '
            '"$budget_stop_marker"; then',
            "    if :; then",
            1,
        ),
        execute.replace(stdout_wait, stdout_wait + "\n" + stdout_wait, 1),
        execute.replace(stderr_wait + "\n", "", 1),
        execute.replace(
            '      terminate_owned_go_child_group\n'
            '      reap_owned_go_child_group\n'
            '      wait "$stdout_reader_pid"',
            '      wait "$stdout_reader_pid"\n'
            '      terminate_owned_go_child_group\n'
            '      reap_owned_go_child_group',
            1,
        ),
        execute.replace(go_wait, go_wait + "\n" + go_wait, 1),
        execute.replace(
            "report_budget_truncated_due_to_budget=true",
            "report_budget_truncated_due_to_budget=false",
            1,
        ),
        execute.replace('      return 75', '      break', 1),
        execute.replace('      stage_exit="$child_exit"', "      stage_exit=0", 1),
        execute.replace('  return "$stage_exit"', "  return 0", 1),
        execute + "\nrun_stage () {\n  :\n}\n",
    ):
        assert hostile != execute
        assert not _hard2g_run_stage_automaton_is_closed(hostile)

    assert not _hard2g_run_stage_automaton_is_closed(None)
    assert not _hard2g_run_stage_automaton_is_closed(
        execute.replace("\n", "\r\n", 1)
    )
    assert not _hard2g_run_stage_automaton_is_closed(execute.removesuffix("\n"))

    reader_only = execute.replace(
        'stream_tmp="$GOTMPDIR/stream-capture/'
        '$stage.$stream.$package_ordinal.tmp"',
        'stream_tmp="$stream_path.raw"',
        1,
    )
    assert reader_only != execute
    assert _hard2g_run_stage_automaton_is_closed(reader_only)
    assert not _hard2g_execute_budget_bridge_is_closed(reader_only)


def test_hard2g_c02_execute_surface_is_closed() -> None:
    execute = _hard2g_c02_durable_budget_execute_fixture()
    assert _HARD2G_C02_EXECUTE_EXTENSION_ALLOWLIST == ()
    assert _hard2g_c02_execute_surface_is_closed(execute)

    expected_inventory = ("read_bounded_stage_stream", "run_stage")
    inventory = _hard2g_c02_function_inventory(execute)
    assert inventory is not None
    assert tuple(inventory) == expected_inventory

    extra_function = "\nnoop() {\n  :\n}\n"
    for hostile in (
        execute.replace(
            "\nrun_stage() {",
            "\ncheck_budget() {\n  :\n}\n\nrun_stage() {",
            1,
        ),
        execute.replace(
            "\nrun_stage() {",
            "\nstage_stream_cap() {\n  printf '%s\\n' 1\n}\n\nrun_stage() {",
            1,
        ),
        execute + extra_function,
        execute.replace("set -eu\n", "set -eu\nPATH=/tmp\n", 1),
        execute.replace(
            "set -eu\n",
            'set -eu\n: > "$run_dir/unowned"\n',
            1,
        ),
        execute + "\n# comment\n" + ': > "$run_dir/unowned"\n',
        execute + "\nrun_stage () {\n  :\n}\n",
        execute.replace("#!/bin/sh", "#!/bin/bash", 1),
        execute.replace("#!/bin/sh", "#!/usr/bin/env bash", 1),
        execute.replace("set -eu\n", "set -eu\n# unsafe\x00comment\n", 1),
        execute.replace("set -eu\n", "\vset -eu\n", 1),
        execute.replace("set -eu\n", "set -eu\n\v# not-a-shell-comment\n", 1),
    ):
        assert hostile != execute
        assert not _hard2g_c02_execute_surface_is_closed(hostile)

    assert _hard2g_c02_execute_surface_is_closed(
        execute.replace("set -eu\n", "set -eu\n# comment-only extension\n", 1)
    )
    assert not _hard2g_c02_execute_surface_is_closed(None)
    assert not _hard2g_c02_execute_surface_is_closed(
        execute.replace("\n", "\r\n", 1)
    )
    assert not _hard2g_c02_execute_surface_is_closed(execute.removesuffix("\n"))


def test_hard2g_c02_function_inventory_is_fail_closed_and_preserves_bodies() -> None:
    expected_body = (
        "if true; then",
        "if false; then",
        ":",
        "fi",
        "fi",
    )
    inventory = _hard2g_c02_function_inventory(
        """#!/bin/sh

# preamble comments do not become commands
alpha() {
  if true; then
    # body comments do not become commands
    if false; then
      :
    fi
  fi
}

beta () {
  if true; then
    if false; then
      :
    fi
  fi
}

gamma ()
{
  if true; then
    if false; then
      :
    fi
  fi
}
"""
    )
    assert inventory is not None
    assert tuple(inventory) == ("alpha", "beta", "gamma")
    assert tuple(inventory.values()) == (expected_body,) * 3
    assert _hard2g_c02_function_inventory(
        "one() {\n  :\n}\n\ntwo () {\n  :\n}\n"
    ) is not None

    runtime = _hard2g_c02_function_inventory(
        _hard2g_runtime_budget_fixture(_hard2g_runtime_budget_assignments())
    )
    execute = _hard2g_c02_function_inventory(
        _hard2g_execute_budget_bridge_fixture()
    )
    assert runtime is not None
    assert execute is not None
    assert {
        "stage_stream_cap",
        "check_budget",
        *_HARD2G_C02_RUNTIME_OWNED_HELPERS,
    }.issubset(runtime)
    assert tuple(
        name
        for name in runtime
        if name in _HARD2G_C02_RUNTIME_OWNED_HELPERS
    ) == _HARD2G_C02_RUNTIME_OWNED_HELPERS
    assert all(
        runtime[name] == (":",)
        for name in _HARD2G_C02_RUNTIME_OWNED_HELPERS
    )
    assert {"read_bounded_stage_stream", "run_stage"}.issubset(execute)
    assert not set(_HARD2G_C02_RUNTIME_OWNED_HELPERS).intersection(execute)
    assert runtime["check_budget"][-1].endswith("return 75")
    assert execute["read_bounded_stage_stream"][-1] == "return 75"
    assert execute["run_stage"][-1] == 'return "$stage_exit"'

    for hostile in (
        "one() {\n}\none () {\n}\n",
        "if true; then\n  evil() {\n  }\nfi\n",
        "(\n  evil() {\n  }\n)\n",
        "outer() {\ninner() {\n}\n}\n",
        "outer() {\ninner ()\n{\n}\n}\n",
        "one(){\n}\n",
        "function one() {\n}\n",
        '"$dynamic"() {\n}\n',
        "one() {;\n}\n",
        "one() { trailing\n}\n",
        "}\n",
        "one() {\n  :\n",
        "one ()\nnot-a-brace\n",
        "one ()\n# a skipped comment\ntwo () {\n}\n",
    ):
        assert _hard2g_c02_function_inventory(hostile) is None


def test_hard2g_c02_shell_body_lexically_closes_only_supported_forms() -> None:
    runtime = _hard2g_c02_function_inventory(
        _hard2g_runtime_budget_fixture(_hard2g_runtime_budget_assignments())
    )
    execute = _hard2g_c02_function_inventory(
        _hard2g_execute_budget_bridge_fixture()
    )
    assert runtime is not None
    assert execute is not None
    assert all(
        _hard2g_c02_shell_body_lexically_closed(body)
        for body in (*runtime.values(), *execute.values())
    )

    for accepted in (
        ('printf \'%s\\n\' "${SAFE}"',),
        ('report_bytes="$(report_regular_file_bytes)"',),
        ('metadata_bytes="$(report_metadata_bytes)"',),
        ('stream_cap="$(stage_stream_cap "$stage" "$stream")"',),
        ('stream_used="$(cat "$stream_used_path")"',),
        ('stream_bytes="$(wc -c < "$stream_tmp")"',),
        ('stat_line="$(cat "$stat_path")"',),
        ('value=$((left + 1))',),
        ('other=$((9 - right))',),
        ('if true; then',),
        ('while true; do',),
        ('*) return 1 ;;',),
        ('"$go_child_pid (go) Z "*|"$go_child_pid (go) X "*) return 1 ;;',),
        ('"$go_child_pid (go) "*) return 0 ;;',),
    ):
        assert _hard2g_c02_shell_body_lexically_closed(accepted)

    for hostile in (
        ('cat <<EOF',),
        ('cat <<< "unsafe"',),
        ('echo `unsafe`',),
        ('cat <(unsafe)',),
        ('cat >(unsafe)',),
        ('left | right',),
        ('eval unsafe',),
        ('exec unsafe',),
        ('. ./unsafe',),
        ('source ./unsafe',),
        ('alias x=unsafe',),
        ('command unsafe',),
        ('builtin unsafe',),
        ('sh -c unsafe',),
        ('/bin/sh -c unsafe',),
        ('/bin/sh -e -c unsafe',),
        ('/bin/busybox sh -c unsafe',),
        ('"$helper" unsafe',),
        ('"${helper}" unsafe',),
        ('"$@" unsafe',),
        ('"$*" unsafe',),
        ('$1 unsafe',),
        ('/$helper unsafe',),
        ('bin$helper unsafe',),
        ('./${helper} unsafe',),
        ('if "${helper}"; then',),
        ('SAFE=1 "${helper}" unsafe',),
        ('if SAFE=1 "${helper}"; then',),
        ('if SAFE=1 "$@"; then',),
        ('if SAFE=1 ./${helper}; then',),
        ('SAFE=1 > "$target" "${helper}" unsafe',),
        ('&>out unsafe',),
        ('&>>out unsafe',),
        ('true && unsafe',),
        ('true || unsafe',),
        ('nested() {',),
        ('function nested',),
        ('{ unsafe; }',),
        ('value="${x:-unsafe}"',),
        ('value="$(unsafe)"',),
        ('value="$(cat "$(unsafe)")"',),
        ('value=$((left * right))',),
        ('value=$((left + $(unsafe)))',),
        ('if true; then; unsafe',),
        ('*) return 1 ;; unsafe',),
        ('unsafe \\',),
        ('unsafe\r',),
        ('unsafe\nnext',),
        ('unsafe\x00',),
        ('  unsafe',),
        ('# comment',),
        ('',),
    ):
        assert not _hard2g_c02_shell_body_lexically_closed(hostile)


def test_hard2g_c02_shell_control_flow_is_closed() -> None:
    runtime = _hard2g_c02_function_inventory(
        _hard2g_runtime_budget_fixture(_hard2g_runtime_budget_assignments())
    )
    execute = _hard2g_c02_function_inventory(
        _hard2g_execute_budget_bridge_fixture()
    )
    assert runtime is not None
    assert execute is not None
    assert all(
        _hard2g_c02_shell_control_flow_is_closed(body)
        for body in (*runtime.values(), *execute.values())
    )

    for accepted in (
        ("if ! true; then", "fi"),
        (
            "while true; do",
            "if true; then",
            "else",
            "fi",
            "done",
        ),
        (
            'case "$stage:$stream" in',
            "build:stdout) return 0 ;;",
            "*) return 64 ;;",
            "esac",
        ),
        ("while true; do", 'done < "$plan_path"'),
    ):
        assert _hard2g_c02_shell_control_flow_is_closed(accepted)

    for hostile in (
        ("fi",),
        ("done",),
        ('done < "$plan_path"',),
        ("else",),
        ("esac",),
        ("if true; then", "else", "else", "fi"),
        ("if true; then", "while true; do", "fi", "done"),
        ("if true; then",),
        ("while true; do",),
        ('case "$stage" in', "*) return 64 ;;"),
        ("if; then",),
        ("if true then",),
        ("if then; then", "fi"),
        ("if case x in; then", "fi"),
        ("if X=1 then; then", "fi"),
        ("if X=1 ! if true; then", "fi"),
        ("if ! X=1 case x in; then", "fi"),
        ("if true # ; then", "fi"),
        ("if true\\; then", "fi"),
        ("while; do",),
        ("while if true; do", "done"),
        ("while X=1 ! if true; do", "done"),
        ("while true # ; do", "done"),
        ("while true\\; do", "done"),
        ('case in',),
        ('case "$stage"',),
        ('case "$stage" # in', "*) return 64 ;;", "esac"),
        ('case "$stage" extra in', "*) return 64 ;;", "esac"),
        ('case "$stage" in', "esac"),
        (
            'case "$stage" in',
            "printf '%s\\n' \"${SAFE}\"",
            "*) return 64 ;;",
            "esac",
        ),
        (
            'case "$stage" in',
            "*) return 64 ;;",
            "printf '%s\\n' \"${SAFE}\"",
            "esac",
        ),
        (
            'case "$stage" in',
            "if true; then",
            "fi",
            "*) return 64 ;;",
            "esac",
        ),
        ("*) return 64 ;;",),
        ("while true; do", 'done < "$plan_path" > "$other"'),
        ("then",),
        ("do",),
        ("in",),
        ("! true",),
        ("! X=1 true",),
        ("X=1 ! true",),
        ("! if true; then",),
        ("X=1 if true; then",),
        ("X=1 fi",),
        ("elif true; then",),
        ("for item in items",),
        ("until true; do",),
        ("select item in items",),
        ("function item",),
        ("break",),
        ("continue",),
        ("exit",),
    ):
        assert not _hard2g_c02_shell_control_flow_is_closed(hostile)


def test_hard2g_c02_runtime_helper_ownership_is_closed() -> None:
    runtime = _hard2g_runtime_budget_fixture(_hard2g_runtime_budget_assignments())
    plan = "#!/bin/sh\nset -eu\n"
    execute = _hard2g_execute_budget_bridge_fixture()
    receipt = "#!/bin/sh\nset -eu\n"
    assert _hard2g_c02_runtime_helper_ownership_is_closed(
        runtime, plan, execute, receipt
    )
    runtime_with_real_helper = runtime.replace(
        "start_owned_go_child_group() {\n  :\n}\n\n",
        "start_owned_go_child_group() {\n  return 0\n}\n\n",
        1,
    )
    assert runtime_with_real_helper != runtime
    assert _hard2g_c02_runtime_helper_ownership_is_closed(
        runtime_with_real_helper, plan, execute, receipt
    )

    missing = runtime.replace(
        "start_owned_go_child_group() {\n  :\n}\n\n", "", 1
    )
    empty = runtime.replace(
        "start_owned_go_child_group() {\n  :\n}\n\n",
        "start_owned_go_child_group() {\n}\n\n",
        1,
    )
    duplicate = runtime + "\nstart_owned_go_child_group() {\n  :\n}\n"
    for hostile in (missing, empty, duplicate):
        assert hostile != runtime
        assert not _hard2g_c02_runtime_helper_ownership_is_closed(
            hostile, plan, execute, receipt
        )

    for definition in (
        "start_owned_go_child_group() {\n  :\n}\n",
        "start_owned_go_child_group () {\n  :\n}\n",
        "start_owned_go_child_group ()\n{\n  :\n}\n",
        "start_owned_go_child_group(){\n  :\n}\n",
        "function start_owned_go_child_group() {\n  :\n}\n",
    ):
        hostile = execute + "\n" + definition
        assert hostile != execute
        assert not _hard2g_c02_runtime_helper_ownership_is_closed(
            runtime, plan, hostile, receipt
        )

    for declaration in (
        "start_owned_go_child_group=payload\n",
        "readonly start_owned_go_child_group\n",
        "alias start_owned_go_child_group=payload\n",
        "unset start_owned_go_child_group\n",
    ):
        hostile = execute + declaration
        assert hostile != execute
        assert not _hard2g_c02_runtime_helper_ownership_is_closed(
            runtime, plan, hostile, receipt
        )

    for source_name in ("plan", "receipt"):
        for fragment in (
            "start_owned_go_child_group() {\n  :\n}\n",
            "start_owned_go_child_group=payload\n",
            "printf '%s\\n' 'start_owned_go_child_group'\n",
        ):
            hostile_plan = plan + fragment if source_name == "plan" else plan
            hostile_receipt = receipt + fragment if source_name == "receipt" else receipt
            assert hostile_plan != plan or hostile_receipt != receipt
            assert not _hard2g_c02_runtime_helper_ownership_is_closed(
                runtime, hostile_plan, execute, hostile_receipt
            )

    for fragment in (
        ". /payload\n",
        "source /payload\n",
        "alias unsafe=payload\n",
    ):
        hostile = plan + fragment
        assert hostile != plan
        assert not _hard2g_c02_runtime_helper_ownership_is_closed(
            runtime, hostile, execute, receipt
        )

    for source_name in ("plan", "execute", "receipt"):
        hostile_plan = plan + "trap ':' EXIT\n" if source_name == "plan" else plan
        hostile_execute = (
            execute + "trap ':' EXIT\n" if source_name == "execute" else execute
        )
        hostile_receipt = (
            receipt + "trap ':' EXIT\n" if source_name == "receipt" else receipt
        )
        assert (
            hostile_plan != plan
            or hostile_execute != execute
            or hostile_receipt != receipt
        )
        assert not _hard2g_c02_runtime_helper_ownership_is_closed(
            runtime, hostile_plan, hostile_execute, hostile_receipt
        )

    missing_call = execute.replace(
        '    start_owned_go_child_group "$stage" "$package"\n', "    :\n", 1
    )
    extra_call = execute.replace(
        '    start_owned_go_child_group "$stage" "$package"\n',
        '    start_owned_go_child_group "$stage" "$package"\n'
        '    start_owned_go_child_group "$stage" "$package"\n',
        1,
    )
    wrong_call = execute.replace('"$package"', '"$wrong_package"', 1)
    for hostile in (missing_call, extra_call, wrong_call):
        assert hostile != execute
        assert not _hard2g_c02_runtime_helper_ownership_is_closed(
            runtime, plan, hostile, receipt
        )


def test_hard2g_c02_durable_budget_fixture_topology_is_closed() -> None:
    runtime = _hard2g_c02_durable_budget_runtime_fixture()
    execute = _hard2g_c02_durable_budget_execute_fixture()
    receipt = _hard2g_container_receipt_fixture()
    assert _hard2g_c02_durable_budget_fixture_topology_is_closed(
        runtime, execute, receipt
    )

    marker_false = "report_budget_truncated_due_to_budget=false"
    marker_true = "report_budget_truncated_due_to_budget=true"
    runtime_marker_in_function = runtime.replace(
        f"{marker_false}\n\n", "", 1
    ).replace(
        "stage_stream_cap() {\n",
        f"stage_stream_cap() {{\n  {marker_false}\n",
        1,
    )
    for hostile in (
        runtime.replace(f"{marker_false}\n", "", 1),
        runtime.replace(marker_false, f"{marker_false}\n{marker_false}", 1),
        runtime.replace(marker_false, marker_true, 1),
        runtime.replace(marker_false, 'report_budget_truncated_due_to_budget="$flag"', 1),
        runtime_marker_in_function,
    ):
        assert hostile != runtime
        assert not _hard2g_c02_durable_budget_fixture_topology_is_closed(
            hostile, execute, receipt
        )

    execute_marker_in_reader = execute.replace(
        f"  {marker_true}\n", "", 1
    ).replace(
        "read_bounded_stage_stream() {\n",
        f"read_bounded_stage_stream() {{\n  {marker_true}\n",
        1,
    )
    for hostile in (
        execute.replace(f"  {marker_true}\n", "  :\n", 1),
        execute.replace(marker_true, f"{marker_true}\n  {marker_true}", 1),
        execute.replace(marker_true, marker_false, 1),
        execute.replace(marker_true, 'report_budget_truncated_due_to_budget="$flag"', 1),
        execute_marker_in_reader,
        execute.replace(marker_true, "truncation_evidence=budget-stop", 1),
    ):
        assert hostile != execute
        assert not _hard2g_c02_durable_budget_fixture_topology_is_closed(
            runtime, hostile, receipt
        )

    writer = '  write_atomic "$run_dir/container-receipt.json" "$receipt"\n'
    for hostile in (
        receipt.replace(writer, "", 1),
        receipt.replace(writer, writer + writer, 1),
        receipt.replace("container-receipt.json", "receipt-copy.json", 1),
        receipt.replace(
            writer,
            '  write_atomic "$run_dir/$receipt_name" "$receipt"\n',
            1,
        ),
    ):
        assert hostile != receipt
        assert not _hard2g_c02_durable_budget_fixture_topology_is_closed(
            runtime, execute, hostile
        )


def test_hard2g_report_budget_truncation_receipt_binding_is_closed() -> None:
    runtime = _hard2g_c02_durable_budget_runtime_fixture()
    plan = "#!/bin/sh\nset -eu\n"
    execute = _hard2g_c02_durable_budget_execute_fixture()
    receipt = _hard2g_container_receipt_fixture()
    assert _hard2g_report_budget_truncation_receipt_binding_is_closed(
        runtime, plan, execute, receipt
    )

    hostile_sources = (
        (
            runtime.replace("report_budget_truncated_due_to_budget=false\n", "", 1),
            plan,
            execute,
            receipt,
        ),
        (
            runtime,
            plan + "report_budget_truncated_due_to_budget=false\n",
            execute,
            receipt,
        ),
        (
            runtime,
            plan + ". /payload\n",
            execute,
            receipt,
        ),
        (
            runtime,
            plan,
            execute.replace(
                "report_budget_truncated_due_to_budget=true\n", "", 1
            ),
            receipt,
        ),
        (
            runtime,
            plan,
            execute.replace(
                "report_budget_truncated_due_to_budget=true",
                "truncation_evidence=budget-stop",
                1,
            ),
            receipt,
        ),
        (
            runtime,
            plan,
            execute,
            receipt.replace(
                'case "${report_budget_truncated_due_to_budget-}" in',
                "case false in",
                1,
            ),
        ),
        (
            runtime,
            plan,
            execute,
            receipt.replace(
                '$report_budget_json}"',
                'false}"',
                1,
            ),
        ),
    )
    for (
        hostile_runtime,
        hostile_plan,
        hostile_execute,
        hostile_receipt,
    ) in hostile_sources:
        assert (hostile_runtime, hostile_plan, hostile_execute, hostile_receipt) != (
            runtime,
            plan,
            execute,
            receipt,
        )
        assert not _hard2g_report_budget_truncation_receipt_binding_is_closed(
            hostile_runtime,
            hostile_plan,
            hostile_execute,
            hostile_receipt,
        )


def test_hard2g_image_and_build_input_bindings_are_exact() -> None:
    contract = load_root_compose_contract(_REPO_ROOT)
    policy = _mapping_value(contract, "policy")
    config = _mapping_value(policy, "hard2g_observer")
    dockerfile = _required_asset_text(
        _REPO_ROOT / "testframework" / "docker" / "Dockerfile.qa"
    )
    assert _hard2g_provenance_bindings_are_closed(
        _REPO_ROOT,
        policy,
        dockerfile,
    )
    image_inspect_argvs = config["image_inspect_argvs"]
    assert type(image_inspect_argvs) is tuple
    hostile_argvs = (
        (
            *image_inspect_argvs[:-1],
            (*image_inspect_argvs[-1][:-1], "hostile-backend-image"),
        ),
        (
            (*image_inspect_argvs[0], "--no-trunc"),
            *image_inspect_argvs[1:],
        ),
    )
    for argv in hostile_argvs:
        hostile_config = dict(config)
        hostile_config["image_inspect_argvs"] = argv
        hostile_policy = dict(policy)
        hostile_policy["hard2g_observer"] = MappingProxyType(hostile_config)
        assert not _hard2g_provenance_bindings_are_closed(
            _REPO_ROOT,
            MappingProxyType(hostile_policy),
            dockerfile,
        )

    manifest = _canonical_build_input_manifest_bytes(
        _REPO_ROOT,
        config["canonical_build_input_paths"],
    )
    assert type(manifest) is bytes
    manifest_sha256 = "sha256:" + hashlib.sha256(manifest).hexdigest()
    assert _hard2g_exact_manifest_binding_is_closed(
        _REPO_ROOT,
        manifest,
        manifest_sha256,
        config["canonical_build_input_paths"],
    )


def test_hard2g_report_budget_is_exact_and_fail_closed() -> None:
    assert _hard2g_report_budget_is_closed(_HARD2G_REPORT_BUDGET)
    too_small = dict(_HARD2G_REPORT_BUDGET)
    too_small["total_bytes"] = 63 * 1024 * 1024
    assert not _hard2g_report_budget_is_closed(MappingProxyType(too_small))
    stream_mismatch = dict(_HARD2G_REPORT_BUDGET)
    streams = dict(_mapping_value(_HARD2G_REPORT_BUDGET, "streams"))
    streams["race"] = MappingProxyType(
        {"stdout": 23 * 1024 * 1024, "stderr": 4 * 1024 * 1024}
    )
    stream_mismatch["streams"] = MappingProxyType(streams)
    assert not _hard2g_report_budget_is_closed(
        MappingProxyType(stream_mismatch)
    )


def test_hard2g_canonical_build_input_manifest_expands_wildcards_and_hashes_bytes(
    tmp_path: Path,
) -> None:
    (tmp_path / "docker-compose.local.yml").write_bytes(b"compose\n")
    observers = tmp_path / "testframework" / "observers" / "backend_unit"
    observers.mkdir(parents=True)
    (observers / "observe.py").write_bytes(b"observe\n")
    (observers / "contracts.py").write_bytes(b"contracts\n")
    policy_paths = (
        "docker-compose.local.yml",
        "testframework/observers/**",
    )
    manifest = _canonical_build_input_manifest_bytes(tmp_path, policy_paths)
    assert type(manifest) is bytes
    assert manifest == (
        b"docker-compose.local.yml\tsha256:"
        + hashlib.sha256(b"compose\n").hexdigest().encode()
        + b"\n"
        + b"testframework/observers/backend_unit/contracts.py\tsha256:"
        + hashlib.sha256(b"contracts\n").hexdigest().encode()
        + b"\n"
        + b"testframework/observers/backend_unit/observe.py\tsha256:"
        + hashlib.sha256(b"observe\n").hexdigest().encode()
        + b"\n"
    )
    manifest_sha256 = "sha256:" + hashlib.sha256(manifest).hexdigest()
    assert _hard2g_exact_manifest_binding_is_closed(
        tmp_path,
        manifest,
        manifest_sha256,
        policy_paths,
    )
    assert not _hard2g_exact_manifest_binding_is_closed(
        tmp_path,
        manifest.replace(b"\n", b"\r\n", 1),
        manifest_sha256,
        policy_paths,
    )
    assert not _hard2g_exact_manifest_binding_is_closed(
        tmp_path,
        manifest,
        "sha256:" + "0" * 64,
        policy_paths,
    )
    (observers / "late.py").write_bytes(b"late\n")
    assert not _hard2g_exact_manifest_binding_is_closed(
        tmp_path,
        manifest,
        manifest_sha256,
        policy_paths,
    )


def test_hard2g_canonical_build_input_manifest_rejects_symlink_parent(
    tmp_path: Path,
) -> None:
    observers = tmp_path / "testframework" / "observers"
    observers.mkdir(parents=True)
    (observers / "safe.py").write_bytes(b"safe\n")
    policy_paths = ("testframework/observers/**",)
    manifest = _canonical_build_input_manifest_bytes(tmp_path, policy_paths)
    assert type(manifest) is bytes
    manifest_sha256 = "sha256:" + hashlib.sha256(manifest).hexdigest()
    outside = tmp_path / "outside"
    outside.mkdir()
    (outside / "escaped.py").write_bytes(b"escaped\n")
    try:
        (observers / "linked-parent").symlink_to(
            outside,
            target_is_directory=True,
        )
    except (NotImplementedError, OSError):
        pytest.skip("the OS does not permit directory symlink creation")
    assert _canonical_build_input_manifest_bytes(tmp_path, policy_paths) is _INVALID
    assert not _hard2g_exact_manifest_binding_is_closed(
        tmp_path,
        manifest,
        manifest_sha256,
        policy_paths,
    )


def test_hard2g_container_publication_order_is_closed() -> None:
    canonical_receipt = _hard2g_container_receipt_fixture()
    assert _hard2g_container_publication_is_closed(canonical_receipt)
    builder_call = "  build_container_receipt\n"
    terminal_writer = '  write_atomic "$run_dir/container-receipt.json" "$receipt"\n'
    receipt_builder_line = (
        r'  receipt="${receipt%?},\"report_budget\":$report_budget_json}"' + "\n"
    )
    for hostile_builder in (
        canonical_receipt.replace(builder_call, "", 1),
        canonical_receipt.replace(builder_call, builder_call + builder_call, 1),
        canonical_receipt.replace(
            builder_call + terminal_writer,
            terminal_writer + builder_call,
            1,
        ),
        canonical_receipt.replace(
            'case "${report_budget_truncated_due_to_budget-}" in',
            "case false in",
            1,
        ),
        canonical_receipt.replace(
            'case "${report_budget_truncated_due_to_budget-}" in',
            'case "$report_budget_truncated_due_to_budget" in',
            1,
        ),
        canonical_receipt.replace(
            "*'\"report_budget\"'*)",
            "*'\"other\"'*)",
            1,
        ),
        canonical_receipt.replace("\\{*\\})", "*)", 1),
        canonical_receipt.replace(
            '"truncated_due_to_budget":false',
            '"truncated_due_to_budget":"false"',
            1,
        ),
        canonical_receipt.replace(
            '"truncated_due_to_budget":false}',
            '"truncated_due_to_budget":false,'
            '"truncated_due_to_budget":true}',
            1,
        ),
        canonical_receipt.replace(
            receipt_builder_line,
            '  report_budget_truncated_due_to_budget=false\n'
            + receipt_builder_line,
            1,
        ),
        canonical_receipt.replace(
            receipt_builder_line,
            '  write_atomic "$run_dir/budget.json" "$report_budget_json"\n'
            + receipt_builder_line,
            1,
        ),
    ):
        assert hostile_builder != canonical_receipt
        assert not _hard2g_container_publication_is_closed(hostile_builder)
    assert not _hard2g_container_publication_is_closed(
        canonical_receipt + "}\n"
    )
    assert not _hard2g_container_publication_is_closed(
        canonical_receipt.replace(
            '"$run_dir/CONTAINER_COMPLETE"',
            '"$run_dir/container-receipt.json"',
            1,
        )
    )
    assert not _hard2g_container_publication_is_closed(
        canonical_receipt + 'write_atomic "$run_dir/SEAL" "$seal"\n'
    )
    assert not _hard2g_container_publication_is_closed(
        canonical_receipt
        + 'write_atomic "$run_dir/observer-receipt.json" "$receipt"\n'
    )
    assert not _hard2g_container_publication_is_closed(
        'mv -- "$run_dir/final-receipt.json" "$run_dir/result.json"\n'
        + canonical_receipt
    )
    assert not _hard2g_container_publication_is_closed(
        'mv -- "$run_dir/$stage.$artifact.tmp" "$run_dir/$stage.$artifact"\n'
        + canonical_receipt
    )
    assert not _hard2g_container_publication_is_closed(
        canonical_receipt.replace(
            '"$run_dir/container-receipt.json"',
            '"$run_dir/$receipt_name"',
            1,
        )
    )
    assert not _hard2g_container_publication_is_closed(
        canonical_receipt
        + "final_name='final-'receipt.json\n"
        + 'write_atomic "$run_dir/$final_name" "$receipt"\n'
    )
    assert not _hard2g_container_publication_is_closed(
        canonical_receipt
        + "final_part=final-\n"
        + "receipt_part=receipt.json\n"
        + 'write_atomic "$run_dir/$final_part$receipt_part" "$receipt"\n'
    )
    assert not _hard2g_container_publication_is_closed(
        canonical_receipt + "eval 'write_owner'\n"
    )
    assert not _hard2g_container_publication_is_closed(
        canonical_receipt.replace(
            "  write_inventory\n  : >",
            "  : >\n  write_inventory\n  : >",
            1,
        )
    )
    for hostile_early_return in (
        canonical_receipt.replace(
            "  write_package_plans\n",
            "  return 0\n  write_package_plans\n",
            1,
        ),
        canonical_receipt.replace(
            "  write_inventory\n",
            "  write_inventory\n  return 0\n",
            1,
        ),
    ):
        assert hostile_early_return != canonical_receipt
        assert not _hard2g_container_publication_is_closed(
            hostile_early_return
        )
    for hostile in (
        canonical_receipt
        + 'if true; then write_atomic "$run_dir/SEAL" "$seal"; fi\n',
        canonical_receipt + "if true; then arbitrary_writer; fi\n",
        canonical_receipt + "true && arbitrary_writer\n",
        canonical_receipt + "false || arbitrary_writer\n",
        canonical_receipt + "true | arbitrary_writer\n",
        canonical_receipt
        + 'mv -- "$run_dir/$stage.$artifact.tmp" "$run_dir/$stage.$artifact"; '
        + 'write_atomic "$run_dir/SEAL" "$seal"\n',
        canonical_receipt
        + 'write_atomic "$run_dir/$(printf SEAL)" "$seal"\n',
        canonical_receipt
        + 'receipt_target="$run_dir/container-receipt.json"\n'
        + 'write_atomic "$receipt_target" "$receipt"\n',
        canonical_receipt
        + 'mv -- "$run_dir/${stage}.${artifact}.tmp" '
        + '"$run_dir/$stage.$artifact"\n',
        canonical_receipt
        + 'mv -- "$run_dir/$stage.$artifact.tmp" '
        + '"$run_dir/$stage.$artifact"\n'
        + 'mv -- "$run_dir/packet.partial" "$run_dir/packet"\n',
        canonical_receipt + 'printf x > "$run_dir/SEAL"\n',
    ):
        assert not _hard2g_container_publication_is_closed(hostile)
    for hostile_simple_command in (
        "arbitrary_writer\n",
        "python -c 'pass'\n",
        "/tmp/writer\n",
    ):
        assert not _hard2g_container_publication_is_closed(
            canonical_receipt + hostile_simple_command
        )
    for mutation in (
        "read run_dir\n",
        "read -r run_dir\n",
        "run_dir=/tmp\n",
        "export run_dir=/tmp\n",
        "readonly run_dir=/tmp\n",
        "local run_dir=/tmp\n",
        "declare run_dir=/tmp\n",
        "typeset run_dir=/tmp\n",
        "unset run_dir\n",
    ):
        assert not _hard2g_container_publication_is_closed(
            canonical_receipt + mutation
        )
    assert not _hard2g_container_publication_is_closed(
        canonical_receipt.replace(
            '  mv -- "$run_dir/$stage.$artifact.tmp" "$run_dir/$stage.$artifact"',
            "  :",
            1,
        )
    )


def test_hard2g_readme_metadata_is_copied_without_becoming_an_oracle(
    tmp_path: Path,
) -> None:
    dockerfile = "\n".join(
        (
            "FROM python-base AS python-tests",
            *_HARD2G_METADATA_README_COPY_INSTRUCTIONS,
            *sorted(_HARD2G_METADATA_ALLOWED_PYTHON_EXECUTABLE_INSTRUCTIONS),
        )
    ) + "\n"
    assert _hard2g_readme_metadata_copies_are_closed(dockerfile)
    assert _hard2g_readme_metadata_copies_are_closed(
        dockerfile + "ENV README_NOTE=human-metadata\n"
    )
    for copy_instruction in (
        *_HARD2G_PYTHON_TESTS_CURRENT_NON_METADATA_COPY_INSTRUCTIONS,
        _HARD2G_PYTHON_TESTS_FUTURE_OBSERVER_COPY_INSTRUCTION,
    ):
        assert _hard2g_readme_metadata_copies_are_closed(
            dockerfile + copy_instruction + "\n"
        )
    for hostile_copy_instruction in (
        "COPY testframework/README.* ./",
        "COPY testframework/* ./",
        "COPY testframework/ ./testframework/",
    ):
        assert not _hard2g_readme_metadata_copies_are_closed(
            dockerfile + hostile_copy_instruction + "\n"
        )
    assert not _hard2g_readme_metadata_copies_are_closed(
        "FROM python@sha256:base AS python-base\n"
        "COPY testframework/src/ ./src/\n"
        + dockerfile
    )
    assert not _hard2g_readme_metadata_copies_are_closed(
        dockerfile.replace(
            "COPY testframework/schemas/ ./schemas/",
            "COPY testframework/schemas/README.md ./schemas/README.md",
            1,
        )
    )
    assert not _hard2g_readme_metadata_copies_are_closed(
        dockerfile + "COPY testframework/docker/README.md ./docker/README.md\n"
    )
    assert not _hard2g_readme_metadata_copies_are_closed(
        dockerfile + "RUN cat /opt/hub/testframework/docker/README.md\n"
    )
    assert not _hard2g_readme_metadata_copies_are_closed(
        dockerfile
        + 'RUN stem=READ; cat "/opt/hub/testframework/$stem"ME.md\n'
    )
    assert not _hard2g_readme_metadata_copies_are_closed(
        dockerfile
        + "RUN printf '\\122\\105\\101\\104\\115\\105\\056\\155\\144\\n' | xargs cat\n"
    )
    assert not _hard2g_readme_metadata_copies_are_closed(
        dockerfile + 'CMD ["cat", "/opt/hub/testframework/README.md"]\n'
    )
    assert not _hard2g_readme_metadata_copies_are_closed(
        dockerfile + 'ENTRYPOINT ["/usr/local/bin/hostile"]\n'
    )
    assert not _hard2g_readme_metadata_copies_are_closed(
        dockerfile + "COPY --from=python-base /README.md ./README.md\n"
    )
    assert not _hard2g_readme_metadata_copies_are_closed(
        dockerfile + 'RUN cat /opt/hub/testframework/READ""ME.md\n'
    )
    assert not _hard2g_readme_metadata_copies_are_closed(
        dockerfile + "RUN cat /opt/hub/testframework/READ${part}ME.md\n"
    )
    assert not _hard2g_readme_metadata_copies_are_closed(
        dockerfile + "COPY testframework/runner/README.md ./runner/README.md\n"
    )
    assert _HARD2G_METADATA_README_PATHS == frozenset(
        {
            "testframework/README.md",
            "testframework/schemas/README.md",
            "testframework/docker/README.md",
        }
    )
    assert "testframework/runner/README.md" not in _HARD2G_METADATA_README_PATHS
    assert "testframework/runner/README.md" in _RETIRED_PATHS

    _write_valid_layout(tmp_path)
    for relative in _HARD2G_METADATA_README_PATHS:
        path = tmp_path / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text("/qa-gate is metadata only.\n", encoding="utf-8")

    authority_paths = _active_authority_paths(tmp_path)
    assert isinstance(authority_paths, tuple)
    assert all(
        path.relative_to(tmp_path).as_posix() not in _HARD2G_METADATA_README_PATHS
        for path in authority_paths
    )
    assert _active_authority_is_clean(tmp_path, _FORBIDDEN_TOKENS)


def test_hard2g_observer_sources_require_safe_path_read_authority() -> None:
    observer_source = "testframework/observers/backend_unit/observe.py"
    harmless = (
        "# README is metadata, not an observer authority.\n"
        'readme_error = ValueError("source path is denied")\n'
    )
    literal_safe_wrapper = (
        "from observers.backend_unit.safe_paths import open_regular_relative\n"
        'descriptor = open_regular_relative(root, "owner.json", flags)\n'
    )
    dynamic_safe_wrapper = (
        "from observers.backend_unit.safe_paths import open_regular_relative\n"
        "relative = event.relative\n"
        "descriptor = open_regular_relative(root, relative, flags)\n"
    )
    valid_safe_paths = (
        "import os\n"
        "def open_regular_relative(root, relative, flags):\n"
        "    candidate = root / relative\n"
        '    if candidate.name.casefold() == "readme.md":\n'
        '        raise ValueError("README paths are not readable")\n'
        "    return os.open(candidate, flags)\n"
    )
    thread_positive = "thread = lambda: None\nthread()\n"
    assert _hard2g_observer_source_has_closed_read_authority(observer_source, harmless)
    assert _hard2g_observer_source_has_closed_read_authority(
        observer_source, literal_safe_wrapper
    )
    assert _hard2g_observer_source_has_closed_read_authority(
        observer_source, dynamic_safe_wrapper
    )
    assert _hard2g_observer_source_has_closed_read_authority(
        observer_source, thread_positive
    )
    assert _hard2g_observer_source_has_closed_read_authority(
        _HARD2G_OBSERVER_SAFE_PATHS_SOURCE_PATH,
        valid_safe_paths,
    )

    hostile_sources = (
        'Path("/reports/owner.json").open()\n',
        'Path("/reports/owner.json").read_text()\n',
        'Path("/reports/owner.json").read_bytes()\n',
        'Path("/reports").joinpath(name).read_text()\n',
        'Path(f"/reports/{name}.json").read_text()\n',
        "candidate.read_text()\n",
        'getattr(candidate, "read_text")()\n',
        "open(dynamic_path)\n",
        "os.open(dynamic_path, 0)\n",
        "from io import open as file_open\nfile_open(dynamic_path)\n",
        "reader = open\nreader(dynamic_path)\n",
        "reader = candidate.read_text\nreader()\n",
        "reader = os.open\nreader(dynamic_path, 0)\n",
        "exec(source)\n",
        "eval(expression)\n",
        '__builtins__["open"](dynamic_path)\n',
        "program = exec\nprogram(source)\n",
        "program = eval\nprogram(expression)\n",
        "lookup = getattr\nreader = lookup(candidate, 'read_text')\nreader()\n",
        "from builtins import getattr as lookup\n"
        "reader = lookup(candidate, 'read_text')\nreader()\n",
        "lookup = getattr\nlookup(candidate, method_name)()\n",
        "readers = {x: open}\nreaders[x](path)\n",
        "def use(reader=open):\n    reader(path)\nuse()\n",
        "def use(reader):\n    reader(path)\nuse(open)\n",
        "reader, = (open,)\nreader(path)\n",
        "[reader] = [open]\nreader(path)\n",
        "runners = {x: exec}\nrunners[x](source)\n",
        "def use(runner):\n    runner(source)\nuse(exec)\n",
        "lookups = {x: getattr}\nlookups[x](candidate, method_name)\n",
        "def use(lookup):\n    lookup(candidate, method_name)\nuse(getattr)\n",
        "first = middle = last = open\nlast(path)\n",
        "def use(reader):\n    alias = reader\n    alias(path)\nuse(open)\n",
        "def outer():\n"
        "    def inner():\n"
        "        open(path)\n"
        "    inner()\n"
        "outer()\n",
        "class Reader:\n"
        "    def read(self):\n"
        "        open(path)\n"
        "Reader().read()\n",
        "def use(*args):\n    args[0](path)\nuse(open)\n",
        "def use(**kwargs):\n"
        "    kwargs['reader'](path)\n"
        "use(reader=open)\n",
        "def use(reader):\n"
        "    reader(path)\n"
        "use(*(open,))\n",
        "def use(reader):\n"
        "    reader(path)\n"
        'use(**{"reader": open})\n',

        "reader = lambda value: value\n"
        "def use():\n"
        "    global reader\n"
        "    reader = open\n"
        "    reader(path)\n"
        "use()\n",
        "def outer():\n"
        "    reader = lambda value: value\n"
        "    def use():\n"
        "        nonlocal reader\n"
        "        reader = open\n"
        "        reader(path)\n"
        "    use()\n"
        "outer()\n",
    )
    assert all(
        not _hard2g_observer_source_has_closed_read_authority(
            observer_source,
            source,
        )
        for source in hostile_sources
    )
    harmless_callable_sources = (
        "readers = {x: lambda value: value}\nreaders[x](path)\n",
        "def use(reader=lambda value: value):\n    reader(path)\nuse()\n",
        "def use(reader):\n    reader(path)\nuse(lambda value: value)\n",
        "def use(reader):\n    reader(path)\nuse(reader=lambda value: value)\n",
        "reader, = (lambda value: value,)\nreader(path)\n",
        "[reader] = [lambda value: value]\nreader(path)\n",
        "runners = {x: lambda value: value}\nrunners[x](source)\n",
        "def use(runner):\n    runner(source)\nuse(lambda value: value)\n",
        "lookups = {x: lambda value, name: None}\n"
        "lookups[x](candidate, method_name)\n",
        "def use(lookup):\n    lookup(candidate, method_name)\n"
        "use(lambda value, name: None)\n",
        "def ignore(reader):\n    return None\n"
        "def use(reader):\n    reader(path)\n"
        "ignore(open)\n"
        "use(lambda value: value)\n",
        "reader = open\n"
        "def use(reader):\n    reader(path)\n"
        "use(lambda value: value)\n",
        "def set_reader():\n    reader = open\n"
        "def use(reader):\n    reader(path)\n"
        "set_reader()\n"
        "use(lambda value: value)\n",
        "def use(reader, values):\n"
        "    reader(path)\n"
        'use(lambda value: value, ("data",))\n',
        "def use(reader, options):\n"
        "    reader(path)\n"
        'use(lambda value: value, {"reader": "data"})\n',

    )
    assert all(
        _hard2g_observer_source_has_closed_read_authority(
            observer_source,
            source,
        )
        for source in harmless_callable_sources
    )
    for hostile_safe_paths in (
        (
            "import os\n"
            "def open_regular_relative(root, relative, flags):\n"
            "    candidate = root / relative\n"
            "    return os.open(candidate, flags)\n"
        ),
        (
            "import os\n"
            "def open_regular_relative(root, relative, flags):\n"
            "    candidate = root / relative\n"
            "    descriptor = os.open(candidate, flags)\n"
            '    if candidate.name.casefold() == "readme.md":\n'
            '        raise ValueError("README paths are not readable")\n'
            "    return descriptor\n"
        ),
        (
            "import os\n"
            "def open_regular_relative(root, relative, flags):\n"
            "    candidate = root / relative\n"
            '    if candidate.name.casefold() == "readme.md":\n'
            '        raise ValueError("README paths are not readable")\n'
            "    def delayed_open():\n"
            "        return os.open(candidate, flags)\n"
            "    return delayed_open()\n"
        ),
    ):
        assert not _hard2g_observer_source_has_closed_read_authority(
            _HARD2G_OBSERVER_SAFE_PATHS_SOURCE_PATH,
            hostile_safe_paths,
        )


def test_hard2g_observer_source_manifest_rejects_uncovered_modules(
    tmp_path: Path,
) -> None:
    for relative in _HARD2G_OBSERVER_SOURCE_PATHS:
        path = tmp_path / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text("pass\n", encoding="utf-8")
    assert _hard2g_observer_source_manifest_is_closed(
        tmp_path,
        _HARD2G_OBSERVER_SOURCE_PATHS,
    )
    extra = tmp_path / "testframework/observers/backend_unit/uncovered.py"
    extra.write_text("pass\n", encoding="utf-8")
    assert not _hard2g_observer_source_manifest_is_closed(
        tmp_path,
        _HARD2G_OBSERVER_SOURCE_PATHS,
    )


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
            "unset GOMAXPROCS",
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
        hard2g_observer = _parse_hard2g_observer_config(
            document["hard2g_observer"]
        )
        if not isinstance(hard2g_observer, Mapping):
            return _INVALID
        return MappingProxyType(
            {
                "schema": document["schema"],
                "root_compose": document["root_compose"],
                "public_services": tuple(document["public_services"]),
                "retained_docker_paths": tuple(document["retained_docker_paths"]),
                "retired_paths": tuple(document["retired_paths"]),
                "forbidden_tokens": tuple(document["forbidden_tokens"]),
                "hard2g_observer": hard2g_observer,
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


def _parse_hard2g_observer_config(value: object) -> Mapping[str, object] | object:
    try:
        if type(value) is not dict or tuple(value) != _HARD2G_OBSERVER_CONFIG_KEYS:
            return _INVALID
        source_paths = _string_tuple(value["source_paths"])
        selftest_paths = _string_tuple(value["selftest_paths"])
        image_inspect_argvs = _nested_string_argvs(value["image_inspect_argvs"])
        build_input_paths = _string_tuple(value["canonical_build_input_paths"])
        if (
            value["module"] != _HARD2G_OBSERVER_MODULE
            or value["python_stage"] != "python-tests"
            or source_paths != _HARD2G_OBSERVER_SOURCE_PATHS
            or selftest_paths != _HARD2G_OBSERVER_SELFTEST_PATHS
            or value["policy_path"] != _HARD2G_OBSERVER_POLICY_PATH
            or image_inspect_argvs != _HARD2G_IMAGE_INSPECT_ARGVS
            or build_input_paths != _HARD2G_BUILD_INPUT_PATHS
            or value["snapshot_assertion_source"]
            != _HARD2G_OBSERVER_SNAPSHOT_ASSERTION_SOURCE
            or value["fixture_path"] != _HARD2G_OBSERVER_FIXTURE_PATH
        ):
            return _INVALID
        return MappingProxyType(
            {
                "module": value["module"],
                "python_stage": value["python_stage"],
                "source_paths": source_paths,
                "selftest_paths": selftest_paths,
                "policy_path": value["policy_path"],
                "image_inspect_argvs": image_inspect_argvs,
                "canonical_build_input_paths": build_input_paths,
                "snapshot_assertion_source": value["snapshot_assertion_source"],
                "fixture_path": value["fixture_path"],
            }
        )
    except Exception:
        return _INVALID


def _string_tuple(value: object) -> tuple[str, ...] | object:
    if (
        type(value) is not list
        or any(type(item) is not str for item in value)
        or len(value) != len(set(value))
    ):
        return _INVALID
    return tuple(value)


def _nested_string_argvs(value: object) -> tuple[tuple[str, ...], ...] | object:
    if type(value) is not list:
        return _INVALID
    result: list[tuple[str, ...]] = []
    for argv in value:
        if type(argv) is not list or any(type(item) is not str for item in argv):
            return _INVALID
        result.append(tuple(argv))
    return tuple(result)


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


def _compose_service_inventory_is_closed(services: object) -> bool:
    try:
        return isinstance(services, Mapping) and tuple(services) == (
            _COMPOSE_SERVICE_NAMES
        )
    except Exception:
        return False


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
                        "GOENV": "off",
                        "GOFLAGS": "-p=1",
                        "GOPROXY": "off",
                        "GOTOOLCHAIN": "local",
                        "GOTMPDIR": "/var/tmp/hub-go",
                        "GOWORK": "off",
                    }
                ),
                "cpus": 6.0,
                "mem_limit": "2g",
                "memswap_limit": "2g",
                "tmpfs": (
                    "/tmp:rw,noexec,nosuid,nodev,size=64m",
                    "/var/cache/hub-go:rw,noexec,nosuid,nodev,"
                    "size=768m,uid=10001,gid=10001,mode=1770",
                    "/var/tmp/hub-go:rw,nosuid,nodev,exec,"
                    "size=384m,uid=10001,gid=10001,mode=1770",
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
                        "COPY testframework/docker/backend_unit/ "
                        "/usr/local/lib/hub/backend-unit/",
                        "RUN find /usr/local/lib/hub/backend-unit -type f -exec "
                        "sed -i 's/\\r$//' {} + && chmod 0555 "
                        "/usr/local/lib/hub/backend-unit/runtime.sh "
                        "/usr/local/lib/hub/backend-unit/plan.sh "
                        "/usr/local/lib/hub/backend-unit/execute.sh "
                        "/usr/local/lib/hub/backend-unit/receipt.sh",
                        "ENV GOCACHE=/var/cache/hub-go",
                        "ENV GOTMPDIR=/var/tmp/hub-go",
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
                        "COPY testframework/observers/ ./observers/",
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
                        "ENV PYTHONPATH=/opt/hub/testframework/src:/opt/hub/testframework",
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
    return _hard2g_backend_entrypoint_is_closed(text)


def _hard2g_backend_entrypoint_is_closed(text: object) -> bool:
    try:
        if (
            type(text) is not str
            or "\r" in text
            or not text.endswith("\n")
            or not _shell_source_is_safe(text)
        ):
            return False
        lines = tuple(text.splitlines())
        return lines == (
            "#!/bin/sh",
            "set -eu",
            '[ "$#" -eq 0 ] || exit 64',
            "unset GOMAXPROCS",
            ". /usr/local/lib/hub/backend-unit/runtime.sh",
            ". /usr/local/lib/hub/backend-unit/plan.sh",
            ". /usr/local/lib/hub/backend-unit/execute.sh",
            ". /usr/local/lib/hub/backend-unit/receipt.sh",
            "backend_unit_main",
        )
    except Exception:
        return False


def _hard2g_runtime_budget_assignments() -> tuple[str, ...] | object:
    try:
        budget = _HARD2G_REPORT_BUDGET
        if not _hard2g_report_budget_is_closed(budget):
            return _INVALID
        total = _mapping_value(budget, "total_bytes")
        metadata = _mapping_value(budget, "reserved_metadata_bytes")
        streams = _mapping_value(budget, "streams")
        if (
            type(total) is not int
            or type(metadata) is not int
            or not isinstance(streams, Mapping)
        ):
            return _INVALID
        assignments = [
            f"REPORT_BUDGET_BYTES={total}",
            f"REPORT_METADATA_BUDGET_BYTES={metadata}",
        ]
        for stage in ("build", "vet", "unit", "race"):
            stage_budget = _mapping_value(streams, stage)
            if not isinstance(stage_budget, Mapping):
                return _INVALID
            for stream in ("stdout", "stderr"):
                cap = _mapping_value(stage_budget, stream)
                if type(cap) is not int:
                    return _INVALID
                assignments.append(
                    f"{stage.upper()}_{stream.upper()}_CAP_BYTES={cap}"
                )
        return tuple(assignments)
    except Exception:
        return _INVALID


def _hard2g_runtime_stage_stream_cap_body(
    assignments: tuple[str, ...],
) -> tuple[str, ...] | object:
    try:
        cap_symbols = tuple(
            assignment.partition("=")[0] for assignment in assignments[2:]
        )
        if len(cap_symbols) != 8 or any(not symbol for symbol in cap_symbols):
            return _INVALID
        cases = tuple(
            f'{stage}:{stream}) printf \'%s\\n\' "${cap_symbols[index]}" ;;'
            for index, (stage, stream) in enumerate(
                (stage, stream)
                for stage in ("build", "vet", "unit", "race")
                for stream in ("stdout", "stderr")
            )
        )
        return (
            'stage="$1"',
            'stream="$2"',
            'case "$stage:$stream" in',
            *cases,
            "*) return 64 ;;",
            "esac",
        )
    except Exception:
        return _INVALID


def _hard2g_runtime_budget_contract_is_closed(text: object) -> bool:
    try:
        assignments = _hard2g_runtime_budget_assignments()
        if (
            type(text) is not str
            or "\r" in text
            or type(assignments) is not tuple
            or not _shell_source_is_safe(text)
        ):
            return False
        lines = _shell_active_lines(text)
        if not isinstance(lines, tuple) or any(
            lines.count(assignment) != 1 for assignment in assignments
        ):
            return False
        symbols = tuple(assignment.partition("=")[0] for assignment in assignments)
        if any(
            line.startswith(f"{symbol}=") and line not in assignments
            for line in lines
            for symbol in symbols
        ):
            return False
        stage_stream_cap = _shell_function_body(lines, "stage_stream_cap")
        check_budget = _shell_function_body(lines, "check_budget")
        return (
            stage_stream_cap
            == _hard2g_runtime_stage_stream_cap_body(assignments)
            and check_budget
            == (
                'report_bytes="$(report_regular_file_bytes)"',
                'metadata_bytes="$(report_metadata_bytes)"',
                '[ "$report_bytes" -le "$REPORT_BUDGET_BYTES" ] || return 75',
                '[ "$metadata_bytes" -le "$REPORT_METADATA_BUDGET_BYTES" ] '
                "|| return 75",
            )
        )
    except Exception:
        return False


def _hard2g_runtime_budget_fixture(assignments: tuple[str, ...]) -> str:
    stage_stream_cap = _hard2g_runtime_stage_stream_cap_body(assignments)
    assert type(stage_stream_cap) is tuple
    return "\n".join(
        (
            "#!/bin/sh",
            "set -eu",
            "",
            *assignments,
            "",
            "stage_stream_cap() {",
            *(f"  {line}" for line in stage_stream_cap),
            "}",
            "",
            "check_budget() {",
            '  report_bytes="$(report_regular_file_bytes)"',
            '  metadata_bytes="$(report_metadata_bytes)"',
            '  [ "$report_bytes" -le "$REPORT_BUDGET_BYTES" ] || return 75',
            '  [ "$metadata_bytes" -le "$REPORT_METADATA_BUDGET_BYTES" ] '
            "|| return 75",
            "}",
            "",
            *(
                line
                for name in _HARD2G_C02_RUNTIME_OWNED_HELPERS
                for line in (f"{name}() {{", "  :", "}", "")
            ),
        )
    )


def _hard2g_container_receipt_fixture() -> str:
    return "\n".join(
        (
            "#!/bin/sh",
            "set -eu",
            "",
            "rename_stage_artifact() {",
            '  mv -- "$run_dir/$stage.$artifact.tmp" "$run_dir/$stage.$artifact"',
            "}",
            "",
            "build_container_receipt() {",
            *(f"  {line}" for line in _HARD2G_RECEIPT_BUDGET_BUILDER_BODY),
            "}",
            "",
            "finalize_container() {",
            "  write_owner",
            "  write_package_plans",
            "  write_stage_artifacts",
            "  write_inventory",
            '  : > "$run_dir/CONTAINER_COMPLETE"',
            "  build_container_receipt",
            '  write_atomic "$run_dir/container-receipt.json" "$receipt"',
            "  return 0",
            "}",
            "",
        )
    )


def _hard2g_c02_durable_budget_runtime_fixture() -> str:
    runtime = _hard2g_runtime_budget_fixture(_hard2g_runtime_budget_assignments())
    insertion_point = "set -eu\n\n"
    assert runtime.count(insertion_point) == 1
    return runtime.replace(
        insertion_point,
        insertion_point + "report_budget_truncated_due_to_budget=false\n\n",
        1,
    )


def _hard2g_c02_durable_budget_execute_fixture() -> str:
    execute = _hard2g_execute_budget_bridge_fixture()
    legacy = "truncation_evidence=budget-stop"
    assert execute.count(legacy) == 3
    return execute.replace(
        legacy,
        "report_budget_truncated_due_to_budget=true",
    )


def _hard2g_execute_budget_bridge_fixture() -> str:
    return '''#!/bin/sh
set -eu

read_bounded_stage_stream() {
  stage="$1"
  stream="$2"
  fifo_path="$3"
  stream_path="$4"
  stream_used_path="$5"
  package_ordinal="$6"
  stream_cap="$(stage_stream_cap "$stage" "$stream")"
  stream_used="$(cat "$stream_used_path")"
  remaining=$((stream_cap - stream_used))
  read_limit=$((remaining + 1))
  stream_tmp="$GOTMPDIR/stream-capture/$stage.$stream.$package_ordinal.tmp"
  if [ "$remaining" -lt 0 ]; then
    : > "$budget_stop_marker"
    return 75
  fi
  if check_budget; then
    :
  else
    : > "$budget_stop_marker"
    return 75
  fi
  if head -c "$read_limit" < "$fifo_path" > "$stream_tmp"; then
    :
  else
    : > "$budget_stop_marker"
    return 75
  fi
  stream_bytes="$(wc -c < "$stream_tmp")"
  if [ "$stream_bytes" -gt "$remaining" ]; then
    : > "$budget_stop_marker"
    return 75
  fi
  stream_used=$((stream_used + stream_bytes))
  if check_budget; then
    cat -- "$stream_tmp" >> "$stream_path"
  else
    : > "$budget_stop_marker"
    return 75
  fi
  if check_budget; then
    printf '%s\n' "$stream_used" > "$stream_used_path"
    rm -f -- "$stream_tmp"
    return 0
  fi
  : > "$budget_stop_marker"
  return 75
}

run_stage() {
  stage="$1"
  plan_path="$2"
  stage_exit=0
  package_ordinal=0
  while IFS= read -r package; do
    package_ordinal=$((package_ordinal + 1))
    child_exit=0
    stdout_reader_exit=0
    stderr_reader_exit=0
    budget_stop_marker="$stage_stop_marker"
    rm -f -- "$budget_stop_marker"
    start_owned_go_child_group "$stage" "$package"
    read_bounded_stage_stream "$stage" stdout "$stdout_fifo" "$stdout_path" "$stdout_used_path" "$package_ordinal" &
    stdout_reader_pid="$!"
    read_bounded_stage_stream "$stage" stderr "$stderr_fifo" "$stderr_path" "$stderr_used_path" "$package_ordinal" &
    stderr_reader_pid="$!"
    if monitor_owned_budget_stop "$go_child_pid" "$budget_stop_marker"; then
      :
    else
      truncation_evidence=budget-stop
      terminate_owned_go_child_group
      reap_owned_go_child_group
      wait "$stdout_reader_pid" || stdout_reader_exit=$?
      wait "$stderr_reader_pid" || stderr_reader_exit=$?
      return 75
    fi
    wait "$stdout_reader_pid" || stdout_reader_exit=$?
    wait "$stderr_reader_pid" || stderr_reader_exit=$?
    if [ -e "$budget_stop_marker" ] || [ "$stdout_reader_exit" -eq 75 ] || [ "$stderr_reader_exit" -eq 75 ]; then
      truncation_evidence=budget-stop
      terminate_owned_go_child_group
      reap_owned_go_child_group
      return 75
    fi
    if check_budget; then
      :
    else
      truncation_evidence=budget-stop
      terminate_owned_go_child_group
      reap_owned_go_child_group
      return 75
    fi
    wait "$go_child_pid" || child_exit=$?
    if [ "$child_exit" -ne 0 ] && [ "$stage_exit" -eq 0 ]; then
      stage_exit="$child_exit"
    fi
  done < "$plan_path"
  return "$stage_exit"
}
'''


def _hard2g_c02_execute_surface_is_closed(text: object) -> bool:
    try:
        if (
            type(text) is not str
            or "\r" in text
            or "\x00" in text
            or any(
                character.isspace() and character not in " \t\n"
                for character in text
            )
            or not text.endswith("\n")
            or not _shell_source_is_safe(text)
            or _HARD2G_C02_EXECUTE_EXTENSION_ALLOWLIST != ()
        ):
            return False
        source_lines = text.splitlines()
        if not source_lines or source_lines[0] != "#!/bin/sh":
            return False
        canonical = _hard2g_c02_durable_budget_execute_fixture()
        lines = _shell_active_lines(text)
        canonical_lines = _shell_active_lines(canonical)
        inventory = _hard2g_c02_function_inventory(text)
        canonical_inventory = _hard2g_c02_function_inventory(canonical)
        expected_inventory = ("read_bounded_stage_stream", "run_stage")
        return (
            type(lines) is tuple
            and type(canonical_lines) is tuple
            and inventory is not None
            and canonical_inventory is not None
            and tuple(inventory) == expected_inventory
            and tuple(canonical_inventory) == expected_inventory
            and lines == canonical_lines
        )
    except Exception:
        return False


def _hard2g_read_bounded_stage_stream_automaton_is_closed(text: object) -> bool:
    try:
        if (
            type(text) is not str
            or "\r" in text
            or not text.endswith("\n")
            or not _shell_source_is_safe(text)
        ):
            return False
        lines = _shell_active_lines(text)
        inventory = _hard2g_c02_function_inventory(text)
        canonical_inventory = _hard2g_c02_function_inventory(
            _hard2g_c02_durable_budget_execute_fixture()
        )
        if (
            type(lines) is not tuple
            or inventory is None
            or canonical_inventory is None
        ):
            return False
        name = "read_bounded_stage_stream"
        canonical_body = canonical_inventory.get(name)
        return (
            lines.count(f"{name}() {{") == 1
            and type(canonical_body) is tuple
            and bool(canonical_body)
            and inventory.get(name) == canonical_body
        )
    except Exception:
        return False


def _hard2g_run_stage_automaton_is_closed(text: object) -> bool:
    try:
        if (
            type(text) is not str
            or "\r" in text
            or not text.endswith("\n")
            or not _shell_source_is_safe(text)
        ):
            return False
        lines = _shell_active_lines(text)
        inventory = _hard2g_c02_function_inventory(text)
        canonical_inventory = _hard2g_c02_function_inventory(
            _hard2g_c02_durable_budget_execute_fixture()
        )
        if (
            type(lines) is not tuple
            or inventory is None
            or canonical_inventory is None
        ):
            return False
        expected_inventory = ("read_bounded_stage_stream", "run_stage")
        body = inventory.get("run_stage")
        canonical_body = canonical_inventory.get("run_stage")
        stdout_wait = 'wait "$stdout_reader_pid" || stdout_reader_exit=$?'
        stderr_wait = 'wait "$stderr_reader_pid" || stderr_reader_exit=$?'
        go_wait = 'wait "$go_child_pid" || child_exit=$?'
        return (
            tuple(inventory) == expected_inventory
            and tuple(canonical_inventory) == expected_inventory
            and lines.count("run_stage() {") == 1
            and type(body) is tuple
            and bool(body)
            and body == canonical_body
            and body.count("report_budget_truncated_due_to_budget=true") == 3
            and body.count(stdout_wait) == 2
            and body.count(stderr_wait) == 2
            and body.count(go_wait) == 1
        )
    except Exception:
        return False


def _hard2g_execute_budget_bridge_is_closed(text: object) -> bool:
    try:
        return (
            _hard2g_c02_execute_surface_is_closed(text)
            and _hard2g_read_bounded_stage_stream_automaton_is_closed(text)
            and _hard2g_run_stage_automaton_is_closed(text)
        )
    except Exception:
        return False


def _hard2g_backend_unit_service_is_closed(service: object) -> bool:
    try:
        return service == _expected_public_service("backend-unit")
    except Exception:
        return False


def _hard2g_runner_helpers_are_closed(repo_root: Path) -> bool:
    try:
        helper_paths = (
            repo_root / "testframework/docker/backend_unit/runtime.sh",
            repo_root / "testframework/docker/backend_unit/plan.sh",
            repo_root / "testframework/docker/backend_unit/execute.sh",
            repo_root / "testframework/docker/backend_unit/receipt.sh",
        )
        helper_sources = tuple(_read_bounded_regular(path) for path in helper_paths)
        if not all(type(source) is str for source in helper_sources):
            return False
        runtime = helper_sources[0]
        plan = helper_sources[1]
        execute = helper_sources[2]
        receipt = helper_sources[-1]
        assert type(runtime) is str
        assert type(plan) is str
        assert type(execute) is str
        assert type(receipt) is str
        return (
            all(_shell_source_is_safe(source) for source in helper_sources)
            and _hard2g_runtime_budget_contract_is_closed(runtime)
            and _hard2g_execute_budget_bridge_is_closed(execute)
            and _hard2g_c02_runtime_helper_ownership_is_closed(
                runtime, plan, execute, receipt
            )
            and _hard2g_report_budget_truncation_receipt_binding_is_closed(
                runtime, plan, execute, receipt
            )
            and _hard2g_container_publication_is_closed(receipt)
        )
    except Exception:
        return False


_HARD2G_DIRECT_PYTHON_FILESYSTEM_READS = frozenset(
    {
        "exists",
        "fdopen",
        "glob",
        "is_dir",
        "is_file",
        "is_symlink",
        "iterdir",
        "listdir",
        "lstat",
        "open",
        "read",
        "read1",
        "read_bytes",
        "readinto",
        "readinto1",
        "readline",
        "readlines",
        "readlink",
        "read_text",
        "rglob",
        "samefile",
        "scandir",
        "stat",
        "walk",
    }
)
_HARD2G_DYNAMIC_EXECUTION_NAMES = frozenset({"eval", "exec"})
_HARD2G_V4_CONTENT_READ_NAMES = frozenset(
    {
        "fdopen",
        "open",
        "read",
        "read1",
        "read_bytes",
        "readinto",
        "readinto1",
        "readline",
        "readlines",
        "read_text",
    }
)
_HARD2G_V4_DANGEROUS_CALLABLE_KINDS = frozenset(
    {"read", "execution", "getattr"}
)
_HARD2G_METADATA_README_COPY_INSTRUCTIONS = (
    "COPY testframework/pyproject.toml testframework/uv.lock "
    "testframework/.python-version testframework/README.md ./",
    "COPY testframework/schemas/ ./schemas/",
    "COPY testframework/docker/ ./docker/",
)
_HARD2G_PYTHON_BASE_COPY_INSTRUCTIONS = frozenset()
_HARD2G_PYTHON_TESTS_CURRENT_NON_METADATA_COPY_INSTRUCTIONS = frozenset(
    {
        "COPY testframework/src/ ./src/",
        "COPY docker-compose.local.yml /opt/hub/docker-compose.local.yml",
        "COPY .planning/phases/130-tf-traceability-gates/PLAN.md "
        "/opt/hub/.planning/phases/130-tf-traceability-gates/PLAN.md",
        "COPY testframework/selftests/ ./selftests/",
        "COPY testframework/tests/ ./tests/",
        "COPY testframework/cases/ ./cases/",
        "COPY testframework/policies/ ./policies/",
        "COPY testframework/load/ ./load/",
        "COPY testframework/docker/entrypoints/python-qa.sh "
        "/usr/local/bin/python-qa",
    }
)
_HARD2G_PYTHON_TESTS_FUTURE_OBSERVER_COPY_INSTRUCTION = (
    "COPY testframework/observers/ ./observers/"
)
_HARD2G_PYTHON_TESTS_COPY_INSTRUCTIONS = frozenset(
    (
        *_HARD2G_METADATA_README_COPY_INSTRUCTIONS,
        *_HARD2G_PYTHON_TESTS_CURRENT_NON_METADATA_COPY_INSTRUCTIONS,
        _HARD2G_PYTHON_TESTS_FUTURE_OBSERVER_COPY_INSTRUCTION,
    )
)
_HARD2G_METADATA_ALLOWED_PYTHON_BASE_EXECUTABLE_INSTRUCTIONS = frozenset(
    {
        "RUN groupadd --gid 10001 hubqa && useradd --uid 10001 --gid "
        "10001 --create-home --home-dir /home/hubqa hubqa",
        "RUN python -m pip install --no-cache-dir uv==0.12.0",
        "RUN mkdir -p /opt/hub/testframework /opt/hub/.venv && "
        "chown -R 10001:10001 /opt/hub/testframework /opt/hub/.venv",
    }
)
_HARD2G_METADATA_ALLOWED_PYTHON_EXECUTABLE_INSTRUCTIONS = frozenset(
    {
        "RUN UV_PROJECT_ENVIRONMENT=/opt/hub/.venv "
        "uv sync --frozen --all-groups --no-install-project",
        "RUN UV_PROJECT_ENVIRONMENT=/opt/hub/.venv "
        "uv sync --frozen --all-groups",
        "RUN find /opt/hub/docker-compose.local.yml "
        "/opt/hub/.planning/phases/130-tf-traceability-gates/PLAN.md "
        "/opt/hub/testframework/docker -type f -exec "
        "sed -i 's/\\r$//' {} +",
        "RUN find /opt/hub/testframework/cases "
        "/opt/hub/testframework/policies "
        "/opt/hub/testframework/schemas -type f -exec "
        "sed -i 's/\\r$//' {} +",
        "RUN sed -i 's/\\r$//' /usr/local/bin/python-qa && "
        "chmod 0555 /usr/local/bin/python-qa && "
        "chown -R 10001:10001 /opt/hub/testframework /opt/hub/.venv",
        'ENTRYPOINT ["/usr/local/bin/python-qa"]',
    }
)


def _hard2g_observer_source_manifest_is_closed(
    repo_root: Path,
    source_paths: object,
) -> bool:
    try:
        if type(source_paths) is not tuple or source_paths != _HARD2G_OBSERVER_SOURCE_PATHS:
            return False
        discovered = _regular_files_below(
            repo_root / "testframework/observers",
            ".py",
            (),
        )
        expected = tuple(sorted(repo_root / path for path in source_paths))
        return isinstance(discovered, tuple) and tuple(sorted(discovered)) == expected
    except Exception:
        return False


def _hard2g_observer_source_has_closed_read_authority(
    relative: object,
    source: object,
) -> bool:
    try:
        if (
            type(relative) is not str
            or relative not in _HARD2G_OBSERVER_SOURCE_PATHS
            or type(source) is not str
        ):
            return False
        return _hard2g_v4_observer_source_is_closed(relative, source)
    except Exception:
        return False


def _hard2g_observer_sources_have_closed_read_authority(
    source_paths: object,
    sources: object,
) -> bool:
    try:
        return (
            type(source_paths) is tuple
            and source_paths == _HARD2G_OBSERVER_SOURCE_PATHS
            and type(sources) is tuple
            and len(sources) == len(source_paths)
            and all(
                _hard2g_observer_source_has_closed_read_authority(path, source)
                for path, source in zip(source_paths, sources)
            )
        )
    except Exception:
        return False


def _hard2g_authority_aliases(
    tree: ast.AST,
) -> tuple[frozenset[str], frozenset[str], frozenset[str]]:
    reads: set[str] = set()
    executions: set[str] = set()
    paths: set[str] = set()
    for node in ast.walk(tree):
        if not isinstance(node, ast.ImportFrom) or type(node.module) is not str:
            continue
        for alias in node.names:
            bound = alias.asname or alias.name
            if (
                node.module in {"builtins", "io", "os"}
                and alias.name in _HARD2G_DIRECT_PYTHON_FILESYSTEM_READS
            ):
                reads.add(bound)
            if (
                node.module == "builtins"
                and alias.name in _HARD2G_DYNAMIC_EXECUTION_NAMES
            ):
                executions.add(bound)
            if (
                node.module in {"pathlib", "os.path"}
                and alias.name in {"Path", "PurePath", "join"}
            ):
                paths.add(bound)
    return frozenset(reads), frozenset(executions), frozenset(paths)


def _hard2g_python_call_is_direct_filesystem_read(
    call: ast.Call,
    read_aliases: frozenset[str],
) -> bool:
    if isinstance(call.func, ast.Name):
        return call.func.id == "open" or call.func.id in read_aliases
    if isinstance(call.func, ast.Attribute):
        return call.func.attr in _HARD2G_DIRECT_PYTHON_FILESYSTEM_READS
    return _hard2g_getattr_call_names(call, _HARD2G_DIRECT_PYTHON_FILESYSTEM_READS)


def _hard2g_python_call_is_dynamic_execution(
    call: ast.Call,
    execution_aliases: frozenset[str],
) -> bool:
    if isinstance(call.func, ast.Name):
        return (
            call.func.id in _HARD2G_DYNAMIC_EXECUTION_NAMES
            or call.func.id in execution_aliases
        )
    if isinstance(call.func, ast.Attribute):
        return call.func.attr in _HARD2G_DYNAMIC_EXECUTION_NAMES
    return _hard2g_getattr_call_names(call, _HARD2G_DYNAMIC_EXECUTION_NAMES)


def _hard2g_getattr_call_names(call: ast.Call, names: frozenset[str]) -> bool:
    getter = call.func
    if (
        not isinstance(getter, ast.Call)
        or not isinstance(getter.func, ast.Name)
        or getter.func.id != "getattr"
        or len(getter.args) < 2
    ):
        return False
    return _hard2g_ast_value_matches_names(getter.args[1], names)


def _hard2g_safe_paths_has_no_readme_authority(
    tree: ast.AST,
    calls: tuple[ast.Call, ...],
    read_aliases: frozenset[str],
    path_aliases: frozenset[str],
) -> bool:
    readme_names = _hard2g_readme_bound_names(tree)
    if any(
        isinstance(node, ast.BinOp)
        and isinstance(node.op, ast.Div)
        and _hard2g_ast_value_mentions_readme(node, readme_names)
        for node in ast.walk(tree)
    ):
        return False
    return not any(
        (
            _hard2g_python_call_is_direct_filesystem_read(call, read_aliases)
            or _hard2g_python_call_constructs_path(call, path_aliases)
        )
        and any(
            _hard2g_ast_value_mentions_readme(argument, readme_names)
            for argument in (*call.args, *(keyword.value for keyword in call.keywords))
        )
        for call in calls
    )


def _hard2g_readme_bound_names(tree: ast.AST) -> frozenset[str]:
    names: set[str] = set()
    for node in ast.walk(tree):
        value = node.value if isinstance(node, (ast.Assign, ast.AnnAssign)) else None
        if value is None or not _hard2g_ast_value_mentions_readme(value, frozenset()):
            continue
        targets = node.targets if isinstance(node, ast.Assign) else (node.target,)
        names.update(target.id for target in targets if isinstance(target, ast.Name))
    return frozenset(names)


def _hard2g_ast_value_mentions_readme(
    value: ast.AST,
    readme_names: frozenset[str],
) -> bool:
    if any(isinstance(node, ast.Name) and node.id in readme_names for node in ast.walk(value)):
        return True
    return "readme" in "".join(
        character
        for character in ast.unparse(value).casefold()
        if character.isalpha()
    )


def _hard2g_ast_value_matches_names(value: ast.AST, names: frozenset[str]) -> bool:
    compact = "".join(
        character
        for character in ast.unparse(value).casefold()
        if character.isalpha()
    )
    return any(
        "".join(character for character in name if character.isalpha()) in compact
        for name in names
    )


def _hard2g_python_call_constructs_path(
    call: ast.Call,
    path_aliases: frozenset[str],
) -> bool:
    if isinstance(call.func, ast.Name):
        return call.func.id in {"Path", "PurePath"} | path_aliases
    if not isinstance(call.func, ast.Attribute):
        return False
    return call.func.attr in {"joinpath", "with_name", "with_suffix"} or (
        call.func.attr == "join"
        and (
            isinstance(call.func.value, ast.Name)
            and call.func.value.id in path_aliases
            or isinstance(call.func.value, ast.Attribute)
            and call.func.value.attr == "path"
        )
    )


def _hard2g_v4_lexical_shape_is_closed(tree: ast.AST) -> bool:
    if not isinstance(tree, ast.Module):
        return False
    top_level_function_ids = {
        id(node)
        for node in tree.body
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef))
    }
    for node in ast.walk(tree):
        if isinstance(node, (ast.ClassDef, ast.Global, ast.Nonlocal)):
            return False
        if isinstance(node, ast.Call) and (
            any(isinstance(argument, ast.Starred) for argument in node.args)
            or any(keyword.arg is None for keyword in node.keywords)
        ):
            return False
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)) and (
            id(node) not in top_level_function_ids
            or node.args.vararg is not None
            or node.args.kwarg is not None
        ):
            return False
    return True


def _hard2g_v4_observer_source_is_closed(relative: str, source: str) -> bool:
    tree = ast.parse(source)
    if not _hard2g_v4_lexical_shape_is_closed(tree):
        return False
    aliases = _hard2g_v4_authority_aliases(tree)
    module_calls = _hard2g_v4_scope_calls(tree)
    all_calls = tuple(node for node in ast.walk(tree) if isinstance(node, ast.Call))
    function_aliases = _hard2g_v4_function_aliases(tree, aliases)
    scoped_calls = _hard2g_v4_scoped_call_aliases(tree, function_aliases)
    if any(_hard2g_v4_bad_indirection(call, aliases) for call in module_calls):
        return False
    if any(
        _hard2g_v4_bad_indirection(call, scope_aliases)
        for call, scope_aliases in scoped_calls
    ):
        return False
    if relative == _HARD2G_OBSERVER_SAFE_PATHS_SOURCE_PATH:
        return _hard2g_v4_safe_paths_boundary_is_closed(tree, all_calls, aliases)
    if any(_hard2g_v4_direct_read(call, aliases) for call in module_calls):
        return False
    return not any(
        _hard2g_v4_direct_read(call, scope_aliases)
        for call, scope_aliases in scoped_calls
    )


def _hard2g_v4_authority_aliases(tree: ast.AST) -> Mapping[str, frozenset[str]]:
    aliases: dict[str, set[str]] = {
        "read": set(),
        "execution": set(),
        "getattr": set(),
        "builtins": set(),
    }
    nodes = _hard2g_v4_scope_nodes(tree)
    assignments = tuple(
        node
        for node in nodes
        if isinstance(node, (ast.Assign, ast.AnnAssign, ast.NamedExpr))
    )
    _hard2g_v4_seed_import_aliases(nodes, aliases)
    for _ in range(len(assignments) + 1):
        changed = False
        for assignment in assignments:
            changed = _hard2g_v4_bind_assignment(assignment, aliases) or changed
        if not changed:
            break
    return MappingProxyType(
        {name: frozenset(bound) for name, bound in aliases.items()}
    )


def _hard2g_v4_scope_nodes(scope: ast.AST) -> tuple[ast.AST, ...]:
    if not isinstance(scope, (ast.Module, ast.FunctionDef, ast.AsyncFunctionDef)):
        return ()
    nodes: list[ast.AST] = []

    def visit(node: ast.AST) -> None:
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)):
            return
        nodes.append(node)
        for child in ast.iter_child_nodes(node):
            visit(child)

    for statement in scope.body:
        visit(statement)
    return tuple(nodes)


def _hard2g_v4_scope_calls(scope: ast.AST) -> tuple[ast.Call, ...]:
    return tuple(
        node
        for node in _hard2g_v4_scope_nodes(scope)
        if isinstance(node, ast.Call)
    )


def _hard2g_v4_seed_import_aliases(
    nodes: tuple[ast.AST, ...],
    aliases: Mapping[str, frozenset[str] | set[str]],
) -> None:
    for node in nodes:
        if isinstance(node, ast.Import):
            for imported in node.names:
                if imported.name == "builtins":
                    aliases["builtins"].add(imported.asname or imported.name)  # type: ignore[attr-defined]
        elif isinstance(node, ast.ImportFrom) and type(node.module) is str:
            for imported in node.names:
                bound = imported.asname or imported.name
                if (
                    node.module in {"builtins", "io", "os"}
                    and imported.name in _HARD2G_DIRECT_PYTHON_FILESYSTEM_READS
                ):
                    aliases["read"].add(bound)  # type: ignore[attr-defined]
                if (
                    node.module == "builtins"
                    and imported.name in _HARD2G_DYNAMIC_EXECUTION_NAMES
                ):
                    aliases["execution"].add(bound)  # type: ignore[attr-defined]
                if node.module == "builtins" and imported.name == "getattr":
                    aliases["getattr"].add(bound)  # type: ignore[attr-defined]


def _hard2g_v4_top_level_functions(
    tree: ast.AST,
) -> tuple[ast.FunctionDef | ast.AsyncFunctionDef, ...]:
    if not isinstance(tree, ast.Module):
        return ()
    return tuple(
        node
        for node in tree.body
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef))
    )


def _hard2g_v4_function_parameter_names(
    function: ast.FunctionDef | ast.AsyncFunctionDef,
) -> frozenset[str]:
    arguments = function.args
    names = {
        argument.arg
        for argument in (
            *arguments.posonlyargs,
            *arguments.args,
            *arguments.kwonlyargs,
        )
    }
    if arguments.vararg is not None:
        names.add(arguments.vararg.arg)
    if arguments.kwarg is not None:
        names.add(arguments.kwarg.arg)
    return frozenset(names)


def _hard2g_v4_function_aliases(
    tree: ast.AST,
    global_aliases: Mapping[str, frozenset[str]],
) -> Mapping[
    ast.FunctionDef | ast.AsyncFunctionDef,
    Mapping[str, frozenset[str]],
]:
    functions = _hard2g_v4_top_level_functions(tree)
    functions_by_name = {function.name: function for function in functions}
    overlays: dict[
        ast.FunctionDef | ast.AsyncFunctionDef,
        dict[str, set[str]],
    ] = {
        function: {kind: set(names) for kind, names in global_aliases.items()}
        for function in functions
    }
    scoped_nodes = {
        function: _hard2g_v4_scope_nodes(function)
        for function in functions
    }
    for function, aliases in overlays.items():
        parameter_names = _hard2g_v4_function_parameter_names(function)
        for kind in _HARD2G_V4_DANGEROUS_CALLABLE_KINDS:
            aliases[kind].difference_update(parameter_names)
        _hard2g_v4_seed_import_aliases(scoped_nodes[function], aliases)
    module_calls = _hard2g_v4_scope_calls(tree)
    function_calls = {
        function: tuple(
            node for node in scoped_nodes[function] if isinstance(node, ast.Call)
        )
        for function in functions
    }
    local_assignments = {
        function: tuple(
            node
            for node in scoped_nodes[function]
            if isinstance(node, (ast.Assign, ast.AnnAssign, ast.NamedExpr))
        )
        for function in functions
    }
    for _ in range(
        len(module_calls)
        + sum(len(calls) + len(assignments) for calls, assignments in zip(
            function_calls.values(),
            local_assignments.values(),
        ))
        + 1
    ):
        changed = False
        for function, assignments in local_assignments.items():
            for assignment in assignments:
                changed = _hard2g_v4_bind_assignment(
                    assignment,
                    overlays[function],
                ) or changed
        for call in module_calls:
            function = _hard2g_v4_called_function(call, functions_by_name)
            if function is not None:
                changed = _hard2g_v4_bind_function_call(
                    function,
                    call,
                    overlays[function],
                    global_aliases,
                    global_aliases,
                ) or changed
        for caller, calls in function_calls.items():
            caller_aliases = overlays[caller]
            for call in calls:
                function = _hard2g_v4_called_function(call, functions_by_name)
                if function is not None:
                    changed = _hard2g_v4_bind_function_call(
                        function,
                        call,
                        overlays[function],
                        caller_aliases,
                        global_aliases,
                    ) or changed
        if not changed:
            break
    return MappingProxyType(
        {
            function: MappingProxyType(
                {kind: frozenset(names) for kind, names in aliases.items()}
            )
            for function, aliases in overlays.items()
        }
    )


def _hard2g_v4_scoped_call_aliases(
    tree: ast.AST,
    function_aliases: Mapping[
        ast.FunctionDef | ast.AsyncFunctionDef,
        Mapping[str, frozenset[str]],
    ],
) -> tuple[tuple[ast.Call, Mapping[str, frozenset[str]]], ...]:
    pairs: list[tuple[ast.Call, Mapping[str, frozenset[str]]]] = []
    for function in _hard2g_v4_top_level_functions(tree):
        aliases = function_aliases.get(function)
        if aliases is None:
            continue
        pairs.extend(
            (call, aliases) for call in _hard2g_v4_scope_calls(function)
        )
    return tuple(pairs)


def _hard2g_v4_bind_assignment(
    assignment: ast.Assign | ast.AnnAssign | ast.NamedExpr,
    aliases: Mapping[str, frozenset[str] | set[str]],
) -> bool:
    value = assignment.value
    if not isinstance(value, ast.AST):
        return False
    targets = (
        assignment.targets
        if isinstance(assignment, ast.Assign)
        else (assignment.target,)
    )
    changed = False
    for target in targets:
        changed = _hard2g_v4_bind_value_to_target(target, value, aliases) or changed
    return changed


def _hard2g_v4_bind_value_to_target(
    target: ast.AST,
    value: ast.AST,
    aliases: Mapping[str, frozenset[str] | set[str]],
) -> bool:
    if isinstance(target, ast.Name):
        return _hard2g_v4_bind_kinds(
            target.id,
            _hard2g_v4_callable_kinds(value, aliases),
            aliases,
        )
    if isinstance(target, ast.Starred):
        return _hard2g_v4_bind_value_to_target(target.value, value, aliases)
    if not isinstance(target, (ast.Tuple, ast.List)):
        return False
    if isinstance(value, (ast.Tuple, ast.List)) and len(target.elts) == len(value.elts):
        changed = False
        for element, item in zip(target.elts, value.elts):
            changed = _hard2g_v4_bind_value_to_target(element, item, aliases) or changed
        return changed
    changed = False
    for element in target.elts:
        changed = _hard2g_v4_bind_value_to_target(element, value, aliases) or changed
    return changed


def _hard2g_v4_bind_kinds(
    name: str,
    kinds: frozenset[str],
    aliases: Mapping[str, frozenset[str] | set[str]],
) -> bool:
    changed = False
    for kind in kinds & _HARD2G_V4_DANGEROUS_CALLABLE_KINDS:
        bound = aliases[kind]
        if name not in bound:
            bound.add(name)  # type: ignore[attr-defined]
            changed = True
    return changed


def _hard2g_v4_called_function(
    call: ast.Call,
    functions: Mapping[str, ast.FunctionDef | ast.AsyncFunctionDef],
) -> ast.FunctionDef | ast.AsyncFunctionDef | None:
    if not isinstance(call.func, ast.Name):
        return None
    return functions.get(call.func.id)


def _hard2g_v4_bind_function_call(
    function: ast.FunctionDef | ast.AsyncFunctionDef,
    call: ast.Call,
    target_aliases: Mapping[str, frozenset[str] | set[str]],
    caller_aliases: Mapping[str, frozenset[str] | set[str]],
    global_aliases: Mapping[str, frozenset[str] | set[str]],
) -> bool:
    positional = (*function.args.posonlyargs, *function.args.args)
    defaults = {
        argument.arg: default
        for argument, default in zip(
            positional[-len(function.args.defaults) :],
            function.args.defaults,
        )
    }
    keywords = {
        keyword.arg: keyword.value
        for keyword in call.keywords
        if keyword.arg is not None
    }
    changed = False
    for index, argument in enumerate(positional):
        if index < len(call.args):
            value = call.args[index]
            value_aliases = caller_aliases
        elif argument.arg in keywords:
            value = keywords[argument.arg]
            value_aliases = caller_aliases
        else:
            value = defaults.get(argument.arg)
            value_aliases = global_aliases
        if isinstance(value, ast.AST):
            changed = _hard2g_v4_bind_kinds(
                argument.arg,
                _hard2g_v4_callable_kinds(value, value_aliases),
                target_aliases,
            ) or changed
    for argument, default in zip(function.args.kwonlyargs, function.args.kw_defaults):
        if argument.arg in keywords:
            value = keywords[argument.arg]
            value_aliases = caller_aliases
        else:
            value = default
            value_aliases = global_aliases
        if isinstance(value, ast.AST):
            changed = _hard2g_v4_bind_kinds(
                argument.arg,
                _hard2g_v4_callable_kinds(value, value_aliases),
                target_aliases,
            ) or changed
    return changed


def _hard2g_v4_assignment_targets(
    assignment: ast.Assign | ast.AnnAssign | ast.NamedExpr,
) -> frozenset[str]:
    targets = assignment.targets if isinstance(assignment, ast.Assign) else (assignment.target,)
    return frozenset(target.id for target in targets if isinstance(target, ast.Name))


def _hard2g_v4_callable_kinds(
    value: ast.AST,
    aliases: Mapping[str, frozenset[str] | set[str]],
) -> frozenset[str]:
    kinds: set[str] = set()
    if isinstance(value, ast.Name):
        if value.id == "open" or value.id in aliases["read"]:
            kinds.add("read")
        if value.id in _HARD2G_DYNAMIC_EXECUTION_NAMES or value.id in aliases["execution"]:
            kinds.add("execution")
        if value.id == "getattr" or value.id in aliases["getattr"]:
            kinds.add("getattr")
    elif isinstance(value, ast.Attribute):
        if value.attr in _HARD2G_DIRECT_PYTHON_FILESYSTEM_READS:
            kinds.add("read")
        if value.attr in _HARD2G_DYNAMIC_EXECUTION_NAMES:
            kinds.add("execution")
        if (
            value.attr == "getattr"
            and isinstance(value.value, ast.Name)
            and value.value.id in aliases["builtins"]
        ):
            kinds.add("getattr")
    elif isinstance(value, ast.Subscript):
        if (
            _hard2g_v4_is_builtins_namespace(value.value, aliases)
            and isinstance(value.slice, ast.Constant)
            and type(value.slice.value) is str
        ):
            name = value.slice.value
            if name in _HARD2G_DIRECT_PYTHON_FILESYSTEM_READS:
                kinds.add("read")
            if name in _HARD2G_DYNAMIC_EXECUTION_NAMES:
                kinds.add("execution")
            if name == "getattr":
                kinds.add("getattr")
        kinds.update(_hard2g_v4_callable_kinds(value.value, aliases))
    elif isinstance(value, (ast.Tuple, ast.List, ast.Set)):
        for element in value.elts:
            kinds.update(_hard2g_v4_callable_kinds(element, aliases))
    elif isinstance(value, ast.Dict):
        for element in value.values:
            if isinstance(element, ast.AST):
                kinds.update(_hard2g_v4_callable_kinds(element, aliases))
    elif isinstance(value, ast.IfExp):
        kinds.update(_hard2g_v4_callable_kinds(value.body, aliases))
        kinds.update(_hard2g_v4_callable_kinds(value.orelse, aliases))
    elif isinstance(value, ast.Call):
        resolved = _hard2g_v4_getattr_method_kind(value, aliases)
        if resolved in {"read", "execution"}:
            kinds.add(resolved)
        elif resolved == "unknown":
            kinds.add("getattr")
    return frozenset(kinds)


def _hard2g_v4_is_builtins_namespace(
    value: ast.AST,
    aliases: Mapping[str, frozenset[str] | set[str]],
) -> bool:
    return isinstance(value, ast.Name) and (
        value.id == "__builtins__" or value.id in aliases["builtins"]
    )


def _hard2g_v4_getattr_method_kind(
    call: ast.Call,
    aliases: Mapping[str, frozenset[str] | set[str]],
) -> str | None:
    if "getattr" not in _hard2g_v4_callable_kinds(call.func, aliases):
        return None
    if (
        len(call.args) < 2
        or not isinstance(call.args[1], ast.Constant)
        or type(call.args[1].value) is not str
    ):
        return "unknown"
    method = call.args[1].value
    if method in _HARD2G_DIRECT_PYTHON_FILESYSTEM_READS:
        return "read"
    if method in _HARD2G_DYNAMIC_EXECUTION_NAMES:
        return "execution"
    return "allowed"


def _hard2g_v4_bad_indirection(
    call: ast.Call,
    aliases: Mapping[str, frozenset[str] | set[str]],
) -> bool:
    kinds = _hard2g_v4_callable_kinds(call.func, aliases)
    if "execution" in kinds:
        return True
    return "getattr" in kinds and (
        _hard2g_v4_getattr_method_kind(call, aliases) != "allowed"
    )


def _hard2g_v4_direct_read(
    call: ast.Call,
    aliases: Mapping[str, frozenset[str] | set[str]],
) -> bool:
    return "read" in _hard2g_v4_callable_kinds(call.func, aliases)


def _hard2g_v4_safe_paths_boundary_is_closed(
    tree: ast.AST,
    calls: tuple[ast.Call, ...],
    aliases: Mapping[str, frozenset[str] | set[str]],
) -> bool:
    functions = tuple(
        node
        for node in ast.walk(tree)
        if isinstance(node, ast.FunctionDef) and node.name == "open_regular_relative"
    )
    if len(functions) != 1:
        return False
    function = functions[0]
    if any(
        isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef))
        for node in ast.walk(function)
        if node is not function
    ) or not {"root", "relative"} <= {arg.arg for arg in function.args.args}:
        return False
    candidates = tuple(
        (statement, _hard2g_v4_candidate_name(statement))
        for statement in function.body
        if _hard2g_v4_candidate_name(statement) is not None
    )
    if len(candidates) != 1:
        return False
    candidate_statement, candidate = candidates[0]
    assert type(candidate) is str
    if sum(
        candidate in _hard2g_v4_assignment_targets(node)
        for node in ast.walk(function)
        if isinstance(node, (ast.Assign, ast.AnnAssign, ast.NamedExpr))
    ) != 1:
        return False
    guards = tuple(
        statement
        for statement in function.body
        if _hard2g_v4_is_readme_guard(statement, candidate)
    )
    reads = tuple(
        call for call in calls if _hard2g_v4_content_read(call, aliases)
    )
    return (
        len(guards) == len(reads) == 1
        and candidate_statement.lineno < guards[0].lineno < reads[0].lineno
        and reads[0] in ast.walk(function)
        and isinstance(reads[0].func, ast.Attribute)
        and reads[0].func.attr == "open"
        and isinstance(reads[0].func.value, ast.Name)
        and reads[0].func.value.id == "os"
        and len(reads[0].args) >= 2
        and isinstance(reads[0].args[0], ast.Name)
        and reads[0].args[0].id == candidate
    )


def _hard2g_v4_candidate_name(statement: ast.stmt) -> str | None:
    if (
        not isinstance(statement, ast.Assign)
        or len(statement.targets) != 1
        or not isinstance(statement.targets[0], ast.Name)
        or not isinstance(statement.value, ast.BinOp)
        or not isinstance(statement.value.op, ast.Div)
        or not isinstance(statement.value.left, ast.Name)
        or statement.value.left.id != "root"
        or not isinstance(statement.value.right, ast.Name)
        or statement.value.right.id != "relative"
    ):
        return None
    return statement.targets[0].id


def _hard2g_v4_is_readme_guard(statement: ast.stmt, candidate: str) -> bool:
    if (
        not isinstance(statement, ast.If)
        or len(statement.body) != 1
        or not isinstance(statement.body[0], ast.Raise)
        or not isinstance(statement.test, ast.Compare)
        or len(statement.test.ops) != 1
        or not isinstance(statement.test.ops[0], ast.Eq)
        or len(statement.test.comparators) != 1
        or not isinstance(statement.test.comparators[0], ast.Constant)
        or statement.test.comparators[0].value != "readme.md"
    ):
        return False
    value = statement.test.left
    return (
        isinstance(value, ast.Call)
        and not value.args
        and isinstance(value.func, ast.Attribute)
        and value.func.attr == "casefold"
        and isinstance(value.func.value, ast.Attribute)
        and value.func.value.attr == "name"
        and isinstance(value.func.value.value, ast.Name)
        and value.func.value.value.id == candidate
    )


def _hard2g_v4_content_read(
    call: ast.Call,
    aliases: Mapping[str, frozenset[str] | set[str]],
) -> bool:
    return _hard2g_v4_direct_read(call, aliases) and (
        not isinstance(call.func, ast.Attribute)
        or call.func.attr in _HARD2G_V4_CONTENT_READ_NAMES
    )


def _hard2g_container_publication_is_closed(text: object) -> bool:
    try:
        if type(text) is not str or not _shell_source_is_safe(text):
            return False
        lines = _shell_active_lines(text)
        if not isinstance(lines, tuple):
            return False
        if any(
            not _hard2g_shell_line_has_no_observer_publication_authority(line)
            for line in lines
        ):
            return False
        if not _hard2g_stage_renames_are_owned(lines):
            return False
        if not _hard2g_compound_publication_grammar_is_closed(lines):
            return False
        finalizer = _shell_function_body(lines, "finalize_container")
        builder = _shell_function_body(lines, "build_container_receipt")
        if (
            not isinstance(finalizer, tuple)
            or builder != _HARD2G_RECEIPT_BUDGET_BUILDER_BODY
        ):
            return False
        required_lines = (
            "write_owner",
            "write_package_plans",
            "write_stage_artifacts",
            "write_inventory",
            ': > "$run_dir/CONTAINER_COMPLETE"',
            "build_container_receipt",
            'write_atomic "$run_dir/container-receipt.json" "$receipt"',
        )
        positions = tuple(
            finalizer.index(line) if finalizer.count(line) == 1 else -1
            for line in required_lines
        )
        ordered_finalizer = (*required_lines, "return 0")
        return (
            all(lines.count(line) == 1 for line in required_lines)
            and lines.count("return 0") == 1
            and all(position >= 0 for position in positions)
            and positions == tuple(sorted(positions))
            and finalizer == ordered_finalizer
        )
    except Exception:
        return False


def _hard2g_readme_metadata_copies_are_closed(text: object) -> bool:
    try:
        if (
            _HARD2G_METADATA_README_PATHS
            != frozenset(
                {
                    "testframework/README.md",
                    "testframework/schemas/README.md",
                    "testframework/docker/README.md",
                }
            )
            or "testframework/runner/README.md" not in _RETIRED_PATHS
        ):
            return False
        stages = _docker_stages_safe(text)
        if not isinstance(stages, Mapping):
            return False
        python_tests = stages.get("python-tests")
        if type(python_tests) is not tuple or len(python_tests) != 2:
            return False
        python_base = stages.get("python-base")
        base_instructions = (
            () if python_base is None else python_base[1]
        )
        if (
            type(base_instructions) is not tuple
            or not all(
                _hard2g_v4_docker_instruction_is_closed(
                    instruction,
                    _HARD2G_METADATA_ALLOWED_PYTHON_BASE_EXECUTABLE_INSTRUCTIONS,
                    _HARD2G_PYTHON_BASE_COPY_INSTRUCTIONS,
                )
                for instruction in base_instructions
            )
        ):
            return False
        instructions = python_tests[1]
        if type(instructions) is not tuple:
            return False
        if not all(
            instruction in instructions
            for instruction in _HARD2G_METADATA_README_COPY_INSTRUCTIONS
        ):
            return False
        return all(
            _hard2g_v4_docker_instruction_is_closed(
                instruction,
                _HARD2G_METADATA_ALLOWED_PYTHON_EXECUTABLE_INSTRUCTIONS,
                _HARD2G_PYTHON_TESTS_COPY_INSTRUCTIONS,
            )
            for instruction in instructions
        )
    except Exception:
        return False


def _hard2g_v4_docker_instruction_is_closed(
    instruction: str,
    executable_allowlist: frozenset[str],
    copy_allowlist: frozenset[str],
) -> bool:
    opcode = instruction.partition(" ")[0]
    if opcode in {"RUN", "CMD", "ENTRYPOINT"}:
        return instruction in executable_allowlist
    if opcode == "COPY":
        return instruction in copy_allowlist
    return opcode in {"ENV", "USER", "WORKDIR"}


def _hard2g_provenance_bindings_are_closed(
    repo_root: Path,
    policy: Mapping[str, object],
    dockerfile: object,
) -> bool:
    try:
        config = _mapping_value(policy, "hard2g_observer")
        if not _hard2g_observer_config_is_closed(config):
            return False
        if not _hard2g_observer_imports_are_python_tests_only(
            repo_root,
            config,
            dockerfile,
        ):
            return False
        observer_policy = _read_json_mapping(
            repo_root / _HARD2G_OBSERVER_POLICY_PATH
        )
        if not isinstance(observer_policy, Mapping):
            return False
        if (
            observer_policy.get("module") != config["module"]
            or _nested_string_argvs(observer_policy.get("image_inspect_argvs"))
            != config["image_inspect_argvs"]
            or _string_tuple(observer_policy.get("canonical_build_input_paths"))
            != config["canonical_build_input_paths"]
        ):
            return False
        observer_test = _read_bounded_regular(
            repo_root / _HARD2G_OBSERVER_SNAPSHOT_ASSERTION_SOURCE
        )
        if type(observer_test) is not str or not _hard2g_task2_provenance_oracle_chain_is_closed(
            observer_test
        ):
            return False
        fixture = _read_json_mapping(repo_root / _HARD2G_OBSERVER_FIXTURE_PATH)
        if not isinstance(fixture, Mapping):
            return False
        fixed_image_ids = fixture.get("fixed_image_ids")
        final_receipt = fixture.get("final_receipt")
        manifest_bytes = fixture.get("canonical_build_input_manifest_bytes")
        manifest_sha256 = fixture.get("canonical_build_input_manifest_sha256")
        return (
            isinstance(fixed_image_ids, Mapping)
            and tuple(fixed_image_ids) == ("python_contract", "backend_unit")
            and _is_sha256_digest(fixed_image_ids["python_contract"])
            and _is_sha256_digest(fixed_image_ids["backend_unit"])
            and fixture.get("bound_backend_image_id")
            == fixed_image_ids["backend_unit"]
            and fixture.get("container_image")
            == fixed_image_ids["backend_unit"]
            and isinstance(final_receipt, Mapping)
            and final_receipt.get("backend_image_id")
            == fixed_image_ids["backend_unit"]
            and all(
                final_receipt.get(field) == fixture.get(field)
                for field in ("run_id", "owner_nonce", "container_id")
            )
            and "observer_receipt_sha256" not in final_receipt
            and type(manifest_bytes) is str
            and _canonical_manifest_paths_are_closed(
                manifest_bytes,
                config["canonical_build_input_paths"],
            )
            and _is_sha256_digest(manifest_sha256)
            and final_receipt.get("canonical_build_input_manifest_sha256")
            == manifest_sha256
            and manifest_sha256
            == "sha256:"
            + hashlib.sha256(
                manifest_bytes.encode("utf-8", errors="strict")
            ).hexdigest()
        )
    except Exception:
        return False


def _hard2g_observer_config_is_closed(value: object) -> bool:
    try:
        return _parse_hard2g_observer_config(
            {
                "module": value["module"],
                "python_stage": value["python_stage"],
                "source_paths": list(value["source_paths"]),
                "selftest_paths": list(value["selftest_paths"]),
                "policy_path": value["policy_path"],
                "image_inspect_argvs": [
                    list(argv) for argv in value["image_inspect_argvs"]
                ],
                "canonical_build_input_paths": list(
                    value["canonical_build_input_paths"]
                ),
                "snapshot_assertion_source": value["snapshot_assertion_source"],
                "fixture_path": value["fixture_path"],
            }
        ) == value
    except Exception:
        return False


def _hard2g_c02_function_inventory(
    source: str,
) -> Mapping[str, tuple[str, ...]] | None:
    try:
        if type(source) is not str:
            return None
        functions: dict[str, tuple[str, ...]] = {}
        active_name: str | None = None
        pending_name: str | None = None
        body: list[str] = []
        for raw_line in source.splitlines():
            line = raw_line.strip()
            if not line or line.startswith("#"):
                continue
            if pending_name is not None:
                if line != "{":
                    return None
                active_name = pending_name
                pending_name = None
                body = []
                continue
            if active_name is not None:
                if line == "}":
                    if active_name in functions:
                        return None
                    functions[active_name] = tuple(body)
                    active_name = None
                    body = []
                    continue
                brace_safe_line = re.sub(
                    r"\$\{[A-Za-z_][A-Za-z0-9_]*\}", "", line
                )
                if (
                    "{" in brace_safe_line
                    or "}" in brace_safe_line
                    or "()" in line
                    or re.match(r"function\b", line)
                ):
                    return None
                body.append(line)
                continue

            if re.match(
                r"(?:if|then|else|elif|fi|case|esac|while|do|done|for|until|select)\b",
                line,
            ) or line.endswith(("; then", "; do")) or re.match(
                r"(?:!\s*)?\(", line
            ) or line.startswith(")"):
                return None
            name: str | None = None
            immediate = False
            if line.endswith("() {"):
                prefix = line[:-4]
                if prefix and prefix[-1] not in " \t":
                    name = prefix
                    immediate = True
                elif prefix.endswith(" "):
                    name = prefix[:-1]
                    immediate = True
            elif line.endswith(" ()"):
                name = line[:-3]

            if name is not None:
                if (
                    not name
                    or not (
                        "A" <= name[0] <= "Z" or "a" <= name[0] <= "z"
                        or name[0] == "_"
                    )
                    or any(
                        not (
                            "A" <= character <= "Z"
                            or "a" <= character <= "z"
                            or "0" <= character <= "9"
                            or character == "_"
                        )
                        for character in name[1:]
                    )
                    or name in functions
                ):
                    return None
                if immediate:
                    active_name = name
                    body = []
                else:
                    pending_name = name
                continue
            if (
                "()" in line
                or "{" in line
                or "}" in line
                or re.match(r"function\b", line)
            ):
                return None
        if active_name is not None or pending_name is not None:
            return None
        return MappingProxyType(functions)
    except Exception:
        return None


def _hard2g_c02_shell_body_lexically_closed(body: object) -> bool:
    try:
        if type(body) is not tuple or not body:
            return False
        name = r"[A-Za-z_][A-Za-z0-9_]*"
        simple_parameter = rf"(?<!\$)\$\{{{name}\}}"
        sanctioned_substitutions = {
            'report_bytes="$(report_regular_file_bytes)"',
            'metadata_bytes="$(report_metadata_bytes)"',
            'stream_cap="$(stage_stream_cap "$stage" "$stream")"',
            'stream_used="$(cat "$stream_used_path")"',
            'stream_bytes="$(wc -c < "$stream_tmp")"',
            'stat_line="$(cat "$stat_path")"',
        }
        arithmetic_assignment = re.compile(
            rf"{name}=\$\(\((?:{name}|[0-9]+) [+-] (?:{name}|[0-9]+)\)\)"
        )
        canonical_pipe_case = (
            '"$go_child_pid (go) Z "*|"$go_child_pid (go) X "*) '
            "return 1 ;;"
        )
        canonical_live_case = '"$go_child_pid (go) "*) return 0 ;;'
        ordinary_case_arm = re.compile(
            rf"(?:[A-Za-z0-9_:-]+|\*)\) (?:printf '%s\\n' \"\$\{{{name}\}}\"|return [0-9]+) ;;"
        )
        forbidden_words = {
            ".",
            "alias",
            "builtin",
            "busybox",
            "command",
            "env",
            "eval",
            "exec",
            "source",
        }
        for line in body:
            if (
                type(line) is not str
                or not line
                or line != line.strip()
                or "\r" in line
                or "\n" in line
                or "\x00" in line
                or line.startswith("#")
                or line.endswith("\\")
                or "<<" in line
                or "`" in line
                or "<(" in line
                or ">(" in line
                or re.search(r"\bfunction\b", line, re.IGNORECASE)
            ):
                return False
            is_sanctioned_substitution = line in sanctioned_substitutions
            is_arithmetic_assignment = bool(
                re.fullmatch(arithmetic_assignment, line)
            )
            if "$(" in line and not (
                is_sanctioned_substitution or is_arithmetic_assignment
            ):
                return False
            unexpanded = re.sub(simple_parameter, "", line)
            if "${" in unexpanded or "{" in unexpanded or "}" in unexpanded:
                return False
            is_case_arm = (
                line in {canonical_pipe_case, canonical_live_case}
                or bool(re.fullmatch(ordinary_case_arm, line))
            )
            if ";;" in line:
                if not is_case_arm or line.count(";") != 2:
                    return False
            elif ";" in line and (
                line.count(";") != 1 or not line.endswith(("; then", "; do"))
            ):
                return False
            if "|" in line.replace("||", "") and line != canonical_pipe_case:
                return False
            words = _shell_words(line)
            if not isinstance(words, tuple):
                return False
            direct_index = 0
            if direct_index < len(words) and words[direct_index] in {
                "if",
                "until",
                "while",
            }:
                direct_index += 1
            if direct_index < len(words) and words[direct_index] == "!":
                direct_index += 1
            while direct_index < len(words) and re.fullmatch(
                rf"{name}=.*", words[direct_index]
            ):
                direct_index += 1
            if not is_case_arm and direct_index < len(words) and (
                words[direct_index].startswith(("<", ">", "&>", "&>>"))
                or (
                    re.fullmatch(r"[0-9]+", words[direct_index])
                    and direct_index + 1 < len(words)
                    and words[direct_index + 1].startswith(("<", ">"))
                )
                or "$" in words[direct_index]
                or re.fullmatch(
                    r"(?:.*/)?(?:busybox|env)", words[direct_index]
                )
            ):
                return False
            if (
                any(word in forbidden_words for word in words)
                or any(
                    re.fullmatch(
                        r"(?:.*/)?(?:sh|bash|dash|ash)", words[index]
                    )
                    and "-c" in words[index + 1 :]
                    for index in range(len(words))
                )
            ):
                return False
            if "&&" in line or "||" in line:
                bracket_atom = r"\[ [^\[\];|&(){}]+ \]"
                if not (
                    re.fullmatch(
                        rf"if {bracket_atom}(?: (?:&&|\|\|) {bracket_atom})*; then",
                        line,
                    )
                    or re.fullmatch(rf"{bracket_atom} \|\| return [0-9]+", line)
                    or re.fullmatch(
                        rf'wait "\${name}" \|\| {name}=\$\?', line
                    )
                ):
                    return False
            if not (
                is_case_arm
                or is_sanctioned_substitution
                or is_arithmetic_assignment
            ) and ("(" in unexpanded or ")" in unexpanded):
                return False
        return True
    except Exception:
        return False


def _hard2g_c02_shell_control_flow_is_closed(body: object) -> bool:
    try:
        if not _hard2g_c02_shell_body_lexically_closed(body):
            return False
        name = r"[A-Za-z_][A-Za-z0-9_]*"
        canonical_pipe_case = (
            '"$go_child_pid (go) Z "*|"$go_child_pid (go) X "*) '
            "return 1 ;;"
        )
        canonical_live_case = '"$go_child_pid (go) "*) return 0 ;;'
        ordinary_case_arm = re.compile(
            rf"(?:[A-Za-z0-9_:-]+|\*)\) (?:printf '%s\\n' \"\$\{{{name}\}}\"|return [0-9]+) ;;"
        )
        reserved_control_words = {
            "if",
            "while",
            "case",
            "then",
            "do",
            "in",
            "elif",
            "for",
            "until",
            "select",
            "function",
            "break",
            "continue",
            "exit",
            "else",
            "fi",
            "done",
            "esac",
        }
        stack: list[tuple[str, bool]] = []
        for line in body:
            words = _shell_words(line)
            if not isinstance(words, tuple) or not words:
                return False
            is_case_arm = (
                line in {canonical_pipe_case, canonical_live_case}
                or bool(re.fullmatch(ordinary_case_arm, line))
            )
            if stack and stack[-1][0] == "CASE":
                if is_case_arm and line.endswith(";;"):
                    stack[-1] = ("CASE", True)
                    continue
                if line == "esac" and stack[-1][1]:
                    stack.pop()
                    continue
                return False
            if line.endswith(";;") or line == "esac":
                return False
            if re.fullmatch(r"if .+; then", line):
                condition_words = words[1:-2]
                condition_index = 0
                if condition_words and condition_words[0] == "!":
                    condition_index += 1
                while condition_index < len(condition_words) and re.fullmatch(
                    r"[A-Za-z_][A-Za-z0-9_]*=.*",
                    condition_words[condition_index],
                ):
                    condition_index += 1
                if (
                    "#" in line
                    or "\\" in line
                    or words[0] != "if"
                    or words[-2:] != (";", "then")
                    or condition_index >= len(condition_words)
                    or condition_words[condition_index] == "!"
                    or condition_words[condition_index]
                    in reserved_control_words
                ):
                    return False
                stack.append(("IF", False))
                continue
            if re.fullmatch(r"while .+; do", line):
                condition_words = words[1:-2]
                condition_index = 0
                if condition_words and condition_words[0] == "!":
                    condition_index += 1
                while condition_index < len(condition_words) and re.fullmatch(
                    r"[A-Za-z_][A-Za-z0-9_]*=.*",
                    condition_words[condition_index],
                ):
                    condition_index += 1
                if (
                    "#" in line
                    or "\\" in line
                    or words[0] != "while"
                    or words[-2:] != (";", "do")
                    or condition_index >= len(condition_words)
                    or condition_words[condition_index] == "!"
                    or condition_words[condition_index]
                    in reserved_control_words
                ):
                    return False
                stack.append(("WHILE", False))
                continue
            if re.fullmatch(r"case .+ in", line):
                if (
                    "#" in line
                    or "\\" in line
                    or words[0] != "case"
                    or len(words) != 3
                    or words[-1] != "in"
                ):
                    return False
                stack.append(("CASE", False))
                continue
            if line == "else":
                if not stack or stack[-1] != ("IF", False):
                    return False
                stack[-1] = ("IF", True)
                continue
            if line == "fi":
                if not stack or stack[-1][0] != "IF":
                    return False
                stack.pop()
                continue
            if line in {"done", 'done < "$plan_path"'}:
                if not stack or stack[-1][0] != "WHILE":
                    return False
                stack.pop()
                continue
            direct_index = 0
            if words[direct_index] == "!":
                return False
            while direct_index < len(words) and re.fullmatch(
                r"[A-Za-z_][A-Za-z0-9_]*=.*", words[direct_index]
            ):
                direct_index += 1
            if direct_index < len(words) and words[direct_index] == "!":
                return False
            if (
                direct_index < len(words)
                and words[direct_index] in reserved_control_words
            ):
                return False
        return not stack
    except Exception:
        return False


def _hard2g_c02_runtime_helper_ownership_is_closed(
    runtime: object,
    plan: object,
    execute: object,
    receipt: object,
) -> bool:
    try:
        sources = (runtime, plan, execute, receipt)
        if any(
            type(source) is not str
            or "\r" in source
            or not source.endswith("\n")
            or not _shell_source_is_safe(source)
            for source in sources
        ):
            return False
        runtime_text, plan_text, execute_text, receipt_text = sources
        assert type(runtime_text) is str
        assert type(plan_text) is str
        assert type(execute_text) is str
        assert type(receipt_text) is str
        active_sources = tuple(_shell_active_lines(source) for source in sources)
        if not all(type(lines) is tuple for lines in active_sources):
            return False
        runtime_lines, plan_lines, execute_lines, receipt_lines = active_sources
        assert type(runtime_lines) is tuple
        assert type(plan_lines) is tuple
        assert type(execute_lines) is tuple
        assert type(receipt_lines) is tuple
        runtime_inventory = _hard2g_c02_function_inventory(runtime_text)
        execute_inventory = _hard2g_c02_function_inventory(execute_text)
        if runtime_inventory is None or execute_inventory is None:
            return False
        if (
            not all(
                name in runtime_inventory and runtime_inventory[name]
                for name in _HARD2G_C02_RUNTIME_OWNED_HELPERS
            )
            or any(
                name in execute_inventory
                for name in _HARD2G_C02_RUNTIME_OWNED_HELPERS
            )
            or any(
                not _hard2g_c02_shell_body_lexically_closed(body)
                or not _hard2g_c02_shell_control_flow_is_closed(body)
                for inventory in (runtime_inventory, execute_inventory)
                for body in inventory.values()
            )
        ):
            return False
        protected_names = _HARD2G_C02_RUNTIME_OWNED_HELPERS
        loader_words = {".", "source", "eval", "exec", "alias", "command", "builtin"}
        declaration_words = {
            "local",
            "export",
            "readonly",
            "declare",
            "typeset",
            "read",
            "unset",
            "alias",
        }
        for source_index, source_lines in enumerate(active_sources):
            for line in source_lines:
                words = _shell_words(line)
                if type(words) is not tuple:
                    return False
                if (
                    any(word in loader_words or word == "function" for word in words)
                    or (source_index != 0 and "trap" in words)
                    or re.search(r"\$[A-Za-z_{][^\s]*\s*\(\)", line)
                ):
                    return False
                for name in protected_names:
                    if any(word.startswith(f"{name}=") for word in words):
                        return False
                    if (
                        any(
                            word in declaration_words
                            and any(
                                target == name or target.startswith(f"{name}=")
                                for target in words[index + 1 :]
                            )
                            for index, word in enumerate(words)
                        )
                        or (
                            words
                            and words[0] in {"for", "select"}
                            and len(words) > 1
                            and words[1] == name
                        )
                        or (
                            words
                            and words[0] == "getopts"
                            and any(target == name for target in words[1:])
                        )
                    ):
                        return False
        identifier_occurrences = {
            name: re.compile(
                rf"(?<![A-Za-z0-9_]){re.escape(name)}(?![A-Za-z0-9_])"
            )
            for name in protected_names
        }
        if any(
            pattern.search(line)
            for source_lines in (plan_lines, receipt_lines)
            for line in source_lines
            for pattern in identifier_occurrences.values()
        ):
            return False
        expected_execute_occurrences = (
            (
                "start_owned_go_child_group",
                'start_owned_go_child_group "$stage" "$package"',
            ),
            (
                "monitor_owned_budget_stop",
                'if monitor_owned_budget_stop "$go_child_pid" "$budget_stop_marker"; then',
            ),
            ("terminate_owned_go_child_group", "terminate_owned_go_child_group"),
            ("reap_owned_go_child_group", "reap_owned_go_child_group"),
            ("terminate_owned_go_child_group", "terminate_owned_go_child_group"),
            ("reap_owned_go_child_group", "reap_owned_go_child_group"),
            ("terminate_owned_go_child_group", "terminate_owned_go_child_group"),
            ("reap_owned_go_child_group", "reap_owned_go_child_group"),
        )
        actual_execute_occurrences = tuple(
            (name, line)
            for line in execute_lines
            for name, pattern in identifier_occurrences.items()
            if pattern.search(line)
        )
        return actual_execute_occurrences == expected_execute_occurrences
    except Exception:
        return False


def _hard2g_c02_durable_budget_fixture_topology_is_closed(
    runtime: object,
    execute: object,
    receipt: object,
) -> bool:
    try:
        sources = (runtime, execute, receipt)
        if any(
            type(source) is not str
            or "\r" in source
            or not source.endswith("\n")
            or not _shell_source_is_safe(source)
            for source in sources
        ):
            return False
        runtime_text, execute_text, receipt_text = sources
        assert type(runtime_text) is str
        assert type(execute_text) is str
        assert type(receipt_text) is str
        if "truncation_evidence=budget-stop" in runtime_text or (
            "truncation_evidence=budget-stop" in execute_text
        ):
            return False
        runtime_lines = _shell_active_lines(runtime_text)
        execute_lines = _shell_active_lines(execute_text)
        receipt_lines = _shell_active_lines(receipt_text)
        if not all(
            type(lines) is tuple
            for lines in (runtime_lines, execute_lines, receipt_lines)
        ):
            return False
        assert type(runtime_lines) is tuple
        assert type(execute_lines) is tuple
        assert type(receipt_lines) is tuple
        runtime_inventory = _hard2g_c02_function_inventory(runtime_text)
        execute_inventory = _hard2g_c02_function_inventory(execute_text)
        if runtime_inventory is None or execute_inventory is None:
            return False
        symbol = "report_budget_truncated_due_to_budget"
        pattern = re.compile(
            rf"(?<![A-Za-z0-9_]){re.escape(symbol)}(?![A-Za-z0-9_])"
        )
        runtime_occurrences = tuple(
            line for line in runtime_lines if pattern.search(line)
        )
        if runtime_occurrences != (f"{symbol}=false",) or any(
            pattern.search(line)
            for body in runtime_inventory.values()
            for line in body
        ):
            return False
        execute_occurrences = tuple(
            line for line in execute_lines if pattern.search(line)
        )
        run_stage = _shell_function_body(execute_lines, "run_stage")
        reader = _shell_function_body(execute_lines, "read_bounded_stage_stream")
        if type(run_stage) is not tuple or type(reader) is not tuple:
            return False
        expected_true = f"{symbol}=true"
        if (
            execute_occurrences != (expected_true, expected_true, expected_true)
            or run_stage.count(expected_true) != 3
            or any(pattern.search(line) for line in reader)
        ):
            return False
        writer = 'write_atomic "$run_dir/container-receipt.json" "$receipt"'
        return (
            receipt_lines.count(writer) == 1
            and _hard2g_container_publication_is_closed(receipt_text)
        )
    except Exception:
        return False


def _hard2g_report_budget_truncation_receipt_binding_is_closed(
    runtime: object,
    plan: object,
    execute: object,
    receipt: object,
) -> bool:
    try:
        if (
            type(plan) is not str
            or "\r" in plan
            or not plan.endswith("\n")
            or not _shell_source_is_safe(plan)
        ):
            return False
        plan_lines = _shell_active_lines(plan)
        if type(plan_lines) is not tuple:
            return False
        symbol = "report_budget_truncated_due_to_budget"
        pattern = re.compile(
            rf"(?<![A-Za-z0-9_]){re.escape(symbol)}(?![A-Za-z0-9_])"
        )
        if any(pattern.search(line) for line in plan_lines):
            return False
        return (
            _hard2g_c02_runtime_helper_ownership_is_closed(
                runtime, plan, execute, receipt
            )
            and _hard2g_runtime_budget_contract_is_closed(runtime)
            and _hard2g_execute_budget_bridge_is_closed(execute)
            and _hard2g_c02_durable_budget_fixture_topology_is_closed(
                runtime, execute, receipt
            )
            and _hard2g_container_publication_is_closed(receipt)
        )
    except Exception:
        return False


def _shell_active_lines(text: str) -> tuple[str, ...] | object:
    try:
        return tuple(
            line.strip()
            for line in text.splitlines()
            if line.strip() and not line.lstrip().startswith("#")
        )
    except Exception:
        return _INVALID


def _shell_function_body(
    lines: tuple[str, ...], name: str
) -> tuple[str, ...] | object:
    try:
        header = f"{name}() {{"
        if lines.count(header) != 1:
            return _INVALID
        start = lines.index(header)
        for end in range(start + 1, len(lines)):
            if lines[end] == "}":
                return lines[start + 1 : end]
        return _INVALID
    except Exception:
        return _INVALID


def _shell_words(line: str) -> tuple[str, ...] | object:
    try:
        lexer = shlex.shlex(line, posix=True, punctuation_chars=True)
        lexer.whitespace_split = True
        lexer.commenters = "#"
        return tuple(lexer)
    except Exception:
        return _INVALID


def _hard2g_shell_line_has_no_observer_publication_authority(line: str) -> bool:
    words = _shell_words(line)
    if not isinstance(words, tuple):
        return False
    if words and (
        words[0].startswith("$")
        or words[0] in {".", "command", "eval", "exec", "source"}
    ):
        return False
    raw = line.casefold()
    if (
        line != 'write_atomic "$run_dir/container-receipt.json" "$receipt"'
        and any(
            fragment in raw
            for fragment in ("final-", "observer-", "receipt.json", "seal")
        )
    ):
        return False
    return not any(
        reserved in word.casefold()
        for word in words
        for reserved in ("final-receipt.json", "observer-receipt.json", "seal")
    )


def _hard2g_stage_rename_is_owned(line: str) -> bool:
    words = _shell_words(line)
    if not isinstance(words, tuple):
        return False
    return words == (
        "mv",
        "--",
        "$run_dir/$stage.$artifact.tmp",
        "$run_dir/$stage.$artifact",
    )


def _hard2g_stage_renames_are_owned(lines: tuple[str, ...]) -> bool:
    try:
        all_moves = tuple(
            line
            for line in lines
            if (words := _shell_words(line))
            and any(word == "mv" or word.endswith("/mv") for word in words)
        )
        if not all_moves:
            return False
        helper = _shell_function_body(lines, "rename_stage_artifact")
        if not isinstance(helper, tuple):
            return False
        helper_moves = tuple(
            line
            for line in helper
            if (words := _shell_words(line))
            and any(word == "mv" or word.endswith("/mv") for word in words)
        )
        return (
            len(all_moves) == 1
            and len(helper_moves) == 1
            and all_moves == helper_moves
            and _hard2g_stage_rename_is_owned(helper_moves[0])
        )
    except Exception:
        return False


def _hard2g_line_mutates_run_dir(words: tuple[str, ...]) -> bool:
    try:
        if any(word.startswith("run_dir=") for word in words):
            return True
        return (
            bool(words)
            and words[0]
            in {"export", "readonly", "local", "declare", "typeset", "unset"}
            and any(word == "run_dir" for word in words[1:])
        )
    except Exception:
        return True


def _hard2g_compound_publication_grammar_is_closed(
    lines: tuple[str, ...],
) -> bool:
    try:
        builder = _shell_function_body(lines, "build_container_receipt")
        if builder != _HARD2G_RECEIPT_BUDGET_BUILDER_BODY:
            return False
        builder_lines = frozenset(builder)
        if any(lines.count(line) != builder.count(line) for line in builder_lines):
            return False
        allowed_commands = {
            "set -eu": ("set", "-eu"),
            "write_owner": ("write_owner",),
            "write_package_plans": ("write_package_plans",),
            "write_stage_artifacts": ("write_stage_artifacts",),
            "write_inventory": ("write_inventory",),
            "build_container_receipt": ("build_container_receipt",),
            ': > "$run_dir/CONTAINER_COMPLETE"': (
                ":",
                ">",
                "$run_dir/CONTAINER_COMPLETE",
            ),
            'write_atomic "$run_dir/container-receipt.json" "$receipt"': (
                "write_atomic",
                "$run_dir/container-receipt.json",
                "$receipt",
            ),
            'mv -- "$run_dir/$stage.$artifact.tmp" "$run_dir/$stage.$artifact"': (
                "mv",
                "--",
                "$run_dir/$stage.$artifact.tmp",
                "$run_dir/$stage.$artifact",
            ),
            "return 0": ("return", "0"),
        }
        allowed_structure = {
            "rename_stage_artifact() {",
            "build_container_receipt() {",
            "finalize_container() {",
            "}",
        }
        structural_lines = tuple(
            line for line in lines if line in allowed_structure
        )
        if structural_lines != (
            "rename_stage_artifact() {",
            "}",
            "build_container_receipt() {",
            "}",
            "finalize_container() {",
            "}",
        ):
            return False
        for line in lines:
            words = _shell_words(line)
            if not isinstance(words, tuple):
                return False
            if _hard2g_line_mutates_run_dir(words):
                return False
            if line in allowed_structure:
                continue
            if line in builder_lines:
                continue
            if allowed_commands.get(line) != words:
                return False
        return True
    except Exception:
        return False


def _hard2g_observer_imports_are_python_tests_only(
    repo_root: Path,
    config: Mapping[str, object],
    dockerfile: object,
) -> bool:
    try:
        stages = _docker_stages_safe(dockerfile)
        if not isinstance(stages, Mapping):
            return False
        python_stage = stages.get(config["python_stage"])
        if type(python_stage) is not tuple or len(python_stage) != 2:
            return False
        python_instructions = python_stage[1]
        if (
            type(python_instructions) is not tuple
            or "COPY testframework/observers/ ./observers/"
            not in python_instructions
            or "COPY testframework/selftests/ ./selftests/"
            not in python_instructions
            or "ENV PYTHONPATH=/opt/hub/testframework/src:/opt/hub/testframework"
            not in python_instructions
            or "ENV PYTHONSAFEPATH=1" not in python_instructions
        ):
            return False
        if any(
            "testframework/observers/" in instruction
            for stage_name, stage in stages.items()
            if stage_name != config["python_stage"]
            for instruction in stage[1]
        ):
            return False
        source_paths = config["source_paths"]
        selftest_paths = config["selftest_paths"]
        if type(source_paths) is not tuple or type(selftest_paths) is not tuple:
            return False
        source_root = repo_root / "testframework/observers"
        if not source_root.is_dir() or source_root.is_symlink():
            return False
        if not _hard2g_observer_source_manifest_is_closed(repo_root, source_paths):
            return False
        observer_sources = tuple(
            _read_bounded_regular(repo_root / path) for path in source_paths
        )
        if not all(type(source) is str for source in observer_sources):
            return False
        if not _hard2g_observer_sources_have_closed_read_authority(
            source_paths,
            observer_sources,
        ):
            return False
        selftest_sources = {
            path: _read_bounded_regular(repo_root / path) for path in selftest_paths
        }
        if not all(type(source) is str for source in selftest_sources.values()):
            return False
        module = config["module"]
        if type(module) is not str or not all(
            _python_source_imports_module(source, module)
            for source in selftest_sources.values()
            if type(source) is str
        ):
            return False
        python_sources = _regular_files_below(
            repo_root / "testframework",
            ".py",
            (repo_root / "testframework" / "reports",),
        )
        if not isinstance(python_sources, tuple):
            return False
        allowed_selftests = {repo_root / path for path in selftest_paths}
        for source_path in python_sources:
            source = _read_bounded_regular(source_path)
            if type(source) is not str or not _python_source_imports_module(source, module):
                continue
            if source_path not in allowed_selftests and source_root not in source_path.parents:
                return False
        return True
    except Exception:
        return False


def _python_source_imports_module(text: str, module: str) -> bool:
    try:
        tree = ast.parse(text)
        return any(
            (
                isinstance(node, ast.ImportFrom)
                and type(node.module) is str
                and (node.module == module or node.module.startswith(f"{module}."))
            )
            or (
                isinstance(node, ast.Import)
                and any(
                    alias.name == module or alias.name.startswith(f"{module}.")
                    for alias in node.names
                )
            )
            for node in ast.walk(tree)
        )
    except Exception:
        return False


def _hard2g_task2_provenance_oracle_chain_is_closed(text: str) -> bool:
    try:
        tree = ast.parse(text)
        functions = {
            node.name: node
            for node in tree.body
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef))
        }

        def literals(function: ast.AST) -> set[str]:
            return {
                node.value
                for node in ast.walk(function)
                if isinstance(node, ast.Constant) and type(node.value) is str
            }

        def has_call(function: ast.AST, attribute: str) -> bool:
            return any(
                isinstance(node, ast.Call)
                and isinstance(node.func, ast.Attribute)
                and node.func.attr == attribute
                for node in ast.walk(function)
            )

        def has_literal_tuple(function: ast.AST, expected: tuple[str, ...]) -> bool:
            return any(
                isinstance(node, ast.Tuple)
                and tuple(
                    item.value if isinstance(item, ast.Constant) else None
                    for item in node.elts
                )
                == expected
                for node in ast.walk(function)
            )

        def has_snapshot_chain(function: ast.AST) -> bool:
            return any(
                isinstance(node, ast.Call)
                and isinstance(node.func, ast.Attribute)
                and node.func.attr == "verify_snapshot_chain"
                and tuple(
                    argument.id if isinstance(argument, ast.Name) else ""
                    for argument in node.args
                )
                == ("prebuild", "postgreen", "observer_snapshot")
                for node in ast.walk(function)
            )

        def is_runtime_published(value: ast.AST) -> bool:
            return (
                isinstance(value, ast.Attribute)
                and value.attr == "published"
                and isinstance(value.value, ast.Name)
                and value.value.id == "runtime"
            )

        def is_exact_runtime_construction(statement: ast.stmt) -> bool:
            if (
                not isinstance(statement, ast.Assign)
                or len(statement.targets) != 1
                or not isinstance(statement.targets[0], ast.Name)
                or statement.targets[0].id != "runtime"
                or not isinstance(statement.value, ast.Call)
            ):
                return False
            call = statement.value
            return (
                isinstance(call.func, ast.Name)
                and call.func.id == "_RawObserverRuntime"
                and not call.args
                and len(call.keywords) == 1
                and call.keywords[0].arg == "raw_observations"
                and isinstance(call.keywords[0].value, ast.Name)
                and call.keywords[0].value.id == "raw_observations"
            )

        def is_live_publication_unpack(statement: ast.stmt) -> bool:
            return (
                isinstance(statement, ast.Assign)
                and len(statement.targets) == 1
                and isinstance(statement.targets[0], ast.Tuple)
                and tuple(
                    element.id if isinstance(element, ast.Name) else ""
                    for element in statement.targets[0].elts
                )
                == ("final", "observer")
                and is_runtime_published(statement.value)
            )

        def is_runtime_published_guard(statement: ast.stmt) -> bool:
            return (
                isinstance(statement, ast.Assert)
                and isinstance(statement.test, ast.Compare)
                and len(statement.test.ops) == 1
                and isinstance(statement.test.ops[0], ast.IsNot)
                and len(statement.test.comparators) == 1
                and isinstance(statement.test.comparators[0], ast.Constant)
                and statement.test.comparators[0].value is None
                and is_runtime_published(statement.test.left)
            )

        def target_is_rooted_in(
            target: ast.expr,
            names: frozenset[str],
        ) -> bool:
            while isinstance(target, (ast.Attribute, ast.Subscript)):
                target = target.value
            return isinstance(target, ast.Name) and target.id in names

        def statement_writes_names_or_alias_roots(
            statement: ast.stmt,
            names: frozenset[str],
        ) -> bool:
            return any(
                (
                    isinstance(node, ast.Name)
                    and node.id in names
                    and isinstance(node.ctx, (ast.Store, ast.Del))
                )
                or (
                    isinstance(node, (ast.Attribute, ast.Subscript))
                    and isinstance(node.ctx, (ast.Store, ast.Del))
                    and target_is_rooted_in(node, names)
                )
                for node in ast.walk(statement)
            )

        def statement_rebinds_role_aliases(statement: ast.stmt) -> bool:
            names = frozenset({"final", "observer"})
            return statement_writes_names_or_alias_roots(
                statement,
                names,
            ) or any(
                isinstance(node, ast.Call)
                and isinstance(node.func, ast.Attribute)
                and target_is_rooted_in(node.func, names)
                for node in ast.walk(statement)
            )

        def statement_rebinds_runtime_or_published(statement: ast.stmt) -> bool:
            return statement_writes_names_or_alias_roots(
                statement,
                frozenset({"runtime"}),
            ) or any(
                (
                    isinstance(node, ast.Call)
                    and isinstance(node.func, ast.Name)
                    and node.func.id in {"setattr", "delattr"}
                    and len(node.args) >= 1
                    and isinstance(node.args[0], ast.Name)
                    and node.args[0].id == "runtime"
                )
                or (
                    isinstance(node, ast.Call)
                    and isinstance(node.func, ast.Attribute)
                    and target_is_rooted_in(
                        node.func,
                        frozenset({"runtime"}),
                    )
                )
                for node in ast.walk(statement)
            )

        def binding_path(value: ast.AST) -> tuple[str, tuple[str, ...]] | None:
            keys: list[str] = []
            while isinstance(value, ast.Subscript):
                if not isinstance(value.slice, ast.Constant) or type(value.slice.value) is not str:
                    return None
                keys.append(value.slice.value)
                value = value.value
            if not isinstance(value, ast.Name):
                return None
            return value.id, tuple(reversed(keys))

        def equality_chain(statement: ast.stmt) -> tuple[object, ...] | None:
            if not isinstance(statement, ast.Assert) or not isinstance(statement.test, ast.Compare):
                return None
            comparison = statement.test
            if not all(isinstance(operator, ast.Eq) for operator in comparison.ops):
                return None
            return tuple(
                binding_path(value)
                for value in (comparison.left, *comparison.comparators)
            )

        def has_role_specific_publication_bindings(function: ast.AST) -> bool:
            backend_chain = (
                ("final", ("backend_image_id",)),
                ("observer", ("container_image",)),
                ("observer", ("fixed_image_ids", "backend_unit")),
                ("_BACKEND_IMAGE_ID", ()),
            )
            python_binding = (
                ("observer", ("fixed_image_ids", "python_contract")),
                ("_PYTHON_IMAGE_ID", ()),
            )
            statements = tuple(function.body)
            runtime_constructions = tuple(
                index
                for index, statement in enumerate(statements)
                if is_exact_runtime_construction(statement)
            )
            live_unpacks = tuple(
                index
                for index, statement in enumerate(statements)
                if is_live_publication_unpack(statement)
            )
            guards = tuple(
                index
                for index, statement in enumerate(statements)
                if is_runtime_published_guard(statement)
            )
            backend_assertions = tuple(
                index
                for index, statement in enumerate(statements)
                if equality_chain(statement) == backend_chain
            )
            python_assertions = tuple(
                index
                for index, statement in enumerate(statements)
                if equality_chain(statement) == python_binding
            )
            if (
                len(runtime_constructions) != 1
                or len(live_unpacks) != 1
                or len(guards) != 1
                or len(backend_assertions) != 1
                or len(python_assertions) != 1
            ):
                return False
            construction_index = runtime_constructions[0]
            unpack_index = live_unpacks[0]
            backend_index = backend_assertions[0]
            python_index = python_assertions[0]
            if not (
                construction_index
                < guards[0]
                < unpack_index
                < backend_index
                < python_index
            ):
                return False
            if any(
                isinstance(node, ast.Return)
                for statement in statements[: python_index + 1]
                for node in ast.walk(statement)
            ):
                return False
            if any(
                index != unpack_index
                and statement_rebinds_role_aliases(statement)
                for index, statement in enumerate(
                    statements[: python_index + 1]
                )
            ):
                return False
            return not any(
                index != construction_index
                and statement_rebinds_runtime_or_published(statement)
                for index, statement in enumerate(
                    statements[: python_index + 1]
                )
            )

        fixed_images = functions.get(
            "test_docker_adapter_has_only_literal_read_argvs_and_fixed_images"
        )
        candidate_rejection = functions.get(
            "test_candidate_binding_rejects_every_foreign_identity_and_window_boundary"
        )
        bound_candidate = functions.get(
            "test_candidate_binding_requires_hostname_exact_prefix_and_open_window"
        )
        snapshots = functions.get(
            "test_lease_snapshots_bind_exact_manifest_images_and_no_acl_claim"
        )
        publication = functions.get(
            "test_publication_validates_before_write_then_hashes_seals_and_renames"
        )
        main_lifecycle = functions.get(
            "test_main_derives_every_sealed_classification_from_raw_observations"
        )
        payloads = functions.get("_publication_payloads")
        payload_dict_keys = tuple(
            {
                key.value
                for key in node.keys
                if isinstance(key, ast.Constant) and type(key.value) is str
            }
            for node in ast.walk(payloads)
            if isinstance(node, ast.Dict)
        )
        return (
            has_call(fixed_images, "inspect_fixed_images")
            and has_literal_tuple(
                fixed_images,
                (
                    "docker",
                    "image",
                    "inspect",
                    "--format",
                    "{{json .}}",
                    "worktree-master-python-contract",
                ),
            )
            and has_literal_tuple(
                fixed_images,
                (
                    "docker",
                    "image",
                    "inspect",
                    "--format",
                    "{{json .}}",
                    "worktree-master-backend-unit",
                ),
            )
            and has_call(candidate_rejection, "bind_unique_candidate")
            and has_call(candidate_rejection, "raises")
            and {"image_id", "container_image"} <= literals(candidate_rejection)
            and has_call(bound_candidate, "bind_unique_candidate")
            and "backend_image_id" in literals(bound_candidate)
            and has_snapshot_chain(snapshots)
            and has_call(publication, "publish")
            and any(
                isinstance(node, ast.Constant)
                and type(node.value) is bytes
                and node.value.startswith(b"observer_receipt_sha256=")
                for node in ast.walk(publication)
            )
            and has_role_specific_publication_bindings(main_lifecycle)
            and {"run_id", "owner_nonce", "container_id"} <= literals(main_lifecycle)
            and any(
                {
                    "run_id",
                    "owner_nonce",
                    "container_id",
                    "bound_backend_image_id",
                    "container_image",
                    "fixed_image_ids",
                }
                <= keys
                for keys in payload_dict_keys
            )
        )
    except Exception:
        return False


def _read_json_mapping(path: Path) -> Mapping[str, object] | object:
    text = _read_bounded_regular(path)
    document = _load_json_document_safe(text)
    return document if isinstance(document, Mapping) else _INVALID


def _is_sha256_digest(value: object) -> bool:
    return (
        type(value) is str
        and len(value) == 71
        and value.startswith("sha256:")
        and set(value[7:]) <= set("0123456789abcdef")
    )


def _canonical_manifest_paths_are_closed(
    manifest_bytes: str,
    policy_paths: object,
) -> bool:
    try:
        if type(policy_paths) is not tuple or not manifest_bytes.endswith("\n"):
            return False
        rows = tuple(manifest_bytes.removesuffix("\n").split("\n"))
        parsed_paths: list[str] = []
        for row in rows:
            path, separator, digest = row.partition("\t")
            if not separator or not path or not _is_sha256_digest(digest):
                return False
            parsed_paths.append(path)
        required_paths = tuple(path for path in policy_paths if "*" not in path)
        return len(parsed_paths) == len(set(parsed_paths)) and all(
            path in parsed_paths for path in required_paths
        )
    except Exception:
        return False


def _canonical_build_input_manifest_bytes(
    repo_root: Path,
    policy_paths: object,
) -> bytes | object:
    try:
        if type(policy_paths) is not tuple or any(
            type(path) is not str or not path or path.startswith("/")
            for path in policy_paths
        ):
            return _INVALID
        expanded: dict[str, Path] = {}
        for pattern in policy_paths:
            candidates = (
                tuple(
                    Path(path)
                    for path in glob.glob(str(repo_root / pattern), recursive=True)
                )
                if "*" in pattern
                else (repo_root / pattern,)
            )
            if not candidates:
                return _INVALID
            regular: list[tuple[str, Path]] = []
            for candidate in candidates:
                relative = _hard2g_contained_regular_relative(repo_root, candidate)
                if relative is _INVALID:
                    return _INVALID
                if type(relative) is str:
                    regular.append((relative, candidate))
            if "*" in pattern and not regular:
                return _INVALID
            if "*" not in pattern and len(regular) != len(candidates):
                return _INVALID
            for relative, candidate in regular:
                if relative in expanded:
                    return _INVALID
                expanded[relative] = candidate
        rows: list[bytes] = []
        for relative, path in sorted(expanded.items()):
            content = _read_bounded_regular(path)
            if type(content) is not str:
                return _INVALID
            rows.append(
                relative.encode("utf-8")
                + b"\tsha256:"
                + hashlib.sha256(content.encode("utf-8")).hexdigest().encode()
                + b"\n"
            )
        return b"".join(rows) if rows else _INVALID
    except Exception:
        return _INVALID


def _hard2g_contained_regular_relative(
    repo_root: Path,
    candidate: Path,
) -> str | None | object:
    try:
        root_stat = repo_root.lstat()
        if not _hard2g_lstat_component_is_closed(root_stat, directory=True):
            return _INVALID
        relative = candidate.relative_to(repo_root)
        if not relative.parts or any(
            part in {"", ".", ".."} for part in relative.parts
        ):
            return _INVALID
        current = repo_root
        for index, component in enumerate(relative.parts):
            current /= component
            current_stat = current.lstat()
            final_component = index == len(relative.parts) - 1
            if final_component:
                if _hard2g_lstat_component_is_closed(current_stat, directory=False):
                    return relative.as_posix()
                if _hard2g_lstat_component_is_closed(current_stat, directory=True):
                    return None
                return _INVALID
            if not _hard2g_lstat_component_is_closed(current_stat, directory=True):
                return _INVALID
        return _INVALID
    except Exception:
        return _INVALID


def _hard2g_lstat_component_is_closed(
    value: os.stat_result,
    *,
    directory: bool,
) -> bool:
    reparse_flag = getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", 0)
    attributes = getattr(value, "st_file_attributes", 0)
    return (
        (stat.S_ISDIR(value.st_mode) if directory else stat.S_ISREG(value.st_mode))
        and not attributes & reparse_flag
    )


def _hard2g_exact_manifest_binding_is_closed(
    repo_root: Path,
    manifest_bytes: object,
    manifest_sha256: object,
    policy_paths: object,
) -> bool:
    try:
        if (
            type(manifest_bytes) is not bytes
            or b"\r" in manifest_bytes
            or not manifest_bytes.endswith(b"\n")
            or not _is_sha256_digest(manifest_sha256)
        ):
            return False
        expected = _canonical_build_input_manifest_bytes(repo_root, policy_paths)
        return (
            type(expected) is bytes
            and manifest_bytes == expected
            and manifest_sha256
            == "sha256:" + hashlib.sha256(manifest_bytes).hexdigest()
        )
    except Exception:
        return False


def _hard2g_report_budget_is_closed(budget: object) -> bool:
    try:
        if budget != _HARD2G_REPORT_BUDGET or not isinstance(budget, Mapping):
            return False
        streams = _mapping_value(budget, "streams")
        allocated = sum(
            _mapping_value(streams, stage)[stream]
            for stage in ("build", "vet", "unit", "race")
            for stream in ("stdout", "stderr")
        )
        return allocated == budget["total_bytes"] - budget["reserved_metadata_bytes"]
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
        or relative in _HARD2G_METADATA_README_PATHS
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
        "hard2g_observer": {
            "module": _HARD2G_OBSERVER_MODULE,
            "python_stage": "python-tests",
            "source_paths": list(_HARD2G_OBSERVER_SOURCE_PATHS),
            "selftest_paths": list(_HARD2G_OBSERVER_SELFTEST_PATHS),
            "policy_path": _HARD2G_OBSERVER_POLICY_PATH,
            "image_inspect_argvs": [
                list(argv) for argv in _HARD2G_IMAGE_INSPECT_ARGVS
            ],
            "canonical_build_input_paths": list(_HARD2G_BUILD_INPUT_PATHS),
            "snapshot_assertion_source": _HARD2G_OBSERVER_SNAPSHOT_ASSERTION_SOURCE,
            "fixture_path": _HARD2G_OBSERVER_FIXTURE_PATH,
        },
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
