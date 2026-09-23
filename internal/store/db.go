// Package store provides persistence for the home manager using SQLite.
package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, no CGO required
)

// Store wraps the SQLite database connection.
type Store struct {
	db *sql.DB
}

// schema is idempotent; it is executed on every startup.
const schema = `
CREATE TABLE IF NOT EXISTS grocery_lists (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	name        TEXT    NOT NULL,
	created_at  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS grocery_items (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	list_id     INTEGER NOT NULL REFERENCES grocery_lists(id) ON DELETE CASCADE,
	name        TEXT    NOT NULL,
	quantity    TEXT    NOT NULL DEFAULT '',
	checked     INTEGER NOT NULL DEFAULT 0,
	sort_order  INTEGER NOT NULL DEFAULT 0,
	created_at  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	title       TEXT    NOT NULL,
	description TEXT    NOT NULL DEFAULT '',
	starts_at   TEXT    NOT NULL,
	ends_at     TEXT    NOT NULL DEFAULT '',
	location    TEXT    NOT NULL DEFAULT '',
	created_at  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	title       TEXT    NOT NULL,
	details     TEXT    NOT NULL DEFAULT '',
	priority    TEXT    NOT NULL DEFAULT 'medium',
	due_date    TEXT,
	done        INTEGER NOT NULL DEFAULT 0,
	sort_order  INTEGER NOT NULL DEFAULT 0,
	created_at  TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS notes (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	title       TEXT    NOT NULL,
	body        TEXT    NOT NULL DEFAULT '',
	created_at  TEXT    NOT NULL,
	updated_at  TEXT    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_grocery_items_list    ON grocery_items(list_id, sort_order);
CREATE INDEX IF NOT EXISTS idx_events_starts         ON events(starts_at);
CREATE INDEX IF NOT EXISTS idx_tasks_done_due        ON tasks(done, due_date);
CREATE INDEX IF NOT EXISTS idx_notes_updated         ON notes(updated_at);
`

// Open opens (creating if necessary) the SQLite database at path and applies
// the schema. A single writer connection prevents SQLITE_BUSY errors.
func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func timeNow() time.Time { return time.Now() }

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool { return i != 0 }
