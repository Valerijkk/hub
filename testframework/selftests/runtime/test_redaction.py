from __future__ import annotations

from pathlib import Path

import pytest

from hub_testkit.runtime import RedactionError, Redactor


PROJECT_ROOT = Path(__file__).resolve().parents[2]
POLICY_PATH = PROJECT_ROOT / "policies" / "redaction.toml"
REDACTION_HEADER = (
    'schema = "hub.testframework.redaction/v1"\n'
    'replacement = "[REDACTED]"\n'
    'sensitive_key_fragments = ["secret"]\n'
)


def _redactor() -> Redactor:
    return Redactor.from_policy(POLICY_PATH)


def test_redactor_recursively_removes_key_and_value_secret_canaries() -> None:
    redactor = _redactor()
    payload = {
        "headers": {"Authorization": "Bearer top-secret-token"},
        "contact": "alice@example.test",
        "opaque": "0123456789abcdef0123456789abcdef",
        "endpoint": "https://alice:password@example.test/api",
        "nested": ["eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.signature", ("safe", 7)],
    }

    redacted = redactor.redact_mapping(payload)

    assert redacted == {
        "headers": {"Authorization": "[REDACTED]"},
        "contact": "[REDACTED]",
        "opaque": "[REDACTED]",
        "endpoint": "[REDACTED]",
        "nested": ["[REDACTED]", ("safe", 7)],
    }
    assert redactor.redact_mapping(redacted) == redacted


def test_redact_text_removes_a_bearer_canary_without_changing_safe_text() -> None:
    redactor = _redactor()

    assert redactor.redact_text("status=ok") == "status=ok"
    assert redactor.redact_text("Authorization: Bearer secret-token") == "Authorization: [REDACTED]"


def test_redactor_replaces_any_sensitive_key_without_inspecting_its_value() -> None:
    redactor = _redactor()

    assert redactor.redact_mapping({"future_api_secret_blob": object()}) == {
        "future_api_secret_blob": "[REDACTED]"
    }


def test_redactor_fails_closed_for_cycles_unsupported_values_and_non_string_keys() -> None:
    redactor = _redactor()
    cyclic: dict[str, object] = {"visible": "safe"}
    cyclic["self"] = cyclic

    for payload in (cyclic, {"visible": object()}, {1: "not-a-string-key"}):
        with pytest.raises(RedactionError) as error:
            redactor.redact_mapping(payload)  # type: ignore[arg-type]
        assert "not-a-string-key" not in str(error.value)


@pytest.mark.parametrize(
    "contents",
    [
        REDACTION_HEADER + 'value_patterns = ["token"]\n' + 'unexpected = "secret-policy-canary"\n',
        REDACTION_HEADER + 'value_patterns = ["("]\n',
    ],
)
def test_redaction_policy_rejects_unknown_fields_and_invalid_patterns(
    tmp_path: Path,
    contents: str,
) -> None:
    path = tmp_path / "redaction.toml"
    path.write_text(contents, encoding="utf-8")

    with pytest.raises(RedactionError) as error:
        Redactor.from_policy(path)

    assert "secret-policy-canary" not in str(error.value)
    assert error.value.__cause__ is None
