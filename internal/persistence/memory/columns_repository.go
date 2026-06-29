package memory

import (
	"errors"
	"sort"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
)

type ColumnsRepository struct {
	items map[uuid.UUID]*models.Column
}

func NewColumnsRepository() *ColumnsRepository {
	return &ColumnsRepository{
		items: make(map[uuid.UUID]*models.Column),
	}
}

func (r *ColumnsRepository) Save(item *models.Column) error {
	r.items[item.ID] = item
	return nil
}

func (r *ColumnsRepository) FindByID(id uuid.UUID) (*models.Column, error) {
	item, ok := r.items[id]
	if !ok {
		return nil, errors.New("column not found")
	}
	return item, nil
}

func (r *ColumnsRepository) FindByBoardID(boardID uuid.UUID) ([]*models.Column, error) {
	out := make([]*models.Column, 0)
	for _, item := range r.items {
		if item.BoardID == boardID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Position < out[j].Position
	})
	return out, nil
}

func (r *ColumnsRepository) Delete(id uuid.UUID) error {
	delete(r.items, id)
	return nil
}
