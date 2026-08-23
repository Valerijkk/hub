from __future__ import annotations

from pathlib import Path

import pytest

from hub_testkit.pytest_plugin.collection import discover_case_paths
from hub_testkit.pytest_plugin.markers import CollectionError


pytest_plugins = ("pytester",)


PROJECT_ROOT = Path(__file__).resolve().parents[3]
SCHEMA_PATH = PROJECT_ROOT / "schemas" / "test-case.schema.json"
MARKER_POLICY_PATH = PROJECT_ROOT / "policies" / "markers.toml"
REDACTION_POLICY_PATH = PROJECT_ROOT / "policies" / "redaction.toml"
NODE_ID = "testframework/tests/api/health/test_live.py::test_live"


def _pytest_automation(node_id: str = NODE_ID) -> str:
    return "\n".join(
        (
            "[automation]",
            'kind = "pytest"',
            'layer = "api"',
            f'node_id = "{node_id}"',
        )
    )


def _manual_automation() -> str:
    return "\n".join(
        (
            "[automation]",
            'kind = "manual"',
            'evidence_owner = "qa-owner"',
        )
    )


def _case_text(
    case_id: str,
    *,
    status: str = "active",
    automation: str | None = None,
) -> str:
    return "\n".join(
        (
            "+++",
            'schema = "hub.testframework.case/v1"',
            f'id = "{case_id}"',
            'title = "Plugin case"',
            f'status = "{status}"',
            'task = "TASK-206"',
            'requirement = "TF-PLUGIN"',
            'risk = "RISK-HIGH"',
            'scenario = "happy"',
            'subject = "plugin collection"',
            'preconditions = ["isolated fixture"]',
            'oracle = "item is linked"',
            'cleanup = "temporary file removed"',
            automation or _pytest_automation(),
            "+++",
            "# Plugin case",
            "",
            "This case is collected by the public plugin.",
        )
    )


def _case_root(pytester: pytest.Pytester) -> Path:
    root = pytester.path / "testframework" / "cases"
    root.mkdir(parents=True)
    return root


def _write_case(root: Path, case_id: str, contents: str) -> Path:
    path = root / f"{case_id}.md"
    path.write_text(contents, encoding="utf-8", newline="\n")
    return path


def _write_item(pytester: pytest.Pytester, source: str) -> Path:
    path = pytester.path / "testframework" / "tests" / "api" / "health" / "test_live.py"
    path.parent.mkdir(parents=True)
    path.write_text(source, encoding="utf-8", newline="\n")
    return path


def _plugin_args(case_root: Path) -> tuple[str, ...]:
    return (
        "-p",
        "no:playwright",
        "-p",
        "hub_testkit.pytest_plugin.plugin",
        f"--hub-case-root={case_root}",
        f"--hub-case-schema={SCHEMA_PATH}",
        f"--hub-marker-policy={MARKER_POLICY_PATH}",
        f"--hub-redaction-policy={REDACTION_POLICY_PATH}",
    )


def _assert_usage(result: pytest.RunResult) -> None:
    assert result.ret == pytest.ExitCode.USAGE_ERROR


def test_plugin_requires_all_explicit_configuration_options(pytester: pytest.Pytester) -> None:
    _assert_usage(pytester.runpytest_inprocess("-p", "hub_testkit.pytest_plugin.plugin"))


def test_plugin_registers_only_the_traceability_markers(pytester: pytest.Pytester) -> None:
    case_root = _case_root(pytester)
    _write_case(case_root, "TC-PLUGIN-0001", _case_text("TC-PLUGIN-0001"))

    result = pytester.runpytest_inprocess(*_plugin_args(case_root), "--markers")

    assert result.ret == pytest.ExitCode.OK
    marker_output = result.stdout.str()
    for marker in ("case_id(id)", "contract", "api", "integration", "e2e"):
        assert marker in marker_output


