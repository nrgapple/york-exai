package store

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *Store) CreateFieldCheckIn(ctx context.Context, input FieldCheckInInput) (string, []Task, error) {
	checkInID := newID("checkin")
	tasks := []Task{}

	err := s.withTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO field_checkins(id, source_channel, route_day_id, linked_job_id, input_type, short_status, blockers_json, raw_text, confidence, created_at)
			 VALUES(?, ?, NULLIF(?, ''), NULLIF(?, ''), 'text', ?, ?, ?, ?, ?)`,
			checkInID,
			input.SourceChannel,
			input.RouteDayID,
			input.JobID,
			input.ShortStatus,
			sliceJSON(input.Blockers),
			input.RawText,
			input.Confidence,
			nowRFC3339(),
		)
		if err != nil {
			return fmt.Errorf("insert field check-in: %w", err)
		}

		if input.JobID == "" || input.Confidence < 0.75 {
			task, err := createTask(ctx, tx, "ops", "job_link_unclear", "field_checkin", checkInID, "Field check-in requires job review.")
			if err != nil {
				return err
			}
			tasks = append(tasks, task)
		}

		if err := writeEvent(ctx, tx, "field_checkin.received", "field_checkin", checkInID, map[string]any{
			"source_channel": input.SourceChannel,
			"linked_job_id":  input.JobID,
			"confidence":     input.Confidence,
		}); err != nil {
			return err
		}

		if len(tasks) > 0 {
			if err := writeEvent(ctx, tx, "feedback.detected", "task", tasks[0].ID, map[string]any{
				"reason":            tasks[0].Reason,
				"linked_checkin_id": checkInID,
			}); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return "", nil, err
	}

	return checkInID, tasks, nil
}

func (s *Store) RecordArtifact(ctx context.Context, kind string, linkedEntityType string, linkedEntityID string, relativePath string, originalName string, sha256 string) (Artifact, error) {
	artifact := Artifact{
		ID:               newID("artifact"),
		Kind:             kind,
		LinkedEntityType: linkedEntityType,
		LinkedEntityID:   linkedEntityID,
		RelativePath:     relativePath,
		OriginalName:     originalName,
		SHA256:           sha256,
	}

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO artifacts(id, kind, linked_entity_type, linked_entity_id, relative_path, original_name, sha256, created_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		artifact.ID,
		artifact.Kind,
		artifact.LinkedEntityType,
		artifact.LinkedEntityID,
		artifact.RelativePath,
		artifact.OriginalName,
		artifact.SHA256,
		nowRFC3339(),
	)
	if err != nil {
		return Artifact{}, fmt.Errorf("insert artifact: %w", err)
	}

	return artifact, nil
}

func (s *Store) RecordFieldPhoto(ctx context.Context, jobID string, sourceChannel string, artifact Artifact) (string, error) {
	checkInID := newID("checkin")
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO field_checkins(id, source_channel, route_day_id, linked_job_id, input_type, short_status, blockers_json, raw_text, confidence, created_at)
			 VALUES(?, ?, NULL, NULLIF(?, ''), 'photo', 'photo attached', '[]', ?, 1.0, ?)`,
			checkInID,
			sourceChannel,
			jobID,
			artifact.RelativePath,
			nowRFC3339(),
		)
		if err != nil {
			return fmt.Errorf("insert photo field check-in: %w", err)
		}

		if jobID == "" {
			if _, err := createTask(ctx, tx, "ops", "job_link_unclear", "field_checkin", checkInID, "Photo requires job review."); err != nil {
				return err
			}
		}

		return writeEvent(ctx, tx, "field_checkin.received", "field_checkin", checkInID, map[string]any{
			"input_type":  "photo",
			"artifact_id": artifact.ID,
		})
	})
	if err != nil {
		return "", err
	}

	return checkInID, nil
}

