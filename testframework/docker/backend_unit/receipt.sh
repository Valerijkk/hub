#!/bin/sh
set -eu

RECEIPT_SCHEMA=hub.qa.backend-unit-container-receipt/v1
RECEIPT_TARGET=backend-unit
RECEIPT_SOURCE_ROOT=/opt/hub/source/backend

# The entrypoint installs the TERM trap only after every runner source file is
# loaded. Until backend_unit_main has established a receipt-safe checkpoint,
# a TERM is deliberately fail-closed: there is no authority to seal an abort.
backend_unit_term_phase=cold
backend_unit_term_safe=false
backend_unit_term_handling=false

_receipt_decimal() {
  [ "$#" -eq 1 ] || return 64
  case "$1" in ''|*[!0-9]*) return 75 ;; *) return 0 ;; esac
}

_receipt_sha256_file() {
  [ "$#" -eq 1 ] || return 64
  [ -f "$1" ] && [ ! -L "$1" ] || return 75
  receipt_hash_line="$(sha256sum -- "$1")" || return 75
  receipt_hash="$(printf '%s\n' "$receipt_hash_line" | awk 'NF >= 1 && $1 ~ /^[0-9a-f]{64}$/ { print $1; exit }')" || return 75
  case "$receipt_hash" in [0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]) ;; *) return 75 ;; esac
  printf 'sha256:%s\n' "$receipt_hash"
}

_receipt_sha256_text() {
  [ "$#" -eq 1 ] || return 64
  receipt_text_hash="$(printf '%s' "$1" | sha256sum | awk 'NF >= 1 && $1 ~ /^[0-9a-f]{64}$/ { print $1; exit }')" || return 75
  case "$receipt_text_hash" in [0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]) ;; *) return 75 ;; esac
  printf 'sha256:%s\n' "$receipt_text_hash"
}

_receipt_file_size() {
  [ "$#" -eq 1 ] || return 64
  receipt_size="$(stat -c '%s' -- "$1")" || return 75
  _receipt_decimal "$receipt_size" || return 75
  printf '%s\n' "$receipt_size"
}

_receipt_safe_json_text() {
  [ "$#" -eq 1 ] || return 64
  case "$1" in *\"*|*\\*) return 75 ;; esac
  if printf '%s' "$1" | LC_ALL=C grep -q '[^ -~]'; then return 75; fi
  printf '%s\n' "$1"
}

_receipt_stage_name() {
  [ "$#" -eq 1 ] || return 64
  case "$1" in build|vet|unit|race) return 0 ;; *) return 75 ;; esac
}

_receipt_run_dir_current() {
  [ "$#" -eq 0 ] || return 64
  [ -n "${receipt_run_identity-}" ] || return 75
  _owned_run_dir_is_current "$receipt_run_identity"
}

_receipt_regular_is_owned() {
  [ "$#" -eq 1 ] || return 64
  _receipt_run_dir_current || return 75
  _run_dir_target_is_safe "$1" || return 75
  _owned_regular_identity "$1" >/dev/null || return 75
  receipt_links="$(_owned_regular_links "$1")" || return 75
  [ "$receipt_links" = 1 ] || return 75
}

_receipt_remove_owned_regular() {
  [ "$#" -eq 2 ] || return 64
  receipt_remove_path="$1"
  receipt_remove_identity="$2"
  _receipt_run_dir_current || return 75
  _run_dir_target_is_safe "$receipt_remove_path" || return 75
  _owned_temp_is_current "$receipt_remove_path" "$receipt_remove_identity" 1 || return 75
  rm -- "$receipt_remove_path" || return 75
  [ ! -e "$receipt_remove_path" ] && [ ! -L "$receipt_remove_path" ] || return 75
}

_receipt_validate_owner() {
  [ "$#" -eq 0 ] || return 64
  receipt_owner_path="$run_dir/owner.json"
  _receipt_regular_is_owned "$receipt_owner_path" || return 75
  receipt_owner_text="$(cat -- "$receipt_owner_path")" || return 75
  receipt_owner_expected="{\"run_id\":\"$run_id\",\"owner_nonce\":\"$owner_nonce\",\"container_hostname\":\"$container_hostname\"}"
  [ "$receipt_owner_text" = "$receipt_owner_expected" ] || return 75
  receipt_owner_sha256="$(_receipt_sha256_file "$receipt_owner_path")" || return 75
}

_receipt_snapshot_decimal_value() {
  [ "$#" -eq 2 ] || return 64
  receipt_snapshot_field="$1"
  receipt_snapshot_key="$2"
  case "$receipt_snapshot_field" in "$receipt_snapshot_key="*) receipt_snapshot_value=${receipt_snapshot_field#"$receipt_snapshot_key="} ;; *) return 75 ;; esac
  _receipt_decimal "$receipt_snapshot_value" || return 75
  printf '%s\n' "$receipt_snapshot_value"
}

