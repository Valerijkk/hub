# A2-W3 SUMMARY — messenger + ws + calls (Stage A2 re-audit)

> TASK-008 · Stage A2 · READ-ONLY · 2026-07-28  
> Периметр: `W3-ASSIGNMENTS.md` · **131/131** файлов  
> Отчёты: `A2W3-1` … `A2W3-5`

## Severity (открытые: NEW + REGRESS + RESIDUAL)

| Агент | Зона | Файлов | P0 | P1 | P2 | P3 | Σ open |
|-------|------|-------:|---:|---:|---:|---:|-------:|
| A2W3-1 | messenger handlers | 26 | 0 | 0 | 2 | 1 | **3** |
| A2W3-2 | messenger services | 28 | 0 | 0 | 3 | 1 | **4** |
| A2W3-3 | messenger repos | 20 | 0 | 0 | 4 | 0 | **4** |
| A2W3-4 | ws | 15 | 0 | 0 | 1 | 1 | **2** |
| A2W3-5 | calls | 42 | 0 | 0 | 0 | 0 | **0** |
| **Σ** | | **131** | **0** | **0** | **10** | **3** | **13** |

## По тегам

| Tag | Count |
|-----|------:|
| NEW | **2** |
| REGRESS | **0** |
| RESIDUAL | **11** |
| FIXED-OK (справочно, не в Σ open) | **~59** Stage A findings закрыты |

## NEW (2)

| ID | Sev | Суть |
|----|-----|------|
| A2W3-3-004 | P2 | Mongo `AddReaction`: concurrent first-react TOCTOU → двойной `$push` после MatchedCount-фикса |
| A2W3-2-004 | P3 | AcceptInvite: `IsMember` err проглочен (`_`) |

## REGRESS

Нет. P0/P1 IDOR/RBAC/NF-02 guards после W13–W17 + W21 не откатились.

## Топ RESIDUAL

1. **Invite Accept RMW + orphan CreateChannel/DM** (A2W3-2-001/002, A2W3-3-001) — TASK-038 messenger out-of-scope  
2. **Unread TTL / unbounded Ent lists** (A2W3-3-002/003) — TASK-042 deferred  
3. **Recording orphan egress + finalizeCreate** (A2W3-5-001/002) — **FIXED TASK-110**  
4. **WS revoke end-to-end** (A2W3-4-002): Hub Unregister готов, auth не публикует revoked  
5. **JWT `?token=`** (A2W3-4-001) — product ADR  
6. **GetPresence** без shared-channel scope (A2W3-1-001); **wave114** wrong poll URL (A2W3-1-002)

## Stage A → сейчас

| Stage A Σ | Было | A2 open |
|-----------|-----:|--------:|
| P0 | 1 | **0** (promote RBAC FIXED) |
| P1 | 17 | **0** |
| P2 | 32 | **10** residual (+1 NEW P2) |
| P3 | 18 | **3** residual (+1 NEW P3) |

## Пути артефактов

```
.planning/audits/total-2026-07/A2W3-1-messenger-handlers.md
.planning/audits/total-2026-07/A2W3-2-messenger-services.md
.planning/audits/total-2026-07/A2W3-3-messenger-repos.md
.planning/audits/total-2026-07/A2W3-4-ws.md
.planning/audits/total-2026-07/A2W3-5-calls.md
.planning/audits/total-2026-07/A2-W3-SUMMARY.md
```

→ Следующая волна A2 по плану: W4 platform + apigateway (если оркестратор идёт по порядку).
