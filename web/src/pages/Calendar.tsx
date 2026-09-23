import { useEffect, useMemo, useState } from 'react'
import { api, localDayKey } from '../api'
import type { HomeEvent } from '../types'
import { EmptyState } from '../components/EmptyState'
import { Modal } from '../components/Modal'
import { EventForm, type EventDraft } from '../components/EventForm'
import { ChevronLeft, ChevronRight, Clock, Pencil, Plus, Trash } from '../components/icons'

function monthLabel(d: Date): string {
  return d.toLocaleDateString(undefined, { month: 'long', year: 'numeric' })
}

function dayKey(d: Date): string {
  return localDayKey(d.toISOString())
}

// A Monday-first 6x7 grid covering the given month.
function buildGrid(cursor: Date): Date[] {
  const first = new Date(cursor.getFullYear(), cursor.getMonth(), 1)
  const offset = (first.getDay() + 6) % 7 // Monday = 0
  const start = new Date(first)
  start.setDate(first.getDate() - offset)
  return Array.from({ length: 42 }, (_, i) => {
    const d = new Date(start)
    d.setDate(start.getDate() + i)
    return d
  })
}

const timeOf = (iso: string) => new Date(iso).toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' })

export function CalendarPage() {
  const today = new Date()
  const [cursor, setCursor] = useState(new Date(today.getFullYear(), today.getMonth(), 1))
  const [events, setEvents] = useState<HomeEvent[]>([])
  const [selected, setSelected] = useState<string>(dayKey(today))
  const [adding, setAdding] = useState(false)
  const [editing, setEditing] = useState<HomeEvent | null>(null)

  const grid = useMemo(() => buildGrid(cursor), [cursor])
  const rangeFrom = useMemo(() => {
    const d = new Date(grid[0])
    d.setHours(0, 0, 0, 0)
    return d.toISOString()
  }, [grid])
  const rangeTo = useMemo(() => {
    const d = new Date(grid[grid.length - 1])
    d.setHours(23, 59, 59, 999)
    return d.toISOString()
  }, [grid])

  const load = () => {
    api.events.list(rangeFrom, rangeTo).then(setEvents)
  }
  useEffect(load, [rangeFrom, rangeTo])

  const byDay = useMemo(() => {
    const map = new Map<string, HomeEvent[]>()
    for (const e of events) {
      const k = localDayKey(e.startsAt)
      const list = map.get(k) ?? []
      list.push(e)
      map.set(k, list)
    }
    return map
  }, [events])

  const selectedEvents = byDay.get(selected) ?? []
  const todayK = dayKey(today)

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

  const shiftMonth = (delta: number) => {
    setCursor(new Date(cursor.getFullYear(), cursor.getMonth() + delta, 1))
  }

  const goToday = () => {
    setCursor(new Date(today.getFullYear(), today.getMonth(), 1))
    setSelected(todayK)
  }

  return (
    <div className="stack">
      <div className="page-head">
        <h1 className="page-title">Schedule</h1>
        <button className="btn" onClick={() => setAdding(true)}>
          <Plus size={16} /> Add event
        </button>
      </div>

      <div className="card">
        <div className="cal-head">
          <button className="icon-btn" onClick={() => shiftMonth(-1)} aria-label="Previous month">
            <ChevronLeft size={18} />
          </button>
          <h3>{monthLabel(cursor)}</h3>
          <div className="cal-head-actions">
            <button className="btn ghost sm" onClick={goToday}>
              Today
            </button>
            <button className="icon-btn" onClick={() => shiftMonth(1)} aria-label="Next month">
              <ChevronRight size={18} />
            </button>
          </div>
        </div>

        <div className="cal-grid">
          {['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'].map((d) => (
            <div key={d} className="cal-dow">
              {d}
            </div>
          ))}
          {grid.map((d) => {
            const k = dayKey(d)
            const dayEvents = byDay.get(k) ?? []
            const classes = ['cal-cell']
            if (d.getMonth() !== cursor.getMonth()) classes.push('outside')
            if (k === todayK) classes.push('today')
            if (k === selected) classes.push('selected')
            return (
              <button key={k} className={classes.join(' ')} onClick={() => setSelected(k)}>
                <span className="cal-daynum">{d.getDate()}</span>
                <span className="cal-events">
                  {dayEvents.slice(0, 2).map((e) => (
                    <span key={e.id} className="cal-chip">
                      {timeOf(e.startsAt)} {e.title}
                    </span>
                  ))}
                  {dayEvents.length > 2 && <span className="cal-more">+{dayEvents.length - 2} more</span>}
                </span>
              </button>
            )
          })}
        </div>
      </div>

      <section className="card">
        <div className="card-head">
          <h3>{new Date(selected + 'T00:00:00').toLocaleDateString(undefined, { weekday: 'long', month: 'long', day: 'numeric' })}</h3>
          <button className="card-link" onClick={() => setAdding(true)}>
            <Plus size={14} /> Add
          </button>
        </div>
        {selectedEvents.length === 0 ? (
          <EmptyState title="Nothing planned" hint="No events on this day." />
        ) : (
          <ul className="list">
            {selectedEvents.map((e) => (
              <li key={e.id} className="list-row">
                <span className="event-dot" />
                <div className="list-info">
                  <span className="task-title">{e.title}</span>
                  <span className="task-due">
                    <Clock size={13} /> {timeOf(e.startsAt)}
                    {e.endsAt ? ` – ${timeOf(e.endsAt)}` : ''}
                    {e.location ? ` · ${e.location}` : ''}
                  </span>
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
        )}
      </section>

      {adding && (
        <Modal title="New event" onClose={() => setAdding(false)}>
          <EventForm
            initial={{ startsAt: new Date(selected + 'T09:00:00').toISOString() }}
            submitLabel="Create"
            onSubmit={addEvent}
            onCancel={() => setAdding(false)}
          />
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