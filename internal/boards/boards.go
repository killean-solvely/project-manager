package boards

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/killeanjohnson/projectmanager/internal/models"
	"github.com/killeanjohnson/projectmanager/internal/persistence"

	"github.com/google/uuid"
)

// doneColumnName is the default column whose tasks count as completed.
const doneColumnName = "Done"

// defaultColumns are created the first time a project's board is accessed.
var defaultColumns = []string{"Backlog", "Todo", "In progress", doneColumnName}

// Service owns the kanban side of a project: the board, its columns, and its
// tasks. Grouping them in one service keeps the board's invariants in one place.
type Service struct {
	boards  persistence.BoardsRepository
	columns persistence.ColumnsRepository
	tasks   persistence.TasksRepository
}

func NewService(
	boardsRepo persistence.BoardsRepository,
	columnsRepo persistence.ColumnsRepository,
	tasksRepo persistence.TasksRepository,
) *Service {
	return &Service{boards: boardsRepo, columns: columnsRepo, tasks: tasksRepo}
}

// GetBoardForProject returns the project's board, lazily creating it (with the
// default columns) the first time it is requested.
func (s *Service) GetBoardForProject(projectID uuid.UUID) (*models.Board, error) {
	board, err := s.boards.FindByProjectID(projectID)
	if err != nil {
		return nil, err
	}
	if board != nil {
		return board, nil
	}

	now := time.Now()
	board = &models.Board{
		ID:        uuid.New(),
		ProjectID: projectID,
		Name:      "Board",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.boards.Save(board); err != nil {
		return nil, err
	}
	for i, name := range defaultColumns {
		col := &models.Column{
			ID:        uuid.New(),
			BoardID:   board.ID,
			Name:      name,
			Position:  i,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.columns.Save(col); err != nil {
			return nil, err
		}
	}
	return board, nil
}

func (s *Service) ListColumns(boardID uuid.UUID) ([]*models.Column, error) {
	return s.columns.FindByBoardID(boardID)
}

// AddColumn appends a new column to the end of a board.
func (s *Service) AddColumn(boardID uuid.UUID, name string) (*models.Column, error) {
	if name == "" {
		return nil, errors.New("column name is required")
	}
	existing, err := s.columns.FindByBoardID(boardID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	col := &models.Column{
		ID:        uuid.New(),
		BoardID:   boardID,
		Name:      name,
		Position:  len(existing),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.columns.Save(col); err != nil {
		return nil, err
	}
	return col, nil
}

func (s *Service) RenameColumn(columnID uuid.UUID, name string) (*models.Column, error) {
	if name == "" {
		return nil, errors.New("column name is required")
	}
	col, err := s.columns.FindByID(columnID)
	if err != nil {
		return nil, err
	}
	col.Name = name
	col.UpdatedAt = time.Now()
	if err := s.columns.Save(col); err != nil {
		return nil, err
	}
	return col, nil
}

// RemoveColumn deletes an empty column. Columns that still hold tasks cannot be
// removed (move or delete the tasks first).
func (s *Service) RemoveColumn(columnID uuid.UUID) error {
	col, err := s.columns.FindByID(columnID)
	if err != nil {
		return err
	}
	tasks, err := s.tasks.FindByBoardID(col.BoardID)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if t.ColumnID == columnID {
			return errors.New("cannot remove a column that still has tasks")
		}
	}
	return s.columns.Delete(columnID)
}

func (s *Service) ListTasks(boardID uuid.UUID) ([]*models.Task, error) {
	return s.tasks.FindByBoardID(boardID)
}

func (s *Service) GetTask(id uuid.UUID) (*models.Task, error) {
	return s.tasks.FindByID(id)
}

// CreateTask adds a card to a column. The column must belong to the given board.
func (s *Service) CreateTask(boardID, columnID uuid.UUID, title, description string, priority models.TaskPriority, labels []string, documentID *uuid.UUID) (*models.Task, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}
	col, err := s.columns.FindByID(columnID)
	if err != nil {
		return nil, err
	}
	if col.BoardID != boardID {
		return nil, errors.New("column does not belong to this board")
	}
	if priority == "" {
		priority = models.TaskPriorityNone
	}
	if !validPriority(priority) {
		return nil, fmt.Errorf("invalid priority %q", priority)
	}
	if labels == nil {
		labels = []string{}
	}

	tasks, err := s.tasks.FindByBoardID(boardID)
	if err != nil {
		return nil, err
	}
	position := 0
	for _, t := range tasks {
		if t.ColumnID == columnID {
			position++
		}
	}

	now := time.Now()
	task := &models.Task{
		ID:          uuid.New(),
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       title,
		Description: description,
		Priority:    priority,
		Labels:      labels,
		DocumentID:  documentID,
		Position:    position,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.tasks.Save(task); err != nil {
		return nil, err
	}
	return task, nil
}

// UpdateTask applies a partial update to a task's descriptive fields. Use
// MoveTask to change which column a task is in.
func (s *Service) UpdateTask(id uuid.UUID, title, description *string, priority *models.TaskPriority, labels *[]string) (*models.Task, error) {
	task, err := s.tasks.FindByID(id)
	if err != nil {
		return nil, err
	}
	if title != nil {
		if *title == "" {
			return nil, errors.New("title cannot be empty")
		}
		task.Title = *title
	}
	if description != nil {
		task.Description = *description
	}
	if priority != nil {
		if !validPriority(*priority) {
			return nil, fmt.Errorf("invalid priority %q", *priority)
		}
		task.Priority = *priority
	}
	if labels != nil {
		task.Labels = *labels
	}
	task.UpdatedAt = time.Now()
	if err := s.tasks.Save(task); err != nil {
		return nil, err
	}
	return task, nil
}

// MoveTask moves a task to a column at a position, inserting it there and
// renumbering that column so positions stay contiguous (0..n-1). That contiguous
// renumber is what makes drag-reordering land correctly — within a column or
// across columns. Moving into a "Done" column stamps CompletedAt; moving out clears it.
func (s *Service) MoveTask(taskID, columnID uuid.UUID, position int) (*models.Task, error) {
	task, err := s.tasks.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	col, err := s.columns.FindByID(columnID)
	if err != nil {
		return nil, err
	}
	if col.BoardID != task.BoardID {
		return nil, errors.New("target column belongs to a different board")
	}

	// Destination column order (sorted by position), excluding the moved task,
	// then insert it at the requested index.
	boardTasks, err := s.tasks.FindByBoardID(task.BoardID)
	if err != nil {
		return nil, err
	}
	dest := make([]*models.Task, 0, len(boardTasks))
	for _, t := range boardTasks {
		if t.ColumnID == columnID && t.ID != taskID {
			dest = append(dest, t)
		}
	}
	if position < 0 {
		position = 0
	}
	if position > len(dest) {
		position = len(dest)
	}
	dest = append(dest, nil)
	copy(dest[position+1:], dest[position:])
	dest[position] = task

	now := time.Now()
	task.ColumnID = columnID
	if strings.EqualFold(col.Name, doneColumnName) {
		task.CompletedAt = &now
	} else {
		task.CompletedAt = nil
	}
	task.UpdatedAt = now

	// Renumber the destination column. The moved task is always saved; siblings
	// only when their position actually changed. (The source column's remaining
	// tasks keep their relative order, so they need no renumber.)
	for i, t := range dest {
		if t.ID == taskID {
			t.Position = i
			if err := s.tasks.Save(t); err != nil {
				return nil, err
			}
			continue
		}
		if t.Position != i {
			t.Position = i
			if err := s.tasks.Save(t); err != nil {
				return nil, err
			}
		}
	}
	return task, nil
}

// LinkDocument attaches (or, with nil, detaches) a document to a task.
func (s *Service) LinkDocument(taskID uuid.UUID, documentID *uuid.UUID) (*models.Task, error) {
	task, err := s.tasks.FindByID(taskID)
	if err != nil {
		return nil, err
	}
	task.DocumentID = documentID
	task.UpdatedAt = time.Now()
	if err := s.tasks.Save(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Service) DeleteTask(id uuid.UUID) error {
	if _, err := s.tasks.FindByID(id); err != nil {
		return err
	}
	return s.tasks.Delete(id)
}

func validPriority(p models.TaskPriority) bool {
	switch p {
	case models.TaskPriorityNone, models.TaskPriorityLow, models.TaskPriorityMedium, models.TaskPriorityHigh:
		return true
	}
	return false
}
