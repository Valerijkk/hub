# W7 — CI / GitHub Actions (repo-scan 2026-07-23)

> **Provenance class:** PRE_EXISTING · **Confidence:** MIXED · **current_verified:** false (pending code/source recheck)
> Embedded original verdict below = historical self-check only — **not** independent acceptance. See [`PROVENANCE.md`](./PROVENANCE.md).

> Агент: `devops-engineer` (Grok 4.5, **readonly**) · Дата: **2026-07-23**  
> Scope: `.github/workflows/{ci,pr-guard,deploy}.yml`  
> Фокус: **minutes waste** (~3000 мин/мес), **`DEV_MODE`**, **secrets**  
> Код **не менялся**.

## Verdict

| Метрика | Значение |
|---------|----------|
| CRITICAL | 0 |
| HIGH | 0 |
| MEDIUM | 4 |
| LOW | 4 |
| INFO / accepted | 3 |
| **verified** | **false** — открытые MEDIUM (экономия минут + слабый secret-scan) |
| Блокер релиза / prod | **нет** (`DEV_MODE` в deploy отсутствует) |

**Кратко:** прод-инварианты соблюдены — `deploy.yml` без `DEV_MODE`, деплой только после green CI (`workflow_run.conclusion == 'success'` + `event == 'push'`), SSH/GHCR через `secrets.*`. Главные gaps: **нет path-filters** (полный CI на docs-only), **`e2e-full` на каждом PR**, **дубли `go generate`/`npm ci`/release-build**, **pr-guard secret-scan слишком узкий**.

---

## Checklist

| Контроль | Статус | Evidence |
|----------|--------|----------|
| `DEV_MODE=true` в prod deploy | PASS | `deploy.yml` — 0 совпадений `DEV_MODE` |
| `DEV_MODE` только в CI E2E | ACCEPTED | `ci.yml:225` job `e2e-full` |
| Deploy gated on green CI | PASS | `deploy.yml:16,50` — `conclusion == 'success'` |
| Deploy только push (не PR) | PASS | `deploy.yml:16,50` — `event == 'push'` |
| Concurrency / cancel-in-progress | PASS | `ci.yml:10-12` |
| Coverage не на PR | PASS | `ci.yml:69-71` `if: != pull_request` |
| Lighthouse non-blocking + не на PR | PASS | `ci.yml:296-301` |
| Backend release build в CI | PASS | `ci.yml:63-67`, `176-180` `-tags release` |
| Secrets в workflow как plaintext prod | PASS | SSH/GHCR — `${{ secrets.* }}` |
| Hardcoded prod credentials в YAML | PASS | только CI-тестовые `hub`/`local-dev-secret…` |
| Secret-scan в PR Guard | PARTIAL | `pr-guard.yml:27` — узкий regex |
| Path filters (экономия минут) | FAIL | нет `paths` / `paths-ignore` ни в одном workflow |
| Image tag `:latest` в GHCR push | INFO | `deploy.yml:39,46` — рядом с SHA; deploy script тянет **SHA** |

---

## Findings

