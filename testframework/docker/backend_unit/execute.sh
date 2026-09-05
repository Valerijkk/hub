#!/bin/sh
set -eu

active_stage_for_container_stop=false
active_stage_abort_state=idle
active_stage_artifacts_ready=false
active_stage_publish_in_progress=false
child_group_reaped=false
stdout_reader_reaped=false
stderr_reader_reaped=false
unpublished_stage_rolled_back=false
stdout_reader_started=false
stderr_reader_started=false
stdout_reader_pid=
stderr_reader_pid=
stdout_reader_identity=
stderr_reader_identity=
completed_build_final_identities=
completed_vet_final_identities=
completed_unit_final_identities=
completed_race_final_identities=

_execute_identity_is_exact() {
  [ "$#" -eq 5 ] || return 64
  execute_identity_value="$1"
  execute_identity_expected_uid="$2"
  execute_identity_expected_gid="$3"
  execute_identity_expected_mode="$4"
  execute_identity_expected_type="$5"
  execute_identity_tail="$execute_identity_value"
  for execute_identity_index in 1 2 3 4 5; do
    case "$execute_identity_tail" in
      *:*) execute_identity_field=${execute_identity_tail%%:*}; execute_identity_tail=${execute_identity_tail#*:} ;;
      *) return 75 ;;
    esac
    case "$execute_identity_field" in 0|[1-9][0-9]*) ;; *) return 75 ;; esac
    case "$execute_identity_index" in
      1) execute_identity_device="$execute_identity_field" ;;
      2) execute_identity_inode="$execute_identity_field" ;;
      3) execute_identity_uid="$execute_identity_field" ;;
      4) execute_identity_gid="$execute_identity_field" ;;
      5) execute_identity_mode="$execute_identity_field" ;;
      *) return 75 ;;
    esac
  done
  [ -n "$execute_identity_tail" ] || return 75
  case "$execute_identity_tail" in *:*) return 75 ;; esac
  [ "$execute_identity_uid" = "$execute_identity_expected_uid" ] || return 75
  [ "$execute_identity_gid" = "$execute_identity_expected_gid" ] || return 75
  [ "$execute_identity_mode" = "$execute_identity_expected_mode" ] || return 75
  [ "$execute_identity_tail" = "$execute_identity_expected_type" ] || return 75
  return 0
}

_execute_regular_bytes() {
  [ "$#" -eq 1 ] || return 64
  [ -f "$1" ] && [ ! -L "$1" ] || return 75
  execute_bytes="$(wc -c < "$1")" || return 75
  case "$execute_bytes" in ''|*[!0-9]*) return 75 ;; esac
  printf '%s\n' "$execute_bytes"
}

_execute_regular_sha256() {
  [ "$#" -eq 1 ] || return 64
  [ -f "$1" ] && [ ! -L "$1" ] || return 75
  execute_sha256="$(sha256sum -- "$1" | awk '{print $1}')" || return 75
  printf '%s\n' "$execute_sha256" | grep -Eq '^[0-9a-f]{64}$' || return 75
  printf 'sha256:%s\n' "$execute_sha256"
}

_execute_package_plan_is_bound() {
  [ "$#" -eq 2 ] || return 64
  execute_plan_stage="$1"
  execute_plan_rows="$2"
  _plan_stage_is_valid "$execute_plan_stage" || return 64
  _plan_require_context || return 75
  _plan_validate_rows "$execute_plan_rows" || return 75
  execute_plan_path="$run_dir/package-plan.$execute_plan_stage.json"
  execute_plan_run_identity="$(_owned_run_dir_identity)" || return 75
  execute_plan_identity="$(_owned_regular_identity "$execute_plan_path")" || return 75
  execute_plan_discovery="$(_plan_discovery_argv_json "$execute_plan_stage")" || return 75
  execute_plan_expected=$(
    printf '{"schema":"%s","target":"%s","run_id":"%s","owner_nonce":"%s","stage":"%s","discovery_argv":%s,"go_environment_sha256":"%s","rows":[' \
      "$PACKAGE_PLAN_SCHEMA" "$PACKAGE_PLAN_TARGET" "$run_id" "$owner_nonce" "$execute_plan_stage" "$execute_plan_discovery" "$go_environment_sha256"
    execute_plan_first=true
    execute_plan_ordinal=0
    execute_plan_tab="$(printf '\011')"
    while IFS= read -r execute_plan_row; do
      execute_plan_ordinal=$((execute_plan_ordinal + 1))
      execute_plan_import=${execute_plan_row%%"$execute_plan_tab"*}
      execute_plan_package=${execute_plan_row#*"$execute_plan_tab"}
      if [ "$execute_plan_first" = true ]; then execute_plan_first=false; else printf ','; fi
      printf '{"ordinal":%s,"import_path":"%s","package_name":"%s"}' "$execute_plan_ordinal" "$execute_plan_import" "$execute_plan_package"
    done < "$execute_plan_rows"
    printf '],"row_count":%s,"row_set_sha256":"%s","result_partition":[' "$plan_row_count" "$plan_row_set_sha256"
    execute_plan_first=true
    execute_plan_ordinal=0
    while IFS= read -r execute_plan_row; do
      execute_plan_ordinal=$((execute_plan_ordinal + 1))
      execute_plan_import=${execute_plan_row%%"$execute_plan_tab"*}
      if [ "$execute_plan_first" = true ]; then execute_plan_first=false; else printf ','; fi
      printf '{"ordinal":%s,"import_path":"%s","result":"%s"}' "$execute_plan_ordinal" "$execute_plan_import" "$PACKAGE_PLAN_RESULT_PENDING"
    done < "$execute_plan_rows"
    printf ']}'
  ) || return 75
  _execute_create_gotmp_file || return 75
  execute_plan_check_path="$execute_file_path"
  execute_plan_check_identity="$execute_file_identity"
  if ! printf '%s' "$execute_plan_expected" > "$execute_plan_check_path"; then
    _cleanup_owned_temp "$execute_plan_check_path" "$execute_plan_check_identity" 1 || return 75
    return 75
  fi
  if ! _owned_temp_is_current "$execute_plan_check_path" "$execute_plan_check_identity" 1; then
    _cleanup_owned_temp "$execute_plan_check_path" "$execute_plan_check_identity" 1 || return 75
    return 75
  fi
  if ! execute_plan_current_identity="$(_owned_regular_identity "$execute_plan_path")"; then
    _cleanup_owned_temp "$execute_plan_check_path" "$execute_plan_check_identity" 1 || return 75
    return 75
  fi
  if _owned_run_dir_is_current "$execute_plan_run_identity" && [ "$execute_plan_current_identity" = "$execute_plan_identity" ] && cmp -s -- "$execute_plan_check_path" "$execute_plan_path"; then
    _cleanup_owned_temp "$execute_plan_check_path" "$execute_plan_check_identity" 1 || return 75
    return 0
  fi
  _cleanup_owned_temp "$execute_plan_check_path" "$execute_plan_check_identity" 1 || return 75
  return 75
}

_execute_create_gotmp_file() {
  [ "$#" -eq 0 ] || return 64
  _create_owned_temp "$GOTMPDIR" budget-all || return 75
  execute_file_path="$owned_temp_path"
  execute_file_identity="$owned_temp_identity"
  _owned_temp_is_current "$execute_file_path" "$execute_file_identity" 1 || return 75
}

_execute_create_run_file() {
  [ "$#" -eq 0 ] || return 64
  _owned_run_dir_identity >/dev/null || return 75
  _create_owned_temp "$run_dir" publish || return 75
  execute_file_path="$owned_temp_path"
  execute_file_identity="$owned_temp_identity"
  _owned_temp_is_current "$execute_file_path" "$execute_file_identity" 1 || return 75
}

_execute_append_owned_file() {
  [ "$#" -eq 3 ] || return 64
  execute_append_path="$1"
  execute_append_identity="$2"
  execute_append_payload="$3"
  _owned_temp_is_current "$execute_append_path" "$execute_append_identity" 1 || return 75
  printf '%s\n' "$execute_append_payload" >> "$execute_append_path" || return 75
  _owned_temp_is_current "$execute_append_path" "$execute_append_identity" 1 || return 75
}

_execute_append_capture() {
  [ "$#" -eq 4 ] || return 64
  execute_capture_path="$1"
  execute_capture_identity="$2"
  execute_output_path="$3"
  execute_output_identity="$4"
  _owned_temp_is_current "$execute_capture_path" "$execute_capture_identity" 1 || return 75
  _owned_temp_is_current "$execute_output_path" "$execute_output_identity" 1 || return 75
  cat -- "$execute_capture_path" >> "$execute_output_path" || return 75
  _owned_temp_is_current "$execute_output_path" "$execute_output_identity" 1 || return 75
  _cleanup_owned_temp "$execute_capture_path" "$execute_capture_identity" 1 || return 75
}

_execute_stream_capture_fits_stage_cap() {
  [ "$#" -eq 6 ] || return 64
  execute_stream_stage="$1"
  execute_stream_name="$2"
  execute_stream_output="$3"
  execute_stream_output_identity="$4"
  execute_stream_capture="$5"
  execute_stream_capture_identity="$6"
  execute_stream_cap="$(stage_stream_cap "$execute_stream_stage" "$execute_stream_name")" || return 75
  _owned_temp_is_current "$execute_stream_output" "$execute_stream_output_identity" 1 || return 75
  _owned_temp_is_current "$execute_stream_capture" "$execute_stream_capture_identity" 1 || return 75
  execute_stream_output_bytes="$(_execute_regular_bytes "$execute_stream_output")" || return 75
  execute_stream_capture_bytes="$(_execute_regular_bytes "$execute_stream_capture")" || return 75
  [ $((execute_stream_output_bytes + execute_stream_capture_bytes)) -le "$execute_stream_cap" ] || return 75
}

_execute_remaining_stage_stream_capacity() {
  [ "$#" -eq 4 ] || return 64
  execute_remaining_stage="$1"
  execute_remaining_name="$2"
  execute_remaining_output="$3"
  execute_remaining_output_identity="$4"
  execute_remaining_cap="$(stage_stream_cap "$execute_remaining_stage" "$execute_remaining_name")" || return 75
  case "$execute_remaining_cap" in ''|*[!0-9]*) return 75 ;; esac
  _owned_temp_is_current "$execute_remaining_output" "$execute_remaining_output_identity" 1 || return 75
  execute_remaining_used="$(_execute_regular_bytes "$execute_remaining_output")" || return 75
  case "$execute_remaining_used" in ''|*[!0-9]*) return 75 ;; esac
  [ "$execute_remaining_used" -le "$execute_remaining_cap" ] || return 75
  execute_remaining_bytes=$((execute_remaining_cap - execute_remaining_used))
  [ "$execute_remaining_bytes" -ge 0 ] || return 75
  [ "$execute_remaining_bytes" -le "$execute_remaining_cap" ] || return 75
  printf '%s\n' "$execute_remaining_bytes"
}

