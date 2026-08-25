#!/bin/sh
set -eu

readonly EX_USAGE=64
readonly EX_CONFIG=78
readonly REVIEW_FILE_LIMIT_BYTES=33554432
readonly SOURCE_ROOT=/opt/hub/backend
readonly LINEAGE_DIR="$SOURCE_ROOT/migrations/atlas/mail-calendar-v1"
readonly WORK_ROOT=/work/mail-calendar-schema
readonly GENERATE_FIRST_ROOT="$WORK_ROOT/generate-first"
readonly GENERATE_SECOND_ROOT="$WORK_ROOT/generate-second"
readonly GENERATE_FIRST_MANIFEST="$WORK_ROOT/generate-first.manifest"
readonly GENERATE_SECOND_MANIFEST="$WORK_ROOT/generate-second.manifest"
readonly GENERATE_FIRST_SOURCE_BEFORE_MANIFEST="$WORK_ROOT/generate-first.sources.before.manifest"
readonly GENERATE_FIRST_SOURCE_AFTER_MANIFEST="$WORK_ROOT/generate-first.sources.after.manifest"
readonly GENERATE_SECOND_SOURCE_BEFORE_MANIFEST="$WORK_ROOT/generate-second.sources.before.manifest"
readonly GENERATE_SECOND_SOURCE_AFTER_MANIFEST="$WORK_ROOT/generate-second.sources.after.manifest"
readonly GO_RUN_EXEC_ROOT=/run/go-run-exec
readonly GENERATE_FIRST_EXEC_DIR="$GO_RUN_EXEC_ROOT/generate-first"
readonly GENERATE_SECOND_EXEC_DIR="$GO_RUN_EXEC_ROOT/generate-second"
readonly BASELINE_FILENAME=20260823000000_mail_calendar_baseline.sql
readonly EXPECTED_DATABASE_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema?sslmode=disable&search_path=public'
readonly EXPECTED_DEV_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema?sslmode=disable&search_path=atlas_dev'
readonly ATLAS_MODULE='ariga.io/atlas'
readonly ATLAS_VERSION='v0.19.1-0.20240203083654-5948b60a8e43'
readonly ATLAS_MODULE_SUM='h1:GwdJbXydHCYPedeeLt4x/lrlIISQ4JTH1mRWuE5ZZ14='
readonly ATLAS_MODULE_GO_MOD_SUM='h1:uj3pm+hUTVN/X5yfdBexHlZv+1Xu5u5ZbZx7+CDavNU='
readonly ENT_MODULE='entgo.io/ent'
readonly ENT_VERSION='v0.14.1'
readonly ENT_MODULE_SUM='h1:fUERL506Pqr92EPHJqr8EYxbPioflJo6PudkrEA8a/s='
readonly ENT_MODULE_GO_MOD_SUM='h1:MH6XLG0KXpkcDQhKiHfANZSzR55TJyPL5IGNpI8wpco='
readonly ENT_GENERATE_DIRECTIVE_NORMALIZED_SHA256='8138e39380b362cbad50f0646805f2e950973b5460fb9c4e404f03f8834979a5'
readonly SOURCE_GO_MOD_NORMALIZED_SHA256='3f7ff699a420ecf0501693c47ee18108a02a8827a547578bb6b800f83bcc2be6'
readonly SOURCE_GO_MOD_NORMALIZED_BYTES=9763
readonly SOURCE_GO_SUM_NORMALIZED_SHA256='9a03c4db3f998434e3e9058a67b7d2143fda8c5a7ccee6895e9093d6e77d2510'
readonly SOURCE_GO_SUM_NORMALIZED_BYTES=55553

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

require_non_symlink() {
  [ ! -L "$1" ] || fail "schema-maintenance: generate source contains a symlink: $1" "$EX_CONFIG"
}

