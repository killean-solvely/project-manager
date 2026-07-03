package server_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/killeanjohnson/projectmanager/internal/boards"
	"github.com/killeanjohnson/projectmanager/internal/docs"
	"github.com/killeanjohnson/projectmanager/internal/mcpserver"
	"github.com/killeanjohnson/projectmanager/internal/persistence/sqlite"
	"github.com/killeanjohnson/projectmanager/internal/projects"
	"github.com/killeanjohnson/projectmanager/internal/server"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// newTestServer builds a Server backed by the same real service graph the
// production composition root uses, over a fresh temp-file sqlite store. The
// MCP server is built from those exact services, so a write through an MCP tool
// and a read through the REST API hit one database - the in-process version of
// the two-process TestStdioBinarySharesDatabase.
func newTestServer(t *testing.T, mcpEnabled bool) *server.Server {
	t.Helper()
	db, err := sqlite.OpenAt(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	projectsSvc := projects.NewService(sqlite.NewProjectsRepository(db))
	docsSvc := docs.NewService(sqlite.NewDocumentsRepository(db))
	boardsSvc := boards.NewService(
		sqlite.NewBoardsRepository(db),
		sqlite.NewColumnsRepository(db),
		sqlite.NewTasksRepository(db),
	)

	mcpSrv := mcpserver.New(projectsSvc, docsSvc, boardsSvc)
	return server.NewServer("0", projectsSvc, docsSvc, boardsSvc, mcpSrv, mcpEnabled)
}

// mcpClient connects an MCP client to the /mcp endpoint of ts over the
// streamable-HTTP transport, the way a remote client does.
func mcpClient(t *testing.T, ts *httptest.Server) *mcp.ClientSession {
	t.Helper()
	client := mcp.NewClient(&mcp.Implementation{Name: "smoke", Version: "0.0.0"}, nil)
	cs, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: ts.URL + "/mcp"}, nil)
	if err != nil {
		t.Fatalf("connect over streamable HTTP: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

// TestMCPHTTPTransportListsTools proves the /mcp mount and streamable-HTTP
// transport work end to end: a client connects and enumerates the tools,
// mirroring the stdio smoke test's >= 18 assertion.
func TestMCPHTTPTransportListsTools(t *testing.T) {
	ts := httptest.NewServer(newTestServer(t, true).Handler())
	defer ts.Close()

	cs := mcpClient(t, ts)
	tools, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools over HTTP: %v", err)
	}
	if len(tools.Tools) < 18 {
		t.Fatalf("expected >= 18 tools over HTTP, got %d", len(tools.Tools))
	}
}

// TestMCPAndRESTShareServices proves the single-composition-root goal: a project
// created through an MCP tool is immediately visible through the REST API on the
// same server instance, because both front doors share one service graph and one
// DB handle.
func TestMCPAndRESTShareServices(t *testing.T) {
	srv := newTestServer(t, true)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	cs := mcpClient(t, ts)
	if _, err := cs.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "create_idea",
		Arguments: map[string]any{"name": "Cross-transport idea"},
	}); err != nil {
		t.Fatalf("create_idea over MCP: %v", err)
	}

	resp, err := http.Get(ts.URL + "/api/projects")
	if err != nil {
		t.Fatalf("GET /api/projects: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/projects status = %d", resp.StatusCode)
	}
	var projs []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&projs); err != nil {
		t.Fatalf("decode projects: %v", err)
	}
	found := false
	for _, p := range projs {
		if p.Name == "Cross-transport idea" {
			found = true
		}
	}
	if !found {
		t.Fatalf("project written via MCP tool not visible via REST: %+v", projs)
	}
}

// TestMCPDisabledOmitsRoute confirms the MCP_HTTP_ENABLED gate: with MCP
// disabled the /mcp route is not mounted, while REST keeps working.
func TestMCPDisabledOmitsRoute(t *testing.T) {
	ts := httptest.NewServer(newTestServer(t, false).Handler())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/mcp", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("POST /mcp: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("with MCP disabled, /mcp status = %d, want 404", resp.StatusCode)
	}

	health, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer health.Body.Close()
	if health.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d, want 200", health.StatusCode)
	}
}

// TestMCPHTTPDNSRebindingRejected documents the SDK's default DNS-rebinding
// protection (plan §7.1): a request that reaches the loopback listener carrying
// a non-loopback Host header is rejected with 403. A future reader hitting a 403
// behind a proxy/hostname should find this test and the DisableLocalhostProtection
// escape hatch rather than be surprised.
func TestMCPHTTPDNSRebindingRejected(t *testing.T) {
	ts := httptest.NewServer(newTestServer(t, true).Handler())
	defer ts.Close()

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/mcp", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// The connection still dials the loopback test server, but the Host header
	// claims a non-loopback name - exactly the DNS-rebinding shape the SDK guards.
	req.Host = "evil.example.com"

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("non-loopback Host: status = %d, want 403", resp.StatusCode)
	}
}

// TestGracefulShutdown confirms Start/Shutdown form a clean lifecycle: Start
// serves until Shutdown is called, Shutdown returns well before its deadline,
// and Start reports no error (the http.ErrServerClosed from a graceful stop is
// swallowed).
func TestGracefulShutdown(t *testing.T) {
	// Grab a free port, then hand it to the server. A brief window exists between
	// closing the probe listener and Start rebinding, which is acceptable here.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("probe listen: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	srv := server.NewServer(strconv.Itoa(port), nil, nil, nil, nil, false)

	startErr := make(chan error, 1)
	go func() { startErr <- srv.Start() }()

	base := "http://127.0.0.1:" + strconv.Itoa(port)
	waitForHealth(t, base)

	// Shutdown should return promptly and well within the deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	select {
	case err := <-startErr:
		if err != nil {
			t.Fatalf("Start returned error on graceful shutdown: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after Shutdown")
	}

	// The listener is closed after a graceful shutdown.
	if _, err := http.Get(base + "/health"); err == nil {
		t.Fatal("expected request to fail after shutdown")
	}
}

// TestShutdownBeforeStart confirms Shutdown is safe to call if Start never ran.
func TestShutdownBeforeStart(t *testing.T) {
	srv := server.NewServer("0", nil, nil, nil, nil, false)
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown before Start: %v", err)
	}
}

func waitForHealth(t *testing.T, base string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(base + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("server did not become healthy in time")
}
