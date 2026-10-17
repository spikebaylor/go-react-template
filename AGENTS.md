# AGENTS.md

A Go API and a React UI, built into container images. `README.md` covers what it
is and how it ships; this file is how to work in it.

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
- **Deploy shape**: `deploy/base` is what the app needs to run, changed in the
  same PR as the code that needs it. Its images are logical names (`app`) that a
  deploying overlay binds to a registry and digest.

## Commits and versions

PRs are squash-merged, so the PR title becomes the commit and decides the next
version: `feat:` minor, `fix:` patch, `feat!:` or `BREAKING CHANGE:` major;
`docs:`, `chore:`, `test:`, `refactor:`, `ci:` release nothing. Put an issue
number in the scope when there is one: `feat(12): add note search`.
