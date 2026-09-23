import { useEffect, useRef, useState } from 'react'
import { api } from '../api'
import type { Note } from '../types'
import { EmptyState } from '../components/EmptyState'
import { Plus, Trash } from '../components/icons'

function relativeTime(iso: string): string {
  const then = new Date(iso).getTime()
  const now = Date.now()
  const diff = Math.floor((now - then) / 1000)
  if (diff < 60) return 'just now'
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  const d = new Date(iso)
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
}

export function NotesPage() {
  const [notes, setNotes] = useState<Note[]>([])
  const [activeId, setActiveId] = useState<number | null>(null)
  const [title, setTitle] = useState('')
  const [body, setBody] = useState('')
  const [dirty, setDirty] = useState(false)
  const [loading, setLoading] = useState(true)
  const [confirmDelete, setConfirmDelete] = useState<Note | null>(null)
  const [confirmDiscard, setConfirmDiscard] = useState<Note | null>(null)
  const bodyRef = useRef<HTMLTextAreaElement>(null)

  const active = notes.find((n) => n.id === activeId) ?? null

  const load = () => {
    api.notes.list().then((all) => {
      setNotes(all)
      setLoading(false)
    })
  }
  useEffect(load, [])

  useEffect(() => {
    if (active && !dirty) {
      setTitle(active.title)
      setBody(active.body)
    }
  }, [activeId])

  const selectNote = (n: Note) => {
    if (dirty && active && (title !== active.title || body !== active.body)) {
      setConfirmDiscard(n)
      return
    }
    setActiveId(n.id)
    setTitle(n.title)
    setBody(n.body)
    setDirty(false)
  }

  const createNote = () => {
    api.notes.create('Untitled', '').then((n) => {
      setNotes((prev) => [n, ...prev])
      setActiveId(n.id)
      setTitle(n.title)
      setBody(n.body)
      setDirty(false)
      setTimeout(() => bodyRef.current?.focus(), 0)
    })
  }

  const save = () => {
    if (!active) return
    const t = title.trim() || 'Untitled'
    api.notes.update(active.id, t, body).then(() => {
      setDirty(false)
      load()
    })
  }

  const handleDelete = () => {
    if (!confirmDelete) return
    api.notes.remove(confirmDelete.id).then(() => {
      setConfirmDelete(null)
      if (activeId === confirmDelete.id) {
        setActiveId(null)
        setTitle('')
        setBody('')
        setDirty(false)
      }
      load()
    })
  }

  const handleDiscard = () => {
    if (!confirmDiscard) return
    setDirty(false)
    setActiveId(confirmDiscard.id)
    setTitle(confirmDiscard.title)
    setBody(confirmDiscard.body)
    setConfirmDiscard(null)
  }

  if (loading) return <div className="loading">Loading…</div>

  return (
    <div className="notes">
      <div className="page-head">
        <h1 className="page-title">Notes</h1>
        <button className="btn" onClick={createNote}>
          <Plus size={16} /> New note
        </button>
      </div>

      <div className="notes-body">
        <div className="notes-list card">
          {notes.length === 0 ? (
            <EmptyState
              title="No notes"
              hint="Jot things down: shopping ideas, recipes, links…"
              action={
                <button className="btn" onClick={createNote}>
                  <Plus size={16} /> New note
                </button>
              }
            />
          ) : (
            <ul className="list">
              {notes.map((n) => (
                <li
                  key={n.id}
                  className={`note-item ${n.id === activeId ? 'active' : ''}`}
                  onClick={() => selectNote(n)}
                >
                  <span className="note-item-title">{n.title}</span>
                  <span className="note-item-time">{relativeTime(n.updatedAt)}</span>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className="notes-editor card">
          {!active ? (
            <EmptyState title="No note selected" hint="Pick a note on the left, or create a new one." />
          ) : (
            <>
              <div className="card-head">
                <input
                  className="note-title-input"
                  value={title}
                  onChange={(e) => {
                    setTitle(e.target.value)
                    setDirty(true)
                  }}
                  placeholder="Note title"
                />
                <div className="card-actions">
                  {dirty && <span className="dirty-badge">unsaved</span>}
                  <button className="btn sm" onClick={save} disabled={!dirty}>
                    Save
                  </button>
                  <button className="icon-btn danger" onClick={() => setConfirmDelete(active)} aria-label="Delete note">
                    <Trash size={15} />
                  </button>
                </div>
              </div>
              <textarea
                ref={bodyRef}
                className="note-body"
                value={body}
                onChange={(e) => {
                  setBody(e.target.value)
                  setDirty(true)
                }}
                onKeyDown={(e) => {
                  if ((e.metaKey || e.ctrlKey) && e.key === 's') {
                    e.preventDefault()
                    save()
                  }
                }}
                placeholder="Write something…"
              />
            </>
          )}
        </div>
      </div>

      {confirmDelete && (
        <div className="modal-backdrop" onMouseDown={() => setConfirmDelete(null)}>
          <div className="modal" role="dialog" aria-modal="true" onMouseDown={(e) => e.stopPropagation()}>
            <div className="modal-head">
              <h2>Delete this note?</h2>
            </div>
            <div className="modal-body">
              <p>
                “{confirmDelete.title}” will be permanently deleted.
              </p>
            </div>
            <div className="modal-foot">
              <button className="btn ghost" onClick={() => setConfirmDelete(null)}>
                Cancel
              </button>
              <button className="btn danger" onClick={handleDelete}>
                Delete
              </button>
            </div>
          </div>
        </div>
      )}

      {confirmDiscard && (
        <div className="modal-backdrop" onMouseDown={() => setConfirmDiscard(null)}>
          <div className="modal" role="dialog" aria-modal="true" onMouseDown={(e) => e.stopPropagation()}>
            <div className="modal-head">
              <h2>Unsaved changes</h2>
            </div>
            <div className="modal-body">
              <p>You have unsaved changes to “{active?.title}”. What do you want to do?</p>
            </div>
            <div className="modal-foot">
              <button className="btn ghost" onClick={() => setConfirmDiscard(null)}>
                Keep editing
              </button>
              <button className="btn ghost" onClick={handleDiscard}>
                Discard
              </button>
              <button
                className="btn"
                onClick={() => {
                  save()
                  setConfirmDiscard(null)
                  setDirty(false)
                }}
              >
                Save &amp; switch
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}