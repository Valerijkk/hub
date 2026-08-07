#!/usr/bin/env python3
"""Deterministic, stdlib-only inventory for the Hub test rebase.

The scanner is deliberately allow-listed.  It never walks the repository root;
every input is below one of the constants in SOURCE_ROOTS or TARGET_ROOTS.
"""

from __future__ import annotations

import argparse
import fnmatch
import hashlib
import json
import os
import re
import subprocess
import sys
import tempfile
from copy import deepcopy
from pathlib import Path
from typing import Any, Iterable


TOOL_VERSION = "2.2.0"
REPO_ROOT = Path(__file__).resolve().parents[3]
OUTPUT_PATH = Path(__file__).with_name("inventory.json")
SCHEMA_PATH = Path(__file__).with_name("inventory.schema.json")

SOURCE_ROOTS = {
    "go": Path("backend/testframework"),
    "go_server": Path("backend/cmd/server"),
    "vitest": Path("testframework/stages/unit/frontend-test"),
    "playwright": Path("testframework/stages/e2e/domains"),
    "playwright_support": Path("testframework/stages/e2e"),
    "hooks": Path("backend/internal"),
}
TARGET_ROOTS = {
    "go": "backend/tests/",
    "vitest": "frontend/tests/",
    "api": "testframework/api/",
    "integration": "testframework/integration/",
    "contract": "testframework/contract/",
    "frontend_contract": "testframework/contract/frontend/",
    "playwright": "testframework/e2e/",
    "playwright_support": "testframework/e2e/support/",
    "python": "testframework/",
    "load": "testframework/load/",
}
PYTHON_COMPONENT_CLASSES = frozenset(
    {"python-api", "python-integration", "python-contract"}
)
COMPONENT_CLASS_TARGET_KEYS = {
    "python-api": "api",
    "python-integration": "integration",
    "python-contract": "contract",
}
PROTECTED_PREFIX = "pro-qa-ai/"

# These are the acceptance facts captured in evidence.md.  Source-tree drift is
# a deliberate RED result until the ledger and evidence are reviewed together.
EXPECTED = {
    "go_files": 830,
    "go_legacy_files": 829,
    "go_server_files": 1,
    "go_declared_tests": 3667,
    "go_legacy_declared_tests": 3665,
    "go_server_declared_tests": 2,
    "go_unit_files": 594,
    "go_component_files": 236,
    "vitest_files": 180,
    "frontend_unit_files": 156,
    "frontend_static_contract_files": 24,
    "frontend_support_files": 3,
    "playwright_files": 20,
    "playwright_cases": 53,
    "playwright_support_files": 14,
    "canonical_hook_files": 39,
    "target_python_tests": 0,
    "unmapped_files": 0,
}

GO_TEST_RE = re.compile(r"(?m)^func\s+((?:Test|Benchmark|Example)[A-Za-z0-9_]*)\s*\(")
TS_CASE_RE = re.compile(
    r"\b(?:it|test)(?:\.(?:only|skip|todo|fixme))?\s*\(\s*"
    r'(?:"((?:\\.|[^"\\])*)"|\'((?:\\.|[^\'\\])*)\'|`((?:\\.|[^`\\])*)`)'
)
TRACE_ID_RE = re.compile(r"\b(?:TC|FR|GZ|ECC)-[A-Za-z0-9._-]+\b")
GO_SKIP_RE = re.compile(r"\b(?:t|b|suite)\.(Skip|Skipf|SkipNow)\s*\(")
TS_SKIP_RE = re.compile(r"\b(?:it|test|describe)\.(skip|todo|skipIf|runIf)\s*\(")
HOOK_DECL_RE = re.compile(r"(?m)^(?:func|var|const|type)\s+(Export[A-Za-z0-9_]*)\b")
GO_IMPORT_RE = re.compile(
    r'(?m)^\s*(?:(?P<alias>[A-Za-z_.][A-Za-z0-9_.]*)\s+)?"(?P<path>[^"\r\n]+)"'
)
GO_SINGLE_IMPORT_RE = re.compile(
    r'(?m)^\s*import\s+(?:(?P<alias>[A-Za-z_.][A-Za-z0-9_.]*)\s+)?"(?P<path>[^"\r\n]+)"'
)
TS_IMPORT_RE = re.compile(
    r"(?:\bfrom\s+|\brequire\s*\(\s*|\bimport\s*)[\"']([^\"']+)[\"']"
)

SERVICE_PATTERNS = {
    "postgres": ("POSTGRES_TEST_URL", "postgresql://", 'enttest.Open(t, "postgres"'),
    "sqlite": ("enttest.Open", "sqlite3", "modernc.org/sqlite"),
    "mongodb": ("MONGODB_TEST_URL", "mongodb://", "mongo.Connect"),
    "redis": ("REDIS_TEST_URL", "redis.NewClient", "miniredis"),
    "nats": ("NATS_URL", "nats-server", "natsclient"),
    "object-storage": ("MINIO", "minio.New", "s3.New"),
    "http-server": ("httptest.NewServer", "net.Listen("),
    "external-process": ("exec.Command", "os/exec"),
    "browser": ("@playwright/test",),
    "dom": ("@testing-library/", "jsdom"),
}

# A Go file is component-level only when its content proves that it crosses a
# concrete process, network, or persistence boundary.  Directory names and
# historical filenames are intentionally absent from this decision.
GO_COMPONENT_MARKERS = (
    "enttest.Open",
    "POSTGRES_TEST_URL",
    "MONGODB_TEST_URL",
    "REDIS_TEST_URL",
    "testcontainers",
    "dockertest",
    "httptest.NewServer",
    "net.Listen(",
    "exec.Command",
    "nats-server",
)
GO_COMPONENT_SUPPORT_IMPORTS = (
    "/backend/testframework/env",
    "/backend/testframework/mongotest",
)
FRONTEND_STATIC_CONTRACT_MARKERS = (
    "node:fs",
    "node:path",
    "node:child_process",
)