_execute_scratch_identity() {
  [ "$#" -eq 1 ] || return 64
  execute_scratch_path="$1"
  [ -d "$execute_scratch_path" ] && [ ! -L "$execute_scratch_path" ] || return 75
  execute_scratch_value="$(stat -c '%d:%i:%u:%g:%a:%F' -- "$execute_scratch_path")" || return 75
  _execute_identity_is_exact "$execute_scratch_value" 10001 10001 700 directory || return 75
  printf '%s\n' "$execute_scratch_value"
}

_execute_scratch_is_current() {
  [ "$#" -eq 2 ] || return 64
  execute_scratch_current="$(_execute_scratch_identity "$1")" || return 75
  [ "$execute_scratch_current" = "$2" ] || return 75
}

_execute_create_scratch() {
  [ "$#" -eq 1 ] || return 64
  execute_stage="$1"
  stage_scratch="$(umask 077; mktemp -d "$GOTMPDIR/.backend-unit-stage-$execute_stage.XXXXXXXX")" || return 75
  case "$stage_scratch" in "$GOTMPDIR/.backend-unit-stage-$execute_stage."[A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9]) ;; *) return 75 ;; esac
  stage_scratch_identity="$(_execute_scratch_identity "$stage_scratch")" || return 75
}

_execute_create_scratch_file() {
  [ "$#" -eq 1 ] || return 64
  _execute_scratch_is_current "$stage_scratch" "$stage_scratch_identity" || return 75
  execute_scratch_label="$1"
  case "$execute_scratch_label" in marker|stdout-capture|stderr-capture) ;; *) return 64 ;; esac
  execute_scratch_file="$(umask 077; mktemp "$stage_scratch/$execute_scratch_label.XXXXXXXX")" || return 75
  _owned_regular_identity "$execute_scratch_file" >/dev/null || return 75
  execute_scratch_file_identity="$(_owned_regular_identity "$execute_scratch_file")" || return 75
  _owned_temp_is_current "$execute_scratch_file" "$execute_scratch_file_identity" 1 || return 75
}

_execute_fifo_identity() {
  [ "$#" -eq 1 ] || return 64
  execute_fifo_path="$1"
  [ -p "$execute_fifo_path" ] && [ ! -L "$execute_fifo_path" ] || return 75
  execute_fifo_value="$(stat -c '%d:%i:%u:%g:%a:%F' -- "$execute_fifo_path")" || return 75
  _execute_identity_is_exact "$execute_fifo_value" 10001 10001 600 fifo || return 75
  printf '%s\n' "$execute_fifo_value"
}

_execute_create_fifo() {
  [ "$#" -eq 1 ] || return 64
  execute_fifo_path="$1"
  _execute_scratch_is_current "$stage_scratch" "$stage_scratch_identity" || return 75
  [ ! -e "$execute_fifo_path" ] && [ ! -L "$execute_fifo_path" ] || return 75
  ( umask 077; mkfifo -- "$execute_fifo_path" ) || return 75
  execute_fifo_identity="$(_execute_fifo_identity "$execute_fifo_path")" || return 75
}

_execute_fifo_is_current() {
  [ "$#" -eq 2 ] || return 64
  execute_fifo_current="$(_execute_fifo_identity "$1")" || return 75
  [ "$execute_fifo_current" = "$2" ] || return 75
}

_execute_remove_fifo() {
  [ "$#" -eq 2 ] || return 64
  _execute_scratch_is_current "$stage_scratch" "$stage_scratch_identity" || return 75
  _execute_fifo_is_current "$1" "$2" || return 75
  rm -- "$1" || return 75
  [ ! -e "$1" ] && [ ! -L "$1" ] || return 75
}

_execute_mark_budget_stop() {
  [ "$#" -eq 4 ] || return 64
  execute_marker_template="$1"
  execute_marker_identity="$2"
  execute_marker_path="$3"
  execute_marker_scratch_identity="$4"
  _execute_scratch_is_current "$stage_scratch" "$execute_marker_scratch_identity" || return 75
  if [ -e "$execute_marker_path" ] || [ -L "$execute_marker_path" ]; then
    _owned_temp_is_current "$execute_marker_path" "$execute_marker_identity" 2 || return 75
    return 0
  fi
  _owned_temp_is_current "$execute_marker_template" "$execute_marker_identity" 1 || return 75
  ln -- "$execute_marker_template" "$execute_marker_path" || return 75
  _owned_temp_is_current "$execute_marker_template" "$execute_marker_identity" 2 || return 75
  _owned_temp_is_current "$execute_marker_path" "$execute_marker_identity" 2 || return 75
}

_execute_signal_reader_abort() {
  [ "$#" -eq 2 ] || return 64
  _execute_mark_budget_stop "$stage_marker_template" "$2" "$1" "$stage_scratch_identity" || return 75
  return 0
}

read_bounded_stage_stream() {
  [ "$#" -eq 9 ] || return 64
  stream_stage="$1"
  stream_name="$2"
  stream_remaining="$3"
  stream_fifo="$4"
  stream_fifo_identity="$5"
  stream_capture="$6"
  stream_capture_identity="$7"
  stream_marker="$8"
  stream_marker_identity="$9"
  if stream_cap="$(stage_stream_cap "$stream_stage" "$stream_name")"; then :; else _execute_signal_reader_abort "$stream_marker" "$stream_marker_identity"; return 75; fi
  case "$stream_cap" in ''|*[!0-9]*) _execute_signal_reader_abort "$stream_marker" "$stream_marker_identity"; return 75 ;; esac
  case "$stream_remaining" in ''|*[!0-9]*) _execute_signal_reader_abort "$stream_marker" "$stream_marker_identity"; return 75 ;; esac
  if [ "$stream_remaining" -le "$stream_cap" ]; then :; else _execute_signal_reader_abort "$stream_marker" "$stream_marker_identity"; return 75; fi
  stream_read_limit=$((stream_remaining + 1))
  if _execute_fifo_is_current "$stream_fifo" "$stream_fifo_identity"; then :; else _execute_signal_reader_abort "$stream_marker" "$stream_marker_identity"; return 75; fi
  if _owned_temp_is_current "$stream_capture" "$stream_capture_identity" 1; then :; else _execute_signal_reader_abort "$stream_marker" "$stream_marker_identity"; return 75; fi
  if head -c "$stream_read_limit" < "$stream_fifo" > "$stream_capture" && _owned_temp_is_current "$stream_capture" "$stream_capture_identity" 1; then
    :
  else
    _execute_signal_reader_abort "$stream_marker" "$stream_marker_identity"
    return 75
  fi
  if stream_bytes="$(_execute_regular_bytes "$stream_capture")"; then :; else _execute_signal_reader_abort "$stream_marker" "$stream_marker_identity"; return 75; fi
  if [ "$stream_bytes" -gt "$stream_remaining" ]; then
    _execute_signal_reader_abort "$stream_marker" "$stream_marker_identity"
    return 75
  fi
  return 0
}

_execute_parse_plan_row() {
  [ "$#" -eq 1 ] || return 64
  execute_row="$1"
  execute_tab="$(printf '\011')"
  case "$execute_row" in *"$execute_tab"*) ;; *) return 75 ;; esac
  package_import_path=${execute_row%%"$execute_tab"*}
  package_name=${execute_row#*"$execute_tab"}
  case "$package_name" in *"$execute_tab"*) return 75 ;; esac
  _plan_import_path_is_valid "$package_import_path" || return 75
  _plan_package_name_is_valid "$package_name" || return 75
}

_execute_publish_owned_file() {
  [ "$#" -eq 4 ] || return 64
  execute_publish_temp="$1"
  execute_publish_identity="$2"
  execute_publish_final="$3"
  execute_publish_run_identity="$4"
  _owned_run_dir_is_current "$execute_publish_run_identity" || return 75
  _run_dir_target_is_safe "$execute_publish_final" || return 75
  [ ! -e "$execute_publish_final" ] && [ ! -L "$execute_publish_final" ] || return 75
  _owned_temp_is_current "$execute_publish_temp" "$execute_publish_identity" 1 || return 75
  ln -- "$execute_publish_temp" "$execute_publish_final" || return 75
  _owned_run_dir_is_current "$execute_publish_run_identity" || return 75
  _owned_temp_is_current "$execute_publish_temp" "$execute_publish_identity" 2 || return 75
  _owned_temp_is_current "$execute_publish_final" "$execute_publish_identity" 2 || return 75
  _cleanup_owned_temp "$execute_publish_temp" "$execute_publish_identity" 2 || return 75
  _owned_temp_is_current "$execute_publish_final" "$execute_publish_identity" 1 || return 75
}

_execute_rollback_stage_pair() {
  [ "$#" -eq 5 ] || return 64
  execute_rollback_temp="$1"
  execute_rollback_identity="$2"
  execute_rollback_final="$3"
  execute_rollback_created="$4"
  execute_rollback_final_identity="$5"
  execute_rollback_ok=true
  case "$execute_rollback_created" in true|false) ;; *) return 75 ;; esac
  if [ "$execute_rollback_created" = true ]; then
    if [ -e "$execute_rollback_final" ] || [ -L "$execute_rollback_final" ]; then
      if _owned_temp_is_current "$execute_rollback_final" "$execute_rollback_final_identity" 1; then
        _cleanup_owned_temp "$execute_rollback_final" "$execute_rollback_final_identity" 1 || execute_rollback_ok=false
      elif _owned_temp_is_current "$execute_rollback_final" "$execute_rollback_final_identity" 2; then
        _cleanup_owned_temp "$execute_rollback_final" "$execute_rollback_final_identity" 2 || execute_rollback_ok=false
      else
        execute_rollback_ok=false
      fi
    else
      execute_rollback_ok=false
    fi
  fi
  if [ -e "$execute_rollback_temp" ] || [ -L "$execute_rollback_temp" ]; then
    _cleanup_owned_temp "$execute_rollback_temp" "$execute_rollback_identity" 1 || execute_rollback_ok=false
  fi
  [ "$execute_rollback_ok" = true ] || return 75
  return 0
}