_receipt_snapshot_object() {
  [ "$#" -eq 1 ] || return 64
  receipt_snapshot="$1"
  case "$receipt_snapshot" in
    "schema=hub.qa.cgroup-snapshot/v1,family=v2,memory.max=$CGROUP_MEMORY_BYTES,memory.swap.max=0,cpu.max=$CGROUP_CPU_QUOTA_US/$CGROUP_CPU_PERIOD_US,memory.events.oom="*)
      receipt_snapshot_tail=${receipt_snapshot#"schema=hub.qa.cgroup-snapshot/v1,family=v2,memory.max=$CGROUP_MEMORY_BYTES,memory.swap.max=0,cpu.max=$CGROUP_CPU_QUOTA_US/$CGROUP_CPU_PERIOD_US,"}
      receipt_snapshot_oom_field=${receipt_snapshot_tail%%,*}
      receipt_snapshot_tail=${receipt_snapshot_tail#"$receipt_snapshot_oom_field,"}
      [ "$receipt_snapshot_tail" != "$receipt_snapshot_oom_field" ] || return 75
      receipt_snapshot_oom="$(_receipt_snapshot_decimal_value "$receipt_snapshot_oom_field" memory.events.oom)" || return 75
      receipt_snapshot_oom_kill="$(_receipt_snapshot_decimal_value "$receipt_snapshot_tail" memory.events.oom_kill)" || return 75
      printf '{"memory.events":{"oom":%s,"oom_kill":%s},"memory.max":"%s","memory.swap.max":"0"}' "$receipt_snapshot_oom" "$receipt_snapshot_oom_kill" "$CGROUP_MEMORY_BYTES"
      ;;
    "schema=hub.qa.cgroup-snapshot/v1,family=v1,memory.limit_in_bytes=$CGROUP_MEMORY_BYTES,memory.memsw.limit_in_bytes=$CGROUP_MEMORY_BYTES,cpu.cfs_quota_us=$CGROUP_CPU_QUOTA_US,cpu.cfs_period_us=$CGROUP_CPU_PERIOD_US,memory.failcnt="*)
      receipt_snapshot_tail=${receipt_snapshot#"schema=hub.qa.cgroup-snapshot/v1,family=v1,memory.limit_in_bytes=$CGROUP_MEMORY_BYTES,memory.memsw.limit_in_bytes=$CGROUP_MEMORY_BYTES,cpu.cfs_quota_us=$CGROUP_CPU_QUOTA_US,cpu.cfs_period_us=$CGROUP_CPU_PERIOD_US,"}
      receipt_snapshot_failcnt_field=${receipt_snapshot_tail%%,*}
      receipt_snapshot_tail=${receipt_snapshot_tail#"$receipt_snapshot_failcnt_field,"}
      [ "$receipt_snapshot_tail" != "$receipt_snapshot_failcnt_field" ] || return 75
      receipt_snapshot_oom_disable_field=${receipt_snapshot_tail%%,*}
      receipt_snapshot_tail=${receipt_snapshot_tail#"$receipt_snapshot_oom_disable_field,"}
      [ "$receipt_snapshot_tail" != "$receipt_snapshot_oom_disable_field" ] || return 75
      receipt_snapshot_failcnt="$(_receipt_snapshot_decimal_value "$receipt_snapshot_failcnt_field" memory.failcnt)" || return 75
      receipt_snapshot_oom_disable="$(_receipt_snapshot_decimal_value "$receipt_snapshot_oom_disable_field" memory.oom_control.oom_kill_disable)" || return 75
      receipt_snapshot_under_oom="$(_receipt_snapshot_decimal_value "$receipt_snapshot_tail" memory.oom_control.under_oom)" || return 75
      printf '{"memory.failcnt":%s,"memory.limit_in_bytes":"%s","memory.memsw.limit_in_bytes":"%s","memory.oom_control":{"oom_kill_disable":%s,"under_oom":%s}}' "$receipt_snapshot_failcnt" "$CGROUP_MEMORY_BYTES" "$CGROUP_MEMORY_BYTES" "$receipt_snapshot_oom_disable" "$receipt_snapshot_under_oom"
      ;;
    *) return 75 ;;
  esac
}

_receipt_results_are_canonical() {
  [ "$#" -eq 2 ] || return 64
  receipt_results_path="$1"
  receipt_expected_rows="$2"
  awk -v expected="$receipt_expected_rows" '
    BEGIN { ordinal = 0; invalid = 0 }
    /^[^\r\n]+$/ {
      ordinal += 1
      prefix = "{\"ordinal\":" ordinal ",\"import_path\":\""
      if (index($0, prefix) != 1) { invalid = 1; next }
      if ($0 !~ /,\"package_name\":\"[A-Za-z_][A-Za-z0-9_]*\",\"result\":\"(PASS|FAIL)\",\"exit_code\":[0-9]+}$/) { invalid = 1; next }
      if ($0 ~ /\"result\":\"PASS\",\"exit_code\":0}$/) next
      if ($0 ~ /\"result\":\"FAIL\",\"exit_code\":[1-9][0-9]*}$/) next
      invalid = 1
    }
    END { if (invalid || ordinal != expected) exit 75 }
  ' "$receipt_results_path"
}

_receipt_parse_stage_json() {
  [ "$#" -eq 1 ] || return 64
  receipt_stage_line="$1"
  receipt_stage_prefix="{\"schema\":\"hub.qa.backend-unit-stage-result/v1\",\"target\":\"backend-unit\",\"run_id\":\"$run_id\",\"owner_nonce\":\"$owner_nonce\",\"stage\":\"$receipt_stage\",\"package_plan_sha256\":\"$receipt_stage_plan_sha\",\"package_result_sha256\":\"$receipt_stage_results_sha\",\"package_count\":"
  case "$receipt_stage_line" in "$receipt_stage_prefix"*) receipt_stage_remaining=${receipt_stage_line#"$receipt_stage_prefix"} ;; *) return 75 ;; esac
  receipt_stage_package_count=${receipt_stage_remaining%%,*}
  _receipt_decimal "$receipt_stage_package_count" || return 75
  receipt_stage_after_count=${receipt_stage_remaining#"$receipt_stage_package_count,\"package_result_count\":"}
  [ "$receipt_stage_after_count" != "$receipt_stage_remaining" ] || return 75
  receipt_stage_result_count=${receipt_stage_after_count%%,*}
  _receipt_decimal "$receipt_stage_result_count" || return 75
  receipt_stage_after_results=${receipt_stage_after_count#"$receipt_stage_result_count,\"first_exit\":"}
  [ "$receipt_stage_after_results" != "$receipt_stage_after_count" ] || return 75
  receipt_stage_first_exit=${receipt_stage_after_results%%,*}
  _receipt_decimal "$receipt_stage_first_exit" || return 75
  receipt_stage_after_exit=${receipt_stage_after_results#"$receipt_stage_first_exit,\"status\":\""}
  [ "$receipt_stage_after_exit" != "$receipt_stage_after_results" ] || return 75
  receipt_stage_status=${receipt_stage_after_exit%%\"*}
  case "$receipt_stage_status" in PASS|FAIL) ;; *) return 75 ;; esac
  receipt_stage_suffix="\",\"stdout_sha256\":\"$receipt_stage_stdout_sha\",\"stdout_size\":$receipt_stage_stdout_size,\"stderr_sha256\":\"$receipt_stage_stderr_sha\",\"stderr_size\":$receipt_stage_stderr_size,\"cgroup_baseline\":\"$container_cgroup_baseline\",\"cgroup_final\":\"$container_cgroup_final\",\"report_budget_truncated_due_to_budget\":false}"
  [ "$receipt_stage_after_exit" = "$receipt_stage_status$receipt_stage_suffix" ] || return 75
  [ "$receipt_stage_package_count" -gt 0 ] || return 75
  [ "$receipt_stage_package_count" = "$receipt_stage_result_count" ] || return 75
  case "$receipt_stage_status:$receipt_stage_first_exit" in PASS:0|FAIL:[1-9]*) ;; *) return 75 ;; esac
}

_receipt_validate_stage() {
  [ "$#" -eq 1 ] || return 64
  receipt_stage="$1"
  _receipt_stage_name "$receipt_stage" || return 75
  receipt_rows_path="$(package_plan_rows_path "$receipt_stage")" || return 75
  _owned_regular_identity "$receipt_rows_path" >/dev/null || return 75
  _plan_validate_rows "$receipt_rows_path" || return 75
  _execute_package_plan_is_bound "$receipt_stage" "$receipt_rows_path" || return 75
  receipt_stage_plan_path="$run_dir/package-plan.$receipt_stage.json"
  receipt_stage_results_path="$run_dir/package-results.$receipt_stage.jsonl"
  receipt_stage_stdout_path="$run_dir/$receipt_stage.stdout.log"
  receipt_stage_stderr_path="$run_dir/$receipt_stage.stderr.log"
  receipt_stage_json_path="$run_dir/stage-$receipt_stage.json"
  for receipt_stage_file in "$receipt_stage_plan_path" "$receipt_stage_results_path" "$receipt_stage_stdout_path" "$receipt_stage_stderr_path" "$receipt_stage_json_path"; do
    _receipt_regular_is_owned "$receipt_stage_file" || return 75
  done
  receipt_stage_plan_sha="$(_receipt_sha256_file "$receipt_stage_plan_path")" || return 75
  receipt_stage_results_sha="$(_receipt_sha256_file "$receipt_stage_results_path")" || return 75
  receipt_stage_stdout_sha="$(_receipt_sha256_file "$receipt_stage_stdout_path")" || return 75
  receipt_stage_stderr_sha="$(_receipt_sha256_file "$receipt_stage_stderr_path")" || return 75
  receipt_stage_stdout_size="$(_receipt_file_size "$receipt_stage_stdout_path")" || return 75
  receipt_stage_stderr_size="$(_receipt_file_size "$receipt_stage_stderr_path")" || return 75
  receipt_stdout_cap="$(stage_stream_cap "$receipt_stage" stdout)" || return 75
  receipt_stderr_cap="$(stage_stream_cap "$receipt_stage" stderr)" || return 75
  [ "$receipt_stage_stdout_size" -le "$receipt_stdout_cap" ] || return 75
  [ "$receipt_stage_stderr_size" -le "$receipt_stderr_cap" ] || return 75
  receipt_stage_rows_count="$(wc -l < "$receipt_rows_path")" || return 75
  _receipt_decimal "$receipt_stage_rows_count" || return 75
  receipt_stage_line="$(cat -- "$receipt_stage_json_path")" || return 75
  _receipt_parse_stage_json "$receipt_stage_line" || return 75
  [ "$receipt_stage_rows_count" = "$receipt_stage_package_count" ] || return 75
  _receipt_results_are_canonical "$receipt_stage_results_path" "$receipt_stage_result_count" || return 75
  receipt_stage_sha256="$(_receipt_sha256_file "$receipt_stage_json_path")" || return 75
  case "$receipt_stage" in
    build) build_plan_sha256="$receipt_stage_plan_sha"; build_stage_sha256="$receipt_stage_sha256"; build_stdout_size="$receipt_stage_stdout_size"; build_stderr_size="$receipt_stage_stderr_size" ;;
    vet) vet_plan_sha256="$receipt_stage_plan_sha"; vet_stage_sha256="$receipt_stage_sha256"; vet_stdout_size="$receipt_stage_stdout_size"; vet_stderr_size="$receipt_stage_stderr_size" ;;
    unit) unit_plan_sha256="$receipt_stage_plan_sha"; unit_stage_sha256="$receipt_stage_sha256"; unit_stdout_size="$receipt_stage_stdout_size"; unit_stderr_size="$receipt_stage_stderr_size" ;;
    race) race_plan_sha256="$receipt_stage_plan_sha"; race_stage_sha256="$receipt_stage_sha256"; race_stdout_size="$receipt_stage_stdout_size"; race_stderr_size="$receipt_stage_stderr_size" ;;
  esac
  if [ "$receipt_stage_first_exit" -ne 0 ] && [ "$receipt_first_stage_exit" -eq 0 ]; then receipt_first_stage_exit="$receipt_stage_first_exit"; fi
}

_receipt_precomplete_tree() {
  [ "$#" -eq 0 ] || return 64
  receipt_tree="$(find "$run_dir" -xdev -mindepth 1 -maxdepth 1 -printf '%f\t%y\t%n\n' | LC_ALL=C sort)" || return 75
  receipt_expected_tree='build.stderr.log	f	1
build.stdout.log	f	1
owner.json	f	1
package-plan.build.json	f	1
package-plan.race.json	f	1
package-plan.unit.json	f	1
package-plan.vet.json	f	1
package-results.build.jsonl	f	1
package-results.race.jsonl	f	1
package-results.unit.jsonl	f	1
package-results.vet.jsonl	f	1
race.stderr.log	f	1
race.stdout.log	f	1
stage-build.json	f	1
stage-race.json	f	1
stage-unit.json	f	1
stage-vet.json	f	1
unit.stderr.log	f	1
unit.stdout.log	f	1
vet.stderr.log	f	1
vet.stdout.log	f	1'
  [ "$receipt_tree" = "$receipt_expected_tree" ] || return 75
}

_receipt_complete_tree() {
  [ "$#" -eq 0 ] || return 64
  receipt_tree="$(find "$run_dir" -xdev -mindepth 1 -maxdepth 1 -printf '%f\t%y\t%n\n' | LC_ALL=C sort)" || return 75
  receipt_expected_tree='CONTAINER_COMPLETE	f	1
build.stderr.log	f	1
build.stdout.log	f	1
owner.json	f	1
package-plan.build.json	f	1
package-plan.race.json	f	1
package-plan.unit.json	f	1
package-plan.vet.json	f	1
package-results.build.jsonl	f	1
package-results.race.jsonl	f	1
package-results.unit.jsonl	f	1
package-results.vet.jsonl	f	1
race.stderr.log	f	1
race.stdout.log	f	1
stage-build.json	f	1
stage-race.json	f	1
stage-unit.json	f	1
stage-vet.json	f	1
unit.stderr.log	f	1
unit.stdout.log	f	1
vet.stderr.log	f	1
vet.stdout.log	f	1'
  [ "$receipt_tree" = "$receipt_expected_tree" ] || return 75
}

_receipt_create_complete_marker() {
  [ "$#" -eq 0 ] || return 64
  receipt_marker_path="$run_dir/CONTAINER_COMPLETE"
  _receipt_run_dir_current || return 75
  _run_dir_target_is_safe "$receipt_marker_path" || return 75
  [ ! -e "$receipt_marker_path" ] && [ ! -L "$receipt_marker_path" ] || return 75
  ( umask 077; set -C; : > "$receipt_marker_path" ) || return 75
  receipt_marker_identity="$(_owned_regular_identity "$receipt_marker_path")" || return 75
  receipt_marker_links="$(_owned_regular_links "$receipt_marker_path")" || return 75
  [ "$receipt_marker_links" = 1 ] || return 75
  receipt_marker_size="$(_receipt_file_size "$receipt_marker_path")" || return 75
  [ "$receipt_marker_size" = 0 ] || return 75
}

_receipt_inventory() {
  [ "$#" -eq 0 ] || return 64
  receipt_inventory_json=
  receipt_inventory_lines=
  receipt_inventory_first=true
  for receipt_inventory_name in CONTAINER_COMPLETE build.stderr.log build.stdout.log owner.json package-plan.build.json package-plan.race.json package-plan.unit.json package-plan.vet.json package-results.build.jsonl package-results.race.jsonl package-results.unit.jsonl package-results.vet.jsonl race.stderr.log race.stdout.log stage-build.json stage-race.json stage-unit.json stage-vet.json unit.stderr.log unit.stdout.log vet.stderr.log vet.stdout.log; do
    receipt_inventory_path="$run_dir/$receipt_inventory_name"
    _receipt_regular_is_owned "$receipt_inventory_path" || return 75
    receipt_inventory_sha="$(_receipt_sha256_file "$receipt_inventory_path")" || return 75
    receipt_inventory_size="$(_receipt_file_size "$receipt_inventory_path")" || return 75
    if [ "$receipt_inventory_first" = true ]; then
      receipt_inventory_json="{\"path\":\"$receipt_inventory_name\",\"sha256\":\"$receipt_inventory_sha\",\"size\":$receipt_inventory_size}"
      receipt_inventory_first=false
    else
      receipt_inventory_json="$receipt_inventory_json,{\"path\":\"$receipt_inventory_name\",\"sha256\":\"$receipt_inventory_sha\",\"size\":$receipt_inventory_size}"
    fi
    receipt_inventory_lines="${receipt_inventory_lines}${receipt_inventory_name}	${receipt_inventory_sha}	${receipt_inventory_size}
"
  done
  receipt_inventory_sha256="$(_receipt_sha256_text "$receipt_inventory_lines")" || return 75
}

_receipt_source_manifest() {
  [ "$#" -eq 0 ] || return 64
  [ -d "$RECEIPT_SOURCE_ROOT" ] && [ ! -L "$RECEIPT_SOURCE_ROOT" ] || return 75
  _create_owned_temp "$GOTMPDIR" budget-all || return 75
  receipt_manifest_path="$owned_temp_path"
  receipt_manifest_identity="$owned_temp_identity"
  if _owned_temp_is_current "$receipt_manifest_path" "$receipt_manifest_identity" 1 && find "$RECEIPT_SOURCE_ROOT" -xdev -type f -exec sha256sum -- {} \; | LC_ALL=C sort > "$receipt_manifest_path" && _owned_temp_is_current "$receipt_manifest_path" "$receipt_manifest_identity" 1; then
    :
  else
    _cleanup_owned_temp "$receipt_manifest_path" "$receipt_manifest_identity" 1 || return 75
    return 75
  fi
  receipt_manifest_sha256="$(_receipt_sha256_file "$receipt_manifest_path")" || {
    _cleanup_owned_temp "$receipt_manifest_path" "$receipt_manifest_identity" 1 || return 75
    return 75
  }
  _cleanup_owned_temp "$receipt_manifest_path" "$receipt_manifest_identity" 1 || return 75
  printf '%s\n' "$receipt_manifest_sha256"
}

_receipt_environment_observation() {
  [ "$#" -eq 1 ] || return 64
  receipt_env_name="$1"
  case "$receipt_env_name" in
    GODEBUG)
      if [ -n "${GODEBUG+x}" ]; then receipt_env_value="$GODEBUG"; else printf '{"name":"GODEBUG","present":false,"value":null}'; return 0; fi
      ;;
    GOEXPERIMENT)
      if [ -n "${GOEXPERIMENT+x}" ]; then receipt_env_value="$GOEXPERIMENT"; else printf '{"name":"GOEXPERIMENT","present":false,"value":null}'; return 0; fi
      ;;
    GOMEMLIMIT)
      if [ -n "${GOMEMLIMIT+x}" ]; then receipt_env_value="$GOMEMLIMIT"; else printf '{"name":"GOMEMLIMIT","present":false,"value":null}'; return 0; fi
      ;;
    *) return 64 ;;
  esac
  _receipt_safe_json_text "$receipt_env_value" >/dev/null || return 75
  printf '{"name":"%s","present":true,"value":"%s"}' "$receipt_env_name" "$receipt_env_value"
}

