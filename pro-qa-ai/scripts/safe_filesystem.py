#!/usr/bin/env python3
"""Fail-closed filesystem and bounded-text helpers.

The helpers in this module deliberately use ``lstat`` before resolving paths.
That order is important on Windows, where junctions and other reparse points
are not necessarily reported by ``Path.is_symlink()``.
"""

from __future__ import annotations

import hashlib
import os
import stat
from contextlib import contextmanager
from dataclasses import dataclass
from pathlib import Path
from typing import BinaryIO, Callable, Dict, Iterator, Mapping, Optional, Sequence, Tuple


BUFFER_SIZE = 64 * 1024


class FileSafetyError(ValueError):
    """A filesystem boundary or bounded-text policy was violated."""


@dataclass(frozen=True)
class TextPolicy:
    allowed_text_extensions: Tuple[str, ...]
    allowed_text_names: Tuple[str, ...]
    max_file_bytes: int
    max_total_bytes: int
    max_file_count: int


@dataclass(frozen=True)
class InspectedText:
    text: str
    data: bytes
    size: int
    sha256: str
    modified_at_ns: int


@dataclass(frozen=True)
class TreeEntry:
    """Immutable metadata captured while a :class:`SafeTree` is enumerated."""

    relative_path: str
    path: Path
    metadata: os.stat_result
    identity: Tuple[int, ...]
    is_directory: bool


def parse_text_policy(config: Mapping[str, object], field: str = "file_policy") -> TextPolicy:
    """Validate a text/size policy mapping without accepting booleans as ints."""

    raw = config.get(field)
    if not isinstance(raw, Mapping):
        raise FileSafetyError("%s must be an object" % field)

    def string_tuple(name: str, *, extension: bool) -> Tuple[str, ...]:
        values = raw.get(name)
        if not isinstance(values, list) or not values:
            raise FileSafetyError("%s.%s must be a non-empty array" % (field, name))
        normalized = []
        for index, value in enumerate(values):
            if not isinstance(value, str) or not value or "\x00" in value:
                raise FileSafetyError(
                    "%s.%s[%d] must be a non-empty string" % (field, name, index)
                )
            item = value.lower() if extension else value
            if extension and (not item.startswith(".") or "/" in item or "\\" in item):
                raise FileSafetyError(
                    "%s.%s[%d] must be a dot-prefixed extension"
                    % (field, name, index)
                )
            if not extension and ("/" in item or "\\" in item):
                raise FileSafetyError(
                    "%s.%s[%d] must be a basename" % (field, name, index)
                )
            if item in normalized:
                raise FileSafetyError(
                    "%s.%s[%d] duplicates an earlier value" % (field, name, index)
                )
            normalized.append(item)
        return tuple(normalized)

    def positive_int(name: str) -> int:
        value = raw.get(name)
        if type(value) is not int or value <= 0:
            raise FileSafetyError("%s.%s must be a positive integer" % (field, name))
        return value

    policy = TextPolicy(
        allowed_text_extensions=string_tuple(
            "allowed_text_extensions", extension=True
        ),
        allowed_text_names=string_tuple("allowed_text_names", extension=False),
        max_file_bytes=positive_int("max_file_bytes"),
        max_total_bytes=positive_int("max_total_bytes"),
        max_file_count=positive_int("max_file_count"),
    )
    if policy.max_total_bytes < policy.max_file_bytes:
        raise FileSafetyError(
            "%s.max_total_bytes must be at least max_file_bytes" % field
        )
    return policy


def text_path_allowed(path: Path, policy: TextPolicy) -> bool:
    """Return whether a basename or extension is explicitly text-approved."""

    name = path.name
    if name in policy.allowed_text_names:
        return True
    lower_name = name.lower()
    return any(lower_name.endswith(extension) for extension in policy.allowed_text_extensions)


def _windows_runtime(windows: Optional[bool]) -> bool:
    return os.name == "nt" if windows is None else windows


def is_link_like(path: Path, *, windows: Optional[bool] = None) -> bool:
    """Detect symlinks and Windows reparse points using fail-closed metadata.

    ``windows`` exists only to make the Windows attribute branch unit-testable
    on hosts where creating symlinks or junctions needs extra privileges.
    """

    try:
        metadata = os.lstat(path)
    except OSError as exc:
        raise FileSafetyError(
            "cannot inspect filesystem metadata (%s)" % type(exc).__name__
        ) from exc
    return _metadata_is_link_like(metadata, windows=windows)


