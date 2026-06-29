package sqlite

import (
	"database/sql"
	"errors"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
)

type ColumnsRepository struct {
	db *sql.DB
}

func NewColumnsRepository(db *sql.DB) *ColumnsRepository {
	return &ColumnsRepository{db: db}
}

const columnColumns = `id, board_id, name, position, created_at, updated_at`

func (r *ColumnsRepository) Save(item *models.Column) error {
	_, err := r.db.Exec(`
		INSERT INTO board_columns (`+columnColumns+`)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			board_id=excluded.board_id, name=excluded.name, position=excluded.position,
			created_at=excluded.created_at, updated_at=excluded.updated_at`,
		item.ID.String(), item.BoardID.String(), item.Name, item.Position,
		formatTime(item.CreatedAt), formatTime(item.UpdatedAt),
	)
	return err
}

func (r *ColumnsRepository) FindByID(id uuid.UUID) (*models.Column, error) {
	row := r.db.QueryRow(`SELECT `+columnColumns+` FROM board_columns WHERE id = ?`, id.String())
	c, err := scanColumn(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("column not found")
	}
	return c, err
}

func (r *ColumnsRepository) FindByBoardID(boardID uuid.UUID) ([]*models.Column, error) {
	rows, err := r.db.Query(`SELECT `+columnColumns+` FROM board_columns WHERE board_id = ? ORDER BY position`, boardID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*models.Column, 0)
	for rows.Next() {
		c, err := scanColumn(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ColumnsRepository) Delete(id uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM board_columns WHERE id = ?`, id.String())
	return err
}

func scanColumn(s rowScanner) (*models.Column, error) {
	var idStr, boardID, name, createdAt, updatedAt string
	var position int
	if err := s.Scan(&idStr, &boardID, &name, &position, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}
	bid, err := uuid.Parse(boardID)
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
	return &models.Column{
		ID:        id,
		BoardID:   bid,
		Name:      name,
		Position:  position,
		CreatedAt: created,
		UpdatedAt: updated,
	}, nil
}
