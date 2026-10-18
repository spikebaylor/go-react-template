---
status: accepted
---

# Tests get their own Postgres, from Testcontainers or from TEST_DATABASE_URL

Database tests call `dbtest.Main` from `TestMain` and `dbtest.New(t)` per test.
With `TEST_DATABASE_URL` unset, `dbtest` starts `postgres:18` with
Testcontainers; with it set, it uses that server. Either way it migrates a
template database once and hands each test a fresh copy.

## Why

- **Real Postgres, no mocks.** The store's SQL is the thing most likely to be
  wrong; testing it against anything else proves little.
- **No setup step.** An agent in a fresh container runs `just verify` and gets a
  database without being told to start one.
- **Runs where Docker doesn't.** Tekton pods have no Docker daemon, so
  Testcontainers can't run there; a Postgres sidecar plus `TEST_DATABASE_URL`
  can. Template names are per process, so packages running in parallel can
  share one server.
- **Fails, never skips.** With neither Docker nor `TEST_DATABASE_URL`, the tests
  fail and say why. A green run that tested nothing is worse than a red one.