def _metadata_is_link_like(
    metadata: os.stat_result, *, windows: Optional[bool] = None
) -> bool:
    """Classify already-captured metadata without a second path lookup."""

    if stat.S_ISLNK(metadata.st_mode):
        return True
    if _windows_runtime(windows):
        marker = getattr(stat, "FILE_ATTRIBUTE_REPARSE_POINT", None)
        attributes = getattr(metadata, "st_file_attributes", None)
        if marker is None or type(attributes) is not int:
            raise FileSafetyError("cannot inspect Windows reparse attributes")
        return bool(attributes & marker)
    return False


def lexical_absolute(path: Path) -> Path:
    """Make a path absolute without resolving links."""

    expanded = path.expanduser()
    if "\x00" in os.fspath(expanded):
        raise FileSafetyError("path contains an invalid character")
    return Path(os.path.abspath(os.fspath(expanded)))


def _is_unc(path: Path) -> bool:
    rendered = os.fspath(path)
    return rendered.startswith("\\\\") or rendered.startswith("//")


def assert_no_link_components(
    path: Path,
    *,
    allow_missing_tail: bool,
    windows: Optional[bool] = None,
) -> None:
    """Reject link-like existing components from the filesystem anchor down."""

    absolute = lexical_absolute(path)
    if _is_unc(absolute):
        raise FileSafetyError("UNC/network paths are not permitted")
    parts = absolute.parts
    if not parts:
        raise FileSafetyError("path has no filesystem components")
    current = Path(parts[0])
    start = 1
    if not absolute.anchor:
        current = Path()
        start = 0
    for part in parts[start:]:
        current = current / part
        try:
            linked = is_link_like(current, windows=windows)
        except FileSafetyError as exc:
            cause = exc.__cause__
            if allow_missing_tail and isinstance(cause, FileNotFoundError):
                return
            raise
        if linked:
            raise FileSafetyError("path contains a symbolic link or reparse point")


def _is_relative_to(path: Path, root: Path) -> bool:
    try:
        common = os.path.commonpath(
            [os.path.normcase(os.fspath(path)), os.path.normcase(os.fspath(root))]
        )
    except ValueError:
        return False
    return common == os.path.normcase(os.fspath(root))


def resolve_directory_root(path: Path, label: str = "root") -> Path:
    """Resolve an existing directory only after checking every lexical component."""

    absolute = lexical_absolute(path)
    assert_no_link_components(absolute, allow_missing_tail=False)
    try:
        metadata = os.lstat(absolute)
        resolved = absolute.resolve(strict=True)
    except (OSError, RuntimeError) as exc:
        raise FileSafetyError("%s is unavailable (%s)" % (label, type(exc).__name__)) from exc
    if not stat.S_ISDIR(metadata.st_mode):
        raise FileSafetyError("%s must be a directory" % label)
    assert_no_link_components(resolved, allow_missing_tail=False)
    return resolved


def resolve_new_or_empty_directory(path: Path, label: str = "output") -> Path:
    """Resolve a prospective output directory without following link components."""

    absolute = lexical_absolute(path)
    assert_no_link_components(absolute, allow_missing_tail=True)
    try:
        if absolute.exists():
            if is_link_like(absolute):
                raise FileSafetyError(
                    "%s must not be a symbolic link or reparse point" % label
                )
            metadata = os.lstat(absolute)
            if not stat.S_ISDIR(metadata.st_mode):
                raise FileSafetyError("%s must be a directory" % label)
            resolved = absolute.resolve(strict=True)
        else:
            parent = absolute.parent.resolve(strict=True)
            assert_no_link_components(parent, allow_missing_tail=False)
            resolved = parent / absolute.name
    except (OSError, RuntimeError) as exc:
        raise FileSafetyError(
            "%s path is unavailable (%s)" % (label, type(exc).__name__)
        ) from exc
    return resolved


