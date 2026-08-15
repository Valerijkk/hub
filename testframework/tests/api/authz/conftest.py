from __future__ import annotations

from collections.abc import Iterator

import pytest

from hub_testkit.harness.authz import AuthzHarness
from hub_testkit.harness.pytest_authz import (
    cleanup_authz_harness,
    create_authz_harness_from_config,
)


@pytest.fixture
def authz_harness(request: pytest.FixtureRequest) -> Iterator[AuthzHarness]:
    harness = create_authz_harness_from_config(request.config)
    try:
        yield harness
    finally:
        cleanup_authz_harness(harness)
