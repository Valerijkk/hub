from __future__ import annotations

import ast
import copy
import tomllib
import traceback
from pathlib import Path

import pytest

from hub_testkit.http import HttpExchange, HttpRequest
from hub_testkit.harness.project_outbox import (
    ProjectOutboxAssertionError,
    ProjectOutboxRuntime,
    ProjectOutboxRuntimeError,
    assert_project_create_returns_consistent_published_final_state,
)
from hub_testkit.harness.pytest_project_outbox import (
    cleanup_project_outbox_runtime,
    create_project_outbox_runtime_from_config,
    provide_project_outbox_harness_factory,
)
from hub_testkit.integration.project_outbox import (
    OwnedProjectOutboxState,
    ProjectCreatedReceipt,
    ProjectOutboxModelError,
    ProjectOutboxSeed,
    ProjectProjection,
    ProvisionedProjectOutbox,
)


PROJECT_ROOT = Path(__file__).resolve().parents[3]
DOMAIN_ROOT = PROJECT_ROOT / "src" / "hub_testkit" / "integration" / "project_outbox"
HARNESS_ROOT = PROJECT_ROOT / "src" / "hub_testkit" / "harness"
SUITE_ROOT = (
    PROJECT_ROOT / "tests" / "integration" / "redis-nats" / "project_outbox"
)
DOMAIN_PATHS = tuple(
    DOMAIN_ROOT / name
    for name in ("__init__.py", "api.py", "builders.py", "models.py", "oracles.py")
)

ASSERTIONS = (
    "assert_project_create_returns_consistent_published_final_state",
    "assert_duplicate_project_key_conflict_leaves_one_consistent_final_state",
    "assert_anonymous_project_create_is_rejected_with_zero_owned_final_counts",
    "assert_malformed_project_json_is_rejected_with_zero_owned_final_counts",
    "assert_durable_project_created_observer_replays_after_observer_reconnect",
)
TEST_FUNCTIONS = (
    "test_project_create_returns_consistent_published_final_state",
    "test_duplicate_project_key_conflict_leaves_one_consistent_final_state",
    "test_anonymous_project_create_is_rejected_with_zero_owned_final_counts",
    "test_malformed_project_json_is_rejected_with_zero_owned_final_counts",
    "test_durable_project_created_observer_replays_after_observer_reconnect",
)
TEST_TO_ASSERTION = {
    "test_project_create_returns_consistent_published_final_state": (
        "assert_project_create_returns_consistent_published_final_state"
    ),
    "test_duplicate_project_key_conflict_leaves_one_consistent_final_state": (
        "assert_duplicate_project_key_conflict_leaves_one_consistent_final_state"
    ),
    "test_anonymous_project_create_is_rejected_with_zero_owned_final_counts": (
        "assert_anonymous_project_create_is_rejected_with_zero_owned_final_counts"
    ),
    "test_malformed_project_json_is_rejected_with_zero_owned_final_counts": (
        "assert_malformed_project_json_is_rejected_with_zero_owned_final_counts"
    ),
    "test_durable_project_created_observer_replays_after_observer_reconnect": (
        "assert_durable_project_created_observer_replays_after_observer_reconnect"
    ),
}
TEST_TO_CASE_ID = {
    "test_project_create_returns_consistent_published_final_state": (
        "TC-OUTBOX-0001"
    ),
    "test_duplicate_project_key_conflict_leaves_one_consistent_final_state": (
        "TC-OUTBOX-0002"
    ),
    "test_anonymous_project_create_is_rejected_with_zero_owned_final_counts": (
        "TC-OUTBOX-0003"
    ),
    "test_malformed_project_json_is_rejected_with_zero_owned_final_counts": (
        "TC-OUTBOX-0004"
    ),
    "test_durable_project_created_observer_replays_after_observer_reconnect": (
        "TC-OUTBOX-0005"
    ),
}
CASE_PATHS = tuple(
    PROJECT_ROOT / "cases" / "outbox" / f"TC-OUTBOX-000{index}.md"
    for index in range(1, 6)
)
RUNTIME_PROOF_OVERCLAIMS = (
    "puback",
    "one transaction",
    "single transaction",
    "same transaction",
    "atomic transaction",
    "transactional",
    "commits and publishes",
    "commit and publish atomically",
    "before the row is marked",
    "before published",
)
FORBIDDEN_IMPORT_ROOTS = frozenset(
    {
        "aiohttp",
        "asyncpg",
        "docker",
        "httpx",
        "motor",
        "nats",
        "os",
        "pathlib",
        "psycopg",
        "psycopg2",
        "pymongo",
        "redis",
        "requests",
        "socket",
        "subprocess",
        "time",
        "urllib",
    }
)
GLOBAL_FORBIDDEN_CALLS = frozenset(
    {
        "AsyncClient",
        "Client",
        "HttpxTransport",
        "MongoClient",
        "Popen",
        "Redis",
        "connect",
        "create_connection",
        "create_subprocess_exec",
        "create_subprocess_shell",
        "getenv",
        "open",
        "sleep",
        "system",
    }
)
SUITE_DIRECT_CALLS = frozenset(
    {
        "cleanup",
        "cleanup_project_outbox_runtime",
        "create",
        "disconnect_observer",
        "observe_project_created",
        "provision",
        "reconnect_observer",
        "send",
        "state",
    }
)
CONSTRUCTION_CANARY = "project-outbox-opaque-canary-secret"