def enforce_containment(
    root: Path,
    path: Path,
    *,
    must_exist: bool,
    allow_directory: bool = False,
) -> Path:
    """Recheck components, resolution, type, and containment immediately before I/O."""

    safe_root = resolve_directory_root(root, "declared root")
    candidate = lexical_absolute(path)
    if not _is_relative_to(candidate, safe_root):
        raise FileSafetyError("path escapes its declared root")
    assert_no_link_components(candidate, allow_missing_tail=not must_exist)
    try:
        if must_exist:
            resolved = candidate.resolve(strict=True)
            metadata = os.lstat(candidate)
            if not _is_relative_to(resolved, safe_root):
                raise FileSafetyError("resolved path escapes its declared root")
            if allow_directory:
                if not (stat.S_ISREG(metadata.st_mode) or stat.S_ISDIR(metadata.st_mode)):
                    raise FileSafetyError("path is not a regular file or directory")
            elif not stat.S_ISREG(metadata.st_mode):
                raise FileSafetyError("path is not a regular file")
        else:
            parent = candidate.parent.resolve(strict=True)
            if not _is_relative_to(parent, safe_root):
                raise FileSafetyError("output parent escapes its declared root")
            assert_no_link_components(parent, allow_missing_tail=False)
            resolved = parent / candidate.name
    except (OSError, RuntimeError) as exc:
        raise FileSafetyError(
            "cannot resolve contained path (%s)" % type(exc).__name__
        ) from exc
    return resolved


def _identity(metadata: os.stat_result, *, mutable_size: bool = False) -> Tuple[int, ...]:
    values = [
        metadata.st_dev,
        metadata.st_ino,
        metadata.st_mode,
        getattr(metadata, "st_file_attributes", 0),
    ]
    if not mutable_size:
        values.extend(
            [
                metadata.st_size,
                metadata.st_mtime_ns,
                getattr(metadata, "st_ctime_ns", 0),
            ]
        )
    return tuple(values)


@contextmanager
def safe_open_read(root: Path, path: Path) -> Iterator[BinaryIO]:
    """Open a regular file by descriptor and verify identity before and after use."""

    candidate = enforce_containment(root, path, must_exist=True)
    try:
        before = os.lstat(candidate)
        flags = os.O_RDONLY | getattr(os, "O_BINARY", 0) | getattr(os, "O_NOFOLLOW", 0)
        descriptor = os.open(candidate, flags)
    except OSError as exc:
        raise FileSafetyError("cannot open contained source (%s)" % type(exc).__name__) from exc
    handle: Optional[BinaryIO] = None
    try:
        opened = os.fstat(descriptor)
        if not stat.S_ISREG(opened.st_mode) or _identity(before) != _identity(opened):
            raise FileSafetyError("source identity changed before open")
        enforce_containment(root, candidate, must_exist=True)
        handle = os.fdopen(descriptor, "rb", closefd=True)
        descriptor = -1
        try:
            yield handle
        except BaseException:
            raise
        else:
            after_fd = os.fstat(handle.fileno())
            after_path = os.lstat(candidate)
            enforce_containment(root, candidate, must_exist=True)
            if _identity(opened) != _identity(after_fd) or _identity(opened) != _identity(
                after_path
            ):
                raise FileSafetyError("source file changed during read")
    finally:
        if handle is not None:
            handle.close()
        elif descriptor >= 0:
            os.close(descriptor)


def safe_mkdir_parents(root: Path, directory: Path) -> None:
    """Create missing directories one component at a time under a safe root."""

    safe_root = resolve_directory_root(root, "output root")
    candidate = lexical_absolute(directory)
    if not _is_relative_to(candidate, safe_root):
        raise FileSafetyError("output directory escapes its declared root")
    relative = candidate.relative_to(safe_root)
    current = safe_root
    for part in relative.parts:
        current = current / part
        try:
            os.mkdir(current)
        except FileExistsError:
            pass
        except OSError as exc:
            raise FileSafetyError(
                "cannot create output directory (%s)" % type(exc).__name__
            ) from exc
        checked = enforce_containment(
            safe_root, current, must_exist=True, allow_directory=True
        )
        if not checked.is_dir():
            raise FileSafetyError("output component is not a directory")


