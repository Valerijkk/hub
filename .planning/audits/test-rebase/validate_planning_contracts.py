#!/usr/bin/env python3
"""Fail-closed validation for planning decisions referenced by runtime tests."""

from __future__ import annotations

import argparse
import tempfile
from pathlib import Path


CONTRACTS: tuple[tuple[str, tuple[str, ...]], ...] = (
    (
        ".planning/decisions/ADR-ENT-CHANNELMEMBER-LWW-2026.md",
        (
            "LWW accepted",
            "A3W8-1-002",
            "Owner lock / approve LWW path",
            "Accepted 2026-07-29",
        ),
    ),
)


def validate(root: Path) -> list[str]:
    errors: list[str] = []
    for relative_path, required_markers in CONTRACTS:
        path = root / relative_path
        if not path.is_file():
            errors.append(f"missing planning contract: {relative_path}")
            continue
        body = path.read_text(encoding="utf-8")
        for marker in required_markers:
            if marker not in body:
                errors.append(
                    f"missing marker {marker!r} in planning contract {relative_path}"
                )
    return errors


def self_test() -> None:
    relative_path, required_markers = CONTRACTS[0]
    with tempfile.TemporaryDirectory() as temp_dir:
        root = Path(temp_dir)
        path = root / relative_path
        path.parent.mkdir(parents=True)
        path.write_text("\n".join(required_markers), encoding="utf-8")
        assert validate(root) == []
        path.write_text("\n".join(required_markers[:-1]), encoding="utf-8")
        assert validate(root), "validator must reject a missing marker"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--self-test", action="store_true")
    args = parser.parse_args()

    if args.self_test:
        self_test()
        print("planning-contract validator self-test: PASS")
        return 0

    repo_root = Path(__file__).resolve().parents[3]
    errors = validate(repo_root)
    if errors:
        for error in errors:
            print(f"ERROR: {error}")
        return 1

    print(f"planning-contract validation: PASS ({len(CONTRACTS)} contract)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
