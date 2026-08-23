from __future__ import annotations

import ast
from pathlib import Path


PROJECT_ROOT = Path(__file__).resolve().parents[3]
SOURCE_ROOT = PROJECT_ROOT / "src" / "hub_testkit"
EXECUTABLE_ROOTS = (PROJECT_ROOT / "selftests", PROJECT_ROOT / "tests")
PLUGIN_ROOT = SOURCE_ROOT / "pytest_plugin"
ALLOWED_HOOKS = {
    "pytest_addoption",
    "pytest_configure",
    "pytest_collection_finish",
}


def _python_files(root: Path) -> tuple[Path, ...]:
    return tuple(sorted(path for path in root.rglob("*.py") if path.is_file()))


_FORBIDDEN_PLUGIN_IMPORTS = frozenset({"subprocess", "socket", "httpx", "requests", "docker"})


def _violations(tree: ast.AST, *, plugin_source: bool) -> set[str]:
    visitor = _BoundaryVisitor(plugin_source=plugin_source)
    visitor.visit(tree)
    return visitor.violations


def test_ast_guards_distinguish_plugin_io_from_selftest_fixtures() -> None:
    fixture_tree = ast.parse(
        'from pathlib import Path\npath = Path("case.md")\npath.write_text("fixture")'
    )
    write_open_tree = ast.parse(
        'from pathlib import Path\npath = Path("case.md")\npath.open("w")'
    )
    sleep_tree = ast.parse("import time\ntime.sleep(1)")

    assert _violations(fixture_tree, plugin_source=False) == set()
    assert _violations(fixture_tree, plugin_source=True) == {"plugin-write"}
    assert _violations(write_open_tree, plugin_source=True) == {"plugin-write"}
    assert _violations(sleep_tree, plugin_source=False) == {"sleep"}


def test_ast_suite_import_guard_rejects_bare_and_package_namespaces() -> None:
    for namespace in (
        "tests",
        "tests.api",
        "selftests",
        "selftests.collection",
        "cases",
        "cases.authentication",
        "load",
        "load.scenarios",
        "testframework.tests",
        "testframework.tests.api",
        "testframework.selftests",
        "testframework.selftests.collection",
        "testframework.cases",
        "testframework.cases.authentication",
        "testframework.load",
        "testframework.load.scenarios",
    ):
        assert _is_suite_namespace(namespace)
    assert not _is_suite_namespace("hub_testkit.traceability")
    assert _violations(ast.parse("from .load import value"), plugin_source=False) == set()
    assert _violations(
        ast.parse("from testframework import tests, selftests, cases, load"),
        plugin_source=False,
    ) == {"suite-import"}


def test_alias_resolving_visitor_detects_collection_bypasses() -> None:
    tree = ast.parse(
        """import importlib as loader
import sys as system
import time as clock
from builtins import open as writer
from pathlib import Path as CasePath
from time import sleep as pause

clock.sleep(1)
pause(1)
system.path.append("escape")
system.path[0] = "escape"
writer("report.xml", "w")
write_case = CasePath.write_text
write_case(CasePath("case.md"), "contents")
loader.import_module(target)
__import__("tests.api")
while pending:
    poll()
"""
    )

    violations = _violations(tree, plugin_source=True)

    assert violations == {
        "dynamic-import",
        "plugin-write",
        "sleep",
        "suite-import",
        "sys-path",
        "unbounded-loop",
    }


def test_alias_resolving_visitor_keeps_safe_fixture_and_bounded_read_loop() -> None:
    tree = ast.parse(
        """import time as clock
from pathlib import Path

fixture = Path("case.md")
fixture.write_text("fixture", encoding="utf-8")
while len(data) < limit:
    data.extend(read())
while clock.monotonic() < deadline:
    poll()
"""
    )

    assert _violations(tree, plugin_source=False) == set()


