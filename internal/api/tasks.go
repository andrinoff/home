package api

import (
	"net/http"

	"github.com/andrinoff/home/internal/store"
)

func (s *Server) taskRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/tasks", s.listTasks)
	mux.HandleFunc("POST /api/tasks", s.createTask)
	mux.HandleFunc("PUT /api/tasks/{id}", s.updateTask)
	mux.HandleFunc("PATCH /api/tasks/{id}/toggle", s.toggleTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", s.deleteTask)
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.store.ListTasks(r.URL.Query().Get("status"))
	if err != nil {
		handleStoreError(w, "list tasks", err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title     string `json:"title"`
		Details   string `json:"details"`
		Priority  string `json:"priority"`
		DueDate   string `json:"dueDate"`
		SortOrder int    `json:"sortOrder"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !requireText(req.Title) {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if !validPriority(req.Priority) {
		writeError(w, http.StatusBadRequest, "priority must be low, medium or high")
		return
	}
	task, err := s.store.CreateTask(&store.Task{
		Title:     req.Title,
		Details:   req.Details,
		Priority:  req.Priority,
		DueDate:   req.DueDate,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		handleStoreError(w, "create task", err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	var req struct {
		Title     string `json:"title"`
		Details   string `json:"details"`
		Priority  string `json:"priority"`
		DueDate   string `json:"dueDate"`
		Done      bool   `json:"done"`
		SortOrder int    `json:"sortOrder"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !requireText(req.Title) {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if !validPriority(req.Priority) {
		writeError(w, http.StatusBadRequest, "priority must be low, medium or high")
		return
	}
	task, err := s.store.UpdateTask(&store.Task{
		ID:        id,
		Title:     req.Title,
		Details:   req.Details,
		Priority:  req.Priority,
		DueDate:   req.DueDate,
		Done:      req.Done,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		handleStoreError(w, "update task", err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) toggleTask(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	task, err := s.store.ToggleTask(id)
	if err != nil {
		handleStoreError(w, "toggle task", err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	if err := s.store.DeleteTask(id); err != nil {
		handleStoreError(w, "delete task", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validPriority(p string) bool {
	return p == "low" || p == "medium" || p == "high"
}