@contextmanager
def safe_open_write(root: Path, path: Path) -> Iterator[BinaryIO]:
    """Create a new regular file exclusively below an already-safe output root."""

    safe_root = resolve_directory_root(root, "output root")
    candidate = lexical_absolute(path)
    safe_mkdir_parents(safe_root, candidate.parent)
    resolved = enforce_containment(safe_root, candidate, must_exist=False)
    flags = (
        os.O_WRONLY
        | os.O_CREAT
        | os.O_EXCL
        | getattr(os, "O_BINARY", 0)
        | getattr(os, "O_NOFOLLOW", 0)
    )
    try:
        descriptor = os.open(resolved, flags, 0o600)
    except OSError as exc:
        raise FileSafetyError(
            "cannot create contained output (%s)" % type(exc).__name__
        ) from exc
    handle: Optional[BinaryIO] = None
    try:
        opened = os.fstat(descriptor)
        after_path = os.lstat(resolved)
        enforce_containment(safe_root, resolved, must_exist=True)
        if not stat.S_ISREG(opened.st_mode) or _identity(
            opened, mutable_size=True
        ) != _identity(after_path, mutable_size=True):
            raise FileSafetyError("output identity changed during creation")
        handle = os.fdopen(descriptor, "wb", closefd=True)
        descriptor = -1
        yield handle
        handle.flush()
        os.fsync(handle.fileno())
        after_fd = os.fstat(handle.fileno())
        after_path = os.lstat(resolved)
        enforce_containment(safe_root, resolved, must_exist=True)
        if _identity(after_fd, mutable_size=True) != _identity(
            after_path, mutable_size=True
        ):
            raise FileSafetyError("output identity changed during write")
    finally:
        if handle is not None:
            handle.close()
        elif descriptor >= 0:
            os.close(descriptor)


def inspect_text_file(
    root: Path,
    path: Path,
    relative_path: str,
    policy: TextPolicy,
) -> InspectedText:
    """Boundedly read, UTF-8 decode, and hash one approved text file."""

    if not text_path_allowed(Path(relative_path), policy):
        raise FileSafetyError("unapproved text file type: %s" % relative_path)
    digest = hashlib.sha256()
    chunks = []
    total = 0
    with safe_open_read(root, path) as source:
        while True:
            chunk = source.read(min(BUFFER_SIZE, policy.max_file_bytes + 1 - total))
            if not chunk:
                break
            if b"\x00" in chunk:
                raise FileSafetyError("binary/NUL content is not permitted: %s" % relative_path)
            chunks.append(chunk)
            digest.update(chunk)
            total += len(chunk)
            if total > policy.max_file_bytes:
                raise FileSafetyError("file exceeds byte limit: %s" % relative_path)
    data = b"".join(chunks)
    try:
        text = data.decode("utf-8-sig")
    except UnicodeDecodeError as exc:
        raise FileSafetyError("file is not valid UTF-8 text: %s" % relative_path) from exc
    try:
        modified_at_ns = os.lstat(path).st_mtime_ns
    except OSError as exc:
        raise FileSafetyError(
            "cannot inspect source timestamp (%s)" % type(exc).__name__
        ) from exc
    enforce_containment(root, path, must_exist=True)
    return InspectedText(text, data, total, digest.hexdigest(), modified_at_ns)


