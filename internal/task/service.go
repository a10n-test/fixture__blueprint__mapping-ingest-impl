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
	store    Store
	handlers map[string]func(string) error
}

func New(store Store) *Service {
	s := &Service{store: store}
	s.handlers = map[string]func(string) error{
		"audit": s.audit,
	}
	return s
}

// CreateTask dispatches through a registry map to reach a marked helper.
// The AST cannot trace handlers["audit"](id) to s.audit, so ingest cannot reach
// the store :reads marker that s.audit carries — no @reads must be fabricated on
// create_task. create_task only directly performs a store :writes.
// a10n:blueprint Components.TaskService.Commands.create_task
// a10n:blueprint Products.TaskEngine.Features.TaskLifecycle.task_persisted
func (s *Service) CreateTask(title string) (string, error) {
	id := generateID()
	if h, ok := s.handlers["audit"]; ok {
		_ = h(id)
	}
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:writes
	return id, s.store.CreateTask(TaskRow{ID: id, Title: title})
}

// audit is reached ONLY via the registry map dispatch (handlers["audit"]). It
// carries a store :reads marker. Because the dispatch is indirect, the AST walk
// from create_task never reaches this marker, so create_task must NOT gain a
// @reads edge from it. Mapping check (not ingest) is responsible for flagging a
// spec @reads that has no traceable impl path.
func (s *Service) audit(id string) error {
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:reads
	_, err := s.store.GetTask(id)
	return err
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
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:reads
	if _, err := s.store.GetTask(id); err != nil {
		return err
	}
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:writes
	return s.store.UpdateTask(id, done)
}

func generateID() string { return "task-" + "001" }
