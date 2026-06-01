package task

import (
	"errors"
	"net/http"
)

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
	store  Store
	client *http.Client
}

func New(store Store) *Service {
	return &Service{store: store, client: http.DefaultClient}
}

// CreateTask creates a new task.
// a10n:blueprint Components.TaskService.Commands.create_task
func (s *Service) CreateTask(title string) (string, error) {
	id := generateID()
	_ = s.fetchOther(id)
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:writes
	return id, s.store.CreateTask(TaskRow{ID: id, Title: title})
}

// fetchOther is an HTTP CLIENT wrapper: it makes an outbound network request to
// another service. The marker names the boundary op it crosses to.
// a10n:blueprint Components.OtherService.Http.fetch
func (s *Service) fetchOther(id string) error {
	resp, err := s.client.Get("http://other/" + id)
	if err != nil {
		return err
	}
	return resp.Body.Close()
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
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:reads
	if _, err := s.store.GetTask(id); err != nil {
		return err
	}
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:writes
	return s.store.UpdateTask(id, done)
}

func generateID() string { return "task-" + "001" }
