from __future__ import annotations

import ctypes
import os
import stat
from pathlib import Path, PurePosixPath, PureWindowsPath
from typing import Any


SafePathError = type("SafePathError", (RuntimeError,), {})

O_NOFOLLOW = getattr(os, "O_NOFOLLOW", 0)
_O_DIRECTORY = getattr(os, "O_DIRECTORY", 0)
_O_CLOEXEC = getattr(os, "O_CLOEXEC", 0)
_O_BINARY = getattr(os, "O_BINARY", 0)
_O_NONBLOCK = getattr(os, "O_NONBLOCK", 0)
_AT_FDCWD = -100
_FILE_ATTRIBUTE_DIRECTORY = 0x00000010
_FILE_ATTRIBUTE_REPARSE_POINT = 0x00000400
_FILE_ATTRIBUTE_NORMAL = 0x00000080
FILE_FLAG_OPEN_REPARSE_POINT = 0x00200000
_FILE_FLAG_BACKUP_SEMANTICS = 0x02000000
_GENERIC_READ = 0x80000000
_GENERIC_WRITE = 0x40000000
_FILE_SHARE_READ = 0x00000001
_FILE_SHARE_WRITE = 0x00000002
_FILE_SHARE_DELETE = 0x00000004
_OPEN_EXISTING = 3
_CREATE_NEW = 1
_MOVEFILE_WRITE_THROUGH = 0x00000008
_FILE_TYPE_DISK = 0x0001
_FILE_READ_DATA = 0x0001
_FILE_WRITE_DATA = 0x0002
_FILE_READ_ATTRIBUTES = 0x0080
_FILE_WRITE_ATTRIBUTES = 0x0100
_DELETE = 0x00010000
_SYNCHRONIZE = 0x00100000
_FILE_OPEN = 1
_FILE_CREATE = 2
_FILE_DIRECTORY_FILE = 0x00000001
_FILE_NON_DIRECTORY_FILE = 0x00000040
_FILE_SYNCHRONOUS_IO_NONALERT = 0x00000020
_OBJ_CASE_INSENSITIVE = 0x00000040
_FILE_RENAME_INFO = 3
_FILE_DISPOSITION_INFO = 4
_RENAME_NOREPLACE = 0x00000001


def _is_windows(ops: Any | None) -> bool:
    return bool(getattr(ops, "windows", os.name == "nt"))


def _relative_parts(relative: str, windows: bool) -> tuple[str, ...]:
    if type(relative) is not str or not relative:
        raise SafePathError("relative path is required")
    lexical = PureWindowsPath(relative) if windows else PurePosixPath(relative)
    windows_lexical = PureWindowsPath(relative)
    if lexical.is_absolute() or windows_lexical.is_absolute() or windows_lexical.drive:
        raise SafePathError("absolute path is denied")
    if not windows and "\\" in relative:
        raise SafePathError("ambiguous path separator")
    parts = tuple(lexical.parts)
    if not parts or any(part in ("", ".", "..") for part in parts):
        raise SafePathError("parent traversal is denied")
    if windows and any(":" in part for part in parts):
        raise SafePathError("Windows alternate data streams are denied")
    return parts


def _close(ops: Any | None, descriptor: int) -> None:
    try:
        if ops is None:
            if os.name == "nt":
                ctypes.windll.kernel32.CloseHandle(ctypes.c_void_p(descriptor))
            else:
                os.close(descriptor)
        else:
            ops.close(descriptor)
    except Exception:
        pass


def _fstat(ops: Any | None, descriptor: int) -> Any:
    try:
        if ops is not None:
            return ops.fstat(descriptor)
        if os.name == "nt":
            return _windows_fstat(descriptor)
        return os.fstat(descriptor)
    except Exception as exc:
        raise SafePathError("descriptor inspection failed") from exc


def _stat_values(value: Any) -> dict[str, int | bool]:
    try:
        if isinstance(value, dict):
            return {
                "device": int(value["device"]),
                "inode": int(value["inode"]),
                "size": int(value["size"]),
                "regular": bool(value["regular"]),
                "nlink": int(value["nlink"]),
                "reparse": bool(value.get("reparse", False)),
                "directory": bool(value.get("directory", False)),
                "disk": bool(value.get("disk", True)),
            }
        return {
            "device": int(value.st_dev),
            "inode": int(value.st_ino),
            "size": int(value.st_size),
            "regular": stat.S_ISREG(value.st_mode),
            "nlink": int(value.st_nlink),
            "reparse": False,
            "directory": stat.S_ISDIR(value.st_mode),
            "disk": True,
        }
    except Exception as exc:
        raise SafePathError("unusable descriptor metadata") from exc


