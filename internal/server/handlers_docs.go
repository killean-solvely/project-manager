package server

import (
	"net/http"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	list, err := s.docs.ListDocuments(id)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, documentDTOs(list))
}

func (s *Server) handleMissingDocuments(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	missing, err := s.docs.GetMissingDocs(id)
	if err != nil {
		respondServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, missingDocDTOs(missing))
}

func (s *Server) handleUpsertDocument(w http.ResponseWriter, r *http.Request) {
	id, err := uuidParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	docType := models.DocumentType(chi.URLParam(r, "type"))
	var req upsertDocumentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	doc, err := s.docs.UpsertDocument(id, docType, req.Title, req.Content, models.DocumentStatus(req.Status))
	if err != nil {
		respondServiceError(w, err)
		return
	}
	var dto documentDTO
	dto.FromModel(doc)
	writeJSON(w, http.StatusOK, dto)
}
