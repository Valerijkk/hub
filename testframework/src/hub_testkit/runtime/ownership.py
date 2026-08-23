from __future__ import annotations

from collections.abc import Iterable, Mapping
from dataclasses import dataclass
from types import MappingProxyType

from .ids import RuntimeIds


class OwnershipError(ValueError):
    pass


_MANDATORY_LABELS = (
    "hub.qa.task",
    "hub.qa.run",
    "hub.qa.worker",
    "hub.qa.nonce",
)


@dataclass(frozen=True, slots=True)
class OwnedResource:
    kind: str
    resource_id: str
    name: str
    nonce: str
    labels: Mapping[str, str]

    def __post_init__(self) -> None:
        if any(type(value) is not str or not value for value in self._identity_fields()):
            raise OwnershipError("owned resource identity is invalid")
        if not isinstance(self.labels, Mapping):
            raise OwnershipError("owned resource labels are invalid")
        labels = dict(self.labels)
        if any(type(key) is not str or type(value) is not str for key, value in labels.items()):
            raise OwnershipError("owned resource labels are invalid")
        object.__setattr__(self, "labels", MappingProxyType(labels))

    def _identity_fields(self) -> tuple[str, str, str, str]:
        return (self.kind, self.resource_id, self.name, self.nonce)


class OwnershipLedger:
    def __init__(self, ids: RuntimeIds) -> None:
        if not isinstance(ids, RuntimeIds):
            raise OwnershipError("runtime IDs are invalid")
        self._ids = ids
        self._resources: dict[tuple[str, str], OwnedResource] = {}
        self._names: set[str] = set()

    def register(self, resource: OwnedResource) -> None:
        self._assert_owned(resource)
        identity = (resource.kind, resource.resource_id)
        if identity in self._resources or resource.name in self._names:
            raise OwnershipError("owned resource is already registered")
        self._resources[identity] = resource
        self._names.add(resource.name)

    def snapshot(self) -> tuple[OwnedResource, ...]:
        return tuple(self._resources.values())

    def assert_cleanup_scope(self, resources: Iterable[OwnedResource]) -> tuple[OwnedResource, ...]:
        if isinstance(resources, (str, bytes)):
            raise OwnershipError("cleanup scope is invalid")
        try:
            candidates = tuple(resources)
        except TypeError:
            raise OwnershipError("cleanup scope is invalid") from None
        identities: set[tuple[str, str]] = set()
        names: set[str] = set()
        for candidate in candidates:
            if not isinstance(candidate, OwnedResource):
                raise OwnershipError("cleanup scope is invalid")
            identity = (candidate.kind, candidate.resource_id)
            registered = self._resources.get(identity)
            if registered is None or registered != candidate:
                raise OwnershipError("cleanup scope is not exactly owned")
            if identity in identities or candidate.name in names:
                raise OwnershipError("cleanup scope contains duplicates")
            identities.add(identity)
            names.add(candidate.name)
        return candidates

    def mark_cleaned(self, resources: Iterable[OwnedResource]) -> None:
        for resource in self.assert_cleanup_scope(resources):
            identity = (resource.kind, resource.resource_id)
            del self._resources[identity]
            self._names.remove(resource.name)

    def _assert_owned(self, resource: OwnedResource) -> None:
        if not isinstance(resource, OwnedResource):
            raise OwnershipError("owned resource is invalid")
        expected_labels = dict(
            zip(
                _MANDATORY_LABELS,
                (
                    self._ids.task_id,
                    self._ids.run_id,
                    self._ids.worker_id,
                    self._ids.nonce,
                ),
                strict=True,
            )
        )
        if resource.nonce != self._ids.nonce or dict(resource.labels) != expected_labels:
            raise OwnershipError("owned resource does not belong to this runtime")
