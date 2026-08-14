[CmdletBinding()]
param()

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

# Static source contract only: no child process, Docker, test, or target runs.
$script:HubQaG0RunnerV2Status = 'NOT_READY/NOT_RUN'

function Read-HubQaStrictUtf8 {
    param([Parameter(Mandatory = $true)][string]$Path)
    if (-not [IO.File]::Exists($Path)) { throw "Missing static-contract input: $Path" }
    [byte[]]$bytes = [IO.File]::ReadAllBytes($Path)
    if ($bytes.Length -eq 0 -or ($bytes.Length -ge 3 -and $bytes[0] -eq 0xef -and $bytes[1] -eq 0xbb -and $bytes[2] -eq 0xbf)) { throw "Empty or BOM-prefixed input: $Path" }
    try { return ([Text.UTF8Encoding]::new($false, $true)).GetString($bytes) }
    catch { throw "Input is not strict UTF-8: $Path" }
}

function Read-HubQaPowerShellSyntax {
    param([Parameter(Mandatory = $true)][string]$Path)
    $tokens = $null
    $errors = $null
    $ast = [Management.Automation.Language.Parser]::ParseFile($Path, [ref]$tokens, [ref]$errors)
    if (@($errors).Count -ne 0) { throw "PowerShell source is not parseable: $Path" }
    return [pscustomobject][ordered]@{ Path = $Path; Ast = $ast; Tokens = @($tokens) }
}

function Get-HubQaFunctionAst {
    param([object]$Syntax, [string]$Name)
    $matches = @($Syntax.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] -and $node.Name -ceq $Name }, $true))
    if ($matches.Count -ne 1) { throw "Expected one active function $Name in $($Syntax.Path)." }
    return $matches[0]
}

function Get-HubQaAssignments {
    param([Management.Automation.Language.Ast]$Ast, [string]$Variable)
    return @($Ast.FindAll({
        param($node)
        if ($node -isnot [Management.Automation.Language.AssignmentStatementAst]) { return $false }
        $variables = if ($node.Left -is [Management.Automation.Language.VariableExpressionAst]) { @($node.Left) } else { @($node.Left.FindAll({ param($child) $child -is [Management.Automation.Language.VariableExpressionAst] }, $true)) }
        return $variables.Count -eq 1 -and ('$' + $variables[0].VariablePath.UserPath) -ceq $Variable
    }, $true))
}

function Get-HubQaSingleAssignment {
    param([Management.Automation.Language.Ast]$Ast, [string]$Variable)
    $matches = @(Get-HubQaAssignments -Ast $Ast -Variable $Variable)
    if ($matches.Count -ne 1) { throw "Expected one active assignment to $Variable." }
    return $matches[0]
}

function Get-HubQaLiteralStrings {
    param([Management.Automation.Language.Ast]$Ast)
    return @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.StringConstantExpressionAst] }, $true) | ForEach-Object { [string]$_.Value })
}

function Assert-HubQaExactSequence {
    param([object[]]$Actual, [object[]]$Expected, [string]$Identity)
    if ($Actual.Count -ne $Expected.Count) { throw "$Identity has the wrong cardinality: $($Actual.Count), expected $($Expected.Count)." }
    for ($index = 0; $index -lt $Expected.Count; $index++) {
        if ([string]$Actual[$index] -cne [string]$Expected[$index]) { throw "$Identity is reordered or changed at index $index." }
    }
}

function Assert-HubQaAssignedString {
    param([object]$Syntax, [string]$Variable, [string]$Expected)
    $assignment = Get-HubQaSingleAssignment -Ast $Syntax.Ast -Variable $Variable
    if ([string]$assignment.Right.SafeGetValue() -cne $Expected) { throw "$Variable does not equal its frozen value." }
}

function Assert-HubQaAssignedInteger {
    param([object]$Syntax, [string]$Variable, [int64]$Expected)
    $assignment = Get-HubQaSingleAssignment -Ast $Syntax.Ast -Variable $Variable
    if ([int64]$assignment.Right.SafeGetValue() -ne $Expected) { throw "$Variable does not equal its frozen value." }
}

function Assert-HubQaAssignmentShape {
    param([Management.Automation.Language.Ast]$Ast, [string]$Variable, [string]$Pattern, [string]$Identity)
    $assignment = Get-HubQaSingleAssignment -Ast $Ast -Variable $Variable
    if ((($assignment.Right.Extent.Text -replace '\s+', ' ').Trim()) -cnotmatch $Pattern) { throw "$Identity has the wrong active assignment shape." }
}

function Assert-HubQaAssignedStringArray {
    param([object]$Syntax, [string]$Variable, [string[]]$Expected)
    $assignment = Get-HubQaSingleAssignment -Ast $Syntax.Ast -Variable $Variable
    Assert-HubQaExactSequence -Actual @(Get-HubQaLiteralStrings -Ast $assignment.Right) -Expected $Expected -Identity $Variable
}

function Get-HubQaOuterHashtableKeys {
    param([Management.Automation.Language.Ast]$Ast, [string]$Identity)
    $tables = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.HashtableAst] }, $true) | Sort-Object { -($_.Extent.EndOffset - $_.Extent.StartOffset) })
    if ($tables.Count -lt 1) { throw "$Identity has no active hashtable." }
    return @($tables[0].KeyValuePairs | ForEach-Object { [string]$_.Item1.SafeGetValue() })
}

function Assert-HubQaHashtableAssignmentKeys {
    param([Management.Automation.Language.Ast]$Ast, [string]$Variable, [string[]]$Expected, [string]$Identity)
    $assignment = Get-HubQaSingleAssignment -Ast $Ast -Variable $Variable
    Assert-HubQaExactSequence -Actual @(Get-HubQaOuterHashtableKeys -Ast $assignment.Right -Identity $Identity) -Expected $Expected -Identity $Identity
}

function Assert-HubQaHashtableLiteralValue {
    param([Management.Automation.Language.Ast]$Ast, [string]$Variable, [string]$Key, [object]$Expected, [string]$Identity)
    $assignment = Get-HubQaSingleAssignment -Ast $Ast -Variable $Variable
    $tables = @($assignment.Right.FindAll({ param($node) $node -is [Management.Automation.Language.HashtableAst] }, $true) | Sort-Object { -($_.Extent.EndOffset - $_.Extent.StartOffset) })
    if ($tables.Count -lt 1) { throw "$Identity has no active hashtable." }
    $pairs = @($tables[0].KeyValuePairs | Where-Object { [string]$_.Item1.SafeGetValue() -ceq $Key })
    if ($pairs.Count -ne 1) { throw "$Identity does not assign key $Key exactly once." }
    if (@($pairs[0].Item2.Statements).Count -ne 1) { throw "$Identity does not assign one literal statement to $Key." }
    $expression = $pairs[0].Item2.Statements[0].GetPureExpression()
    if ($null -eq $expression -or $expression.SafeGetValue() -cne $Expected) { throw "$Identity assigns the wrong literal to $Key." }
}

function Assert-HubQaHashtableArrayFieldSequence {
    param([Management.Automation.Language.Ast]$Ast, [string]$Variable, [string]$Key, [string[]]$Expected, [string]$Identity)
    $assignment = Get-HubQaSingleAssignment -Ast $Ast -Variable $Variable
    $actual = New-Object 'System.Collections.Generic.List[string]'
    foreach ($table in @($assignment.Right.FindAll({ param($node) $node -is [Management.Automation.Language.HashtableAst] }, $true) | Sort-Object { $_.Extent.StartOffset })) {
        foreach ($pair in $table.KeyValuePairs) {
            if ([string]$pair.Item1.SafeGetValue() -cne $Key -or @($pair.Item2.Statements).Count -ne 1) { continue }
            $expression = $pair.Item2.Statements[0].GetPureExpression()
            if ($null -eq $expression) { throw "$Identity has a non-literal $Key value." }
            [void]$actual.Add([string]$expression.SafeGetValue())
        }
    }
    Assert-HubQaExactSequence -Actual @($actual.ToArray()) -Expected $Expected -Identity $Identity
}

function Get-HubQaCommandAsts {
    param([Management.Automation.Language.Ast]$Ast, [string]$Name)
    return @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] -and $node.GetCommandName() -ceq $Name }, $true))
}

function Assert-HubQaFunctionCalls {
    param([Management.Automation.Language.Ast]$Ast, [string[]]$Expected, [string]$Identity)
    foreach ($name in $Expected) {
        if (@(Get-HubQaCommandAsts -Ast $Ast -Name $name).Count -lt 1) { throw "$Identity does not actively call $name." }
    }
}

function Assert-HubQaCommandOrder {
    param([Management.Automation.Language.Ast]$Ast, [string[]]$Expected, [string]$Identity)
    $commands = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] }, $true) | Sort-Object { $_.Extent.StartOffset })
    $cursor = -1
    foreach ($name in $Expected) {
        $next = @($commands | Where-Object { $_.Extent.StartOffset -gt $cursor -and $_.GetCommandName() -ceq $name } | Select-Object -First 1)
        if ($next.Count -ne 1) { throw "$Identity omits or reorders active call $name." }
        $cursor = $next[0].Extent.StartOffset
    }
}

function Assert-HubQaCommandShape {
    param([Management.Automation.Language.Ast]$Ast, [string]$Name, [string]$Pattern, [string]$Identity)
    $matches = @(Get-HubQaCommandAsts -Ast $Ast -Name $Name | Where-Object { (($_.Extent.Text -replace '\s+', ' ').Trim()) -cmatch $Pattern })
    if ($matches.Count -lt 1) { throw "$Identity is missing the active $Name invocation shape." }
}

function Assert-HubQaCommandLiteralSequence {
    param([Management.Automation.Language.Ast]$Ast, [string]$Name, [string[]]$Expected, [string]$Identity)
    $matches = @(Get-HubQaCommandAsts -Ast $Ast -Name $Name | Where-Object {
        $actual = @(Get-HubQaLiteralStrings -Ast $_)
        if ($actual.Count -gt 0 -and $actual[0] -ceq $Name) { $actual = @($actual | Select-Object -Skip 1) }
        if ($actual.Count -ne $Expected.Count) { return $false }
        for ($i = 0; $i -lt $Expected.Count; $i++) { if ($actual[$i] -cne $Expected[$i]) { return $false } }
        return $true
    })
    if ($matches.Count -lt 1) { throw "$Identity is missing the active exact literal sequence." }
}

function Get-HubQaContainingStatement {
    param([Management.Automation.Language.Ast]$Node, [string]$Identity)
    $statement = $Node
    while ($null -ne $statement.Parent -and $statement.Parent -isnot [Management.Automation.Language.StatementBlockAst]) { $statement = $statement.Parent }
    if ($statement.Parent -isnot [Management.Automation.Language.StatementBlockAst]) { throw "$Identity is not contained by an active statement block." }
    return $statement
}

function Test-HubQaStatementIsExactCommand {
    param([Management.Automation.Language.Ast]$Statement, [string]$Name)
    $commands = @($Statement.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] }, $true))
    return $commands.Count -eq 1 -and $commands[0].GetCommandName() -ceq $Name
}

function Assert-HubQaCommandsGuardedBeforeAfter {
    param([Management.Automation.Language.Ast]$Ast, [string]$Operation, [string]$Guard, [string]$Identity)
    $calls = @(Get-HubQaCommandAsts -Ast $Ast -Name $Operation)
    if ($calls.Count -lt 1) { throw "$Identity has no active $Operation call." }
    foreach ($call in $calls) {
        $statement = Get-HubQaContainingStatement -Node $call -Identity $Identity
        $statements = @($statement.Parent.Statements)
        $index = -1
        for ($candidate = 0; $candidate -lt $statements.Count; $candidate++) {
            if ($statements[$candidate].Extent.StartOffset -eq $statement.Extent.StartOffset) { $index = $candidate; break }
        }
        if ($index -lt 1 -or $index -ge ($statements.Count - 1) -or -not (Test-HubQaStatementIsExactCommand -Statement $statements[$index - 1] -Name $Guard) -or -not (Test-HubQaStatementIsExactCommand -Statement $statements[$index + 1] -Name $Guard)) {
            throw "$Identity does not guard every active $Operation immediately before and after."
        }
    }
}

function Assert-HubQaMembersGuardedBeforeAfter {
    param([Management.Automation.Language.Ast]$Ast, [string]$Member, [string]$ExpressionPattern, [string]$Guard, [string]$Identity)
    $calls = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true) | Where-Object {
        $memberName = try { [string]$_.Member.SafeGetValue() } catch { '' }
        $memberName -ceq $Member -and (($_.Expression.Extent.Text -replace '\s+', ' ').Trim()) -cmatch $ExpressionPattern
    })
    if ($calls.Count -lt 1) { throw "$Identity has no active $Member member call." }
    foreach ($call in $calls) {
        $statement = Get-HubQaContainingStatement -Node $call -Identity $Identity
        $statements = @($statement.Parent.Statements)
        $index = -1
        for ($candidate = 0; $candidate -lt $statements.Count; $candidate++) {
            if ($statements[$candidate].Extent.StartOffset -eq $statement.Extent.StartOffset) { $index = $candidate; break }
        }
        if ($index -lt 1 -or $index -ge ($statements.Count - 1) -or -not (Test-HubQaStatementIsExactCommand -Statement $statements[$index - 1] -Name $Guard) -or -not (Test-HubQaStatementIsExactCommand -Statement $statements[$index + 1] -Name $Guard)) {
            throw "$Identity does not guard every active $Member immediately before and after."
        }
    }
}

function Assert-HubQaDeletesGuardedByFinally {
    param([Management.Automation.Language.Ast]$Ast, [string]$Guard, [string]$Identity)
    $deletes = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true) | Where-Object {
        $memberName = try { [string]$_.Member.SafeGetValue() } catch { '' }
        $memberName -ceq 'Delete' -and (($_.Expression.Extent.Text -replace '\s+', ' ').Trim()) -cmatch '^\[System\.IO\.(File|Directory)\]$'
    })
    if (@($deletes | Where-Object { $_.Expression.Extent.Text -cmatch '\.File\]' }).Count -ne 1 -or @($deletes | Where-Object { $_.Expression.Extent.Text -cmatch '\.Directory\]' }).Count -ne 2) { throw "$Identity does not own the exact file/directory delete surface." }
    foreach ($delete in $deletes) {
        $owner = $delete.Parent
        while ($null -ne $owner -and $owner -isnot [Management.Automation.Language.TryStatementAst]) { $owner = $owner.Parent }
        if ($owner -isnot [Management.Automation.Language.TryStatementAst] -or $null -eq $owner.Finally -or @(Get-HubQaCommandAsts -Ast $owner.Finally -Name $Guard).Count -lt 1) { throw "$Identity has a delete without an unconditional post-delete deadline guard." }
        $statements = @($owner.Parent.Statements)
        $index = -1
        for ($candidate = 0; $candidate -lt $statements.Count; $candidate++) {
            if ($statements[$candidate].Extent.StartOffset -eq $owner.Extent.StartOffset) { $index = $candidate; break }
        }
        if ($index -lt 1 -or -not (Test-HubQaStatementIsExactCommand -Statement $statements[$index - 1] -Name $Guard)) { throw "$Identity has a delete without an immediate pre-delete deadline guard." }
    }
}

