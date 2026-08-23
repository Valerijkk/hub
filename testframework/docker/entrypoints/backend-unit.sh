#!/bin/sh
set -eu

[ "$#" -eq 0 ] || exit 64

report_dir="/reports/backend-unit"
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

mkdir -p "$report_dir" "$report_dir/go-cache" "$report_dir/tmp"
cd /opt/hub/source/backend

if go build ./...; then
  write_step build PASS 0
else
  code=$?
  record_failure "$code"
  write_step build FAIL "$code"
fi > "$report_dir/build.stdout.log" 2> "$report_dir/build.stderr.log"

if go vet -tags hubtest ./...; then
  write_step vet PASS 0
else
  code=$?
  record_failure "$code"
  write_step vet FAIL "$code"
fi > "$report_dir/vet.stdout.log" 2> "$report_dir/vet.stderr.log"

if go test -tags hubtest -json -count=1 -timeout=150s ./tests/...; then
  write_step unit PASS 0
else
  code=$?
  record_failure "$code"
  write_step unit FAIL "$code"
fi > "$report_dir/unit.stdout.log" 2> "$report_dir/unit.stderr.log"

if go test -tags hubtest -json -race -count=1 -timeout=150s ./tests/...; then
  write_step race PASS 0
else
  code=$?
  record_failure "$code"
  write_step race FAIL "$code"
fi > "$report_dir/race.stdout.log" 2> "$report_dir/race.stderr.log"

if [ "$first_exit" -eq 0 ]; then
  status=PASS
else
  status=FAIL
fi
printf '{"target":"backend-unit","status":"%s","first_exit":%s}\n' \
  "$status" "$first_exit" > "$result_path"
exit "$first_exit"