GO_OWNER_WAVES = {
    "auth": "P133/TASK-210",
    "rbac": "P133/TASK-210",
    "audit": "P133/TASK-210",
    "ws": "P133/TASK-210",
    "calls": "P133/TASK-210",
    "platform": "P133/TASK-211",
    "apigateway": "P133/TASK-211",
    "dev": "P133/TASK-211",
    "ent": "P133/TASK-211",
    "env": "P133/TASK-211",
    "status": "P133/TASK-211",
    "tracker": "P134/TASK-212",
    "project": "P134/TASK-212",
    "sprint": "P134/TASK-212",
    "tms": "P134/TASK-212",
    "messenger": "P134/TASK-213",
    "notification": "P134/TASK-213",
    "accounting": "P135/TASK-214",
    "calendar": "P135/TASK-214",
    "intranet": "P135/TASK-214",
    "crm": "P135/TASK-214",
    "hr": "P135/TASK-214",
    "bases": "P135/TASK-214",
    "dashboard": "P135/TASK-214",
    "ai": "P135/TASK-215",
    "search": "P135/TASK-215",
    "storage": "P135/TASK-215",
    "backup": "P135/TASK-215",
    "imex": "P135/TASK-215",
    "github": "P135/TASK-215",
    "gitlab": "P135/TASK-215",
    "onec": "P135/TASK-215",
    "marketplace": "P135/TASK-215",
    "mail": "P135/TASK-215",
}
COMPONENT_OWNER_WAVES = {
    **{
        domain: "P133/TASK-227"
        for domain in (
            "auth",
            "rbac",
            "audit",
            "ws",
            "calls",
            "platform",
            "apigateway",
            "dev",
            "ent",
            "env",
            "status",
        )
    },
    **{
        domain: "P134/TASK-228"
        for domain in ("tracker", "project", "sprint", "tms", "messenger", "notification")
    },
    **{
        domain: "P135/TASK-229"
        for domain in (
            "accounting",
            "calendar",
            "intranet",
            "crm",
            "hr",
            "bases",
            "dashboard",
            "ai",
            "search",
            "storage",
            "backup",
            "imex",
            "github",
            "gitlab",
            "onec",
            "marketplace",
            "mail",
        )
    },
}
FRONTEND_OWNER_WAVES = {
    "auth": "P130/TASK-206",
    "platform": "P130/TASK-206",
    "tracker": "P131/TASK-207",
    "messenger": "P131/TASK-207",
    "tms": "P131/TASK-207",
    "calls": "P131/TASK-207",
    "admin": "P131/TASK-208",
    "ai": "P131/TASK-208",
    "dashboard": "P131/TASK-208",
    "hr": "P131/TASK-208",
    "notifications": "P131/TASK-208",
    "search": "P131/TASK-208",
    "v4": "P131/TASK-208",
    "support": "P130/TASK-206",
}
COMPLEX_E2E_SOURCES = {
    "testframework/stages/e2e/domains/ai/ai-hub.spec.ts",
    "testframework/stages/e2e/domains/messenger/messenger.spec.ts",
    "testframework/stages/e2e/domains/platform/a11y.spec.ts",
    "testframework/stages/e2e/domains/tms/tms.spec.ts",
    "testframework/stages/e2e/domains/tracker/project-lifecycle.spec.ts",
}
OWNER_WAVES = set(GO_OWNER_WAVES.values()) | set(FRONTEND_OWNER_WAVES.values()) | {
    *COMPONENT_OWNER_WAVES.values(),
    "P131/TASK-230",
    "P138/TASK-195",
    "P138/TASK-196",
    "P137/TASK-194",
}


def posix(path: Path) -> str:
    return path.as_posix()


def rel(path: Path) -> str:
    return posix(path.resolve().relative_to(REPO_ROOT.resolve()))


def sha256_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def read_allowed(path: Path, root: Path) -> tuple[bytes, str]:
    resolved = path.resolve()
    allowed = (REPO_ROOT / root).resolve()
    if not resolved.is_relative_to(allowed):
        raise ValueError(f"path escaped allow-listed root {root}: {path}")
    source_path = rel(resolved)
    if source_path.startswith(PROTECTED_PREFIX):
        raise ValueError(f"protected path rejected: {source_path}")
    data = resolved.read_bytes()
    return data, data.decode("utf-8", errors="replace")


def sorted_files(root: Path, pattern: str) -> list[Path]:
    absolute = REPO_ROOT / root
    if not absolute.exists():
        return []
    files = []
    excluded = {"node_modules", ".git", "coverage", "dist"}
    for current, directories, names in os.walk(absolute, followlinks=False):
        directories[:] = sorted(directory for directory in directories if directory not in excluded)
        for name in sorted(names):
            if not fnmatch.fnmatch(name, pattern):
                continue
            candidate = Path(current) / name
            # Resolve every candidate before accepting it so a symlink cannot
            # turn an allow-listed walk into a protected/external read.
            read_allowed(candidate, root)
            files.append(candidate.resolve())
    return sorted(files, key=rel)


def revision() -> str:
    result = subprocess.run(
        ["git", "rev-parse", "HEAD"],
        cwd=REPO_ROOT,
        check=True,
        capture_output=True,
        text=True,
    )
    value = result.stdout.strip()
    if not re.fullmatch(r"[0-9a-f]{40}", value):
        raise ValueError(f"unexpected git revision: {value!r}")
    return value


def revision_is_ancestor(value: str) -> bool:
    if not re.fullmatch(r"[0-9a-f]{40}", value):
        return False
    result = subprocess.run(
        ["git", "merge-base", "--is-ancestor", value, "HEAD"],
        cwd=REPO_ROOT,
        check=False,
        capture_output=True,
        text=True,
    )
    return result.returncode == 0


def line_sites(text: str, pattern: re.Pattern[str]) -> list[str]:
    return [f"L{text.count(chr(10), 0, match.start()) + 1}:{match.group(1)}" for match in pattern.finditer(text)]


def ts_cases(text: str) -> list[str]:
    cases = []
    for match in TS_CASE_RE.finditer(text):
        title = next(group for group in match.groups() if group is not None)
        cases.append(title.replace("\\`", "`").replace("\\\"", '"').replace("\\'", "'"))
    return cases


def go_import_pairs(text: str) -> list[tuple[str | None, str]]:
    pairs = set()
    for pattern in (GO_IMPORT_RE, GO_SINGLE_IMPORT_RE):
        for match in pattern.finditer(text):
            alias = match.group("alias")
            # GO_IMPORT_RE sees the keyword in `import "path"` as an alias;
            # GO_SINGLE_IMPORT_RE supplies the canonical pair for that form.
            if alias == "import":
                continue
            pairs.add((alias, match.group("path")))
    return sorted(pairs, key=lambda item: ((item[0] or ""), item[1]))


def external_go_imports(text: str) -> list[str]:
    imports = []
    for _, value in go_import_pairs(text):
        if "." in value.split("/", 1)[0]:
            imports.append(value)
    return sorted(set(imports))


def external_ts_imports(text: str) -> list[str]:
    imports = []
    for value in TS_IMPORT_RE.findall(text):
        if value.startswith((".", "/", "@/")):
            continue
        if value.startswith("@"):
            parts = value.split("/")
            imports.append("/".join(parts[:2]))
        else:
            imports.append(value.split("/", 1)[0])
    return sorted(set(imports))


def service_hints(text: str) -> list[str]:
    return sorted(
        service
        for service, markers in SERVICE_PATTERNS.items()
        if any(marker in text for marker in markers)
    )


def domain_for(relative: Path, runner: str) -> str:
    parts = relative.parts
    if runner in {"vitest", "vitest-support"}:
        if parts and parts[0] == "domains":
            return parts[1] if len(parts) > 1 else "unknown"
        return "support"
    if runner == "playwright" and parts:
        return parts[1] if parts[0] == "domains" and len(parts) > 1 else parts[0]
    if runner == "go" and parts:
        return parts[0]
    return "unknown"


def go_target(relative: Path) -> str:
    return posix(Path(TARGET_ROOTS["go"]) / relative)


def is_go_component(text: str, dependencies: list[str]) -> bool:
    return any(marker in text for marker in GO_COMPONENT_MARKERS) or any(
        marker in dependency
        for dependency in dependencies
        for marker in GO_COMPONENT_SUPPORT_IMPORTS
    )


