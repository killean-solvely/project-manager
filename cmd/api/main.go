package main

import (
	"log"

	"github.com/killeanjohnson/projectmanager/internal/boards"
	"github.com/killeanjohnson/projectmanager/internal/config"
	"github.com/killeanjohnson/projectmanager/internal/docs"
	"github.com/killeanjohnson/projectmanager/internal/persistence/sqlite"
	"github.com/killeanjohnson/projectmanager/internal/projects"
	"github.com/killeanjohnson/projectmanager/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := sqlite.OpenAt(cfg.DBPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()
	log.Printf("using database at %s", cfg.DBPath)

	projectsRepo := sqlite.NewProjectsRepository(db)
	documentsRepo := sqlite.NewDocumentsRepository(db)
	boardsRepo := sqlite.NewBoardsRepository(db)
	columnsRepo := sqlite.NewColumnsRepository(db)
	tasksRepo := sqlite.NewTasksRepository(db)

	projectsSvc := projects.NewService(projectsRepo)
	docsSvc := docs.NewService(documentsRepo)
	boardsSvc := boards.NewService(boardsRepo, columnsRepo, tasksRepo)

	srv := server.NewServer(cfg.Port, projectsSvc, docsSvc, boardsSvc)
	log.Printf("listening on :%s", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
