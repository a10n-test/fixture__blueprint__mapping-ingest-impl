package task

import "errors"

var ErrNotFound = errors.New("not found")

type TaskRow struct {
	ID    string
	Title string
	Done  bool
}

type Store interface {
	CreateTask(t TaskRow) error
	GetTask(id string) (*TaskRow, error)
	UpdateTask(id string, done bool) error
}

type Service struct {
	store Store
}

func New(store Store) *Service {
	return &Service{store: store}
}

// CreateTask creates a new task.
//
// Through the unmarked helper reconcile, create_task reaches BOTH the marked
// get_task op (suppressed below) and the marked update_task op (not suppressed).
// a10n:blueprint Components.TaskService.Commands.create_task
// a10n:blueprint:ignore-call Components.TaskService.Commands.get_task
func (s *Service) CreateTask(title string) (string, error) {
	id := generateID()
	s.reconcile(id)
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:writes
	return id, s.store.CreateTask(TaskRow{ID: id, Title: title})
}

// reconcile is an UNMARKED helper that fans out to the two marked ops. The
// indexer cannot trace same-receiver method calls, so the edges are declared
// explicitly with a10n:index:call (the deterministic escape hatch for
// AST-untraceable call edges).
// a10n:index:call task.Service.GetTask
// a10n:index:call task.Service.UpdateTask
func (s *Service) reconcile(id string) {
	if t, err := s.GetTask(id); err == nil && t != nil {
		_ = s.UpdateTask(id, t.Done)
	}
}

// GetTask retrieves a task by ID.
// a10n:blueprint Components.TaskService.Commands.get_task
func (s *Service) GetTask(id string) (*TaskRow, error) {
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:reads
	return s.store.GetTask(id)
}

// UpdateTask marks a task done or not.
// a10n:blueprint Components.TaskService.Commands.update_task
func (s *Service) UpdateTask(id string, done bool) error {
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:writes
	return s.store.UpdateTask(id, done)
}

func generateID() string { return "task-" + "001" }
