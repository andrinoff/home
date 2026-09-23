package store

import (
	"database/sql"
	"fmt"
)

// Note is a free-form text note.
type Note struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ListNotes returns all notes, newest update first.
func (s *Store) ListNotes() ([]Note, error) {
	rows, err := s.db.Query(`SELECT id, title, body, created_at, updated_at FROM notes ORDER BY updated_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	notes := []Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// GetNote returns a single note by id.
func (s *Store) GetNote(id int64) (*Note, error) {
	var n Note
	err := s.db.QueryRow(`SELECT id, title, body, created_at, updated_at FROM notes WHERE id = ?`, id).
		Scan(&n.ID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get note: %w", err)
	}
	return &n, nil
}

// CreateNote inserts a new note.
func (s *Store) CreateNote(n *Note) (*Note, error) {
	created := now()
	res, err := s.db.Exec(`INSERT INTO notes (title, body, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		n.Title, n.Body, created, created)
	if err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("create note: %w", err)
	}
	return s.GetNote(id)
}

// UpdateNote replaces title and body of a note.
func (s *Store) UpdateNote(n *Note) (*Note, error) {
	res, err := s.db.Exec(`UPDATE notes SET title = ?, body = ?, updated_at = ? WHERE id = ?`,
		n.Title, n.Body, now(), n.ID)
	if err != nil {
		return nil, fmt.Errorf("update note: %w", err)
	}
	if err := ensureAffected(res, "note"); err != nil {
		return nil, err
	}
	return s.GetNote(n.ID)
}

// DeleteNote removes a note.
func (s *Store) DeleteNote(id int64) error {
	res, err := s.db.Exec(`DELETE FROM notes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	return ensureAffected(res, "note")
}
