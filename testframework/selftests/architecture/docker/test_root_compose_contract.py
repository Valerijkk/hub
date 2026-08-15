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
from yaml.nodes import MappingNode
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
        "backend-unit": (
            "/opt/hub/testframework/docker/entrypoints/backend-unit.sh",
            "--hub-qa-target=backend-unit",
        ),
        "frontend-unit": (
            "/opt/hub/testframework/docker/entrypoints/frontend-unit.sh",
            "--hub-qa-target=frontend-unit",
        ),
        "python-contract": (
            "/opt/hub/testframework/docker/entrypoints/python-qa.sh",
            "--hub-qa-target=python-contract",
        ),
        "python-api": (
            "/opt/hub/testframework/docker/entrypoints/python-qa.sh",
            "--hub-qa-target=python-api",
        ),
        "python-integration": (
            "/opt/hub/testframework/docker/entrypoints/python-qa.sh",
            "--hub-qa-target=python-integration",
        ),
        "python-e2e-smoke": (
            "/opt/hub/testframework/docker/entrypoints/python-qa.sh",
            "--hub-qa-target=python-e2e-smoke",
        ),
        "python-e2e-full": (
            "/opt/hub/testframework/docker/entrypoints/python-qa.sh",
            "--hub-qa-target=python-e2e-full",
        ),
        "python-load": (
            "/opt/hub/testframework/docker/entrypoints/python-qa.sh",
            "--hub-qa-target=python-load",
        ),
    }
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


def test_qa_services_have_fixed_commands_and_security_options() -> None:
    services = _compose_services(load_root_compose_contract(_REPO_ROOT))

    for name in _PUBLIC_SERVICES:
        service = _mapping_value(services, name)
        assert service["profiles"] == ("qa",)
        assert service["command"] == _FIXED_COMMANDS[name]
        assert service["read_only"] is True
        assert service["security_opt"] == ("no-new-privileges:true",)
        assert service["cap_drop"] == ("ALL",)
        assert service["init"] is True
        assert _resource_limits_are_bounded(service)
        assert _has_reports_mount(service)
        assert "ports" not in service
        assert not _contains_token(service, "docker.sock")


def test_qa_infrastructure_is_unpublished_and_disposable() -> None:
    services = _compose_services(load_root_compose_contract(_REPO_ROOT))

    for name in _QA_INFRASTRUCTURE:
        service = _mapping_value(services, name)
        assert service["profiles"] == ("qa",)
        assert "ports" not in service
        assert _mapping_value(service, "healthcheck")
        assert _resource_limits_are_bounded(service)
        assert _storage_is_disposable(service)
        assert not _contains_token(service, "docker.sock")


def test_docker_directory_contains_only_manifested_assets() -> None:
    contract = load_root_compose_contract(_REPO_ROOT)
    policy = _mapping_value(contract, "policy")

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
    result: dict[str, object] = {}
    for key_node, value_node in node.value:
        key = loader.construct_object(key_node, deep=deep)
        if type(key) is not str or key in result:
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
