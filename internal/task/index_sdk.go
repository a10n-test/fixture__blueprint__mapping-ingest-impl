package task

// IndexSdk is the deeper SDK component. Its SyncIndex method is the REALIZING op
// (sync_index): it carries its OWN ownership op marker plus its OWN co-located
// markers — a store :reads, a @projects invariant, and a @publishes event.
//
// When create_task's BFS walk reaches SyncIndex, the realizing-op attribution
// rule treats SyncIndex as a boundary_op_symbol (it bears a Components.* SDK
// ownership marker). So create_task gets ONLY the @calls edge to sync_index;
// sync_index's OWN @reads / @projects / @publishes are emitted by sync_index's
// OWN seed walk at depth 0 and must NOT bubble up to create_task. This is the
// exact bug the fix repairs (a shallow consumer absorbing a deep op's markers).
type IndexSdk struct {
	store Store
}

// SyncIndex syncs the task into the search index.
// a10n:blueprint Components.IndexSdk.Operations.sync_index
// a10n:blueprint Products.TaskEngine.Features.Indexing.index_synced
// a10n:blueprint Components.IndexSdk.Events.IndexEvents.synced
func (x *IndexSdk) SyncIndex(id string) error {
	// a10n:blueprint Components.TaskRelationalStore.TaskRow:reads
	_, err := x.store.GetTask(id)
	return err
}
