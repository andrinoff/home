import { useEffect, useState } from 'react'
import { api } from '../api'
import type { GroceryItem, GroceryList } from '../types'
import { EmptyState } from '../components/EmptyState'
import { Modal } from '../components/Modal'
import { Plus, Pencil, Trash } from '../components/icons'

export function GroceryPage() {
  const [lists, setLists] = useState<GroceryList[]>([])
  const [loading, setLoading] = useState(true)
  const [activeId, setActiveId] = useState<number | null>(null)
  const [showAddList, setShowAddList] = useState(false)
  const [editingList, setEditingList] = useState<GroceryList | null>(null)
  const [confirmDeleteList, setConfirmDeleteList] = useState<GroceryList | null>(null)

  // per-item inline editors
  const [editingItem, setEditingItem] = useState<GroceryItem | null>(null)
  const [itemDraft, setItemDraft] = useState({ name: '', quantity: '' })

  // add-item form
  const [newName, setNewName] = useState('')
  const [newQty, setNewQty] = useState('')

  const load = () => {
    api.grocery.lists().then((data) => {
      setLists(data)
      if (activeId === null || !data.some((l) => l.id === activeId)) {
        setActiveId(data[0]?.id ?? null)
      }
      setLoading(false)
    })
  }
  useEffect(load, [])

  const active = lists.find((l) => l.id === activeId) ?? null
  const uncheckedCount = active ? active.items.filter((i) => !i.checked).length : 0

  const addItem = (e: React.FormEvent) => {
    e.preventDefault()
    const name = newName.trim()
    if (!name || !active) return
    api.grocery.addItem(active.id, name, newQty.trim()).then(() => {
      setNewName('')
      setNewQty('')
      load()
    })
  }

  const toggle = (it: GroceryItem) => {
    api.grocery.toggleItem(it.id).then(load)
  }

  const startEditItem = (it: GroceryItem) => {
    setEditingItem(it)
    setItemDraft({ name: it.name, quantity: it.quantity })
  }

  const saveItem = () => {
    if (!editingItem) return
    const name = itemDraft.name.trim()
    if (!name) return
    api.grocery.updateItem({ ...editingItem, name, quantity: itemDraft.quantity.trim() }).then(() => {
      setEditingItem(null)
      load()
    })
  }

  const createList = () => {
    const el = document.getElementById('new-list-name') as HTMLInputElement | null
    const name = el?.value.trim()
    if (!name) return
    api.grocery.createList(name).then((l) => {
      setShowAddList(false)
      setActiveId(l.id)
      load()
    })
  }

  if (loading) return <div className="loading">Loading…</div>

  return (
    <div className="grocery">
      <div className="page-head">
        <h1 className="page-title">Groceries</h1>
        <button className="btn" onClick={() => setShowAddList(true)}>
          <Plus size={16} /> New list
        </button>
      </div>

      <div className="grocery-body">
        <div className="list-tabs">
          {lists.length === 0 ? (
            <p className="muted">No lists yet.</p>
          ) : (
            lists.map((l) => (
              <button
                key={l.id}
                className={`list-tab ${l.id === active?.id ? 'active' : ''}`}
                onClick={() => setActiveId(l.id)}
              >
                {l.name}
              </button>
            ))
          )}
          <button className="list-tab add" onClick={() => setShowAddList(true)} aria-label="Add list">
            <Plus size={16} />
          </button>
        </div>

        <div className="card grocery-card">
          {!active ? (
            <EmptyState
              title="Nothing here yet"
              hint="Create a list to start adding groceries."
              action={
                <button className="btn" onClick={() => setShowAddList(true)}>
                  <Plus size={16} /> New list
                </button>
              }
            />
          ) : (
            <>
              <div className="card-head">
                <h3>{active.name}</h3>
                <div className="card-actions">
                  <span className="count-badge">{uncheckedCount} to buy</span>
                  <button className="icon-btn" onClick={() => setEditingList(active)} aria-label="Rename list">
                    <Pencil size={15} />
                  </button>
                  <button
                    className="icon-btn danger"
                    onClick={() => setConfirmDeleteList(active)}
                    aria-label="Delete list"
                  >
                    <Trash size={15} />
                  </button>
                </div>
              </div>

              <form className="add-form" onSubmit={addItem}>
                <input
                  className="input"
                  placeholder="Add item…"
                  value={newName}
                  onChange={(e) => setNewName(e.target.value)}
                  autoFocus
                />
                <input
                  className="input qty"
                  placeholder="Qty"
                  value={newQty}
                  onChange={(e) => setNewQty(e.target.value)}
                />
                <button className="btn" type="submit" disabled={!newName.trim()}>
                  <Plus size={16} /> Add
                </button>
              </form>

              {active.items.length === 0 ? (
                <EmptyState title="List is empty" hint="Add your first item above." />
              ) : (
                <>
                  <ul className="list">
                    {active.items.map((it) => (
                      <li key={it.id} className={`list-row ${it.checked ? 'checked-row' : ''}`}>
                        <input
                          type="checkbox"
                          className="checkbox"
                          checked={it.checked}
                          onChange={() => toggle(it)}
                          aria-label={`Toggle ${it.name}`}
                        />
                        <div className="list-info">
                          <span className={`grocery-name ${it.checked ? 'done' : ''}`}>{it.name}</span>
                          {it.quantity && <span className="qty-badge">{it.quantity}</span>}
                        </div>
                        <div className="row-actions">
                          <button className="icon-btn" onClick={() => startEditItem(it)} aria-label={`Edit ${it.name}`}>
                            <Pencil size={15} />
                          </button>
                          <button className="icon-btn danger" onClick={() => api.grocery.deleteItem(it.id).then(load)} aria-label={`Delete ${it.name}`}>
                            <Trash size={15} />
                          </button>
                        </div>
                      </li>
                    ))}
                  </ul>
                  {uncheckedCount < active.items.length && (
                    <div className="card-foot">
                      <button className="btn ghost" onClick={() => api.grocery.clearChecked(active.id).then(load)}>
                        Clear checked
                      </button>
                    </div>
                  )}
                </>
              )}
            </>
          )}
        </div>
      </div>

      {/* New list modal */}
      {showAddList && (
        <Modal
          title="New list"
          onClose={() => setShowAddList(false)}
          footer={
            <>
              <button className="btn ghost" onClick={() => setShowAddList(false)}>
                Cancel
              </button>
              <button className="btn" onClick={createList}>
                Create
              </button>
            </>
          }
        >
          <input id="new-list-name" className="input" placeholder="List name e.g. Weekly" autoFocus onKeyDown={(e) => e.key === 'Enter' && createList()} />
        </Modal>
      )}

      {/* Rename list modal */}
      {editingList && (
        <Modal
          title="Rename list"
          onClose={() => setEditingList(null)}
          footer={
            <>
              <button className="btn ghost" onClick={() => setEditingList(null)}>
                Cancel
              </button>
              <button
                className="btn"
                onClick={() => {
                  const el = document.getElementById('rename-list') as HTMLInputElement | null
                  const name = el?.value.trim()
                  if (!name) return
                  api.grocery.renameList(editingList.id, name).then(() => {
                    setEditingList(null)
                    load()
                  })
                }}
              >
                Save
              </button>
            </>
          }
        >
          <input id="rename-list" className="input" defaultValue={editingList.name} autoFocus onKeyDown={(e) => e.key === 'Enter' && e.currentTarget.form?.requestSubmit()} />
        </Modal>
      )}

      {/* Edit item modal */}
      {editingItem && (
        <Modal
          title="Edit item"
          onClose={() => setEditingItem(null)}
          footer={
            <>
              <button className="btn ghost" onClick={() => setEditingItem(null)}>
                Cancel
              </button>
              <button className="btn" onClick={saveItem}>
                Save
              </button>
            </>
          }
        >
          <form
            className="modal-form"
            onSubmit={(e) => {
              e.preventDefault()
              saveItem()
            }}
          >
            <label>
              Name
              <input className="input" value={itemDraft.name} onChange={(e) => setItemDraft({ ...itemDraft, name: e.target.value })} autoFocus />
            </label>
            <label>
              Quantity (optional)
              <input className="input" value={itemDraft.quantity} onChange={(e) => setItemDraft({ ...itemDraft, quantity: e.target.value })} />
            </label>
          </form>
        </Modal>
      )}

      {/* Delete list confirm */}
      {confirmDeleteList && (
        <Modal
          title="Delete this list?"
          onClose={() => setConfirmDeleteList(null)}
          footer={
            <>
              <button className="btn ghost" onClick={() => setConfirmDeleteList(null)}>
                Cancel
              </button>
              <button
                className="btn danger"
                onClick={() => {
                  api.grocery.deleteList(confirmDeleteList.id).then(load)
                  setConfirmDeleteList(null)
                }}
              >
                Delete
              </button>
            </>
          }
        >
          <p>
            “{confirmDeleteList.name}” and its {confirmDeleteList.items.length} item
            {confirmDeleteList.items.length === 1 ? '' : 's'} will be deleted. This cannot be undone.
          </p>
        </Modal>
      )}
    </div>
  )
}