def _require_regular(snapshot: Any) -> dict[str, int | bool]:
    values = _stat_values(snapshot)
    if (
        values["regular"] is not True
        or values["nlink"] != 1
        or values["reparse"] is not False
        or values["disk"] is not True
        or values["size"] < 0
    ):
        raise SafePathError("path is not a stable regular file")
    return values


def _require_stable(before: Any, after: Any) -> None:
    first = _require_regular(before)
    second = _require_regular(after)
    if any(first[key] != second[key] for key in ("device", "inode", "size")):
        raise SafePathError("file identity changed while opened")


def _require_directory(snapshot: Any) -> dict[str, int | bool]:
    values = _stat_values(snapshot)
    if values["reparse"] is True or values["directory"] is not True:
        raise SafePathError("unsafe directory")
    return values


def _posix_openat(directory_fd: int, name: str, flags: int, mode: int = 0) -> int:
    try:
        library = ctypes.CDLL(None, use_errno=True)
        call = library.openat
        encoded = os.fsencode(name)
        descriptor = call(directory_fd, encoded, flags, mode) if mode else call(directory_fd, encoded, flags)
        if descriptor < 0:
            raise OSError(ctypes.get_errno(), "openat failed", name)
        return int(descriptor)
    except Exception as exc:
        raise SafePathError("anchored openat failed") from exc


def _posix_open_directory(root: Path) -> tuple[int, int]:
    if O_NOFOLLOW == 0 or not root.is_absolute():
        raise SafePathError("anchored POSIX directory access is unavailable")
    flags = os.O_RDONLY | _O_DIRECTORY | O_NOFOLLOW | _O_CLOEXEC | _O_BINARY
    descriptor = _posix_openat(_AT_FDCWD, os.sep, flags)
    try:
        _require_directory(_fstat(None, descriptor))
        for part in root.parts[1:]:
            child = _posix_openat(descriptor, part, flags)
            try:
                _require_directory(_fstat(None, child))
            except Exception:
                _close(None, child)
                raise
            _close(None, descriptor)
            descriptor = child
        root_values = _require_directory(_fstat(None, descriptor))
        return descriptor, int(root_values["device"])
    except Exception:
        _close(None, descriptor)
        raise


def _openat(ops: Any | None, directory_fd: int, name: str, flags: int) -> int:
    if ops is None:
        return _posix_openat(directory_fd, name, flags)
    try:
        return int(ops.openat(directory_fd, name, flags))
    except Exception as exc:
        raise SafePathError("anchored openat failed") from exc


def _open_posix_leaf(root: Path, parts: tuple[str, ...], ops: Any | None) -> tuple[int, int | None]:
    root_is_owned = False
    if ops is None:
        root_fd, root_device = _posix_open_directory(root)
    else:
        open_root = getattr(ops, "open_root", None)
        if callable(open_root):
            root_fd = int(open_root(root))
            try:
                root_values = _require_directory(_fstat(ops, root_fd))
                root_device = int(root_values["device"])
            except Exception:
                _close(ops, root_fd)
                raise
            root_is_owned = True
        else:
            raise SafePathError("rooted POSIX ops are required")
    current = root_fd
    native = ops is None
    try:
        for ordinal, part in enumerate(parts):
            child = -1
            flags = os.O_RDONLY | O_NOFOLLOW | _O_CLOEXEC | _O_BINARY | _O_NONBLOCK
            if ordinal + 1 != len(parts):
                flags |= _O_DIRECTORY
            child = _openat(ops, current, part, flags)
            try:
                if native or root_is_owned:
                    child_values = _stat_values(_fstat(ops, child))
                    if child_values["device"] != root_device:
                        raise SafePathError("path crossed a device boundary")
                    if ordinal + 1 != len(parts):
                        _require_directory(child_values)
            except Exception:
                _close(ops, child)
                child = -1
                raise
            if native or root_is_owned or current != root_fd:
                _close(ops, current)
            current = child
        return current, root_device
    except Exception:
        _close(ops, current)
        raise


