# The repo's contract: every check and build an agent, a person or CI runs goes
# through these recipes, so all three get the same answer. CI's YAML only calls
# them. Run `just` to list them.

set shell := ["bash", "-euo", "pipefail", "-c"]

go_tool := "go tool -modfile=tools/go.mod"
step := "scripts/step.sh"

# List the recipes
default:
    @just --list --unsorted

# Everything CI checks before building an image. Run it before calling work done.
verify: _web-deps
    @{{step}} "Generated code is current" just _check-generated
    @{{step}} "Go lint" {{go_tool}} golangci-lint run ./...
    @{{step}} "Go tests" {{go_tool}} gotestsum --junitfile out/reports/go-junit.xml -- ./...
    @{{step}} "Kubernetes manifests render" {{go_tool}} kustomize build deploy/example -o /dev/null
    @{{step}} "Web typecheck" npm --prefix web run typecheck
    @{{step}} "Web lint" npm --prefix web run lint
    @{{step}} "Web format" npm --prefix web run format:check
    @{{step}} "Web tests" npm --prefix web test
    @{{step}} "Web build" npm --prefix web run build
    @echo "✓ verify passed"

# verify, then build the images and run the end-to-end tests against them
ci: verify e2e

# Format everything and apply automatic lint fixes
fix: _web-deps
    {{go_tool}} golangci-lint fmt ./...
    {{go_tool}} golangci-lint run --fix ./... || true
    cd web && npx oxlint --fix || true
    cd web && npx prettier --write . >/dev/null
    @echo "fixed what could be fixed; run just verify for what's left"

# Regenerate the Go server and the TypeScript client from api/openapi.yaml
generate: _web-deps
    {{go_tool}} oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml
    npm --prefix web run --silent generate

# Build images for this machine: app (combined), api, web; or name some
image *targets:
    docker buildx bake --load {{targets}}

# Add tags to images already in the registry, without rebuilding: every image
# (app, -api, -web) tagged FROM gets each of the comma-separated TAGS
retag from tags:
    scripts/retag.sh {{from}} {{tags}}

# Build and push the images for every platform. CI sets IMAGE, VERSION, TAGS,
# PLATFORMS and SOURCE (see docker-bake.hcl) and logs in to the registry first.
publish:
    docker buildx bake --push

# Build the combined image and run the end-to-end tests against it
e2e: _web-deps (image "app")
    @{{step}} "End-to-end tests" scripts/e2e.sh app:dev

# Run the end-to-end tests against an existing image, e.g. one CI pushed
e2e-image image: _web-deps
    @{{step}} "End-to-end tests" scripts/e2e.sh {{image}}

# The version this commit would release as; with an id, a pre-release of it
version prerelease="":
    @scripts/version.sh {{prerelease}}

# Build the combined image and run it with a Postgres at http://localhost:8080
up: (image "app")
    docker compose up -d --wait db
    docker compose up -d server
    @for _ in $(seq 1 30); do curl -sf localhost:8080/api/health >/dev/null && break; sleep 1; done
    @curl -sf localhost:8080/api/health && echo && echo "running at http://localhost:8080; just down to stop"

# Stop what `just up` or `just dev-db` started (add -v to wipe the database)
down *args:
    docker compose down {{args}}

# Development: a Postgres on localhost:5432 for dev-api
dev-db:
    docker compose up -d --wait db

# Development: the Go API on :8080 against dev-db
dev-api:
    APP_DATABASE_URL="postgres://app:app@localhost:5432/app?sslmode=disable" go run ./cmd/app

# Development: the UI with hot reload on :5173, proxying /api to dev-api
dev-web: _web-deps
    npm --prefix web run dev

_check-generated:
    just generate
    @git diff --exit-code --stat -- internal/api web/src/api || (echo "generated code was stale: commit the output of just generate" && exit 1)

_web-deps:
    @[ web/node_modules/.package-lock.json -nt web/package-lock.json ] || npm --prefix web ci --no-audit --no-fund
