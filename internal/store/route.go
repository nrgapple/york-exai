package store

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *Store) CreateRouteDay(ctx context.Context, input CreateRouteDayInput) (RouteDay, error) {
	var existing RouteDay
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, route_date, status, created_at, updated_at
		 FROM route_days
		 WHERE route_date = ?`,
		input.RouteDate,
	).Scan(&existing.ID, &existing.RouteDate, &existing.Status, &existing.CreatedAt, &existing.UpdatedAt)
	if err == nil {
		return existing, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return RouteDay{}, fmt.Errorf("query route day: %w", err)
	}

	route := RouteDay{
		ID:        newID("route"),
		RouteDate: input.RouteDate,
		Status:    "scheduled",
	}

	err = s.withTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO route_days(id, route_date, status, created_at, updated_at)
			 VALUES(?, ?, ?, ?, ?)`,
			route.ID,
			route.RouteDate,
			route.Status,
			nowRFC3339(),
			nowRFC3339(),
		)
		if err != nil {
			return fmt.Errorf("insert route day: %w", err)
		}

		return writeEvent(ctx, tx, "followup.scheduled", "route_day", route.ID, map[string]any{
			"route_date": route.RouteDate,
			"source":     "route.create",
		})
	})
	if err != nil {
		return RouteDay{}, err
	}

	return route, nil
}

func (s *Store) AddStop(ctx context.Context, input AddStopInput, urgent bool, requestedPosition int) (RouteStop, Job, error) {
	var routeDayID string
	if input.RouteDayID != "" {
		routeDayID = input.RouteDayID
	} else {
		route, err := s.CreateRouteDay(ctx, CreateRouteDayInput{RouteDate: input.RouteDate})
		if err != nil {
			return RouteStop{}, Job{}, err
		}
		routeDayID = route.ID
	}

	stop := RouteStop{}
	job := Job{}
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		if input.JobID != "" {
			existingJob, err := scanJob(tx.QueryRowContext(
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
				return fmt.Errorf("load existing job: %w", err)
			}
			job = existingJob
		} else {
			job = Job{
				ID:                newID("job"),
				CustomerName:      input.CustomerName,
				PropertyAddress:   input.PropertyAddress,
				JobType:           input.JobType,
				PestTarget:        input.PestTarget,
				Priority:          input.Priority,
				ScheduledWindow:   input.ScheduledWindow,
				Status:            "scheduled",
				CloseoutState:     "pending",
				BillingHoldReason: "none",
				PaperworkRequired: input.PaperworkRequired,
				RequiresPhotos:    input.RequiresPhotos,
				RequiresPrep:      input.RequiresPrep,
				RequiresFollowUp:  input.RequiresFollowUp,
			}

			_, err := tx.ExecContext(
				ctx,
				`INSERT INTO jobs(id, customer_name, property_address, job_type, pest_target, priority, scheduled_window,
				                  status, closeout_state, billing_hold_reason, paperwork_required, paperwork_ready,
				                  requires_photos, requires_prep, requires_follow_up, created_at, updated_at)
				 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, ?, ?, ?)`,
				job.ID,
				job.CustomerName,
				job.PropertyAddress,
				job.JobType,
				job.PestTarget,
				job.Priority,
				job.ScheduledWindow,
				job.Status,
				job.CloseoutState,
				job.BillingHoldReason,
				boolToInt(job.PaperworkRequired),
				boolToInt(job.RequiresPhotos),
				boolToInt(job.RequiresPrep),
				boolToInt(job.RequiresFollowUp),
				nowRFC3339(),
				nowRFC3339(),
			)
			if err != nil {
				return fmt.Errorf("insert job: %w", err)
			}

			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO document_packets(id, job_id, updated_at) VALUES(?, ?, ?)`,
				newID("packet"),
				job.ID,
				nowRFC3339(),
			)
			if err != nil {
				return fmt.Errorf("create document packet: %w", err)
			}
		}

		position := requestedPosition
		if position <= 0 {
			if err := tx.QueryRowContext(
				ctx,
				`SELECT COALESCE(MAX(position), 0) + 1 FROM route_stops WHERE route_day_id = ?`,
				routeDayID,
			).Scan(&position); err != nil {
				return fmt.Errorf("get next route position: %w", err)
			}
		} else if urgent {
			if _, err := tx.ExecContext(
				ctx,
				`UPDATE route_stops
				 SET position = position + 1000, updated_at = ?
				 WHERE route_day_id = ? AND position >= ?`,
				nowRFC3339(),
				routeDayID,
				position,
			); err != nil {
				return fmt.Errorf("shift urgent positions upward: %w", err)
			}
			if _, err := tx.ExecContext(
				ctx,
				`UPDATE route_stops
				 SET position = position - 999, updated_at = ?
				 WHERE route_day_id = ? AND position >= ?`,
				nowRFC3339(),
				routeDayID,
				position+1000,
			); err != nil {
				return fmt.Errorf("normalize urgent positions: %w", err)
			}
		}

		stop = RouteStop{
			ID:         newID("stop"),
			RouteDayID: routeDayID,
			JobID:      job.ID,
			Position:   position,
			Status:     "scheduled",
			IsUrgent:   urgent,
			Blocker:    "",
			Unresolved: "",
		}

		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO route_stops(id, route_day_id, job_id, position, status, is_urgent, blocker, unresolved_reason, created_at, updated_at)
			 VALUES(?, ?, ?, ?, ?, ?, '', '', ?, ?)`,
			stop.ID,
			stop.RouteDayID,
			stop.JobID,
			stop.Position,
			stop.Status,
			boolToInt(stop.IsUrgent),
			nowRFC3339(),
			nowRFC3339(),
		)
		if err != nil {
			return fmt.Errorf("insert route stop: %w", err)
		}

		eventName := "followup.scheduled"
		eventPayload := map[string]any{
			"route_day_id": routeDayID,
			"job_id":       job.ID,
			"position":     stop.Position,
			"urgent":       urgent,
		}

		if input.CallbackOfJobID != "" {
			callbackID := newID("callback")
			_, err := tx.ExecContext(
				ctx,
				`INSERT INTO callbacks(id, origin_job_id, callback_job_id, reason, urgency, margin_impact_note, status, created_at)
				 VALUES(?, ?, ?, ?, ?, '', 'open', ?)`,
				callbackID,
				input.CallbackOfJobID,
				job.ID,
				"callback return",
				job.Priority,
				nowRFC3339(),
			)
			if err != nil {
				return fmt.Errorf("insert callback: %w", err)
			}

			if err := writeEvent(ctx, tx, "callback.requested", "job", input.CallbackOfJobID, map[string]any{
				"callback_id":     callbackID,
				"callback_job_id": job.ID,
			}); err != nil {
				return err
			}
			if err := writeEvent(ctx, tx, "callback.scheduled", "job", job.ID, map[string]any{
				"callback_id":   callbackID,
				"origin_job_id": input.CallbackOfJobID,
			}); err != nil {
				return err
			}

			eventName = "callback.scheduled"
			eventPayload["callback_of_job_id"] = input.CallbackOfJobID
		}

		return writeEvent(ctx, tx, eventName, "route_stop", stop.ID, eventPayload)
	})
	if err != nil {
		return RouteStop{}, Job{}, err
	}

	stop.CustomerName = job.CustomerName
	stop.JobType = job.JobType
	stop.PestTarget = job.PestTarget
	stop.Priority = job.Priority
	return stop, job, nil
}

