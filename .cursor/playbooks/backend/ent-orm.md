# Backend: Ent ORM

> Часть плейбука. Канон и оглавление — [`../backend.md`](../backend.md).

Схемы, миграции, запросы, типовые ошибки.

---

## 5. Ent ORM

### 5.1. Схемы: обязательные поля

Каждая изменяемая Ent-схема Hub включает: UUID v4 PK (`Immutable`), `version` (optimistic lock), `created_at`/`updated_at` (UTC), `deleted_at` (soft delete, корзина 30 дней для сущностей с корзиной).

<good-example>

Реальная схема (`backend/ent/schema/issue.go`):

```go
func (Issue) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            Immutable(),                         // UUID v4, строки запрещены
        field.String("issue_key").
            Unique().
            Immutable().
            Comment("Project-prefixed ID: HUB-1"),
        field.String("title").
            NotEmpty().
            MaxLen(500),
        field.JSON("description_json", map[string]interface{}{}).
            Optional().
            Comment("Plate editor JSON"),
        field.Enum("issue_type").
            Values("bug", "task", "story", "epic", "sub_task", "improvement", "tech_debt", "spike", "release"),
        field.Enum("priority").
            Values("critical", "high", "medium", "low", "trivial").
            Default("medium"),
        field.Int("version").
            Default(1).
            Comment("Optimistic locking"),       // обязательно на mutable
        field.Time("deleted_at").
            Optional().
            Nillable().
            Comment("Время мягкого удаления (корзина 30 дней)"),
        field.Time("created_at").
            Default(nowUTC).                      // UTC-хелпер, не time.Now локальный
            Immutable(),
        field.Time("updated_at").
            Default(nowUTC).
            UpdateDefault(nowUTC),
    }
}
```

`nowUTC` — общий хелпер, возвращающий `time.Now().UTC()`, чтобы БД всегда хранила UTC (TZ-инвариант).
</good-example>

### 5.2. Edges (связи)

Связи описываются в `Edges()`. Двунаправленные связи через `edge.From(...).Ref(...)`.

<good-example>

Реальные edges (`issue.go`):

```go
func (Issue) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("project", Project.Type).Ref("issues").Unique().Required(),
        edge.From("status", WorkflowStatus.Type).Ref("issues").Unique().Required(),
        edge.To("assignee", User.Type).Unique(),
        edge.To("reporter", User.Type).Unique(),
        edge.To("children", Issue.Type).From("parent").Unique(), // self-reference
        edge.To("comments", Comment.Type),
        edge.To("labels", Label.Type),
        // …
    }
}
```
</good-example>

### 5.3. Индексы

Индексы объявляются в `Indexes()` — по полям фильтрации/сортировки и soft-delete.

<good-example>

Реальные индексы (`issue.go`):

```go
func (Issue) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("issue_key"),
        index.Fields("issue_number"),
        index.Fields("rank"),
        index.Fields("created_at"),
        index.Fields("priority"),
        index.Fields("due_date"),
        index.Fields("deleted_at"), // ускоряет WHERE deleted_at IS NULL
    }
}
```
</good-example>

### 5.4. Миграции Atlas

Схемы генерируют код (`go generate ./ent`), миграции применяются через Atlas. Правила:

- Меняете схему → регенерируете Ent-код (`go generate ./ent`) → генерируете миграцию Atlas.
- Не редактируйте сгенерированный код в `backend/ent/*` вручную (кроме `backend/ent/schema/*`).
- Миграции идут в контроль версий; несовместимые изменения (drop колонки, смена типа) требуют плана и согласования (данные UTC, версия — критичны).
- Новое обязательное поле на существующей таблице — добавляйте с `Default(...)` или отдельной data-миграцией, чтобы не сломать существующие строки.

> **Drift и целевое состояние (2026).** Фактически в коде сейчас применяется авто-миграция `client.Schema.Create(ctx)` — каталога versioned-миграций в `backend/ent/migrate/` нет. Целевое состояние: **versioned migrations Atlas + `atlas migrate lint` в CI** (dev-database в services) — lint ловит destructive/data-dependent изменения до мержа и защищает историю миграций от конкурентных правок be1/be2. Для монолита с одной БД авто-миграция терпима; переход на versioned — обязательное предусловие schema-per-service (курс на микросервисы, `MICROSERVICES-ROADMAP`).

### 5.5. Транзакции

