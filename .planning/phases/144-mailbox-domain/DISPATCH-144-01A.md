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

- Historical packet sources: `MAIL-SPEC.md` 0.1.7 §§1–3, 5, 7. **Do not redispatch this
  surface:** replacement work must use independently accepted `ENGINE-CONTRACT-0.1.8.md`;
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

[HISTORICAL 0.1.7 PORT — DO NOT EXECUTE]

The following interface is retained only as source-history evidence. It is not a current
dispatch contract; the exact replacement surface and DTO fields are exclusively in
`docs/mail/ENGINE-CONTRACT-0.1.8.md`:

```go
type MailEngine interface {
    Capabilities(context.Context) (EngineCapabilities, error)
    GetDomain(context.Context, string) (EngineDomainState, error)
    EnsureDomain(context.Context, DomainSpec) (EngineDomainState, error)
    EnsureAccount(context.Context, AccountSpec) (EngineAccount, error)
    ListFolders(context.Context, string) ([]EngineFolder, error)
    CreateFolder(context.Context, string, FolderSpec) (EngineFolder, error)
    RenameFolder(context.Context, string, string, string) error
    DeleteFolder(context.Context, string, string) error
    ListMessages(context.Context, EngineMessageQuery) (EngineMessagePage, error)
    GetMessage(context.Context, string, string) (EngineMessage, error)
    GetBlob(context.Context, string, string) (EngineBlob, error)
    Changes(context.Context, string, string) (EngineChangePage, error)
    EnumerateProjection(context.Context, string, string) (EngineEnumerationPage, error)
    SaveDraft(context.Context, EngineDraft) (EngineMessage, error)
    ReconcileDraft(context.Context, EngineDraftReconciliation) (EngineDraftResult, error)
    Submit(context.Context, EngineSubmission) (EngineSubmissionResult, error)
    ReconcileSubmission(context.Context, string, string) (EngineSubmissionResult, error)
    MoveMessages(context.Context, string, []string, string) error
    SetFlags(context.Context, string, []string, FlagPatch) error
    SetKeywords(context.Context, string, []string, KeywordPatch) error
    DeleteMessages(context.Context, string, []string, bool) error
    GetSieve(context.Context, string, string) (EngineSieveState, error)
    UpsertSieve(context.Context, string, SieveScript) (EngineSieveState, error)
    ClassifyMessages(context.Context, EngineSpamTraining) (EngineSpamTrainingResult, error)
    Health(context.Context) (EngineHealth, error)
}
```

Semantic requirements:

- Все account/folder/message/blob/state/submission ids — opaque bounded strings.
- Engine search query/result DTO is structurally absent. Public search is a separate
  owner-filtered Hub/Meilisearch port and can hydrate only through account-scoped reads.
- `EngineDraft` содержит account id, предыдущую opaque engine version, stable Hub
  mutation id и письмо целиком только как request value; никакой persistence.
- `EngineDraftResult` различает applied, absent/unchanged и indeterminate; reconciliation
  не имеет права вызывать save.
- Historical `EngineSubmission` carried a stable key plus a second message value. Mail
  0.1.8 replaces it with account/draft/message references, confirmed version, stable key
  and final-MIME ceiling only; no second content value is permitted.
- `EngineSubmissionResult` различает `sent`, `absent`, `indeterminate` и несет result ids
  только когда они подтверждены; оба success id — непрозрачные bounded engine strings, не
  Hub UUID, а zero value не должен выглядеть как success.
- `AccountSpec` содержит immutable stable provisioning key, normalized reserved local part,
  selected domain, captured generation/hash, enabled state и exact primary/alias set.
  `EnsureAccount` idempotent на key + desired spec;
  password/token структурно отсутствуют во входе и выходе.
- `DomainSpec` содержит stable provisioning key, normalized ASCII domain, captured desired
  generation/hash и desired present/enabled state. `EnsureDomain` возвращает только opaque
  id, public DNS requirements, verification/DKIM and applied/indeterminate state;
  `GetDomain` read-only reconciles unknown outcome. DKIM private material отсутствует.
- `EngineSieveState` раскрывает только fixed script name, active flag и bounded Hub
  generation/hash marker. `GetSieve` — read-only reconciliation; `UpsertSieve` возвращает
  applied/indeterminate state и не принимает public raw Sieve.
