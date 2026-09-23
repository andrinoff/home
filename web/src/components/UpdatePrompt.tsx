import { useEffect, useState } from 'react'

// Polls the server health endpoint and shows a badge when the server is
// unreachable, so a dead server looks intentional instead of like broken UI.
export function UpdatePrompt() {
  const [online, setOnline] = useState(true)

  useEffect(() => {
    let cancelled = false
    const check = async () => {
      try {
        const res = await fetch('/api/health', { cache: 'no-store' })
        if (!cancelled) setOnline(res.ok)
      } catch {
        if (!cancelled) setOnline(false)
      }
    }
    const timer = setInterval(check, 15000)
    check()
    return () => {
      cancelled = true
      clearInterval(timer)
    }
  }, [])

  if (online) return null
  return <span className="offline-badge">server unreachable</span>
}