function Assert-HubQaForEachDeadlineFence {
    param([Management.Automation.Language.Ast]$Ast, [string]$Variable, [string]$Collection, [string]$Guard, [string]$Identity)
    $loops = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.ForEachStatementAst] }, $true) | Where-Object {
        ('$' + $_.Variable.VariablePath.UserPath) -ceq $Variable -and (($_.Condition.Extent.Text -replace '\s+', ' ').Trim()) -ceq $Collection
    })
    if ($loops.Count -ne 1) { throw "$Identity loop is missing or ambiguous." }
    $loop = $loops[0]
    $statements = @($loop.Parent.Statements)
    $index = -1
    for ($candidate = 0; $candidate -lt $statements.Count; $candidate++) {
        if ($statements[$candidate].Extent.StartOffset -eq $loop.Extent.StartOffset) { $index = $candidate; break }
    }
    if ($index -lt 1 -or -not (Test-HubQaStatementIsExactCommand -Statement $statements[$index - 1] -Name $Guard)) { throw "$Identity has no deadline guard before enumeration." }
    $bodyStatements = @($loop.Body.Statements)
    if ($bodyStatements.Count -lt 2 -or -not (Test-HubQaStatementIsExactCommand -Statement $bodyStatements[0] -Name $Guard)) { throw "$Identity has no deadline guard immediately after enumeration." }
    $terminal = $bodyStatements[$bodyStatements.Count - 1]
    $terminalGuarded = Test-HubQaStatementIsExactCommand -Statement $terminal -Name $Guard
    if (-not $terminalGuarded -and $terminal -is [Management.Automation.Language.TryStatementAst] -and $null -ne $terminal.Finally) {
        $terminalGuarded = @(Get-HubQaCommandAsts -Ast $terminal.Finally -Name $Guard).Count -ge 1
    }
    if (-not $terminalGuarded) { throw "$Identity has no deadline guard on its terminal iteration path." }
}

function Get-HubQaThrowCodes {
    param([Management.Automation.Language.Ast]$Ast)
    return @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.ThrowStatementAst] }, $true) | ForEach-Object {
        $expression = $_.Pipeline.GetPureExpression()
        if ($null -ne $expression) { try { [string]$expression.SafeGetValue() } catch { $null } }
    } | Where-Object { $null -ne $_ })
}

function Assert-HubQaThrowCodes {
    param([Management.Automation.Language.Ast]$Ast, [string[]]$Expected, [string]$Identity)
    $actual = @(Get-HubQaThrowCodes -Ast $Ast)
    foreach ($code in $Expected) { if ($actual -cnotcontains $code) { throw "$Identity does not actively throw $code." } }
}

function Assert-HubQaThrowPrefix {
    param([Management.Automation.Language.Ast]$Ast, [string]$Expected, [string]$Identity)
    if (@(Get-HubQaThrowCodes -Ast $Ast | Where-Object { $_.StartsWith($Expected, [StringComparison]::Ordinal) }).Count -lt 1) {
        throw "$Identity does not actively throw the $Expected failure family."
    }
}

function Assert-HubQaConditionalLiteral {
    param([Management.Automation.Language.Ast]$Ast, [string]$Expected, [string]$Identity)
    $matches = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.IfStatementAst] }, $true) | Where-Object {
        $values = @(Get-HubQaLiteralStrings -Ast $_)
        $values -ccontains $Expected
    })
    if ($matches.Count -lt 1) { throw "$Identity does not actively branch on $Expected." }
}

function Assert-HubQaConditionalInteger {
    param([Management.Automation.Language.Ast]$Ast, [int64]$Expected, [string]$Identity)
    $matches = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.IfStatementAst] }, $true) | Where-Object {
        $found = $false
        foreach ($clause in $_.Clauses) {
            if (@($clause.Item1.FindAll({
                param($child)
                if ($child -isnot [Management.Automation.Language.ConstantExpressionAst]) { return $false }
                try { return [int64]$child.Value -eq $Expected } catch { return $false }
            }, $true)).Count -gt 0) { $found = $true }
        }
        $found
    })
    if ($matches.Count -lt 1) { throw "$Identity does not actively branch on integer $Expected." }
}

function Assert-HubQaIfConditionShape {
    param([Management.Automation.Language.Ast]$Ast, [string]$Pattern, [string]$Identity)
    $conditions = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.IfStatementAst] }, $true) | ForEach-Object {
        foreach ($clause in $_.Clauses) { (($clause.Item1.Extent.Text -replace '\s+', ' ').Trim()) }
    })
    if (@($conditions | Where-Object { $_ -cmatch $Pattern }).Count -lt 1) { throw "$Identity has no matching active if-condition." }
}

function Assert-HubQaMemberInvocations {
    param([Management.Automation.Language.Ast]$Ast, [string[]]$Expected, [string]$Identity)
    $actual = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true) | ForEach-Object {
        try { [string]$_.Member.SafeGetValue() } catch { $null }
    } | Where-Object { $null -ne $_ })
    foreach ($member in $Expected) { if ($actual -cnotcontains $member) { throw "$Identity does not actively invoke $member." } }
}

function Assert-HubQaMemberShape {
    param([Management.Automation.Language.Ast]$Ast, [string]$Name, [string]$Pattern, [string]$Identity)
    $matches = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true) | Where-Object {
        $memberName = try { [string]$_.Member.SafeGetValue() } catch { '' }
        $memberName -ceq $Name -and (($_.Extent.Text -replace '\s+', ' ').Trim()) -cmatch $Pattern
    })
    if ($matches.Count -lt 1) { throw "$Identity has no matching active $Name invocation." }
}

function Assert-HubQaBoundedOverrideMap {
    param([Management.Automation.Language.Ast]$Ast, [System.Collections.IDictionary]$Expected)
    $assignments = @(Get-HubQaAssignments -Ast $Ast -Variable '$boundedOverrides' | Where-Object { @($_.Right.FindAll({ param($node) $node -is [Management.Automation.Language.HashtableAst] }, $true)).Count -gt 0 })
    if ($assignments.Count -ne 1) { throw 'Bounded override map is missing or ambiguous.' }
    $assignment = $assignments[0]
    $outerTables = @($assignment.Right.FindAll({ param($node) $node -is [Management.Automation.Language.HashtableAst] }, $true) | Sort-Object { -($_.Extent.EndOffset - $_.Extent.StartOffset) })
    if ($outerTables.Count -lt 1) { throw 'Bounded override map has no active outer hashtable.' }
    $outer = $outerTables[0]
    Assert-HubQaExactSequence -Actual @($outer.KeyValuePairs | ForEach-Object { [string]$_.Item1.SafeGetValue() }) -Expected @($Expected.Keys) -Identity 'bounded private-source override names'
    foreach ($outerPair in $outer.KeyValuePairs) {
        $overrideName = [string]$outerPair.Item1.SafeGetValue()
        $innerTables = @($outerPair.Item2.FindAll({ param($node) $node -is [Management.Automation.Language.HashtableAst] }, $true) | Sort-Object { -($_.Extent.EndOffset - $_.Extent.StartOffset) })
        if ($innerTables.Count -ne 1) { throw "Bounded override $overrideName is not one closed active record." }
        $inner = $innerTables[0]
        Assert-HubQaExactSequence -Actual @($inner.KeyValuePairs | ForEach-Object { [string]$_.Item1.SafeGetValue() }) -Expected @('RequiredCall','Body') -Identity "bounded override $overrideName"
        $requiredPair = @($inner.KeyValuePairs | Where-Object { [string]$_.Item1.SafeGetValue() -ceq 'RequiredCall' })[0]
        $requiredExpression = $requiredPair.Item2.Statements[0].GetPureExpression()
        if ($null -eq $requiredExpression -or [string]$requiredExpression.SafeGetValue() -cne [string]$Expected[$overrideName]) { throw "Bounded override $overrideName points at the wrong B-owned implementation." }
        $bodyPair = @($inner.KeyValuePairs | Where-Object { [string]$_.Item1.SafeGetValue() -ceq 'Body' })[0]
        $bodyExpression = $bodyPair.Item2.Statements[0].GetPureExpression()
        if ($bodyExpression -isnot [Management.Automation.Language.ScriptBlockExpressionAst]) { throw "Bounded override $overrideName has no active script-block body." }
        $bodyCommands = @($bodyExpression.ScriptBlock.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] }, $true))
        if ($bodyCommands.Count -ne 1 -or $bodyCommands[0].GetCommandName() -cne [string]$Expected[$overrideName]) { throw "Bounded override $overrideName is not an exact one-call B-owned wrapper." }
    }
    return $assignment
}

function Assert-HubQaNoUnboundedWaitOrRead {
    param([object]$Syntax)
    foreach ($member in @($Syntax.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true))) {
        $name = try { [string]$member.Member.SafeGetValue() } catch { '' }
        if ($name -ceq 'ReadToEnd') { throw 'Package B retains an unbounded synchronous stream read.' }
        if ($name -ceq 'WaitForExit' -and @($member.Arguments).Count -ne 1) { throw 'Package B retains an unbounded WaitForExit call.' }
    }
}

function Assert-HubQaNoParentOnlyKill {
    param([Management.Automation.Language.Ast]$Ast, [string]$Identity)
    foreach ($member in @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true))) {
        $name = try { [string]$member.Member.SafeGetValue() } catch { '' }
        if ($name -ceq 'Kill' -and @($member.Arguments).Count -ne 1) { throw "$Identity retains a parent-only Kill fallback." }
    }
}

function Assert-HubQaCumulativeCleanupBudgetScope {
    param([Management.Automation.Language.Ast]$Ast, [string]$Identity)
    Assert-HubQaAssignmentShape $Ast '$script:HubQaBootstrapCleanupBudgetStopwatch' '\[System\.Diagnostics\.Stopwatch\]::new\(\)' "$Identity shared stopwatch construction"
    Assert-HubQaIfConditionShape $Ast '\$null -eq \$script:HubQaBootstrapCleanupBudgetStopwatch' "$Identity lazy shared stopwatch"
    Assert-HubQaIfConditionShape $Ast '-not \$script:HubQaBootstrapCleanupBudgetStopwatch\.IsRunning' "$Identity restartable shared stopwatch"
    $starts = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true) | Where-Object {
        $memberName = try { [string]$_.Member.SafeGetValue() } catch { '' }
        $memberName -ceq 'Start' -and $_.Extent.Text -cmatch '\$script:HubQaBootstrapCleanupBudgetStopwatch\.Start\(\)'
    })
    $stops = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true) | Where-Object {
        $memberName = try { [string]$_.Member.SafeGetValue() } catch { '' }
        $memberName -ceq 'Stop' -and $_.Extent.Text -cmatch '\$script:HubQaBootstrapCleanupBudgetStopwatch\.Stop\(\)'
    })
    if ($starts.Count -ne 1 -or $stops.Count -ne 1 -or $starts[0].Extent.StartOffset -ge $stops[0].Extent.StartOffset) { throw "$Identity does not start then stop the shared cleanup budget exactly once." }
    $finallyOwners = @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.TryStatementAst] -and $null -ne $node.Finally }, $true) | Where-Object {
        @($_.Finally.FindAll({ param($child) $child -is [Management.Automation.Language.InvokeMemberExpressionAst] -and $child.Extent.Text -cmatch '\$script:HubQaBootstrapCleanupBudgetStopwatch\.Stop\(\)' }, $true)).Count -eq 1
    })
    if ($finallyOwners.Count -ne 1) { throw "$Identity does not stop the shared cleanup budget in one unconditional finally." }
}

function Get-HubQaIfStatementsByCondition {
    param([Management.Automation.Language.Ast]$Ast, [string]$Pattern)
    return @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.IfStatementAst] }, $true) | Where-Object {
        @($_.Clauses | Where-Object { (($_.Item1.Extent.Text -replace '\s+', ' ').Trim()) -cmatch $Pattern }).Count -eq 1
    })
}

function Assert-HubQaTimeoutRethrowCatchCount {
    param([Management.Automation.Language.Ast]$Ast, [int]$Expected, [string]$Identity)
    $matches = @()
    foreach ($catchClause in @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.CatchClauseAst] }, $true))) {
        $timeoutBranches = @(Get-HubQaIfStatementsByCondition -Ast $catchClause.Body -Pattern '^\[string\]\$_\.Exception\.Message -ceq ''E_B0_TIMEOUT_124''$')
        if ($timeoutBranches.Count -eq 0) { continue }
        if ($timeoutBranches.Count -ne 1) { throw "$Identity has an ambiguous timeout rethrow branch." }
        $timeoutBranch = $timeoutBranches[0]
        if ($timeoutBranch.Clauses.Count -ne 1 -or $null -ne $timeoutBranch.ElseClause) { throw "$Identity timeout rethrow is not one closed if branch." }
        $bareThrows = @($timeoutBranch.Clauses[0].Item2.FindAll({ param($node) $node -is [Management.Automation.Language.ThrowStatementAst] -and $node.Extent.Text.Trim() -ceq 'throw' }, $true))
        if ($bareThrows.Count -ne 1 -or @($timeoutBranch.Clauses[0].Item2.Statements).Count -ne 1 -or $timeoutBranch.Clauses[0].Item2.Statements[0] -isnot [Management.Automation.Language.ThrowStatementAst]) { throw "$Identity does not immediately preserve timeout identity with one active bare rethrow." }
        $matches += $catchClause
    }
    if ($matches.Count -ne $Expected) { throw "$Identity has $($matches.Count) timeout-preserving catches, expected $Expected." }
}

function Assert-HubQaStoredTimeoutRethrowCatchCount {
    param([Management.Automation.Language.Ast]$Ast, [int]$Expected, [string]$Guard, [string]$Identity)
    $matches = @()
    foreach ($catchClause in @($Ast.FindAll({ param($node) $node -is [Management.Automation.Language.CatchClauseAst] }, $true))) {
        $timeoutBranches = @(Get-HubQaIfStatementsByCondition -Ast $catchClause.Body -Pattern '^\$cleanupError -ceq ''E_B0_TIMEOUT_124''$')
        if ($timeoutBranches.Count -eq 0) { continue }
        if ($timeoutBranches.Count -ne 1) { throw "$Identity has an ambiguous stored-timeout rethrow branch." }
        $assignments = @(Get-HubQaAssignments -Ast $catchClause.Body -Variable '$cleanupError')
        $guards = @(Get-HubQaCommandAsts -Ast $catchClause.Body -Name $Guard)
        $timeoutBranch = $timeoutBranches[0]
        $bareThrows = @($timeoutBranch.Clauses[0].Item2.FindAll({ param($node) $node -is [Management.Automation.Language.ThrowStatementAst] -and $node.Extent.Text.Trim() -ceq 'throw' }, $true))
        if ($assignments.Count -ne 1 -or ((($assignments[0].Right.Extent.Text -replace '\s+', ' ').Trim()) -cne '[string]$_.Exception.Message') -or $guards.Count -ne 1 -or $timeoutBranch.Clauses.Count -ne 1 -or $null -ne $timeoutBranch.ElseClause -or $bareThrows.Count -ne 1 -or @($timeoutBranch.Clauses[0].Item2.Statements).Count -ne 1) { throw "$Identity stored timeout identity is not a closed assignment/guard/bare-rethrow path." }
        if ($assignments[0].Extent.StartOffset -ge $guards[0].Extent.StartOffset -or $guards[0].Extent.StartOffset -ge $timeoutBranch.Extent.StartOffset) { throw "$Identity does not preserve stored timeout identity across the combined deadline guard." }
        $matches += $catchClause
    }
    if ($matches.Count -ne $Expected) { throw "$Identity has $($matches.Count) stored-timeout catches, expected $Expected." }
}