require_exact_go_dependency_pins() {
  lock_file="$SOURCE_ROOT/docker/schema-tool.lock.json"
  require_non_symlink "$lock_file"
  [ -f "$lock_file" ] || fail 'schema-maintenance: schema tool lock is absent' "$EX_CONFIG"

  if grep -Eq '^[[:space:]]*(replace|exclude)([[:space:]]|$)' "$SOURCE_ROOT/go.mod"; then
    fail 'schema-maintenance: generate source go.mod must not contain replace/exclude directives' "$EX_CONFIG"
  fi
  grep -Fqx "$ATLAS_MODULE $ATLAS_VERSION $ATLAS_MODULE_SUM" "$SOURCE_ROOT/go.sum" ||
    fail 'schema-maintenance: Atlas module checksum does not match the reviewed lock' "$EX_CONFIG"
  grep -Fqx "$ATLAS_MODULE $ATLAS_VERSION/go.mod $ATLAS_MODULE_GO_MOD_SUM" "$SOURCE_ROOT/go.sum" ||
    fail 'schema-maintenance: Atlas go.mod checksum does not match the reviewed lock' "$EX_CONFIG"
  grep -Fqx "$ENT_MODULE $ENT_VERSION $ENT_MODULE_SUM" "$SOURCE_ROOT/go.sum" ||
    fail 'schema-maintenance: Ent module checksum does not match the reviewed lock' "$EX_CONFIG"
  grep -Fqx "$ENT_MODULE $ENT_VERSION/go.mod $ENT_MODULE_GO_MOD_SUM" "$SOURCE_ROOT/go.sum" ||
    fail 'schema-maintenance: Ent go.mod checksum does not match the reviewed lock' "$EX_CONFIG"
  grep -Fqx '    "module": "ariga.io/atlas",' "$lock_file" || fail 'schema-maintenance: locked Atlas module is absent' "$EX_CONFIG"
  grep -Fqx '    "version": "v0.19.1-0.20240203083654-5948b60a8e43",' "$lock_file" || fail 'schema-maintenance: locked Atlas version is absent' "$EX_CONFIG"
  grep -Fqx '    "module_sum": "h1:GwdJbXydHCYPedeeLt4x/lrlIISQ4JTH1mRWuE5ZZ14=",' "$lock_file" || fail 'schema-maintenance: locked Atlas sum is absent' "$EX_CONFIG"
  grep -Fqx '    "module_go_mod_sum": "h1:uj3pm+hUTVN/X5yfdBexHlZv+1Xu5u5ZbZx7+CDavNU=",' "$lock_file" || fail 'schema-maintenance: locked Atlas go.mod sum is absent' "$EX_CONFIG"
  grep -Fqx '    "module": "entgo.io/ent",' "$lock_file" || fail 'schema-maintenance: locked Ent module is absent' "$EX_CONFIG"
  grep -Fqx '    "version": "v0.14.1",' "$lock_file" || fail 'schema-maintenance: locked Ent version is absent' "$EX_CONFIG"
  grep -Fqx '    "module_sum": "h1:fUERL506Pqr92EPHJqr8EYxbPioflJo6PudkrEA8a/s=",' "$lock_file" || fail 'schema-maintenance: locked Ent sum is absent' "$EX_CONFIG"
  grep -Fqx '    "module_go_mod_sum": "h1:MH6XLG0KXpkcDQhKiHfANZSzR55TJyPL5IGNpI8wpco=",' "$lock_file" || fail 'schema-maintenance: locked Ent go.mod sum is absent' "$EX_CONFIG"

  grep -Fqx '  "source_module_lock": {' "$lock_file" ||
    fail 'schema-maintenance: source module lock object is absent' "$EX_CONFIG"
  grep -Fqx '    "go_mod_normalized_sha256": "3f7ff699a420ecf0501693c47ee18108a02a8827a547578bb6b800f83bcc2be6",' "$lock_file" ||
    fail 'schema-maintenance: locked go.mod hash is absent' "$EX_CONFIG"
  grep -Fqx '    "go_mod_normalized_bytes": 9763,' "$lock_file" ||
    fail 'schema-maintenance: locked go.mod byte count is absent' "$EX_CONFIG"
  grep -Fqx '    "go_sum_normalized_sha256": "9a03c4db3f998434e3e9058a67b7d2143fda8c5a7ccee6895e9093d6e77d2510",' "$lock_file" ||
    fail 'schema-maintenance: locked go.sum hash is absent' "$EX_CONFIG"
  grep -Fqx '    "go_sum_normalized_bytes": 55553' "$lock_file" ||
    fail 'schema-maintenance: locked go.sum byte count is absent' "$EX_CONFIG"

  actual_hash="$(tr -d '\r' < "$SOURCE_ROOT/go.mod" | sha256sum | cut -d ' ' -f1)"
  actual_bytes="$(tr -d '\r' < "$SOURCE_ROOT/go.mod" | wc -c | tr -d '[:space:]')"
  [ "$actual_hash" = "$SOURCE_GO_MOD_NORMALIZED_SHA256" ] || fail 'schema-maintenance: normalized go.mod hash does not match the reviewed lock' "$EX_CONFIG"
  [ "$actual_bytes" = "$SOURCE_GO_MOD_NORMALIZED_BYTES" ] || fail 'schema-maintenance: normalized go.mod byte count does not match the reviewed lock' "$EX_CONFIG"
  actual_hash="$(tr -d '\r' < "$SOURCE_ROOT/go.sum" | sha256sum | cut -d ' ' -f1)"
  actual_bytes="$(tr -d '\r' < "$SOURCE_ROOT/go.sum" | wc -c | tr -d '[:space:]')"
  [ "$actual_hash" = "$SOURCE_GO_SUM_NORMALIZED_SHA256" ] || fail 'schema-maintenance: normalized go.sum hash does not match the reviewed lock' "$EX_CONFIG"
  [ "$actual_bytes" = "$SOURCE_GO_SUM_NORMALIZED_BYTES" ] || fail 'schema-maintenance: normalized go.sum byte count does not match the reviewed lock' "$EX_CONFIG"
}