| Sev | Файл:строка | Наблюдение | Рекомендация | Класс | Статус |
|-----|-------------|------------|--------------|-------|--------|
| MEDIUM | `ci.yml:3-8` (+ все jobs) | Нет `paths`/`paths-ignore`. Любой push/PR в `master` (в т.ч. только `.planning/`, `docs/`, markdown) гоняет backend-test + frontend-test + build + e2e-smoke + **e2e-full**. На студенческом пакете ~3000 мин/мес это основной waste. | Добавить path-filters: backend → `backend/**`, frontend → `frontend/**`/`testframework/**`, docs-only → только `pr-guard`/`docs-link-check`. Либо top-level `paths-ignore: ['**.md', '.planning/**']` с осторожностью для doc-gates. | minutes | open |
| MEDIUM | `ci.yml:220-273` `e2e-full` | Full E2E (Docker infra + native backend + Playwright, `timeout-minutes: 30`) на **каждом PR**. Самый дорогой job после coverage. Coverage уже вынесен с PR; e2e-full — нет. | Варианты: (a) `if: github.event_name != 'pull_request'` + оставлять `e2e-smoke` на PR; (b) path-filter на `backend/**`/`frontend/**`/`testframework/**`; (c) `workflow_dispatch` / label `run-e2e-full`. | minutes | open |
| MEDIUM | `ci.yml:63-67` + `166-189` | Дублирование release-сборки: `backend-test` уже делает `go generate` + `go build -tags release`; job `build` повторяет `go generate` + build backend **и** `npm ci && npm run build` после того как `frontend-test` уже ставил deps. Три независимых `npm ci` на PR-пути (frontend-test, build, e2e-smoke) + четвёртый в e2e-full. | Убрать дубль release-build из `backend-test` **или** сделать `build` артефакт-only / merge в needs без повторного generate; кэшировать Playwright browsers; не гонять `build` на docs-only (см. path-filter). | minutes | open |
| MEDIUM | `pr-guard.yml:20-31` | Secret-scan ловит только `password=…`, `api_key=…`, `BEGIN (RSA\|OPENSSH) PRIVATE KEY`. Не ловит: `JWT_SECRET=`, `RESEND_API_KEY`/`re_…`, Telegram `digits:token`, `LIVEKIT_API_SECRET`, `INTEGRATION_ENCRYPTION_KEY`, AWS keys, base64-секреты, `.env` содержимое без `= "…"`. | Расширить patterns (или `gitleaks`/`trufflehog` action с allowlist на фикстуры); сканировать и filenames (`.env`, `livekit-prod.yaml` без `.example`). Оставить как **дополнение** к ручному rule 50, не замена. | secrets | open |
| LOW | `ci.yml:53-55` + `63-66` | Внутри `backend-test` `go generate ./ent` вызывается **дважды** (vet step + build step). | Один generate на job. | minutes | open |
| LOW | `ci.yml:135-164`, `166-189`, `192-216` | У `frontend-test`, `build`, `e2e-smoke` нет `timeout-minutes` — зависание npm/playwright жжёт минуты до дефолтного лимита runner. | Выставить timeouts: frontend-test ~15, build ~15, e2e-smoke ~20 (как у backend/e2e-full). | minutes | open |
| LOW | `pr-guard.yml:3-7`, `13-18` | `secret-scan` на **push + PR**, `fetch-depth: 0` (полный clone). На каждый push в master — лишний полный history clone ради diff `before..HEAD`. | Ограничить secret-scan `if: github.event_name == 'pull_request'` (как docs-link / path-labels) **или** `fetch-depth: 2` для push. | minutes | open |
| LOW | `ci.yml` (root) | Нет явного `permissions:` (в отличие от `pr-guard.yml:9-10`). Default token на `push` часто шире `contents: read`. | Добавить least-privilege: `permissions: contents: read` на workflow; точечно расширять только если понадобится. | secrets / hygiene | open |
| INFO | `ci.yml:225`, `:233` | `DEV_MODE: 'true'` + `JWT_SECRET: local-dev-secret-change-in-production` в `e2e-full` — ожидаемо для ephemeral CI, **не** prod-путь. | Не копировать в `deploy.yml` / `docker-compose.prod.yml`. | DEV_MODE | **accepted** |
| INFO | `deploy.yml:39,46` | Push тегов `:latest` **и** `:${SHA}`. Deploy script (`:67-68`) корректно тянет **SHA**. Риск: ручной `docker pull …:latest` на сервере → drift. TZ forbidden явно про infra (`seaweedfs:latest`), не про app GHCR — но pinning-философия та же. | Опционально убрать `:latest` из push; в runbook запретить pull latest. | pinning | open (optional) |
| INFO | `deploy.yml:53` | `continue-on-error: true` на deploy — failed SSH/health не краснит workflow. Сознательный best-effort (сервер/секреты). | После стабилизации FA-06/SSH — убрать continue-on-error, чтобы деплой-фейлы были видны. | ops | accepted (пока) |

---

## Minutes map (оценка порядка)

