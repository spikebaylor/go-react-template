#!/usr/bin/env bash
# e2e.sh IMAGE: run the Playwright tests against IMAGE (a combined image) beside
# a Postgres, all in containers, so the only thing it needs is Docker.
#
# The browser runs in Playwright's own image, pinned to the @playwright/test
# version web/ installed, so there's no browser to install and nothing that
# differs between a laptop, an agent's sandbox and CI.
set -euo pipefail

image=$1
project=app-e2e
pw_version=$(node -p "require('./web/node_modules/@playwright/test/package.json').version")

compose() { APP_IMAGE=$image APP_PORT=0 DB_PORT=0 docker compose -p "$project" "$@"; }
cleanup() {
  status=$?
  [[ $status -eq 0 ]] || compose logs server | tail -40
  compose down -v --remove-orphans >/dev/null 2>&1
  exit $status
}
trap cleanup EXIT

compose up -d --quiet-pull
# The app's image has no shell to run a healthcheck in, so wait from outside.
for _ in $(seq 1 60); do
  if docker run --rm --network "${project}_default" curlimages/curl:8.17.0 -sf http://server:8080/api/health >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

docker run --rm --network "${project}_default" \
  --user "$(id -u):$(id -g)" -e HOME=/tmp \
  -e CI -e BASE_URL=http://server:8080 -e EXPECT_VERSION \
  -v "$PWD:/work" -w /work/web \
  "mcr.microsoft.com/playwright:v${pw_version}-noble" \
  npx playwright test
