#!/bin/sh
set -eu

CGROUP_MEMORY_BYTES=2147483648
CGROUP_CPU_QUOTA_US=600000
CGROUP_CPU_PERIOD_US=100000
TMPFS_TMP_BYTES=67108864
TMPFS_GOCACHE_BYTES=805306368
TMPFS_GOTMPDIR_BYTES=402653184
GOCACHE_WATERMARK_BYTES=536870912
REPORT_BUDGET_BYTES=67108864
REPORT_METADATA_BUDGET_BYTES=8388608
BUILD_STDOUT_CAP_BYTES=2097152
BUILD_STDERR_CAP_BYTES=2097152
VET_STDOUT_CAP_BYTES=2097152
VET_STDERR_CAP_BYTES=2097152
UNIT_STDOUT_CAP_BYTES=16777216
UNIT_STDERR_CAP_BYTES=4194304
RACE_STDOUT_CAP_BYTES=25165824
RACE_STDERR_CAP_BYTES=4194304

report_budget_truncated_due_to_budget=false
go_child_pid=
go_child_session=
go_child_starttime=
go_child_pgid=
go_child_identity=
go_child_reaped=false

_cgroup_file_equals() {
  [ "$#" -eq 2 ] || return 64
  cgroup_file="$1"
  cgroup_expected="$2"
  [ -r "$cgroup_file" ] || return 75
  cgroup_actual="$(cat -- "$cgroup_file")" || return 75
  [ "$cgroup_actual" = "$cgroup_expected" ] || return 75
  return 0
}

_cgroup_v2_events_are_readable() {
  [ "$#" -eq 1 ] || return 64
  cgroup_events_file="$1"
  [ -r "$cgroup_events_file" ] || return 75
  grep -Eq '^oom [0-9]+$' "$cgroup_events_file" || return 75
  grep -Eq '^oom_kill [0-9]+$' "$cgroup_events_file" || return 75
  return 0
}

assert_cgroup() {
  [ "$#" -eq 0 ] || return 64

  if [ -r /sys/fs/cgroup/cgroup.controllers ]; then
    cgroup_family=v2
    cgroup_root=/sys/fs/cgroup
    cgroup_memory_limit_file="$cgroup_root/memory.max"
    cgroup_swap_limit_file="$cgroup_root/memory.swap.max"
    cgroup_cpu_limit_file="$cgroup_root/cpu.max"
    cgroup_failure_file="$cgroup_root/memory.events"
    _cgroup_file_equals "$cgroup_memory_limit_file" "$CGROUP_MEMORY_BYTES" || return 75
    _cgroup_file_equals "$cgroup_swap_limit_file" 0 || return 75
    _cgroup_file_equals "$cgroup_cpu_limit_file" "$CGROUP_CPU_QUOTA_US $CGROUP_CPU_PERIOD_US" || return 75
    _cgroup_v2_events_are_readable "$cgroup_failure_file" || return 75
    cgroup_memory_limit_value="$(cat -- "$cgroup_memory_limit_file")" || return 75
    cgroup_swap_limit_value="$(cat -- "$cgroup_swap_limit_file")" || return 75
    cgroup_cpu_limit_value="$(cat -- "$cgroup_cpu_limit_file")" || return 75
    cgroup_failure_value="$(cat -- "$cgroup_failure_file")" || return 75
    cgroup_oom_control_file=
    return 0
  fi

  if [ -r /sys/fs/cgroup/memory/memory.limit_in_bytes ] && [ -r /sys/fs/cgroup/memory/memory.memsw.limit_in_bytes ]; then
    cgroup_family=v1
    cgroup_root=/sys/fs/cgroup/memory
    cgroup_memory_limit_file="$cgroup_root/memory.limit_in_bytes"
    cgroup_swap_limit_file="$cgroup_root/memory.memsw.limit_in_bytes"
    cgroup_failure_file="$cgroup_root/memory.failcnt"
    _cgroup_file_equals "$cgroup_memory_limit_file" "$CGROUP_MEMORY_BYTES" || return 75
    _cgroup_file_equals "$cgroup_swap_limit_file" "$CGROUP_MEMORY_BYTES" || return 75
    [ -r "$cgroup_failure_file" ] || return 75
    grep -Eq '^[0-9]+$' "$cgroup_failure_file" || return 75

    if [ -r /sys/fs/cgroup/cpu/cpu.cfs_quota_us ] && [ -r /sys/fs/cgroup/cpu/cpu.cfs_period_us ]; then
      cgroup_cpu_root=/sys/fs/cgroup/cpu
    elif [ -r /sys/fs/cgroup/cpu,cpuacct/cpu.cfs_quota_us ] && [ -r /sys/fs/cgroup/cpu,cpuacct/cpu.cfs_period_us ]; then
      cgroup_cpu_root=/sys/fs/cgroup/cpu,cpuacct
    else
      return 75
    fi
    cgroup_cpu_limit_file="$cgroup_cpu_root/cpu.cfs_quota_us"
    cgroup_cpu_period_file="$cgroup_cpu_root/cpu.cfs_period_us"
    _cgroup_file_equals "$cgroup_cpu_limit_file" "$CGROUP_CPU_QUOTA_US" || return 75
    _cgroup_file_equals "$cgroup_cpu_period_file" "$CGROUP_CPU_PERIOD_US" || return 75
    cgroup_memory_limit_value="$(cat -- "$cgroup_memory_limit_file")" || return 75
    cgroup_swap_limit_value="$(cat -- "$cgroup_swap_limit_file")" || return 75
    cgroup_cpu_limit_value="$(cat -- "$cgroup_cpu_limit_file")" || return 75
    cgroup_cpu_period_value="$(cat -- "$cgroup_cpu_period_file")" || return 75
    cgroup_failure_value="$(cat -- "$cgroup_failure_file")" || return 75
    if [ -r "$cgroup_root/memory.oom_control" ]; then
      cgroup_oom_control_file="$cgroup_root/memory.oom_control"
    else
      cgroup_oom_control_file=
    fi
    return 0
  fi

  return 75
}