def go_component_target_class(services: list[str], dependencies: list[str]) -> str:
    if "http-server" in services:
        return "python-api"
    integration_services = {
        "postgres",
        "sqlite",
        "mongodb",
        "redis",
        "nats",
        "object-storage",
        "external-process",
    }
    integration_dependencies = ("testcontainers", "dockertest") + GO_COMPONENT_SUPPORT_IMPORTS
    if integration_services.intersection(services) or any(
        marker in dependency
        for dependency in dependencies
        for marker in integration_dependencies
    ):
        return "python-integration"
    # A component marker without a resolved runtime dependency still needs an
    # explicit Python parity destination; it is not silently treated as Go unit.
    return "python-contract"


def go_component_target(relative: Path, target_class: str) -> str:
    target_key = COMPONENT_CLASS_TARGET_KEYS.get(target_class)
    if target_key is None:
        raise ValueError(f"unsupported Go component target class: {target_class}")
    parts = list(relative.parts)
    filename = parts[-1]
    stem = filename[: -len("_test.go")] if filename.endswith("_test.go") else Path(filename).stem
    parts[-1] = f"test_{stem.replace('-', '_')}.py"
    return posix(Path(TARGET_ROOTS[target_key]).joinpath(*parts))


def vitest_target(relative: Path) -> str:
    parts = list(relative.parts)
    if parts and parts[0] == "domains":
        parts = parts[1:]
    return posix(Path(TARGET_ROOTS["vitest"]).joinpath(*parts))


def frontend_support_target(relative: Path) -> str:
    parts = list(relative.parts)
    if parts and parts[0] == "domains":
        parts = parts[1:]
    return posix(Path(TARGET_ROOTS["vitest"]).joinpath("support", *parts))


def python_test_filename(filename: str) -> str:
    suffixes = (".test.tsx", ".spec.tsx", ".test.ts", ".spec.ts", "_test.go")
    stem = next(
        (filename[: -len(suffix)] for suffix in suffixes if filename.endswith(suffix)),
        Path(filename).stem,
    )
    normalized = re.sub(r"[^A-Za-z0-9_]+", "_", stem).strip("_")
    return f"test_{normalized}.py"


def frontend_static_contract_target(relative: Path) -> str:
    parts = list(relative.parts)
    if parts and parts[0] == "domains":
        parts = parts[1:]
    parts[-1] = python_test_filename(parts[-1])
    return posix(Path(TARGET_ROOTS["frontend_contract"]).joinpath(*parts))


def e2e_target(relative: Path) -> str:
    parts = list(relative.parts)
    if parts and parts[0] == "domains":
        parts = parts[1:]
    parts[-1] = python_test_filename(parts[-1])
    return posix(Path(TARGET_ROOTS["playwright"]).joinpath(*parts))


def e2e_support_target(relative: Path) -> str:
    parts = list(relative.parts)
    filename = parts[-1]
    stem = filename[: -len(".tsx")] if filename.endswith(".tsx") else filename[: -len(".ts")]
    parts[-1] = f"{re.sub(r'[^A-Za-z0-9_]+', '_', stem).strip('_')}.py"
    return posix(Path(TARGET_ROOTS["playwright_support"]).joinpath(*parts))


def replacement_status(target_path: str) -> str:
    return "candidate-present" if (REPO_ROOT / target_path).is_file() else "not-started"


def canonical_owner_wave(
    source_path: str, runner: str, domain: str, target_class: str
) -> str:
    if runner == "go-test":
        wave = (
            COMPONENT_OWNER_WAVES.get(domain)
            if target_class in PYTHON_COMPONENT_CLASSES
            else GO_OWNER_WAVES.get(domain)
        )
    elif runner in {"vitest", "vitest-support"}:
        wave = (
            "P131/TASK-230"
            if target_class == "python-static-contract"
            else FRONTEND_OWNER_WAVES.get(domain)
        )
    elif runner == "playwright-ts":
        wave = "P138/TASK-196" if source_path in COMPLEX_E2E_SOURCES else "P138/TASK-195"
    elif runner == "playwright-support-ts":
        wave = "P137/TASK-194"
    else:
        wave = None
    if wave is None:
        raise ValueError(
            f"no owner wave for {runner}/{domain}/{target_class}: {source_path}"
        )
    return wave


def base_row(
    *,
    source_path: str,
    source_hash: str,
    runner: str,
    domain: str,
    test_ids: list[str],
    skip_sites: list[str],
    dependencies: list[str],
    services: list[str],
    target_class: str,
    target_path: str,
    coverage_bearing: bool,
) -> dict[str, Any]:
    return {
        "source_path": source_path,
        "sha256": source_hash,
        "runner": runner,
        "domain": domain,
        "owner_wave": canonical_owner_wave(source_path, runner, domain, target_class),
        "test_ids": test_ids,
        "test_count": len(test_ids),
        "skips": {"count": len(skip_sites), "sites": skip_sites},
        "service_hints": services,
        "dependency_hints": dependencies,
        "target_class": target_class,
        "target_path": target_path,
        "coverage_bearing": coverage_bearing,
        "hook_consumers": [],
        "baseline_status": "not-executed-in-task-202",
        "replacement_status": replacement_status(target_path),
        "parity_status": "not-assessed",
        "disposition": "retain-until-parity",
        "evidence": [f"sha256:{source_hash}", f"declared-tests:{len(test_ids)}"],
        "trace_ids": sorted(set(TRACE_ID_RE.findall("\n".join(test_ids)))),
    }


def go_imports_by_module(text: str, package_names: dict[str, str]) -> dict[str, str]:
    result = {}
    marker = "/internal/"
    for alias, import_path in go_import_pairs(text):
        if marker not in import_path:
            continue
        module = import_path.split(marker, 1)[1]
        if module not in package_names:
            continue
        if alias == "_":
            continue
        result[module] = alias if alias and alias != "." else ("." if alias == "." else package_names[module])
    return result


