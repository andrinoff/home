import { useEffect, useMemo, useState } from 'react'
import { api, formatDateTime, localDayKey } from '../api'
import type { HomeEvent } from '../types'
import { EmptyState } from '../components/EmptyState'
import { Modal } from '../components/Modal'
import { EventForm, type EventDraft } from '../components/EventForm'
import { Clock, Pencil, Plus, Trash } from '../components/icons'

function isPast(e: HomeEvent): boolean {
  const end = new Date(e.endsAt || e.startsAt)
  return end.getTime() < Date.now()
}

function relativeLabel(k: string): string {
  const today = new Date()
  const tk = localDayKey(today.toISOString())
  const tomorrow = new Date()
  tomorrow.setDate(today.getDate() + 1)
  const tmk = localDayKey(tomorrow.toISOString())
  if (k === tk) return 'Today'
  if (k === tmk) return 'Tomorrow'
  const d = new Date(k + 'T00:00:00')
  return d.toLocaleDateString(undefined, { weekday: 'long', month: 'long', day: 'numeric' })
}

export function EventsPage() {
  const [events, setEvents] = useState<HomeEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [showPast, setShowPast] = useState(false)
  const [adding, setAdding] = useState(false)
  const [editing, setEditing] = useState<HomeEvent | null>(null)

  const load = () => {
    api.events.list('', '').then((all) => {
      setEvents(all)
      setLoading(false)
    })
  }
  useEffect(load, [])

  const { upcoming, past } = useMemo(() => {
    const up: HomeEvent[] = []
    const pa: HomeEvent[] = []
    for (const e of events) (isPast(e) ? pa : up).push(e)
    up.sort((a, b) => a.startsAt.localeCompare(b.startsAt))
    return { upcoming: up, past: pa }
  }, [events])

  const grouped = useMemo(() => {
    const map = new Map<string, HomeEvent[]>()
    for (const e of upcoming) {
      const k = localDayKey(e.startsAt)
      const list = map.get(k) ?? []
      list.push(e)
      map.set(k, list)
    }
    return [...map.entries()].sort(([a], [b]) => a.localeCompare(b))
  }, [upcoming])

  const addEvent = (draft: EventDraft) => {
    api.events
      .create(draft)
      .then(() => {
        setAdding(false)
        load()
      })
  }
  const saveEvent = (draft: EventDraft) => {
    if (!editing) return
    api.events.update(editing.id, draft).then(() => {
      setEditing(null)
      load()
    })
  }

  if (loading) return <div className="loading">Loading…</div>

  return (
    <div className="stack">
      <div className="page-head">
        <h1 className="page-title">Events</h1>
        <button className="btn" onClick={() => setAdding(true)}>
          <Plus size={16} /> Add event
        </button>
      </div>

      {grouped.length === 0 ? (
        <div className="card">
          <EmptyState
            title="No upcoming events"
            hint="Events you schedule will be listed here."
            action={
              <button className="btn" onClick={() => setAdding(true)}>
                <Plus size={16} /> New event
              </button>
            }
          />
        </div>
      ) : (
        grouped.map(([k, list]) => (
          <section key={k} className="card">
            <div className="card-head">
              <h3>{relativeLabel(k)}</h3>
              <span className="count-badge">{list.length}</span>
            </div>
            <ul className="list">
              {list.map((e) => (
                <li key={e.id} className="list-row">
                  <span className="event-dot" />
                  <div className="list-info">
                    <span className="task-title">{e.title}</span>
                    <span className="task-due">
                      <Clock size={13} /> {formatDateTime(e.startsAt)}
                      {e.location ? `, ${e.location}` : ''}
                    </span>
                    {e.description && <span className="muted event-desc">{e.description}</span>}
                  </div>
                  <div className="row-actions">
                    <button className="icon-btn" onClick={() => setEditing(e)} aria-label={`Edit ${e.title}`}>
                      <Pencil size={15} />
                    </button>
                    <button className="icon-btn danger" onClick={() => api.events.remove(e.id).then(load)} aria-label={`Delete ${e.title}`}>
                      <Trash size={15} />
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          </section>
        ))
      )}

      {past.length > 0 && (
        <section className="card">
          <button className="collapse-toggle" onClick={() => setShowPast((s) => !s)}>
            {showPast ? 'Hide' : 'Show'} past events ({past.length})
          </button>
          {showPast && (
            <ul className="list">
              {past.map((e) => (
                <li key={e.id} className="list-row">
                  <span className="event-dot faded" />
                  <div className="list-info">
                    <span className="task-title">{e.title}</span>
                    <span className="task-due">
                      <Clock size={13} /> {formatDateTime(e.startsAt)}
                    </span>
                  </div>
                  <div className="row-actions">
                    <button className="icon-btn danger" onClick={() => api.events.remove(e.id).then(load)} aria-label={`Delete ${e.title}`}>
                      <Trash size={15} />
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </section>
      )}

      {adding && (
        <Modal title="New event" onClose={() => setAdding(false)}>
          <EventForm submitLabel="Create" onSubmit={addEvent} onCancel={() => setAdding(false)} />
        </Modal>
      )}

      {editing && (
        <Modal title="Edit event" onClose={() => setEditing(null)}>
          <EventForm initial={editing} submitLabel="Save" onSubmit={saveEvent} onCancel={() => setEditing(null)} />
        </Modal>
      )}
    </div>
  )
}