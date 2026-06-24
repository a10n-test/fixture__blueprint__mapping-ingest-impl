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
// a10n:blueprint Components.TaskService.Commands.create_task
// a10n:blueprint Products.TaskEngine.Features.TaskLifecycle.task_persisted
// a10n:blueprint:call Components.TaskService.Http.ping
func (s *Service) CreateTask(title string) (string, error) {
	id := generateID()
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:reads
	t, err := s.store.GetTask(id) // check idempotency
	if err == nil && t != nil {
		return t.ID, nil
	}
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:writes
	return id, s.store.CreateTask(TaskRow{ID: id, Title: title})
}

// GetTask retrieves a task by ID.
// a10n:blueprint Components.TaskService.Commands.get_task
// a10n:blueprint Products.TaskEngine.Features.TaskLifecycle.task_retrievable
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