def _is_suite_namespace(name: str) -> bool:
    prefixes = (
        "tests",
        "selftests",
        "cases",
        "load",
        "testframework.tests",
        "testframework.selftests",
        "testframework.cases",
        "testframework.load",
    )
    return any(name == prefix or name.startswith(f"{prefix}.") for prefix in prefixes)


class _BoundaryVisitor(ast.NodeVisitor):
    def __init__(self, *, plugin_source: bool) -> None:
        self.plugin_source = plugin_source
        self.violations: set[str] = set()
        self._scopes: list[dict[str, str]] = [{}]

    def visit_Import(self, node: ast.Import) -> None:
        for alias in node.names:
            if _is_suite_namespace(alias.name):
                self.violations.add("suite-import")
            if alias.name == "_pytest" or alias.name.startswith("_pytest."):
                self.violations.add("private-pytest")
            root = alias.name.split(".", 1)[0]
            local = alias.asname or root
            self._bind(local, _module_token(root))
            if self.plugin_source and root in _FORBIDDEN_PLUGIN_IMPORTS:
                self.violations.add("plugin-operational-import")

    def visit_ImportFrom(self, node: ast.ImportFrom) -> None:
        module = node.module or ""
        if node.level == 0 and _is_suite_namespace(module):
            self.violations.add("suite-import")
        if module == "_pytest" or module.startswith("_pytest."):
            self.violations.add("private-pytest")
        root = module.split(".", 1)[0]
        if self.plugin_source and root in _FORBIDDEN_PLUGIN_IMPORTS:
            self.violations.add("plugin-operational-import")
        for alias in node.names:
            if node.level == 0 and module == "testframework" and alias.name in {
                "tests",
                "selftests",
                "cases",
                "load",
            }:
                self.violations.add("suite-import")
            local = alias.asname or alias.name
            self._bind(local, _from_import_token(module, alias.name))

    def visit_Assign(self, node: ast.Assign) -> None:
        self.visit(node.value)
        token = self._resolve(node.value)
        for target in node.targets:
            self._assign_target(target, token)

    def visit_AnnAssign(self, node: ast.AnnAssign) -> None:
        if node.value is not None:
            self.visit(node.value)
            self._assign_target(node.target, self._resolve(node.value))
        else:
            self._assign_target(node.target, None)

    def visit_AugAssign(self, node: ast.AugAssign) -> None:
        if self._is_sys_path(node.target):
            self.violations.add("sys-path")
        self.visit(node.target)
        self.visit(node.value)

    def visit_NamedExpr(self, node: ast.NamedExpr) -> None:
        self.visit(node.value)
        self._assign_target(node.target, self._resolve(node.value))

    def visit_Call(self, node: ast.Call) -> None:
        token = self._resolve(node.func)
        if token == "sleep":
            self.violations.add("sleep")
        if token in {"dynamic-import", "import-module"}:
            self.violations.add("dynamic-import")
            literal = _literal_string(node.args[0]) if node.args else None
            if literal is not None and _is_suite_namespace(literal):
                self.violations.add("suite-import")
        if self.plugin_source and token == "path-write":
            self.violations.add("plugin-write")
        if self.plugin_source and token == "path-open" and _has_write_mode(
            node.args,
            node.keywords,
        ):
            self.violations.add("plugin-write")
        if self.plugin_source and token == "open" and _has_write_mode(
            node.args[1:],
            node.keywords,
        ):
            self.violations.add("plugin-write")
        if isinstance(node.func, ast.Attribute) and node.func.attr in {
            "append",
            "extend",
            "insert",
            "remove",
            "pop",
        } and self._is_sys_path(node.func.value):
            self.violations.add("sys-path")
        self.generic_visit(node)

    def visit_While(self, node: ast.While) -> None:
        if _is_unbounded_loop(node.test, self):
            self.violations.add("unbounded-loop")
        self.generic_visit(node)

    def visit_FunctionDef(self, node: ast.FunctionDef) -> None:
        self._bind(node.name, None)
        self._visit_scoped(node.args, node.body, node.decorator_list)

    def visit_AsyncFunctionDef(self, node: ast.AsyncFunctionDef) -> None:
        self._bind(node.name, None)
        self._visit_scoped(node.args, node.body, node.decorator_list)

    def visit_Lambda(self, node: ast.Lambda) -> None:
        self._push_scope()
        self._bind_arguments(node.args)
        self.visit(node.body)
        self._pop_scope()

    def visit_ClassDef(self, node: ast.ClassDef) -> None:
        self._bind(node.name, None)
        for decorator in node.decorator_list:
            self.visit(decorator)
        self._push_scope()
        for statement in node.body:
            self.visit(statement)
        self._pop_scope()

    def _visit_scoped(
        self,
        arguments: ast.arguments,
        body: list[ast.stmt],
        decorators: list[ast.expr],
    ) -> None:
        for decorator in decorators:
            self.visit(decorator)
        for default in (*arguments.defaults, *arguments.kw_defaults):
            if default is not None:
                self.visit(default)
        self._push_scope()
        self._bind_arguments(arguments)
        for statement in body:
            self.visit(statement)
        self._pop_scope()

    def _bind_arguments(self, arguments: ast.arguments) -> None:
        all_arguments = (
            *arguments.posonlyargs,
            *arguments.args,
            *arguments.kwonlyargs,
        )
        for argument in all_arguments:
            self._bind(argument.arg, None)
        if arguments.vararg is not None:
            self._bind(arguments.vararg.arg, None)
        if arguments.kwarg is not None:
            self._bind(arguments.kwarg.arg, None)

    def _assign_target(self, target: ast.expr, token: str | None) -> None:
        if self._is_sys_path(target):
            self.violations.add("sys-path")
        if isinstance(target, ast.Name):
            self._bind(target.id, token)
        elif isinstance(target, (ast.Tuple, ast.List)):
            for item in target.elts:
                self._assign_target(item, None)
        else:
            self.visit(target)

    def _resolve(self, value: ast.AST) -> str | None:
        if isinstance(value, ast.Name):
            return self._lookup(value.id)
        if isinstance(value, ast.Attribute):
            base = self._resolve(value.value)
            if base == "time" and value.attr == "sleep":
                return "sleep"
            if base == "time" and value.attr == "monotonic":
                return "monotonic"
            if base == "sys" and value.attr == "path":
                return "sys-path"
            if base == "importlib" and value.attr == "import_module":
                return "import-module"
            if base == "builtins" and value.attr == "__import__":
                return "dynamic-import"
            if base == "pathlib" and value.attr == "Path":
                return "path"
            if base in {"path", "path-instance"} and value.attr.startswith("write"):
                return "path-write"
            if base in {"path", "path-instance"} and value.attr == "open":
                return "path-open"
            return None
        if isinstance(value, ast.Call):
            call_target = self._resolve(value.func)
            if call_target == "path":
                return "path-instance"
            if call_target == "monotonic":
                return "monotonic"
        return None

    def _is_sys_path(self, value: ast.AST) -> bool:
        if isinstance(value, ast.Subscript):
            return self._is_sys_path(value.value)
        return self._resolve(value) == "sys-path"

    def _lookup(self, name: str) -> str | None:
        for scope in reversed(self._scopes):
            if name in scope:
                return scope[name]
        return {"__import__": "dynamic-import", "open": "open"}.get(name)

    def _bind(self, name: str, token: str | None) -> None:
        self._scopes[-1][name] = token or "unknown"

    def _push_scope(self) -> None:
        self._scopes.append({})

    def _pop_scope(self) -> None:
        self._scopes.pop()


