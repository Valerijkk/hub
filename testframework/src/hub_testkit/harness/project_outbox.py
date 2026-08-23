from __future__ import annotations

import secrets
from dataclasses import dataclass, field
from typing import NoReturn, Protocol, TypeGuard, runtime_checkable

from hub_testkit.http import HttpExchange, HttpRequest
from hub_testkit.integration.project_outbox.api import ProjectOutboxApi
from hub_testkit.integration.project_outbox.builders import build_project_outbox_seed
from hub_testkit.integration.project_outbox.models import (
    OwnedProjectOutboxState,
    ProjectCreatedReceipt,
    ProjectOutboxFlow,
    ProjectOutboxSeed,
    ProjectProjection,
    ProvisionedProjectOutbox,
    _provisioned_details,
    _receipt_details,
    _seed_details,
    _state_details,
)
from hub_testkit.integration.project_outbox.oracles import (
    parse_created_project,
    project_matches_seed,
    receipt_matches_project_and_owner,
    require_project_outbox_status,
    state_matches_counts,
)


class ProjectOutboxRuntimeError(ValueError):
    pass


class ProjectOutboxAssertionError(AssertionError):
    pass


@runtime_checkable
class ProjectOutboxHarnessDelegate(Protocol):
    """Launcher-owned boundary for one exclusive nonce-bound topology."""

    def send(self, request: HttpRequest) -> HttpExchange:
        """Send one admitted product request through the owned server."""
        ...

    def provision(self, seed: ProjectOutboxSeed) -> ProvisionedProjectOutbox:
        """Provision one owner after creating its durable InterestPolicy observer."""
        ...

    def observe_project_created(self) -> ProjectCreatedReceipt:
        """Return one typed receipt without asserting broker acknowledgement order."""
        ...

    def state(self) -> OwnedProjectOutboxState:
        """Return closed final counts without asserting transaction boundaries."""
        ...

    def disconnect_observer(self) -> None:
        """Disconnect only the exact durable observer owned by this delegate."""
        ...

    def reconnect_observer(self) -> None:
        """Reconnect only the same durable observer without creating another."""
        ...

    def cleanup(self) -> None:
        """Quiesce and confirm the exact server, then delete owned resources.

        After confirmed process exit, cleanup removes only the exact observer,
        JetStream stream/node, MongoDB database, Redis instance/database, and
        PostgreSQL database, in that LIFO order. On bounded shutdown timeout it
        terminates only the recorded process and confirms exit before deletion.
        Prefix scans, shared mutation, foreign adoption, and cleanup-as-oracle
        are forbidden. The caller invokes this method exactly once; the future
        delegate owns internal completion and must reconcile partial cleanup
        failures before returning rather than requiring a caller retry.
        """
        ...


@runtime_checkable
class ProjectOutboxHarnessFactory(Protocol):
    """Future launcher supplies one fresh healthy exclusive delegate per test."""

    def __call__(self) -> ProjectOutboxHarnessDelegate:
        """Create full owned services and the durable InterestPolicy EVENTS observer."""
        ...