def _windows_fstat(descriptor: int) -> dict[str, int | bool]:
    try:
        field = ctypes.c_uint32
        info_type = type(
            "_ByHandleFileInformation",
            (ctypes.Structure,),
            {
                "_fields_": [
                    ("attributes", field),
                    ("creation_low", field),
                    ("creation_high", field),
                    ("access_low", field),
                    ("access_high", field),
                    ("write_low", field),
                    ("write_high", field),
                    ("volume", field),
                    ("size_high", field),
                    ("size_low", field),
                    ("links", field),
                    ("index_high", field),
                    ("index_low", field),
                ]
            },
        )
        info = info_type()
        get_info = ctypes.windll.kernel32.GetFileInformationByHandle
        get_info.argtypes = (ctypes.c_void_p, ctypes.POINTER(info_type))
        get_info.restype = ctypes.c_int
        if not get_info(ctypes.c_void_p(descriptor), ctypes.byref(info)):
            raise OSError(ctypes.get_last_error(), "GetFileInformationByHandle failed")
        file_type = ctypes.windll.kernel32.GetFileType(ctypes.c_void_p(descriptor))
        return {
            "device": int(info.volume),
            "inode": (int(info.index_high) << 32) | int(info.index_low),
            "size": (int(info.size_high) << 32) | int(info.size_low),
            "regular": not bool(info.attributes & _FILE_ATTRIBUTE_DIRECTORY),
            "nlink": int(info.links),
            "reparse": bool(info.attributes & _FILE_ATTRIBUTE_REPARSE_POINT),
            "directory": bool(info.attributes & _FILE_ATTRIBUTE_DIRECTORY),
            "disk": file_type == _FILE_TYPE_DISK,
        }
    except Exception as exc:
        raise SafePathError("GetFileInformationByHandle failed") from exc


def _nt_volume_anchor(path: Path) -> tuple[str, tuple[str, ...]]:
    raw = str(path)
    if raw.startswith("\\\\?\\UNC\\"):
        raw = "\\\\" + raw[8:]
    elif raw.startswith("\\\\?\\"):
        raw = raw[4:]
    lexical = PureWindowsPath(raw)
    if not lexical.is_absolute():
        raise SafePathError("Windows root must be absolute")
    parts = tuple(str(part) for part in lexical.parts[1:])
    if any(":" in part for part in parts):
        raise SafePathError("Windows alternate data streams are denied")
    if lexical.drive.startswith("\\\\"):
        return "\\??\\UNC\\" + lexical.drive.lstrip("\\") + "\\", parts
    if len(lexical.drive) == 2 and lexical.drive[1] == ":":
        return "\\??\\" + lexical.drive + "\\", parts
    raise SafePathError("unsupported Windows volume root")


