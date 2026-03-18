package store

import (
	"context"
	"encoding/json"
	"fmt"
)

func (s *Store) GetCloseoutStatus(ctx context.Context, jobID string) (CloseoutStatus, error) {
	job, err := s.GetJob(ctx, jobID)
	if err != nil {
		return CloseoutStatus{}, err
	}

	var packet DocumentPacket
	var mediaRaw string
	var formsRaw string
	var missingItemsRaw string
	var prepRequired int
	var prepComplete int

	err = s.db.QueryRowContext(
		ctx,
		`SELECT id, job_id, notes, media_json, forms_json, follow_up_instructions, completeness_status,
		        missing_items_json, prep_required, prep_complete, follow_up_plan, updated_at
		 FROM document_packets
		 WHERE job_id = ?`,
		jobID,
	).Scan(
		&packet.ID,
		&packet.JobID,
		&packet.Notes,
		&mediaRaw,
		&formsRaw,
		&packet.FollowUpInstructions,
		&packet.CompletenessStatus,
		&missingItemsRaw,
		&prepRequired,
		&prepComplete,
		&packet.FollowUpPlan,
		&packet.UpdatedAt,
	)
	if err != nil {
		return CloseoutStatus{}, fmt.Errorf("load document packet status: %w", err)
	}

	packet.MediaIDs = decodeSlice(mediaRaw)
	packet.FormIDs = decodeSlice(formsRaw)
	packet.MissingItems = decodeSlice(missingItemsRaw)
	packet.PrepRequired = prepRequired == 1
	packet.PrepComplete = prepComplete == 1

	return CloseoutStatus{
		Job:            job,
		DocumentPacket: packet,
	}, nil
}

func (s *Store) ListEvents(ctx context.Context, entityType string, entityID string, eventName string) ([]EventRecord, error) {
	query := `SELECT id, event_name, entity_type, entity_id, occurred_at, payload_json
	          FROM events
	          WHERE (? = '' OR entity_type = ?)
	            AND (? = '' OR entity_id = ?)
	            AND (? = '' OR event_name = ?)
	          ORDER BY occurred_at ASC, id ASC`

	rows, err := s.db.QueryContext(ctx, query, entityType, entityType, entityID, entityID, eventName, eventName)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	events := []EventRecord{}
	for rows.Next() {
		var event EventRecord
		var payloadRaw string
		if err := rows.Scan(
			&event.ID,
			&event.EventName,
			&event.EntityType,
			&event.EntityID,
			&event.OccurredAt,
			&payloadRaw,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		event.Payload = map[string]any{}
		if payloadRaw != "" {
			if err := json.Unmarshal([]byte(payloadRaw), &event.Payload); err != nil {
				return nil, fmt.Errorf("decode event payload: %w", err)
			}
		}
		events = append(events, event)
	}

	return events, nil
}