@dataclass(frozen=True, slots=True, repr=False, eq=False, init=False)
class ProjectOutboxRuntime:
    _delegate: ProjectOutboxHarnessDelegate = field(repr=False)
    _api: ProjectOutboxApi = field(repr=False)
    _runtime_binding: object = field(repr=False)
    _nonce: str = field(repr=False)
    _cleanup_called: bool = field(repr=False)

    def __init__(self) -> None:
        _raise_invalid()

    @classmethod
    def _create(cls, delegate: ProjectOutboxHarnessDelegate) -> ProjectOutboxRuntime:
        result = _runtime_create_safe(cls, delegate)
        del delegate
        if type(result) is cls:
            return result
        del result
        _raise_invalid()

    def provision(self, seed: ProjectOutboxSeed) -> ProvisionedProjectOutbox:
        result = _runtime_provision_safe(self, seed)
        del seed
        if type(result) is ProvisionedProjectOutbox:
            return result
        del result
        _raise_invalid()

    def create(
        self,
        seed: ProjectOutboxSeed,
        provisioned: ProvisionedProjectOutbox,
        correlation_id: str,
    ) -> HttpExchange:
        result = _runtime_create_request_safe(
            self,
            seed,
            provisioned,
            correlation_id,
            "authenticated",
        )
        del seed, provisioned, correlation_id
        if type(result) is HttpExchange:
            return result
        del result
        _raise_invalid()

    def create_anonymous(
        self,
        seed: ProjectOutboxSeed,
        provisioned: ProvisionedProjectOutbox,
        correlation_id: str,
    ) -> HttpExchange:
        result = _runtime_create_request_safe(
            self,
            seed,
            provisioned,
            correlation_id,
            "anonymous",
        )
        del seed, provisioned, correlation_id
        if type(result) is HttpExchange:
            return result
        del result
        _raise_invalid()

    def create_malformed(
        self,
        seed: ProjectOutboxSeed,
        provisioned: ProvisionedProjectOutbox,
        correlation_id: str,
    ) -> HttpExchange:
        result = _runtime_create_request_safe(
            self,
            seed,
            provisioned,
            correlation_id,
            "malformed",
        )
        del seed, provisioned, correlation_id
        if type(result) is HttpExchange:
            return result
        del result
        _raise_invalid()

    def observe_project_created(self) -> ProjectCreatedReceipt:
        result = _runtime_receipt_safe(self)
        if type(result) is ProjectCreatedReceipt:
            return result
        del result
        _raise_invalid()

    def state(self) -> OwnedProjectOutboxState:
        result = _runtime_state_safe(self)
        if type(result) is OwnedProjectOutboxState:
            return result
        del result
        _raise_invalid()

    def disconnect_observer(self) -> None:
        result = _runtime_observer_action_safe(self, "disconnect")
        if result is True:
            return
        del result
        _raise_invalid()

    def reconnect_observer(self) -> None:
        result = _runtime_observer_action_safe(self, "reconnect")
        if result is True:
            return
        del result
        _raise_invalid()

    def _cleanup(self) -> None:
        result = _runtime_cleanup_safe(self)
        if result is True:
            return
        del result
        _raise_invalid()

    def __copy__(self) -> NoReturn:
        _raise_invalid()

    def __deepcopy__(self, memo: object) -> NoReturn:
        del memo
        _raise_invalid()

    def __repr__(self) -> str:
        return "ProjectOutboxRuntime()"


def assert_project_create_returns_consistent_published_final_state(
    runtime: ProjectOutboxRuntime,
) -> None:
    result = _consistent_final_state_scenario_safe(runtime)
    del runtime
    if result is True:
        return
    del result
    _raise_assertion()


def assert_duplicate_project_key_conflict_leaves_one_consistent_final_state(
    runtime: ProjectOutboxRuntime,
) -> None:
    result = _duplicate_scenario_safe(runtime)
    del runtime
    if result is True:
        return
    del result
    _raise_assertion()


def assert_anonymous_project_create_is_rejected_with_zero_owned_final_counts(
    runtime: ProjectOutboxRuntime,
) -> None:
    result = _anonymous_scenario_safe(runtime)
    del runtime
    if result is True:
        return
    del result
    _raise_assertion()


def assert_malformed_project_json_is_rejected_with_zero_owned_final_counts(
    runtime: ProjectOutboxRuntime,
) -> None:
    result = _malformed_scenario_safe(runtime)
    del runtime
    if result is True:
        return
    del result
    _raise_assertion()


def assert_durable_project_created_observer_replays_after_observer_reconnect(
    runtime: ProjectOutboxRuntime,
) -> None:
    result = _observer_replay_scenario_safe(runtime)
    del runtime
    if result is True:
        return
    del result
    _raise_assertion()


def _consistent_final_state_scenario_safe(runtime: object) -> bool:
    try:
        parts = _runtime_parts_safe(runtime)
        if not isinstance(parts, tuple) or type(runtime) is not ProjectOutboxRuntime:
            return False
        _, _, binding, nonce = parts
        seed = build_project_outbox_seed("create", nonce)
        provisioned = runtime.provision(seed)
        exchange = runtime.create(
            seed,
            provisioned,
            _correlation_id(nonce, "create"),
        )
        project = parse_created_project(exchange)
        receipt = runtime.observe_project_created()
        state = runtime.state()
        return (
            project_matches_seed(project, seed)
            and receipt_matches_project_and_owner(
                receipt,
                project,
                seed,
                provisioned,
                binding,
            )
            and state_matches_counts(state, (1, 1, 1, 1), binding)
        )
    except Exception:
        return False