def hook_inventory(go_sources: list[tuple[Path, str]]) -> tuple[list[dict[str, Any]], dict[str, list[dict[str, str]]]]:
    hook_paths = sorted_files(SOURCE_ROOTS["hooks"], "export_hooks*.go")
    raw_hooks = []
    package_names: dict[str, str] = {}
    for path in hook_paths:
        data, text = read_allowed(path, SOURCE_ROOTS["hooks"])
        relative = path.relative_to((REPO_ROOT / SOURCE_ROOTS["hooks"]).resolve())
        module = posix(relative.parent)
        package_match = re.search(r"(?m)^package\s+([A-Za-z0-9_]+)\s*$", text)
        package_name = package_match.group(1) if package_match else relative.parent.name
        package_names[module] = package_name
        raw_hooks.append((path, data, text, module, package_name))

    consumers_by_test: dict[str, list[dict[str, str]]] = {rel(path): [] for path, _ in go_sources}
    records = []
    records_by_module: dict[str, list[dict[str, Any]]] = {}
    for path, data, text, module, package_name in raw_hooks:
        symbols = sorted(set(HOOK_DECL_RE.findall(text)))
        record = {
            "source_path": rel(path),
            "sha256": sha256_bytes(data),
            "module": module,
            "package": package_name,
            "canonical": path.name == "export_hooks.go",
            "symbols": [{"name": symbol, "consumer_test_files": []} for symbol in symbols],
            "deletion_decision": "none-in-task-202",
        }
        records.append(record)
        records_by_module.setdefault(module, []).append(record)

    # Resolve only modules actually imported by each test.  This is both more
    # precise and dramatically cheaper than comparing every exported symbol to
    # every test file.
    for test_path, test_text in go_sources:
        test_rel = rel(test_path)
        for module, alias in go_imports_by_module(test_text, package_names).items():
            for record in records_by_module.get(module, []):
                for symbol_record in record["symbols"]:
                    symbol = symbol_record["name"]
                    needle = symbol if alias == "." else f"{alias}.{symbol}"
                    if needle not in test_text:
                        continue
                    usage = rf"(?<![A-Za-z0-9_]){re.escape(symbol)}\b" if alias == "." else rf"\b{re.escape(alias)}\.{re.escape(symbol)}\b"
                    if re.search(usage, test_text):
                        symbol_record["consumer_test_files"].append(test_rel)
                        consumers_by_test[test_rel].append(
                            {"hook_source": record["source_path"], "symbol": symbol}
                        )
    for record in records:
        for symbol_record in record["symbols"]:
            symbol_record["consumer_test_files"].sort()
    for links in consumers_by_test.values():
        links.sort(key=lambda item: (item["hook_source"], item["symbol"]))
    return records, consumers_by_test


def target_python_test_count() -> int:
    return sum(
        candidate.name.startswith("test_") or candidate.name.endswith("_test.py")
        for candidate in sorted_files(Path(TARGET_ROOTS["python"]), "*.py")
    )


def discover_frontend_support_paths() -> list[Path]:
    return sorted(
        {
            path
            for pattern in ("render.tsx", "*.fixture.ts", "*.fixture.tsx")
            for path in sorted_files(SOURCE_ROOTS["vitest"], pattern)
        },
        key=rel,
    )


def discover_e2e_support_paths() -> list[Path]:
    e2e_domain_root = (REPO_ROOT / SOURCE_ROOTS["playwright"]).resolve()
    return sorted(
        {
            path
            for pattern in ("*.ts", "*.tsx")
            for path in sorted_files(SOURCE_ROOTS["playwright_support"], pattern)
            if not path.is_relative_to(e2e_domain_root)
        },
        key=rel,
    )


def build_inventory() -> dict[str, Any]:
    go_paths = sorted_files(SOURCE_ROOTS["go"], "*_test.go")
    go_server_paths = sorted_files(SOURCE_ROOTS["go_server"], "*_test.go")
    vitest_paths = sorted(
        {
            path
            for pattern in ("*.test.ts", "*.test.tsx", "*.spec.ts", "*.spec.tsx")
            for path in sorted_files(SOURCE_ROOTS["vitest"], pattern)
        },
        key=rel,
    )
    frontend_support_paths = discover_frontend_support_paths()
    e2e_paths = sorted_files(SOURCE_ROOTS["playwright"], "*.spec.ts")
    e2e_support_paths = discover_e2e_support_paths()

    go_sources = []
    for path in go_paths:
        _, text = read_allowed(path, SOURCE_ROOTS["go"])
        relative = path.relative_to((REPO_ROOT / SOURCE_ROOTS["go"]).resolve())
        go_sources.append(
            (path, text, SOURCE_ROOTS["go"], relative, domain_for(relative, "go"), relative)
        )
    for path in go_server_paths:
        _, text = read_allowed(path, SOURCE_ROOTS["go_server"])
        relative = path.relative_to((REPO_ROOT / SOURCE_ROOTS["go_server"]).resolve())
        go_sources.append(
            (
                path,
                text,
                SOURCE_ROOTS["go_server"],
                relative,
                "platform",
                Path("server") / relative,
            )
        )
    go_sources.sort(key=lambda item: rel(item[0]))
    hooks, consumers_by_test = hook_inventory(
        [(path, text) for path, text, _, _, _, _ in go_sources]
    )

    rows = []
    for path, text, source_root, relative, domain, target_relative in go_sources:
        data, _ = read_allowed(path, source_root)
        dependencies = external_go_imports(text)
        component = is_go_component(text, dependencies)
        services = service_hints(text)
        if component:
            target_class = go_component_target_class(services, dependencies)
            target_path = go_component_target(target_relative, target_class)
        else:
            target_class = "go-unit"
            target_path = go_target(target_relative)
        row = base_row(
            source_path=rel(path),
            source_hash=sha256_bytes(data),
            runner="go-test",
            domain=domain,
            test_ids=GO_TEST_RE.findall(text),
            skip_sites=line_sites(text, GO_SKIP_RE),
            dependencies=dependencies,
            services=services,
            target_class=target_class,
            target_path=target_path,
            coverage_bearing=True,
        )
        if component:
            mapping_hints = services or dependencies or ["component-marker-fallback"]
            row["evidence"].append(
                f"conversion-target:{target_class};hints:{','.join(mapping_hints)}"
            )
        else:
            row["evidence"].append("target:go-unit")
        row["hook_consumers"] = consumers_by_test[row["source_path"]]
        row["trace_ids"] = sorted(set(TRACE_ID_RE.findall(text)))
        rows.append(row)

    for path in vitest_paths:
        data, text = read_allowed(path, SOURCE_ROOTS["vitest"])
        relative = path.relative_to((REPO_ROOT / SOURCE_ROOTS["vitest"]).resolve())
        static_contract = any(marker in text for marker in FRONTEND_STATIC_CONTRACT_MARKERS)
        target_class = "python-static-contract" if static_contract else "frontend-unit"
        target_path = (
            frontend_static_contract_target(relative)
            if static_contract
            else vitest_target(relative)
        )
        row = base_row(
            source_path=rel(path),
            source_hash=sha256_bytes(data),
            runner="vitest",
            domain=domain_for(relative, "vitest"),
            test_ids=ts_cases(text),
            skip_sites=line_sites(text, TS_SKIP_RE),
            dependencies=external_ts_imports(text),
            services=service_hints(text),
            target_class=target_class,
            target_path=target_path,
            coverage_bearing=not static_contract,
        )
        row["evidence"].append(
            "conversion-target:python-static-contract;hints:node-fs-path-process"
            if static_contract
            else "target:frontend-unit"
        )
        row["trace_ids"] = sorted(set(TRACE_ID_RE.findall(text)))
        rows.append(row)

    for path in frontend_support_paths:
        data, text = read_allowed(path, SOURCE_ROOTS["vitest"])
        relative = path.relative_to((REPO_ROOT / SOURCE_ROOTS["vitest"]).resolve())
        row = base_row(
            source_path=rel(path),
            source_hash=sha256_bytes(data),
            runner="vitest-support",
            domain=domain_for(relative, "vitest-support"),
            test_ids=[],
            skip_sites=[],
            dependencies=external_ts_imports(text),
            services=service_hints(text),
            target_class="frontend-support",
            target_path=frontend_support_target(relative),
            coverage_bearing=False,
        )
        row["evidence"].append("target:frontend-support")
        row["trace_ids"] = sorted(set(TRACE_ID_RE.findall(text)))
        rows.append(row)

    for path in e2e_paths:
        data, text = read_allowed(path, SOURCE_ROOTS["playwright"])
        relative = path.relative_to((REPO_ROOT / SOURCE_ROOTS["playwright"]).resolve())
        row = base_row(
            source_path=rel(path),
            source_hash=sha256_bytes(data),
            runner="playwright-ts",
            domain=domain_for(relative, "playwright"),
            test_ids=ts_cases(text),
            skip_sites=line_sites(text, TS_SKIP_RE),
            dependencies=external_ts_imports(text),
            services=service_hints(text),
            target_class="python-e2e",
            target_path=e2e_target(relative),
            coverage_bearing=False,
        )
        row["trace_ids"] = sorted(set(TRACE_ID_RE.findall(text)))
        rows.append(row)

    for path in e2e_support_paths:
        data, text = read_allowed(path, SOURCE_ROOTS["playwright_support"])
        relative = path.relative_to(
            (REPO_ROOT / SOURCE_ROOTS["playwright_support"]).resolve()
        )
        row = base_row(
            source_path=rel(path),
            source_hash=sha256_bytes(data),
            runner="playwright-support-ts",
            domain="support",
            test_ids=[],
            skip_sites=[],
            dependencies=external_ts_imports(text),
            services=service_hints(text),
            target_class="python-e2e-support",
            target_path=e2e_support_target(relative),
            coverage_bearing=False,
        )
        row["evidence"].append("conversion-target:python-e2e-support")
        row["trace_ids"] = sorted(set(TRACE_ID_RE.findall(text)))
        rows.append(row)

    rows.sort(key=lambda item: item["source_path"])
    counts = {
        "go_files": len(go_paths) + len(go_server_paths),
        "go_legacy_files": len(go_paths),
        "go_server_files": len(go_server_paths),
        "go_declared_tests": sum(row["test_count"] for row in rows if row["runner"] == "go-test"),
        "go_legacy_declared_tests": sum(
            row["test_count"]
            for row in rows
            if str(row["source_path"]).startswith("backend/testframework/")
        ),
        "go_server_declared_tests": sum(
            row["test_count"]
            for row in rows
            if str(row["source_path"]).startswith("backend/cmd/server/")
        ),
        "go_unit_files": sum(row["target_class"] == "go-unit" for row in rows),
        "go_component_files": sum(row["target_class"] in PYTHON_COMPONENT_CLASSES for row in rows),
        "go_component_api_files": sum(row["target_class"] == "python-api" for row in rows),
        "go_component_integration_files": sum(
            row["target_class"] == "python-integration" for row in rows
        ),
        "go_component_contract_files": sum(
            row["target_class"] == "python-contract" for row in rows
        ),
        "vitest_files": len(vitest_paths),
        "vitest_declared_tests": sum(row["test_count"] for row in rows if row["runner"] == "vitest"),
        "frontend_unit_files": sum(row["target_class"] == "frontend-unit" for row in rows),
        "frontend_static_contract_files": sum(
            row["target_class"] == "python-static-contract" for row in rows
        ),
        "frontend_support_files": len(frontend_support_paths),
        "playwright_files": len(e2e_paths),
        "playwright_cases": sum(row["test_count"] for row in rows if row["runner"] == "playwright-ts"),
        "playwright_skip_sites": sum(row["skips"]["count"] for row in rows if row["runner"] == "playwright-ts"),
        "playwright_support_files": len(e2e_support_paths),
        "canonical_hook_files": sum(item["canonical"] for item in hooks),
        "hook_fragment_files": sum(not item["canonical"] for item in hooks),
        "hook_files_total": len(hooks),
        "hook_symbols": sum(len(item["symbols"]) for item in hooks),
        "hook_symbols_with_consumers": sum(
            bool(symbol["consumer_test_files"]) for item in hooks for symbol in item["symbols"]
        ),
        "target_python_tests": target_python_test_count(),
        "unmapped_files": sum(not row.get("target_path") for row in rows),
    }
    return {
        "schema_version": 2,
        "tool_version": TOOL_VERSION,
        "revision": revision(),
        "roots": {
            "sources": {key: posix(value) for key, value in SOURCE_ROOTS.items()},
            "targets": TARGET_ROOTS,
            "protected_exclusion": PROTECTED_PREFIX,
        },
        "expected": EXPECTED,
        "counts": counts,
        "rows": rows,
        "hooks": hooks,
    }