def test_project_outbox_opaque_objects_reject_construction_copy_and_mutation() -> None:
    expectations = (
        (ProjectOutboxSeed, ProjectOutboxModelError, "project outbox model is invalid"),
        (
            ProvisionedProjectOutbox,
            ProjectOutboxModelError,
            "project outbox model is invalid",
        ),
        (
            ProjectCreatedReceipt,
            ProjectOutboxModelError,
            "project outbox model is invalid",
        ),
        (
            OwnedProjectOutboxState,
            ProjectOutboxModelError,
            "project outbox model is invalid",
        ),
        (ProjectProjection, ProjectOutboxModelError, "project outbox model is invalid"),
        (
            ProjectOutboxRuntime,
            ProjectOutboxRuntimeError,
            "project outbox runtime is invalid",
        ),
    )

    for constructor, expected_error, expected_message in expectations:
        _assert_generic_failure(constructor, expected_error, expected_message)
        instance = _opaque_instance_with_canary(constructor)
        _assert_generic_failure(instance.__init__, expected_error, expected_message)
        _assert_generic_failure(
            lambda: copy.copy(instance),
            expected_error,
            expected_message,
        )
        _assert_generic_failure(
            lambda: copy.deepcopy(instance),
            expected_error,
            expected_message,
        )
        assert CONSTRUCTION_CANARY not in repr(instance)
        with pytest.raises((AttributeError, TypeError)):
            setattr(instance, "injected", CONSTRUCTION_CANARY)


def test_project_outbox_factory_boundary_is_single_install_and_canary_safe() -> None:
    config = _Config()
    factory = _ExplodingFactory()
    provide_project_outbox_harness_factory(config, factory)

    with pytest.raises(pytest.UsageError) as duplicate_info:
        provide_project_outbox_harness_factory(config, factory)
    duplicate_error = duplicate_info.value
    assert str(duplicate_error) == "project outbox harness is invalid"
    assert duplicate_error.__cause__ is None
    assert duplicate_error.__suppress_context__ is True

    try:
        create_project_outbox_runtime_from_config(config)
    except pytest.UsageError as error:
        rendered = "".join(
            traceback.TracebackException.from_exception(
                error,
                capture_locals=True,
            ).format()
        )
        assert str(error) == "project outbox harness is unavailable"
        assert error.__cause__ is None
        assert error.__suppress_context__ is True
        assert CONSTRUCTION_CANARY not in rendered
    else:
        raise AssertionError("hostile project outbox factory was admitted")


