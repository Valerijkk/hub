# PW4-REPORT — Требования и решения

**Дата:** 2026-07-27

## Сделано

| ID | Результат | Коммит |
|----|-----------|--------|
| D1 | `NFR-24-VERSION-CLASSIFICATION.md` — все схемы + 2 опции владельцу | `3f6157a` |
| D2 | `NFR-11-TRASH-CLASSIFICATION.md` | `1c9685a` |
| D3 | PLAN phases for FR-INTR-08 (`126-…`) and FR-CAL-05 (`127-…`) — код не писался | `48a3bbc`, `1a78da4` |
| D4 | `DECISIONS-PENDING.md` — 14 `[default]` as owner questions | `934aa96` |
| D5 | GDPR/wiki invalidity covered in DECISIONS / GDPR notes | with D4/D5 |
| D6 | `GRAY-ZONE-MARKERS-STATUS.md` | `5b9f80d` |

## Владельцу (приоритет)

1. GZ-SEC-17/18/19 GDPR — решение ссылается на wiki cascade, модуля нет.
2. GZ-SEC-13/14 AES-GCM HR/зарплаты.
3. NFR-24 / NFR-11: раскатать mixin/корзину **или** сузить формулировку (не оставлять молчание).
4. Утвердить PLAN 126/127 перед кодом.

## Хуже ожидаемого

- Разрыв NFR-24/NFR-11 не «мелкий» — без решения владельца любой статус «готово» врёт.