def strings(value: Any) -> Iterable[str]:
    if isinstance(value, str):
        yield value
    elif isinstance(value, dict):
        for item in value.values():
            yield from strings(item)
    elif isinstance(value, list):
        for item in value:
            yield from strings(item)


SCHEMA_KEYWORDS = {
    "$schema",
    "$id",
    "$defs",
    "$ref",
    "title",
    "type",
    "required",
    "properties",
    "additionalProperties",
    "items",
    "enum",
    "const",
    "pattern",
    "minimum",
}


def resolve_local_ref(root_schema: dict[str, Any], reference: str) -> dict[str, Any]:
    if not reference.startswith("#/"):
        raise ValueError(f"only local schema refs are supported: {reference}")
    current: Any = root_schema
    for raw_part in reference[2:].split("/"):
        part = raw_part.replace("~1", "/").replace("~0", "~")
        if not isinstance(current, dict) or part not in current:
            raise ValueError(f"unresolved schema ref: {reference}")
        current = current[part]
    if not isinstance(current, dict):
        raise ValueError(f"schema ref does not resolve to an object: {reference}")
    return current


def schema_contract_errors(
    value: Any,
    schema: dict[str, Any],
    *,
    root_schema: dict[str, Any],
    path: str = "$",
) -> list[str]:
    """Validate the explicit JSON-Schema subset used by inventory.schema.json.

    This is intentionally not advertised as a general JSON Schema engine.  It
    rejects unsupported assertion keywords so the check cannot silently pretend
    to enforce a contract it does not understand.
    """

    errors = []
    unsupported = sorted(set(schema) - SCHEMA_KEYWORDS)
    if unsupported:
        errors.append(f"{path}: unsupported schema keywords {unsupported}")
        return errors
    if "$ref" in schema:
        try:
            referenced = resolve_local_ref(root_schema, schema["$ref"])
        except ValueError as error:
            return [f"{path}: {error}"]
        return schema_contract_errors(value, referenced, root_schema=root_schema, path=path)

    expected_type = schema.get("type")
    type_checks = {
        "object": lambda item: isinstance(item, dict),
        "array": lambda item: isinstance(item, list),
        "string": lambda item: isinstance(item, str),
        "integer": lambda item: isinstance(item, int) and not isinstance(item, bool),
        "boolean": lambda item: isinstance(item, bool),
    }
    if expected_type is not None:
        checker = type_checks.get(expected_type)
        if checker is None:
            errors.append(f"{path}: unsupported schema type {expected_type!r}")
            return errors
        if not checker(value):
            errors.append(f"{path}: expected {expected_type}, got {type(value).__name__}")
            return errors

    if "const" in schema and value != schema["const"]:
        errors.append(f"{path}: expected const {schema['const']!r}, got {value!r}")
    if "enum" in schema and value not in schema["enum"]:
        errors.append(f"{path}: value {value!r} is not in enum {schema['enum']!r}")
    if "pattern" in schema and isinstance(value, str) and re.search(schema["pattern"], value) is None:
        errors.append(f"{path}: string does not match pattern {schema['pattern']!r}")
    if "minimum" in schema and isinstance(value, int) and value < schema["minimum"]:
        errors.append(f"{path}: value {value} is below minimum {schema['minimum']}")

    if isinstance(value, dict):
        required = schema.get("required", [])
        for key in required:
            if key not in value:
                errors.append(f"{path}: missing required property {key!r}")
        properties = schema.get("properties", {})
        for key, child_schema in properties.items():
            if key in value:
                errors.extend(
                    schema_contract_errors(
                        value[key], child_schema, root_schema=root_schema, path=f"{path}.{key}"
                    )
                )
        if schema.get("additionalProperties") is False:
            for key in sorted(set(value) - set(properties)):
                errors.append(f"{path}: additional property {key!r} is forbidden")
    if isinstance(value, list) and "items" in schema:
        for index, item in enumerate(value):
            errors.extend(
                schema_contract_errors(
                    item, schema["items"], root_schema=root_schema, path=f"{path}[{index}]"
                )
            )
    return errors