_receipt_capture_go_environment() {
  [ "$#" -eq 0 ] || return 64
  [ "${GOFLAGS-}" = -p=1 ] || return 75
  [ "${GOENV-}" = off ] || return 75
  [ "${GOWORK-}" = off ] || return 75
  [ "${GOTOOLCHAIN-}" = local ] || return 75
  [ "${GOPROXY-}" = off ] || return 75
  [ -z "${GOMAXPROCS+x}" ] || return 75
  [ "${GOCACHE-}" = /var/cache/hub-go ] || return 75
  [ "${GOTMPDIR-}" = /var/tmp/hub-go ] || return 75
  receipt_go_version="$(go version)" || return 75
  receipt_cgo_enabled="$(go env CGO_ENABLED)" || return 75
  receipt_goos="$(go env GOOS)" || return 75
  receipt_goarch="$(go env GOARCH)" || return 75
  _receipt_safe_json_text "$receipt_go_version" >/dev/null || return 75
  _receipt_safe_json_text "$receipt_cgo_enabled" >/dev/null || return 75
  _receipt_safe_json_text "$receipt_goos" >/dev/null || return 75
  _receipt_safe_json_text "$receipt_goarch" >/dev/null || return 75
  receipt_env_godebug="$(_receipt_environment_observation GODEBUG)" || return 75
  receipt_env_goexperiment="$(_receipt_environment_observation GOEXPERIMENT)" || return 75
  receipt_env_gomemlimit="$(_receipt_environment_observation GOMEMLIMIT)" || return 75
  go_environment="{\"CGO_ENABLED\":\"$receipt_cgo_enabled\",\"GOARCH\":\"$receipt_goarch\",\"GOCACHE\":\"$GOCACHE\",\"GOENV\":\"off\",\"GOFLAGS\":\"-p=1\",\"GOMAXPROCS\":null,\"GOOS\":\"$receipt_goos\",\"GOPROXY\":\"off\",\"GOTMPDIR\":\"$GOTMPDIR\",\"GOTOOLCHAIN\":\"local\",\"GOWORK\":\"off\",\"go_version\":\"$receipt_go_version\",\"runtime_affecting_environment\":[$receipt_env_godebug,$receipt_env_goexperiment,$receipt_env_gomemlimit]}"
  go_environment_sha256="$(_receipt_sha256_text "$go_environment")" || return 75
}

