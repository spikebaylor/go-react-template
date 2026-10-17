#!/usr/bin/env bash
# step.sh NAME COMMAND...: run one step of a check with bounded output.
#
# Locally it prints one line per step and, on failure, the last 60 lines of
# that step's log, so an agent's context isn't flooded with passing output.
# Under CI (CI=true) it streams everything, since the log is the record there.
set -uo pipefail

name=$1
shift

if [[ "${CI:-}" == "true" ]]; then
  echo "::group::$name"
  "$@"
  status=$?
  echo "::endgroup::"
  [[ $status -eq 0 ]] || echo "✗ $name failed (exit $status)"
  exit $status
fi

mkdir -p out/logs
log="out/logs/$(echo "$name" | tr -c 'a-zA-Z0-9\n' '-').log"
printf '▶ %s...' "$name"
start=$SECONDS
"$@" >"$log" 2>&1
status=$?
if [[ $status -eq 0 ]]; then
  printf ' ✓ (%ds)\n' $((SECONDS - start))
  exit 0
fi
printf ' ✗ (%ds)\n\n===== %s failed: last 60 lines (full log: %s) =====\n' $((SECONDS - start)) "$name" "$log"
tail -60 "$log"
exit $status
