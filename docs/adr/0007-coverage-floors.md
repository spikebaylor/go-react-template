---
status: accepted
---

# Coverage floors: 85% Go, 60% web, checked in `just verify`

`just verify` fails when Go statement coverage is below 85% in total, or when
any of the web's four Vitest metrics (statements, branches, functions, lines)
is below 60%. Generated, vendored and wiring code is excluded. README.md's
"Test coverage" section lists the exclusions and how to change them.

## Why

- **A floor, checked where everything else is.** Coverage is one more `just
  verify` step (ADR-0001), so an agent, a laptop and CI get the same answer.
  Go is checked by `scripts/coverage.sh` over `go tool cover` (ADR-0008); the
  web by `@vitest/coverage-v8`, pinned in `web/package.json`. No coverage
  service is involved, so the check works under Tekton and air-gapped.
- **85% for Go, not 90%.** The backend is thin handlers and SQL, tested
  against a real Postgres (ADR-0004), so high coverage is cheap. Most of what
  sits between 85% and 90% is `if err != nil { return err }` after a database
  call, and reaching it means breaking the database on purpose. A gate that
  demands it pushes people and agents toward tests that run lines without
  checking anything.
- **60% for the web.** UI tests are slower to write and more brittle than API
  tests, and the e2e suite covers the app in a real browser too. All four
  metrics are held, so weak branch coverage can't hide behind strong line
  coverage.
- **The total, measured across packages.** `go test -coverpkg=./...` counts a
  test toward every package it exercises. Without it, the server tests' use of
  the store counts for nothing and the number tracks package layout rather
  than testing.
- **Exclusions are for code we don't write or can't sensibly unit test.**
  Generated code (`internal/api`, `schema.d.ts`), vendored shadcn components,
  test scaffolding (`internal/dbtest`) and startup wiring (`cmd/`, `main.tsx`)
  would measure someone else's work or force a refactor for the metric's sake.
- **A fixed floor, not a ratchet.** A ratchet that only lets coverage rise
  needs stored state and fails a fresh copy of the template for reasons it
  can't see. A number in a config file passes in every copy, and a project
  raises it deliberately.

## Consequences

- The template ships above both floors, so a copy passes unedited.
- Dropping below a floor means adding tests. An exclusion is for another file
  of the kinds above, never for code that's merely hard to reach.
- Reports land in `out/reports/` (`go-coverage.out`, `web-coverage/`), which CI
  already uploads.
