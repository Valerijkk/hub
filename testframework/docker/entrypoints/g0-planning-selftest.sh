#!/bin/sh
set -u

target='g0-planning-selftest'
report_dir='/reports/g0-planning-selftest'
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

if [ "$#" -ne 1 ] || [ "$1" != '--hub-qa-target=g0-planning-selftest' ]; then fail 64 'fixed-command-required'; fi
if [ ! -d /reports ]; then exit 73; fi
if [ "${A1_BINDING_STATUS:-}" != 'BOUND' ]; then fail 66 'binding-unavailable'; fi
for binding in /opt/hub/controls/dependency-bundle.manifest.v1.json /opt/hub/controls/source-policy.v2.json /opt/hub/controls/build-manifest.v3.json; do
  if [ ! -s "$binding" ]; then fail 66 'binding-unavailable'; fi
done
if [ ! -f /opt/hub/testframework/docker/runner-fixtures/g0-planning-selftest.py ]; then fail 66 'runner-fixture-unavailable'; fi

mkdir -p "$report_dir" || exit 73
if python3 /opt/hub/testframework/docker/runner-fixtures/g0-planning-selftest.py > "$report_dir/fixture.stdout.log" 2> "$report_dir/fixture.stderr.log"; then
  write_result 'PASS' 'shared-python-admission-fixtures-complete' 0
  exit 0
else
  first_exit=$?
  fail "$first_exit" 'shared-python-admission-fixtures'
fi
