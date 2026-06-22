# Rewire unexported search symbols in backend/testframework/search to export_hooks.
$ErrorActionPreference = "Stop"
$dir = Join-Path (Split-Path -Parent $PSScriptRoot) "backend/testframework/search"

Get-ChildItem -Path $dir -Filter "*_test.go" | ForEach-Object {
    $c = Get-Content $_.FullName -Raw -Encoding UTF8
    $orig = $c

    $c = $c -replace '\bgetUserID\(', 'search.ExportGetUserID('
    $c = $c -replace '\bsavedFilterToDTO\(', 'search.ExportSavedFilterToDTO('
    $c = $c -replace '\bdateFieldPredicate\(', 'search.ExportDateFieldPredicate('
    $c = $c -replace '\bidx\.host\b', 'search.ExportMeiliIndexerHost(idx)'
    $c = $c -replace '\bissuesIndex\b', 'search.ExportIssuesIndex()'
    $c = $c -replace '\bidx\.upsertDocuments\(', 'search.ExportMeiliUpsertDocuments(idx, '

    $c = $c -replace '\[\]meiliIssueDoc\{\{ID: "1"\}\}', '[search.ExportMeiliIssueDoc{ID: "1"}]'

    $c = $c -replace '&search\.MeiliIndexer\{host:\s*([^,]+),\s*client:\s*([^}]+)\}', 'search.ExportNewMeiliIndexerTest($1, "", $2)'

    $c = $c -replace 'svc\.searchIssues\(', 'search.ExportSearchIssues(svc, '
    $c = $c -replace 'svc\.searchTestCases\(', 'search.ExportSearchTestCases(svc, '
    $c = $c -replace 'svc\.searchUsers\(', 'search.ExportSearchUsers(svc, '
    $c = $c -replace 'svc\.searchUsersUnscoped\(', 'search.ExportSearchUsersUnscoped(svc, '
    $c = $c -replace 'adminSvc\.searchUsers\(', 'search.ExportSearchUsers(adminSvc, '
    $c = $c -replace 'memberSvc\.searchUsers\(', 'search.ExportSearchUsers(memberSvc, '
    $c = $c -replace 'memberSvc\.searchUsersUnscoped\(', 'search.ExportSearchUsersUnscoped(memberSvc, '
    $c = $c -replace 'svc\.searchProjects\(', 'search.ExportSearchProjects(svc, '
    $c = $c -replace 'svc\.searchFiles\(', 'search.ExportSearchFiles(svc, '
    $c = $c -replace 'svc\.searchChannels\(', 'search.ExportSearchChannels(svc, '
    $c = $c -replace 'svc\.searchMessages\(', 'search.ExportSearchMessages(svc, '

    $c = $c -replace 'svc\.getUserChannelIDs\(', 'search.ExportGetUserChannelIDs(svc, '
    $c = $c -replace 'svc\.getUserProjectIDs\(', 'search.ExportGetUserProjectIDs(svc, '

    $c = $c -replace 'svc\.repo\s*=\s*([^;\r\n]+)', 'search.ExportSetSearchServiceRepo(svc, $1)'

    $c = $c -replace '&search\.SearchService\{sqlDB:\s*newScopeTestDB\(t\),\s*repo:\s*&mockSearchRepo\{\}\}', 'search.ExportNewSearchServiceRepoSQL(&mockSearchRepo{}, newScopeTestDB(t))'
    $c = $c -replace '&search\.SearchService\{repo:\s*&mockSearchRepo\{\},\s*sqlDB:\s*([^,]+),\s*mongoDB:\s*nil\}', 'search.ExportNewSearchServiceFull(&mockSearchRepo{}, $1, nil)'
    $c = $c -replace '&search\.SearchService\{sqlDB:\s*([^,]+),\s*mongoDB:\s*nil\}', 'search.ExportNewSearchServiceFull(nil, $1, nil)'
    $c = $c -replace '&search\.SearchService\{repo:\s*&mockSearchRepo\{isAdmin:\s*true\}\}', 'search.ExportNewSearchServiceRepoOnly(&mockSearchRepo{isAdmin: true})'
    $c = $c -replace '&search\.SearchService\{sqlDB:\s*newScopeTestDB\(t\)\}', 'search.ExportNewSearchServiceSQL(newScopeTestDB(t))'
    $c = $c -replace '&search\.SearchService\{sqlDB:\s*db\}', 'search.ExportNewSearchServiceSQL(db)'

    # Multiline admin repo struct
    $c = [regex]::Replace($c, '(?ms)&search\.SearchService\{\s*sqlDB:\s*([^,]+),\s*repo:\s*&adminSearchRepo\{admin:\s*true\},\s*\}', 'search.ExportNewSearchServiceRepoSQL(&adminSearchRepo{admin: true}, $1)')
    $c = [regex]::Replace($c, '(?ms)&search\.SearchService\{\s*sqlDB:\s*([^,]+),\s*repo:\s*&mockSearchRepo\{isAdmin:\s*true\},\s*\}', 'search.ExportNewSearchServiceRepoSQL(&mockSearchRepo{isAdmin: true}, $1)')
    $c = [regex]::Replace($c, '(?ms)&search\.SearchService\{\s*sqlDB:\s*([^,]+),\s*repo:\s*&mockSearchRepo\{isAdmin:\s*false\},\s*\}', 'search.ExportNewSearchServiceRepoSQL(&mockSearchRepo{isAdmin: false}, $1)')

    $c = [regex]::Replace($c, '(?ms)idx := &search\.MeiliIndexer\{\s*host:\s*([^,]+),\s*apiKey:\s*"([^"]*)",\s*client:\s*([^,]+),\s*\}', 'idx := search.ExportNewMeiliIndexerTest($1, "$2", $3)')

    if ($c -match '\bsearch\.Export' -and $c -notmatch 'github.com/valerii/hub/backend/internal/search') {
        if ($c -match '(?ms)(import \(\r?\n)') {
            $c = $c -replace '(?ms)(import \(\r?\n)', "`${1}`t`"github.com/valerii/hub/backend/internal/search`"`r`n"
        }
    }

    if ($c -ne $orig) {
        Set-Content -Path $_.FullName -Value $c -Encoding UTF8 -NoNewline
        Write-Host "Fixed: $($_.Name)"
    }
}

Write-Host "Done."
