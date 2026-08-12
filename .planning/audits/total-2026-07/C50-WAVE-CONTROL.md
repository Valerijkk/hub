# C50 wave control

> Оркестратор: волны **по 5** (owner 2026-07-29). 50/10 разом — hang.

| Wave | Slots | Status |
|------|-------|--------|
| W0 | verify×4 + TASK-190 | **в W1** |
| W1 | 190 + verify×4 + C50-01…05 | **RUNNING** (max 10) |
| W2 | C50-06…15 | queued |
| W3 | C50-16…25 | queued |
| W4 | C50-26…35 | queued |
| W5 | C50-36…45 | queued |
| W6 | C50-46…50 | queued |

ABORT все предыдущие C50 слоты при hang; не ждать.