def test_project_outbox_delegate_and_cleanup_failures_are_generic_and_canary_safe() -> None:
    config = _Config()
    factory = _HostileFactory()
    provide_project_outbox_harness_factory(config, factory)
    runtime = create_project_outbox_runtime_from_config(config)

    try:
        assert_project_create_returns_consistent_published_final_state(runtime)
    except ProjectOutboxAssertionError as assertion_error:
        rendered = "".join(
            traceback.TracebackException.from_exception(
                assertion_error,
                capture_locals=True,
            ).format()
        )
        assert str(assertion_error) == "project outbox contract was not satisfied"
        assert assertion_error.__cause__ is None
        assert assertion_error.__suppress_context__ is True
        assert CONSTRUCTION_CANARY not in rendered
    else:
        raise AssertionError("hostile project outbox delegate was admitted")

    try:
        cleanup_project_outbox_runtime(runtime)
    except pytest.UsageError as cleanup_error:
        rendered = "".join(
            traceback.TracebackException.from_exception(
                cleanup_error,
                capture_locals=True,
            ).format()
        )
        assert str(cleanup_error) == "project outbox harness cleanup failed"
        assert cleanup_error.__cause__ is None
        assert cleanup_error.__suppress_context__ is True
        assert CONSTRUCTION_CANARY not in rendered
    else:
        raise AssertionError("hostile project outbox cleanup was admitted")

    with pytest.raises(pytest.UsageError) as retry_info:
        cleanup_project_outbox_runtime(runtime)
    retry_error = retry_info.value
    retry_rendered = "".join(
        traceback.TracebackException.from_exception(
            retry_error,
            capture_locals=True,
        ).format()
    )
    assert str(retry_error) == "project outbox harness cleanup failed"
    assert retry_error.__cause__ is None
    assert retry_error.__suppress_context__ is True
    assert CONSTRUCTION_CANARY not in retry_rendered
    assert factory.cleanup_calls == 1


def test_project_outbox_cases_claim_only_observable_consistent_final_state() -> None:
    unsafe_claims = (
        "PubAck is observed before the row is marked published",
        "Project, membership, and outbox are committed in one transaction",
        "The transactional create commits and publishes atomically",
    )
    safe_claim = (
        "One durable typed receipt and closed final counts agree without "
        "inferring write atomicity or acknowledgement order"
    )

    assert all(not _runtime_claim_is_bounded(claim) for claim in unsafe_claims)
    assert _runtime_claim_is_bounded(safe_claim)
    assert _runtime_claim_is_bounded(" ".join(ASSERTIONS + TEST_FUNCTIONS))
    assert tuple(path.name for path in CASE_PATHS) == tuple(
        f"TC-OUTBOX-000{index}.md" for index in range(1, 6)
    )
    for path in CASE_PATHS:
        assert _runtime_claim_is_bounded(_case_claim_text(path))


def test_project_outbox_boundary_detects_alias_dynamic_private_and_io_bypasses() -> None:
    tree = ast.parse(
        """import httpx as web
import importlib as loader
import nats as broker
import os as system
import psycopg as database
import pymongo as documents
import redis as cache
import subprocess as process
from builtins import __import__ as import_any
from builtins import getattr as lookup
from builtins import open as writer
from hub_testkit.http import HttpxTransport as DefaultTransport
from pathlib import Path as P
from time import sleep as pause

alias = project_outbox_runtime
typed_alias: object = project_outbox_runtime
if False:
    alias.provision(seed)
    typed_alias.state()
    project_outbox_runtime.disconnect_observer()
    project_outbox_runtime.reconnect_observer()
    project_outbox_runtime.observe_project_created()
    delegate.send(request)
    delegate.cleanup()
    cleanup_project_outbox_runtime(project_outbox_runtime)
    project_outbox_runtime._delegate
    lookup(project_outbox_runtime, "_runtime_binding")
    object.__getattribute__(project_outbox_runtime, "_delegate")
    type(project_outbox_runtime).__dataclass_fields__["_delegate"].__get__(
        project_outbox_runtime,
        type(project_outbox_runtime),
    )
    broker.connect()
    database.connect(dsn)
    documents.MongoClient(uri)
    cache.Redis()
    process.Popen(command)
    web.Client()
    DefaultTransport(client, policy, redactor)
    pause(1)
    system.getenv("SECRET")
    P(".").read_text()
    writer("report.txt", "w")
    loader.import_module(module_name)
    import_any(module_name)
    while True:
        alias.state()
"""
    )

    assert _violations(tree, suite=True) == {
        "direct-runtime-access",
        "dynamic-import",
        "forbidden-import",
        "forbidden-io",
        "private-access",
        "suite-import-not-allowed",
    }


