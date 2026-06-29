package sqlite

import (
	"database/sql"
	"errors"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
)

type ProjectsRepository struct {
	db *sql.DB
}

func NewProjectsRepository(db *sql.DB) *ProjectsRepository {
	return &ProjectsRepository{db: db}
}

const projectColumns = `id, name, summary, description, status, mode, tags,
	archived_reason, promoted_at, archived_at, created_at, updated_at`

func (r *ProjectsRepository) Save(item *models.Project) error {
	_, err := r.db.Exec(`
		INSERT INTO projects (`+projectColumns+`)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name=excluded.name, summary=excluded.summary, description=excluded.description,
			status=excluded.status, mode=excluded.mode, tags=excluded.tags,
			archived_reason=excluded.archived_reason, promoted_at=excluded.promoted_at,
			archived_at=excluded.archived_at, created_at=excluded.created_at,
			updated_at=excluded.updated_at`,
		item.ID.String(), item.Name, item.Summary, item.Description,
		string(item.Status), string(item.Mode), encodeStrings(item.Tags),
		item.ArchivedReason, nullableTime(item.PromotedAt), nullableTime(item.ArchivedAt),
		formatTime(item.CreatedAt), formatTime(item.UpdatedAt),
	)
	return err
}

func (r *ProjectsRepository) FindByID(id uuid.UUID) (*models.Project, error) {
	row := r.db.QueryRow(`SELECT `+projectColumns+` FROM projects WHERE id = ?`, id.String())
	p, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("project not found")
	}
	return p, err
}

func (r *ProjectsRepository) FindAll() ([]*models.Project, error) {
	rows, err := r.db.Query(`SELECT ` + projectColumns + ` FROM projects ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*models.Project, 0)
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProjectsRepository) Delete(id uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM projects WHERE id = ?`, id.String())
	return err
}

func scanProject(s rowScanner) (*models.Project, error) {
	var (
		idStr, name, summary, description, status, mode, tags, archivedReason string
		promotedAt, archivedAt                                                sql.NullString
		createdAt, updatedAt                                                  string
	)
	if err := s.Scan(&idStr, &name, &summary, &description, &status, &mode, &tags,
		&archivedReason, &promotedAt, &archivedAt, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	id, err := uuid.Parse(idStr)
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
	promoted, err := parseNullableTime(promotedAt)
	if err != nil {
		return nil, err
	}
	archived, err := parseNullableTime(archivedAt)
	if err != nil {
		return nil, err
	}

	return &models.Project{
		ID:             id,
		Name:           name,
		Summary:        summary,
		Description:    description,
		Status:         models.ProjectStatus(status),
		Mode:           models.ProjectMode(mode),
		Tags:           decodeStrings(tags),
		ArchivedReason: archivedReason,
		PromotedAt:     promoted,
		ArchivedAt:     archived,
		CreatedAt:      created,
		UpdatedAt:      updated,
	}, nil
}
