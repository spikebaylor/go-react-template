import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Trash2 } from "lucide-react"
import { useState, type FormEvent } from "react"

import { api } from "@/api/client"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"

export default function App() {
  const queryClient = useQueryClient()
  const [text, setText] = useState("")

  const health = useQuery({
    queryKey: ["health"],
    queryFn: async () => (await api.GET("/api/health")).data,
  })
  const notes = useQuery({
    queryKey: ["notes"],
    queryFn: async () => {
      const { data, error } = await api.GET("/api/notes")
      if (error) throw new Error("could not load notes")
      return data
    },
  })

  const refresh = () => queryClient.invalidateQueries({ queryKey: ["notes"] })
  const create = useMutation({
    mutationFn: async (text: string) => {
      const { error } = await api.POST("/api/notes", { body: { text } })
      if (error) throw new Error(error.message)
    },
    onSuccess: () => {
      setText("")
      return refresh()
    },
  })
  const remove = useMutation({
    mutationFn: (id: number) =>
      api.DELETE("/api/notes/{id}", { params: { path: { id } } }),
    onSuccess: refresh,
  })

  function submit(e: FormEvent) {
    e.preventDefault()
    if (text.trim()) create.mutate(text)
  }

  return (
    <main className="mx-auto flex min-h-screen max-w-xl flex-col gap-6 p-6">
      <Card>
        <CardHeader>
          <CardTitle>Notes</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-4">
          <form onSubmit={submit} className="flex gap-2">
            <Input
              aria-label="New note"
              placeholder="Write a note"
              value={text}
              maxLength={500}
              onChange={(e) => setText(e.target.value)}
            />
            <Button type="submit" disabled={create.isPending}>
              Add
            </Button>
          </form>
          {create.error && (
            <p className="text-destructive text-sm">{create.error.message}</p>
          )}
          {notes.isError && (
            <p className="text-destructive text-sm">Could not load notes.</p>
          )}
          <ul className="flex flex-col gap-2">
            {notes.data?.map((note) => (
              <li
                key={note.id}
                className="flex items-center justify-between rounded-md border px-3 py-2"
              >
                <span>{note.text}</span>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Delete "${note.text}"`}
                  onClick={() => remove.mutate(note.id)}
                >
                  <Trash2 />
                </Button>
              </li>
            ))}
          </ul>
          {notes.data?.length === 0 && (
            <p className="text-muted-foreground text-sm">No notes yet.</p>
          )}
        </CardContent>
      </Card>
      <footer className="text-muted-foreground text-center text-xs">
        {health.data ? `v${health.data.version}` : " "}
      </footer>
    </main>
  )
}