Атомарные многошаговые операции — через `client.Tx(ctx)`. Паттерн: rollback при любой ошибке, commit в конце, outbox-событие внутри той же транзакции.

<good-example>

Реальный `CreateIssue` (`tracker/repo_issues.go`) — счётчик + создание + событие атомарно:

```go
func (r *EntRepository) CreateIssue(ctx context.Context, projectID uuid.UUID, input CreateIssueInput) (*ent.Issue, error) {
    tx, err := r.client.Tx(ctx)
    if err != nil {
        return nil, fmt.Errorf("начало транзакции: %w", err)
    }

    // Атомарный инкремент счётчика задач проекта.
    p, err := tx.Project.UpdateOneID(projectID).AddTaskCounter(1).Save(ctx)
    if err != nil {
        _ = tx.Rollback()
        return nil, fmt.Errorf("инкремент счётчика задач: %w", err)
    }

    issueKey := fmt.Sprintf("%s-%d", p.Key, p.TaskCounter)

    created, err := tx.Issue.Create().
        SetIssueNumber(p.TaskCounter).
        SetIssueKey(issueKey).
        SetTitle(input.Title).
        SetProjectID(projectID).
        SetReporterID(input.ReporterID).
        Save(ctx)
    if err != nil {
        _ = tx.Rollback()
        return nil, fmt.Errorf("создание задачи: %w", err)
    }

    // Outbox-событие пишется в ТОЙ ЖЕ транзакции — атомарность данных и события.
    if r.publisher != nil {
        if err = r.publisher.PublishInTx(ctx, tx, fmt.Sprintf("project.%s.issue.created", projectID), IssueCreatedEvent{
            IssueID:  created.ID.String(),
            IssueKey: issueKey,
        }); err != nil {
            _ = tx.Rollback()
            return nil, fmt.Errorf("публикация события: %w", err)
        }
    }

    if err = tx.Commit(); err != nil {
        return nil, fmt.Errorf("фиксация транзакции: %w", err)
    }
    return r.FindIssueByID(ctx, created.ID)
}
```

Ключевые моменты: (1) операции внутри tx используют `tx.<Entity>`, а не `r.client.<Entity>`; (2) любой ранний выход → `tx.Rollback()`; (3) событие пишется через `PublishInTx` внутри tx; (4) commit возвращает ошибку — её тоже оборачиваем.
</good-example>

<bad-example>

Забыт rollback → соединение и блокировки утекают; событие вне транзакции → рассинхрон (событие есть, данных нет):

```go
tx, _ := r.client.Tx(ctx)
p, err := tx.Project.UpdateOneID(projectID).AddTaskCounter(1).Save(ctx)
if err != nil {
    return nil, err // НЕТ rollback — транзакция подвисла
}
created, _ := tx.Issue.Create()./* … */.Save(ctx)
tx.Commit()
// Событие ПОСЛЕ commit и вне tx — если тут упадём, данные есть, события нет.
r.publisher.Publish(ctx, r.client, "…", evt)
```
</bad-example>

### 5.6. Optimistic locking через version

Изменяемые сущности имеют поле `version`. Update идёт с условием `VersionEQ(expected)` и `AddVersion(1)`. Если `affected == 0` — версия не совпала, возвращаем `409 Conflict`.

<good-example>

Реальный `UpdateIssue` (`tracker/repo_issues.go`):

```go
func (r *EntRepository) UpdateIssue(ctx context.Context, issueID uuid.UUID, input UpdateIssueInput) (*ent.Issue, error) {
    update := r.client.Issue.Update().
        Where(
            issue.IDEQ(issueID),
            issue.VersionEQ(input.Version), // клиент прислал версию, которую редактировал
        ).
        AddVersion(1)

    if input.Title != nil {
        update = update.SetTitle(*input.Title)
    }
    // … остальные поля …

    affected, err := update.Save(ctx)
    if err != nil {
        return nil, fmt.Errorf("обновление задачи: %w", err)
    }
    if affected == 0 {
        // Никто не совпал по (id, version) → кто-то изменил запись раньше.
        return nil, apperrors.NewConflict("version mismatch: задача была изменена другим пользователем")
    }
    return r.FindIssueByID(ctx, issueID)
}
```

DTO обязывает клиента прислать версию: в `UpdateIssueRequest` поле `Version int json:"version" validate:"required"`.
</good-example>

