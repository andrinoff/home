package api

import (
	"net/http"

	"github.com/andrinoff/home/internal/store"
)

func (s *Server) eventRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/events", s.listEvents)
	mux.HandleFunc("POST /api/events", s.createEvent)
	mux.HandleFunc("PUT /api/events/{id}", s.updateEvent)
	mux.HandleFunc("DELETE /api/events/{id}", s.deleteEvent)
}

func (s *Server) listEvents(w http.ResponseWriter, r *http.Request) {
	events, err := s.store.ListEvents(r.URL.Query().Get("from"), r.URL.Query().Get("to"))
	if err != nil {
		handleStoreError(w, "list events", err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (s *Server) createEvent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		StartsAt    string `json:"startsAt"`
		EndsAt      string `json:"endsAt"`
		Location    string `json:"location"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !requireText(req.Title) {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if !requireText(req.StartsAt) {
		writeError(w, http.StatusBadRequest, "startsAt is required")
		return
	}
	event, err := s.store.CreateEvent(&store.Event{
		Title:       req.Title,
		Description: req.Description,
		StartsAt:    req.StartsAt,
		EndsAt:      req.EndsAt,
		Location:    req.Location,
	})
	if err != nil {
		handleStoreError(w, "create event", err)
		return
	}
	writeJSON(w, http.StatusCreated, event)
}

func (s *Server) updateEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		StartsAt    string `json:"startsAt"`
		EndsAt      string `json:"endsAt"`
		Location    string `json:"location"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !requireText(req.Title) {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if !requireText(req.StartsAt) {
		writeError(w, http.StatusBadRequest, "startsAt is required")
		return
	}
	event, err := s.store.UpdateEvent(&store.Event{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		StartsAt:    req.StartsAt,
		EndsAt:      req.EndsAt,
		Location:    req.Location,
	})
	if err != nil {
		handleStoreError(w, "update event", err)
		return
	}
	writeJSON(w, http.StatusOK, event)
}

func (s *Server) deleteEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid event id")
		return
	}
	if err := s.store.DeleteEvent(id); err != nil {
		handleStoreError(w, "delete event", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
