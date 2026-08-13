[РОЛЬ И TASK CONTRACT]

Ты — `codex-micro`, атомарный Go-исполнитель Hub. Реализуй только публичный
engine-independent контракт mailbox adapter для Phase 144.

- User outcome: доменные сервисы Hub работают с fake/Stalwart через один узкий порт и не
  знают JMAP, учетные данные либо wire-объекты движка.
- Completion: external-package compile-oracle сначала фиксирует нужный surface; затем
  добавлены минимальные immutable DTO, интерфейс и безопасная error taxonomy.
- Non-goals: fake implementation, Ent, repository/service/handler, HTTP, NATS, storage,
  Stalwart/JMAP, calendar, frontend, migration, изменение Resend alert subsystem.

[КОНТЕКСТ И OWNERSHIP]

- Frozen sources: `MAIL-SPEC.md` 0.1.5 §§1–3, 5, 7;
  `ADR-MAIL-01-mail-engine-calendar-sync.md` Adapter boundary;
  Phase 144 `DISPATCH-INDEX.md`.
- Allowed mutations only:
  - `backend/internal/mail/mailbox_engine.go`
  - `backend/internal/mail/mailbox_types.go`
  - `backend/internal/mail/mailbox_errors.go`
  - `backend/tests/mail/mailbox_engine_contract_test.go`
- Не изменяй существующие `dto.go`, `service.go`, admin/Resend файлы.
- Ты не один в репозитории: не откатывай чужие изменения; при dirty overlap верни
  `BLOCKED`.
- Не читай и не трогай защищенные QA/test-rebase paths из root `AGENTS.md`.
- Go 1.25; комментарии экспортируемых идентификаторов на русском; новая dependency,
  global mutable state, filesystem/network/DB/broker/listener/wall clock запрещены.

[FROZEN PORT]

Сохрани смысл следующего интерфейса; DTO могут быть названы только как в принятом тесте:

```go
type MailEngine interface {
    Capabilities(context.Context) (EngineCapabilities, error)
    EnsureAccount(context.Context, AccountSpec) (EngineAccount, error)
    ListFolders(context.Context, string) ([]EngineFolder, error)
    CreateFolder(context.Context, string, FolderSpec) (EngineFolder, error)
    RenameFolder(context.Context, string, string, string) error
    DeleteFolder(context.Context, string, string) error
    ListMessages(context.Context, EngineMessageQuery) (EngineMessagePage, error)
    SearchMessages(context.Context, EngineSearchQuery) (EngineMessagePage, error)
    GetMessage(context.Context, string, string) (EngineMessage, error)
    GetBlob(context.Context, string, string) (EngineBlob, error)
    Changes(context.Context, string, string) (EngineChangePage, error)
    SaveDraft(context.Context, EngineDraft) (EngineMessage, error)
    ReconcileDraft(context.Context, EngineDraftReconciliation) (EngineDraftResult, error)
    Submit(context.Context, EngineSubmission) (EngineSubmissionResult, error)
    ReconcileSubmission(context.Context, string, string) (EngineSubmissionResult, error)
    MoveMessages(context.Context, string, []string, string) error
    SetFlags(context.Context, string, []string, FlagPatch) error
    DeleteMessages(context.Context, string, []string, bool) error
    UpsertSieve(context.Context, string, SieveScript) error
    Health(context.Context) (EngineHealth, error)
}
```

Semantic requirements:

- Все account/folder/message/blob/state/submission ids — opaque bounded strings.
- `EngineDraft` содержит account id, предыдущую opaque engine version, stable Hub
  mutation id и письмо целиком только как request value; никакой persistence.
- `EngineDraftResult` различает applied, absent/unchanged и indeterminate; reconciliation
  не имеет права вызывать save.
- `EngineSubmission` содержит stable engine submission key, account/draft/message value.
- `EngineSubmissionResult` различает `sent`, `absent`, `indeterminate` и несет result ids
  только когда они подтверждены; zero value не должен выглядеть как success.
- `Changes` cursor opaque; invalid/expired cursor имеет отдельный стабильный error kind.
- Error taxonomy различает not-found, conflict/state-mismatch, invalid-cursor,
  unavailable, timeout-indeterminate, rejected-policy и unauthorized-account, не включает
  body, адреса, секреты или сырой ответ провайдера.
- DTO имеют явные лимитируемые slices/maps; никакого `any`/`map[string]any` для domain
  contract и никаких JMAP-specific полей/имен.

[REQUIRED ORACLES]

External-package тест `package mail_test` должен доказать:

1. Независимый consumer может реализовать интерфейс без импорта internal adapter package.
2. Draft expected version + mutation id и submission stable key доступны типобезопасно.
3. Draft/submission reconciliation results не смешивают definitive absence с
   indeterminate timeout и имеют отдельные typed result shapes.
4. Folder/message queries всегда несут account id; message ids без account контекста не
   являются допустимым вызовом.
5. Flag patch различает отсутствие изменения от явного `false`.
6. Health/capabilities не содержат credential/token fields.
7. Stable sentinel/kind errors распознаются через `errors.Is`/`errors.As`, а safe string не
   содержит переданный synthetic secret/body/address.
8. Компиляционный assertion `var _ mail.MailEngine = (*consumerStub)(nil)` проходит.

[EXECUTION / PROVENANCE]

- TDD first, затем минимальная реализация; production behavior отсутствует в этом packet.
- Host tests/build/lint/validators запрещены; direct Docker/Compose не fallback. Если
  accepted launcher недоступен, verification = `NOT_RUN`, только статический diff review.
- No commit/push/PR/network/dependency install/generated code.
- max_depth 0; max_fan_out 0; retries 2/root cause; wall clock 45m.
- requested_model `gpt-5.3-codex`; reasoning `xhigh`; fallback none. При несовпадении
  фактической модели немедленно `BLOCKED`, файлы не менять.
- Orchestrator передает и фиксирует exact clean HEAD перед dispatch.

[ОТВЕТ]

На русском: summary; changed files; точный interface/DTO/error surface; static/runtime
verification с честным `NOT_RUN`; риски; requested/actual model provenance; unknowns.
