#!/bin/sh
set -u

target='frontend-unit'
report_dir='/reports/frontend-unit'
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

if [ "$#" -ne 1 ] || [ "$1" != '--hub-qa-target=frontend-unit' ]; then fail 64 'fixed-command-required'; fi
if [ ! -d /reports ]; then exit 73; fi
if [ ! -x /usr/local/bin/node ]; then fail 66 'node-unavailable'; fi
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
cd /opt/hub/source/frontend || fail 72 'frontend-source-unavailable'
export HOME='/reports/frontend-unit/home'
export TMPDIR='/reports/frontend-unit/tmp'
mkdir -p "$HOME" "$TMPDIR" || exit 73

run_step 'typescript' /usr/local/bin/node /opt/hub/source/frontend/node_modules/typescript/bin/tsc -b
run_step 'biome' /usr/local/bin/node /opt/hub/source/frontend/node_modules/@biomejs/biome/bin/biome check ./src
run_step 'vite' /usr/local/bin/node /opt/hub/source/frontend/node_modules/vite/bin/vite.js build --outDir /reports/frontend-unit/vite-dist
run_step 'vitest' /usr/local/bin/node /opt/hub/source/frontend/node_modules/vitest/vitest.mjs run --config /opt/hub/config/vitest.frontend-unit.config.ts --reporter=json
write_result 'PASS' 'complete' 0
