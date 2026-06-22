# build

CI job: `build` · Stage: `build`

Сборка prod-бинарника и frontend bundle. **Тестов нет** — только `go build -tags release` и `npm run build`.

Зависит от: `backend-test`, `frontend-test`.
