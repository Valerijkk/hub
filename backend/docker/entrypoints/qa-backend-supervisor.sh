#!/bin/ash

set -eu
set -f
umask 077

FIRST_CHILD_TIMEOUT_SECONDS=30
current_pid=''
current_dir=''
current_fifo=''
control_open=0
control_record=''
record_token=''
child_exit_status=0

close_control_fd() {
  if [ "$control_open" -eq 1 ]; then
    exec 4<&-
    control_open=0
  fi
}

cleanup_owned_dir() {
  close_control_fd

  if [ -n "$current_fifo" ]; then
    rm -f -- "$current_fifo" || :
    current_fifo=''
  fi

  if [ -n "$current_dir" ]; then
    case "$current_dir" in
      /tmp/qa-backend-supervisor.??????)
        if ! rmdir "$current_dir"; then
          printf '%s\n' 'qa-backend-supervisor: cleanup failure' >&2
        fi
        ;;
      *)
        printf '%s\n' 'qa-backend-supervisor: cleanup failure' >&2
        ;;
    esac
    current_dir=''
  fi
}

terminate_and_reap_current() {
  close_control_fd
  if [ -n "$current_pid" ]; then
    kill -TERM "$current_pid" 2>/dev/null || :
    wait "$current_pid" 2>/dev/null || :
    current_pid=''
  fi
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
  printf '%s\n' 'qa-backend-supervisor: signal received' >&2
  close_control_fd
  if [ -n "$current_pid" ]; then
    case "$signal_name" in
      TERM) kill -TERM "$current_pid" 2>/dev/null || : ;;
      INT) kill -INT "$current_pid" 2>/dev/null || : ;;
    esac
    wait "$current_pid" 2>/dev/null || :
    current_pid=''
  fi
  cleanup_owned_dir
  trap - EXIT
  exit 128
}

trap 'cleanup_owned_dir' EXIT
trap 'forward_signal TERM' TERM
trap 'forward_signal INT' INT

if [ "$#" -ne 0 ]; then
  fail_launch
fi

launch_child() {
  launch_generation=$1
  launch_mode=$2

  current_dir=$(mktemp -d /tmp/qa-backend-supervisor.XXXXXX) || fail_launch
  case "$current_dir" in
    /tmp/qa-backend-supervisor.??????) ;;
    *) fail_launch ;;
  esac
  [ -d "$current_dir" ] || fail_launch

  current_fifo="$current_dir/control"
  mkfifo "$current_fifo" || fail_launch
  chmod 0600 "$current_fifo" || fail_launch
  [ -p "$current_fifo" ] || fail_launch

  (
    HUB_QA_CHILD_CONTROL_FD=3
    HUB_QA_CHILD_GENERATION="$launch_generation"
    HUB_QA_CHILD_MODE="$launch_mode"
    export HUB_QA_CHILD_CONTROL_FD HUB_QA_CHILD_GENERATION HUB_QA_CHILD_MODE
    exec 3>"$current_fifo"
    exec /app/hub-server
  ) &
  current_pid=$!

  if ! exec 4<"$current_fifo"; then
    fail_launch
  fi
  control_open=1
  if ! rm -f -- "$current_fifo"; then
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
  close_control_fd
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
cleanup_owned_dir

if [ "$normal_restart_seen" -ne 1 ] || [ "$child_exit_status" -ne 0 ]; then
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
cleanup_owned_dir
exit "$child_exit_status"
