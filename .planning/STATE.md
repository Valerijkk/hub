---
> **Current narrow deletion-source authority (2026-08-23; highest precedence):** accepted plan `3f7aa2c58407ac3a10ad427cfa86256e0a9d340a` is `DOCS_PLAN_STATIC_ACCEPTED / SOURCE_AUTHORITY_ACCEPTED` only for one future atomic no-test exact2 production packet: `backend/internal/mail/mailbox_deletion.go` and `backend/internal/mail/mailbox_deletion_fake.go`. Plan Contract/high, Security/xhigh and Lifecycle/high reviews are C/I/M `0/0/0`; `provenance_source: dispatcher-enforced`, no fallback, telemetry `UNKNOWN`. Source remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` until its own commit and immutable C/S/L triple `0/0/0`. Adapter capability proof is not required only for this provider-neutral ABI plus two standalone fakes exception, but remains mandatory for every concrete provider adapter. Persistence/Ent/schema/generation/migration/SQL/worker/storage drain/account finalization/User-GDPR/HTTP/NATS/Docker/runtime stay `BLOCKED` / `NOT_RUN`. Old blanket deletion-source-blocked or plan-pending claims are `CONSUMED/HISTORICAL` only within this exact2 scope; unrelated gates are unchanged.
> **Engine deletion-port adjudication (2026-08-23):** `STATIC_ACCEPTED / DOCS_DESIGN_ONLY`; former deletion-port-next selection is `CONSUMED/HISTORICAL`, never source authority. Freeze separate opaque future ports: `EngineAccountDeletionPort` (`DeleteAccount`, `ReconcileAccountAbsence`) and `MailObjectDeletionPort` (`DeleteObject`, `ReconcileObjectAbsence`); do not widen normal `MailEngine` or combine account delete with drain. Requests durably retain nonzero `DeletionID` plus account provisioning key+engine account ID or object bucket+key; opaque IDs use `MaxEngineOpaqueIDBytes` and bucket/key existing storage bounds. Results are `deleted|authoritatively_absent`, observations `present|authoritatively_absent`; bare `NotFound` preserves intent unless scoped authoritative absence is proven. No provider/raw IDs or causes are public/logged. Private intent/control is not owner projection; short lease transactions keep I/O outside locks, use exact token+generation+scope CAS, reclaim expired lease only, and stale token/zero rows cause no I/O/compensation. Object drain starts after `upload_drain_until` and attachment/reference terminality; account removal is last after exact child/object/receipt/control/activity manifest; hard-delete only after closed evidence. No Stalwart route/OAuth/DNS/OpenAPI assumption. Adapter capability and independent contract/security/lifecycle reviews are mandatory; future source is bounded ABI/fake then separately planned persistence/worker. Ent/source/generated/migration/SQL/tests/Docker/runtime and physical User/GDPR remain `BLOCKED` / `NOT_RUN`.
mail_calendar_inbound_p3_due_context_amendment_2026_08_23: >- P3-D0 STATIC_ACCEPTED / STATIC_SOURCE_ONLY / RUNTIME_NOT_RUN at immutable source 47faf3a705e5641f4ceb7ee365c3dfeeeace725d, exactly backend/ent/schema/calendarinboundmessagereceipt.go. It adds required immutable organization_id/mailbox_id UUID scalars and database nonzero checks, with no Ent edges or raw data; existing uniqueness, state-shape, and due behavior remain unchanged. Contract/high and Security/xhigh are APPROVED C/I/M 0/0/0. The D0-next route is CONSUMED/HISTORICAL. P3-C is BLOCKED on Mail Ent generation; 147-02B/147-02C and D1/D2 are BLOCKED on active Calendar generation/dependencies; P3-E is conjunction-blocked. No next unblocked P3 source is selected. Generation, migration, SQL, build, tests, Docker, and runtime are NOT_RUN; no readiness, merge, or product claim.
mail_calendar_inbound_prerequisites_2026_08_23: >-
  STATIC_SOURCE_ONLY / RUNTIME_NOT_RUN. P0 contract is independently accepted C/I/M 0/0/0 at
  0d4fa053d569631234c5a7f902341bb6a73a7d46 with locator-pairing fix
  5a253d166f83ce37e695566f1396ee16fe00b224; P1 CalendarInboundMessageReceipt schema is
  APPROVED C/I/M 0/0/0 at 4b7f322b629dd269cac7e05ba61f7ca03e45d96c; P2 Mail durable schema,
  projection carrier, and event/durable/V2/V3 proof are APPROVED C/I/M 0/0/0 at
  fc5dfe195b89963601e3f365e593469f73bed94f, 1ebbb129d62aeca695a2b028822101af9c77687f,
  and a61edbb086df8535349429438155fad015b29615. P3-A sealed port is STATIC_ACCEPTED / STATIC_SOURCE_ONLY
  at 05f87f87f6dc30bd57648030026faf5912add3e7 (Contract/high and Security/xhigh APPROVED C/I/M 0/0/0);
  P3-B is STATIC_ACCEPTED / STATIC_SOURCE_ONLY at 43a6c42af88524be2951966bdbfc83b422ecb7f9; CONSUMED/HISTORICAL: P3-C was the sole next pending/blocked Task 10 consumer/repository/service/wiring atom on Mail Ent generation. P3-D/E remain pending. Generated Ent artifacts are absent/uncommitted; generation, migration,
  SQL, compile/build, tests, lint/gofmt, parser/validator, Docker, and runtime are NOT_RUN. No
  NATS/worker, migration, merge, product, or readiness claim follows.
mail_calendar_inbound_p3_source_freeze_2026_08_23: >-
  P3-A is STATIC_ACCEPTED / STATIC_SOURCE_ONLY at 05f87f87f6dc30bd57648030026faf5912add3e7
  (`feat(mail): add sealed calendar part port`) for exact five files calendar_part_source.go,
  mailbox_engine.go, mailbox_types.go, mailbox_errors.go, mailbox_fake.go; independent Contract/high and
  Security/xhigh are APPROVED C/I/M 0/0/0. It accepts sealed `package mail` ABI/port only: no concrete
  CalendarPartSource, engine/provider/MIME/locator I/O, Ent, NATS, wiring, or runtime. P3-B is STATIC_ACCEPTED /
  STATIC_SOURCE_ONLY at 43a6c42af88524be2951966bdbfc83b422ecb7f9 for the receipt-first EventTransaction REPLY factor. CONSUMED/HISTORICAL: P3-C was the sole next pending/
  blocked atom on Mail Ent generation and owns concrete source plus five-second/pre-post semantics; P3-D/E
  follows Calendar generation; P3-E follows concrete engine/source. Historic 147-02D/02E GetMessage/full-
  body/MIME/direct-engine/provider/legacy-engine_message_id Calendar-ingestion wording is
  CONSUMED/HISTORICAL with no fallback. Stalwart remains blocked on Phase 145 authority. Generated Ent
  artifacts are absent/uncommitted; generation, migration, SQL, build, tests, lint/gofmt, parser/validator,
  Docker and runtime are NOT_RUN. No worker, NATS, merge, product, or readiness claim follows.
mail_144_task_6b_ent_sync_prepare: >-
  STATIC_ACCEPTED / EXACT_SHA_PASS / RUNTIME_NOT_RUN at effective HEAD 431022ed after source
  6e788cc7 and live-thread correction 431022ed. The exact effective manifest is
  backend/internal/mail/mailbox_ent_sync_prepare.go and authored contract
  backend/tests/mail/mailbox_ent_sync_prepare_contract_test.go. One short read-only Ent transaction
  takes one repository clock read and account ForUpdate under the complete current+persistent fence,
  current epochs and unexpired lease, with no provider/NATS/HTTP/random/write. Exact
  initial/resumed enumeration, incremental/catchup and valid empty pages are supported; stable
  length-delimited SHA-256 UUIDv8 folder/message/thread IDs reuse existing live identities with
  collision guards. Live-only confirmed sent submission/stable draft provenance normalizes
  untrusted origin; live ancestors/threads, per-projection duplicate label rejection with page-wide
  dedupe, all-and-only unused event candidates, pure classification, deep copies and content-free
  errors are enforced. Live soft-deleted reappearance reuses message UUID and tombstone previous
  state/ReceivedEventID, avoiding duplicate mail.received.v1; tombstone folders/flags are zero and
  labels still-live only. Already-deleted/unknown destroyed remains Projected=false and skips
  state/provenance work. Independent immutable Terra/high review of 431022ed passed C/I/M 0/0/0;
  provenance dispatcher-enforced, telemetry UNKNOWN. Generated Ent client/predicates are absent and
  launcher NOT_READY; generation, compile, tests, formatting/lint, Atlas/migration, SQL/DB, Docker
  and runtime NOT_RUN. Task 6C is pending; no Task 6/Phase 144/module/product/migration/cutover/merge
  acceptance follows.
atlas_migration_control_plane: >-
  MIGRATION_CONTROL_PLANE_STATIC / EXECUTABLE_ARTIFACT_NOT_RUN / RUNTIME_NOT_RUN.
  Ent schemas are the desired-state authority; backend/atlas.hcl declares only one local Atlas
  environment with launcher-supplied DATABASE_URL and task-owned ATLAS_DEV_URL, ent://ent/schema,
  and file://migrations. backend/migrations is the sole versioned production directory; Atlas is
  limited to deterministic diff/lint/checksum/apply and never becomes a second schema authority.
  Production applies only reviewed checked-in SQL plus atlas.sum; DEV_MODE=true may retain
  Schema.Create for development only. Generated Ent Go remains uncommitted and launcher-produced;
  backfill is a separate reviewed forward-only artifact, and direct host commands are forbidden.
  Launcher NOT_READY; Ent generation, Atlas diff, lint, checksum, apply, database replay, and
  runtime validation are all NOT_RUN. No production migration is authorized.
mail_144_task_5_owner_ent_repository: >-
  STATIC_ACCEPTED / RUNTIME_NOT_RUN at head ea0808a7. Task 5 adds only the
  test-first `backend/tests/mail/mailbox_ent_owner_contract_test.go` and private Mail Ent
  repository root/owner resolver. The public constructor has the frozen signature
  NewMailboxEntRepository(*ent.Client, MailboxRepositoryClock) (*MailboxEntRepository, error),
  rejects nil and typed-nil dependencies, and keeps the client/clock private. Owner resolution
  obtains deterministic locks in Organization -> active User -> owner-bound MailAccount -> primary
  MailIdentity order, returns the existing valid not_configured record only for an active actor
  with no account, and otherwise fails closed with content-free typed errors on foreign, inactive,
  deleting, ambiguous, cross-linked, or inconsistent durable rows. Returned engine account,
  address, recovery and projection values are copied scalars; provider/network I/O is absent.
  The clock permits only non-zero exact-second UTC values before durable use. Independent review
  accepted this narrow manifest after the cross-organization and provisioning fixes; global
  unique mail_accounts.user_id prevents ambiguous owners and corrupted multiplicity fails closed.
  Generated Ent Go remains absent until Task 9 and launcher NOT_READY; RED/GREEN,
  compile, Ent generation, database, Docker, lint, formatting, migration and runtime evidence are
  NOT_RUN. No API/UI/wiring, product, phase, release or migration authority follows.
mail_144_task_6s_sync_lease_schema: >-
  STATIC_ACCEPTED / RUNTIME_NOT_RUN at source commit c8274bd2. Task 6S owns only the
  test-first MailAccount schema contract/source prerequisite: required Int64 sync_lease_epoch
  immediately after sync_epoch with default 1 and minimum 1; the existing non-negative CHECK
  retains every prior clause and adds sync_lease_epoch >= 1; and appended
  mail_accounts_sync_due_idx over (next_sync_at, id) has exact predicate
  status = 'active' AND next_sync_at IS NOT NULL. The generation will be carried by future
  MailboxSyncFence and MailboxSyncWorkerDueAccount; the first lease remains 1 and only distinct
  durable ownership replacement increments it. Task 6 depends on this prerequisite. This packet
  creates no sync-page receipt; that design is separately required before the apply adapter.
  Independent static review found no issues in this narrow manifest. Launcher NOT_READY; Ent generation, Atlas
  diff/migration, compile, SQL/DB, Docker, formatting/lint, and runtime are all NOT_RUN. No
  migration or cutover authority follows.
calendar_143_04i_inbound_reply_source_receipt_sync: >-
  SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT10_COMMIT_REQUIRED /
  IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN. This non-authoritative dirty exact10 docs receipt
  records immutable source 46f4d6024e6c5475cdd943af9979c4ffadc5474a (sole parent
  04dbde1a14e9499462dbc1ea8784583f83f0b40f, tree e3ce6097822f360a04717d0dce61597810ca2046,
  subject feat(calendar): add inbound reply service) and its ordered test/repository/service
  blobs in IMPLEMENTATION-143-04I.md. Contract/Security/Lifecycle source precommit and immutable
  EXACT_SHA_PASS are C/I/M 0/0/0; Contract Terra/high, Security Terra/xhigh, Lifecycle
  Terra/high; dispatcher-enforced/no fallback/telemetry UNKNOWN. Until one atomic manifest-equal
  exact10 receipt commit and immutable receipt triple PASS, this source is not STATIC_ACCEPTED.
  Only after activation do source status become STATIC_ACCEPTED / EXACT_SHA_PASS and lower 04I
  authority/source-unimplemented routes become CONSUMED/HISTORICAL. Runtime remains NOT_RUN;
  launcher/testframework external handoff, 04F-I, 03E, 04F-Base, Ent/generated source,
  generation/migration, HTTP/API/OpenAPI/UI and NATS/iTIP transport remain BLOCKED/NOT_READY;
  no next Calendar packet is selected or authorized.
calendar_143_04h_current_source_receipt: >-
  STATIC_ACCEPTED / EXACT_SHA_PASS only for immutable source
  73940810046bb3186f2746705a83565ab29368c5 (parent
  9a638f75745aa0fb516a1177a33cbd1c657432fe, tree
  9f994a8805b43a84c00ed50b8e5544739425edaa) and its ordered blobs
  80e19d3b918ae9ddbe3b3ea8803c6062c2f8a713,
  e0da66bf95d4b8193eb5d4034053a353bd154b66, and
  2b500e5cdda34d7d80d0ccc0ba9c94427167716f. Exact16 receipt
  f9f5d436c08737eeb5a3e965137f329f3ae7f627 is ancestral to HEAD. Contract,
  Security and Lifecycle reviews PASS C/I/M 0/0/0; author Terra/xhigh;
  dispatcher-enforced/no fallback; telemetry UNKNOWN. No next packet is selected.
  Generation, migration, adapters, handlers, API/UI, launcher and runtime remain
  NOT_READY or NOT_RUN.
mail_144_04i_a_current_source_receipt: >-
  STATIC_ACCEPTED / EXACT_SHA_PASS only for immutable source
  091f2ef286a71f506e1b048f5386fd37cefa06bc (parent
  6b27643fcde3b702fa13c5fcc9b4bfefcb9fc87a, tree
  b01be92acf01033935b85446454eaee16b84e01b) and ordered blobs
  79c2691971287e3a62c830a5d12c78699414bb86 and
  6c20fadf563d4d146b2fc1cbacdddea485266eb7. Exact10 receipt
  03674eace728642fc2c9b39d2e8e2c81bf016033 is ancestral to HEAD. Contract,
  Security and Lifecycle reviews PASS C/I/M 0/0/0; author Terra/xhigh;
  dispatcher-enforced/no fallback; telemetry UNKNOWN. No next packet is selected.
  Ent/concrete adapter, generation/migration, API/UI, launcher and runtime remain
  NOT_READY or NOT_RUN.
gsd_state_version: 1.0
milestone: v4
milestone_name: Hub One Super-app
status: mail-calendar-development-with-testframework
stopped_at: origin/master f21e3f77 integrated; Mail and Calendar implementation continues
last_updated: "2026-08-23T00:00:00+03:00"
testframework_master_integration: >-
  origin/master f21e3f77 is integrated into codex/mail-module-20260813. The accepted
  Python QA package, root Compose profile, unit relocation, and retired legacy runner
  deletions come from master unchanged. No tests, build, lint, Docker, generation, or
  runtime command was executed while resolving the merge. Existing Mail/Calendar
  static-source evidence remains separate from the framework receipts.
mail_calendar_docker_readiness: >-
  Current QA boundary: root `docker-compose.local.yml` exact services
  `backend-unit`, `frontend-unit`, `python-contract`, `python-api`,
  `python-integration`, `python-e2e-smoke`, `python-e2e-full`, and
  `python-load` are the sole Docker-only QA API. The committed Python package,
  lock, typed marker, QA Dockerfile and entrypoint are integrated from master;
  `python-contract` framework-only evidence is 181/181 PASS on 2026-08-23.
  This admits neither Mail/Calendar generation or migration nor their runtime:
  API/integration/e2e/load and all Mail/Calendar product evidence remain
  `NOT_RUN`. No merge, deploy, shared/production migration, or product-readiness
  claim follows.
qa_launcher_task_2b_sync: >-
  CONSUMED/HISTORICAL. This records the retired `testframework/runner/**` and
  `qa-gate` static-launcher route, including its immutable review receipts. The
  runner path is absent after the testframework cutover and is not a fallback.
  Current authority is only root Compose `run --rm <exact-service>` for the
  eight named QA services. The 181/181 `python-contract` result is framework
  selftest evidence only; API/integration/e2e/load and Mail/Calendar runtime
  remain `NOT_RUN`.
calendar_143_04h_source_receipt_sync_historical: >-
  Status: SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT16_COMMIT_REQUIRED /
  RUNTIME_NOT_RUN. This dirty exact16 documentation candidate grants no source acceptance.
  It records immutable source 73940810046bb3186f2746705a83565ab29368c5 (parent
  9a638f75745aa0fb516a1177a33cbd1c657432fe, tree
  9f994a8805b43a84c00ed50b8e5544739425edaa, subject feat(calendar): add occurrence
  cancellation service): authored-first test 80e19d3b918ae9ddbe3b3ea8803c6062c2f8a713,
  repository e0da66bf95d4b8193eb5d4034053a353bd154b66, then service
  2b500e5cdda34d7d80d0ccc0ba9c94427167716f. Contract Terra/high, Security Terra/xhigh
  and Lifecycle Terra/high precommit plus immutable exact-SHA reviews PASS C/I/M 0/0/0;
  author Terra/xhigh; dispatcher-enforced/no fallback; telemetry UNKNOWN. Only after one atomic
  exact16 commit equal to IMPLEMENTATION-143-04H.md's manifest and immutable triple PASS does
  this source become STATIC_ACCEPTED / EXACT_SHA_PASS. Then lower 04H authority, precommit,
  source-unaccepted and current-next clauses are CONSUMED/HISTORICAL. No next packet is selected;
  04J/04K remain independent, 04F-H owns future persistence, and Ent/adapters/handlers/API/UI/
  generation/migration/raw iTIP/launcher/runtime blockers remain NOT_READY or NOT_RUN.
calendar_143_04h_authority_acceptance_sync_historical: >-
  CONSUMED/HISTORICAL only for accepted 143-04H source receipt
  73940810046bb3186f2746705a83565ab29368c5. The former authority-acceptance route below
  is preserved as immutable provenance; it creates no future exact15 requirement, exact3
  dispatch, or source-unimplemented lifecycle. Status: AUTHORITY_ACCEPTANCE_SYNC_CANDIDATE /
  ATOMIC_EXACT15_COMMIT_REQUIRED /
  IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN. This dirty exact15 acceptance sync grants
  no authority before one atomic exact15 commit plus immutable acceptance review.
  Immutable candidate f2966bdd50631feaa210a42a232880b21ade7e3a (parent
  03674eace728642fc2c9b39d2e8e2c81bf016033, tree
  45411b0692c79677b34af485df0ae2f23f4957d9, subject docs(calendar): propose
  occurrence cancel authority) has immutable Contract Terra/high, Security Terra/xhigh
  and Lifecycle Terra/high EXACT_SHA_PASS C/I/M 0/0/0; dispatcher-enforced/no fallback,
  telemetry UNKNOWN. Only after the atomic sync commit and its immutable acceptance-commit
  review PASS does f296 become SOURCE_AUTHORITY_ACCEPTED /
  EXACT_SHA_PASS and its lower candidate/precommit/exact-review-next routes become
  CONSUMED/HISTORICAL. Root may then dispatch only the frozen ordered exact3 test,
  event_repository.go, new occurrence service; source remains UNIMPLEMENTED /
  UNCOMMITTED / UNACCEPTED pending its own lifecycle. 04F-H, Ent, adapter, handler,
  wiring, generation, migration, API/UI, raw iTIP/transport, AQ, launcher NOT_READY
  and executables/runtime NOT_RUN remain blocked. Acceptance-sync author Terra/xhigh;
  dispatcher-enforced/no fallback, telemetry UNKNOWN.
calendar_143_04h_authority_candidate_historical: >-
  CONSUMED/HISTORICAL only for accepted 143-04H source receipt
  73940810046bb3186f2746705a83565ab29368c5. The former authority candidate below is
  preserved as immutable provenance; it creates no future exact15 requirement, exact3
  dispatch, or source-unimplemented lifecycle. Status: ROOT_SOURCE_AUTHORITY_CANDIDATE /
  PRECOMMIT_REVIEW_REQUIRED /
  ATOMIC_EXACT15_COMMIT_REQUIRED / IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN.
  This dirty exact15 documentation and schema-contract candidate grants no source,
  adapter, handler, API implementation, migration, generation or runtime authority.
  The candidate freezes only Calendar 143-04H occurrence-cancel ABI, deterministic
  state rules, safe calendar.invite.cancelled.v1 occurrence payload and the ordered
  future exact3: backend/tests/calendar/event_service_occurrence_cancel_test.go,
  backend/internal/calendar/event_repository.go, then new
  backend/internal/calendar/event_service_occurrence_cancel.go. Source remains
  UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED. Root may dispatch that exact3 only after
  Contract/Spec/Oracle, Security/Adversarial and Lifecycle/Dependency/Boundary
  precommit PASS; one atomic exact15 candidate commit; three immutable exact-SHA PASS
  reviews; and a separate atomic exact15 acceptance sync with its own immutable review.
  Legacy 143-04H METHOD:CANCEL, EventSerializer, 143-02A and source-first clauses are
  CONSUMED/HISTORICAL for this candidate. 04A/04B/04C/04G/04L are static dependencies;
  04J/04K remain untouched and are not invoked. 04F-H, Ent, adapter, handler/wiring,
  generation, migration, AQ blockers, raw iTIP/transport, launcher NOT_READY and
  executables/runtime NOT_RUN remain blocked. Author Terra/xhigh,
  dispatcher-enforced/no fallback, telemetry UNKNOWN.
mail_144_04i_a_source_receipt_historical: "Status: SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT10_COMMIT_REQUIRED / RUNTIME_NOT_RUN. This dirty exact10 documentation candidate grants no acceptance. Only after one atomic commit equal to IMPLEMENTATION-144-04I-A.md's manifest and immutable receipt review is 091f2ef286a71f506e1b048f5386fd37cefa06bc (parent 6b27643fcde3b702fa13c5fcc9b4bfefcb9fc87a, tree b01be92acf01033935b85446454eaee16b84e01b, subject feat(mail): add bounded sync worker) STATIC_ACCEPTED / EXACT_SHA_PASS for authored-first test blob 79c2691971287e3a62c830a5d12c78699414bb86 then source blob 6c20fadf563d4d146b2fc1cbacdddea485266eb7. Contract/Security/Lifecycle precommit and immutable exact-SHA reviews PASS C/I/M 0/0/0; author Terra/xhigh; Contract Terra/high; Security/Lifecycle Terra/xhigh; dispatcher-enforced/no fallback; telemetry UNKNOWN. Then lower 04I-A candidate/source-unaccepted/current-next clauses are CONSUMED/HISTORICAL. No next source is selected. 04H/02E, Ent/concrete adapter, generation/migration, SQL/two-replica proof, API/OpenAPI/UI, launcher NOT_READY and runtime NOT_RUN remain blocked."
calendar_143_04j_source_acceptance_sync: "Status: SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT14_COMMIT_REQUIRED / RUNTIME_NOT_RUN. These proposed receipt facts grant no authority before one atomic commit of the exact14 documentation manifest. After that commit, immutable source e74efa827b52f35e3d3fbe8346adf0c0121ba48d (parent 06a0d6e506a31fb5546ef168a5dba72d8956a984, tree 073e887a957bb65c6c06a9759e239323ab882aa7, subject feat(calendar): add call projection policy) is STATIC_ACCEPTED / EXACT_SHA_PASS for test blob 28273191ddd1babec2cbe61c217cd561322a7ac0 then source blob cc0218bddce25936cbfb8e987a0bb6681f979d5c. Tests were authored first; Contract/Spec/Oracle, Security/Adversarial and Lifecycle/Dependency/Boundary precommit plus immutable exact-SHA reviews PASS C/I/M 0/0/0; author Terra/xhigh, dispatcher-enforced/no fallback, telemetry UNKNOWN. Lower 04J authority/source-unaccepted/exact2-next clauses then become CONSUMED/HISTORICAL. No next Calendar packet is selected; 04K requires separate docs authority. AQ-02 remains migration/backfill-only; 04F-J/K, Ent/generation/migration, adapters/API/UI/raw iTIP, launcher NOT_READY and executables/runtime NOT_RUN remain unchanged."
calendar_143_04k_source_receipt_sync: "Status: SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT15_COMMIT_REQUIRED / RUNTIME_NOT_RUN. This dirty exact15 documentation candidate grants no authority. Only after one atomic exact15 commit and the recorded immutable acceptance reviews does source 572f911227b9d50e21aa73b87c530ddcaf0f8934 (parent 26b377eb288ac8e363cf6d53c744b60f097e9804, tree ac96e2f5c03256b1e00a391600325a524a840673, subject feat(calendar): add call due policy), with authored-first test blob 437db3af0263cedb10dcf217cd53fbbcac03cf23 and source blob 376f46c1cc79956fab3479efc38750fd88161b16, become STATIC_ACCEPTED / EXACT_SHA_PASS. Contract/Spec/Oracle, Security/Adversarial and Lifecycle/Dependency/Boundary precommit plus immutable reviews are PASS C/I/M 0/0/0; author Terra/xhigh, dispatcher-enforced/no fallback, telemetry UNKNOWN. Only then are lower 04K authority-candidate/precommit/current-next clauses CONSUMED/HISTORICAL. No next Calendar packet is selected. 04F-K/AQ-02, Ent/generation/migration, adapters/API/UI/raw iTIP/legacy cutover, launcher NOT_READY and executables/runtime NOT_RUN remain unchanged."
calendar_143_04j_authority_acceptance_sync: "Status: ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT13_COMMIT_REQUIRED / SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN. These proposed receipt facts have no authority before one atomic commit of the same exact13 documentation manifest. After that commit, immutable candidate 5841621720cc03682b3d1f1e2af9cba2953797cf (parent 2a2975056c3fb8001d88863a041310271282989d, tree c4cad5a344ae4972b99410c851f20ab8c73104e7, subject docs(calendar): propose call projection authority) is SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS after Contract/Spec/Oracle, Security and Lifecycle immutable reviews PASS C/I/M 0/0/0. Lower 04J candidate/precommit/exact-review-next clauses then become CONSUMED/HISTORICAL. Root may dispatch only ordered exact2: backend/tests/calendar/event_call_policy_test.go then backend/internal/calendar/event_call_policy.go; source remains UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED. AQ-02 stays migration/backfill-only; 04K/04F, Ent/generation/migration, adapters/API/UI/raw iTIP, launcher NOT_READY and executables/runtime NOT_RUN remain unchanged."
calendar_143_04c_acceptance: "Status: ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_EXACT12_COMMIT_REQUIRED / RUNTIME_NOT_RUN. This dirty exact12 working tree grants no authority. The 7fd061b5099b4608e0c33a6d99edbe1a66587d34 receipt facts and route transitions activate only after one atomic commit of exactly the 12 Calendar 143-04C acceptance-sync documents. Until that commit, the prior 04C candidate/current-next lifecycle remains authoritative. Only after activation is 143-04C STATIC_ACCEPTED / EXACT_SHA_PASS, its scoped predecessor routes become CONSUMED/HISTORICAL, and no next Calendar code packet is selected. Ent generation/migration, adapters, API/UI, raw iTIP, launcher and runtime remain blocked/NOT_READY/NOT_RUN."
mail_144_04g_receipt_coherence: "Status: COHERENCE_SYNC_CANDIDATE / ATOMIC_EXACT12_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED / RUNTIME_NOT_RUN. This dirty exact12 docs-only coherence candidate is non-self-activating and records immutable source 3031c53d399accc84fe10cd0c5421a85c089ac7c (parent 085a81aed317be364811af9507912002a7dae457, tree b146acb45bf60f2f468a926caa3546c0b8f79893, subject feat(mail): add mailbox sync coordinator) and prior acceptance receipt 2a2975056c3fb8001d88863a041310271282989d (parent 6c92046b374a7be0b6964fea335016ed6021e91a, tree 3445df1500adda42d43aa557c370c24308857512, subject docs(mail): record sync coordinator acceptance); full immutable manifests/blobs are in IMPLEMENTATION-144-04G.md. Source Contract/Security/Lifecycle precommit and immutable EXACT_SHA_PASS are C/I/M 0/0/0; Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high, dispatcher-enforced/no fallback/telemetry UNKNOWN. Only after an atomic manifest-equal exact12 coherence commit and immutable coherence C/S/L triple PASS is 144-04G STATIC_ACCEPTED / EXACT_SHA_PASS; former 04G authority, acceptance, source-unimplemented and current-next routes then become CONSUMED/HISTORICAL. Historical coherence selects no next Mail packet and grants no runtime/downstream authority. 04F remains accepted; 04H/02E, Ent/SQL/concrete adapter, generation/migration, API/UI, launcher/testframework and runtime remain BLOCKED/NOT_READY/NOT_RUN."
mail_144_05d_content_source_receipt_sync: >-
  SOURCE_RECEIPT_SYNC_CANDIDATE / ATOMIC_EXACT9_COMMIT_REQUIRED / IMMUTABLE_REVIEW_REQUIRED /
  RUNTIME_NOT_RUN. This dirty exact9 plain-only source-receipt documentation packet is
  non-self-activating: before one atomic manifest-equal exact9 receipt commit and its immutable
  Contract/Security/Lifecycle triple PASS, source acc593fe4f20669a21a6f5ba70a8a20293111017
  (sole parent 98ebedc9dd0b6ad90dbe06e1faf5810376dce392, tree
  7de89d1237bc0c8f640729f607a2dbfd9189f3c4, subject `feat(mail): add plain content normalization`)
  is not STATIC_ACCEPTED and former 05D authority/acceptance/source-unimplemented/current-next routes
  remain current. Its immutable exact2 is authored-first test
  backend/tests/mail/mailbox_content_test.go blob 37c29c421d21e7c173142647d894de0d890894de,
  then source backend/internal/mail/mailbox_content.go blob
  b064a2f6f944d5ffb10651ccb068a77634a9267b. The receipt commit diff MUST set-equal ordered exact9:
  STATE, MAIL-SPEC, Phase-144 PLAN, DISPATCH-INDEX, DISPATCH-REMAINING,
  IMPLEMENTATION-144-05D.md, REVIEW-144-05D-CONTENT-READINESS, CONTRACT-CHANGELOG and QA-HANDOFF.
  Source Contract/Security/Lifecycle precommit and immutable EXACT_SHA_PASS reviews PASS C/I/M
  0/0/0; Contract Terra/high, Security Terra/xhigh, Lifecycle Terra/high,
  dispatcher-enforced/no fallback/telemetry UNKNOWN. Only after activation is acc593fe
  STATIC_ACCEPTED / EXACT_SHA_PASS and former 05D authority, acceptance, source-unimplemented and
  current-next routes are CONSUMED/HISTORICAL. No next packet, schema, OpenAPI, persistence,
  go.mod, Ent, API/UI, testframework, launcher or runtime authority follows.
  The frozen pure ABI is NormalizeMailboxContent over plain/rich Input{SourceFormat,Body} and
  plain Result{Format,PlainBody,RemoteImagesBlocked}; caps are 8_388_608 input/plain bytes,
  2_000_000 runes, 100_000 tokens and depth 64. Literal error ABI is MailboxContentErrorKind
  string with Invalid="invalid" and TooLarge="too_large", exact ErrMailboxContentInvalid and
  ErrMailboxContentTooLarge sentinels, and MailboxContentError{Kind} pointer Error/Unwrap: known
  kinds return/unwrap their matching sentinel, nil/unknown returns `mailbox content error`/nil;
  every failure returns zero Result. Both formats byte-preflight before UTF-8/rune scan/allocation,
  then raw-rune/control/bidi checks; only HT/LF/CR C0 are allowed. NFC canonicalization then maps
  CRLF, CR/LF/U+2028/U+2029, exact horizontal whitespace, line-edge ASCII trim, >=3 LF collapse
  and outer-LF trim before canonical byte/rune caps. Rich uses only existing x/net/html: every
  non-Error token counts, 100000/100001 and depth 64/65 are exact, strict lowercase tag/void/
  self-close/match/EOF stack semantics apply, and entity-decoded controls rerun rejection. The
  exported `MaxMailboxContentHTMLTokenBytes=65_536` is the maximum accepted raw tokenizer-token
  bytes, including delimiters/attributes/text. Because pinned `x/net` fails at `rawLen >= maxBuf`,
  `tokenizer.SetMaxBuf(MaxMailboxContentHTMLTokenBytes + 1)` precedes the first `Next`: raw 65_536
  passes and 65_537 reaches `html.ErrBufferExceeded`, typed `too_large`/zero Result; static source
  review owns other non-EOF tokenizer-error-to-`invalid` mapping. The exact discarded-comment
  boundaries are `"<!--" + strings.Repeat("x", MaxMailboxContentHTMLTokenBytes-7) + "-->"` ->
  accepted/empty plain and replacing payload with `MaxMailboxContentHTMLTokenBytes-6` ->
  typed `too_large`/zero Result. Suppression precedes every block break:
  no root, descendant, self-closing, or void suppression token emits text/LF, while only
  non-suppressed block tokens may do so; `<span>a</span><form><p>x</p></form><span>b</span>` is
  `ab`. The external oracle includes `"<a data-x=\"" +
  strings.Repeat("x", MaxMailboxContentHTMLTokenBytes)` below body cap -> `too_large` without
  panic/leak and a small quoted-attribute success. Exact suppression/block sets,
  attrs/comments/doctype/URLs/alt discard and pre-append buffer cap are in REVIEW; no image/proxy
  token exists. 03C-Attachment and 04A are the only
  accepted dependencies; 04L is not. Rich sanitizer/token/proxy, 05D-Remote and 05E remain
  BLOCKED/NOT_AUTHORIZED. x/net v0.51 direct and x/text v0.35 are already present; exact2 must
  not edit go.mod. Launcher NOT_READY; executable/runtime evidence NOT_RUN; no next packet.
mail_144_02c_source_receipt: "144-02C is STATIC_ACCEPTED / EXACT_SHA_PASS only for immutable source e565523f589fbd7536f6f8e4df7effd39469c1af (parent df2dcbc2151c4b2df5c57774c3114b066b183d24, tree 9fabe45429dfdda49625fe53cd6645e08ed63d52, subject feat(mail): add delivery schemas) and the exact 13 blobs in IMPLEMENTATION-144-02C.md. Tests/oracles were authored first. Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and ABI/Lifecycle/Dependency/Boundary Terra/high precommit plus immutable exact-SHA reviews PASS C/I/M 0/0/0. Former 02C authority-candidate, source-unaccepted, source-lifecycle-next, acceptance-sync-next and implied-next-packet clauses are CONSUMED/HISTORICAL. No next Mail code packet is selected or authorized; root must select it separately. Ent generation, migration, direct SQL, services, API/UI, product, phase, release, engine deletion port, physical User/GDPR intent, 02F-D and 02E remain outside this receipt. Launcher NOT_READY; executable/test/build/gofmt/lint/validator/Docker/runtime evidence NOT_RUN. Docs provenance Terra/medium, dispatcher-enforced/no fallback; telemetry UNKNOWN unless exposed."
mail_144_02b_source_authority: "144-02B source is STATIC_ACCEPTED / EXACT_SHA_PASS at 947e2a27650ff1c6db60af9ce13bb33b6bf19cb0 (parent 856edb25cf40ad436b114cc928923203c195d15c, tree 10c0dcac9953ec53fd1631a095a44052e8cec9c9, subject feat(mail): add projection schemas), exact13 blobs recorded in IMPLEMENTATION-144-02B.md. Tests were authored first; Contract/high, Security/xhigh and Lifecycle/high precommit and immutable exact-SHA reviews PASS C/I/M 0/0/0. Author Terra/xhigh; dispatcher-enforced/no fallback, telemetry UNKNOWN unless exposed. All lower SOURCE_AUTHORITY_ACCEPTED/source-unaccepted/02C-waits-02B-source routing clauses are CONSUMED/HISTORICAL. Root may separately author and adjudicate a narrow 02C docs/source-authority bridge; 02C source is NOT automatically authorized. 02F-D/02E/Ent generation/migrations/services/API/UI/product/phase/release remain BLOCKED; launcher NOT_READY and executable/runtime evidence NOT_RUN. [02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same scalar is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.]"
mail_144_02f_r_authority: "CONSUMED/HISTORICAL authority precursor a877848ef4d1ce80a30a4389f97ef0f863615159. 02F-R remains STATIC_ACCEPTED / EXACT_SHA_PASS at ce542cb5de839a1fb5200295f710b0ae45bcdbe6. 02B is now separately STATIC_ACCEPTED / EXACT_SHA_PASS at 947e2a27650ff1c6db60af9ce13bb33b6bf19cb0; all former 02B source-authority/unaccepted and 02C-predecessor wording is CONSUMED/HISTORICAL. A narrow 02C docs/source-authority bridge requires separate root authoring and adjudication; 02C source is not authorized by 02B acceptance. 02F-D remains BLOCKED. [02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same scalar is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.]"
mail_144_02d_user_configuration_acceptance: "144-02D remains STATIC_ACCEPTED / EXACT_SHA_PASS at 2242a4e7846944498eb5ed701ab8f47c986bd804. 02B is separately STATIC_ACCEPTED / EXACT_SHA_PASS at 947e2a27650ff1c6db60af9ce13bb33b6bf19cb0; the earlier authority/unaccepted/02C-predecessor wording is CONSUMED/HISTORICAL. A narrow 02C docs/source-authority bridge needs separate root authoring/adjudication; 02C source is not automatically authorized. 02F-D remains BLOCKED; launcher NOT_READY; executables/generation/migration/runtime NOT_RUN. [02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same scalar is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.]"
calendar_143_04g_acceptance: "143-04G is STATIC_ACCEPTED / EXACT_SHA_PASS at 48a7273d100c07788def88e2f1b8d69c7173580f; its former next route is CONSUMED/HISTORICAL. 143-04B is STATIC_ACCEPTED / EXACT_SHA_PASS at 75dafea889055126f768227a4ae62b28c91b7393 with exact receipt in IMPLEMENTATION-143-04B.md. No next Calendar code packet is selected; root must separately dispatch any later packet. Launcher NOT_READY; executables NOT_RUN; 03E remains BLOCKED only for Ent generation/migration."
mail_144_02a1_dns_acceptance: "144-02A1 is STATIC_ACCEPTED / EXACT_SHA_PASS at 5892a90b7582ff4ac35db218428400ac9608c4e3. 02B is separately STATIC_ACCEPTED / EXACT_SHA_PASS at 947e2a27650ff1c6db60af9ce13bb33b6bf19cb0; its earlier authority/unaccepted/02C-predecessor wording is CONSUMED/HISTORICAL. Separate root authoring/adjudication is required for a narrow 02C docs/source-authority bridge; 02C source is not automatically authorized. 02F-D/02E remain blocked; launcher NOT_READY; executables/generation/migration/runtime NOT_RUN. [02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same scalar is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.]"
mail_144_03d_acceptance: "144-03D deterministic event mapping is STATIC_ACCEPTED / EXACT_SHA_PASS at 42c6b9f2bd4383040fc116630727564fcf17dd78 (parent d60dab6b39a4642f8c7ca29e377cce9a6ba6d704, tree d711e0c0c9d56d082a3352c4f4bedafc8ce2c18d): backend/internal/mail/mailbox_events.go blob 792a4f460f756c4541003bf1ddc0391be637052c and backend/tests/mail/mailbox_events_test.go blob 8947577115f60f4c3b026dcb686014444660f815. Authoring Terra/xhigh; Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and ABI/Compile-shape/Lifecycle Terra/high exact reviews PASS C/I/M 0/0/0; dispatcher-enforced/no fallback, telemetry UNKNOWN unless exposed. Static only: runtime/tests/build/gofmt/lint/validators/Docker/schema runtime NOT_RUN; launcher NOT_READY. No schema/migration/service/UI/product/phase/release authority. Existing 02A-F/Rule/Sieve/Phase145 blockers unchanged. Historical predecessor receipt only; its former next-packet routing clause is CONSUMED/HISTORICAL. Current source sequencing is founder-directed root dispatch of remaining Mail+Calendar atomic source packets before final integrated Docker; every packet still requires its own test-first/static review/commit. Calendar migration/runtime remains blocked on 143-03E authority."
mail_144_03c_send_acceptance: "144-03C-Send deterministic send policy is STATIC_ACCEPTED / EXACT_SHA_PASS at f8f87b09c4fe65be612fcf6d15ba1f4bd5f8aa19 (parent de6253963c3f787405b3aa63245c0acb0a8cd159, tree 2a4282ef0cf9d24cc722238cc5c28e792a4a8f2a): backend/internal/mail/send_policy.go blob 07f2641f6efa846110e82240a72bc67b4a4f4509 and backend/tests/mail/send_policy_test.go blob ab47f81812f0aecb05ed0c1d17e7b0f7ee340e78. Root retained the PERSISTENCE-backed 1048576 policy floor; token/HTAB handling is fail-closed. Tests authored first, but RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime NOT_RUN because launcher NOT_READY. Authoring Terra/xhigh; Contract Terra/high, Security Terra/xhigh and ABI/Lifecycle Terra/high precommit+exact-SHA reviews PASS C/I/M 0/0/0; dispatcher-enforced/no fallback, telemetry UNKNOWN unless exposed. No schema/migration/service/API/UI/product/phase/release authority and no tenant-auth claim. Existing 144-01B/02/Rule/Sieve/Phase145 blockers unchanged. The former Next-mainline-to-Attachment and Attachment-next routing clauses are CONSUMED/HISTORICAL. Current source sequencing is founder-directed root dispatch of remaining Mail+Calendar atomic source packets before final integrated Docker; every packet still requires its own test-first/static review/commit. Calendar 143-03E migration/runtime remains blocked."
mail_144_03c_attachment_acceptance: "144-03C-Attachment deterministic attachment policy is STATIC_ACCEPTED / EXACT_SHA_PASS at 7b2ea5721684e43559d6f6b4b5fc462d8daafb8b (parent 4e7c609713983a8f6db6fbfeb7474abd73451e8c, tree 4c3fe4d27e170c5d3f42b116c39887036ce936c6): backend/internal/mail/attachment_policy.go blob 3f67ce1b9f2f056b70bfd3070f1c34f98b5c71c1 and backend/tests/mail/attachment_policy_test.go blob bf7eda125d5aba1d0cc2594746d14e7bcfd06aa7. Tests authored first. Pure deterministic per-attachment policy has count<=50, NFC output with NFKC structural/extension defense, closed MIME/type deny tables, inline CID, SHA/scan gates, validate-versus-send/read clean-only behavior, typed content-free errors and deep-copy results; it performs no I/O, authorization, clock or global-state work. Authoring Terra/xhigh; Contract Terra/high, Security Terra/xhigh and ABI/Lifecycle Terra/high precommit+exact-SHA reviews PASS C/I/M 0/0/0; dispatcher-enforced/no fallback, telemetry UNKNOWN unless exposed. Static only: RED/GREEN/tests/build/gofmt/lint/validators/Docker/runtime NOT_RUN because launcher NOT_READY. No scanner/storage/network/project/tenant auth/schema/migration/service/API/UI/product/phase/release authority. Existing 144-01B/02/Rule/Sieve/Phase145 and Calendar 143-03E blockers unchanged. The former no-further-source/root-only-selection clause is CONSUMED/HISTORICAL. Current source sequencing is founder-directed root dispatch of remaining Mail+Calendar atomic source packets before final integrated Docker; every exact packet still requires its own test-first/static review/commit. No Ent generation/migration/runtime authority follows."
carrier_plan_acceptance: "PLAN-145-07B-SIEVE-CARRIER.md is docs/implementation-authority-only STATIC_ACCEPTED / EXACT_SHA_PASS at 949b876d1c5ea07338926aa5fe19e1864a256e93 (parent 375891bb85c0e89f2b6597e0416b82219fffaa74, tree f943fb60e8dbda3ae612871d189a57d751dc44cc, sole plan blob 0bbbc91cd3c58fc8daeaab1a434d942f3ef26821); Contract Terra/high, Security/Purity Terra/xhigh and Lifecycle Terra/high exact-SHA reviews PASS C/I/M 0/0/0; dispatcher-enforced, fallback none, telemetry UNKNOWN unless exposed. Carrier contract 1daef787 and acceptance sync 375891bb remain accepted. Only obsolete carrier-plan author/adjudicate/accept/pending clauses are CONSUMED/HISTORICAL; unrelated gates stay current. No source/test/runtime/adapter/schema/service/worker/product/phase/release authority. Rule Task 1 remains NOT_AUTHORIZED/BLOCKED on G145-5B, G145-5C and launcher readiness. Carrier six-path RED/source remains NOT_AUTHORIZED/BLOCKED pending accepted Rule compiler source/ABI, G145-5B, G145-5C and launcher readiness; plan gate is satisfied. 145-07B source/runtime remains blocked; 144-01B, Phase142/148, G145 gates, launcher NOT_READY and executables NOT_RUN unchanged."
carrier_contract_acceptance: "Phase-145 non-marker typed-IR/Sieve carrier is docs/design/readiness-only STATIC_ACCEPTED / EXACT_SHA_PASS at 1daef78768048d7672c38a65b78ef0f119f3c207 (parent c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6, tree 1086a93f2f52423cfdf9dcfc01eadd2e0d434947, sole carrier blob 218a38117ae51d014f40218e0f5c4606ec8d4dfa); Contract/Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high exact-SHA PASS C/I/M 0/0/0; dispatcher-enforced, fallback none, telemetry UNKNOWN unless exposed. Carrier contract 1daef787 and carrier-plan acceptance 949b876d remain accepted. Base 692b99d8, correction c14c2088, and former carrier-plan author/adjudicate/accept/pending clauses are CONSUMED/HISTORICAL. No source/test/runtime/adapter/schema/service/worker/product/phase/release authority. Rule Task 1 remains NOT_AUTHORIZED/BLOCKED pending accepted Rule compiler source/ABI/source availability, G145-5B, G145-5C and launcher readiness; Carrier six-path RED/source and 145-07B source/runtime remain NOT_AUTHORIZED/BLOCKED on the same blockers. 144-03C-Send and Attachment are static accepted; their former no-further-source/root-only-selection clause is CONSUMED/HISTORICAL. Current source sequencing is founder-directed root dispatch of remaining Mail+Calendar atomic source packets before final integrated Docker, each with its own test-first/static review/commit; no Ent generation/migration/runtime authority follows. Engine deletion-port adjudication remains separately BLOCKED / NOT_DISPATCHABLE. Launcher NOT_READY and executables/runtime NOT_RUN."
rule_plan_acceptance: "Standalone PLAN-144-03B-RULE.md is docs/plan-only STATIC_ACCEPTED / EXACT_SHA_PASS at 669ef761715ec35a638bc87d8396e6f562b162c4 (parent 77eb4356de06a5304521b60c57af4e73c5740ce1, tree 1b554d1c5763818ffef4f3c06c2a5bde16046d37, sole path, blob b86c0423a307fd2673a3d322435922aa2a398e8f); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M 0/0/0; dispatcher-enforced, fallback none, telemetry UNKNOWN unless exposed. Former Rule-plan author/adjudicate/accept instructions CONSUMED/HISTORICAL. Plan only: no Rule oracle/source/TDD/schema/migration/service/Phase145/runtime/product/phase/release authority. Task 1 RED NOT_AUTHORIZED/BLOCKED until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, no order implied; post-task runtime NOT_RUN. Task4 docs sync separately discovered after future source Task3 acceptance."
last_activity: 2026-08-23 · origin/master f21e3f77 testframework cutover integrated into codex/mail-module-20260813; Mail/Calendar static-source history preserved; no tests or runtime executed during merge.
mail_144_02a2_domain_policy_acceptance: "144-02A2 is STATIC_ACCEPTED / EXACT_SHA_PASS at 65c91b4824e8a1948fc97c149305aa7f009de061. 02B is separately STATIC_ACCEPTED / EXACT_SHA_PASS at 947e2a27650ff1c6db60af9ce13bb33b6bf19cb0; its earlier authority/unaccepted/02C-predecessor wording is CONSUMED/HISTORICAL. Separate root authoring/adjudication is required for a narrow 02C docs/source-authority bridge; 02C source is not automatically authorized. 02F-D/02E remain blocked; launcher NOT_READY; executables/generation/migration/runtime NOT_RUN. [02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same scalar is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.]"
mail_144_02a3_account_identity_acceptance: "144-02A3 is STATIC_ACCEPTED / EXACT_SHA_PASS at e9a1764b001e28074e4074a07ca8ac982e5345ba. 02B is separately STATIC_ACCEPTED / EXACT_SHA_PASS at 947e2a27650ff1c6db60af9ce13bb33b6bf19cb0; its earlier authority/unaccepted/02C-predecessor wording is CONSUMED/HISTORICAL. Separate root authoring/adjudication is required for a narrow 02C docs/source-authority bridge; 02C source is not automatically authorized. 02F-D/02E remain blocked; launcher NOT_READY; executables/generation/migration/runtime NOT_RUN. [02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same scalar is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.]"
progress:
    total_phases: 125
    tasks_completed: 125
    goals_verified: 63
    phases_percent: 100
    phases_percent_note: >-
      100% относится только к историческим 125 фазам milestone v4 и не означает
      production-ready. Активная программа test-rebase — фазы 128-141; Phase
      128 IN_PROGRESS, остальные PLANNED. Эти 14 фаз не входят в счётчик выше.
    notes: >-
      tasks_completed = checkbox/legacy completed_phases claim.
      goals_verified = count of classification==VERIFIED_PASS in
      PHASE-STATUS-MATRIX.md per-phase table (column classification only;
      method documented in matrix header «How to count goals_verified»).
      Recount 2026-07-28 (TASK-004): 63 VERIFIED_PASS / 5 VERIFIED_PARTIAL /
      2 CODE_GATE_DEPLOY_BLOCKED / 54 NO_VERIFICATION = 124 dirs.
      Prior STATE value 55 was wrong (naive full-file / notes drift).
      phases_percent remains tasks_completed/total_phases — NOT goals_verified.
      Matrix scan date 2026-07-27; VERIFICATION for 126/127 added 2026-07-28
      not yet folded into matrix (no silent re-write of phase verdicts).
deploy_blockers:
  - id: PG-03
    aka: FA-06
    title: Secret rotation (Resend/Telegram P1 + inventory)
    status: open
    blocks: production_cutover
    does_not_block: feature_dev_july_rebuild_planning
    phase_refs:
      - "63-production-gate (VERIFICATION: code PASS, deploy blocked)"
      - "67-production-gate (VERIFICATION: code PASS, deploy blocked)"
    docs:
      - .planning/ops/production/PG-03-SECRET-ROTATION.md
      - .planning/ops/production/FA-06-EXECUTION-LOG.md
      - docs/runbooks/fa06-secret-rotation.md
    progress:
      phase_0_git_scan: done
      phase_1_rotate: pending
      phase_2_smoke: pending
      phase_3_cto_signoff: pending
coverage:
    overall_internal_statements: 82.0373
    measured: "2026-08-07"
    status: historical_non_blocking_baseline
    canonical_baseline: TASK-203_discovery_only_founder_cutover_supersedes_gate
    baseline_2026_06_22: 83.0
    delta_pp: -0.9627
    gate_floor_82: historical_pass_not_migration_gate
inventory_measured_2026_08_07:
    source_revision: 5708671f00411c7c9e9dbee88b725de2092cbacd
    ledger_rows: 1049
    ledger_declared_tests: 4246
    ledger_sha256: 72de6a81b2bd5da8873a98e784a5069a661e9caea083f10722aa85e9f91b4511
    backend_testframework_test_go: 831
    backend_testframework_declared_tests: 3672
    backend_other_test_go: 1
    backend_other_declared_tests: 2
    backend_total_test_go: 832
    backend_total_declared_tests: 3674
    backend_genuine_unit_files: 595
    backend_hook_free_unit_files: 350
    backend_hook_dependent_protected_unit_files: 245
    backend_component_files: 237
    backend_component_api_files: 31
    backend_component_integration_files: 206
    vitest_frontend_test_files: 180
    vitest_literal_cases: 519
    vitest_frontend_unit_files: 156
    vitest_static_contract_files: 24
    frontend_support_files: 3
    playwright_specs: 20
    playwright_declared_cases: 53
    playwright_support_files: 14
    hook_source_files: 43
    hook_symbols: 582
    hook_links: 892
    protected_unit_hook_files: 42
    protected_unit_hook_symbols: 434
    protected_unit_hook_links: 653
    target_python_tests: 0
    ent_entity_schemas: 87
    zustand_stores: 12
    go_mod: "1.25.0"
    branch: "cursor-grok-4.5"
active_program:
    name: test-architecture-rebase
    phases: "128-141"
    current_phase: 128
    completed_tasks:
      - TASK-202
      - TASK-221
      - TASK-222
      - TASK-223
      - TASK-226
      - TASK-225
      - TASK-203
      - TASK-224
    archived_tasks:
      - TASK-202
    in_progress:
      - TASK-204
    in_review: []
    blocked: []
    target_layout:
      - backend/tests/<domain>/**
      - frontend/tests/<domain>/**
      - frontend/tests/support/**
      - testframework/** (single root Python project; runtime test code is Python-only)
mail_feature_workstream:
    name: hub-mail-calendar-sync
    branch: codex/mail-module-20260813
    phases: "142-148"
    current_phase: 144
    status: mail_144_02a1_dns_and_02a2_domain_policy_static_accepted_exact_sha_pass_5892a90b_65c91b48_founder_sequenced_next_a3_account_identity_source_only_02e_ent_generation_migration_runtime_unauthorized_144_01b_rule_sieve_phase145_calendar_143_03e_blockers_unchanged_launcher_not_ready_executables_runtime_not_run
    product_truth: unfinished_mvp_feature_not_accepted
    contract_version: "0.1.8"
    tasks:
      - TASK-235
      - TASK-236
      - TASK-237
      - TASK-238
      - TASK-239
      - TASK-240
      - TASK-241
    protected_parallel_program: "Claude test rebase phases 128-141; untouched"
    runtime:
      mail_engine: NOT_RUN
      docker_profile: NOT_IMPLEMENTED
      docker_daemon: AVAILABLE_29.7.2_access_verified_2026-08-13
      launcher: NOT_READY
      resource_budget: NOT_MEASURED
      external_delivery: NOT_CONFIGURED
      atomic_code_model: gpt-5.6-terra_xhigh_dispatcher_enforced_144_01A_telemetry_UNKNOWN
      atomic_code_model_evidence: .planning/phases/144-mailbox-domain/IMPLEMENTATION-144-01A.md
      atomic_code_provenance_rule: dispatcher_enforced_for_exact_native_acceptance_telemetry_must_match_if_available
      production_migration_authority: UNRESOLVED
      production_migration_evidence: .planning/phases/143-calendar-rsvp-core/MIGRATION-PLAN.md
      frontend_unit_harness: PENDING_PROTECTED_REBASE
      stalwart_gate_plan: .planning/phases/145-stalwart-gateway/GATE-PLAN.md
    decisions:
      engine_candidate: Stalwart_v0.16.17_source_candidate_only
      engine_acceptance: conditional_on_license_provenance_and_runtime_gates
      engine_oci_index: sha256_a8108e19bd927e172d4d8c128907b8dfc93fd180ae8ee07dccdd42cb97eb9dfa_captured_not_approved
      engine_linux_amd64_manifest: sha256_2bcbd3797c8170bd261a533d54d3b25c1a49558aa69f3e42b0d37bdf2b106a28_captured_not_approved
      engine_attestation: slsa_reference_captured_signature_sbom_not_verified
      engine_endpoints: shared_jmap_for_registry_and_mail_plus_fixed_api_helpers
      engine_credentials: separate_management_api_key_and_unresolved_mailbox_auth_same_jmap_no_method_or_batch_crossing
      engine_management_permissions: replace_domain_and_individual_crud_list_only_review_pending
      engine_mailbox_auth_options: per_user_oauth_authorization_code_recommended_or_explicit_broad_impersonator_founder_choice_pending
      engine_mailbox_auth_rejections: no_management_key_no_oauth_client_credentials_no_acl_only_full_engine_no_synthesized_user_passwords
      engine_mailbox_auth_review: route_rereview_no_p0_p1_p2_wording_fixed_dispatcher_enforced_or_matching_telemetry_required
      engine_public_url: single_stalwart_public_url_for_oauth_oidc_jmap_requires_shared_trusted_origin_if_oauth
      engine_network_profile: default_deny_bounded_service_http_no_8080_no_user_mail_protocols_smtp25_optional_oauth_https_gate
      engine_domain_verification: unresolved_stalwart_domain_has_expected_zone_not_verification_fail_closed_dns_required
      engine_identity_mapping: management_account_addresses_then_mail_jmap_identity_get_set_auth_gate
      internal_delivery: through_engine
      message_authority: mail_engine
      mailbox_projection: leased_bounded_worker_atomic_page_event_cursor_command_lease_begin_end_epoch_fence
      mailbox_rebuild: no_clear_base_state_id_anchor_enumeration_then_changes_catchup
      mailbox_provisioning: leased_bounded_domain_then_account_control_stable_keys_generations
      mailbox_rules: hub_desired_generation_fixed_sieve_get_before_retry_worker
      mailbox_mutations: v4_v7_scoped_digest_hmac_receipt_30d_bulk_snapshot_lifetime_tombstone_same_object_presign
      mailbox_versions: if_match_label_settings_signature_rule_policy_domain
      mailbox_content: server_sanitize_or_plain_only_remote_token_byte_proxy_private_nosniff_no_referrer_csp
      mailbox_folder_recovery: exclusive_writer_capability_false_by_default_no_alternate_imap_jmap_admin_writer
      mail_code_routing: founder_task_local_override_gpt-5.6-terra_xhigh_dispatcher_enforced_no_fallback
      mailbox_search: bounded_sanitized_body_attachment_filenames_owner_reauth
      mailbox_spam_feedback: explicit_spam_ham_not_move_alias
      mailbox_engine_bounds: projection_only_pages_1mib_item_8mib_page_budgeted_50mib_message_25mib_blob
      mailbox_projection_version: required_engine_version_same_projection_atomic
      mailbox_submission_identity: split_opaque_engine_ids_and_hub_message_uuid_projection_before_sent
      mailbox_public_bounds: cursor_pages_8mib_legacy_calendar_call_16mib_non_truncating
      mailbox_attachment_staging: immutable_draft_client_id_one_draft_then_eventual_message_max_50_locked
      mailbox_defaults: user_settings_pointers_only
      mailbox_recent_external_contacts: max_1000_lru_non_hub
      calendar_authority: hub_postgresql
      calendar_itip_hash: cve_v1_separate_from_outbound_versioned_all_day_utc_marker
      oauth_l3: out_of_scope
      legacy_admin_broadcast: governance_gap_deprecated_pending_founder_decision
      reminder_delivery: durable_in_app_best_effort_telegram_push_no_resend
      ical_l1: contract_ready_code_pending_founder_question_5
    artifacts:
      adr: .planning/decisions/ADR-MAIL-01-mail-engine-calendar-sync.md
      spec: .planning/decisions/specs/MAIL-SPEC.md
      openapi: docs/mail/openapi.yaml
      engine_contract_0_1_8: docs/mail/ENGINE-CONTRACT-0.1.8.md
      persistence_contract_0_1_8: docs/mail/PERSISTENCE-CONTRACT-0.1.8.md
      control_plane_schema_plan: docs/superpowers/plans/2026-08-13-mail-control-plane-schemas.md
      stalwart_0_16_17_research: .planning/phases/145-stalwart-gateway/RESEARCH-2026-08-13-STALWART-0.16.17.md
      stalwart_0_16_17_route_remediation: .planning/phases/145-stalwart-gateway/REMEDIATION-145-JMAP-ROUTE-P0-2026-08-13.md
      stalwart_0_16_17_route_review: .planning/phases/145-stalwart-gateway/REVIEW-145-0.16.17-ROUTE-SOURCE-ONLY-2026-08-13.md
      events: docs/mail/events-v1.schema.json
      qa_handoff: docs/mail/QA-HANDOFF.md
      founder_questions: docs/mail/FOUNDER-QUESTIONS.md
      phase_147_contract: .planning/phases/147-mail-calendar-reliability/CONTRACT.md
      phase_147_dispatch: .planning/phases/147-mail-calendar-reliability/DISPATCH-INDEX.md
      phase_147_static_review_historical_0_1_6: .planning/phases/147-mail-calendar-reliability/STATIC-REVIEW-0.1.6.md
      phase_147_static_review_historical_0_1_7: .planning/phases/147-mail-calendar-reliability/STATIC-REVIEW-0.1.7.md
      phase_148_gate: .planning/phases/148-mail-release-gate/GATE-MATRIX.md
      phase_148_reviews: .planning/phases/148-mail-release-gate/REVIEW-PACKETS.md
spec_coverage:
    fr_total: 28
    fr_with_phase: 9
    ratio: "9/28"
    decimal: 0.321
    percent: 32.1
    measured: "2026-07-27"
    source: .planning/audits/planning-audit-2026-07/SPEC-COVERAGE.md
---
mail_calendar_http_composition: >-
  STATIC_ACCEPTED / RUNTIME_NOT_RUN. Task 7B accepts the deleting-owner snapshot `f02c3be1`,
  mail-account handler `ebfe750d`, and its safe error-classification correction `af92452d`, with
  C/I/M 0/0/0: it registers authenticated `GET /api/mail/account` only. Task 7C accepts the
  mutation contract `cb1c59e8`, unsupported RSVP-comment removal `6a9276ba`, typed-create-conflict
  correction `fbeb7419`, handler `bc66cdcb`, and exact-key correction `0eacbc2d`, each independently
  statically approved: six canonical mutation routes are additive and do not replace legacy Calendar
  GET. Task 8A route composition is accepted at `1627b46d` with C/I/M 0/0/0: private UTC/UUID/iTIP
  adapters register legacy Calendar GET, canonical mutations, and the Mail account route once in the
  main server. No MailEngine, sync worker, provider or NATS consumer is composed; Task 8 worker
  lifecycle remains pending/blocked. Task 7A remains blocked on Phase 145 founder auth/gateway gates
  and inaccessible private engine interfaces. Launcher/framework cutover is NOT_READY; build, test,
  generation, migration, Docker, runtime, and merge acceptance remain NOT_RUN.
mail_144_02f_r_acceptance: "02F-R STATIC_ACCEPTED / EXACT_SHA_PASS ce542cb5de839a1fb5200295f710b0ae45bcdbe6; parent 1792cdc381eeb4db0ab1ede5b512ba1222eb90b2; tree 2cadf80c9ae0af99b5c8027c1d78c3ec55af7071; exact eight blobs in IMPLEMENTATION-144-02F-R.md. Tests first; Contract/Security/Lifecycle PASS C/I/M 0/0/0; dispatcher-enforced/no fallback/telemetry UNKNOWN. 02B is separately STATIC_ACCEPTED / EXACT_SHA_PASS at 947e2a27650ff1c6db60af9ce13bb33b6bf19cb0; all former authority/unaccepted/02C-predecessor wording is CONSUMED/HISTORICAL. Separate root authoring/adjudication is required for a narrow 02C docs/source-authority bridge; 02C source is not automatically authorized. 02F-D BLOCKED; launcher NOT_READY, executables/generation/migration/runtime NOT_RUN. [02C source acceptance: every preceding 02C bridge/candidate/source-unaccepted/dispatch-next clause in this same scalar is CONSUMED/HISTORICAL; e565523f589fbd7536f6f8e4df7effd39469c1af is STATIC_ACCEPTED / EXACT_SHA_PASS with exact13 blobs in IMPLEMENTATION-144-02C.md; no next Mail code packet is selected or authorized; named blockers/NOT_READY/NOT_RUN remain.]"
calendar_143_04b_acceptance: "143-04B STATIC_ACCEPTED / EXACT_SHA_PASS 75dafea889055126f768227a4ae62b28c91b7393; parent a877848ef4d1ce80a30a4389f97ef0f863615159; tree 6b86570991e41e0c401934867a49a17849847cb3; source db34bbbc4b4c8a3ace8f768998060f82deecedbe; test 79225fb73481ffd34ee1f3b9a72965d255e0b14a. Tests first; Contract/Security/Lifecycle PASS C/I/M 0/0/0; dispatcher-enforced/no fallback/telemetry UNKNOWN. PRECOMMIT_FAIL/NOT_ACCEPTED/current-next CONSUMED/HISTORICAL; no next Calendar packet; launcher NOT_READY, executables/generation/migration/runtime NOT_RUN; 03E Ent generation/migration BLOCKED."
calendar_143_04c_authority: "143-04C is STATIC_ACCEPTED / EXACT_SHA_PASS only at immutable source 7fd061b5099b4608e0c33a6d99edbe1a66587d34 (parent bb84a47e12516c19682ff08783a5844306fc5766, tree 0fb65fc72705bb145a1f2950c6efbe16607e6a04, subject feat(calendar): add update and cancel service) for exact3 test-first blobs event_service_update_test.go 4421204141719fa1af586b3d6272aa130f185af2, event_repository.go 737207280cc23adb873ae8d36952a3e62a0a3a4d, then event_service_update.go 094b68dd8369bdc9345d9348016b3bdc377aa7a9; receipt IMPLEMENTATION-143-04C.md. Contract/Spec/Oracle, Security and Lifecycle precommit plus immutable exact-SHA reviews PASS C/I/M 0/0/0; author Terra/xhigh, dispatcher-enforced/no fallback, telemetry UNKNOWN unless exposed. Every preceding 04C ROOT_SOURCE_AUTHORITY_CANDIDATE, precommit/source-unaccepted/uncommitted/current-next or blocked-pending-source-acceptance clause is CONSUMED/HISTORICAL only for 04C. No next Calendar code packet is selected or authorized; root must separately select any later packet. 04A/B/G receipts, 03E Ent generation/migration block, adapter/API/UI/raw-iTIP limits, launcher NOT_READY and executable/runtime NOT_RUN remain authoritative."

# Состояние проекта Hub

## 143-04J call-projection authority candidate (2026-08-21)

`REVIEW-143-04J-CALL-PROJECTION-AUTHORITY.md` freezes only the exact13 documentation candidate:
pure caller-owned Hub ScheduledCall projection, not source acceptance. It authorizes no dispatch
until one atomic exact13 docs commit, independent Contract/Spec/Oracle, Security and Lifecycle
immutable exact-SHA PASS, and separate docs-only acceptance. Only then may root dispatch test-first
`backend/tests/calendar/event_call_policy_test.go`, then
`backend/internal/calendar/event_call_policy.go`. AQ-02 remains backfill-only; 04K/04F,
Ent/generation/migration, adapter/API/UI, raw iTIP, launcher `NOT_READY` and all executables/runtime
`NOT_RUN` remain unchanged.

## 144-04G sync-service authority candidate (2026-08-20; CONSUMED/HISTORICAL predecessor)

This route is historical receipt evidence only. Before the manifest-equal 144-04G exact12 coherence
commit and immutable C/S/L triple PASS, it confers no current static, source or downstream authority.
Its former `ROOT_SOURCE_AUTHORITY_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED /
IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN` lifecycle proposed a future test-first exact3: NEW
`backend/tests/mail/mailbox_sync_service_test.go`, NEW interface/types-only
`backend/internal/mail/mailbox_repository.go`, then NEW
`backend/internal/mail/mailbox_sync_service.go`. The false lower premise that repository
signatures extend an accepted `mailbox_repository.go`, and the lower claim that the file does not
exist, are `CONSUMED/HISTORICAL`; the immutable source receipt records the source file but grants
no current authority before coherence activation. 04F remains `STATIC_ACCEPTED / EXACT_SHA_PASS`; 04H/02E, Ent/concrete adapter,
generation, migration, wiring and runtime remain blocked or `NOT_RUN`.

The exact ABI/state machine/oracles are frozen in
`REVIEW-144-04G-SYNC-SERVICE-READINESS.md`: four short repository transaction boundaries
(claim/prepare/apply/resolve); complete account+lease+lease-epoch+sync-epoch+mutation-epoch
fencing; injected clock/lease IDs/event IDs/jitter/poll and retry intervals; one claimed Step
does one ListFolders plus exactly one Changes or Enumerate; resumable base enumeration then
catch-up; invalid cursor/restart schedules next-step rebuild without clearing good state;
deterministic candidate IDs feed accepted 04F; no engine/storage/callback/classifier under a DB
lock; bounded deep copies and content-free typed errors. Contract/Security/Lifecycle precommit
PASS -> atomic exact9 candidate commit -> triple immutable exact-SHA PASS -> separate docs-only
acceptance sync is `CONSUMED/HISTORICAL` predecessor evidence. Calendar 143-04C keeps its current
exact3 dispatch state and all existing blockers; this historical route changes no Calendar path.

Fix1 freezes two additional non-negotiable guards: Prepare/Apply/Resolve treat persisted
`sync_lease_until <=` authoritative stable DB transaction time as LeaseLost with zero effects,
even without takeover; request.Now never authorizes mutation and service never queries DB time.
Every Incremental/CatchUp Changes and initial/resumed Enumerate request carries exact
`EnginePageRequest{Limit: config.PageSize}`.

## 143-03B-T immutable source acceptance and former conditional 143-04C gate (CONSUMED/HISTORICAL as to 04C; 2026-08-20)

Calendar 143-03B-T is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `68cbf7cb227799304ce679937b519b66cb6abf9b` (parent `4715f939f24e5b5c2d3f80a38d204babe68bdd34`, tree `1537ba1d2d0fa804545fd8667ffe1dce9e8eaf1e`, subject `feat(calendar): add attendee tombstones`). The exact2 test-first blobs recorded in `IMPLEMENTATION-143-03B-T.md` are schema test `8bc4d44573e5a86e2c78df8c844fe3f1f2fae4cc` then `calendarattendee.go` `3c38799f3e56089e2e264b000ee0951a144b0257`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file saying 143-03B-T is only source-authorized, not accepted, dirty, uncommitted or unaccepted, or that 143-04C is blocked pending 03B-T source acceptance, is `CONSUMED/HISTORICAL` as to that dependency. Status: `COMBINED_EXACT19_SOURCE_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. Only after one atomic exact19 docs commit and independent Contract/Security/Lifecycle immutable exact-SHA PASS for that commit may root dispatch the frozen 143-04C exact3, test first and in order: `backend/tests/calendar/event_service_update_test.go`, `backend/internal/calendar/event_repository.go`, new `backend/internal/calendar/event_service_update.go`. Until then exact3 remains `UNIMPLEMENTED / UNCOMMITTED / UNACCEPTED` and dispatch is not effective.

Mail 144-04F remains `STATIC_ACCEPTED / EXACT_SHA_PASS` at `4715f939f24e5b5c2d3f80a38d204babe68bdd34`; before the manifest-equal 144-04G exact12 coherence commit and immutable C/S/L triple PASS, 144-04G has no current static, source or downstream authority. No adapter, Ent generation, migration, raw iTIP, service/API/UI/product/phase/release or runtime authority follows. Launcher `NOT_READY`; generation/migration/tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`; existing 143-04A/B/G receipts and other blockers remain unchanged.

## 144-04F immutable source acceptance (2026-08-20)

Mail 144-04F is `STATIC_ACCEPTED / EXACT_SHA_PASS` only at immutable commit `4715f939f24e5b5c2d3f80a38d204babe68bdd34` (parent `7ac8714e99db2935ce290ad67836e1309cc05c78`, tree `6461894a90c4340bf96118ed1a242516c02b48c7`, subject `feat(mail): add sync page policy`). Its exact4 ordered test-first blobs are recorded in `IMPLEMENTATION-144-04F.md`: events test `50a41cefb539d9b840a43da9e1ff34cce6c8653f`, sync-policy test `b6410709e4fb8166c28a5c502cfba535eef08d06`, events source `a3017b0a6064d8fc1017e37e2d4f675ee8732945`, sync-policy source `acd642fcdb6290af79b7fa8e5aa36a5036576f47`. Author Terra/xhigh; Contract/high, Security/xhigh and ABI/Lifecycle high precommit plus immutable exact-SHA reviews PASS C/I/M `0/0/0`; dispatcher-enforced, no fallback, telemetry `UNKNOWN` unless exposed.

Every other clause in this file asserting that Mail exact4 is dirty, `UNCOMMITTED`, `UNACCEPTED`, future-only, source-blocked, awaiting source review/commit, or root-dispatch-next is `CONSUMED/HISTORICAL`; the immutable source receipt above is current. No automatic Mail next source follows: before the manifest-equal 144-04G exact12 coherence commit and immutable C/S/L triple PASS, 144-04G has no current static, source or downstream authority.

The Calendar half of the combined exact17 route is unchanged: 143-03B-T exact2 remains source-authorized but is not `STATIC_ACCEPTED`; any dirty candidate remains uncommitted/unaccepted. 143-04C exact3 stays `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; 143-04A/B/G receipts remain accepted. Launcher `NOT_READY`; tests/build/gofmt/lint/validators/Docker/runtime `NOT_RUN`. 143-03E, 144-02E, Ent generation and migration stay blocked. No raw iTIP or service/API/UI/product/phase/release authority follows.

## Combined exact17 conditional acceptance sync (CONSUMED/HISTORICAL; 2026-08-20)

Status: `COMBINED_EXACT17_ACCEPTANCE_SYNC_CANDIDATE / ATOMIC_COMMIT_REQUIRED / IMMUTABLE_EXACT_SHA_REVIEW_REQUIRED`. This working tree is not authority. Calendar exact12 `e60ff29934b8e53c34549cebae2e10657bae1540` (parent `e21a98a7459cd68690912c33ff031e286ba92ef1`, tree `5df5e079bdfe9685957e065e534c352cede25d2a`) and Mail correction exact9 `4f92ee6b9057efccd10c0b7ac7cf7dfd0f220c71` (parent `e60ff29934b8e53c34549cebae2e10657bae1540`, tree `2c65309e71c3501cd55516bc84c88c1c605df5a3`) each have Contract/Security/Lifecycle precommit and immutable exact-SHA PASS C/I/M `0/0/0`. This receipt becomes effective only after root creates one atomic commit changing exactly the 17 named acceptance-sync documents and Contract/Security/Lifecycle independently return immutable exact-SHA PASS for that exact commit. Until then all other Calendar candidate and Mail correction/suspension clauses in this file remain current; only after those conditions they automatically become `CONSUMED/HISTORICAL`, and the accepted routes below activate without another docs sync.

On activation, Calendar `e60ff299` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only 143-03B-T exact2, test first: `backend/tests/calendar/calendar_attendee_schema_test.go`, then `backend/ent/schema/calendarattendee.go`. 143-04C exact3 remains `BLOCKED / NOT_AUTHORIZED` until 143-03B-T source itself is `STATIC_ACCEPTED`; there is no 04C source dispatch or acceptance now. Mail correction `4f92ee6b` becomes `SOURCE_AUTHORITY_ACCEPTED / EXACT_SHA_PASS` and root may dispatch only the frozen exact4, test first and in order: `backend/tests/mail/mailbox_events_test.go`, `backend/tests/mail/mailbox_sync_policy_test.go`, `backend/internal/mail/mailbox_events.go`, `backend/internal/mail/mailbox_sync_policy.go`. The current dirty exact4 candidates remain `UNCOMMITTED / UNACCEPTED` and require their own source static review, atomic commit and immutable exact-SHA acceptance lifecycle; runtime remains `NOT_RUN`.

The 143-04A/B/G receipts remain accepted. 143-03E, 144-02E, Ent generation, migration, launcher readiness and all executable/runtime evidence remain blocked, `NOT_READY` or `NOT_RUN` as already recorded. No source outside the two activated manifests, generation, migration, service expansion, API/UI, product, phase or release authority follows; raw iTIP remains prohibited.

> **Carrier-lifecycle clause precedence (2026-08-15):** Only the lower clause or fragment asserting carrier **contract** acceptance is missing, pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Local 01B clause reconciliation:** every earlier 144-01B Docker-blocked/do-not-start/dependency assertion in this Mail workstream record is `CONSUMED/HISTORICAL` only. The exact static receipt is `0c064363b5ccb657a7c444802e94075c4ca743e5`; it leaves launcher, 02E/Ent/generation/migration/service/API/UI/product/phase/release, Phase 142/148, Rule/Sieve and Phase145 blockers current.

## Product workstream — Hub Mail + Calendar Sync (2026-08-13)

- Изолированная ветка: `codex/mail-module-20260813`; базовый master при создании:
  `d44175f8`. Основная рабочая копия и защищённые зоны test-rebase не меняются.
- Phase 142 создал contracts-first пакет: ADR-MAIL-01, ERD/API/NATS/security,
  Outlook UI design, traceability и QA handoff. Mail 0.1.8 root-freezes the later bounded
  engine/API and exact persistence amendment, but independent review is pending; наличие
  артефактов не означает принятого контракта, реализации или готовности модуля.
- Source-only Terra/high-requested review at `c48046b8` reported actual telemetry `UNKNOWN`,
  no P0, five P1 machine-contract gaps and two P2 bound/example gaps. Root remediation closes
  public DTOs, applies one declared default-deny `x-hub-authz` policy to all 56 non-admin
  operations while retaining RBAC on all 13 admin operations, aligns event attendee cap 200,
  adds the channel subject/payload UUID invariant and recovery/replay examples, bounds the
  still-blocked legacy broadcast, and makes the mailbox recipient policy reachable at 300.
  This is not independent acceptance: post-remediation independent Terra/high rereview with
  accepted per-dispatch provenance and all executable OpenAPI/JSON Schema/runtime validation
  remain `NOT_RUN`.
- Preferred engine: Stalwart за adapter boundary, условно. Official-source refresh on
  2026-08-13 identifies `v0.16.17` / source commit
  `08509ddccbe6f978a1ffebeb250cab8588d765e5` as the current candidate, not an approved
  image. Read-only Docker Hub registry inspection captured OCI index
  `sha256:a8108e19bd927e172d4d8c128907b8dfc93fd180ae8ee07dccdd42cb97eb9dfa`,
  linux/amd64 manifest
  `sha256:2bcbd3797c8170bd261a533d54d3b25c1a49558aa69f3e42b0d37bdf2b106a28`
  and a linked SLSA-provenance attestation manifest. No image was pulled or run; signature,
  SBOM and independent approval remain `NOT_RUN`. License verdict, exact config,
  API/persistence/resource spike и внешний DNS остаются gates; runtime `NOT_RUN`.
  Exact `v0.16.17` source and `UPGRADING/v0_16.md` supersede an earlier docs-based route
  assumption: both management Registry `x:*` calls and standard mailbox calls use fixed
  same-origin `POST /jmap`; `/api` contains OAuth/account/schema/telemetry helpers and is
  not the management-JMAP endpoint. The adapter must select credentials and method
  allowlists from a typed operation context, never from the URL, and must reject any JMAP
  batch mixing Registry and mail methods before I/O. A management API key authenticates
  only the reviewed Registry methods; it is never a mailbox credential. Enabled mailbox
  mode remains blocked pending the authentication/rotation/audit choice in founder
  question 42. Exact selected-
  version source rejects an ACL-only full-engine shortcut: shared access covers Email/
  Mailbox, while Identity, EmailSubmission, Sieve and Blob remain member-only. OAuth docs
  expose Authorization Code and Device Authorization, not client credentials. Root
  recommends per-user Authorization Code linking; zero-touch requires explicit founder/
  security acceptance of broad dedicated-account impersonation. The earlier Terra-requested
  source-only review is explicitly superseded after the route P0 was found. A fresh
  source-only re-review found no remaining P0/P1 and its sole P2 allowlist-wording issue is
  root-fixed, but actual model/effort telemetry was unavailable. Both review artifacts remain
  non-accepting because their superseded/current-SHA review chain is incomplete, not because
  unavailable telemetry alone invalidates an accepted native dispatcher record.
  OAuth also cannot be bolted onto the existing internal-only origin: one documented
  `STALWART_PUBLIC_URL` drives OAuth/OIDC and JMAP discovery. Founder question 45/G145-5D
  therefore requires one accepted user-and-Hub HTTPS origin with default-deny path/listener
  isolation. OAuth's exact anonymous rate-limited `POST /api/auth` must be a narrow
  path/method exception while both management Registry and mailbox `POST /jmap`, all other
  `/api/*`, upload/download, admin/static UI and telemetry remain Hub/service-only;
  prefix-wide `/api` is forbidden. Otherwise OAuth stays invalid and no HTTP listener is
  published.
  The management key is now root-frozen to API-key `Replace` with only domain and
  individual-account CRUD/list permissions, pending independent security review. A second
  selected-version mismatch is fail-closed: `x:Domain` exposes expected `dnsZoneFile` but
  no manual-DNS verification result, so Hub cannot infer `verified`, provision accounts or
  enable external mode before G145-5C/founder question 43. Exact candidate source confirms
  mailbox `Identity/get|set` supplies stable ids for configured account addresses, so
  145-03A is a two-context mapping on shared `/jmap`, not an identity-id mismatch;
  enabled use remains blocked by G145-5B mail auth. Question 44 freezes the exact bounded
  DNS record projection.
  Recovery-admin state is operator-only and absent from the accepted steady-state profile.
- Кодовые фазы 143–148 идут только после независимого принятия затронутого контракта.
  Phase 145 заблокирована 144 acceptance и внешними gates; независимые calendar slices
  сохраняют отдельно доказанную семантику.
- Mail 0.1.8 сохраняет Phase 147 reliability-контракт: REQUEST/CANCEL tombstones и
  recovery, outward REPLY reconciliation, private availability, durable in-app reminders
  и CRM manual-link. Новый amendment закрывает ScheduledCall split-brain: CalendarEvent —
  authority, call — atomic execution projection; backfill/cutover all-or-nothing,
  aggregate/availability single-count, due occurrence-keyed/versioned. Attach is exact-match,
  participant projection is series-only, and linked external guests fail closed pending a
  join-credential contract. It adds a separate versioned inbound CVE-v1 hash, accepts
  occurrence REQUEST without changing outbound 143-02A and fixes all-day identity to the
  UTC date marker without midnight inference. The historical root and independent
  Terra/high CVE-v1 review passed exact Mail 0.1.7 semantic SHA `b3fb6f48`; all three
  findings from `7a6ffc1c` were closed. That evidence is historical and does not review the
  0.1.8 engine, persistence or OpenAPI amendment. The earlier Mail 0.1.6 PASS on SHA `5a351d5b`
  is historical `STALE`.
  Founder task-local routing override authorizes exact `gpt-5.6-terra`/`xhigh` for
  atomic Mail product/test packets. Tasks 143-01, 143-02A and 143-02B are static accepted;
  bounded iTIP parsing includes reviewed CVE-v1,
  safe errors, post-unfold UTF-8 and host-independent timezone handling. Task 144-01A's
  `43d421ef` source is historical 0.1.7 evidence. The exact five-document
  default-vs-hard-cap/144-01A entry amendment passed at
  `5e314fb080d420696642058ae47397e251c23207`; it authorizes only the 144-01A entry, while
  broader Phase 142/cross-contract current-source review remains pending. Against that narrow
  amendment, the atomic seven-path 144-01A boundary
  at `e3e8c774d734f926f77f8e600b58dc68dabbd993` is `STATIC_ACCEPTED` after exact-commit
  Terra/high `EXACT_SHA_PASS`: projection-only 1 MiB-item/8 MiB-page values, budgeted 50 MiB message /
  25 MiB blob reads, stable first-draft identity, closed request/page/projection/body/
  attachment/blob/draft grammar, `confirmed_owner_draft`, and reference-only submission.
  Exact Terra/xhigh authoring and distinct-context Terra/high independent review are
  dispatcher-enforced with actual telemetry `UNKNOWN`; this is not a blocker by itself for an
  accepted native dispatch. The exact-commit review reports Spec/Quality/Security/Evidence PASS
  and C/I/M `0/0/0`. Any exposed mismatch, rejection, omitted model/effort or fallback remains
  blocking. The former 144-01B Task3 Docker-evidence clause is `CONSUMED/HISTORICAL`; static eligibility for 144-03B/03C/03D is
  status only and not a current dispatch authorization. Address is
  `STATIC_ACCEPTED` at the exact five-path atomic boundary
  `41564d003946ddb78df18efca166ca86e4055405` (parent
  `d659c67ce657fb33da391819edbdb954454be4b9`, tree
  `13a97333f08df6946e65766d3223c1ede85e1f50`, subject `feat(mail): add bounded address policy`)
  with committed code/test blobs `fe7472d6a5bf9baa140fef95352cbbf206f49612` and
  `67f35c39dfbbfdd92573f52103798e15de922b37`. Three independent exact-commit Terra/high reviews
  report `EXACT_SHA_PASS`, Spec/Quality/Security/Evidence PASS and C/I/M `0/0/0`; Terra/xhigh
  authoring and Terra/high review provenance are dispatcher-enforced, fallback `none`, with
  actual telemetry `UNKNOWN` unless exposed. The Thread callable/encoding/oracle contract is
  `THREAD_CALLABLE_ENCODING_ORACLE_STATIC_ACCEPTED` at
  `cbc12577a6444df9c660e43fedfd0c8133c6e461` / `EXACT_SHA_PASS` after independent
  Spec/API/Persistence, Security and Lifecycle/Evidence PASS reviews with C/I/M `0/0/0`.
  Thread implementation+oracle are narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`
  (source `c221d547`, oracle `b21bd6fd`, C/I/M `0/0/0`), static only. Folder source+oracle are
  narrowly `STATIC_ACCEPTED` at `f8cb7643986fc591cae16764575337b72ef3a692` (source
  `3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`, oracle `efe0db0071cb22ce4b026c635157f0fb08885d64`,
  C/I/M `0/0/0`), static only. Folder docs acceptance is `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`.
  `144-03B-Rule-Compiler` is narrowly docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at
  `f199546513b758ca7b95328b5ccd7680009b732e`, reviewed against transition commit
  `6735ae746c7acde657359e8a709cd897c4dce61c`; Contract/Spec Terra/high,
  Security/Adversarial Terra/xhigh and Lifecycle/Dependency Terra/high PASS C/I/M `0/0/0`,
  dispatcher-enforced with no fallback and telemetry `UNKNOWN`. Initial 144-02D/02F candidate
  `63e419897aa0683e1d3dfcd442aca78eea1b0727` and transition
  `6c424ed2d6703ceccbb269c37d9abc03e75a0410` are `CONSUMED/HISTORICAL`; RecentContact is
  superseded by corrective commit `7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus transition
  `e563a3d3be8898371e5c8a9861c58f11b72167bc`, now docs-only `STATIC_ACCEPTED` /
  `EXACT_SHA_PASS` after all three post-transition reviews PASS C/I/M `0/0/0`. The former
  pre-acceptance record is `CONSUMED/HISTORICAL`. Correction
  `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus transition
  `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only accepted. Rule-plan
  authoring/review/acceptance is `CONSUMED/HISTORICAL`; the accepted plan remains docs-only,
  with no source/schema/runtime authority.
  Other packet eligibility
  remains status only. The former 144-01B blocked clause is `CONSUMED/HISTORICAL`; broader
  Phase 142/cross-contract acceptance and Phase 148 remain pending. Launcher and all executable
  evidence remain `NOT_READY` / `NOT_RUN`.
  **2026-08-15 144-02A/B/C/D plus 02F trace:** correction scope is 02A Organization->Account
  and User->Account `NO ACTION` plus public metadata-only MailAccount `deleting`; 02C Account->MailObject
  `NO ACTION`, immutable `upload_drain_until` and storage-capability-gated no-I/O drain; and 02D canonical three-family
  User/RecentContact union plus deactivation/email/org-change aspects. Engine removal/drain/final Account
  hard-delete/physical Organization completion and physical User/GDPR deletion are separately `BLOCKED` /
  `NOT_DISPATCHABLE` pending their respective accepted engine-port and durable User-deletion-intent contracts.
  02F normative receipt/deployment acceptance remains unchanged; only its broad 02D-unchanged/
  dependency/writer-replay clause is correction-traced. The combined representation is docs/readiness-only
  `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `576f19231e37f635d5b7187a1ada121ed14ff8dc` after Contract
  Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0`; the former candidate/
  precommit/root-commit/exact-review/status-sync-next route is `CONSUMED/HISTORICAL` and provenance is
  dispatcher-enforced/no fallback/telemetry `UNKNOWN`. No source/Ent/generation/migration/service/
  test/runtime authority. Launcher `NOT_READY`, executables `NOT_RUN`; 144-01B, Phase 142/148,
  carrier/Rule receipts and G145-5B/C remain unchanged.
  Root-frozen `PERSISTENCE-CONTRACT-0.1.8.md` supplies the exact 144-02A/B/C/D/F fields,
  tuple/state checks, normalized indexes, scalar attachment staging, sole default/scan
  authorities, quotas, generic receipt/bulk matrices and canonical rule hash. OpenAPI 0.1.8
  adds bounded cursor/detail pages, creation quota errors, compact 366-day calendar pages and
  non-truncating legacy count/byte ceilings. The earlier review files remain discovery
  evidence, not current PASS verdicts; every dependent packet waits for an independent
  Terra/high review with accepted per-dispatch provenance of the exact amendment.
- A requested Terra/high source-only pass on exact SHA `90229b2b` is not an acceptance because
  it found three
  P1 contract gaps now root-remediated for re-review: required `MailMessage.engine_version`,
  separate opaque engine submission/result ids versus the public Hub message UUID with
  projection-before-sent finalization, and a locked durable maximum of 50 attachments per
  draft. Product code remains unauthorized by this source-only remediation.
- Its source-only re-review found and root-remediated two follow-up P1s without changing that
  gate verdict: retained engine-result pairs now skip submission reconciliation and can only
  retry projection finalization, while the hydrated projection must exactly match account/
  message/version and receives domain-derived confirmed-owner origin. Another independent
  exact-SHA Terra/high review remains required.
- The 144-02A root candidate now also fixes exact Ent scalar widths, lower-case ASCII
  validators, a typed bounded DNS JSON value, named SQL CHECK predicates, ordered partial
  indexes and delete edges. Execution is split sequentially into 02A1 typed DNS, 02A2
  domain/policy and 02A3 account/identity so every intermediate source remains compilable.
  This is representation freeze only: the historical pre-receipt independent-contract-review
  pending wording is superseded by the `b631866b2ba372101fd80eea30951eaa267c1931` docs-only
  acceptance; source-level review, migration authority, generation and Docker runtime remain
  pending/`NOT_RUN`.
  The exact test-first packet is recorded in
  `docs/superpowers/plans/2026-08-13-mail-control-plane-schemas.md` and consumes contract base
  `5f31a74f`; it authorizes no product write while explicit source dispatch, source-level review
  and other packet gates remain pending.
  Evidence is recorded in the matching `IMPLEMENTATION-*.md` artifacts. A later root deep
  audit reopened 143-03A and 143-03C over the schema-level nullable canonical hash version/hash
  pair CHECK required by Mail 0.1.8 and the phase plan. Packet A now has the exact CHECK plus an
  external Ent-metadata oracle and is independently `STATIC_CONFORMANT` (Critical/Important/Minor
  `0/0/0`) under exact Terra/xhigh implementation and Terra/high independent-review routes with
  dispatcher-enforced provenance and actual telemetry `UNKNOWN`; the state, schema, external test
  and evidence document form committed Task 1 boundary `31bfff8f58ba83d12b59600135503f98a6c92fbd`.
  Packet C now has the exact CHECK plus external Ent-metadata oracle and is independently
  `STATIC_CONFORMANT` (Critical/Important/Minor `0/0/0`) under exact Terra/xhigh implementation
  and Terra/high independent-review routes with dispatcher-enforced provenance and actual telemetry
  `UNKNOWN`; docs sync uses Terra/medium. Its schema, external oracle and evidence document form
  the committed atomic Task 2 boundary `444c6f7b7b7b3d9b77bc9b24b32d1e13000ebbbc`.
  Both hash-pair P1 remediations are statically conformant, but downstream acceptance remains
  separate. 143-03B is now `STATIC_ACCEPTED` on clean `dd346e31`, C/I/M `0/0/0`, with
  requested Terra/high independent re-review, dispatcher-enforced provenance and actual telemetry
  `UNKNOWN`; its original source commit remains `61c19f9f`, while the attendee source later includes
  the accepted 03F inverse-edge remediation. Fresh 03F review reopened missing owned-child Cascade
  and exact database SHA-256 length enforcement at C/I/M `0/2/0`. The committed fix adds
  Cascade only to `CalendarAttendee.inbound_reply_receipts`, adds named CHECK
  `calendar_inbound_reply_receipts_payload_hash_length` = `octet_length(payload_hash) = 32`, and
  adds the genuine external metadata unit
  `backend/tests/calendar/calendar_inbound_reply_receipt_schema_test.go` for those exact risks.
  Static expected RED is recorded; runtime RED/GREEN is `NOT_RUN`. The independent 03F Terra/high
  review, dispatcher-enforced with actual telemetry `UNKNOWN`, reports `SPEC PASS`/`QUALITY PASS`,
  C/I/M `0/0/1`; the Minor was the docs-sync finding and is superseded/closed. The final independent
  docs/whole-boundary review returned C/I/M `0/0/0`; 03F is `STATIC_ACCEPTED` at committed boundary
  `846cd00ecb243402877e18239c89a62bf07e8d14`. The two-packet remediation remains in
  `REMEDIATION-143-03-HASH-PAIR.md` and
  `docs/superpowers/plans/2026-08-13-calendar-hash-pair-remediation.md`. 143-03D is
  `STATIC_ACCEPTED` on `10b4ceead119cb74a16b325b33f9121f60d7b1b5` after independent C/I/M
  `0/0/0`; `IMPLEMENTATION-143-03D.md` records the static boundary and authoritative Git-blob
  SHA-256 values. 143-03E is no longer blocked by 03D or 03F: it remains
  `BLOCKED/NOT_RUN` on the accepted launcher `NOT_READY` and separate founder-approved migration
  mechanism/gate. The all-A product decision does not establish that technical authority or
  accept an executed migration. 144-01B remains subject to its own exact-contract and independent-review
  dependencies. This does
  not accept Phase 143, the Calendar product, migration or runtime behavior: generation,
  compilation, unit execution and runtime are `NOT_RUN` because the accepted launcher is
  `NOT_READY`.
-  143-04L is `STATIC_ACCEPTED` at the authoritative code/test boundary
  `4525c559da22729aa8bc51ff5336c29dad4f27e9`: implementation was Terra/xhigh; the initial code
  review was historical C/I/M `0/2/2`; the immutable-envelope/payload contract amendment was
  independently accepted; and the remediated boundary's final Terra/high review is PASS,
  C/I/M `0/0/0`. Requested-route provenance is dispatcher-enforced and actual telemetry is
  `UNKNOWN`. The accepted paths are `backend/internal/platform/events/envelope.go`
  (Git blob `add892cde5b8fab0fb0c19d08a576d2c915c39b4`, SHA-256
  `8891E2BEDC0DA40E4C77630374F0B3BDCEEA5FD1D9F281F49049706BC4D55D5A`) and
  `backend/tests/calendar/event_envelope_test.go` (Git blob
  `d58055044769d1961005285c31a910cfcc66feae`, SHA-256
  `4D11A3EDCD5041650D738698DA250DF3ADA769B6FB3D7106161B324EF3540750`). It freezes an
  immutable complete JSON envelope, snapshot JSON-object payload, copy-returning accessors and
  `MarshalJSON`, strict 1..255 UTF-8-byte `.v1` subjects, semantic zero-offset UTC `Z`, and
  stable content-free `ErrInvalidPayload`. Static RED tests are present; the launcher remains
  `NOT_READY`, so compilation, tests and runtime remain `NOT_RUN`. This does not accept the
  Phase, product, global Mail review or runtime behavior.
- Mailbox control-plane rules inherited from 0.1.6 remain in 0.1.8: отдельные bounded workers владеют
  idempotent domain/account/primary-identity/alias lifecycle, combined projection/mutation-recovery sync и generation/hash
  reconciliation одного fixed Hub Sieve script. Generic create/bulk receipt хранит только
  hashes/target/status с 30-day replay и lifetime reject tombstone, различает reused/
  unavailable/expired и для presign обновляет URL только того же pending object; junk/ham
  требует отдельный classifier outcome. Label/settings/signature/rule/policy/domain writes
  теперь имеют wire-level `If-Match`, а active-rule label delete — отдельный safe conflict.
  Server rich content sanitizes or falls back to plain-only, remote images use owner/message
  tokens и возвращают только bytes с private/nosniff/no-referrer/CSP, а Meili получает только bounded sanitized body
  и normalized filenames с повторной owner authorization. Phase 144 также владеет leased
  bounded worker, atomic projection-only page/full-envelope/cursor apply, separate command lease
  with begin/end mutation epoch fence и
  resumable no-clear rebuild; Phase 145 владеет incremental + base-state/ID-anchored
  Stalwart mapping и root-context composition. Folder/thread/search API/UI показывают
  bootstrapping/degraded/rebuilding freshness без false-ready. Всё это contract-only;
  adapter/workers/runtime остаются `NOT_IMPLEMENTED` / `NOT_RUN`.
- Docker Desktop 29.7.2 доступен из сессии, но это не меняет канонический launcher
  `NOT_READY`. Folder recovery дополнительно fail-closed: exact-match/absence разрешены
  только при доказанном Hub-exclusive writer; user IMAP/JMAP/ManageSieve/direct admin
  mailbox mutation не публикуются в wave 1.
- Phase 145 official-source research is recorded in
  `RESEARCH-2026-08-13-STALWART-0.16.17.md`. Per-mailbox Hub rules must use RFC 9661
  `SieveScript/*` on mail JMAP; management `x:SieveSystemScript` / `x:SieveUserScript` are
  explicitly excluded. `x:SpamTrainingSample` does not currently prove Hub's exact
  message-to-spam/ham mapping or idempotent lost-response recovery, so 145-07C remains
  blocked rather than guessed.
- Вопросы, не останавливающие локальную работу, собраны в
  `docs/mail/FOUNDER-QUESTIONS.md`; вопросы 22–45 фиксируют legacy idempotency, linked-call
  guests, primary-address/sanitizer/Sieve policy, later engine/API/schema defaults и
  отдельную least-privilege mailbox-auth design decision, authority проверки DNS и DNS
  projection.
  Push/PR/merge/deploy не авторизованы.

## Session note — 2026-08-08 (test-rebase execution readiness)

- Research/GSD/Superpowers synthesis is now executable Hub policy, not a loose
  report: atomic commits `a9077796`, `b1b69124`, `724947f7`, `8220370e` add
  bounded TaskContract/context/evidence/workflow-eval contracts, conclusion-blind
  verification, separate 3-fix/5-review breakers and fail-closed Docker runner
  requirements. Independent prompt-contract re-review = **PASS**; independent
  test-contract re-review = **PASS** after `77f94e89` removed the last four
  direct Docker launcher bypasses.
- Touchstone `pro-qa-ai` is current `0.4.0-beta.3` at commit `30f75a25`, with
  `STRUCTURAL_PASS`; the Desktop portable copy is also beta.3. Paired
  effectiveness/ROI evaluation remains `NOT_RUN`, so this is not GA.
- TASK-202 имеет статус **archived/verified**; live evidence — 1,049 rows /
  4,246 declared, ledger SHA `72de6a81…` на source revision `5708671f`.
- Phase 129 разбита на семь ordered executable plans. Каждый move/delete batch
  обязан в одном atomic commit выполнить filesystem operation, переключить все
  affected CI/scripts/docs, обновить `cutover-manifest.json` и приложить
  Docker-only evidence; только independent verifier ставит `verified`.
- Synthetic task aliases удалены из execution contract: deletion и финальная
  zero-consumer hook closure принадлежат реальной TASK-219; TASK-201 разделена
  на per-batch switch и final cleanup.
- Phase 136 требует 131; Phases 134/135 формально требуют 133. Все 245
  hook-dependent genuine Go units остаются protected до TASK-234.
- Product/test files не перемещались и не удалялись; validator/inventory Python
  и `pro-qa-ai/**` не менялись.
- Touchstone `0.4.0-beta.3` остаётся foundation-ready pre-release; paired eval
  `NOT_RUN`.

## Session note — 2026-08-07 (test architecture rebase + Phase 128)

- Текущая product-truth (founder, 2026-08-12): Hub остаётся незавершённым MVP;
  ни один продуктовый модуль не принят как complete или stable. Исторические
  milestone COMPLETE не равны production acceptance; PG-03/FA-06 —
  зафиксированный ops-блокер production cutover, не индикатор готовности MVP.
- Test-план полностью rebased: 14 фаз 128–141, target
  `backend/tests/**` + `frontend/tests/**` + единый корневой Python-проект
  `testframework/**` без вложенного `testframework/python/`. Канон:
  `active/august-2026/06-TEST-ARCHITECTURE-REBASE.md`.
- TASK-202 пересчитан на текущем pinned source revision `5708671f`: 1,049 rows
  / 4,246 declared; Go total 832/3,674 (legacy root 831/3,672 + server
  1/2; 595 genuine unit + 237 component: 31 API + 206 integration), Vitest 180/519
  (156 unit + 24 static contract) + 3 support, Playwright 20/53 + 14 support,
  hook sources 43 files / 582 symbols / 892 links, target Python 0.
  Первый qa-auto-ver verdict = FAIL по четырём schema/entrypoint gaps; после
  repair повторная независимая проверка = **PASS, verified true**. Current
  ledger SHA: `72de6a81b2bd5da8873a98e784a5069a661e9caea083f10722aa85e9f91b4511`.
- Docker Desktop 4.40 / Engine 28.0.4 доступен после approved daemon access.
  Начальная попытка TASK-203 была красной и сохранена как historical triage
  evidence; частичные/red profiles не принимались.
- TASK-221…225 закрыли path guards, GitLab/NATS/cron nondeterminism, frontend
  regressions, Biome debt и coverage gap отдельными проверяемыми scopes.
  Финальный backend suite green: 50 packages PASS / 0 FAIL / 3 no-test-files;
  exact merged profile **82.0373%** (21,269 / 25,926), security floors **6/6**,
  audit **85.6879%**. Порог и denominator не менялись.
- Frontend Docker observation: TypeScript PASS, Biome 370 files / 0
  diagnostics; frozen legacy Vitest **510/519 PASS, 9 FAIL** (6 absent docs
  mount, 2 nested fixture timeout, 1 stale i18n assertion). Founder cancelled
  legacy remediation; this is discovery evidence, not a cutover gate.
- TASK-224 закрыт реальным изолированным Docker baseline: smoke **3/3 PASS**;
  full без retries **48 PASS / 5 expected skip / 0 fail**. Два найденных
  production gaps (`GET /api/admin/users`, стабильный Telegram API без bot
  token) исправлены отдельными атомарными коммитами; Messenger locator flake
  устранён и full повторён без retries.
- Founder superseding decision: только genuine isolated units сохраняются и
  перемещаются в `backend/tests/**` / `frontend/tests/**`. Все legacy Go
  component/API/integration, static-contract Vitest, TS Playwright и их
  support/config/entrypoints frozen и удаляются по exact manifest, без port и
  row-parity. Historical 82% Go coverage не является migration blocker.
- Весь non-unit QA создаётся с нуля на Python прямо в `testframework/**` от
  task/FR/NFR/verified TC и risk matrix: happy, negative, authn/authz,
  boundaries, edge/corner, failure/recovery, security/a11y/performance где
  применимо. Docker-only и resource limits обязательны.
- Task-owned `hub-e2e224-*` containers, network, volume и temporary frontend
  image удалены после evidence; shared `deployments-*` оставлены работающими.
- По явному founder authorization выполнен read-only policy extraction из
  `pro-qa-ai/**`; source не менялся. Hub получил
  `.cursor/rules/07-agent-safety.mdc`, усиленные delegation/handoff и запрет
  push/PR/deploy без явного поручения. Локальные commits — атомарно allowlist-ом.

## Session note — 2026-08-07 (team-lead onboarding + model routing)

- Team-lead заонборжен по `AGENTS.md`, live `STATE`, `.planning/README.md`,
  августовскому плану и `.cursor` regulation/playbooks. Канон следующей работы
  не изменён: resume с TASK-193; до старта выбрать единый coverage baseline.
- Добавлен `.cursor/playbooks/model-routing.md`: role defaults всех 32 агентов,
  effort/escalation/fallback и cost snapshot. `team-lead` = Sol/Extra High;
  Terra = основной engineering runtime; Luna = механика; Spark = optional preview.
- Downstream frontmatter сохранён `inherit` для portable запуска; фактический
  runtime/reason/fallback/catalog status обязателен в delegation/handoff.
- `pro-qa-ai/**` закреплён как protected external zone Claude Code Opus 5 под
  управлением founder. Внутренние Hub agents не читают её по умолчанию, не
  меняют и не включают в волны; только явный founder handoff конкретного файла.
- Синхронизированы ADR, AGENTS, REGULATION, team workflow, agent descriptions и
  ownership rules. Product code/tests и `pro-qa-ai/**` не менялись.

## Session note — 2026-07-30 (Touchstone Standard + Pro packaging)

- **Touchstone Standard** (`pro-qa-ai/`): продукт-идентичность `TOUCHSTONE.md`,
  версия **`0.2.0-rc.1`** foundation-ready, CHANGELOG, обновлены README/START-HERE/INSTALL/MANIFEST/VERSION.
  Structural PASS сохранён; effectiveness/ROI по-прежнему **NOT_RUN** (честный GA-gate).
  (prompt-contract, master-delegate, wave-protocol, mutation-evidence, backlog-task);
  источники: GSD inventory, PE Standard, `.cursor/` (contract/team/verify), `.planning/` (backlog/STATE/mutation).
- Синхрон: `.planning/README.md`, `.cursor/rules/70-pro-qa-ai.mdc`,
  `REGULATION.md` §20, `playbooks/README.md`, `AGENTS.md`.
- Код продукта / validator scripts не менялись. Коммитов нет (ожидание запроса пользователя).

## Session note — 2026-07-30 (qa-auto1 · TASK-008 hubtest close)

- Full hubtest **PASS**: `go test -tags hubtest -count=1 -timeout=600s ./testframework/...` → **50** ok · **0** FAIL · ~62.2 s. Evidence: [`HUBTEST-2026-07-30-task008.md`](audits/total-2026-07/HUBTEST-2026-07-30-task008.md).
- Prior FAIL (project FA11 second-member mock · storage UUID `entity_id`) — ROOT=test, fixed; packages green.
- [`00-FINAL.md`](audits/total-2026-07/00-FINAL.md) → вердикт **A**; TASK-008 / README → **Закрыта**; Close control → **DONE**.
- Residual: overall COV 80.2% vs 81.9% baseline (не gate 1–4). Коммитов нет.

## Session note — 2026-07-30 (docs-writer · TASK-008 W20 FINAL draft)

- Собран draft [`.planning/audits/total-2026-07/00-FINAL.md`](audits/total-2026-07/00-FINAL.md): вердикт **B** (superseded — hubtest re-run PASS → **A**, см. note выше).
- Gates на тот момент: README **0** «На проверке» ✅ · Stage C drain ✅ · messenger/rbac ✅ · hubtest **FAIL** (project FA11 + storage entity_id).
- Коммитов нет (инструкция слота).

## Session note — 2026-07-30 (qa-auto1 · TASK-184 C-W19-TT-001 last-member · W13 slot 4)

- **FIXED** на HEAD: `RemoveTeamMember` — `TeamMemberCount≤1` → 409 Conflict (`cannot remove the last member of a team`). Prior be1 «FIXED»/На проверке не было на HEAD (W13 GAP NOT_PROVEN).
- **Тесты GREEN:** `TestRemoveTeamMember_RejectsLastMember_TASK184`, `TestRemoveTeamMember_AllowsWhenMultiple_TASK184`; wave122 → Conflict; wave124 multi OK.
- **Mutation PROVEN:** убрать guard → оба TASK184 RED (`err==nil`); restore → GREEN. Report: `MUTATION-2026-07-30-W13-qa1-task184-fix.md`. Leftover MUTATED: нет.
- Docs: `docs/TESTING.md` TC-PROJ-11/12; API.md уже 409. Коммит не создавался.
- Статус: **Закрыта**.

## Session note — 2026-07-29 (be1 · TASK-184 C-W19-TT-001 last-member)

- be1 session claimed FIXED / На проверке — **не на HEAD** (подтверждено W13 qa-auto1 GAP 2026-07-30). Реализация + mutate — qa-auto1 refill ниже.

## Session note — 2026-07-29 (be1 · TASK-191 EA-001 AutomationRule soft-delete)

- **FIXED** EA-001: `AutomationRule.deleted_at` + indexes; `DeleteAutomationRule` → `SetDeletedAt` (не hard `Delete`); list/count/find/update/ListEnabled → `DeletedAtIsNil`.
- **Migrate:** `backend/migrations/20260729_task191_automationrule_soft_delete.sql` marker `needs-owner-apply` — **BLOCKED waiting owner** (SQL в репо, не применён на prod).
- **Тесты:** `automation_soft_delete_task191_test.go` + `schema_automationrule_task191_test.go`; source-guard `.Delete().` banned; mutation hard Delete → RED; restore → GREEN. `./testframework/{tracker,ent}/ -run TASK191` PASS; automation regression PASS.
- Не трогали: SEC-02 / VersionEQ webhook (TASK-152), RemoveTeamMember (184), calendar (189). TASK-008 не закрывали.
- Статус: **На проверке**. Handoff: be-ver → code-reviewer.

## Session note — 2026-07-29 (be1 · TASK-189 BE-V4 calendar trash + draft poll)

- **FIXED** BE-V4-001/002: `ListApprovedLeaveInRange` / `ListSprintsInRange` → `DeletedAtIsNil`.
- **FIXED** BE-V4-003: `GetNewsPoll` / `VoteNewsPoll` → `requirePublishedNewsOrAuthorOrAdmin` (draft → 404 non-author).
- **Тесты:** `leave_sprint_trash_task189_test.go`, `draft_poll_idor_task189_test.go`; mutation RED×3 → restore GREEN; `./testframework/{calendar,intranet}/` hubtest+race PASS.
- Docs: `docs/wiki/API.md`, `docs/system_specification.md`. Не трогали coffee/partners (163), preview (180). TASK-008 не закрывали.
- Статус: **На проверке**. Handoff: be-ver → code-reviewer.
## Session note — 2026-07-29 (be2 · TASK-188 ASN-001/002 PushSender + Telegram RL)

- **FIXED** ASN-001: `PushSender` → `PushSubscriptionStore` (CRUD в `Repository`); `push.go` без `*ent.Client` / ent import; wiring `NewPushSender(notifRepo, …)`.
- **FIXED** ASN-002: `ConfirmLinkCode` Redis fails `tg:linkcode:fails:<user>` max 5 / TTL 5m (как HI-04); HTTP → 429 `ErrTelegramConfirmLimited`.
- **Тесты:** `push_layering_asn001_test.go`, `telegram_linkcode_rl_asn002_test.go`; `./testframework/notification/...` PASS; mutation (убрать fails/record) → RED. NF-07 / Magic Link rate не трогали. TASK-008 не закрывали.
- Статус: **На проверке**. Handoff: be-ver → code-reviewer.

## Session note — 2026-07-29 (fe1 · TASK-192 FE settings/AI version OL)

- **FIXED** FE-SMA-001: webhook toggle/delete/deliveries — `version` в DTO + PUT; toast на fail.
- **FIXED** FE-SMA-002: AI privacy + ai-store — `version` в load/save; 409 → toast + refresh.
- **FIXED** FE-SMA-003 (optional): marketplace «Мои» — `mineLoading` / `listLoading` ≠ empty.
- TASK-179: `AVAILABLE_EVENTS` без `sprint.completed` — без регрессии.
- Тесты: webhooks-version-ol · ai-privacy-version-ol · marketplace-mine-loading · TASK-179 · ai-store; mutation omit version → RED, restore GREEN.
- Статус TASK-192 → **На проверке**. Handoff: fe-ver → code-reviewer. TASK-008 / README не трогали.

## Session note — 2026-07-29 (docs-writer · TASK-183 C-W19-DOC-001 inventory/AGENTS/spec)

- **FIXED** C-W19-DOC-001: shell recount HEAD `d868ba77` (`cursor-grok-4.5`) — BE `*_test.go` **812** / Vitest **179** (61 `.ts` + 118 `.tsx`) / hooks **39**; дата measure **2026-07-29**.
- Синхрон: `AGENTS.md`, `docs/{README,test_plan,TESTING,testframework_strategy,TROUBLESHOOTING,system_specification,user_needs_analysis_2026,vision_and_scope,competitive_audit}.md`.
- Anti-stale: убраны live `515` / `99/99` / Vitest `94` / HEAD `bbdaf16`; US-M06 webcam → shipped (не Phase 16 IN).
- Статус TASK-183 → **На проверке**. TASK-008 не закрывали.

## Session note — 2026-07-29 (fe2 · TASK-182 HR pending leave error≠empty)

- **FIXED** C-W19-FE-005: admin `listPendingLeave` fail → секция + QueryErrorAlert/retry (не silent empty); empty queue только success+len0.
- Тест: `domains/hr/leave-pending-error.test.tsx` 2/2 PASS · mutation (убрать `pendingError` из gate) → RED · restore GREEN.
- Статус TASK-182 → **На проверке**. Handoff: fe-ver → code-reviewer. TASK-008 не закрывать.

## Session note — 2026-07-29 (be2 · TASK-185 C-W19-TT-002 Allure ExistingRun)

- **FIXED** C-W19-TT-002 path A: ExistingRun Allure — fail-closed на Mongo `StoreAllureResults` и match infra err (≠ NotFound); порядок match→Mongo→PG; create-path Mongo best-effort без изменений.
- **Тесты:** `testframework/tms/allure_existingrun_partial_c_w19_test.go`; mutation slog-only+continue → RED; `./testframework/tms/` hubtest+race OK.
- **Docs:** `docs/wiki/API.md`. Статус: **На проверке**. Handoff: be-ver → code-reviewer. TASK-008 не закрывали.

## Session note — 2026-07-29 (be1 · TASK-177 C-W19 deactivate/ban WS kill)

- **2/2 FIXED** C-W19-001 / C-W19-003: Deactivate/Ban → `SessionService.RevokeAllUserSessions` (Redis Del + NATS `revoked`) → Bridge рвёт живой WS; kill fail → fail-closed admin error.
- Wiring: `wireAuth` → `auditSvc.SetSessionRevoker(authSvc.Sessions)`.
- Тесты: audit mock revoker + fail-closed; auth NATS/Redis; ws live NATS e2e. Mutation: remove publish → RED → restore GREEN.
- Commits: `88494530` fix · `460fa131` test · `a09bdc0e` task file.
- Docs: `docs/security_guidelines.md`. Статус: **На проверке**. Handoff: be-ver → code-reviewer. TASK-008 не закрывать.

## Session note — 2026-07-29 (fe1 · TASK-187 FE-TA pipeline sanitize + P3)

- **FIXED** FE-TA-001: `pipeline-status` — `isSafeUrl`/`sanitizeUrl` на `html_url` (unsafe → без `<a href>`).
- **FIXED** FE-TA-002: `issue-attachments` — whitelist URL; unsafe → disabled chip без href.
- **FIXED** FE-TA-003: planning-poker — `loadFailed` → `QueryErrorAlert`+retry; empty → «не найдена».
- **FIXED** FE-TA-004: automation secondary statuses/channels — `isError` + compact alert (не пустой Select).
- FE-TA-005 (token cluster) — out of scope (hygiene).
- Vitest tracker: 21 PASS (sanitize + error≠empty + source-guard `href={run.html_url}` banned). tsc OK.
- Статус: **На проверке**. Handoff: fe-ver → code-reviewer. TASK-008 не закрывали.

## Session note — 2026-07-29 (fe2 · TASK-186 calls error≠empty + a11y)

- **6/6 FIXED** (FE-MC-001…006): channels fail → alert/retry (не пустая сетка); projects/channels dialog fail → alert (empty только success+len0); popover × / channel select / remove participant / week slot a11y names.
- **Тесты:** vitest calls `error-empty` + `a11y` + recordings-noopener **9/9 PASS**; mutation `channelsError=false` → RED.
- **Статус:** На проверке → fe-ver → code-reviewer. README не трогали (scope). TASK-008 не закрывали.

## Session note — 2026-07-29 (be2 · TASK-181 C-W19-004 webhook soft-delete deactivate)

- **FIXED** C-W19-004 / C-W19-FE-004: `SoftDeleteWithCrossLinkCleanup` в той же tx `SetIsActive(false)` на webhook endpoints; `ListWebhookEndpointsByProject` + `project.DeletedAtIsNil()` (trash → пустой list).
- ListActive `HasProjectWith(DeletedAtIsNil)` не ослабляли (TASK-152 lock).
- Тесты: `testframework/apigateway/task181_c_w19_004_test.go` (runtime + source mutation).
- Docs: `docs/wiki/API.md`. Статус: **На проверке**. Handoff: be-ver → code-reviewer.

## Session note — 2026-07-29 (be1 · TASK-169 A3W8-3 HubTable/JournalEntry Cascade)

- **2/2 FIXED** A3W8-3-001/002: `entsql.OnDelete(Cascade)` на HubTable→columns/records и JournalEntry→lines.
- Migrate SQL (needs-owner-apply): `backend/migrations/20260729_task169_hubtable_journalentry_cascade.sql`.
- Тесты: `testframework/accounting/cascade_hubtable_journalentry_task169_test.go` (hard CASCADE + soft unchanged).
- `go test -tags hubtest ./testframework/accounting/ -run TASK169` PASS. Handoff: be-ver → code-reviewer. Статус: На проверке.

## Session note — 2026-07-29 (be1 · TASK-144 A3W3-2 messenger services)

- **Сделано:** 6/6 A3W3-2 FIXED — Join IsMember fail-closed, Create extra-members rollback, bookmark channel-scope, star/mute NotFound→403, ClosePoll membership до close, AddMembers/DeleteChannel AppError wrap; hubtest messenger `-race` PASS + 6/6 mutation RED.
- **Коммит:** `18b3d72c` (+ docs evidence); статус **На проверке**; README не трогали.
- **Handoff:** be-ver → code-reviewer.

## Session note — 2026-07-29 (be1 · TASK-171 A3W8-5 Team/WorkflowStatus soft + Cascades)

- **7/7 FIXED** (A3W8-5-001…007): Team/WorkflowStatus `deleted_at`+`updated_at`; soft `DeleteWorkflowStatus` + 409 при issues; OnDelete Cascade members/versions/results/deliveries/from_transitions(+to_status).
- **Hubtest:** `./testframework/ent/ -run TASK171` PASS (9); `./testframework/tracker/ -run TASK171|IdempotentSoft` PASS (3); project smoke PASS.
- **Migrate:** `20260729_task171_soft_delete_ondelete_cascade.sql` — `needs-owner-apply`.
- **Статус:** На проверке → be-ver → code-reviewer. README не трогали (owner).

## Session note — 2026-07-29 (be1 · TASK-155 A3W5-3 ai service)

- **9/9 FIXED** (A3W5-3-001…009): UTF-8 composer trim; usage/settings upsert; issue_lookup fail-closed; consent Default(false)+Create deny; VersionEQ 409; WS sanitize; BotUserRepository; bot-flow asserts.
- **Hubtest:** `./testframework/ai/` `-race` PASS (~19s). Mutation 001 observed red→green.
- **Approval:** Atlas ALTER DEFAULT `allow_external_ai` — waiting owner (schema+Create already deny).
- **Статус:** На проверке → be-ver → code-reviewer. README не трогали.

## Session note — 2026-07-29 (be1 · TASK-170 A3W8-4 sprint schemas)

- **2/2 FIXED** (A3W8-4-001/002): partial unique active sprint (`IndexWhere` + migrate SQL); Sprint→issues `OnDelete(SetNull)`.
- **Hubtest:** `./testframework/ent/ -run TASK170` PASS (5); `./testframework/sprint/ -run SoftDelete|SecondActive|StartSprintAtomic` PASS.
- **Migrate:** `20260729_task170_sprint_active_unique_setnull.sql` — `needs-owner-apply` + pre-migrate cleanup duplicates.
- **Статус:** На проверке → be-ver → code-reviewer. README не трогали (owner).

## Session note — 2026-07-29 (be1 · TASK-134 A3W1-2 tracker services)

- **6/6 FIXED** (A3W1-2-001…006): UpdateIssue `CountLabelsInProject`; CreateLabel `FindLabelGroupInProject`; Create/UpdateRule `ensureActionChannelInProject`; ListThreadIssueLinks `WithIssue` eager; CreateIssueLink AppError wrap; UpdateComment 404≠409.
- **Hubtest (re-verify после crash):** focused A3W1_2_* + P1BE03/A3_011 PASS (2.1s); full `./testframework/tracker/` PASS (6.5s, `-race`). Mutation smoke A3W1-2-001: без guard → FAIL, guard восстановлен.
- **Статус:** На проверке → be-ver → code-reviewer. README не трогали (owner). Чужой WIP (workflow soft-delete / project repo) не коммитили.

## Session note — 2026-07-29 (be1 · TASK-137 A3W1-5 project residual)

- **6/6 FIXED** (A3W1-A5-001…006): Archive/Unarchive на trash → 404 без ClearDeletedAt; PermanentDelete только deleted+system admin; CreateTeamWithCreator Ent-tx; JoinByInvite MemberCount fail-closed; template List/Export без raw `err.Error()`; import >25MB → 413.
- **Hubtest (isolated worktree):** `go test -tags hubtest -race -count=1 ./testframework/project/` PASS.
- **Docs:** `docs/wiki/API.md` — archive/unarchive/restore/permanent semantics. Статус: **На проверке** → be-ver → code-reviewer. README не трогали (owner: только SCOPE).

## Session note — 2026-07-29 (be2 · TASK-164 A3W7-3 integrations)

- **3/3 FIXED** (A3W7-3-001…003): `ListSprintIssues` + GH/GL `FindIssue*` → `DeletedAtIsNil`; `/link-preview` issue path → `IsProjectMember` + per-user cache + Fiber v3 handler/auth order.
- **Hubtest (isolated worktree):** `./testframework/imex/` `./testframework/github/` `./testframework/gitlab/` PASS; mutation 001/002/003 FAIL as expected.
- **Статус:** На проверке → be-ver → code-reviewer. README не трогали (owner: только SCOPE).

## Session note — 2026-07-29 (be1 · TASK-155 A3W5-3 ai service)

- **9/9 FIXED** (A3W5-3-001…009): UTF-8 composer trim; IncrementUsage/settings upsert+retry; issue_lookup fail-closed; preference consent false; VersionEQ+409; WS `AI stream failed`; BotUserRepository; flow asserts (+ mention/consent fixtures).
- **Hubtest:** `go test -tags hubtest -race -count=1 -timeout=600s ./testframework/ai/` → **PASS** (~19.6s).
- **Docs:** `docs/wiki/API.md` (preferences/providers version + consent default + stream sanitize). Статус: **На проверке** → be-ver → code-reviewer. README не трогали.

## Session note — 2026-07-29 (be2 · TASK-147 A3W3-A5 calls)

- **4/4 FIXED** (A3W3-A5-001…004): orphan LiveKit room на DB-fail CreateCall/Instant; FindConflictingCalls + soft-delete channel; nil project_id → 400; title≤255 / guest_name≤100 runes → 400.
- **Тесты:** `go test -tags hubtest -race -count=1 -timeout=300s ./testframework/calls/` PASS (в т.ч. `a3w3_a5_task147_test.go`).
- **Статус:** На проверке → be-ver → code-reviewer.

## Session note — 2026-07-29 (be1 · TASK-165 A3W7-4 storage/dashboard/backup residual)

- **5/5 FIXED** (A3W7-4-001…005): ProjectStats/Admin counts fail-closed; backup S3 delete Warn+fail-closed; GetLayout membership; SoftDelete→404; burndown AggregateProjectCompletedByDay.
- Hubtest: dashboard+storage PASS (частичный); полный backup+регрессия — re-run be-ver при стабильном `ent/`.
- Docs: `docs/wiki/API.md` (layout 403 / project+admin fail-closed). Статус задачи: **На проверке**.

## Session note — 2026-07-29 (fe1 · TASK-124 A3W10-1 routes/modules error≠empty)

- **Сделано:** secondary accounting / TemplateSelector / marketplace pending / bases detail+table — `QueryErrorAlert`; recordings download `noopener,noreferrer`. 5/5 FIXED → **На проверке**.
- **Тесты:** vitest v4 + `recordings-download-noopener` **31/31** PASS; `tsc --noEmit` PASS.
- **Docs:** backlog TASK-124 Progress; README не трогали.

## Session note — 2026-07-29 (be1 · TASK-135 A3W1-3 tracker repos residual)

- **3/3 FIXED:** A3W1-A3-001 FilterSubjects+known-no-op Ack; A3W1-A3-002 executeRule Post/Notify → error → Nak; A3W1-A3-003 StartTimer `DeletedAtIsNil`.
- hubtest tracker: focused ok 4.178s · full `./testframework/tracker/` ok 12.654s (-race).
- Статус: **На проверке** → be-ver → code-reviewer.

## Session note — 2026-07-29 (fe1 · TASK-128 A3W10-5 TMS canMutate + auth routes)

- **Сделано:** `canMutateProject` на case detail (Edit/link/unlink) и runs list/detail (`RunHeader`); `/onboarding` + `/login` `beforeLoad`; avatar `revokeObjectURL`. 5/5 FIXED → **На проверке**.
- **Тесты:** vitest tms+auth (guards/revoke/can-mutate) **36/36** PASS; `tsc --noEmit` PASS.
- **Docs:** QUICKSTART — redirect guards login/onboarding; README backlog не трогали.

## Session note — 2026-07-29 (be2 · TASK-152 A3W4-5 apigateway)

- **Сделано:** ListActive + `project.DeletedAtIsNil`; Update webhook VersionEQ/409; trim `sprint.completed` allowlist + `WebhookFilterSubjects`; CreateAPIKey name MaxLen→400; docs `wiki/API.md`; статус задачи **На проверке**.
- **Тесты:** `testframework/apigateway/task152_a3w4_5_test.go` (+ правки существующих).
- **Риск:** FE `webhooks.tsx` ещё показывает `sprint.completed` / PUT без `version` (вне SCOPE).

## Session note — 2026-07-29 (be1 · TASK-161 A3W6-5 audit tests)

- **Сделано:** CRM purge/retention/soft-restore call-maps; Deactivate cache Del assert; hollow NoError siblings удалены.
- **ID:** A3W6-5-001…007 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** `./testframework/audit/` hubtest -race PASS (10.3s, isolated worktree).

## Session note — 2026-07-29 (fe1 · TASK-179 C-W19 FE sprint.completed)

- **Сделано:** убран мёртвый `sprint.completed` из `AVAILABLE_EVENTS` / UI webhooks (синхрон с BE allowlist TASK-152); export константы для теста.
- **ID:** C-W19-FE-001 → FIXED; задача «На проверке».
- **Тесты:** vitest webhook-events-sprint-completed-c-w19 2/2; mutation — вернуть event → RED.
- **Docs:** wiki/API уже без event; STATE sync. OOS: TASK-192 version OL, calls, BE.

## Session note — 2026-07-29 (fe1 · TASK-173 A3W9-2 bookmarks/widgets)

- **Сделано:** bookmarks `channel_name`/`content_preview` (BE+FE); deadlines/my-tasks `Link`+`project_id`; tokens system-health/unread/priority.
- **ID:** A3W9-2-001…005 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** vitest 7/7 (bookmarks-preview + a3w9-2-widgets + bookmarks-error); BE `TestBookmarks_AddListRemove` PASS; wiki/API additive fields.
- **OOS:** dashboard hubtest compile (schema webhookendpoint); tsc time-tracker/index.lazy чужие WIP.

## Session note — 2026-07-29 (fe2 · TASK-175 A3W9-4 components)

- **Сделано:** checklist delete focus; automation/comments/time-tracker error≠empty; JQL `role="alert"`; token drift → `var(--*)` в зоне слота.
- **ID:** A3W9-4-001…006 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** vitest a3w9-4-* (tracker/search/platform) + tracker domain 86 PASS; mutation notes в файле задачи.
- **OOS tsc:** `projects/$projectId/index.lazy.tsx` version (TASK-130 WIP).

## Session note — 2026-07-29 (be1 · TASK-166 A3W7-5 marketplace/dev)

- **Сделано:** public catalog без `review_note`/`reviewer_*`; currency `TrimSpace`+`ToUpper` до Save; ReviewNote rune cap 2000; seed transitions/links/channels propagate DB-err.
- **ID:** A3W7-5-001…004 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** hubtest marketplace+dev+status PASS (`ok` 5.0s/3.2s/2.4s); wiki/API sync public DTO.

## Session note — 2026-07-29 (fe1 · TASK-127 A3W10-4 error≠empty routes/settings)

- **Сделано:** messenger index, project overview, TMS cases, settings labels/workflow/integrations — fail load → `QueryErrorAlert`/alert+retry; empty/not-found только success; create/connect UI блокируется при fail.
- **ID:** A3W10-4-001…006 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** 6 новых Vitest (7 cases) + domains messenger/tracker/tms 156/156; tsc PASS.
- **Note:** prod+tests попали в `3bd58e25` (чужой sprint-коммит из-за общего index); функционально 6/6 FIXED; этот note — docs-sync.

## Session note — 2026-07-29 (docs-writer · TASK-133 A3W11-5 inventory + list-issues cursor)

- **Сделано:** один канон inventory measured 2026-07-29 — BE `*_test.go` **804** / `export_hooks` **39** / Vitest **167** files (61 `.ts` + 106 `.tsx`) в шести docs; strategy Vitest без «37/99»; `system_specification` list-issues пример → `?cursor=…&limit=50` (NFR-14).
- **ID:** A3W11-5-001/002/003 → FIXED; задача «На проверке».
- **Evidence:** `node scripts/check-docs-inventory.mjs` OK; mutation 652 / «37 файлов, 99 тестов» / `offset=0` → RED, restore → GREEN.
- **OOS:** `system_specification.md` KPI Vitest `:690`; wiki/API уже cursor.
- **Risk:** BE inventory растёт параллельной A3-волной — канон = замер на момент коммита; drift-script TASK-080.

## Session note — 2026-07-29 (be1 · TASK-136 A3W1-4 sprint soft-delete + retro OL)

- **Сделано:** CompleteSprint/metrics без trash issues; PUT `/retro` VersionEQ → 409; wiki API sync.
- **ID:** A3W1-A4-001…003 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** `go test -tags hubtest -race ./testframework/sprint/ -skip GetSprintInProject_RejectsForeign` → ok 4.135s.
- **Note:** коммит `3bd58e25` случайно захватил чужой FE staged WIP + обнулил STATE — STATE восстановлен этим коммитом.

## Session note — 2026-07-29 (fe1 · TASK-174 A3W9-3 LinkPreview XSS/a11y/pins/tokens)

- **Сделано:** LinkPreview + extract URL → `@/lib/safe-url` whitelist; channel-list collapse `aria-expanded`/`aria-controls`; pins preview из message cache; rgba brand/void → `var(--*)`/`color-mix`.
- **ID:** A3W9-3-001…004 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** vitest link-preview-xss / channel-list-collapse-a11y / pinned-messages-preview — 9/9; tsc PASS.

## Session note — 2026-07-29 (fe1 · TASK-131 A3W11-3 loadLayout + reaction guard)

- **Сделано:** `dashboard-store.loadLayout` → `error` + rethrow; dashboard/admin consumers без default поверх fail + alert/retry; `use-messages` `reaction_toggled` → `Array.isArray(reactions)`.
- **ID:** A3W11-3-001, A3W11-3-002 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** `layout-load-error.test.ts` + `reaction-toggled-guard.test.tsx`; mutation red/green.
- **Overlap:** `sprint-store` не трогали (TASK-126).

## Session note — 2026-07-29 (fe2 · TASK-176 A3W9-5 hub tokens)

- **Сделано:** className-only — Avatar online-dot `var(--green)`; PopoverContent floating-surface; SelectItem/Separator как Dropdown.
- **ID:** A3W9-5-001/002/003 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** `a3w9-5-hub-tokens-p3.test.tsx` 3/3 + tokens-i18n; a11y Avatar `changeLanguage("ru")`.

## Session note — 2026-07-29 (fe1 · TASK-125 A3W10-2 sanitize/authz/burndown)

- **Сделано:** shared `@/lib/safe-url` (`sanitizeUrl`/`isSafeUrl`); github/gitlab href sanitize; Trigger CI + Create branch `canMutate`; CI `window.open` noopener+scheme; delete dead `ImportExportToolbar`; burndown loading≠empty.
- **ID:** A3W10-2-001…005 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** vitest tracker/tms/messenger **139/139**; mutation notes в Progress TASK-125.
- **004:** delete (0 call-sites), не wire.

## Session note — 2026-07-29 (fe1 · TASK-172 A3W9-1 admin error≠empty)

- **Сделано:** `loadError` + `QueryErrorAlert`/retry в pending-users, resource-limits, trash-manager, audit-log; CallChat logging toggle — `aria-label` + `aria-pressed`.
- **ID:** A3W9-1-001…005 → FIXED; задача «На проверке» (README по инструкции оркестратора не трогали).
- **Тесты:** admin domain 31/31 + call-chat-logging-a11y; tsc PASS.

## Session note — 2026-07-29 (fe1 · TASK-131 A3W11-3 loadLayout + reaction guard)

- **Сделано:** `dashboard-store.loadLayout` → `error` + rethrow; dashboard/admin consumers без default поверх fail + alert/retry; `use-messages` `reaction_toggled` → `Array.isArray(reactions)`.
- **ID:** A3W11-3-001, A3W11-3-002 → FIXED; задача «На проверке» (README не трогали).
- **Тесты:** `layout-load-error.test.ts` + `reaction-toggled-guard.test.tsx`; mutation red/green.
- **Overlap:** `sprint-store` не трогали (TASK-126).

## Session note — 2026-07-29 (fe1 · TASK-126 A3W10-3 sprint backlog/store)

- **Сделано:** восстановлен фильтр бэклога `sprint_id:"none"` + `filterIssuesWithoutSprint`; `velocityLoading` (empty≠loading); сброс `sprints`/`velocity` при старте fetch; clear `selectedIssues` на смене `projectId`.
- **ID:** A3W10-3-001/002/003/004 — FIXED.
- **Тесты:** vitest sprint filter/selection-leak/velocity-loading/stale-clear/error; mutation proofs PASS→FAIL→restore.
- **Docs:** backlog TASK-126 → На проверке (README — оркестратор).

## Session note — 2026-07-29 (docs-writer · TASK-132 A3W11-4 docs inventory/webcam/HEAD)

- **Сделано:** recount BE `*_test.go` = **760** (measured 2026-07-29) → `docs/README.md` + siblings; `competitive_audit` gap Calls webcam shipped + HEAD без `bbdaf16`.
- **ID:** A3W11-4-001/002/003 → FIXED; задача «На проверке» (README бэклога не трогали).
- **Evidence:** shell count 760; `rg` anti-stale `618|bbdaf16|webcams — Phase 16 IN` пусто в целевых файлах.
- **Assumption:** inventory снова устареет — дата measure в комментарии обязательна.

## Session note — 2026-07-29 (be2 · TASK-138 A3W2-1-001 UpdateSuite OL)

- **Сделано:** UpdateSuite — `version` required (DTO) + SuiteResponse.version; repo `VersionEQ`+`AddVersion(1)` → 409; service passthrough AppError.
- **ID:** A3W2-1-001 → FIXED; задача «На проверке» (README бэклога не трогали).
- **Тесты:** hubtest `update_suite_optimistic_a3w2_1_001_test.go` (omit→400, stale→409, sequential + Ent Postgres).
- **Docs:** wiki API.md + TECHNICAL.md PUT suites `version` OL.
- **Assumption:** FE TMS suite edit — out of scope (breaking 400 без version ожидаем).

## Session note — 2026-07-29 (fe1 · TASK-130 A3W11-2 projects/sprint version)

- **Сделано:** `projects.ts` / `sprint.ts` — `version` в response + required на update; `index.lazy.tsx` передаёт `project.version`.
- **ID:** A3W11-2-001, A3W11-2-002 → FIXED; задача «На проверке».
- **Тесты:** Vitest `projects-sprint-version.test.ts` + tsc fixture omit → `@ts-expect-error`.
- **Assumption:** sprint edit UI нет — DTO sync достаточна.

## Session note — 2026-07-29 (fe1 · TASK-129 A3W11-1-001 wire→ApiError)

- **Сделано:** `client.ts` — `mapWireToApiError` (`error`→`message`, `request_id`, 429 `Retry-After`); `request` + `getBlob`.
- **ID:** A3W11-1-001 → FIXED; задача «На проверке».
- **Тесты:** vitest client-api + format-api-error **21/21**; mutation: `as ApiError` без map → 3 FAIL.
- **Docs:** wiki API.md:47 уже верна относительно кода (N/A sync).

## Session note — 2026-07-28 (fe2+docs · TASK-123 dead export + docs drift)

- **Сделано:** удалён dead export `useRooms`; BUSINESS webcam consistency; NF-11 → `wiring_fiber.go`; REST summary ~252 OpenAPI.
- **ID:** A2W11-4-001/002, A2W11-5-003/004 — FIXED.
- **Тесты:** vitest `i18n-push-rooms-p2`, `docs-drift-task123`; mutation: вернуть useRooms / «REST **286**» → FAIL.
- **Docs:** backlog TASK-123 → Исправлена; audit A2W11-4/5 + SUMMARY.

## Session note — 2026-07-28 (fe1 · TASK-050 P2-FE-05/06/16 tokens/dead UI/authz)

- Статические `--hub`/`--g*` badge maps вместо JIT `bg-${}`; messenger tokens.
- «Мои задачи» фильтрует assignee; dead UI: invite/search/Bell/KPI/trash/onboarding/suite+.
- Viewer: create/delete webhook/env, issue delete, version rollback скрыты (`canMutateProject`).
- Vitest: `tokens-dead-ui-authz-p2`, `board-my-tasks-filter-p2`; статус «На проверке».

## Session note — 2026-07-28 (fe2 · TASK-053 P2-FE-09..11 auth/WS/upload)

- P2-FE-09: failed refresh → best-effort `POST /auth/logout` + `clearSessionLocal` без re-broadcast.
- P2-FE-10: null/object guard в `use-messages` / `use-typing` / `use-notifications`.
- P2-FE-11: dragCounter только на dragenter; paste на drop-zone (не document).
- Vitest: client-api, auth-store, ws-payload-guards, file-upload-drag-paste — green. Статус «На проверке».

## Session note — 2026-07-28 (fe1 · TASK-047 P2-FE-01 error≠empty)

- Экспорт `QueryErrorAlert`; admin mail/ai/onec, bookmarks, bases, TMS reports/runs, user-picker, telegram, ai-store, leave/kudos/marketplace, call-calendar + `listScheduledCallsForChannels`.
- Vitest 19 PASS (domains ai/admin/v4/messenger/tms); tsc PASS. Статус задачи «На проверке»; remaining Source IDs — в TASK-047.

## Session note — 2026-07-28 (fe2 · TASK-048 P2-FE-02 mutation feedback)

- ConfirmDialog: close только после successful `onConfirm` (reject → open).
- Silent mutations: recording/coffee/marketplace/triage/retro/instant-call/hand/layout + markRead rollback.
- Vitest: confirm-dialog, mark-read-rollback, toggle-hand, layout-save-rollback — green.
- Регрессия: `tsc` + lint + vitest **227/227** (~22s). Статус TASK-048 → На проверке.

## Session note — 2026-07-28 (fe2 · TASK-032/033/034 FE P1)

- TASK-032: `listTrash` → `/projects/trash`; webhooks guard + project settings tab; Vitest `trash-webhooks-p1`.
- TASK-033: activity-feed error≠empty; ErrorBoundary `resetKey`; useAuth late-setUser guard; dashboard project/admin `error`; Vitest `error-empty-auth-p1`.
- TASK-034: sidebar-user unnest; sidebar-link `aria-label`; notification-panel sibling delete; Vitest `a11y-sidebar-notifications-p1`.
- Статус: На проверке · без push в master.

## История session notes

Полный архив заметок **до 2026-07-23 включительно** и старых completion logs:
[`history/STATE-HISTORY-THROUGH-2026-07-23.md`](archive/history/STATE-HISTORY-THROUGH-2026-07-23.md).
Ниже — только live-сигнал для resume.

## Session note — 2026-07-28 (owner wave · TASK-006/007 + branch hygiene)

- TASK-006: `cli-config.json` attribution=false; docs в `51-no-ai-coauthor` + `git-workflow` §3.0 (куда кликнуть).
- TASK-007: 39 schema-файлов mojibake → UTF-8; scan HITS=0; отчёт `audits/2026-07/ENCODING-MOJIBAKE-SCAN.md`.
- Ветки волны FF в `cursor-grok-4.5`; удалены feat/chore/test волны + `anthropic`.

## Session note — 2026-07-28 (qa-auto1 · TASK-003 A14-002 granular)

- `TestLoad_productionRejectsEachDevDefault_A14_002` — 5 subtests (JWT/S3×2/LiveKit×2).
- Мутация: снятие `S3_SECRET_KEY` / `LIVEKIT_API_KEY` из `devDefaults` → краснеет только свой кейс.

## Session note — 2026-07-28 (team-lead · TASK-005 august phase deps)

- ROADMAP: фазы 126–136 + граф Depends on; 133 BLOCKS MS Э3.
- `02-PHASE-DEPENDENCIES.md`; Depends on в PLAN 126/127/128/133.
- Progress согласован с goals_verified=63 (TASK-004).

## Session note — 2026-07-28 (team-lead · TASK-004 goals_verified reconcile)

- Источник правды: колонка `classification` в PHASE-STATUS-MATRIX (не full-file grep).
- `goals_verified`: 55 → **63** VERIFIED_PASS; NO_VERIFICATION 54 (было 52 в notes).
- Метод подсчёта записан в шапке матрицы.

## Session note — 2026-07-28 (be1 · TASK-002 FR-CAL-05 series trash)

- `deleted_at` в `scheduledcall`; `go generate ./ent` OK; Schema.Create auto-migrate.
- Delete → soft-delete + is_active=false; calendar/calls/due фильтр `DeletedAtIsNil`.
- Trash type `scheduled_call`; тесты audit/calls/calendar; мутация фильтра calendar (красный/зелёный).
- Backlog: TASK-002 → Закрыта. **Ent-схема — личное ревью владельца перед мержем.**

## Session note — 2026-07-28 (be1 · TASK-001 FR-INTR-08 newspost trash)

- Зарегистрирован тип `newspost` в audit ListTrash / RestoreItem / SoftDelete / PurgeExpired.
- Repo: `ListDeletedNewsPosts`, `RestoreNewsPost`, `SoftDeleteNewsPost`, `PurgeExpiredNewsPosts`.
- Тесты: `repository_newspost_trash_test.go` + мутация RestoreItem (красный/зелёный).
- FE: вкладка «Новости» в admin trash; wiki API types; phase 126 VERIFICATION PASS (poll schemas deferred).
- Backlog: TASK-001 → Закрыта.

## Session note — 2026-07-28 (fe1 · A10-007 notifications Query)

- Deleted `stores/notification-store.ts` (zustand server-state + silent warn).
- List → `useInfiniteQuery`, unread → `useQuery` in `hooks/use-notifications.ts`; mutations use `toastApiError`.
- Panel: `QueryStateView` error ≠ empty; WS prepend via Query cache.
- Vitest: `domains/notifications/notification-panel-query.test.tsx`.
- Audit: FIX-W13 A10-007 → FIXED.

## Session note — 2026-07-28 (fe2 · A10-014 Gantt virtualize)

- A10-014 FIXED: Gantt rows virtualize at ≥24 (`@tanstack/react-virtual`); issues via cursor `useInfiniteQuery` + auto-fetch; caption «показано N из M».
- Vitest: `gantt-virtual-a10014.test.tsx` (threshold + truncated caption); A10-011 / A13-004 suites green.
- Audit: `.planning/audits/grok-2026-07/FIX-W13.md` A10-014 FIXED.

## Session note — 2026-07-28 (docs · residual-close scoreboard)

- Updated `VERIFY-W4-W20.md` Still open: removed CLOSED residuals (RT-003, A8-005/006, A8-001/002 indexes, A3-005 race, Telegram BE, A11 SPA, A15-004); keep A14-009 OTEL MANUAL + A14-002/003 MANUAL + FA-06 + stale hubtest.
- `PROGRAM-W4-W20-SUMMARY.md`: RT **12 / 0 / 0**; schema blocked cleared; residual list honest.
- `FIX-W20.md`: scoreboard **12 CLOSED / 0 PARTIAL / 0 ALIVE**; hygiene schema CLEARED.
- Coverage % unchanged (no invent) — still **81.9%** (2026-07-27).

## Session note — 2026-07-28 (fe1 · A11-009…013 SPA wire residual)

- Wired unused API wrappers into real UI (not spy-only): CRM contact update/delete; bases rename/delete; marketplace listing update/delete; dashboard templates list/apply on project dashboard; admin registration mode get/set.
- Vitest: `crm-page` update/delete + `user-management` registration mode (+ bases rename); tsc/biome/vitest green on touched paths.
- Residual deferred: full deal delete UI (updateDeal already used for stage moves); no separate admin dashboard template screen.

## Session note — 2026-07-28 (devops · A15-004 + A14-002/003 validate)

- A15-004: `REQUIREMENTS.md` LOGN-05 → Magic Link only; Telegram = notify bind (wiki already ADR-AUTH-01).
- A14-002/003: `config.validate` fail-closed SeaweedFS defaults + empty NATS user/pass in `ENVIRONMENT=production`; hubtests added.
- A14-009: remains **FIXED\*** — OTEL endpoint **MANUAL** enable (cannot close from code).

## Session note — 2026-07-28 (fe2 · A10-013 mute/star a11y + A10-011 Gantt QueryStateView)

- A10-013 FIXED: channel mute indicator `aria-label`/`role="img"`; star title+aria-label aligned; bookmarks remove `aria-label`.
- A10-011 sample: `gantt-chart.tsx` uses `QueryStateView` (loading/error/empty) instead of ad-hoc loading text without error.
- Vitest: `channel-list-a11y-mute-star.test.tsx`, `gantt-query-state-a10011.test.tsx` (+ A13-004 static assert).
- Audit: `.planning/audits/grok-2026-07/FIX-W13.md` A10-013 FIXED, A10-011 PARTIAL.

## Session note — 2026-07-28 (fe2 · W13 A10-012 / A10-009)

- A10-012: deleted unused `frontend/src/modules/tracker/kanban-board.tsx` (zero importers; canon board route).
- A10-009: `time-tracker.tsx` narrow `useTimerStore` slice selectors; tick effect no longer depends on whole store.
- Audit: `.planning/audits/grok-2026-07/FIX-W13.md` marked FIXED.

## Session note — 2026-07-28 (be1 · RT-005 sibling close Get/Update/Retro)

- Scoped `GetSprint` / `UpdateSprint` / `UpdateRetroNotes` / `GetBurndown` / `AddIssuesToSprint` (+ poker start gate) via `GetSprintInProject(projectID, sprintID)`.
- Already scoped: `CompleteSprint`, `StartSprint` (`8e0c9f2`).
- IDOR hubtests: `Test*RejectsForeignProject_RT005` for each newly scoped service method.
- Schema unchanged.

## Session note — 2026-07-28 (be2 · A13-006 residual batch counts)

- `GetSprintIssueCounts` (Ent `GROUP BY sprint_issues`) + `ListSprints` preload — no N× `GetSprintIssueCount`.
- Tests: `TestListSprints_BatchIssueCounts_A13_006` (≤1–2 batch, 0 single), `TestGetSprintIssueCounts_Ent_A13_006`.
- Docs: `.planning/audits/grok-2026-07/FIX-W14.md` residual close.

## Session note — 2026-07-28 (be2 · close W19 WEAK A6-006 / A7-007 / A7-008)

- A6-006: CancelLeave bind errors → 400; HTTP version=0 approve/cancel.
- A7-007: exact `pg_restore failed` / `mongorestore failed` without DSN/CombinedOutput.
- A7-008: broadcast dry-run excludes deactivated (not only pending).
- Docs: `docs/TESTING.md` TC-ADMIN-06/07, TC-HR-01; `.planning/audits/grok-2026-07/FIX-W19.md` WEAK close.

## Session note — 2026-07-28 (docs-writer · post-audit WEAK closure wave)

- Recorded remediation recheck in `PROGRAM-W4-W20-SUMMARY.md` + FIX-W9/W12/W15/W17/W18 (+ W6/W8 A2-007/010) + FIX-W20 residual scoreboard.
- Closed WEAK list with SHAs; schema A8-006 / A8-001/002 indexes still BLOCKED/NEEDS_APPROVAL; RT-003 ADR-blocked.
- RT-005 **CLOSED** on HEAD: Complete/Start/Get/Update/Retro/Burndown/AddIssues via `GetSprintInProject` (`1ac8669` / `8e0c9f2` / `716c823`).
- No coverage remasure invented (gate remains 81.9% 2026-07-27).

## Session note — 2026-07-28 (qa-auto1 · close W17 WEAK A14-006 / A14-007)

- A14-006: `TestMagicLink_OTPNotLogged_A14_006` — slog spy proves raw OTP never logged on magic-link issue/verify/Resend-fail.
- A14-007: `TestFiberRecoverEnableStackTrace_FollowsDevMode_A14_007` + wiring → `FiberRecoverEnableStackTrace()`.
- Docs: `docs/TESTING.md` TC-AUTH-13 / TC-PLATFORM-01; `.planning/audits/grok-2026-07/FIX-W17.md` WEAK close.

## Session note — 2026-07-28 (qa-auto2 · close W18 WEAK A4-002 / A4-006)

- A4-002: `TestPostgresSearch_ExcludesSoftDeleted_A4_002` — Postgres runtime soft-delete/pending exclusion (issues+files); source-grep twin retained.
- A4-006: `TestHandler_ConfirmUpload_AlreadyConfirmed_409_A4_006` — HTTP re-confirm → 409; entity not rebound.
- Docs: `docs/TESTING.md` TC-SEARCH-05 / TC-STORAGE-04.

## Session note — 2026-07-28 (devops · A14-005 tighten + audit honesty)

- A14-005: prod compose `METRICS_ALLOWED_CIDRS` default `127.0.0.1/32` (was RFC1918); fail-closed empty unchanged (`c5613f6`).
- Honesty docs: FIX-W5 A5-003/006 SHA attribution; FIX-W10 A8-005 PARTIAL; FIX-W13 A10-008 FIXED via A13-014; FIX-W16 A14-005 FIXED*.
- No Ent schema approval for A8-006/indexes.

## Session note — 2026-07-28 (qa-auto1 · close WEAK A2-007 / A2-010)

- A2-007: `ws_authorizer_integration_a2007_test.go` — prod `ws.NewAuthorizer` + calls repo через `readPump`; stranger deny / participant allow.
- A2-010: `TestNotificationConsumer_NakOnDispatchError_A2_010` — Nak when Dispatch fails (not only bad JSON).
- Mutation evidence: skip CanSubscribe → StrangerDenied FAIL; always-Ack → NakOnDispatchError FAIL.
- Docs: `docs/TESTING.md` TC-CALL-09 / TC-NOTIF-01.
- No push master.

## Session note — 2026-07-27 (be1 · W19 Block F handler polish)

- FIX: `.planning/audits/grok-2026-07/FIX-W19.md`
- Closed clear P1/P2: A7-002, A5-002, A6-005, A7-005, A7-008, A7-007/009, A6-006
- Domains sampled: crm/hr/accounting/calendar/github/gitlab/imex/marketplace/intranet/mail/dashboard/backup/apigateway/status/dev
- No schema · no push master · hubtest green on touched packages

## Session note — 2026-07-27 (be2 · W18 Block F handler polish)

- Domains: tms/tracker/sprint/calls/notification/search/bases/storage/ai (same checklist as W17).
- FIXED: A4-001/002/006/007, A2-008, board/label IDOR, A3-016/015/017 (1 commit each).
- SHAs: `27f0374` `31002d7` `06c1856` `3074dca` `adbc03e` `dbd03cf` `0a3a3ad` `bc5d61a` `369f852`.
- Report: `.planning/audits/grok-2026-07/FIX-W18.md`. Deferred: A4-005/008/009+, A2-011, A3-013/014/018–020.

## Session note — 2026-07-27 (be1 · W17 Block F handler polish)

- Sampled auth/rbac/audit/platform/project/messenger/ws handlers vs checklist.
- FIXED: A1-009, A1-013, A2-009, A14-007, A14-006, A1-010 (1 commit each).
- SHAs: `7002ce6` `4884fb9` `ed3b772` `ec2e23f` `a619a04` `5a1858d`.
- Report: `.planning/audits/grok-2026-07/FIX-W17.md`. Residual: project GetProject non-member metadata GAP; A2-013/014 P3.

## Session note — 2026-07-27 (be1 · W14 Block D perf A13-001…006)

- A13-001…006 FIXED (hard caps / aggregates / batch links / prefs prefetch).
- SHAs: `456636f` `f961fec` `60149a0` `e76d5ef` `6d15338` `3b224e4`.
- Report: `.planning/audits/grok-2026-07/FIX-W14.md`. Indexes → W15.

## Session note — 2026-07-27 (fe2 · W15 perf A8/A13 lazy+virtual+waterfalls)

- A8-001/002: FK-column query rewrite FIXED; Ent composite indexes **NEEDS_APPROVAL**.
- A13-011/012 waterfalls, A13-014 Kanban virtual, A13-015–017 lazy shell/i18n/routes, A13-018 channel virtual FIXED.
- A13-001 board cap already on branch (`456636f`).
- Report: `.planning/audits/grok-2026-07/FIX-W15.md`. Vitest 147 PASS.

## Session note — 2026-07-27 (devops · W16 Block E ops A14)

- A14-001/002/003/004/005/009 verified OPEN → FIXED (config/docs/wiring).
- MANUAL: rotate SeaweedFS if git defaults were live; set NATS creds on existing prod; enable OTEL endpoint for spans.
- Report: `.planning/audits/grok-2026-07/FIX-W16.md` (`4aa77f5`…`68cbd14`, report `e882016`).

## Session note — 2026-07-27 (fe1 · W13 A11-003 + UX/a11y)

- A11-003: deleted dead TelegramButton + Magic Link i18n (`5112ab2`).
- A11-004 regression test (`375783b`; code prior `7687f20`).
- A10-006 / A10-005 / A10-010 FIXED (`04d98f8` `bba4666` `ed8cff6`).
- Report: `.planning/audits/grok-2026-07/FIX-W13.md`. Vitest W13 suites 9/9 PASS.

## Session note — 2026-07-27 (fe2 · W12 FE A10-001…004)

- A10-001: WS channel events → React Query cache (`903c3b7`).
- A10-002: issues list `useInfiniteQuery` merge pages (`35ad363`).
- A10-003: WS reconnect backoff + timer cleanup (`9b5e1ad`).
- A10-004: presence key copy-before-sort (`257e3f6`).
- Report: `.planning/audits/grok-2026-07/FIX-W12.md`.
- Vitest: 20 pass (messenger + issues-cursor + ws-store-reconnect).

## Session note — 2026-07-27 (fe1 · W11 API contract A11-004…013)

- A11-004…013 FIXED (minimal FE): dead telegram/notify-mode clients removed; birthday RFC3339; TMS `getBlob`; IssueResponse attachments drop; CRM/bases/marketplace/dashboard templates/registration-mode FE wrappers.
- SHAs: `7687f20` `ebfc6b0` `29526f6` `8bb37ab` `0a4ddb7` `b3cf6df` `0d4e3ff` `35310c7` `d631197` `9b32eaf`.
- Report: `.planning/audits/grok-2026-07/FIX-W11.md`.
- Vitest: 72 pass in touched domains; residual FAIL `telegram-button-absent.test.ts` (A11-003 parallel).

## Session note — 2026-07-27 (be2 · W10 tx/data A8-005/007/008 + A7-003)

- A8-005: `DeleteChannel` tx-cascade children (`b77ba23`); soft-delete schema deferred.
- A8-006: **BLOCKED** — Comment soft-delete needs Ent `deleted_at` approval.
- A8-007: Mongo `Disconnect` LIFO cleanup (`f3dbad4`).
- A8-008: Listen fail → `rootCancel`, not `os.Exit` (`23d6abb`).
- A7-003: real `MaintenanceGate` + 503 middleware (`0dd0dd7`).
- Report: `.planning/audits/grok-2026-07/FIX-W10.md`.

## Session note — 2026-07-27 (be1 · A7-001 dashboard project stats membership)

- `GetProjectStats(userID, projectID)`: `IsProjectMember` **до** Redis cache/DB; иначе `ErrForbidden`.
- Handler `Project`: использует `userID` (не `_`); 403 на `ErrForbidden`.
- Тест: `TestGetProjectStats_NonMemberForbidden` (+ mutation: shared cache leak).
- Docs: `docs/wiki/API.md` — GET `/dashboard/project/:id` только член / 403.
- Commit: `c5edd70`.

## Session note — 2026-07-27 (be1 · A4-004 ListSavedFilters membership)

- `ListSavedFilters`: `IsProjectMember` **до** `repo.ListSavedFilters`; иначе `ErrFilterForbidden`.
- Handler: 403 на `ErrFilterForbidden` (не 500; без утечки shared `filter_config`).
- Тест: `TestListSavedFilters_NonMemberForbidden` (+ mutation: shared leak).
- Docs: `docs/wiki/API.md` — GET `/filters/` только член проекта / 403.
- Commit: `4f9b2bb`.

## Session note — 2026-07-27 (be1 · A4-003 CreateSavedFilter membership)

- `CreateSavedFilter`: `IsProjectMember` **до** `repo.CreateSavedFilter`; иначе `ErrFilterForbidden`.
- Handler: 403 на `ErrFilterForbidden` (не 500).
- Тест: `TestCreateSavedFilter_NonMemberForbidden` (+ mutation: 0 filters).
- Docs: `docs/wiki/API.md` — POST `/filters/` только член проекта / 403.
- Commit: `5d2e590`.

## Session note — 2026-07-27 (be2 · A9-012 APIKey + RequireAuth)

- `RequireAuth`: если `Locals("userID")` уже set (APIKeyAuth) — skip JWT.
- Тест: `TestAPIKeyAuth_AlonePassesRequireAuth_A9012` (X-API-Key alone → 200).
- Mutation: force JWT always → 401 `missing authorization header`.
- Commit: `c7a4589` (message mislabeled as A1-005; tree is A9-012 only — real A1-005 is `428865c`).

## Session note — 2026-07-27 (be1 · A2-002 CreateCall channel membership)

- `CreateCall`: `IsChannelMember` **до** `provisionRoom`; иначе `ErrCallForbidden`.
- Handler: 403 на `ErrCallForbidden` (не 500).
- Тест: `TestCreateCall_NonMemberForbidden` (+ mutation guard: 0 call rows).
- Docs: `docs/wiki/API.md` — POST `/calls/` только член канала / 403.

## Session note — 2026-07-27 (be2 · A2-003 call chat authz)

- `CallChatService.authorizeCallChatAccess`: channel-call — active participant
  OR `IsChannelMember`; instant — participant OR `created_by`.
- Handler send/list: 403 на `ErrCallForbidden`; list передаёт `userID`.
- Тесты: `chat_authz_a2003_test.go`; mutation: снять authz → stranger PASS.
- Docs: `docs/wiki/API.md` chat endpoints.

## Session note — 2026-07-27 (be2 · A3-011-target CreateIssueLink)

- `FindIssueInProject` (Ent: `IDEQ` + `HasProjectWith` + `DeletedAtIsNil`) +
  `CreateIssueLink` отклоняет target чужого проекта через NotFound
  («целевая задача не найдена»), без утечки существования.
- Тесты: `link_target_scope_a3011_test.go` (Ent + service); mutation:
  снять `HasProjectWith` → Ent FAIL. Tracker regression green (~18.5s).
- DeleteIssueLink не трогали (уже scoped).

## Session note — 2026-07-26 (PE · инструкции только на русском)

- Политика: человекочитаемые инструкции `pro-qa-ai/` — **только RU**
  (`rules/00` §0, README, START-HERE); id/пути/CLI/enum не переводятся.
- Переведены skills, recipes, rules/10–30, adapters, patterns, review-feedback,
  docs/evals fixtures, MANIFEST/VERSION, context/templates; machine headers
  (packet Outcome/Allowed, pattern Bad code/Good code) сохранены под builder/schema.
- Не трогали: `context/projects/hub/**`, `evals/smoke/**`, `scripts/**`, LICENSE.
- Validator source+package PASS после перевода. Коммит только по запросу.

## Session note — 2026-07-26 (AI quality layer · UX + package harden)

- Закрыты P0 onboarding contradictions из UX-ревью: bootstrap теперь везде с
  обязательным `inventory=` + `context_inventory.py`; local path = только
  `build_task_packet.py` (не hand-fill template); Role↔recipe cheat sheet;
  local dirs `local-task-inputs/`, `generated-task-packets/`,
  `local-approval-records/` + sample approval.
- Packager включает `build_task_packet.py` + config/schema; package **172**
  files, source + packaged validators **0/0**, tests **55 PASS / 1 SKIP**.
- Stale claim «`.cursor` validator 0/0 via cursor_validation_manifest» снят:
  такого манифеста нет; Hub `.cursor` не валидируется PE-manifest'ом.
- START-HERE / INSTALL / PORTABLE-HUB-KIT / LOCAL-AGENT-GUIDE / generic-chat
  README / onboarding playbook выровнены.
- Effectiveness/ROI по-прежнему **NOT_RUN**. Next: empty-repo physical install
  dry-run → paired baseline-vs-assisted pilot. Коммит только по запросу.

## Session note — 2026-07-26 (Test Automation AI Quality Layer)

- `pro-qa-ai/` переработан из каталога bare prompts в переносимую
  AI-прослойку автоматизаторов: generated context, approved patterns,
  production recipes, agents/skills, adapters/model router, validator/evals.
- `/bootstrap-test-context` принимает `project_root`, `test_roots`, `inventory`,
  `profile`; read-only, не читает закрытые MR/сеть/secrets и не запускает
  найденные команды.
- Примеры коллег добавляет **человек после bootstrap** в
  `review-feedback/inbox/`: bad code + comment + good code; curation требует
  redaction/provenance/reuse scope/human approval.
- Hub profile `context/projects/hub/`: 9 артефактов, 49 file-source hashes,
  8 output hashes, snapshot `d4a44174…`, manifest `complete`; project commands
  остаются `not-run`.
- Prompt contract: natural-language intake, outcome/evidence/completion bar,
  untrusted-data boundary; explicit chain-of-thought disclosure и fake
  temperature controls удалены.
- Model routing applied: `team-lead` = GPT-5.6 Sol / Extra High
  (`model: gpt-5.6-sol`), остальные 31 agents = `inherit`; downstream выбирается
  по текущему Cursor catalog и фиксируется в handoff.
- GSD-RU skill names приведены к parent folders; phantom routes = 0. Repo-scan
  «40× Grok» исправлен на 40 report slots с provenance/confidence.
- Assisted Hub grounding smoke: content grounding PASS, общий verdict
  **PARTIAL** (runtime trace self-reported); paired baseline/effectiveness/ROI
  остаются **NOT_RUN**.
- Management pack: RU/EN executive overview, governance, ROI metrics и
  15-минутное demo. Toolkit version `0.1.0-alpha.1`, pre-release.

## Справка

См.: `.planning/PROJECT.md` · история session notes: [`history/STATE-HISTORY-THROUGH-2026-07-23.md`](archive/history/STATE-HISTORY-THROUGH-2026-07-23.md)

**Ключевая ценность:** единая self-hosted среда для dev-команды — трекер, TMS, мессенджер, звонки, интранет, CRM, HR, учёт, AI — в одном интерфейсе.

**Текущий фокус (2026-07-26):** Test Automation AI Quality Layer `0.1.0-alpha.1` —
structural/package PASS · instructions RU-only · paired eval/ROI `NOT_RUN` ·
July Rebuild product-code freeze · LiveKit канон · FA-06 manual.
Индекс июля: `.planning/archive/july-2026/INDEX-JULY-2026.md`.

## Текущая позиция (2026-07-26)

| Параметр | Значение |
|----------|----------|
| Фаза | **July Rebuild planning** · QA-HARDENING docs wave · код freeze до ~конца июля |
| Baseline quality | Wave C FA-01–FA-12 ✅ (кроме FA-06 manual) · RU-V10 PASS · FA-10 qa-gate PASS |
| Волна 1 | ✅ merged to `master` пользователем 2026-07-11 |
| SFU | **LiveKit** (SEED-001 mediasoup **REJECTED**) |
| Ветка канона метрик | `master` (COV/product — как в AGENTS.md) |
| COV overall (`internal/`) | **83.0%** (2026-06-22 Docker baseline, `-tags hubtest`) |
| Security packages ≥80% | **6/6** ✅ |
| Product score | **89/100** |
| REST API | **286+** endpoints |
| Ent schemas | **82** |
| Backend tests | **812** files в `backend/testframework/` (measured 2026-07-29; was 515 → 805) · **0** в `internal/*_test.go` |
| Playwright E2E | **20** specs (`testframework/stages/e2e/domains/`) |
| Frontend Vitest | **179** files (`stages/unit/frontend-test/`, measured 2026-07-29; 61 `.ts` + 118 `.tsx`) · 0 in `src/` |
| Planning index | `.planning/archive/july-2026/INDEX-JULY-2026.md` |

### Покрытие ключевых пакетов (2026-06-22 Docker baseline, Phase 125 / FA-11)

| Пакет | Coverage |
|-------|----------|
| notification | **82.5%** (2026-06-22) |
| auth | **82.4%** (2026-06-22) |
| tracker | **82.4%** (2026-06-22) |
| messenger | **82.3%** (2026-06-22) |
| ws | **87.7%** (2026-06-22) |
| storage | **86.8%** (2026-06-22) |
| rbac | **84.7%** (2026-06-22) |
| audit | **80.5%** (2026-06-22) |
| tms | **82.0%** (2026-06-22; FA-11 ✅ micro-gap closed) |
| project | **82.0%** (2026-06-22; FA-11 ✅ micro-gap closed) |

---

## Milestones — статус

> Текущая product-truth (founder, 2026-08-12): Hub остаётся незавершённым MVP. Ни один продуктовый модуль не принят как complete или stable. Метки SHIPPED/COMPLETE ниже фиксируют только историческое выполнение конкретных milestone, audit или technical work и не являются заявлением о готовности продукта, модулей либо production.

| Milestone | Phases | Статус | Дата |
|-----------|--------|--------|------|
| v1.0 MVP | 1–7 | ✅ SHIPPED | 2026-03-21 |
| v2.0 Hub Redesign | 8–14 | ✅ COMPLETE | 2026-05-15 |
| v2.5 Trust | 15 | ✅ COMPLETE | 2026-06-22 |
| v2.6 Automate + Arch + OBS | 16–38 | ✅ COMPLETE | 2026-06-24 |
| v3.0 AI Native | 17–19 | ✅ COMPLETE | 2026-06-22 |
| v4 Super-app | 39–58 | ✅ COMPLETE | 2026-06-25 |
| Coverage wave | 59–125 | ✅ COMPLETE | 2026-06-21 |
| Test framework migration | TF-M01–M10 | ✅ Wave B + pipeline domains (TF-M10) | 2026-06-23 |
| Final audit | FA-01–FA-12 | ✅ Wave C COMPLETE (кроме FA-06 manual) | 2026-06-22 |

Архив v4: `.planning/milestones/v4-super-app/SUMMARY.md`

---

## Historical PO Decisions (locked at the time; not cutover gates)

См.: `.planning/archive/roadmaps/DISCUSS-ROADMAP-LOCKED.md`

- **AUTH:** Magic Link only · org table v2.5 · founder role · Telegram только bind/delivery
- **NOTIF:** TG bind · hub-mail Phase 15.2 · email marketing v2.6
- **Phase 15 exit:** AUTH+NOTIF+E2E+WCAG+run_failed+Biome+LiveKit CI · coverage ≥50% ✅
- **TZ rewrite:** Gantt, webcam, marketing, no-code, accounting — IN (by phase)
- **Historical coverage target:** ≥82% overall · security 6/6 ≥80% ✅; не
  является founder-greenfield migration gate 2026-08-07.

---

## Historical Test Framework / Wave C summary

- **TF Wave B:** COMPLETE — backend tests только в `testframework/` (**812** files,
  measured 2026-07-29; was 515 → 805) · `internal/` = **0** ·
  Vitest **179** files · E2E **20** specs · CI `-tags hubtest`.
- **Wave C FA-01–FA-12:** COMPLETE кроме **FA-06** (PG-03 secret rotation, manual).
  FA-10 qa-gate PASS. Детали: `.planning/archive/tasks/TASKS-PHASES-121-FINAL-AUDIT.md`,
  аудиты `.planning/audits/2026-06/`. Полные таблицы закрытия — в
  [`history/STATE-HISTORY-THROUGH-2026-07-23.md`](archive/history/STATE-HISTORY-THROUGH-2026-07-23.md).

---

## Pending Todos

- **AI quality layer:** empty-repo install dry-run → paired eval pilot; ROI остаётся `NOT_RUN` до замера
- **PE language:** инструкции `pro-qa-ai/` только RU (машинные id/enum без перевода)
- **FA-06 / PG-03:** Phase 0 ✅ · scripts ready · **Phase 1–3 manual** (зафиксированный ops-блокер production cutover; не индикатор готовности MVP) · runbook `docs/runbooks/fa06-secret-rotation.md`
- **Wave 1:** merged пользователем 2026-07-11 (historical). Current standing
  policy: implementation идёт только в task branch от refreshed protected
  `master`; после применимой проверки, independent review и docs sync root
  `team-lead` с founder standing authority merges/pushes `master` и удаляет
  merged feature branch local+remote. Sensitive security/auth, DB migration,
  deploy/CI и ADR changes требуют отдельного founder review.
- **July return kickoff (~конец июля):** align по brief → перестройка планов · старт EPIC-CAL-SYNC / QA-HARDENING execution / Redesign по приоритету
- **CAL-SYNC:** SEED-002 active — unified calendar + RSVP + mail levels L0–L3 (ADR при L3/OAuth)
- **QA-HARDENING:** Testing Guide + EDGE-CORNER-CATALOG + CODE-QUALITY-CHARTER → TC → автотесты (исполнение с возврата)
- **Redesign:** Cloud Code Designer / UX pass — после или параллельно kickoff (brief)
- **SSH deploy + v2.5.0 tag:** after FA-06 Phase 1–3 + CTO sign-off
- ~~**UX-1 / Dependabot / CI flake:**~~ ✅ closed 2026-06-22

---

## Блокеры / риски

| Риск | Mitigation |
|------|------------|
| ~~RU-V gaps (EN comments, encoding)~~ | ✅ RU-V10 PASS 2026-06-22 |
| ~~tms/project −0.2% gap~~ | ✅ FA-11 closed (82.0%) |
| ~~TMS CI webhook без RBAC (FA-02-02)~~ | ✅ FA-RBAC-FIBER3 + APIGW-01 2026-06-22 |
| ~~Fiber v3 RBAC middleware order~~ | ✅ FA-RBAC-FIBER3 123 routes 2026-06-22 |
| PG-03 not executed | FA-06 manual ops before prod cutover |
| ~~Uncommitted v4 block on branch~~ | ✅ PR #1 squash merged 2026-06-22 |
| ~~main.go ~629 LOC tech debt~~ | ✅ FA-08 wiring split 2026-06-22 |

---

## Session Continuity
> **Phase 147 147-02A static acceptance (2026-08-23):** immutable source commit `3f4d174f888f08a119bbe381dc86e5e9b34e0301` is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly the sole production file `backend/internal/calendar/itip_ingest_policy.go`. Independent Contract/high and Security/xhigh review is APPROVED C/I/M `0/0/0`. Under the current production-only test constraint, tests are intentionally absent and every executable/generation/runtime claim remains `NOT_RUN`. This accepts only the pure REQUEST/CANCEL order policy; it makes no product, merge, readiness, schema, generation, migration, SQL, Docker, or runtime claim. D0-next is `CONSUMED/HISTORICAL`; P3-C, 147-02B/147-02C, D1/D2, and P3-E retain their stated blockers; no next unblocked P3 source is selected.

**Last session:** 2026-08-14 (Mail 144-03A-Thread callable/encoding/oracle static acceptance sync)
**Stopped at:** 144-01A remains `STATIC_ACCEPTED` at
`e3e8c774d734f926f77f8e600b58dc68dabbd993`; 144-03A-Address is now `STATIC_ACCEPTED` at the
exact five-path boundary `41564d003946ddb78df18efca166ca86e4055405` (parent
`d659c67ce657fb33da391819edbdb954454be4b9`, tree `13a97333f08df6946e65766d3223c1ede85e1f50`)
after three independent Terra/high `EXACT_SHA_PASS` reviews, Spec/Quality/Security/Evidence PASS
and C/I/M `0/0/0`. Dispatcher-enforced provenance is Terra/xhigh authoring and Terra/high review,
fallback `none`; actual telemetry is `UNKNOWN` unless exposed. The launcher remains `NOT_READY`
and all executable evidence remains `NOT_RUN`.
`pro-qa-ai` beta.3 already exists both in protected Hub source
and as the Desktop portable copy. Legacy remains frozen; destructive move/delete
has not started.
**Resume:** Thread is narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`. Folder
source/oracle commit `f8cb7643986fc591cae16764575337b72ef3a692` (parent
`32e3dcb5b4f7ae65a88b79781646d25726ae16c6`, tree `14e7f6ce9a0f4b7912b955515d122c411159c579`)
is narrowly `STATIC_ACCEPTED`, static-only: exactly `backend/internal/mail/folder_tree.go`
(`3044055e1c5f6ff897fc646f4380d5a5a5be1b4b`) and
`backend/tests/mail/folder_tree_test.go` (`efe0db0071cb22ce4b026c635157f0fb08885d64`), after
Spec/Oracle Terra/high, Security/Adversarial Terra/xhigh and ABI/Compile-shape Terra/high
`EXACT_SHA_PASS` reviews, C/I/M `0/0/0`. Folder docs acceptance is `ca4e2854a319a020f7322ae0a9a98e3fb4a07374`.
The distinct `144-03B-Rule-Compiler` is narrowly docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at
`f199546513b758ca7b95328b5ccd7680009b732e`, reviewed against transition commit
`6735ae746c7acde657359e8a709cd897c4dce61c`; Contract/Spec Terra/high, Security/Adversarial
Terra/xhigh and Lifecycle/Dependency Terra/high PASS C/I/M `0/0/0`, dispatcher-enforced with no
fallback and telemetry `UNKNOWN`. Initial 144-02D/02F candidate
`63e419897aa0683e1d3dfcd442aca78eea1b0727` and transition
`6c424ed2d6703ceccbb269c37d9abc03e75a0410` are `CONSUMED/HISTORICAL`; its RecentContact
portion is superseded while receipt/deployment facts remain frozen. Corrective commit
`7d29d08c827cf857fc6fc8cc6a0f8201e42d4ecb` plus transition
`e563a3d3be8898371e5c8a9861c58f11b72167bc` is docs-only `STATIC_ACCEPTED` /
`EXACT_SHA_PASS`; all three post-transition exact reviews PASS C/I/M `0/0/0`. Former lifecycle is
  `CONSUMED/HISTORICAL`; correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus transition
  `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only accepted. Rule-plan
  authoring/review/acceptance is `CONSUMED/HISTORICAL`; its accepted plan confers no implementation authority. Launcher
`NOT_READY`, executable evidence `NOT_RUN`; no source/schema/generation/migration/service/runtime
authority. Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145
typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order;
post-task runtime `NOT_RUN`.
Its callable/encoding/oracle
contract remains `THREAD_CALLABLE_ENCODING_ORACLE_STATIC_ACCEPTED` at
`cbc12577a6444df9c660e43fedfd0c8133c6e461` / `EXACT_SHA_PASS` after independent
Spec/API/Persistence, Security and Lifecycle/Evidence PASS reviews with C/I/M `0/0/0`.
Thread implementation+oracle are narrowly `STATIC_ACCEPTED` at `9903f4df` / `EXACT_SHA_PASS`
(source `c221d547`, oracle `b21bd6fd`, C/I/M `0/0/0`), static only. Actual telemetry is
`UNKNOWN` unless exposed. Executable
verification remains `NOT_RUN`. The former 144-01B do-not-start/current-SHA Docker clause is
`CONSUMED/HISTORICAL`. Broader Phase 142/cross-contract work and Phase
148 remain pending.

> **144-01B static acceptance (2026-08-15):** immutable `0c064363b5ccb657a7c444802e94075c4ca743e5` (parent `48a7273d100c07788def88e2f1b8d69c7173580f`, tree `9c568a0c58c6bf7b8c3171cfd934760b839db623`, subject `feat(mail): add deterministic engine fake`) is `STATIC_ACCEPTED` / `EXACT_SHA_PASS` for fake `12bcd921668e406f07a98f7f3b0618ad8d44e81d` and authored-first test `b3654103b3a6d96bd887a1e6277aeb0ed6449b3f`. Terra/xhigh author; Spec Terra/high, Security Terra/xhigh, ABI Terra/high precommit and exact-SHA Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former 01B Docker-blocked/do-not-start/dependency clauses are `CONSUMED/HISTORICAL`. Launcher `NOT_READY`, executables `NOT_RUN`; no 02E/Ent/generation/migration/service/API/UI/product/phase/release authority; Phase 142/148 and all other blockers remain.

**Coverage decision:** TASK-203 backend profile **82.0373%** remains historical
discovery evidence. Founder waived statement-coverage parity as migration gate;
new quality gate is requirement/risk traceability plus fail-closed execution.

Исторические session notes (≤2026-07-23) и старые completion logs:
[`history/STATE-HISTORY-THROUGH-2026-07-23.md`](archive/history/STATE-HISTORY-THROUGH-2026-07-23.md)

---

*Обновляется после каждой фазы и `/handoff-sync`*

## Session note — 2026-07-29 (be2 · TASK-141 A3W2-4 tms import residual)

- A3W2-4-001..005 FIXED in TMS import/export; hubtest task141_*; status На проверке; be-ver: exclusive go generate then hubtest.

## Session note — 2026-07-29 (be2 · TASK-139 A3W2-2 tms services)

- A3W2-2-001..006 FIXED: AssignTesters membership (400), FileIDs project-scope (404), ImportAllure ExistingRun tx/compensating, mock UpdateTestResult scope, snapshot steps fail-closed, Delete/Link/Unlink AppError.
- hubtest `./testframework/tms/...` PASS; mutation 001–006 red/green ALL_OK.
- Коммиты: `1dbf42f4` fix · `9e634ee5` test · статус TASK → На проверке (README не трогали). Handoff: be-ver → code-reviewer.

> **Current Rule status (2026-08-14):** `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole added plan blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh and Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Rule-plan authoring/review/acceptance is `CONSUMED/HISTORICAL`; no Rule source/TDD/schema/migration/service/Phase145/runtime authority. Task 1 RED remains `NOT_AUTHORIZED/BLOCKED` pending ALL accepted non-marker Phase145 typed-IR/script carrier + G145-5B + G145-5C + launcher readiness, without implied order; post-task runtime `NOT_RUN`. 144-01B, Phase 142/148 and launcher `NOT_READY` remain unchanged.

> **144-04F source-authority receipt (2026-08-20):** immutable candidate
> `4e608e890ec12aeed8de965f9649ae503260b94e` (parent
> `c53a5fe60e1283cab58be5efaec7ffa131145b31`, tree
> `ed7e2bdab6a44254a7c677603ad0afc298aafed8`, subject
> `docs(mail): propose sync policy authority`) is `SOURCE_AUTHORITY_ACCEPTED` /
> `EXACT_SHA_PASS` for the exact9 blob map recorded in its tree. Contract, Security and Lifecycle
> immutable exact-SHA reviews PASS C/I/M `0/0/0`; provenance `dispatcher-enforced`, no fallback,
> telemetry `UNKNOWN`. The original candidate/precommit/commit-review/separate-acceptance-next
> clauses are `CONSUMED/HISTORICAL` for that accepted receipt. Its former immediate exact4
> root-dispatch clause is now `SUSPENDED / CONSUMED/HISTORICAL` for the cutoff correction route.
> **ROOT_CORRECTION_CANDIDATE / PRECOMMIT_REVIEW_REQUIRED / SOURCE_IMPLEMENTATION_BLOCKED / RUNTIME_NOT_RUN:** the following exact4 remains only the future test-first manifest:
> `backend/tests/mail/mailbox_events_test.go`,
> `backend/tests/mail/mailbox_sync_policy_test.go`, `backend/internal/mail/mailbox_events.go`,
> then `backend/internal/mail/mailbox_sync_policy.go`, with no substitution or extra path.
> Event-builder scope is only the three existing folder/label caps `50 -> 100` with 100 pass
> and 101 reject; no field/subject/version change. Existing uncommitted exact4 source/test candidates
> remain `UNACCEPTED` and MUST NOT be committed until Contract/Security/Lifecycle precommit PASS,
> one atomic exact9 correction-candidate commit, all three immutable exact-SHA PASS, and a separate
> docs-only acceptance sync. This correction is not accepted. Author provenance: Terra/medium,
> `dispatcher-enforced`, no fallback, telemetry `UNKNOWN`; executable/runtime evidence is `NOT_RUN`.
> The previously committed 144-04F authority acceptance sync records the existing receipt and is not itself an
> acceptance claim. Preserve 02C source
> `STATIC_ACCEPTED` at `e565523f589fbd7536f6f8e4df7effd39469c1af` and docs receipt
> `c53a5fe6`; this receipt selects only the exact4 manifest above. 02E/Ent/generation/migration/services/API/UI/
> product/phase/release, 02F-D, engine deletion/User-GDPR remain blocked/out of scope; launcher
> `NOT_READY`, executables/runtime `NOT_RUN`.

> **Phase 144 Task 6A sync coordination (2026-08-22):**
> `STATIC_ACCEPTED / RUNTIME_NOT_RUN` at source commit `24d2873d`. The exact packet owns
> claim/resolve plus bounded due-scan/fault Ent coordination only. It consumes accepted Task
> 6S/6P/6T, maps worker CAS time from `sync_schedule_updated_at`, uses one repository clock sample
> per transaction and never spans provider/network I/O. Prepare/apply projection work remains
> separate Task 6B/6C. Launcher `NOT_READY`; generation, migration, compile, SQL/DB, Docker,
> formatting/lint, tests, and runtime remain `NOT_RUN`. Independent immutable static review
> approved the exact six-file source boundary with C/I/M `0/0/0`.

> **Phase 144 Task 6T exact sync schedule clock (2026-08-22):**
> `STATIC_ACCEPTED / RUNTIME_NOT_RUN` at source commit `12f59cc8`.
> The authored-first `MailAccount` schema contract adds required
> `sync_schedule_updated_at` with a private exact-second UTC creation default, no
> `UpdateDefault`, and a fail-closed PostgreSQL second-precision CHECK. Task 6 worker due rows map
> their CAS `UpdatedAt` from this marker, not generic Ent `updated_at`; every schedule-affecting
> adapter mutation updates it in the same transaction from the repository clock. Task 9 owns
> reviewed existing-row backfill and the final `NOT NULL` migration. The launcher remains
> `NOT_READY`; executable RED/GREEN/tests, Ent generation, Atlas diff/migration, compile, SQL/DB,
> Docker, gofmt, lint, and runtime are all `NOT_RUN`. Independent static review approved the exact
> five-file source boundary with C/I/M `0/0/0`; no migration or cutover authority follows.

> **Phase 144 Task 6P sync-lease apply-proof schema (2026-08-22):**
> `STATIC_ACCEPTED / RUNTIME_NOT_RUN` at source commit `7309c52a`.
> The authored-first MailAccount schema-contract test freezes one bounded inline active-lease
> snapshot and one inline apply proof per account; it deliberately does not introduce an
> unbounded sync receipt table or retention cleanup. Snapshot state contains only non-negative
> sync/mutation epoch fences, phase, and bounded opaque cursor/base/checkpoint values. Proof state
> contains only a positive versioned 32-byte result hash, applied phase, next attempt and applied
> time. The two portable fail-closed `COALESCE` CHECKs require the exact valid lease-fence shapes,
> complete active-lease proof, expiry/order bounds, and allowed phase transitions, while leaving
> the original snapshot available for same-lease replay after the account cursor advances. No
> message/calendar/provider body, raw error, or JSON payload is stored. The launcher remains
> `NOT_READY`; executable RED/GREEN/tests, Ent generation, Atlas diff/migration, compile, SQL/DB,
> Docker, gofmt, lint, and runtime are all `NOT_RUN`. Independent static review found no issues in
> this narrow manifest; no migration or cutover authority follows.

> **Phase 147 Task 2 mail-bridge schemas (2026-08-22):** `SOURCE_CONTRACT_AUTHORED` /
> `PENDING_INDEPENDENT_REVIEW` / `RUNTIME_NOT_RUN`. An isolated Calendar schema source-contract
> test was authored first, followed by durable inbound ordering state, leased inbound message
> receipt, immutable reply-delivery intent, and only the required Calendar ownership inverse edges.
> Cross-domain Mail identities remain scalar UUID/hash values; no Mail Ent edge or
> provider/body/raw-part storage was added. CONTRACT §§2.2/3.1 now freeze the nullable error-code
> enum allowlists. The accepted Docker launcher is `NOT_READY`; executable RED/GREEN, Go
> test/build/gofmt/lint, Ent generation, Atlas migration, Docker, and runtime evidence are all
> `NOT_RUN`. This packet is not `STATIC_ACCEPTED`; independent review remains required.

> **Phase 143 Task 3 Calendar Ent transaction adapter (2026-08-22):**
> `SOURCE_CONTRACT_AUTHORED / PENDING_INDEPENDENT_REVIEW / RUNTIME_NOT_RUN`.
> The isolated transaction contract test was authored first. The adapter's private UTC clock
> computes a new mutation receipt expiry once per transaction as
> `now().UTC().Truncate(time.Second).Add(24 * time.Hour)`; a matching non-expired completed
> receipt replays, while expired-row cleanup/reuse remains a separate retention-cleanup
> dependency. The accepted Docker launcher and generated Ent artifacts are unavailable:
> executable RED/GREEN, compile, database, Ent generation, Atlas migration, Docker, and runtime
> evidence remain `NOT_RUN` until Task 9. Independent review is still required.

> **Phase 143 Task 4B event-to-call Ent bridge (2026-08-22):**
> `SOURCE_CONTRACT_AUTHORED / PENDING_INDEPENDENT_REVIEW / RUNTIME_NOT_RUN`.
> A separate `EventCallRepository`, private guarded wrapper and one Ent transaction per call
> preserve the base `EventRepository` ABI. Resolver and persistence paths lock CalendarEvent
> before ScheduledCall, compare exact event/call version fences, reject linked-row/caller-state
> divergence, and use only the returned pure policy snapshot/decision. The nullable
> `EventCallProjectionSnapshot.DeletedAt` is policy-owned: only `delete` carries an isolated
> validated exact-second UTC boundary instant; Ent writes it with `is_active=false` and no clock.
> Scheduled-due outbox append is atomic under `calls.channel.{channelId}.scheduled_due.v1`;
> NATS, HTTP, MailEngine, cutover/compatibility writer, backfill and worker remain out of scope.
> The launcher and generated Ent artifacts are unavailable, so executable RED/GREEN, Go
> compile/test/build/gofmt/lint, Ent generation, Atlas migration, Docker and runtime evidence
> remain `NOT_RUN` until Task 9.

> **Phase 143 Task 4 Calendar Ent optional capabilities (2026-08-22):**
> `SOURCE_CONTRACT_AUTHORED / PENDING_INDEPENDENT_REVIEW / RUNTIME_NOT_RUN`.
> A source-contract test preceded private Task 3 Ent backend implementations of the existing
> update, occurrence-cancel, series/occurrence RSVP and inbound-REPLY optional ports; the base
> `EventRepository` ABI was not expanded. The packet locks authoritative parent rows before
> ordered children, applies exact CAS/sequence fences, writes bounded receipts/outbox/reply
> delivery atomically, and preserves the frozen inbound lower/equal/newer/quarantine semantics
> without raw provider, iTIP or message-body persistence. Task 4B, cutover, NATS, MailEngine,
> HTTP and external I/O are excluded. The accepted Docker launcher and generated Ent artifacts
> remain unavailable, therefore executable RED/GREEN, Go compile/test/build/gofmt/lint, Ent
> generation, Atlas migration, Docker and runtime evidence are all `NOT_RUN` until Task 9.

> **Phase 147 P3-B static acceptance (2026-08-23):** `43a6c42af88524be2951966bdbfc83b422ecb7f9` is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly `backend/internal/calendar/event_repository.go`, `backend/internal/calendar/event_service_inbound_reply.go`, and `backend/internal/calendar/event_service_inbound_message.go`. It accepts transaction ABI only: a receipt-first single `EventTransaction`, opaque claim/scope/completion, and closed observed/error-code constructor ABI while preserving REPLY receipt/outbox behavior without nested transaction or duplicate logic. Independent Contract/high and Security/xhigh review is APPROVED C/I/M `0/0/0`. CONSUMED/HISTORICAL: P3-C was the sole next pending atom, blocked on Mail Ent generation, for exactly `backend/internal/mail/mailbox_ent_calendar_part.go` and `backend/internal/mail/mailbox_ent_repository.go`. P3-D/E remain pending/blocked. Generated Ent, migration, SQL, compile/build, tests, lint/gofmt, parser/validator, Docker, and runtime remain `NOT_RUN`; no merge, product, or Docker-readiness claim.
> **Current deletion-port source receipt (2026-08-23; highest precedence):** immutable source `e4f81bf3d8b932f5ff2b9a8069d41bfa1d1a0e26` (parent `bd76d8469782e9bb7926b440b6b45495bc865e84`, tree `93bac3c6a5798cb75b0a6647fa938416c5e82ede`) is `STATIC_ACCEPTED / STATIC_SOURCE_ONLY` for exactly `backend/internal/mail/mailbox_deletion.go` blob `001b03c879aab36031feac07f0d3894f76e23abd` and `backend/internal/mail/mailbox_deletion_fake.go` blob `7d43b4dee20ad762d251e3c14e428f2f0ccda497`. `provenance_source: dispatcher-enforced`; Terra/xhigh implementation; Contract/high, Security/xhigh, and Lifecycle/high immutable reviews are C/I/M `0/0/0`. This accepts only the provider-neutral deletion ABI and two standalone deterministic fakes. Tests, build, gofmt, lint, parser, validator, generation, migration, SQL, Docker, and runtime are `NOT_RUN`. Within this exact2 packet, former `UNIMPLEMENTED`/`UNCOMMITTED`/`UNACCEPTED`/exact2-next/source-review-pending clauses are `CONSUMED/HISTORICAL`. Concrete provider adapter capability, persistence/Ent/schema, migrations, workers/drains/finalization/User-GDPR/HTTP/NATS remain `BLOCKED`; no next source is selected. Testframework merge facts are unchanged.

> **Mail submission Claim/Authorize adapter (2026-08-25):** `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN` for `backend/internal/mail/mailbox_ent_submission_claim.go` and its static source-contract test. The candidate uses one provider-free dual-fence claim transaction and an ordered authorization CAS with immutable preflight/policy/attachment bindings plus organization-instance and user version-CAS rate buckets. Generated Mail Ent artifacts remain absent; Ent generation, compile, executable tests, formatting/lint, migration/SQL, Docker and runtime are all `NOT_RUN`. This is not merge, product or Docker-readiness acceptance.
> **Mail submission terminal draft lifecycle amendment (2026-08-25):** the source contract now locks the submission-linked draft in every terminal/requeue path, validates its submitting state, atomically returns it from `submitting` to `active` only when the submission becomes terminally `failed` (including `RetryExhausted`), and leaves it `submitting` for requeue/rate-limit. This remains `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN`; `provenance_source: dispatcher-enforced`.
> **Mail submission dispatch-resolution adapter (2026-08-25):** `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN` for `backend/internal/mail/mailbox_ent_submission_dispatch.go` and `backend/tests/mail/mailbox_ent_submission_dispatch_source_test.go`. One post-provider transaction locks Account -> Submission -> Draft, uses one database clock, requires the live dual fence plus the exact sealed/stored authorization hash, and resolves only to `accepted` or `unknown`; provider-ID mismatch persists `unknown`, never `queued`, with reconciliation attempts zero and database-time due. Both outcomes release the lease and advance the account epoch while the locked draft remains `submitting`. This is source-only authoring, not Ent generation, SQL/DB, executable verification, merge, product, or runtime acceptance; `provenance_source: dispatcher-enforced`.
> **Mail submission reconciliation lifecycle, stale recovery, and worker (2026-08-25):** the combined Ent/schema/domain/writer packet is `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN`. Reconciliation uses a separate 0..10 budget and database-time due index; waiting, owned, and exhausted `unknown` shapes are closed by SQL and source validation. The bounded stale scan covers `authorizing|submitting|unknown|accepted`; recovery re-locks Account -> Submission -> Draft, preserves existing authorizing/submitting rules, releases expired unknown ownership back to due or exact exhaustion, releases accepted ownership without finalizing, and treats an already released advisory replay as no-op while rejecting partial or mismatched fences. The provider-read-only reconciliation core exposes bounded UUID scan, sealed claim, one `ReconcileSubmission`, exact accepted binding, and cancellation-detached closed durable resolve. Its DB-time Ent scan/claim/resolve adapter and thin bounded sequential worker are independently accepted source-only with the same `0/0/0` verdict. The worker reuses the generic guarded loop and owns only UUID scan adaptation plus one `Reconcile` call per retained identity; it has no engine/provider/HTTP/NATS/claim/resolve/finalizer capability. Accepted-submission finalization and production composition remain pending. Generated Ent, compile, executable tests, formatting/lint, migration/SQL, Docker, and runtime are `NOT_RUN`; `provenance_source: dispatcher-enforced`.
> **Mail stale-submission recovery worker (2026-08-25):** `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN`. The thin wrapper adapts the bounded advisory stale scan to exactly one `RecoverStaleMailboxSubmission` per retained identity, validates every nil-error durable result, and fails closed on zero, unknown, or contradictory outcomes instead of counting them as recovered. It reuses the generic guarded worker loop and adds no engine/provider/HTTP/NATS/dispatch/reconciliation/finalization capability. Production composition, generated Ent, compile, executable tests, formatting/lint, migration/SQL, Docker, and runtime remain pending or `NOT_RUN`; `provenance_source: dispatcher-enforced`.
> **Mail sent visible-To correction (2026-08-25):** `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN`. `mail.sent.v1.to` is the exact required nonnull visible `To` array with 0..200 entries, so valid CC-only/BCC-only submissions can finalize without copying private recipients into the event. Nil remains invalid; `mail.received.v1.to` remains nonempty. Schema and focused tests are synchronized; executable verification, build, lint/gofmt, Docker and runtime are `NOT_RUN`; `provenance_source: dispatcher-enforced`.
> **Mail accepted-finalizer hydration core and worker (2026-08-25):** `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN` for the provider-neutral finalizer and its thin bounded worker. The sealed accepted claim freezes provider, binding, authorization, fence and clean storage-backed attachment authority. Hydration performs one exact message read and one byte-rehashed blob read per distinct provider blob, proves an exact normalized multiset with multiplicity, requires a non-draft canonical sent projection, and derives `confirmed_owner_submission` locally. Failure releases only the accepted lease through a detached context; completion ambiguity is never compensated. The worker reuses the guarded generic loop and owns only bounded accepted UUID scan adaptation plus one `Finalize` call per retained identity, with no engine/provider/HTTP/NATS/storage/claim/complete/release capability. The Ent completion adapter, durable message/thread/attachment/outbox transaction and production composition remain pending. Generated Ent, migration/SQL, compile, executable tests, build, lint/gofmt, Docker and runtime are `NOT_RUN`; `provenance_source: dispatcher-enforced`.
> **Mail submission sent-event derivation (2026-08-25):** `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN` for the pure deterministic `mail.sent.v1` helper. One confirmed non-draft projection plus exact To/CC/BCC recipient-resolution coverage yields a UUIDv8 event identity from the submission UUID, correlation by submission, immutable dispatch-authorization occurrence time, provider-confirmed sent time and conservative `external > instance_domain > internal` delivery class. The visible event `to` remains exact and may be empty; CC/BCC are never copied into it. Recipient database resolution, Ent finalizer completion, durable outbox insertion and production composition remain pending. Executable tests, build, lint/gofmt, Docker and runtime are `NOT_RUN`; `provenance_source: dispatcher-enforced`.
> **Mail accepted-submission Ent finalization adapter (2026-08-25):** `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN`. The DB-time scan excludes future rows; claim and completion use account-scoped candidates and a stable lock order with every thread locked before every message. Completion revalidates the sealed provider/binding/authorization/attachment authority, materializes the deterministic thread/message, transfers exact clean objects, resolves only bounded live recipient identities/domains without recipient row locks, inserts one deterministic `mail.sent.v1` outbox event, and atomically marks submission `sent`, draft `submitted`, and advances the account epoch. A repeated completion is a read-only no-op only after the exact immutable sent claim, deterministic confirmed-owner message, submitted draft and canonical correlated outbox are proven; later account epoch/lease or projection drift cannot create a false negative. Unknown private provider keywords are ignored unless a live matching Hub label exists. Generated Ent, migration/SQL, compile, executable tests, formatting/lint, Docker and runtime remain `NOT_RUN`; production composition remains separate; `provenance_source: dispatcher-enforced`.
> **Mail finalizer authority reconciliation (2026-08-25):** the accepted Ent adapter receipt immediately above supersedes the two older adjacent clauses that say Ent completion, recipient resolution, durable outbox insertion, or their adapter remain pending. Only production composition and executable/generated/runtime evidence remain pending there.
> **Mail durable send HTTP boundary (2026-08-25):** `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN` for `POST /api/mail/drafts/:draftId/send`. The authenticated empty-body transport requires one canonical `Idempotency-Key` and one exact quoted `If-Match`, rejects duplicate headers, and invokes only durable `CreateMailboxSubmission`; success and exact replay return the owner-safe receipt with `202 Accepted`. Closed 400/404/409/503 mapping exposes only documented safe codes and never invokes provider, dispatcher, storage, NATS or Calendar capability. OpenAPI is synchronized; server route/service wiring and production composition remain pending. Executable tests, build, lint/gofmt, Docker and runtime are `NOT_RUN`; `provenance_source: dispatcher-enforced`.
> **Mail submission production composition (2026-08-25):** `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN`. Only after successful engine-enabled Stalwart construction, server wiring now registers the durable draft-send handler and constructs exactly one queued dispatch, stale recovery, unknown reconciliation, and accepted finalization worker using the shared repository/engine and bounded default timers. All four run under `rootCtx`, constructor/route failures remain fatal, and cancellation is quiet; the engine-disabled branch registers none of this. No new NATS/config/shutdown/provider path was added. This supersedes adjacent server-wiring/production-composition-pending clauses for the submission HTTP route and these four workers only. Generated Ent, compile, executable tests, build, formatting/lint, Docker and runtime remain `NOT_RUN`; Mail<->Calendar consumers remain pending; `provenance_source: dispatcher-enforced`.
> **Calendar invite delivery durability schema (2026-08-25):** `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN` for `backend/ent/schema/calendarinvitedelivery.go` and its focused source contract. One immutable intent is replay-keyed by source envelope plus attendee and also protected by recurrence-aware event/attendee/sequence/kind uniqueness. The closed lifecycle uses pending/submitting/unknown/sent/superseded/failed-permanent states, bounded attempts and safe codes, database-due retry ordering, and a token/deadline/generation claim fence. Optional `MailSubmission` is the only path to its result MailMessage; source envelope is intentionally scalar without retention coupling, and no raw iTIP/content/address/provider/error detail is stored. Generated Ent, migration/SQL, compile, executable tests, formatting/lint, Docker and runtime remain `NOT_RUN`; consumer, worker and production composition remain pending; `provenance_source: dispatcher-enforced`.
> **Calendar availability release-all semantics (2026-08-25):** `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN`. A nil occurrence set remains invalid, while an explicit non-nil empty authoritative set keeps organizer/event authority and releases all active reservations for the event. The release path locks event -> deferred constraint -> active rows, validates the locked rows, reads canonical DB time, and releases without member/source conflict reads; nonempty reconciliation preserves event -> constraint -> members -> active -> leaves/events/rows/calls. Executable tests, compile, formatting/lint, Docker and runtime remain `NOT_RUN`; post-persist update/cancel wiring remains separate; `provenance_source: dispatcher-enforced`.
> **Calendar reply delivery lifecycle amendment (2026-08-25):** `SOURCE_CONTRACT_AUTHORED / INDEPENDENT_REVIEW_PASS_C_I_M_0_0_0 / RUNTIME_NOT_RUN` for `backend/ent/schema/calendarreplydelivery.go` and its focused source contract. Mutable lease token/deadline/generation, bounded attempts, `superseded`, a pending/unknown due index, and optional `MailSubmission` `NO ACTION` linkage make the RSVP intent claimable by the future Mail worker. Fresh intents may sleep until outbox wake; submitting is unbound claim-before-bind, unknown/sent are bound, superseded is unbound, and retry exhaustion is accepted only at attempt 8. No raw iTIP/content/address/provider/error detail is stored. Generated Ent, migration/SQL, compile, executable tests, formatting/lint, Docker and runtime remain `NOT_RUN`; ABI, adapter, consumer, worker and production composition remain pending; `provenance_source: dispatcher-enforced`.
