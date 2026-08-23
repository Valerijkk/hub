from __future__ import annotations

import pytest

from hub_testkit.domains.identity.api import IdentityApi
from hub_testkit.domains.identity.builders import actor_from_session
from hub_testkit.domains.identity.models import IdentitySeed
from hub_testkit.domains.identity.oracles import (
    parse_me_identity,
    parse_session_identities,
    parse_token_pair,
)
from hub_testkit.harness.identity import IdentityHarness
from hub_testkit.http.transport import HttpRequest


@pytest.mark.case_id("TC-IDENTITY-0001")
@pytest.mark.api
def test_refresh_rejects_invalid_and_deactivated_credentials(
    identity_harness: IdentityHarness,
) -> None:
    api = IdentityApi(identity_harness)
    malformed = HttpRequest(
        method="POST",
        path="/api/auth/refresh",
        actor=None,
        correlation_id="identity-refresh-malformed",
        idempotency_key=None,
        body=b"{",
        content_type="application/json",
    )
    assert identity_harness.send(malformed).response.status_code == 400
    assert (
        api.refresh("invalid-refresh-token", "identity-refresh-invalid").response.status_code
        == 401
    )

    deactivated = identity_harness.provision(
        IdentitySeed(label="identity-deactivated", status="deactivated", session_count=1)
    )

    assert deactivated.status == "deactivated"
    assert len(deactivated.sessions) == 1
    assert (
        api.refresh(
            deactivated.sessions[0].refresh_token,
            "identity-refresh-deactivated",
        ).response.status_code
        == 401
    )


@pytest.mark.case_id("TC-IDENTITY-0002")
@pytest.mark.api
def test_refresh_rotates_and_replay_is_rejected(identity_harness: IdentityHarness) -> None:
    api = IdentityApi(identity_harness)
    provisioned = identity_harness.provision(
        IdentitySeed(label="identity-rotation", status="active", session_count=1)
    )
    credential = provisioned.sessions[0]

    rotated = parse_token_pair(
        api.refresh(credential.refresh_token, "identity-refresh-rotate")
    )

    assert rotated.refresh_token != credential.refresh_token
    assert (
        api.refresh(credential.refresh_token, "identity-refresh-replay").response.status_code
        == 401
    )


@pytest.mark.case_id("TC-IDENTITY-0003")
@pytest.mark.api
def test_me_requires_active_bearer_and_logout_revokes_refresh(
    identity_harness: IdentityHarness,
) -> None:
    api = IdentityApi(identity_harness)
    active = identity_harness.provision(
        IdentitySeed(label="identity-active", status="active", session_count=1)
    )
    deactivated = identity_harness.provision(
        IdentitySeed(label="identity-disabled", status="deactivated", session_count=1)
    )
    active_actor = actor_from_session(active.user_id, active.sessions[0])
    deactivated_actor = actor_from_session(deactivated.user_id, deactivated.sessions[0])

    assert api.me(None, "identity-me-anonymous").response.status_code == 401
    me = parse_me_identity(api.me(active_actor, "identity-me-active"))
    assert (me.id, me.email, me.status) == (active.user_id, active.email, active.status)
    assert active.email not in repr(active)
    assert me.email not in repr(me)
    assert api.me(deactivated_actor, "identity-me-deactivated").response.status_code == 403
    assert api.logout(active_actor, "identity-logout-active").response.status_code == 200
    assert (
        api.refresh(
            active.sessions[0].refresh_token,
            "identity-refresh-after-logout",
        ).response.status_code
        == 401
    )


@pytest.mark.case_id("TC-IDENTITY-0004")
@pytest.mark.api
def test_sessions_are_owner_scoped_and_revoke_invalidates_target_refresh(
    identity_harness: IdentityHarness,
) -> None:
    api = IdentityApi(identity_harness)
    owner = identity_harness.provision(
        IdentitySeed(label="identity-owner", status="active", session_count=1)
    )
    foreign = identity_harness.provision(
        IdentitySeed(label="identity-foreign", status="active", session_count=1)
    )
    owner_actor = actor_from_session(owner.user_id, owner.sessions[0])
    listed = parse_session_identities(api.list_sessions(owner_actor, "identity-sessions-list"))

    assert {session.session_id for session in listed} == {
        session.session_id for session in owner.sessions
    }
    assert foreign.sessions[0].session_id not in {
        session.session_id for session in listed
    }
    assert len(listed) == 1
    assert sum(session.is_current for session in listed) == 1
    current = next(session for session in listed if session.is_current)
    assert current.session_id == owner.sessions[0].session_id
    assert (
        api.revoke_session(
            owner_actor,
            foreign.sessions[0].session_id,
            "identity-sessions-foreign",
        ).response.status_code
        == 404
    )
    foreign_refresh = parse_token_pair(
        api.refresh(foreign.sessions[0].refresh_token, "identity-refresh-foreign")
    )
    assert foreign_refresh.refresh_token != foreign.sessions[0].refresh_token
    assert (
        api.revoke_session(
            owner_actor,
            owner.sessions[0].session_id,
            "identity-sessions-revoke",
        ).response.status_code
        == 200
    )
    relisted = parse_session_identities(
        api.list_sessions(owner_actor, "identity-sessions-relist")
    )
    assert relisted == ()
    assert (
        api.refresh(owner.sessions[0].refresh_token, "identity-refresh-revoked").response.status_code
        == 401
    )