- `EngineSpamTraining` связывает account, unique bounded message ids, stable receipt/message
  command ids и явное `spam|ham`.
  Result различает per-message accepted/rejected/indeterminate training от folder/flag
  mutation; обычный move не может сконструировать accepted feedback. Indeterminate command
  cannot be called again unless a later accepted port revision adds explicit reconciliation.
- `KeywordPatch` содержит только server-resolved opaque reserved markers with explicit
  add/remove sets. Public label name/id/raw keyword cannot construct it; hydrated messages
  expose bounded reserved markers separately from public flag enum.
- `Changes` возвращает bounded projection-only page с точными opaque `from_cursor` и
  `next_cursor`; invalid/expired cursor имеет отдельный стабильный error kind.
- `EnumerateProjection` возвращает bounded resumable enumeration page, привязанную к
  одному pre-enumeration composite base cursor. DTO не обещает frozen provider state:
  adapter обязан доказать стабильный resume boundary и последующий Changes
  catch-up. Missing anchor, unsupported stable enumeration или expired base history —
  отдельный restart-required error kind, а не пустой cursor или тихая новая enumeration.
- Cursor/checkpoint/base-cursor DTO остаются adapter/repository-only: они не являются
  публичными API/event/log полями.
- Error taxonomy различает not-found, conflict/state-mismatch, invalid-cursor,
  unavailable, timeout-indeterminate, rejected-policy и unauthorized-account, не включает
  body, адреса, секреты или сырой ответ провайдера.
- DTO имеют явные лимитируемые slices/maps; никакого `any`/`map[string]any` для domain
  contract и никаких JMAP-specific полей/имен.
- Hydrated message DTO несёт server-derived projection origin enum. Только подтверждённая
  submission/draft provenance может подавить inbound event; `From`/alias или Sent folder
  сами по себе никогда не кодируются как confirmed-owner proof, unknown остаётся
  отдельным kind.

[REQUIRED ORACLES]

External-package тест `package mail_test` должен доказать:

1. Независимый consumer может реализовать интерфейс без импорта internal adapter package.
2. Draft expected version + mutation id и submission stable key доступны типобезопасно.
3. Draft/submission reconciliation results не смешивают definitive absence с
   indeterminate timeout и имеют отдельные typed result shapes. Submission success требует
   два непустых independently bounded opaque id и не навязывает UUID grammar.
4. Folder/message queries всегда несут account id; message ids без account контекста не
   являются допустимым вызовом.
5. Flag patch различает отсутствие изменения от явного `false`.
6. Health/capabilities не содержат credential/token fields.
7. Stable sentinel/kind errors распознаются через `errors.Is`/`errors.As`, а safe string не
   содержит переданный synthetic secret/body/address.
8. Incremental и enumeration pages несут явные bounds, exact cursor/checkpoint metadata и
   typed invalid-cursor/restart-required outcomes без provider-specific объектов.
9. Spoofed owner address and mutable Sent membership cannot construct confirmed-owner
   origin; confirmed submission, draft and inbound-or-unknown remain distinct typed values.
10. Компиляционный assertion `var _ mail.MailEngine = (*consumerStub)(nil)` проходит.
11. Один provisioning key/spec допускает повторный ensure без второго logical account;
    измененный spec — typed conflict, secret fields структурно отсутствуют.
12. Sieve read/apply marker и spam/ham outcomes — bounded typed DTO; raw script/provider
    response отсутствуют, move-only нельзя представить как trained.
13. Domain ensure/get distinguish verified, DNS-required, absent and indeterminate; only
    public bounded DNS records exist and no private key/provider body can compile into DTO.
14. Reserved keyword add/remove is account-scoped and structurally distinct from public
    flags/labels; unknown keywords cannot construct a Hub label/event field.

[EXECUTION / PROVENANCE]

- TDD first, затем минимальная реализация; production behavior отсутствует в этом packet.
- Host tests/build/lint/validators запрещены; direct Docker/Compose не fallback. Если
  accepted launcher недоступен, verification = `NOT_RUN`, только статический diff review.
- No commit/push/PR/network/dependency install/generated code.
- max_depth 0; max_fan_out 0; retries 2/root cause; wall clock 45m.
- requested_model `gpt-5.6-terra`; reasoning `xhigh`; fallback none. При несовпадении
  фактической модели немедленно `BLOCKED`, файлы не менять.
- Orchestrator передает и фиксирует exact clean HEAD перед dispatch.

[ОТВЕТ]

На русском: summary; changed files; точный interface/DTO/error surface; static/runtime
verification с честным `NOT_RUN`; риски; requested/actual model provenance; unknowns.
