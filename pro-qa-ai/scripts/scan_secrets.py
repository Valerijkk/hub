"""Поиск конфиденциальных значений — в рабочем дереве и во всей истории.

Валидатор (`secrets`) смотрит только текущее состояние файлов пакета.
Этого мало: значение, убранное из последней версии, остаётся в предыдущих
ревизиях и достаётся оттуда одной командой. Поиск, выполненный по HEAD,
даёт ложное «чисто».

Скрипт ничего не правит и не удаляет. Он сообщает путь, ревизию, номер
строки и класс — само значение не печатается никогда.

Правку истории выполняет человек: она необратима и задевает всех, у кого
есть клон репозитория (rules/60 §4).

Использование:
    python scripts/scan_secrets.py                 рабочее дерево
    python scripts/scan_secrets.py --history       плюс вся история git
    python scripts/scan_secrets.py --root <путь>   другой корень

Экзит-коды: 0 чисто, 1 найдено, 2 запуск не удался.
"""

from __future__ import annotations

import argparse
import io
import re
import subprocess
import sys
from pathlib import Path
from typing import Iterable, List, Tuple

sys.path.insert(0, str(Path(__file__).resolve().parent))

from secret_rules import secret_rule_matches  # noqa: E402

TEXT_SUFFIXES = {
    ".md", ".py", ".json", ".yaml", ".yml", ".txt", ".cfg", ".ini",
    ".toml", ".sh", ".ps1", ".env", ".conf", ".xml", ".properties",
}
SKIP_DIRS = {".git", "__pycache__", "node_modules", ".venv", "dist", ".pytest_cache"}
GIT_TIMEOUT = 120
_HEX = re.compile(r"[0-9a-f]{40}")


def _force_utf8_output() -> None:
    """Вывод в UTF-8 независимо от кодировки консоли.

    На Windows консоль по умолчанию в cp1251, и печать любого символа вне
    неё роняет скрипт целиком. Инструмент безопасности, падающий на
    собственном отчёте, хуже отсутствующего: он оставляет впечатление,
    что проверка прошла.
    """
    for stream_name in ("stdout", "stderr"):
        stream = getattr(sys, stream_name, None)
        buffer = getattr(stream, "buffer", None)
        if buffer is None:
            continue
        setattr(
            sys,
            stream_name,
            io.TextIOWrapper(buffer, encoding="utf-8", errors="replace", line_buffering=True),
        )


def _run_git(root: Path, *args: str) -> str:
    """git с таймаутом; при отказе возвращает пустую строку.

    Таймаут обязателен: без него зависший вызов останавливает проверку
    целиком и выглядит как «сканирование идёт».
    """
    try:
        done = subprocess.run(
            ["git", "-C", str(root), *args],
            capture_output=True,
            timeout=GIT_TIMEOUT,
            check=False,
        )
    except (OSError, subprocess.TimeoutExpired):
        return ""
    if done.returncode != 0:
        return ""
    # Байты, а не text=True: в истории попадаются файлы в чужой кодировке и
    # двоичные. Декодирование с заменой не роняет проверку целиком — иначе
    # один нечитаемый файл прячет все находки после него.
    return done.stdout.decode("utf-8", errors="replace")


def scan_text(text: str) -> List[Tuple[str, int]]:
    """Пары «класс правила, номер строки».

    `secret_rule_matches` отдаёт смещение в символах, а не строку. Разница
    незаметна на коротком файле и врёт на длинном: смещение 331 в файле из
    двадцати трёх строк выглядит как строка 331, которой нет. Человек идёт
    по указанному адресу, не находит ничего и перестаёт верить отчёту.
    """
    return [
        (rule_id, text.count(chr(10), 0, offset) + 1)
        for rule_id, offset in secret_rule_matches(text)
    ]


def scan_worktree(root: Path) -> List[str]:
    findings: List[str] = []
    for path in sorted(root.rglob("*")):
        if not path.is_file() or path.suffix.lower() not in TEXT_SUFFIXES:
            continue
        if any(part in SKIP_DIRS for part in path.parts):
            continue
        try:
            text = path.read_text(encoding="utf-8", errors="replace")
        except OSError:
            continue
        for rule_id, line in scan_text(text):
            relative = path.relative_to(root).as_posix()
            findings.append("рабочее дерево  %s:%d  класс %s" % (relative, line, rule_id))
    return findings


def scan_history(root: Path, max_commits: int = 0) -> List[str]:
    """Все ревизии за один проход `git log -p`.

    Перебор «ревизия x файл» через `git show` даёт тысячи вызовов git и на
    реальном репозитории не укладывается ни в какое разумное время: 2000
    коммитов по сотне файлов — это 200 000 запусков процесса.

    Здесь один вызов отдаёт все добавленные строки сразу. Ищем именно
    добавленные (`+`): момент, когда значение появилось, — это и есть
    момент утечки.
    """
    args = ["log", "--all", "-p", "--unified=0", "--no-color", "--format=%H"]
    if max_commits:
        args.insert(1, "--max-count=%d" % max_commits)
    stream = _run_git(root, *args)
    if not stream:
        return ["история недоступна: не репозиторий git, нет коммитов либо превышен таймаут"]

    findings: List[str] = []
    revision = "?"
    current_file = "?"
    seen = set()
    for line in stream.splitlines():
        if len(line) == 40 and _HEX.fullmatch(line):
            revision = line
            continue
        if line.startswith("+++ b/"):
            current_file = line[6:]
            continue
        if not line.startswith("+") or line.startswith("+++"):
            continue
        for rule_id, _ in scan_text(line[1:]):
            key = (revision, current_file, rule_id)
            if key in seen:
                continue
            seen.add(key)
            findings.append(
                "история %s  %s  класс %s" % (revision[:10], current_file, rule_id)
            )
    return findings


def main(argv: Iterable[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", default=".", help="корень проверки")
    parser.add_argument("--history", action="store_true", help="проверить всю историю git")
    parser.add_argument("--max-commits", type=int, default=0, help="ограничить число коммитов (0 — все)")
    args = parser.parse_args(list(argv) if argv is not None else None)
    _force_utf8_output()

    root = Path(args.root).resolve()
    if not root.is_dir():
        print("корень не найден: %s" % root, file=sys.stderr)
        return 2

    findings = scan_worktree(root)
    if args.history:
        findings.extend(scan_history(root, args.max_commits))

    if not findings:
        print("чисто: совпадений нет (%s)" % ("дерево и история" if args.history else "рабочее дерево"))
        return 0

    print("НАЙДЕНО совпадений: %d" % len(findings))
    print("Значения не печатаются намеренно — см. rules/50 §3.2.\n")
    for item in findings:
        print("  " + item)
    print("\nПорядок действий: остановиться → сообщить человеку → ротация → очистка.")
    print("Правку истории выполняет человек (rules/60 §4).")
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