_cgroup_decimal_file_value() {
  [ "$#" -eq 1 ] || return 64
  cgroup_decimal_file="$1"
  [ -r "$cgroup_decimal_file" ] || return 75
  cgroup_decimal_value="$(cat -- "$cgroup_decimal_file")" || return 75
  case "$cgroup_decimal_value" in
    ''|*[!0-9]*) return 75 ;;
    *) printf '%s\n' "$cgroup_decimal_value" ;;
  esac
}

_cgroup_v2_oom_snapshot_fields() {
  [ "$#" -eq 1 ] || return 64
  cgroup_events_snapshot_file="$1"
  [ -r "$cgroup_events_snapshot_file" ] || return 75
  awk '
    $1 == "oom" {
      if (NF != 2 || $2 !~ /^[0-9]+$/ || seen_oom) invalid = 1
      seen_oom = 1
      oom = $2
    }
    $1 == "oom_kill" {
      if (NF != 2 || $2 !~ /^[0-9]+$/ || seen_oom_kill) invalid = 1
      seen_oom_kill = 1
      oom_kill = $2
    }
    END {
      if (invalid || !seen_oom || !seen_oom_kill) exit 75
      printf "memory.events.oom=%s,memory.events.oom_kill=%s\n", oom, oom_kill
    }
  ' "$cgroup_events_snapshot_file" || return 75
}

_cgroup_v1_oom_control_snapshot_fields() {
  [ "$#" -eq 1 ] || return 64
  cgroup_oom_control_snapshot_file="$1"
  [ -r "$cgroup_oom_control_snapshot_file" ] || return 75
  awk '
    $1 == "oom_kill_disable" {
      if (NF != 2 || $2 !~ /^[0-9]+$/ || seen_oom_kill_disable) invalid = 1
      seen_oom_kill_disable = 1
      oom_kill_disable = $2
    }
    $1 == "under_oom" {
      if (NF != 2 || $2 !~ /^[0-9]+$/ || seen_under_oom) invalid = 1
      seen_under_oom = 1
      under_oom = $2
    }
    END {
      if (invalid || !seen_oom_kill_disable || !seen_under_oom) exit 75
      printf "memory.oom_control.oom_kill_disable=%s,memory.oom_control.under_oom=%s\n", oom_kill_disable, under_oom
    }
  ' "$cgroup_oom_control_snapshot_file" || return 75
}

_cgroup_snapshot_take_field() {
  [ "$#" -eq 0 ] || return 64
  case "$cgroup_snapshot_remaining" in
    *,*)
      cgroup_snapshot_field=${cgroup_snapshot_remaining%%,*}
      cgroup_snapshot_remaining=${cgroup_snapshot_remaining#*,}
      ;;
    *)
      cgroup_snapshot_field=$cgroup_snapshot_remaining
      cgroup_snapshot_remaining=
      ;;
  esac
  [ -n "$cgroup_snapshot_field" ] || return 75
}

_cgroup_snapshot_require_literal() {
  [ "$#" -eq 1 ] || return 64
  _cgroup_snapshot_take_field || return 75
  [ "$cgroup_snapshot_field" = "$1" ] || return 75
}

_cgroup_snapshot_require_decimal() {
  [ "$#" -eq 1 ] || return 64
  cgroup_snapshot_decimal_key="$1"
  _cgroup_snapshot_take_field || return 75
  case "$cgroup_snapshot_field" in
    "$cgroup_snapshot_decimal_key="*) cgroup_snapshot_decimal_value=${cgroup_snapshot_field#"$cgroup_snapshot_decimal_key="} ;;
    *) return 75 ;;
  esac
  case "$cgroup_snapshot_decimal_value" in
    ''|*[!0-9]*) return 75 ;;
    *) return 0 ;;
  esac
}

_cgroup_snapshot_is_canonical() {
  [ "$#" -eq 1 ] || return 64
  cgroup_snapshot_remaining="$1"
  _cgroup_snapshot_require_literal 'schema=hub.qa.cgroup-snapshot/v1' || return 75
  _cgroup_snapshot_take_field || return 75
  case "$cgroup_snapshot_field" in
    family=v2)
      _cgroup_snapshot_require_literal "memory.max=$CGROUP_MEMORY_BYTES" || return 75
      _cgroup_snapshot_require_literal 'memory.swap.max=0' || return 75
      _cgroup_snapshot_require_literal "cpu.max=$CGROUP_CPU_QUOTA_US/$CGROUP_CPU_PERIOD_US" || return 75
      _cgroup_snapshot_require_decimal 'memory.events.oom' || return 75
      _cgroup_snapshot_require_decimal 'memory.events.oom_kill' || return 75
      ;;
    family=v1)
      _cgroup_snapshot_require_literal "memory.limit_in_bytes=$CGROUP_MEMORY_BYTES" || return 75
      _cgroup_snapshot_require_literal "memory.memsw.limit_in_bytes=$CGROUP_MEMORY_BYTES" || return 75
      _cgroup_snapshot_require_literal "cpu.cfs_quota_us=$CGROUP_CPU_QUOTA_US" || return 75
      _cgroup_snapshot_require_literal "cpu.cfs_period_us=$CGROUP_CPU_PERIOD_US" || return 75
      _cgroup_snapshot_require_decimal 'memory.failcnt' || return 75
      _cgroup_snapshot_require_decimal 'memory.oom_control.oom_kill_disable' || return 75
      _cgroup_snapshot_require_decimal 'memory.oom_control.under_oom' || return 75
      ;;
    *) return 75 ;;
  esac
  [ -z "$cgroup_snapshot_remaining" ] || return 75
  return 0
}

