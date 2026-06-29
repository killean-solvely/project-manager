package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/killeanjohnson/projectmanager/internal/boards"
	"github.com/killeanjohnson/projectmanager/internal/docs"
	"github.com/killeanjohnson/projectmanager/internal/projects"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// Server wires HTTP handlers to the service layer. Handlers do three things only:
// parse the request, call a service, format the response. No business logic here.
type Server struct {
	port     string
	projects *projects.Service
	docs     *docs.Service
	boards   *boards.Service
}

func NewServer(port string, projectsSvc *projects.Service, docsSvc *docs.Service, boardsSvc *boards.Service) *Server {
	return &Server{
		port:     port,
		projects: projectsSvc,
		docs:     docsSvc,
		boards:   boardsSvc,
	}
}

func (s *Server) Start() error {
	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	s.registerRoutes(mux)

	return http.ListenAndServe(":"+s.port, mux)
}

// --- response helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// respondServiceError maps a service error to an HTTP status: "not found" errors
// become 404, everything else (validation, illegal transitions) becomes 400.
func respondServiceError(w http.ResponseWriter, err error) {
	if strings.Contains(err.Error(), "not found") {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeError(w, http.StatusBadRequest, err.Error())
}

// --- request helpers ---

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// decodeJSONAllowEmpty is for endpoints whose body is optional (e.g. promote with
// no explicit mode). An empty body is fine; malformed JSON is still an error.
func decodeJSONAllowEmpty(r *http.Request, v any) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

func uuidParam(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, key))
}
