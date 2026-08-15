from __future__ import annotations

import ast
from pathlib import Path


PROJECT_ROOT = Path(__file__).resolve().parents[3]
CONTRACT_ROOT = PROJECT_ROOT / "src" / "hub_testkit" / "contracts"
FORBIDDEN_IMPORT_ROOTS = frozenset(
    {
        "httpx",
        "os",
        "pathlib",
        "requests",
        "socket",
        "subprocess",
        "urllib",
    }
)


def _violations(tree: ast.AST, *, contract_suite: bool = False) -> set[str]:
    visitor = _ContractBoundaryVisitor(contract_suite=contract_suite)
    visitor.visit(tree)
    return visitor.violations


def test_contract_boundary_visitor_detects_io_environment_and_private_bypasses() -> None:
    tree = ast.parse(
        """import importlib as loader
import os as system
from pathlib import Path
from subprocess import run
from builtins import getattr as lookup
from builtins import __import__ as import_any
from importlib import import_module as direct_load

system.environ[\"SECRET\"]
Path(\"source\").read_bytes()
open(\"source\")
run([\"git\", \"show\"])
alias = bundle
alias._snapshot
lookup(alias, "_content_by_path")
object.__getattribute__(alias, "_source_commit")
if False:
    loader.import_module("hub_testkit.contracts.source_bundle")
    __import__("hub_testkit.contracts.source_bundle")
    import_any("hub_testkit.contracts.source_bundle")
    direct_load("hub_testkit.contracts.source_bundle")
"""
    )

    assert _violations(tree) == {
        "forbidden-import",
        "forbidden-io",
        "private-facade-access",
        "dynamic-import",
    }


def test_contract_boundary_visitor_keeps_hash_toml_stash_and_closed_selector_safe() -> None:
    tree = ast.parse(
        """import hashlib
import tomllib
import pytest

digest = hashlib.sha256(raw).hexdigest()
parsed = tomllib.loads(text)
key = pytest.StashKey()
selected = bundle.select(selector)
"""
    )

    assert _violations(tree) == set()


def test_contract_suite_visitor_rejects_raw_snapshot_factory_selector_and_introspection() -> None:
    tree = ast.parse(
        """import hub_testkit.contracts as contracts
from hub_testkit.contracts import ContractSourceBundle as Bundle
from hub_testkit.contracts import SourceBlob as Blob
from hub_testkit.contracts import SourceSnapshot
from hub_testkit.contracts.pytest_source_bundle import ContractSourceBundleFactory as Factory
from hub_testkit.contracts.pytest_source_bundle import """
        "provide_contract_source_bundle_factory as provide\n"
        """

if False:
    snapshot = SourceSnapshot("0123456789abcdef0123456789abcdef01234567", ())
    blob = Blob(None, b"", "0" * 64)
    bundle = Bundle(snapshot)
    provide(config, Factory())
    raw_select = bundle.select
    raw_select(selector)
    bundle.select(selector)
    contract_source_bundle.select(selector)
    raw_getattr_select = getattr(unknown_receiver, "select")
    raw_getattr_select(selector)
    typed_alias: object = contract_source_bundle
    typed_alias.select(selector)
    type(bundle).__dataclass_fields__
    getattr(type(bundle), "__dataclass_fields__")
    Bundle.source_commit.__get__(bundle, Bundle)
    getattr(Bundle.source_commit, "fget")
    getattr(unknown_receiver, "_secret")
    object.__getattribute__(unknown_receiver, "_secret")
    contracts.SourceBlob

def helper(value):
    if False:
        type(value).__dataclass_fields__["_content_by_path"].__get__(value, type(value))
        value._content_by_path
"""
    )

    assert _violations(tree, contract_suite=True) == {
        "suite-raw-contract-api",
        "suite-raw-selector",
        "suite-raw-introspection",
        "suite-contract-import-not-allowed",
        "private-facade-access",
    }


def test_contract_suite_visitor_allows_only_injected_fixture_and_semantic_helpers() -> None:
    tree = ast.parse(
        """from hub_testkit.contracts import contract_source_bundle, load_f1_04_risk_ledger
from hub_testkit.contracts.openapi import assert_magic_verify_error_contract
from hub_testkit.contracts.openapi import assert_openapi_canonical_document
from hub_testkit.contracts.openapi import assert_tms_webhook_auth_contract
from hub_testkit.contracts.public_error import assert_public_error_artifact_contract

def test_contract(contract_source_bundle):
    ledger = load_f1_04_risk_ledger(contract_source_bundle)
    assert_openapi_canonical_document(contract_source_bundle)
    assert_magic_verify_error_contract(contract_source_bundle)
    assert_tms_webhook_auth_contract(contract_source_bundle)
    assert_public_error_artifact_contract(contract_source_bundle)
    return ledger.edge_for_node_id("""
        '"testframework/tests/contract/openapi/'
        "test_openapi_security_contract.py::"
        'test_swagger_document_is_valid_with_single_api_prefix")\n'
    )

    assert _violations(tree, contract_suite=True) == set()


