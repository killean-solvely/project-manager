package sqlite

import (
	"database/sql"
	"errors"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
)

type BoardsRepository struct {
	db *sql.DB
}

func NewBoardsRepository(db *sql.DB) *BoardsRepository {
	return &BoardsRepository{db: db}
}

const boardColumns = `id, project_id, name, created_at, updated_at`

func (r *BoardsRepository) Save(item *models.Board) error {
	_, err := r.db.Exec(`
		INSERT INTO boards (`+boardColumns+`)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			project_id=excluded.project_id, name=excluded.name,
			created_at=excluded.created_at, updated_at=excluded.updated_at`,
		item.ID.String(), item.ProjectID.String(), item.Name,
		formatTime(item.CreatedAt), formatTime(item.UpdatedAt),
	)
	return err
}

func (r *BoardsRepository) FindByID(id uuid.UUID) (*models.Board, error) {
	row := r.db.QueryRow(`SELECT `+boardColumns+` FROM boards WHERE id = ?`, id.String())
	b, err := scanBoard(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("board not found")
	}
	return b, err
}

func (r *BoardsRepository) FindByProjectID(projectID uuid.UUID) (*models.Board, error) {
	row := r.db.QueryRow(`SELECT `+boardColumns+` FROM boards WHERE project_id = ?`, projectID.String())
	b, err := scanBoard(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // no board yet — the boards service lazily creates one
	}
	return b, err
}

func scanBoard(s rowScanner) (*models.Board, error) {
	var idStr, projectID, name, createdAt, updatedAt string
	if err := s.Scan(&idStr, &projectID, &name, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, err
	}
	created, err := parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	updated, err := parseTime(updatedAt)
	if err != nil {
		return nil, err
	}
	return &models.Board{
		ID:        id,
		ProjectID: pid,
		Name:      name,
		CreatedAt: created,
		UpdatedAt: updated,
	}, nil
}