class SafeTree:
    """Bounded, fail-closed snapshot and inspection cache for one directory tree.

    Enumeration captures canonical root-relative paths and ``lstat`` identities.
    Later reads must still match that snapshot before, during, and after the
    descriptor-based read.  Callers never need to perform their own filesystem
    probes, which prevents a check from accidentally bypassing the boundary.
    """

    def __init__(
        self,
        root: Path,
        policy: TextPolicy,
        *,
        exclude: Optional[Callable[[str, bool], bool]] = None,
    ) -> None:
        self.root = resolve_directory_root(root, "inspection root")
        self.policy = policy
        self._exclude = exclude or (lambda _relative, _is_directory: False)
        self._files: Dict[str, TreeEntry] = {}
        self._directories: Dict[str, TreeEntry] = {}
        self._inspection_cache: Dict[str, InspectedText] = {}
        self._total_bytes = 0
        self._snapshot()

    @staticmethod
    def _key(relative_path: str) -> str:
        return os.path.normcase(relative_path.replace("/", os.sep))

    @staticmethod
    def _entry_identity(metadata: os.stat_result) -> Tuple[int, ...]:
        return _identity(metadata)

    def _capture(self, path: Path, relative_path: str) -> TreeEntry:
        try:
            metadata = os.lstat(path)
        except OSError as exc:
            raise FileSafetyError(
                "cannot inspect tree entry %s (%s)"
                % (relative_path or ".", type(exc).__name__)
            ) from exc
        if _metadata_is_link_like(metadata):
            raise FileSafetyError(
                "tree entry is a symbolic link or reparse point: %s"
                % (relative_path or ".")
            )
        is_directory = stat.S_ISDIR(metadata.st_mode)
        if not is_directory and not stat.S_ISREG(metadata.st_mode):
            raise FileSafetyError(
                "tree entry is not a regular file or directory: %s"
                % (relative_path or ".")
            )
        return TreeEntry(
            relative_path,
            path,
            metadata,
            self._entry_identity(metadata),
            is_directory,
        )

    def _verify_entry(self, entry: TreeEntry) -> os.stat_result:
        try:
            metadata = os.lstat(entry.path)
        except OSError as exc:
            raise FileSafetyError(
                "snapshotted entry is unavailable: %s (%s)"
                % (entry.relative_path or ".", type(exc).__name__)
            ) from exc
        if _metadata_is_link_like(metadata):
            raise FileSafetyError(
                "snapshotted entry became a symbolic link or reparse point: %s"
                % (entry.relative_path or ".")
            )
        if self._entry_identity(metadata) != entry.identity:
            raise FileSafetyError(
                "snapshotted entry identity changed: %s"
                % (entry.relative_path or ".")
            )
        enforce_containment(
            self.root,
            entry.path,
            must_exist=True,
            allow_directory=entry.is_directory,
        )
        return metadata

    def _snapshot(self) -> None:
        root_entry = self._capture(self.root, "")
        if not root_entry.is_directory:
            raise FileSafetyError("inspection root must be a directory")
        self._directories[self._key("")] = root_entry
        pending = [root_entry]
        while pending:
            directory = pending.pop()
            self._verify_entry(directory)
            try:
                with os.scandir(directory.path) as iterator:
                    names = sorted(entry.name for entry in iterator)
            except OSError as exc:
                raise FileSafetyError(
                    "cannot enumerate directory %s (%s)"
                    % (directory.relative_path or ".", type(exc).__name__)
                ) from exc
            self._verify_entry(directory)

            children = []
            for name in names:
                if (
                    not name
                    or name in {".", ".."}
                    or "\x00" in name
                    or "/" in name
                    or "\\" in name
                ):
                    raise FileSafetyError("directory contains an invalid entry name")
                relative = (
                    name
                    if not directory.relative_path
                    else directory.relative_path + "/" + name
                )
                candidate = directory.path / name
                child = self._capture(candidate, relative)
                if self._exclude(relative, child.is_directory):
                    continue
                key = self._key(relative)
                if key in self._files or key in self._directories:
                    raise FileSafetyError(
                        "tree contains colliding canonical paths: %s" % relative
                    )
                if child.is_directory:
                    self._directories[key] = child
                    children.append(child)
                    continue
                if child.metadata.st_size < 0:
                    raise FileSafetyError("tree entry reports an invalid size: %s" % relative)
                count_exceeded = len(self._files) >= self.policy.max_file_count
                total_exceeded = (
                    self._total_bytes + child.metadata.st_size
                    > self.policy.max_total_bytes
                )
                if count_exceeded or total_exceeded:
                    reasons = []
                    if count_exceeded:
                        reasons.append("max file count")
                    if total_exceeded:
                        reasons.append("max total bytes")
                    raise FileSafetyError("tree exceeds " + " and ".join(reasons))
                self._files[key] = child
                self._total_bytes += child.metadata.st_size
            self._verify_entry(directory)
            pending.extend(reversed(children))

    def _relative_lookup(self, path: Path | str) -> Tuple[str, str]:
        raw = Path(path)
        candidate = lexical_absolute(raw if raw.is_absolute() else self.root / raw)
        if not _is_relative_to(candidate, self.root):
            raise FileSafetyError("path escapes its declared root")
        relative_os = os.path.relpath(os.fspath(candidate), os.fspath(self.root))
        relative = "" if relative_os == "." else Path(relative_os).as_posix()
        parts = Path(relative).parts
        if any(part in {"", ".", ".."} for part in parts):
            raise FileSafetyError("path is not canonical below its declared root")
        return relative, self._key(relative)

    @property
    def file_count(self) -> int:
        return len(self._files)

    @property
    def total_bytes(self) -> int:
        return self._total_bytes

    @property
    def inspection_cache(self) -> Mapping[str, InspectedText]:
        return dict(self._inspection_cache)

    def relative_path(self, path: Path | str) -> str:
        relative, _ = self._relative_lookup(path)
        return relative

    def files(self) -> Tuple[Path, ...]:
        return tuple(
            entry.path
            for entry in sorted(
                self._files.values(), key=lambda item: self._key(item.relative_path)
            )
        )

    def directories(self) -> Tuple[Path, ...]:
        return tuple(
            entry.path
            for entry in sorted(
                self._directories.values(), key=lambda item: self._key(item.relative_path)
            )
        )

    def is_file(self, path: Path | str) -> bool:
        _, key = self._relative_lookup(path)
        return key in self._files

    def is_dir(self, path: Path | str) -> bool:
        _, key = self._relative_lookup(path)
        return key in self._directories

    def exists(self, path: Path | str) -> bool:
        _, key = self._relative_lookup(path)
        return key in self._files or key in self._directories

    def files_in(
        self,
        directory: Path | str,
        *,
        suffix: Optional[str] = None,
        recursive: bool = False,
    ) -> Tuple[Path, ...]:
        relative, key = self._relative_lookup(directory)
        if key not in self._directories:
            return ()
        prefix = relative + "/" if relative else ""
        output = []
        for entry in self._files.values():
            if not entry.relative_path.startswith(prefix):
                continue
            remainder = entry.relative_path[len(prefix) :]
            if not recursive and "/" in remainder:
                continue
            if suffix is not None and not entry.path.name.lower().endswith(suffix.lower()):
                continue
            output.append(entry.path)
        return tuple(sorted(output, key=lambda item: self._key(self.relative_path(item))))

    def directories_in(
        self, directory: Path | str, *, recursive: bool = False
    ) -> Tuple[Path, ...]:
        relative, key = self._relative_lookup(directory)
        if key not in self._directories:
            return ()
        prefix = relative + "/" if relative else ""
        output = []
        for entry in self._directories.values():
            if entry.relative_path == relative or not entry.relative_path.startswith(prefix):
                continue
            remainder = entry.relative_path[len(prefix) :]
            if not recursive and "/" in remainder:
                continue
            output.append(entry.path)
        return tuple(sorted(output, key=lambda item: self._key(self.relative_path(item))))

    def metadata(self, path: Path | str) -> os.stat_result:
        _, key = self._relative_lookup(path)
        entry = self._files.get(key) or self._directories.get(key)
        if entry is None:
            raise FileSafetyError("path was not present in the safe snapshot")
        return entry.metadata

    def inspect_text(self, path: Path | str) -> InspectedText:
        relative, key = self._relative_lookup(path)
        cached = self._inspection_cache.get(key)
        if cached is not None:
            return cached
        entry = self._files.get(key)
        if entry is None:
            raise FileSafetyError("file was not present in the safe snapshot: %s" % relative)
        if not text_path_allowed(Path(relative), self.policy):
            raise FileSafetyError("unapproved text file type: %s" % relative)
        self._verify_entry(entry)

        digest = hashlib.sha256()
        chunks = []
        total = 0
        with safe_open_read(self.root, entry.path) as source:
            while True:
                remaining = self.policy.max_file_bytes + 1 - total
                chunk = source.read(min(BUFFER_SIZE, remaining))
                if not chunk:
                    break
                if any(
                    byte == 0 or byte < 9 or 13 < byte < 32 or byte == 127
                    for byte in chunk
                ):
                    raise FileSafetyError(
                        "binary/control content is not permitted: %s" % relative
                    )
                chunks.append(chunk)
                digest.update(chunk)
                total += len(chunk)
                if total > self.policy.max_file_bytes:
                    raise FileSafetyError("file exceeds byte limit: %s" % relative)
        self._verify_entry(entry)
        data = b"".join(chunks)
        try:
            text = data.decode("utf-8-sig")
        except UnicodeDecodeError as exc:
            raise FileSafetyError(
                "file is not valid UTF-8 text: %s" % relative
            ) from exc
        inspected = InspectedText(
            text,
            data,
            total,
            digest.hexdigest(),
            entry.metadata.st_mtime_ns,
        )
        self._inspection_cache[key] = inspected
        return inspected
