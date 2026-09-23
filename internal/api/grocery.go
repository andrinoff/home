package api

import (
	"net/http"
)

func (s *Server) groceryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/grocery/lists", s.listGroceryLists)
	mux.HandleFunc("POST /api/grocery/lists", s.createGroceryList)
	mux.HandleFunc("PUT /api/grocery/lists/{id}", s.updateGroceryList)
	mux.HandleFunc("DELETE /api/grocery/lists/{id}", s.deleteGroceryList)
	mux.HandleFunc("GET /api/grocery/lists/{id}/items", s.listGroceryItems)
	mux.HandleFunc("POST /api/grocery/lists/{id}/items", s.createGroceryItem)
	mux.HandleFunc("PUT /api/grocery/items/{id}", s.updateGroceryItem)
	mux.HandleFunc("PATCH /api/grocery/items/{id}/toggle", s.toggleGroceryItem)
	mux.HandleFunc("DELETE /api/grocery/items/{id}", s.deleteGroceryItem)
	mux.HandleFunc("POST /api/grocery/lists/{id}/clear-checked", s.clearCheckedGroceryItems)
}

func (s *Server) listGroceryLists(w http.ResponseWriter, r *http.Request) {
	lists, err := s.store.ListGroceryLists()
	if err != nil {
		handleStoreError(w, "list grocery lists", err)
		return
	}
	writeJSON(w, http.StatusOK, lists)
}

func (s *Server) createGroceryList(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !requireText(req.Name) {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	list, err := s.store.CreateGroceryList(req.Name)
	if err != nil {
		handleStoreError(w, "create grocery list", err)
		return
	}
	writeJSON(w, http.StatusCreated, list)
}

func (s *Server) updateGroceryList(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid list id")
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !requireText(req.Name) {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := s.store.RenameGroceryList(id, req.Name); err != nil {
		handleStoreError(w, "update grocery list", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"id": id})
}

func (s *Server) deleteGroceryList(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid list id")
		return
	}
	if err := s.store.DeleteGroceryList(id); err != nil {
		handleStoreError(w, "delete grocery list", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listGroceryItems(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid list id")
		return
	}
	items, err := s.store.ListGroceryItems(id)
	if err != nil {
		handleStoreError(w, "list grocery items", err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) createGroceryItem(w http.ResponseWriter, r *http.Request) {
	listID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid list id")
		return
	}
	var req struct {
		Name     string `json:"name"`
		Quantity string `json:"quantity"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !requireText(req.Name) {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	item, err := s.store.CreateGroceryItem(listID, req.Name, req.Quantity)
	if err != nil {
		handleStoreError(w, "create grocery item", err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) updateGroceryItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}
	var req struct {
		Name      string `json:"name"`
		Quantity  string `json:"quantity"`
		SortOrder int    `json:"sortOrder"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	if !requireText(req.Name) {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	item, err := s.store.UpdateGroceryItem(id, req.Name, req.Quantity, req.SortOrder)
	if err != nil {
		handleStoreError(w, "update grocery item", err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) toggleGroceryItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}
	item, err := s.store.ToggleGroceryItem(id)
	if err != nil {
		handleStoreError(w, "toggle grocery item", err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) deleteGroceryItem(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid item id")
		return
	}
	if err := s.store.DeleteGroceryItem(id); err != nil {
		handleStoreError(w, "delete grocery item", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) clearCheckedGroceryItems(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid list id")
		return
	}
	if err := s.store.ClearCheckedGroceryItems(id); err != nil {
		handleStoreError(w, "clear checked grocery items", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