_receipt_create_owner_run() {
  [ "$#" -eq 0 ] || return 64
  container_hostname="$(hostname)" || return 75
  printf '%s\n' "$container_hostname" | grep -Eq '^[0-9a-f]{12,64}$' || return 75
  owner_nonce="$(od -An -N16 -tx1 /dev/urandom | tr -d ' \n')" || return 75
  printf '%s\n' "$owner_nonce" | grep -Eq '^[0-9a-f]{32}$' || return 75
  receipt_timestamp="$(date -u '+%Y%m%dT%H%M%SZ')" || return 75
  printf '%s\n' "$receipt_timestamp" | grep -Eq '^[0-9]{8}T[0-9]{6}Z$' || return 75
  run_id="backend-unit-$receipt_timestamp-$container_hostname-${owner_nonce%????????????????????????}"
  printf '%s\n' "$run_id" | grep -Eq '^backend-unit-[0-9]{8}T[0-9]{6}Z-[0-9a-f]{12,64}-[0-9a-f]{8}$' || return 75
  run_dir="/reports/backend-unit-$container_hostname.partial"
  [ ! -e "$run_dir" ] && [ ! -L "$run_dir" ] || return 75
  mkdir -m 700 -- "$run_dir" || return 75
  receipt_run_identity="$(_owned_run_dir_identity)" || return 75
  owner_payload="{\"run_id\":\"$run_id\",\"owner_nonce\":\"$owner_nonce\",\"container_hostname\":\"$container_hostname\"}"
  write_atomic "$run_dir/owner.json" "$owner_payload" || return 75
  _receipt_validate_owner || return 75
}

_receipt_budget_json() {
  [ "$#" -eq 2 ] || return 64
  receipt_metadata_used="$1"
  receipt_total_used="$2"
  _receipt_decimal "$receipt_metadata_used" || return 75
  _receipt_decimal "$receipt_total_used" || return 75
  [ "$receipt_metadata_used" -le "$REPORT_METADATA_BUDGET_BYTES" ] || return 75
  [ "$receipt_total_used" -le "$REPORT_BUDGET_BYTES" ] || return 75
  [ "${report_budget_truncated_due_to_budget-}" = false ] || return 75
  printf '{"metadata_used_bytes":%s,"raw_output_partial":false,"reserved_metadata_bytes":%s,"streams":{"build":{"stderr":{"cap_bytes":%s,"truncated":false,"used_bytes":%s},"stdout":{"cap_bytes":%s,"truncated":false,"used_bytes":%s}},"race":{"stderr":{"cap_bytes":%s,"truncated":false,"used_bytes":%s},"stdout":{"cap_bytes":%s,"truncated":false,"used_bytes":%s}},"unit":{"stderr":{"cap_bytes":%s,"truncated":false,"used_bytes":%s},"stdout":{"cap_bytes":%s,"truncated":false,"used_bytes":%s}},"vet":{"stderr":{"cap_bytes":%s,"truncated":false,"used_bytes":%s},"stdout":{"cap_bytes":%s,"truncated":false,"used_bytes":%s}}},"total_bytes":%s,"total_used_bytes":%s,"truncated_due_to_budget":false}' "$receipt_metadata_used" "$REPORT_METADATA_BUDGET_BYTES" "$BUILD_STDERR_CAP_BYTES" "$build_stderr_size" "$BUILD_STDOUT_CAP_BYTES" "$build_stdout_size" "$RACE_STDERR_CAP_BYTES" "$race_stderr_size" "$RACE_STDOUT_CAP_BYTES" "$race_stdout_size" "$UNIT_STDERR_CAP_BYTES" "$unit_stderr_size" "$UNIT_STDOUT_CAP_BYTES" "$unit_stdout_size" "$VET_STDERR_CAP_BYTES" "$vet_stderr_size" "$VET_STDOUT_CAP_BYTES" "$vet_stdout_size" "$REPORT_BUDGET_BYTES" "$receipt_total_used"
}

_receipt_container_payload_is_canonical() {
  [ "$#" -eq 0 ] || return 64
  case "$receipt" in *'
') ;; *) return 75 ;; esac
  receipt_payload_body=${receipt%?}
  [ "$receipt_payload_body" != "$receipt" ] || return 75
  case "$receipt_payload_body" in *'
'*) return 75 ;; esac
  case "$receipt_payload_body" in
    '{"artifact_inventory":['*',"terminal":{"exit_code":'*',"signal":null}}') return 0 ;;
    *) return 75 ;;
  esac
}

_receipt_build_container_json() {
  [ "$#" -eq 1 ] || return 64
  receipt_budget_json="$1"
  receipt_cgroup_json="{\"baseline\":$container_cgroup_baseline_json,\"family\":\"$cgroup_family\",\"final\":$container_cgroup_final_json}"
  receipt="{\"artifact_inventory\":[$receipt_inventory_json],\"artifact_inventory_sha256\":\"$receipt_inventory_sha256\",\"cgroup\":$receipt_cgroup_json,\"command_profile_sha256\":\"$command_profile_sha256\",\"container_hostname\":\"$container_hostname\",\"go_environment\":$go_environment,\"launcher_sha256\":\"$launcher_sha256\",\"owner_nonce\":\"$owner_nonce\",\"package_plan_sha256\":{\"build\":\"$build_plan_sha256\",\"race\":\"$race_plan_sha256\",\"unit\":\"$unit_plan_sha256\",\"vet\":\"$vet_plan_sha256\"},\"report_budget\":$receipt_budget_json,\"run_id\":\"$run_id\",\"schema\":\"$RECEIPT_SCHEMA\",\"source_manifest_after_sha256\":\"$source_manifest_after_sha256\",\"source_manifest_before_sha256\":\"$source_manifest_before_sha256\",\"stage_receipt_sha256\":{\"build\":\"$build_stage_sha256\",\"race\":\"$race_stage_sha256\",\"unit\":\"$unit_stage_sha256\",\"vet\":\"$vet_stage_sha256\"},\"state\":\"CONTAINER_COMPLETE\",\"target\":\"$RECEIPT_TARGET\",\"terminal\":{\"exit_code\":$receipt_first_stage_exit,\"signal\":null}}"
  receipt="$receipt
"
  _receipt_container_payload_is_canonical
}

_receipt_converge_budget() {
  [ "$#" -eq 0 ] || return 64
  receipt_regular_before="$(report_regular_file_bytes)" || return 75
  receipt_metadata_before="$(report_metadata_bytes)" || return 75
  receipt_metadata_used="$receipt_metadata_before"
  receipt_total_used="$receipt_regular_before"
  receipt_iteration=0
  while [ "$receipt_iteration" -lt 5 ]; do
    receipt_iteration=$((receipt_iteration + 1))
    receipt_budget="$(_receipt_budget_json "$receipt_metadata_used" "$receipt_total_used")" || return 75
    _receipt_build_container_json "$receipt_budget" || return 75
    receipt_size="$(printf '%s' "$receipt" | wc -c)" || return 75
    _receipt_decimal "$receipt_size" || return 75
    receipt_next_metadata=$((receipt_metadata_before + receipt_size))
    receipt_next_total=$((receipt_regular_before + receipt_size))
    [ "$receipt_next_metadata" -le "$REPORT_METADATA_BUDGET_BYTES" ] || return 75
    [ "$receipt_next_total" -le "$REPORT_BUDGET_BYTES" ] || return 75
    if [ "$receipt_next_metadata" = "$receipt_metadata_used" ] && [ "$receipt_next_total" = "$receipt_total_used" ]; then
      return 0
    fi
    receipt_metadata_used="$receipt_next_metadata"
    receipt_total_used="$receipt_next_total"
  done
  return 75
}

