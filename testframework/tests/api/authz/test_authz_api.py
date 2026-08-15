from __future__ import annotations

import pytest

from hub_testkit.domains.authz import (
    AuthzApi,
    PrincipalSeed,
    actor_from_principal,
    parse_management_success,
    parse_registration_mode,
    parse_success_message,
    pending_approval_is_non_enumerating,
    pending_contains_principals,
    pending_excludes_principals,
)
from hub_testkit.harness.authz import AuthzHarness


def _active_seed(label: str, kind: str) -> PrincipalSeed:
    return PrincipalSeed(label=label, kind=kind, state="active", credential_count=1)


def _pending_seed(label: str) -> PrincipalSeed:
    return PrincipalSeed(label=label, kind="member", state="pending", credential_count=0)


@pytest.mark.case_id("TC-AUTHZ-0001")
@pytest.mark.api
def test_system_admin_read_matrix(authz_harness: AuthzHarness) -> None:
    api = AuthzApi(authz_harness)
    founder = authz_harness.provision_principal(_active_seed("authz-founder", "founder"))
    admin = authz_harness.provision_principal(_active_seed("authz-admin", "admin"))
    member = authz_harness.provision_principal(_active_seed("authz-member", "member"))

    assert api.registration_mode(None, "authz-mode-anonymous").response.status_code == 401
    assert (
        api.registration_mode(actor_from_principal(member), "authz-mode-member").response.status_code
        == 403
    )
    assert parse_registration_mode(
        api.registration_mode(actor_from_principal(founder), "authz-mode-founder")
    ).mode in {"open", "invite_only"}
    assert parse_registration_mode(
        api.registration_mode(actor_from_principal(admin), "authz-mode-admin")
    ).mode in {"open", "invite_only"}


@pytest.mark.case_id("TC-AUTHZ-0002")
@pytest.mark.api
def test_pending_approval_and_rejection_are_admin_only(authz_harness: AuthzHarness) -> None:
    api = AuthzApi(authz_harness)
    admin = authz_harness.provision_principal(_active_seed("authz-admin", "admin"))
    member = authz_harness.provision_principal(_active_seed("authz-member", "member"))
    pending_approve = authz_harness.provision_principal(_pending_seed("authz-approve"))
    pending_reject = authz_harness.provision_principal(_pending_seed("authz-reject"))
    assert (
        api.list_pending(actor_from_principal(member), "authz-pending-member-list")
        .response.status_code
        == 403
    )
    assert (
        api.approve_pending(
            actor_from_principal(member),
            pending_approve.user_id,
            "authz-pending-member-approve",
        )
        .response.status_code
        == 403
    )
    assert (
        api.reject_pending(
            actor_from_principal(member),
            pending_reject.user_id,
            "authz-pending-member-reject",
        )
        .response.status_code
        == 403
    )
    assert pending_contains_principals(
        api.list_pending(actor_from_principal(admin), "authz-pending-admin-list"),
        (pending_approve, pending_reject),
    )
    assert (
        parse_success_message(
            api.approve_pending(
                actor_from_principal(admin),
                pending_approve.user_id,
                "authz-pending-approve",
            )
        ).message
        == "user approved"
    )
    assert (
        parse_success_message(
            api.reject_pending(
                actor_from_principal(admin),
                pending_reject.user_id,
                "authz-pending-reject",
            )
        ).message
        == "user rejected"
    )
    assert pending_excludes_principals(
        api.list_pending(actor_from_principal(admin), "authz-pending-admin-after"),
        (pending_approve, pending_reject),
    )


@pytest.mark.case_id("TC-AUTHZ-0003")
@pytest.mark.api
def test_deactivation_revokes_credentials_and_preserves_founder(
    authz_harness: AuthzHarness,
) -> None:
    api = AuthzApi(authz_harness)
    founder = authz_harness.provision_principal(_active_seed("authz-founder", "founder"))
    admin = authz_harness.provision_principal(_active_seed("authz-admin", "admin"))
    ordinary = authz_harness.provision_principal(_active_seed("authz-member", "member"))
    assert parse_management_success(
        api.deactivate(
            actor_from_principal(admin),
            ordinary.user_id,
            "authz-deactivate-member",
        )
    ).success is True
    assert api.me(actor_from_principal(ordinary), "authz-me-deactivated").response.status_code == 403
    assert (
        api.refresh(ordinary.credentials[0].refresh_token, "authz-refresh-deactivated")
        .response.status_code
        == 401
    )
    assert (
        api.deactivate(
            actor_from_principal(admin),
            founder.user_id,
            "authz-deactivate-founder",
        ).response.status_code
        == 403
    )
    assert api.me(actor_from_principal(founder), "authz-me-founder").response.status_code == 200


@pytest.mark.case_id("TC-AUTHZ-0004")
@pytest.mark.api
def test_pending_admin_endpoint_does_not_enumerate_active_users(
    authz_harness: AuthzHarness,
) -> None:
    api = AuthzApi(authz_harness)
    admin = authz_harness.provision_principal(_active_seed("authz-admin", "admin"))
    active = authz_harness.provision_principal(_active_seed("authz-active", "member"))
    assert pending_approval_is_non_enumerating(
        api.approve_pending(
            actor_from_principal(admin),
            active.user_id,
            "authz-enumeration-known",
        ),
        api.approve_pending(
            actor_from_principal(admin),
            "00000000-0000-4000-8000-000000000001",
            "authz-enumeration-unknown",
        ),
        active,
    )
