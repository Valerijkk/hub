# Infrastructure & DevOps Audit — Hub v2.0

Scope: `docker-compose.*.yml`, `Dockerfile`, `deployments/`, `.github/workflows/`, `Makefile`, `.env.example`, observability, backup/recovery, docs.
All paths absolute.

---

## 1. Docker Compose

### `C:\Users\Valerii\Desktop\hub\docker-compose.local.yml`
- Versions pinned: postgres:17, mongo:8, redis:7.4-alpine, nats:2.11, livekit:v1.8, seaweedfs:**latest** (BAD — non-reproducible).
- Healthchecks present for postgres/mongo/redis/nats. Missing for `seaweedfs` and `livekit`.
- `depends_on: service_healthy` correct chain for backend, frontend uses `service_started` only.
- **SECRETS HARDCODED**: `RESEND_API_KEY`, `TELEGRAM_BOT_TOKEN` are **real production-looking tokens** committed in plaintext (lines 115, 117). These must be revoked + moved out of the file.
- LiveKit dev creds in plain — OK for local but they double as the same string in `deployments/livekit.yaml`.
- `S3_USE_SSL=false`, `JWT_SECRET=local-dev-secret-change-in-production` — fine in dev.

### `C:\Users\Valerii\Desktop\hub\deployments\docker-compose.prod.yml`
- All secrets via `${VAR}` — good. Resource limits set per service — good.
- Inline LiveKit config generation via heredoc (lines 113-134) — clever, but **does NOT include `departure_timeout: 120`** that you just added to `livekit.yaml`. The 30s-call-death fix WILL NOT REACH PRODUCTION via this compose file. Critical regression.
- mongo pinned to `mongo:7` here vs `mongo:8` in local — version drift.
- seaweedfs still `chrislusf/seaweedfs:latest`.
- `S3_USE_SSL: "false"` hardcoded — not env-driven.
- No `networks:` block — implicit default network. Frontend service has no ports published (relies on Caddy) — correct.
- `deploy.resources.limits` only honored under Swarm/k8s; plain `docker compose up -d` ignores them. Use top-level `mem_limit`/`cpus` for compose v2.

### `C:\Users\Valerii\Desktop\hub\deployments\docker-compose.yml`
- Legacy dev-only file used by `make dev-up`. No healthchecks on redis/nats/seaweedfs/mongo/livekit, no `restart`, livekit volume not read-only.
- Caddy mounts `Caddyfile` from cwd but `make dev-up` runs from repo root → relative path `./Caddyfile` resolves to `deployments/Caddyfile` only because the compose file is in `deployments/`. Confusing.

### `C:\Users\Valerii\Desktop\hub\deployments\docker-compose.test.yml`
- Mongo lacks healthcheck; nats lacks JS storage volume. No backend/frontend — appears CI-only but **not referenced by `ci.yml`**.

---

## 2. Dockerfiles

### `C:\Users\Valerii\Desktop\hub\backend\Dockerfile`
- Multi-stage: golang:1.24-alpine → alpine:3.20. Good.
- Final image installs `postgresql16-client mongodb-tools` for backup/restore — adds attack surface but needed for `backup/service.go` (which shells out to `pg_dump`/`mongodump`).
- **No `USER non-root`** — container runs as root. Big finding.
- `CGO_ENABLED=0` — good for static binary, but no `-ldflags="-s -w"` (no symbol strip) and no `-trimpath`.
- No build args for version/SHA → image has no provenance label.
- `HEALTHCHECK` baked in — but compose overrides it anyway.

### `C:\Users\Valerii\Desktop\hub\frontend\Dockerfile`
- node:22-alpine builder → nginx:1.27-alpine. Good.
- Nginx config built via inline `echo` heredoc — fragile, hard to review. Move to `nginx.conf` mounted COPY.
- **Runs as root** (default nginx image runs worker as nginx user, but master as root).
- No `npm ci --omit=dev` — installs full devDependencies before `npm run build`. Acceptable, but build cache layer bloats.