def test_project_outbox_boundary_allows_only_five_semantic_suite_calls() -> None:
    source = "import pytest\nfrom hub_testkit.harness.project_outbox import (\n"
    source += "".join(f"    {name},\n" for name in ASSERTIONS)
    source += ")\n"
    for function_name, assertion_name in TEST_TO_ASSERTION.items():
        source += (
            f'@pytest.mark.case_id("{TEST_TO_CASE_ID[function_name]}")\n'
            "@pytest.mark.integration\n"
            f"def {function_name}(project_outbox_runtime):\n"
            f"    {assertion_name}(project_outbox_runtime)\n"
        )

    assert _violations(ast.parse(source), suite=True) == set()


def test_project_outbox_boundary_rejects_swapped_unknown_markers_and_body_escape() -> None:
    swapped = ast.parse(
        """@pytest.mark.case_id("TC-OUTBOX-0002")
@pytest.mark.integration
def test_project_create_returns_consistent_published_final_state(project_outbox_runtime):
    assert_project_create_returns_consistent_published_final_state(project_outbox_runtime)
"""
    ).body[0]
    unknown = ast.parse(
        """@pytest.mark.case_id("TC-OUTBOX-9999")
@pytest.mark.integration
def test_project_create_returns_consistent_published_final_state(project_outbox_runtime):
    assert_project_create_returns_consistent_published_final_state(project_outbox_runtime)
"""
    ).body[0]
    wrong_layer = ast.parse(
        """@pytest.mark.case_id("TC-OUTBOX-0001")
@pytest.mark.api
def test_project_create_returns_consistent_published_final_state(project_outbox_runtime):
    assert_project_create_returns_consistent_published_final_state(project_outbox_runtime)
"""
    ).body[0]
    escaped = ast.parse(
        """@pytest.mark.case_id("TC-OUTBOX-0001")
@pytest.mark.integration
def test_project_create_returns_consistent_published_final_state(project_outbox_runtime):
    value = project_outbox_runtime.state()
    assert value
"""
    )

    for node in (swapped, unknown, wrong_layer):
        assert isinstance(node, ast.FunctionDef)
        assert not _test_function_is_exact(
            node,
            "assert_project_create_returns_consistent_published_final_state",
            "TC-OUTBOX-0001",
        )
    assert _violations(escaped, suite=True) == {"direct-runtime-access"}


def test_project_outbox_boundary_rejects_pytest_controls_and_nonliteral_imports() -> None:
    tree = ast.parse(
        """import pytest
from hub_testkit.harness.project_outbox import ProjectOutboxRuntime, assert_unknown
from hub_testkit.harness.pytest_project_outbox import cleanup_project_outbox_runtime

@pytest.mark.parametrize("value", [1])
@pytest.mark.skip
def test_escape(project_outbox_runtime):
    assert_unknown(project_outbox_runtime)
    cleanup_project_outbox_runtime(project_outbox_runtime)
"""
    )

    assert _violations(tree, suite=True) == {
        "direct-runtime-access",
        "suite-import-not-allowed",
        "suite-pytest-control",
    }


def test_project_outbox_suite_is_exactly_five_thin_mapped_nodes() -> None:
    path = SUITE_ROOT / "test_project_outbox.py"
    tree = ast.parse(path.read_text(encoding="utf-8"))

    assert _suite_imports_are_exact(tree)
    assert len(tree.body) == 2 + len(TEST_FUNCTIONS)
    assert all(
        isinstance(node, (ast.Import, ast.ImportFrom, ast.FunctionDef))
        for node in tree.body
    )
    functions = tuple(
        node for node in tree.body if isinstance(node, ast.FunctionDef)
    )
    assert tuple(node.name for node in functions) == TEST_FUNCTIONS
    for node in functions:
        assert _test_function_is_exact(
            node,
            TEST_TO_ASSERTION[node.name],
            TEST_TO_CASE_ID[node.name],
        )
    assert _violations(tree, suite=True) == set()


