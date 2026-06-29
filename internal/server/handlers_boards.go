package server

import (
	"net/http"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
)

func (s *Server) handleGetBoard(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	board, err := s.boards.GetBoardForProject(id)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	cols, err := s.boards.ListColumns(board.ID)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newBoardDTO(board, cols))
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	board, err := s.boards.GetBoardForProject(id)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	tasks, err := s.boards.ListTasks(board.ID)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, taskDTOs(tasks))
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	var req createTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	columnID, err := uuid.Parse(req.ColumnID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid column id")
		return
	}
	var documentID *uuid.UUID
	if req.DocumentID != nil && *req.DocumentID != "" {
		did, err := uuid.Parse(*req.DocumentID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid document id")
			return
		}
		documentID = &did
	}
	board, err := s.boards.GetBoardForProject(id)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	task, err := s.boards.CreateTask(board.ID, columnID, req.Title, req.Description, models.TaskPriority(req.Priority), req.Labels, documentID)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	var dto taskDTO
	dto.FromModel(task)
	writeJSON(w, http.StatusCreated, dto)
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	var req updateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	var priority *models.TaskPriority
	if req.Priority != nil {
		p := models.TaskPriority(*req.Priority)
		priority = &p
	}
	task, err := s.boards.UpdateTask(id, req.Title, req.Description, priority, req.Labels)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	var dto taskDTO
	dto.FromModel(task)
	writeJSON(w, http.StatusOK, dto)
}

func (s *Server) handleMoveTask(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	var req moveTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	columnID, err := uuid.Parse(req.ColumnID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid column id")
		return
	}
	task, err := s.boards.MoveTask(id, columnID, req.Position)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	var dto taskDTO
	dto.FromModel(task)
	writeJSON(w, http.StatusOK, dto)
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	if err := s.boards.DeleteTask(id); err != nil {
		respondServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
