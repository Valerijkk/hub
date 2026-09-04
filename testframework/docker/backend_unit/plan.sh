#!/bin/sh
set -eu

PACKAGE_PLAN_SCHEMA=hub.qa.backend-unit-package-plan/v1
PACKAGE_PLAN_TARGET=backend-unit
PACKAGE_PLAN_RESULT_PENDING=PENDING
PACKAGE_DISCOVERY_FORMAT='{{printf "%s\t%s" .ImportPath .Name}}'

_plan_require_context() {
  [ "$#" -eq 0 ] || return 64
  [ -n "${run_dir-}" ] && [ -d "$run_dir" ] && [ ! -L "$run_dir" ] || return 75
  [ -n "${GOTMPDIR-}" ] && [ -d "$GOTMPDIR" ] && [ ! -L "$GOTMPDIR" ] || return 75
  [ -n "${run_id-}" ] || return 75
  [ -n "${owner_nonce-}" ] || return 75
  [ -n "${go_environment_sha256-}" ] || return 75
  printf '%s\n' "$run_id" | grep -Eq '^backend-unit-[0-9]{8}T[0-9]{6}Z-[0-9a-f]{12,64}-[0-9a-f]{8}$' || return 75
  printf '%s\n' "$owner_nonce" | grep -Eq '^[0-9a-f]{32}$' || return 75
  printf '%s\n' "$go_environment_sha256" | grep -Eq '^sha256:[0-9a-f]{64}$' || return 75
  return 0
}

_plan_stage_is_valid() {
  [ "$#" -eq 1 ] || return 64
  case "$1" in
    build|vet|unit|race) return 0 ;;
    *) return 64 ;;
  esac
}

_plan_import_path_is_valid() {
  [ "$#" -eq 1 ] || return 64
  case "$1" in
    ''|/*|*/|*//*|*'\t'*|*[!A-Za-z0-9._~+/-]*) return 75 ;;
    *) ;;
  esac
  plan_import_path_remaining="$1"
  while :; do
    plan_import_path_segment=${plan_import_path_remaining%%/*}
    case "$plan_import_path_segment" in
      .|..) return 75 ;;
      *) ;;
    esac
    [ "$plan_import_path_remaining" = "$plan_import_path_segment" ] && break
    plan_import_path_remaining=${plan_import_path_remaining#*/}
  done
  return 0
}

_plan_package_name_is_valid() {
  [ "$#" -eq 1 ] || return 64
  case "$1" in
    [A-Za-z_]* ) ;;
    *) return 75 ;;
  esac
  case "$1" in
    *[!A-Za-z0-9_]*|*'\t'*) return 75 ;;
    *) return 0 ;;
  esac
}

_plan_prepare_owned_directories() {
  [ "$#" -eq 0 ] || return 64
  plan_tmp_dir="$GOTMPDIR/package-plans"
  plan_bin_dir="$GOTMPDIR/bin"
  mkdir -p -- "$plan_tmp_dir" "$plan_bin_dir" || return 75
  [ -d "$plan_tmp_dir" ] && [ ! -L "$plan_tmp_dir" ] || return 75
  [ -d "$plan_bin_dir" ] && [ ! -L "$plan_bin_dir" ] || return 75
  return 0
}