### `.dockerignore`
- backend: minimal — misses `.env`, `.air.toml`, `tmp/`, `coverage/`, `*.test`, `node_modules`. The `.git` exclude works but `*.md` is too broad (excludes any docs the binary might embed, currently none).
- frontend: minimal — misses `.env*`, `.cache`, `coverage`, `playwright-report`.

### Makefile
- `make build` → `go build -o bin/server ./cmd/server` — no build tags, no release flags. **No `build-prod`/`release` target exists**. README mentions it but not implemented.
- No `docker build` / `docker push` targets — CI handles directly.
- No `migrate` target.

---

## 3. Deployment configs (`deployments/*.yaml`)

### `C:\Users\Valerii\Desktop\hub\deployments\livekit.yaml`
- Has the new `departure_timeout: 120` fix. Good comment. **Local only**.
- `use_external_ip: false` — fine for docker host networking.

### `C:\Users\Valerii\Desktop\hub\deployments\livekit-prod.yaml`
- **STALE/UNUSED**: keys `devkey: secret` (dev creds in a file named "prod"). 
- Not mounted by `docker-compose.prod.yml` (heredoc generates config inline instead). File should either be deleted or wired in — currently misleading.
- **Missing** `departure_timeout: 120`.

### `seaweedfs-s3.json`
- Hardcoded creds `hub` / `hubsecret`. Mounted as `./seaweedfs-s3.json` in prod compose but creds don't match env-driven `${S3_ACCESS_KEY}` / `${S3_SECRET_KEY}` → **prod backend will fail to auth against SeaweedFS** unless `.env` happens to use `hub`/`hubsecret` (which violates `.env.example`'s `CHANGE_ME_STRONG_PASSWORD`).
- Should be templated or generated from env.

### Caddyfile.prod
- Single-domain config to `770470.cloud4box.ru`. No HSTS header, no `encode gzip`, no client IP header forwarding to backend. WS proxy fine.

---

## 4. Migrations

- `backend/ent/migrate/` is **generated code only** (migrate.go, schema.go). No Atlas SQL migration files, no `migrations/` dir.
- On boot, `database.NewPostgresClient` calls `client.Schema.Create(ctx)` → **auto-migrate in production**. Comment in code says "Use Atlas for production" but Atlas is not wired up.
- Risk: schema drift / accidental destructive change on deploy. No `WithDropColumn`, but column rename = drop+add silently in Ent.

---

## 5. CI/CD

### `.github/workflows/ci.yml`
- Triggers on `push: master` only — **no PR checks**, no branch protection signal.
- Builds only (no tests, no lint, no swagger, no docker build). Frontend `npm test` never runs.
- **No security scanning** (Trivy, gosec, govulncheck, npm audit, dependabot config).
- No SBOM, no image signing (cosign).
- Go version mismatch: workflow `1.25`, Dockerfile `1.24`. Pick one.

### `.github/workflows/deploy.yml`
- Triggers on every push to master → auto-deploy without staging/manual approval.
- `git reset --hard origin/master` on server overwrites local state → forces server to mirror repo. OK.
- **`sed -i 's/DEV_MODE: "false"/DEV_MODE: "true"/'`** in production deploy script (line 64). This flips production into dev mode — exposes `/api/dev/*` endpoints, disables auth rate limit. Almost certainly a leftover hack. **Must remove**.
- `docker rm -f frontend` then `docker run -d --name frontend` outside of compose → frontend lifecycle diverges from compose state. Will confuse `docker compose ps`.
- `docker system prune -f --volumes` is **dangerous**: prunes any unused named volume → if a service is down at prune time you can lose data. At minimum drop `--volumes`.
- Tags: only `:latest` + `:${sha}` — no semver, no rollback alias.
- No deploy status check / smoke test after restart.

