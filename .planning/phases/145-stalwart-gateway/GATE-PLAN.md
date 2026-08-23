# Phase 145 — external and runtime gate plan
> **Engine deletion-port adjudication (2026-08-23):** `STATIC_ACCEPTED / DOCS_DESIGN_ONLY`; former deletion-port-next selection is `CONSUMED/HISTORICAL`, never source authority. Freeze separate opaque future ports: `EngineAccountDeletionPort` (`DeleteAccount`, `ReconcileAccountAbsence`) and `MailObjectDeletionPort` (`DeleteObject`, `ReconcileObjectAbsence`); do not widen normal `MailEngine` or combine account delete with drain. Requests durably retain nonzero `DeletionID` plus account provisioning key+engine account ID or object bucket+key; opaque IDs use `MaxEngineOpaqueIDBytes` and bucket/key existing storage bounds. Results are `deleted|authoritatively_absent`, observations `present|authoritatively_absent`; bare `NotFound` preserves intent unless scoped authoritative absence is proven. No provider/raw IDs or causes are public/logged. Private intent/control is not owner projection; short lease transactions keep I/O outside locks, use exact token+generation+scope CAS, reclaim expired lease only, and stale token/zero rows cause no I/O/compensation. Object drain starts after `upload_drain_until` and attachment/reference terminality; account removal is last after exact child/object/receipt/control/activity manifest; hard-delete only after closed evidence. No Stalwart route/OAuth/DNS/OpenAPI assumption. Adapter capability and independent contract/security/lifecycle reviews are mandatory; future source is bounded ABI/fake then separately planned persistence/worker. Ent/source/generated/migration/SQL/tests/Docker/runtime and physical User/GDPR remain `BLOCKED` / `NOT_RUN`.

> **Current Mail docs/readiness receipt (2026-08-15):** `576f19231e37f635d5b7187a1ada121ed14ff8dc` is the exact14 docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS` receipt. Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high reviews PASS C/I/M `0/0/0`; `provenance_source: dispatcher-enforced`, no fallback, telemetry `UNKNOWN` unless exposed. It grants no source/schema/runtime/product/phase/release authority. All lower active 02D/02F-only, candidate, precommit, root-commit, exact-review and status-sync-next routes are `CONSUMED/HISTORICAL` as applicable while retaining unrelated historical facts. 02F accepted and Rule/Sieve plan facts remain accepted, without source/schema/runtime authority. 144-01B, Phase 142/148, G145-5B/G145-5C and all other gates remain unchanged; launcher `NOT_READY`, executables `NOT_RUN`. Engine deletion-port and User deletion-intent contracts remain future blockers; the nearest permitted work is docs-only engine deletion-port adjudication, not code.
> **Carrier-plan acceptance receipt (2026-08-15):** `PLAN-145-07B-SIEVE-CARRIER.md` is docs/implementation-authority-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `949b876d1c5ea07338926aa5fe19e1864a256e93` (parent `375891bb85c0e89f2b6597e0416b82219fffaa74`, tree `f943fb60e8dbda3ae612871d189a57d751dc44cc`, sole blob `0bbbc91cd3c58fc8daeaab1a434d942f3ef26821`); three exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Contract `1daef787` and sync `375891bb` remain accepted. Only obsolete plan-pending fragments are `CONSUMED/HISTORICAL`; unrelated G145 gates remain current. No source/test/runtime authority. Rule Task 1 awaits G145-5B/G145-5C/launcher. Carrier six-path RED/source additionally awaits accepted Rule compiler source/ABI. 145-07B blocked; `NOT_READY`/`NOT_RUN` unchanged.

> **Carrier-lifecycle clause precedence:** Only the lower clause or fragment asserting carrier **contract** acceptance is pending, unaccepted, or open is superseded and `CONSUMED/HISTORICAL`; every unrelated/current fact in the same record remains authoritative, including 02F accepted, Phase 145 blocked, implementation-plan/source, G145-5A/B/C/D/6-9, launcher and runtime status. Contract `1daef787` is docs/design accepted.

> **Carrier acceptance receipt (2026-08-15):** Carrier `1daef78768048d7672c38a65b78ef0f119f3c207` (parent `c14c208889e8aa1ce7f005a3d9ce03d1a3bb6dc6`, tree `1086a93f2f52423cfdf9dcfc01eadd2e0d434947`, blob `218a38117ae51d014f40218e0f5c4606ec8d4dfa`) is docs/design/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`, all exact reviews PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`; base/correction lifecycle `CONSUMED/HISTORICAL`. No carrier plan/source/test/runtime authority. Root-only separate carrier-plan acceptance, G145-5B/G145-5C and launcher readiness keep Rule Task 1/145-07B `NOT_AUTHORIZED/BLOCKED`; launcher `NOT_READY`, executables `NOT_RUN`.

> **Current upstream lifecycle (2026-08-14):** correction `fcb5b84fed4a417e736ddb4053256e494a0c25f6` plus exact-28 transition `2f4381a2cf12a1a2b07bdee5ce6e8309d2f10260` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. Former lifecycle and Rule-plan-next instructions are `CONSUMED/HISTORICAL`; G145-5B/G145-5C stay open. Accepted Rule plan `669ef761` does not authorize Rule Task 1, which remains `NOT_AUTHORIZED/BLOCKED` on the independent carrier/G145-5B/G145-5C/launcher prerequisites without implied order.

> **Status:** BLOCKED_BY_144_ACCEPTANCE_AND_EXTERNAL_GATES
> **No implementation authorization:** this plan does not approve a license, image,
> network egress, DNS change, secret, deployment or runtime command.

> **Upstream lifecycle update:** 144-02D remains docs-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; 02F `fcb5b84f` plus `2f4381a2` is docs/readiness-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` after Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former lifecycle is `CONSUMED/HISTORICAL`; this does not close G145-5B/G145-5C or authorize runtime.

