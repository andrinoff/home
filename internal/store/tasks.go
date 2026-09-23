package store

import (
	"database/sql"
	"fmt"
)

// Task is a to-do item, optionally due on a specific day.
type Task struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Details   string `json:"details"`
	Priority  string `json:"priority"` // low | medium | high
	DueDate   string `json:"dueDate"`  // YYYY-MM-DD or ""
	Done      bool   `json:"done"`
	SortOrder int    `json:"sortOrder"`
	CreatedAt string `json:"createdAt"`
}

// ListTasks returns tasks, optionally filtered by done state.
// status: "all", "open" (default) or "done".
func (s *Store) ListTasks(status string) ([]Task, error) {
	q := `SELECT id, title, details, priority, IFNULL(due_date, ''), done, sort_order, created_at
		FROM tasks`
	args := []any{}
	switch status {
	case "done":
		q += " WHERE done = 1"
	case "all":
		// no filter
	default:
		q += " WHERE done = 0"
	}
	q += " ORDER BY done, due_date IS NOT NULL DESC, due_date, sort_order, id"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks, err := scanTasks(rows)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// TasksForDashboard returns open tasks that are due today or overdue,
// followed by open tasks without a due date, capped at limit.
func (s *Store) TasksForDashboard(limit int) ([]Task, error) {
	today := timeNow().UTC().Format("2006-01-02")
	rows, err := s.db.Query(`SELECT id, title, details, priority, IFNULL(due_date, ''), done, sort_order, created_at
		FROM tasks
		WHERE done = 0 AND (due_date IS NULL OR due_date <= ?)
		ORDER BY due_date IS NULL, due_date, sort_order, id
		LIMIT ?`, today, limit)
	if err != nil {
		return nil, fmt.Errorf("list dashboard tasks: %w", err)
	}
	defer rows.Close()
	return scanTasks(rows)
}

// CreateTask inserts a new task.
func (s *Store) CreateTask(t *Task) (*Task, error) {
	created := now()
	due := nullableString(t.DueDate)
	res, err := s.db.Exec(`INSERT INTO tasks (title, details, priority, due_date, done, sort_order, created_at)
		VALUES (?, ?, ?, ?, 0, ?, ?)`,
		t.Title, t.Details, t.Priority, due, t.SortOrder, created)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return s.getTask(id)
}

// UpdateTask updates an existing task.
func (s *Store) UpdateTask(t *Task) (*Task, error) {
	res, err := s.db.Exec(`UPDATE tasks SET title = ?, details = ?, priority = ?, due_date = ?, done = ?, sort_order = ?
		WHERE id = ?`,
		t.Title, t.Details, t.Priority, nullableString(t.DueDate), boolToInt(t.Done), t.SortOrder, t.ID)
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	if err := ensureAffected(res, "task"); err != nil {
		return nil, err
	}
	return s.getTask(t.ID)
}

// ToggleTask flips the done state.
func (s *Store) ToggleTask(id int64) (*Task, error) {
	res, err := s.db.Exec(`UPDATE tasks SET done = 1 - done WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("toggle task: %w", err)
	}
	if err := ensureAffected(res, "task"); err != nil {
		return nil, err
	}
	return s.getTask(id)
}

// DeleteTask removes a task.
func (s *Store) DeleteTask(id int64) error {
	res, err := s.db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return ensureAffected(res, "task")
}

func (s *Store) getTask(id int64) (*Task, error) {
	var t Task
	var done int
	err := s.db.QueryRow(`SELECT id, title, details, priority, IFNULL(due_date, ''), done, sort_order, created_at
		FROM tasks WHERE id = ?`, id).
		Scan(&t.ID, &t.Title, &t.Details, &t.Priority, &t.DueDate, &done, &t.SortOrder, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	t.Done = intToBool(done)
	return &t, nil
}

func scanTasks(rows *sql.Rows) ([]Task, error) {
	tasks := []Task{}
	for rows.Next() {
		var t Task
		var done int
		if err := rows.Scan(&t.ID, &t.Title, &t.Details, &t.Priority, &t.DueDate, &done, &t.SortOrder, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		t.Done = intToBool(done)
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