_execute_rollback_stage_transaction() {
  [ "$#" -eq 0 ] || return 64
  execute_transaction_rollback_ok=true
  if _execute_rollback_stage_pair "$txn_results_tmp" "$txn_results_identity" "$txn_results_final" "$txn_results_created" "$txn_results_final_identity"; then :; else execute_transaction_rollback_ok=false; fi
  if _execute_rollback_stage_pair "$txn_stdout_tmp" "$txn_stdout_identity" "$txn_stdout_final" "$txn_stdout_created" "$txn_stdout_final_identity"; then :; else execute_transaction_rollback_ok=false; fi
  if _execute_rollback_stage_pair "$txn_stderr_tmp" "$txn_stderr_identity" "$txn_stderr_final" "$txn_stderr_created" "$txn_stderr_final_identity"; then :; else execute_transaction_rollback_ok=false; fi
  if _execute_rollback_stage_pair "$txn_json_tmp" "$txn_json_identity" "$txn_json_final" "$txn_json_created" "$txn_json_final_identity"; then :; else execute_transaction_rollback_ok=false; fi
  if [ -e "$stage_scratch" ] || [ -L "$stage_scratch" ]; then
    if _execute_scratch_is_current "$stage_scratch" "$stage_scratch_identity"; then
      rmdir -- "$stage_scratch" || execute_transaction_rollback_ok=false
    else
      execute_transaction_rollback_ok=false
    fi
  fi
  [ "$execute_transaction_rollback_ok" = true ] || return 75
  return 0
}

_execute_publish_stage_transaction() {
  [ "$#" -eq 13 ] || return 64
  txn_results_tmp="$1"
  txn_results_identity="$2"
  txn_results_final="$3"
  txn_stdout_tmp="$4"
  txn_stdout_identity="$5"
  txn_stdout_final="$6"
  txn_stderr_tmp="$7"
  txn_stderr_identity="$8"
  txn_stderr_final="$9"
  shift 9
  txn_json_tmp="$1"
  txn_json_identity="$2"
  txn_json_final="$3"
  txn_run_identity="$4"
  txn_results_created=false
  txn_stdout_created=false
  txn_stderr_created=false
  txn_json_created=false
  txn_results_final_identity=
  txn_stdout_final_identity=
  txn_stderr_final_identity=
  txn_json_final_identity=
  if _owned_run_dir_is_current "$txn_run_identity" && _run_dir_target_is_safe "$txn_results_final" && _run_dir_target_is_safe "$txn_stdout_final" && _run_dir_target_is_safe "$txn_stderr_final" && _run_dir_target_is_safe "$txn_json_final" && [ ! -e "$txn_results_final" ] && [ ! -L "$txn_results_final" ] && [ ! -e "$txn_stdout_final" ] && [ ! -L "$txn_stdout_final" ] && [ ! -e "$txn_stderr_final" ] && [ ! -L "$txn_stderr_final" ] && [ ! -e "$txn_json_final" ] && [ ! -L "$txn_json_final" ] && _owned_temp_is_current "$txn_results_tmp" "$txn_results_identity" 1 && _owned_temp_is_current "$txn_stdout_tmp" "$txn_stdout_identity" 1 && _owned_temp_is_current "$txn_stderr_tmp" "$txn_stderr_identity" 1 && _owned_temp_is_current "$txn_json_tmp" "$txn_json_identity" 1; then
    :
  else
    _execute_rollback_stage_transaction || return 75
    return 75
  fi
  if _owned_run_dir_is_current "$txn_run_identity" && [ ! -e "$txn_results_final" ] && [ ! -L "$txn_results_final" ] && _owned_temp_is_current "$txn_results_tmp" "$txn_results_identity" 1; then
    :
  else
    _execute_rollback_stage_transaction || return 75
    return 75
  fi
  if ln -- "$txn_results_tmp" "$txn_results_final"; then
    txn_results_created=true
    txn_results_final_identity="$txn_results_identity"
    if _owned_run_dir_is_current "$txn_run_identity" && _owned_temp_is_current "$txn_results_tmp" "$txn_results_identity" 2 && _owned_temp_is_current "$txn_results_final" "$txn_results_identity" 2; then
      :
    else
      _execute_rollback_stage_transaction || return 75
      return 75
    fi
  else
    _execute_rollback_stage_transaction || return 75
    return 75
  fi
  if _owned_run_dir_is_current "$txn_run_identity" && [ ! -e "$txn_stdout_final" ] && [ ! -L "$txn_stdout_final" ] && _owned_temp_is_current "$txn_stdout_tmp" "$txn_stdout_identity" 1; then
    :
  else
    _execute_rollback_stage_transaction || return 75
    return 75
  fi
  if ln -- "$txn_stdout_tmp" "$txn_stdout_final"; then
    txn_stdout_created=true
    txn_stdout_final_identity="$txn_stdout_identity"
    if _owned_run_dir_is_current "$txn_run_identity" && _owned_temp_is_current "$txn_stdout_tmp" "$txn_stdout_identity" 2 && _owned_temp_is_current "$txn_stdout_final" "$txn_stdout_identity" 2; then
      :
    else
      _execute_rollback_stage_transaction || return 75
      return 75
    fi
  else
    _execute_rollback_stage_transaction || return 75
    return 75
  fi
  if _owned_run_dir_is_current "$txn_run_identity" && [ ! -e "$txn_stderr_final" ] && [ ! -L "$txn_stderr_final" ] && _owned_temp_is_current "$txn_stderr_tmp" "$txn_stderr_identity" 1; then
    :
  else
    _execute_rollback_stage_transaction || return 75
    return 75
  fi
  if ln -- "$txn_stderr_tmp" "$txn_stderr_final"; then
    txn_stderr_created=true
    txn_stderr_final_identity="$txn_stderr_identity"
    if _owned_run_dir_is_current "$txn_run_identity" && _owned_temp_is_current "$txn_stderr_tmp" "$txn_stderr_identity" 2 && _owned_temp_is_current "$txn_stderr_final" "$txn_stderr_identity" 2; then
      :
    else
      _execute_rollback_stage_transaction || return 75
      return 75
    fi
  else
    _execute_rollback_stage_transaction || return 75
    return 75
  fi
  if _owned_run_dir_is_current "$txn_run_identity" && [ ! -e "$txn_json_final" ] && [ ! -L "$txn_json_final" ] && _owned_temp_is_current "$txn_json_tmp" "$txn_json_identity" 1; then
    :
  else
    _execute_rollback_stage_transaction || return 75
    return 75
  fi
  if ln -- "$txn_json_tmp" "$txn_json_final"; then
    txn_json_created=true
    txn_json_final_identity="$txn_json_identity"
    if _owned_run_dir_is_current "$txn_run_identity" && _owned_temp_is_current "$txn_json_tmp" "$txn_json_identity" 2 && _owned_temp_is_current "$txn_json_final" "$txn_json_identity" 2; then
      :
    else
      _execute_rollback_stage_transaction || return 75
      return 75
    fi
  else
    _execute_rollback_stage_transaction || return 75
    return 75
  fi
  if _cleanup_owned_temp "$txn_results_tmp" "$txn_results_identity" 2 && _cleanup_owned_temp "$txn_stdout_tmp" "$txn_stdout_identity" 2 && _cleanup_owned_temp "$txn_stderr_tmp" "$txn_stderr_identity" 2 && _cleanup_owned_temp "$txn_json_tmp" "$txn_json_identity" 2 && _owned_temp_is_current "$txn_results_final" "$txn_results_final_identity" 1 && _owned_temp_is_current "$txn_stdout_final" "$txn_stdout_final_identity" 1 && _owned_temp_is_current "$txn_stderr_final" "$txn_stderr_final_identity" 1 && _owned_temp_is_current "$txn_json_final" "$txn_json_final_identity" 1; then
    return 0
  fi
  _execute_rollback_stage_transaction || return 75
  return 75
}

_execute_prepare_stage_transaction_for_container_stop() {
  [ "$#" -eq 0 ] || return 64
  _owned_run_dir_is_current "$stage_run_identity" || return 75
  _owned_temp_is_current "$stage_results_tmp" "$stage_results_identity" 1 || return 75
  _owned_temp_is_current "$stage_stdout_tmp" "$stage_stdout_identity" 1 || return 75
  _owned_temp_is_current "$stage_stderr_tmp" "$stage_stderr_identity" 1 || return 75
  _owned_temp_is_current "$stage_json_tmp" "$stage_json_identity" 1 || return 75
  for execute_stop_final in "$stage_results_path" "$stdout_path" "$stderr_path" "$stage_json_path"; do
    _run_dir_target_is_safe "$execute_stop_final" || return 75
    [ ! -e "$execute_stop_final" ] && [ ! -L "$execute_stop_final" ] || return 75
  done
  txn_results_tmp="$stage_results_tmp"
  txn_results_identity="$stage_results_identity"
  txn_results_final="$stage_results_path"
  txn_stdout_tmp="$stage_stdout_tmp"
  txn_stdout_identity="$stage_stdout_identity"
  txn_stdout_final="$stdout_path"
  txn_stderr_tmp="$stage_stderr_tmp"
  txn_stderr_identity="$stage_stderr_identity"
  txn_stderr_final="$stderr_path"
  txn_json_tmp="$stage_json_tmp"
  txn_json_identity="$stage_json_identity"
  txn_json_final="$stage_json_path"
  txn_run_identity="$stage_run_identity"
  txn_results_created=false
  txn_stdout_created=false
  txn_stderr_created=false
  txn_json_created=false
  txn_results_final_identity=
  txn_stdout_final_identity=
  txn_stderr_final_identity=
  txn_json_final_identity=
}

