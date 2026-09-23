import type { SVGProps } from 'react'

type IconProps = SVGProps<SVGSVGElement> & { size?: number }

function base(size: number | undefined, props: IconProps): IconProps {
  const s = size ?? 20
  const { size: _size, ...rest } = props
  return {
    width: s,
    height: s,
    viewBox: '0 0 24 24',
    fill: 'none',
    stroke: 'currentColor',
    strokeWidth: 2,
    strokeLinecap: 'round',
    strokeLinejoin: 'round',
    ...rest,
  }
}

export function HomeIcon(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
      <polyline points="9 22 9 12 15 12 15 22" />
    </svg>
  )
}

export function Cart(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <circle cx="9" cy="21" r="1" />
      <circle cx="20" cy="21" r="1" />
      <path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6" />
    </svg>
  )
}

export function Calendar(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
      <line x1="16" y1="2" x2="16" y2="6" />
      <line x1="8" y1="2" x2="8" y2="6" />
      <line x1="3" y1="10" x2="21" y2="10" />
    </svg>
  )
}

export function CalendarClock(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <path d="M16 2v4M2 2h20" />
      <path d="M11 14.5 9 18h9l-2-3.5" />
      <path d="M12 22a7 7 0 1 0 0-14 7 7 0 0 0 0 14" />
      <path d="M12 11v4h2" />
    </svg>
  )
}

export function CheckSquare(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <polyline points="9 11 12 14 22 4" />
      <path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11" />
    </svg>
  )
}

export function Note(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <path d="M5 3h14a2 2 0 0 1 2 2v12l-5 4H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2z" />
      <path d="M8 8h8M8 12h6M8 16h4" />
    </svg>
  )
}

export function Plus(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <line x1="12" y1="5" x2="12" y2="19" />
      <line x1="5" y1="12" x2="19" y2="12" />
    </svg>
  )
}

export function Trash(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <polyline points="3 6 5 6 21 6" />
      <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
    </svg>
  )
}

export function Pencil(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z" />
    </svg>
  )
}

export function ChevronLeft(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <polyline points="15 18 9 12 15 6" />
    </svg>
  )
}

export function ChevronRight(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <polyline points="9 18 15 12 9 6" />
    </svg>
  )
}

export function X(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <line x1="18" y1="6" x2="6" y2="18" />
      <line x1="6" y1="6" x2="18" y2="18" />
    </svg>
  )
}

export function Clock(p: IconProps) {
  const props = base(p.size, p)
  return (
    <svg {...props}>
      <circle cx="12" cy="12" r="10" />
      <polyline points="12 6 12 12 16 14" />
    </svg>
  )
}