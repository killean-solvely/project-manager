package sqlite

import (
	"database/sql"
	"errors"

	"github.com/killeanjohnson/projectmanager/internal/models"

	"github.com/google/uuid"
)

type TasksRepository struct {
	db *sql.DB
}

func NewTasksRepository(db *sql.DB) *TasksRepository {
	return &TasksRepository{db: db}
}

const taskColumns = `id, board_id, column_id, title, description, priority, labels,
	document_id, position, completed_at, created_at, updated_at`

func (r *TasksRepository) Save(item *models.Task) error {
	_, err := r.db.Exec(`
		INSERT INTO tasks (`+taskColumns+`)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			board_id=excluded.board_id, column_id=excluded.column_id, title=excluded.title,
			description=excluded.description, priority=excluded.priority, labels=excluded.labels,
			document_id=excluded.document_id, position=excluded.position,
			completed_at=excluded.completed_at, created_at=excluded.created_at,
			updated_at=excluded.updated_at`,
		item.ID.String(), item.BoardID.String(), item.ColumnID.String(), item.Title,
		item.Description, string(item.Priority), encodeStrings(item.Labels),
		nullableUUID(item.DocumentID), item.Position, nullableTime(item.CompletedAt),
		formatTime(item.CreatedAt), formatTime(item.UpdatedAt),
	)
	return err
}

func (r *TasksRepository) FindByID(id uuid.UUID) (*models.Task, error) {
	row := r.db.QueryRow(`SELECT `+taskColumns+` FROM tasks WHERE id = ?`, id.String())
	t, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("task not found")
	}
	return t, err
}

func (r *TasksRepository) FindByBoardID(boardID uuid.UUID) ([]*models.Task, error) {
	rows, err := r.db.Query(`SELECT `+taskColumns+` FROM tasks WHERE board_id = ? ORDER BY position, created_at`, boardID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*models.Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TasksRepository) Delete(id uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM tasks WHERE id = ?`, id.String())
	return err
}

func scanTask(s rowScanner) (*models.Task, error) {
	var (
		idStr, boardID, columnID, title, description, priority, labels string
		documentID, completedAt                                        sql.NullString
		position                                                       int
		createdAt, updatedAt                                           string
	)
	if err := s.Scan(&idStr, &boardID, &columnID, &title, &description, &priority, &labels,
		&documentID, &position, &completedAt, &createdAt, &updatedAt); err != nil {
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
	cid, err := uuid.Parse(columnID)
	if err != nil {
		return nil, err
	}
	docID, err := parseNullableUUID(documentID)
	if err != nil {
		return nil, err
	}
	completed, err := parseNullableTime(completedAt)
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
	return &models.Task{
		ID:          id,
		BoardID:     bid,
		ColumnID:    cid,
		Title:       title,
		Description: description,
		Priority:    models.TaskPriority(priority),
		Labels:      decodeStrings(labels),
		DocumentID:  docID,
		Position:    position,
		CompletedAt: completed,
		CreatedAt:   created,
		UpdatedAt:   updated,
	}, nil
}
