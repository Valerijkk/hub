# Anti-AI checks

Reference for Phase 1 of `/style-check`. Source of truth for prohibited style
patterns: `.cursor/rules/10-style.mdc`.

## Narrating comments

Find comments that restate the next operation instead of explaining why:

```bash
rg -n "// (Handle|Return|Define|Import|Initialize|Create|Set|Get|Check|Update|Delete) the " backend/ frontend/src/ --type go --type ts --type tsx
```

Remove comments such as `// Return the result` when the code is self-explanatory.
Keep comments that document a non-obvious reason, invariant, or trade-off.

## Echo comments

```bash
rg -n "// (increment|decrement|return result|return error|log error|parse request|send response)" backend/ frontend/src/ --type go --type ts --type tsx
```

Treat each result as a review candidate, not an automatic deletion.

## Placeholder text

```bash
rg -n "(TODO:|FIXME:|HACK:|placeholder|lorem ipsum|example\.com|test@test\.com|dummy|fake data)" backend/internal/ frontend/src/ --type go --type ts --type tsx
```

Classify fixtures and deliberate examples separately. Production placeholders
must be replaced, moved to test data, or tied to a real tracked task.

## Hub-specific deprecated patterns

```bash
# SeaweedFS uses aws-sdk-go-v2, never minio-go.
rg -n "minio-go|minio\.New\(" backend/ --type go

# WebSocket uses Fiber, never Gorilla.
rg -n "gorilla/websocket" backend/ --type go
```

Logging, `console.*`, SSRF, unbounded reads, secret comparisons, and other
security/style checks are owned by the later `/style-check` phases and are not
duplicated here.

## Reporting

Report each confirmed finding as:

```text
🔴 [ANTI_AI_TYPE] path:line — concise explanation
```

If there are no confirmed findings, report `✅ No anti-AI patterns found`.
