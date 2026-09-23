import type { Dashboard, GroceryItem, GroceryList, HomeEvent, Note, Priority, Task } from './types'
import { toast } from './components/Toast'

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(path, {
    method,
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    let msg = `${res.status} ${res.statusText}`
    try {
      const data = await res.json()
      if (data?.error) msg = data.error
    } catch {
      /* keep default message */
    }
    toast(msg)
    throw new Error(msg)
  }
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

export const api = {
  dashboard: () => request<Dashboard>('GET', '/api/dashboard'),

  grocery: {
    lists: () => request<GroceryList[]>('GET', '/api/grocery/lists'),
    createList: (name: string) => request<GroceryList>('POST', '/api/grocery/lists', { name }),
    renameList: (id: number, name: string) => request<{ id: number }>('PUT', `/api/grocery/lists/${id}`, { name }),
    deleteList: (id: number) => request<void>('DELETE', `/api/grocery/lists/${id}`),
    addItem: (listId: number, name: string, quantity: string) =>
      request<GroceryItem>('POST', `/api/grocery/lists/${listId}/items`, { name, quantity }),
    updateItem: (item: GroceryItem) =>
      request<GroceryItem>('PUT', `/api/grocery/items/${item.id}`, {
        name: item.name,
        quantity: item.quantity,
        sortOrder: item.sortOrder,
      }),
    toggleItem: (id: number) => request<GroceryItem>('PATCH', `/api/grocery/items/${id}/toggle`),
    deleteItem: (id: number) => request<void>('DELETE', `/api/grocery/items/${id}`),
    clearChecked: (listId: number) => request<void>('POST', `/api/grocery/lists/${listId}/clear-checked`),
  },

  events: {
    list: (from: string, to: string) => request<HomeEvent[]>('GET', `/api/events?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`),
    create: (e: Omit<HomeEvent, 'id' | 'createdAt'>) => request<HomeEvent>('POST', '/api/events', e),
    update: (id: number, e: Omit<HomeEvent, 'id' | 'createdAt'>) => request<HomeEvent>('PUT', `/api/events/${id}`, e),
    remove: (id: number) => request<void>('DELETE', `/api/events/${id}`),
  },

  tasks: {
    list: (status: 'open' | 'done' | 'all' = 'open') => request<Task[]>('GET', `/api/tasks?status=${status}`),
    create: (t: { title: string; details: string; priority: Priority; dueDate: string }) =>
      request<Task>('POST', '/api/tasks', t),
    update: (id: number, t: { title: string; details: string; priority: Priority; dueDate: string; done: boolean }) =>
      request<Task>('PUT', `/api/tasks/${id}`, t),
    toggle: (id: number) => request<Task>('PATCH', `/api/tasks/${id}/toggle`),
    remove: (id: number) => request<void>('DELETE', `/api/tasks/${id}`),
  },

  notes: {
    list: () => request<Note[]>('GET', '/api/notes'),
    get: (id: number) => request<Note>('GET', `/api/notes/${id}`),
    create: (title: string, body: string) => request<Note>('POST', '/api/notes', { title, body }),
    update: (id: number, title: string, body: string) => request<Note>('PUT', `/api/notes/${id}`, { title, body }),
    remove: (id: number) => request<void>('DELETE', `/api/notes/${id}`),
  },
}

// --- shared formatting helpers ---

export function localDayKey(iso: string): string {
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export function formatDayKey(key: string): string {
  const [y, m, d] = key.split('-').map(Number)
  return new Date(y, m - 1, d).toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })
}

export function formatDateTime(iso: string): string {
  const d = new Date(iso)
  const date = d.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })
  const time = d.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' })
  return `${date} · ${time}`
}

export function toDatetimeLocal(iso: string): string {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function fromDatetimeLocal(value: string): string {
  return new Date(value).toISOString()
}

export function todayKey(): string {
  return localDayKey(new Date().toISOString())
}