_receipt_published_tree() {
  [ "$#" -eq 0 ] || return 64
  receipt_tree="$(find "$run_dir" -xdev -mindepth 1 -maxdepth 1 -printf '%f\t%y\t%n\n' | LC_ALL=C sort)" || return 75
  receipt_expected_tree='CONTAINER_COMPLETE	f	1
build.stderr.log	f	1
build.stdout.log	f	1
container-receipt.json	f	1
owner.json	f	1
package-plan.build.json	f	1
package-plan.race.json	f	1
package-plan.unit.json	f	1
package-plan.vet.json	f	1
package-results.build.jsonl	f	1
package-results.race.jsonl	f	1
package-results.unit.jsonl	f	1
package-results.vet.jsonl	f	1
race.stderr.log	f	1
race.stdout.log	f	1
stage-build.json	f	1
stage-race.json	f	1
stage-unit.json	f	1
stage-vet.json	f	1
unit.stderr.log	f	1
unit.stdout.log	f	1
vet.stderr.log	f	1
vet.stdout.log	f	1'
  [ "$receipt_tree" = "$receipt_expected_tree" ] || return 75
}

_receipt_cleanup_terminal() {
  [ "$#" -eq 0 ] || return 64
  receipt_cleanup_failed=false
  if [ -n "${receipt_marker_identity-}" ] && { [ -e "$receipt_marker_path" ] || [ -L "$receipt_marker_path" ]; }; then
    _receipt_remove_owned_regular "$receipt_marker_path" "$receipt_marker_identity" || receipt_cleanup_failed=true
  fi
  if [ -n "${receipt_final_identity-}" ] && { [ -e "$receipt_final_path" ] || [ -L "$receipt_final_path" ]; }; then
    _receipt_remove_owned_regular "$receipt_final_path" "$receipt_final_identity" || receipt_cleanup_failed=true
  fi
  [ "$receipt_cleanup_failed" = false ] || return 75
}

_receipt_abort_stage_is_completed() {
  [ "$#" -eq 1 ] || return 64
  case "$1" in
    build) [ "$receipt_abort_completed_count" -ge 1 ] ;;
    vet) [ "$receipt_abort_completed_count" -ge 2 ] ;;
    unit) [ "$receipt_abort_completed_count" -ge 3 ] ;;
    race) [ "$receipt_abort_completed_count" -ge 4 ] ;;
    *) return 64 ;;
  esac
}

_receipt_abort_name_is_required() {
  [ "$#" -eq 1 ] || return 64
  case "$1" in
    owner.json) return 0 ;;
    package-plan.build.json|package-plan.vet.json|package-plan.unit.json|package-plan.race.json)
      [ "$receipt_abort_phase" = execution ]
      ;;
    build.stderr.log|build.stdout.log|package-results.build.jsonl|stage-build.json) _receipt_abort_stage_is_completed build ;;
    vet.stderr.log|vet.stdout.log|package-results.vet.jsonl|stage-vet.json) _receipt_abort_stage_is_completed vet ;;
    unit.stderr.log|unit.stdout.log|package-results.unit.jsonl|stage-unit.json) _receipt_abort_stage_is_completed unit ;;
    race.stderr.log|race.stdout.log|package-results.race.jsonl|stage-race.json) _receipt_abort_stage_is_completed race ;;
    *) return 64 ;;
  esac
}

_receipt_abort_expected_tree() {
  [ "$#" -eq 2 ] || return 64
  case "$1:$2" in false:false|true:false|true:true) ;; *) return 64 ;; esac
  receipt_abort_marker_expected="$1"
  receipt_abort_final_expected="$2"
  receipt_abort_expected_tree=$(for receipt_abort_name in CONTAINER_ABORTED build.stderr.log build.stdout.log container-receipt.json owner.json package-plan.build.json package-plan.race.json package-plan.unit.json package-plan.vet.json package-results.build.jsonl package-results.race.jsonl package-results.unit.jsonl package-results.vet.jsonl race.stderr.log race.stdout.log stage-build.json stage-race.json stage-unit.json stage-vet.json unit.stderr.log unit.stdout.log vet.stderr.log vet.stdout.log; do
    case "$receipt_abort_name" in
      CONTAINER_ABORTED) [ "$receipt_abort_marker_expected" = true ] || continue ;;
      container-receipt.json) [ "$receipt_abort_final_expected" = true ] || continue ;;
      *) _receipt_abort_name_is_required "$receipt_abort_name" || continue ;;
    esac
    printf '%s\tf\t1\n' "$receipt_abort_name"
  done | LC_ALL=C sort) || return 75
  receipt_abort_actual_tree="$(find "$run_dir" -xdev -mindepth 1 -maxdepth 1 -printf '%f\t%y\t%n\n' | LC_ALL=C sort)" || return 75
  [ "$receipt_abort_actual_tree" = "$receipt_abort_expected_tree" ] || return 75
}

_receipt_create_aborted_marker() {
  [ "$#" -eq 0 ] || return 64
  receipt_marker_path="$run_dir/CONTAINER_ABORTED"
  _receipt_run_dir_current || return 75
  _run_dir_target_is_safe "$receipt_marker_path" || return 75
  [ ! -e "$receipt_marker_path" ] && [ ! -L "$receipt_marker_path" ] || return 75
  ( umask 077; set -C; : > "$receipt_marker_path" ) || return 75
  receipt_marker_identity="$(_owned_regular_identity "$receipt_marker_path")" || return 75
  receipt_marker_links="$(_owned_regular_links "$receipt_marker_path")" || return 75
  [ "$receipt_marker_links" = 1 ] || return 75
  receipt_marker_size="$(_receipt_file_size "$receipt_marker_path")" || return 75
  [ "$receipt_marker_size" = 0 ] || return 75
}

_receipt_abort_inventory() {
  [ "$#" -eq 0 ] || return 64
  receipt_inventory_json=
  receipt_inventory_lines=
  receipt_abort_inventory_identities=
  receipt_inventory_first=true
  for receipt_abort_name in CONTAINER_ABORTED build.stderr.log build.stdout.log owner.json package-plan.build.json package-plan.race.json package-plan.unit.json package-plan.vet.json package-results.build.jsonl package-results.race.jsonl package-results.unit.jsonl package-results.vet.jsonl race.stderr.log race.stdout.log stage-build.json stage-race.json stage-unit.json stage-vet.json unit.stderr.log unit.stdout.log vet.stderr.log vet.stdout.log; do
    case "$receipt_abort_name" in
      CONTAINER_ABORTED) ;;
      *) _receipt_abort_name_is_required "$receipt_abort_name" || continue ;;
    esac
    receipt_inventory_path="$run_dir/$receipt_abort_name"
    _receipt_regular_is_owned "$receipt_inventory_path" || return 75
    receipt_inventory_identity="$(_owned_regular_identity "$receipt_inventory_path")" || return 75
    receipt_inventory_links="$(_owned_regular_links "$receipt_inventory_path")" || return 75
    [ "$receipt_inventory_links" = 1 ] || return 75
    receipt_inventory_sha="$(_receipt_sha256_file "$receipt_inventory_path")" || return 75
    receipt_inventory_size="$(_receipt_file_size "$receipt_inventory_path")" || return 75
    if [ "$receipt_inventory_first" = true ]; then
      receipt_inventory_json="{\"path\":\"$receipt_abort_name\",\"sha256\":\"$receipt_inventory_sha\",\"size\":$receipt_inventory_size}"
      receipt_inventory_first=false
    else
      receipt_inventory_json="$receipt_inventory_json,{\"path\":\"$receipt_abort_name\",\"sha256\":\"$receipt_inventory_sha\",\"size\":$receipt_inventory_size}"
    fi
    receipt_inventory_lines="${receipt_inventory_lines}${receipt_abort_name}	${receipt_inventory_sha}	${receipt_inventory_size}
"
    receipt_abort_inventory_identities="${receipt_abort_inventory_identities}${receipt_abort_name}	${receipt_inventory_identity}	${receipt_inventory_sha}	${receipt_inventory_size}
"
  done
  receipt_inventory_sha256="$(_receipt_sha256_text "$receipt_inventory_lines")" || return 75
}

_receipt_abort_inventory_is_current() {
  [ "$#" -eq 0 ] || return 64
  [ -n "${receipt_abort_inventory_captured_identities-}" ] || return 75
  [ -n "${receipt_abort_inventory_captured_json-}" ] || return 75
  [ -n "${receipt_abort_inventory_captured_lines-}" ] || return 75
  [ -n "${receipt_abort_inventory_captured_sha256-}" ] || return 75
  _receipt_abort_inventory || return 75
  [ "$receipt_abort_inventory_identities" = "$receipt_abort_inventory_captured_identities" ] || return 75
  [ "$receipt_inventory_json" = "$receipt_abort_inventory_captured_json" ] || return 75
  [ "$receipt_inventory_lines" = "$receipt_abort_inventory_captured_lines" ] || return 75
  [ "$receipt_inventory_sha256" = "$receipt_abort_inventory_captured_sha256" ] || return 75
}