def _nt_create_file(
    root_directory: int | None,
    name: str,
    directory: bool,
    write: bool,
    create_new: bool,
    tolerate_missing: bool,
    tolerate_collision: bool,
) -> int:
    try:
        unicode_type = type(
            "_UnicodeString",
            (ctypes.Structure,),
            {"_fields_": [("Length", ctypes.c_ushort), ("MaximumLength", ctypes.c_ushort), ("Buffer", ctypes.c_void_p)]},
        )
        object_type = type(
            "_ObjectAttributes",
            (ctypes.Structure,),
            {
                "_fields_": [
                    ("Length", ctypes.c_uint32),
                    ("RootDirectory", ctypes.c_void_p),
                    ("ObjectName", ctypes.POINTER(unicode_type)),
                    ("Attributes", ctypes.c_uint32),
                    ("SecurityDescriptor", ctypes.c_void_p),
                    ("SecurityQualityOfService", ctypes.c_void_p),
                ]
            },
        )
        status_type = type(
            "_IoStatusBlock",
            (ctypes.Structure,),
            {"_fields_": [("Status", ctypes.c_long), ("Information", ctypes.c_void_p)]},
        )
        buffer = ctypes.create_unicode_buffer(name)
        encoded = name.encode("utf-16-le")
        unicode = unicode_type(
            len(encoded),
            len(encoded) + ctypes.sizeof(ctypes.c_wchar),
            ctypes.cast(buffer, ctypes.c_void_p),
        )
        attributes = object_type(
            ctypes.sizeof(object_type),
            ctypes.c_void_p(root_directory or 0),
            ctypes.pointer(unicode),
            _OBJ_CASE_INSENSITIVE,
            None,
            None,
        )
        io_status = status_type()
        handle = ctypes.c_void_p()
        desired_access = _FILE_READ_ATTRIBUTES | _SYNCHRONIZE
        desired_access |= _FILE_WRITE_DATA | _FILE_WRITE_ATTRIBUTES | _DELETE if write else _FILE_READ_DATA
        options = FILE_FLAG_OPEN_REPARSE_POINT | _FILE_SYNCHRONOUS_IO_NONALERT
        options |= _FILE_DIRECTORY_FILE if directory else _FILE_NON_DIRECTORY_FILE
        create = ctypes.windll.ntdll.NtCreateFile
        create.argtypes = (
            ctypes.POINTER(ctypes.c_void_p),
            ctypes.c_uint32,
            ctypes.POINTER(object_type),
            ctypes.POINTER(status_type),
            ctypes.c_void_p,
            ctypes.c_uint32,
            ctypes.c_uint32,
            ctypes.c_uint32,
            ctypes.c_uint32,
            ctypes.c_void_p,
            ctypes.c_uint32,
        )
        create.restype = ctypes.c_long
        status = create(
            ctypes.byref(handle),
            desired_access,
            ctypes.byref(attributes),
            ctypes.byref(io_status),
            None,
            _FILE_ATTRIBUTE_NORMAL if create_new else 0,
            _FILE_SHARE_READ | _FILE_SHARE_WRITE | _FILE_SHARE_DELETE,
            _FILE_CREATE if create_new else _FILE_OPEN,
            options,
            None,
            0,
        )
        if status < 0:
            code = status & 0xFFFFFFFF
            if tolerate_missing and code in (0xC0000034, 0xC000003A):
                return -1
            if tolerate_collision and code == 0xC0000035:
                return -1
            raise OSError(code, "NtCreateFile failed", name)
        if handle.value is None:
            raise OSError("NtCreateFile returned no handle")
        return int(handle.value)
    except Exception as exc:
        raise SafePathError("NtCreateFile failed") from exc


def _nt_open_absolute_directory(path: Path) -> tuple[int, tuple[int, ...], int]:
    anchor, parts = _nt_volume_anchor(path)
    anchor_handle = _nt_create_file(None, anchor, True, False, False, False, False)
    ancestors = [anchor_handle]
    current = anchor_handle
    child = -1
    try:
        anchor_values = _require_directory(_fstat(None, anchor_handle))
        root_device = int(anchor_values["device"])
        for part in parts:
            child = _nt_open_relative(current, part, True)
            try:
                child_values = _require_directory(_fstat(None, child))
                if child_values["device"] != root_device:
                    raise SafePathError("path crossed a device boundary")
            except Exception:
                _close(None, child)
                child = -1
                raise
            ancestors.append(child)
            current = child
            child = -1
        return current, tuple(ancestors), root_device
    except Exception:
        if child >= 0:
            _close(None, child)
        for ancestor in reversed(ancestors):
            _close(None, ancestor)
        raise


def _nt_open_relative(
    parent: int,
    name: str,
    directory: bool,
    write: bool = False,
    create_new: bool = False,
    tolerate_missing: bool = False,
    tolerate_collision: bool = False,
) -> int:
    return _nt_create_file(
        parent,
        name,
        directory,
        write,
        create_new,
        tolerate_missing,
        tolerate_collision,
    )


def _windows_write(handle: int, payload: bytes) -> None:
    try:
        write = ctypes.windll.kernel32.WriteFile
        write.argtypes = (ctypes.c_void_p, ctypes.c_void_p, ctypes.c_uint32, ctypes.POINTER(ctypes.c_uint32), ctypes.c_void_p)
        write.restype = ctypes.c_int
        buffer = ctypes.create_string_buffer(payload)
        offset = 0
        while offset < len(payload):
            count = min(len(payload) - offset, 0x7FFFF000)
            written = ctypes.c_uint32()
            if not write(
                ctypes.c_void_p(handle), ctypes.byref(buffer, offset), count, ctypes.byref(written), None
            ) or written.value == 0:
                raise OSError(ctypes.get_last_error(), "WriteFile failed")
            offset += int(written.value)
        flush = ctypes.windll.kernel32.FlushFileBuffers
        flush.argtypes = (ctypes.c_void_p,)
        flush.restype = ctypes.c_int
        if not flush(ctypes.c_void_p(handle)):
            raise OSError(ctypes.get_last_error(), "FlushFileBuffers failed")
    except Exception as exc:
        raise SafePathError("Windows temporary write failed") from exc


