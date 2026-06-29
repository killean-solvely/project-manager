package boards_test

import (
	"slices"
	"testing"

	"github.com/killeanjohnson/projectmanager/internal/boards"
	"github.com/killeanjohnson/projectmanager/internal/persistence/memory"

	"github.com/google/uuid"
)

func newService() *boards.Service {
	return boards.NewService(
		memory.NewBoardsRepository(),
		memory.NewColumnsRepository(),
		memory.NewTasksRepository(),
	)
}

func columnTitles(t *testing.T, s *boards.Service, boardID, columnID uuid.UUID) []string {
	t.Helper()
	tasks, err := s.ListTasks(boardID)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}
	var out []string
	for _, tk := range tasks {
		if tk.ColumnID == columnID {
			out = append(out, tk.Title)
		}
	}
	return out
}

func TestMoveTaskReordersWithinColumn(t *testing.T) {
	s := newService()
	board, err := s.GetBoardForProject(uuid.New())
	if err != nil {
		t.Fatalf("board: %v", err)
	}
	cols, _ := s.ListColumns(board.ID)
	todo := cols[1] // Backlog, Todo, In progress, Done

	for _, name := range []string{"A", "B", "C"} {
		if _, err := s.CreateTask(board.ID, todo.ID, name, "", "", nil, nil); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}
	cID := findByTitle(t, s, board.ID, "C")

	// Move C to the front.
	if _, err := s.MoveTask(cID, todo.ID, 0); err != nil {
		t.Fatalf("move: %v", err)
	}
	if got := columnTitles(t, s, board.ID, todo.ID); !slices.Equal(got, []string{"C", "A", "B"}) {
		t.Fatalf("want [C A B], got %v", got)
	}

	// Move C to the end.
	if _, err := s.MoveTask(cID, todo.ID, 2); err != nil {
		t.Fatalf("move: %v", err)
	}
	if got := columnTitles(t, s, board.ID, todo.ID); !slices.Equal(got, []string{"A", "B", "C"}) {
		t.Fatalf("want [A B C], got %v", got)
	}
}

func TestMoveTaskAcrossColumns(t *testing.T) {
	s := newService()
	board, _ := s.GetBoardForProject(uuid.New())
	cols, _ := s.ListColumns(board.ID)
	todo, done := cols[1], cols[3]

	for _, name := range []string{"A", "B"} {
		if _, err := s.CreateTask(board.ID, todo.ID, name, "", "", nil, nil); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}
	aID := findByTitle(t, s, board.ID, "A")

	moved, err := s.MoveTask(aID, done.ID, 0)
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	if moved.ColumnID != done.ID {
		t.Fatalf("column not updated")
	}
	if moved.CompletedAt == nil {
		t.Fatalf("moving into Done should set completed_at")
	}
	if got := columnTitles(t, s, board.ID, todo.ID); !slices.Equal(got, []string{"B"}) {
		t.Fatalf("todo want [B], got %v", got)
	}
	if got := columnTitles(t, s, board.ID, done.ID); !slices.Equal(got, []string{"A"}) {
		t.Fatalf("done want [A], got %v", got)
	}
}

func findByTitle(t *testing.T, s *boards.Service, boardID uuid.UUID, title string) uuid.UUID {
	t.Helper()
	tasks, err := s.ListTasks(boardID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, tk := range tasks {
		if tk.Title == title {
			return tk.ID
		}
	}
	t.Fatalf("task %q not found", title)
	return uuid.Nil
}
