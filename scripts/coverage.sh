#!/usr/bin/env bash
# coverage.sh PROFILE: fail when Go statement coverage, outside the excluded
# paths, is below the floor.
#
# It uses only Go's own `go tool cover`, so it adds no dependency to vet, scan
# or mirror; docs/adr/0008-coverage-check-is-a-script.md says why, and what a
# dedicated tool would add. README.md's "Test coverage" section explains the
# floor and the exclusions.
set -euo pipefail

# The floor, as a percentage of statements across all non-excluded code.
floor=85

# Paths left out of the total: an extended regex matched against each file's
# path from the repo root.
#   internal/api/     generated from api/openapi.yaml by `just generate`
#   internal/dbtest/  test scaffolding: it only ever runs inside tests
#   cmd/              startup wiring; `just e2e` runs it in the real image
exclude='^(internal/api|internal/dbtest|cmd)/'

profile=$1
module="$(go list -m)/"
filtered=$(mktemp)
trap 'rm -f "$filtered"' EXIT

# Keep the "mode:" line, then every block whose file isn't excluded.
awk -v module="$module" -v exclude="$exclude" '
  NR == 1 { print; next }
  {
    path = $0
    if (index(path, module) == 1) path = substr(path, length(module) + 1)
    if (path !~ exclude) print
  }
' "$profile" >"$filtered"

report=$(go tool cover -func="$filtered")
total=$(awk '/^total:/ { sub(/%/, "", $NF); print $NF }' <<<"$report")

if awk -v t="$total" -v f="$floor" 'BEGIN { exit !(t < f) }'; then
  echo "Go coverage is ${total}%, below the ${floor}% floor."
  echo "Least-covered functions (profile: $profile):"
  grep -v '^total:' <<<"$report" | awk '$NF != "100.0%"' | sort -t$'\t' -k3 -n | sed "s|^$module||"
  exit 1
fi
echo "Go coverage is ${total}% (floor ${floor}%)."
