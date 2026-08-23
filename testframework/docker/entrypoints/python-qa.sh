#!/bin/sh
set -eu

[ "$#" -eq 1 ] || exit 64

cd /opt/hub/testframework
export PYTHONSAFEPATH=1

case "$1" in
  contract)
    exec /opt/hub/.venv/bin/pytest selftests --strict-markers
    ;;
  api)
    exec /opt/hub/.venv/bin/pytest -p hub_testkit.pytest_plugin.plugin tests/api --hub-case-root cases --hub-case-schema schemas/test-case.schema.json --hub-marker-policy policies/markers.toml --hub-redaction-policy policies/redaction.toml --strict-markers
    ;;
  integration)
    exec /opt/hub/.venv/bin/pytest -p hub_testkit.pytest_plugin.plugin tests/integration --hub-case-root cases --hub-case-schema schemas/test-case.schema.json --hub-marker-policy policies/markers.toml --hub-redaction-policy policies/redaction.toml --strict-markers
    ;;
  e2e-smoke)
    exec /opt/hub/.venv/bin/pytest -p hub_testkit.pytest_plugin.plugin tests/e2e -m smoke --hub-case-root cases --hub-case-schema schemas/test-case.schema.json --hub-marker-policy policies/markers.toml --hub-redaction-policy policies/redaction.toml --strict-markers
    ;;
  e2e-full)
    exec /opt/hub/.venv/bin/pytest -p hub_testkit.pytest_plugin.plugin tests/e2e --hub-case-root cases --hub-case-schema schemas/test-case.schema.json --hub-marker-policy policies/markers.toml --hub-redaction-policy policies/redaction.toml --strict-markers
    ;;
  load)
    exec /opt/hub/.venv/bin/locust -f load/locustfile.py --headless
    ;;
  *)
    exit 64
    ;;
esac
