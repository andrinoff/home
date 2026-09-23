package store

import (
	"database/sql"
	"fmt"
	"strings"
)

// Event is a calendar item, used by both the Schedule (calendar view) and
// Events (upcoming list) frontends.
type Event struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StartsAt    string `json:"startsAt"`
	EndsAt      string `json:"endsAt"`
	Location    string `json:"location"`
	CreatedAt   string `json:"createdAt"`
}

// ListEvents returns events overlapping [from, to]. Empty bounds are ignored.
// Values are RFC3339 UTC strings, which compare lexicographically in order.
func (s *Store) ListEvents(from, to string) ([]Event, error) {
	q := `SELECT id, title, description, starts_at, ends_at, location, created_at
		FROM events`
	args := []any{}
	conds := []string{}
	if from != "" {
		conds = append(conds, "ends_at = '' OR ends_at >= ?")
		args = append(args, from)
	}
	if to != "" {
		conds = append(conds, "starts_at <= ?")
		args = append(args, to)
	}
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY starts_at, id"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	events, err := scanEvents(rows)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// UpcomingEvents returns the next limit events that have not ended yet.
func (s *Store) UpcomingEvents(limit int) ([]Event, error) {
	rows, err := s.db.Query(`SELECT id, title, description, starts_at, ends_at, location, created_at
		FROM events WHERE ends_at = '' OR ends_at >= ?
		ORDER BY starts_at, id LIMIT ?`, now(), limit)
	if err != nil {
		return nil, fmt.Errorf("list upcoming events: %w", err)
	}
	defer rows.Close()
	return scanEvents(rows)
}

// CreateEvent inserts a new event.
func (s *Store) CreateEvent(e *Event) (*Event, error) {
	created := now()
	res, err := s.db.Exec(`INSERT INTO events (title, description, starts_at, ends_at, location, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`, e.Title, e.Description, e.StartsAt, e.EndsAt, e.Location, created)
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	return s.getEvent(id)
}

// UpdateEvent updates an existing event.
func (s *Store) UpdateEvent(e *Event) (*Event, error) {
	res, err := s.db.Exec(`UPDATE events SET title = ?, description = ?, starts_at = ?, ends_at = ?, location = ?
		WHERE id = ?`, e.Title, e.Description, e.StartsAt, e.EndsAt, e.Location, e.ID)
	if err != nil {
		return nil, fmt.Errorf("update event: %w", err)
	}
	if err := ensureAffected(res, "event"); err != nil {
		return nil, err
	}
	return s.getEvent(e.ID)
}

// DeleteEvent removes an event.
func (s *Store) DeleteEvent(id int64) error {
	res, err := s.db.Exec(`DELETE FROM events WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete event: %w", err)
	}
	return ensureAffected(res, "event")
}

func (s *Store) getEvent(id int64) (*Event, error) {
	var e Event
	err := s.db.QueryRow(`SELECT id, title, description, starts_at, ends_at, location, created_at
		FROM events WHERE id = ?`, id).
		Scan(&e.ID, &e.Title, &e.Description, &e.StartsAt, &e.EndsAt, &e.Location, &e.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}
	return &e, nil
}

func scanEvents(rows *sql.Rows) ([]Event, error) {
	events := []Event{}
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.StartsAt, &e.EndsAt, &e.Location, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
