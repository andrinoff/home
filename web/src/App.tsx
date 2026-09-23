import { NavLink, Navigate, Route, Routes } from 'react-router-dom'
import { Calendar, CalendarClock, Cart, CheckSquare, HomeIcon, Note } from './components/icons'
import { CalendarPage } from './pages/Calendar'
import { DashboardPage } from './pages/Dashboard'
import { EventsPage } from './pages/Events'
import { GroceryPage } from './pages/Grocery'
import { NotesPage } from './pages/Notes'
import { TasksPage } from './pages/Tasks'
import { UpdatePrompt } from './components/UpdatePrompt'

const nav = [
  { to: '/', label: 'Home', icon: HomeIcon, end: true },
  { to: '/grocery', label: 'Groceries', icon: Cart },
  { to: '/calendar', label: 'Calendar', icon: Calendar },
  { to: '/events', label: 'Events', icon: CalendarClock },
  { to: '/tasks', label: 'Tasks', icon: CheckSquare },
  { to: '/notes', label: 'Notes', icon: Note },
]

export function App() {
  return (
    <div className="app">
      <header className="topbar">
        <span className="brand">
          <HomeIcon size={20} />
          Home
        </span>
        <UpdatePrompt />
      </header>

      <aside className="sidebar">
        <div className="brand sidebar-brand">
          <HomeIcon size={20} />
          Home
        </div>
        <nav>
          {nav.map(({ to, label, icon: Icon, end }) => (
            <NavLink key={to} to={to} end={end} className={({ isActive }) => (isActive ? 'nav-link active' : 'nav-link')}>
              <Icon size={18} />
              {label}
            </NavLink>
          ))}
        </nav>
      </aside>

      <main className="content">
        <Routes>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/grocery" element={<GroceryPage />} />
          <Route path="/calendar" element={<CalendarPage />} />
          <Route path="/events" element={<EventsPage />} />
          <Route path="/tasks" element={<TasksPage />} />
          <Route path="/notes" element={<NotesPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </main>

      <nav className="bottomnav">
        {nav.map(({ to, label, icon: Icon, end }) => (
          <NavLink key={to} to={to} end={end} className={({ isActive }) => (isActive ? 'bottom-link active' : 'bottom-link')}>
            <Icon size={20} />
            <span>{label}</span>
          </NavLink>
        ))}
      </nav>
    </div>
  )
}