# W3 SUMMARY — messenger + ws + calls

> TASK-008 · Стадия A · read-only · 131/131

| Агент | Зона | Файлов | P0 | P1 | P2 | P3 |
|-------|------|-------:|---:|---:|---:|---:|
| A1 | messenger handlers | 26 | 0 | 3 | 7 | 2 |
| A2 | messenger services | 28 | 1 | 7 | 8 | 2 |
| A3 | messenger repos | 20 | 0 | 2 | 5 | 5 |
| A4 | ws | 15 | 0 | 1 | 4 | 5 |
| A5 | calls | 42 | 0 | 4 | 8 | 4 |
| **Σ** | | **131** | **1** | **17** | **32** | **18** |

Сырых находок: **68**. Отчёты: A3-1…A3-5.

## Тесты
messenger / ws / calls — все **ok** (`-tags hubtest -race`).

## Топ
- **P0** (A2): критичный membership/IDOR в messenger service — см. A3-2
- WS NF-02 / calls LiveKit authz / recording webhook — см. A3-4, A3-5

→ W4 platform + apigateway
