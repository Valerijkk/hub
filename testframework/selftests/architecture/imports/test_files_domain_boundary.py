from __future__ import annotations

import ast
import traceback
from pathlib import Path

from hub_testkit.domains.files import (
    DownloadTicket,
    FilesModelError,
    UploadContent,
    UploadTicket,
)
from hub_testkit.harness.files import FilesRuntime, FilesRuntimeError


PROJECT_ROOT = Path(__file__).resolve().parents[3]
FILES_DOMAIN_ROOT = PROJECT_ROOT / "src" / "hub_testkit" / "domains" / "files"
HARNESS_ROOT = PROJECT_ROOT / "src" / "hub_testkit" / "harness"
SUITE_ROOT = PROJECT_ROOT / "tests" / "api" / "files"

ASSERTIONS = (
    "assert_presigned_lifecycle_download_and_soft_delete",
    "assert_confirm_missing_object_then_retry_succeeds",
    "assert_confirm_and_delete_are_single_transition",
    "assert_files_require_authz_and_hide_storage_artifacts",
    "assert_upload_rejects_bounded_size_and_unsafe_media",
)
TEST_FUNCTIONS = (
    "test_presigned_lifecycle_download_and_soft_delete",
    "test_confirm_missing_object_then_retry_succeeds",
    "test_confirm_and_delete_are_single_transition",
    "test_files_require_authz_and_hide_storage_artifacts",
    "test_upload_rejects_bounded_size_and_unsafe_media",
)
TEST_TO_ASSERTION = {
    "test_presigned_lifecycle_download_and_soft_delete": (
        "assert_presigned_lifecycle_download_and_soft_delete"
    ),
    "test_confirm_missing_object_then_retry_succeeds": (
        "assert_confirm_missing_object_then_retry_succeeds"
    ),
    "test_confirm_and_delete_are_single_transition": (
        "assert_confirm_and_delete_are_single_transition"
    ),
    "test_files_require_authz_and_hide_storage_artifacts": (
        "assert_files_require_authz_and_hide_storage_artifacts"
    ),
    "test_upload_rejects_bounded_size_and_unsafe_media": (
        "assert_upload_rejects_bounded_size_and_unsafe_media"
    ),
}
TEST_TO_CASE_ID = {
    "test_presigned_lifecycle_download_and_soft_delete": "TC-FILES-0001",
    "test_confirm_missing_object_then_retry_succeeds": "TC-FILES-0002",
    "test_confirm_and_delete_are_single_transition": "TC-FILES-0003",
    "test_files_require_authz_and_hide_storage_artifacts": "TC-FILES-0004",
    "test_upload_rejects_bounded_size_and_unsafe_media": "TC-FILES-0005",
}
FORBIDDEN_IMPORT_ROOTS = frozenset(
    {
        "httpx",
        "os",
        "pathlib",
        "requests",
        "socket",
        "subprocess",
        "time",
        "urllib",
    }
)
FORBIDDEN_CALLS = frozenset(
    {
        "AsyncClient",
        "Client",
        "HttpxTransport",
        "cleanup",
        "fetch",
        "getenv",
        "open",
        "provision",
        "put",
        "send",
        "sleep",
    }
)
GLOBAL_FORBIDDEN_CALLS = frozenset(
    {"AsyncClient", "Client", "HttpxTransport", "getenv", "open", "sleep"}
)
CONSTRUCTION_CANARY = "opaque-construction-canary-secret"