def _duplicate_scenario_safe(runtime: object) -> bool:
    try:
        parts = _runtime_parts_safe(runtime)
        if not isinstance(parts, tuple) or type(runtime) is not ProjectOutboxRuntime:
            return False
        _, _, binding, nonce = parts
        seed = build_project_outbox_seed("duplicate", nonce)
        provisioned = runtime.provision(seed)
        first_exchange = runtime.create(
            seed,
            provisioned,
            _correlation_id(nonce, "duplicate-first"),
        )
        project = parse_created_project(first_exchange)
        receipt = runtime.observe_project_created()
        second_exchange = runtime.create(
            seed,
            provisioned,
            _correlation_id(nonce, "duplicate-second"),
        )
        require_project_outbox_status(second_exchange, 409)
        state = runtime.state()
        return (
            project_matches_seed(project, seed)
            and receipt_matches_project_and_owner(
                receipt,
                project,
                seed,
                provisioned,
                binding,
            )
            and state_matches_counts(state, (1, 1, 1, 1), binding)
        )
    except Exception:
        return False


def _anonymous_scenario_safe(runtime: object) -> bool:
    try:
        parts = _runtime_parts_safe(runtime)
        if not isinstance(parts, tuple) or type(runtime) is not ProjectOutboxRuntime:
            return False
        _, _, binding, nonce = parts
        seed = build_project_outbox_seed("anonymous", nonce)
        provisioned = runtime.provision(seed)
        exchange = runtime.create_anonymous(
            seed,
            provisioned,
            _correlation_id(nonce, "anonymous"),
        )
        require_project_outbox_status(exchange, 401)
        state = runtime.state()
        return state_matches_counts(state, (0, 0, 0, 0), binding)
    except Exception:
        return False


def _malformed_scenario_safe(runtime: object) -> bool:
    try:
        parts = _runtime_parts_safe(runtime)
        if not isinstance(parts, tuple) or type(runtime) is not ProjectOutboxRuntime:
            return False
        _, _, binding, nonce = parts
        seed = build_project_outbox_seed("malformed", nonce)
        provisioned = runtime.provision(seed)
        exchange = runtime.create_malformed(
            seed,
            provisioned,
            _correlation_id(nonce, "malformed"),
        )
        require_project_outbox_status(exchange, 400)
        state = runtime.state()
        return state_matches_counts(state, (0, 0, 0, 0), binding)
    except Exception:
        return False


def _observer_replay_scenario_safe(runtime: object) -> bool:
    disconnected = False
    try:
        parts = _runtime_parts_safe(runtime)
        if not isinstance(parts, tuple) or type(runtime) is not ProjectOutboxRuntime:
            return False
        _, _, binding, nonce = parts
        seed = build_project_outbox_seed("observer-replay", nonce)
        provisioned = runtime.provision(seed)
        runtime.disconnect_observer()
        disconnected = True
        exchange = runtime.create(
            seed,
            provisioned,
            _correlation_id(nonce, "observer-replay"),
        )
        project = parse_created_project(exchange)
        runtime.reconnect_observer()
        disconnected = False
        receipt = runtime.observe_project_created()
        state = runtime.state()
        return (
            project_matches_seed(project, seed)
            and receipt_matches_project_and_owner(
                receipt,
                project,
                seed,
                provisioned,
                binding,
            )
            and state_matches_counts(state, (1, 1, 1, 1), binding)
        )
    except Exception:
        return False
    finally:
        if disconnected and type(runtime) is ProjectOutboxRuntime:
            _runtime_observer_action_safe(runtime, "reconnect")


def _runtime_create_safe(
    cls: type[ProjectOutboxRuntime],
    delegate: object,
) -> ProjectOutboxRuntime | object:
    try:
        if not _delegate_is_valid_safe(delegate):
            return object()
        nonce = secrets.token_hex(16)
        if not _nonce_is_valid_safe(nonce):
            return object()
        api = ProjectOutboxApi(sender=delegate)
        result = object.__new__(cls)
        object.__setattr__(result, "_delegate", delegate)
        object.__setattr__(result, "_api", api)
        object.__setattr__(result, "_runtime_binding", delegate)
        object.__setattr__(result, "_nonce", nonce)
        object.__setattr__(result, "_cleanup_called", False)
        return result
    except Exception:
        return object()


def _runtime_parts_safe(
    runtime: object,
) -> tuple[ProjectOutboxHarnessDelegate, ProjectOutboxApi, object, str] | object:
    try:
        if type(runtime) is not ProjectOutboxRuntime or runtime._cleanup_called is not False:
            return object()
        delegate = runtime._delegate
        api = runtime._api
        binding = runtime._runtime_binding
        nonce = runtime._nonce
        if (
            not _delegate_is_valid_safe(delegate)
            or type(api) is not ProjectOutboxApi
            or api.sender is not delegate
            or binding is not delegate
            or not _nonce_is_valid_safe(nonce)
        ):
            return object()
        return delegate, api, binding, nonce
    except Exception:
        return object()


