package memory

import (
	"errors"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
)

type BoardsRepository struct {
	items map[uuid.UUID]*models.Board
}

func NewBoardsRepository() *BoardsRepository {
	return &BoardsRepository{
		items: make(map[uuid.UUID]*models.Board),
	}
}

func (r *BoardsRepository) Save(item *models.Board) error {
	r.items[item.ID] = item
	return nil
}

func (r *BoardsRepository) FindByID(id uuid.UUID) (*models.Board, error) {
	item, ok := r.items[id]
	if !ok {
		return nil, errors.New("board not found")
	}
	return item, nil
}

func (r *BoardsRepository) FindByProjectID(projectID uuid.UUID) (*models.Board, error) {
	for _, item := range r.items {
		if item.ProjectID == projectID {
			return item, nil
		}
	}
	return nil, nil
}
