export interface GroceryItem {
  id: number
  listId: number
  name: string
  quantity: string
  checked: boolean
  sortOrder: number
  createdAt: string
}

export interface GroceryList {
  id: number
  name: string
  createdAt: string
  items: GroceryItem[]
}

export interface HomeEvent {
  id: number
  title: string
  description: string
  startsAt: string
  endsAt: string
  location: string
  createdAt: string
}

export type Priority = 'low' | 'medium' | 'high'

export interface Task {
  id: number
  title: string
  details: string
  priority: Priority
  dueDate: string
  done: boolean
  sortOrder: number
  createdAt: string
}

export interface Note {
  id: number
  title: string
  body: string
  createdAt: string
  updatedAt: string
}

export interface Dashboard {
  tasks: Task[]
  events: HomeEvent[]
  groceries: GroceryItem[]
  groceryTotal: number
  openTaskCount: number
  noteCount: number
}