def _windows_delete_handle(handle: int) -> None:
    try:
        delete = ctypes.c_int(1)
        set_info = ctypes.windll.kernel32.SetFileInformationByHandle
        set_info.argtypes = (ctypes.c_void_p, ctypes.c_int, ctypes.c_void_p, ctypes.c_uint32)
        set_info.restype = ctypes.c_int
        if not set_info(
            ctypes.c_void_p(handle), _FILE_DISPOSITION_INFO, ctypes.byref(delete), ctypes.sizeof(delete)
        ):
            raise OSError(ctypes.get_last_error(), "SetFileInformationByHandle delete failed")
    except Exception as exc:
        raise SafePathError("Windows temporary cleanup failed") from exc


def _windows_rename_relative(handle: int, parent: int, destination: str) -> None:
    try:
        header_type = type(
            "_FileRenameInformation",
            (ctypes.Structure,),
            {
                "_fields_": [
                    ("ReplaceIfExists", ctypes.c_ubyte),
                    ("RootDirectory", ctypes.c_void_p),
                    ("FileNameLength", ctypes.c_uint32),
                    ("FileName", ctypes.c_wchar * 1),
                ]
            },
        )
        encoded = destination.encode("utf-16-le")
        offset = header_type.FileName.offset
        size = offset + len(encoded) + ctypes.sizeof(ctypes.c_wchar)
        storage = ctypes.create_string_buffer(size)
        header = ctypes.cast(storage, ctypes.POINTER(header_type)).contents
        header.ReplaceIfExists = 0
        header.RootDirectory = ctypes.c_void_p(parent)
        header.FileNameLength = len(encoded)
        ctypes.memmove(ctypes.addressof(storage) + offset, encoded, len(encoded))
        set_info = ctypes.windll.kernel32.SetFileInformationByHandle
        set_info.argtypes = (ctypes.c_void_p, ctypes.c_int, ctypes.c_void_p, ctypes.c_uint32)
        set_info.restype = ctypes.c_int
        if not set_info(ctypes.c_void_p(handle), _FILE_RENAME_INFO, storage, size):
            raise OSError(ctypes.get_last_error(), "SetFileInformationByHandle rename failed")
    except Exception as exc:
        raise SafePathError("Windows no-replace rename failed") from exc


def _open_windows_leaf(
    root: Path, parts: tuple[str, ...], ops: Any | None
) -> tuple[int, tuple[int, ...], int | None]:
    if ops is not None:
        open_root = getattr(ops, "nt_open_absolute_directory", None)
        open_relative = getattr(ops, "nt_open_relative", None)
        if not (callable(open_root) and callable(open_relative)):
            raise SafePathError("rooted Windows ops are required")
        root_handle = int(open_root(root))
        ancestors = [root_handle]
        current = root_handle
        child = -1
        try:
            root_values = _require_directory(_fstat(ops, root_handle))
            root_device = int(root_values["device"])
            for ordinal, part in enumerate(parts):
                directory = ordinal + 1 != len(parts)
                child = int(open_relative(current, part, directory))
                child_values = _stat_values(_fstat(ops, child))
                if child_values["device"] != root_device:
                    raise SafePathError("path crossed a device boundary")
                if child_values["reparse"] is True:
                    raise SafePathError("reparse point is denied")
                if directory:
                    _require_directory(child_values)
                    ancestors.append(child)
                    current = child
                    child = -1
            return current if child == -1 else child, tuple(ancestors), root_device
        except Exception:
            if child >= 0:
                _close(ops, child)
            for ancestor in reversed(ancestors):
                _close(ops, ancestor)
            raise
    root_handle, root_ancestors, root_device = _nt_open_absolute_directory(root)
    ancestors = list(root_ancestors)
    current = root_handle
    child = -1
    try:
        for ordinal, part in enumerate(parts):
            directory = ordinal + 1 != len(parts)
            child = _nt_open_relative(current, part, directory)
            try:
                snapshot = _fstat(None, child)
                child_values = _stat_values(snapshot)
                if child_values["device"] != root_device:
                    raise SafePathError("path crossed a device boundary")
                if child_values["reparse"] is True:
                    raise SafePathError("reparse point is denied")
                if directory:
                    _require_directory(child_values)
            except Exception:
                _close(None, child)
                child = -1
                raise
            if directory:
                ancestors.append(child)
                current = child
                child = -1
        return current if child == -1 else child, tuple(ancestors), root_device
    except Exception:
        if child >= 0:
            _close(None, child)
        for ancestor in reversed(ancestors):
            _close(None, ancestor)
        raise


