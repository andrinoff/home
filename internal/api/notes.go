package api

import (
	"net/http"

	"github.com/andrinoff/home/internal/store"
)

func (s *Server) noteRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/notes", s.listNotes)
	mux.HandleFunc("POST /api/notes", s.createNote)
	mux.HandleFunc("GET /api/notes/{id}", s.getNote)
	mux.HandleFunc("PUT /api/notes/{id}", s.updateNote)
	mux.HandleFunc("DELETE /api/notes/{id}", s.deleteNote)
}

func (s *Server) listNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := s.store.ListNotes()
	if err != nil {
		handleStoreError(w, "list notes", err)
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

func (s *Server) createNote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !requireText(req.Title) {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	note, err := s.store.CreateNote(&store.Note{Title: req.Title, Body: req.Body})
	if err != nil {
		handleStoreError(w, "create note", err)
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

func (s *Server) getNote(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid note id")
		return
	}
	note, err := s.store.GetNote(id)
	if err != nil {
		handleStoreError(w, "get note", err)
		return
	}
	writeJSON(w, http.StatusOK, note)
}

func (s *Server) updateNote(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid note id")
		return
	}
	var req struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !requireText(req.Title) {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	note, err := s.store.UpdateNote(&store.Note{ID: id, Title: req.Title, Body: req.Body})
	if err != nil {
		handleStoreError(w, "update note", err)
		return
	}
	writeJSON(w, http.StatusOK, note)
}

func (s *Server) deleteNote(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid note id")
		return
	}
	if err := s.store.DeleteNote(id); err != nil {
		handleStoreError(w, "delete note", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	dash, err := s.store.DashboardData()
	if err != nil {
		handleStoreError(w, "dashboard", err)
		return
	}
	writeJSON(w, http.StatusOK, dash)
}
