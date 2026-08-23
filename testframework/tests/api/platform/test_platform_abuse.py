from __future__ import annotations

import hashlib

import pytest

from hub_testkit.domains.platform_abuse import (
    OwnedVerifyScenario,
    PlatformAbuseApi,
    PlatformAbuseSeed,
    rate_limit_sequence_is_bounded,
    verify_replay_is_rejected,
)
from hub_testkit.harness.platform_abuse import PlatformAbuseHarness


def _has_go_client_hash(scenario: object) -> bool:
    try:
        if type(scenario) is not OwnedVerifyScenario:
            return False
        return scenario.server_observed_client_hash == hashlib.sha256(
            scenario.server_observed_client.encode("utf-8")
        ).hexdigest()[:16]
    except Exception:
        return False


@pytest.mark.case_id("TC-PLATFORM-0001")
@pytest.mark.api
def test_magic_verify_has_bounded_production_rate_limit(
    platform_abuse_harness: PlatformAbuseHarness,
) -> None:
    scenario = platform_abuse_harness.provision_scenario(
        PlatformAbuseSeed(label="platform-rate-limit", flow="unknown-code-rate-limit")
    )
    assert _has_go_client_hash(scenario)
    api = PlatformAbuseApi(platform_abuse_harness)
    assert rate_limit_sequence_is_bounded(
        api.verify(scenario, "platform-rate-one"),
        api.verify(scenario, "platform-rate-two"),
        api.verify(scenario, "platform-rate-three"),
        api.verify(scenario, "platform-rate-four"),
        scenario,
    )


@pytest.mark.case_id("TC-PLATFORM-0002")
@pytest.mark.api
def test_magic_verify_rejects_replay_without_secret_leak(
    platform_abuse_harness: PlatformAbuseHarness,
) -> None:
    scenario = platform_abuse_harness.provision_scenario(
        PlatformAbuseSeed(label="platform-replay", flow="one-time-code-replay")
    )
    api = PlatformAbuseApi(platform_abuse_harness)
    assert verify_replay_is_rejected(
        api.verify(scenario, "platform-replay-first"),
        api.verify(scenario, "platform-replay-second"),
        scenario,
    )
