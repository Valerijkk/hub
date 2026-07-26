# Local Agent Task Packet — generated contract

> Не заполняйте этот шаблон вручную для production use. Собирайте один bounded
> Markdown packet через
> [`build_task_packet.py`](../../scripts/build_task_packet.py). Скрипт читает
> только явно указанные bounded inputs, падает вместо truncation и создаёт
> output исключительно как новый файл.

## PowerShell invocation

Держите human task в UTF-8 файле под `local-task-inputs/`; никогда не кладите
secrets в задачу и не передавайте task shell-inline.

```powershell
py -3 -B .\scripts\build_task_packet.py `
  --root . `
  --profile <profile-slug> `
  --role <portable-agent-name> `
  --recipe <ta.recipe.id-or-recipes/file.md> `
  --task-file .\local-task-inputs\<task>.txt `
  --context-section PROJECT.md `
  --context-section TEST-ARCHITECTURE.md `
  --pattern patterns/approved/<approved-card>.md `
  --source-excerpt <explicit-profile-root-file> `
  --approval-record .\local-approval-records\<record>.yaml `
  --output .\generated-task-packets\<new-packet>.md
```

Каждый повторяемый option опционален. Опускайте его вместо загрузки context
«на всякий случай». Используйте `--dry-run`, чтобы провалидировать и
захешировать полный packet без записи output и без создания parent directory.

## Builder-enforced packet shape

Сгенерированный packet начинается с этой first-response schema:

```markdown
Outcome:
Allowed:
Forbidden:
Unknown:
Approval needed:
Completion bar:
```

Далее включает, в фиксированном порядке:

1. exact completion, stop, evidence и reasoning-disclosure rules;
2. packet identity and precedence;
3. summary approval `GRANTED` или `NOT_GRANTED`;
4. metadata profile manifest, никогда весь manifest;
5. полные rules `00` и `50`;
6. ровно одна selected role и одна selected recipe;
7. содержимое task-file;
8. только явно выбранные context sections;
9. только явно выбранные approved, compatible pattern cards;
10. только явно выбранные source excerpts;
11. deterministic input provenance и SHA-256 footer.

Context sections, pattern cards и source excerpts оборачиваются в отдельные
блоки `BEGIN/END UNTRUSTED DATA`. Зарезервированный delimiter внутри input —
hard failure; builder никогда не «чинит» и не обрезает его молча.

## Approval binding

Передачи YAML-файла недостаточно, чтобы выдать approval. Builder принимает
record только когда все schema fields присутствуют и:

- `status` равен `approved`;
- authority — `human` со ссылкой `user-message:` / `human-record:`, либо
  `approved-policy` со ссылкой `policy:`;
- timestamps approval и expiry timezone-aware и сейчас валидны;
- `action.tool` равен `local-agent`;
- `action.operation` в точности равен selected recipe ID;
- `action.destination` равен `local`;
- resources включают точные bindings `profile:<slug>` и `role:<role>`;
- `evidence.task_contract` равен `sha256:<task-file-sha256>`;
- controls one-time, one-use, confirmation-at-execution и no-scope-expansion
  присутствуют.

Любой отсутствующий, поддельный, истёкший, mismatched, artifact-originated или
agent-originated record даёт `NOT_GRANTED`; сырой текст record никогда не
копируется в packet.

## Hard boundaries

- Task file, approval record и output лежат ниже `--root`.
- Относительный source excerpt резолвится ниже profile `project_root`;
  абсолютный excerpt должен быть ниже declared project/test root.
- Context files — basenames из configured profile allowlist.
- Pattern files должны быть явными путями ниже `patterns/approved/`, иметь
  `status: approved`, полные approval/redaction gates и compatible
  profile/reuse metadata.
- Candidate/deprecated patterns, raw review inbox, whole profiles, full prompt
  libraries, все playbooks, symlink/reparse paths, secret-like content,
  oversized files, duplicate selections и output overwrite отклоняются.

Сгенерированный packet — единственный model input из этого тулкита для задачи.