require_ent_generate_directive_alignment() {
  lock_file="$SOURCE_ROOT/docker/schema-tool.lock.json"
  grep -Fqx '    "generate_go_normalized_sha256": "8138e39380b362cbad50f0646805f2e950973b5460fb9c4e404f03f8834979a5",' "$lock_file" ||
    fail 'schema-maintenance: schema tool lock does not carry the reviewed Ent directive hash' "$EX_CONFIG"
  grep -Fqx '    "generate_command": "go run -mod=readonly entgo.io/ent/cmd/ent generate ./ent/schema",' "$lock_file" ||
    fail 'schema-maintenance: schema tool lock does not carry the fixed Ent command' "$EX_CONFIG"
  directive_hash="$(tr -d '\r' < "$SOURCE_ROOT/ent/generate.go" | sha256sum | cut -d ' ' -f1)"
  [ "$directive_hash" = "$ENT_GENERATE_DIRECTIVE_NORMALIZED_SHA256" ] ||
    fail 'schema-maintenance: Ent generate directive does not match the reviewed lock' "$EX_CONFIG"
}

require_generate_source_files() {
  for source_path in \
    "$SOURCE_ROOT" \
    "$SOURCE_ROOT/go.mod" \
    "$SOURCE_ROOT/go.sum" \
    "$SOURCE_ROOT/ent" \
    "$SOURCE_ROOT/ent/generate.go" \
    "$SOURCE_ROOT/ent/schema"; do
    require_non_symlink "$source_path"
  done
  [ -f "$SOURCE_ROOT/go.mod" ] || fail 'schema-maintenance: generate source go.mod is absent' "$EX_CONFIG"
  [ -f "$SOURCE_ROOT/go.sum" ] || fail 'schema-maintenance: generate source go.sum is absent' "$EX_CONFIG"
  [ -f "$SOURCE_ROOT/ent/generate.go" ] || fail 'schema-maintenance: generate source directive is absent' "$EX_CONFIG"
  [ -d "$SOURCE_ROOT/ent/schema" ] || fail 'schema-maintenance: generate source schema is absent' "$EX_CONFIG"
  if find "$SOURCE_ROOT/ent/schema" -mindepth 1 \( -type l -o \( ! -type f -a ! -type d \) \) -print -quit | grep -q .; then
    fail 'schema-maintenance: generate source schema contains a symlink or non-regular entry' "$EX_CONFIG"
  fi
  if ! find "$SOURCE_ROOT/ent/schema" -mindepth 1 -type f -print -quit | grep -q .; then
    fail 'schema-maintenance: generate source schema is empty' "$EX_CONFIG"
  fi
  require_exact_go_dependency_pins
  require_ent_generate_directive_alignment
}