def _module_token(name: str) -> str | None:
    return {
        "builtins": "builtins",
        "importlib": "importlib",
        "pathlib": "pathlib",
        "sys": "sys",
        "time": "time",
    }.get(name)


def _from_import_token(module: str, name: str) -> str | None:
    return {
        ("builtins", "__import__"): "dynamic-import",
        ("builtins", "open"): "open",
        ("importlib", "import_module"): "import-module",
        ("pathlib", "Path"): "path",
        ("sys", "path"): "sys-path",
        ("time", "monotonic"): "monotonic",
        ("time", "sleep"): "sleep",
    }.get((module, name))


def _has_write_mode(
    positional: list[ast.expr],
    keywords: list[ast.keyword],
) -> bool:
    modes = (
        *positional,
        *(keyword.value for keyword in keywords if keyword.arg == "mode"),
    )
    return any(
        isinstance(argument, ast.Constant)
        and isinstance(argument.value, str)
        and any(flag in argument.value for flag in ("w", "a", "x", "+"))
        for argument in modes
    )


def _literal_string(value: ast.expr) -> str | None:
    if isinstance(value, ast.Constant) and type(value.value) is str:
        return value.value
    return None


def _is_unbounded_loop(test: ast.expr, visitor: _BoundaryVisitor) -> bool:
    if isinstance(test, ast.Constant) and test.value is True:
        return True
    if isinstance(test, ast.Name):
        name = test.id.casefold()
        if name in {"queue", "stack"} or any(
            fragment in name
            for fragment in ("limit", "max", "count", "remaining", "size", "length")
        ):
            return False
    for comparison in (node for node in ast.walk(test) if isinstance(node, ast.Compare)):
        if _has_clock_or_deadline(comparison, visitor) or _has_finite_bound(comparison):
            return False
    return True


