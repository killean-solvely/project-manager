package sqlite

import (
	"database/sql"
	"errors"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
)

type DocumentsRepository struct {
	db *sql.DB
}

func NewDocumentsRepository(db *sql.DB) *DocumentsRepository {
	return &DocumentsRepository{db: db}
}

const documentColumns = `id, project_id, type, title, content, status, created_at, updated_at`

func (r *DocumentsRepository) Save(item *models.Document) error {
	_, err := r.db.Exec(`
		INSERT INTO documents (`+documentColumns+`)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			project_id=excluded.project_id, type=excluded.type, title=excluded.title,
			content=excluded.content, status=excluded.status,
			created_at=excluded.created_at, updated_at=excluded.updated_at`,
		item.ID.String(), item.ProjectID.String(), string(item.Type), item.Title,
		item.Content, string(item.Status), formatTime(item.CreatedAt), formatTime(item.UpdatedAt),
	)
	return err
}

func (r *DocumentsRepository) FindByID(id uuid.UUID) (*models.Document, error) {
	row := r.db.QueryRow(`SELECT `+documentColumns+` FROM documents WHERE id = ?`, id.String())
	d, err := scanDocument(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("document not found")
	}
	return d, err
}

func (r *DocumentsRepository) FindByProjectID(projectID uuid.UUID) ([]*models.Document, error) {
	rows, err := r.db.Query(`SELECT `+documentColumns+` FROM documents WHERE project_id = ? ORDER BY type`, projectID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*models.Document, 0)
	for rows.Next() {
		d, err := scanDocument(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *DocumentsRepository) FindByProjectIDAndType(projectID uuid.UUID, docType models.DocumentType) (*models.Document, error) {
	row := r.db.QueryRow(`SELECT `+documentColumns+` FROM documents WHERE project_id = ? AND type = ?`,
		projectID.String(), string(docType))
	d, err := scanDocument(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // absence is normal — the project hasn't written this type yet
	}
	return d, err
}

func (r *DocumentsRepository) Delete(id uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM documents WHERE id = ?`, id.String())
	return err
}

func scanDocument(s rowScanner) (*models.Document, error) {
	var idStr, projectID, docType, title, content, status, createdAt, updatedAt string
	if err := s.Scan(&idStr, &projectID, &docType, &title, &content, &status, &createdAt, &updatedAt); err != nil {
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
	return &models.Document{
		ID:        id,
		ProjectID: pid,
		Type:      models.DocumentType(docType),
		Title:     title,
		Content:   content,
		Status:    models.DocumentStatus(status),
		CreatedAt: created,
		UpdatedAt: updated,
	}, nil
}
