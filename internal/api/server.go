// Package api exposes the JSON REST API and serves the embedded frontend.
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/andrinoff/home/internal/store"
)

// Server holds the dependencies for all HTTP handlers.
type Server struct {
	store  *store.Store
	static fs.FS
}

// NewServer builds the HTTP handler serving the API and the static frontend.
// A nil static fs is allowed (API-only, e.g. during frontend development).
func NewServer(st *store.Store, static fs.FS) http.Handler {
	s := &Server{store: st, static: static}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/dashboard", s.handleDashboard)

	s.groceryRoutes(mux)
	s.eventRoutes(mux)
	s.taskRoutes(mux)
	s.noteRoutes(mux)

	mux.Handle("/", s.staticHandler())
	return logRequests(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Microsecond))
	})
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			log.Printf("write json: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return false
	}
	return true
}

func pathID(r *http.Request, name string) (int64, bool) {
	raw := r.PathValue(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func requireText(v string) bool {
	return strings.TrimSpace(v) != ""
}

// handleStoreError maps store errors to HTTP responses.
func handleStoreError(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, op+": not found")
	default:
		log.Printf("%s: %v", op, err)
		writeError(w, http.StatusInternalServerError, op+": internal error")
	}
}

// staticHandler serves the embedded SPA, falling back to index.html for
// client-side routes. If the frontend was never built, it explains how.
func (s *Server) staticHandler() http.Handler {
	if s.static == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			writeError(w, http.StatusServiceUnavailable, "frontend not built; run `make build`")
		})
	}
	fileServer := http.FileServerFS(s.static)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if p == "" || p == "." {
			p = "index.html"
		}
		if _, err := fs.Stat(s.static, p); err != nil {
			if idx, readErr := fs.ReadFile(s.static, "index.html"); readErr == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Cache-Control", "no-store")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, string(idx))
				return
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}
