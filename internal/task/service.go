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

// IndexEntry is the row the index SDK reads while syncing.
type IndexEntry struct {
	TaskID string
	Term   string
}

type Service struct {
	store Store
	index *IndexSdk
}

func New(store Store) *Service {
	return &Service{store: store, index: &IndexSdk{store: store}}
}

// CreateTask is the SHALLOW CONSUMER (seed) op. It directly performs a store
// :writes AND @calls into the deeper SDK op sync_index (an AST-traceable call to
// s.index.SyncIndex). The realizing-op attribution rule means create_task must
// get a @calls edge to sync_index but must NOT absorb sync_index's OWN markers
// (its @projects invariant, its store :reads, its @publishes event) — those
// belong to sync_index, the nearest enclosing ownership op. create_task DOES,
// however, absorb the @projects of the op-LESS helper auditLog it reaches
// (accepted op-less fan-out).
// a10n:blueprint Components.TaskService.Commands.create_task
// a10n:blueprint Products.TaskEngine.Features.TaskLifecycle.task_persisted
func (s *Service) CreateTask(title string) (string, error) {
	id := generateID()
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:writes
	if err := s.store.CreateTask(TaskRow{ID: id, Title: title}); err != nil {
		return "", err
	}
	s.auditLog(id)            // op-less helper — its @projects fans out to create_task
	return id, s.index.SyncIndex(id) // @calls into the SDK op sync_index
}

// auditLog is an op-LESS helper (NO ownership op marker). It carries a @projects
// invariant. Because it is not an ownership op, the realizing-op rule does NOT
// re-root here: its @projects legitimately fans out to every seed that reaches
// it (here, create_task). This is the ACCEPTED op-less fan-out the fix must NOT
// change.
// a10n:blueprint Products.TaskEngine.Features.TaskLifecycle.audit_recorded
func (s *Service) auditLog(id string) {
	_ = id
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