def test_contract_suite_visitor_rejects_every_nonwhitelisted_contract_import() -> None:
    tree = ast.parse(
        """from hub_testkit.contracts import RiskLedgerError
from hub_testkit.contracts.openapi import OpenApiContractError
from hub_testkit.contracts.openapi import unexpected_semantic_helper
from hub_testkit.contracts.public_error import PublicErrorContractError
"""
    )

    assert _violations(tree, contract_suite=True) == {"suite-contract-import-not-allowed"}


def test_contract_source_modules_have_no_io_environment_private_or_second_loader_escape() -> None:
    source_paths = tuple(sorted(CONTRACT_ROOT.glob("*.py")))
    contract_test_root = PROJECT_ROOT / "tests" / "contract"
    contract_test_paths = (
        tuple(sorted(contract_test_root.rglob("*.py"))) if contract_test_root.is_dir() else ()
    )

    assert tuple(path.name for path in source_paths) == (
        "__init__.py",
        "openapi.py",
        "public_error.py",
        "pytest_source_bundle.py",
        "risk_ledger.py",
        "source_bundle.py",
    )
    for path in source_paths:
        tree = ast.parse(path.read_text(encoding="utf-8"))
        assert _violations(tree) == set()
    for path in contract_test_paths:
        tree = ast.parse(path.read_text(encoding="utf-8"))
        assert _violations(tree, contract_suite=True) == set()

    risk_tree = ast.parse((CONTRACT_ROOT / "risk_ledger.py").read_text(encoding="utf-8"))
    loads = [
        node.name
        for node in ast.walk(risk_tree)
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef))
        and node.name.startswith("load")
    ]
    assert loads == ["load_f1_04_risk_ledger"]
    openapi_tree = ast.parse((CONTRACT_ROOT / "openapi.py").read_text(encoding="utf-8"))
    public_openapi_functions = tuple(
        node.name
        for node in openapi_tree.body
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef))
        and not node.name.startswith("_")
    )
    assert public_openapi_functions == (
        "assert_openapi_canonical_document",
        "assert_magic_verify_error_contract",
        "assert_tms_webhook_auth_contract",
    )
    public_error_tree = ast.parse(
        (CONTRACT_ROOT / "public_error.py").read_text(encoding="utf-8")
    )
    public_error_functions = tuple(
        node.name
        for node in public_error_tree.body
        if isinstance(node, (ast.FunctionDef, ast.AsyncFunctionDef))
        and not node.name.startswith("_")
    )
    assert public_error_functions == ("assert_public_error_artifact_contract",)


