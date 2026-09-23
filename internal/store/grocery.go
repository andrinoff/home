package store

import (
	"database/sql"
	"fmt"
)

// GroceryList is a named collection of grocery items.
type GroceryList struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	CreatedAt string        `json:"createdAt"`
	Items     []GroceryItem `json:"items"`
}

// GroceryItem is a single line on a grocery list.
type GroceryItem struct {
	ID        int64  `json:"id"`
	ListID    int64  `json:"listId"`
	Name      string `json:"name"`
	Quantity  string `json:"quantity"`
	Checked   bool   `json:"checked"`
	SortOrder int    `json:"sortOrder"`
	CreatedAt string `json:"createdAt"`
}

// ListGroceryLists returns all lists with their items populated.
func (s *Store) ListGroceryLists() ([]GroceryList, error) {
	rows, err := s.db.Query(`SELECT id, name, created_at FROM grocery_lists ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("list grocery lists: %w", err)
	}
	defer rows.Close()

	lists := []GroceryList{}
	for rows.Next() {
		var l GroceryList
		if err := rows.Scan(&l.ID, &l.Name, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan grocery list: %w", err)
		}
		l.Items = []GroceryItem{}
		lists = append(lists, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate grocery lists: %w", err)
	}

	items, err := s.allGroceryItems()
	if err != nil {
		return nil, err
	}
	byList := map[int64][]GroceryItem{}
	for _, it := range items {
		byList[it.ListID] = append(byList[it.ListID], it)
	}
	for i := range lists {
		lists[i].Items = byList[lists[i].ID]
		if lists[i].Items == nil {
			lists[i].Items = []GroceryItem{}
		}
	}
	return lists, nil
}

func (s *Store) allGroceryItems() ([]GroceryItem, error) {
	rows, err := s.db.Query(`SELECT id, list_id, name, quantity, checked, sort_order, created_at
		FROM grocery_items ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("list grocery items: %w", err)
	}
	defer rows.Close()

	items := []GroceryItem{}
	for rows.Next() {
		var it GroceryItem
		var checked int
		if err := rows.Scan(&it.ID, &it.ListID, &it.Name, &it.Quantity, &checked, &it.SortOrder, &it.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan grocery item: %w", err)
		}
		it.Checked = intToBool(checked)
		items = append(items, it)
	}
	return items, rows.Err()
}

// ListGroceryItems returns the items of a single list.
func (s *Store) ListGroceryItems(listID int64) ([]GroceryItem, error) {
	rows, err := s.db.Query(`SELECT id, list_id, name, quantity, checked, sort_order, created_at
		FROM grocery_items WHERE list_id = ? ORDER BY sort_order, id`, listID)
	if err != nil {
		return nil, fmt.Errorf("list grocery items: %w", err)
	}
	defer rows.Close()

	items := []GroceryItem{}
	for rows.Next() {
		var it GroceryItem
		var checked int
		if err := rows.Scan(&it.ID, &it.ListID, &it.Name, &it.Quantity, &checked, &it.SortOrder, &it.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan grocery item: %w", err)
		}
		it.Checked = intToBool(checked)
		items = append(items, it)
	}
	return items, rows.Err()
}

// CreateGroceryList adds a new list.
func (s *Store) CreateGroceryList(name string) (*GroceryList, error) {
	res, err := s.db.Exec(`INSERT INTO grocery_lists (name, created_at) VALUES (?, ?)`, name, now())
	if err != nil {
		return nil, fmt.Errorf("create grocery list: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("create grocery list: %w", err)
	}
	return &GroceryList{ID: id, Name: name, CreatedAt: now(), Items: []GroceryItem{}}, nil
}

// RenameGroceryList changes a list's name.
func (s *Store) RenameGroceryList(id int64, name string) error {
	res, err := s.db.Exec(`UPDATE grocery_lists SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		return fmt.Errorf("rename grocery list: %w", err)
	}
	return ensureAffected(res, "grocery list")
}

// DeleteGroceryList removes a list and (via cascade) its items.
func (s *Store) DeleteGroceryList(id int64) error {
	res, err := s.db.Exec(`DELETE FROM grocery_lists WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete grocery list: %w", err)
	}
	return ensureAffected(res, "grocery list")
}

// CreateGroceryItem adds an item to a list at the end.
func (s *Store) CreateGroceryItem(listID int64, name, quantity string) (*GroceryItem, error) {
	var maxOrder sql.NullInt64
	if err := s.db.QueryRow(`SELECT MAX(sort_order) FROM grocery_items WHERE list_id = ?`, listID).Scan(&maxOrder); err != nil {
		return nil, fmt.Errorf("create grocery item: %w", err)
	}
	sortOrder := 0
	if maxOrder.Valid {
		sortOrder = int(maxOrder.Int64) + 1
	}
	created := now()
	res, err := s.db.Exec(`INSERT INTO grocery_items (list_id, name, quantity, checked, sort_order, created_at)
		VALUES (?, ?, ?, 0, ?, ?)`, listID, name, quantity, sortOrder, created)
	if err != nil {
		return nil, fmt.Errorf("create grocery item: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("create grocery item: %w", err)
	}
	return &GroceryItem{
		ID: id, ListID: listID, Name: name, Quantity: quantity,
		Checked: false, SortOrder: sortOrder, CreatedAt: created,
	}, nil
}

// UpdateGroceryItem changes name, quantity and position of an item.
func (s *Store) UpdateGroceryItem(id int64, name, quantity string, sortOrder int) (*GroceryItem, error) {
	res, err := s.db.Exec(`UPDATE grocery_items SET name = ?, quantity = ?, sort_order = ? WHERE id = ?`,
		name, quantity, sortOrder, id)
	if err != nil {
		return nil, fmt.Errorf("update grocery item: %w", err)
	}
	if err := ensureAffected(res, "grocery item"); err != nil {
		return nil, err
	}
	return s.getGroceryItem(id)
}

// ToggleGroceryItem flips the checked state of an item.
func (s *Store) ToggleGroceryItem(id int64) (*GroceryItem, error) {
	res, err := s.db.Exec(`UPDATE grocery_items SET checked = 1 - checked WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("toggle grocery item: %w", err)
	}
	if err := ensureAffected(res, "grocery item"); err != nil {
		return nil, err
	}
	return s.getGroceryItem(id)
}

// DeleteGroceryItem removes an item.
func (s *Store) DeleteGroceryItem(id int64) error {
	res, err := s.db.Exec(`DELETE FROM grocery_items WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete grocery item: %w", err)
	}
	return ensureAffected(res, "grocery item")
}

// ClearCheckedGroceryItems removes all checked items from a list.
func (s *Store) ClearCheckedGroceryItems(listID int64) error {
	_, err := s.db.Exec(`DELETE FROM grocery_items WHERE list_id = ? AND checked = 1`, listID)
	if err != nil {
		return fmt.Errorf("clear checked grocery items: %w", err)
	}
	return nil
}

func (s *Store) getGroceryItem(id int64) (*GroceryItem, error) {
	var it GroceryItem
	var checked int
	err := s.db.QueryRow(`SELECT id, list_id, name, quantity, checked, sort_order, created_at
		FROM grocery_items WHERE id = ?`, id).
		Scan(&it.ID, &it.ListID, &it.Name, &it.Quantity, &checked, &it.SortOrder, &it.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get grocery item: %w", err)
	}
	it.Checked = intToBool(checked)
	return &it, nil
}

func ensureAffected(res sql.Result, what string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update %s: %w", what, err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