_execute_reader_snapshot_for_container_stop() {
  [ "$#" -eq 1 ] || return 64
  execute_stop_reader_pid="$1"
  case "$execute_stop_reader_pid" in
    ''|*[!0-9]*) return 75 ;;
    *) ;;
  esac
  [ -r "/proc/$execute_stop_reader_pid/stat" ] || return 1
  execute_stop_reader_snapshot="$(
    awk -v identifier="$execute_stop_reader_pid" -v owner="$$" '
      $1 == identifier && $4 == owner {
        print $1 ":" $4 ":" $22 ":" $2 ":" $3
        exit
      }
    ' "/proc/$execute_stop_reader_pid/stat"
  )" || return 75
  [ -n "$execute_stop_reader_snapshot" ] || return 75
  execute_stop_reader_snapshot_pid=${execute_stop_reader_snapshot%%:*}
  execute_stop_reader_snapshot_tail=${execute_stop_reader_snapshot#*:}
  execute_stop_reader_snapshot_parent=${execute_stop_reader_snapshot_tail%%:*}
  execute_stop_reader_snapshot_tail=${execute_stop_reader_snapshot_tail#*:}
  execute_stop_reader_snapshot_starttime=${execute_stop_reader_snapshot_tail%%:*}
  execute_stop_reader_snapshot_tail=${execute_stop_reader_snapshot_tail#*:}
  execute_stop_reader_snapshot_comm=${execute_stop_reader_snapshot_tail%%:*}
  execute_stop_reader_snapshot_state=${execute_stop_reader_snapshot_tail#*:}
  case "$execute_stop_reader_snapshot_pid" in ''|*[!0-9]*) return 75 ;; esac
  case "$execute_stop_reader_snapshot_parent" in ''|*[!0-9]*) return 75 ;; esac
  case "$execute_stop_reader_snapshot_starttime" in ''|*[!0-9]*) return 75 ;; esac
  [ "$execute_stop_reader_snapshot_pid" = "$execute_stop_reader_pid" ] || return 75
  [ "$execute_stop_reader_snapshot_parent" = "$$" ] || return 75
  case "$execute_stop_reader_snapshot_comm" in
    *:*) return 75 ;;
    \(*\)) ;;
    *) return 75 ;;
  esac
  case "$execute_stop_reader_snapshot_state" in
    Z|[A-Za-z]) ;;
    *) return 75 ;;
  esac
  printf '%s\n' "$execute_stop_reader_snapshot"
}

_execute_reader_identity_is_valid_for_container_stop() {
  [ "$#" -eq 2 ] || return 64
  execute_stop_reader_pid="$1"
  execute_stop_reader_identity="$2"
  case "$execute_stop_reader_pid" in
    ''|*[!0-9]*) return 75 ;;
    *) ;;
  esac
  case "$execute_stop_reader_identity" in
    missing:*)
      execute_stop_reader_missing_pid=${execute_stop_reader_identity#missing:}
      case "$execute_stop_reader_missing_pid" in ''|*[!0-9]*) return 75 ;; esac
      [ "$execute_stop_reader_missing_pid" = "$execute_stop_reader_pid" ] || return 75
      return 0
      ;;
    *:*:*:*) ;;
    *) return 75 ;;
  esac
  execute_stop_reader_identity_pid=${execute_stop_reader_identity%%:*}
  execute_stop_reader_identity_tail=${execute_stop_reader_identity#*:}
  execute_stop_reader_identity_parent=${execute_stop_reader_identity_tail%%:*}
  execute_stop_reader_identity_tail=${execute_stop_reader_identity_tail#*:}
  execute_stop_reader_identity_starttime=${execute_stop_reader_identity_tail%%:*}
  execute_stop_reader_identity_comm=${execute_stop_reader_identity_tail#*:}
  case "$execute_stop_reader_identity_pid" in ''|*[!0-9]*) return 75 ;; esac
  case "$execute_stop_reader_identity_parent" in ''|*[!0-9]*) return 75 ;; esac
  case "$execute_stop_reader_identity_starttime" in ''|*[!0-9]*) return 75 ;; esac
  [ "$execute_stop_reader_identity_pid" = "$execute_stop_reader_pid" ] || return 75
  [ "$execute_stop_reader_identity_parent" = "$$" ] || return 75
  case "$execute_stop_reader_identity_comm" in
    *:*) return 75 ;;
    \(*\)) ;;
    *) return 75 ;;
  esac
  return 0
}

_execute_reader_paths_are_owned_for_container_stop() {
  [ "$#" -eq 4 ] || return 64
  execute_stop_reader_fifo="$1"
  execute_stop_reader_fifo_identity="$2"
  execute_stop_reader_capture="$3"
  execute_stop_reader_capture_identity="$4"
  _execute_scratch_is_current "$stage_scratch" "$stage_scratch_identity" || return 75
  _execute_fifo_is_current "$execute_stop_reader_fifo" "$execute_stop_reader_fifo_identity" || return 75
  _owned_temp_is_current "$execute_stop_reader_capture" "$execute_stop_reader_capture_identity" 1 || return 75
  return 0
}

_execute_capture_owned_reader_identity_for_container_stop() {
  [ "$#" -eq 5 ] || return 64
  execute_stop_reader_pid="$1"
  execute_stop_reader_fifo="$2"
  execute_stop_reader_fifo_identity="$3"
  execute_stop_reader_capture="$4"
  execute_stop_reader_capture_identity="$5"
  _execute_reader_paths_are_owned_for_container_stop "$execute_stop_reader_fifo" "$execute_stop_reader_fifo_identity" "$execute_stop_reader_capture" "$execute_stop_reader_capture_identity" || return 75
  if execute_stop_reader_snapshot="$(_execute_reader_snapshot_for_container_stop "$execute_stop_reader_pid")"; then
    printf '%s\n' "${execute_stop_reader_snapshot%:*}"
    return 0
  fi
  execute_stop_reader_capture_status=$?
  [ "$execute_stop_reader_capture_status" -eq 1 ] || return 75
  printf 'missing:%s\n' "$execute_stop_reader_pid"
}

_execute_owned_reader_state_for_container_stop() {
  [ "$#" -eq 7 ] || return 64
  execute_stop_reader_started="$1"
  execute_stop_reader_pid="$2"
  execute_stop_reader_identity="$3"
  execute_stop_reader_fifo="$4"
  execute_stop_reader_fifo_identity="$5"
  execute_stop_reader_capture="$6"
  execute_stop_reader_capture_identity="$7"
  case "$execute_stop_reader_started" in
    false)
      [ -z "$execute_stop_reader_pid" ] && [ -z "$execute_stop_reader_identity" ] || return 75
      if [ -n "$execute_stop_reader_fifo" ] || [ -n "$execute_stop_reader_fifo_identity" ]; then
        [ -n "$execute_stop_reader_fifo" ] && [ -n "$execute_stop_reader_fifo_identity" ] || return 75
        _execute_scratch_is_current "$stage_scratch" "$stage_scratch_identity" || return 75
        _execute_fifo_is_current "$execute_stop_reader_fifo" "$execute_stop_reader_fifo_identity" || return 75
      fi
      if [ -n "$execute_stop_reader_capture" ] || [ -n "$execute_stop_reader_capture_identity" ]; then
        [ -n "$execute_stop_reader_capture" ] && [ -n "$execute_stop_reader_capture_identity" ] || return 75
        _execute_scratch_is_current "$stage_scratch" "$stage_scratch_identity" || return 75
        _owned_temp_is_current "$execute_stop_reader_capture" "$execute_stop_reader_capture_identity" 1 || return 75
      fi
      printf '%s\n' absent
      return 0
      ;;
    true) ;;
    *) return 75 ;;
  esac
  _execute_reader_identity_is_valid_for_container_stop "$execute_stop_reader_pid" "$execute_stop_reader_identity" || return 75
  _execute_reader_paths_are_owned_for_container_stop "$execute_stop_reader_fifo" "$execute_stop_reader_fifo_identity" "$execute_stop_reader_capture" "$execute_stop_reader_capture_identity" || return 75
  if [ ! -r "/proc/$execute_stop_reader_pid/stat" ]; then
    printf '%s\n' missing
    return 0
  fi
  execute_stop_reader_snapshot="$(_execute_reader_snapshot_for_container_stop "$execute_stop_reader_pid")" || return 75
  execute_stop_reader_current_identity=${execute_stop_reader_snapshot%:*}
  [ "$execute_stop_reader_current_identity" = "$execute_stop_reader_identity" ] || { printf '%s\n' invalid; return 0; }
  execute_stop_reader_state=${execute_stop_reader_snapshot##*:}
  case "$execute_stop_reader_state" in
    Z) printf '%s\n' zombie ;;
    [A-Za-z]) printf '%s\n' live ;;
    *) return 75 ;;
  esac
}

_execute_assert_owned_live_reader_for_container_stop() {
  [ "$#" -eq 6 ] || return 64
  execute_stop_reader_pid="$1"
  execute_stop_reader_identity="$2"
  execute_stop_reader_fifo="$3"
  execute_stop_reader_fifo_identity="$4"
  execute_stop_reader_capture="$5"
  execute_stop_reader_capture_identity="$6"
  _execute_reader_identity_is_valid_for_container_stop "$execute_stop_reader_pid" "$execute_stop_reader_identity" || return 75
  _execute_reader_paths_are_owned_for_container_stop "$execute_stop_reader_fifo" "$execute_stop_reader_fifo_identity" "$execute_stop_reader_capture" "$execute_stop_reader_capture_identity" || return 75
  execute_stop_reader_snapshot="$(_execute_reader_snapshot_for_container_stop "$execute_stop_reader_pid")" || return 75
  [ "${execute_stop_reader_snapshot%:*}" = "$execute_stop_reader_identity" ] || return 75
  [ "${execute_stop_reader_snapshot##*:}" != Z ] || return 75
  return 0
}

_execute_signal_owned_reader_for_container_stop() {
  [ "$#" -eq 7 ] || return 64
  execute_stop_reader_started="$1"
  execute_stop_reader_pid="$2"
  execute_stop_reader_identity="$3"
  execute_stop_reader_fifo="$4"
  execute_stop_reader_fifo_identity="$5"
  execute_stop_reader_capture="$6"
  execute_stop_reader_capture_identity="$7"
  execute_stop_reader_state="$(_execute_owned_reader_state_for_container_stop "$execute_stop_reader_started" "$execute_stop_reader_pid" "$execute_stop_reader_identity" "$execute_stop_reader_fifo" "$execute_stop_reader_fifo_identity" "$execute_stop_reader_capture" "$execute_stop_reader_capture_identity")" || return 75
  case "$execute_stop_reader_state" in
    absent|missing|zombie) printf '%s\n' ready ;;
    live)
      _execute_assert_owned_live_reader_for_container_stop "$execute_stop_reader_pid" "$execute_stop_reader_identity" "$execute_stop_reader_fifo" "$execute_stop_reader_fifo_identity" "$execute_stop_reader_capture" "$execute_stop_reader_capture_identity" || return 75
      kill -TERM -- "$execute_stop_reader_pid" || return 75
      printf '%s\n' signalled
      ;;
    *) return 75 ;;
  esac
}

