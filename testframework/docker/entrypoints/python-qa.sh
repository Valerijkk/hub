#!/bin/sh
set -eu

[ "$#" -eq 1 ] || exit 64

cd /opt/hub/testframework
export PYTHONSAFEPATH=1

case "$1" in
  contract)
    exec /opt/hub/.venv/bin/pytest selftests tests/contract --hub-case-root cases --strict-markers
    ;;
  api)
    exec /opt/hub/.venv/bin/pytest tests/api --hub-case-root cases --strict-markers
    ;;
  integration)
    exec /opt/hub/.venv/bin/pytest tests/integration --hub-case-root cases --strict-markers
    ;;
  e2e-smoke)
    exec /opt/hub/.venv/bin/pytest tests/e2e -m smoke --hub-case-root cases --strict-markers
    ;;
  e2e-full)
    exec /opt/hub/.venv/bin/pytest tests/e2e --hub-case-root cases --strict-markers
    ;;
  load)
    exec /opt/hub/.venv/bin/locust -f load/locustfile.py --headless
    ;;
  *)
    exit 64
    ;;
esac