_plan_selftest_discovery_contract() {
  [ "$#" -eq 0 ] || return 64
  [ "$PACKAGE_DISCOVERY_FORMAT" = '{{printf "%s\t%s" .ImportPath .Name}}' ] || return 75
  template_tail=${PACKAGE_DISCOVERY_FORMAT#*'\t'}
  [ "$template_tail" != "$PACKAGE_DISCOVERY_FORMAT" ] || return 75
  case "$template_tail" in
    *'\t'*) return 75 ;;
    *) ;;
  esac
  delimiter="$(printf '\011')"
  delimiter_hex="$(printf '%s' "$delimiter" | od -An -v -tx1 | tr -d ' \n')"
  [ "$delimiter_hex" = 09 ] || return 75
  probe="example.invalid/package${delimiter}package"
  case "$probe" in
    *'\t'*) return 75 ;;
    *"$delimiter"*) ;;
    *) return 75 ;;
  esac
  _plan_import_path_is_valid 'example.invalid/module-v1/pkg_2~x+test' || return 75
  _plan_import_path_is_valid '/x' && return 75
  _plan_import_path_is_valid '../x' && return 75
  _plan_import_path_is_valid 'a//b' && return 75
  _plan_import_path_is_valid 'a/./b' && return 75
  _plan_import_path_is_valid 'a/../b' && return 75
  _plan_import_path_is_valid 'a/' && return 75
  [ "$(_plan_discovery_argv_json build)" = '["go","list","-f","{{printf \"%s\\t%s\" .ImportPath .Name}}","./..."]' ] || return 75
  [ "$(_plan_discovery_argv_json vet)" = '["go","list","-tags","hubtest","-f","{{printf \"%s\\t%s\" .ImportPath .Name}}","./..."]' ] || return 75
  [ "$(_plan_discovery_argv_json unit)" = '["go","list","-tags","hubtest","-f","{{printf \"%s\\t%s\" .ImportPath .Name}}","./tests/..."]' ] || return 75
  [ "$(_plan_discovery_argv_json race)" = '["go","list","-tags","hubtest","-f","{{printf \"%s\\t%s\" .ImportPath .Name}}","./tests/..."]' ] || return 75
  return 0
}

_plan_create_from_command() {
  [ "$#" -eq 2 ] || return 64
  plan_destination="$1"
  plan_selector="$2"
  [ ! -e "$plan_destination" ] && [ ! -L "$plan_destination" ] || return 75
  case "$plan_selector" in
    build)
      ( umask 077; set -C; go list -f '{{printf "%s\t%s" .ImportPath .Name}}' ./... > "$plan_destination" ) || return 75
      ;;
    vet)
      ( umask 077; set -C; go list -tags hubtest -f '{{printf "%s\t%s" .ImportPath .Name}}' ./... > "$plan_destination" ) || return 75
      ;;
    tests)
      ( umask 077; set -C; go list -tags hubtest -f '{{printf "%s\t%s" .ImportPath .Name}}' ./tests/... > "$plan_destination" ) || return 75
      ;;
    *) return 64 ;;
  esac
  [ -f "$plan_destination" ] && [ ! -L "$plan_destination" ] || return 75
  return 0
}

_plan_sort_rows() {
  [ "$#" -eq 2 ] || return 64
  plan_source="$1"
  plan_destination="$2"
  [ -f "$plan_source" ] && [ ! -L "$plan_source" ] || return 75
  [ ! -e "$plan_destination" ] && [ ! -L "$plan_destination" ] || return 75
  ( umask 077; set -C; sort -- "$plan_source" > "$plan_destination" ) || return 75
  [ -f "$plan_destination" ] && [ ! -L "$plan_destination" ] || return 75
  return 0
}

_plan_validate_rows() {
  [ "$#" -eq 1 ] || return 64
  plan_rows_path="$1"
  [ -f "$plan_rows_path" ] && [ ! -L "$plan_rows_path" ] || return 75
  [ -s "$plan_rows_path" ] || return 75
  plan_last_byte="$(tail -c 1 -- "$plan_rows_path" | od -An -v -tx1 | tr -d ' \n')"
  [ "$plan_last_byte" = 0a ] || return 75

  plan_tab="$(printf '\011')"
  plan_previous_import_path=
  plan_row_count=0
  while IFS= read -r plan_row; do
    case "$plan_row" in
      *'\t'*) return 75 ;;
      *"$plan_tab"*) ;;
      *) return 75 ;;
    esac
    plan_import_path=${plan_row%%"$plan_tab"*}
    plan_package_name=${plan_row#*"$plan_tab"}
    case "$plan_package_name" in
      *"$plan_tab"*) return 75 ;;
      *) ;;
    esac
    _plan_import_path_is_valid "$plan_import_path" || return 75
    _plan_package_name_is_valid "$plan_package_name" || return 75
    if [ -n "$plan_previous_import_path" ] && [ "$plan_import_path" \> "$plan_previous_import_path" ]; then
      :
    elif [ -n "$plan_previous_import_path" ]; then
      return 75
    fi
    plan_previous_import_path="$plan_import_path"
    plan_row_count=$((plan_row_count + 1))
  done < "$plan_rows_path"
  [ "$plan_row_count" -gt 0 ] || return 75
  plan_row_set_sha256="sha256:$(sha256sum -- "$plan_rows_path" | awk '{print $1}')"
  printf '%s\n' "$plan_row_set_sha256" | grep -Eq '^sha256:[0-9a-f]{64}$' || return 75
  return 0
}