prepare_generate_module_root() {
  module_root="$1"
  rm -rf "$module_root"
  mkdir -p "$module_root/ent"
  cp "$SOURCE_ROOT/go.mod" "$module_root/go.mod"
  cp "$SOURCE_ROOT/go.sum" "$module_root/go.sum"
  cp "$SOURCE_ROOT/ent/generate.go" "$module_root/ent/generate.go"
  cp -R "$SOURCE_ROOT/ent/schema" "$module_root/ent/schema"
}

require_generate_module_source_files() {
  module_root="$1"
  for source_path in \
    "$module_root" \
    "$module_root/go.mod" \
    "$module_root/go.sum" \
    "$module_root/ent" \
    "$module_root/ent/generate.go" \
    "$module_root/ent/schema"; do
    require_non_symlink "$source_path"
  done
  [ -f "$module_root/go.mod" ] || fail 'schema-maintenance: isolated generate root go.mod is absent' "$EX_CONFIG"
  [ -f "$module_root/go.sum" ] || fail 'schema-maintenance: isolated generate root go.sum is absent' "$EX_CONFIG"
  [ -f "$module_root/ent/generate.go" ] || fail 'schema-maintenance: isolated generate root directive is absent' "$EX_CONFIG"
  [ -d "$module_root/ent/schema" ] || fail 'schema-maintenance: isolated generate root schema is absent' "$EX_CONFIG"
  if find "$module_root/ent/schema" -mindepth 1 \( -type l -o \( ! -type f -a ! -type d \) \) -print -quit | grep -q .; then
    fail 'schema-maintenance: isolated generate root schema contains a symlink or non-regular entry' "$EX_CONFIG"
  fi
}

