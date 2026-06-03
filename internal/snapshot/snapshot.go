package snapshot

// This file models the AGENT-GRAPH redundant-behind topology for the ingest
// graph-node strip (work5-strip-aggregated-graph-edges). It mirrors the real
// gap C (start → read_object behind a snapshot node's own store write).
//
// The agent's only ownership-marked seed is the entry node Run (Graph.start).
// The non-entry graph node `snapshot` carries NO impl marker — its @calls are
// hand-authored in the SPEC only. call_graph_integrity AGGREGATES the snapshot
// node's spec @calls onto the entry node and checks only the entry:
//
//	Run (entry node = Graph.start, the seed)
//	  ├─ blob.ListObjects()  →  list_objects  (SDK op marker, FIRST-ON-PATH → got
//	  │                          → snapshot @calls list_objects is KEPT)
//	  └─ writeMeta()  (op-LESS helper, MetaRow:writes — walk_first STOPS here)
//	        └─ blob.ReadObject()  →  read_object  (SDK op marker, reached only
//	                                  BEHIND MetaRow:writes → boundaryGot, NOT got
//	                                  → snapshot @calls read_object is STRIPPED)
//
// archive_object has NO impl at all — never AST-reached, absent from got and
// boundaryGot → snapshot @calls archive_object is KEPT (boundary case: the strip
// must not crash on or wrongly strip an unreachable target).

type Object struct{ Key string }

// Blobstore is the external SDK the snapshot agent calls.
type Blobstore interface {
	ListObjects(prefix string) ([]Object, error)
	ReadObject(key string) (*Object, error)
}

// MetaStore persists snapshot meta rows.
type MetaStore interface {
	WriteMeta(id, payload string) error
}

type Agent struct {
	blob Blobstore
	meta MetaStore
}

func New(blob Blobstore, meta MetaStore) *Agent {
	return &Agent{blob: blob, meta: meta}
}

// Run is the agent graph ENTRY node (is_entrypoint). It is the unit's
// ownership-marked seed; call_graph_integrity walks the whole unit AST from here.
// a10n:blueprint Components.SnapshotAgent.Graph.start
func (a *Agent) Run(key string) error {
	// list_objects is reached FIRST-ON-PATH (a direct SDK call before any store
	// marker), so it lands in the entry's `got`. The snapshot node's hand-authored
	// @calls(list_objects) aggregates onto the entry and is covered → KEPT.
	if _, err := a.listObjects(key); err != nil {
		return err
	}
	// writeMeta bears the MetaRow:writes store marker — walk_first STOPS here, so
	// anything writeMeta reaches deeper (read_object) is BEHIND it for the entry.
	return a.writeMeta(key)
}

// listObjects forwards to the SDK list op (first-on-path SDK marker).
func (a *Agent) listObjects(prefix string) ([]Object, error) {
	// a10n:blueprint Components.BlobstoreSdk.Operations.list_objects
	return a.blob.ListObjects(prefix)
}

// writeMeta is an op-LESS helper bearing the MetaRow:writes store marker. The
// check's walk_first STOPS at this system-layer marker, so read_object (reached
// deeper) is never first-on-path for the entry.
func (a *Agent) writeMeta(key string) error {
	// a10n:blueprint Components.MetaRelationalStore.MetaRow:writes
	if err := a.meta.WriteMeta(key, "snapshot"); err != nil {
		return err
	}
	return a.readObject(key)
}

// readObject forwards to the SDK read op. It is reached ONLY through writeMeta's
// store marker, so for the entry it is redundant-behind: in walk_boundary
// `boundaryGot` but not first-on-path `got`. The snapshot node's hand-authored
// @calls(read_object) aggregates onto the entry and the check flags it
// spec_redundant_marker_behind → ingest STRIPS it from the snapshot node.
func (a *Agent) readObject(key string) error {
	// a10n:blueprint Components.BlobstoreSdk.Operations.read_object
	_, err := a.blob.ReadObject(key)
	return err
}
