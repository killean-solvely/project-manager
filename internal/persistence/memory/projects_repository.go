package memory

import (
	"errors"
	"sort"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
)

type ProjectsRepository struct {
	items map[uuid.UUID]*models.Project
}

func NewProjectsRepository() *ProjectsRepository {
	return &ProjectsRepository{
		items: make(map[uuid.UUID]*models.Project),
	}
}

func (r *ProjectsRepository) Save(item *models.Project) error {
	r.items[item.ID] = item
	return nil
}

func (r *ProjectsRepository) FindByID(id uuid.UUID) (*models.Project, error) {
	item, ok := r.items[id]
	if !ok {
		return nil, errors.New("project not found")
	}
	return item, nil
}

func (r *ProjectsRepository) FindAll() ([]*models.Project, error) {
	out := make([]*models.Project, 0, len(r.items))
	for _, item := range r.items {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out, nil
}

func (r *ProjectsRepository) Delete(id uuid.UUID) error {
	delete(r.items, id)
	return nil
}
