#!/usr/bin/env bash
# SessionStart: in a Claude Code cloud session, make the tools `just verify`
# needs available. Cloud containers ship dockerd but don't run it, and may not
# have just. Anywhere else this does nothing.
set -u

[ "${CLAUDE_CODE_REMOTE:-}" = "true" ] || exit 0

if ! command -v just >/dev/null 2>&1; then
  pip install --quiet rust-just >/dev/null 2>&1 || echo "session-start: could not install just" >&2
fi

if ! docker info >/dev/null 2>&1 && command -v dockerd >/dev/null 2>&1; then
  # setsid: its own session, so it outlives the hook.
  setsid nohup dockerd >/tmp/dockerd.log 2>&1 &
  for _ in $(seq 1 30); do
    docker info >/dev/null 2>&1 && break
    sleep 1
  done
  docker info >/dev/null 2>&1 || echo "session-start: dockerd did not start; see /tmp/dockerd.log" >&2
fi
exit 0