append_generate_source_manifest_entry() {
  module_root="$1"
  source_file="$2"
  manifest="$3"
  case "$source_file" in
    "$module_root"/*) relative_path="${source_file#"$module_root"/}" ;;
    *) fail 'schema-maintenance: source-file path escaped its module root' "$EX_CONFIG" ;;
  esac
  case "$relative_path" in
    ""|/*|*..*|-*|*[!abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._/-]*) fail 'schema-maintenance: source-file path is unsafe' "$EX_CONFIG" ;;
  esac
  [ -f "$source_file" ] || fail 'schema-maintenance: source manifest contains a non-regular path' "$EX_CONFIG"
  source_bytes="$(wc -c < "$source_file" | tr -d '[:space:]')"
  source_sha256="$(sha256sum "$source_file" | cut -d ' ' -f1)"
  printf '%s %s %s\n' "$relative_path" "$source_bytes" "$source_sha256" >> "$manifest"
}

write_generate_source_manifest() {
  module_root="$1"
  manifest="$2"
  source_list="$manifest.list"
  require_generate_module_source_files "$module_root"
  : > "$manifest"
  append_generate_source_manifest_entry "$module_root" "$module_root/go.mod" "$manifest"
  append_generate_source_manifest_entry "$module_root" "$module_root/go.sum" "$manifest"
  append_generate_source_manifest_entry "$module_root" "$module_root/ent/generate.go" "$manifest"
  LC_ALL=C find "$module_root/ent/schema" -type f -print | LC_ALL=C sort > "$source_list"
  while IFS= read -r source_file; do
    [ -n "$source_file" ] || continue
    append_generate_source_manifest_entry "$module_root" "$source_file" "$manifest"
  done < "$source_list"
  rm -f "$source_list"
  [ -s "$manifest" ] || fail 'schema-maintenance: generate source manifest is empty' "$EX_CONFIG"
  manifest_size="$(wc -c < "$manifest" | tr -d '[:space:]')"
  [ "$manifest_size" -le "$REVIEW_FILE_LIMIT_BYTES" ] || fail 'schema-maintenance: generate source manifest exceeds its fixed bound' "$EX_CONFIG"
}

run_generate_module_root() {
  module_root="$1"
  source_before_manifest="$2"
  source_after_manifest="$3"
  module_exec_tmp="$4"
  module_cache="$module_root/.gocache"
  module_home="$module_root/.home"
  module_xdg_cache="$module_root/.xdg-cache"
  mkdir -p "$module_cache" "$module_home" "$module_xdg_cache"
  case "$module_exec_tmp" in
    "$GENERATE_FIRST_EXEC_DIR"|"$GENERATE_SECOND_EXEC_DIR") ;;
    *) fail 'schema-maintenance: unexpected Go execution tmpfs path' "$EX_CONFIG" ;;
  esac
  require_non_symlink "$GO_RUN_EXEC_ROOT"
  [ -d "$GO_RUN_EXEC_ROOT" ] || fail 'schema-maintenance: Go execution tmpfs is absent' "$EX_CONFIG"
  rm -rf "$module_exec_tmp"
  mkdir -p "$module_exec_tmp"
  require_non_symlink "$module_exec_tmp"
  write_generate_source_manifest "$module_root" "$source_before_manifest"
  if ! ent_version="$(
    cd "$module_root"
    HOME="$module_home" XDG_CACHE_HOME="$module_xdg_cache" TMPDIR="$module_exec_tmp" GOTOOLCHAIN=local GOWORK=off GOPROXY=off GOSUMDB=off GOPATH=/opt/hub/go GOMODCACHE=/opt/hub/go/pkg/mod GOFLAGS=-mod=readonly GOCACHE="$module_cache" GOTMPDIR="$module_exec_tmp" go list -m -f '{{.Version}}' entgo.io/ent
  )"; then
    fail 'schema-maintenance: pinned Ent module is unavailable from the offline read-only cache' "$EX_CONFIG"
  fi
  [ "$ent_version" = "$ENT_VERSION" ] || fail 'schema-maintenance: Ent module version does not match the reviewed v0.14.1 pin' "$EX_CONFIG"
  if ! (
    cd "$module_root"
    HOME="$module_home" XDG_CACHE_HOME="$module_xdg_cache" TMPDIR="$module_exec_tmp" GOTOOLCHAIN=local GOWORK=off GOPROXY=off GOSUMDB=off GOPATH=/opt/hub/go GOMODCACHE=/opt/hub/go/pkg/mod GOFLAGS=-mod=readonly GOCACHE="$module_cache" GOTMPDIR="$module_exec_tmp" go run -mod=readonly entgo.io/ent/cmd/ent generate ./ent/schema
  ) >/dev/null 2>&1; then
    fail 'schema-maintenance: Ent generation failed inside its isolated offline module root' "$EX_CONFIG"
  fi
  write_generate_source_manifest "$module_root" "$source_after_manifest"
  cmp -s "$source_before_manifest" "$source_after_manifest" || fail 'schema-maintenance: Ent generation modified a declared source input' "$EX_CONFIG"
}

write_generate_manifest() {
  module_root="$1"
  manifest="$2"
  ent_root="$module_root/ent"
  manifest_list="$module_root/generated-files.list"

  if find "$ent_root" -mindepth 1 ! -type f ! -type d -print -quit | grep -q .; then
    fail 'schema-maintenance: Ent generation produced a non-regular entry' "$EX_CONFIG"
  fi
  LC_ALL=C find "$ent_root" -type f ! -path "$ent_root/generate.go" ! -path "$ent_root/schema/*" -print | LC_ALL=C sort > "$manifest_list"
  : > "$manifest"
  while IFS= read -r generated_file; do
    [ -n "$generated_file" ] || continue
    case "$generated_file" in
      "$ent_root"/*) relative_path="$(echo "$generated_file" | sed "s|^$ent_root/||")" ;;
      *) fail 'schema-maintenance: generated-file path escaped its module root' "$EX_CONFIG" ;;
    esac
    case "$relative_path" in
      ""|/*|*..*|-*|*[!abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._/-]*) fail 'schema-maintenance: generated-file path is unsafe' "$EX_CONFIG" ;;
    esac
    [ -f "$generated_file" ] || fail 'schema-maintenance: generated-file manifest contains a non-regular path' "$EX_CONFIG"
    generated_bytes="$(wc -c < "$generated_file" | tr -d ' ')"
    generated_sha256="$(sha256sum "$generated_file" | cut -d ' ' -f1)"
    echo "$relative_path $generated_bytes $generated_sha256" >> "$manifest"
  done < "$manifest_list"
  rm -f "$manifest_list"
  [ -s "$manifest" ] || fail 'schema-maintenance: Ent generation produced no regular generated files' "$EX_CONFIG"
  manifest_size="$(wc -c < "$manifest" | tr -d ' ')"
  [ "$manifest_size" -le "$REVIEW_FILE_LIMIT_BYTES" ] || fail 'schema-maintenance: generated-file manifest exceeds its fixed bound' "$EX_CONFIG"
}

run_generate() {
  require_generate_source_files
  trap 'rm -rf "$WORK_ROOT" "$GENERATE_FIRST_EXEC_DIR" "$GENERATE_SECOND_EXEC_DIR"' EXIT HUP INT TERM
  rm -rf "$WORK_ROOT"
  mkdir -p "$WORK_ROOT"
  umask 077

  prepare_generate_module_root "$GENERATE_FIRST_ROOT"
  prepare_generate_module_root "$GENERATE_SECOND_ROOT"
  run_generate_module_root "$GENERATE_FIRST_ROOT" "$GENERATE_FIRST_SOURCE_BEFORE_MANIFEST" "$GENERATE_FIRST_SOURCE_AFTER_MANIFEST" "$GENERATE_FIRST_EXEC_DIR"
  run_generate_module_root "$GENERATE_SECOND_ROOT" "$GENERATE_SECOND_SOURCE_BEFORE_MANIFEST" "$GENERATE_SECOND_SOURCE_AFTER_MANIFEST" "$GENERATE_SECOND_EXEC_DIR"
  write_generate_manifest "$GENERATE_FIRST_ROOT" "$GENERATE_FIRST_MANIFEST"
  write_generate_manifest "$GENERATE_SECOND_ROOT" "$GENERATE_SECOND_MANIFEST"
  cmp -s "$GENERATE_FIRST_MANIFEST" "$GENERATE_SECOND_MANIFEST" || fail 'schema-maintenance: independent Ent generation manifests differ' "$EX_CONFIG"
  emit_review_file "$GENERATE_FIRST_MANIFEST"
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

[ "$#" -eq 1 ] || fail 'usage: mail-calendar-schema {baseline|diff|generate|verify}' "$EX_USAGE"

case "$1" in
  baseline)
    run_baseline
    ;;
  diff)
    run_diff
    ;;
  generate)
    run_generate
    ;;
  verify)
    run_verify
    ;;
  *)
    fail 'usage: mail-calendar-schema {baseline|diff|generate|verify}' "$EX_USAGE"
    ;;
esac
