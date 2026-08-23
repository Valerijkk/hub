from __future__ import annotations

import ast
import importlib.metadata
import importlib.resources
import re
import tomllib
from pathlib import Path

import hub_testkit


PROJECT_ROOT = Path(__file__).resolve().parents[3]
PYPROJECT_PATH = PROJECT_ROOT / "pyproject.toml"
PACKAGE_ROOT = PROJECT_ROOT / "src" / "hub_testkit"


def _project() -> dict[str, object]:
    with PYPROJECT_PATH.open("rb") as stream:
        return tomllib.load(stream)


def _dependency_name(requirement: str) -> str:
    return re.split(r"[<>=!~ ;\[]", requirement, maxsplit=1)[0].lower().replace("_", "-")


def test_build_and_python_contract_is_closed() -> None:
    document = _project()
    build_system = document["build-system"]
    project = document["project"]
    hatch = document["tool"]["hatch"]["build"]["targets"]["wheel"]

    assert build_system["build-backend"] == "hatchling.build"
    assert {_dependency_name(item) for item in build_system["requires"]} == {"hatchling"}
    assert project["name"] == "hub-testkit"
    assert project["requires-python"] == ">=3.12,<3.13"
    assert hatch["packages"] == ["src/hub_testkit"]
    assert (PROJECT_ROOT / ".python-version").read_text(encoding="ascii") == "3.12.13\n"


def test_dependency_groups_have_one_authoritative_membership() -> None:
    document = _project()
    project_dependencies = {
        _dependency_name(item) for item in document["project"]["dependencies"]
    }
    groups = {
        name: {_dependency_name(item) for item in items}
        for name, items in document["dependency-groups"].items()
    }

    assert project_dependencies == {"jsonschema", "pydantic"}
    assert groups == {
        "browser": {"playwright", "pytest-playwright"},
        "contract": {"openapi-spec-validator"},
        "framework": {"pytest", "pytest-randomly", "pytest-timeout", "pyyaml"},
        "http": {"httpx"},
        "load": {"locust"},
        "quality": {"import-linter", "mypy", "ruff"},
    }


def test_pytest_configuration_is_strict_without_path_injection() -> None:
    pytest_config = _project()["tool"]["pytest"]["ini_options"]

    assert pytest_config["testpaths"] == ["selftests", "tests"]
    assert pytest_config["xfail_strict"] is True
    assert "--strict-config" in pytest_config["addopts"]
    assert "--strict-markers" in pytest_config["addopts"]
    assert "pythonpath" not in pytest_config


def test_installed_distribution_matches_public_package() -> None:
    assert importlib.metadata.version("hub-testkit") == hub_testkit.__version__
    assert importlib.resources.files("hub_testkit").joinpath("py.typed").is_file()
    assert hub_testkit.__all__ == ()


def test_public_package_boundary_has_no_suite_or_path_hack() -> None:
    tree = ast.parse((PACKAGE_ROOT / "__init__.py").read_text(encoding="utf-8"))
    imported_roots: set[str] = set()
    for node in ast.walk(tree):
        if isinstance(node, ast.Import):
            imported_roots.update(alias.name.split(".", 1)[0] for alias in node.names)
        elif isinstance(node, ast.ImportFrom) and node.module:
            imported_roots.add(node.module.split(".", 1)[0])
        elif isinstance(node, ast.Attribute) and isinstance(node.value, ast.Name):
            assert not (node.value.id == "sys" and node.attr == "path")

    assert imported_roots.isdisjoint({"cases", "selftests", "tests"})
