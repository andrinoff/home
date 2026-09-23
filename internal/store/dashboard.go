package store

import "fmt"

// Dashboard is the aggregate payload for the home page.
type Dashboard struct {
	Tasks         []Task        `json:"tasks"`     // open tasks due today or overdue
	Events        []Event       `json:"events"`    // next upcoming events
	GroceryItems  []GroceryItem `json:"groceries"` // unchecked grocery items (capped)
	GroceryTotal  int           `json:"groceryTotal"`
	OpenTaskCount int           `json:"openTaskCount"`
	NoteCount     int           `json:"noteCount"`
}

// DashboardData aggregates everything the home page needs.
func (s *Store) DashboardData() (*Dashboard, error) {
	d := &Dashboard{}

	tasks, err := s.TasksForDashboard(15)
	if err != nil {
		return nil, err
	}
	d.Tasks = tasks

	events, err := s.UpcomingEvents(5)
	if err != nil {
		return nil, err
	}
	d.Events = events

	items, err := s.allGroceryItems()
	if err != nil {
		return nil, err
	}
	d.GroceryTotal = len(items)
	uncapped := []GroceryItem{}
	for _, it := range items {
		if !it.Checked {
			uncapped = append(uncapped, it)
		}
	}
	d.GroceryItems = uncapped
	if len(d.GroceryItems) > 10 {
		d.GroceryItems = d.GroceryItems[:10]
	}

	if err := s.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE done = 0`).Scan(&d.OpenTaskCount); err != nil {
		return nil, fmt.Errorf("count open tasks: %w", err)
	}
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM notes`).Scan(&d.NoteCount); err != nil {
		return nil, fmt.Errorf("count notes: %w", err)
	}
	return d, nil
}
