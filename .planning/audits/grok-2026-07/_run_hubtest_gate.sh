set -o pipefail
apt-get update -qq
apt-get install -y -qq gcc libc6-dev libsqlite3-dev >/dev/null
go test -tags hubtest -race -count=1 -timeout=900s ./testframework/... 2>&1 | tee /artifacts/w20-gate-hubtest.txt
ec=${PIPESTATUS[0]}
echo EXIT_CODE=$ec | tee -a /artifacts/w20-gate-hubtest.txt
exit $ec