capture_cgroup_snapshot() {
  [ "$#" -eq 0 ] || return 64
  assert_cgroup || return 75
  case "$cgroup_family" in
    v2)
      cgroup_snapshot_failure_fields="$(_cgroup_v2_oom_snapshot_fields "$cgroup_failure_file")" || return 75
      cgroup_snapshot="schema=hub.qa.cgroup-snapshot/v1,family=v2,memory.max=$CGROUP_MEMORY_BYTES,memory.swap.max=0,cpu.max=$CGROUP_CPU_QUOTA_US/$CGROUP_CPU_PERIOD_US,$cgroup_snapshot_failure_fields"
      ;;
    v1)
      [ -n "${cgroup_oom_control_file-}" ] || return 75
      cgroup_snapshot_failcnt="$(_cgroup_decimal_file_value "$cgroup_failure_file")" || return 75
      cgroup_snapshot_oom_control_fields="$(_cgroup_v1_oom_control_snapshot_fields "$cgroup_oom_control_file")" || return 75
      cgroup_snapshot="schema=hub.qa.cgroup-snapshot/v1,family=v1,memory.limit_in_bytes=$CGROUP_MEMORY_BYTES,memory.memsw.limit_in_bytes=$CGROUP_MEMORY_BYTES,cpu.cfs_quota_us=$CGROUP_CPU_QUOTA_US,cpu.cfs_period_us=$CGROUP_CPU_PERIOD_US,memory.failcnt=$cgroup_snapshot_failcnt,$cgroup_snapshot_oom_control_fields"
      ;;
    *) return 75 ;;
  esac
  _cgroup_snapshot_is_canonical "$cgroup_snapshot" || return 75
  printf '%s\n' "$cgroup_snapshot"
}

verify_cgroup_snapshot_unchanged() {
  [ "$#" -eq 1 ] || return 64
  cgroup_snapshot_baseline="$1"
  _cgroup_snapshot_is_canonical "$cgroup_snapshot_baseline" || return 75
  cgroup_snapshot_final="$(capture_cgroup_snapshot)" || return 75
  [ "$cgroup_snapshot_final" = "$cgroup_snapshot_baseline" ] || return 75
  return 0
}

_assert_tmpfs_mount() {
  [ "$#" -eq 6 ] || return 64
  tmpfs_path="$1"
  tmpfs_expected_bytes="$2"
  tmpfs_expected_uid="$3"
  tmpfs_expected_gid="$4"
  tmpfs_expected_mode="$5"
  tmpfs_execution_policy="$6"

  [ -d "$tmpfs_path" ] && [ ! -L "$tmpfs_path" ] || return 75
  tmpfs_resolved="$(readlink -f -- "$tmpfs_path")" || return 75
  [ "$tmpfs_resolved" = "$tmpfs_path" ] || return 75
  tmpfs_owner="$(stat -c '%u:%g:%a' -- "$tmpfs_path")" || return 75
  [ "$tmpfs_owner" = "$tmpfs_expected_uid:$tmpfs_expected_gid:$tmpfs_expected_mode" ] || return 75
  tmpfs_mount_record="$(awk -v target="$tmpfs_path" '
    $5 == target {
      count += 1
      for (mount_field = 7; mount_field <= NF; mount_field += 1) {
        if ($mount_field == "-") {
          if (mount_field + 3 <= NF) {
            type = $(mount_field + 1)
            options = $6 "," $(mount_field + 3)
          }
          break
        }
      }
    }
    END {
      if (count != 1 || type != "tmpfs" || options == "") {
        exit 75
      }
      print type ":" options
    }
  ' /proc/self/mountinfo)" || return 75
  case "$tmpfs_mount_record" in
    tmpfs:*) tmpfs_mount_options=${tmpfs_mount_record#tmpfs:} ;;
    *) return 75 ;;
  esac
  case ",$tmpfs_mount_options," in
    *,rw,*) ;;
    *) return 75 ;;
  esac
  case ",$tmpfs_mount_options," in
    *,nosuid,*) ;;
    *) return 75 ;;
  esac
  case ",$tmpfs_mount_options," in
    *,nodev,*) ;;
    *) return 75 ;;
  esac
  case "$tmpfs_execution_policy" in
    noexec)
      case ",$tmpfs_mount_options," in
        *,noexec,*) ;;
        *) return 75 ;;
      esac
      ;;
    exec)
      case ",$tmpfs_mount_options," in
        *,noexec,*) return 75 ;;
        *) ;;
      esac
      ;;
    *) return 64 ;;
  esac
  tmpfs_size="$(df -B1 --output=size "$tmpfs_path" 2>/dev/null | awk 'NR == 2 { print $1; exit }')" || return 75
  [ "$tmpfs_size" = "$tmpfs_expected_bytes" ] || return 75
  return 0
}

assert_owned_tmpfs() {
  [ "$#" -eq 0 ] || return 64
  [ "${GOCACHE-}" = /var/cache/hub-go ] || return 75
  [ "${GOTMPDIR-}" = /var/tmp/hub-go ] || return 75
  _assert_tmpfs_mount /tmp "$TMPFS_TMP_BYTES" 0 0 1777 noexec || return 75
  _assert_tmpfs_mount /var/cache/hub-go "$TMPFS_GOCACHE_BYTES" 10001 10001 1770 noexec || return 75
  _assert_tmpfs_mount /var/tmp/hub-go "$TMPFS_GOTMPDIR_BYTES" 10001 10001 1770 exec || return 75
  return 0
}

_owned_run_dir_identity() {
  [ "$#" -eq 0 ] || return 64
  [ -n "${run_dir-}" ] && [ -d "$run_dir" ] && [ ! -L "$run_dir" ] || return 75
  owned_run_dir_identity="$(stat -c '%d:%i:%u:%g:%a:%F' -- "$run_dir")" || return 75
  case "$owned_run_dir_identity" in
    [0-9]*:[0-9]*:10001:10001:700:directory) ;;
    *) return 75 ;;
  esac
  printf '%s\n' "$owned_run_dir_identity"
}

