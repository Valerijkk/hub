#!/bin/sh
set -eu

[ "$#" -eq 0 ] || exit 64
unset GOMAXPROCS
. /usr/local/lib/hub/backend-unit/runtime.sh
. /usr/local/lib/hub/backend-unit/plan.sh
. /usr/local/lib/hub/backend-unit/execute.sh
. /usr/local/lib/hub/backend-unit/receipt.sh
trap 'backend_unit_handle_term' TERM
backend_unit_main