def validate_inventory(
    data: dict[str, Any],
    *,
    live_source_paths: set[str] | None = None,
    enforce_expected: bool = True,
) -> list[str]:
    errors = []
    rows = data.get("rows", [])
    source_paths = [row.get("source_path") for row in rows]
    unique_paths = set(source_paths)
    if len(unique_paths) != len(source_paths):
        duplicates = sorted(path for path in unique_paths if source_paths.count(path) > 1)
        errors.append(f"duplicate source rows: {duplicates}")
    if live_source_paths is not None:
        missing = sorted(live_source_paths - unique_paths)
        extra = sorted(unique_paths - live_source_paths)
        if missing:
            errors.append(f"missing source rows: {missing}")
        if extra:
            errors.append(f"unknown source rows: {extra}")
    target_paths = [row.get("target_path") for row in rows]
    duplicate_targets = sorted(path for path in set(target_paths) if target_paths.count(path) > 1)
    if duplicate_targets:
        errors.append(f"duplicate target paths: {duplicate_targets}")

    allowed = {
        "go-test": {
            "go-unit": TARGET_ROOTS["go"],
            "python-api": TARGET_ROOTS["api"],
            "python-integration": TARGET_ROOTS["integration"],
            "python-contract": TARGET_ROOTS["contract"],
        },
        "vitest": {
            "frontend-unit": TARGET_ROOTS["vitest"],
            "python-static-contract": TARGET_ROOTS["frontend_contract"],
        },
        "vitest-support": {
            "frontend-support": f"{TARGET_ROOTS['vitest']}support/",
        },
        "playwright-ts": {"python-e2e": TARGET_ROOTS["playwright"]},
        "playwright-support-ts": {
            "python-e2e-support": TARGET_ROOTS["playwright_support"],
        },
    }
    for row in rows:
        runner = row.get("runner")
        if runner not in allowed:
            errors.append(f"bad runner for {row.get('source_path')}: {runner}")
            continue
        classes = allowed[runner]
        target_class = row.get("target_class")
        if target_class not in classes:
            errors.append(f"bad target class for {row.get('source_path')}: {row.get('target_class')}")
            prefix = ""
        else:
            prefix = classes[target_class]
        try:
            expected_owner = canonical_owner_wave(
                str(row.get("source_path")),
                str(runner),
                str(row.get("domain")),
                str(target_class),
            )
        except ValueError as error:
            errors.append(str(error))
        else:
            if row.get("owner_wave") != expected_owner or row.get("owner_wave") not in OWNER_WAVES:
                errors.append(
                    f"bad owner wave for {row.get('source_path')}: "
                    f"expected={expected_owner} actual={row.get('owner_wave')}"
                )
        if not prefix or not str(row.get("target_path", "")).startswith(prefix):
            errors.append(f"bad target root for {row.get('source_path')}: {row.get('target_path')}")
        target_path = str(row.get("target_path", ""))
        python_classes = PYTHON_COMPONENT_CLASSES | {
            "python-static-contract",
            "python-e2e",
            "python-e2e-support",
        }
        if target_class in python_classes and not target_path.endswith(".py"):
            errors.append(f"non-Python target for {row.get('source_path')}: {row.get('target_path')}")
        if target_class == "go-unit" and not target_path.endswith("_test.go"):
            errors.append(f"non-Go unit target for {row.get('source_path')}: {row.get('target_path')}")
        if target_class in {"frontend-unit", "frontend-support"} and not target_path.endswith(
            (".ts", ".tsx")
        ):
            errors.append(f"non-TypeScript frontend target for {row.get('source_path')}: {target_path}")
        expected_coverage = runner == "go-test" or target_class == "frontend-unit"
        if row.get("coverage_bearing") is not expected_coverage:
            errors.append(f"bad coverage-bearing flag for {row.get('source_path')}")
        if row.get("test_count") != len(row.get("test_ids", [])):
            errors.append(f"test count mismatch for {row.get('source_path')}")
        skips = row.get("skips", {})
        if skips.get("count") != len(skips.get("sites", [])):
            errors.append(f"skip count mismatch for {row.get('source_path')}")
        if row.get("disposition") != "retain-until-parity":
            errors.append(f"unsafe disposition for {row.get('source_path')}: {row.get('disposition')}")
        if row.get("parity_status") != "not-assessed":
            errors.append(f"unproven parity status for {row.get('source_path')}: {row.get('parity_status')}")
        if row.get("replacement_status") not in {"not-started", "candidate-present"}:
            errors.append(
                f"bad replacement status for {row.get('source_path')}: {row.get('replacement_status')}"
            )
        if not row.get("evidence"):
            errors.append(f"missing evidence for {row.get('source_path')}")

    counts = data.get("counts", {})
    calculated = {
        "go_files": sum(row.get("runner") == "go-test" for row in rows),
        "go_legacy_files": sum(
            str(row.get("source_path", "")).startswith("backend/testframework/")
            for row in rows
        ),
        "go_server_files": sum(
            str(row.get("source_path", "")).startswith("backend/cmd/server/")
            for row in rows
        ),
        "go_declared_tests": sum(
            row.get("test_count", 0) for row in rows if row.get("runner") == "go-test"
        ),
        "go_legacy_declared_tests": sum(
            row.get("test_count", 0)
            for row in rows
            if str(row.get("source_path", "")).startswith("backend/testframework/")
        ),
        "go_server_declared_tests": sum(
            row.get("test_count", 0)
            for row in rows
            if str(row.get("source_path", "")).startswith("backend/cmd/server/")
        ),
        "go_unit_files": sum(row.get("target_class") == "go-unit" for row in rows),
        "go_component_files": sum(
            row.get("target_class") in PYTHON_COMPONENT_CLASSES for row in rows
        ),
        "go_component_api_files": sum(row.get("target_class") == "python-api" for row in rows),
        "go_component_integration_files": sum(
            row.get("target_class") == "python-integration" for row in rows
        ),
        "go_component_contract_files": sum(
            row.get("target_class") == "python-contract" for row in rows
        ),
        "vitest_files": sum(row.get("runner") == "vitest" for row in rows),
        "frontend_unit_files": sum(row.get("target_class") == "frontend-unit" for row in rows),
        "frontend_static_contract_files": sum(
            row.get("target_class") == "python-static-contract" for row in rows
        ),
        "frontend_support_files": sum(row.get("runner") == "vitest-support" for row in rows),
        "playwright_files": sum(row.get("runner") == "playwright-ts" for row in rows),
        "playwright_cases": sum(row.get("test_count", 0) for row in rows if row.get("runner") == "playwright-ts"),
        "playwright_support_files": sum(
            row.get("runner") == "playwright-support-ts" for row in rows
        ),
        "canonical_hook_files": sum(hook.get("canonical") is True for hook in data.get("hooks", [])),
        "unmapped_files": sum(not row.get("target_path") for row in rows),
    }
    for key, value in calculated.items():
        if counts.get(key) != value:
            errors.append(f"count reconciliation failed for {key}: metadata={counts.get(key)} rows={value}")
    if enforce_expected and (counts.get("playwright_files") != 20 or counts.get("playwright_cases") != 53):
        errors.append(
            "Playwright reconciliation failed: "
            f"files={counts.get('playwright_files')} declared_cases={counts.get('playwright_cases')}"
        )
    if enforce_expected:
        if data.get("tool_version") != TOOL_VERSION:
            errors.append(f"tool version mismatch: ledger={data.get('tool_version')} current={TOOL_VERSION}")
        ledger_revision = str(data.get("revision", ""))
        if not revision_is_ancestor(ledger_revision):
            errors.append(
                "ledger revision is not an ancestor of current HEAD: "
                f"ledger={ledger_revision} current={revision()}"
            )
        for key, value in EXPECTED.items():
            if counts.get(key) != value:
                errors.append(f"expected baseline drift for {key}: expected={value} actual={counts.get(key)}")

    row_lookup = {row.get("source_path"): row for row in rows}
    hooks = data.get("hooks", [])
    hook_sources = [hook.get("source_path") for hook in hooks]
    duplicate_hooks = sorted(path for path in set(hook_sources) if hook_sources.count(path) > 1)
    if duplicate_hooks:
        errors.append(f"duplicate hook files: {duplicate_hooks}")
    forward_links = set()
    for hook in hooks:
        hook_source = hook.get("source_path")
        if not str(hook_source).startswith("backend/internal/") or not str(hook_source).split("/")[-1].startswith("export_hooks"):
            errors.append(f"bad hook source path: {hook_source}")
        for symbol in hook.get("symbols", []):
            name = symbol.get("name")
            for consumer in symbol.get("consumer_test_files", []):
                forward_links.add((hook_source, name, consumer))
                row = row_lookup.get(consumer)
                if row is None or row.get("runner") != "go-test":
                    errors.append(f"bad hook consumer link {hook_source}:{name} -> {consumer}")
                    continue
                expected_link = {"hook_source": hook_source, "symbol": name}
                if expected_link not in row.get("hook_consumers", []):
                    errors.append(f"missing reverse hook link {hook_source}:{name} -> {consumer}")
    for consumer, row in row_lookup.items():
        for link in row.get("hook_consumers", []):
            candidate = (link.get("hook_source"), link.get("symbol"), consumer)
            if candidate not in forward_links:
                errors.append(
                    "orphan reverse hook link "
                    f"{link.get('hook_source')}:{link.get('symbol')} <- {consumer}"
                )

    for value in strings(data):
        normalized = value.replace("\\", "/").lstrip("./")
        if normalized.startswith(PROTECTED_PREFIX) and value != PROTECTED_PREFIX:
            errors.append(f"protected path leaked into ledger: {value}")
    return errors


