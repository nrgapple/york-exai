package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type rowScanner interface {
	Scan(dest ...any) error
}

func Open(ctx context.Context, dbPath string, autoInit bool) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create database dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetConnMaxIdleTime(5 * time.Minute)

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout = 5000;"); err != nil {
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	st := &Store{db: db}
	if autoInit {
		if err := st.ensureSchema(ctx); err != nil {
			return nil, err
		}
	}

	return st, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) ensureSchema(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	if _, err := s.db.ExecContext(
		ctx,
		"INSERT OR IGNORE INTO schema_migrations(version, applied_at) VALUES(?, ?)",
		schemaVersion,
		nowRFC3339(),
	); err != nil {
		return fmt.Errorf("record schema version: %w", err)
	}

	return nil
}

func (s *Store) SchemaVersion(ctx context.Context) (int, error) {
	var version int
	err := s.db.QueryRowContext(ctx, "SELECT MAX(version) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("query schema version: %w", err)
	}
	return version, nil
}

func (s *Store) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func sliceJSON(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	data, _ := json.Marshal(values)
	return string(data)
}

func mapJSON(values map[string]any) string {
	if len(values) == 0 {
		return "{}"
	}
	data, _ := json.Marshal(values)
	return string(data)
}

func decodeSlice(raw string) []string {
	if raw == "" {
		return []string{}
	}

	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return []string{}
	}
	return values
}

func checksumFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read artifact: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func normalizePestValue(jobType string, pestTarget string) string {
	return strings.ToLower(strings.TrimSpace(jobType + " " + pestTarget))
}

func isTermiteOrWDI(jobType string, pestTarget string) bool {
	value := normalizePestValue(jobType, pestTarget)
	return strings.Contains(value, "termite") || strings.Contains(value, "wdi")
}

func isBedBug(jobType string, pestTarget string) bool {
	value := normalizePestValue(jobType, pestTarget)
	return strings.Contains(value, "bed bug")
}

func writeEvent(ctx context.Context, tx *sql.Tx, eventName string, entityType string, entityID string, payload map[string]any) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO events(id, event_name, entity_type, entity_id, occurred_at, payload_json)
		 VALUES(?, ?, ?, ?, ?, ?)`,
		newID("evt"),
		eventName,
		entityType,
		entityID,
		nowRFC3339(),
		mapJSON(payload),
	)
	if err != nil {
		return fmt.Errorf("write event %s: %w", eventName, err)
	}
	return nil
}

func createTask(ctx context.Context, tx *sql.Tx, owner string, reason string, linkedEntityType string, linkedEntityID string, details string) (Task, error) {
	task := Task{
		ID:               newID("task"),
		Owner:            owner,
		Reason:           reason,
		ResolutionState:  "open",
		LinkedEntityType: linkedEntityType,
		LinkedEntityID:   linkedEntityID,
		Details:          details,
	}

	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO tasks(id, owner, due_date, linked_entity_type, linked_entity_id, reason, resolution_state, details, created_at, updated_at)
		 VALUES(?, ?, '', ?, ?, ?, ?, ?, ?, ?)`,
		task.ID,
		task.Owner,
		task.LinkedEntityType,
		task.LinkedEntityID,
		task.Reason,
		task.ResolutionState,
		task.Details,
		nowRFC3339(),
		nowRFC3339(),
	)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	return task, nil
}

func scanJob(scanner rowScanner) (Job, error) {
	var job Job
	var paperworkRequired int
	var paperworkReady int
	var requiresPhotos int
	var requiresPrep int
	var requiresFollowUp int

	err := scanner.Scan(
		&job.ID,
		&job.CustomerName,
		&job.PropertyAddress,
		&job.JobType,
		&job.PestTarget,
		&job.Priority,
		&job.ScheduledWindow,
		&job.Status,
		&job.CloseoutState,
		&job.BillingHoldReason,
		&job.ServiceSummary,
		&job.InspectionNotes,
		&job.TreatmentNotes,
		&paperworkRequired,
		&paperworkReady,
		&requiresPhotos,
		&requiresPrep,
		&requiresFollowUp,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return Job{}, err
	}

	job.PaperworkRequired = paperworkRequired == 1
	job.PaperworkReady = paperworkReady == 1
	job.RequiresPhotos = requiresPhotos == 1
	job.RequiresPrep = requiresPrep == 1
	job.RequiresFollowUp = requiresFollowUp == 1
	return job, nil
}
