package store

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *Store) DraftInvoice(ctx context.Context, jobID string) (InvoiceDraftResult, error) {
	job, err := s.GetJob(ctx, jobID)
	if err != nil {
		return InvoiceDraftResult{}, err
	}

	result := InvoiceDraftResult{
		InvoiceDraftID:    newID("invoice"),
		JobID:             jobID,
		State:             "draft",
		BillingHoldReason: "none",
		Reasons:           []string{},
	}

	if job.CloseoutState != "complete" {
		result.State = "hold"
		result.BillingHoldReason = "closeout_incomplete"
		result.Reasons = append(result.Reasons, job.CloseoutState)
	}

	var callbackCount int
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT COUNT(1)
		 FROM callbacks
		 WHERE (origin_job_id = ? OR callback_job_id = ?) AND status = 'open'`,
		jobID,
		jobID,
	).Scan(&callbackCount); err != nil {
		return InvoiceDraftResult{}, fmt.Errorf("query callback hold state: %w", err)
	}
	if callbackCount > 0 {
		result.State = "hold"
		result.BillingHoldReason = "callback_under_review"
		result.Reasons = append(result.Reasons, "open_callback")
	}

	if result.State == "hold" && result.BillingHoldReason == "none" {
		result.BillingHoldReason = "missing_records"
	}

	err = s.withTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO invoice_drafts(id, job_id, state, billable_items_json, billing_hold_reason, total_amount, delivery_state, payment_status, created_at, updated_at)
			 VALUES(?, ?, ?, ?, ?, 0, 'pending', 'unpaid', ?, ?)
			 ON CONFLICT(job_id) DO UPDATE SET
			   state = excluded.state,
			   billable_items_json = excluded.billable_items_json,
			   billing_hold_reason = excluded.billing_hold_reason,
			   updated_at = excluded.updated_at`,
			result.InvoiceDraftID,
			jobID,
			result.State,
			sliceJSON([]string{job.JobType, job.PestTarget}),
			result.BillingHoldReason,
			nowRFC3339(),
			nowRFC3339(),
		)
		if err != nil {
			return fmt.Errorf("upsert invoice draft: %w", err)
		}

		_, err = tx.ExecContext(
			ctx,
			`UPDATE jobs SET billing_hold_reason = ?, updated_at = ? WHERE id = ?`,
			result.BillingHoldReason,
			nowRFC3339(),
			jobID,
		)
		if err != nil {
			return fmt.Errorf("update job billing hold: %w", err)
		}

		if result.State == "hold" {
			return writeEvent(ctx, tx, "feedback.detected", "job", jobID, map[string]any{
				"billing_hold_reason": result.BillingHoldReason,
			})
		}
		return writeEvent(ctx, tx, "invoice.drafted", "invoice_draft", result.InvoiceDraftID, map[string]any{
			"job_id": jobID,
		})
	})
	if err != nil {
		return InvoiceDraftResult{}, err
	}

	return result, nil
}

func (s *Store) ListBillingHolds(ctx context.Context) ([]InvoiceDraftResult, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, job_id, state, billing_hold_reason, billable_items_json
		 FROM invoice_drafts
		 WHERE state = 'hold'
		 ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query billing holds: %w", err)
	}
	defer rows.Close()

	results := []InvoiceDraftResult{}
	for rows.Next() {
		var result InvoiceDraftResult
		var reasonsRaw string
		if err := rows.Scan(
			&result.InvoiceDraftID,
			&result.JobID,
			&result.State,
			&result.BillingHoldReason,
			&reasonsRaw,
		); err != nil {
			return nil, fmt.Errorf("scan billing hold: %w", err)
		}
		result.Reasons = decodeSlice(reasonsRaw)
		results = append(results, result)
	}
	return results, nil
}
