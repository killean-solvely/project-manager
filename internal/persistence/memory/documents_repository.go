package memory

import (
	"errors"
	"sort"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
)

type DocumentsRepository struct {
	items map[uuid.UUID]*models.Document
}

func NewDocumentsRepository() *DocumentsRepository {
	return &DocumentsRepository{
		items: make(map[uuid.UUID]*models.Document),
	}
}

func (r *DocumentsRepository) Save(item *models.Document) error {
	r.items[item.ID] = item
	return nil
}

func (r *DocumentsRepository) FindByID(id uuid.UUID) (*models.Document, error) {
	item, ok := r.items[id]
	if !ok {
		return nil, errors.New("document not found")
	}
	return item, nil
}

func (r *DocumentsRepository) FindByProjectID(projectID uuid.UUID) ([]*models.Document, error) {
	out := make([]*models.Document, 0)
	for _, item := range r.items {
		if item.ProjectID == projectID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Type < out[j].Type
	})
	return out, nil
}

func (r *DocumentsRepository) FindByProjectIDAndType(projectID uuid.UUID, docType models.DocumentType) (*models.Document, error) {
	for _, item := range r.items {
		if item.ProjectID == projectID && item.Type == docType {
			return item, nil
		}
	}
	return nil, nil
}

func (r *DocumentsRepository) Delete(id uuid.UUID) error {
	delete(r.items, id)
	return nil
}