function Assert-HubQaFirstExitCatchMapping {
    param(
        [Management.Automation.Language.StatementBlockAst]$Body,
        [int]$NonTimeoutExit,
        [bool]$Rethrows,
        [string]$Identity
    )
    $guards = @(Get-HubQaIfStatementsByCondition -Ast $Body -Pattern '^\$firstExit -eq 0$')
    if ($guards.Count -ne 1) { throw "$Identity does not have one active first-exit guard." }
    $guard = $guards[0]
    if ($guard.Clauses.Count -ne 1 -or $null -ne $guard.ElseClause) { throw "$Identity first-exit guard is not closed." }
    $timeoutBranches = @(Get-HubQaIfStatementsByCondition -Ast $guard.Clauses[0].Item2 -Pattern '^\[string\]\$_\.Exception\.Message -ceq ''E_B0_TIMEOUT_124''$')
    if ($timeoutBranches.Count -ne 1) { throw "$Identity does not have one active timeout-identity branch." }
    $timeoutBranch = $timeoutBranches[0]
    if ($timeoutBranch.Clauses.Count -ne 1 -or $null -eq $timeoutBranch.ElseClause) { throw "$Identity timeout mapping is not an exact if/else." }
    $timeoutAssignments = @(Get-HubQaAssignments -Ast $timeoutBranch.Clauses[0].Item2 -Variable '$firstExit')
    $otherAssignments = @(Get-HubQaAssignments -Ast $timeoutBranch.ElseClause -Variable '$firstExit')
    if ($timeoutAssignments.Count -ne 1 -or $otherAssignments.Count -ne 1) { throw "$Identity timeout mapping is not singular." }
    $timeoutValue = try { [int]$timeoutAssignments[0].Right.SafeGetValue() } catch { -1 }
    $otherValue = try { [int]$otherAssignments[0].Right.SafeGetValue() } catch { -1 }
    if ($timeoutValue -ne 124 -or $otherValue -ne $NonTimeoutExit) { throw "$Identity does not map timeout/general failure to 124/$NonTimeoutExit." }
    $allAssignments = @(Get-HubQaAssignments -Ast $guard.Clauses[0].Item2 -Variable '$firstExit')
    if ($allAssignments.Count -ne 2) { throw "$Identity has an extra firstExit mutation inside its guard." }
    $bareThrows = @($Body.FindAll({ param($node) $node -is [Management.Automation.Language.ThrowStatementAst] -and $node.Extent.Text.Trim() -ceq 'throw' }, $true))
    if (($Rethrows -and $bareThrows.Count -ne 1) -or (-not $Rethrows -and $bareThrows.Count -ne 0)) { throw "$Identity has the wrong active rethrow policy." }
}

function Assert-HubQaExitValues {
    param([object]$Syntax, [int[]]$Expected, [string]$Identity)
    $actual = @($Syntax.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.ExitStatementAst] }, $true) | ForEach-Object {
        $expression = $_.Pipeline.GetPureExpression()
        if ($null -ne $expression) { [int]$expression.SafeGetValue() }
    })
    foreach ($value in $Expected) { if ($actual -cnotcontains $value) { throw "$Identity does not actively exit $value." } }
}

function Read-HubQaJsonObject {
    param([string]$Path)
    $text = Read-HubQaStrictUtf8 -Path $Path
    try { $value = $text | ConvertFrom-Json -ErrorAction Stop } catch { throw "Invalid JSON: $Path" }
    return [pscustomobject][ordered]@{ Text = $text; Value = $value }
}

function Assert-HubQaJsonKeys {
    param([object]$Value, [string[]]$Expected, [string]$Identity)
    Assert-HubQaExactSequence -Actual @($Value.PSObject.Properties.Name) -Expected $Expected -Identity $Identity
}

function Get-HubQaActiveTextLines {
    param([string]$Path)
    return @((Read-HubQaStrictUtf8 -Path $Path) -split "`r?`n" | ForEach-Object { $_.Trim() } | Where-Object { $_.Length -gt 0 -and -not $_.StartsWith('#') -and -not $_.StartsWith('//') })
}

function Assert-HubQaActiveLine {
    param([string[]]$Lines, [string]$Pattern, [string]$Identity)
    if (@($Lines | Where-Object { $_ -cmatch $Pattern }).Count -lt 1) { throw "$Identity has no matching active line." }
}

function Assert-HubQaActiveLineSequence {
    param([string[]]$Lines, [string[]]$Patterns, [string]$Identity)
    $cursor = -1
    foreach ($pattern in $Patterns) {
        $next = -1
        for ($lineIndex = $cursor + 1; $lineIndex -lt $Lines.Count; $lineIndex++) {
            if ($Lines[$lineIndex] -cmatch $pattern) { $next = $lineIndex; break }
        }
        if ($next -lt 0) { throw "$Identity omits or reorders active native source pattern $pattern." }
        $cursor = $next
    }
}

function Get-HubQaActiveAddTypeSourceLines {
    param([object]$Syntax, [string]$ClassName)
    $addTypeCommands = @(Get-HubQaCommandAsts -Ast $Syntax.Ast -Name 'Add-Type' | Where-Object { $_.Extent.Text -cmatch '-TypeDefinition' })
    if ($addTypeCommands.Count -ne 1) { throw 'Package B must compile exactly one native process/file source.' }
    $sourceNodes = @($addTypeCommands[0].FindAll({ param($node) $node -is [Management.Automation.Language.StringConstantExpressionAst] -and $node.StringConstantType -eq [Management.Automation.Language.StringConstantType]::SingleQuotedHereString }, $true) | Where-Object { $_.Value -cmatch ('\bclass\s+' + [regex]::Escape($ClassName) + '\b') })
    if ($sourceNodes.Count -ne 1) { throw "Package B does not actively compile the unique $ClassName native source." }
    return @(([string]$sourceNodes[0].Value -split "`r?`n") | ForEach-Object { $_.Trim() } | Where-Object { $_.Length -gt 0 -and -not $_.StartsWith('//') })
}

$repositoryRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..\..\..'))
$path = [ordered]@{
    BootstrapContract = Join-Path $repositoryRoot 'testframework\runner\contracts\hub.qa-bootstrap.v1.json'
    Bootstrap = Join-Path $repositoryRoot 'testframework\runner\invoke-qa-bootstrap-v1.ps1'
    PublicGate = Join-Path $repositoryRoot 'testframework\runner\invoke-qa-gate-v2.ps1'
    Runner = Join-Path $repositoryRoot 'testframework\runner\run-task-v2.ps1'
    Snapshot = Join-Path $repositoryRoot 'testframework\runner\new-repository-snapshot.ps1'
    Envelope = Join-Path $repositoryRoot 'testframework\runner\new-run-envelope.ps1'
    Activation = Join-Path $repositoryRoot 'testframework\runner\write-g0-activation.ps1'
    SnapshotSchema = Join-Path $repositoryRoot 'testframework\schemas\repository-snapshot.schema.v1.json'
    EnvelopeSchema = Join-Path $repositoryRoot 'testframework\schemas\run-envelope.schema.v1.json'
    AdmissionLauncher = Join-Path $repositoryRoot 'testframework\docker\admit-local-dependency-assets.ps1'
    AdmissionModule = Join-Path $repositoryRoot 'testframework\docker\admission\admit_local_dependency_assets.py'
    Attestor = Join-Path $repositoryRoot 'testframework\docker\source-image-attestor.ps1'
    BackendDockerfile = Join-Path $repositoryRoot 'testframework\docker\Dockerfile.backend-unit'
    FrontendDockerfile = Join-Path $repositoryRoot 'testframework\docker\Dockerfile.frontend-unit'
    OpenApiDockerfile = Join-Path $repositoryRoot 'testframework\docker\Dockerfile.backend-openapi'
}

$ps = [ordered]@{
    Bootstrap = Read-HubQaPowerShellSyntax -Path $path.Bootstrap
    PublicGate = Read-HubQaPowerShellSyntax -Path $path.PublicGate
    Runner = Read-HubQaPowerShellSyntax -Path $path.Runner
    Snapshot = Read-HubQaPowerShellSyntax -Path $path.Snapshot
    Envelope = Read-HubQaPowerShellSyntax -Path $path.Envelope
    Activation = Read-HubQaPowerShellSyntax -Path $path.Activation
    AdmissionLauncher = Read-HubQaPowerShellSyntax -Path $path.AdmissionLauncher
    Attestor = Read-HubQaPowerShellSyntax -Path $path.Attestor
}
$bootstrapContract = Read-HubQaJsonObject -Path $path.BootstrapContract
$snapshotSchema = Read-HubQaJsonObject -Path $path.SnapshotSchema
$envelopeSchema = Read-HubQaJsonObject -Path $path.EnvelopeSchema

$sourceCommit = '61b38573682d7c01e6d9381babc5085537d7990f'
$a01Commit = '56a288dd51e261bbb91ed693d7b58a8a97994faf'
$a02Paths = @(
    'testframework/docker/dependency-bundle.schema.v1.json', 'testframework/docker/admission/verify_local_dependency_bundle.py',
    'testframework/docker/admit-local-dependency-assets.ps1', 'testframework/docker/entrypoints/backend-unit.sh',
    'testframework/docker/vitest.frontend-unit.config.ts', 'testframework/docker/bootstrap-control.v1.json',
    'testframework/docker/source-image-attestor.ps1', 'testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1',
    'testframework/selftests/runner/g0-image-contract.ps1'
)
$a02OverlapPaths = @(
    'testframework/docker/dependency-bundle.schema.v1.json', 'testframework/docker/admission/verify_local_dependency_bundle.py',
    'testframework/docker/admit-local-dependency-assets.ps1', 'testframework/docker/bootstrap-control.v1.json',
    'testframework/docker/source-image-attestor.ps1', 'testframework/selftests/runner/g0-bootstrap-attestor-contract.ps1',
    'testframework/selftests/runner/g0-image-contract.ps1'
)
$a01RemainingPaths = @(
    'testframework/docker/admission/admit_local_dependency_assets.py', 'testframework/docker/Dockerfile.backend-unit',
    'testframework/docker/Dockerfile.frontend-unit', 'testframework/docker/Dockerfile.backend-openapi',
    'testframework/docker/entrypoints/frontend-unit.sh', 'testframework/docker/runner-fixtures/g0-planning-selftest.py'
)
$a1AssetKeys = @('asset_id','targets','relative_path','sha256','bytes','archive_format','member_count','expanded_bytes','layout_sha256','license_spdx','license_evidence_sha256','provenance_sha256')
$cReceiptKeys = @('nonce_consumption','owner_marker','local_base_inspect','helper_base_verification','helper_image_attestation','ingress_copy','helper_admission','helper_cleanup','source_tree_verification','local_image_inspection','candidate_lock','planning_selftest_negative_red','planning_selftest_positive_green','command_receipts','run','cleanup')
$fixtureNames = @('noncanonical-bootstrap-json','wrong-version-or-producer','short-or-foreign-run','serialized-role-claim','direct-planning-selftest','release-nonce-consumption','successful-run-receipt-binding','snapshot-envelope-drift','missing-prebuild-source-receipt','partial-postbuild-attestation','postbuild-base-receipt-binding','d-evidence-acceptance-contract','activation-r-c1-d-chain','first-exit-cleanup')

Assert-HubQaAssignedString -Syntax $ps.Bootstrap -Variable '$script:HubQaBootstrapSourceCommit' -Expected $sourceCommit
Assert-HubQaAssignedString -Syntax $ps.Activation -Variable '$script:HubQaActivationSourceCommit' -Expected $sourceCommit
Assert-HubQaAssignedString -Syntax $ps.Activation -Variable '$script:HubQaActivationA0Commit' -Expected $a01Commit
Assert-HubQaAssignedString -Syntax $ps.Snapshot -Variable '$script:HubQaSnapshotSourceCommit' -Expected $sourceCommit
Assert-HubQaAssignedString -Syntax $ps.Attestor -Variable '$script:SourceBaseline' -Expected $sourceCommit
Assert-HubQaAssignedStringArray -Syntax $ps.Bootstrap -Variable '$script:HubQaBootstrapA02Paths' -Expected $a02Paths
Assert-HubQaAssignedStringArray -Syntax $ps.Bootstrap -Variable '$script:HubQaBootstrapA02OverlapPaths' -Expected $a02OverlapPaths
Assert-HubQaAssignedStringArray -Syntax $ps.Bootstrap -Variable '$script:HubQaBootstrapA01RemainingPaths' -Expected $a01RemainingPaths
Assert-HubQaAssignedStringArray -Syntax $ps.Activation -Variable '$script:HubQaActivationA02Paths' -Expected $a02Paths
Assert-HubQaAssignedStringArray -Syntax $ps.Activation -Variable '$script:HubQaActivationA02OverlapPaths' -Expected $a02OverlapPaths
Assert-HubQaAssignedStringArray -Syntax $ps.Activation -Variable '$script:HubQaActivationA01RemainingPaths' -Expected $a01RemainingPaths
Assert-HubQaAssignedStringArray -Syntax $ps.Bootstrap -Variable '$script:HubQaBootstrapCReceiptKeys' -Expected $cReceiptKeys
Assert-HubQaAssignedStringArray -Syntax $ps.Bootstrap -Variable '$script:HubQaBootstrapFixedTargets' -Expected @('backend-unit','frontend-unit','backend-openapi','g0-planning-selftest')
Assert-HubQaAssignedInteger -Syntax $ps.Bootstrap -Variable '$script:HubQaBootstrapInspectSeconds' -Expected 30
Assert-HubQaAssignedInteger -Syntax $ps.Bootstrap -Variable '$script:HubQaBootstrapCommandSeconds' -Expected 180
Assert-HubQaAssignedInteger -Syntax $ps.Bootstrap -Variable '$script:HubQaBootstrapTargetSeconds' -Expected 780
Assert-HubQaAssignedInteger -Syntax $ps.Bootstrap -Variable '$script:HubQaBootstrapOuterSeconds' -Expected 840
Assert-HubQaAssignedInteger -Syntax $ps.Bootstrap -Variable '$script:HubQaBootstrapCleanupSeconds' -Expected 30
Assert-HubQaAssignedInteger -Syntax $ps.Bootstrap -Variable '$script:HubQaBootstrapOrchestrationSlackSeconds' -Expected 30
Assert-HubQaNoUnboundedWaitOrRead -Syntax $ps.Bootstrap
Assert-HubQaNoUnboundedWaitOrRead -Syntax $ps.Snapshot
Assert-HubQaNoUnboundedWaitOrRead -Syntax $ps.Activation
Assert-HubQaNoParentOnlyKill -Ast $ps.Bootstrap.Ast -Identity 'Package-B termination path'
$bootstrapKillCalls = @($ps.Bootstrap.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true) | Where-Object { try { [string]$_.Member.SafeGetValue() -ceq 'Kill' } catch { $false } })
if ($bootstrapKillCalls.Count -ne 0) { throw 'Package B has an active Process.Kill fallback instead of whole-Job termination.' }
$directBootstrapExecutables = @($ps.Bootstrap.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] }, $true) | Where-Object { $_.GetCommandName() -cin @('git','docker','docker.exe') })
if ($directBootstrapExecutables.Count -ne 0) { throw 'Package B has an active direct Git/Docker command outside the bounded process runner.' }
$directBootstrapProcesses = @(Get-HubQaCommandAsts -Ast $ps.Bootstrap.Ast -Name 'New-Object' | Where-Object { @(Get-HubQaLiteralStrings -Ast $_) -cmatch '^(System\.)?Diagnostics\.Process$' })
if ($directBootstrapProcesses.Count -ne 0) { throw 'Package B has an active direct managed-process constructor outside its Job Object owner.' }
$nativeProcessLines = @(Get-HubQaActiveAddTypeSourceLines -Syntax $ps.Bootstrap -ClassName 'HubQaG0AssignedProcess')
foreach ($nativePattern in @(
    '^private const UInt32 CREATE_SUSPENDED = 0x00000004;$',
    '^private const UInt32 JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE = 0x00002000;$',
    '^private static extern IntPtr CreateJobObjectW\(',
    '^private static extern bool SetInformationJobObject\(',
    '^private static extern bool AssignProcessToJobObject\(',
    '^private static extern UInt32 ResumeThread\(',
    '^private static extern bool TerminateJobObject\('
)) { Assert-HubQaActiveLine -Lines $nativeProcessLines -Pattern $nativePattern -Identity 'native Job Object API' }
Assert-HubQaActiveLineSequence -Lines $nativeProcessLines -Patterns @(
    '^job = CreateJobObjectW\(IntPtr\.Zero, null\);$',
    '^limits\.BasicLimitInformation\.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE;$',
    '^if \(!SetInformationJobObject\(',
    '^if \(!CreateProcessW\(.*CREATE_SUSPENDED \| CREATE_NO_WINDOW,.*\)$',
    '^if \(!AssignProcessToJobObject\(job, information\.hProcess\)\)$',
    '^if \(ResumeThread\(information\.hThread\) == UInt32\.MaxValue\)$'
) -Identity 'suspended-create assign-before-resume Job Object sequence'
Assert-HubQaActiveLine -Lines $nativeProcessLines -Pattern '^if \(_job == IntPtr\.Zero \|\| !TerminateJobObject\(_job, exitCode\)\)$' -Identity 'runtime whole-job termination'
Assert-HubQaActiveLine -Lines $nativeProcessLines -Pattern '^else if \(created && information\.hProcess != IntPtr\.Zero\) TerminateProcess\(information\.hProcess, 124\);$' -Identity 'never-resumed setup rollback'
if (@($nativeProcessLines | Where-Object { $_ -cmatch '\.Kill\s*\(' -or $_ -cmatch '\bProcess\.Start\s*\(' -or $_ -cmatch '\.WaitForExit\s*\(\s*\)' -or $_ -cmatch '\.ReadToEnd\s*\(' }).Count -ne 0) { throw 'Native Package-B process owner contains an unbounded or unassigned process fallback.' }