def open_regular_relative(root: Path, relative: str, *, ops: Any | None = None) -> int:
    candidate = root / relative
    if candidate.name.casefold() == "readme.md":
        raise SafePathError("README paths are not readable")
    windows = _is_windows(ops)
    parts = _relative_parts(relative, windows)
    retained: tuple[int, ...] = ()
    root_device: int | None = None
    if windows:
        descriptor, retained, root_device = _open_windows_leaf(root, parts, ops)
    else:
        descriptor, root_device = _open_posix_leaf(root, parts, ops)
    try:
        before = _fstat(ops, descriptor)
        if root_device is not None and _stat_values(before)["device"] != root_device:
            raise SafePathError("leaf crossed a device boundary")
        _require_regular(before)
        after = _fstat(ops, descriptor)
        if root_device is not None and _stat_values(after)["device"] != root_device:
            raise SafePathError("leaf crossed a device boundary")
        _require_stable(before, after)
        return descriptor
    except Exception:
        _close(ops, descriptor)
        raise
    finally:
        for ancestor in reversed(retained):
            _close(ops, ancestor)


def _output_name(path: Path, windows: bool) -> str:
    name = path.name
    if windows:
        lexical = PureWindowsPath(str(path))
        for index, component in enumerate(lexical.parts):
            if ":" in str(component) and not (
                index == 0 and lexical.is_absolute() and str(component) == lexical.anchor
            ):
                raise SafePathError("Windows alternate data streams are denied")
    if (
        not name
        or name in (".", "..")
        or "/" in name
        or "\\" in name
        or windows and ":" in name
    ):
        raise SafePathError("output name is unsafe")
    return name


def _native_destination_exists(directory_fd: int, name: str) -> bool:
    try:
        library = ctypes.CDLL(None, use_errno=True)
        call = library.fstatat
        probe_type = ctypes.c_byte * 256
        probe = probe_type()
        result = call(directory_fd, os.fsencode(name), ctypes.byref(probe), 0x100)
        if result == 0:
            return True
        errno = ctypes.get_errno()
        if errno == 2:
            return False
        raise OSError(errno, "fstatat failed", name)
    except SafePathError:
        raise
    except Exception as exc:
        raise SafePathError("destination inspection failed") from exc


def _native_write_new(directory_fd: int, destination: str, payload: bytes) -> tuple[str, int]:
    temporary = f".{destination}.{os.urandom(16).hex()}.tmp"
    flags = os.O_WRONLY | os.O_CREAT | os.O_EXCL | O_NOFOLLOW | _O_CLOEXEC | _O_BINARY
    descriptor = _posix_openat(directory_fd, temporary, flags, 0o600)
    try:
        offset = 0
        while offset < len(payload):
            written = os.write(descriptor, payload[offset:])
            if written <= 0:
                raise SafePathError("temporary write did not advance")
            offset += written
        os.fsync(descriptor)
        before = _fstat(None, descriptor)
        _require_stable(before, _fstat(None, descriptor))
        return temporary, int(_stat_values(before)["device"])
    except Exception:
        try:
            library = ctypes.CDLL(None, use_errno=True)
            library.unlinkat(directory_fd, os.fsencode(temporary), 0)
        except Exception:
            pass
        raise
    finally:
        _close(None, descriptor)