> **Current Rule-plan acceptance:** Standalone `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4` (parent `77eb4356de06a5304521b60c57af4e73c5740ce1`, tree `1b554d1c5763818ffef4f3c06c2a5bde16046d37`, sole path, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`); Contract/Spec/Oracle Terra/high, Security/Adversarial/Purity Terra/xhigh, Lifecycle/Dependency/Boundary Terra/high PASS C/I/M `0/0/0`; dispatcher-enforced/no fallback/telemetry `UNKNOWN` unless exposed. This accepts only the plan: former Rule-plan author/adjudicate/accept instructions are `CONSUMED/HISTORICAL`; no Rule oracle/source/TDD, schema, migration, service, Phase145 carrier, runtime, product, phase or release authority follows. Rule Task 1 RED is `NOT_AUTHORIZED/BLOCKED` until all accepted non-marker Phase145 typed-IR/script carrier, G145-5B, G145-5C and launcher readiness, without implied order; post-task runtime `NOT_RUN`. Task4 docs sync remains separately discovered after future source Task3 acceptance. 02D/02F docs-only acceptance, 144-01B Docker blocker, Phase142/148 and release gates, launcher `NOT_READY`, executable tests/build/lint/validators/Docker/runtime `NOT_RUN` remain unchanged.

## Gate order

| Gate | Evidence required | Owner / authority | Current state | Failure action |
|---|---|---|---|---|
| G145-0 engine boundary | independently accepted projection-only pages, bounded content/blob request shape, hard admin maxima and stable over-limit behavior | founder + architecture/security review | 0.1.8 ROOT_FROZEN / INDEPENDENT_REVIEW_PENDING | no fake/adapter implementation against an unaccepted port |
| G145-1 license | written choice: unmodified sidecar under AGPL-3.0 or accepted SELv2/commercial terms | founder/legal | UNRESOLVED | select docker-mailserver fallback behind the same port; re-plan adapter |
| G145-2 runtime model | exact `gpt-5.6-terra` / `xhigh` requested and actual provenance for each code packet | founder/runtime operator | ROUTE_AUTHORIZED / new packet actual telemetry unavailable | reject any mismatch or fallback |
| G145-3 launcher | versioned task-owned outer launcher implementing limits, timeout, report owner marker and exact cleanup | test architecture owner | NOT_READY | no direct Docker/Compose fallback; runtime remains NOT_RUN |
| G145-4 provenance | refreshed official release, immutable digest, signature/SBOM/source and approval record | DevOps/security/founder | `v0.16.17` Docker Hub OCI index + linux/amd64 manifest + SLSA attestation references CAPTURED / SIGNATURE AND SBOM NOT_VERIFIED / NOT_APPROVED | do not pull or reference mutable image |
| G145-5 config contract | reviewed Stalwart config/API/session/health/volume/listener/access-control paths for the selected immutable version | adapter reviewer | ROUTE P0 ROOT-CORRECTED: shared `/jmap` Registry+mail contexts, `/api` helpers, health, volume and default-deny listener baseline; SOURCE-ONLY REREVIEW NO P0/P1, P2 ROOT-FIXED, ACTUAL TELEMETRY UNKNOWN / NOT GATE ACCEPTANCE; exact image config/export/checksum and feedback semantics NOT_FROZEN | stop rather than guess version-specific fields; reject route-by-credential logic, remove unused listeners and never inherit allow-all HTTP |
| G145-5A folder writer exclusivity | reviewed config, internal-network policy and published-port evidence prove only Hub can create/rename/delete custom folders; system roles are distinguishable | adapter + security reviewer | NOT_RUN | capability stays false; automatic folder recovery makes zero mutation and remains degraded |
| G145-5B typed credential contexts | reviewed least-privilege management API key for exact Registry methods on shared `/jmap` plus separately scoped mailbox JMAP authentication/delegation, rotation, revocation and audit | founder + security + adapter reviewer | MANAGEMENT TEN-PERMISSION REPLACE SCOPE + EXACT AUTH/ACL SOURCE ROOT_FROZEN / PRIOR SOURCE REVIEW SUPERSEDED BY ROUTE P0 / OAUTH AUTHORIZATION-CODE RECOMMENDED / IMPERSONATION HIGH-RISK FALLBACK / FOUNDER CHOICE + ACCEPTED RE-REVIEW PENDING | keep engine disabled; reject mixed Registry/mail batches, URL-selected auth, management-key mail use, OAuth client-credentials assumptions, ACL-only full-engine mode and invented per-user passwords |
| G145-5C domain verification authority | reviewed authority proves the exact bounded public DNS projection before Hub may report `verified`; selected Stalwart `x:Domain` has no manual-DNS verification field | founder + security + adapter reviewer | UNRESOLVED / P1 SELECTED-VERSION MISMATCH | adapter returns at most `dns_required`; account provisioning and external mode make zero engine/mail calls |
| G145-5D OAuth origin boundary | one user-reachable HTTPS origin, exact OAuth UI/endpoint allowlist and service-only shared `/jmap` enforcement if Authorization Code is selected | founder + security + DevOps + adapter reviewer | SINGLE `STALWART_PUBLIC_URL` + EXACT `POST /api/auth` EXCEPTION CONSTRAINTS ROOT_FROZEN / TOPOLOGY AND COMPLETE UI PATHS UNRESOLVED | OAuth mode remains invalid; no public JMAP, `/api` prefix allow or guessed reverse proxy |
| G145-6 local profile | bounded health, authenticated JMAP, internal delivery, restart persistence and exact cleanup | accepted launcher | NOT_RUN | gateway remains disabled/not_configured |
| G145-7 resource budget | observed whole added profile ≤4 CPU and ≤4 GiB under idle + representative load | accepted launcher | NOT_RUN | reduce/rebalance or reject candidate |
| G145-8 backup/restore | dedicated engine-volume backup and restore exercise with current SHA/digest | ops + accepted launcher | NOT_RUN | no production readiness claim |
| G145-9 external mail | approved domain/DNS/operator, PTR coordination, SPF/DKIM/DMARC/MX and controlled deliverability exercise | founder/operator | UNRESOLVED | keep external mode disabled |

Gates are sequential where evidence depends on an earlier choice. G145-1 never derives
from repository text. G145-4 must be refreshed immediately before an authorized pull/run;
the 2026-08-13 `v0.16.17` research snapshot and read-only registry digest capture are not
an immutable-image approval. G145-5B
cannot be inferred from “administrator credentials”: official Stalwart API keys cannot
authenticate JMAP mail, while impersonation is a separate broad permission. The management
half uses API-key `Replace` with only domain CRUD/list and individual-account CRUD/list;
that root freeze still needs independent security review. G145-5C cannot be inferred from
domain existence, `isEnabled` or a non-empty `dnsZoneFile`. Account identities are separate
stable RFC 8621 `Identity/get|set` objects on mailbox JMAP; using them remains blocked by
the unresolved G145-5B mailbox-auth design, not by an id-derivation fallback.
Exact candidate source also prevents an ACL-only shortcut: Email and Mailbox methods may
use shared-account access, but Identity, EmailSubmission, Sieve and Blob paths require
account membership. Stalwart documents Authorization Code and Device Authorization only,
not an unattended OAuth client-credentials grant. The default recommendation is therefore
per-user Authorization Code OAuth; zero-touch operation requires explicit acceptance of a
broad dedicated `impersonate` principal and its incident-grade controls.
Exact selected-version source also corrects the former route split: Registry `x:*` and
standard mail methods share `POST /jmap`; `/api` is a helper/auth namespace. The two typed
contexts still use different capabilities and credentials, may never be mixed in one batch,
and select authorization before request construction rather than from the URL.
If Authorization Code is selected, G145-5D is also mandatory: one
`STALWART_PUBLIC_URL` supplies OAuth/OIDC and JMAP discovery URLs, so separate internal Hub
and browser origins cannot be invented. The accepted topology may expose exact rate-limited
`POST /api/auth` for OAuth login, but must keep both Registry and mailbox `POST /jmap`, all
unrelated `/api/*`, upload/download and operator surfaces Hub/service-only. Prefix-wide
`/api` access is forbidden.

## Disabled/default behavior

- `MAILBOX_ENGINE_ENABLED=false` produces `not_configured` and does not affect existing
  Resend alert readiness.
- Enabled but unreachable/misconfigured engine produces mailbox API `503`/degraded health;
  server readiness policy must be explicitly wired and tested, not silently reported ready.
- `MAILBOX_EXTERNAL_ENABLED=false` remains the default even after local engine setup.
- `MAILBOX_EXCLUSIVE_FOLDER_WRITER=false` remains the safe default. It may be true only in
  an accepted profile that publishes no user-facing IMAP/JMAP/ManageSieve or direct
  engine-admin mailbox mutation surface.
- Missing domain or secret refuses external mode without printing values.
- Without G145-5C, a present Stalwart domain remains `dns_required`; Hub neither reports
  `verified` nor provisions accounts against it.
- Invalid/expired JMAP cursor marks the affected mailbox projection degraded and triggers
  a bounded owner-scoped rebuild. It does not silently reset freshness.
- Enabled startup owns exactly one bounded domain-control, account-control, combined projection/recovery and rule-deployment
  worker on the process root context; disabled startup owns none. Shutdown cancels and
  drains them through registered reverse-order cleanup.

## Secret and endpoint rules

- One trusted internal origin comes only from process configuration. Management Registry
  and mailbox methods share fixed relative `/jmap`; mail discovery and upload/download URLs
  are accepted only under the reviewed same-origin policy. `/api` is limited to individually
  allowlisted helper/auth path+method pairs. No request parameter, domain record or session may redirect
  either typed context.
- Management and mailbox authorization providers are distinct even on shared `/jmap`.
  The scoped management API key is sent only for an exact Registry method allowlist; it is
  never sent for standard mail methods, upload/download, `/api/*` helpers or mail protocols.
  Mailbox credentials never authorize `x:*` Registry methods. Mixed-context batches and
  URL-selected authorization fail before I/O. No guessed fallback creates or persists one
  plaintext password per user.
- Shared-mailbox ACL is not an accepted full-engine authentication mode. Before G145-5B is
  accepted, configuration accepts only `disabled`; it must reject OAuth client-credentials,
  ACL-only, impersonation and per-user password modes before any network I/O.
- Stalwart HTTP access control defaults allow-all. The accepted profile is default-deny by
  listener/source/path/method, has explicit connection caps, publishes no steady-state 8080,
  and removes every unused IMAP/POP3/submission/ManageSieve listener. OAuth adds no public
  HTTP surface before G145-5D; `STALWART_PUBLIC_URL` never silently rewrites Hub discovery
  to another trust origin.
- Management keys, mailbox credentials, DKIM private keys and generated tokens are
  secret-provider or engine-volume state. Example files contain names/placeholders only.
- `STALWART_RECOVERY_ADMIN`, recovery mode and the HTTP configuration listener are
  operator-only bootstrap/break-glass mechanisms. They are absent from an accepted
  steady-state profile and are never reused by Hub.
- Logs, health and API responses redact authorization headers, session capability URLs,
  account ids if credential-like, raw provider bodies and SMTP recipient/body content.
- Selected reverse-proxy/access-log configuration logs no query string for mail search and
  no concrete remote-image token path (constant route label or disabled path logging).
  Application route-template logging is necessary but not accepted as proxy proof.
- Production transport ports are not published by default. Wave 1 may explicitly publish
  SMTP delivery ports after G145-9; it never publishes user-facing IMAP/JMAP/ManageSieve or
  direct engine-admin mailbox mutation. Internal JMAP/admin exposure is service-network
  scoped to Hub. Any alternate writer keeps the exclusivity capability false.

## Runtime evidence manifest

When authorized, the accepted launcher report must record exact repository SHA, launcher
version, image tag + digest, config checksum without secrets, container/resource names,
limits, timestamps, wall-clock timeout, cases, observed CPU/RSS and unconditional cleanup.
Required cases:

1. disabled/not_configured startup and unchanged Resend alert routes;
2. engine health, separately authenticated management Registry compatibility and mailbox
   JMAP session discovery on shared `/jmap`; prove wrong-context credentials and mixed
   Registry/mail batches are rejected before I/O, and `/api` is not used for provisioning;
3. account provisioning and default folders;
3b. domain create/DNS-required/delete desired state reconciles through Ensure/Get without
    automatic DNS change or private DKIM exposure; before G145-5C, existence, enabled state
    and expected zone text never produce `verified`; after G145-5C, the accepted authority
    alone can promote the domain;
3a. provisioning lost response repeats the same stable key, two replicas create one logical
    account/address set, then mailbox `Identity/get|set` reconciles exactly one stable opaque
    identity id per desired address across reorder/restart/lost response; duplicate or
    unexpected identities degrade safely, and disabled/missing-domain paths make zero engine calls;
4. internal submit creates one Sent and one recipient Inbox view;
5. lost-submit response reconciles without duplicate;
6. draft version conflict and lost-save reconciliation;
7. change cursor advance, invalid cursor degraded state, base-state/ID-anchored enumeration,
   query-state drift catch-up, typed restart on anchor/base-history loss, and command-lease
   begin/during/end exclusion from sync capture/apply;
8. crash/resume during enumeration and catch-up, two-replica lease exclusion, worker context
   cancellation/drain, expired command-lease reclaim/late-owner reconciliation,
   engine-volume persistence and Hub projection recovery;
8a. fixed Hub Sieve generation/hash apply, lost-response Get-before-retry, stale-generation
    non-finalization and external script drift repair; explicit spam and ham feedback proves
    accepted/rejected/indeterminate outcomes separately from Junk folder movement;
8b. selected-version feedback idempotency/reconciliation capability is evidenced; if
    absent, a lost response stays terminal unknown and the stable command is never retried;
8c. folder create persists one pending intent before engine I/O; with proven exclusive
    writer capability a lost response survives restart and exact non-system match/
    definitive-absence recovery cannot duplicate the folder. Rename/delete recover only the
    persisted desired intent. False/unknown capability, an exposed alternate writer,
    system-role candidate or ambiguous state makes zero recovery mutation and remains
    degraded;
8d. enabled startup owns exactly four worker schedulers (domain, account, combined
    projection/recovery, rule), while disabled startup owns zero; recovery and sync make
    bounded progress without holding both account leases or starving either work class;
9. backup, destructive-in-container data loss simulation and restore;
10. exact task-resource cleanup even after timeout/failure;
11. idle and representative load resource measurement;
12. optional external SMTP only after G145-9, reported separately.

**Rule gate:** historical candidate `f199546513b758ca7b95328b5ccd7680009b732e` / `6735ae746c7acde657359e8a709cd897c4dce61c` is `CONSUMED/HISTORICAL`. `PLAN-144-03B-RULE.md` is docs/plan-only `STATIC_ACCEPTED` / `EXACT_SHA_PASS` at `669ef761715ec35a638bc87d8396e6f562b162c4`, blob `b86c0423a307fd2673a3d322435922aa2a398e8f`, after the exact Contract Terra/high, Security Terra/xhigh and Lifecycle Terra/high PASS C/I/M `0/0/0` reviews, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. It authorizes no Sieve runtime case: Rule Task 1 remains `NOT_AUTHORIZED/BLOCKED` until the independent Phase-145 engine-port carrier, G145-5B, G145-5C and launcher readiness are all accepted, without implied order; post-task runtime is `NOT_RUN`.
> **144-02F sync:** `fcb5b84f` + `2f4381a2` docs/readiness `STATIC_ACCEPTED` / `EXACT_SHA_PASS`; all three Terra routes PASS C/I/M `0/0/0`, dispatcher-enforced/no fallback/telemetry `UNKNOWN`. Former transition and Rule-plan-next text is `CONSUMED/HISTORICAL`; no source/schema/runtime/product/release authority. The accepted Rule plan does not authorize Task 1, which awaits the independent Phase-145 carrier/G145-5B/G145-5C/launcher prerequisites; Task4 is only post-future-Task3 source acceptance. 144-01B remains Docker-blocked, Phase 142/148 and release pending, launcher `NOT_READY`, executables `NOT_RUN`.