_owned_run_dir_is_current() {
  [ "$#" -eq 1 ] || return 64
  owned_run_dir_expected_identity="$1"
  owned_run_dir_current_identity="$(_owned_run_dir_identity)" || return 75
  [ "$owned_run_dir_current_identity" = "$owned_run_dir_expected_identity" ] || return 75
  return 0
}

_run_dir_target_is_safe() {
  [ "$#" -eq 1 ] || return 64
  atomic_target="$1"
  run_dir_safety_identity="$(_owned_run_dir_identity)" || return 75
  case "$atomic_target" in
    "$run_dir/"*) atomic_relative=${atomic_target#"$run_dir/"} ;;
    *) return 75 ;;
  esac
  case "$atomic_relative" in
    ''|.|..|./*|../*|*/./*|*/../*|*/.|*/..|*'//'*) return 75 ;;
    */*) return 75 ;;
    *) return 0 ;;
  esac
}

_owned_regular_identity() {
  [ "$#" -eq 1 ] || return 64
  owned_regular_path="$1"
  [ -f "$owned_regular_path" ] && [ ! -L "$owned_regular_path" ] || return 75
  owned_regular_identity="$(stat -c '%d:%i:%u:%g:%a:%F' -- "$owned_regular_path")" || return 75
  case "$owned_regular_identity" in
    [0-9]*:[0-9]*:10001:10001:600:'regular file') ;;
    [0-9]*:[0-9]*:10001:10001:600:'regular empty file')
      owned_regular_identity=${owned_regular_identity%'regular empty file'}'regular file'
      ;;
    *) return 75 ;;
  esac
  printf '%s\n' "$owned_regular_identity"
}

_owned_regular_links() {
  [ "$#" -eq 1 ] || return 64
  owned_links_path="$1"
  [ -f "$owned_links_path" ] && [ ! -L "$owned_links_path" ] || return 75
  owned_regular_links="$(stat -c '%h' -- "$owned_links_path")" || return 75
  case "$owned_regular_links" in
    ''|*[!0-9]*) return 75 ;;
    *) printf '%s\n' "$owned_regular_links" ;;
  esac
}

_owned_temp_name_is_safe() {
  [ "$#" -eq 3 ] || return 64
  owned_temp_directory="$1"
  owned_temp_label="$2"
  owned_temp_candidate="$3"
  [ -d "$owned_temp_directory" ] && [ ! -L "$owned_temp_directory" ] || return 75
  case "$owned_temp_label" in
    budget-all|budget-metadata|publish) ;;
    *) return 64 ;;
  esac
  case "$owned_temp_candidate" in
    "$owned_temp_directory/"*) owned_temp_leaf=${owned_temp_candidate#"$owned_temp_directory/"} ;;
    *) return 75 ;;
  esac
  case "$owned_temp_leaf" in
    ".backend-unit-$owned_temp_label."[A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9]) return 0 ;;
    *) return 75 ;;
  esac
}

_create_owned_temp() {
  [ "$#" -eq 2 ] || return 64
  owned_temp_directory="$1"
  owned_temp_label="$2"
  owned_temp_path="$(umask 077; mktemp "$owned_temp_directory/.backend-unit-$owned_temp_label.XXXXXXXX")" || return 75
  _owned_temp_name_is_safe "$owned_temp_directory" "$owned_temp_label" "$owned_temp_path" || return 75
  owned_temp_identity="$(_owned_regular_identity "$owned_temp_path")" || return 75
  owned_temp_links="$(_owned_regular_links "$owned_temp_path")" || return 75
  [ "$owned_temp_links" = 1 ] || return 75
  return 0
}

_owned_temp_is_current() {
  [ "$#" -eq 3 ] || return 64
  owned_current_path="$1"
  owned_expected_identity="$2"
  owned_expected_links="$3"
  owned_current_identity="$(_owned_regular_identity "$owned_current_path")" || return 75
  [ "$owned_current_identity" = "$owned_expected_identity" ] || return 75
  owned_current_links="$(_owned_regular_links "$owned_current_path")" || return 75
  [ "$owned_current_links" = "$owned_expected_links" ] || return 75
  return 0
}

_cleanup_owned_temp() {
  [ "$#" -eq 3 ] || return 64
  owned_cleanup_path="$1"
  owned_cleanup_identity="$2"
  owned_cleanup_links="$3"
  _owned_temp_is_current "$owned_cleanup_path" "$owned_cleanup_identity" "$owned_cleanup_links" || return 75
  rm -- "$owned_cleanup_path" || return 75
  [ ! -e "$owned_cleanup_path" ] && [ ! -L "$owned_cleanup_path" ] || return 75
  return 0
}

write_atomic() {
  [ "$#" -eq 2 ] || return 64
  atomic_target="$1"
  atomic_payload="$2"
  atomic_run_dir_identity="$(_owned_run_dir_identity)" || return 75
  _run_dir_target_is_safe "$atomic_target" || return 75
  _owned_run_dir_is_current "$atomic_run_dir_identity" || return 75
  [ ! -e "$atomic_target" ] && [ ! -L "$atomic_target" ] || return 75
  _create_owned_temp "$run_dir" publish || return 75
  atomic_temporary="$owned_temp_path"
  atomic_identity="$owned_temp_identity"
  _owned_run_dir_is_current "$atomic_run_dir_identity" || return 75
  if _owned_temp_is_current "$atomic_temporary" "$atomic_identity" 1 && printf '%s' "$atomic_payload" > "$atomic_temporary" && _owned_run_dir_is_current "$atomic_run_dir_identity" && _owned_temp_is_current "$atomic_temporary" "$atomic_identity" 1; then
    :
  else
    _owned_run_dir_is_current "$atomic_run_dir_identity" || return 75
    _cleanup_owned_temp "$atomic_temporary" "$atomic_identity" 1 || return 75
    return 75
  fi
  _owned_run_dir_is_current "$atomic_run_dir_identity" || return 75
  if ln -- "$atomic_temporary" "$atomic_target"; then
    :
  else
    _owned_run_dir_is_current "$atomic_run_dir_identity" || return 75
    _cleanup_owned_temp "$atomic_temporary" "$atomic_identity" 1 || return 75
    return 75
  fi
  _owned_run_dir_is_current "$atomic_run_dir_identity" || return 75
  _owned_temp_is_current "$atomic_temporary" "$atomic_identity" 2 || return 75
  atomic_target_identity="$(_owned_regular_identity "$atomic_target")" || return 75
  atomic_target_links="$(_owned_regular_links "$atomic_target")" || return 75
  [ "$atomic_target_identity" = "$atomic_identity" ] && [ "$atomic_target_links" = 2 ] || return 75
  _owned_run_dir_is_current "$atomic_run_dir_identity" || return 75
  _cleanup_owned_temp "$atomic_temporary" "$atomic_identity" 2 || return 75
  _owned_run_dir_is_current "$atomic_run_dir_identity" || return 75
  _owned_temp_is_current "$atomic_target" "$atomic_identity" 1 || return 75
  return 0
}

_report_bytes_from_find() {
  [ "$#" -eq 1 ] || return 64
  report_find_mode="$1"
  [ -n "${run_dir-}" ] && [ -d "$run_dir" ] && [ ! -L "$run_dir" ] || return 75
  [ -n "${GOTMPDIR-}" ] && [ -d "$GOTMPDIR" ] && [ ! -L "$GOTMPDIR" ] || return 75
  case "$report_find_mode" in
    all) report_temp_label=budget-all ;;
    metadata) report_temp_label=budget-metadata ;;
    *) return 64 ;;
  esac
  _create_owned_temp "$GOTMPDIR" "$report_temp_label" || return 75
  report_byte_list="$owned_temp_path"
  report_byte_identity="$owned_temp_identity"
  case "$report_find_mode" in
    all)
      if _owned_temp_is_current "$report_byte_list" "$report_byte_identity" 1 && find "$run_dir" -xdev -type f -links 1 -printf '%s\n' > "$report_byte_list" && _owned_temp_is_current "$report_byte_list" "$report_byte_identity" 1; then
        :
      else
        _cleanup_owned_temp "$report_byte_list" "$report_byte_identity" 1 || return 75
        return 75
      fi
      ;;
    metadata)
      if _owned_temp_is_current "$report_byte_list" "$report_byte_identity" 1 && find "$run_dir" -xdev -type f -links 1 ! -name '*.stdout.log' ! -name '*.stderr.log' -printf '%s\n' > "$report_byte_list" && _owned_temp_is_current "$report_byte_list" "$report_byte_identity" 1; then
        :
      else
        _cleanup_owned_temp "$report_byte_list" "$report_byte_identity" 1 || return 75
        return 75
      fi
      ;;
  esac
  if _owned_temp_is_current "$report_byte_list" "$report_byte_identity" 1 && awk 'BEGIN { total = 0; invalid = 0 } /^[0-9]+$/ { total += $1; next } { invalid = 1 } END { if (invalid) exit 75; printf "%.0f\n", total }' "$report_byte_list" && _owned_temp_is_current "$report_byte_list" "$report_byte_identity" 1; then
    :
  else
    _cleanup_owned_temp "$report_byte_list" "$report_byte_identity" 1 || return 75
    return 75
  fi
  _cleanup_owned_temp "$report_byte_list" "$report_byte_identity" 1 || return 75
  return 0
}

report_regular_file_bytes() {
  [ "$#" -eq 0 ] || return 64
  _report_bytes_from_find all
}

report_metadata_bytes() {
  [ "$#" -eq 0 ] || return 64
  _report_bytes_from_find metadata
}

stage_stream_cap() {
  stage="$1"
  stream="$2"
  case "$stage:$stream" in
    build:stdout) printf '%s\n' "$BUILD_STDOUT_CAP_BYTES" ;;
    build:stderr) printf '%s\n' "$BUILD_STDERR_CAP_BYTES" ;;
    vet:stdout) printf '%s\n' "$VET_STDOUT_CAP_BYTES" ;;
    vet:stderr) printf '%s\n' "$VET_STDERR_CAP_BYTES" ;;
    unit:stdout) printf '%s\n' "$UNIT_STDOUT_CAP_BYTES" ;;
    unit:stderr) printf '%s\n' "$UNIT_STDERR_CAP_BYTES" ;;
    race:stdout) printf '%s\n' "$RACE_STDOUT_CAP_BYTES" ;;
    race:stderr) printf '%s\n' "$RACE_STDERR_CAP_BYTES" ;;
    *) return 64 ;;
  esac
}

check_budget() {
  report_bytes="$(report_regular_file_bytes)"
  metadata_bytes="$(report_metadata_bytes)"
  [ "$report_bytes" -le "$REPORT_BUDGET_BYTES" ] || return 75
  [ "$metadata_bytes" -le "$REPORT_METADATA_BUDGET_BYTES" ] || return 75
}

_owned_go_fifos_are_safe() {
  [ "$#" -eq 0 ] || return 64
  [ -n "${stdout_fifo-}" ] && [ -n "${stderr_fifo-}" ] || return 75
  [ -p "$stdout_fifo" ] && [ ! -L "$stdout_fifo" ] || return 75
  [ -p "$stderr_fifo" ] && [ ! -L "$stderr_fifo" ] || return 75
  return 0
}

start_owned_go_child_group() {
  [ "$#" -eq 2 ] || return 64
  stage="$1"
  package="$2"
  _owned_go_fifos_are_safe || return 75
  case "${go_child_pid-}" in
    '') ;;
    *) return 75 ;;
  esac
  case "$stage" in
    build)
      [ -n "${owned_output-}" ] || return 75
      case "$owned_output" in
        "$GOTMPDIR/bin/"*) ;;
        *) return 75 ;;
      esac
      setsid go build -p=1 -o "$owned_output" "$package" > "$stdout_fifo" 2> "$stderr_fifo" &
      ;;
    vet)
      setsid go vet -tags hubtest -p=1 "$package" > "$stdout_fifo" 2> "$stderr_fifo" &
      ;;
    unit)
      setsid go test -tags hubtest -json -p=1 -count=1 -timeout=150s "$package" > "$stdout_fifo" 2> "$stderr_fifo" &
      ;;
    race)
      setsid go test -tags hubtest -json -race -p=1 -count=1 -timeout=150s "$package" > "$stdout_fifo" 2> "$stderr_fifo" &
      ;;
    *) return 64 ;;
  esac
  go_child_pid="$!"
  go_child_session=
  go_child_starttime=
  go_child_pgid=
  go_child_identity=
  go_child_reaped=false
  return 0
}

_go_child_identity_from_proc() {
  [ "$#" -eq 1 ] || return 64
  go_identity_pid="$1"
  case "$go_identity_pid" in
    ''|*[!0-9]*) return 75 ;;
    *) ;;
  esac
  [ -r "/proc/$go_identity_pid/stat" ] || return 75
  go_identity_value="$(awk -v identifier="$go_identity_pid" '$1 == identifier && $2 == "(go)" { print $1 ":" $5 ":" $6 ":" $22 ":go"; exit }' "/proc/$go_identity_pid/stat")" || return 75
  case "$go_identity_value" in
    "$go_identity_pid:$go_identity_pid:$go_identity_pid:"[0-9]*:go) ;;
    *) return 75 ;;
  esac
  printf '%s\n' "$go_identity_value"
}

_capture_owned_go_child_identity() {
  [ "$#" -eq 0 ] || return 64
  case "${go_child_pid-}" in
    ''|*[!0-9]*) return 75 ;;
    *) ;;
  esac
  go_child_identity="$(_go_child_identity_from_proc "$go_child_pid")" || return 75
  go_identity_tail=${go_child_identity#*:}
  go_child_pgid=${go_identity_tail%%:*}
  go_identity_tail=${go_identity_tail#*:}
  go_child_session=${go_identity_tail%%:*}
  go_identity_tail=${go_identity_tail#*:}
  go_child_starttime=${go_identity_tail%%:*}
  case "$go_child_pgid:$go_child_session:$go_child_starttime" in
    "$go_child_pid:$go_child_pid:"[0-9]*) ;;
    *) return 75 ;;
  esac
  return 0
}

owned_go_child_is_live() {
  [ "$#" -eq 0 ] || return 64
  [ -n "${go_child_identity-}" ] || return 1
  case "${go_child_pid-}" in
    ''|*[!0-9]*) return 1 ;;
    *) ;;
  esac
  go_child_current_identity="$(_go_child_identity_from_proc "$go_child_pid")" || return 1
  [ "$go_child_current_identity" = "$go_child_identity" ] || return 1
  return 0
}

_owned_go_leader_state() {
  [ "$#" -eq 0 ] || return 64
  case "${go_child_pid-}" in ''|*[!0-9]*) return 75 ;; esac
  case "${go_child_pgid-}" in ''|*[!0-9]*) return 75 ;; esac
  case "${go_child_session-}" in ''|*[!0-9]*) return 75 ;; esac
  case "${go_child_starttime-}" in ''|*[!0-9]*) return 75 ;; esac
  case "${go_child_identity-}" in
    "$go_child_pid:$go_child_pgid:$go_child_session:$go_child_starttime:go") ;;
    *) return 75 ;;
  esac
  if [ ! -r "/proc/$go_child_pid/stat" ]; then
    printf '%s\n' missing
    return 0
  fi
  go_leader_proc="$(
    awk -v identifier="$go_child_pid" '$1 == identifier { print $1 ":" $2 ":" $3 ":" $5 ":" $6 ":" $22; exit }' "/proc/$go_child_pid/stat"
  )" || return 75
  if [ -z "$go_leader_proc" ]; then
    printf '%s\n' missing
    return 0
  fi
  go_leader_pid=${go_leader_proc%%:*}
  go_leader_tail=${go_leader_proc#*:}
  go_leader_comm=${go_leader_tail%%:*}
  go_leader_tail=${go_leader_tail#*:}
  go_leader_stat=${go_leader_tail%%:*}
  go_leader_tail=${go_leader_tail#*:}
  go_leader_pgid=${go_leader_tail%%:*}
  go_leader_tail=${go_leader_tail#*:}
  go_leader_session=${go_leader_tail%%:*}
  go_leader_starttime=${go_leader_tail#*:}
  case "$go_leader_pid:$go_leader_comm:$go_leader_pgid:$go_leader_session:$go_leader_starttime" in
    "$go_child_pid:(go):$go_child_pgid:$go_child_session:$go_child_starttime") ;;
    *) printf '%s\n' invalid; return 0 ;;
  esac
  case "$go_leader_stat" in
    Z) printf '%s\n' valid-zombie ;;
    [A-Za-z]) printf '%s\n' valid-live ;;
    *) printf '%s\n' invalid ;;
  esac
}

_assert_owned_signal_identity() {
  [ "$#" -eq 0 ] || return 64
  [ -n "${go_child_identity-}" ] || return 75
  case "${go_child_pid-}" in ''|*[!0-9]*) return 75 ;; esac
  case "${go_child_pgid-}" in ''|*[!0-9]*) return 75 ;; esac
  case "${go_child_session-}" in ''|*[!0-9]*) return 75 ;; esac
  case "${go_child_starttime-}" in ''|*[!0-9]*) return 75 ;; esac
  case "$go_child_identity" in
    "$go_child_pid:$go_child_pgid:$go_child_session:$go_child_starttime:go") ;;
    *) return 75 ;;
  esac
  go_child_signal_identity="$(_go_child_identity_from_proc "$go_child_pid")" || return 75
  [ "$go_child_signal_identity" = "$go_child_identity" ] || return 75
  go_child_signal_group_state="$(_owned_go_group_state)" || return 75
  [ "$go_child_signal_group_state" = valid-live ] || return 75
  return 0
}

_owned_go_group_state() {
  [ "$#" -eq 0 ] || return 64
  case "${go_child_pgid-}" in
    ''|*[!0-9]*) return 75 ;;
    *) ;;
  esac
  case "${go_child_session-}" in
    ''|*[!0-9]*) return 75 ;;
    *) ;;
  esac
  go_group_state="$(ps -e -o pgid= -o sid= -o stat= | awk -v group="$go_child_pgid" -v session="$go_child_session" '$1 == group { count += 1; if ($2 != session) invalid = 1; if ($3 !~ /^Z/) live = 1 } END { if (invalid) print "invalid"; else if (count == 0) print "none"; else if (live) print "valid-live"; else print "valid-zombie" }')" || return 75
  case "$go_group_state" in
    none|valid-live|valid-zombie|invalid) printf '%s\n' "$go_group_state" ;;
    *) return 75 ;;
  esac
}

_owned_go_group_reaches_terminal_state() {
  [ "$#" -eq 1 ] || return 64
  group_grace_remaining="$1"
  case "$group_grace_remaining" in
    ''|*[!0-9]*) return 64 ;;
    *) ;;
  esac
  while [ "$group_grace_remaining" -gt 0 ]; do
    group_state="$(_owned_go_group_state)" || return 75
    case "$group_state" in
      none) return 0 ;;
      valid-live|valid-zombie) sleep 1 ;;
      invalid) return 75 ;;
    esac
    group_grace_remaining=$((group_grace_remaining - 1))
  done
  group_state="$(_owned_go_group_state)" || return 75
  case "$group_state" in
    none) return 0 ;;
    valid-live|valid-zombie) return 1 ;;
    invalid) return 75 ;;
  esac
}

monitor_owned_budget_stop() {
  [ "$#" -eq 2 ] || return 64
  monitored_pid="$1"
  budget_stop_marker="$2"
  [ "$monitored_pid" = "${go_child_pid-}" ] || return 75
  [ "${GOTMPDIR-}" = /var/tmp/hub-go ] || return 75
  case "${stage-}" in build|vet|unit|race) ;; *) return 75 ;; esac
  case "${stage_scratch-}" in
    "$GOTMPDIR/.backend-unit-stage-$stage."[A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9][A-Za-z0-9]) ;;
    *) return 75 ;;
  esac
  _execute_scratch_is_current "$stage_scratch" "${stage_scratch_identity-}" || return 75
  case "${package_ordinal-}" in ''|*[!0-9]*) return 75 ;; esac
  [ "$package_ordinal" -ge 1 ] || return 75
  [ "$budget_stop_marker" = "$stage_scratch/stop.$package_ordinal" ] || return 75
  [ ! -L "$budget_stop_marker" ] || return 75
  if [ -z "${go_child_identity-}" ]; then
    if _capture_owned_go_child_identity; then
      :
    else
      return 75
    fi
  fi
  while :; do
    leader_state="$(_owned_go_leader_state)" || return 75
    group_state="$(_owned_go_group_state)" || return 75
    case "$leader_state:$group_state" in
      valid-live:valid-live)
        [ ! -e "$budget_stop_marker" ] || return 1
        sleep 1
        ;;
      valid-zombie:valid-zombie|missing:none)
        [ ! -e "$budget_stop_marker" ] || return 1
        return 0
        ;;
      valid-zombie:valid-live) return 75 ;;
      valid-live:valid-zombie|valid-live:none|valid-zombie:none|invalid:*|missing:*|*:invalid) return 75 ;;
      *) return 75 ;;
    esac
  done
}

terminate_owned_go_child_group() {
  [ "$#" -eq 0 ] || return 64
  if [ -z "${go_child_identity-}" ]; then
    _capture_owned_go_child_identity || return 75
  fi
  group_state="$(_owned_go_group_state)" || return 75
  leader_state="$(_owned_go_leader_state)" || return 75
  case "$leader_state:$group_state" in
    valid-zombie:valid-zombie) return 0 ;;
    valid-live:valid-live|valid-zombie:valid-live) ;;
    *) return 75 ;;
  esac
  _assert_owned_signal_identity || return 75
  kill -TERM "-$go_child_pgid" || return 75
  if _owned_go_group_reaches_terminal_state 5; then
    return 0
  fi
  leader_state="$(_owned_go_leader_state)" || return 75
  group_state="$(_owned_go_group_state)" || return 75
  case "$leader_state:$group_state" in
    valid-zombie:valid-zombie) return 0 ;;
    valid-live:valid-live|valid-zombie:valid-live) ;;
    *) return 75 ;;
  esac
  _assert_owned_signal_identity || return 75
  kill -KILL "-$go_child_pgid" || return 75
  if _owned_go_group_reaches_terminal_state 5; then
    return 0
  fi
  leader_state="$(_owned_go_leader_state)" || return 75
  group_state="$(_owned_go_group_state)" || return 75
  [ "$leader_state:$group_state" = valid-zombie:valid-zombie ] || return 75
  return 0
}

_owned_go_group_reaches_reapable_state_for_container_stop() {
  [ "$#" -eq 1 ] || return 64
  container_stop_grace_remaining="$1"
  case "$container_stop_grace_remaining" in
    1) ;;
    *) return 64 ;;
  esac
  while :; do
    container_stop_leader_state="$(_owned_go_leader_state)" || return 75
    container_stop_group_state="$(_owned_go_group_state)" || return 75
    case "$container_stop_leader_state:$container_stop_group_state" in
      valid-zombie:valid-zombie) return 0 ;;
      valid-live:valid-live|valid-zombie:valid-live)
        [ "$container_stop_grace_remaining" -eq 1 ] || return 1
        sleep 1
        container_stop_grace_remaining=0
        ;;
      *) return 75 ;;
    esac
  done
}

abort_owned_go_child_group_for_container_stop() {
  [ "$#" -eq 0 ] || return 64
  case "${go_child_reaped-false}" in
    true)
      [ -z "${go_child_pid-}" ] && [ -z "${go_child_pgid-}" ] && [ -z "${go_child_session-}" ] && [ -z "${go_child_starttime-}" ] && [ -z "${go_child_identity-}" ] || return 75
      case "${go_child_exit-}" in ''|*[!0-9]*) return 75 ;; *) return 0 ;; esac
      ;;
    false) ;;
    *) return 75 ;;
  esac
  case "${go_child_pid-}" in ''|*[!0-9]*) return 75 ;; *) ;; esac
  case "${go_child_pgid-}" in ''|*[!0-9]*) return 75 ;; *) ;; esac
  case "${go_child_session-}" in ''|*[!0-9]*) return 75 ;; *) ;; esac
  case "${go_child_starttime-}" in ''|*[!0-9]*) return 75 ;; *) ;; esac
  case "$go_child_pid:$go_child_pgid:$go_child_session:$go_child_starttime" in
    "$go_child_pid:$go_child_pid:$go_child_pid:"[0-9]*) ;;
    *) return 75 ;;
  esac
  case "$go_child_identity" in
    "$go_child_pid:$go_child_pgid:$go_child_session:$go_child_starttime:go") ;;
    *) return 75 ;;
  esac
  container_stop_leader_state="$(_owned_go_leader_state)" || return 75
  container_stop_group_state="$(_owned_go_group_state)" || return 75
  case "$container_stop_leader_state:$container_stop_group_state" in
    valid-zombie:valid-zombie) ;;
    valid-live:valid-live|valid-zombie:valid-live)
      _assert_owned_signal_identity || return 75
      kill -TERM "-$go_child_pgid" || return 75
      if _owned_go_group_reaches_reapable_state_for_container_stop 1; then
        :
      else
        container_stop_wait_status=$?
        [ "$container_stop_wait_status" -eq 1 ] || return 75
        container_stop_leader_state="$(_owned_go_leader_state)" || return 75
        container_stop_group_state="$(_owned_go_group_state)" || return 75
        case "$container_stop_leader_state:$container_stop_group_state" in
          valid-live:valid-live|valid-zombie:valid-live) ;;
          *) return 75 ;;
        esac
        _assert_owned_signal_identity || return 75
        kill -KILL "-$go_child_pgid" || return 75
        _owned_go_group_reaches_reapable_state_for_container_stop 1 || return 75
      fi
      ;;
    *) return 75 ;;
  esac
  container_stop_leader_state="$(_owned_go_leader_state)" || return 75
  container_stop_group_state="$(_owned_go_group_state)" || return 75
  [ "$container_stop_leader_state:$container_stop_group_state" = valid-zombie:valid-zombie ] || return 75
  if wait "$go_child_pid"; then
    go_child_exit=0
  else
    go_child_exit=$?
  fi
  go_child_pid=
  go_child_reaped=true
  container_stop_group_state="$(_owned_go_group_state)" || return 75
  [ "$container_stop_group_state" = none ] || return 75
  go_child_session=
  go_child_starttime=
  go_child_pgid=
  go_child_identity=
  return 0
}

_post_reaped_gocache_cleanup() {
  [ "$#" -eq 0 ] || return 64
  [ "${go_child_reaped-}" = true ] && [ -z "${go_child_pid-}" ] || return 75
  [ "${GOCACHE-}" = /var/cache/hub-go ] || return 75
  _assert_tmpfs_mount /var/cache/hub-go "$TMPFS_GOCACHE_BYTES" 10001 10001 1770 noexec || return 75
  gocache_used_bytes="$(df -B1 --output=used "$GOCACHE" 2>/dev/null | awk 'NR == 2 { print $1; exit }')" || return 75
  case "$gocache_used_bytes" in
    ''|*[!0-9]*) return 75 ;;
    *) ;;
  esac
  if [ "$gocache_used_bytes" -ge "$GOCACHE_WATERMARK_BYTES" ]; then
    go clean -cache || return 75
    _assert_tmpfs_mount /var/cache/hub-go "$TMPFS_GOCACHE_BYTES" 10001 10001 1770 noexec || return 75
  fi
  return 0
}

reap_owned_go_child_group() {
  [ "$#" -eq 0 ] || return 64
  [ "${go_child_reaped-}" = false ] || return 75
  case "${go_child_pid-}" in
    ''|*[!0-9]*) return 75 ;;
    *) ;;
  esac
  if [ -z "${go_child_identity-}" ]; then
    _capture_owned_go_child_identity || return 75
  fi
  leader_state="$(_owned_go_leader_state)" || return 75
  group_state="$(_owned_go_group_state)" || return 75
  case "$leader_state:$group_state" in
    valid-zombie:valid-zombie|missing:none) ;;
    *) return 75 ;;
  esac
  if wait "$go_child_pid"; then
    go_child_exit=0
  else
    go_child_exit=$?
  fi
  go_child_pid=
  go_child_reaped=true
  if _owned_go_group_reaches_terminal_state 5; then
    :
  else
    return 75
  fi
  _post_reaped_gocache_cleanup || return 75
  go_child_session=
  go_child_starttime=
  go_child_pgid=
  go_child_identity=
  return 0
}
