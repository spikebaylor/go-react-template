---
status: accepted
---

# The justfile is the contract; CI only calls it

Every check and build is a `just` recipe, and the CI workflows do nothing but set
up tools, log in and call recipes. `just verify` is the same command for an agent
finishing a change, a person before pushing, and CI before building images.

## Why

- **One answer.** When the agent, the laptop and CI run different commands, they
  disagree, and the agent can't tell "done" from "done until CI runs". A single
  `verify` makes "it passed locally" mean "it will pass in CI".
- **Portable.** These repos also have to work under Tekton (and whatever comes
  next), where the pipeline is owned by someone else. A pipeline that calls
  `just verify`, `just publish` and `just e2e-image` needs no knowledge of Go,
  npm or Playwright. Moving CI systems means rewriting a thin YAML file, not the
  logic.
- **Bounded output for agents.** `scripts/step.sh` prints one line per step and
  only the failing step's log tail, so an agent's context isn't flooded with
  passing output. Under `CI=true` it streams everything instead.

## Consequences

- Logic belongs in a recipe or a `scripts/` file, never inline in workflow YAML.
  Workflow-only concerns (checkout, credentials, the GitHub step summary) stay
  in the YAML.
- Dev tools are pinned in `tools/go.mod` and run with `go tool`, so the only
  prerequisites are Go, Node, Docker and just.
