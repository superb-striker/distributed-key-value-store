package api

import (
	"io"
	"net/http"
	"strings"

	"distkv/internal/store"
)

// Server is one node's HTTP API.
type Server struct {
	NodeID string
	Store  *store.Store
}

// NewServer creates an HTTP API server backed by a local store.
func NewServer(nodeID string, store *store.Store) *Server {
	return &Server{
		NodeID: nodeID,
		Store:  store,
	}
}

// Handler returns the HTTP handler for this node.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/kv/", s.handleKV)
	mux.HandleFunc("/healthz", s.handleHealthz)
	return mux
}

// dispatches PUT, GET, and DELETE requests for a key.
func (s *Server) handleKV(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/kv/")

	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.get(w, key)

	case http.MethodPut:
		s.put(w, r, key)

	case http.MethodDelete:
		s.delete(w, key)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// get retrieves a value from the local store.
func (s *Server) get(w http.ResponseWriter, key string) {
	value, err := s.Store.Get(key)

	if err == store.ErrKeyNotFound {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(value); err != nil {
		return
	}
}

// put stores a value in the local store.
func (s *Server) put(w http.ResponseWriter, r *http.Request, key string) {
	value, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	if err := s.Store.Put(key, value); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// delete removes a key from the local store.
func (s *Server) delete(w http.ResponseWriter, key string) {
	if err := s.Store.Delete(key); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// provides a minimal health endpoint for local development and cluster-level testing.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte("ok"))
}
