package memory

import (
	"errors"
	"sort"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
)

type TasksRepository struct {
	items map[uuid.UUID]*models.Task
}

func NewTasksRepository() *TasksRepository {
	return &TasksRepository{
		items: make(map[uuid.UUID]*models.Task),
	}
}

func (r *TasksRepository) Save(item *models.Task) error {
	r.items[item.ID] = item
	return nil
}

func (r *TasksRepository) FindByID(id uuid.UUID) (*models.Task, error) {
	item, ok := r.items[id]
	if !ok {
		return nil, errors.New("task not found")
	}
	return item, nil
}

func (r *TasksRepository) FindByBoardID(boardID uuid.UUID) ([]*models.Task, error) {
	out := make([]*models.Task, 0)
	for _, item := range r.items {
		if item.BoardID == boardID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Position != out[j].Position {
			return out[i].Position < out[j].Position
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out, nil
}

func (r *TasksRepository) Delete(id uuid.UUID) error {
	delete(r.items, id)
	return nil
}
