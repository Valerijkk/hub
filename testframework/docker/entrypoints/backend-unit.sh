#!/bin/sh
set -eu

[ "$#" -eq 0 ] || exit 64
unset GOMAXPROCS
. /usr/local/lib/hub/backend-unit/runtime.sh
. /usr/local/lib/hub/backend-unit/plan.sh
. /usr/local/lib/hub/backend-unit/execute.sh
. /usr/local/lib/hub/backend-unit/receipt.sh
backend_unit_main