func (s *Store) UpdateStop(ctx context.Context, input UpdateStopInput) (RouteStop, error) {
	stop := RouteStop{}
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(
			ctx,
			`SELECT id, route_day_id, job_id, position, status, is_urgent, blocker, unresolved_reason
			 FROM route_stops
			 WHERE id = ?`,
			input.RouteStopID,
		).Scan(
			&stop.ID,
			&stop.RouteDayID,
			&stop.JobID,
			&stop.Position,
			&stop.Status,
			&stop.IsUrgent,
			&stop.Blocker,
			&stop.Unresolved,
		); err != nil {
			return fmt.Errorf("load route stop: %w", err)
		}

		stop.Status = input.Status
		stop.Blocker = input.Blocker
		stop.Unresolved = input.UnresolvedReason

		_, err := tx.ExecContext(
			ctx,
			`UPDATE route_stops
			 SET status = ?, blocker = ?, unresolved_reason = ?, updated_at = ?
			 WHERE id = ?`,
			stop.Status,
			stop.Blocker,
			stop.Unresolved,
			nowRFC3339(),
			stop.ID,
		)
		if err != nil {
			return fmt.Errorf("update route stop: %w", err)
		}

		jobStatus := stop.Status
		if jobStatus == "complete" {
			jobStatus = "field_complete"
		}
		_, err = tx.ExecContext(
			ctx,
			`UPDATE jobs SET status = ?, updated_at = ? WHERE id = ?`,
			jobStatus,
			nowRFC3339(),
			stop.JobID,
		)
		if err != nil {
			return fmt.Errorf("update job status: %w", err)
		}

		return writeEvent(ctx, tx, "field_checkin.received", "route_stop", stop.ID, map[string]any{
			"status":            stop.Status,
			"blocker":           stop.Blocker,
			"unresolved_reason": stop.Unresolved,
		})
	})
	if err != nil {
		return RouteStop{}, err
	}

	return stop, nil
}

func (s *Store) RouteSummary(ctx context.Context, routeDate string) (RouteDay, []RouteStop, error) {
	var route RouteDay
	err := s.db.QueryRowContext(
		ctx,
		`SELECT id, route_date, status, created_at, updated_at
		 FROM route_days
		 WHERE route_date = ?`,
		routeDate,
	).Scan(&route.ID, &route.RouteDate, &route.Status, &route.CreatedAt, &route.UpdatedAt)
	if err != nil {
		return RouteDay{}, nil, fmt.Errorf("load route day: %w", err)
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT rs.id, rs.route_day_id, rs.job_id, rs.position, rs.status, rs.is_urgent, rs.blocker, rs.unresolved_reason,
		        j.customer_name, j.job_type, j.pest_target, j.priority
		 FROM route_stops rs
		 JOIN jobs j ON j.id = rs.job_id
		 WHERE rs.route_day_id = ?
		 ORDER BY rs.position ASC`,
		route.ID,
	)
	if err != nil {
		return RouteDay{}, nil, fmt.Errorf("query route stops: %w", err)
	}
	defer rows.Close()

	stops := []RouteStop{}
	for rows.Next() {
		var stop RouteStop
		if err := rows.Scan(
			&stop.ID,
			&stop.RouteDayID,
			&stop.JobID,
			&stop.Position,
			&stop.Status,
			&stop.IsUrgent,
			&stop.Blocker,
			&stop.Unresolved,
			&stop.CustomerName,
			&stop.JobType,
			&stop.PestTarget,
			&stop.Priority,
		); err != nil {
			return RouteDay{}, nil, fmt.Errorf("scan route stop: %w", err)
		}
		stops = append(stops, stop)
	}

	return route, stops, nil
}