def test_project_outbox_conftest_and_source_tree_keep_closed_boundary() -> None:
    conftest_tree = ast.parse(
        (SUITE_ROOT / "conftest.py").read_text(encoding="utf-8")
    )
    assert _conftest_imports_are_exact(conftest_tree)
    functions = tuple(
        node for node in conftest_tree.body if isinstance(node, ast.FunctionDef)
    )
    assert len(conftest_tree.body) == 6
    assert len(functions) == 1
    assert _conftest_fixture_is_exact(functions[0])
    assert _violations(conftest_tree, conftest=True) == set()

    assert tuple(path.name for path in DOMAIN_PATHS) == (
        "__init__.py",
        "api.py",
        "builders.py",
        "models.py",
        "oracles.py",
    )
    source_paths = DOMAIN_PATHS + (
        HARNESS_ROOT / "__init__.py",
        HARNESS_ROOT / "project_outbox.py",
        HARNESS_ROOT / "pytest_project_outbox.py",
    )
    for path in source_paths:
        tree = ast.parse(path.read_text(encoding="utf-8"))
        assert _violations(tree, domain=path.parent == DOMAIN_ROOT) == set()
        assert _runtime_claim_is_bounded(_source_claim_text(tree))
        owners = _object_new_owners(tree)
        if path == DOMAIN_ROOT / "models.py":
            assert owners == {
                "_projection_create_safe",
                "_provisioned_create_safe",
                "_receipt_create_safe",
                "_seed_create_safe",
                "_state_create_safe",
            }
        elif path == HARNESS_ROOT / "project_outbox.py":
            assert owners == {"_runtime_create_safe"}
        else:
            assert owners == set()

    harness_tree = ast.parse(
        (HARNESS_ROOT / "project_outbox.py").read_text(encoding="utf-8")
    )
    public_assertions = tuple(
        node
        for node in harness_tree.body
        if isinstance(node, ast.FunctionDef) and node.name.startswith("assert_")
    )
    assert tuple(node.name for node in public_assertions) == ASSERTIONS
    assert all(_assertion_signature_is_exact(node) for node in public_assertions)
    runtime_class = next(
        node
        for node in harness_tree.body
        if isinstance(node, ast.ClassDef) and node.name == "ProjectOutboxRuntime"
    )
    public_methods = tuple(
        node.name
        for node in runtime_class.body
        if isinstance(node, ast.FunctionDef) and not node.name.startswith("_")
    )
    assert public_methods == (
        "provision",
        "create",
        "create_anonymous",
        "create_malformed",
        "observe_project_created",
        "state",
        "disconnect_observer",
        "reconnect_observer",
    )


def _assert_generic_failure(
    operation: object,
    expected_error: type[Exception],
    expected_message: str,
) -> None:
    try:
        assert callable(operation)
        operation()
    except Exception as error:
        rendered = "".join(
            traceback.TracebackException.from_exception(
                error,
                capture_locals=True,
            ).format()
        )
        assert type(error) is expected_error
        assert str(error) == expected_message
        assert error.__cause__ is None
        assert error.__suppress_context__ is True
        assert CONSTRUCTION_CANARY not in rendered
    else:
        raise AssertionError("opaque project outbox object was admitted")


def _opaque_instance_with_canary(constructor: type[object]) -> object:
    instance = object.__new__(constructor)
    if constructor is ProjectOutboxSeed:
        object.__setattr__(instance, "_nonce", CONSTRUCTION_CANARY)
    elif constructor is ProvisionedProjectOutbox:
        object.__setattr__(instance, "_owner_token", CONSTRUCTION_CANARY)
    elif constructor is ProjectCreatedReceipt:
        object.__setattr__(instance, "_event_id", CONSTRUCTION_CANARY)
    elif constructor is OwnedProjectOutboxState:
        object.__setattr__(instance, "_runtime_binding", CONSTRUCTION_CANARY)
    elif constructor is ProjectProjection:
        object.__setattr__(instance, "_project_id", CONSTRUCTION_CANARY)
    elif constructor is ProjectOutboxRuntime:
        object.__setattr__(instance, "_delegate", CONSTRUCTION_CANARY)
    return instance


def _runtime_claim_is_bounded(claim: object) -> bool:
    if type(claim) is not str:
        return False
    normalized = " ".join(claim.casefold().split())
    return all(fragment not in normalized for fragment in RUNTIME_PROOF_OVERCLAIMS)


def _case_claim_text(path: Path) -> str:
    text = path.read_text(encoding="utf-8")
    opening, frontmatter, body = text.split("+++\n", 2)
    assert opening == ""
    metadata = tomllib.loads(frontmatter)
    claim_fields = (
        metadata["title"],
        metadata["subject"],
        *metadata["preconditions"],
        metadata["oracle"],
        metadata["cleanup"],
        body,
    )
    assert all(type(value) is str for value in claim_fields)
    return " ".join(claim_fields)


