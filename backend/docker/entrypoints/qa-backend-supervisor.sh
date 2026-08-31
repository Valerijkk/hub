#!/bin/ash

set -eu
set -f
umask 077

FIRST_CHILD_TIMEOUT_SECONDS=30
FAILURE_TERM_GRACE_SECONDS=5
SIGNAL_GRACE_SECONDS=5
current_pid=''
current_dir=''
current_fifo=''
control_open=0
control_record=''
record_token=''
child_exit_status=0
launch_in_progress=0
recovery_forbidden=0
cleanup_failed=0
signal_handling=0

close_control_fd() {
  if [ "$control_open" -eq 1 ]; then
    if ! { exec 4<&-; } 2>/dev/null; then
      return 1
    fi
    control_open=0
  fi
  return 0
}

cleanup_owned_dir() {
  cleanup_status=0
  if ! close_control_fd; then
    cleanup_status=1
  fi

  if [ -n "$current_fifo" ]; then
    if ! rm -f -- "$current_fifo" 2>/dev/null; then
      cleanup_status=1
    else
      current_fifo=''
    fi
  fi

  if [ -n "$current_dir" ]; then
    case "$current_dir" in
      /tmp/qa-backend-supervisor.??????)
        if ! rmdir "$current_dir" 2>/dev/null; then
          cleanup_status=1
        else
          current_dir=''
        fi
        ;;
      *)
        cleanup_status=1
        ;;
    esac
  fi

  if [ "$cleanup_status" -ne 0 ]; then
    cleanup_failed=1
    printf '%s\n' 'qa-backend-supervisor: cleanup failure' >&2
    return 1
  fi
  return 0
}

reap_failed_child() {
  if [ -n "$current_pid" ]; then
    kill -TERM "$current_pid" 2>/dev/null || :
    # A failed child gets a bounded TERM grace period before forced reap.
    sleep "$FAILURE_TERM_GRACE_SECONDS" 2>/dev/null || :
    if kill -0 "$current_pid" 2>/dev/null; then
      kill -KILL "$current_pid" 2>/dev/null || :
    fi
    wait "$current_pid" 2>/dev/null || :
    current_pid=''
  fi
}

reap_signaled_child() {
  received_signal=$1
  if [ -n "$current_pid" ]; then
    case "$received_signal" in
      TERM) kill -TERM "$current_pid" 2>/dev/null || : ;;
      INT) kill -INT "$current_pid" 2>/dev/null || : ;;
    esac
    # PID 1 must not remain in a signal trap behind a stuck child.
    sleep "$SIGNAL_GRACE_SECONDS" 2>/dev/null || :
    if kill -0 "$current_pid" 2>/dev/null; then
      kill -KILL "$current_pid" 2>/dev/null || :
    fi
    wait "$current_pid" 2>/dev/null || :
    current_pid=''
  fi
}

terminate_and_reap_current() {
  recovery_forbidden=1
  close_control_fd || :
  reap_failed_child
  cleanup_owned_dir
}

fail_launch() {
  printf '%s\n' 'qa-backend-supervisor: child launch failure' >&2
  terminate_and_reap_current
  exit 1
}

fail_protocol() {
  printf '%s\n' 'qa-backend-supervisor: protocol failure' >&2
  terminate_and_reap_current
  exit 1
}

forward_signal() {
  signal_name=$1
  if [ "$signal_handling" -ne 0 ]; then
    trap '' TERM INT
    exit 128
  fi
  signal_handling=1
  trap '' TERM INT
  recovery_forbidden=1
  # BusyBox ash dispatches pending traps at command boundaries; after `)&`,
  # `$!` is the wrapper PID until the parent records it below.
  if [ -z "$current_pid" ] && [ "$launch_in_progress" -eq 1 ]; then
    current_pid=$!
  fi
  printf '%s\n' 'qa-backend-supervisor: signal received' >&2
  close_control_fd || :
  reap_signaled_child "$signal_name"
  cleanup_owned_dir || :
  trap - EXIT
  exit 128
}

finish_exit() {
  final_status=$?
  if [ "$cleanup_failed" -ne 0 ] || ! cleanup_owned_dir; then
    trap - EXIT
    exit 1
  fi
  trap - EXIT
  exit "$final_status"
}

trap 'finish_exit' EXIT
trap 'forward_signal TERM' TERM
trap 'forward_signal INT' INT

if [ "$#" -ne 0 ]; then
  fail_launch
fi

launch_child() {
  launch_generation=$1
  launch_mode=$2

  if ! current_dir=$(mktemp -d /tmp/qa-backend-supervisor.XXXXXX 2>/dev/null); then
    fail_launch
  fi
  case "$current_dir" in
    /tmp/qa-backend-supervisor.??????) ;;
    *) fail_launch ;;
  esac
  [ -d "$current_dir" ] || fail_launch

  current_fifo="$current_dir/control"
  if ! mkfifo "$current_fifo" 2>/dev/null; then
    fail_launch
  fi
  if ! chmod 0600 "$current_fifo" 2>/dev/null; then
    fail_launch
  fi
  [ -p "$current_fifo" ] || fail_launch

  # Immutable image preflight keeps launch validation on fixed diagnostics.
  if [ ! -e /app/hub-server ] || [ ! -x /app/hub-server ]; then
    fail_launch
  fi

  launch_in_progress=1
  (
    HUB_QA_CHILD_CONTROL_FD=3
    HUB_QA_CHILD_GENERATION="$launch_generation"
    HUB_QA_CHILD_MODE="$launch_mode"
    export HUB_QA_CHILD_CONTROL_FD HUB_QA_CHILD_GENERATION HUB_QA_CHILD_MODE
    if ! { exec 3>"$current_fifo"; } 2>/dev/null; then
      exit 125
    fi
    exec /app/hub-server
  ) &
  current_pid=$!
  launch_in_progress=0

  if ! { exec 4<"$current_fifo"; } 2>/dev/null; then
    fail_launch
  fi
  control_open=1
  if ! rm -f -- "$current_fifo" 2>/dev/null; then
    fail_launch
  fi
  current_fifo=''
}

