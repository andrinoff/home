import { useEffect, useState } from 'react'

// Tiny pub/sub so api.ts can surface errors without context plumbing.
type Listener = (msg: string) => void
const listeners = new Set<Listener>()

export function toast(msg: string) {
  listeners.forEach((l) => l(msg))
}

interface ToastItem {
  id: number
  msg: string
}

let nextId = 1

export function Toaster() {
  const [items, setItems] = useState<ToastItem[]>([])

  useEffect(() => {
    const listener: Listener = (msg) => {
      const id = nextId++
      setItems((prev) => [...prev, { id, msg }])
      setTimeout(() => setItems((prev) => prev.filter((t) => t.id !== id)), 4000)
    }
    listeners.add(listener)
    return () => {
      listeners.delete(listener)
    }
  }, [])

  return (
    <div className="toaster" role="status" aria-live="polite">
      {items.map((t) => (
        <button
          key={t.id}
          className="toast"
          onClick={() => setItems((prev) => prev.filter((x) => x.id !== t.id))}
        >
          {t.msg}
        </button>
      ))}
    </div>
  )
}