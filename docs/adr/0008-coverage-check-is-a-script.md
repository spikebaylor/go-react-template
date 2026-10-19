---
status: accepted
---

# The Go coverage check is a script, not a tool

`scripts/coverage.sh` enforces the Go coverage floor (ADR-0007). It removes the
excluded paths from the coverage profile, totals the rest with Go's own `go
tool cover`, and fails below the floor, listing the least-covered functions.
It adds no dependency.

## Context

A dependency is code you vet, scan, mirror and update for as long as the
project lives. Go is this template's backend language partly because it keeps
that list short: the standard library covers most of what a service needs, a
build is a single static binary, and a short module list means short security
scans and an easy handover to an air-gapped network. Dev tools aren't exempt.
`tools/go.mod` is mirrored, scanned and vendored like the app's own `go.mod`.

The obvious tool for a coverage floor is
[go-test-coverage](https://github.com/vladopajic/go-test-coverage), and it was
tried first. It works well, but pinning it adds 21 modules to `tools/go.mod`,
including the AWS SDK, go-github and a font rasteriser. They're there for
features this template doesn't use: publishing a coverage badge to S3 or a git
branch, and drawing it. Its command-line binary imports them unconditionally,
so they come along whether or not you use those features.

## Decision

Check Go coverage with a short shell script over `go tool cover`, which ships
with Go. The floor and the exclusions are variables at the top of the script.

As a rule for this template: before adding a dev tool, weigh what it brings
in against what a few lines over Go's own toolchain would do. Add the tool
when it does real work a script can't do well, as golangci-lint and
oapi-codegen do.

## Consequences

- No new modules to vet. The check uses the same `go` the build does.
- You get one total and nothing else. The script has no per-package or
  per-file floors, no badges and no diff coverage. On failure it lists
  functions, not line ranges; `go tool cover -html=out/reports/go-coverage.out`
  shows the lines.
- The script is about 40 lines of bash and awk to maintain. It's simple, but
  it's ours.

## Choosing differently

Use a dedicated tool if a project needs what the script doesn't do:
per-package floors, a coverage badge, or failing a PR that lowers coverage on
the lines it changed. That decision is per project and doesn't affect any
other ADR. To switch to go-test-coverage:

1. `go get -modfile=tools/go.mod -tool github.com/vladopajic/go-test-coverage/v2@latest`
2. Add a `.testcoverage.yml` with `profile: out/reports/go-coverage.out`,
   `threshold.total`, and the exclusions under `exclude.paths` (the same
   regexes as the script's `exclude`).
3. In the justfile, replace the "Go coverage" step's command with
   `{{go_tool}} go-test-coverage --config .testcoverage.yml`, then delete
   `scripts/coverage.sh` and update README.md's "Test coverage" section.

The web side needs no equivalent decision: Vitest's coverage provider is its
own package and checks thresholds itself.