def _runtime_provision_safe(
    runtime: object,
    seed: object,
) -> ProvisionedProjectOutbox | object:
    try:
        parts = _runtime_parts_safe(runtime)
        seed_details = _seed_details(seed)
        if (
            not isinstance(parts, tuple)
            or type(seed) is not ProjectOutboxSeed
            or not isinstance(seed_details, tuple)
        ):
            return object()
        delegate, _, binding, nonce = parts
        if seed_details[2] != nonce:
            return object()
        provisioned = delegate.provision(seed)
        details = _provisioned_details(provisioned)
        if not isinstance(details, tuple):
            return object()
        if details[3] != nonce or details[4] is not binding:
            return object()
        return provisioned
    except Exception:
        return object()


def _runtime_create_request_safe(
    runtime: object,
    seed: object,
    provisioned: object,
    correlation_id: object,
    mode: object,
) -> HttpExchange | object:
    try:
        parts = _runtime_parts_safe(runtime)
        seed_details = _seed_details(seed)
        provisioned_details = _provisioned_details(provisioned)
        if (
            not isinstance(parts, tuple)
            or type(seed) is not ProjectOutboxSeed
            or type(provisioned) is not ProvisionedProjectOutbox
            or not isinstance(seed_details, tuple)
            or not isinstance(provisioned_details, tuple)
            or type(correlation_id) is not str
        ):
            return object()
        _, api, binding, nonce = parts
        if (
            seed_details[2] != nonce
            or provisioned_details[3] != nonce
            or provisioned_details[4] is not binding
        ):
            return object()
        if mode == "authenticated":
            return api.create(seed, provisioned, correlation_id)
        if mode == "anonymous":
            return api.create_anonymous(seed, provisioned, correlation_id)
        if mode == "malformed":
            return api.create_malformed(seed, provisioned, correlation_id)
        return object()
    except Exception:
        return object()


def _runtime_receipt_safe(runtime: object) -> ProjectCreatedReceipt | object:
    try:
        parts = _runtime_parts_safe(runtime)
        if not isinstance(parts, tuple):
            return object()
        delegate, _, binding, _ = parts
        receipt = delegate.observe_project_created()
        details = _receipt_details(receipt)
        if not isinstance(details, tuple) or details[6] is not binding:
            return object()
        return receipt
    except Exception:
        return object()


def _runtime_state_safe(runtime: object) -> OwnedProjectOutboxState | object:
    try:
        parts = _runtime_parts_safe(runtime)
        if not isinstance(parts, tuple):
            return object()
        delegate, _, binding, _ = parts
        state = delegate.state()
        details = _state_details(state)
        if not isinstance(details, tuple) or details[4] is not binding:
            return object()
        return state
    except Exception:
        return object()


def _runtime_observer_action_safe(runtime: object, action: object) -> bool:
    try:
        parts = _runtime_parts_safe(runtime)
        if not isinstance(parts, tuple):
            return False
        delegate = parts[0]
        if action == "disconnect":
            delegate.disconnect_observer()
        elif action == "reconnect":
            delegate.reconnect_observer()
        else:
            return False
        return True
    except Exception:
        return False


def _runtime_cleanup_safe(runtime: object) -> bool:
    try:
        parts = _runtime_parts_safe(runtime)
        if not isinstance(parts, tuple) or type(runtime) is not ProjectOutboxRuntime:
            return False
        delegate = parts[0]
        object.__setattr__(runtime, "_cleanup_called", True)
        delegate.cleanup()
        return True
    except Exception:
        return False


def _delegate_is_valid_safe(value: object) -> TypeGuard[ProjectOutboxHarnessDelegate]:
    try:
        return isinstance(value, ProjectOutboxHarnessDelegate)
    except Exception:
        return False


def _correlation_id(nonce: str, suffix: str) -> str:
    return f"outbox-{nonce}-{suffix}"


def _nonce_is_valid_safe(value: object) -> bool:
    try:
        return (
            type(value) is str
            and len(value) == 32
            and all(character in "0123456789abcdef" for character in value)
        )
    except Exception:
        return False


def _raise_invalid() -> NoReturn:
    raise ProjectOutboxRuntimeError("project outbox runtime is invalid") from None


def _raise_assertion() -> NoReturn:
    raise ProjectOutboxAssertionError("project outbox contract was not satisfied") from None
