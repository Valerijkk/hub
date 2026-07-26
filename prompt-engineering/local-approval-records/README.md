# Local approval records

Опциональные YAML records по схеме
[`adapters/APPROVAL-RECORD.schema.yaml`](../adapters/APPROVAL-RECORD.schema.yaml).

Передавайте явный файл в `build_task_packet.py --approval-record …` только когда
recipe/action требует write, command, network или destructive scope.

Git-ignored, кроме этого README и sample ниже.
