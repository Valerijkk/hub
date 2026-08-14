#!/bin/sh
set -u

target='backend-openapi'
report_dir='/reports/backend-openapi'
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

if [ "$#" -ne 1 ] || [ "$1" != '--hub-qa-target=backend-openapi' ]; then fail 64 'fixed-command-required'; fi
if [ ! -d /reports ]; then exit 73; fi
if [ "${A1_BINDING_STATUS:-}" != 'BOUND' ]; then fail 66 'binding-unavailable'; fi
for binding in /opt/hub/bindings/dependency-bundle.manifest.v1.json /opt/hub/bindings/source-policy.v2.json /opt/hub/bindings/build-manifest.v3.json; do
  if [ ! -s "$binding" ]; then fail 66 'binding-unavailable'; fi
done
if [ ! -x /opt/hub/offline/swag/swag ]; then fail 66 'swag-unavailable'; fi

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

write_manifest() {
  generation="$1"
  manifest="$2"
  if (cd "$generation" && find . -type f -print | LC_ALL=C sort | xargs -r sha256sum) > "$manifest" 2> "$report_dir/manifest.stderr.log"; then
    return 0
  else
    first_exit=$?
    fail "$first_exit" 'generation-manifest'
  fi
}

assert_equal() {
  step="$1"
  left="$2"
  right="$3"
  if diff -u "$left" "$right" > "$report_dir/$step.diff" 2>&1; then
    return 0
  else
    first_exit=$?
    fail "$first_exit" "$step"
  fi
}

mkdir -p "$report_dir" || exit 73
cd /opt/hub/source/backend || fail 72 'backend-source-unavailable'
export GOCACHE='/reports/backend-openapi/go-cache'

run_step 'generation-1' /opt/hub/offline/swag/swag init -g cmd/server/main.go -o /reports/backend-openapi/generation-1 --parseDependency --parseInternal
run_step 'generation-2' /opt/hub/offline/swag/swag init -g cmd/server/main.go -o /reports/backend-openapi/generation-2 --parseDependency --parseInternal
write_manifest '/reports/backend-openapi/generation-1' '/reports/backend-openapi/generation-1.sha256'
write_manifest '/reports/backend-openapi/generation-2' '/reports/backend-openapi/generation-2.sha256'
assert_equal 'generation-equality' '/reports/backend-openapi/generation-1.sha256' '/reports/backend-openapi/generation-2.sha256'
for generated in docs.go swagger.json swagger.yaml; do
  assert_equal "source-sync-$generated" "/reports/backend-openapi/generation-1/$generated" "/opt/hub/source/backend/docs/$generated"
done
write_result 'PASS' 'complete' 0
