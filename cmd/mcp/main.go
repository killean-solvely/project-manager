package main

import (
	"context"
	"log"

	"github.com/killeanjohnson/projectmanager/internal/boards"
	"github.com/killeanjohnson/projectmanager/internal/docs"
	"github.com/killeanjohnson/projectmanager/internal/mcpserver"
	"github.com/killeanjohnson/projectmanager/internal/persistence/sqlite"
	"github.com/killeanjohnson/projectmanager/internal/projects"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Same DB path as cmd/api by default (PM_DB_PATH, else ~/.projectmanager/...),
	// so the API and MCP server share one store.
	db, path, err := sqlite.Open()
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	projectsRepo := sqlite.NewProjectsRepository(db)
	documentsRepo := sqlite.NewDocumentsRepository(db)
	boardsRepo := sqlite.NewBoardsRepository(db)
	columnsRepo := sqlite.NewColumnsRepository(db)
	tasksRepo := sqlite.NewTasksRepository(db)

	projectsSvc := projects.NewService(projectsRepo)
	docsSvc := docs.NewService(documentsRepo)
	boardsSvc := boards.NewService(boardsRepo, columnsRepo, tasksRepo)

	srv := mcpserver.New(projectsSvc, docsSvc, boardsSvc)

	// log goes to stderr, so it won't corrupt the stdio JSON-RPC stream on stdout.
	log.Printf("projectmanager MCP server running on stdio (db: %s)", path)
	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