def _source_claim_text(tree: ast.Module) -> str:
    claims: list[str] = []
    for node in ast.walk(tree):
        if isinstance(node, (ast.ClassDef, ast.FunctionDef, ast.AsyncFunctionDef)):
            claims.append(node.name)
            docstring = ast.get_docstring(node, clean=False)
            if docstring is not None:
                claims.append(docstring)
        elif isinstance(node, ast.Raise) and node.exc is not None:
            claims.extend(
                child.value
                for child in ast.walk(node.exc)
                if isinstance(child, ast.Constant) and type(child.value) is str
            )
    return " ".join(claims)


class _Config:
    def __init__(self) -> None:
        self.stash = pytest.Stash()


class _ExplodingFactory:
    def __call__(self) -> object:
        raise RuntimeError(CONSTRUCTION_CANARY)

    def __repr__(self) -> str:
        return "_ExplodingFactory()"


class _HostileFactory:
    def __init__(self) -> None:
        self.cleanup_calls = 0

    def __call__(self) -> _HostileDelegate:
        return _HostileDelegate(self)

    def __repr__(self) -> str:
        return "_HostileFactory()"


class _HostileDelegate:
    def __init__(self, factory: _HostileFactory) -> None:
        self._factory = factory

    def send(self, request: HttpRequest) -> HttpExchange:
        del request
        raise RuntimeError(CONSTRUCTION_CANARY)

    def provision(self, seed: ProjectOutboxSeed) -> ProvisionedProjectOutbox:
        del seed
        raise RuntimeError(CONSTRUCTION_CANARY)

    def observe_project_created(self) -> ProjectCreatedReceipt:
        raise RuntimeError(CONSTRUCTION_CANARY)

    def state(self) -> OwnedProjectOutboxState:
        raise RuntimeError(CONSTRUCTION_CANARY)

    def disconnect_observer(self) -> None:
        raise RuntimeError(CONSTRUCTION_CANARY)

    def reconnect_observer(self) -> None:
        raise RuntimeError(CONSTRUCTION_CANARY)

    def cleanup(self) -> None:
        self._factory.cleanup_calls += 1
        raise RuntimeError(CONSTRUCTION_CANARY)

    def __repr__(self) -> str:
        return "_HostileDelegate()"


def _suite_imports_are_exact(tree: ast.Module) -> bool:
    imports = tuple(
        node for node in tree.body if isinstance(node, (ast.Import, ast.ImportFrom))
    )
    if len(imports) != 2:
        return False
    first, second = imports
    return (
        isinstance(first, ast.Import)
        and tuple(alias.name for alias in first.names) == ("pytest",)
        and isinstance(second, ast.ImportFrom)
        and second.module == "hub_testkit.harness.project_outbox"
        and tuple(alias.name for alias in second.names) == ASSERTIONS
        and all(alias.asname is None for alias in second.names)
    )


def _conftest_imports_are_exact(tree: ast.Module) -> bool:
    imports = tuple(
        node for node in tree.body if isinstance(node, (ast.Import, ast.ImportFrom))
    )
    actual = tuple(
        (
            "import" if isinstance(node, ast.Import) else node.module,
            tuple(alias.name for alias in node.names),
        )
        for node in imports
    )
    return actual == (
        ("__future__", ("annotations",)),
        ("collections.abc", ("Iterator",)),
        ("import", ("pytest",)),
        ("hub_testkit.harness.project_outbox", ("ProjectOutboxRuntime",)),
        (
            "hub_testkit.harness.pytest_project_outbox",
            (
                "cleanup_project_outbox_runtime",
                "create_project_outbox_runtime_from_config",
            ),
        ),
    )


def _test_function_is_exact(
    node: ast.FunctionDef,
    assertion: str,
    case_id: str,
) -> bool:
    if len(node.decorator_list) != 2:
        return False
    if len(node.args.args) != 1 or node.args.args[0].arg != "project_outbox_runtime":
        return False
    if node.args.vararg is not None or node.args.kwarg is not None:
        return False
    if node.args.kwonlyargs or node.args.defaults or node.args.kw_defaults:
        return False
    if not _decorator_is_case_id(node.decorator_list[0], case_id):
        return False
    if not _decorator_is_integration(node.decorator_list[1]):
        return False
    if len(node.body) != 1 or not isinstance(node.body[0], ast.Expr):
        return False
    call = node.body[0].value
    return (
        isinstance(call, ast.Call)
        and isinstance(call.func, ast.Name)
        and call.func.id == assertion
        and len(call.args) == 1
        and isinstance(call.args[0], ast.Name)
        and call.args[0].id == "project_outbox_runtime"
        and not call.keywords
    )


