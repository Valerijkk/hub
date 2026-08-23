#!/bin/sh
set -eu

readonly EX_USAGE=64
readonly EX_CONFIG=78
readonly REVIEW_FILE_LIMIT_BYTES=33554432
readonly SOURCE_ROOT=/opt/hub/backend
readonly LINEAGE_DIR="$SOURCE_ROOT/migrations/atlas/mail-calendar-v1"
readonly WORK_ROOT=/work/mail-calendar-schema
readonly BASELINE_FILENAME=20260823000000_mail_calendar_baseline.sql
readonly EXPECTED_DATABASE_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema?sslmode=disable&search_path=public'
readonly EXPECTED_DEV_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema?sslmode=disable&search_path=atlas_dev'

fail() {
  printf '%s\n' "$1" >&2
  exit "$2"
}

require_internal_database_urls() {
  [ "${DATABASE_URL:-}" = "$EXPECTED_DATABASE_URL" ] || fail 'schema-maintenance: DATABASE_URL must be the fixed internal disposable target' "$EX_CONFIG"
  [ "${ATLAS_DEV_URL:-}" = "$EXPECTED_DEV_URL" ] || fail 'schema-maintenance: ATLAS_DEV_URL must be the fixed internal disposable dev schema' "$EX_CONFIG"
}

require_exact_files() {
  expected_count="$1"
  shift

  [ -d "$LINEAGE_DIR" ] || fail 'schema-maintenance: reviewed Mail/Calendar lineage is absent' "$EX_CONFIG"
  non_regular_entry="$(find "$LINEAGE_DIR" -mindepth 1 -maxdepth 1 ! -type f -print -quit)"
  [ -z "$non_regular_entry" ] || fail 'schema-maintenance: lineage inputs contain a non-regular entry' "$EX_CONFIG"
  actual_count="$(find "$LINEAGE_DIR" -mindepth 1 -maxdepth 1 -type f | wc -l | tr -d ' ')"
  [ "$actual_count" -eq "$expected_count" ] || fail 'schema-maintenance: lineage inputs are incomplete or contain unreviewed files' "$EX_CONFIG"

  for required_name in "$@"; do
    [ -f "$LINEAGE_DIR/$required_name" ] || fail 'schema-maintenance: required reviewed lineage input is absent' "$EX_CONFIG"
  done
}

require_exact_candidate_files() {
  candidate_dir="$1"
  expected_count="$2"
  shift 2

  [ -d "$candidate_dir" ] || fail 'schema-maintenance: candidate directory is absent' "$EX_CONFIG"
  if find "$candidate_dir" -mindepth 1 -maxdepth 1 ! -type f -print -quit | grep -q .; then
    fail 'schema-maintenance: candidate contains a non-regular output entry' "$EX_CONFIG"
  fi
  if ! find "$candidate_dir" -mindepth 1 -maxdepth 1 -type f | wc -l | tr -d ' ' | grep -qx "$expected_count"; then
    fail 'schema-maintenance: candidate contains unreviewed output files' "$EX_CONFIG"
  fi

  for required_name in "$@"; do
    [ -f "$candidate_dir/$required_name" ] || fail 'schema-maintenance: required candidate output is absent' "$EX_CONFIG"
  done
}

emit_review_file() {
  review_file="$1"
  review_size="$(wc -c < "$review_file" | tr -d ' ')"
  [ "$review_size" -le "$REVIEW_FILE_LIMIT_BYTES" ] || fail 'schema-maintenance: candidate review output exceeds its fixed bound' "$EX_CONFIG"

  review_hash="$(sha256sum "$review_file" | awk '{print $1}')"
  printf '%s\n' "--- BEGIN REVIEW FILE $(basename "$review_file") sha256=$review_hash bytes=$review_size ---"
  cat "$review_file"
  printf '%s\n' "--- END REVIEW FILE $(basename "$review_file") ---"
}

run_baseline() {
  require_internal_database_urls

  candidate_dir="$WORK_ROOT/baseline"
  rm -rf "$candidate_dir"
  mkdir -p "$candidate_dir"
  trap 'rm -rf "$WORK_ROOT"' EXIT HUP INT TERM

  atlas schema inspect \
    --config "$SOURCE_ROOT/atlas.hcl" \
    --env prefeature_baseline \
    --url env://schema.src \
    --format '{{ sql . }}' > "$candidate_dir/$BASELINE_FILENAME"
  atlas migrate hash --dir "file://$candidate_dir"

  require_exact_candidate_files "$candidate_dir" 2 "$BASELINE_FILENAME" atlas.sum
  emit_review_file "$candidate_dir/$BASELINE_FILENAME"
  emit_review_file "$candidate_dir/atlas.sum"
}

run_diff() {
  require_internal_database_urls
  require_exact_files 2 20260823000000_mail_calendar_baseline.sql atlas.sum

  candidate_dir="$WORK_ROOT/diff"
  rm -rf "$candidate_dir"
  mkdir -p "$candidate_dir"
  trap 'rm -rf "$WORK_ROOT"' EXIT HUP INT TERM
  cp "$LINEAGE_DIR/20260823000000_mail_calendar_baseline.sql" "$candidate_dir/"
  cp "$LINEAGE_DIR/atlas.sum" "$candidate_dir/"

  atlas migrate diff mail_calendar_core \
    --config "$SOURCE_ROOT/atlas.hcl" \
    --env local \
    --dir "file://$candidate_dir"
  atlas migrate hash --dir "file://$candidate_dir"

  non_regular_entry="$(find "$candidate_dir" -mindepth 1 -maxdepth 1 ! -type f -print -quit)"
  [ -z "$non_regular_entry" ] || fail 'schema-maintenance: Atlas candidate contains a non-regular output entry' "$EX_CONFIG"
  require_candidate_files="$(find "$candidate_dir" -mindepth 1 -maxdepth 1 -type f -name '*_mail_calendar_core.sql' | wc -l | tr -d ' ')"
  [ "$require_candidate_files" -eq 1 ] || fail 'schema-maintenance: Atlas did not produce exactly one bounded Mail/Calendar candidate' "$EX_CONFIG"
  candidate_sql="$(find "$candidate_dir" -mindepth 1 -maxdepth 1 -type f -name '*_mail_calendar_core.sql')"
  [ "$(find "$candidate_dir" -mindepth 1 -maxdepth 1 -type f | wc -l | tr -d ' ')" -eq 3 ] || fail 'schema-maintenance: Atlas candidate contains unreviewed output files' "$EX_CONFIG"

  emit_review_file "$candidate_sql"
  emit_review_file "$candidate_dir/atlas.sum"
}

run_verify() {
  require_internal_database_urls
  require_exact_files 3 20260823000000_mail_calendar_baseline.sql 20260823000001_mail_calendar_core.sql atlas.sum

  atlas migrate validate --config "$SOURCE_ROOT/atlas.hcl" --env local
  atlas migrate lint --config "$SOURCE_ROOT/atlas.hcl" --env local --latest 2
}

[ "$#" -eq 1 ] || fail 'usage: mail-calendar-schema {baseline|diff|verify}' "$EX_USAGE"

case "$1" in
  baseline)
    run_baseline
    ;;
  diff)
    run_diff
    ;;
  verify)
    run_verify
    ;;
  *)
    fail 'usage: mail-calendar-schema {baseline|diff|verify}' "$EX_USAGE"
    ;;
esac