_receipt_abort_capture_stream_sizes() {
  [ "$#" -eq 0 ] || return 64
  build_stdout_size=0; build_stderr_size=0
  vet_stdout_size=0; vet_stderr_size=0
  unit_stdout_size=0; unit_stderr_size=0
  race_stdout_size=0; race_stderr_size=0
  for receipt_abort_stage in build vet unit race; do
    _receipt_abort_stage_is_completed "$receipt_abort_stage" || continue
    receipt_abort_stdout_path="$run_dir/$receipt_abort_stage.stdout.log"
    receipt_abort_stderr_path="$run_dir/$receipt_abort_stage.stderr.log"
    _receipt_regular_is_owned "$receipt_abort_stdout_path" || return 75
    _receipt_regular_is_owned "$receipt_abort_stderr_path" || return 75
    receipt_abort_stdout_size="$(_receipt_file_size "$receipt_abort_stdout_path")" || return 75
    receipt_abort_stderr_size="$(_receipt_file_size "$receipt_abort_stderr_path")" || return 75
    receipt_abort_stdout_cap="$(stage_stream_cap "$receipt_abort_stage" stdout)" || return 75
    receipt_abort_stderr_cap="$(stage_stream_cap "$receipt_abort_stage" stderr)" || return 75
    [ "$receipt_abort_stdout_size" -le "$receipt_abort_stdout_cap" ] || return 75
    [ "$receipt_abort_stderr_size" -le "$receipt_abort_stderr_cap" ] || return 75
    case "$receipt_abort_stage" in
      build) build_stdout_size="$receipt_abort_stdout_size"; build_stderr_size="$receipt_abort_stderr_size" ;;
      vet) vet_stdout_size="$receipt_abort_stdout_size"; vet_stderr_size="$receipt_abort_stderr_size" ;;
      unit) unit_stdout_size="$receipt_abort_stdout_size"; unit_stderr_size="$receipt_abort_stderr_size" ;;
      race) race_stdout_size="$receipt_abort_stdout_size"; race_stderr_size="$receipt_abort_stderr_size" ;;
    esac
  done
}

_receipt_abort_progress() {
  [ "$#" -eq 3 ] || return 64
  receipt_abort_phase="$1"
  receipt_abort_completed_prefix="$2"
  receipt_abort_active_stage="$3"
  case "$receipt_abort_phase:$receipt_abort_completed_prefix:$receipt_abort_active_stage" in
    bootstrap::null)
      receipt_abort_completed_count=0
      receipt_abort_completed_json='[]'
      receipt_abort_planned_json='[]'
      receipt_abort_active_json=null
      [ -z "${completed_build_final_identities-}" ] && [ -z "${completed_vet_final_identities-}" ] && [ -z "${completed_unit_final_identities-}" ] && [ -z "${completed_race_final_identities-}" ] || return 75
      ;;
    execution::build)
      receipt_abort_completed_count=0
      receipt_abort_completed_json='[]'
      receipt_abort_planned_json='["build","vet","unit","race"]'
      receipt_abort_active_json='"build"'
      [ -z "${completed_build_final_identities-}" ] && [ -z "${completed_vet_final_identities-}" ] && [ -z "${completed_unit_final_identities-}" ] && [ -z "${completed_race_final_identities-}" ] || return 75
      ;;
    execution:build:vet)
      _execute_completed_stage_final_identities_are_current build "${completed_build_final_identities-}" || return 75
      [ -z "${completed_vet_final_identities-}" ] && [ -z "${completed_unit_final_identities-}" ] && [ -z "${completed_race_final_identities-}" ] || return 75
      receipt_abort_completed_count=1
      receipt_abort_completed_json='["build"]'
      receipt_abort_planned_json='["build","vet","unit","race"]'
      receipt_abort_active_json='"vet"'
      ;;
    execution:build,vet:unit)
      _execute_completed_stage_final_identities_are_current build "${completed_build_final_identities-}" || return 75
      _execute_completed_stage_final_identities_are_current vet "${completed_vet_final_identities-}" || return 75
      [ -z "${completed_unit_final_identities-}" ] && [ -z "${completed_race_final_identities-}" ] || return 75
      receipt_abort_completed_count=2
      receipt_abort_completed_json='["build","vet"]'
      receipt_abort_planned_json='["build","vet","unit","race"]'
      receipt_abort_active_json='"unit"'
      ;;
    execution:build,vet,unit:race)
      _execute_completed_stage_final_identities_are_current build "${completed_build_final_identities-}" || return 75
      _execute_completed_stage_final_identities_are_current vet "${completed_vet_final_identities-}" || return 75
      _execute_completed_stage_final_identities_are_current unit "${completed_unit_final_identities-}" || return 75
      [ -z "${completed_race_final_identities-}" ] || return 75
      receipt_abort_completed_count=3
      receipt_abort_completed_json='["build","vet","unit"]'
      receipt_abort_planned_json='["build","vet","unit","race"]'
      receipt_abort_active_json='"race"'
      ;;
    execution:build,vet,unit,race:null)
      _execute_completed_stage_final_identities_are_current build "${completed_build_final_identities-}" || return 75
      _execute_completed_stage_final_identities_are_current vet "${completed_vet_final_identities-}" || return 75
      _execute_completed_stage_final_identities_are_current unit "${completed_unit_final_identities-}" || return 75
      _execute_completed_stage_final_identities_are_current race "${completed_race_final_identities-}" || return 75
      receipt_abort_completed_count=4
      receipt_abort_completed_json='["build","vet","unit","race"]'
      receipt_abort_planned_json='["build","vet","unit","race"]'
      receipt_abort_active_json=null
      ;;
    *) return 75 ;;
  esac
}

_receipt_aborted_payload_is_canonical() {
  [ "$#" -eq 0 ] || return 64
  case "$receipt" in *'
') ;; *) return 75 ;; esac
  receipt_payload_body=${receipt%?}
  [ "$receipt_payload_body" != "$receipt" ] || return 75
  case "$receipt_payload_body" in *'
'*) return 75 ;; esac
  case "$receipt_payload_body" in
    '{"abort":'*',"terminal":{"exit_code":143,"signal":"TERM"}}') return 0 ;;
    *) return 75 ;;
  esac
}

_receipt_build_aborted_container_json() {
  [ "$#" -eq 1 ] || return 64
  receipt_budget_json="$1"
  [ "$child_group_reaped" = true ] && [ "$stdout_reader_reaped" = true ] && [ "$stderr_reader_reaped" = true ] && [ "$unpublished_stage_rolled_back" = true ] || return 75
  receipt_abort_object="{\"active_stage\":$receipt_abort_active_json,\"child_group_reaped\":true,\"completed_stages\":$receipt_abort_completed_json,\"phase\":\"$receipt_abort_phase\",\"planned_stages\":$receipt_abort_planned_json,\"signal\":\"TERM\",\"stderr_reader_reaped\":true,\"stdout_reader_reaped\":true,\"unpublished_stage_rolled_back\":true}"
  receipt_cgroup_json="{\"baseline\":$container_cgroup_baseline_json,\"family\":\"$cgroup_family\",\"final\":$container_cgroup_final_json}"
  receipt="{\"abort\":$receipt_abort_object,\"artifact_inventory\":[$receipt_inventory_json],\"artifact_inventory_sha256\":\"$receipt_inventory_sha256\",\"cgroup\":$receipt_cgroup_json,\"command_profile_sha256\":\"$command_profile_sha256\",\"container_hostname\":\"$container_hostname\",\"go_environment\":$go_environment,\"launcher_sha256\":\"$launcher_sha256\",\"owner_nonce\":\"$owner_nonce\",\"report_budget\":$receipt_budget_json,\"run_id\":\"$run_id\",\"schema\":\"$RECEIPT_SCHEMA\",\"source_manifest_after_sha256\":\"$source_manifest_after_sha256\",\"source_manifest_before_sha256\":\"$source_manifest_before_sha256\",\"state\":\"CONTAINER_ABORTED\",\"target\":\"$RECEIPT_TARGET\",\"terminal\":{\"exit_code\":143,\"signal\":\"TERM\"}}"
  receipt="$receipt
"
  _receipt_aborted_payload_is_canonical
}