Assert-HubQaJsonKeys -Value $bootstrapContract.Value -Expected @('schema','contract','status','task','run_pattern','targets','internal_role','nonce_consumption','report_root','owner_marker','base_receipts','source_tree_gate','post_build_attestation','run_receipt','release_authority','cleanup_contract','direct_invocation') -Identity 'bootstrap contract'
if ($bootstrapContract.Value.schema -cne 'hub.qa.bootstrap-contract/v1' -or $bootstrapContract.Value.contract -cne 'hub.qa-bootstrap/v1' -or $bootstrapContract.Value.status -cne 'NOT_READY') { throw 'Bootstrap contract identity drifted.' }
Assert-HubQaJsonKeys -Value $bootstrapContract.Value.nonce_consumption -Expected @('schema','root','path_template','keys','status','release_id_pattern','failure_code') -Identity 'bootstrap nonce contract'
Assert-HubQaExactSequence -Actual @($bootstrapContract.Value.nonce_consumption.keys) -Expected @('schema','task','run','release_id','release_commit','release_raw_sha256','one_shot_nonce','consumed_at_utc','status') -Identity 'nonce receipt keys'
if ([int]$bootstrapContract.Value.direct_invocation.exit_code -ne 64 -or [bool]$bootstrapContract.Value.direct_invocation.public_bootstrap -or [bool]$bootstrapContract.Value.direct_invocation.public_g0_planning_selftest) { throw 'Bootstrap direct invocation contract drifted.' }
if ($snapshotSchema.Value.'$id' -cne 'hub.qa.repository-snapshot/v1' -or [bool]$snapshotSchema.Value.additionalProperties) { throw 'Snapshot schema drifted.' }
if ($envelopeSchema.Value.'$id' -cne 'hub.qa.run-envelope/v1' -or [bool]$envelopeSchema.Value.additionalProperties) { throw 'Envelope schema drifted.' }
$limits = $envelopeSchema.Value.properties.limits.properties
if ($limits.command_seconds.const -ne 180 -or $limits.target_seconds.const -ne 780 -or $limits.inspect_seconds.const -ne 30 -or $limits.outer_seconds.const -ne 840 -or $limits.cleanup_seconds.const -ne 30 -or $limits.orchestration_slack_seconds.const -ne 30) { throw 'Run envelope limits drifted.' }
Assert-HubQaExactSequence -Actual @($envelopeSchema.Value.properties.limits.required) -Expected @('command_seconds','target_seconds','inspect_seconds','outer_seconds','cleanup_seconds','orchestration_slack_seconds') -Identity 'required envelope limits'

$bootstrapMain = Get-HubQaFunctionAst $ps.Bootstrap 'Invoke-HubQaBootstrapV1'
$releaseAuthority = Get-HubQaFunctionAst $ps.Bootstrap 'Assert-HubQaBootstrapReleaseAuthority'
$privateSourceImporter = Get-HubQaFunctionAst $ps.Bootstrap 'Import-HubQaBootstrapPrivateSources'
$privatePrimitiveContract = Get-HubQaFunctionAst $ps.Bootstrap 'Assert-HubQaBootstrapPrivateProcessPrimitiveContract'
$a02Authority = Get-HubQaFunctionAst $ps.Bootstrap 'Assert-HubQaBootstrapA02Authority'
$commitPaths = Get-HubQaFunctionAst $ps.Bootstrap 'Get-HubQaBootstrapCommitPaths'
$currentSources = Get-HubQaFunctionAst $ps.Bootstrap 'Assert-HubQaBootstrapCurrentAcceptedSources'
$authoritySelector = Get-HubQaFunctionAst $ps.Bootstrap 'Get-HubQaBootstrapAuthorityCommitForPath'
$a1Asset = Get-HubQaFunctionAst $ps.Bootstrap 'Assert-HubQaBootstrapA1Asset'
$deadline = Get-HubQaFunctionAst $ps.Bootstrap 'Assert-HubQaBootstrapDeadline'
$outerDeadline = Get-HubQaFunctionAst $ps.Bootstrap 'Assert-HubQaBootstrapOuterDeadline'
$cleanupAndOuterDeadline = Get-HubQaFunctionAst $ps.Bootstrap 'Assert-HubQaBootstrapCleanupAndOuterDeadline'
$timeout = Get-HubQaFunctionAst $ps.Bootstrap 'Get-HubQaBootstrapEffectiveTimeoutMilliseconds'
$process = Get-HubQaFunctionAst $ps.Bootstrap 'Invoke-HubQaBootstrapProcess'
$processStart = Get-HubQaFunctionAst $ps.Bootstrap 'Start-HubQaBootstrapAssignedProcess'
$processStop = Get-HubQaFunctionAst $ps.Bootstrap 'Stop-HubQaBootstrapProcessJobWithinDeadline'
$git = Get-HubQaFunctionAst $ps.Bootstrap 'Invoke-HubQaBootstrapGit'
$gitObject = Get-HubQaFunctionAst $ps.Bootstrap 'Get-HubQaBootstrapGitObjectBytes'
$attestorGitBlob = Get-HubQaFunctionAst $ps.Bootstrap 'Get-HubQaBootstrapAttestorGitBlobBytes'
$attestorGitTree = Get-HubQaFunctionAst $ps.Bootstrap 'Get-HubQaBootstrapAttestorGitTreeEntry'
$attestorSourceTree = Get-HubQaFunctionAst $ps.Bootstrap 'Get-HubQaBootstrapAttestorGitObjectSourceTree'
$boundedAdmission = Get-HubQaFunctionAst $ps.Bootstrap 'Invoke-HubQaBootstrapBoundedG0AdmissionFromBootstrapContext'
$safeFileOpen = Get-HubQaFunctionAst $ps.Bootstrap 'Open-HubQaBootstrapSafeRegularFile'
$docker = Get-HubQaFunctionAst $ps.Bootstrap 'Invoke-HubQaBootstrapDocker'
$dockerProbe = Get-HubQaFunctionAst $ps.Bootstrap 'Invoke-HubQaBootstrapDockerProbe'
$fourBuilds = Get-HubQaFunctionAst $ps.Bootstrap 'Invoke-HubQaBootstrapFourFixedImageBuilds'
$imageInspection = Get-HubQaFunctionAst $ps.Bootstrap 'New-HubQaBootstrapLocalImageInspectionReceipt'
$ownedReportTreeCleanup = Get-HubQaFunctionAst $ps.Bootstrap 'Remove-HubQaBootstrapOwnedReportTree'
$exactCleanup = Get-HubQaFunctionAst $ps.Bootstrap 'Invoke-HubQaBootstrapExactCleanup'
$helperCleanup = Get-HubQaFunctionAst $ps.Bootstrap 'Invoke-HubQaBootstrapHelperCleanup'
$nonceParser = Get-HubQaFunctionAst $ps.Bootstrap 'ConvertFrom-HubQaBootstrapNonceConsumption'
$nonceWriter = Get-HubQaFunctionAst $ps.Bootstrap 'New-HubQaBootstrapNonceConsumptionReceipt'
$commandWriter = Get-HubQaFunctionAst $ps.Bootstrap 'Write-HubQaBootstrapCommandReceipts'
$cChainValidator = Get-HubQaFunctionAst $ps.Bootstrap 'Assert-HubQaBootstrapCReceiptChain'
$cChainWriter = Get-HubQaFunctionAst $ps.Bootstrap 'New-HubQaBootstrapCReceiptChain'
$runWriter = Get-HubQaFunctionAst $ps.Bootstrap 'Write-HubQaBootstrapRunReceipt'
$sourceTreeWriter = Get-HubQaFunctionAst $ps.Bootstrap 'New-HubQaBootstrapSourceTreeReceipt'
$postBuild = Get-HubQaFunctionAst $ps.Bootstrap 'Invoke-HubQaBootstrapPostBuildAttestation'
$bootstrapAdmission = Get-HubQaFunctionAst $ps.Bootstrap 'Invoke-HubQaBootstrapAdmission'
$snapshotGitEvidenceWriter = Get-HubQaFunctionAst $ps.Bootstrap 'New-HubQaBootstrapSnapshotGitEvidence'
$snapshotContextWriter = Get-HubQaFunctionAst $ps.Bootstrap 'New-HubQaBootstrapSnapshotPrivateContext'
$boundedGitBlobDigest = Get-HubQaFunctionAst $ps.Bootstrap 'Get-HubQaBootstrapBoundedGitBlobDigest'
$snapshotIndexRecords = Get-HubQaFunctionAst $ps.Bootstrap 'Get-HubQaBootstrapIndexPathRecords'
$snapshotWorktreeRecords = Get-HubQaFunctionAst $ps.Bootstrap 'Get-HubQaBootstrapWorktreePathRecords'
$closedDockerClientPath = Get-HubQaFunctionAst $ps.Bootstrap 'Get-HubQaBootstrapClosedDockerClientPath'
$runnerPublic = Get-HubQaFunctionAst $ps.Runner 'Invoke-HubQaRunnerV2PublicRequest'
$runnerContext = Get-HubQaFunctionAst $ps.Runner 'Test-HubQaRunnerV2ExactContext'
$snapshotWriter = Get-HubQaFunctionAst $ps.Snapshot 'New-HubQaRepositorySnapshot'
$envelopeWriter = Get-HubQaFunctionAst $ps.Envelope 'New-HubQaRunEnvelope'
$activationA02 = Get-HubQaFunctionAst $ps.Activation 'Assert-HubQaActivationA02Authority'
$activationRelease = Get-HubQaFunctionAst $ps.Activation 'Get-HubQaActivationRelease'
$activationD = Get-HubQaFunctionAst $ps.Activation 'Get-HubQaActivationDRecord'
$activationWriter = Get-HubQaFunctionAst $ps.Activation 'Write-HubQaG0Activation'
$attestorDependency = Get-HubQaFunctionAst $ps.Attestor 'Assert-HubQaDependencyBundleContract'
$attestorSourceTreeOriginal = Get-HubQaFunctionAst $ps.Attestor 'Get-HubQaGitObjectSourceTree'
$attestorImageOriginal = Get-HubQaFunctionAst $ps.Attestor 'Invoke-HubQaBootstrapImageAttestation'
$admissionDependency = Get-HubQaFunctionAst $ps.AdmissionLauncher 'Assert-HubQaPrivateDependencyBundleManifest'
$admissionOriginal = Get-HubQaFunctionAst $ps.AdmissionLauncher 'Invoke-HubQaG0AdmissionFromBootstrapContext'