| Job | PR | Push master | Порядок стоимости | Примечание |
|-----|----|-------------|-------------------|------------|
| `backend-test` | ✅ | ✅ | высокий | race + 600s timeout + mongo service |
| `frontend-test` | ✅ | ✅ | средний | npm ci + tsc + vitest + biome |
| `build` | ✅ | ✅ | средний | **дубль** generate/build/npm |
| `e2e-smoke` | ✅ | ✅ | средний | playwright chromium + deps |
| `e2e-full` | ✅ | ✅ | **очень высокий** | infra + backend + full PW; кандидат на skip PR |
| `coverage-gate` | ❌ | ✅ | высокий | уже оптимизирован |
| `lighthouse` | ❌ | ✅ | средний | non-blocking; apt chrome + rebuild |
| `pr-guard/*` | ✅ | secret-scan only | низкий | fetch-depth:0 на push — зря |
| `deploy` build+SSH | ❌ | после CI | высокий (образы) | gated; SSH best-effort |

**Уже хорошо:** concurrency cancel; coverage/lighthouse off PR; deploy только success+push; e2e-smoke без backend.

---

## DEV_MODE detail

| Место | Значение | Вердикт |
|-------|----------|---------|
| `ci.yml:225` `e2e-full.env` | `'true'` | ACCEPTED (CI-only) |
| `deploy.yml` | отсутствует | PASS |
| Prod compose (вне scope, сверка) | по W1-security: `"false"` | PASS (cross-ref) |

Инвариант INF-02 / TZ: **не переносить `DEV_MODE=true` в deploy** — сейчас соблюдён.

---

## Secrets detail

### OK
- GHCR login: `secrets.GITHUB_TOKEN` (`deploy.yml:31-32`)
- SSH: `SSH_HOST` / `SSH_USER` / `SSH_KEY` (`deploy.yml:59-61`) — не в plaintext
- Checkout deploy по `workflow_run.head_sha` — не из PR fork произвольного кода без green CI push

### Gaps
1. Слабый `pr-guard` regex (см. MEDIUM).
2. CI fixture secrets (`hub`/`hub`, `local-dev-secret-change-in-production`) — норма для test; не коммитить аналоги в prod compose.
3. Нет скана на `DEV_MODE=true` в diff deploy/prod-файлов (можно добавить строку в pr-guard для `deploy.yml` / `docker-compose.prod.yml`).

---

## Positive controls (не findings)

1. `concurrency.group: ci-${{ github.ref }}` + `cancel-in-progress: true`.
2. Coverage ≥82% только вне PR.
3. Deploy: `workflow_run` → CI success + push only.
4. Health-check loop после SSH deploy (`deploy.yml:95-103`).
5. Image prune 72h на сервере.
6. BOM-guard для Go (`ci.yml:34-44`).
7. `go build -tags release` в CI (ME-09).
8. Artifact retention ограничена (PW 7d, Lighthouse 14d).

---

## Рекомендуемый порядок фиксов (не в этом PR)

1. **Path-filters / skip docs-only** — максимальный ROI по минутам.  
2. **e2e-full off PR** (оставить smoke) — второй по ROI.  
3. Расширить **secret-scan** (+ опционально запрет `DEV_MODE=true` в prod-путях diff).  
4. Убрать двойной `go generate` / упростить job `build`.  
5. `timeout-minutes` на frontend/build/e2e-smoke; `permissions: contents: read` на `ci.yml`.

---

## Verification

| Этап | Результат |
|------|-----------|
| Readonly review трёх YAML | done |
| Grep `DEV_MODE` / secrets / `:latest` / timeouts | done |
| Код / workflow **не менялись** | confirmed |
| Docs product | n/a (audit artifact only) |

**verified: false** — 2026-07-23 · открыты 4× MEDIUM (minutes ×3 + secret-scan).  
CRITICAL/HIGH по DEV_MODE/prod-secrets: **0**.  
Handoff: `team-lead` / волна 7 слоты deploy·compose; фиксы CI — отдельная chore-ветка после code-freeze продукта или с явным ок на infra.
)
