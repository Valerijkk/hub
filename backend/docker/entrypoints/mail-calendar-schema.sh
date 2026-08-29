#!/bin/sh
set -eu

readonly EX_USAGE=64
readonly EX_CONFIG=78
readonly REVIEW_FILE_LIMIT_BYTES=33554432
readonly SOURCE_ROOT=/opt/hub/backend
readonly SCHEMA_TOOL_LOCK="$SOURCE_ROOT/docker/schema-tool.lock.json"
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
readonly CORE_FILENAME=20260823000001_mail_calendar_core.sql
readonly DEFERRED_RECEIPT_EXPIRY_FILENAME=20260826000000_calendar_deferred_receipt_expiry.sql
readonly INBOUND_MESSAGE_RECEIPT_EVENT_MESSAGE_IDX_FILENAME=20260826000001_calendar_inbound_message_receipt_event_message_idx.sql
readonly PASSWORD_ENROLLMENT_FILENAME=20260827000000_mail_account_password_enrollment.sql
readonly DELIVERY_MATERIALIZATION_INVALID_FILENAME=20260827000001_calendar_delivery_materialization_invalid.sql
readonly INVITE_SEQUENCE_NON_NEGATIVE_FILENAME=20260827000002_calendar_invite_sequence_non_negative.sql
readonly ROUNDTRIP_RECONCILIATION_FILENAME=20260829000000_mail_calendar_schema_roundtrip_reconciliation.sql
readonly EXPECTED_LINEAGE_SQL_FILE_COUNT=8
readonly EXPECTED_LINEAGE_ENTRY_COUNT=9
readonly EXPECTED_LATEST_VERSION=20260829000000
readonly SNAPSHOT_FILENAME=20260829000000_current_snapshot.sql
readonly EXPECTED_DATABASE_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema?sslmode=disable&search_path=public'
readonly EXPECTED_PREFEATURE_SCHEMA_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema_prefeature?sslmode=disable&search_path=hub_mail_calendar'
readonly EXPECTED_CURRENT_SCHEMA_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema_current?sslmode=disable&search_path=hub_mail_calendar'
readonly EXPECTED_VERIFICATION_SCHEMA_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema_verification?sslmode=disable&search_path=hub_mail_calendar'
readonly EXPECTED_ATLAS_DIFF_DEV_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema_atlas_diff_dev?sslmode=disable&search_path=hub_mail_calendar'
readonly EXPECTED_ATLAS_VALIDATE_DEV_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema_atlas_validate_dev?sslmode=disable&search_path=hub_mail_calendar'
readonly EXPECTED_ATLAS_LINT_DEV_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema_atlas_lint_dev?sslmode=disable&search_path=hub_mail_calendar'
readonly EXPECTED_ATLAS_COMPARE_DEV_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema_atlas_compare_dev?sslmode=disable&search_path=hub_mail_calendar'
readonly EXPECTED_ATLAS_REVISIONS_URL='postgres://schema@schema-maintenance-postgres:5432/hub_schema_verification?sslmode=disable&search_path=atlas_revisions'
readonly MATERIALIZER_SOURCE="$SOURCE_ROOT/docker/schema-maintenance/materializer/main.go"
readonly MATERIALIZER_LOCK="$SOURCE_ROOT/docker/schema-materializer.lock.json"
readonly MATERIALIZER_BIN="$GO_RUN_EXEC_ROOT/schema-materializer"
readonly MATERIALIZER_READY_FILE="$WORK_ROOT/materializer.lock.ready"
readonly MATERIALIZER_SOURCE_MANIFEST="$WORK_ROOT/materializer.sources.manifest"
readonly MATERIALIZER_NORMALIZED_SHA256='e9158d05c2795216fbd7d60ffb179b46946bb38167585d80be669ecc3a3e7ae3'
MATERIALIZER_PID=''
readonly ATLAS_MODULE='ariga.io/atlas'
readonly ATLAS_MODULE_VERSION='v0.19.1-0.20240203083654-5948b60a8e43'
readonly ATLAS_MODULE_SUM='h1:GwdJbXydHCYPedeeLt4x/lrlIISQ4JTH1mRWuE5ZZ14='
readonly ATLAS_MODULE_GO_MOD_SUM='h1:uj3pm+hUTVN/X5yfdBexHlZv+1Xu5u5ZbZx7+CDavNU='
readonly ATLAS_CLI='/usr/local/bin/atlas'
readonly ATLAS_CLI_REFERENCE='docker.io/arigaio/atlas@sha256:ca911e2db73ab32a9b85953a70484ab70cda9f9fadc8b5c7c594910c5dc87478'
readonly ATLAS_CLI_PLATFORM='linux/amd64'
readonly ATLAS_CLI_RUNTIME_ARCH='x86_64'
readonly ATLAS_CLI_SOURCE_BINARY='/atlas'
readonly ATLAS_CLI_VERSION_OUTPUT_FIRST_LINE='atlas community version v1.3.2'
readonly ATLAS_CLI_BINARY_SHA256='bdae7d2b6b903fdc4c40313fedfdf87f8a08ce1f2a4fd5c62fea2308419d526e'
readonly ENT_MODULE='entgo.io/ent'
readonly ENT_VERSION='v0.14.1'
readonly ENT_MODULE_SUM='h1:fUERL506Pqr92EPHJqr8EYxbPioflJo6PudkrEA8a/s='
readonly ENT_MODULE_GO_MOD_SUM='h1:MH6XLG0KXpkcDQhKiHfANZSzR55TJyPL5IGNpI8wpco='
readonly ENT_GENERATE_DIRECTIVE_NORMALIZED_SHA256='9342a9f687de2a67c517afc060884b16655e88191bccae67de52f4a7e1a9a918'
readonly SOURCE_GO_MOD_NORMALIZED_SHA256='3f7ff699a420ecf0501693c47ee18108a02a8827a547578bb6b800f83bcc2be6'
readonly SOURCE_GO_MOD_NORMALIZED_BYTES=9763
readonly SOURCE_GO_SUM_NORMALIZED_SHA256='9a03c4db3f998434e3e9058a67b7d2143fda8c5a7ccee6895e9093d6e77d2510'
readonly SOURCE_GO_SUM_NORMALIZED_BYTES=55553

fail() {
  printf '%s\n' "$1" >&2
  exit "$2"
}

normalized_file_has_exact_line() {
  tr -d '\r' < "$2" | grep -Fqx -- "$1"
}