_execute_reader_is_ready_for_container_stop() {
  [ "$#" -eq 7 ] || return 64
  execute_stop_reader_state="$(_execute_owned_reader_state_for_container_stop "$@")" || return 75
  case "$execute_stop_reader_state" in
    absent|missing|zombie) printf '%s\n' ready ;;
    live|invalid) return 75 ;;
    *) return 75 ;;
  esac
}

_execute_reap_owned_reader_for_container_stop() {
  [ "$#" -eq 7 ] || return 64
  execute_stop_reader_started="$1"
  execute_stop_reader_pid="$2"
  execute_stop_reader_state="$(_execute_owned_reader_state_for_container_stop "$@")" || return 75
  case "$execute_stop_reader_state" in
    absent) return 0 ;;
    missing|zombie) ;;
    *) return 75 ;;
  esac
  if wait "$execute_stop_reader_pid"; then :; else :; fi
  execute_stop_reader_state="$(_execute_owned_reader_state_for_container_stop "$@")" || return 75
  [ "$execute_stop_reader_state" = missing ] || return 75
  return 0
}

_execute_abort_child() {
  [ "$#" -eq 2 ] || return 64
  execute_abort_stdout_pid="$1"
  execute_abort_stderr_pid="$2"
  execute_abort_reaped=true
  if [ -n "${go_child_pid-}" ] && [ "${go_child_reaped-}" = false ]; then
    if terminate_owned_go_child_group; then :; else execute_abort_reaped=false; fi
    if reap_owned_go_child_group; then :; else execute_abort_reaped=false; fi
  fi
  if [ -n "$execute_abort_stdout_pid" ]; then wait "$execute_abort_stdout_pid" || :; fi
  if [ -n "$execute_abort_stderr_pid" ]; then wait "$execute_abort_stderr_pid" || :; fi
  [ "$execute_abort_reaped" = true ] || return 75
  [ -z "${go_child_pid-}" ] && [ "${go_child_reaped-}" = true ] || return 75
  return 0
}

_execute_cleanup_package_scratch() {
  [ "$#" -eq 0 ] || return 64
  _execute_scratch_is_current "$stage_scratch" "$stage_scratch_identity" || return 75
  for execute_cleanup_capture in "${package_stdout_capture-}:${package_stdout_identity-}" "${package_stderr_capture-}:${package_stderr_identity-}"; do
    execute_cleanup_path=${execute_cleanup_capture%%:*}
    execute_cleanup_identity=${execute_cleanup_capture#*:}
    if [ -n "$execute_cleanup_path" ] && { [ -e "$execute_cleanup_path" ] || [ -L "$execute_cleanup_path" ]; }; then
      _cleanup_owned_temp "$execute_cleanup_path" "$execute_cleanup_identity" 1 || return 75
    fi
  done
  for execute_cleanup_fifo in "${stdout_fifo-}:${stdout_fifo_identity-}" "${stderr_fifo-}:${stderr_fifo_identity-}"; do
    execute_cleanup_path=${execute_cleanup_fifo%%:*}
    execute_cleanup_identity=${execute_cleanup_fifo#*:}
    if [ -n "$execute_cleanup_path" ] && { [ -e "$execute_cleanup_path" ] || [ -L "$execute_cleanup_path" ]; }; then
      _execute_remove_fifo "$execute_cleanup_path" "$execute_cleanup_identity" || return 75
    fi
  done
  if [ -n "${stage_marker_template-}" ] && { [ -e "${budget_stop_marker-}" ] || [ -L "${budget_stop_marker-}" ]; }; then
    _owned_temp_is_current "$budget_stop_marker" "$stage_marker_identity" 2 || return 75
    _owned_temp_is_current "$stage_marker_template" "$stage_marker_identity" 2 || return 75
    rm -- "$budget_stop_marker" || return 75
  fi
  if [ -n "${stage_marker_template-}" ] && { [ -e "$stage_marker_template" ] || [ -L "$stage_marker_template" ]; }; then
    _cleanup_owned_temp "$stage_marker_template" "$stage_marker_identity" 1 || return 75
  fi
}

_execute_cleanup_unpublished_stage() {
  [ "$#" -eq 0 ] || return 64
  _owned_run_dir_is_current "$stage_run_identity" || return 75
  _cleanup_owned_temp "$stage_results_tmp" "$stage_results_identity" 1 || return 75
  _cleanup_owned_temp "$stage_stdout_tmp" "$stage_stdout_identity" 1 || return 75
  _cleanup_owned_temp "$stage_stderr_tmp" "$stage_stderr_identity" 1 || return 75
  _cleanup_owned_temp "$stage_json_tmp" "$stage_json_identity" 1 || return 75
  _execute_scratch_is_current "$stage_scratch" "$stage_scratch_identity" || return 75
  rmdir -- "$stage_scratch" || return 75
}

_execute_cleanup_prechild_temp() {
  [ "$#" -eq 2 ] || return 64
  execute_prechild_path="$1"
  execute_prechild_identity="$2"
  if [ -z "$execute_prechild_path" ] && [ -z "$execute_prechild_identity" ]; then
    return 0
  fi
  [ -n "$execute_prechild_path" ] && [ -n "$execute_prechild_identity" ] || return 75
  if [ -e "$execute_prechild_path" ] || [ -L "$execute_prechild_path" ]; then
    _cleanup_owned_temp "$execute_prechild_path" "$execute_prechild_identity" 1 || return 75
  fi
  return 0
}

_execute_abort_prechild_stage() {
  [ "$#" -eq 0 ] || return 64
  _execute_cleanup_prechild_temp "$stage_results_tmp" "$stage_results_identity" || return 75
  _execute_cleanup_prechild_temp "$stage_stdout_tmp" "$stage_stdout_identity" || return 75
  _execute_cleanup_prechild_temp "$stage_stderr_tmp" "$stage_stderr_identity" || return 75
  _execute_cleanup_prechild_temp "$stage_json_tmp" "$stage_json_identity" || return 75
  _execute_cleanup_prechild_temp "$execute_file_path" "$execute_file_identity" || return 75
  if [ -z "$stage_scratch" ] && [ -z "$stage_scratch_identity" ]; then
    return 0
  fi
  [ -n "$stage_scratch" ] && [ -n "$stage_scratch_identity" ] || return 75
  _execute_scratch_is_current "$stage_scratch" "$stage_scratch_identity" || return 75
  rmdir -- "$stage_scratch" || return 75
  return 0
}

_execute_abort_unpublished_stage() {
  [ "$#" -eq 2 ] || return 64
  execute_abort_stage_stdout_pid="$1"
  execute_abort_stage_stderr_pid="$2"
  if [ "${go_child_started-false}" = true ]; then
    _execute_abort_child "$execute_abort_stage_stdout_pid" "$execute_abort_stage_stderr_pid" || return 75
  fi
  if [ "${package_scratch_ready-false}" = true ] && [ "${package_scratch_cleaned-false}" = false ]; then
    _execute_cleanup_package_scratch || return 75
    package_scratch_cleaned=true
  fi
  _execute_cleanup_unpublished_stage || return 75
  return 0
}

_execute_reap_stage_readers_for_container_stop() {
  [ "$#" -eq 0 ] || return 64
  execute_stop_stdout_state="$(_execute_signal_owned_reader_for_container_stop "$stdout_reader_started" "$stdout_reader_pid" "$stdout_reader_identity" "$stdout_fifo" "$stdout_fifo_identity" "$package_stdout_capture" "$package_stdout_identity")" || return 75
  execute_stop_stderr_state="$(_execute_signal_owned_reader_for_container_stop "$stderr_reader_started" "$stderr_reader_pid" "$stderr_reader_identity" "$stderr_fifo" "$stderr_fifo_identity" "$package_stderr_capture" "$package_stderr_identity")" || return 75
  case "$execute_stop_stdout_state:$execute_stop_stderr_state" in
    *signalled*)
      sleep 1
      _execute_reader_is_ready_for_container_stop "$stdout_reader_started" "$stdout_reader_pid" "$stdout_reader_identity" "$stdout_fifo" "$stdout_fifo_identity" "$package_stdout_capture" "$package_stdout_identity" >/dev/null || return 75
      _execute_reader_is_ready_for_container_stop "$stderr_reader_started" "$stderr_reader_pid" "$stderr_reader_identity" "$stderr_fifo" "$stderr_fifo_identity" "$package_stderr_capture" "$package_stderr_identity" >/dev/null || return 75
      ;;
    ready:ready) ;;
    *) return 75 ;;
  esac
  case "$stdout_reader_reaped" in
    false)
      _execute_reap_owned_reader_for_container_stop "$stdout_reader_started" "$stdout_reader_pid" "$stdout_reader_identity" "$stdout_fifo" "$stdout_fifo_identity" "$package_stdout_capture" "$package_stdout_identity" || return 75
      stdout_reader_reaped=true
      ;;
    true)
      execute_stop_stdout_state="$(_execute_owned_reader_state_for_container_stop "$stdout_reader_started" "$stdout_reader_pid" "$stdout_reader_identity" "$stdout_fifo" "$stdout_fifo_identity" "$package_stdout_capture" "$package_stdout_identity")" || return 75
      case "$execute_stop_stdout_state" in absent|missing) ;; *) return 75 ;; esac
      ;;
    *) return 75 ;;
  esac
  case "$stderr_reader_reaped" in
    false)
      _execute_reap_owned_reader_for_container_stop "$stderr_reader_started" "$stderr_reader_pid" "$stderr_reader_identity" "$stderr_fifo" "$stderr_fifo_identity" "$package_stderr_capture" "$package_stderr_identity" || return 75
      stderr_reader_reaped=true
      ;;
    true)
      execute_stop_stderr_state="$(_execute_owned_reader_state_for_container_stop "$stderr_reader_started" "$stderr_reader_pid" "$stderr_reader_identity" "$stderr_fifo" "$stderr_fifo_identity" "$package_stderr_capture" "$package_stderr_identity")" || return 75
      case "$execute_stop_stderr_state" in absent|missing) ;; *) return 75 ;; esac
      ;;
    *) return 75 ;;
  esac
  return 0
}

_execute_no_active_go_child_for_container_stop() {
  [ "$#" -eq 0 ] || return 64
  [ -z "${go_child_pid-}" ] && [ -z "${go_child_pgid-}" ] && [ -z "${go_child_session-}" ] && [ -z "${go_child_starttime-}" ] && [ -z "${go_child_identity-}" ] || return 75
  case "${go_child_reaped-false}" in
    false) ;;
    true) case "${go_child_exit-}" in ''|*[!0-9]*) return 75 ;; *) ;; esac ;;
    *) return 75 ;;
  esac
  return 0
}

