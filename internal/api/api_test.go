package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/andrinoff/home/internal/store"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return NewServer(st, nil)
}

func doJSON(t *testing.T, h http.Handler, method, path string, body any) (int, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func expectStatus(t *testing.T, got, want int, what string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got status %d, want %d", what, got, want)
	}
}

func idOf(t *testing.T, m map[string]any) int64 {
	t.Helper()
	if m["id"] == nil {
		raw, _ := json.Marshal(m)
		t.Fatalf("no id in response %s", raw)
	}
	return int64(m["id"].(float64))
}

func piggybackGet(t *testing.T, h http.Handler, path string) []map[string]any {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: status %d", path, rec.Code)
	}
	var out []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return out
}

func TestHealth(t *testing.T) {
	h := newTestServer(t)
	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	expectStatus(t, rec.Code, 200, "health")
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"ok"`)) {
		t.Fatalf("unexpected health body: %s", rec.Body.Bytes())
	}
}

func TestGroceryLifecycle(t *testing.T) {
	h := newTestServer(t)

	code, _ := doJSON(t, h, "GET", "/api/grocery/lists", nil)
	expectStatus(t, code, 200, "GET lists")

	code, list := doJSON(t, h, "POST", "/api/grocery/lists", map[string]string{"name": "Weekly"})
	expectStatus(t, code, 201, "POST list")
	listID := idOf(t, list)

	// create item on real list
	code, item := doJSON(t, h, "POST", "/api/grocery/lists/"+fmt.Sprint(listID)+"/items",
		map[string]string{"name": "Milk", "quantity": "2"})
	expectStatus(t, code, 201, "POST item")
	itemID := idOf(t, item)

	// invalid id -> 400
	code, _ = doJSON(t, h, "POST", "/api/grocery/lists/notanumber/items", map[string]string{"name": "x"})
	expectStatus(t, code, 400, "POST item bad id")

	// toggle
	code, toggled := doJSON(t, h, "PATCH", "/api/grocery/items/"+fmt.Sprint(itemID)+"/toggle", nil)
	expectStatus(t, code, 200, "PATCH toggle")
	if toggled["checked"] != true {
		t.Fatalf("expected checked=true after toggle, got %v", toggled["checked"])
	}

	// rename item
	code, renamed := doJSON(t, h, "PUT", "/api/grocery/items/"+fmt.Sprint(itemID),
		map[string]any{"name": "Soy milk", "quantity": "1", "sortOrder": 0})
	expectStatus(t, code, 200, "PUT item")
	if renamed["name"] != "Soy milk" {
		t.Fatalf("expected renamed name, got %v", renamed["name"])
	}

	// rename list
	code, _ = doJSON(t, h, "PUT", "/api/grocery/lists/"+fmt.Sprint(listID), map[string]string{"name": "Monthly"})
	expectStatus(t, code, 200, "PUT list")

	// clear checked removes the toggled item
	code, _ = doJSON(t, h, "POST", "/api/grocery/lists/"+fmt.Sprint(listID)+"/clear-checked", nil)
	expectStatus(t, code, 204, "clear checked")
	for _, it := range piggybackGet(t, h, "/api/grocery/lists/"+fmt.Sprint(listID)+"/items") {
		if it["id"] == itemID {
			t.Fatalf("checked item should have been cleared")
		}
	}

	// delete list
	code, _ = doJSON(t, h, "DELETE", "/api/grocery/lists/"+fmt.Sprint(listID), nil)
	expectStatus(t, code, 204, "DELETE list")

	// delete again -> 404
	code, _ = doJSON(t, h, "DELETE", "/api/grocery/lists/"+fmt.Sprint(listID), nil)
	expectStatus(t, code, 404, "DELETE list again")
}

func TestEventsRange(t *testing.T) {
	h := newTestServer(t)

	start := "2026-10-01T08:00:00Z"
	end := "2026-10-01T10:00:00Z"
	code, ev := doJSON(t, h, "POST", "/api/events", map[string]any{
		"title": "Dentist", "startsAt": start, "endsAt": end, "location": "Clinic",
	})
	expectStatus(t, code, 201, "POST event")
	evID := idOf(t, ev)

	// inside range
	events := piggybackGet(t, h, "/api/events?from=2026-10-01T00:00:00Z&to=2026-10-02T00:00:00Z")
	if len(events) != 1 {
		t.Fatalf("expected 1 event in range, got %d", len(events))
	}

	// outside range
	events = piggybackGet(t, h, "/api/events?from=2026-11-01T00:00:00Z&to=2026-11-02T00:00:00Z")
	if len(events) != 0 {
		t.Fatalf("expected 0 events outside range, got %d", len(events))
	}

	// update
	code, ev = doJSON(t, h, "PUT", "/api/events/"+fmt.Sprint(evID), map[string]any{
		"title": "Dentist (rescheduled)", "startsAt": start, "endsAt": end, "location": "",
	})
	expectStatus(t, code, 200, "PUT event")
	if ev["title"] != "Dentist (rescheduled)" {
		t.Fatalf("expected rescheduled title, got %v", ev["title"])
	}

	// missing title -> 400
	code, _ = doJSON(t, h, "POST", "/api/events", map[string]any{"startsAt": start})
	expectStatus(t, code, 400, "POST event missing title")

	// delete
	code, _ = doJSON(t, h, "DELETE", "/api/events/"+fmt.Sprint(evID), nil)
	expectStatus(t, code, 204, "DELETE event")
}

func TestTasks(t *testing.T) {
	h := newTestServer(t)

	code, task := doJSON(t, h, "POST", "/api/tasks", map[string]any{
		"title": "Take out trash", "priority": "high", "dueDate": "2026-09-24",
	})
	expectStatus(t, code, 201, "POST task")
	taskID := idOf(t, task)

	// invalid priority -> 400
	code, _ = doJSON(t, h, "POST", "/api/tasks", map[string]any{"title": "x", "priority": "urgent"})
	expectStatus(t, code, 400, "POST task bad priority")

	tasks := piggybackGet(t, h, "/api/tasks?status=open")
	if len(tasks) != 1 {
		t.Fatalf("expected 1 open task, got %d", len(tasks))
	}

	code, toggled := doJSON(t, h, "PATCH", "/api/tasks/"+fmt.Sprint(taskID)+"/toggle", nil)
	expectStatus(t, code, 200, "PATCH task toggle")
	if toggled["done"] != true {
		t.Fatalf("expected task done after toggle")
	}

	tasks = piggybackGet(t, h, "/api/tasks?status=done")
	if len(tasks) != 1 {
		t.Fatalf("expected 1 done task, got %d", len(tasks))
	}

	code, _ = doJSON(t, h, "DELETE", "/api/tasks/"+fmt.Sprint(taskID), nil)
	expectStatus(t, code, 204, "DELETE task")
}

func TestNotes(t *testing.T) {
	h := newTestServer(t)

	code, note := doJSON(t, h, "POST", "/api/notes", map[string]string{"title": "WiFi", "body": "password is secret"})
	expectStatus(t, code, 201, "POST note")
	noteID := idOf(t, note)

	code, got := doJSON(t, h, "GET", "/api/notes/"+fmt.Sprint(noteID), nil)
	expectStatus(t, code, 200, "GET note")
	if got["body"] != "password is secret" {
		t.Fatalf("unexpected note body: %v", got["body"])
	}

	code, updated := doJSON(t, h, "PUT", "/api/notes/"+fmt.Sprint(noteID), map[string]string{"title": "WiFi", "body": "new password"})
	expectStatus(t, code, 200, "PUT note")
	if updated["body"] != "new password" {
		t.Fatalf("note not updated: %v", updated["body"])
	}

	noteList := piggybackGet(t, h, "/api/notes")
	if len(noteList) != 1 {
		t.Fatalf("expected 1 note, got %d", len(noteList))
	}
}

func TestDashboard(t *testing.T) {
	h := newTestServer(t)

	req := httptest.NewRequest("GET", "/api/dashboard", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	expectStatus(t, rec.Code, 200, "dashboard")

	var dash map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &dash); err != nil {
		t.Fatalf("decode dashboard: %v", err)
	}
	for _, key := range []string{"tasks", "events", "groceries", "groceryTotal", "openTaskCount", "noteCount"} {
		if _, ok := dash[key]; !ok {
			t.Fatalf("dashboard missing key %q", key)
		}
	}
}

func TestUnauthenticated400s(t *testing.T) {
	h := newTestServer(t)

	code, _ := doJSON(t, h, "POST", "/api/grocery/lists", map[string]string{"name": "  "})
	expectStatus(t, code, 400, "blank name")

	req := httptest.NewRequest("POST", "/api/grocery/lists", bytes.NewBufferString("{not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	expectStatus(t, rec.Code, 400, "bad json")
}