require_internal_database_urls() {
  [ "${DATABASE_URL:-}" = "$EXPECTED_DATABASE_URL" ] || fail 'schema-maintenance: DATABASE_URL must be the fixed internal disposable target' "$EX_CONFIG"
  [ "${PREFEATURE_SCHEMA_URL:-}" = "$EXPECTED_PREFEATURE_SCHEMA_URL" ] || fail 'schema-maintenance: PREFEATURE_SCHEMA_URL must be the fixed internal prefeature schema' "$EX_CONFIG"
  [ "${CURRENT_SCHEMA_URL:-}" = "$EXPECTED_CURRENT_SCHEMA_URL" ] || fail 'schema-maintenance: CURRENT_SCHEMA_URL must be the fixed internal current schema' "$EX_CONFIG"
  [ "${VERIFICATION_SCHEMA_URL:-}" = "$EXPECTED_VERIFICATION_SCHEMA_URL" ] || fail 'schema-maintenance: VERIFICATION_SCHEMA_URL must be the fixed internal verification schema' "$EX_CONFIG"
  [ "${ATLAS_DIFF_DEV_URL:-}" = "$EXPECTED_ATLAS_DIFF_DEV_URL" ] || fail 'schema-maintenance: ATLAS_DIFF_DEV_URL must be the fixed internal diff dev schema' "$EX_CONFIG"
  [ "${ATLAS_VALIDATE_DEV_URL:-}" = "$EXPECTED_ATLAS_VALIDATE_DEV_URL" ] || fail 'schema-maintenance: ATLAS_VALIDATE_DEV_URL must be the fixed internal validate dev schema' "$EX_CONFIG"
  [ "${ATLAS_LINT_DEV_URL:-}" = "$EXPECTED_ATLAS_LINT_DEV_URL" ] || fail 'schema-maintenance: ATLAS_LINT_DEV_URL must be the fixed internal lint dev schema' "$EX_CONFIG"
  [ "${ATLAS_COMPARE_DEV_URL:-}" = "$EXPECTED_ATLAS_COMPARE_DEV_URL" ] || fail 'schema-maintenance: ATLAS_COMPARE_DEV_URL must be the fixed internal compare dev schema' "$EX_CONFIG"
  [ "${ATLAS_REVISIONS_URL:-}" = "$EXPECTED_ATLAS_REVISIONS_URL" ] || fail 'schema-maintenance: ATLAS_REVISIONS_URL must be the fixed internal revisions schema' "$EX_CONFIG"
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

require_eight_sql_lineage_inputs() {
  require_exact_files "$EXPECTED_LINEAGE_ENTRY_COUNT" \
    "$BASELINE_FILENAME" \
    "$CORE_FILENAME" \
    "$DEFERRED_RECEIPT_EXPIRY_FILENAME" \
    "$INBOUND_MESSAGE_RECEIPT_EVENT_MESSAGE_IDX_FILENAME" \
    "$PASSWORD_ENROLLMENT_FILENAME" \
    "$DELIVERY_MATERIALIZATION_INVALID_FILENAME" \
    "$INVITE_SEQUENCE_NON_NEGATIVE_FILENAME" \
    "$ROUNDTRIP_RECONCILIATION_FILENAME" \
    atlas.sum
}

copy_eight_sql_lineage_inputs() {
  destination="$1"

  cp "$LINEAGE_DIR/$BASELINE_FILENAME" "$destination/$BASELINE_FILENAME" ||
    fail 'schema-maintenance: cannot copy baseline lineage input for hashing' "$EX_CONFIG"
  cp "$LINEAGE_DIR/$CORE_FILENAME" "$destination/$CORE_FILENAME" ||
    fail 'schema-maintenance: cannot copy core lineage input for hashing' "$EX_CONFIG"
  cp "$LINEAGE_DIR/$DEFERRED_RECEIPT_EXPIRY_FILENAME" "$destination/$DEFERRED_RECEIPT_EXPIRY_FILENAME" ||
    fail 'schema-maintenance: cannot copy deferred receipt expiry lineage input for hashing' "$EX_CONFIG"
  cp "$LINEAGE_DIR/$INBOUND_MESSAGE_RECEIPT_EVENT_MESSAGE_IDX_FILENAME" "$destination/$INBOUND_MESSAGE_RECEIPT_EVENT_MESSAGE_IDX_FILENAME" ||
    fail 'schema-maintenance: cannot copy inbound receipt event index lineage input for hashing' "$EX_CONFIG"
  cp "$LINEAGE_DIR/$PASSWORD_ENROLLMENT_FILENAME" "$destination/$PASSWORD_ENROLLMENT_FILENAME" ||
    fail 'schema-maintenance: cannot copy password enrollment lineage input for hashing' "$EX_CONFIG"
  cp "$LINEAGE_DIR/$DELIVERY_MATERIALIZATION_INVALID_FILENAME" "$destination/$DELIVERY_MATERIALIZATION_INVALID_FILENAME" ||
    fail 'schema-maintenance: cannot copy delivery materialization lineage input for hashing' "$EX_CONFIG"
  cp "$LINEAGE_DIR/$INVITE_SEQUENCE_NON_NEGATIVE_FILENAME" "$destination/$INVITE_SEQUENCE_NON_NEGATIVE_FILENAME" ||
    fail 'schema-maintenance: cannot copy invite sequence lineage input for hashing' "$EX_CONFIG"
  cp "$LINEAGE_DIR/$ROUNDTRIP_RECONCILIATION_FILENAME" "$destination/$ROUNDTRIP_RECONCILIATION_FILENAME" ||
    fail 'schema-maintenance: cannot copy roundtrip reconciliation lineage input for hashing' "$EX_CONFIG"
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
  lock_file="$SCHEMA_TOOL_LOCK"
  require_non_symlink "$lock_file"
  [ -f "$lock_file" ] || fail 'schema-maintenance: schema tool lock is absent' "$EX_CONFIG"
  normalized_file_has_exact_line '  "lock_version": 2,' "$lock_file" ||
    fail 'schema-maintenance: schema tool lock version is absent' "$EX_CONFIG"
  normalized_file_has_exact_line '  "atlas_module": {' "$lock_file" ||
    fail 'schema-maintenance: Atlas source module lock object is absent' "$EX_CONFIG"
  normalized_file_has_exact_line '    "purpose": "source_module_graph_dependency",' "$lock_file" ||
    fail 'schema-maintenance: Atlas source module purpose is absent' "$EX_CONFIG"

  if grep -Eq '^[[:space:]]*(replace|exclude)([[:space:]]|$)' "$SOURCE_ROOT/go.mod"; then
    fail 'schema-maintenance: generate source go.mod must not contain replace/exclude directives' "$EX_CONFIG"
  fi
  normalized_file_has_exact_line "$ATLAS_MODULE $ATLAS_MODULE_VERSION $ATLAS_MODULE_SUM" "$SOURCE_ROOT/go.sum" ||
    fail 'schema-maintenance: Atlas module checksum does not match the reviewed lock' "$EX_CONFIG"
  normalized_file_has_exact_line "$ATLAS_MODULE $ATLAS_MODULE_VERSION/go.mod $ATLAS_MODULE_GO_MOD_SUM" "$SOURCE_ROOT/go.sum" ||
    fail 'schema-maintenance: Atlas go.mod checksum does not match the reviewed lock' "$EX_CONFIG"
  normalized_file_has_exact_line "$ENT_MODULE $ENT_VERSION $ENT_MODULE_SUM" "$SOURCE_ROOT/go.sum" ||
    fail 'schema-maintenance: Ent module checksum does not match the reviewed lock' "$EX_CONFIG"
  normalized_file_has_exact_line "$ENT_MODULE $ENT_VERSION/go.mod $ENT_MODULE_GO_MOD_SUM" "$SOURCE_ROOT/go.sum" ||
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

require_exact_atlas_cli() {
  lock_file="$SCHEMA_TOOL_LOCK"
  [ "$(uname -m)" = "$ATLAS_CLI_RUNTIME_ARCH" ] ||
    fail 'schema-maintenance: Atlas CLI runtime architecture does not match the reviewed lock' "$EX_CONFIG"
  require_non_symlink "$ATLAS_CLI"
  [ -f "$ATLAS_CLI" ] && [ ! -L "$ATLAS_CLI" ] ||
    fail 'schema-maintenance: Atlas CLI must be a regular non-symlink file' "$EX_CONFIG"
  [ "$(stat -c '%U:%G:%a' "$ATLAS_CLI")" = 'root:root:555' ] ||
    fail 'schema-maintenance: Atlas CLI ownership or mode does not match the reviewed lock' "$EX_CONFIG"

  normalized_file_has_exact_line '  "lock_version": 2,' "$lock_file" ||
    fail 'schema-maintenance: Atlas CLI lock version is absent' "$EX_CONFIG"
  normalized_file_has_exact_line '    "atlas_cli": {' "$lock_file" ||
    fail 'schema-maintenance: Atlas CLI lock object is absent' "$EX_CONFIG"
  normalized_file_has_exact_line '      "reference": "docker.io/arigaio/atlas@sha256:ca911e2db73ab32a9b85953a70484ab70cda9f9fadc8b5c7c594910c5dc87478",' "$lock_file" ||
    fail 'schema-maintenance: Atlas CLI reference does not match the reviewed lock' "$EX_CONFIG"
  normalized_file_has_exact_line '      "platform": "linux/amd64",' "$lock_file" ||
    fail 'schema-maintenance: Atlas CLI platform does not match the reviewed lock' "$EX_CONFIG"
  normalized_file_has_exact_line '      "expected_runtime_arch": "x86_64",' "$lock_file" ||
    fail 'schema-maintenance: Atlas CLI architecture does not match the reviewed lock' "$EX_CONFIG"
  normalized_file_has_exact_line '      "source_binary": "/atlas",' "$lock_file" ||
    fail 'schema-maintenance: Atlas CLI source path does not match the reviewed lock' "$EX_CONFIG"
  normalized_file_has_exact_line '      "expected_version_output_first_line": "atlas community version v1.3.2",' "$lock_file" ||
    fail 'schema-maintenance: Atlas CLI version does not match the reviewed lock' "$EX_CONFIG"
  normalized_file_has_exact_line '      "binary_sha256": "bdae7d2b6b903fdc4c40313fedfdf87f8a08ce1f2a4fd5c62fea2308419d526e"' "$lock_file" ||
    fail 'schema-maintenance: Atlas CLI hash is absent from the reviewed lock' "$EX_CONFIG"

  actual_hash="$(sha256sum "$ATLAS_CLI" | cut -d ' ' -f1)"
  [ "$actual_hash" = "$ATLAS_CLI_BINARY_SHA256" ] ||
    fail 'schema-maintenance: Atlas CLI binary hash does not match the reviewed lock' "$EX_CONFIG"
  if ! actual_atlas_version="$("$ATLAS_CLI" version 2>&1 | sed -n '1p')"; then
    fail 'schema-maintenance: Atlas CLI version command failed' "$EX_CONFIG"
  fi
  [ "$actual_atlas_version" = "$ATLAS_CLI_VERSION_OUTPUT_FIRST_LINE" ] ||
    fail 'schema-maintenance: Atlas CLI version output does not match the reviewed lock' "$EX_CONFIG"
}

require_ent_generate_directive_alignment() {
  lock_file="$SCHEMA_TOOL_LOCK"
  grep -Fqx '    "generate_go_normalized_sha256": "9342a9f687de2a67c517afc060884b16655e88191bccae67de52f4a7e1a9a918",' "$lock_file" ||
    fail 'schema-maintenance: schema tool lock does not carry the reviewed Ent directive hash' "$EX_CONFIG"
  grep -Fqx '    "generate_command": "go run -mod=readonly entgo.io/ent/cmd/ent generate --feature sql/lock ./ent/schema",' "$lock_file" ||
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
    "$SOURCE_ROOT/ent/schema" \
    "$SOURCE_ROOT/internal" \
    "$SOURCE_ROOT/internal/mailvalue" \
    "$SOURCE_ROOT/internal/mailvalue/dns.go" \
    "$SOURCE_ROOT/internal/mail" \
    "$SOURCE_ROOT/internal/mail/dns_types.go"; do
    require_non_symlink "$source_path"
  done
  [ -f "$SOURCE_ROOT/go.mod" ] || fail 'schema-maintenance: generate source go.mod is absent' "$EX_CONFIG"
  [ -f "$SOURCE_ROOT/go.sum" ] || fail 'schema-maintenance: generate source go.sum is absent' "$EX_CONFIG"
  [ -f "$SOURCE_ROOT/ent/generate.go" ] || fail 'schema-maintenance: generate source directive is absent' "$EX_CONFIG"
  [ -d "$SOURCE_ROOT/ent/schema" ] || fail 'schema-maintenance: generate source schema is absent' "$EX_CONFIG"
  [ -f "$SOURCE_ROOT/internal/mailvalue/dns.go" ] || fail 'schema-maintenance: generate source mailvalue DNS helper is absent' "$EX_CONFIG"
  [ -f "$SOURCE_ROOT/internal/mail/dns_types.go" ] || fail 'schema-maintenance: generate source mail DNS types are absent' "$EX_CONFIG"
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
  mkdir -p "$module_root/ent" "$module_root/internal/mailvalue" "$module_root/internal/mail"
  cp "$SOURCE_ROOT/go.mod" "$module_root/go.mod"
  cp "$SOURCE_ROOT/go.sum" "$module_root/go.sum"
  cp "$SOURCE_ROOT/ent/generate.go" "$module_root/ent/generate.go"
  cp "$SOURCE_ROOT/internal/mailvalue/dns.go" "$module_root/internal/mailvalue/dns.go"
  cp "$SOURCE_ROOT/internal/mail/dns_types.go" "$module_root/internal/mail/dns_types.go"
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
    "$module_root/ent/schema" \
    "$module_root/internal" \
    "$module_root/internal/mailvalue" \
    "$module_root/internal/mailvalue/dns.go" \
    "$module_root/internal/mail" \
    "$module_root/internal/mail/dns_types.go"; do
    require_non_symlink "$source_path"
  done
  [ -f "$module_root/go.mod" ] || fail 'schema-maintenance: isolated generate root go.mod is absent' "$EX_CONFIG"
  [ -f "$module_root/go.sum" ] || fail 'schema-maintenance: isolated generate root go.sum is absent' "$EX_CONFIG"
  [ -f "$module_root/ent/generate.go" ] || fail 'schema-maintenance: isolated generate root directive is absent' "$EX_CONFIG"
  [ -d "$module_root/ent/schema" ] || fail 'schema-maintenance: isolated generate root schema is absent' "$EX_CONFIG"
  [ -f "$module_root/internal/mailvalue/dns.go" ] || fail 'schema-maintenance: isolated generate root mailvalue DNS helper is absent' "$EX_CONFIG"
  [ -f "$module_root/internal/mail/dns_types.go" ] || fail 'schema-maintenance: isolated generate root mail DNS types are absent' "$EX_CONFIG"
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
  append_generate_source_manifest_entry "$module_root" "$module_root/internal/mailvalue/dns.go" "$manifest"
  append_generate_source_manifest_entry "$module_root" "$module_root/internal/mail/dns_types.go" "$manifest"
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
  diagnostic_file="$module_root/ent-generate.stderr"
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
    ulimit -f 65536
    cd "$module_root"
    HOME="$module_home" XDG_CACHE_HOME="$module_xdg_cache" TMPDIR="$module_exec_tmp" GOTOOLCHAIN=local GOWORK=off GOPROXY=off GOSUMDB=off GOPATH=/opt/hub/go GOMODCACHE=/opt/hub/go/pkg/mod GOFLAGS='-mod=readonly -p=2' GOCACHE="$module_cache" GOTMPDIR="$module_exec_tmp" go run -mod=readonly entgo.io/ent/cmd/ent generate --feature sql/lock ./ent/schema
  ) >"$diagnostic_file" 2>&1; then
    require_non_symlink "$diagnostic_file"
    [ -f "$diagnostic_file" ] || fail 'schema-maintenance: Ent generation diagnostic is absent or non-regular' "$EX_CONFIG"
    emit_review_file "$diagnostic_file" >&2
    fail 'schema-maintenance: Ent generation failed inside its isolated offline module root' "$EX_CONFIG"
  fi
  rm -f "$diagnostic_file"
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

require_regular_non_symlink() {
  materializer_path="$1"
  materializer_label="$2"
  require_non_symlink "$materializer_path"
  [ -f "$materializer_path" ] || fail "$materializer_label must be a regular file: $materializer_path" "$EX_CONFIG"
}

require_safe_nonempty_schema_tree() {
  materializer_path="$1"
  materializer_label="$2"
  require_non_symlink "$materializer_path"
  [ -d "$materializer_path" ] || fail "$materializer_label must be a directory: $materializer_path" "$EX_CONFIG"
  materializer_first_entry="$(find "$materializer_path" -mindepth 1 -print -quit)"
  [ -n "$materializer_first_entry" ] || fail "$materializer_label must not be empty: $materializer_path" "$EX_CONFIG"
  materializer_unsafe_entry="$(find "$materializer_path" -mindepth 1 -type l -print -quit)"
  [ -z "$materializer_unsafe_entry" ] || fail "$materializer_label must not contain symlinks: $materializer_path" "$EX_CONFIG"
  materializer_unsafe_entry="$(find "$materializer_path" -mindepth 1 \( ! -type d -a ! -type f \) -print -quit)"
  [ -z "$materializer_unsafe_entry" ] || fail "$materializer_label must not contain non-regular entries: $materializer_path" "$EX_CONFIG"
}

require_materializer_lock() {
  normalized_file_has_exact_line '  "lock_version": 1,' "$MATERIALIZER_LOCK" ||
    fail "materializer lock version is absent" "$EX_CONFIG"
  normalized_file_has_exact_line '    "path": "docker/schema-maintenance/materializer/main.go",' "$MATERIALIZER_LOCK" ||
    fail "materializer lock source path is absent" "$EX_CONFIG"
  normalized_file_has_exact_line "    \"normalized_sha256\": \"$MATERIALIZER_NORMALIZED_SHA256\"" "$MATERIALIZER_LOCK" ||
    fail "materializer lock source hash is absent" "$EX_CONFIG"
  normalized_file_has_exact_line '  "runtime_source_mounts": [],' "$MATERIALIZER_LOCK" ||
    fail "materializer lock runtime mount policy is absent" "$EX_CONFIG"
  normalized_file_has_exact_line '  "network": "schema-maintenance-internal only",' "$MATERIALIZER_LOCK" ||
    fail "materializer lock network policy is absent" "$EX_CONFIG"
  normalized_file_has_exact_line '    "destructive_operations": false,' "$MATERIALIZER_LOCK" ||
    fail "materializer lock destructive-operation policy is absent" "$EX_CONFIG"
  normalized_file_has_exact_line '    "drop_schema": false' "$MATERIALIZER_LOCK" ||
    fail "materializer lock drop-schema policy is absent" "$EX_CONFIG"
}

require_materializer_source() {
  require_non_symlink "$SOURCE_ROOT"
  [ -d "$SOURCE_ROOT" ] || fail "SOURCE_ROOT must be a directory: $SOURCE_ROOT" "$EX_CONFIG"
  require_regular_non_symlink "$SOURCE_ROOT/go.mod" "current go.mod"
  require_regular_non_symlink "$SOURCE_ROOT/go.sum" "current go.sum"
  require_safe_nonempty_schema_tree "$SOURCE_ROOT/ent/schema" "current ent schema"
  require_regular_non_symlink "$SOURCE_ROOT/internal/mailvalue/dns.go" "current mailvalue DNS helper"
  require_regular_non_symlink "$SOURCE_ROOT/internal/mail/dns_types.go" "current mail DNS helper"
  require_regular_non_symlink "$MATERIALIZER_SOURCE" "materializer source"
  require_regular_non_symlink "$MATERIALIZER_LOCK" "materializer lock"
  require_non_symlink "/opt/hub/prefeature/backend"
  [ -d "/opt/hub/prefeature/backend" ] || fail "prefeature backend root must be a directory" "$EX_CONFIG"
  require_regular_non_symlink "/opt/hub/prefeature/backend/go.mod" "prefeature go.mod"
  require_regular_non_symlink "/opt/hub/prefeature/backend/go.sum" "prefeature go.sum"
  require_safe_nonempty_schema_tree "/opt/hub/prefeature/backend/ent/schema" "prefeature ent schema"
  require_exact_go_dependency_pins
  require_materializer_lock

  materializer_carriage_return="$(printf '\r')"
  materializer_normalized_sha256="$(LC_ALL=C sed "s/${materializer_carriage_return}\$//" "$MATERIALIZER_SOURCE" | sha256sum | awk '{print $1}')"
  [ "$materializer_normalized_sha256" = "$MATERIALIZER_NORMALIZED_SHA256" ] ||
    fail "materializer source hash does not match MATERIALIZER_NORMALIZED_SHA256" "$EX_CONFIG"
}

append_materializer_source_manifest_entry() {
  materializer_manifest_prefix="$1"
  materializer_manifest_root="$2"
  materializer_manifest_file="$3"
  materializer_manifest_path="$4"
  case "$materializer_manifest_file" in
    "$materializer_manifest_root"/*) materializer_manifest_relative=${materializer_manifest_file#"$materializer_manifest_root"/} ;;
    *) fail "materializer manifest entry escapes its fixed root" "$EX_CONFIG" ;;
  esac
  case "$materializer_manifest_relative" in
    ''|/*|*'//'|*'/./'*|*'/../'*|*[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._/-]*)
      fail "materializer manifest entry has an unsafe relative path" "$EX_CONFIG"
      ;;
  esac
  require_regular_non_symlink "$materializer_manifest_file" "materializer manifest source entry"
  materializer_manifest_bytes="$(wc -c < "$materializer_manifest_file" | tr -d ' ')"
  materializer_manifest_sha256="$(sha256sum "$materializer_manifest_file" | awk '{print $1}')"
  printf '%s/%s %s %s\n' "$materializer_manifest_prefix" "$materializer_manifest_relative" "$materializer_manifest_bytes" "$materializer_manifest_sha256" >> "$materializer_manifest_path" ||
    fail "cannot append materializer source manifest entry" "$EX_CONFIG"
}

write_materializer_source_manifest() {
  [ ! -e "$MATERIALIZER_SOURCE_MANIFEST" ] && [ ! -L "$MATERIALIZER_SOURCE_MANIFEST" ] ||
    fail "materializer source manifest already exists" "$EX_CONFIG"
  : > "$MATERIALIZER_SOURCE_MANIFEST" || fail "cannot create materializer source manifest" "$EX_CONFIG"

  append_materializer_source_manifest_entry current "$SOURCE_ROOT" "$SOURCE_ROOT/go.mod" "$MATERIALIZER_SOURCE_MANIFEST"
  append_materializer_source_manifest_entry current "$SOURCE_ROOT" "$SOURCE_ROOT/go.sum" "$MATERIALIZER_SOURCE_MANIFEST"
  append_materializer_source_manifest_entry current "$SOURCE_ROOT" "$SOURCE_ROOT/docker/schema-maintenance/materializer/main.go" "$MATERIALIZER_SOURCE_MANIFEST"
  append_materializer_source_manifest_entry current "$SOURCE_ROOT" "$SOURCE_ROOT/docker/schema-materializer.lock.json" "$MATERIALIZER_SOURCE_MANIFEST"
  append_materializer_source_manifest_entry current "$SOURCE_ROOT" "$SOURCE_ROOT/internal/mailvalue/dns.go" "$MATERIALIZER_SOURCE_MANIFEST"
  append_materializer_source_manifest_entry current "$SOURCE_ROOT" "$SOURCE_ROOT/internal/mail/dns_types.go" "$MATERIALIZER_SOURCE_MANIFEST"
  find "$SOURCE_ROOT/ent/schema" -type f -print |
    LC_ALL=C sort |
    while IFS= read -r materializer_manifest_schema_file; do
      append_materializer_source_manifest_entry current "$SOURCE_ROOT" "$materializer_manifest_schema_file" "$MATERIALIZER_SOURCE_MANIFEST"
    done ||
    fail "cannot append current Ent schema manifest entries" "$EX_CONFIG"

  append_materializer_source_manifest_entry prefeature "/opt/hub/prefeature/backend" "/opt/hub/prefeature/backend/go.mod" "$MATERIALIZER_SOURCE_MANIFEST"
  append_materializer_source_manifest_entry prefeature "/opt/hub/prefeature/backend" "/opt/hub/prefeature/backend/go.sum" "$MATERIALIZER_SOURCE_MANIFEST"
  find "/opt/hub/prefeature/backend/ent/schema" -type f -print |
    LC_ALL=C sort |
    while IFS= read -r materializer_manifest_schema_file; do
      append_materializer_source_manifest_entry prefeature "/opt/hub/prefeature/backend" "$materializer_manifest_schema_file" "$MATERIALIZER_SOURCE_MANIFEST"
    done ||
    fail "cannot append prefeature Ent schema manifest entries" "$EX_CONFIG"

  require_regular_non_symlink "$MATERIALIZER_SOURCE_MANIFEST" "materializer source manifest"
  materializer_manifest_size="$(wc -c < "$MATERIALIZER_SOURCE_MANIFEST" | tr -d ' ')"
  [ "$materializer_manifest_size" -gt 0 ] || fail "materializer source manifest must not be empty" "$EX_CONFIG"
  [ "$materializer_manifest_size" -le "$REVIEW_FILE_LIMIT_BYTES" ] ||
    fail "materializer source manifest exceeds review size limit" "$EX_CONFIG"
}

prepare_materializer() {
  materializer_build_root="$GO_RUN_EXEC_ROOT/materializer-build"
  materializer_tmp_root="$materializer_build_root/tmp"
  mkdir -p "$WORK_ROOT/home" "$WORK_ROOT/xdg/cache" "$WORK_ROOT/xdg/config" "$WORK_ROOT/xdg/data" "$WORK_ROOT/gocache" ||
    fail "cannot create materializer build directories" "$EX_CONFIG"

  if ! (
    cd "$SOURCE_ROOT"
    HOME="$WORK_ROOT/home" XDG_CACHE_HOME="$WORK_ROOT/xdg/cache" XDG_CONFIG_HOME="$WORK_ROOT/xdg/config" XDG_DATA_HOME="$WORK_ROOT/xdg/data" TMPDIR="$materializer_tmp_root" GOTMPDIR="$materializer_tmp_root" GOCACHE="$WORK_ROOT/gocache" GOTOOLCHAIN=local GOWORK=off GOPROXY=off GOSUMDB=off GOPATH=/opt/hub/go GOMODCACHE=/opt/hub/go/pkg/mod GOFLAGS='-mod=readonly -p=2' go build -o "$MATERIALIZER_BIN" "$MATERIALIZER_SOURCE"
  ); then
    fail "cannot build schema materializer" "$EX_CONFIG"
  fi
  require_regular_non_symlink "$MATERIALIZER_BIN" "materializer binary"
  [ -x "$MATERIALIZER_BIN" ] || fail "materializer binary is not executable: $MATERIALIZER_BIN" "$EX_CONFIG"
}

run_materializer() {
  HOME="$WORK_ROOT/home" \
  XDG_CACHE_HOME="$WORK_ROOT/xdg/cache" \
  XDG_CONFIG_HOME="$WORK_ROOT/xdg/config" \
  XDG_DATA_HOME="$WORK_ROOT/xdg/data" \
  TMPDIR="$GO_RUN_EXEC_ROOT/materializer-build/tmp" \
  GOTMPDIR="$GO_RUN_EXEC_ROOT/materializer-build/tmp" \
  GOCACHE="$WORK_ROOT/gocache" \
  GOTOOLCHAIN=local \
  GOWORK=off \
  GOPROXY=off \
  GOSUMDB=off \
  GOPATH=/opt/hub/go \
  GOMODCACHE=/opt/hub/go/pkg/mod \
  GOFLAGS='-mod=readonly -p=2' \
  "$MATERIALIZER_BIN" "$@"
}

stop_materializer_lock() {
  [ -n "$MATERIALIZER_PID" ] || return 0
  if kill -0 "$MATERIALIZER_PID" 2>/dev/null; then
    kill -TERM "$MATERIALIZER_PID" 2>/dev/null || :
  fi
  wait "$MATERIALIZER_PID" 2>/dev/null || :
  MATERIALIZER_PID=''
}

acquire_materializer_lock() {
  materializer_attempt=0
  HOME="$WORK_ROOT/home" \
  XDG_CACHE_HOME="$WORK_ROOT/xdg/cache" \
  XDG_CONFIG_HOME="$WORK_ROOT/xdg/config" \
  XDG_DATA_HOME="$WORK_ROOT/xdg/data" \
  TMPDIR="$GO_RUN_EXEC_ROOT/materializer-build/tmp" \
  GOTMPDIR="$GO_RUN_EXEC_ROOT/materializer-build/tmp" \
  GOCACHE="$WORK_ROOT/gocache" \
  GOTOOLCHAIN=local \
  GOWORK=off \
  GOPROXY=off \
  GOSUMDB=off \
  GOPATH=/opt/hub/go \
  GOMODCACHE=/opt/hub/go/pkg/mod \
  GOFLAGS='-mod=readonly -p=2' \
  "$MATERIALIZER_BIN" lock &
  MATERIALIZER_PID=$!
  while [ "$materializer_attempt" -lt 150 ]; do
    if [ -f "$MATERIALIZER_READY_FILE" ] && [ ! -L "$MATERIALIZER_READY_FILE" ]; then
      if kill -0 "$MATERIALIZER_PID" 2>/dev/null; then
        return 0
      fi
      stop_materializer_lock
      fail "materializer lock exited after reporting ready" "$EX_CONFIG"
    fi
    if ! kill -0 "$MATERIALIZER_PID" 2>/dev/null; then
      stop_materializer_lock
      fail "materializer lock exited before reporting ready" "$EX_CONFIG"
    fi
    sleep 0.2
    materializer_attempt=$((materializer_attempt + 1))
  done
  stop_materializer_lock
  fail "timed out waiting for materializer lock readiness" "$EX_CONFIG"
}

cleanup_materializer_job() {
  materializer_cleanup_status=$?
  trap - 0 HUP INT TERM
  set +e
  stop_materializer_lock
  exit "$materializer_cleanup_status"
}

start_materializer_job() {
  require_materializer_source
  [ ! -e "$WORK_ROOT" ] && [ ! -L "$WORK_ROOT" ] ||
    fail "materializer work root already exists: $WORK_ROOT" "$EX_CONFIG"
  [ ! -e "$GO_RUN_EXEC_ROOT/materializer-build" ] && [ ! -L "$GO_RUN_EXEC_ROOT/materializer-build" ] ||
    fail "materializer build root already exists: $GO_RUN_EXEC_ROOT/materializer-build" "$EX_CONFIG"
  require_non_symlink "$GO_RUN_EXEC_ROOT"
  [ -d "$GO_RUN_EXEC_ROOT" ] || fail "Go run-exec root must be a directory: $GO_RUN_EXEC_ROOT" "$EX_CONFIG"

  umask 077
  mkdir "$WORK_ROOT" || fail "cannot create materializer work root" "$EX_CONFIG"
  mkdir -p "$GO_RUN_EXEC_ROOT/materializer-build/tmp" || fail "cannot create materializer temporary root" "$EX_CONFIG"
  write_materializer_source_manifest
  emit_review_file "$MATERIALIZER_SOURCE_MANIFEST"
  prepare_materializer
  acquire_materializer_lock
  trap cleanup_materializer_job 0 HUP INT TERM
  run_materializer initialize || fail "cannot initialize materializer schemas" "$EX_CONFIG"
}

run_baseline() {
  require_exact_atlas_cli
  require_internal_database_urls
  start_materializer_job

  candidate_dir="$WORK_ROOT/baseline"
  mkdir "$candidate_dir" || fail 'schema-maintenance: cannot create baseline candidate directory' "$EX_CONFIG"

  run_materializer materialize prefeature || fail 'schema-maintenance: cannot materialize prefeature schema' "$EX_CONFIG"
  "$ATLAS_CLI" schema inspect \
    --url "$PREFEATURE_SCHEMA_URL" \
    --format '{{ sql . }}' > "$candidate_dir/$BASELINE_FILENAME" ||
    fail 'schema-maintenance: cannot inspect prefeature schema' "$EX_CONFIG"
  "$ATLAS_CLI" migrate hash --dir "file://$candidate_dir" ||
    fail 'schema-maintenance: cannot hash baseline migration directory' "$EX_CONFIG"

  require_exact_candidate_files "$candidate_dir" 2 "$BASELINE_FILENAME" atlas.sum
  emit_review_file "$candidate_dir/$BASELINE_FILENAME"
  emit_review_file "$candidate_dir/atlas.sum"
}

run_diff() {
  require_exact_atlas_cli
  require_internal_database_urls
  require_exact_files 3 "$BASELINE_FILENAME" "$CORE_FILENAME" atlas.sum
  start_materializer_job

  candidate_dir="$WORK_ROOT/diff"
  mkdir "$candidate_dir" || fail 'schema-maintenance: cannot create diff candidate directory' "$EX_CONFIG"
  cp "$LINEAGE_DIR/20260823000000_mail_calendar_baseline.sql" "$candidate_dir/"
  "$ATLAS_CLI" migrate hash --dir "file://$candidate_dir" ||
    fail 'schema-maintenance: cannot create baseline-only Atlas checksum' "$EX_CONFIG"
  require_exact_candidate_files "$candidate_dir" 2 "$BASELINE_FILENAME" atlas.sum

  run_materializer materialize current || fail 'schema-maintenance: cannot materialize current schema' "$EX_CONFIG"
  "$ATLAS_CLI" migrate diff mail_calendar_core \
    --dev-url "$ATLAS_DIFF_DEV_URL" \
    --to "$CURRENT_SCHEMA_URL" \
    --dir "file://$candidate_dir" ||
    fail 'schema-maintenance: cannot create Mail/Calendar migration candidate' "$EX_CONFIG"
  non_regular_entry="$(find "$candidate_dir" -mindepth 1 -maxdepth 1 ! -type f -print -quit)"
  [ -z "$non_regular_entry" ] || fail 'schema-maintenance: Atlas candidate contains a non-regular output entry' "$EX_CONFIG"
  require_candidate_files="$(find "$candidate_dir" -mindepth 1 -maxdepth 1 -type f -name '*_mail_calendar_core.sql' | wc -l | tr -d ' ')"
  [ "$require_candidate_files" -eq 1 ] || fail 'schema-maintenance: Atlas did not produce exactly one bounded Mail/Calendar candidate' "$EX_CONFIG"
  candidate_sql="$(find "$candidate_dir" -mindepth 1 -maxdepth 1 -type f -name '*_mail_calendar_core.sql')"
  [ "$(find "$candidate_dir" -mindepth 1 -maxdepth 1 -type f | wc -l | tr -d ' ')" -eq 3 ] || fail 'schema-maintenance: Atlas candidate contains unreviewed output files' "$EX_CONFIG"
  require_non_symlink "$candidate_sql"
  [ -f "$candidate_sql" ] || fail 'schema-maintenance: Mail/Calendar diff candidate is not regular' "$EX_CONFIG"
  core_sql="$candidate_dir/$CORE_FILENAME"
  if [ "$candidate_sql" != "$core_sql" ]; then
    [ ! -e "$core_sql" ] && [ ! -L "$core_sql" ] ||
      fail 'schema-maintenance: fixed core migration path already exists' "$EX_CONFIG"
    mv "$candidate_sql" "$core_sql" ||
      fail 'schema-maintenance: cannot rename candidate to fixed core migration filename' "$EX_CONFIG"
  fi
  "$ATLAS_CLI" migrate hash --dir "file://$candidate_dir" ||
    fail 'schema-maintenance: cannot hash renamed migration candidate directory' "$EX_CONFIG"
  require_exact_candidate_files "$candidate_dir" 3 "$BASELINE_FILENAME" "$CORE_FILENAME" atlas.sum

  emit_review_file "$core_sql"
  emit_review_file "$candidate_dir/atlas.sum"
}

run_hash() {
  require_exact_atlas_cli
  require_eight_sql_lineage_inputs

  [ ! -e "$WORK_ROOT" ] && [ ! -L "$WORK_ROOT" ] ||
    fail 'schema-maintenance: hash work root already exists' "$EX_CONFIG"
  mkdir "$WORK_ROOT" ||
    fail 'schema-maintenance: cannot create hash work root' "$EX_CONFIG"
  candidate_dir="$WORK_ROOT/lineage-hash"
  [ ! -e "$candidate_dir" ] && [ ! -L "$candidate_dir" ] ||
    fail 'schema-maintenance: hash candidate directory already exists' "$EX_CONFIG"
  mkdir "$candidate_dir" ||
    fail 'schema-maintenance: cannot create hash candidate directory' "$EX_CONFIG"

  copy_eight_sql_lineage_inputs "$candidate_dir"
  "$ATLAS_CLI" migrate hash --dir "file://$candidate_dir" > /dev/null ||
    fail 'schema-maintenance: cannot hash reviewed Mail/Calendar lineage inputs' "$EX_CONFIG"
  require_exact_candidate_files "$candidate_dir" "$EXPECTED_LINEAGE_ENTRY_COUNT" \
    "$BASELINE_FILENAME" \
    "$CORE_FILENAME" \
    "$DEFERRED_RECEIPT_EXPIRY_FILENAME" \
    "$INBOUND_MESSAGE_RECEIPT_EVENT_MESSAGE_IDX_FILENAME" \
    "$PASSWORD_ENROLLMENT_FILENAME" \
    "$DELIVERY_MATERIALIZATION_INVALID_FILENAME" \
    "$INVITE_SEQUENCE_NON_NEGATIVE_FILENAME" \
    "$ROUNDTRIP_RECONCILIATION_FILENAME" \
    atlas.sum

  emit_review_file "$candidate_dir/atlas.sum"
}

run_verify() {
  require_exact_atlas_cli
  require_internal_database_urls
  require_eight_sql_lineage_inputs
  start_materializer_job

  run_materializer materialize current || fail 'schema-maintenance: cannot materialize current schema' "$EX_CONFIG"
  "$ATLAS_CLI" migrate validate \
    --dir "file://$LINEAGE_DIR" \
    --dev-url "$ATLAS_VALIDATE_DEV_URL" ||
    fail 'schema-maintenance: Mail/Calendar lineage validation failed' "$EX_CONFIG"
  "$ATLAS_CLI" migrate lint \
    --dir "file://$LINEAGE_DIR" \
    --dev-url "$ATLAS_LINT_DEV_URL" \
    --latest "$EXPECTED_LINEAGE_SQL_FILE_COUNT" ||
    fail 'schema-maintenance: Mail/Calendar lineage lint failed' "$EX_CONFIG"
  "$ATLAS_CLI" migrate apply \
    --dir "file://$LINEAGE_DIR" \
    --url "$VERIFICATION_SCHEMA_URL" \
    --revisions-schema atlas_revisions ||
    fail 'schema-maintenance: cannot apply Mail/Calendar lineage to verification schema' "$EX_CONFIG"

  materializer_verification_status="$WORK_ROOT/verification.status"
  "$ATLAS_CLI" migrate status \
    --dir "file://$LINEAGE_DIR" \
    --url "$VERIFICATION_SCHEMA_URL" \
    --revisions-schema atlas_revisions > "$materializer_verification_status" ||
    fail 'schema-maintenance: cannot inspect verification migration status' "$EX_CONFIG"
  require_non_symlink "$materializer_verification_status"
  [ -f "$materializer_verification_status" ] ||
    fail 'schema-maintenance: verification migration status is not regular' "$EX_CONFIG"
  [ -s "$materializer_verification_status" ] ||
    fail 'schema-maintenance: verification migration status is empty' "$EX_CONFIG"
  materializer_verification_status_size="$(wc -c < "$materializer_verification_status" | tr -d ' ')"
  [ "$materializer_verification_status_size" -le "$REVIEW_FILE_LIMIT_BYTES" ] ||
    fail 'schema-maintenance: verification migration status exceeds review size limit' "$EX_CONFIG"
  if ! grep -Eq '^[[:space:]]*Migration Status:[[:space:]]*OK[[:space:]]*$' "$materializer_verification_status" ||
    ! grep -Eq "^[[:space:]]*--[[:space:]]+Current Version:[[:space:]]*$EXPECTED_LATEST_VERSION[[:space:]]*\$" "$materializer_verification_status" ||
    ! grep -Eq '^[[:space:]]*--[[:space:]]+Next Version:[[:space:]]*Already at latest version[[:space:]]*$' "$materializer_verification_status" ||
    ! grep -Eq "^[[:space:]]*--[[:space:]]+Executed Files:[[:space:]]*$EXPECTED_LINEAGE_SQL_FILE_COUNT[[:space:]]*\$" "$materializer_verification_status" ||
    ! grep -Eq '^[[:space:]]*--[[:space:]]+Pending Files:[[:space:]]*0[[:space:]]*$' "$materializer_verification_status"; then
    emit_review_file "$materializer_verification_status" >&2
    fail 'schema-maintenance: verification migration status did not converge' "$EX_CONFIG"
  fi
  emit_review_file "$materializer_verification_status"

  materializer_verification_schema="$WORK_ROOT/verification.schema.sql"
  materializer_current_schema="$WORK_ROOT/current.schema.sql"
  "$ATLAS_CLI" schema inspect \
    --url "$VERIFICATION_SCHEMA_URL" \
    --format '{{ sql . }}' > "$materializer_verification_schema" ||
    fail 'schema-maintenance: cannot inspect verification schema' "$EX_CONFIG"
  "$ATLAS_CLI" schema inspect \
    --url "$CURRENT_SCHEMA_URL" \
    --format '{{ sql . }}' > "$materializer_current_schema" ||
    fail 'schema-maintenance: cannot inspect current schema' "$EX_CONFIG"

  require_non_symlink "$materializer_verification_schema"
  [ -f "$materializer_verification_schema" ] ||
    fail 'schema-maintenance: verification schema inspection is not regular' "$EX_CONFIG"
  [ -s "$materializer_verification_schema" ] ||
    fail 'schema-maintenance: verification schema inspection is empty' "$EX_CONFIG"
  materializer_verification_schema_size="$(wc -c < "$materializer_verification_schema" | tr -d ' ')"
  [ "$materializer_verification_schema_size" -le "$REVIEW_FILE_LIMIT_BYTES" ] ||
    fail 'schema-maintenance: verification schema inspection exceeds review size limit' "$EX_CONFIG"

  require_non_symlink "$materializer_current_schema"
  [ -f "$materializer_current_schema" ] ||
    fail 'schema-maintenance: current schema inspection is not regular' "$EX_CONFIG"
  [ -s "$materializer_current_schema" ] ||
    fail 'schema-maintenance: current schema inspection is empty' "$EX_CONFIG"
  materializer_current_schema_size="$(wc -c < "$materializer_current_schema" | tr -d ' ')"
  [ "$materializer_current_schema_size" -le "$REVIEW_FILE_LIMIT_BYTES" ] ||
    fail 'schema-maintenance: current schema inspection exceeds review size limit' "$EX_CONFIG"

  materializer_snapshot_dir="$WORK_ROOT/current-snapshot"
  materializer_snapshot_file="$materializer_snapshot_dir/$SNAPSHOT_FILENAME"
  materializer_snapshot_apply_log="$WORK_ROOT/current-snapshot.apply.log"
  materializer_snapshot_status="$WORK_ROOT/current-snapshot.status"
  materializer_compare_roundtrip_schema="$WORK_ROOT/compare.roundtrip.schema.sql"
  mkdir "$materializer_snapshot_dir" ||
    fail 'schema-maintenance: cannot create current schema snapshot directory' "$EX_CONFIG"
  cp "$materializer_current_schema" "$materializer_snapshot_file" ||
    fail 'schema-maintenance: cannot create current schema snapshot migration' "$EX_CONFIG"
  "$ATLAS_CLI" migrate hash --dir "file://$materializer_snapshot_dir" ||
    fail 'schema-maintenance: cannot hash current schema snapshot' "$EX_CONFIG"
  require_exact_candidate_files "$materializer_snapshot_dir" 2 "$SNAPSHOT_FILENAME" atlas.sum

  : > "$materializer_snapshot_apply_log" ||
    fail 'schema-maintenance: cannot create current schema snapshot apply log' "$EX_CONFIG"
  if ! "$ATLAS_CLI" migrate apply \
    --dir "file://$materializer_snapshot_dir" \
    --url "$ATLAS_COMPARE_DEV_URL" \
    --revisions-schema atlas_revisions > "$materializer_snapshot_apply_log" 2>&1; then
    require_non_symlink "$materializer_snapshot_apply_log"
    [ -f "$materializer_snapshot_apply_log" ] ||
      fail 'schema-maintenance: current schema snapshot apply log is not regular' "$EX_CONFIG"
    materializer_snapshot_apply_log_size="$(wc -c < "$materializer_snapshot_apply_log" | tr -d ' ')"
    [ "$materializer_snapshot_apply_log_size" -le "$REVIEW_FILE_LIMIT_BYTES" ] ||
      fail 'schema-maintenance: current schema snapshot apply log exceeds review size limit' "$EX_CONFIG"
    emit_review_file "$materializer_snapshot_apply_log" >&2
    fail 'schema-maintenance: cannot apply current schema snapshot to comparison database' "$EX_CONFIG"
  fi
  rm -f "$materializer_snapshot_apply_log" ||
    fail 'schema-maintenance: cannot remove current schema snapshot apply log' "$EX_CONFIG"

  "$ATLAS_CLI" migrate status \
    --dir "file://$materializer_snapshot_dir" \
    --url "$ATLAS_COMPARE_DEV_URL" \
    --revisions-schema atlas_revisions > "$materializer_snapshot_status" ||
    fail 'schema-maintenance: cannot inspect current schema snapshot migration status' "$EX_CONFIG"
  require_non_symlink "$materializer_snapshot_status"
  [ -f "$materializer_snapshot_status" ] ||
    fail 'schema-maintenance: current schema snapshot migration status is not regular' "$EX_CONFIG"
  [ -s "$materializer_snapshot_status" ] ||
    fail 'schema-maintenance: current schema snapshot migration status is empty' "$EX_CONFIG"
  materializer_snapshot_status_size="$(wc -c < "$materializer_snapshot_status" | tr -d ' ')"
  [ "$materializer_snapshot_status_size" -le "$REVIEW_FILE_LIMIT_BYTES" ] ||
    fail 'schema-maintenance: current schema snapshot migration status exceeds review size limit' "$EX_CONFIG"
  if ! grep -Eq '^[[:space:]]*Migration Status:[[:space:]]*OK[[:space:]]*$' "$materializer_snapshot_status" ||
    ! grep -Eq "^[[:space:]]*--[[:space:]]+Current Version:[[:space:]]*$EXPECTED_LATEST_VERSION[[:space:]]*\$" "$materializer_snapshot_status" ||
    ! grep -Eq '^[[:space:]]*--[[:space:]]+Next Version:[[:space:]]*Already at latest version[[:space:]]*$' "$materializer_snapshot_status" ||
    ! grep -Eq '^[[:space:]]*--[[:space:]]+Executed Files:[[:space:]]*1[[:space:]]*$' "$materializer_snapshot_status" ||
    ! grep -Eq '^[[:space:]]*--[[:space:]]+Pending Files:[[:space:]]*0[[:space:]]*$' "$materializer_snapshot_status"; then
    emit_review_file "$materializer_snapshot_status" >&2
    fail 'schema-maintenance: current schema snapshot migration status did not converge' "$EX_CONFIG"
  fi
  emit_review_file "$materializer_snapshot_status"

  "$ATLAS_CLI" schema inspect \
    --url "$ATLAS_COMPARE_DEV_URL" \
    --format '{{ sql . }}' > "$materializer_compare_roundtrip_schema" ||
    fail 'schema-maintenance: cannot inspect comparison schema after snapshot apply' "$EX_CONFIG"
  require_non_symlink "$materializer_compare_roundtrip_schema"
  [ -f "$materializer_compare_roundtrip_schema" ] ||
    fail 'schema-maintenance: comparison schema inspection is not regular' "$EX_CONFIG"
  [ -s "$materializer_compare_roundtrip_schema" ] ||
    fail 'schema-maintenance: comparison schema inspection is empty' "$EX_CONFIG"
  materializer_compare_roundtrip_schema_size="$(wc -c < "$materializer_compare_roundtrip_schema" | tr -d ' ')"
  [ "$materializer_compare_roundtrip_schema_size" -le "$REVIEW_FILE_LIMIT_BYTES" ] ||
    fail 'schema-maintenance: comparison schema inspection exceeds review size limit' "$EX_CONFIG"

  if [ "$materializer_verification_schema_size" -ne "$materializer_compare_roundtrip_schema_size" ]; then
    emit_review_file "$materializer_verification_schema" >&2
    emit_review_file "$materializer_compare_roundtrip_schema" >&2
    fail 'schema-maintenance: verification and snapshot roundtrip schema inspection sizes differ' "$EX_CONFIG"
  fi

  materializer_verification_to_compare_diff="$WORK_ROOT/verification-to-compare.diff"
  materializer_compare_to_verification_diff="$WORK_ROOT/compare-to-verification.diff"
  "$ATLAS_CLI" schema diff \
    --from "$VERIFICATION_SCHEMA_URL" \
    --to "$ATLAS_COMPARE_DEV_URL" \
    --dev-url "$ATLAS_DIFF_DEV_URL" \
    --format '{{ sql . }}' > "$materializer_verification_to_compare_diff" ||
    fail 'schema-maintenance: cannot diff verification schema to snapshot roundtrip schema' "$EX_CONFIG"
  require_non_symlink "$materializer_verification_to_compare_diff"
  [ -f "$materializer_verification_to_compare_diff" ] ||
    fail 'schema-maintenance: forward schema diff is not regular' "$EX_CONFIG"
  materializer_verification_to_compare_diff_size="$(wc -c < "$materializer_verification_to_compare_diff" | tr -d ' ')"
  [ "$materializer_verification_to_compare_diff_size" -le "$REVIEW_FILE_LIMIT_BYTES" ] ||
    fail 'schema-maintenance: forward schema diff exceeds review size limit' "$EX_CONFIG"
  if [ -s "$materializer_verification_to_compare_diff" ]; then
    emit_review_file "$materializer_verification_to_compare_diff" >&2
    fail 'schema-maintenance: verification schema differs from snapshot roundtrip schema' "$EX_CONFIG"
  fi

  "$ATLAS_CLI" schema diff \
    --from "$ATLAS_COMPARE_DEV_URL" \
    --to "$VERIFICATION_SCHEMA_URL" \
    --dev-url "$ATLAS_DIFF_DEV_URL" \
    --format '{{ sql . }}' > "$materializer_compare_to_verification_diff" ||
    fail 'schema-maintenance: cannot diff snapshot roundtrip schema to verification schema' "$EX_CONFIG"
  require_non_symlink "$materializer_compare_to_verification_diff"
  [ -f "$materializer_compare_to_verification_diff" ] ||
    fail 'schema-maintenance: reverse schema diff is not regular' "$EX_CONFIG"
  materializer_compare_to_verification_diff_size="$(wc -c < "$materializer_compare_to_verification_diff" | tr -d ' ')"
  [ "$materializer_compare_to_verification_diff_size" -le "$REVIEW_FILE_LIMIT_BYTES" ] ||
    fail 'schema-maintenance: reverse schema diff exceeds review size limit' "$EX_CONFIG"
  if [ -s "$materializer_compare_to_verification_diff" ]; then
    emit_review_file "$materializer_compare_to_verification_diff" >&2
    fail 'schema-maintenance: snapshot roundtrip schema differs from verification schema' "$EX_CONFIG"
  fi

  materializer_verification_schema_sha256="$(sha256sum "$materializer_verification_schema" | awk '{print $1}')"
  materializer_compare_roundtrip_schema_sha256="$(sha256sum "$materializer_compare_roundtrip_schema" | awk '{print $1}')"
  printf '%s\n' "schema-maintenance: verification_schema_sha256=$materializer_verification_schema_sha256 compare_schema_sha256=$materializer_compare_roundtrip_schema_sha256 bytes=$materializer_verification_schema_size"
}

[ "$#" -eq 1 ] || fail 'usage: mail-calendar-schema {baseline|diff|generate|hash|verify}' "$EX_USAGE"

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
  hash)
    run_hash
    ;;
  verify)
    run_verify
    ;;
  *)
    fail 'usage: mail-calendar-schema {baseline|diff|generate|hash|verify}' "$EX_USAGE"
    ;;
esac
