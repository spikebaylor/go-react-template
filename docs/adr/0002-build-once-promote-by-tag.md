---
status: accepted
---

# Images are built once and promoted by tag; only releases touch git

Every PR and every push to `main` builds and pushes images, versioned as the
release they'd become (`1.3.0-pr.17.3`, `1.3.0-dev.42`). The end-to-end tests
then run against the *pushed* image, and only when they pass does it get its
floating tags: `pr-17`, or `dev` and `sha-<commit>`. Release retags the tested
`sha-` images as `1.3.0`, `1.3`, `latest`, then tags git and publishes a GitHub
Release. Nothing is rebuilt.

## Why

- **What ships is what was tested.** A rebuild at release time picks up new base
  images and dependencies, so the release is bytes nobody tested. At work the
  release is an air-gapped handover that can sit in approval for weeks; it has
  to be exactly the artifact that passed.
- **A tag means something.** `sha-<commit>` exists only if that image passed e2e,
  so release can find "the tested image for this commit" by name and refuse
  when there isn't one.
- **Pre-releases stay out of git.** Momentum's dev builds put a floating GitHub
  Release in git and committed changelogs back to `main`; both were clutter and
  needed branch-protection bypasses. Here pre-releases live only in the
  registry, and cleanup removes them.
- **One version calculator.** `scripts/version.sh` (svu over conventional commits)
  computes every version, pre-release or release, so they can't disagree. PR
  titles are linted because squash merges make them the commit subjects.

## Consequences

- Tag after the thing the tag claims: git tags only after images are promoted,
  floating image tags only after e2e passes.
- Cleanup must never delete a release. It deletes tags, and an image goes only
  when all its tags are gone, so a released image survives losing its `dev` tag.
- For work, release branches (`release/x.y`) will be needed for patching a
  handed-over version; `version.sh` already reads tags from the current branch.
