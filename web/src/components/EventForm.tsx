import { useState } from 'react'
import type { HomeEvent } from '../types'
import { fromDatetimeLocal, toDatetimeLocal } from '../api'

export interface EventDraft {
  title: string
  description: string
  startsAt: string
  endsAt: string
  location: string
}

interface EventFormProps {
  initial?: Partial<HomeEvent>
  submitLabel: string
  onSubmit: (draft: EventDraft) => void
  onCancel: () => void
}

export function EventForm({ initial, submitLabel, onSubmit, onCancel }: EventFormProps) {
  const [title, setTitle] = useState(initial?.title ?? '')
  const [description, setDescription] = useState(initial?.description ?? '')
  const [startsAt, setStartsAt] = useState(initial?.startsAt ? toDatetimeLocal(initial.startsAt) : '')
  const [endsAt, setEndsAt] = useState(initial?.endsAt ? toDatetimeLocal(initial.endsAt) : '')
  const [location, setLocation] = useState(initial?.location ?? '')

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim() || !startsAt) return
    onSubmit({
      title: title.trim(),
      description: description.trim(),
      startsAt: fromDatetimeLocal(startsAt),
      endsAt: endsAt ? fromDatetimeLocal(endsAt) : '',
      location: location.trim(),
    })
  }

  return (
    <form className="modal-form" onSubmit={submit} id="event-form">
      <label>
        Title
        <input className="input" value={title} onChange={(e) => setTitle(e.target.value)} autoFocus required />
      </label>
      <div className="form-row">
        <label>
          Starts
          <input className="input" type="datetime-local" value={startsAt} onChange={(e) => setStartsAt(e.target.value)} required />
        </label>
        <label>
          Ends (optional)
          <input className="input" type="datetime-local" value={endsAt} onChange={(e) => setEndsAt(e.target.value)} />
        </label>
      </div>
      <label>
        Location (optional)
        <input className="input" value={location} onChange={(e) => setLocation(e.target.value)} placeholder="Where?" />
      </label>
      <label>
        Notes (optional)
        <textarea className="input" rows={3} value={description} onChange={(e) => setDescription(e.target.value)} />
      </label>
      <div className="modal-form-actions">
        <button type="button" className="btn ghost" onClick={onCancel}>
          Cancel
        </button>
        <button type="submit" className="btn">
          {submitLabel}
        </button>
      </div>
    </form>
  )
}