_plan_require_subset() {
  [ "$#" -eq 2 ] || return 64
  plan_subset_path="$1"
  plan_superset_path="$2"
  [ -f "$plan_subset_path" ] && [ ! -L "$plan_subset_path" ] || return 75
  [ -f "$plan_superset_path" ] && [ ! -L "$plan_superset_path" ] || return 75
  while IFS= read -r plan_subset_row; do
    grep -F -x -- "$plan_subset_row" "$plan_superset_path" > /dev/null || return 75
  done < "$plan_subset_path"
  return 0
}

_plan_discovery_argv_json() {
  [ "$#" -eq 1 ] || return 64
  case "$1" in
    build)
      printf '%s' '["go","list","-f","{{printf \"%s\\t%s\" .ImportPath .Name}}","./..."]'
      ;;
    vet)
      printf '%s' '["go","list","-tags","hubtest","-f","{{printf \"%s\\t%s\" .ImportPath .Name}}","./..."]'
      ;;
    unit|race)
      printf '%s' '["go","list","-tags","hubtest","-f","{{printf \"%s\\t%s\" .ImportPath .Name}}","./tests/..."]'
      ;;
    *) return 64 ;;
  esac
}

_plan_write_json() {
  [ "$#" -eq 2 ] || return 64
  plan_stage="$1"
  plan_rows_path="$2"
  _plan_stage_is_valid "$plan_stage" || return 64
  _plan_validate_rows "$plan_rows_path" || return 75
  plan_json_path="$run_dir/package-plan.$plan_stage.json"
  [ ! -e "$plan_json_path" ] && [ ! -L "$plan_json_path" ] || return 75
  plan_discovery_argv="$(_plan_discovery_argv_json "$plan_stage")" || return 75
  plan_payload=$(
    printf '{"schema":"%s","target":"%s","run_id":"%s","owner_nonce":"%s","stage":"%s","discovery_argv":%s,"go_environment_sha256":"%s","rows":[' \
      "$PACKAGE_PLAN_SCHEMA" "$PACKAGE_PLAN_TARGET" "$run_id" "$owner_nonce" "$plan_stage" "$plan_discovery_argv" "$go_environment_sha256"
    plan_first=true
    plan_ordinal=0
    plan_tab="$(printf '\011')"
    while IFS= read -r plan_row; do
      plan_ordinal=$((plan_ordinal + 1))
      plan_import_path=${plan_row%%"$plan_tab"*}
      plan_package_name=${plan_row#*"$plan_tab"}
      if [ "$plan_first" = true ]; then
        plan_first=false
      else
        printf ','
      fi
      printf '{"ordinal":%s,"import_path":"%s","package_name":"%s"}' \
        "$plan_ordinal" "$plan_import_path" "$plan_package_name"
    done < "$plan_rows_path"
    printf '],"row_count":%s,"row_set_sha256":"%s","result_partition":[' \
      "$plan_row_count" "$plan_row_set_sha256"
    plan_first=true
    plan_ordinal=0
    while IFS= read -r plan_row; do
      plan_ordinal=$((plan_ordinal + 1))
      plan_import_path=${plan_row%%"$plan_tab"*}
      if [ "$plan_first" = true ]; then
        plan_first=false
      else
        printf ','
      fi
      printf '{"ordinal":%s,"import_path":"%s","result":"%s"}' \
        "$plan_ordinal" "$plan_import_path" "$PACKAGE_PLAN_RESULT_PENDING"
    done < "$plan_rows_path"
    printf ']}'
  ) || return 75
  write_atomic "$plan_json_path" "$plan_payload" || return 75
  check_budget || return 75
  return 0
}

package_plan_rows_path() {
  [ "$#" -eq 1 ] || return 64
  _plan_stage_is_valid "$1" || return 64
  printf '%s/%s.rows\n' "$GOTMPDIR/package-plans" "$1"
}

prepare_package_plans() {
  [ "$#" -eq 0 ] || return 64
  _plan_require_context || return 75
  assert_owned_tmpfs || return 75
  check_budget || return 75
  _plan_selftest_discovery_contract || return 75
  _plan_prepare_owned_directories || return 75

  LC_ALL=C
  export LC_ALL
  _plan_create_from_command "$plan_tmp_dir/discovery.build.raw" build || return 75
  _plan_create_from_command "$plan_tmp_dir/discovery.vet.raw" vet || return 75
  _plan_create_from_command "$plan_tmp_dir/discovery.tests.raw" tests || return 75
  _plan_sort_rows "$plan_tmp_dir/discovery.build.raw" "$plan_tmp_dir/build.rows" || return 75
  _plan_sort_rows "$plan_tmp_dir/discovery.vet.raw" "$plan_tmp_dir/vet.rows" || return 75
  _plan_sort_rows "$plan_tmp_dir/discovery.tests.raw" "$plan_tmp_dir/unit.rows" || return 75
  [ ! -e "$plan_tmp_dir/race.rows" ] && [ ! -L "$plan_tmp_dir/race.rows" ] || return 75
  ( umask 077; set -C; cp -- "$plan_tmp_dir/unit.rows" "$plan_tmp_dir/race.rows" ) || return 75
  cmp -s -- "$plan_tmp_dir/unit.rows" "$plan_tmp_dir/race.rows" || return 75

  _plan_validate_rows "$plan_tmp_dir/build.rows" || return 75
  _plan_validate_rows "$plan_tmp_dir/vet.rows" || return 75
  _plan_validate_rows "$plan_tmp_dir/unit.rows" || return 75
  _plan_validate_rows "$plan_tmp_dir/race.rows" || return 75
  _plan_require_subset "$plan_tmp_dir/build.rows" "$plan_tmp_dir/vet.rows" || return 75
  _plan_require_subset "$plan_tmp_dir/unit.rows" "$plan_tmp_dir/vet.rows" || return 75

  _plan_write_json build "$plan_tmp_dir/build.rows" || return 75
  _plan_write_json vet "$plan_tmp_dir/vet.rows" || return 75
  _plan_write_json unit "$plan_tmp_dir/unit.rows" || return 75
  _plan_write_json race "$plan_tmp_dir/race.rows" || return 75
  check_budget || return 75
  return 0
}

prepare_owned_output() {
  [ "$#" -eq 2 ] || return 64
  package_ordinal="$1"
  package_import_path="$2"
  _plan_prepare_owned_directories || return 75
  case "$package_ordinal" in
    ''|*[!0-9]*) return 75 ;;
    *) ;;
  esac
  [ "$package_ordinal" -ge 1 ] && [ "$package_ordinal" -le 999999 ] || return 75
  _plan_import_path_is_valid "$package_import_path" || return 75
  package_import_sha256="$(printf '%s' "$package_import_path" | sha256sum | awk '{print $1}')"
  printf '%s\n' "$package_import_sha256" | grep -Eq '^[0-9a-f]{64}$' || return 75
  package_ordinal_padded="$(printf '%06d' "$package_ordinal")" || return 75
  owned_output="$plan_bin_dir/$package_ordinal_padded-$package_import_sha256"
  case "$owned_output" in
    "$GOTMPDIR/bin/"[0-9][0-9][0-9][0-9][0-9][0-9]-[0-9a-f][0-9a-f]*) ;;
    *) return 75 ;;
  esac
  [ ! -e "$owned_output" ] && [ ! -L "$owned_output" ] || return 75
  return 0
}

unlink_owned_output() {
  [ "$#" -eq 0 ] || return 64
  [ -n "${owned_output-}" ] || return 75
  case "$owned_output" in
    "$GOTMPDIR/bin/"[0-9][0-9][0-9][0-9][0-9][0-9]-[0-9a-f][0-9a-f]*) ;;
    *) return 75 ;;
  esac
  if [ -e "$owned_output" ] || [ -L "$owned_output" ]; then
    [ -f "$owned_output" ] && [ ! -L "$owned_output" ] || return 75
    [ "$(stat -c '%h' -- "$owned_output")" = 1 ] || return 75
    rm -- "$owned_output" || return 75
  fi
  owned_output=
  return 0
}