class _ContractBoundaryVisitor(ast.NodeVisitor):
    def __init__(self, *, contract_suite: bool) -> None:
        self.contract_suite = contract_suite
        self.violations: set[str] = set()
        self._aliases: dict[str, str] = {}
        self._bundle_aliases: set[str] = {
            "bundle",
            "contract_source_bundle",
            "facade",
            "source_bundle",
        }
        self._contract_module_aliases: set[str] = set()
        self._raw_contract_names: set[str] = set()
        self._raw_selector_aliases: set[str] = set()

    def visit_Import(self, node: ast.Import) -> None:
        for alias in node.names:
            root = alias.name.split(".", 1)[0]
            local = alias.asname or root
            self._aliases[local] = root
            if self.contract_suite and alias.name.startswith("hub_testkit.contracts"):
                self._contract_module_aliases.add(local)
                self.violations.add("suite-contract-import-not-allowed")
            if root in FORBIDDEN_IMPORT_ROOTS:
                self.violations.add("forbidden-import")

    def visit_ImportFrom(self, node: ast.ImportFrom) -> None:
        root = (node.module or "").split(".", 1)[0]
        if root in FORBIDDEN_IMPORT_ROOTS:
            self.violations.add("forbidden-import")
        for alias in node.names:
            local = alias.asname or alias.name
            if self.contract_suite and (node.module or "").startswith("hub_testkit.contracts"):
                if not _suite_contract_import_allowed(node.module or "", alias.name):
                    self.violations.add("suite-contract-import-not-allowed")
            if self.contract_suite and _is_raw_contract_symbol(node.module, alias.name):
                self._raw_contract_names.add(local)
                self.violations.add("suite-raw-contract-api")
            if self.contract_suite and node.module == "hub_testkit" and alias.name == "contracts":
                self._contract_module_aliases.add(local)
            if node.module == "builtins" and alias.name == "getattr":
                self._aliases[local] = "getattr"
            elif node.module == "builtins" and alias.name == "__import__":
                self._aliases[local] = "dynamic-import"
            elif node.module == "importlib" and alias.name == "import_module":
                self._aliases[local] = "dynamic-import"
            else:
                self._aliases[local] = root

    def visit_Assign(self, node: ast.Assign) -> None:
        self.generic_visit(node)
        if isinstance(node.value, ast.Name) and node.value.id in self._bundle_aliases:
            for target in node.targets:
                if isinstance(target, ast.Name):
                    self._bundle_aliases.add(target.id)
        if self.contract_suite and _is_raw_selector_expression(node.value, self._bundle_aliases):
            for target in node.targets:
                if isinstance(target, ast.Name):
                    self._raw_selector_aliases.add(target.id)
        if self.contract_suite and _is_selector_getattr_call(node.value, self._aliases):
            for target in node.targets:
                if isinstance(target, ast.Name):
                    self._raw_selector_aliases.add(target.id)

    def visit_AnnAssign(self, node: ast.AnnAssign) -> None:
        self.generic_visit(node)
        if (
            isinstance(node.target, ast.Name)
            and isinstance(node.value, ast.Name)
            and node.value.id in self._bundle_aliases
        ):
            self._bundle_aliases.add(node.target.id)

    def visit_NamedExpr(self, node: ast.NamedExpr) -> None:
        self.generic_visit(node)
        if isinstance(node.target, ast.Name) and isinstance(node.value, ast.Name):
            if node.value.id in self._bundle_aliases:
                self._bundle_aliases.add(node.target.id)

    def visit_Attribute(self, node: ast.Attribute) -> None:
        if self.contract_suite and node.attr.startswith("_"):
            self.violations.add("private-facade-access")
        if self.contract_suite and _is_raw_contract_attribute(
            node,
            self._contract_module_aliases,
        ):
            self.violations.add("suite-raw-contract-api")
        if self.contract_suite and _is_raw_selector_expression(node, self._bundle_aliases):
            self.violations.add("suite-raw-selector")
        if self.contract_suite and _is_raw_introspection(node, self._bundle_aliases):
            self.violations.add("suite-raw-introspection")
        if node.attr.startswith("_") and isinstance(node.value, ast.Name):
            if node.value.id in self._bundle_aliases:
                self.violations.add("private-facade-access")
        if node.attr in {"environ", "getenv", "read_text", "read_bytes", "open"}:
            self.violations.add("forbidden-io")
        self.generic_visit(node)

    def visit_Call(self, node: ast.Call) -> None:
        if self.contract_suite and isinstance(node.func, ast.Name):
            if node.func.id in self._raw_contract_names or node.func.id in _RAW_CONTRACT_SYMBOLS:
                self.violations.add("suite-raw-contract-api")
            if node.func.id in self._raw_selector_aliases:
                self.violations.add("suite-raw-selector")
        if isinstance(node.func, ast.Name) and node.func.id == "open":
            self.violations.add("forbidden-io")
        if (
            isinstance(node.func, ast.Name)
            and self._aliases.get(node.func.id) in FORBIDDEN_IMPORT_ROOTS
        ):
            self.violations.add("forbidden-io")
        if isinstance(node.func, ast.Name) and node.func.id == "__import__":
            self.violations.add("dynamic-import")
        if isinstance(node.func, ast.Name) and self._aliases.get(node.func.id) == "dynamic-import":
            self.violations.add("dynamic-import")
        if isinstance(node.func, ast.Attribute):
            if (
                isinstance(node.func.value, ast.Name)
                and self._aliases.get(node.func.value.id) == "importlib"
                and node.func.attr == "import_module"
            ):
                self.violations.add("dynamic-import")
            if (
                isinstance(node.func.value, ast.Name)
                and node.func.value.id == "object"
                and node.func.attr == "__getattribute__"
                and _has_private_literal(node.args)
            ):
                self.violations.add("private-facade-access")
        if isinstance(node.func, ast.Name) and (
            node.func.id == "getattr" or self._aliases.get(node.func.id) == "getattr"
        ) and _has_private_literal(node.args):
            self.violations.add("private-facade-access")
        if self.contract_suite and _is_selector_getattr_call(node, self._aliases):
            self.violations.add("suite-raw-selector")
        if self.contract_suite and isinstance(node.func, ast.Name) and (
            node.func.id == "getattr" or self._aliases.get(node.func.id) == "getattr"
        ) and _is_type_introspection_lookup(node.args, self._bundle_aliases):
            self.violations.add("suite-raw-introspection")
        self.generic_visit(node)