def render(data: dict[str, Any]) -> bytes:
    return (json.dumps(data, ensure_ascii=False, indent=2, sort_keys=True) + "\n").encode("utf-8")


def live_test_paths() -> set[str]:
    paths = {rel(path) for path in sorted_files(SOURCE_ROOTS["go"], "*_test.go")}
    paths.update(rel(path) for path in sorted_files(SOURCE_ROOTS["go_server"], "*_test.go"))
    for pattern in ("*.test.ts", "*.test.tsx", "*.spec.ts", "*.spec.tsx"):
        paths.update(rel(path) for path in sorted_files(SOURCE_ROOTS["vitest"], pattern))
    paths.update(rel(path) for path in discover_frontend_support_paths())
    paths.update(rel(path) for path in sorted_files(SOURCE_ROOTS["playwright"], "*.spec.ts"))
    paths.update(rel(path) for path in discover_e2e_support_paths())
    return paths


def run_check() -> int:
    if not OUTPUT_PATH.is_file():
        print(f"RED: missing generated ledger {rel(OUTPUT_PATH)}", file=sys.stderr)
        return 1
    if not SCHEMA_PATH.is_file():
        print(f"RED: missing schema contract {rel(SCHEMA_PATH)}", file=sys.stderr)
        return 1
    on_disk = json.loads(OUTPUT_PATH.read_text(encoding="utf-8"))
    schema = json.loads(SCHEMA_PATH.read_text(encoding="utf-8"))
    errors = [
        f"schema contract: {error}"
        for error in schema_contract_errors(on_disk, schema, root_schema=schema)
    ]
    errors.extend(validate_inventory(on_disk, live_source_paths=live_test_paths()))
    expected = build_inventory()
    # `revision` is provenance for the source snapshot, not a cachebuster for
    # every later planning-only commit. validate_inventory already proves that
    # it is an ancestor; compare all live inputs against that recorded value.
    expected["revision"] = on_disk.get("revision")
    expected_bytes = render(expected)
    actual_bytes = OUTPUT_PATH.read_bytes()
    if actual_bytes != expected_bytes:
        errors.append("generated ledger differs from current allow-listed inputs")
    if errors:
        for error in errors:
            print(f"RED: {error}", file=sys.stderr)
        return 1
    print(
        "GREEN: ledger matches allow-listed inputs; "
        f"sha256={sha256_bytes(actual_bytes)} rows={len(on_disk['rows'])}"
    )
    return 0


def sentinel_row(source_path: str) -> dict[str, Any]:
    return {
        "source_path": source_path,
        "sha256": "0" * 64,
        "runner": "go-test",
        "domain": "auth",
        "owner_wave": "P133/TASK-210",
        "test_ids": ["TestSentinel"],
        "test_count": 1,
        "skips": {"count": 0, "sites": []},
        "service_hints": [],
        "dependency_hints": [],
        "target_class": "go-unit",
        "target_path": "backend/tests/sentinel/sentinel_test.go",
        "coverage_bearing": True,
        "hook_consumers": [],
        "baseline_status": "sentinel",
        "replacement_status": "not-started",
        "parity_status": "not-assessed",
        "disposition": "retain-until-parity",
        "evidence": ["sentinel"],
        "trace_ids": [],
    }


