from __future__ import annotations

from dataclasses import dataclass

import pytest

from hub_testkit.runtime.redaction import Redactor
from hub_testkit.traceability import CaseIndex, load_marker_policy

from .collection import CollectionError, load_case_index, validate_collection


@dataclass(frozen=True, slots=True)
class _PluginState:
    case_index: CaseIndex
    redactor: Redactor


_PLUGIN_STATE_KEY: pytest.StashKey[_PluginState] = pytest.StashKey()
_OPTIONS = (
    "--hub-case-root",
    "--hub-case-schema",
    "--hub-marker-policy",
    "--hub-redaction-policy",
)
_MARKERS = (
    "case_id(id): Hub traceability case identifier",
    "contract: Hub contract test layer",
    "api: Hub API test layer",
    "integration: Hub integration test layer",
    "e2e: Hub end-to-end test layer",
)


def pytest_addoption(parser: pytest.Parser) -> None:
    group = parser.getgroup("hub-testframework")
    for option in _OPTIONS:
        group.addoption(option, action="store", default=None)


def pytest_configure(config: pytest.Config) -> None:
    try:
        for marker in _MARKERS:
            config.addinivalue_line("markers", marker)
        case_root, schema_path, marker_policy_path, redaction_policy_path = _option_values(config)
        marker_policy = load_marker_policy(marker_policy_path)
        redactor = Redactor.from_policy(redaction_policy_path)
        case_index = load_case_index(case_root, schema_path, marker_policy)
        config.stash[_PLUGIN_STATE_KEY] = _PluginState(case_index=case_index, redactor=redactor)
    except Exception:
        raise pytest.UsageError("Hub test framework configuration is invalid") from None


@pytest.hookimpl(trylast=True)
def pytest_collection_finish(session: pytest.Session) -> None:
    try:
        state = session.config.stash.get(_PLUGIN_STATE_KEY, None)
        if not isinstance(state, _PluginState):
            raise CollectionError("plugin state is invalid")
        validate_collection(session.items, state.case_index, state.redactor)
    except Exception:
        raise pytest.UsageError("Hub test framework collection is invalid") from None


def _option_values(config: pytest.Config) -> tuple[str, str, str, str]:
    case_root = _required_option(config, _OPTIONS[0])
    schema_path = _required_option(config, _OPTIONS[1])
    marker_policy_path = _required_option(config, _OPTIONS[2])
    redaction_policy_path = _required_option(config, _OPTIONS[3])
    return case_root, schema_path, marker_policy_path, redaction_policy_path


def _required_option(config: pytest.Config, option: str) -> str:
    value = config.getoption(option)
    if type(value) is not str or not value:
        raise CollectionError("plugin options are invalid")
    return value
