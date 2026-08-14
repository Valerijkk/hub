#!/bin/sh
set -u

target='backend-unit'
report_dir='/reports/backend-unit'
result_path="$report_dir/result.json"
first_exit=0

write_result() {
  status="$1"
  step="$2"
  code="$3"
  mkdir -p "$report_dir" || exit 73
  printf '{"target":"%s","status":"%s","step":"%s","first_exit":%s}\n' "$target" "$status" "$step" "$code" > "$result_path"
}

fail() {
  first_exit="$1"
  write_result 'FAIL' "$2" "$first_exit"
  exit "$first_exit"
}

if [ "$#" -ne 1 ] || [ "$1" != '--hub-qa-target=backend-unit' ]; then fail 64 'fixed-command-required'; fi
if [ ! -d /reports ]; then exit 73; fi
if [ "${A1_BINDING_STATUS:-}" != 'BOUND' ]; then fail 66 'binding-unavailable'; fi
for binding in /opt/hub/bindings/dependency-bundle.manifest.v1.json /opt/hub/bindings/source-policy.v2.json /opt/hub/bindings/build-manifest.v3.json; do
  if [ ! -s "$binding" ]; then fail 66 'binding-unavailable'; fi
done

run_step() {
  step="$1"
  shift
  if "$@" > "$report_dir/$step.stdout.log" 2> "$report_dir/$step.stderr.log"; then
    return 0
  else
    first_exit=$?
    fail "$first_exit" "$step"
  fi
}

mkdir -p "$report_dir" || exit 73
cd /opt/hub/source/backend || fail 72 'backend-source-unavailable'
export GOCACHE='/reports/backend-unit/go-cache'

run_step 'build' go build ./...
run_step 'vet' go vet ./...
run_step 'unit' go test -tags hubtest -json -count=1 -timeout=150s ./tests/...
run_step 'race' go test -tags hubtest -json -race -count=1 -timeout=150s ./tests/...
write_result 'PASS' 'complete' 0
