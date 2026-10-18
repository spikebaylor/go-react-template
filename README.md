# go-react-template

A starting point for web apps and services: a Go API, a React + shadcn/ui UI,
Postgres, and a CI pipeline that ships container images. It's set up so an AI
agent can build, test and verify changes on its own, and the same checks run
for agents, people and CI.

The example app is a small notes list. It exists to exercise every part of the
pipeline; replace it with yours.

## What's in it

| Path | What it is |
|---|---|
| `api/openapi.yaml` | The API contract. Go server code and TypeScript types are generated from it. |
| `cmd/app`, `internal/` | The Go server: stdlib `net/http`, `pgx`, `goose` migrations, `slog` JSON logs. |
| `web/` | Vite + React + TypeScript, Tailwind, shadcn/ui, TanStack Query, Vitest, Playwright. |
| `Dockerfile`, `docker-bake.hcl` | Three images from one Dockerfile: `app` (UI embedded, one port), `app-api`, `app-web`. |
| `deploy/base` | A kustomize base with logical image names, for whoever deploys it to bind. |
| `justfile` | Every check and build. CI calls these recipes and nothing else. |
| `tools/go.mod` | Pinned dev tools (golangci-lint, oapi-codegen, svu, gotestsum, kustomize), run with `go tool`. |
| `AGENTS.md` | Instructions for AI agents (`CLAUDE.md` imports it). |
| `docs/adr/` | Why it's built this way: the decisions to keep, or change deliberately. |

## Prerequisites

Go (version in `go.mod`), Node (version in `web/.nvmrc`), Docker, and
[just](https://just.systems). The Go dev tools need nothing installed: `go tool`
builds the pinned versions on first use.

## Day to day

```sh
just verify      # every check CI runs before building images
just e2e         # build the combined image and test it in a browser
just fix         # format and autofix
just generate    # after editing api/openapi.yaml
just up          # build and run the combined image at http://localhost:8080
just dev-db      # dev loop: Postgres on :5432 ...
just dev-api     # ... the API on :8080 ...
just dev-web     # ... and the UI with hot reload on :5173
```

## CI at a glance

| Workflow | Runs when | Steps, in order (each needs the one before) |
|---|---|---|
| **CI** (`ci.yml`) | A PR is opened or pushed to | PR title is a conventional commit → `verify` → `image`: build amd64 + arm64, push as `x.y.z-pr.N.run` → `e2e` against the pushed image → tag it `pr-N` |
| **CI** (`ci.yml`) | A push to `main` (a merge) | `verify` → `image`: push as `x.y.z-dev.run` → `e2e` → tag it `dev` and `sha-<commit>` |
| **Release** (`release.yml`) | By hand: Actions → Release | Work out the version → find the tested `sha-` image for `main`'s head → show the notes (a *dry run* stops here) → retag as `x.y.z`, `x.y`, `latest` → tag git `vx.y.z` and publish a GitHub Release |
| **Clean up images** (`cleanup.yml`) | A PR closes | For each image (`app`, `-api`, `-web`): delete that PR's tags |
| **Clean up images** (`cleanup.yml`) | Weekly, or by hand | For each image: keep the 20 newest `main` builds and delete the rest; by hand with a PR number, redo that PR's cleanup instead |

`verify` is `just verify`, the same command you run locally. A PR from a fork
runs only the title check and `verify`, since its token can't push images.

## Builds and releases

Every PR and every push to `main` produces versioned images; only releases
touch git.

| When | Image tags | Git |
|---|---|---|
| PR #17, CI run 3 | `1.3.0-pr.17.3`, then `pr-17` once e2e passes | nothing |
| Push to `main`, run 42 | `1.3.0-dev.42`, then `dev` and `sha-<commit>` once e2e passes | nothing |
| **Release** (run by hand) | the tested `sha-` images retagged `1.3.0`, `1.3`, `latest` | tag `v1.3.0` and a GitHub Release |

- The version is the next release's, worked out from the conventional-commit
  PR titles since the last `v*` tag (`just version`), so a pre-release always
  shows what it would ship as.
- Images are built once. E2E tests run against the pushed image, and the
  floating tags (`pr-17`, `dev`, `sha-…`) are added only after they pass.
  Release promotes those exact images without rebuilding.
- To try a PR: `APP_IMAGE=ghcr.io/<owner>/<repo>:pr-17 docker compose up`.
- To release: Actions → **Release** → Run workflow. Tick *dry run* to see the
  version and notes first. It refuses when nothing since the last release is a
  `feat`, `fix` or breaking change.
- Cleanup never deletes a release: it removes tags, and an image goes only when
  all its tags have.

## Starting a new app from this template

1. **Use this template** on GitHub, then clone it.
2. Rename the Go module: replace `github.com/spikebaylor/go-react-template`
   everywhere (`grep -rl` it), then `go mod tidy`.
3. Delete the "Working on the template itself" section of `AGENTS.md` (between
   the `template-only` markers); it's for maintaining the template, not your app.
4. In the repo's settings:
   - **General → Pull Requests**: allow squash merging only, with the default
     message set to *Pull request title*.
   - **Rules → Rulesets** for `main`: require a pull request and the status
     checks `verify`, `e2e` and `PR title is a conventional commit`.
5. Open a PR. CI publishes images to `ghcr.io/<owner>/<repo>` (and `-api`,
   `-web`) with no further setup. Packages pushed from a workflow inherit the
   repository's access; make them public from the package settings if you want
   to pull without logging in.

Image names, Kubernetes names and the app name all default to `app`; change
them in `docker-bake.hcl`, `compose.yaml` and `deploy/` when it matters.

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `APP_DATABASE_URL` | (required) | Postgres connection string, e.g. `postgres://user:pass@host:5432/db?sslmode=disable`. |
| `APP_ADDR` | `:8080` | Address the HTTP server listens on. |

Migrations run on startup under a Postgres advisory lock, so replicas starting
together apply each one once.