def _has_clock_or_deadline(comparison: ast.Compare, visitor: _BoundaryVisitor) -> bool:
    return any(
        _expression_name(value, visitor) in {"monotonic", "deadline"}
        for value in (comparison.left, *comparison.comparators)
    )


def _has_finite_bound(comparison: ast.Compare) -> bool:
    bound_fragments = ("limit", "max", "count", "remaining", "size", "length", "cursor")
    names = _lexical_names(comparison)
    if "len" in names:
        return True
    return any(
        any(fragment in name.casefold() for fragment in bound_fragments)
        for name in names
    )


def _expression_name(value: ast.expr, visitor: _BoundaryVisitor) -> str | None:
    resolved = visitor._resolve(value)
    if resolved == "monotonic":
        return "monotonic"
    if isinstance(value, ast.Name) and "deadline" in value.id.casefold():
        return "deadline"
    if isinstance(value, ast.Attribute) and "deadline" in value.attr.casefold():
        return "deadline"
    return None


def _lexical_names(value: ast.AST) -> tuple[str, ...]:
    names: list[str] = []
    for node in ast.walk(value):
        if isinstance(node, ast.Name):
            names.append(node.id)
        elif isinstance(node, ast.Attribute):
            names.append(node.attr)
    return tuple(names)


def test_python_boundaries_have_no_suite_or_path_escape_hacks() -> None:
    for root in (SOURCE_ROOT, *EXECUTABLE_ROOTS):
        for path in _python_files(root):
            assert path.name != "utils.py"
            tree = ast.parse(path.read_text(encoding="utf-8"))
            assert not _violations(tree, plugin_source=path.is_relative_to(PLUGIN_ROOT))
            for node in ast.walk(tree):
                assert not (isinstance(node, ast.ClassDef) and node.name == "BaseTest")
                assert not (
                    isinstance(node, ast.While)
                    and isinstance(node.test, ast.Constant)
                    and node.test.value is True
                )


def test_plugin_uses_only_public_hooks_and_no_forbidden_operational_imports() -> None:
    hook_names: set[str] = set()
    for path in _python_files(PLUGIN_ROOT):
        tree = ast.parse(path.read_text(encoding="utf-8"))
        assert not _violations(tree, plugin_source=True)
        for node in ast.walk(tree):
            if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef)) and node.name.startswith(
                "pytest_"
            ):
                hook_names.add(node.name)
    assert hook_names == ALLOWED_HOOKS