def test_opaque_files_objects_reject_direct_construction_without_local_leaks() -> None:
    expectations = (
        (UploadTicket, FilesModelError, "files model is invalid"),
        (DownloadTicket, FilesModelError, "files model is invalid"),
        (UploadContent, FilesModelError, "files model is invalid"),
        (FilesRuntime, FilesRuntimeError, "files runtime is invalid"),
    )

    for constructor, expected_error, expected_message in expectations:
        try:
            constructor()
        except Exception as direct_error:
            assert type(direct_error) is expected_error
            assert str(direct_error) == expected_message
            assert direct_error.__cause__ is None
            assert direct_error.__suppress_context__ is True
        else:
            raise AssertionError("opaque direct construction was admitted")

        instance = _opaque_instance_with_canary(constructor)
        try:
            instance.__init__()
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
            assert CONSTRUCTION_CANARY not in repr(instance)
            assert CONSTRUCTION_CANARY not in rendered
        else:
            raise AssertionError("opaque direct construction was admitted")


def _opaque_instance_with_canary(constructor: type[object]) -> object:
    instance = object.__new__(constructor)
    if constructor is UploadTicket:
        object.__setattr__(instance, "_upload_url", CONSTRUCTION_CANARY)
    elif constructor is DownloadTicket:
        object.__setattr__(instance, "_download_url", CONSTRUCTION_CANARY)
    elif constructor is UploadContent:
        object.__setattr__(instance, "_body", CONSTRUCTION_CANARY.encode("utf-8"))
    elif constructor is FilesRuntime:
        object.__setattr__(instance, "_delegate", CONSTRUCTION_CANARY)
    return instance


