from __future__ import annotations

import re
from typing import Any


CASE_ID_PATTERN = r"^TC-[A-Z][A-Z0-9]*(?:-[A-Z][A-Z0-9]*)*-[0-9]{4}$"
LAYER_MARKERS = ("contract", "api", "integration", "e2e")
BLOCKED_MARKERS = ("skip", "skipif", "xfail")

_CASE_ID_RE = re.compile(CASE_ID_PATTERN)


class CollectionError(ValueError):
    pass


def case_id_from_item(item: Any) -> str:
    markers = _markers(item, "case_id")
    if len(markers) != 1:
        raise CollectionError("case marker is invalid")
    marker = markers[0]
    if marker.kwargs or len(marker.args) != 1:
        raise CollectionError("case marker is invalid")
    case_id = marker.args[0]
    if type(case_id) is not str or _CASE_ID_RE.fullmatch(case_id) is None:
        raise CollectionError("case marker is invalid")
    return case_id


def layer_from_item(item: Any) -> str:
    markers = [marker for layer in LAYER_MARKERS for marker in _markers(item, layer)]
    if len(markers) != 1:
        raise CollectionError("layer marker is invalid")
    marker = markers[0]
    if marker.args or marker.kwargs or marker.name not in LAYER_MARKERS:
        raise CollectionError("layer marker is invalid")
    return marker.name


def reject_blocked_markers(item: Any) -> None:
    if any(_markers(item, name) for name in BLOCKED_MARKERS):
        raise CollectionError("collection marker is not allowed")


def _markers(item: Any, name: str) -> tuple[Any, ...]:
    try:
        return tuple(item.iter_markers(name=name))
    except Exception:
        raise CollectionError("collection markers are invalid") from None
