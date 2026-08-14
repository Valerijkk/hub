from __future__ import annotations

import re
import tomllib
from pathlib import Path


PROJECT_ROOT = Path(__file__).resolve().parents[3]
REPOSITORY_ROOT = PROJECT_ROOT.parent
POLICY_PATH = PROJECT_ROOT / "policies" / "profiles.toml"
PLAN_PATH = (
    REPOSITORY_ROOT
    / ".planning"
    / "phases"
    / "130-tf-traceability-gates"
    / "PLAN.md"
)
PROFILE_PATTERN = re.compile(
    r'^<profile name="(?P<name>[^"]+)" '
    r'image_role="(?P<image_role>[^"]+)" '
    r'limit_class="(?P<limit_class>[^"]+)" '
    r'target="(?P<target>[^"]+)" '
    r'deadline="(?P<deadline_seconds>[0-9]+)"/>$',
    re.MULTILINE,
)
PROFILE_KEYS = ["name", "image_role", "limit_class", "target", "deadline_seconds"]


def _plan_profiles() -> list[dict[str, object]]:
    profiles: list[dict[str, object]] = []
    for match in PROFILE_PATTERN.finditer(PLAN_PATH.read_text(encoding="utf-8")):
        profile: dict[str, object] = match.groupdict()
        profile["deadline_seconds"] = int(profile["deadline_seconds"])
        profiles.append(profile)
    return profiles


def _policy() -> dict[str, object]:
    with POLICY_PATH.open("rb") as stream:
        return tomllib.load(stream)


def test_profile_policy_is_the_literal_phase_130_registry() -> None:
    policy = _policy()
    profiles = policy["profiles"]

    assert list(policy) == ["schema", "profile_count", "profiles"]
    assert policy["schema"] == "hub.testframework.profiles/v1"
    assert policy["profile_count"] == 47
    assert len(profiles) == 47
    assert profiles == _plan_profiles()


def test_profile_records_are_closed_unique_and_bounded() -> None:
    profiles = _policy()["profiles"]
    names: set[str] = set()

    for profile in profiles:
        assert list(profile) == PROFILE_KEYS
        assert profile["name"] not in names
        names.add(profile["name"])
        assert profile["image_role"] in {"browser", "ci", "integration", "load", "python-core"}
        assert profile["limit_class"] in {"large", "medium", "small"}
        assert re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", profile["name"])
        assert re.fullmatch(r"[a-z0-9]+(?:[/-][a-z0-9]+)*", profile["target"])
        assert 1 <= profile["deadline_seconds"] <= 1800