def _native_renameat(source_parent_fd: int, source: str, destination_parent_fd: int, destination: str) -> None:
    if source_parent_fd != destination_parent_fd:
        raise SafePathError("cross-directory publication is denied")
    try:
        library = ctypes.CDLL(None, use_errno=True)
        rename = library.renameat2
        rename.argtypes = (ctypes.c_int, ctypes.c_char_p, ctypes.c_int, ctypes.c_char_p, ctypes.c_uint)
        rename.restype = ctypes.c_int
        if rename(
            source_parent_fd,
            os.fsencode(source),
            destination_parent_fd,
            os.fsencode(destination),
            _RENAME_NOREPLACE,
        ) != 0:
            errno = ctypes.get_errno()
            if errno in (22, 38, 95):
                raise SafePathError("renameat2 RENAME_NOREPLACE is unavailable")
            raise OSError(errno, "renameat2 publication failed", destination)
    except Exception as exc:
        if isinstance(exc, SafePathError):
            raise
        raise SafePathError("atomic same-directory publication failed") from exc


def _native_unlinkat(directory_fd: int, name: str) -> None:
    try:
        ctypes.CDLL(None, use_errno=True).unlinkat(directory_fd, os.fsencode(name), 0)
    except Exception:
        pass


def _windows_atomic_write(path: Path, destination: str, payload: bytes) -> None:
    parent, parent_ancestors, parent_device = _nt_open_absolute_directory(path.parent)
    temporary_handle = -1
    try:
        existing = _nt_open_relative(parent, destination, False, tolerate_missing=True)
        if existing >= 0:
            _close(None, existing)
            raise SafePathError("destination already exists")
        for _attempt in range(8):
            temporary = f".{destination}.{os.urandom(16).hex()}.tmp"
            temporary_handle = _nt_open_relative(
                parent,
                temporary,
                False,
                write=True,
                create_new=True,
                tolerate_collision=True,
            )
            if temporary_handle >= 0:
                break
        if temporary_handle < 0:
            raise SafePathError("could not reserve a temporary output")
        _windows_write(temporary_handle, payload)
        before = _fstat(None, temporary_handle)
        if _stat_values(before)["device"] != parent_device:
            raise SafePathError("temporary crossed a device boundary")
        _require_regular(before)
        after = _fstat(None, temporary_handle)
        if _stat_values(after)["device"] != parent_device:
            raise SafePathError("temporary crossed a device boundary")
        _require_stable(before, after)
        _windows_rename_relative(temporary_handle, parent, destination)
        _close(None, temporary_handle)
        temporary_handle = -1
    except Exception as exc:
        cleanup_error: Exception | None = None
        if temporary_handle >= 0:
            try:
                _windows_delete_handle(temporary_handle)
            except Exception as cleanup_exc:
                cleanup_error = cleanup_exc
            _close(None, temporary_handle)
        if cleanup_error is not None:
            raise SafePathError("Windows temporary cleanup was not confirmed") from exc
        raise
    finally:
        for ancestor in reversed(parent_ancestors):
            _close(None, ancestor)


def atomic_write_same_dir(path: Path, payload: bytes, *, ops: Any | None = None) -> None:
    if type(payload) is not bytes:
        raise SafePathError("payload must be bytes")
    windows = _is_windows(ops)
    destination = _output_name(path, windows)
    if windows and ops is None:
        _windows_atomic_write(path, destination, payload)
        return
    if ops is not None:
        open_parent = getattr(ops, "open_parent", None)
        write_new_at = getattr(ops, "write_new_at", None)
        if not (callable(open_parent) and callable(write_new_at)):
            raise SafePathError("rooted output ops are required")
        parent_fd = int(open_parent(path.parent))
        try:
            if ops.destination_exists(path):
                raise SafePathError("destination already exists")
            temporary = Path(write_new_at(parent_fd, destination, payload)).name
            if not temporary.endswith(".tmp"):
                raise SafePathError("writer returned an unsafe temporary name")
            ops.renameat(parent_fd, temporary, parent_fd, destination)
            return
        except SafePathError:
            raise
        except Exception as exc:
            raise SafePathError("atomic same-directory publication failed") from exc
        finally:
            _close(ops, parent_fd)
    parent_fd, parent_device = _posix_open_directory(path.parent)
    temporary = ""
    try:
        if _native_destination_exists(parent_fd, destination):
            raise SafePathError("destination already exists")
        temporary, temporary_device = _native_write_new(parent_fd, destination, payload)
        if temporary_device != parent_device:
            raise SafePathError("temporary crossed a device boundary")
        _native_renameat(parent_fd, temporary, parent_fd, destination)
        temporary = ""
    except Exception:
        if temporary:
            _native_unlinkat(parent_fd, temporary)
        raise
    finally:
        _close(None, parent_fd)
