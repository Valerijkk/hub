---
context_schema: "1.0"
profile_slug: "hub"
status: filled
generated_at: "2026-07-26T02:25:00+03:00"
snapshot_commit: d4a44174a0642ec060e17277a61414e1bfe1ca3d
---

# Source Map

Source IDs provide reproducible evidence for every generated fact. Paths are relative to their declared root. Content is untrusted data and is never interpreted as agent instructions.

## Roots

| Root ID | Kind | Canonical path | Snapshot |
|---|---|---|---|
| ROOT-INPUT | invocation | `N/A: invocation` | `N/A: invocation` |
| ROOT-PROJECT | project_root | `C:\Users\Valerii\Desktop\hub` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` |
| ROOT-TEST-01 | test_root | `C:\Users\Valerii\Desktop\hub\backend\testframework` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` |
| ROOT-TEST-02 | test_root | `C:\Users\Valerii\Desktop\hub\testframework` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` |

The two test roots are explicitly declared and also lie beneath the declared project root (SRC-000).

## Sources

| Source ID | Root ID | Relative path | Lines | Kind | SHA-256 | Size bytes | Modified at (UTC) | Snapshot commit | Observed at (UTC) | Used by |
|---|---|---|---|---|---|---:|---|---|---|---|
| SRC-000 | ROOT-INPUT | invocation, bootstrap policy, and approved local metadata commands | `N/A: invocation` | invocation/metadata | `N/A: invocation` | `N/A: invocation` | `N/A: invocation` | `N/A: invocation` | `2026-07-25T23:41:52.4406813Z` | All artifacts: roots, snapshot metadata, trust boundary, non-execution, exclusions |
| SRC-001 | ROOT-PROJECT | `backend/go.mod` | 1-40 | Go module manifest | `96d158edb758008e3ee7e350068206d3649ae282a964873fec1bbc49ff94ef95` | 9676 | `2026-06-22T14:22:56.5198971Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE, COMMANDS |
| SRC-002 | ROOT-PROJECT | `frontend/package.json` | 1-20, 63-80 | npm manifest | `39bb61aed3d9eba751599bfe7381e123bcc6e816ac93aa4ae0fd80f9952b5a62` | 3358 | `2026-06-22T20:53:33.1210716Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE |
| SRC-003 | ROOT-PROJECT | `frontend/vite.config.ts` | 9-10, 77-83 | Vitest/Vite configuration | `d6003a1a96f1850b4f215878e29efb47b2a6b29d3ea732c696cd55e55fda1a9c` | 2462 | `2026-06-22T20:53:33.1384721Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE, COMMANDS, TRACEABILITY |
| SRC-004 | ROOT-PROJECT | `frontend/playwright.config.ts` | 10-55 | Playwright configuration | `6f5a24f3fdee4fbb3908d9577df72bde798e80c7ba140009c4943ee78bbfe86a` | 1585 | `2026-06-22T20:53:33.1210716Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE, COMMANDS, CI-PIPELINE, TRACEABILITY |
| SRC-005 | ROOT-TEST-02 | `stages/e2e/playwright.config.ts` | 1-7 | Playwright forwarding configuration | `cb2ac501274d89fd14f8ef861ecfdbb411c5cc35e9691ebee4828be1fa5f405d` | 276 | `2026-06-22T20:53:33.1639653Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | SOURCE-MAP |
| SRC-006 | ROOT-PROJECT | `.github/workflows/ci.yml` | 1-342 | GitHub Actions workflow | `b01f045d228aec2df87d95e589ff342ed3f60d9725de8f84b07d3dcc35aae8d6` | 10810 | `2026-06-22T20:53:33.1006137Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE, COMMANDS, CI-PIPELINE, TEST-DATA |
| SRC-007 | ROOT-TEST-02 | `PIPELINE.md` | 1-76 | Test pipeline documentation | `2a076f1630d63207d64d8ac19d57ea69072e220c1435f76299b9e780e2d36945` | 3297 | `2026-06-22T20:53:33.1418976Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE, COMMANDS, CI-PIPELINE |
| SRC-008 | ROOT-TEST-02 | `pipeline/stages.yaml` | 1-24 | Pipeline stage manifest | `52b4c968edd9834df69c30005e66bdc321e7174d0b807a9e5fe423e1451d4318` | 458 | `2026-06-22T20:53:33.1472590Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, CI-PIPELINE |
| SRC-009 | ROOT-TEST-02 | `pipeline/jobs/backend-test.yaml` | 1-49 | Pipeline job manifest | `7cdf4a0a0e4451c1b2363ff739e4d44478c34d59d85dcc38e1b95ac8e88e2ebf` | 751 | `2026-06-22T20:53:33.1441196Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE, COMMANDS, CI-PIPELINE, TRACEABILITY |
| SRC-010 | ROOT-TEST-02 | `pipeline/jobs/frontend-test.yaml` | 1-60 | Pipeline job manifest | `87bbabbf65431684f357c15adae4b5be4b08773414e67b814cc204470d6ed349` | 1464 | `2026-06-22T20:53:33.1467350Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, COMMANDS, CI-PIPELINE |
| SRC-011 | ROOT-TEST-02 | `pipeline/jobs/e2e-smoke.yaml` | 1-16 | Pipeline job manifest | `16cb081338781d595370384f78dc325377af19b93fa7aaa154fae3fd129d8f52` | 410 | `2026-06-22T20:53:33.1462097Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE, CI-PIPELINE, TRACEABILITY |
| SRC-012 | ROOT-TEST-02 | `pipeline/jobs/e2e-full.yaml` | 1-40 | Pipeline job manifest | `5d3f3137e295019457017c1dcf76e2bf8567f1f8e78914257f891454f76989db` | 873 | `2026-06-22T20:53:33.1457049Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE, CI-PIPELINE, TRACEABILITY |
| SRC-013 | ROOT-TEST-02 | `pipeline/jobs/coverage-gate.yaml` | 1-10 | Pipeline job manifest | `298b225c8eb201010cb22b71c6805336278462ec60075be776f3a823e1743530` | 327 | `2026-06-22T20:53:33.1451608Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | COMMANDS, CI-PIPELINE |
| SRC-014 | ROOT-PROJECT | `docs/vision_and_scope.md` | 8-14 | Product requirements documentation | `4566e21a276167b96a5554f42a97eb49e4df35af5af60a93d91a38688659d58a` | 59935 | `2026-07-25T22:30:18.9632602Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT |
| SRC-015 | ROOT-PROJECT | `docs/architecture.md` | 9-18, 54-115 | Architecture documentation | `e5aafc0cce7b4fbc23aceac3b98567477bb9e91856812bc04138cf3650709eb5` | 14373 | `2026-07-17T22:21:28.9770627Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT |
| SRC-016 | ROOT-PROJECT | `docs/test_plan.md` | 8-29, 46-60, 90-107, 145-157, 170-183 | Test strategy documentation | `d53f667fdeda48aaa106d44a26a1a8410e6d952799c0fa2bb0bad899ae3b6e3b` | 9374 | `2026-07-26T11:33:00.3607756Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-26T11:38:21.5106347Z` | PROJECT, TEST-ARCHITECTURE, TRACEABILITY, conflicts |
| SRC-017 | ROOT-PROJECT | `docs/code_style_guide.md` | 21-46, 107-138, 142-167, 260-270 | Coding and test style documentation | `199e70ef31bf20ffce1cce53d5b9222310f9110e0d00966e1d277b98881b5786` | 9762 | `2026-06-22T20:53:33.1189829Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE, TEST-DATA, TRACEABILITY |
| SRC-018 | ROOT-PROJECT | `docs/security_guidelines.md` | 104-145, 156-183, 232-243 | Security documentation | `2101375cf02031ab719a63795de52f84e8f598b61257f3e6998dc5c0a4badbb8` | 18362 | `2026-07-17T22:21:28.9880633Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, COMMANDS, CI-PIPELINE, TEST-DATA, TRACEABILITY |
| SRC-019 | ROOT-PROJECT | `docs/tz_constraints.md` | 10-110 | Mandatory/forbidden constraints | `de05222e06f3e852924280adaf32dd8014b1d0e69ca50d0be2677c3ed23e34b5` | 7533 | `2026-07-02T23:25:06.3416868Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT |
| SRC-020 | ROOT-PROJECT | `docs/GLOSSARY.md` | 11-36, 109-119 | Identifier and test terminology documentation | `e38ea687a7f517b9870a76108c1b7da3154698e3da4f3d540296f1ebb64e242d` | 12917 | `2026-07-19T15:45:37.3986186Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TRACEABILITY |
| SRC-021 | ROOT-PROJECT | `docs/TROUBLESHOOTING.md` | 11-63, 111-139 | Checked-out troubleshooting documentation | `7ebe24da8a16857c7c10f9e5c89db0f455a5804e0a349e37b49148113cc22712` | 11444 | `2026-07-19T17:05:30.7869060Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE, COMMANDS, conflicts |
| SRC-022 | ROOT-PROJECT | `docs/TESTING.md` | 1-35, 200-219, 335-359 | Test-case and strategy documentation | `fbb3841fbbe61a0915686b068df2cf90243f6a9f12f3e9b6c94693e9c90553a6` | 37816 | `2026-07-19T17:09:48.0617500Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE, TRACEABILITY, conflicts |
| SRC-023 | ROOT-TEST-01 | `env/builder.go` | 1-96 | Test environment helper | `aa45a86ad9b2e8e9fe5ece22c949bd85a0fb93f64d54c6f84926da8ca1961d19` | 2609 | `2026-06-22T20:02:17.1887118Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, CI-PIPELINE, TEST-DATA |
| SRC-024 | ROOT-TEST-01 | `factory/user.go` | 1-37 | Test factory | `e2a469ddf38b4da3d9a2c205a105f60233c7cb877757708df62d7433337377c1` | 1091 | `2026-06-22T20:02:17.1947367Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, TEST-DATA |
| SRC-025 | ROOT-TEST-01 | `mongotest/setup.go` | 1-56 | Mongo test helper | `321c6135488d00f23eeeb5034f40ec0f0804444eb07e746d592ce0ddfd056aee` | 1767 | `2026-06-22T20:02:20.1740558Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, COMMANDS, TEST-DATA |
| SRC-026 | ROOT-TEST-01 | `steps/http.go` | 1-25 | HTTP test step helper | `f5d61e498b3a93ea06fb178ef524a38f7f89e6efb08f1bae9221201269e1e5af` | 698 | `2026-06-22T20:02:20.6301456Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE |
| SRC-027 | ROOT-TEST-01 | `auth/service_test.go` | 117-177 | Go auth tests | `b7c48186202710b53c3c2abfa6bccb49c3439dae4fcfe33b8334ce8c267acf0e` | 7269 | `2026-07-25T22:30:18.9643021Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, TRACEABILITY |
| SRC-028 | ROOT-TEST-01 | `auth/magiclink_test.go` | 19-31, 34-145 | Go Redis/auth tests | `3c78896a9d0e57947a32341d5d516898282e3f1a8fd330adeabc82f8a63c5156` | 11722 | `2026-06-22T20:02:16.3597674Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, CI-PIPELINE, TEST-DATA |
| SRC-029 | ROOT-TEST-01 | `bases/rbac_test.go` | 22-88, 220-236 | Go handler/RBAC tests | `71f0e368267ee72eeb81e92e9b1775e914109e971267f35dd3da929c72d6191e` | 9072 | `2026-07-11T19:04:12.1104018Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, CI-PIPELINE, TEST-DATA |
| SRC-030 | ROOT-TEST-01 | `platform/nats/client_integration_test.go` | 1-107 | Go NATS integration tests | `e2ce8bdac594515160d5aa39d2ccce929b5aa4260bdeabf010ffbadd96246e61` | 2744 | `2026-06-22T20:02:20.3497530Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, CI-PIPELINE, TEST-DATA |
| SRC-031 | ROOT-TEST-01 | `messenger/mongo_repository_test.go` | 17-83 | Go Mongo integration test | `c7dfc8a2e853513ade320545cd60da6ec20aec4890babfd62c8f5d83d35bbf35` | 2780 | `2026-06-22T20:07:20.1308915Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-DATA |
| SRC-032 | ROOT-TEST-02 | `stages/unit/frontend-test/render.tsx` | 7-49 | Frontend test renderer | `6cd2ebd81b49b8211cefba094fef62fd5e394addd3684022a1b2a4eb8bd09fd2` | 1367 | `2026-06-22T20:53:33.2010946Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, TEST-DATA |
| SRC-033 | ROOT-TEST-02 | `stages/unit/frontend-test/domains/auth/auth-store.test.ts` | 1-66 | Vitest test source | `105406cf4049207080831b41706e0a90dfd53bb55e58a157e60d3af0b2402133` | 1688 | `2026-06-22T20:53:33.1767532Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, TEST-DATA, TRACEABILITY |
| SRC-034 | ROOT-TEST-02 | `stages/unit/frontend-test/domains/tracker/issue-form.test.tsx` | 1-80 | Vitest/RTL test source | `fc130373f5d5e21db16c4eb35ac481b9d0c72ac40215607947ae52674be511e5` | 2314 | `2026-06-22T20:53:33.1928692Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, TRACEABILITY |
| SRC-035 | ROOT-TEST-02 | `stages/unit/frontend-test/domains/ai/ai-privacy.test.tsx` | 1-34 | Vitest/RTL test source | `ba00486e3259c13e34aa4b1a51ee283f5096c530bd29cca45daea1f8613a946f` | 1049 | `2026-06-22T20:53:33.1767532Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, TRACEABILITY |
| SRC-036 | ROOT-TEST-02 | `stages/e2e/shared/fixtures.ts` | 1-50 | Playwright fixtures | `9d4bf347068f22bd19cb4f8ed475084333279669bd199cecab1e0aa13378599c` | 1582 | `2026-06-22T20:53:33.1644941Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, COMMANDS, TEST-DATA |
| SRC-037 | ROOT-TEST-02 | `stages/e2e/shared/helpers/backend.ts` | 1-57 | E2E backend helper | `bcaf13e440d5e0892a7f061380787a6d5c95074b62d8e33387b94da1759c4d76` | 1416 | `2026-06-22T20:53:33.1673488Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, COMMANDS, CI-PIPELINE |
| SRC-038 | ROOT-TEST-02 | `stages/e2e/shared/helpers/auth.ts` | 1-72 | E2E auth/data helper | `ea266f19dd35d277529ef402c43bce7d71e3b6b673082685d7521fd69492830c` | 1815 | `2026-06-22T20:53:33.1665829Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, TEST-DATA |
| SRC-039 | ROOT-TEST-02 | `stages/e2e/shared/pages/base.page.ts` | 1-23 | Page object source | `0ddbbe18af870ccff414cc2fe2e4d199818a728985137441cf4ac1ef0fea6bc1` | 787 | `2026-06-22T20:53:33.1694876Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE |
| SRC-040 | ROOT-TEST-02 | `stages/e2e/shared/pages/login.page.ts` | 1-40 | Page object source | `27f3d99d841f56418e9441c8c30b18c007d92a8a5160772d6547ddbf33140bce` | 1393 | `2026-06-22T20:53:33.1710470Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE |
| SRC-041 | ROOT-TEST-02 | `stages/e2e/shared/steps/auth.steps.ts` | 1-126 | E2E step source | `89fff3e7086b18c4009fca307b32d03de11f2a1a3b8f527999eb77bca3ec338c` | 3671 | `2026-06-22T20:53:33.1720880Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, TEST-ARCHITECTURE |
| SRC-042 | ROOT-TEST-02 | `stages/e2e/domains/auth/login.spec.ts` | 1-41 | Playwright test source | `77595d042c5a8df6a1c5cff2ac2582f9eebaf49e16c2a8485e3fe814ce7cd3f5` | 1333 | `2026-06-22T20:53:33.1531384Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, TRACEABILITY |
| SRC-043 | ROOT-TEST-01 | `apigateway/apikey_auth_test.go` | 146-185 | Go security regression test | `6eb876151e15da297c703465176889b4301512a4323038fecce274cae680038b` | 6485 | `2026-06-22T20:02:15.7289353Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TRACEABILITY |
| SRC-044 | ROOT-TEST-01 | `calls/recording_webhook_test.go` | 156-271 | Go security regression tests | `86f47395b07b7be1d011cdf042a655e785a538c2c45f7abf335c74b06c9670eb` | 11713 | `2026-06-22T20:02:17.1074560Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TRACEABILITY |
| SRC-045 | ROOT-TEST-01 | `ws/hub_test.go` | 70-163 | Go concurrency regression test | `8cc634d29658d24f662e9bf19a0d5ccdc504756a9f96494d21db224c134b685d` | 12252 | `2026-06-22T20:02:21.5712355Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, TEST-DATA, TRACEABILITY |
| SRC-046 | ROOT-TEST-02 | `guides/06-edge-corner-coverage.md` | 1-123 | Test design/traceability guide | `cb9a56581d75f3fa167535287b73a07d600dd49c2d3e210cf0397c307450e1fc` | 6207 | `2026-07-17T22:21:29.0195444Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | TEST-ARCHITECTURE, TEST-DATA, TRACEABILITY |
| SRC-047 | ROOT-TEST-02 | `stages/unit/frontend-test/README.md` | 1-27 | Frontend suite documentation | `0cc106c2d3e77d790c297c29ec7de68c68ddde99e346e1b618a5f66534c8f3d8` | 884 | `2026-06-22T20:53:33.1751886Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT, COMMANDS |
| SRC-048 | ROOT-TEST-02 | `stages/e2e/README.md` | 1-24 | E2E suite documentation | `666a842de54d8b9d87d5ed8b694f96c13d93e4ea1ff662e16f0afe4c83d05e23` | 775 | `2026-06-22T20:53:33.1488438Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT |
| SRC-049 | ROOT-PROJECT | `docs/system_specification.md` | 8-76 | Functional/non-functional requirements | `1a08c3f8887a817e3c6bd91b00d06aaf43e71728b538c8072a681f142dc49c5a` | 53673 | `2026-07-17T22:21:28.9910640Z` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-25T23:41:52.4406813Z` | PROJECT |
| SRC-050 | ROOT-TEST-01 | `**/*_test.go` | `N/A: filesystem inventory` | Filesystem inventory | `N/A: inventory` | `N/A: inventory` | `N/A: inventory` | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-26T11:38:21.5106347Z` | PROJECT, TEST-ARCHITECTURE, conflicts |

## Filesystem inventory observations

| Source ID | Root ID | Pattern | Observed count | Tool method | Snapshot commit | Observed at (UTC) |
|---|---|---|---:|---|---|---|
| SRC-050 | ROOT-TEST-01 | `**/*_test.go` | 499 | Dedicated `functions.Glob` file search/count | `d4a44174a0642ec060e17277a61414e1bfe1ca3d` | `2026-07-26T11:38:21.5106347Z` |

## Generated fact map

| Artifact section | Fact summary | Source IDs | Verification |
|---|---|---|---|
| PROJECT.md#identity | Project name, purpose, monorepo shape, and local snapshot commit | SRC-000, SRC-014, SRC-015, SRC-007 | read at snapshot; project commands not run |
| PROJECT.md#declared-scan-boundary | Explicit root paths and existence | SRC-000 | validated with local read/search tools |
| PROJECT.md#technology-inventory | Go, TypeScript, runner, dependency, and build-tool versions | SRC-001, SRC-002, SRC-003, SRC-004, SRC-006 | machine-readable manifests/config read |
| PROJECT.md#repository-and-test-layout | Product/test directory roles | SRC-001, SRC-002, SRC-003, SRC-004, SRC-007, SRC-009, SRC-010, SRC-012, SRC-015, SRC-016, SRC-036, SRC-039, SRC-041, SRC-047, SRC-048 | paths and cited content read |
| PROJECT.md#governing-documentation | Requirements, testing, style, and security authorities | SRC-014, SRC-016, SRC-017, SRC-018, SRC-019, SRC-022, SRC-049 | cited documents read |
| PROJECT.md#verified-constraints | Test selection, E2E split, live backend inventory, and scan restrictions | SRC-000, SRC-003, SRC-004, SRC-006, SRC-009, SRC-011, SRC-012, SRC-016, SRC-018, SRC-050 | cited content and inventory read; project commands not run |
| TEST-ARCHITECTURE.md#test-topology | Observed runners, roots, selectors, and suite counts | SRC-001, SRC-002, SRC-003, SRC-004, SRC-006, SRC-009, SRC-010, SRC-011, SRC-012, SRC-016, SRC-022, SRC-050 | configs/manifests/docs and live file inventory compared |
| TEST-ARCHITECTURE.md#organization | Domain grouping, naming, and mixed unit/integration boundary | SRC-003, SRC-004, SRC-007, SRC-009, SRC-010, SRC-012, SRC-016, SRC-017, SRC-025, SRC-027, SRC-028, SRC-030, SRC-033, SRC-042 | representative files sampled |
| TEST-ARCHITECTURE.md#framework-components | Builders, fixtures, helpers, doubles, clients, POM, and reporting | SRC-004, SRC-006, SRC-023, SRC-024, SRC-026, SRC-028, SRC-029, SRC-030, SRC-032, SRC-034, SRC-035, SRC-036, SRC-037, SRC-038, SRC-039, SRC-040, SRC-041 | representative files sampled |
| TEST-ARCHITECTURE.md#runtime-behavior | Cleanup, isolation, parallelism, time, retries/skips, and artifacts | SRC-004, SRC-006, SRC-023, SRC-025, SRC-028, SRC-032, SRC-033, SRC-036, SRC-040, SRC-045, SRC-046 | static source inspection only |
| COMMANDS.md#discovered-commands | Exact checked-out test/build/quality commands, all not-run | SRC-006, SRC-007, SRC-009, SRC-010, SRC-013, SRC-047 | copied verbatim; not executed |
| COMMANDS.md#required-prerequisites | Observed runtimes, link step, browser, services, and script dependencies | SRC-001, SRC-003, SRC-004, SRC-006, SRC-013, SRC-021, SRC-025, SRC-036 | static source inspection only |
| COMMANDS.md#environment-variable-names | Names and purposes only; no values copied | SRC-004, SRC-006, SRC-018, SRC-025, SRC-037 | names transcribed; values excluded |
| CI-PIPELINE.md#pipeline-inventory | Checked-out workflow and local mirror | SRC-006, SRC-007, SRC-008, SRC-009, SRC-010, SRC-011, SRC-012, SRC-013 | local files only |
| CI-PIPELINE.md#test-jobs | Jobs, dependencies, commands, services, and artifacts | SRC-004, SRC-006, SRC-008, SRC-009, SRC-010, SRC-011, SRC-012, SRC-013, SRC-023, SRC-028, SRC-030 | local files only; no remote state |
| CI-PIPELINE.md#quality-gates | Blocking relations, coverage threshold, and non-blocking Lighthouse | SRC-004, SRC-006, SRC-008, SRC-010, SRC-011, SRC-012, SRC-013 | workflow/manifests read |
| CI-PIPELINE.md#caches-and-generated-artifacts | Cache keys and artifact retention | SRC-006, SRC-013 | workflow read |
| CI-PIPELINE.md#ci-environment | Environment names only | SRC-004, SRC-006, SRC-018, SRC-037 | values excluded |
| TEST-DATA.md#data-builders-and-fixtures | Data builders, factories, renderer, fixtures, and dev seed/session mechanism | SRC-023, SRC-024, SRC-025, SRC-032, SRC-036, SRC-037, SRC-038 | source read; no data created |
| TEST-DATA.md#backing-services | SQLite, Redis, MongoDB, NATS, and CI service lifecycle | SRC-006, SRC-023, SRC-025, SRC-028, SRC-029, SRC-030, SRC-031 | source read; no service started |
| TEST-DATA.md#isolation-and-cleanup | Unique IDs, cleanup, parallelism policy, and time control | SRC-004, SRC-023, SRC-024, SRC-025, SRC-028, SRC-032, SRC-033, SRC-045, SRC-046 | representative source inspection |
| TEST-DATA.md#sensitive-data-controls | Profile exclusion plus project logging guidance | SRC-000, SRC-017, SRC-018 | values excluded |
| TRACEABILITY.md#identifier-conventions | Requirement, test-case, automated-test, and change IDs | SRC-003, SRC-004, SRC-017, SRC-020, SRC-022 | documentation/config read |
| TRACEABILITY.md#requirement-to-test-map | Selected explicit security, TC, and E2E links | SRC-009, SRC-011, SRC-012, SRC-016, SRC-018, SRC-022, SRC-027, SRC-042, SRC-043, SRC-044, SRC-045 | both mapping and target source read |
| TRACEABILITY.md#unlinked-tests | Sampled tests with no inline project identifier | SRC-033, SRC-034, SRC-035 | source read; external linkage remains UNKNOWN |
| TRACEABILITY.md#uncovered-or-unverifiable-requirements | Explicit manual/unspecified AUTH rows and incomplete global chain | SRC-020, SRC-022, SRC-046 | documentation compared |
| TRACEABILITY.md#traceability-rules | Project traceability process and CI mapping | SRC-009, SRC-010, SRC-011, SRC-012, SRC-020, SRC-022, SRC-046 | documentation/manifests read |

Verification means evidence was read at the recorded snapshot. It does not imply a discovered command was executed.

## Excluded sources

| Path/pattern | Reason | Evidence |
|---|---|---|
| `.git/**`, VCS internals, remote-ref caches, review metadata | Source-control internals/history and review sources excluded | SRC-000 |
| `pro-qa-ai/context/**` except the selected output profile and neutral templates used to initialize it | Tool-owned knowledge cannot evidence project facts | SRC-000 |
| `pro-qa-ai/patterns/**` | Pattern corpus excluded; no cards created or promoted | SRC-000 |
| `pro-qa-ai/review-feedback/**` | Review feedback excluded from bootstrap | SRC-000 |
| `**/node_modules/**`, `**/vendor/**`, `**/third_party/**` | Dependencies/vendored code excluded | SRC-000 |
| `**/dist/**`, `**/build/**`, `**/target/**`, `**/coverage/**` | Build/cache/coverage output excluded | SRC-000 |
| Traces, screenshots, Playwright/Lighthouse/generated reports | Generated reports and media excluded | SRC-000, SRC-006 |
| `.env`, `.env.*`, credentials, private keys, certificates, keystores, token dumps | Secret-bearing files excluded; example files were also not read under the user's stricter boundary | SRC-000 |
| Binaries, archives, media, databases, core dumps, minified/generated blobs | Non-safe-text content excluded | SRC-000 |
| Symlink/junction targets outside declared roots | External targets not followed | SRC-000 |

## Conflicts

| Topic | Sources | Conflict | Resolution |
|---|---|---|---|
| Backend test-file count | SRC-016, SRC-021, SRC-022, SRC-050 | Historical profile snapshot recorded 495 versus 499; corrected canonical docs and the live filesystem inventory now report 499 | RESOLVED — dedicated `functions.Glob` count for `backend/testframework/**/*_test.go` returned 499 |
| Universal TC automation claim | SRC-022 | Line 202 says every TC is linked to automated unit tests, while AUTH rows 215-217 explicitly mark manual/E2E or generic middleware coverage without exact automated paths | UNKNOWN |

Unresolved conflicts use `UNKNOWN`; resolved rows retain their history and resolution evidence.

## Manual evidence

<!-- BEGIN MANUAL -->
UNKNOWN
<!-- END MANUAL -->