func (s *Store) RecordVoiceMemo(ctx context.Context, input VoiceMemoInput, artifact Artifact) (string, []Task, error) {
	checkInID := newID("checkin")
	voiceMemoID := newID("voice")
	tasks := []Task{}

	err := s.withTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO field_checkins(id, source_channel, route_day_id, linked_job_id, input_type, short_status, blockers_json, raw_text, confidence, created_at)
			 VALUES(?, ?, NULL, NULLIF(?, ''), 'voice', 'voice memo received', '[]', ?, ?, ?)`,
			checkInID,
			input.SourceChannel,
			input.JobID,
			input.Summary,
			input.Confidence,
			nowRFC3339(),
		)
		if err != nil {
			return fmt.Errorf("insert voice memo check-in: %w", err)
		}

		processingStatus := "ready"
		if input.Transcript == "" {
			processingStatus = "review_required"
			task, err := createTask(ctx, tx, "ops", "transcription_unavailable", "voice_memo", voiceMemoID, "Voice memo retained but transcription is unavailable.")
			if err != nil {
				return err
			}
			tasks = append(tasks, task)
		}
		if input.JobID == "" || input.Confidence < 0.75 {
			task, err := createTask(ctx, tx, "ops", "job_link_unclear", "voice_memo", voiceMemoID, "Voice memo requires job review.")
			if err != nil {
				return err
			}
			tasks = append(tasks, task)
			processingStatus = "review_required"
		}

		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO voice_memos(id, field_checkin_id, linked_job_id, audio_path, transcript, summary,
			                         pest_facts_json, treatment_facts_json, follow_up_needs_json, billing_changes_json,
			                         content_ideas_json, confidence_json, processing_status, created_at)
			 VALUES(?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			voiceMemoID,
			checkInID,
			input.JobID,
			artifact.RelativePath,
			input.Transcript,
			input.Summary,
			sliceJSON(input.PestFacts),
			sliceJSON(input.TreatmentFacts),
			sliceJSON(input.FollowUpNeeds),
			sliceJSON(input.BillingChanges),
			sliceJSON(input.ContentIdeas),
			mapJSON(map[string]any{"job_match": input.Confidence}),
			processingStatus,
			nowRFC3339(),
		)
		if err != nil {
			return fmt.Errorf("insert voice memo: %w", err)
		}

		if err := writeEvent(ctx, tx, "voice_memo.received", "voice_memo", voiceMemoID, map[string]any{
			"linked_job_id": input.JobID,
			"audio_path":    artifact.RelativePath,
		}); err != nil {
			return err
		}

		if input.Transcript != "" {
			if err := writeEvent(ctx, tx, "voice_memo.transcribed", "voice_memo", voiceMemoID, map[string]any{
				"summary": input.Summary,
			}); err != nil {
				return err
			}
		}

		if len(input.FollowUpNeeds) > 0 {
			if err := writeEvent(ctx, tx, "followup.scheduled", "voice_memo", voiceMemoID, map[string]any{
				"follow_up_needs": input.FollowUpNeeds,
			}); err != nil {
				return err
			}
		}

		for _, task := range tasks {
			if err := writeEvent(ctx, tx, "feedback.detected", "task", task.ID, map[string]any{
				"reason":        task.Reason,
				"voice_memo_id": voiceMemoID,
			}); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return "", nil, err
	}

	return voiceMemoID, tasks, nil
}

func (s *Store) ListOpenReviewTasks(ctx context.Context) ([]Task, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, owner, reason, resolution_state, linked_entity_type, linked_entity_id, details, due_date
		 FROM tasks
		 WHERE resolution_state = 'open'
		 ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query review tasks: %w", err)
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		var task Task
		if err := rows.Scan(
			&task.ID,
			&task.Owner,
			&task.Reason,
			&task.ResolutionState,
			&task.LinkedEntityType,
			&task.LinkedEntityID,
			&task.Details,
			&task.DueDate,
		); err != nil {
			return nil, fmt.Errorf("scan review task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