def test_plugin_collects_one_valid_item_and_sets_only_redacted_properties(
    pytester: pytest.Pytester,
) -> None:
    case_root = _case_root(pytester)
    _write_case(case_root, "TC-PLUGIN-0001", _case_text("TC-PLUGIN-0001"))
    pytester.makeconftest(
        """import pytest


@pytest.hookimpl(tryfirst=True)
def pytest_collection_modifyitems(items):
    items[0].user_properties.append(("unrelated", "preserved"))
"""
    )
    item = _write_item(
        pytester,
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
@pytest.mark.api
def test_live(request):
    assert request.node.user_properties == [
        ("unrelated", "preserved"),
        ("hub.case_id", "TC-PLUGIN-0001"),
        ("hub.layer", "api"),
        ("hub.task", "TASK-206"),
        ("hub.requirement", "TF-PLUGIN"),
        ("hub.risk", "RISK-HIGH"),
    ]
""",
    )

    result = pytester.runpytest_inprocess(*_plugin_args(case_root), str(item))

    result.assert_outcomes(passed=1)


@pytest.mark.parametrize(
    "source",
    (
        "def test_live():\n    pass\n",
        """import pytest

pytestmark = [pytest.mark.case_id("TC-PLUGIN-0001"), pytest.mark.api]


@pytest.mark.case_id("TC-PLUGIN-0001")
def test_live():
    pass
""",
    ),
)
def test_plugin_rejects_missing_or_multiple_case_markers(
    pytester: pytest.Pytester,
    source: str,
) -> None:
    case_root = _case_root(pytester)
    _write_case(case_root, "TC-PLUGIN-0001", _case_text("TC-PLUGIN-0001"))
    item = _write_item(pytester, source)

    _assert_usage(pytester.runpytest_inprocess(*_plugin_args(case_root), str(item)))


@pytest.mark.parametrize(
    "marker",
    (
        '@pytest.mark.case_id(id="TC-PLUGIN-0001")',
        '@pytest.mark.case_id("TC-PLUGIN-000")',
    ),
)
def test_plugin_rejects_noncanonical_or_keyword_case_marker(
    pytester: pytest.Pytester,
    marker: str,
) -> None:
    case_root = _case_root(pytester)
    _write_case(case_root, "TC-PLUGIN-0001", _case_text("TC-PLUGIN-0001"))
    item = _write_item(
        pytester,
        f"""import pytest


{marker}
@pytest.mark.api
def test_live():
    pass
""",
    )

    _assert_usage(pytester.runpytest_inprocess(*_plugin_args(case_root), str(item)))


@pytest.mark.parametrize(
    "source",
    (
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
def test_live():
    pass
""",
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
@pytest.mark.api("argument")
def test_live():
    pass
""",
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
@pytest.mark.api
@pytest.mark.contract
def test_live():
    pass
""",
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
@pytest.mark.contract
def test_live():
    pass
""",
    ),
)
def test_plugin_rejects_missing_multiple_or_wrong_layer_markers(
    pytester: pytest.Pytester,
    source: str,
) -> None:
    case_root = _case_root(pytester)
    _write_case(case_root, "TC-PLUGIN-0001", _case_text("TC-PLUGIN-0001"))
    item = _write_item(pytester, source)

    _assert_usage(pytester.runpytest_inprocess(*_plugin_args(case_root), str(item)))


def test_plugin_rejects_unknown_case_marker(pytester: pytest.Pytester) -> None:
    case_root = _case_root(pytester)
    _write_case(case_root, "TC-PLUGIN-0001", _case_text("TC-PLUGIN-0001"))
    item = _write_item(
        pytester,
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0003")
@pytest.mark.api
def test_live():
    pass
""",
    )

    _assert_usage(pytester.runpytest_inprocess(*_plugin_args(case_root), str(item)))


def test_plugin_rejects_orphan_node_for_an_active_case(pytester: pytest.Pytester) -> None:
    case_root = _case_root(pytester)
    _write_case(case_root, "TC-PLUGIN-0001", _case_text("TC-PLUGIN-0001"))
    item = pytester.path / "testframework" / "tests" / "api" / "health" / "test_other.py"
    item.parent.mkdir(parents=True)
    item.write_text(
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
@pytest.mark.api
def test_other():
    pass
""",
        encoding="utf-8",
        newline="\n",
    )

    _assert_usage(pytester.runpytest_inprocess(*_plugin_args(case_root), str(item)))


def test_plugin_rejects_an_uncollected_active_case(pytester: pytest.Pytester) -> None:
    case_root = _case_root(pytester)
    _write_case(case_root, "TC-PLUGIN-0001", _case_text("TC-PLUGIN-0001"))
    _write_case(
        case_root,
        "TC-PLUGIN-0002",
        _case_text(
            "TC-PLUGIN-0002",
            automation=_pytest_automation(
                "testframework/tests/api/health/test_missing.py::test_missing"
            ),
        ),
    )
    item = _write_item(
        pytester,
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
@pytest.mark.api
def test_live():
    pass
""",
    )

    _assert_usage(pytester.runpytest_inprocess(*_plugin_args(case_root), str(item)))


@pytest.mark.parametrize("decorator", ("skip", "xfail"))
def test_plugin_rejects_skipped_or_expected_failure_items(
    pytester: pytest.Pytester,
    decorator: str,
) -> None:
    case_root = _case_root(pytester)
    _write_case(case_root, "TC-PLUGIN-0001", _case_text("TC-PLUGIN-0001"))
    item = _write_item(
        pytester,
        f'''import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
@pytest.mark.api
@pytest.mark.{decorator}
def test_live():
    pass
''',
    )

    _assert_usage(pytester.runpytest_inprocess(*_plugin_args(case_root), str(item)))




