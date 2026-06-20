## Summary

<!-- What does this PR change and why? Reference PLAN.md task or FEATURE-BACKLOG ID if applicable. -->

## Type

- [ ] feat / fix / refactor / test / docs / chore / ci / perf

## Scope

- [ ] backend
- [ ] frontend
- [ ] infra / deploy
- [ ] docs / planning

## Test plan

- [ ] `cd backend && go test -race -count=1 ./...`
- [ ] `cd frontend && npx tsc --noEmit && npx vitest run`
- [ ] Manual / E2E (describe):

## Security checklist

- [ ] No secrets, tokens, or `.env` values in diff
- [ ] No `DEV_MODE=true` in production paths
- [ ] User-controlled URLs use safe HTTP client (SSRF)
- [ ] Webhook signatures not skipped

## Docs

- [ ] Updated `docs/` or `.planning/` if behavior/API changed