read_first_record() {
  control_record=''
  if ! IFS= read -r -t "$FIRST_CHILD_TIMEOUT_SECONDS" -n 129 control_record <&4; then
    return 1
  fi
  [ "${#control_record}" -le 127 ] || return 2
  return 0
}

read_next_record() {
  control_record=''
  if IFS= read -r -n 129 control_record <&4; then
    [ "${#control_record}" -le 127 ] || return 2
    return 0
  fi
  [ -z "$control_record" ] && return 1
  return 2
}

is_positive_decimal() {
  case "$1" in
    ''|0|0*|*[!0-9]*) return 1 ;;
    *) return 0 ;;
  esac
}

is_lower_hex_token() {
  [ "${#1}" -eq 64 ] || return 1
  case "$1" in
    *[!0123456789abcdef]*) return 1 ;;
    *) return 0 ;;
  esac
}

validate_child_record() {
  expected_pid=$1
  expected_generation=$2
  expected_mode=$3

  if [ -z "$control_record" ]; then
    return 1
  fi
  case "$control_record" in
    *'  '*|' '*|*' ') return 1 ;;
  esac
  saved_ifs=$IFS
  IFS=' '
  set -- $control_record
  IFS=$saved_ifs
  [ "$#" -eq 6 ] || return 1
  [ "$1" = CHILD ] || return 1
  [ "$2" = 1 ] || return 1
  is_positive_decimal "$3" || return 1
  [ "$3" = "$expected_pid" ] || return 1
  is_positive_decimal "$4" || return 1
  [ "$4" = "$expected_generation" ] || return 1
  [ "$5" = "$expected_mode" ] || return 1
  case "$5" in normal|recovery) ;; *) return 1 ;; esac
  is_lower_hex_token "$6" || return 1
  record_token=$6
  return 0
}

validate_restart_record() {
  expected_pid=$1
  expected_generation=$2

  if [ -z "$control_record" ]; then
    return 1
  fi
  case "$control_record" in
    *'  '*|' '*|*' ') return 1 ;;
  esac
  saved_ifs=$IFS
  IFS=' '
  set -- $control_record
  IFS=$saved_ifs
  [ "$#" -eq 4 ] || return 1
  [ "$1" = RESTART ] || return 1
  [ "$2" = 1 ] || return 1
  is_positive_decimal "$3" || return 1
  [ "$3" = "$expected_pid" ] || return 1
  is_positive_decimal "$4" || return 1
  [ "$4" = "$expected_generation" ] || return 1
  return 0
}

consume_normal_stream() {
  normal_restart_seen=0
  while :; do
    if read_next_record; then
      if [ "$normal_restart_seen" -ne 0 ] || ! validate_restart_record "$current_pid" 1; then
        return 2
      fi
      normal_restart_seen=1
      continue
    fi
    read_result=$?
    if [ "$read_result" -eq 1 ]; then
      break
    fi
    return 2
  done
  return "$normal_restart_seen"
}

consume_recovery_stream() {
  while :; do
    if read_next_record; then
      return 2
    fi
    read_result=$?
    if [ "$read_result" -eq 1 ]; then
      return 0
    fi
    return 2
  done
}

wait_for_current_child() {
  if ! close_control_fd; then
    fail_protocol
  fi
  if wait "$current_pid"; then
    child_exit_status=0
  else
    child_exit_status=$?
  fi
  current_pid=''
}

launch_child 1 normal
normal_pid=$current_pid
if ! read_first_record || ! validate_child_record "$normal_pid" 1 normal; then
  fail_protocol
fi
normal_token=$record_token
record_token=''
control_record=''

if consume_normal_stream; then
  normal_restart_seen=0
else
  normal_stream_result=$?
  if [ "$normal_stream_result" -eq 1 ]; then
    normal_restart_seen=1
  else
    fail_protocol
  fi
fi
wait_for_current_child
if ! cleanup_owned_dir; then
  exit 1
fi

if [ "$recovery_forbidden" -ne 0 ] || [ "$normal_restart_seen" -ne 1 ] || [ "$child_exit_status" -ne 0 ]; then
  exit "$child_exit_status"
fi

launch_child 2 recovery
recovery_pid=$current_pid
if ! read_first_record || ! validate_child_record "$recovery_pid" 2 recovery; then
  fail_protocol
fi
if [ "$recovery_pid" = "$normal_pid" ] || [ "$record_token" = "$normal_token" ]; then
  fail_protocol
fi
normal_token=''
record_token=''
control_record=''

if ! consume_recovery_stream; then
  fail_protocol
fi
wait_for_current_child
if ! cleanup_owned_dir; then
  exit 1
fi
exit "$child_exit_status"
