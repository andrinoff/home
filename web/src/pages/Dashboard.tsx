import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, formatDayKey, todayKey } from '../api'
import type { Dashboard } from '../types'
import { EmptyState } from '../components/EmptyState'
import { CalendarClock, Cart, CheckSquare, ChevronRight, Clock } from '../components/icons'

export function DashboardPage() {
  const [dash, setDash] = useState<Dashboard | null>(null)

  const load = () => {
    api.dashboard().then(setDash)
  }
  useEffect(load, [])

  const toggleTask = (id: number) => api.tasks.toggle(id).then(load)
  const toggleItem = (id: number) => api.grocery.toggleItem(id).then(load)

  if (!dash) return <div className="loading">Loading…</div>

  const today = new Date()
  const tk = todayKey()

  const dueLabel = (due: string) => {
    if (!due) return 'no due date'
    if (due < tk) return `overdue · ${formatDayKey(due)}`
    if (due === tk) return 'due today'
    return formatDayKey(due)
  }
  const dueClass = (due: string) => (due && due <= tk ? 'due-overdue' : '')

  const eventDate = (iso: string) => {
    const d = new Date(iso)
    const date = d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' })
    const time = d.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' })
    return `${date} · ${time}`
  }

  return (
    <div className="stack">
      <h1 className="page-title">Overview</h1>
      <p className="day-heading">
        {today.toLocaleDateString(undefined, { weekday: 'long', month: 'long', day: 'numeric' })}
      </p>

      <section className="card">
        <div className="card-head">
          <h3>
            <CheckSquare size={18} /> Tasks
          </h3>
          {dash.openTaskCount > 0 && <span className="count-badge">{dash.openTaskCount} open</span>}
          <Link to="/tasks" className="card-link">
            All tasks <ChevronRight size={14} />
          </Link>
        </div>
        {dash.tasks.length === 0 ? (
          <EmptyState title="Nothing due" hint="No open tasks due today." />
        ) : (
          <ul className="list">
            {dash.tasks.map((t) => (
              <li key={t.id} className="list-row">
                <input
                  type="checkbox"
                  className="checkbox"
                  checked={t.done}
                  onChange={() => toggleTask(t.id)}
                  aria-label={`Toggle ${t.title}`}
                />
                <div className="list-info">
                  <span className={`task-title ${t.done ? 'done' : ''}`}>{t.title}</span>
                  <span className={`task-due ${dueClass(t.dueDate)}`}>{dueLabel(t.dueDate)}</span>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="card">
        <div className="card-head">
          <h3>
            <CalendarClock size={18} /> Upcoming events
          </h3>
          <Link to="/events" className="card-link">
            All events <ChevronRight size={14} />
          </Link>
        </div>
        {dash.events.length === 0 ? (
          <EmptyState title="Nothing scheduled" hint="Upcoming events will appear here." />
        ) : (
          <ul className="list">
            {dash.events.map((e) => (
              <li key={e.id} className="list-row">
                <span className="event-dot" />
                <div className="list-info">
                  <span className="task-title">{e.title}</span>
                  <span className="task-due">
                    <Clock size={13} /> {eventDate(e.startsAt)}
                  </span>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>

      <section className="card">
        <div className="card-head">
          <h3>
            <Cart size={18} /> Groceries
          </h3>
          {dash.groceryTotal > 0 && <span className="count-badge">{dash.groceryTotal} items</span>}
          <Link to="/grocery" className="card-link">
            Lists <ChevronRight size={14} />
          </Link>
        </div>
        {dash.groceries.length === 0 ? (
          <EmptyState title="All stocked up" hint="Unchecked grocery items will show here." />
        ) : (
          <ul className="list">
            {dash.groceries.map((it) => (
              <li key={it.id} className="list-row">
                <input
                  type="checkbox"
                  className="checkbox"
                  checked={it.checked}
                  onChange={() => toggleItem(it.id)}
                  aria-label={`Toggle ${it.name}`}
                />
                <div className="list-info">
                  <span className={`task-title ${it.checked ? 'done' : ''}`}>{it.name}</span>
                  {it.quantity && <span className="task-due">{it.quantity}</span>}
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}