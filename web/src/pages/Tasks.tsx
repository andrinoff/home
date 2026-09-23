import { useEffect, useMemo, useState } from 'react'
import { api, formatDayKey, todayKey } from '../api'
import type { Priority, Task } from '../types'
import { EmptyState } from '../components/EmptyState'
import { Modal } from '../components/Modal'
import { Pencil, Plus, Trash } from '../components/icons'

type Filter = 'open' | 'all' | 'done'

const PRIORITY_LABEL: Record<Priority, string> = { low: 'Low', medium: 'Med', high: 'High' }

export function TasksPage() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<Filter>('open')
  const [title, setTitle] = useState('')
  const [due, setDue] = useState('')
  const [priority, setPriority] = useState<Priority>('medium')

  const [editing, setEditing] = useState<Task | null>(null)
  const [editDraft, setEditDraft] = useState({ title: '', details: '', due: '', priority: 'medium' as Priority })

  const load = () => {
    api.tasks.list('all').then((all) => {
      setTasks(all)
      setLoading(false)
    })
  }
  useEffect(load, [])

  const add = (e: React.FormEvent) => {
    e.preventDefault()
    const t = title.trim()
    if (!t) return
    api.tasks
      .create({ title: t, details: '', priority, dueDate: due })
      .then(() => {
        setTitle('')
        setDue('')
        setPriority('medium')
        load()
      })
  }

  const toggle = (id: number) => api.tasks.toggle(id).then(load)

  const startEdit = (t: Task) => {
    setEditing(t)
    setEditDraft({ title: t.title, details: t.details, due: t.dueDate, priority: t.priority })
  }

  const save = () => {
    if (!editing) return
    const t = editDraft.title.trim()
    if (!t) return
    api.tasks
      .update(editing.id, {
        title: t,
        details: editDraft.details.trim(),
        priority: editDraft.priority,
        dueDate: editDraft.due,
        done: editing.done,
      })
      .then(() => {
        setEditing(null)
        load()
      })
  }

  const sorted = useMemo(() => {
    const tk = todayKey()
    const rank = (p: Priority) => (p === 'high' ? 0 : p === 'medium' ? 1 : 2)
    const s = [...tasks].sort((a, b) => {
      if (a.done !== b.done) return a.done ? 1 : -1
      // overdue first, then by due date, then priority
      const aOver = a.dueDate && a.dueDate < tk ? 0 : 1
      const bOver = b.dueDate && b.dueDate < tk ? 0 : 1
      if (aOver !== bOver) return aOver - bOver
      const aDue = a.dueDate || '9999'
      const bDue = b.dueDate || '9999'
      if (aDue !== bDue) return aDue.localeCompare(bDue)
      return rank(a.priority) - rank(b.priority)
    })
    return s
  }, [tasks])

  const shown = filter === 'all' ? sorted : sorted.filter((t) => (filter === 'done' ? t.done : !t.done))
  const openCount = tasks.filter((t) => !t.done).length
  const doneCount = tasks.length - openCount

  if (loading) return <div className="loading">Loading…</div>

  return (
    <div className="stack">
      <div className="page-head">
        <h1 className="page-title">Tasks</h1>
        <span className="count-badge">{openCount} open</span>
      </div>

      <form className="add-form" onSubmit={add}>
        <input className="input" placeholder="Add a task…" value={title} onChange={(e) => setTitle(e.target.value)} />
        <input className="input date" type="date" value={due} onChange={(e) => setDue(e.target.value)} aria-label="Due date" />
        <select className="input prio" value={priority} onChange={(e) => setPriority(e.target.value as Priority)} aria-label="Priority">
          <option value="low">Low</option>
          <option value="medium">Medium</option>
          <option value="high">High</option>
        </select>
        <button className="btn" type="submit" disabled={!title.trim()}>
          <Plus size={16} /> Add
        </button>
      </form>

      <div className="segmented" role="group" aria-label="Filter tasks">
        <button aria-pressed={filter === 'open'} className={filter === 'open' ? 'active' : ''} onClick={() => setFilter('open')}>
          Open <span className="seg-count">{openCount}</span>
        </button>
        <button aria-pressed={filter === 'all'} className={filter === 'all' ? 'active' : ''} onClick={() => setFilter('all')}>
          All <span className="seg-count">{tasks.length}</span>
        </button>
        <button aria-pressed={filter === 'done'} className={filter === 'done' ? 'active' : ''} onClick={() => setFilter('done')}>
          Done <span className="seg-count">{doneCount}</span>
        </button>
      </div>

      <div className="card">
        {shown.length === 0 ? (
          <EmptyState title={filter === 'done' ? 'Nothing completed yet' : 'All clear'}
            hint={filter === 'done' ? 'Completed tasks will appear here.' : 'No tasks to show.'}
            action={
              filter !== 'done' ? (
                <button className="btn" onClick={() => document.querySelector<HTMLInputElement>('.add-form input')?.focus()}>
                  <Plus size={16} /> Add a task
                </button>
              ) : undefined
            }
          />
        ) : (
          <ul className="list">
            {shown.map((t) => {
              const tk = todayKey()
              const overdue = !t.done && t.dueDate && t.dueDate < tk
              const dueToday = !t.done && t.dueDate === tk
              return (
                <li key={t.id} className={`list-row ${t.done ? 'checked-row' : ''} ${overdue ? 'overdue' : ''}`}>
                  <input
                    type="checkbox"
                    className="checkbox"
                    checked={t.done}
                    onChange={() => toggle(t.id)}
                    aria-label={`Toggle ${t.title}`}
                  />
                  <div className="list-info">
                    <span className={`task-title ${t.done ? 'done' : ''}`}>
                      {t.title}
                      {t.details && <span className="task-details">{t.details}</span>}
                    </span>
                    <span className={`task-due ${overdue ? 'overdue-text' : dueToday ? 'due-today-text' : ''}`}>
                      <span className={`prio-chip ${t.priority}`}>{PRIORITY_LABEL[t.priority]}</span>
                      {t.dueDate && (
                        <>
                          {overdue ? 'overdue · ' : dueToday ? 'due today' : formatDayKey(t.dueDate)}
                        </>
                      )}
                    </span>
                  </div>
                  <div className="row-actions">
                    <button className="icon-btn" onClick={() => startEdit(t)} aria-label={`Edit ${t.title}`}>
                      <Pencil size={15} />
                    </button>
                    <button className="icon-btn danger" onClick={() => api.tasks.remove(t.id).then(load)} aria-label={`Delete ${t.title}`}>
                      <Trash size={15} />
                    </button>
                  </div>
                </li>
              )
            })}
          </ul>
        )}
      </div>

      {editing && (
        <Modal
          title="Edit task"
          onClose={() => setEditing(null)}
          footer={
            <>
              <button className="btn ghost" onClick={() => setEditing(null)}>
                Cancel
              </button>
              <button className="btn" onClick={save}>
                Save
              </button>
            </>
          }
        >
          <form className="modal-form" onSubmit={(e) => { e.preventDefault(); save() }}>
            <label>
              Title
              <input className="input" value={editDraft.title} onChange={(e) => setEditDraft({ ...editDraft, title: e.target.value })} autoFocus />
            </label>
            <label>
              Details (optional)
              <input className="input" value={editDraft.details} onChange={(e) => setEditDraft({ ...editDraft, details: e.target.value })} />
            </label>
            <div className="form-row">
              <label>
                Due date
                <input className="input" type="date" value={editDraft.due} onChange={(e) => setEditDraft({ ...editDraft, due: e.target.value })} />
              </label>
              <label>
                Priority
                <select className="input" value={editDraft.priority} onChange={(e) => setEditDraft({ ...editDraft, priority: e.target.value as Priority })}>
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                </select>
              </label>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}