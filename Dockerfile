# syntax=docker/dockerfile:1

# One Dockerfile, three images (docker-bake.hcl builds them):
#   combined  the Go binary with the UI embedded: one image, one port (default)
#   api       the Go binary alone, for scaling the API separately
#   web       the UI's static files behind nginx
# Keep GO_VERSION and NODE_VERSION in step with go.mod and web/package.json.
ARG GO_VERSION=1.27
ARG NODE_VERSION=24

FROM --platform=$BUILDPLATFORM node:${NODE_VERSION}-slim AS web-build
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY web/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS go-base
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY cmd/ cmd/
COPY internal/ internal/

# Cross-compiled on the build platform, so a multi-arch build needs no emulation.
FROM go-base AS api-build
ARG TARGETOS TARGETARCH VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
    -ldflags "-s -w -X main.version=$VERSION" -o /out/app ./cmd/app

FROM go-base AS combined-build
ARG TARGETOS TARGETARCH VERSION=dev
COPY --from=web-build /src/web/dist/ internal/webui/dist/
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -tags embedui \
    -ldflags "-s -w -X main.version=$VERSION" -o /out/app ./cmd/app

FROM gcr.io/distroless/static-debian13:nonroot AS api
COPY --from=api-build /out/app /app
EXPOSE 8080
ENTRYPOINT ["/app"]

FROM nginxinc/nginx-unprivileged:1-alpine AS web
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=web-build /src/web/dist/ /usr/share/nginx/html/
EXPOSE 8080

FROM gcr.io/distroless/static-debian13:nonroot AS combined
COPY --from=combined-build /out/app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
