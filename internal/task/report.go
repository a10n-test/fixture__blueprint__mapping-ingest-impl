package task

// This file models the REDUNDANT-BEHIND reads/writes/calls topology for the
// ingest strip (work5-ingest-strip-redundant-rwc). The shape mirrors the real
// gaps (e.g. Doctor.doctor → CheckManifest behind an intervening store marker):
//
//	report_run (op)
//	  ├─ directly reads ReportRow            (depth 0 — first-on-path, KEPT)
//	  └─ calls loadConfig()  (op-LESS helper)
//	        ├─ reads ConfigRow               (depth 1 — first-on-path, KEPT)
//	        └─ calls loadManifest()  (op-LESS helper)
//	              └─ reads ManifestRow        (depth 2 — BEHIND ConfigRow's
//	                                           store marker → NOT first-on-path
//	                                           → spec_redundant_marker_behind →
//	                                           STRIPPED by ingest)
//
// The check's walk_first STOPS at ConfigRow's store marker on loadConfig
// (stopping_marked = ANY Components.% marker, incl. :reads), so ManifestRow is
// never first-on-path for report_run — the check flags report_run @reads
// ManifestRow as redundant-behind. The ingest walk's narrower stop-set
// (boundary_op_symbols excludes :reads/:writes) passes through loadConfig and
// over-attributes ManifestRow to report_run unless the rwc strip removes it.
// loadManifest is op-LESS, so this is NOT realizing-op attribution (no op
// boundary) — it is purely the intervening-store-marker shadowing.

type ReportRow struct{ ID string }
type ConfigRow struct{ Key string }
type ManifestRow struct{ Name string }

type ReportStore interface {
	GetReport(id string) (*ReportRow, error)
	GetConfig(key string) (*ConfigRow, error)
	GetManifest(name string) (*ManifestRow, error)
}

type ReportService struct{ store ReportStore }

// report_run reads ReportRow IN ITS OWN BODY (first-on-path, kept) and reaches
// ConfigRow / ManifestRow through the op-less loadConfig → loadManifest chain.
// a10n:blueprint Components.ReportService.Commands.report_run
func (s *ReportService) RunReport(id string) error {
	// a10n:blueprint Components.ReportRelationalStore.ReportRow:reads
	if _, err := s.store.GetReport(id); err != nil {
		return err
	}
	return s.loadConfig()
}

// loadConfig is an op-LESS helper bearing a ConfigRow:reads marker. The check's
// walk_first STOPS here (store marker is a system-layer Components.% marker), so
// anything this helper reaches deeper is BEHIND it for report_run.
func (s *ReportService) loadConfig() error {
	// a10n:blueprint Components.ReportRelationalStore.ConfigRow:reads
	if _, err := s.store.GetConfig("default"); err != nil {
		return err
	}
	return s.loadManifest()
}

// loadManifest is a deeper op-LESS helper bearing a ManifestRow:reads marker —
// reached ONLY through loadConfig's store marker, so it is redundant-behind for
// report_run and must be stripped from report_run's @reads at ingest.
func (s *ReportService) loadManifest() error {
	// a10n:blueprint Components.ReportRelationalStore.ManifestRow:reads
	_, err := s.store.GetManifest("app")
	return err
}