def run_self_test() -> int:
    # TemporaryDirectory documents and enforces that the sentinel never writes
    # into the real audit directory or mutates inventory.json.
    with tempfile.TemporaryDirectory(prefix="hub-test-inventory-sentinel-") as temp_dir:
        expected_e2e_target = "testframework/e2e/platform/test_critical_gate.py"
        actual_e2e_target = e2e_target(Path("platform/critical-gate.spec.ts"))
        if TARGET_ROOTS["python"] != "testframework/" or actual_e2e_target != expected_e2e_target:
            print(
                "RED: root Python/E2E target mapping drifted: "
                f"python={TARGET_ROOTS['python']} e2e={actual_e2e_target}",
                file=sys.stderr,
            )
            return 1
        mapping_checks = (
            (go_component_target_class(["http-server"], []), "python-api"),
            (go_component_target_class(["postgres"], []), "python-integration"),
            (
                go_component_target_class([], ["github.com/ory/dockertest/v3"]),
                "python-integration",
            ),
            (
                is_go_component(
                    "", ["github.com/valerii/hub/backend/testframework/env"]
                ),
                True,
            ),
            (
                go_component_target_class(
                    [], ["github.com/valerii/hub/backend/testframework/mongotest"]
                ),
                "python-integration",
            ),
            (go_component_target_class([], []), "python-contract"),
            (
                go_component_target(
                    Path("audit/coverage_gate_task225_test.go"), "python-integration"
                ),
                "testframework/integration/audit/test_coverage_gate_task225.py",
            ),
            (
                vitest_target(Path("domains/tracker/issue-card.test.tsx")),
                "frontend/tests/tracker/issue-card.test.tsx",
            ),
            (
                frontend_static_contract_target(
                    Path("domains/platform/docs-drift-task123.test.ts")
                ),
                "testframework/contract/frontend/platform/test_docs_drift_task123.py",
            ),
            (
                frontend_support_target(
                    Path("domains/tracker/projects-sprint-version-required.fixture.ts")
                ),
                "frontend/tests/support/tracker/projects-sprint-version-required.fixture.ts",
            ),
            (
                e2e_support_target(Path("shared/pages/base-page.ts")),
                "testframework/e2e/support/shared/pages/base_page.py",
            ),
            (
                go_target(Path("server/wiring_fiber_recover_test.go")),
                "backend/tests/server/wiring_fiber_recover_test.go",
            ),
            (
                canonical_owner_wave("sentinel", "go-test", "auth", "go-unit"),
                "P133/TASK-210",
            ),
            (
                canonical_owner_wave("sentinel", "go-test", "auth", "python-api"),
                "P133/TASK-227",
            ),
            (
                canonical_owner_wave(
                    "sentinel", "go-test", "tracker", "python-integration"
                ),
                "P134/TASK-228",
            ),
            (
                canonical_owner_wave("sentinel", "go-test", "ai", "python-contract"),
                "P135/TASK-229",
            ),
            (
                canonical_owner_wave(
                    "sentinel", "vitest", "platform", "python-static-contract"
                ),
                "P131/TASK-230",
            ),
        )
        if any(actual != expected for actual, expected in mapping_checks):
            print(f"RED: target/owner mapping drifted: {mapping_checks}", file=sys.stderr)
            return 1
        sentinel_path = "backend/testframework/sentinel/sentinel_test.go"
        clean = {
            "counts": {
                "go_files": 1,
                "go_legacy_files": 1,
                "go_server_files": 0,
                "go_declared_tests": 1,
                "go_legacy_declared_tests": 1,
                "go_server_declared_tests": 0,
                "go_unit_files": 1,
                "go_component_files": 0,
                "go_component_api_files": 0,
                "go_component_integration_files": 0,
                "go_component_contract_files": 0,
                "vitest_files": 0,
                "frontend_unit_files": 0,
                "frontend_static_contract_files": 0,
                "frontend_support_files": 0,
                "playwright_files": 0,
                "playwright_cases": 0,
                "playwright_support_files": 0,
                "canonical_hook_files": 0,
                "unmapped_files": 0,
            },
            "rows": [sentinel_row(sentinel_path)],
            "hooks": [],
            "sentinel_temp_dir": temp_dir,
        }
        clean_errors = validate_inventory(clean, live_source_paths={sentinel_path}, enforce_expected=False)
        duplicate = deepcopy(clean)
        duplicate["rows"].append(deepcopy(duplicate["rows"][0]))
        duplicate_errors = validate_inventory(duplicate, live_source_paths={sentinel_path}, enforce_expected=False)
        removed = deepcopy(clean)
        removed["rows"] = []
        removed_errors = validate_inventory(removed, live_source_paths={sentinel_path}, enforce_expected=False)
        misrouted = deepcopy(clean)
        misrouted["rows"][0]["target_class"] = "python-integration"
        misrouted_errors = validate_inventory(
            misrouted, live_source_paths={sentinel_path}, enforce_expected=False
        )
        if clean_errors:
            print(f"RED: clean sentinel unexpectedly failed: {clean_errors}", file=sys.stderr)
            return 1
        if not any("duplicate source rows" in error for error in duplicate_errors):
            print("RED: duplicate sentinel was not detected", file=sys.stderr)
            return 1
        if not any("missing source rows" in error for error in removed_errors):
            print("RED: removal sentinel was not detected", file=sys.stderr)
            return 1
        if not any("bad target root" in error for error in misrouted_errors):
            print("RED: component-to-backend misroute was not detected", file=sys.stderr)
            return 1
        schema = json.loads(SCHEMA_PATH.read_text(encoding="utf-8"))
        row_schema = {"$ref": "#/$defs/testRow"}
        clean_schema_errors = schema_contract_errors(
            clean["rows"][0], row_schema, root_schema=schema, path="$.sentinel"
        )
        missing_owner = deepcopy(clean["rows"][0])
        del missing_owner["owner_wave"]
        missing_owner_errors = schema_contract_errors(
            missing_owner, row_schema, root_schema=schema, path="$.sentinel"
        )
        if clean_schema_errors:
            print(f"RED: clean schema sentinel unexpectedly failed: {clean_schema_errors}", file=sys.stderr)
            return 1
        if not any("missing required property 'owner_wave'" in error for error in missing_owner_errors):
            print("RED: schema sentinel did not reject missing owner_wave", file=sys.stderr)
            return 1
    print(
        "GREEN: root Python/frontend/E2E/component mappings and clean inventory/schema "
        "sentinels passed; duplicate, removal, misroute, and missing-owner "
        "sentinels produced RED"
    )
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    actions = parser.add_mutually_exclusive_group(required=True)
    actions.add_argument("--generate", action="store_true", help="write deterministic inventory.json")
    actions.add_argument("--check", action="store_true", help="validate inventory.json against current inputs")
    actions.add_argument("--self-test", action="store_true", help="run safe in-memory/temp sentinel validation")
    args = parser.parse_args()
    if args.self_test:
        return run_self_test()
    if args.check:
        return run_check()
    data = build_inventory()
    errors = validate_inventory(data, live_source_paths=live_test_paths())
    if errors:
        for error in errors:
            print(f"RED: {error}", file=sys.stderr)
        return 1
    payload = render(data)
    OUTPUT_PATH.write_bytes(payload)
    print(f"WROTE {rel(OUTPUT_PATH)} sha256={sha256_bytes(payload)} rows={len(data['rows'])}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
