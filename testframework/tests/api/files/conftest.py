from __future__ import annotations

from collections.abc import Iterator

import pytest

from hub_testkit.harness.files import FilesRuntime
from hub_testkit.harness.pytest_files import (
    cleanup_files_runtime,
    create_files_runtime_from_config,
)


@pytest.fixture
def files_runtime(request: pytest.FixtureRequest) -> Iterator[FilesRuntime]:
    runtime = create_files_runtime_from_config(request.config)
    try:
        yield runtime
    finally:
        cleanup_files_runtime(runtime)
