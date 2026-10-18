---
status: accepted
---

# One Dockerfile builds a combined image and split api and web images

The Dockerfile has three final targets: `combined` (the Go binary with the UI
embedded through the `embedui` build tag; the default), `api` (the binary
without the UI) and `web` (the built UI behind nginx). `docker-bake.hcl` builds
all three.

## Why

Personal apps want one image on one port. Work apps often want the API and the
UI scaled and deployed separately. The code is the same either way; only the
packaging differs, so both come from one build and the choice is made at deploy
time rather than baked into the repo.

## Consequences

- The Go stages cross-compile on the build platform, so multi-arch builds need
  no emulation; keep the final stages free of `RUN`.
- E2E tests drive the combined image. The split images are built and pushed
  with it; a project that ships them should add e2e coverage for them.
- A project that only ever wants one shape can drop the other targets from the
  bake file's default group.