_RAW_CONTRACT_SYMBOLS = frozenset(
    {
        "CONTRACT_SOURCE_BUNDLE_FACTORY_KEY",
        "ContractSourceBundle",
        "ContractSourceBundleDelegate",
        "ContractSourceBundleFactory",
        "SourceBlob",
        "SourceSnapshot",
        "create_contract_source_bundle_from_config",
        "provide_contract_source_bundle_factory",
        "require_contract_source_bundle_factory",
    }
)
_ALLOWED_CONTRACT_SUITE_IMPORTS = {
    "hub_testkit.contracts": frozenset(
        {
            "contract_source_bundle",
            "load_f1_04_risk_ledger",
        }
    ),
    "hub_testkit.contracts.openapi": frozenset(
        {
            "assert_magic_verify_error_contract",
            "assert_openapi_canonical_document",
            "assert_tms_webhook_auth_contract",
        }
    ),
    "hub_testkit.contracts.public_error": frozenset(
        {"assert_public_error_artifact_contract"}
    ),
}
_INTROSPECTION_ATTRIBUTES = frozenset(
    {
        "__dataclass_fields__",
        "__dataclass_params__",
        "__dict__",
        "__get__",
        "__slots__",
        "fget",
    }
)


def _is_raw_contract_symbol(module: str | None, name: str) -> bool:
    return (
        bool(module)
        and module.startswith("hub_testkit.contracts")
        and name in _RAW_CONTRACT_SYMBOLS
    )


def _suite_contract_import_allowed(module: str, name: str) -> bool:
    allowed = _ALLOWED_CONTRACT_SUITE_IMPORTS.get(module)
    return allowed is not None and name in allowed


def _is_raw_contract_attribute(node: ast.Attribute, module_aliases: set[str]) -> bool:
    return (
        isinstance(node.value, ast.Name)
        and node.value.id in module_aliases
        and node.attr in _RAW_CONTRACT_SYMBOLS
    )


def _is_raw_selector_expression(node: ast.AST, bundle_aliases: set[str]) -> bool:
    del bundle_aliases
    return isinstance(node, ast.Attribute) and node.attr == "select"


def _is_selector_getattr_call(node: ast.AST, aliases: dict[str, str]) -> bool:
    if not isinstance(node, ast.Call) or len(node.args) < 2:
        return False
    if not isinstance(node.func, ast.Name):
        return False
    if node.func.id != "getattr" and aliases.get(node.func.id) != "getattr":
        return False
    name = node.args[1]
    return isinstance(name, ast.Constant) and name.value == "select"


def _has_private_literal(arguments: list[ast.expr]) -> bool:
    if len(arguments) < 2 or not isinstance(arguments[1], ast.Constant):
        return False
    name = arguments[1].value
    return type(name) is str and name.startswith("_")


def _is_raw_introspection(node: ast.Attribute, bundle_aliases: set[str]) -> bool:
    if node.attr not in _INTROSPECTION_ATTRIBUTES:
        return False
    if isinstance(node.value, ast.Call):
        return (
            isinstance(node.value.func, ast.Name)
            and node.value.func.id == "type"
            and len(node.value.args) == 1
            and isinstance(node.value.args[0], ast.Name)
            and node.value.args[0].id in bundle_aliases
        )
    if isinstance(node.value, ast.Attribute):
        return isinstance(node.value.value, ast.Name) and node.value.value.id in {
            *bundle_aliases,
            "ContractSourceBundle",
            "Bundle",
        }
    return False


def _is_type_introspection_lookup(arguments: list[ast.expr], bundle_aliases: set[str]) -> bool:
    if len(arguments) < 2 or not isinstance(arguments[1], ast.Constant):
        return False
    name = arguments[1].value
    if type(name) is not str or name not in _INTROSPECTION_ATTRIBUTES:
        return False
    target = arguments[0]
    if isinstance(target, ast.Call):
        return (
            isinstance(target.func, ast.Name)
            and target.func.id == "type"
            and len(target.args) == 1
            and isinstance(target.args[0], ast.Name)
            and target.args[0].id in bundle_aliases
        )
    if isinstance(target, ast.Attribute):
        return isinstance(target.value, ast.Name) and target.value.id in {
            *bundle_aliases,
            "ContractSourceBundle",
            "Bundle",
        }
    return False
