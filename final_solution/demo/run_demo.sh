#!/usr/bin/env bash
# Load Sphynx demo runner. Boots 5 backends + the balancer.
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

LOG="$ROOT/demo/logs"
mkdir -p "$LOG"

echo ">> Checking Redis..."
if ! redis-cli ping >/dev/null 2>&1; then
  echo "   Redis not reachable on 127.0.0.1:6379."
  echo "   Run:  brew services start redis"
  exit 1
fi
echo "   Redis OK."

echo ">> go mod tidy..."
go mod tidy

echo ">> Booting 5 backend servers on ports 5001-5005..."
for pair in "5001 Server1" "5002 Server2" "5003 Server3" "5004 Server4" "5005 Server5"; do
  port=${pair% *}; name=${pair#* }
  go run demo/backends.go "$port" "$name" > "$LOG/backend-$port.log" 2>&1 &
  echo "   $name on :$port (pid $!)"
done
sleep 1.5

echo ">> Starting Load Sphynx (main.go)..."
echo "   Virtual service 1: http://localhost:8001  (round-robin across :5001, :5002)"
echo "   Virtual service 2: http://localhost:8443  (weighted-RR across :5003/4/5)"
echo "   Dashboard UI:      http://localhost:8080/  (user: bal / pass: 2fourall)"
echo "   Admin API:         http://localhost:8080/access/vs"
echo "----------------------------------------------------------------------------"
# run in foreground so Ctrl-C kills everything
trap 'echo; echo ">> Cleaning up..."; kill $(jobs -p) 2>/dev/null; exit 0' INT TERM
go run main.go
