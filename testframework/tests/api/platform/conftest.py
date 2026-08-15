from __future__ import annotations

from collections.abc import Iterator

import pytest

from hub_testkit.harness.platform_abuse import PlatformAbuseHarness
from hub_testkit.harness.pytest_platform_abuse import (
    cleanup_platform_abuse_harness,
    create_platform_abuse_harness_from_config,
)


@pytest.fixture
def platform_abuse_harness(request: pytest.FixtureRequest) -> Iterator[PlatformAbuseHarness]:
    harness = create_platform_abuse_harness_from_config(request.config)
    try:
        yield harness
    finally:
        cleanup_platform_abuse_harness(harness)