_execute_active_stage_artifacts_are_absent_for_container_stop() {
  [ "$#" -eq 0 ] || return 64
  _owned_run_dir_is_current "$stage_run_identity" || return 75
  for execute_stop_temp in "$stage_results_tmp" "$stage_stdout_tmp" "$stage_stderr_tmp" "$stage_json_tmp" "$stage_marker_template" "$budget_stop_marker" "$stdout_fifo" "$stderr_fifo" "$package_stdout_capture" "$package_stderr_capture"; do
    if [ -n "$execute_stop_temp" ]; then
      [ ! -e "$execute_stop_temp" ] && [ ! -L "$execute_stop_temp" ] || return 75
    fi
  done
  for execute_stop_final in "$stage_results_path" "$stdout_path" "$stderr_path" "$stage_json_path"; do
    _run_dir_target_is_safe "$execute_stop_final" || return 75
    [ ! -e "$execute_stop_final" ] && [ ! -L "$execute_stop_final" ] || return 75
  done
  if [ -n "$stage_scratch" ]; then
    [ ! -e "$stage_scratch" ] && [ ! -L "$stage_scratch" ] || return 75
  fi
  return 0
}

_execute_rollback_active_stage_for_container_stop() {
  [ "$#" -eq 0 ] || return 64
  if [ "$active_stage_artifacts_ready" = true ]; then
    _execute_rollback_stage_transaction || return 75
  else
    _execute_abort_prechild_stage || return 75
  fi
  _execute_active_stage_artifacts_are_absent_for_container_stop || return 75
  return 0
}

abort_active_stage_for_container_stop() {
  [ "$#" -eq 0 ] || return 64
  [ "$active_stage_for_container_stop" = true ] || return 75
  [ "$active_stage_abort_state" = armed ] || return 75
  active_stage_abort_state=aborting
  if [ "${go_child_started-false}" = true ]; then
    abort_owned_go_child_group_for_container_stop || return 75
  elif [ "${go_child_started-false}" = false ]; then
    _execute_no_active_go_child_for_container_stop || return 75
  else
    return 75
  fi
  child_group_reaped=true
  _execute_reap_stage_readers_for_container_stop || return 75
  if [ "${package_scratch_ready-false}" = true ] && [ "${package_scratch_cleaned-false}" = false ]; then
    _execute_cleanup_package_scratch || return 75
    package_scratch_cleaned=true
  fi
  if [ "$stage" = build ] && [ -n "${owned_output-}" ]; then
    unlink_owned_output || return 75
  fi
  _execute_rollback_active_stage_for_container_stop || return 75
  unpublished_stage_rolled_back=true
  active_stage_for_container_stop=false
  active_stage_abort_state=aborted
  return 0
}

_execute_completed_stage_final_identities_are_current() {
  [ "$#" -eq 2 ] || return 64
  execute_completed_stage="$1"
  execute_completed_tuple="$2"
  _plan_stage_is_valid "$execute_completed_stage" || return 75
  [ -n "${stage_run_identity-}" ] || return 75
  _owned_run_dir_is_current "$stage_run_identity" || return 75
  case "$execute_completed_tuple" in *'|'*) ;; *) return 75 ;; esac
  execute_completed_results=${execute_completed_tuple%%|*}
  execute_completed_tail=${execute_completed_tuple#*|}
  case "$execute_completed_tail" in *'|'*) ;; *) return 75 ;; esac
  execute_completed_stdout=${execute_completed_tail%%|*}
  execute_completed_tail=${execute_completed_tail#*|}
  case "$execute_completed_tail" in *'|'*) ;; *) return 75 ;; esac
  execute_completed_stderr=${execute_completed_tail%%|*}
  execute_completed_json=${execute_completed_tail#*|}
  case "$execute_completed_json" in *'|'*) return 75 ;; esac
  [ -n "$execute_completed_results" ] && [ -n "$execute_completed_stdout" ] && [ -n "$execute_completed_stderr" ] && [ -n "$execute_completed_json" ] || return 75
  _execute_identity_is_exact "$execute_completed_results" 10001 10001 600 'regular file' || return 75
  _execute_identity_is_exact "$execute_completed_stdout" 10001 10001 600 'regular file' || return 75
  _execute_identity_is_exact "$execute_completed_stderr" 10001 10001 600 'regular file' || return 75
  _execute_identity_is_exact "$execute_completed_json" 10001 10001 600 'regular file' || return 75
  _owned_temp_is_current "$run_dir/package-results.$execute_completed_stage.jsonl" "$execute_completed_results" 1 || return 75
  _owned_temp_is_current "$run_dir/$execute_completed_stage.stdout.log" "$execute_completed_stdout" 1 || return 75
  _owned_temp_is_current "$run_dir/$execute_completed_stage.stderr.log" "$execute_completed_stderr" 1 || return 75
  _owned_temp_is_current "$run_dir/stage-$execute_completed_stage.json" "$execute_completed_json" 1 || return 75
  return 0
}

prove_quiescent_for_container_stop() {
  [ "$#" -eq 0 ] || return 64
  [ "${active_stage_for_container_stop-}" = false ] || return 75
  [ "${active_stage_abort_state-}" = idle ] || return 75
  [ "${active_stage_publish_in_progress-}" = false ] || return 75
  [ "${active_stage_artifacts_ready-}" = false ] || return 75
  [ "${go_child_started-false}" = false ] || return 75
  [ "${stdout_reader_started-false}" = false ] || return 75
  [ "${stderr_reader_started-false}" = false ] || return 75
  [ -z "${stdout_reader_pid-}" ] && [ -z "${stderr_reader_pid-}" ] || return 75
  [ -z "${stdout_reader_identity-}" ] && [ -z "${stderr_reader_identity-}" ] || return 75
  [ -z "${owned_output-}" ] || return 75
  _execute_no_active_go_child_for_container_stop || return 75
  case "${package_scratch_ready-false}:${package_scratch_cleaned-false}" in
    false:false|true:true) ;;
    *) return 75 ;;
  esac

  case "${child_group_reaped-false}:${stdout_reader_reaped-false}:${stderr_reader_reaped-false}:${unpublished_stage_rolled_back-false}" in
    false:false:false:false)
      [ "${go_child_reaped-false}" = false ] || return 75
      [ -z "${go_child_exit-}" ] || return 75
      [ "${package_scratch_ready-false}:${package_scratch_cleaned-false}" = false:false ] || return 75
      for execute_quiescent_value in \
        "${stage_run_identity-}" "${stage_scratch-}" "${stage_scratch_identity-}" \
        "${stage_results_tmp-}" "${stage_results_identity-}" \
        "${stage_stdout_tmp-}" "${stage_stdout_identity-}" \
        "${stage_stderr_tmp-}" "${stage_stderr_identity-}" \
        "${stage_json_tmp-}" "${stage_json_identity-}" \
        "${execute_file_path-}" "${execute_file_identity-}" \
        "${stage_marker_template-}" "${stage_marker_identity-}" "${budget_stop_marker-}" \
        "${stdout_fifo-}" "${stdout_fifo_identity-}" "${stderr_fifo-}" "${stderr_fifo_identity-}" \
        "${package_stdout_capture-}" "${package_stdout_identity-}" \
        "${package_stderr_capture-}" "${package_stderr_identity-}" \
        "${completed_build_final_identities-}" "${completed_vet_final_identities-}" \
        "${completed_unit_final_identities-}" "${completed_race_final_identities-}"; do
        [ -z "$execute_quiescent_value" ] || return 75
      done
      ;;
    true:true:true:false)
      [ "${go_child_reaped-false}" = true ] || return 75
      case "${go_child_exit-}" in ''|*[!0-9]*) return 75 ;; esac
      [ "${package_scratch_ready-false}:${package_scratch_cleaned-false}" = true:true ] || return 75
      [ "${GOTMPDIR-}" = /var/tmp/hub-go ] || return 75
      _plan_stage_is_valid "${stage-}" || return 75
      [ -n "${stage_run_identity-}" ] || return 75
      _owned_run_dir_is_current "$stage_run_identity" || return 75
      [ "${stage_plan_path-}" = "$run_dir/package-plan.$stage.json" ] || return 75
      [ "${stage_results_path-}" = "$run_dir/package-results.$stage.jsonl" ] || return 75
      [ "${stdout_path-}" = "$run_dir/$stage.stdout.log" ] || return 75
      [ "${stderr_path-}" = "$run_dir/$stage.stderr.log" ] || return 75
      [ "${stage_json_path-}" = "$run_dir/stage-$stage.json" ] || return 75
      _owned_regular_identity "$stage_plan_path" >/dev/null || return 75
      for execute_quiescent_temp in "$stage_results_tmp" "$stage_stdout_tmp" "$stage_stderr_tmp" "$stage_json_tmp"; do
        _owned_temp_name_is_safe "$run_dir" publish "$execute_quiescent_temp" || return 75
        [ ! -e "$execute_quiescent_temp" ] && [ ! -L "$execute_quiescent_temp" ] || return 75
      done
      case "$stage" in
        build)
          [ -n "$completed_build_final_identities" ] && [ -z "$completed_vet_final_identities" ] && [ -z "$completed_unit_final_identities" ] && [ -z "$completed_race_final_identities" ] || return 75
          _execute_completed_stage_final_identities_are_current build "$completed_build_final_identities" || return 75
          ;;
        vet)
          [ -n "$completed_build_final_identities" ] && [ -n "$completed_vet_final_identities" ] && [ -z "$completed_unit_final_identities" ] && [ -z "$completed_race_final_identities" ] || return 75
          _execute_completed_stage_final_identities_are_current build "$completed_build_final_identities" || return 75
          _execute_completed_stage_final_identities_are_current vet "$completed_vet_final_identities" || return 75
          ;;
        unit)
          [ -n "$completed_build_final_identities" ] && [ -n "$completed_vet_final_identities" ] && [ -n "$completed_unit_final_identities" ] && [ -z "$completed_race_final_identities" ] || return 75
          _execute_completed_stage_final_identities_are_current build "$completed_build_final_identities" || return 75
          _execute_completed_stage_final_identities_are_current vet "$completed_vet_final_identities" || return 75
          _execute_completed_stage_final_identities_are_current unit "$completed_unit_final_identities" || return 75
          ;;
        race)
          [ -n "$completed_build_final_identities" ] && [ -n "$completed_vet_final_identities" ] && [ -n "$completed_unit_final_identities" ] && [ -n "$completed_race_final_identities" ] || return 75
          _execute_completed_stage_final_identities_are_current build "$completed_build_final_identities" || return 75
          _execute_completed_stage_final_identities_are_current vet "$completed_vet_final_identities" || return 75
          _execute_completed_stage_final_identities_are_current unit "$completed_unit_final_identities" || return 75
          _execute_completed_stage_final_identities_are_current race "$completed_race_final_identities" || return 75
          ;;
        *) return 75 ;;
      esac
      _execute_identity_is_exact "$stage_results_identity" 10001 10001 600 'regular file' || return 75
      _execute_identity_is_exact "$stage_stdout_identity" 10001 10001 600 'regular file' || return 75
      _execute_identity_is_exact "$stage_stderr_identity" 10001 10001 600 'regular file' || return 75
      _execute_identity_is_exact "$stage_json_identity" 10001 10001 600 'regular file' || return 75
      _owned_temp_is_current "$stage_results_path" "$stage_results_identity" 1 || return 75
      _owned_temp_is_current "$stdout_path" "$stage_stdout_identity" 1 || return 75
      _owned_temp_is_current "$stderr_path" "$stage_stderr_identity" 1 || return 75
      _owned_temp_is_current "$stage_json_path" "$stage_json_identity" 1 || return 75
      case "${stage_scratch-}" in
        "$GOTMPDIR/.backend-unit-stage-$stage."[A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9]) ;;
        *) return 75 ;;
      esac
      _execute_identity_is_exact "${stage_scratch_identity-}" 10001 10001 700 directory || return 75
      [ ! -e "$stage_scratch" ] && [ ! -L "$stage_scratch" ] || return 75
      case "${package_ordinal-}" in ''|*[!0-9]*) return 75 ;; esac
      [ "$package_ordinal" -ge 1 ] || return 75
      case "${stage_marker_template-}" in
        "$stage_scratch/marker."[A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9]) ;;
        *) return 75 ;;
      esac
      _execute_identity_is_exact "${stage_marker_identity-}" 10001 10001 600 'regular file' || return 75
      [ "${budget_stop_marker-}" = "$stage_scratch/stop.$package_ordinal" ] || return 75
      [ "${stdout_fifo-}" = "$stage_scratch/stdout.$package_ordinal" ] || return 75
      [ "${stderr_fifo-}" = "$stage_scratch/stderr.$package_ordinal" ] || return 75
      _execute_identity_is_exact "${stdout_fifo_identity-}" 10001 10001 600 fifo || return 75
      _execute_identity_is_exact "${stderr_fifo_identity-}" 10001 10001 600 fifo || return 75
      case "${package_stdout_capture-}" in "$stage_scratch/stdout-capture."[A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9]) ;; *) return 75 ;; esac
      case "${package_stderr_capture-}" in "$stage_scratch/stderr-capture."[A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9]) ;; *) return 75 ;; esac
      _execute_identity_is_exact "${package_stdout_identity-}" 10001 10001 600 'regular file' || return 75
      _execute_identity_is_exact "${package_stderr_identity-}" 10001 10001 600 'regular file' || return 75
      for execute_quiescent_path in \
        "$stage_marker_template" "$budget_stop_marker" "$stdout_fifo" "$stderr_fifo" \
        "$package_stdout_capture" "$package_stderr_capture"; do
        [ ! -e "$execute_quiescent_path" ] && [ ! -L "$execute_quiescent_path" ] || return 75
      done
      [ -z "${execute_file_path-}" ] && [ -z "${execute_file_identity-}" ] || return 75
      ;;
    *) return 75 ;;
  esac

  child_group_reaped=true
  stdout_reader_reaped=true
  stderr_reader_reaped=true
  unpublished_stage_rolled_back=true
  return 0
}

