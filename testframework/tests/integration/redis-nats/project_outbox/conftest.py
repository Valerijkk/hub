from __future__ import annotations

from collections.abc import Iterator

import pytest

from hub_testkit.harness.project_outbox import ProjectOutboxRuntime
from hub_testkit.harness.pytest_project_outbox import (
    cleanup_project_outbox_runtime,
    create_project_outbox_runtime_from_config,
)


@pytest.fixture
def project_outbox_runtime(
    request: pytest.FixtureRequest,
) -> Iterator[ProjectOutboxRuntime]:
    runtime = create_project_outbox_runtime_from_config(request.config)
    try:
        yield runtime
    finally:
        cleanup_project_outbox_runtime(runtime)
