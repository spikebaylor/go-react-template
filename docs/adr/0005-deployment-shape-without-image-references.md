---
status: accepted
---

# deploy/base names images logically; whoever deploys binds them

`deploy/base` is a kustomize base holding what the app needs to run: ports,
probes, settings, security context. Its images are logical names (`app`). A
deploying overlay binds each name to a registry and digest with an `images:`
block, as `deploy/example` shows.

## Why

At work the source repo's manifests named the image they shipped with. Every
build then had to edit the manifests, a trivial manifest change needed a new
build, and a release script had to keep two places in agreement. With logical
names, the manifests never depend on their own build, so:

- what the code needs (a new setting, port or probe) changes in the same PR as
  the code, here;
- what the environment decides (replicas, hosts, the exact image digest) lives
  with whoever deploys, e.g. a release or GitOps repo, and changes without a
  build.

## Consequences

- No registry, tag or digest in `deploy/base`.
- Required settings appear in `deploy/base/deployment.yaml`;
  `TestSettingsAreDocumentedAndDeployed` checks it, and verify renders the
  example overlay.