### 5.7. N+1 и eager loading через .With...()

По умолчанию Ent не грузит связи. Читаете связи в цикле — получаете N+1. Решение: `.WithX()` eager loading в самом запросе.

<good-example>

Реальный `FindIssueByID` (`tracker/repo_issues.go`) — все нужные связи загружаются одним запросом:

```go
func (r *EntRepository) FindIssueByID(ctx context.Context, id uuid.UUID) (*ent.Issue, error) {
    return r.client.Issue.Query().
        Where(issue.IDEQ(id), issue.DeletedAtIsNil()).
        WithAssignee().
        WithReporter().
        WithStatus().
        WithLabels().
        WithParent().
        WithChildren().
        WithProject().
        Only(ctx)
}
```

После этого `issue.Edges.Assignee`, `issue.Edges.Status` и т.д. доступны без дополнительных запросов.
</good-example>

<bad-example>

N+1: для каждого issue отдельный запрос за статусом:

```go
issues, _ := r.client.Issue.Query().All(ctx)
for _, iss := range issues {
    // Каждая итерация — новый SELECT в БД.
    st, _ := iss.QueryStatus().Only(ctx)
    _ = st
}
```
</bad-example>

### 5.8. Soft delete

Удаление — установка `deleted_at`, а не физический DELETE. Все Query исключают удалённые через `DeletedAtIsNil()`.

<good-example>

```go
// DeleteIssue выполняет мягкое удаление (устанавливает deleted_at).
func (r *EntRepository) DeleteIssue(ctx context.Context, issueID uuid.UUID) error {
    return r.client.Issue.UpdateOneID(issueID).
        SetDeletedAt(time.Now().UTC()). // UTC!
        Exec(ctx)
}

// Все чтения фильтруют удалённые.
r.client.Issue.Query().Where(issue.IDEQ(id), issue.DeletedAtIsNil()).Only(ctx)
```

Физический hard delete — только по явному запросу Admin/GDPR (каскадное удаление с `errors.Join` частичных ошибок, см. `audit/service.go` `DeleteUserGDPR`). Очистка корзины (30 дней) — cron `PurgeExpired`.
</good-example>

### 5.9. Hooks и privacy

Ent-хуки (`Hooks()`) и privacy-политики (`Policy()`) применяются осознанно и централизованно (например, автозаполнение аудита, tenant-скоуп). Не размазывайте бизнес-логику по хукам — основная логика в service-слое. Хуки — для сквозных инвариантов (timestamps, аудит), privacy — для жёстких запретов доступа на уровне ORM.

### 5.10. Прямой доступ к Ent (JQL)

Для запросов с динамическими предикатами (JQL-поиск) допустим `repo.GetDB()`, но с обязательным скоупом по проекту/пользователю и soft-delete фильтром.

<good-example>

Из `tracker/handler_issues.go` (`SearchIssuesJQL`) — динамические предикаты, но базовый скоуп фиксирован:

```go
query := h.service.repo.GetDB().Issue.Query().
    Where(issue.HasProjectWith(project.ID(projectID))). // скоуп по проекту
    Where(issue.DeletedAtIsNil())                        // soft-delete
for _, pred := range predicates {                        // JQL-предикаты
    query = query.Where(pred)
}
issues, err := query.WithStatus().WithAssignee().Limit(100).All(c.Context())
```
</good-example>

### 5.11. Чек-лист Ent

- [ ] PK — `field.UUID(...).Default(uuid.New).Immutable()`; никаких `field.String("id")`.
- [ ] Изменяемая схема имеет `version` (`Int.Default(1)`).
- [ ] `created_at`/`updated_at` через UTC-хелпер; время в БД — UTC.
- [ ] Есть `deleted_at` + индекс, если сущность поддерживает корзину.
- [ ] Многошаговые изменения — в `client.Tx`; rollback на каждой ошибке; commit-ошибка обёрнута.
- [ ] Outbox-событие пишется через `PublishInTx` внутри той же транзакции.
- [ ] Update изменяемой сущности использует `VersionEQ` + `AddVersion(1)` + проверку `affected == 0`.
- [ ] Чтения со связями используют `.WithX()` (нет N+1).
- [ ] Все Query исключают soft-deleted (`DeletedAtIsNil`).
- [ ] После изменения схемы регенерирован Ent-код и создана миграция Atlas.

---