def test_plugin_rejects_spoofed_hub_property(pytester: pytest.Pytester) -> None:
    case_root = _case_root(pytester)
    _write_case(case_root, "TC-PLUGIN-0001", _case_text("TC-PLUGIN-0001"))
    pytester.makeconftest(
        """import pytest


@pytest.hookimpl(tryfirst=True)
def pytest_collection_modifyitems(items):
    items[0].user_properties.append(("hub.case_id", "spoofed"))
"""
    )
    item = _write_item(
        pytester,
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
@pytest.mark.api
def test_live():
    pass
""",
    )

    _assert_usage(pytester.runpytest_inprocess(*_plugin_args(case_root), str(item)))


@pytest.mark.parametrize("status", ("planned", "retired"))
def test_plugin_rejects_planned_or_retired_case_collection(
    pytester: pytest.Pytester,
    status: str,
) -> None:
    case_root = _case_root(pytester)
    _write_case(
        case_root,
        "TC-PLUGIN-0001",
        _case_text(
            "TC-PLUGIN-0001",
            status=status,
            automation=_manual_automation(),
        ),
    )
    item = _write_item(
        pytester,
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
@pytest.mark.api
def test_live():
    pass
""",
    )

    _assert_usage(pytester.runpytest_inprocess(*_plugin_args(case_root), str(item)))


def test_plugin_hides_case_canary_when_loading_fails(pytester: pytest.Pytester) -> None:
    case_root = _case_root(pytester)
    canary = "PLUGIN-SECRET-CANARY"
    _write_case(
        case_root,
        "TC-PLUGIN-0001",
        _case_text("TC-PLUGIN-0001").replace('status = "active"', f'status = "{canary}"'),
    )
    item = _write_item(
        pytester,
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
@pytest.mark.api
def test_live():
    pass
""",
    )

    result = pytester.runpytest_inprocess(*_plugin_args(case_root), str(item))

    _assert_usage(result)
    assert canary not in result.stdout.str()
    assert canary not in result.stderr.str()


@pytest.mark.parametrize(
    "mutation",
    (
        'items[0].add_marker(pytest.mark.skip(reason="late mutation"))',
        'items[0].user_properties.append(("hub.case_id", "late mutation"))',
    ),
)
def test_plugin_rejects_mutations_after_modifyitems_hooks(
    pytester: pytest.Pytester,
    mutation: str,
) -> None:
    case_root = _case_root(pytester)
    _write_case(case_root, "TC-PLUGIN-0001", _case_text("TC-PLUGIN-0001"))
    pytester.makeconftest(
        f"""import pytest


@pytest.hookimpl(hookwrapper=True, trylast=True)
def pytest_collection_modifyitems(items):
    yield
    {mutation}
"""
    )
    item = _write_item(
        pytester,
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
@pytest.mark.api
def test_live():
    pass
""",
    )

    _assert_usage(pytester.runpytest_inprocess(*_plugin_args(case_root), str(item)))


def test_discovery_ignores_readme_documentation(tmp_path: Path) -> None:
    (tmp_path / "README.md").write_text("# Cases\n", encoding="utf-8")
    case_path = tmp_path / "TC-PLUGIN-0001.md"
    case_path.write_text("# case\n", encoding="utf-8")

    assert discover_case_paths(tmp_path) == (case_path,)


def test_discovery_rejects_more_than_512_case_sources(tmp_path: Path) -> None:
    for number in range(513):
        (tmp_path / f"case-{number:04d}.md").write_text("# case\n", encoding="utf-8")

    with pytest.raises(CollectionError):
        discover_case_paths(tmp_path)


def test_discovery_rejects_17_nested_directories(tmp_path: Path) -> None:
    nested = tmp_path
    for number in range(17):
        nested = nested / f"level-{number}"
        nested.mkdir()

    with pytest.raises(CollectionError):
        discover_case_paths(tmp_path)


def test_plugin_rejects_zero_collection_and_symlink_case_discovery(
    pytester: pytest.Pytester,
) -> None:
    case_root = _case_root(pytester)
    _write_case(case_root, "TC-PLUGIN-0001", _case_text("TC-PLUGIN-0001"))

    _assert_usage(pytester.runpytest_inprocess(*_plugin_args(case_root)))

    linked_root = pytester.path / "linked-cases"
    try:
        linked_root.symlink_to(case_root, target_is_directory=True)
    except OSError:
        pytest.skip("symlinks are unavailable on this platform")
    item = _write_item(
        pytester,
        """import pytest


@pytest.mark.case_id("TC-PLUGIN-0001")
@pytest.mark.api
def test_live():
    pass
""",
    )

    _assert_usage(pytester.runpytest_inprocess(*_plugin_args(linked_root), str(item)))
