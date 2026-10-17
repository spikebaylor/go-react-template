#!/usr/bin/env bash
# version.sh [PRERELEASE]: the version this commit would ship as.
#
#   version.sh            the next release, e.g. 1.3.0. Fails if nothing since
#                         the last release warrants one (only docs:, chore:, ...).
#   version.sh dev.42     a pre-release of the next release, e.g. 1.3.0-dev.42.
#                         It always sorts after the last release, even when only
#                         docs changed since.
#
# The version comes from conventional-commit subjects since the last vX.Y.Z tag
# on this branch: feat: bumps minor, fix: bumps patch, a ! or BREAKING CHANGE:
# bumps major. Only releases are tagged in git; pre-releases live in the registry.
set -euo pipefail

pattern='v[0-9]*.[0-9]*.[0-9]*'
svu() {
  # svu fails on a pattern nothing matches, so before the first release let it
  # start from its default of v0.0.0.
  if [[ -n "$(git tag --merged HEAD --list "$pattern")" ]]; then
    set -- "$@" --tag.pattern "$pattern"
  fi
  go tool -modfile=tools/go.mod svu "$@" --tag.mode current
}

if [[ $# -eq 0 ]]; then
  current=$(svu current)
  next=$(svu next)
  if [[ "$next" == "$current" ]]; then
    echo "version.sh: nothing to release since $current: no feat:, fix: or breaking commits" >&2
    exit 1
  fi
  echo "${next#v}"
else
  next=$(svu next --always)
  echo "${next#v}-$1"
fi
