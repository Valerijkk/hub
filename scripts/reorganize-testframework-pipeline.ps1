# Reorganize testframework: stages -> jobs -> domains
# Usage: powershell -ExecutionPolicy Bypass -File scripts/reorganize-testframework-pipeline.ps1
$ErrorActionPreference = "Stop"
$Root = git rev-parse --show-toplevel
Set-Location $Root

function Ensure-Dir($p) {
    if (-not (Test-Path $p)) { New-Item -ItemType Directory -Path $p -Force | Out-Null }
}

$feSrc = "testframework/frontend"
$feDst = "testframework/stages/unit/frontend-test"
$e2eSrc = "testframework/e2e"
$e2eShared = "testframework/stages/e2e/shared"
$e2eDomains = "testframework/stages/e2e/domains"

Ensure-Dir "testframework/stages/unit/backend-test"
Ensure-Dir "testframework/stages/build"
Ensure-Dir "testframework/stages/quality/coverage-gate"
Ensure-Dir $feDst
Ensure-Dir $e2eShared
Ensure-Dir $e2eDomains

Write-Host "=== E2E shared assets ==="
foreach ($dir in @("pages", "steps", "helpers")) {
    $from = Join-Path $e2eSrc $dir
    $to = Join-Path $e2eShared $dir
    if (Test-Path $from) {
        Ensure-Dir (Split-Path $to -Parent)
        if (-not (Test-Path $to)) { git mv $from $to }
    }
}
if (Test-Path "$e2eSrc/fixtures.ts") {
    git mv "$e2eSrc/fixtures.ts" "$e2eShared/fixtures.ts"
}

Write-Host "=== E2E specs -> domains ==="
$e2eMap = @{
    "login.spec.ts" = "auth"
    "logout.spec.ts" = "auth"
    "onboarding-profile.spec.ts" = "auth"
    "navigation.spec.ts" = "navigation"
    "command-palette.spec.ts" = "navigation"
    "shortcuts.spec.ts" = "navigation"
    "admin-audit.spec.ts" = "admin"
    "admin-ai.spec.ts" = "admin"
    "admin-metrics.spec.ts" = "admin"
    "messenger.spec.ts" = "messenger"
    "project-lifecycle.spec.ts" = "tracker"
    "tms.spec.ts" = "tms"
    "ai-hub.spec.ts" = "ai"
    "automation.spec.ts" = "ai"
    "health.spec.ts" = "platform"
    "metrics.spec.ts" = "platform"
    "notifications.spec.ts" = "platform"
    "a11y.spec.ts" = "platform"
    "critical-gate.spec.ts" = "platform"
    "v4-superapp.spec.ts" = "v4"
}
foreach ($entry in $e2eMap.GetEnumerator()) {
    $from = Join-Path "$e2eSrc/specs" $entry.Key
    $domain = $entry.Value
    Ensure-Dir (Join-Path $e2eDomains $domain)
    $to = Join-Path $e2eDomains "$domain/$($entry.Key)"
    if (Test-Path $from) { git mv $from $to }
}

Write-Host "=== Frontend unit -> domains ==="
if (Test-Path "$feSrc/render.tsx") {
    git mv "$feSrc/render.tsx" "$feDst/render.tsx"
}

