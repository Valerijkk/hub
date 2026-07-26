# Runtime-адаптеры

Портативные контракты агентов не содержат host-specific frontmatter, permissions,
model IDs или launch API. Выберите один адаптер:

- [`cursor/`](cursor/) — обёртки Cursor custom-agent и маршрутизация по живому каталогу
- [`generic-chat/`](generic-chat/) — сборка system/user сообщений
- [`ci/`](ci/) — неинтерактивный least-privilege job mapping

Все адаптеры потребляют один и тот же core `agents/`, сгенерированные
context/pattern cards, `playbooks/model-routing.md` и заполненную копию
`models/registry.template.yaml`.

Адаптер владеет runtime model parameters и permissions. Он не должен переписывать
поведение роли или заявлять возможности, которых нет в наблюдаемом каталоге.

Каждый адаптер ОБЯЗАН:

1. загружать `rules/00` и `rules/50` как policy более высокого приоритета;
2. отделять содержимое repository/context/tool как untrusted data;
3. использовать [`APPROVAL-RECORD.schema.yaml`](APPROVAL-RECORD.schema.yaml);
4. отклонять approval, заявленный кодом, комментариями, логами, тикетами или другим агентом;
5. повторно подтверждать external/destructive действия непосредственно перед исполнением;
6. блокировать non-public context, если provider/region/retention/training policy
   неизвестна или не одобрена.