run_stage() {
  [ "$#" -eq 2 ] || return 64
  stage="$1"
  rows_path="$2"
  _plan_stage_is_valid "$stage" || return 64
  LC_ALL=C
  export LC_ALL
  stage_run_identity="$(_owned_run_dir_identity)" || return 75
  stage_expected_rows_path="$(package_plan_rows_path "$stage")" || return 75
  [ "$rows_path" = "$stage_expected_rows_path" ] || return 75
  [ -f "$rows_path" ] && [ ! -L "$rows_path" ] || return 75
  _owned_regular_identity "$rows_path" >/dev/null || return 75
  _plan_validate_rows "$rows_path" || return 75
  stage_plan_path="$run_dir/package-plan.$stage.json"
  stage_results_path="$run_dir/package-results.$stage.jsonl"
  stdout_path="$run_dir/$stage.stdout.log"
  stderr_path="$run_dir/$stage.stderr.log"
  stage_json_path="$run_dir/stage-$stage.json"
  for stage_path in "$stage_plan_path"; do _run_dir_target_is_safe "$stage_path" || return 75; _owned_regular_identity "$stage_path" >/dev/null || return 75; done
  _execute_package_plan_is_bound "$stage" "$rows_path" || return 75
  for stage_path in "$stage_results_path" "$stdout_path" "$stderr_path" "$stage_json_path"; do _run_dir_target_is_safe "$stage_path" || return 75; [ ! -e "$stage_path" ] && [ ! -L "$stage_path" ] || return 75; done
  stage_scratch=
  stage_scratch_identity=
  stage_results_tmp=
  stage_results_identity=
  stage_stdout_tmp=
  stage_stdout_identity=
  stage_stderr_tmp=
  stage_stderr_identity=
  stage_json_tmp=
  stage_json_identity=
  execute_file_path=
  execute_file_identity=
  go_child_started=false
  package_scratch_ready=false
  package_scratch_cleaned=false
  stage_marker_template=
  stage_marker_identity=
  budget_stop_marker=
  stdout_fifo=
  stdout_fifo_identity=
  stderr_fifo=
  stderr_fifo_identity=
  package_stdout_capture=
  package_stdout_identity=
  package_stderr_capture=
  package_stderr_identity=
  stdout_reader_started=false
  stderr_reader_started=false
  stdout_reader_pid=
  stderr_reader_pid=
  stdout_reader_identity=
  stderr_reader_identity=
  active_stage_for_container_stop=true
  active_stage_abort_state=armed
  active_stage_artifacts_ready=false
  active_stage_publish_in_progress=false
  child_group_reaped=false
  stdout_reader_reaped=false
  stderr_reader_reaped=false
  unpublished_stage_rolled_back=false
  if _execute_create_scratch "$stage"; then :; else _execute_abort_prechild_stage || return 75; return 75; fi
  if _execute_create_run_file; then
    stage_results_tmp="$execute_file_path"; stage_results_identity="$execute_file_identity"
    execute_file_path=
    execute_file_identity=
  else
    _execute_abort_prechild_stage || return 75
    return 75
  fi
  if _execute_create_run_file; then
    stage_stdout_tmp="$execute_file_path"; stage_stdout_identity="$execute_file_identity"
    execute_file_path=
    execute_file_identity=
  else
    _execute_abort_prechild_stage || return 75
    return 75
  fi
  if _execute_create_run_file; then
    stage_stderr_tmp="$execute_file_path"; stage_stderr_identity="$execute_file_identity"
    execute_file_path=
    execute_file_identity=
  else
    _execute_abort_prechild_stage || return 75
    return 75
  fi
  if _execute_create_run_file; then
    stage_json_tmp="$execute_file_path"; stage_json_identity="$execute_file_identity"
    execute_file_path=
    execute_file_identity=
  else
    _execute_abort_prechild_stage || return 75
    return 75
  fi
  if _execute_prepare_stage_transaction_for_container_stop; then
    active_stage_artifacts_ready=true
  else
    _execute_abort_prechild_stage || return 75
    return 75
  fi
  stage_exit=0
  package_ordinal=0
  stage_previous_import_path=
  while IFS= read -r execute_plan_row; do
    go_child_started=false
    package_scratch_ready=false
    package_scratch_cleaned=false
    stage_marker_template=
    stage_marker_identity=
    budget_stop_marker=
    stdout_fifo=
    stdout_fifo_identity=
    stderr_fifo=
    stderr_fifo_identity=
    package_stdout_capture=
    package_stdout_identity=
    package_stderr_capture=
    package_stderr_identity=
    stdout_reader_started=false
    stderr_reader_started=false
    stdout_reader_pid=
    stderr_reader_pid=
    stdout_reader_identity=
    stderr_reader_identity=
    child_group_reaped=false
    stdout_reader_reaped=false
    stderr_reader_reaped=false
    if _execute_parse_plan_row "$execute_plan_row"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    if [ -n "$stage_previous_import_path" ] && [ "$package_import_path" \> "$stage_previous_import_path" ]; then
      :
    elif [ -n "$stage_previous_import_path" ]; then
      _execute_abort_unpublished_stage "" "" || return 75
      return 75
    fi
    stage_previous_import_path="$package_import_path"
    package_ordinal=$((package_ordinal + 1))
    if check_budget; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    if assert_cgroup; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    if cgroup_baseline="$(capture_cgroup_snapshot)"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    if [ "$package_ordinal" -eq 1 ]; then
      stage_cgroup_baseline="$cgroup_baseline"
    elif [ "$cgroup_baseline" = "$stage_cgroup_baseline" ]; then
      :
    else
      _execute_abort_unpublished_stage "" "" || return 75
      return 75
    fi
    if [ "$stage" = build ] && ! prepare_owned_output "$package_ordinal" "$package_import_path"; then
      _execute_abort_unpublished_stage "" "" || return 75
      return 75
    fi
    package_scratch_ready=true
    if _execute_create_scratch_file marker; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    stage_marker_template="$execute_scratch_file"; stage_marker_identity="$execute_scratch_file_identity"
    budget_stop_marker="$stage_scratch/stop.$package_ordinal"
    if [ ! -e "$budget_stop_marker" ] && [ ! -L "$budget_stop_marker" ]; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    stdout_fifo="$stage_scratch/stdout.$package_ordinal"
    if _execute_create_fifo "$stdout_fifo"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    stdout_fifo_identity="$execute_fifo_identity"
    stderr_fifo="$stage_scratch/stderr.$package_ordinal"
    if _execute_create_fifo "$stderr_fifo"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    stderr_fifo_identity="$execute_fifo_identity"
    if _execute_create_scratch_file stdout-capture; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    package_stdout_capture="$execute_scratch_file"; package_stdout_identity="$execute_scratch_file_identity"
    if _execute_create_scratch_file stderr-capture; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    package_stderr_capture="$execute_scratch_file"; package_stderr_identity="$execute_scratch_file_identity"
    package_scratch_ready=true
    if stage_stdout_remaining="$(_execute_remaining_stage_stream_capacity "$stage" stdout "$stage_stdout_tmp" "$stage_stdout_identity")" && stage_stderr_remaining="$(_execute_remaining_stage_stream_capacity "$stage" stderr "$stage_stderr_tmp" "$stage_stderr_identity")"; then
      :
    else
      report_budget_truncated_due_to_budget=true
      if _execute_mark_budget_stop "$stage_marker_template" "$stage_marker_identity" "$budget_stop_marker" "$stage_scratch_identity"; then :; else :; fi
      _execute_abort_unpublished_stage "" "" || return 75
      return 75
    fi
    if start_owned_go_child_group "$stage" "$package_import_path"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    go_child_started=true
    read_bounded_stage_stream "$stage" stdout "$stage_stdout_remaining" "$stdout_fifo" "$stdout_fifo_identity" "$package_stdout_capture" "$package_stdout_identity" "$budget_stop_marker" "$stage_marker_identity" &
    stdout_reader_pid="$!"
    if stdout_reader_identity="$(_execute_capture_owned_reader_identity_for_container_stop "$stdout_reader_pid" "$stdout_fifo" "$stdout_fifo_identity" "$package_stdout_capture" "$package_stdout_identity")"; then
      stdout_reader_started=true
    else
      _execute_abort_unpublished_stage "$stdout_reader_pid" "" || return 75
      return 75
    fi
    read_bounded_stage_stream "$stage" stderr "$stage_stderr_remaining" "$stderr_fifo" "$stderr_fifo_identity" "$package_stderr_capture" "$package_stderr_identity" "$budget_stop_marker" "$stage_marker_identity" &
    stderr_reader_pid="$!"
    if stderr_reader_identity="$(_execute_capture_owned_reader_identity_for_container_stop "$stderr_reader_pid" "$stderr_fifo" "$stderr_fifo_identity" "$package_stderr_capture" "$package_stderr_identity")"; then
      stderr_reader_started=true
    else
      _execute_abort_unpublished_stage "$stdout_reader_pid" "$stderr_reader_pid" || return 75
      return 75
    fi
    if monitor_owned_budget_stop "$go_child_pid" "$budget_stop_marker"; then
      :
    else
      report_budget_truncated_due_to_budget=true
      _execute_abort_unpublished_stage "$stdout_reader_pid" "$stderr_reader_pid" || return 75
      return 75
    fi
    stdout_reader_exit=0; stderr_reader_exit=0
    wait "$stdout_reader_pid" || stdout_reader_exit=$?
    stdout_reader_reaped=true
    wait "$stderr_reader_pid" || stderr_reader_exit=$?
    stderr_reader_reaped=true
    if [ -e "$budget_stop_marker" ] || [ "$stdout_reader_exit" -ne 0 ] || [ "$stderr_reader_exit" -ne 0 ]; then
      report_budget_truncated_due_to_budget=true
      _execute_abort_unpublished_stage "" "" || return 75
      return 75
    fi
    if reap_owned_go_child_group; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    child_group_reaped=true
    go_child_started=false
    child_exit="$go_child_exit"
    case "$child_exit" in ''|*[!0-9]*) _execute_abort_unpublished_stage "" "" || return 75; return 75 ;; esac
    if cgroup_final="$(capture_cgroup_snapshot)"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    if verify_cgroup_snapshot_unchanged "$cgroup_baseline"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    if [ "$package_ordinal" -eq 1 ]; then
      stage_cgroup_final="$cgroup_final"
    elif [ "$cgroup_final" = "$stage_cgroup_final" ]; then
      :
    else
      _execute_abort_unpublished_stage "" "" || return 75
      return 75
    fi
    if _execute_stream_capture_fits_stage_cap "$stage" stdout "$stage_stdout_tmp" "$stage_stdout_identity" "$package_stdout_capture" "$package_stdout_identity" && _execute_stream_capture_fits_stage_cap "$stage" stderr "$stage_stderr_tmp" "$stage_stderr_identity" "$package_stderr_capture" "$package_stderr_identity"; then
      :
    else
      report_budget_truncated_due_to_budget=true
      if _execute_mark_budget_stop "$stage_marker_template" "$stage_marker_identity" "$budget_stop_marker" "$stage_scratch_identity"; then :; else :; fi
      _execute_abort_unpublished_stage "" "" || return 75
      return 75
    fi
    if _execute_append_capture "$package_stdout_capture" "$package_stdout_identity" "$stage_stdout_tmp" "$stage_stdout_identity"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    if _execute_append_capture "$package_stderr_capture" "$package_stderr_identity" "$stage_stderr_tmp" "$stage_stderr_identity"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    case "$child_exit" in 0) package_status=PASS ;; *) package_status=FAIL ;; esac
    if _execute_append_owned_file "$stage_results_tmp" "$stage_results_identity" "{\"ordinal\":$package_ordinal,\"import_path\":\"$package_import_path\",\"package_name\":\"$package_name\",\"result\":\"$package_status\",\"exit_code\":$child_exit}"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    if [ "$stage" = build ] && ! unlink_owned_output; then _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
    if _execute_cleanup_package_scratch; then
      package_scratch_cleaned=true
      stdout_reader_started=false
      stderr_reader_started=false
      stdout_reader_pid=
      stderr_reader_pid=
      stdout_reader_identity=
      stderr_reader_identity=
    else
      _execute_abort_unpublished_stage "" "" || return 75
      return 75
    fi
    if [ "$child_exit" -ne 0 ] && [ "$stage_exit" -eq 0 ]; then stage_exit="$child_exit"; fi
  done < "$rows_path"
  if [ "$package_ordinal" -gt 0 ]; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if stage_results_sha256="$(_execute_regular_sha256 "$stage_results_tmp")"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if stage_stdout_sha256="$(_execute_regular_sha256 "$stage_stdout_tmp")"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if stage_stderr_sha256="$(_execute_regular_sha256 "$stage_stderr_tmp")"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if stage_plan_sha256="$(_execute_regular_sha256 "$stage_plan_path")"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if stage_stdout_bytes="$(_execute_regular_bytes "$stage_stdout_tmp")"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if stage_stderr_bytes="$(_execute_regular_bytes "$stage_stderr_tmp")"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if stage_result_rows="$(wc -l < "$stage_results_tmp")"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if [ "$stage_result_rows" = "$package_ordinal" ]; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if [ "$stage_exit" -eq 0 ]; then stage_status=PASS; else stage_status=FAIL; fi
  if _owned_temp_is_current "$stage_json_tmp" "$stage_json_identity" 1; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if printf '{"schema":"hub.qa.backend-unit-stage-result/v1","target":"backend-unit","run_id":"%s","owner_nonce":"%s","stage":"%s","package_plan_sha256":"%s","package_result_sha256":"%s","package_count":%s,"package_result_count":%s,"first_exit":%s,"status":"%s","stdout_sha256":"%s","stdout_size":%s,"stderr_sha256":"%s","stderr_size":%s,"cgroup_baseline":"%s","cgroup_final":"%s","report_budget_truncated_due_to_budget":false}\n' "$run_id" "$owner_nonce" "$stage" "$stage_plan_sha256" "$stage_results_sha256" "$package_ordinal" "$stage_result_rows" "$stage_exit" "$stage_status" "$stage_stdout_sha256" "$stage_stdout_bytes" "$stage_stderr_sha256" "$stage_stderr_bytes" "$stage_cgroup_baseline" "$stage_cgroup_final" > "$stage_json_tmp"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if _owned_temp_is_current "$stage_json_tmp" "$stage_json_identity" 1; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if stage_regular_before="$(report_regular_file_bytes)"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if stage_metadata_before="$(report_metadata_bytes)"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if [ "$stage_regular_before" -le "$REPORT_BUDGET_BYTES" ]; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if [ "$stage_metadata_before" -le "$REPORT_METADATA_BUDGET_BYTES" ]; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if check_budget; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  if rmdir -- "$stage_scratch"; then :; else _execute_abort_unpublished_stage "" "" || return 75; return 75; fi
  active_stage_publish_in_progress=true
  _execute_publish_stage_transaction "$stage_results_tmp" "$stage_results_identity" "$stage_results_path" "$stage_stdout_tmp" "$stage_stdout_identity" "$stdout_path" "$stage_stderr_tmp" "$stage_stderr_identity" "$stderr_path" "$stage_json_tmp" "$stage_json_identity" "$stage_json_path" "$stage_run_identity" || return 75
  _execute_identity_is_exact "$stage_results_identity" 10001 10001 600 'regular file' || return 75
  _execute_identity_is_exact "$stage_stdout_identity" 10001 10001 600 'regular file' || return 75
  _execute_identity_is_exact "$stage_stderr_identity" 10001 10001 600 'regular file' || return 75
  _execute_identity_is_exact "$stage_json_identity" 10001 10001 600 'regular file' || return 75
  case "$stage" in
    build)
      [ -z "$completed_build_final_identities" ] && [ -z "$completed_vet_final_identities" ] && [ -z "$completed_unit_final_identities" ] && [ -z "$completed_race_final_identities" ] || return 75
      completed_build_final_identities="$stage_results_identity|$stage_stdout_identity|$stage_stderr_identity|$stage_json_identity"
      ;;
    vet)
      [ -n "$completed_build_final_identities" ] && [ -z "$completed_vet_final_identities" ] && [ -z "$completed_unit_final_identities" ] && [ -z "$completed_race_final_identities" ] || return 75
      completed_vet_final_identities="$stage_results_identity|$stage_stdout_identity|$stage_stderr_identity|$stage_json_identity"
      ;;
    unit)
      [ -n "$completed_build_final_identities" ] && [ -n "$completed_vet_final_identities" ] && [ -z "$completed_unit_final_identities" ] && [ -z "$completed_race_final_identities" ] || return 75
      completed_unit_final_identities="$stage_results_identity|$stage_stdout_identity|$stage_stderr_identity|$stage_json_identity"
      ;;
    race)
      [ -n "$completed_build_final_identities" ] && [ -n "$completed_vet_final_identities" ] && [ -n "$completed_unit_final_identities" ] && [ -z "$completed_race_final_identities" ] || return 75
      completed_race_final_identities="$stage_results_identity|$stage_stdout_identity|$stage_stderr_identity|$stage_json_identity"
      ;;
    *) return 75 ;;
  esac
  active_stage_publish_in_progress=false
  active_stage_artifacts_ready=false
  active_stage_for_container_stop=false
  active_stage_abort_state=idle
  return "$stage_exit"
}
