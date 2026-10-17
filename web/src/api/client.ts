import createClient from "openapi-fetch"

import type { components, paths } from "./schema"

// Typed from api/openapi.yaml: a path, method or field the contract doesn't
// have is a compile error. Regenerate schema.d.ts with `just generate`.
export const api = createClient<paths>({
  baseUrl: window.location.origin,
  // Looked up per call rather than captured once, so tests can stub fetch.
  fetch: (request) => globalThis.fetch(request),
})

export type Note = components["schemas"]["Note"]
