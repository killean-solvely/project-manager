package server

import (
	"github.com/go-chi/chi/v5"
)

// registerRoutes wires HTTP routes to handler methods on Server, under /api.
func (s *Server) registerRoutes(mux chi.Router) {
	mux.Route("/api", func(r chi.Router) {
		// Portfolio + lifecycle.
		r.Get("/projects", s.handleListProjects)
		r.Post("/projects", s.handleCreateProject)
		r.Get("/projects/{id}", s.handleGetProject)
		r.Patch("/projects/{id}", s.handleUpdateProject)
		r.Post("/projects/{id}/promote", s.handlePromoteProject)
		r.Post("/projects/{id}/mode", s.handleSetMode)
		r.Post("/projects/{id}/archive", s.handleArchiveProject)
		r.Post("/projects/{id}/revive", s.handleReviveProject)

		// Docs library.
		r.Get("/projects/{id}/documents", s.handleListDocuments)
		r.Get("/projects/{id}/documents/missing", s.handleMissingDocuments)
		r.Put("/projects/{id}/documents/{type}", s.handleUpsertDocument)

		// Board + tasks.
		r.Get("/projects/{id}/board", s.handleGetBoard)
		r.Get("/projects/{id}/tasks", s.handleListTasks)
		r.Post("/projects/{id}/tasks", s.handleCreateTask)
		r.Patch("/tasks/{id}", s.handleUpdateTask)
		r.Post("/tasks/{id}/move", s.handleMoveTask)
		r.Post("/tasks/{id}/link", s.handleLinkTaskDocument)
		r.Delete("/tasks/{id}", s.handleDeleteTask)
	})
}