def _conftest_fixture_is_exact(node: ast.FunctionDef) -> bool:
    if node.name != "project_outbox_runtime":
        return False
    if len(node.decorator_list) != 1 or _marker_name(node.decorator_list[0]) != "fixture":
        return False
    if len(node.args.args) != 1 or node.args.args[0].arg != "request":
        return False
    if len(node.body) != 2 or not isinstance(node.body[0], ast.Assign):
        return False
    lifecycle = node.body[1]
    if not isinstance(lifecycle, ast.Try) or len(lifecycle.finalbody) != 1:
        return False
    cleanup = lifecycle.finalbody[0]
    return (
        isinstance(cleanup, ast.Expr)
        and isinstance(cleanup.value, ast.Call)
        and isinstance(cleanup.value.func, ast.Name)
        and cleanup.value.func.id == "cleanup_project_outbox_runtime"
        and len(cleanup.value.args) == 1
        and isinstance(cleanup.value.args[0], ast.Name)
        and cleanup.value.args[0].id == "runtime"
    )


def _assertion_signature_is_exact(node: ast.FunctionDef) -> bool:
    return (
        len(node.args.args) == 1
        and node.args.args[0].arg == "runtime"
        and isinstance(node.args.args[0].annotation, ast.Name)
        and node.args.args[0].annotation.id == "ProjectOutboxRuntime"
        and isinstance(node.returns, ast.Constant)
        and node.returns.value is None
        and node.args.vararg is None
        and node.args.kwarg is None
        and not node.args.kwonlyargs
        and not node.args.defaults
        and not node.args.kw_defaults
    )


def _decorator_is_case_id(node: ast.expr, case_id: str) -> bool:
    return (
        isinstance(node, ast.Call)
        and isinstance(node.func, ast.Attribute)
        and node.func.attr == "case_id"
        and isinstance(node.func.value, ast.Attribute)
        and node.func.value.attr == "mark"
        and isinstance(node.func.value.value, ast.Name)
        and node.func.value.value.id == "pytest"
        and len(node.args) == 1
        and isinstance(node.args[0], ast.Constant)
        and node.args[0].value == case_id
        and not node.keywords
    )


def _decorator_is_integration(node: ast.expr) -> bool:
    return (
        isinstance(node, ast.Attribute)
        and node.attr == "integration"
        and isinstance(node.value, ast.Attribute)
        and node.value.attr == "mark"
        and isinstance(node.value.value, ast.Name)
        and node.value.value.id == "pytest"
    )


def _violations(
    tree: ast.AST,
    *,
    suite: bool = False,
    conftest: bool = False,
    domain: bool = False,
) -> set[str]:
    visitor = _ProjectOutboxBoundaryVisitor(
        suite=suite,
        conftest=conftest,
        domain=domain,
    )
    visitor.visit(tree)
    return visitor.violations


