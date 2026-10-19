import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { afterEach, beforeEach, expect, test, vi } from "vitest"

import App from "./App"
import type { Note } from "./api/client"

let notes: Note[]

// A fake of the API in api/openapi.yaml, enough for the UI's round trip.
function fakeFetch(request: Request): Response {
  const url = new URL(request.url)
  const json = (body: unknown, status = 200) =>
    new Response(JSON.stringify(body), {
      status,
      headers: { "Content-Type": "application/json" },
    })

  if (url.pathname === "/api/health")
    return json({ status: "ok", version: "1.2.3" })
  if (url.pathname === "/api/notes" && request.method === "GET")
    return json(notes)
  if (url.pathname.startsWith("/api/notes/") && request.method === "DELETE") {
    const id = Number(url.pathname.split("/").pop())
    notes = notes.filter((n) => n.id !== id)
    return new Response(null, { status: 204 })
  }
  return new Response(null, { status: 404 })
}

beforeEach(() => {
  notes = [{ id: 1, text: "existing note", createdAt: "2026-01-01T00:00:00Z" }]
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: Request) => {
      if (input.method === "POST") {
        const { text } = (await input.json()) as { text: string }
        notes = [{ id: 2, text, createdAt: "2026-01-02T00:00:00Z" }, ...notes]
        return new Response(JSON.stringify(notes[0]), {
          status: 201,
          headers: { "Content-Type": "application/json" },
        })
      }
      return fakeFetch(input)
    }),
  )
})

afterEach(() => vi.unstubAllGlobals())

function renderApp() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <QueryClientProvider client={client}>
      <App />
    </QueryClientProvider>,
  )
}

test("lists notes and shows the server version", async () => {
  renderApp()

  expect(await screen.findByText("existing note")).toBeInTheDocument()
  expect(await screen.findByText("v1.2.3")).toBeInTheDocument()
})

test("adds a note", async () => {
  renderApp()
  await screen.findByText("existing note")

  await userEvent.type(screen.getByLabelText("New note"), "brand new")
  await userEvent.click(screen.getByRole("button", { name: "Add" }))

  expect(await screen.findByText("brand new")).toBeInTheDocument()
  expect(screen.getByLabelText("New note")).toHaveValue("")
})

test("deletes a note", async () => {
  renderApp()

  await userEvent.click(
    await screen.findByRole("button", { name: 'Delete "existing note"' }),
  )

  expect(await screen.findByText("No notes yet.")).toBeInTheDocument()
  expect(screen.queryByText("existing note")).not.toBeInTheDocument()
})

test("shows why the server rejected a note", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: Request) =>
      input.method === "POST"
        ? new Response(
            JSON.stringify({ message: "text must be 1 to 500 characters" }),
            {
              status: 400,
              headers: { "Content-Type": "application/problem+json" },
            },
          )
        : fakeFetch(input),
    ),
  )
  renderApp()
  await screen.findByText("existing note")

  await userEvent.type(screen.getByLabelText("New note"), "rejected")
  await userEvent.click(screen.getByRole("button", { name: "Add" }))

  expect(
    await screen.findByText("text must be 1 to 500 characters"),
  ).toBeInTheDocument()
  expect(screen.getByLabelText("New note")).toHaveValue("rejected")
})

test("says when notes can't be loaded", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => new Response(null, { status: 500 })),
  )
  renderApp()

  expect(await screen.findByText("Could not load notes.")).toBeInTheDocument()
})
