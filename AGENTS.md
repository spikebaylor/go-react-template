# AGENTS.md

A Go API and a React UI, built into container images. `README.md` covers what it
is and how it ships; this file is how to work in it. The decisions behind its
shape are in `docs/adr/`: read the relevant one before changing CI, versioning,
the images, the test database or `deploy/`.

<!-- template-only: delete from here to "end template-only" in a project made from this template -->
## Working on the template itself

This repo is a template: new apps start as a copy of it, and the notes app is a
placeholder that exists to exercise the pipeline. A change here is a change to
every future project's starting point, so it holds to a stricter bar:

- **A copy works unedited.** A repo made from the template must pass CI and
  publish images with no edits; derive names from the repository
  (`github.repository`), never hardcode this one. Anything a new project must
  change goes in README.md's "Starting a new app" checklist.
- **The decisions in `docs/adr/` are the point.** Build once and promote, the
  justfile as the only CI interface, contract-first API, `TEST_DATABASE_URL`,
  logical image names: these are what make the template portable to other CI
  systems (Tekton at work) and air-gapped handovers. Changing one is a new ADR,
  not a refactor.
- **Keep the example small.** It earns its place by exercising a part of the
  pipeline (database, generated client, UI, e2e); features that exercise
  nothing new belong in real projects.
- **Workflow changes are proven by running them.** Lint them
  (`go run github.com/rhysd/actionlint/cmd/actionlint@latest`) and say in the
  PR which run demonstrates the change.
<!-- end template-only -->

## Done means `just verify` passes

`just verify` runs every check CI runs before building images, with one line per
step and the tail of the log on failure. Run it before calling any change done,
and fix what it reports rather than working around it. For a change the UI or
an image could break, also run `just e2e`: it builds the combined image and
drives it in a browser. Both need Docker; the Go tests start their own Postgres,
or use `TEST_DATABASE_URL` when it's set.

`just` lists the other recipes. `just fix` formats and applies lint autofixes,
so formatting is never worth hand-editing.

## Contract first

`api/openapi.yaml` is the API. The Go handlers (`internal/api`) and the
TypeScript client types (`web/src/api/schema.d.ts`) are generated from it by
`just generate`, and verify fails when they're stale. To change the API: edit
the spec, run `just generate`, then make the Go code compile against the new
`api.StrictServerInterface` and the UI typecheck against the new client.
Generated files change only through `just generate`.

## Conventions the tests hold you to

- **Settings**: `internal/config` reads every setting from the environment. Each
  needs a row in README.md's Configuration table, and a required one needs an
  entry in `deploy/base/deployment.yaml`; `TestSettingsAreDocumentedAndDeployed`
  checks both.
- **Database**: all SQL lives in `internal/store`. A schema change is a new goose
  migration in `internal/store/migrations`; migrations that have shipped are
  history and stay as they are. Tests get a fresh migrated database from
  `dbtest.New(t)`, with `dbtest.Main` in the package's `TestMain`.
- **UI**: shadcn/ui components in `web/src/components/ui` (add more with
  `npx shadcn add <name>` from `web/`), styled with the semantic tokens in
  `web/src/index.css` (`bg-background`, `text-muted-foreground`, ...), so both
  themes follow. Data goes through `web/src/api/client.ts` and TanStack Query.
- **Coverage**: `just verify` fails below 85% Go statements
  (`.testcoverage.yml`) or 60% of any web metric (`web/vite.config.ts`). Meet
  it with tests that check what your change does. Don't add an exclusion or
  lower a floor to get green; README.md's "Test coverage" says what belongs in
  the exclusions.
- **Deploy shape**: `deploy/base` is what the app needs to run, changed in the
  same PR as the code that needs it. Its images are logical names (`app`) that a
  deploying overlay binds to a registry and digest.

## Commits and versions

PRs are squash-merged, so the PR title becomes the commit and decides the next
version: `feat:` minor, `fix:` patch, `feat!:` or `BREAKING CHANGE:` major;
`docs:`, `chore:`, `test:`, `refactor:`, `ci:` release nothing. Put an issue
number in the scope when there is one: `feat(12): add note search`.
