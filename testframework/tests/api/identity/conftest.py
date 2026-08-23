from __future__ import annotations

from collections.abc import Iterator

import pytest

from hub_testkit.harness.identity import IdentityHarness
from hub_testkit.harness.pytest_identity import require_identity_harness_factory


@pytest.fixture
def identity_harness(request: pytest.FixtureRequest) -> Iterator[IdentityHarness]:
    try:
        harness = require_identity_harness_factory(request.config)()
        if not isinstance(harness, IdentityHarness):
            raise ValueError("identity harness is invalid")
    except Exception:
        raise pytest.UsageError("identity harness is unavailable") from None
    try:
        yield harness
    finally:
        try:
            harness.cleanup()
        except Exception:
            raise pytest.UsageError("identity harness cleanup failed") from None
