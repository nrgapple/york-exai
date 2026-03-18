package store

import (
	"context"
	"fmt"
)

func (s *Store) MorningReport(ctx context.Context, routeDate string) (ReportSummary, error) {
	report := ReportSummary{
		RouteDate:  routeDate,
		StopCounts: map[string]int{},
	}

	_, stops, err := s.RouteSummary(ctx, routeDate)
	if err == nil {
		for _, stop := range stops {
			report.StopCounts[stop.Status]++
		}
	}

	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM callbacks WHERE status = 'open'").Scan(&report.OverdueCallbacks); err != nil {
		return ReportSummary{}, fmt.Errorf("count callbacks: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM jobs WHERE closeout_state = 'blocked'").Scan(&report.BlockedCloseouts); err != nil {
		return ReportSummary{}, fmt.Errorf("count blocked closeouts: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM jobs WHERE closeout_state = 'follow_up_needed'").Scan(&report.FollowUpNeeded); err != nil {
		return ReportSummary{}, fmt.Errorf("count follow-up jobs: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM invoice_drafts WHERE state = 'draft'").Scan(&report.InvoiceReady); err != nil {
		return ReportSummary{}, fmt.Errorf("count invoice-ready drafts: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM invoice_drafts WHERE state = 'hold'").Scan(&report.InvoiceHolds); err != nil {
		return ReportSummary{}, fmt.Errorf("count invoice holds: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM tasks WHERE resolution_state = 'open'").Scan(&report.ReviewTaskCount); err != nil {
		return ReportSummary{}, fmt.Errorf("count review tasks: %w", err)
	}

	return report, nil
}

func (s *Store) RouteRiskReport(ctx context.Context, routeDate string) (ReportSummary, error) {
	_, stops, err := s.RouteSummary(ctx, routeDate)
	if err != nil {
		return ReportSummary{}, err
	}

	report := ReportSummary{
		RouteDate:   routeDate,
		StopCounts:  map[string]int{},
		AtRiskStops: []RouteStop{},
	}
	for _, stop := range stops {
		report.StopCounts[stop.Status]++
		if stop.Status == "at_risk" || stop.Status == "blocked" || stop.Status == "unresolved" {
			report.AtRiskStops = append(report.AtRiskStops, stop)
		}
	}
	return report, nil
}

func (s *Store) BlockedCloseoutReport(ctx context.Context) (ReportSummary, error) {
	jobs, err := s.ListBlockedJobs(ctx)
	if err != nil {
		return ReportSummary{}, err
	}
	return ReportSummary{
		BlockedCloseouts: len(jobs),
		BlockedJobs:      jobs,
	}, nil
}

func (s *Store) CallbackPressureReport(ctx context.Context) (ReportSummary, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT reason, urgency, COUNT(1)
		 FROM callbacks
		 WHERE status = 'open'
		 GROUP BY reason, urgency
		 ORDER BY COUNT(1) DESC, urgency DESC`,
	)
	if err != nil {
		return ReportSummary{}, fmt.Errorf("query callback pressure: %w", err)
	}
	defer rows.Close()

	breakdown := []CallbackSummary{}
	total := 0
	for rows.Next() {
		var item CallbackSummary
		if err := rows.Scan(&item.Reason, &item.Urgency, &item.Count); err != nil {
			return ReportSummary{}, fmt.Errorf("scan callback pressure row: %w", err)
		}
		total += item.Count
		breakdown = append(breakdown, item)
	}

	return ReportSummary{
		OverdueCallbacks:  total,
		CallbackBreakdown: breakdown,
	}, nil
}

func (s *Store) EndDayReport(ctx context.Context, routeDate string) (ReportSummary, error) {
	report := ReportSummary{
		RouteDate:       routeDate,
		UnresolvedStops: []RouteStop{},
	}

	_, stops, err := s.RouteSummary(ctx, routeDate)
	if err != nil {
		return ReportSummary{}, err
	}
	for _, stop := range stops {
		if stop.Status == "unresolved" || stop.Status == "blocked" {
			report.UnresolvedStops = append(report.UnresolvedStops, stop)
		}
	}

	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM jobs WHERE closeout_state = 'blocked'").Scan(&report.BlockedCloseouts); err != nil {
		return ReportSummary{}, fmt.Errorf("count blocked closeouts: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM jobs WHERE closeout_state = 'complete'").Scan(&report.InvoiceReady); err != nil {
		return ReportSummary{}, fmt.Errorf("count invoice-ready jobs: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM invoice_drafts WHERE state = 'hold'").Scan(&report.InvoiceHolds); err != nil {
		return ReportSummary{}, fmt.Errorf("count invoice holds: %w", err)
	}

	return report, nil
}
