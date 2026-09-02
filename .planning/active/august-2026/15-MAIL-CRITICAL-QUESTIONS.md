# Mail: критические вопросы для founder

## 1. Восстановление admin auth boundary

После удаления Mail Broadcast обязательный escalation gate блокирует
восстановление ранее принятой Fiber v3.0.0-beta.3 group-based цепочки в
`backend/internal/mail/admin_handler.go` и
`backend/internal/mail/mail_domain_control_handler.go`.

Свежая Docker-проверка фиксирует `401/503`. Нужное точное восстановление:

```text
auth -> preflight (только domain routes) -> requireAdmin -> business
```

Broadcast остаётся удалённым и должен отвечать `404`.

**Вопрос/решение founder:** дать явное security-boundary разрешение на это
восстановление прежней auth chain.

**Граница:** этот вопрос не блокирует другую незаблокированную реализацию или
QA, но блокирует зелёный `backend-unit` и финальную локальную приёмку.

## 2. Внешний Stalwart/JMAP production cutover

Для реальной внешней отправки/приёма нужны предоставленные ops-входы и
разрешение:

- реальный domain и IP;
- DNS: MX, SPF, DKIM и DMARC;
- достижимые SMTP, JMAP и TLS;
- secrets;
- ops authorization на cutover.

**Вопрос/решение founder:** когда и кем будут предоставлены эти входы и
разрешение на внешний cutover?

**Граница:** до этого локальная Docker/manual Mail+Calendar проверка с Fake
продолжается; блокируется только внешняя отправка/приём, а не прочая
незаблокированная реализация или QA.

Provenance: `dispatcher-enforced`.