---

## 6. `.env.example` — Missing Env Documentation

Compared `getEnv` / `os.Getenv` calls in `backend/internal/` against `deployments/.env.example`:

| Variable | Where used | In .env.example? |
|---|---|---|
| `INTEGRATION_ENCRYPTION_KEY` | `config.go:183` (HI-05, required in prod) | **MISSING** — prod will fail-start |
| `VAPID_PUBLIC_KEY` | `config.go:178` (Web Push) | **MISSING** |
| `VAPID_PRIVATE_KEY` | `config.go:179` | **MISSING** |
| `VAPID_SUBJECT` | `config.go:180` | **MISSING** |
| `CORS_ORIGINS` | `middleware/cors.go:24` | **MISSING** (docs/wiki/DEPLOY.md mentions it but example doesn't) |
| `ENVIRONMENT` | `config.go:137` (production gate) | **MISSING** |
| `DEV_MODE` | `config.go:136` | **MISSING** from example, only in compose |
| `FIBER_IMMUTABLE` | `config.go:135` | **MISSING** |
| `SERVER_PORT` | `config.go:132` | **MISSING** |
| `DATABASE_URL` / `MONGODB_URL` / `REDIS_URL` / `NATS_URL` | config.go | Built from parts in compose; not exposed as overrides in example |
| `S3_ENDPOINT`, `S3_USE_SSL` | config.go | **MISSING** |
| `LIVEKIT_URL` | config.go:173 | **MISSING** |
| `MONGODB_TEST_URL` | `calls/chat_test.go:24` | **MISSING** (test-only, ok) |

Net: at least **10 required env vars** are undocumented. `INTEGRATION_ENCRYPTION_KEY` is the worst — without it production won't boot, and operators learn this from a crash.

---

## 7. Backup / Recovery

- `backend/internal/backup/service.go` shells `pg_dump` + `mongodump` → zip → S3 (`hub-backups` bucket).
- Bucket name `"hub-backups"` is **hardcoded** in `main.go:557` — not env-driven.
- `s3Client` passed as `nil` from main (line 557): `backupSvc := backupmod.NewBackupService(db, nil, ...)` → **backups never upload to S3** in current code. PerformBackup will nil-deref or no-op on S3 call.
- Weekly ticker `7 * 24 * time.Hour` — but on every process restart the timer resets (no cron, no persistent schedule). With prod re-deploys on every master push, backups effectively never trigger.
- `maxBackups: 4` retention — only deletes by record, not by S3 lifecycle.
- No restore drill documentation.

---

## 8. Observability

- **Logs**: slog JSON to stdout — good for docker logging driver.
- **Metrics**: `otel` is only a transitive dep (indirect). No `prometheus` package, no `/metrics` endpoint, no instrumentation.
- **Tracing**: not wired (no otel tracer provider, no exporter).
- **Health**: `/api/health` checks postgres+redis+mongo (good). No `/api/ready` vs `/api/live` separation. NATS not pinged.
- No structured request_id propagation visible in logs beyond middleware (would need verification).
- No alerting / SLO definitions. README v2.0 promises "DevOps (monitoring, K3s)" — not delivered.

---

## 9. `.gitignore` & Sensitive Data

- `.env` family ignored.
- `backend/ent/*` excluded except schema/generate.go — but generated code IS checked in (since CI runs `go generate`), so this might cause divergence.
- `documentation/` ignored.
- **Actual leak in repo**: `docker-compose.local.yml` lines 115-117 contain a real-looking Resend API key (`re_QYvUBvFt_...`) and Telegram bot token. These are **committed**. Treat as compromised, rotate immediately.
- No `.env` files at repo root or in deployments — clean on that front.

---

## 10. Docs / Deploy Instructions

- `README.md:201` and `docs/wiki/DEPLOY.md:28` instruct `docker compose -f docker-compose.demo.yml up -d --build` — **no such file exists** in the repo (only `local`, `prod`, `test`, plain `deployments/docker-compose.yml`). Deploy doc is broken.
- README claims "K3s" in production — no k8s manifests/Helm charts in `deployments/`.
- `docs/wiki/DEPLOY.md` env table is incomplete (matches `.env.example` gaps).
- No RUNBOOK.md, no on-call playbook, no rollback procedure documented.

---

## Operations Runbook Readiness: **35 / 100**

Reasoning:
- Boot path works in dev (+25)
- Healthchecks in compose, structured logging, JWT/secret validation in config (+10)
- Production deploys flip DEV_MODE=true via sed in CI (-15)
- LiveKit prod config diverges from local fix (-10)
- Backups don't actually upload (nil s3Client) (-10)
- 10+ undocumented env vars; INTEGRATION_ENCRYPTION_KEY missing (-10)
- No metrics, no tracing, no alerting (-10)
- Deploy doc references non-existent compose file (-5)
- Real-looking secrets committed in dev compose (-10)
- Auto-migrate in prod, no Atlas wired (-5)

---

## Top Recommendations (priority order)

1. **REVOKE & ROTATE** the Resend API key and Telegram bot token visible in `docker-compose.local.yml:115-117`; move to `.env.local` (gitignored).
2. **Sync `livekit.yaml` ↔ `docker-compose.prod.yml` heredoc** — add `departure_timeout: 120` to the prod inline config, or stop generating inline and mount `livekit-prod.yaml` (after fixing its dev creds). Otherwise the call-death bug returns in prod.
3. **Remove the `sed` line** in `.github/workflows/deploy.yml:64` that flips `DEV_MODE=true` in production. Critical security regression.
4. **Add to `deployments/.env.example`**: `INTEGRATION_ENCRYPTION_KEY`, `VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY`, `VAPID_SUBJECT`, `CORS_ORIGINS`, `ENVIRONMENT=production`, `DEV_MODE=false`, `S3_ENDPOINT`, `LIVEKIT_URL`, `BACKUP_S3_BUCKET`.
5. **Fix backup service**: instantiate real S3 client in `cmd/server/main.go` (currently `nil`) or remove the weekly job — current state is silent failure.
6. **Pin `seaweedfs:latest` to a specific tag**; unify mongo:7↔8 between local/prod.
7. **Add `USER nonroot` (UID 10001)** in both Dockerfiles; add `-ldflags="-s -w" -trimpath` to Go build.
8. **CI**: add `go test`, `golangci-lint`, `npm test`, `govulncheck`, `trivy image` scans; trigger on PRs not just push to master.
9. **Pre-deploy smoke test**: after compose up, curl `/api/health`, fail deploy on non-200; gate `docker system prune --volumes` behind a flag.
10. **Remove auto-migrate in production**: wire Atlas (`atlas migrate apply`) as an explicit deploy step; keep `Schema.Create` only behind `DEV_MODE=true`.
11. **Fix docs**: replace `docker-compose.demo.yml` references in `README.md` and `docs/wiki/DEPLOY.md` with actual file (`deployments/docker-compose.prod.yml`); add RUNBOOK.md for rollback + restore drill.
12. **Decide on K3s**: either ship `deployments/k3s/*.yaml` (README promise) or remove the claim.
13. **SeaweedFS S3 creds**: template `seaweedfs-s3.json` from `${S3_ACCESS_KEY}/${S3_SECRET_KEY}` (entrypoint script with `envsubst`) — currently mismatch will break prod.
14. **Observability gap**: add `/metrics` (prometheus client_golang) + basic counters (HTTP, NATS, JetStream consumer lag); wire `otel` tracer (already in indirect deps).
15. **Caddyfile.prod**: add `encode gzip zstd`, security headers (HSTS, X-Content-Type-Options), and `header_up X-Real-IP {remote_host}`.
