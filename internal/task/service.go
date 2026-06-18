package task

import "errors"

var ErrNotFound = errors.New("not found")

type TaskRow struct {
	ID    string
	Title string
	Done  bool
}

// CreateTask is the marked entry point for Components.TaskService.Commands.create_task.
//
// It natively calls writeRow (a connected store write whose marker matches the
// spec @writes(TaskRow) edge) and auditWrite (a symbol marked AuditRow, which
// create_task does NOT declare -> an ORPHAN branch). It does NOT call any
// get_task-marked symbol, so the spec's @calls(get_task) declaration is a BROKEN
// link with no impl marker.
//
// a10n:blueprint Components.TaskService.Commands.create_task
func CreateTask(title string) string {
	id := generateID()
	writeRow(TaskRow{ID: id, Title: title})
	auditWrite(id)
	return id
}

// writeRow is the connected store write — its marker matches the spec @writes edge.
//
// a10n:blueprint Components.TaskRelationalStore.TaskRow:writes
func writeRow(t TaskRow) { _ = t }

// auditWrite is marked AuditRow. CreateTask reaches it natively, but create_task's
// spec declares no edge to AuditRow -> ORPHAN (impl_marker_not_in_spec).
//
// a10n:blueprint Components.TaskRelationalStore.AuditRow:writes
func auditWrite(id string) { _ = id }

func generateID() string { return "task-001" }