_receipt_converge_aborted_budget() {
  [ "$#" -eq 0 ] || return 64
  receipt_regular_before="$(report_regular_file_bytes)" || return 75
  receipt_metadata_before="$(report_metadata_bytes)" || return 75
  receipt_metadata_used="$receipt_metadata_before"
  receipt_total_used="$receipt_regular_before"
  receipt_iteration=0
  while [ "$receipt_iteration" -lt 16 ]; do
    receipt_iteration=$((receipt_iteration + 1))
    receipt_budget="$(_receipt_budget_json "$receipt_metadata_used" "$receipt_total_used")" || return 75
    _receipt_build_aborted_container_json "$receipt_budget" || return 75
    receipt_size="$(printf '%s' "$receipt" | wc -c)" || return 75
    _receipt_decimal "$receipt_size" || return 75
    receipt_next_metadata=$((receipt_metadata_before + receipt_size))
    receipt_next_total=$((receipt_regular_before + receipt_size))
    [ "$receipt_next_metadata" -le "$REPORT_METADATA_BUDGET_BYTES" ] || return 75
    [ "$receipt_next_total" -le "$REPORT_BUDGET_BYTES" ] || return 75
    if [ "$receipt_next_metadata" = "$receipt_metadata_used" ] && [ "$receipt_next_total" = "$receipt_total_used" ]; then
      return 0
    fi
    receipt_metadata_used="$receipt_next_metadata"
    receipt_total_used="$receipt_next_total"
  done
  return 75
}

_receipt_abort_common_evidence_is_current() {
  [ "$#" -eq 0 ] || return 64
  printf '%s\n' "${run_id-}" | grep -Eq '^backend-unit-[0-9]{8}T[0-9]{6}Z-[0-9a-f]{12,64}-[0-9a-f]{8}$' || return 75
  printf '%s\n' "${owner_nonce-}" | grep -Eq '^[0-9a-f]{32}$' || return 75
  printf '%s\n' "${container_hostname-}" | grep -Eq '^[0-9a-f]{12,64}$' || return 75
  for receipt_abort_hash in "${launcher_sha256-}" "${command_profile_sha256-}" "${go_environment_sha256-}"; do
    printf '%s\n' "$receipt_abort_hash" | grep -Eq '^sha256:[0-9a-f]{64}$' || return 75
  done
  [ "$(_receipt_sha256_file /usr/local/bin/backend-unit)" = "$launcher_sha256" ] || return 75
  [ "$(_receipt_sha256_text "$command_profile")" = "$command_profile_sha256" ] || return 75
  [ "$(_receipt_sha256_text "$go_environment")" = "$go_environment_sha256" ] || return 75
}

_receipt_abort_terminal_evidence_is_current() {
  [ "$#" -eq 0 ] || return 64
  _receipt_run_dir_current || return 75
  _owned_temp_is_current "$receipt_marker_path" "$receipt_marker_identity" 1 || return 75
  receipt_marker_size="$(_receipt_file_size "$receipt_marker_path")" || return 75
  [ "$receipt_marker_size" = 0 ] || return 75
  _receipt_abort_inventory_is_current || return 75
  _owned_temp_is_current "$receipt_final_path" "$receipt_final_identity" 1 || return 75
  receipt_abort_final_size="$(_receipt_file_size "$receipt_final_path")" || return 75
  [ "$receipt_abort_final_size" = "$receipt_abort_final_expected_size" ] || return 75
  [ "$(_receipt_sha256_file "$receipt_final_path")" = "$receipt_abort_final_expected_sha256" ] || return 75
}

finalize_aborted_container() {
  [ "$#" -eq 3 ] || return 64
  receipt_marker_path="$run_dir/CONTAINER_ABORTED"
  receipt_marker_identity=
  receipt_final_path="$run_dir/container-receipt.json"
  receipt_final_identity=
  _receipt_run_dir_current || return 75
  [ ! -e "$receipt_marker_path" ] && [ ! -L "$receipt_marker_path" ] || return 75
  [ ! -e "$receipt_final_path" ] && [ ! -L "$receipt_final_path" ] || return 75
  _receipt_validate_owner || return 75
  _receipt_abort_common_evidence_is_current || return 75
  [ "${child_group_reaped-}" = true ] && [ "${stdout_reader_reaped-}" = true ] && [ "${stderr_reader_reaped-}" = true ] && [ "${unpublished_stage_rolled_back-}" = true ] || return 75
  printf '%s\n' "${source_manifest_before_sha256-}" | grep -Eq '^sha256:[0-9a-f]{64}$' || return 75
  [ "$source_manifest_before_sha256" = "${source_manifest_after_sha256-}" ] || return 75
  _cgroup_snapshot_is_canonical "$container_cgroup_baseline" || return 75
  _cgroup_snapshot_is_canonical "$container_cgroup_final" || return 75
  [ "$container_cgroup_baseline" = "$container_cgroup_final" ] || return 75
  container_cgroup_baseline_json="$(_receipt_snapshot_object "$container_cgroup_baseline")" || return 75
  container_cgroup_final_json="$(_receipt_snapshot_object "$container_cgroup_final")" || return 75
  _receipt_abort_progress "$1" "$2" "$3" || return 75
  _receipt_abort_capture_stream_sizes || return 75
  _receipt_abort_expected_tree false false || return 75
  if _receipt_create_aborted_marker; then :; else _receipt_cleanup_terminal || return 75; return 75; fi
  if _receipt_abort_expected_tree true false && _receipt_abort_inventory && _receipt_converge_aborted_budget; then
    :
  else
    _receipt_cleanup_terminal || return 75
    return 75
  fi
  receipt_abort_inventory_captured_identities="$receipt_abort_inventory_identities"
  receipt_abort_inventory_captured_json="$receipt_inventory_json"
  receipt_abort_inventory_captured_lines="$receipt_inventory_lines"
  receipt_abort_inventory_captured_sha256="$receipt_inventory_sha256"
  receipt_abort_final_expected_size="$receipt_size"
  _receipt_decimal "$receipt_abort_final_expected_size" || {
    _receipt_cleanup_terminal || return 75
    return 75
  }
  receipt_abort_final_expected_sha256="$(_receipt_sha256_text "$receipt")" || {
    _receipt_cleanup_terminal || return 75
    return 75
  }
  if write_atomic "$receipt_final_path" "$receipt"; then
    :
  else
    _receipt_cleanup_terminal || return 75
    return 75
  fi
  receipt_final_identity="$(_owned_regular_identity "$receipt_final_path")" || {
    _receipt_cleanup_terminal || return 75
    return 75
  }
  receipt_final_links="$(_owned_regular_links "$receipt_final_path")" || {
    _receipt_cleanup_terminal || return 75
    return 75
  }
  if [ "$receipt_final_links" = 1 ] && _receipt_abort_terminal_evidence_is_current && _receipt_abort_expected_tree true true && check_budget; then
    :
  else
    _receipt_cleanup_terminal || return 75
    return 75
  fi
  receipt_final_total="$(report_regular_file_bytes)" || {
    _receipt_cleanup_terminal || return 75
    return 75
  }
  receipt_final_metadata="$(report_metadata_bytes)" || {
    _receipt_cleanup_terminal || return 75
    return 75
  }
  if [ "$receipt_final_total" = "$receipt_total_used" ] && [ "$receipt_final_metadata" = "$receipt_metadata_used" ] && _receipt_abort_expected_tree true true && check_budget && _receipt_abort_terminal_evidence_is_current; then
    return 0
  fi
  _receipt_cleanup_terminal || return 75
  return 75
}

_backend_unit_term_exact_plan_set_is_current() {
  [ "$#" -eq 0 ] || return 64
  _receipt_run_dir_current || return 75
  _receipt_validate_owner || return 75
  backend_unit_term_expected_tree=$(printf '%s\n' owner.json package-plan.build.json package-plan.race.json package-plan.unit.json package-plan.vet.json | LC_ALL=C sort) || return 75
  backend_unit_term_actual_tree=$(find "$run_dir" -xdev -mindepth 1 -maxdepth 1 -printf '%f\n' | LC_ALL=C sort) || return 75
  [ "$backend_unit_term_actual_tree" = "$backend_unit_term_expected_tree" ] || return 75
  for backend_unit_term_stage in build vet unit race; do
    backend_unit_term_rows="$(package_plan_rows_path "$backend_unit_term_stage")" || return 75
    _plan_validate_rows "$backend_unit_term_rows" || return 75
    _receipt_regular_is_owned "$run_dir/package-plan.$backend_unit_term_stage.json" || return 75
    _execute_package_plan_is_bound "$backend_unit_term_stage" "$backend_unit_term_rows" || return 75
  done
  # _execute_package_plan_is_bound cleans its GOTMPDIR probe before success.
  # Its path variables must be blank at the first quiescent checkpoint.
  execute_file_path=
  execute_file_identity=
  return 0
}