$feMoves = @(
    @("unit/api/admin.test.ts", "domains/admin/admin-api.test.ts"),
    @("unit/api/ai.test.ts", "domains/ai/ai-api.test.ts"),
    @("unit/api/client.test.ts", "domains/auth/client-api.test.ts"),
    @("unit/components/admin/audit-log.test.tsx", "domains/admin/audit-log.test.tsx"),
    @("unit/components/admin/onec-settings.test.tsx", "domains/admin/onec-settings.test.tsx"),
    @("unit/components/admin/operation-metrics.test.tsx", "domains/admin/operation-metrics.test.tsx"),
    @("unit/components/ai/model-picker.test.tsx", "domains/ai/model-picker.test.tsx"),
    @("unit/components/dashboard/widgets/system-health.test.tsx", "domains/dashboard/system-health.test.tsx"),
    @("unit/components/help/context-help-tip.test.tsx", "domains/platform/context-help-tip.test.tsx"),
    @("unit/components/help/shortcuts-dialog.test.tsx", "domains/platform/shortcuts-dialog.test.tsx"),
    @("unit/components/layout/skip-link.test.tsx", "domains/platform/skip-link.test.tsx"),
    @("unit/components/messenger/channel-icon.test.tsx", "domains/messenger/channel-icon.test.tsx"),
    @("unit/components/tracker/issue-card.test.tsx", "domains/tracker/issue-card.test.tsx"),
    @("unit/components/tracker/issue-form.test.tsx", "domains/tracker/issue-form.test.tsx"),
    @("unit/components/tracker/user-picker.test.tsx", "domains/tracker/user-picker.test.tsx"),
    @("unit/components/ui/error-boundary.test.tsx", "domains/platform/error-boundary.test.tsx"),
    @("unit/hooks/use-auth.test.ts", "domains/auth/use-auth.test.ts"),
    @("unit/hooks/use-composer-ai-suggest.test.ts", "domains/ai/use-composer-ai-suggest.test.ts"),
    @("unit/i18n/validation-keys.test.ts", "domains/platform/validation-keys.test.ts"),
    @("unit/lib/format-api-error.test.ts", "domains/platform/format-api-error.test.ts"),
    @("unit/lib/sanitize-color.test.ts", "domains/platform/sanitize-color.test.ts"),
    @("unit/lib/onboarding/profile-schema.test.ts", "domains/auth/profile-schema.test.ts"),
    @("unit/lib/onboarding/project-schema.test.ts", "domains/tracker/project-schema.test.ts"),
    @("unit/modules/accounting/accounting-page.test.tsx", "domains/v4/accounting-page.test.tsx"),
    @("unit/modules/bases/bases-page.test.tsx", "domains/v4/bases-page.test.tsx"),
    @("unit/modules/calendar/unified-calendar.test.tsx", "domains/v4/unified-calendar.test.tsx"),
    @("unit/modules/crm/crm-page.test.tsx", "domains/v4/crm-page.test.tsx"),
    @("unit/modules/hr/leave-requests.test.tsx", "domains/v4/leave-requests.test.tsx"),
    @("unit/modules/intranet/news-feed.test.tsx", "domains/v4/news-feed.test.tsx"),
    @("unit/modules/intranet/org-directory.test.tsx", "domains/v4/org-directory.test.tsx"),
    @("unit/modules/intranet/random-coffee.test.tsx", "domains/v4/random-coffee.test.tsx"),
    @("unit/modules/marketplace/marketplace-page.test.tsx", "domains/v4/marketplace-page.test.tsx"),
    @("unit/modules/settings/ai-privacy.test.tsx", "domains/ai/ai-privacy.test.tsx"),
    @("unit/modules/status/status-page.test.tsx", "domains/v4/status-page.test.tsx"),
    @("unit/modules/tms/components/case-table-row.test.tsx", "domains/tms/case-table-row.test.tsx"),
    @("unit/stores/ai-store.test.ts", "domains/ai/ai-store.test.ts"),
    @("unit/stores/auth-store.test.ts", "domains/auth/auth-store.test.ts")
)
foreach ($m in $feMoves) {
    $from = Join-Path $feSrc $m[0]
    $to = Join-Path $feDst $m[1]
    Ensure-Dir (Split-Path $to -Parent)
    if (Test-Path $from) { git mv $from $to }
}

Write-Host "=== Fix E2E imports (../ -> ../../shared/) ==="
Get-ChildItem "$e2eDomains" -Recurse -Filter "*.spec.ts" | ForEach-Object {
    $c = Get-Content $_.FullName -Raw -Encoding UTF8
    $n = $c `
        -replace 'from "\.\./fixtures"', 'from "../../shared/fixtures"' `
        -replace 'from "\.\./steps"', 'from "../../shared/steps"' `
        -replace 'from "\.\./pages"', 'from "../../shared/pages"' `
        -replace 'from "\.\./helpers/', 'from "../../shared/helpers/'
    if ($n -ne $c) {
        [System.IO.File]::WriteAllText($_.FullName, $n, [System.Text.UTF8Encoding]::new($false))
    }
}

Write-Host "=== Fix shared/fixtures.ts helper imports ==="
$fixtures = "$e2eShared/fixtures.ts"
if (Test-Path $fixtures) {
    $c = Get-Content $fixtures -Raw -Encoding UTF8
    $n = $c -replace 'from "\./helpers/', 'from "./helpers/'
    [System.IO.File]::WriteAllText($fixtures, $n, [System.Text.UTF8Encoding]::new($false))
}

Write-Host "Done. Remove empty legacy dirs manually if needed."
