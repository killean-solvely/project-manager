package server

import (
	"context"
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
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wires HTTP handlers to the service layer. Handlers do three things only:
// parse the request, call a service, format the response. No business logic here.
//
// The same service graph also backs the MCP server mounted at /mcp, so REST and
// MCP share one composition root and one database handle.
type Server struct {
	port     string
	projects *projects.Service
	docs     *docs.Service
	boards   *boards.Service

	// mcpSrv is the transport-agnostic MCP server; when mcpEnabled is true it is
	// mounted at /mcp via the streamable-HTTP handler. It is the same server the
	// stdio binary (cmd/mcp) binds, just over a different transport.
	mcpSrv     *mcp.Server
	mcpEnabled bool

	// httpSrv is built in NewServer so Start and Shutdown share one instance
	// without a startup race.
	httpSrv *http.Server
}

// NewServer builds the REST server. mcpSrv is the MCP server to mount at /mcp;
// when mcpEnabled is false the /mcp route is not registered (mcpSrv may be nil).
func NewServer(port string, projectsSvc *projects.Service, docsSvc *docs.Service, boardsSvc *boards.Service, mcpSrv *mcp.Server, mcpEnabled bool) *Server {
	s := &Server{
		port:       port,
		projects:   projectsSvc,
		docs:       docsSvc,
		boards:     boardsSvc,
		mcpSrv:     mcpSrv,
		mcpEnabled: mcpEnabled,
	}
	// Build the http.Server (and its handler) up front so Shutdown never races
	// Start over the httpSrv field: a signal can arrive before Start's goroutine
	// is scheduled, and Shutdown must still see the server.
	s.httpSrv = &http.Server{Addr: ":" + port, Handler: s.Handler()}
	return s
}

// Handler builds the chi router with the REST routes and, when enabled, the MCP
// handler at /mcp. It is exported so tests can drive the exact production router
// through httptest without opening a listener.
func (s *Server) Handler() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	s.registerRoutes(mux)

	if s.mcpEnabled && s.mcpSrv != nil {
		// Streamable HTTP transport, stateless + JSON responses: a plain
		// request/response tool server with no session bookkeeping to leak on
		// shutdown. The handler does its own method/sub-path dispatch, so a bare
		// Handle on /mcp routes every method to it.
		mcpHandler := mcp.NewStreamableHTTPHandler(
			func(*http.Request) *mcp.Server { return s.mcpSrv },
			&mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
		)
		mux.Handle("/mcp", mcpHandler)
	}

	return mux
}

// Start binds the HTTP listener and serves until Shutdown is called. It returns
// nil on a graceful shutdown (http.ErrServerClosed) and the underlying error
// otherwise.
func (s *Server) Start() error {
	err := s.httpSrv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Shutdown gracefully stops the HTTP server, letting in-flight REST and MCP
// requests finish (or until ctx expires).
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
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