def test_files_boundary_detects_alias_dead_dynamic_private_and_runtime_bypasses() -> None:
    tree = ast.parse(
        """import importlib as loader
import httpx as h
import os as system
from builtins import getattr as lookup
from builtins import __import__ as import_any
from hub_testkit.http import HttpxTransport as DefaultTransport
from pathlib import Path as P
from time import sleep as pause

alias = files_runtime
typed_alias: object = files_runtime
if False:
    alias.provision(seed)
    typed_alias.fetch(ticket)
    files_runtime.put(ticket, content)
    files_runtime.fetch(ticket)
    delegate.send(request)
    delegate.cleanup()
    ticket._upload_url
    lookup(ticket, "_object_key")
    object.__getattribute__(ticket, "_runtime_binding")
    type(ticket).__dataclass_fields__["_upload_url"].__get__(ticket, type(ticket))
    build_ticket(url="http://outside.invalid/object", object_key="foreign/key")
    h.Client()
    DefaultTransport(client, policy, redactor)
    pause(1)
    system.getenv("SECRET")
    P(".").read_text()
    loader.import_module(module_name)
    import_any(module_name)
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


def test_files_boundary_allows_only_the_five_semantic_calls_in_suite_code() -> None:
    source = "import pytest\nfrom hub_testkit.harness.files import (\n"
    source += "".join(f"    {name},\n" for name in ASSERTIONS)
    source += ")\n"
    for function_name, assertion_name in TEST_TO_ASSERTION.items():
        source += (
            f'@pytest.mark.case_id("{TEST_TO_CASE_ID[function_name]}")\n'
            "@pytest.mark.api\n"
            f"def {function_name}(files_runtime):\n"
            f"    {assertion_name}(files_runtime)\n"
        )

    assert _violations(ast.parse(source), suite=True) == set()


def test_files_boundary_rejects_swapped_and_wrong_exact_case_ids() -> None:
    swapped = ast.parse(
        """@pytest.mark.case_id("TC-FILES-0002")
@pytest.mark.api
def test_presigned_lifecycle_download_and_soft_delete(files_runtime):
    assert_presigned_lifecycle_download_and_soft_delete(files_runtime)
"""
    ).body[0]
    wrong = ast.parse(
        """@pytest.mark.case_id("TC-FILES-9999")
@pytest.mark.api
def test_presigned_lifecycle_download_and_soft_delete(files_runtime):
    assert_presigned_lifecycle_download_and_soft_delete(files_runtime)
"""
    ).body[0]

    assert isinstance(swapped, ast.FunctionDef)
    assert isinstance(wrong, ast.FunctionDef)
    assert not _test_function_is_exact(
        swapped,
        "assert_presigned_lifecycle_download_and_soft_delete",
        "TC-FILES-0001",
    )
    assert not _test_function_is_exact(
        wrong,
        "assert_presigned_lifecycle_download_and_soft_delete",
        "TC-FILES-0001",
    )


def test_files_boundary_rejects_nonwhitelisted_suite_imports_and_pytest_controls() -> None:
    tree = ast.parse(
        """import pytest
from hub_testkit.domains.files import UploadTicket
from hub_testkit.harness.files import FilesRuntime, assert_unknown
from hub_testkit.harness.pytest_files import cleanup_files_runtime

@pytest.mark.parametrize("value", [1])
@pytest.mark.skip
def test_escape(files_runtime):
    assert_unknown(files_runtime)
    cleanup_files_runtime(files_runtime)
"""
    )

    assert _violations(tree, suite=True) == {
        "direct-runtime-access",
        "suite-import-not-allowed",
        "suite-pytest-control",
    }


def test_files_suite_is_exactly_five_thin_mapped_nodes() -> None:
    path = SUITE_ROOT / "test_files_api.py"
    tree = ast.parse(path.read_text(encoding="utf-8"))

    assert _suite_imports_are_exact(tree)
    assert len(tree.body) == 2 + len(TEST_FUNCTIONS)
    assert all(
        isinstance(node, (ast.Import, ast.ImportFrom, ast.FunctionDef))
        for node in tree.body
    )
    functions = tuple(
        node for node in tree.body if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef))
    )
    assert tuple(node.name for node in functions) == TEST_FUNCTIONS
    for node in functions:
        assert _test_function_is_exact(
            node,
            TEST_TO_ASSERTION[node.name],
            TEST_TO_CASE_ID[node.name],
        )
    assert _violations(tree, suite=True) == set()


def test_files_conftest_and_source_tree_keep_the_closed_boundary() -> None:
    conftest = SUITE_ROOT / "conftest.py"
    conftest_tree = ast.parse(conftest.read_text(encoding="utf-8"))
    assert _conftest_imports_are_exact(conftest_tree)
    conftest_functions = tuple(
        node
        for node in conftest_tree.body
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef))
    )
    assert len(conftest_tree.body) == 6
    assert all(
        isinstance(node, (ast.Import, ast.ImportFrom, ast.FunctionDef))
        for node in conftest_tree.body
    )
    assert len(conftest_functions) == 1
    assert _conftest_fixture_is_exact(conftest_functions[0])
    assert _violations(conftest_tree, conftest=True) == set()

    domain_paths = tuple(sorted(FILES_DOMAIN_ROOT.glob("*.py")))
    assert tuple(path.name for path in domain_paths) == (
        "__init__.py",
        "api.py",
        "builders.py",
        "models.py",
        "oracles.py",
    )
    source_paths = domain_paths + (
        HARNESS_ROOT / "__init__.py",
        HARNESS_ROOT / "files.py",
        HARNESS_ROOT / "pytest_files.py",
    )
    for path in source_paths:
        tree = ast.parse(path.read_text(encoding="utf-8"))
        assert _violations(tree, domain=path.parent == FILES_DOMAIN_ROOT) == set()

    harness_tree = ast.parse(
        (HARNESS_ROOT / "files.py").read_text(encoding="utf-8")
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
        if isinstance(node, ast.ClassDef) and node.name == "FilesRuntime"
    )
    public_runtime_methods = tuple(
        node.name
        for node in runtime_class.body
        if isinstance(node, ast.FunctionDef) and not node.name.startswith("_")
    )
    assert public_runtime_methods == ("provision", "put", "fetch")


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
        and second.module == "hub_testkit.harness.files"
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
        ("hub_testkit.harness.files", ("FilesRuntime",)),
        (
            "hub_testkit.harness.pytest_files",
            ("cleanup_files_runtime", "create_files_runtime_from_config"),
        ),
    )


def _test_function_is_exact(
    node: ast.FunctionDef | ast.AsyncFunctionDef,
    assertion: str,
    case_id: str,
) -> bool:
    if isinstance(node, ast.AsyncFunctionDef) or len(node.decorator_list) != 2:
        return False
    if len(node.args.args) != 1 or node.args.args[0].arg != "files_runtime":
        return False
    if node.args.vararg is not None or node.args.kwarg is not None:
        return False
    if node.args.kwonlyargs or node.args.defaults or node.args.kw_defaults:
        return False
    if not _decorator_is_case_id(node.decorator_list[0], case_id):
        return False
    if not _decorator_is_api(node.decorator_list[1]):
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
        and call.args[0].id == "files_runtime"
        and not call.keywords
    )


def _conftest_fixture_is_exact(node: ast.FunctionDef | ast.AsyncFunctionDef) -> bool:
    if isinstance(node, ast.AsyncFunctionDef) or node.name != "files_runtime":
        return False
    if len(node.decorator_list) != 1 or _marker_name(node.decorator_list[0]) != "fixture":
        return False
    if len(node.args.args) != 1 or node.args.args[0].arg != "request":
        return False
    if len(node.body) != 2 or not isinstance(node.body[0], ast.Assign):
        return False
    if not isinstance(node.body[1], ast.Try) or len(node.body[1].finalbody) != 1:
        return False
    cleanup = node.body[1].finalbody[0]
    return (
        isinstance(cleanup, ast.Expr)
        and isinstance(cleanup.value, ast.Call)
        and isinstance(cleanup.value.func, ast.Name)
        and cleanup.value.func.id == "cleanup_files_runtime"
        and len(cleanup.value.args) == 1
        and isinstance(cleanup.value.args[0], ast.Name)
        and cleanup.value.args[0].id == "runtime"
    )


def _assertion_signature_is_exact(node: ast.FunctionDef) -> bool:
    return (
        len(node.args.args) == 1
        and node.args.args[0].arg == "runtime"
        and isinstance(node.args.args[0].annotation, ast.Name)
        and node.args.args[0].annotation.id == "FilesRuntime"
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


def _decorator_is_api(node: ast.expr) -> bool:
    return (
        isinstance(node, ast.Attribute)
        and node.attr == "api"
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
    visitor = _FilesBoundaryVisitor(suite=suite, conftest=conftest, domain=domain)
    visitor.visit(tree)
    return visitor.violations


class _FilesBoundaryVisitor(ast.NodeVisitor):
    def __init__(self, *, suite: bool, conftest: bool, domain: bool) -> None:
        self.suite = suite
        self.conftest = conftest
        self.domain = domain
        self.violations: set[str] = set()
        self.aliases: dict[str, str] = {}
        self.runtime_aliases = {"files_runtime"}

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
            module == "hub_testkit.harness.files"
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
            if self.suite and node.func.attr in FORBIDDEN_CALLS:
                self.violations.add("direct-runtime-access")
            if node.func.attr in GLOBAL_FORBIDDEN_CALLS:
                self.violations.add("forbidden-io")
        if resolved in GLOBAL_FORBIDDEN_CALLS:
            self.violations.add("forbidden-io")
        if self.suite and name not in ASSERTIONS and name not in {"case_id"}:
            if name in {"cleanup_files_runtime", "assert_unknown"}:
                self.violations.add("direct-runtime-access")
        if self.suite and any(
            keyword.arg in {
                "download_url",
                "key",
                "object_key",
                "s3_key",
                "upload_url",
                "url",
            }
            for keyword in node.keywords
        ):
            self.violations.add("direct-runtime-access")
        self.generic_visit(node)

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        if self.suite:
            for decorator in node.decorator_list:
                marker = _marker_name(decorator)
                if marker in {"parametrize", "skip", "skipif", "xfail"}:
                    self.violations.add("suite-pytest-control")
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
