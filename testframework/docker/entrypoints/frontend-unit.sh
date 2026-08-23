#!/bin/sh
set -eu

[ "$#" -eq 0 ] || exit 64

report_dir="/reports/frontend-unit"
result_path="$report_dir/result.json"
first_exit=0

record_failure() {
  code="$1"
  if [ "$first_exit" -eq 0 ]; then
    first_exit="$code"
  fi
}

write_step() {
  step="$1"
  status="$2"
  code="$3"
  printf '{"step":"%s","status":"%s","exit":%s}\n' \
    "$step" "$status" "$code" > "$report_dir/$step.result.json"
}

mkdir -p "$report_dir" "$report_dir/home" "$report_dir/tmp"
cd /opt/hub/source/frontend

if /usr/bin/node /opt/hub/source/frontend/node_modules/typescript/bin/tsc -p /opt/hub/source/frontend/tsconfig.json --noEmit; then
  write_step typescript PASS 0
else
  code=$?
  record_failure "$code"
  write_step typescript FAIL "$code"
fi > "$report_dir/typescript.stdout.log" 2> "$report_dir/typescript.stderr.log"

if /usr/bin/node /opt/hub/source/frontend/node_modules/vitest/vitest.mjs run --config /opt/hub/source/frontend/.qa/vitest.frontend-unit.config.ts --configLoader runner; then
  write_step vitest PASS 0
else
  code=$?
  record_failure "$code"
  write_step vitest FAIL "$code"
fi > "$report_dir/vitest.stdout.log" 2> "$report_dir/vitest.stderr.log"

if [ "$first_exit" -eq 0 ]; then
  status=PASS
else
  status=FAIL
fi
printf '{"target":"frontend-unit","status":"%s","first_exit":%s}\n' \
  "$status" "$first_exit" > "$result_path"
exit "$first_exit"
