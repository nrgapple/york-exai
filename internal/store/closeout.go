package store

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *Store) UpdateCloseoutNotes(ctx context.Context, input CloseoutNoteInput) (Job, error) {
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		job, err := scanJob(tx.QueryRowContext(
			ctx,
			`SELECT id, customer_name, property_address, job_type, pest_target, priority, scheduled_window,
			        status, closeout_state, billing_hold_reason, service_summary, inspection_notes,
			        treatment_notes, paperwork_required, paperwork_ready, requires_photos,
			        requires_prep, requires_follow_up, created_at, updated_at
			 FROM jobs
			 WHERE id = ?`,
			input.JobID,
		))
		if err != nil {
			return fmt.Errorf("load job for closeout note: %w", err)
		}

		if input.ServiceSummary != "" {
			job.ServiceSummary = input.ServiceSummary
		}
		if input.InspectionNotes != "" {
			job.InspectionNotes = input.InspectionNotes
		}
		if input.TreatmentNotes != "" {
			job.TreatmentNotes = input.TreatmentNotes
		}
		if input.PaperworkRequired != nil {
			job.PaperworkRequired = *input.PaperworkRequired
		}
		if input.PaperworkReady != nil {
			job.PaperworkReady = *input.PaperworkReady
		}
		if input.RequiresPhotos != nil {
			job.RequiresPhotos = *input.RequiresPhotos
		}
		if input.RequiresPrep != nil {
			job.RequiresPrep = *input.RequiresPrep
		}
		if input.RequiresFollowUp != nil {
			job.RequiresFollowUp = *input.RequiresFollowUp
		}

		_, err = tx.ExecContext(
			ctx,
			`UPDATE jobs
			 SET service_summary = ?, inspection_notes = ?, treatment_notes = ?,
			     paperwork_required = ?, paperwork_ready = ?, requires_photos = ?,
			     requires_prep = ?, requires_follow_up = ?, updated_at = ?
			 WHERE id = ?`,
			job.ServiceSummary,
			job.InspectionNotes,
			job.TreatmentNotes,
			boolToInt(job.PaperworkRequired),
			boolToInt(job.PaperworkReady),
			boolToInt(job.RequiresPhotos),
			boolToInt(job.RequiresPrep),
			boolToInt(job.RequiresFollowUp),
			nowRFC3339(),
			job.ID,
		)
		if err != nil {
			return fmt.Errorf("update closeout notes: %w", err)
		}

		_, err = tx.ExecContext(
			ctx,
			`UPDATE document_packets
			 SET notes = ?, follow_up_instructions = COALESCE(NULLIF(?, ''), follow_up_instructions), updated_at = ?
			 WHERE job_id = ?`,
			job.ServiceSummary,
			input.FollowUpPlan,
			nowRFC3339(),
			job.ID,
		)
		if err != nil {
			return fmt.Errorf("update document packet notes: %w", err)
		}

		return writeEvent(ctx, tx, "inspection.completed", "job", job.ID, map[string]any{
			"service_summary_present":  job.ServiceSummary != "",
			"inspection_notes_present": job.InspectionNotes != "",
			"treatment_notes_present":  job.TreatmentNotes != "",
		})
	})
	if err != nil {
		return Job{}, err
	}

	return s.GetJob(ctx, input.JobID)
}

func (s *Store) SetPrepState(ctx context.Context, input PrepInput) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`UPDATE jobs
			 SET requires_prep = ?, requires_follow_up = ?, updated_at = ?
			 WHERE id = ?`,
			boolToInt(input.Required),
			boolToInt(input.FollowUpPlan != ""),
			nowRFC3339(),
			input.JobID,
		)
		if err != nil {
			return fmt.Errorf("update prep state on job: %w", err)
		}

		_, err = tx.ExecContext(
			ctx,
			`UPDATE document_packets
			 SET prep_required = ?, prep_complete = ?, follow_up_plan = ?, updated_at = ?
			 WHERE job_id = ?`,
			boolToInt(input.Required),
			boolToInt(input.Completed),
			input.FollowUpPlan,
			nowRFC3339(),
			input.JobID,
		)
		if err != nil {
			return fmt.Errorf("update prep state on packet: %w", err)
		}

		eventName := "prep_notice.required"
		if input.Completed {
			eventName = "prep_notice.cleared"
		}
		return writeEvent(ctx, tx, eventName, "job", input.JobID, map[string]any{
			"required":       input.Required,
			"completed":      input.Completed,
			"follow_up_plan": input.FollowUpPlan,
		})
	})
}

func (s *Store) AttachDocumentPhoto(ctx context.Context, jobID string, artifact Artifact) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		var mediaRaw string
		if err := tx.QueryRowContext(ctx, "SELECT media_json FROM document_packets WHERE job_id = ?", jobID).Scan(&mediaRaw); err != nil {
			return fmt.Errorf("load document packet media: %w", err)
		}
		media := decodeSlice(mediaRaw)
		media = append(media, artifact.ID)

		_, err := tx.ExecContext(
			ctx,
			`UPDATE document_packets
			 SET media_json = ?, updated_at = ?
			 WHERE job_id = ?`,
			sliceJSON(media),
			nowRFC3339(),
			jobID,
		)
		if err != nil {
			return fmt.Errorf("update document packet media: %w", err)
		}

		return writeEvent(ctx, tx, "treatment.completed", "job", jobID, map[string]any{
			"artifact_id":   artifact.ID,
			"artifact_kind": artifact.Kind,
		})
	})
}