_backend_unit_term_abort_progress_from_completed_identities() {
  [ "$#" -eq 0 ] || return 64
  case "$backend_unit_term_phase" in
    bootstrap)
      [ -z "${completed_build_final_identities-}" ] && [ -z "${completed_vet_final_identities-}" ] && [ -z "${completed_unit_final_identities-}" ] && [ -z "${completed_race_final_identities-}" ] || return 75
      backend_unit_term_completed_prefix=
      backend_unit_term_active_stage=null
      ;;
    execution)
      if [ -z "${completed_build_final_identities-}" ]; then
        [ -z "${completed_vet_final_identities-}" ] && [ -z "${completed_unit_final_identities-}" ] && [ -z "${completed_race_final_identities-}" ] || return 75
        backend_unit_term_completed_prefix=
        backend_unit_term_active_stage=build
      elif [ -z "${completed_vet_final_identities-}" ]; then
        [ -z "${completed_unit_final_identities-}" ] && [ -z "${completed_race_final_identities-}" ] || return 75
        backend_unit_term_completed_prefix=build
        backend_unit_term_active_stage=vet
      elif [ -z "${completed_unit_final_identities-}" ]; then
        [ -z "${completed_race_final_identities-}" ] || return 75
        backend_unit_term_completed_prefix=build,vet
        backend_unit_term_active_stage=unit
      elif [ -z "${completed_race_final_identities-}" ]; then
        backend_unit_term_completed_prefix=build,vet,unit
        backend_unit_term_active_stage=race
      else
        backend_unit_term_completed_prefix=build,vet,unit,race
        backend_unit_term_active_stage=null
      fi
      ;;
    *) return 75 ;;
  esac
  return 0
}

finalize_container() {
  [ "$#" -eq 0 ] || return 64
  receipt_marker_path="$run_dir/CONTAINER_COMPLETE"
  receipt_marker_identity=
  receipt_final_path="$run_dir/container-receipt.json"
  receipt_final_identity=
  _receipt_run_dir_current || return 75
  [ ! -e "$receipt_marker_path" ] && [ ! -L "$receipt_marker_path" ] || return 75
  [ ! -e "$receipt_final_path" ] && [ ! -L "$receipt_final_path" ] || return 75
  _receipt_validate_owner || return 75
  _cgroup_snapshot_is_canonical "$container_cgroup_baseline" || return 75
  _cgroup_snapshot_is_canonical "$container_cgroup_final" || return 75
  [ "$container_cgroup_baseline" = "$container_cgroup_final" ] || return 75
  _receipt_snapshot_object "$container_cgroup_baseline" >/dev/null || return 75
  container_cgroup_baseline_json="$(_receipt_snapshot_object "$container_cgroup_baseline")" || return 75
  container_cgroup_final_json="$(_receipt_snapshot_object "$container_cgroup_final")" || return 75
  receipt_first_stage_exit=0
  for receipt_stage in build vet unit race; do _receipt_validate_stage "$receipt_stage" || return 75; done
  [ "$receipt_first_stage_exit" = "$backend_unit_first_exit" ] || return 75
  _receipt_precomplete_tree || return 75
  _receipt_create_complete_marker || return 75
  if _receipt_complete_tree && _receipt_inventory && _receipt_converge_budget; then
    :
  else
    _receipt_cleanup_terminal || return 75
    return 75
  fi
  if write_atomic "$run_dir/container-receipt.json" "$receipt"; then
    :
  else
    _receipt_cleanup_terminal || return 75
    return 75
  fi
  receipt_final_identity="$(_owned_regular_identity "$receipt_final_path")" || {
    _receipt_cleanup_terminal || return 75
    return 75
  }
  receipt_final_links="$(_owned_regular_links "$receipt_final_path")" || {
    _receipt_cleanup_terminal || return 75
    return 75
  }
  if [ "$receipt_final_links" = 1 ] && _receipt_published_tree && check_budget; then
    :
  else
    _receipt_cleanup_terminal || return 75
    return 75
  fi
  receipt_final_total="$(report_regular_file_bytes)" || {
    _receipt_cleanup_terminal || return 75
    return 75
  }
  receipt_final_metadata="$(report_metadata_bytes)" || {
    _receipt_cleanup_terminal || return 75
    return 75
  }
  if [ "$receipt_final_total" = "$receipt_total_used" ] && [ "$receipt_final_metadata" = "$receipt_metadata_used" ]; then
    return 0
  fi
  _receipt_cleanup_terminal || return 75
  return 75
}

backend_unit_handle_term() {
  # This is intentionally the first handler action: repeated TERM/INT/HUP
  # cannot interrupt rollback, receipt construction, or cleanup.
  trap '' TERM INT HUP
  case "$backend_unit_term_handling" in
    false) backend_unit_term_handling=true ;;
    *) exit 75 ;;
  esac
  case "$backend_unit_term_safe:$backend_unit_term_phase" in
    true:bootstrap|true:execution) ;;
    *) exit 75 ;;
  esac
  case "${active_stage_for_container_stop-}" in
    true) abort_active_stage_for_container_stop || exit 75 ;;
    false) prove_quiescent_for_container_stop || exit 75 ;;
    *) exit 75 ;;
  esac
  [ "${child_group_reaped-}" = true ] && [ "${stdout_reader_reaped-}" = true ] && [ "${stderr_reader_reaped-}" = true ] && [ "${unpublished_stage_rolled_back-}" = true ] || exit 75
  _backend_unit_term_abort_progress_from_completed_identities || exit 75
  verify_cgroup_snapshot_unchanged "$container_cgroup_baseline" || exit 75
  container_cgroup_final="$(capture_cgroup_snapshot)" || exit 75
  [ "$container_cgroup_baseline" = "$container_cgroup_final" ] || exit 75
  source_manifest_after_sha256="$(_receipt_source_manifest)" || exit 75
  [ "$source_manifest_before_sha256" = "$source_manifest_after_sha256" ] || exit 75
  finalize_aborted_container "$backend_unit_term_phase" "$backend_unit_term_completed_prefix" "$backend_unit_term_active_stage" || exit 75
  exit 143
}

backend_unit_main() {
  [ "$#" -eq 0 ] || return 64
  assert_cgroup || return 75
  assert_owned_tmpfs || return 75
  _receipt_create_owner_run || return 75
  _receipt_capture_go_environment || return 75
  launcher_sha256="$(_receipt_sha256_file /usr/local/bin/backend-unit)" || return 75
  command_profile='{"build":["go","build","-p=1","-o","<owned-output>","<package>"],"vet":["go","vet","-tags","hubtest","-p=1","<package>"],"unit":["go","test","-tags","hubtest","-json","-p=1","-count=1","-timeout=150s","<package>"],"race":["go","test","-tags","hubtest","-json","-race","-p=1","-count=1","-timeout=150s","<package>"]}'
  command_profile_sha256="$(_receipt_sha256_text "$command_profile")" || return 75
  source_manifest_before_sha256="$(_receipt_source_manifest)" || return 75
  container_cgroup_baseline="$(capture_cgroup_snapshot)" || return 75
  backend_unit_term_phase=bootstrap
  backend_unit_term_safe=true
  backend_unit_term_phase=planning
  backend_unit_term_safe=false
  prepare_package_plans || return 75
  _backend_unit_term_exact_plan_set_is_current || return 75
  backend_unit_term_phase=execution
  backend_unit_term_safe=true
  backend_unit_first_exit=0
  for backend_unit_stage in build vet unit race; do
    backend_unit_rows="$(package_plan_rows_path "$backend_unit_stage")" || return 75
    if run_stage "$backend_unit_stage" "$backend_unit_rows"; then
      backend_unit_stage_exit=0
    else
      backend_unit_stage_exit=$?
    fi
    case "$backend_unit_stage_exit" in
      0) ;;
      75) return 75 ;;
      [1-9]|[1-9][0-9]|[1-9][0-9][0-9])
        if [ "$backend_unit_first_exit" -eq 0 ]; then backend_unit_first_exit="$backend_unit_stage_exit"; fi
        ;;
      *) return 75 ;;
    esac
  done
  verify_cgroup_snapshot_unchanged "$container_cgroup_baseline" || return 75
  container_cgroup_final="$(capture_cgroup_snapshot)" || return 75
  source_manifest_after_sha256="$(_receipt_source_manifest)" || return 75
  backend_unit_term_phase=finalizing
  backend_unit_term_safe=false
  finalize_container || return 75
  return "$backend_unit_first_exit"
}