class _ProjectOutboxBoundaryVisitor(ast.NodeVisitor):
    def __init__(self, *, suite: bool, conftest: bool, domain: bool) -> None:
        self.suite = suite
        self.conftest = conftest
        self.domain = domain
        self.violations: set[str] = set()
        self.aliases: dict[str, str] = {}
        self.runtime_aliases = {"project_outbox_runtime"}

    def visit_Import(self, node: ast.Import) -> None:
        for alias in node.names:
            root = alias.name.split(".", 1)[0]
            self.aliases[alias.asname or root] = root
            if root in FORBIDDEN_IMPORT_ROOTS:
                self.violations.add("forbidden-import")
            if _is_forbidden_layer_import(alias.name, domain=self.domain):
                self.violations.add("layer-import")
            if self.suite and root != "pytest":
                self.violations.add("suite-import-not-allowed")

    def visit_ImportFrom(self, node: ast.ImportFrom) -> None:
        module = node.module or ""
        root = module.split(".", 1)[0]
        for alias in node.names:
            self.aliases[alias.asname or alias.name] = alias.name
        if root in FORBIDDEN_IMPORT_ROOTS:
            self.violations.add("forbidden-import")
        if _is_forbidden_layer_import(module, domain=self.domain):
            self.violations.add("layer-import")
        if self.suite and not (
            module == "hub_testkit.harness.project_outbox"
            and all(alias.name in ASSERTIONS and alias.asname is None for alias in node.names)
        ):
            self.violations.add("suite-import-not-allowed")

    def visit_Assign(self, node: ast.Assign) -> None:
        self.generic_visit(node)
        if isinstance(node.value, ast.Name) and node.value.id in self.runtime_aliases:
            for target in node.targets:
                if isinstance(target, ast.Name):
                    self.runtime_aliases.add(target.id)

    def visit_AnnAssign(self, node: ast.AnnAssign) -> None:
        self.generic_visit(node)
        if (
            isinstance(node.target, ast.Name)
            and isinstance(node.value, ast.Name)
            and node.value.id in self.runtime_aliases
        ):
            self.runtime_aliases.add(node.target.id)

    def visit_Attribute(self, node: ast.Attribute) -> None:
        if self.suite and node.attr.startswith("_"):
            self.violations.add("private-access")
        if isinstance(node.value, ast.Name) and node.value.id in self.runtime_aliases:
            self.violations.add("direct-runtime-access")
        if node.attr in {
            "environ",
            "getenv",
            "read_bytes",
            "read_text",
            "write_bytes",
            "write_text",
        }:
            self.violations.add("forbidden-io")
        self.generic_visit(node)

    def visit_Call(self, node: ast.Call) -> None:
        name = _call_name(node.func)
        resolved = self.aliases.get(name, name)
        if name in {"__import__", "import_module"} or resolved in {
            "__import__",
            "import_module",
        }:
            self.violations.add("dynamic-import")
        if name == "getattr" or resolved == "getattr":
            if _has_private_literal(node.args):
                self.violations.add("private-access")
            if self.suite:
                self.violations.add("direct-runtime-access")
        if isinstance(node.func, ast.Attribute):
            if node.func.attr == "import_module":
                self.violations.add("dynamic-import")
            if self.suite and node.func.attr in SUITE_DIRECT_CALLS:
                self.violations.add("direct-runtime-access")
            if node.func.attr in GLOBAL_FORBIDDEN_CALLS:
                self.violations.add("forbidden-io")
        if resolved in GLOBAL_FORBIDDEN_CALLS:
            self.violations.add("forbidden-io")
        if self.suite and name in SUITE_DIRECT_CALLS:
            self.violations.add("direct-runtime-access")
        self.generic_visit(node)

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        if self.suite:
            for decorator in node.decorator_list:
                if _marker_name(decorator) in {"parametrize", "skip", "skipif", "xfail"}:
                    self.violations.add("suite-pytest-control")
        self.generic_visit(node)

    def visit_While(self, node: ast.While) -> None:
        if isinstance(node.test, ast.Constant) and node.test.value is True:
            self.violations.add("forbidden-io")
        self.generic_visit(node)


def _call_name(node: ast.expr) -> str:
    if isinstance(node, ast.Name):
        return node.id
    if isinstance(node, ast.Attribute):
        return node.attr
    return ""


def _marker_name(node: ast.expr) -> str:
    target = node.func if isinstance(node, ast.Call) else node
    return target.attr if isinstance(target, ast.Attribute) else ""


def _has_private_literal(arguments: list[ast.expr]) -> bool:
    if len(arguments) < 2 or not isinstance(arguments[1], ast.Constant):
        return False
    value = arguments[1].value
    return type(value) is str and value.startswith("_")


def _object_new_owners(tree: ast.AST) -> set[str]:
    owners: set[str] = set()
    for function in (
        node for node in ast.walk(tree) if isinstance(node, ast.FunctionDef)
    ):
        if any(
            isinstance(node, ast.Call)
            and isinstance(node.func, ast.Attribute)
            and node.func.attr == "__new__"
            and isinstance(node.func.value, ast.Name)
            and node.func.value.id == "object"
            for node in ast.walk(function)
        ):
            owners.add(function.name)
    return owners


def _is_forbidden_layer_import(module: str, *, domain: bool) -> bool:
    forbidden = (
        "cases",
        "selftests",
        "testframework.cases",
        "testframework.selftests",
        "testframework.tests",
        "tests",
    )
    return module.startswith(forbidden) or (
        domain and module.startswith("hub_testkit.harness")
    )