func (s *Store) EvaluateCloseout(ctx context.Context, jobID string) (CloseoutResult, error) {
	job, err := s.GetJob(ctx, jobID)
	if err != nil {
		return CloseoutResult{}, err
	}

	var mediaRaw string
	var prepRequired int
	var prepComplete int
	var followUpPlan string
	err = s.db.QueryRowContext(
		ctx,
		`SELECT media_json, prep_required, prep_complete, follow_up_plan
		 FROM document_packets
		 WHERE job_id = ?`,
		jobID,
	).Scan(&mediaRaw, &prepRequired, &prepComplete, &followUpPlan)
	if err != nil {
		return CloseoutResult{}, fmt.Errorf("load document packet: %w", err)
	}

	missing := []string{}
	followUps := []string{}

	if job.ServiceSummary == "" {
		missing = append(missing, "service_summary")
	}
	if isTermiteOrWDI(job.JobType, job.PestTarget) {
		if job.InspectionNotes == "" {
			missing = append(missing, "inspection_notes")
		}
		if !job.PaperworkReady {
			missing = append(missing, "paperwork_ready")
		}
	}
	if isBedBug(job.JobType, job.PestTarget) {
		if job.TreatmentNotes == "" {
			missing = append(missing, "treatment_notes")
		}
		if prepRequired == 1 && prepComplete == 0 {
			missing = append(missing, "prep_complete")
		}
		if followUpPlan == "" {
			followUps = append(followUps, "reinspection_plan")
		}
	}
	if job.RequiresPhotos && len(decodeSlice(mediaRaw)) == 0 {
		missing = append(missing, "required_photo")
	}
	if job.RequiresFollowUp && followUpPlan == "" {
		followUps = append(followUps, "follow_up_plan")
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
		return CloseoutResult{}, fmt.Errorf("query callback state: %w", err)
	}
	if callbackCount > 0 {
		followUps = append(followUps, "callback_under_review")
	}

	result := CloseoutResult{
		JobID:            jobID,
		CloseoutState:    "complete",
		MissingItems:     missing,
		FollowUpRequired: followUps,
		CallbackOpen:     callbackCount > 0,
	}
	if len(missing) > 0 {
		result.CloseoutState = "blocked"
	} else if len(followUps) > 0 {
		result.CloseoutState = "follow_up_needed"
	}

	err = s.withTx(ctx, func(tx *sql.Tx) error {
		billingHold := "none"
		if result.CloseoutState == "blocked" {
			billingHold = "missing_records"
		}

		_, err := tx.ExecContext(
			ctx,
			`UPDATE jobs
			 SET closeout_state = ?, billing_hold_reason = ?, updated_at = ?
			 WHERE id = ?`,
			result.CloseoutState,
			billingHold,
			nowRFC3339(),
			jobID,
		)
		if err != nil {
			return fmt.Errorf("update job closeout state: %w", err)
		}

		_, err = tx.ExecContext(
			ctx,
			`UPDATE document_packets
			 SET completeness_status = ?, missing_items_json = ?, updated_at = ?
			 WHERE job_id = ?`,
			result.CloseoutState,
			sliceJSON(result.MissingItems),
			nowRFC3339(),
			jobID,
		)
		if err != nil {
			return fmt.Errorf("update document packet completeness: %w", err)
		}

		if result.CloseoutState == "blocked" {
			if _, err := createTask(ctx, tx, "ops", "closeout_missing_records", "job", jobID, "Closeout is blocked on missing records."); err != nil {
				return err
			}
			return writeEvent(ctx, tx, "job.closeout.blocked", "job", jobID, map[string]any{
				"missing_items": result.MissingItems,
			})
		}
		if result.CloseoutState == "follow_up_needed" {
			return writeEvent(ctx, tx, "followup.scheduled", "job", jobID, map[string]any{
				"follow_up_required": result.FollowUpRequired,
			})
		}
		return writeEvent(ctx, tx, "job.closed", "job", jobID, map[string]any{
			"closeout_state": result.CloseoutState,
		})
	})
	if err != nil {
		return CloseoutResult{}, err
	}

	return result, nil
}

func (s *Store) GetJob(ctx context.Context, jobID string) (Job, error) {
	job, err := scanJob(s.db.QueryRowContext(
		ctx,
		`SELECT id, customer_name, property_address, job_type, pest_target, priority, scheduled_window,
		        status, closeout_state, billing_hold_reason, service_summary, inspection_notes,
		        treatment_notes, paperwork_required, paperwork_ready, requires_photos,
		        requires_prep, requires_follow_up, created_at, updated_at
		 FROM jobs
		 WHERE id = ?`,
		jobID,
	))
	if err != nil {
		return Job{}, fmt.Errorf("load job: %w", err)
	}
	return job, nil
}

func (s *Store) ListBlockedJobs(ctx context.Context) ([]Job, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, customer_name, property_address, job_type, pest_target, priority, scheduled_window,
		        status, closeout_state, billing_hold_reason, service_summary, inspection_notes,
		        treatment_notes, paperwork_required, paperwork_ready, requires_photos,
		        requires_prep, requires_follow_up, created_at, updated_at
		 FROM jobs
		 WHERE closeout_state = 'blocked'
		 ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query blocked jobs: %w", err)
	}
	defer rows.Close()

	jobs := []Job{}
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, fmt.Errorf("scan blocked job: %w", err)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
