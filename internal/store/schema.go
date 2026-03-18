package store

const schemaVersion = 1

const schemaSQL = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS route_days (
  id TEXT PRIMARY KEY,
  route_date TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS jobs (
  id TEXT PRIMARY KEY,
  customer_name TEXT NOT NULL,
  property_address TEXT NOT NULL,
  job_type TEXT NOT NULL,
  pest_target TEXT NOT NULL,
  priority TEXT NOT NULL,
  scheduled_window TEXT NOT NULL,
  status TEXT NOT NULL,
  closeout_state TEXT NOT NULL,
  billing_hold_reason TEXT NOT NULL,
  service_summary TEXT NOT NULL DEFAULT '',
  inspection_notes TEXT NOT NULL DEFAULT '',
  treatment_notes TEXT NOT NULL DEFAULT '',
  paperwork_required INTEGER NOT NULL DEFAULT 0,
  paperwork_ready INTEGER NOT NULL DEFAULT 0,
  requires_photos INTEGER NOT NULL DEFAULT 0,
  requires_prep INTEGER NOT NULL DEFAULT 0,
  requires_follow_up INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS route_stops (
  id TEXT PRIMARY KEY,
  route_day_id TEXT NOT NULL REFERENCES route_days(id) ON DELETE CASCADE,
  job_id TEXT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  position INTEGER NOT NULL,
  status TEXT NOT NULL,
  is_urgent INTEGER NOT NULL DEFAULT 0,
  blocker TEXT NOT NULL DEFAULT '',
  unresolved_reason TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(route_day_id, position),
  UNIQUE(route_day_id, job_id)
);

CREATE TABLE IF NOT EXISTS field_checkins (
  id TEXT PRIMARY KEY,
  source_channel TEXT NOT NULL,
  route_day_id TEXT REFERENCES route_days(id) ON DELETE SET NULL,
  linked_job_id TEXT REFERENCES jobs(id) ON DELETE SET NULL,
  input_type TEXT NOT NULL,
  short_status TEXT NOT NULL,
  blockers_json TEXT NOT NULL DEFAULT '[]',
  raw_text TEXT NOT NULL DEFAULT '',
  confidence REAL NOT NULL DEFAULT 1.0,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS voice_memos (
  id TEXT PRIMARY KEY,
  field_checkin_id TEXT NOT NULL REFERENCES field_checkins(id) ON DELETE CASCADE,
  linked_job_id TEXT REFERENCES jobs(id) ON DELETE SET NULL,
  audio_path TEXT NOT NULL,
  transcript TEXT NOT NULL DEFAULT '',
  summary TEXT NOT NULL DEFAULT '',
  pest_facts_json TEXT NOT NULL DEFAULT '[]',
  treatment_facts_json TEXT NOT NULL DEFAULT '[]',
  follow_up_needs_json TEXT NOT NULL DEFAULT '[]',
  billing_changes_json TEXT NOT NULL DEFAULT '[]',
  content_ideas_json TEXT NOT NULL DEFAULT '[]',
  confidence_json TEXT NOT NULL DEFAULT '{}',
  processing_status TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS document_packets (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL UNIQUE REFERENCES jobs(id) ON DELETE CASCADE,
  notes TEXT NOT NULL DEFAULT '',
  media_json TEXT NOT NULL DEFAULT '[]',
  forms_json TEXT NOT NULL DEFAULT '[]',
  follow_up_instructions TEXT NOT NULL DEFAULT '',
  completeness_status TEXT NOT NULL DEFAULT 'pending',
  missing_items_json TEXT NOT NULL DEFAULT '[]',
  prep_required INTEGER NOT NULL DEFAULT 0,
  prep_complete INTEGER NOT NULL DEFAULT 0,
  follow_up_plan TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS callbacks (
  id TEXT PRIMARY KEY,
  origin_job_id TEXT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  callback_job_id TEXT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  reason TEXT NOT NULL,
  urgency TEXT NOT NULL,
  margin_impact_note TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
  id TEXT PRIMARY KEY,
  owner TEXT NOT NULL,
  due_date TEXT NOT NULL DEFAULT '',
  linked_entity_type TEXT NOT NULL,
  linked_entity_id TEXT NOT NULL,
  reason TEXT NOT NULL,
  resolution_state TEXT NOT NULL,
  details TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS invoice_drafts (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL UNIQUE REFERENCES jobs(id) ON DELETE CASCADE,
  state TEXT NOT NULL,
  billable_items_json TEXT NOT NULL DEFAULT '[]',
  billing_hold_reason TEXT NOT NULL,
  total_amount REAL NOT NULL DEFAULT 0,
  delivery_state TEXT NOT NULL DEFAULT 'pending',
  payment_status TEXT NOT NULL DEFAULT 'unpaid',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS payments (
  id TEXT PRIMARY KEY,
  invoice_draft_id TEXT REFERENCES invoice_drafts(id) ON DELETE SET NULL,
  job_id TEXT REFERENCES jobs(id) ON DELETE SET NULL,
  amount REAL NOT NULL DEFAULT 0,
  method TEXT NOT NULL DEFAULT '',
  received_date TEXT NOT NULL DEFAULT '',
  reconciliation_state TEXT NOT NULL DEFAULT 'pending',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS ledger_entries (
  id TEXT PRIMARY KEY,
  job_id TEXT REFERENCES jobs(id) ON DELETE SET NULL,
  account_bucket TEXT NOT NULL,
  amount REAL NOT NULL DEFAULT 0,
  source_event TEXT NOT NULL,
  period TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS artifacts (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  linked_entity_type TEXT NOT NULL,
  linked_entity_id TEXT NOT NULL,
  relative_path TEXT NOT NULL,
  original_name TEXT NOT NULL,
  sha256 TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
  id TEXT PRIMARY KEY,
  event_name TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  occurred_at TEXT NOT NULL,
  payload_json TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_route_days_date ON route_days(route_date);
CREATE INDEX IF NOT EXISTS idx_route_stops_order ON route_stops(route_day_id, position);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status, closeout_state, billing_hold_reason);
CREATE INDEX IF NOT EXISTS idx_callbacks_status ON callbacks(status, urgency);
CREATE INDEX IF NOT EXISTS idx_events_lookup ON events(event_name, occurred_at);
`
