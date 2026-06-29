package server

import (
	"net/http"

	"github.com/killeanjohnson/projectmanager/internal/models"
)

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	var statusFilter *models.ProjectStatus
	if q := r.URL.Query().Get("status"); q != "" {
		st := models.ProjectStatus(q)
		if !validProjectStatus(st) {
			writeError(w, http.StatusBadRequest, "invalid status filter")
			return
		}
		statusFilter = &st
	}
	list, err := s.projects.ListProjects(statusFilter)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projectDTOs(list))
}

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	p, err := s.projects.CreateIdea(req.Name, req.Summary, req.Description, req.Tags)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	var dto projectDTO
	dto.FromModel(p)
	writeJSON(w, http.StatusCreated, dto)
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	p, err := s.projects.GetProject(id)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	var dto projectDTO
	dto.FromModel(p)
	writeJSON(w, http.StatusOK, dto)
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	var req updateProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	p, err := s.projects.UpdateDetails(id, req.Name, req.Summary, req.Description, req.Tags)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	var dto projectDTO
	dto.FromModel(p)
	writeJSON(w, http.StatusOK, dto)
}

func (s *Server) handlePromoteProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	var req promoteRequest
	if err := decodeJSONAllowEmpty(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	p, err := s.projects.PromoteToActive(id, models.ProjectMode(req.Mode))
	if err != nil {
		respondServiceError(w, err)
		return
	}
	var dto projectDTO
	dto.FromModel(p)
	writeJSON(w, http.StatusOK, dto)
}

func (s *Server) handleSetMode(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	var req setModeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	p, err := s.projects.SetMode(id, models.ProjectMode(req.Mode))
	if err != nil {
		respondServiceError(w, err)
		return
	}
	var dto projectDTO
	dto.FromModel(p)
	writeJSON(w, http.StatusOK, dto)
}

func (s *Server) handleArchiveProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	var req archiveRequest
	if err := decodeJSONAllowEmpty(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	p, err := s.projects.Archive(id, req.Reason)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	var dto projectDTO
	dto.FromModel(p)
	writeJSON(w, http.StatusOK, dto)
}

func (s *Server) handleReviveProject(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	var req reviveRequest
	if err := decodeJSONAllowEmpty(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	p, err := s.projects.Revive(id, models.ProjectMode(req.Mode))
	if err != nil {
		respondServiceError(w, err)
		return
	}
	var dto projectDTO
	dto.FromModel(p)
	writeJSON(w, http.StatusOK, dto)
}

func validProjectStatus(s models.ProjectStatus) bool {
	switch s {
	case models.ProjectStatusIdea, models.ProjectStatusActive, models.ProjectStatusArchived:
		return true
	}
	return false
}
