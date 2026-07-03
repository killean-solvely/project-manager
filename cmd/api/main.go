package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/killeanjohnson/projectmanager/internal/boards"
	"github.com/killeanjohnson/projectmanager/internal/config"
	"github.com/killeanjohnson/projectmanager/internal/docs"
	"github.com/killeanjohnson/projectmanager/internal/mcpserver"
	"github.com/killeanjohnson/projectmanager/internal/persistence/sqlite"
	"github.com/killeanjohnson/projectmanager/internal/projects"
	"github.com/killeanjohnson/projectmanager/internal/server"
)

// shutdownTimeout bounds how long graceful shutdown waits for in-flight
// requests to drain before the process exits.
const shutdownTimeout = 10 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := sqlite.OpenAt(cfg.DBPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() { _ = db.Close() }()
	log.Printf("using database at %s", cfg.DBPath)

	projectsRepo := sqlite.NewProjectsRepository(db)
	documentsRepo := sqlite.NewDocumentsRepository(db)
	boardsRepo := sqlite.NewBoardsRepository(db)
	columnsRepo := sqlite.NewColumnsRepository(db)
	tasksRepo := sqlite.NewTasksRepository(db)

	projectsSvc := projects.NewService(projectsRepo)
	docsSvc := docs.NewService(documentsRepo)
	boardsSvc := boards.NewService(boardsRepo, columnsRepo, tasksRepo)

	// One composition root, one DB handle: the MCP server mounted at /mcp is
	// built from the same service graph as the REST API.
	mcpSrv := mcpserver.New(projectsSvc, docsSvc, boardsSvc)

	srv := server.NewServer(cfg.Port, projectsSvc, docsSvc, boardsSvc, mcpSrv, cfg.MCPHTTPEnabled)

	// Stop the listener on SIGINT/SIGTERM and drain in-flight requests so REST
	// and MCP shut down cleanly (and the deferred db.Close runs).
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if cfg.MCPHTTPEnabled {
			log.Printf("MCP HTTP handler mounted at /mcp")
		}
		log.Printf("listening on :%s", cfg.Port)
		errCh <- srv.Start()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	case <-ctx.Done():
		stop() // restore default signal handling so a second signal force-quits
		log.Println("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("graceful shutdown: %v", err)
		}
	}
}