Assert-HubQaCommandShape $a02Authority.Body 'Invoke-HubQaBootstrapGit' '''rev-list'', ''--first-parent'', \$A02Commit' 'A0.2 A0.1 first-parent membership'
Assert-HubQaCommandShape $a02Authority.Body 'Get-HubQaBootstrapActualFirstParent' '-Commit \$A02Commit' 'A0.2 actual parent'
Assert-HubQaCommandShape $a02Authority.Body 'Invoke-HubQaBootstrapGit' '''rev-list'', ''--first-parent'', \$actualParent' 'A0.2 source lineage'
Assert-HubQaCommandShape $a02Authority.Body 'Assert-HubQaBootstrapExactCommitPaths' '-Commit \$A02Commit -Expected \$script:HubQaBootstrapA02Paths' 'A0.2 nine-path diff'
Assert-HubQaFunctionCalls $commitPaths.Body @('Get-HubQaBootstrapActualFirstParent','Invoke-HubQaBootstrapGit') 'actual-parent diff implementation'
Assert-HubQaCommandShape $commitPaths.Body 'Invoke-HubQaBootstrapGit' '''diff-tree''.*\$parent, \$Commit' 'actual-parent diff implementation'
Assert-HubQaCommandShape $currentSources.Body 'Assert-HubQaBootstrapCurrentPackageBytesAndModes' 'a02_commit.*HubQaBootstrapA02OverlapPaths.*SkipExactCommitScope' 'A0.2 seven current blobs'
Assert-HubQaCommandShape $currentSources.Body 'Assert-HubQaBootstrapCurrentPackageBytesAndModes' 'a0_commit.*HubQaBootstrapA01RemainingPaths.*SkipExactCommitScope' 'A0.1 six current blobs'
Assert-HubQaIfConditionShape $authoritySelector.Body '\$script:HubQaBootstrapA02Paths -ccontains \$Path' 'A0.2 path authority selector'
$a02Returns = @($authoritySelector.Body.FindAll({ param($node) $node -is [Management.Automation.Language.ReturnStatementAst] -and $node.Pipeline.Extent.Text -cmatch 'Release\.Value\.a02_commit' }, $true))
if ($a02Returns.Count -ne 1) { throw 'A0.2 paths do not resolve to a02_commit authority.' }
Assert-HubQaFunctionCalls $releaseAuthority.Body @('Assert-HubQaBootstrapA02Authority') 'release A0.2 authority chain'
Assert-HubQaFunctionCalls $privateSourceImporter.Body @('Assert-HubQaBootstrapCurrentAcceptedSources') 'accepted current-source chain'
Assert-HubQaFunctionCalls $bootstrapMain.Body @('Import-HubQaBootstrapPrivateSources') 'bootstrap private-source import chain'
Assert-HubQaHashtableArrayFieldSequence $privateSourceImporter.Body '$sourceFiles' 'RelativePath' @('testframework/docker/source-image-attestor.ps1','testframework/docker/admit-local-dependency-assets.ps1','testframework/runner/run-task-v2.ps1','testframework/runner/new-run-envelope.ps1','testframework/runner/new-repository-snapshot.ps1') 'accepted private-source list'
Assert-HubQaCommandOrder $privateSourceImporter.Body @('Get-HubQaBootstrapAuthorityCommitForPath','Get-HubQaBootstrapGitObjectBytes','Open-HubQaBootstrapSafeRegularFile','Read-HubQaBootstrapHeldFileBytes','Test-HubQaBootstrapBytesEqual','Assert-HubQaBootstrapPrivateProcessPrimitiveContract') 'accepted held-source byte import'
Assert-HubQaConditionalLiteral $privatePrimitiveContract.Body 'testframework/docker/source-image-attestor.ps1' 'accepted A0.2 attestor primitive identity'
Assert-HubQaConditionalLiteral $privatePrimitiveContract.Body 'testframework/docker/admit-local-dependency-assets.ps1' 'accepted A0.1 admission primitive identity'
Assert-HubQaMemberInvocations $privatePrimitiveContract.Body @('IndexOf') 'accepted historical primitive drift check'
Assert-HubQaThrowCodes $privatePrimitiveContract.Body @('E_B0_PRIVATE_SOURCE') 'accepted historical primitive drift check'
$dotSources = @($bootstrapMain.Body.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] -and $node.InvocationOperator -eq [Management.Automation.Language.TokenKind]::Dot }, $true) | Where-Object { $_.Extent.Text -cmatch '\$privateSource\.Path' })
if ($dotSources.Count -ne 2) { throw 'Bootstrap does not dot-source only the accepted held private-source path.' }
$boundedOverrideCalls = [ordered]@{
    'Get-HubQaDockerClientPath' = 'Get-HubQaBootstrapClosedDockerClientPath'
    'Get-HubQaGitBlobBytes' = 'Get-HubQaBootstrapAttestorGitBlobBytes'
    'Get-HubQaGitTreeEntry' = 'Get-HubQaBootstrapAttestorGitTreeEntry'
    'Get-HubQaGitObjectSourceTree' = 'Get-HubQaBootstrapAttestorGitObjectSourceTree'
    'Invoke-HubQaG0AdmissionFromBootstrapContext' = 'Invoke-HubQaBootstrapBoundedG0AdmissionFromBootstrapContext'
}
$boundedOverrideAssignment = Assert-HubQaBoundedOverrideMap -Ast $bootstrapMain.Body -Expected $boundedOverrideCalls
$importCommand = @(Get-HubQaCommandAsts -Ast $bootstrapMain.Body -Name 'Import-HubQaBootstrapPrivateSources')
$setOverride = @(Get-HubQaCommandAsts -Ast $bootstrapMain.Body -Name 'Set-Item')
$verifyOverride = @(Get-HubQaCommandAsts -Ast $bootstrapMain.Body -Name 'Get-Command')
if ($importCommand.Count -ne 1 -or $setOverride.Count -ne 1 -or $verifyOverride.Count -ne 1) { throw 'Private-source import/override/verification is not singular.' }
Assert-HubQaCommandShape $bootstrapMain.Body 'Set-Item' 'Function:\\.*\+ \$overrideName.*-Value \$boundedOverrides\[\$overrideName\]\.Body -Force' 'bounded private-source override installation'
Assert-HubQaCommandShape $bootstrapMain.Body 'Get-Command' '\$overrideName -CommandType Function -ErrorAction Stop' 'bounded private-source override verification'
Assert-HubQaIfConditionShape $bootstrapMain.Body '\$activeBody\.IndexOf\(\$requiredCall.*Ordinal\) -lt 0' 'bounded private-source active-body verification'
$overrideVerificationIfs = @($bootstrapMain.Body.FindAll({ param($node) $node -is [Management.Automation.Language.IfStatementAst] }, $true) | Where-Object {
    $_.Extent.Text -cmatch '\$activeBody\.IndexOf\(\$requiredCall' -and @(Get-HubQaThrowCodes -Ast $_) -ccontains 'E_B0_PRIVATE_SOURCE'
})
if ($overrideVerificationIfs.Count -ne 1) { throw 'Active bounded override verification is missing or ambiguous.' }
$postImportReads = @(Get-HubQaCommandAsts -Ast $bootstrapMain.Body -Name 'Read-HubQaBootstrapHeldFileBytes' | Where-Object { $_.Extent.Text -cmatch '-Opened \$privateSource' })
if ($postImportReads.Count -ne 1) { throw 'Private-source bytes are not rechecked exactly once after import.' }
Assert-HubQaIfConditionShape $bootstrapMain.Body '\$postImport\.Sha256 -cne \[string\]\$privateSource\.AcceptedSha256' 'post-import held-byte digest check'
$lastDotSourceOffset = ($dotSources | Sort-Object { $_.Extent.EndOffset } | Select-Object -Last 1).Extent.EndOffset
if ($importCommand[0].Extent.EndOffset -ge $lastDotSourceOffset -or $lastDotSourceOffset -ge $postImportReads[0].Extent.StartOffset -or
    $postImportReads[0].Extent.EndOffset -ge $boundedOverrideAssignment.Extent.StartOffset -or
    $boundedOverrideAssignment.Extent.EndOffset -ge $setOverride[0].Extent.StartOffset -or $setOverride[0].Extent.EndOffset -ge $verifyOverride[0].Extent.StartOffset -or
    $verifyOverride[0].Extent.EndOffset -ge $overrideVerificationIfs[0].Extent.StartOffset) {
    throw 'Private-source import, replacement definitions, installation, and active verification are reordered.'
}
$importedProcessNames = @(
    @($ps.Attestor.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] }, $true) | ForEach-Object { $_.Name }) +
    @($ps.AdmissionLauncher.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] }, $true) | ForEach-Object { $_.Name })
)
$directImportedCalls = @($bootstrapMain.Body.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] }, $true) | Where-Object { $importedProcessNames -ccontains $_.GetCommandName() })
$attestorEntryCalls = @($directImportedCalls + @(Get-HubQaCommandAsts -Ast $bootstrapMain.Body -Name 'Invoke-HubQaBootstrapAdmission') + @(Get-HubQaCommandAsts -Ast $bootstrapMain.Body -Name 'Invoke-HubQaBootstrapPostBuildAttestation'))
if ($attestorEntryCalls.Count -lt 3 -or @($attestorEntryCalls | Where-Object { $_.Extent.StartOffset -le $overrideVerificationIfs[0].Extent.EndOffset }).Count -ne 0) {
    throw 'An imported attestor/admission call is reachable before bounded override verification.'
}
$firstAttestorEntry = @($attestorEntryCalls | Sort-Object { $_.Extent.StartOffset } | Select-Object -First 1)
if ($firstAttestorEntry.Count -ne 1 -or $firstAttestorEntry[0].GetCommandName() -cne 'Invoke-HubQaBootstrapAdmission') { throw 'The first post-replacement private process entry is not bounded admission.' }
Assert-HubQaFunctionCalls $bootstrapAdmission.Body @('Invoke-HubQaBootstrapBoundedG0AdmissionFromBootstrapContext') 'bounded admission call graph'
Assert-HubQaFunctionCalls $closedDockerClientPath.Body @('Open-HubQaBootstrapSafeRegularFile','Close-HubQaBootstrapSafeRegularFile') 'closed Docker client path replacement'
Assert-HubQaFunctionCalls $postBuild.Body @('Invoke-HubQaBootstrapImageAttestation') 'post-build attestor override call graph'
Assert-HubQaFunctionCalls $attestorImageOriginal.Body @('Get-HubQaGitObjectSourceTree') 'accepted A0.2 attestor call graph'
Assert-HubQaFunctionCalls $attestorSourceTreeOriginal.Body @('Get-HubQaGitTreeEntry','Get-HubQaGitBlobBytes') 'accepted A0.2 Git primitive call graph'
Assert-HubQaFunctionCalls $admissionOriginal.Body @('Get-HubQaDockerClientPath') 'accepted A0.1 Docker resolver call graph'
$historicalAdmissionCommands = @($admissionOriginal.Body.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] -and $node.InvocationOperator -eq [Management.Automation.Language.TokenKind]::Ampersand }, $true) | Where-Object { $_.Extent.Text -cmatch '\$BootstrapPrivateContext\.PrivateDockerClientPath' })
if ($historicalAdmissionCommands.Count -ne 1) { throw 'Accepted A0.1 admission primitive identity drifted before its bounded replacement.' }
Assert-HubQaFunctionCalls $privateSourceImporter.Body @('Open-HubQaBootstrapSafeRegularFile','Read-HubQaBootstrapHeldFileBytes','Test-HubQaBootstrapBytesEqual','Assert-HubQaBootstrapPrivateProcessPrimitiveContract') 'held exact-byte private-source import'
Assert-HubQaConditionalLiteral $bootstrapMain.Body 'read-only-no-write-no-delete' 'held private-source execution binding'
$noWriteFileOpens = @($safeFileOpen.Body.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true) | Where-Object {
    $memberName = try { [string]$_.Member.SafeGetValue() } catch { '' }
    $shape = (($_.Extent.Text -replace '\s+', ' ').Trim())
    $memberName -ceq 'CreateFileW' -and $shape -cmatch 'GENERIC_READ.*FILE_SHARE_READ.*OPEN_EXISTING.*FILE_FLAG_OPEN_REPARSE_POINT' -and $shape -cnotmatch 'FILE_SHARE_WRITE|FILE_SHARE_DELETE|FILE_FLAG_BACKUP_SEMANTICS'
})
if ($noWriteFileOpens.Count -ne 1) { throw 'Private-source execution is not bound to one exact no-write/no-delete regular-file handle.' }
Assert-HubQaAssignedStringArray ([pscustomobject]@{ Ast = $releaseAuthority.Body }) '$releaseKeys' @('schema','release_id','g0_plan_path','g0_plan_commit','a0_commit','a02_commit','a1_commit','b_commit','task','run','local_only','network','offline_ingress_handoff_id','offline_ingress_owner_marker_sha256','dependency_bundle_manifest_sha256','source_commit','expires_at_utc','one_shot_nonce','founder_approval','created_at_utc')
Assert-HubQaAssignedStringArray ([pscustomobject]@{ Ast = $a1Asset.Body }) '$expected' $a1AssetKeys
Assert-HubQaCommandShape $a1Asset.Body 'Assert-HubQaBootstrapExactKeys' '-Value \$Asset -Expected \$expected' 'A1 exact twelve fields'
Assert-HubQaCommandLiteralSequence $attestorDependency.Body 'Assert-HubQaExactPropertyNames' @('asset_id','targets','relative_path','sha256','bytes','archive_format','member_count','expanded_bytes','layout_sha256','license_spdx','license_evidence_sha256','provenance_sha256') 'attestor A1 exact twelve fields'
Assert-HubQaCommandLiteralSequence $admissionDependency.Body 'Assert-HubQaExactPropertyNames' @('asset_id','targets','relative_path','sha256','bytes','archive_format','member_count','expanded_bytes','layout_sha256','license_spdx','license_evidence_sha256','provenance_sha256','runner-held A1 dependency manifest asset') 'admission A1 exact twelve fields'
Assert-HubQaHashtableAssignmentKeys $bootstrapAdmission.Body '$AdmissionPrivateContext' @('Task','Run','OwnerMarkerSha256','PrivateReportRoot','PrivateIngressRoot','PrivateIngressCopyReceipt','PrivateIngressCopyReceiptSha256','PrivateIngressDigestMap','PrivateIngressDigestMapSha256','PrivateIngressOrderedPaths','PrivateIngressByteCounts','PrivateIngressSourceRoot','PrivateIngressDestinationRoot','PrivateOfflineIngressHandoffId','PrivateOfflineIngressOwnerMarkerSha256','PrivateDependencyBundleManifest','PrivateDependencyBundleManifestSha256','PrivateSourceCommit','PrivateLocalBaseReceipt','PrivateLocalBaseReceiptSha256','PrivateLocalBaseReceiptNonce','PrivateDockerClientPath','PrivateHelperBuildTag','PrivateHelperImageAttestation','PrivateHelperImageAttestationSha256','PrivateHelperLocalImageId','PrivateBootstrapHelperPythonLocalImageId','PrivateBootstrapHelperPythonLocalImageIdSha256') '28-field admission context'
Assert-HubQaHashtableAssignmentKeys $postBuild.Body '$ImageAttestationPrivateContext' @('Task','Run','RoleClaim','OwnerNonce','OwnerMarkerRawBytes','OwnerMarkerSha256','RunnerInspectionRawBytes','RunnerInspectionSha256','LocalBaseReceiptRawBytes','LocalBaseReceiptSha256','LocalBaseReceiptNonce') '11-field image-attestation context'
$snapshotGitEvidenceKeys = @('CandidateLockPath','CandidateLockRawBytes','CandidateLockSha256','G0PlanSha256','BootstrapContractSha256','RunnerContractSha256','DependencyBundleManifestSha256','SourcePolicySha256','BuildManifestSha256','SourceRevisionTree','HeadCommit','HeadTree','IndexPaths','WorktreePaths','Commits','Ancestry')
Assert-HubQaAssignedStringArray ([pscustomobject]@{ Ast = $snapshotContextWriter.Body }) '$evidenceKeys' $snapshotGitEvidenceKeys
Assert-HubQaExactSequence -Actual @(Get-HubQaOuterHashtableKeys -Ast $snapshotGitEvidenceWriter.Body -Identity 'snapshot Git evidence') -Expected $snapshotGitEvidenceKeys -Identity 'snapshot Git evidence'
Assert-HubQaFunctionCalls $snapshotGitEvidenceWriter.Body @('Read-HubQaBootstrapHeldBoundedIngressControl','Get-HubQaBootstrapGitSingle','Invoke-HubQaBootstrapGit','Get-HubQaBootstrapNameStatusEntries','Get-HubQaBootstrapBoundedGitBlobDigest','Get-HubQaBootstrapGitObjectBytes','Get-HubQaBootstrapIndexPathRecords','Get-HubQaBootstrapWorktreePathRecords') 'bounded snapshot Git evidence capture'
Assert-HubQaFunctionCalls $boundedGitBlobDigest.Body @('Get-HubQaBootstrapGitSingle','Get-HubQaBootstrapGitObjectBytes','Get-HubQaBootstrapSha256') 'bounded snapshot blob-digest path'
Assert-HubQaAssignmentShape $snapshotIndexRecords.Body '$objectName' 'HEAD:.*\$entry\.Path.*else.*:.*\$entry\.Path' 'index delete/head and non-delete/index blob selection'
Assert-HubQaCommandShape $snapshotWorktreeRecords.Body 'Get-HubQaBootstrapBoundedGitBlobDigest' '-ObjectName \(.*:.*\+ \$relativePath\)' 'worktree delete/index blob selection'
Assert-HubQaAssignmentShape $snapshotGitEvidenceWriter.Body '$changedObject' '\$parent.*\$entry\.Path.*else.*HEAD:.*\$entry\.Path' 'HEAD commit delete/parent and non-delete/head blob selection'
Assert-HubQaFunctionCalls $snapshotContextWriter.Body @('Assert-HubQaBootstrapExactKeys','Get-HubQaBootstrapSha256') 'snapshot context consumes captured evidence'
$snapshotPrivateContextKeys = @('PositiveGreenPath','PositiveGreenRawBytes','PositiveGreenSha256','CandidateLockPath','CandidateLockRawBytes','CandidateLockSha256','CleanupPath','CleanupRawBytes','CleanupSha256','G0PlanSha256','BootstrapContractSha256','RunnerContractSha256','DependencyBundleManifestSha256','SourcePolicySha256','BuildManifestSha256','SourceRevisionTree','HeadCommit','HeadTree','IndexPaths','WorktreePaths','Commits','Ancestry','StartedAtUtc','FinishedAtUtc')
Assert-HubQaCommandLiteralSequence $snapshotWriter.Body 'Assert-HubQaSnapshotKeys' $snapshotPrivateContextKeys 'snapshot producer private context'
$snapshotProducerFunctionNames = @($ps.Snapshot.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] }, $true) | ForEach-Object { $_.Name } | Where-Object { $_ -cmatch 'Git|Process' })
$snapshotProducerCommands = @($ps.Snapshot.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] }, $true) | Where-Object { $_.GetCommandName() -cmatch '(^git$|Git|Process|Get-Command|Start-Process)' -or $_.InvocationOperator -in @([Management.Automation.Language.TokenKind]::Ampersand,[Management.Automation.Language.TokenKind]::Dot) })
$snapshotProducerProcessTypes = @($ps.Snapshot.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.TypeExpressionAst] -and $node.TypeName.FullName -cmatch 'Diagnostics\.Process' }, $true))
$snapshotProducerProcessMembers = @($ps.Snapshot.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true) | Where-Object { try { [string]$_.Member.SafeGetValue() -cin @('WaitForExit','ReadToEnd','ReadToEndAsync','Kill','Start') } catch { $false } })
if ($snapshotProducerFunctionNames.Count -ne 0 -or $snapshotProducerCommands.Count -ne 0 -or $snapshotProducerProcessTypes.Count -ne 0 -or $snapshotProducerProcessMembers.Count -ne 0) { throw 'Snapshot producer retains a reachable Git/process/dot-source path.' }
$snapshotLocalFunctions = @($ps.Snapshot.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.FunctionDefinitionAst] }, $true) | ForEach-Object { $_.Name })
$snapshotAllowedEnvelopeCalls = @('Set-StrictMode','Get-HubQaEnvelopeSha256','Get-HubQaEnvelopeUtf8','ConvertTo-HubQaEnvelopeCanonicalJson','Get-HubQaEnvelopeRepositoryRoot','Get-HubQaEnvelopeSafeContainedBytes','Open-HubQaEnvelopeHeldDirectoryPath','Assert-HubQaEnvelopeHeldDirectoryChain','Close-HubQaEnvelopeHeldDirectoryChain','ConvertFrom-HubQaEnvelopeCanonicalJson','Assert-HubQaSnapshotEnvelopePayload','Write-HubQaEnvelopeSafeCreateNew')
$snapshotUnexpectedCalls = @($ps.Snapshot.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] }, $true) | Where-Object { $name = $_.GetCommandName(); [string]::IsNullOrWhiteSpace($name) -or ($snapshotLocalFunctions -cnotcontains $name -and $snapshotAllowedEnvelopeCalls -cnotcontains $name) })
if ($snapshotUnexpectedCalls.Count -ne 0) { throw 'Snapshot producer calls outside its exact local/envelope-only allowlist.' }
$snapshotContextUnexpectedCalls = @($snapshotContextWriter.Body.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] }, $true) | Where-Object { $_.GetCommandName() -cnotin @('Assert-HubQaBootstrapExactKeys','Get-HubQaBootstrapSha256') })
if ($snapshotContextUnexpectedCalls.Count -ne 0) { throw 'Post-cleanup snapshot context is not a pure captured-evidence projection.' }

$timeoutVariables = @($timeout.Body.FindAll({ param($node) $node -is [Management.Automation.Language.VariableExpressionAst] }, $true) | ForEach-Object { $_.VariablePath.UserPath })
foreach ($name in @('HubQaBootstrapOuterSeconds','HubQaBootstrapCleanupSeconds','HubQaBootstrapOrchestrationSlackSeconds','HubQaBootstrapTargetSeconds','HubQaBootstrapCleanupBudgetStopwatch')) { if ($timeoutVariables -cnotcontains $name) { throw "Timeout calculator does not actively bind $name." } }
Assert-HubQaThrowCodes $timeout.Body @('E_B0_TIMEOUT_124') 'cumulative timeout calculator'
Assert-HubQaIfConditionShape $deadline.Body '^\$Stopwatch\.Elapsed\.Ticks -ge \(\[System\.TimeSpan\]::FromSeconds\(\$Seconds\)\)\.Ticks$' 'sub-millisecond deadline expiry'
Assert-HubQaThrowCodes $deadline.Body @('E_B0_TIMEOUT_124') 'sub-millisecond deadline expiry'
Assert-HubQaIfConditionShape $outerDeadline.Body '^\$null -eq \$script:HubQaBootstrapRuntimeStopwatch -or -not \$script:HubQaBootstrapRuntimeStopwatch\.IsRunning$' 'outer deadline fail-closed stopwatch'
Assert-HubQaThrowCodes $outerDeadline.Body @('E_B0_TIMEOUT_124') 'outer deadline fail-closed stopwatch'
Assert-HubQaCommandShape $outerDeadline.Body 'Assert-HubQaBootstrapDeadline' '-Stopwatch \$script:HubQaBootstrapRuntimeStopwatch -Seconds \$script:HubQaBootstrapOuterSeconds' 'outer 840-second deadline'
Assert-HubQaIfConditionShape $cleanupAndOuterDeadline.Body '^\$null -eq \$script:HubQaBootstrapCleanupBudgetStopwatch -or -not \$script:HubQaBootstrapCleanupBudgetStopwatch\.IsRunning$' 'combined cleanup deadline fail-closed stopwatch'
Assert-HubQaThrowCodes $cleanupAndOuterDeadline.Body @('E_B0_TIMEOUT_124') 'combined cleanup deadline fail-closed stopwatch'
Assert-HubQaCommandOrder $cleanupAndOuterDeadline.Body @('Assert-HubQaBootstrapOuterDeadline','Assert-HubQaBootstrapDeadline') 'combined outer then cleanup deadline'
Assert-HubQaCommandShape $cleanupAndOuterDeadline.Body 'Assert-HubQaBootstrapDeadline' '-Stopwatch \$script:HubQaBootstrapCleanupBudgetStopwatch -Seconds \$script:HubQaBootstrapCleanupSeconds' 'combined 30-second cleanup deadline'
Assert-HubQaMemberInvocations $process.Body @('CopyToAsync','WaitForExit','Wait') 'bounded process path'
$processCopyTasks = @($process.Body.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true) | Where-Object { try { [string]$_.Member.SafeGetValue() -ceq 'CopyToAsync' } catch { $false } })
if ($processCopyTasks.Count -ne 2) { throw 'Bounded process path does not own exactly two asynchronous output-copy tasks.' }
Assert-HubQaMemberShape $process.Body 'CopyToAsync' '\$stdoutStream\.CopyToAsync\(\$stdoutMemory\)' 'bounded stdout copy'
Assert-HubQaMemberShape $process.Body 'CopyToAsync' '\$stderrStream\.CopyToAsync\(\$stderrMemory\)' 'bounded stderr copy'
Assert-HubQaFunctionCalls $process.Body @('Get-HubQaBootstrapEffectiveTimeoutMilliseconds','Get-HubQaBootstrapOperationRemainingMilliseconds','Start-HubQaBootstrapAssignedProcess','Stop-HubQaBootstrapProcessJobWithinDeadline','Add-HubQaBootstrapCommandReceiptBytes') 'bounded process receipt path'
Assert-HubQaMemberShape $processStart.Body 'Start' '\[HubQaG0AssignedProcess\]::Start\(' 'Job Object process start path'
Assert-HubQaMemberInvocations $processStop.Body @('Terminate','CloseJob','WaitForExit') 'bounded Job Object termination path'
Assert-HubQaCommandShape $git.Body 'Invoke-HubQaBootstrapProcess' '-TimeoutSeconds \$script:HubQaBootstrapInspectSeconds' 'bounded Git inspection'
Assert-HubQaFunctionCalls $gitObject.Body @('Get-HubQaBootstrapEffectiveTimeoutMilliseconds','Get-HubQaBootstrapOperationRemainingMilliseconds','Start-HubQaBootstrapAssignedProcess','Stop-HubQaBootstrapProcessJobWithinDeadline','Add-HubQaBootstrapCommandReceiptBytes') 'bounded Git blob inspection'
Assert-HubQaFunctionCalls $attestorGitBlob.Body @('Get-HubQaBootstrapGitObjectBytes') 'bounded attestor Git blob replacement'
Assert-HubQaFunctionCalls $attestorGitTree.Body @('Invoke-HubQaBootstrapGit') 'bounded attestor Git tree replacement'
Assert-HubQaFunctionCalls $attestorSourceTree.Body @('Invoke-HubQaBootstrapGit','Get-HubQaBootstrapAttestorGitTreeEntry','Get-HubQaBootstrapAttestorGitBlobBytes') 'bounded attestor source-tree replacement'
Assert-HubQaFunctionCalls $boundedAdmission.Body @('Invoke-HubQaBootstrapProcess') 'bounded admission Docker replacement'
Assert-HubQaCommandShape $boundedAdmission.Body 'Invoke-HubQaBootstrapProcess' '-TimeoutSeconds \$script:HubQaBootstrapCommandSeconds.*-CountsAgainstTarget' 'bounded admission Docker replacement'
Assert-HubQaCommandShape $docker.Body 'Invoke-HubQaBootstrapProcess' '-CountsAgainstTarget:\(-not \$CleanupOperation\) -CleanupOperation:\$CleanupOperation' 'cumulative Docker target budget'
Assert-HubQaCommandShape $dockerProbe.Body 'Invoke-HubQaBootstrapProcess' '-TimeoutSeconds \$script:HubQaBootstrapInspectSeconds.*-CountsAgainstTarget:\(-not \$CleanupOperation\)' 'cumulative Docker inspection budget'
Assert-HubQaFunctionCalls $fourBuilds.Body @('Invoke-HubQaBootstrapDocker') 'four fixed target builds'
Assert-HubQaFunctionCalls $imageInspection.Body @('Invoke-HubQaBootstrapDocker') 'four fixed target inspections'
Assert-HubQaFunctionCalls $exactCleanup.Body @('Invoke-HubQaBootstrapDocker','Invoke-HubQaBootstrapDockerProbe') 'bounded cleanup operations'
Assert-HubQaCumulativeCleanupBudgetScope $helperCleanup.Body 'helper cleanup'
Assert-HubQaCumulativeCleanupBudgetScope $exactCleanup.Body 'exact cleanup'
Assert-HubQaTimeoutRethrowCatchCount $helperCleanup.Body 2 'helper cleanup'
Assert-HubQaTimeoutRethrowCatchCount $exactCleanup.Body 3 'exact cleanup'
Assert-HubQaStoredTimeoutRethrowCatchCount $exactCleanup.Body 2 'Assert-HubQaBootstrapCleanupAndOuterDeadline' 'exact cleanup'
$exactCleanupTerminalFinals = @($exactCleanup.Body.FindAll({ param($node) $node -is [Management.Automation.Language.TryStatementAst] -and $null -ne $node.Finally }, $true) | Where-Object {
    @(Get-HubQaCommandAsts -Ast $_.Finally -Name 'Assert-HubQaBootstrapDeadline').Count -eq 1 -and @(Get-HubQaCommandAsts -Ast $_.Finally -Name 'Assert-HubQaBootstrapOuterDeadline').Count -eq 1
})
if ($exactCleanupTerminalFinals.Count -ne 1) { throw 'Exact cleanup lacks one stopped-cleanup plus live-outer terminal fence.' }
$exactCleanupTerminalFinally = $exactCleanupTerminalFinals[0].Finally
Assert-HubQaCommandShape $exactCleanupTerminalFinally 'Assert-HubQaBootstrapDeadline' '-Stopwatch \$script:HubQaBootstrapCleanupBudgetStopwatch -Seconds \$script:HubQaBootstrapCleanupSeconds' 'stopped cumulative cleanup terminal fence'
$sharedCleanupStops = @($exactCleanupTerminalFinally.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true) | Where-Object {
    $memberName = try { [string]$_.Member.SafeGetValue() } catch { '' }
    $memberName -ceq 'Stop' -and (($_.Expression.Extent.Text -replace '\s+', ' ').Trim()) -ceq '$script:HubQaBootstrapCleanupBudgetStopwatch'
})
$stoppedCleanupChecks = @(Get-HubQaCommandAsts -Ast $exactCleanupTerminalFinally -Name 'Assert-HubQaBootstrapDeadline')
$liveOuterChecks = @(Get-HubQaCommandAsts -Ast $exactCleanupTerminalFinally -Name 'Assert-HubQaBootstrapOuterDeadline')
if ($sharedCleanupStops.Count -ne 1 -or $sharedCleanupStops[0].Extent.StartOffset -ge $stoppedCleanupChecks[0].Extent.StartOffset -or $stoppedCleanupChecks[0].Extent.StartOffset -ge $liveOuterChecks[0].Extent.StartOffset) { throw 'Exact cleanup terminal deadline order is not stop -> stopped cleanup check -> live outer check.' }
Assert-HubQaMembersGuardedBeforeAfter $ownedReportTreeCleanup.Body 'GetFiles' '^\$rootDirectory$' 'Assert-HubQaBootstrapCleanupAndOuterDeadline' 'owned report-tree file enumeration'
Assert-HubQaMembersGuardedBeforeAfter $ownedReportTreeCleanup.Body 'GetDirectories' '^\$rootDirectory$' 'Assert-HubQaBootstrapCleanupAndOuterDeadline' 'owned report-tree directory enumeration'
Assert-HubQaForEachDeadlineFence $ownedReportTreeCleanup.Body '$file' '$files' 'Assert-HubQaBootstrapCleanupAndOuterDeadline' 'owned report-tree file loop'
Assert-HubQaForEachDeadlineFence $ownedReportTreeCleanup.Body '$directory' '$directories' 'Assert-HubQaBootstrapCleanupAndOuterDeadline' 'owned report-tree directory loop'
Assert-HubQaDeletesGuardedByFinally $ownedReportTreeCleanup.Body 'Assert-HubQaBootstrapCleanupAndOuterDeadline' 'owned report-tree deletion'
Assert-HubQaForEachDeadlineFence $exactCleanup.Body '$container' '@($script:HubQaBootstrapOwnedContainers.ToArray())' 'Assert-HubQaBootstrapCleanupAndOuterDeadline' 'owned-container cleanup loop'
Assert-HubQaForEachDeadlineFence $exactCleanup.Body '$resource' '@($script:HubQaBootstrapOwnedResources.ToArray())' 'Assert-HubQaBootstrapCleanupAndOuterDeadline' 'owned-image cleanup loop'
foreach ($cleanupOperation in @('Get-HubQaDockerClientPath','Invoke-HubQaBootstrapDockerProbe','Invoke-HubQaBootstrapDocker','Remove-HubQaBootstrapOwnedReportTree','Write-HubQaBootstrapCreateNew')) {
    Assert-HubQaCommandsGuardedBeforeAfter $exactCleanup.Body $cleanupOperation 'Assert-HubQaBootstrapCleanupAndOuterDeadline' "exact cleanup $cleanupOperation"
}
$headBeforeAssignment = Get-HubQaSingleAssignment -Ast $snapshotGitEvidenceWriter.Body -Variable '$headCommit'
$indexCaptureAssignment = Get-HubQaSingleAssignment -Ast $snapshotGitEvidenceWriter.Body -Variable '$indexPaths'
$worktreeCaptureAssignment = Get-HubQaSingleAssignment -Ast $snapshotGitEvidenceWriter.Body -Variable '$worktreePaths'
$headAfterAssignment = Get-HubQaSingleAssignment -Ast $snapshotGitEvidenceWriter.Body -Variable '$headAfterCapture'
Assert-HubQaAssignmentShape $snapshotGitEvidenceWriter.Body '$headCommit' "Get-HubQaBootstrapGitSingle .* -Arguments @\('rev-parse', 'HEAD'\)" 'snapshot HEAD-before capture'
Assert-HubQaAssignmentShape $snapshotGitEvidenceWriter.Body '$headAfterCapture' "Get-HubQaBootstrapGitSingle .* -Arguments @\('rev-parse', 'HEAD'\)" 'snapshot HEAD-after capture'
$headBeforeGuards = @(Get-HubQaIfStatementsByCondition -Ast $snapshotGitEvidenceWriter.Body -Pattern '^\$headCommit -cne \[string\]\$Release\.Commit$')
$headAfterGuards = @(Get-HubQaIfStatementsByCondition -Ast $snapshotGitEvidenceWriter.Body -Pattern '^\$headAfterCapture -cne \$headCommit -or \$headAfterCapture -cne \[string\]\$Release\.Commit$')
if ($headBeforeGuards.Count -ne 1 -or $headAfterGuards.Count -ne 1) { throw 'Snapshot HEAD drift fence is missing or ambiguous.' }
if (-not ($headBeforeAssignment.Extent.StartOffset -lt $headBeforeGuards[0].Extent.StartOffset -and $headBeforeGuards[0].Extent.StartOffset -lt $indexCaptureAssignment.Extent.StartOffset -and $indexCaptureAssignment.Extent.StartOffset -lt $worktreeCaptureAssignment.Extent.StartOffset -and $worktreeCaptureAssignment.Extent.StartOffset -lt $headAfterAssignment.Extent.StartOffset -and $headAfterAssignment.Extent.StartOffset -lt $headAfterGuards[0].Extent.StartOffset)) { throw 'Snapshot HEAD drift fence does not surround ordered index/worktree capture.' }
$targetBudgetStarts = @(Get-HubQaAssignments -Ast $bootstrapMain.Body -Variable '$script:HubQaBootstrapTargetBudgetStopwatch' | Where-Object {
    @($_.Right.FindAll({
        param($node)
        if ($node -isnot [Management.Automation.Language.InvokeMemberExpressionAst]) { return $false }
        try { return [string]$node.Member.SafeGetValue() -ceq 'StartNew' } catch { return $false }
    }, $true)).Count -gt 0
})
if ($targetBudgetStarts.Count -ne 1) { throw 'Bootstrap does not start one cumulative target budget.' }
Assert-HubQaCommandOrder $bootstrapMain.Body @('Assert-HubQaBootstrapReleaseAuthority','New-HubQaBootstrapNonceConsumptionReceipt','New-HubQaBootstrapReportRoot','New-HubQaBootstrapLocalBaseReceipt','New-HubQaBootstrapSourceTreeReceipt','Invoke-HubQaBootstrapFourFixedImageBuilds','New-HubQaBootstrapLocalImageInspectionReceipt','Invoke-HubQaBootstrapPostBuildAttestation','New-HubQaBootstrapSnapshotGitEvidence','Write-HubQaBootstrapCommandReceipts','Invoke-HubQaBootstrapExactCleanup','New-HubQaBootstrapSnapshotPrivateContext','New-HubQaRepositorySnapshot','New-HubQaRunEnvelope','Write-HubQaBootstrapRunReceipt','New-HubQaBootstrapCReceiptChain') 'bootstrap authority and evidence order'
foreach ($finalStage in @('New-HubQaBootstrapSnapshotGitEvidence','Write-HubQaBootstrapCommandReceipts','Invoke-HubQaBootstrapExactCleanup','New-HubQaBootstrapSnapshotPrivateContext','New-HubQaRepositorySnapshot','New-HubQaRunEnvelope','Write-HubQaBootstrapRunReceipt','New-HubQaBootstrapCReceiptChain')) {
    Assert-HubQaCommandsGuardedBeforeAfter $bootstrapMain.Body $finalStage 'Assert-HubQaBootstrapOuterDeadline' "final-stage $finalStage"
}
$snapshotGitCaptureCalls = @(Get-HubQaCommandAsts -Ast $bootstrapMain.Body -Name 'New-HubQaBootstrapSnapshotGitEvidence')
$commandReceiptCalls = @(Get-HubQaCommandAsts -Ast $bootstrapMain.Body -Name 'Write-HubQaBootstrapCommandReceipts')
$snapshotContextCalls = @(Get-HubQaCommandAsts -Ast $bootstrapMain.Body -Name 'New-HubQaBootstrapSnapshotPrivateContext')
if ($snapshotGitCaptureCalls.Count -ne 1 -or $commandReceiptCalls.Count -ne 1 -or $snapshotContextCalls.Count -ne 1) { throw 'Snapshot capture/receipt/pure-context boundary is not singular.' }
$postReceiptGitCalls = @($bootstrapMain.Body.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] }, $true) | Where-Object { $_.Extent.StartOffset -gt $commandReceiptCalls[0].Extent.EndOffset -and $_.GetCommandName() -cmatch 'Git|Process|SnapshotGitEvidence' })
if ($postReceiptGitCalls.Count -ne 0) { throw 'Bootstrap performs Git/process snapshot capture after command receipts are sealed.' }

$planningAssignments = @(Get-HubQaAssignments -Ast $bootstrapMain.Body -Variable '$planningSelftestClosure' | Where-Object { $_.Right -is [Management.Automation.Language.ScriptBlockExpressionAst] })
if ($planningAssignments.Count -ne 1) { throw 'Planning selftest is not one lexical closure.' }
$planningClosure = $planningAssignments[0].Right.ScriptBlock
Assert-HubQaFunctionCalls $planningClosure @('Test-HubQaRunnerV2ExactContext','Invoke-HubQaBootstrapDocker','Write-HubQaRunnerV2CreateNew') 'private planning closure'
Assert-HubQaCommandShape $planningClosure 'Invoke-HubQaBootstrapDocker' "@\('run'.*--hub-qa-target=g0-planning-selftest'\).*HubQaBootstrapCommandSeconds" 'private planning wrapper'
if ($null -ne $planningClosure.ParamBlock -and @($planningClosure.ParamBlock.Parameters).Count -ne 0) { throw 'Planning closure exposes parameters.' }
$runnerDockerCommands = @($ps.Runner.Ast.FindAll({ param($node) $node -is [Management.Automation.Language.CommandAst] -and $node.GetCommandName() -match 'Docker|docker' }, $true))
if ($runnerDockerCommands.Count -ne 0) { throw 'Private runner exports a Docker executor.' }
Assert-HubQaThrowCodes $runnerPublic.Body @('E_B0_TASK_RUN','E_B0_DIRECT_INVOCATION_FORBIDDEN') 'public fail-closed route'
if (@($ps.PublicGate.Ast.ParamBlock.Parameters).Count -ne 0 -or @($ps.Runner.Ast.ParamBlock.Parameters).Count -ne 0 -or @($ps.Bootstrap.Ast.ParamBlock.Parameters).Count -ne 0) { throw 'A runner entrypoint exposes a file parameter binder.' }
[void](Get-HubQaSingleAssignment -Ast $ps.PublicGate.Ast -Variable '$rawInvocationArguments')
[void](Get-HubQaSingleAssignment -Ast $ps.Runner.Ast -Variable '$rawInvocationArguments')
[void](Get-HubQaSingleAssignment -Ast $ps.Bootstrap.Ast -Variable '$rawInvocationArguments')
foreach ($switchName in @('-Target','-Task','-Run')) { Assert-HubQaConditionalLiteral $ps.PublicGate.Ast $switchName 'public raw argv shape' }
Assert-HubQaConditionalInteger $ps.PublicGate.Ast 6 'public raw argv cardinality'
Assert-HubQaConditionalInteger $ps.Runner.Ast 0 'private runner raw argv cardinality'
Assert-HubQaConditionalInteger $ps.Bootstrap.Ast 0 'bootstrap raw argv cardinality'
Assert-HubQaIfConditionShape $ps.PublicGate.Ast '\$rawInvocationArguments\.Count -ne 6.*\[0\].*-Target.*\[2\].*-Task.*\[4\].*-Run' 'public raw argv order'
Assert-HubQaIfConditionShape $ps.Runner.Ast '\$MyInvocation\.InvocationName -ne ''\.''.*\$rawInvocationArguments\.Count -ne 0' 'private runner dot-source boundary'
Assert-HubQaIfConditionShape $ps.Bootstrap.Ast '\$MyInvocation\.InvocationName -ne ''\.''.*\$rawInvocationArguments\.Count -ne 0' 'bootstrap dot-source boundary'
Assert-HubQaFunctionCalls $ps.PublicGate.Ast @('Invoke-HubQaRunnerV2PublicRequest') 'public gate dispatch'
Assert-HubQaExitValues $ps.PublicGate @(0,64) 'public raw argv route'
Assert-HubQaExitValues $ps.Runner @(64) 'private raw argv route'
Assert-HubQaExitValues $ps.Bootstrap @(64) 'bootstrap raw argv route'

$fixtureAssertions = [ordered]@{
    'noncanonical-bootstrap-json' = { Assert-HubQaFunctionCalls $releaseAuthority.Body @('ConvertFrom-HubQaBootstrapClosedJson','Assert-HubQaBootstrapExactKeys') 'canonical release parser' }
    'wrong-version-or-producer' = {
        Assert-HubQaHashtableAssignmentKeys $snapshotWriter.Body '$payload' @('schema','producer','snapshot_id','captured_for','task','run','envelope_id','role_claim','self_attested','source_revision','head','index','worktree','commits','ancestry','control','evidence','lifecycle') 'snapshot payload'
        Assert-HubQaHashtableLiteralValue $snapshotWriter.Body '$payload' 'schema' 'hub.qa.repository-snapshot/v1' 'snapshot payload'
        Assert-HubQaHashtableLiteralValue $snapshotWriter.Body '$payload' 'producer' 'hub.qa.repository-snapshot-producer/v1' 'snapshot payload'
        Assert-HubQaHashtableAssignmentKeys $envelopeWriter.Body '$payload' @('schema','producer','envelope_id','contract','target','task','run','owner_nonce','repository_snapshot_sha256','report_root','limits','cleanup_contract','created_at_utc') 'envelope payload'
        Assert-HubQaHashtableLiteralValue $envelopeWriter.Body '$payload' 'schema' 'hub.qa.run-envelope/v1' 'envelope payload'
        Assert-HubQaHashtableLiteralValue $envelopeWriter.Body '$payload' 'producer' 'hub.qa.run-envelope-producer/v1' 'envelope payload'
        Assert-HubQaHashtableLiteralValue $envelopeWriter.Body '$payload' 'contract' 'hub.qa-runner/v2' 'envelope payload'
    }
    'short-or-foreign-run' = { Assert-HubQaThrowCodes $runnerPublic.Body @('E_B0_TASK_RUN') 'task/run grammar' }
    'serialized-role-claim' = {
        Assert-HubQaAssignedStringArray ([pscustomobject]@{ Ast = $runnerContext.Body }) '$expected' @('Task','Run','RoleClaim','OwnerNonce','OwnerMarkerRawBytes','OwnerMarkerSha256','RunnerInspectionRawBytes','RunnerInspectionSha256','LocalBaseReceiptRawBytes','LocalBaseReceiptSha256','LocalBaseReceiptNonce')
        Assert-HubQaThrowCodes $runnerContext.Body @('E_B0_ROLE_TRANSPORT') 'private role transport'
    }
    'direct-planning-selftest' = {
        Assert-HubQaFunctionCalls $planningClosure @('Invoke-HubQaBootstrapDocker') 'private planning closure'
        Assert-HubQaThrowCodes $planningClosure @('E_B0_ONE_SHOT_REJECTED') 'private planning one-shot closure'
        Assert-HubQaThrowCodes $runnerPublic.Body @('E_B0_DIRECT_INVOCATION_FORBIDDEN') 'public planning rejection'
    }
    'release-nonce-consumption' = {
        Assert-HubQaFunctionCalls $nonceParser.Body @('ConvertFrom-HubQaEnvelopeCanonicalJson','Assert-HubQaEnvelopeKeys') 'nonce ledger parser'
        Assert-HubQaHashtableAssignmentKeys $nonceWriter.Body '$record' @('schema','task','run','release_id','release_commit','release_raw_sha256','one_shot_nonce','consumed_at_utc','status') 'nonce ledger writer'
        Assert-HubQaHashtableLiteralValue $nonceWriter.Body '$record' 'schema' 'hub.qa.g0-nonce-consumption/v1' 'nonce ledger writer'
        Assert-HubQaHashtableLiteralValue $nonceWriter.Body '$record' 'status' 'CONSUMED' 'nonce ledger writer'
        Assert-HubQaFunctionCalls $nonceWriter.Body @('Get-HubQaBootstrapNonceConsumptionRoot','Read-HubQaBootstrapHeldBoundedIngressControl','ConvertFrom-HubQaBootstrapNonceConsumption','Write-HubQaEnvelopeSafeCreateNew') 'durable nonce ledger'
    }
    'successful-run-receipt-binding' = {
        Assert-HubQaHashtableAssignmentKeys $commandWriter.Body '$record' @('schema','producer','task','run','owner_marker_sha256','release_one_shot_nonce','commands','created_at_utc') 'command receipts'
        Assert-HubQaHashtableLiteralValue $commandWriter.Body '$record' 'schema' 'hub.qa.command-receipts/v1' 'command receipts'
        Assert-HubQaHashtableLiteralValue $commandWriter.Body '$record' 'producer' 'hub.qa-bootstrap/v1' 'command receipts'
        Assert-HubQaCommandLiteralSequence $commandWriter.Body 'Assert-HubQaBootstrapExactKeys' @('name','argv','exit_code','timed_out','started_at_utc','finished_at_utc','stdout_sha256','stderr_sha256','E_B0_COMMAND_RECEIPTS') 'command entry shape'
        Assert-HubQaCommandLiteralSequence $commandWriter.Body 'Join-Path' @('commands\command-receipts.v1.json') 'command receipt path'
        Assert-HubQaFunctionCalls $commandWriter.Body @('Write-HubQaBootstrapCreateNew') 'command receipt create-new writer'
        Assert-HubQaHashtableAssignmentKeys $cChainWriter.Body '$chain' $cReceiptKeys 'Package-C receipt chain'
        Assert-HubQaCommandLiteralSequence $cChainValidator.Body 'Assert-HubQaBootstrapExactKeys' @('path','raw_sha256','E_B0_C_RECEIPT_CHAIN') 'Package-C receipt pair shape'
        Assert-HubQaFunctionCalls $cChainWriter.Body @('Assert-HubQaBootstrapCReceiptChain') 'Package-C receipt validation'
    }
    'snapshot-envelope-drift' = {
        Assert-HubQaCommandOrder $bootstrapMain.Body @('New-HubQaRepositorySnapshot','New-HubQaRunEnvelope','Write-HubQaBootstrapRunReceipt') 'snapshot envelope run order'
        Assert-HubQaHashtableAssignmentKeys $runWriter.Body '$record' @('schema','producer','task','run','owner_marker_sha256','release_one_shot_nonce','repository_snapshot','run_envelope','status','created_at_utc') 'run receipt'
        Assert-HubQaHashtableLiteralValue $runWriter.Body '$record' 'schema' 'hub.qa.bootstrap-run/v1' 'run receipt'
        Assert-HubQaHashtableLiteralValue $runWriter.Body '$record' 'producer' 'hub.qa-bootstrap/v1' 'run receipt'
        Assert-HubQaHashtableLiteralValue $runWriter.Body '$record' 'status' 'completed-zero' 'run receipt'
        Assert-HubQaFunctionCalls $runWriter.Body @('ConvertFrom-HubQaBootstrapNonceConsumption','Write-HubQaBootstrapCreateNew') 'run nonce/raw-pair binding'
    }
    'missing-prebuild-source-receipt' = {
        Assert-HubQaFunctionCalls $sourceTreeWriter.Body @('Get-HubQaBootstrapAttestorGitObjectSourceTree','Write-HubQaBootstrapCreateNew') 'source-tree receipt'
        Assert-HubQaThrowCodes $sourceTreeWriter.Body @('E_B0_SOURCE_TREE_PREBUILD') 'source-tree receipt'
    }
    'partial-postbuild-attestation' = {
        Assert-HubQaFunctionCalls $postBuild.Body @('Invoke-HubQaBootstrapImageAttestation') 'post-build attestation'
        Assert-HubQaThrowCodes $postBuild.Body @('E_B0_IMAGE_ATTESTATION_POSTBUILD') 'post-build attestation'
    }
    'postbuild-base-receipt-binding' = {
        foreach ($name in @('Read-HubQaLocalBaseReceipt','Read-HubQaHelperAdmissionReceipt')) {
            $body = Get-HubQaFunctionAst $ps.Attestor $name
            Assert-HubQaThrowPrefix $body.Body 'E_RECEIPT_BINDING_INVALID' $name
        }
    }
    'd-evidence-acceptance-contract' = {
        Assert-HubQaFunctionCalls $activationD.Body @('Get-HubQaActivationNonceConsumptionReceipt','Get-HubQaActivationCanonicalNode','Assert-HubQaSnapshotEnvelopePayload') 'D evidence reader'
        Assert-HubQaThrowCodes $activationD.Body @('E_B0_D_RECEIPT_LINK','E_B0_D_SNAPSHOT_ENVELOPE_LINK','E_B0_D_VERDICT') 'D evidence reader'
        Assert-HubQaCommandLiteralSequence $activationD.Body 'Assert-HubQaActivationNodeKeys' @('schema','acceptance_id','g0_plan_path','g0_plan_commit','a0_commit','a02_commit','a1_commit','b_commit','founder_release','repository_commit','task','run','c_receipts','c1_lock','repository_snapshot','run_envelope','acceptance_verdict','accepted_at_utc','E_B0_D_SCHEMA_CLOSED') 'D exact schema'
        Assert-HubQaAssignedStringArray ([pscustomobject]@{ Ast = $activationD.Body }) '$receiptKeys' $cReceiptKeys
        Assert-HubQaCommandLiteralSequence $activationD.Body 'Assert-HubQaActivationNodeKeys' @('schema','producer','task','run','owner_marker_sha256','release_one_shot_nonce','repository_snapshot','run_envelope','status','created_at_utc','E_B0_D_SNAPSHOT_ENVELOPE_LINK') 'D run receipt shape'
        Assert-HubQaCommandLiteralSequence $activationD.Body 'Assert-HubQaActivationNodeKeys' @('schema','task','run','release_id','release_commit','release_raw_sha256','one_shot_nonce','consumed_at_utc','status','E_B0_D_SNAPSHOT_ENVELOPE_LINK') 'D nonce receipt shape'
        Assert-HubQaCommandLiteralSequence $activationD.Body 'Assert-HubQaActivationNodeKeys' @('path','raw_sha256','E_B0_D_SNAPSHOT_ENVELOPE_LINK') 'D run pair shape'
    }
    'activation-r-c1-d-chain' = {
        Assert-HubQaFunctionCalls $activationWriter.Body @('Get-HubQaActivationRelease','Assert-HubQaActivationPackageCommits','Get-HubQaActivationDRecord','Assert-HubQaActivationA02Authority','Get-HubQaActivationTargets') 'activation R/C1/D chain'
        Assert-HubQaHashtableAssignmentKeys $activationWriter.Body '$payload' @('schema','activation_id','g0_plan_path','required_plan_commit','a0_commit','a02_commit','a1_commit','b_commit','founder_release_path','founder_release_sha256','founder_release_commit','nonce_consumption_path','nonce_consumption_sha256','repository_commit','bootstrap_contract','bootstrap_contract_path','bootstrap_contract_sha256','runner_contract','runner_contract_path','runner_contract_sha256','dependency_bundle_manifest_path','dependency_bundle_manifest_sha256','source_policy_path','source_policy_sha256','build_manifest_path','build_manifest_sha256','images_lock_path','images_lock_sha256','target_entrypoints','planning_selftest','repository_snapshot_schema_path','repository_snapshot_schema_sha256','run_envelope_schema_path','run_envelope_schema_sha256','repository_snapshot_producer_path','repository_snapshot_producer_sha256','run_envelope_producer_path','run_envelope_producer_sha256','report_cleanup_contract','acceptance_verdict','acceptance_commit','created_at_utc') 'activation payload'
        Assert-HubQaFunctionCalls $activationA02.Body @('Test-HubQaActivationFirstParentMember','Get-HubQaActivationGitText','Get-HubQaActivationDiffPaths','Get-HubQaActivationGitBlob') 'activation A0.2 proof'
        Assert-HubQaCommandShape $activationA02.Body 'Get-HubQaActivationDiffPaths' '-From \$actualParent -To \$A02Commit' 'activation A0.2 actual-parent diff'
        Assert-HubQaCommandLiteralSequence $activationRelease.Body 'Assert-HubQaActivationNodeKeys' @('schema','release_id','g0_plan_path','g0_plan_commit','a0_commit','a02_commit','a1_commit','b_commit','task','run','local_only','network','offline_ingress_handoff_id','offline_ingress_owner_marker_sha256','dependency_bundle_manifest_sha256','source_commit','expires_at_utc','one_shot_nonce','founder_approval','created_at_utc') 'activation release keys'
    }
    'first-exit-cleanup' = {
        $tryStatements = @($bootstrapMain.Body.FindAll({ param($node) $node -is [Management.Automation.Language.TryStatementAst] -and $null -ne $node.Finally }, $true))
        if ($tryStatements.Count -lt 1) { throw 'Bootstrap lacks an active finally cleanup path.' }
        $cleanupCalls = @(Get-HubQaCommandAsts -Ast $bootstrapMain.Body -Name 'Invoke-HubQaBootstrapExactCleanup')
        if ($cleanupCalls.Count -ne 2) { throw 'Bootstrap does not have exactly one success cleanup and one finally cleanup.' }
        $cleanupFinals = @($tryStatements | Where-Object { @(Get-HubQaCommandAsts -Ast $_.Finally -Name 'Invoke-HubQaBootstrapExactCleanup').Count -eq 1 })
        if ($cleanupFinals.Count -ne 1) { throw 'Bootstrap lacks one unambiguous finally cleanup call.' }
        $cleanupTryOwners = @()
        foreach ($cleanupCall in $cleanupCalls) {
            $owner = $cleanupCall.Parent
            while ($null -ne $owner -and $owner -isnot [Management.Automation.Language.TryStatementAst]) { $owner = $owner.Parent }
            if ($owner -isnot [Management.Automation.Language.TryStatementAst] -or $owner.CatchClauses.Count -ne 1) { throw 'An exact-cleanup call is not owned by one guarded try/catch.' }
            $cleanupTryOwners += $owner
        }
        if (@($cleanupTryOwners | Select-Object -Unique).Count -ne 2) { throw 'Success and fallback cleanup do not have distinct active try/catch owners.' }
        $successCleanupOwners = @($cleanupTryOwners | Where-Object { @($_.CatchClauses[0].Body.FindAll({ param($node) $node -is [Management.Automation.Language.ThrowStatementAst] -and $node.Extent.Text.Trim() -ceq 'throw' }, $true)).Count -eq 1 })
        $fallbackCleanupOwners = @($cleanupTryOwners | Where-Object { @($_.CatchClauses[0].Body.FindAll({ param($node) $node -is [Management.Automation.Language.ThrowStatementAst] -and $node.Extent.Text.Trim() -ceq 'throw' }, $true)).Count -eq 0 })
        if ($successCleanupOwners.Count -ne 1 -or $fallbackCleanupOwners.Count -ne 1) { throw 'Success/fallback cleanup catches are not structurally distinct.' }
        Assert-HubQaFirstExitCatchMapping $successCleanupOwners[0].CatchClauses[0].Body 71 $true 'success cleanup catch'
        Assert-HubQaFirstExitCatchMapping $fallbackCleanupOwners[0].CatchClauses[0].Body 71 $false 'fallback cleanup catch'
        if ($cleanupFinals[0].CatchClauses.Count -ne 1) { throw 'Bootstrap outer failure catch is missing or ambiguous.' }
        Assert-HubQaFirstExitCatchMapping $cleanupFinals[0].CatchClauses[0].Body 65 $false 'bootstrap outer catch'
        $finalSuccessFences = @($cleanupFinals[0].Finally.FindAll({ param($node) $node -is [Management.Automation.Language.IfStatementAst] }, $true) | Where-Object {
            @($_.Clauses | Where-Object { (($_.Item1.Extent.Text -replace '\s+', ' ').Trim()) -cmatch '^\$firstExit -eq 0$' }).Count -eq 1 -and @(Get-HubQaCommandAsts -Ast $_ -Name 'Assert-HubQaBootstrapOuterDeadline').Count -eq 1
        })
        if ($finalSuccessFences.Count -ne 1) { throw 'Bootstrap final success deadline fence is missing or ambiguous.' }
        $finalFenceTries = @($finalSuccessFences[0].Clauses[0].Item2.FindAll({ param($node) $node -is [Management.Automation.Language.TryStatementAst] }, $true))
        if ($finalFenceTries.Count -ne 1 -or $finalFenceTries[0].CatchClauses.Count -ne 1) { throw 'Bootstrap final success deadline fence is not one guarded try/catch.' }
        $finalFenceAssignments = @(Get-HubQaAssignments -Ast $finalFenceTries[0].CatchClauses[0].Body -Variable '$firstExit')
        $finalFenceExit = if ($finalFenceAssignments.Count -eq 1) { try { [int]$finalFenceAssignments[0].Right.SafeGetValue() } catch { -1 } } else { -1 }
        if ($finalFenceAssignments.Count -ne 1 -or $finalFenceExit -ne 124) { throw 'Bootstrap final success deadline fence does not map expiry to 124.' }
        $runtimeStopCalls = @($cleanupFinals[0].Finally.FindAll({ param($node) $node -is [Management.Automation.Language.InvokeMemberExpressionAst] }, $true) | Where-Object {
            $memberName = try { [string]$_.Member.SafeGetValue() } catch { '' }
            $memberName -ceq 'Stop' -and (($_.Expression.Extent.Text -replace '\s+', ' ').Trim()) -ceq '$stopwatch'
        })
        if ($runtimeStopCalls.Count -ne 1 -or $finalSuccessFences[0].Extent.EndOffset -ge $runtimeStopCalls[0].Extent.StartOffset) { throw 'Bootstrap stops the outer clock before its final success fence.' }
        $stoppedSuccessFences = @($cleanupFinals[0].Finally.FindAll({ param($node) $node -is [Management.Automation.Language.IfStatementAst] }, $true) | Where-Object {
            @($_.Clauses | Where-Object { (($_.Item1.Extent.Text -replace '\s+', ' ').Trim()) -cmatch '^\$firstExit -eq 0$' }).Count -eq 1 -and @(Get-HubQaCommandAsts -Ast $_ -Name 'Assert-HubQaBootstrapDeadline').Count -eq 1
        })
        if ($stoppedSuccessFences.Count -ne 1) { throw 'Bootstrap stopped-outer success fence is missing or ambiguous.' }
        Assert-HubQaCommandShape $stoppedSuccessFences[0] 'Assert-HubQaBootstrapDeadline' '-Stopwatch \$stopwatch -Seconds \$script:HubQaBootstrapOuterSeconds' 'stopped outer terminal fence'
        $stoppedFenceTries = @($stoppedSuccessFences[0].Clauses[0].Item2.FindAll({ param($node) $node -is [Management.Automation.Language.TryStatementAst] }, $true))
        if ($stoppedFenceTries.Count -ne 1 -or $stoppedFenceTries[0].CatchClauses.Count -ne 1) { throw 'Bootstrap stopped-outer success fence is not one guarded try/catch.' }
        $stoppedFenceAssignments = @(Get-HubQaAssignments -Ast $stoppedFenceTries[0].CatchClauses[0].Body -Variable '$firstExit')
        $stoppedFenceExit = if ($stoppedFenceAssignments.Count -eq 1) { try { [int]$stoppedFenceAssignments[0].Right.SafeGetValue() } catch { -1 } } else { -1 }
        $runtimeNullAssignments = @(Get-HubQaAssignments -Ast $cleanupFinals[0].Finally -Variable '$script:HubQaBootstrapRuntimeStopwatch')
        if ($stoppedFenceExit -ne 124 -or $runtimeNullAssignments.Count -ne 1 -or $runtimeStopCalls[0].Extent.EndOffset -ge $stoppedSuccessFences[0].Extent.StartOffset -or $stoppedSuccessFences[0].Extent.EndOffset -ge $runtimeNullAssignments[0].Extent.StartOffset) { throw 'Bootstrap stopped-outer terminal fence does not preserve stop -> check -> clear with exit 124.' }
        $firstExitAssignments = @(Get-HubQaAssignments -Ast $bootstrapMain.Body -Variable '$firstExit')
        $initial = @($firstExitAssignments | Where-Object { try { [int]$_.Right.SafeGetValue() -eq 0 } catch { $false } })
        if ($initial.Count -ne 1) { throw 'Bootstrap does not initialize firstExit exactly once.' }
        foreach ($assignment in @($firstExitAssignments | Where-Object { $_.Extent.StartOffset -ne $initial[0].Extent.StartOffset })) {
            $owner = $assignment.Parent
            $guarded = $false
            while ($null -ne $owner -and $owner -isnot [Management.Automation.Language.FunctionDefinitionAst]) {
                if ($owner -is [Management.Automation.Language.IfStatementAst] -and @($owner.Clauses | Where-Object {
                    $condition = (($_.Item1.Extent.Text -replace '\s+', ' ').Trim())
                    $condition -cmatch '^\$firstExit -eq 0$' -or $condition -cmatch '^-not \[bool\]\$cleanup\.Succeeded -and \$firstExit -eq 0$'
                }).Count -eq 1) { $guarded = $true; break }
                $owner = $owner.Parent
            }
            if (-not $guarded) { throw 'A later firstExit assignment is not guarded by firstExit == 0.' }
        }
        $returns = @($bootstrapMain.Body.FindAll({ param($node) $node -is [Management.Automation.Language.ReturnStatementAst] -and $node.Pipeline.Extent.Text.Trim() -ceq '$firstExit' }, $true))
        if ($returns.Count -ne 1) { throw 'Bootstrap does not return the preserved first exit.' }
        Assert-HubQaConditionalLiteral $bootstrapMain.Body 'E_B0_TIMEOUT_124' 'first-exit timeout mapping'
        Assert-HubQaThrowCodes $bootstrapMain.Body @('E_B0_CLEANUP') 'first-exit cleanup'
    }
}

Assert-HubQaExactSequence -Actual @($fixtureAssertions.Keys) -Expected $fixtureNames -Identity 'Package-B exact fourteen-fixture matrix'
foreach ($fixture in $fixtureNames) { & $fixtureAssertions[$fixture] }

$admissionLines = Get-HubQaActiveTextLines -Path $path.AdmissionModule
Assert-HubQaActiveLine -Lines $admissionLines -Pattern '^parent_parts\s*=\s*\("build-context",\s*target,\s*"bindings",\s*\*parts\[:-1\]\)$' -Identity 'Python held archive topology'
foreach ($dockerfile in @($path.BackendDockerfile,$path.FrontendDockerfile,$path.OpenApiDockerfile)) {
    Assert-HubQaActiveLine -Lines (Get-HubQaActiveTextLines -Path $dockerfile) -Pattern '^ADD\s+bindings/assets/' -Identity 'admitted archive Dockerfile input'
}

[pscustomobject][ordered]@{
    contract = 'hub.qa.g0-runner-v2-static-contract/v1'
    status = $script:HubQaG0RunnerV2Status
    fixtures = $fixtureNames
    runtime = 'NOT_RUN'
}
