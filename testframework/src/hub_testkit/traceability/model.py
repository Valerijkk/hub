from __future__ import annotations

import re
from pathlib import Path
from typing import Literal

from pydantic import BaseModel, ConfigDict, Field, field_validator, model_validator


CASE_ID_PATTERN = r"^TC-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*-[0-9]{4}$"
TASK_PATTERN = r"^TASK-[0-9]{3}$"
UPPER_IDENTIFIER_PATTERN = r"^[A-Z][A-Z0-9]*(?:-[A-Z0-9]+)+$"
PYTEST_NODE_PATTERN = (
    r"^testframework/tests/(?:contract|api|integration|e2e)/"
    r"(?:[A-Za-z0-9][A-Za-z0-9_-]*/)*"
    r"test_[A-Za-z0-9_]+\.py::test_[A-Za-z0-9_]+$"
)
LOCUST_ENTRYPOINT_PATTERN = (
    r"^testframework/load/(?:[A-Za-z0-9][A-Za-z0-9_-]*/)*[A-Za-z0-9_]+\.py$"
)
OWNER_PATTERN = r"^[a-z][a-z0-9-]{2,31}$"

_CASE_ID_RE = re.compile(CASE_ID_PATTERN)
_TASK_RE = re.compile(TASK_PATTERN)
_UPPER_IDENTIFIER_RE = re.compile(UPPER_IDENTIFIER_PATTERN)
_PYTEST_NODE_RE = re.compile(PYTEST_NODE_PATTERN)
_LOCUST_ENTRYPOINT_RE = re.compile(LOCUST_ENTRYPOINT_PATTERN)
_OWNER_RE = re.compile(OWNER_PATTERN)

Status = Literal["active", "planned", "retired"]
Scenario = Literal[
    "happy",
    "negative",
    "authn",
    "authz",
    "boundary",
    "edge",
    "corner",
    "failure-recovery",
    "security",
    "accessibility",
    "performance",
]
PytestLayer = Literal["contract", "api", "integration", "e2e"]


class PytestAutomation(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True, strict=True)

    kind: Literal["pytest"]
    layer: PytestLayer
    node_id: str

    @field_validator("node_id")
    @classmethod
    def _validate_node_id(cls, value: str) -> str:
        if type(value) is not str or _PYTEST_NODE_RE.fullmatch(value) is None:
            raise ValueError("pytest node ID is invalid")
        return value

    @model_validator(mode="after")
    def _require_matching_layer(self) -> PytestAutomation:
        if not self.node_id.startswith(f"testframework/tests/{self.layer}/"):
            raise ValueError("pytest layer is invalid")
        return self


class LocustAutomation(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True, strict=True)

    kind: Literal["locust"]
    entrypoint: str

    @field_validator("entrypoint")
    @classmethod
    def _validate_entrypoint(cls, value: str) -> str:
        if type(value) is not str or _LOCUST_ENTRYPOINT_RE.fullmatch(value) is None:
            raise ValueError("Locust entrypoint is invalid")
        return value


class ManualAutomation(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True, strict=True)

    kind: Literal["manual"]
    evidence_owner: str

    @field_validator("evidence_owner")
    @classmethod
    def _validate_evidence_owner(cls, value: str) -> str:
        if type(value) is not str or _OWNER_RE.fullmatch(value) is None:
            raise ValueError("manual evidence owner is invalid")
        return value


Automation = PytestAutomation | LocustAutomation | ManualAutomation


class MarkerPolicy(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True, strict=True)

    schema: Literal["hub.testframework.markers/v1"]
    case_id_pattern: str
    statuses: tuple[Status, ...] = Field(min_length=1)
    scenarios: tuple[Scenario, ...] = Field(min_length=1)
    pytest_layers: tuple[PytestLayer, ...] = Field(min_length=1)
    automation_kinds: tuple[Literal["pytest", "locust", "manual"], ...] = Field(min_length=1)


class CaseMetadata(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True, strict=True)

    schema: Literal["hub.testframework.case/v1"]
    id: str
    title: str
    status: Status
    task: str
    requirement: str
    risk: str
    scenario: Scenario
    subject: str
    preconditions: tuple[str, ...] = Field(min_length=1, max_length=64)
    oracle: str
    cleanup: str
    automation: Automation

    @field_validator("id")
    @classmethod
    def _validate_case_id(cls, value: str) -> str:
        if type(value) is not str or _CASE_ID_RE.fullmatch(value) is None:
            raise ValueError("case ID is invalid")
        return value

    @field_validator("task")
    @classmethod
    def _validate_task(cls, value: str) -> str:
        if type(value) is not str or _TASK_RE.fullmatch(value) is None:
            raise ValueError("task is invalid")
        return value

    @field_validator("requirement", "risk")
    @classmethod
    def _validate_upper_identifier(cls, value: str) -> str:
        if type(value) is not str or _UPPER_IDENTIFIER_RE.fullmatch(value) is None:
            raise ValueError("identifier is invalid")
        return value

    @field_validator("title", "subject", "oracle", "cleanup")
    @classmethod
    def _validate_text(cls, value: str) -> str:
        return _require_nonblank_text(value)

    @field_validator("preconditions")
    @classmethod
    def _validate_preconditions(cls, values: tuple[str, ...]) -> tuple[str, ...]:
        if len(values) != len(set(values)):
            raise ValueError("preconditions must be unique")
        return tuple(_require_nonblank_text(value) for value in values)

    @model_validator(mode="after")
    def _require_status_automation_pairing(self) -> CaseMetadata:
        if self.status == "active" and not isinstance(self.automation, PytestAutomation):
            raise ValueError("active cases require pytest automation")
        if self.status != "active" and isinstance(self.automation, PytestAutomation):
            raise ValueError("inactive cases require non-pytest automation")
        return self


class CaseRecord(BaseModel):
    model_config = ConfigDict(extra="forbid", frozen=True, strict=True)

    metadata: CaseMetadata
    source_path: Path
    markdown_body: str

    @field_validator("markdown_body")
    @classmethod
    def _validate_markdown_body(cls, value: str) -> str:
        if type(value) is not str or not value.strip() or len(value.encode("utf-8")) > 64 * 1024:
            raise ValueError("Markdown body is invalid")
        return value


def _require_nonblank_text(value: str) -> str:
    if type(value) is not str or not value.strip() or len(value) > 240:
        raise ValueError("text is invalid")
    return value
