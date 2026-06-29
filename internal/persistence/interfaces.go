package persistence

import (
	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
)

// Repository interfaces — one per domain. Methods take and return model pointers
// only. Concrete implementations live under persistence/memory (and later sqlite).

type ProjectsRepository interface {
	Save(item *models.Project) error
	FindByID(id uuid.UUID) (*models.Project, error)
	FindAll() ([]*models.Project, error)
	Delete(id uuid.UUID) error
}

type DocumentsRepository interface {
	Save(item *models.Document) error
	FindByID(id uuid.UUID) (*models.Document, error)
	FindByProjectID(projectID uuid.UUID) ([]*models.Document, error)
	// FindByProjectIDAndType returns (nil, nil) when no such document exists yet —
	// absence is normal (the project simply hasn't written that doc type).
	FindByProjectIDAndType(projectID uuid.UUID, docType models.DocumentType) (*models.Document, error)
	Delete(id uuid.UUID) error
}

type BoardsRepository interface {
	Save(item *models.Board) error
	FindByID(id uuid.UUID) (*models.Board, error)
	// FindByProjectID returns (nil, nil) when the project has no board yet, so the
	// boards service can lazily create one on first access.
	FindByProjectID(projectID uuid.UUID) (*models.Board, error)
}

type ColumnsRepository interface {
	Save(item *models.Column) error
	FindByID(id uuid.UUID) (*models.Column, error)
	FindByBoardID(boardID uuid.UUID) ([]*models.Column, error)
	Delete(id uuid.UUID) error
}

type TasksRepository interface {
	Save(item *models.Task) error
	FindByID(id uuid.UUID) (*models.Task, error)
	FindByBoardID(boardID uuid.UUID) ([]*models.Task, error)
	Delete(id uuid.UUID) error
}
