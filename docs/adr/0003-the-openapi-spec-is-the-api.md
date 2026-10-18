---
status: accepted
---

# The OpenAPI spec is the API; both sides are generated from it

`api/openapi.yaml` defines the API. `just generate` turns it into the Go server
interface (`internal/api`, oapi-codegen strict server on stdlib `net/http`) and
the TypeScript client types (`web/src/api/schema.d.ts`, openapi-typescript).
Verify fails when the generated code is stale.

## Why

An agent changing one side of a hand-written API rarely updates the other, and
nothing notices until runtime. Generated from one spec, a drift is a compile
error on whichever side didn't follow: the Go handlers must satisfy
`api.StrictServerInterface`, and the UI must typecheck against the client. It
also keeps working when the backend and frontend are split into separate
images, or separate teams.

## Consequences

- API changes start in the spec. Generated files change only through
  `just generate`.
- The spec is the place to read the API, for people and agents alike.
