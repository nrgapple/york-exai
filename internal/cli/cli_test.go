package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type cliResponse struct {
	OK       bool    `json:"ok"`
	Code     string  `json:"code"`
	Message  string  `json:"message"`
	Data     any     `json:"data"`
	Warnings []issue `json:"warnings"`
	Errors   []issue `json:"errors"`
}

type issue struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

var builtBinaryPath string

func TestMain(m *testing.M) {
	root, err := findRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "find repo root: %v\n", err)
		os.Exit(1)
	}

	tempDir, err := os.MkdirTemp("", "york-e2e-binary-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temp binary dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	builtBinaryPath = filepath.Join(tempDir, "york")
	if err := buildYorkBinary(root, builtBinaryPath); err != nil {
		fmt.Fprintf(os.Stderr, "build york binary: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestVAL001BootstrapAndRuntimeTrust(t *testing.T) {
	home := t.TempDir()

	initResp := mustRunYork(t, home, 0, "init")
	requireResponseCode(t, initResp, true, "init.ok")

	doctorResp := mustRunYork(t, home, 0, "doctor")
	requireResponseCode(t, doctorResp, true, "doctor.ok")
	doctorData := dataMap(t, doctorResp)

	assertPathExists(t, filepath.Join(home, "config.json"))
	assertPathExists(t, filepath.Join(home, "state", "york.db"))
	assertPathExists(t, filepath.Join(home, "artifacts", "audio"))
	assertPathExists(t, filepath.Join(home, "backups"))
	requireEqualString(t, doctorData["db_path"], filepath.Join(home, "state", "york.db"))
	requireEqualString(t, doctorData["artifacts_dir"], filepath.Join(home, "artifacts"))
}

func TestVAL002MorningRouteSetup(t *testing.T) {
	home := t.TempDir()
	mustRunYork(t, home, 0, "init")

	routeResp := mustRunYork(t, home, 0, "route", "create", "--date", "2026-03-18")
	routeID := dataMap(t, routeResp)["id"].(string)

	generalResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--route-id", routeID,
		"--customer", "Smith Residence",
		"--address", "101 Queen St, York PA",
		"--job-type", "general pest",
		"--pest", "ants",
		"--priority", "medium",
		"--window", "am",
	)
	generalJobID := nestedDataMap(t, generalResp, "job")["id"].(string)

	termiteResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--route-id", routeID,
		"--customer", "Keystone Realty",
		"--address", "455 Market St, Harrisburg PA",
		"--job-type", "termite inspection",
		"--pest", "termite",
		"--priority", "high",
		"--window", "midday",
		"--paperwork-required=true",
	)
	termiteJobID := nestedDataMap(t, termiteResp, "job")["id"].(string)

	callbackResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--route-id", routeID,
		"--customer", "Smith Residence Callback",
		"--address", "101 Queen St, York PA",
		"--job-type", "general pest callback",
		"--pest", "ants",
		"--priority", "medium",
		"--window", "pm",
		"--callback-of", generalJobID,
	)
	callbackJobID := nestedDataMap(t, callbackResp, "job")["id"].(string)

	mustRunYork(t, home, 0, "closeout", "note",
		"--job", callbackJobID,
		"--service-summary", "Returned for recurring ant activity.",
		"--treatment-notes", "Rechecked kitchen and exterior entry points.",
	)
	mustRunYork(t, home, 2, "closeout", "evaluate", "--job", callbackJobID)
	callbackDraft := mustRunYork(t, home, 2, "billing", "draft", "--job", callbackJobID)
	requireResponseCode(t, callbackDraft, false, "billing.draft.ok")
	requireEqualString(t, dataMap(t, callbackDraft)["billing_hold_reason"], "callback_under_review")

	mustRunYork(t, home, 0, "closeout", "note",
		"--job", termiteJobID,
		"--service-summary", "Evidence found at rear sill plate.",
		"--inspection-notes", "Mud tubes visible at foundation seam.",
	)
	mustRunYork(t, home, 2, "closeout", "evaluate", "--job", termiteJobID)

	morningResp := mustRunYork(t, home, 0, "report", "morning", "--date", "2026-03-18")
	requireResponseCode(t, morningResp, true, "report.morning.ok")
	morningData := dataMap(t, morningResp)
	stopCounts := anyToMap(t, morningData["stop_counts"])
	if int(stopCounts["scheduled"].(float64)) != 3 {
		t.Fatalf("expected 3 scheduled stops, got %#v", stopCounts)
	}
	if int(morningData["overdue_callbacks"].(float64)) != 1 {
		t.Fatalf("expected 1 overdue callback, got %#v", morningData["overdue_callbacks"])
	}
	if int(morningData["blocked_closeouts"].(float64)) != 1 {
		t.Fatalf("expected 1 blocked closeout, got %#v", morningData["blocked_closeouts"])
	}
	if int(morningData["invoice_holds"].(float64)) != 1 {
		t.Fatalf("expected 1 invoice hold, got %#v", morningData["invoice_holds"])
	}

	summaryResp := mustRunYork(t, home, 0, "route", "summary", "--date", "2026-03-18")
	requireResponseCode(t, summaryResp, true, "route.summary.ok")
	stops := nestedDataSlice(t, summaryResp, "stops")
	if len(stops) != 3 {
		t.Fatalf("expected 3 route stops, got %d", len(stops))
	}
}

func TestVAL003NormalRecurringStopToInvoiceReady(t *testing.T) {
	home := t.TempDir()
	fixtureDir := t.TempDir()
	mustRunYork(t, home, 0, "init")
	mustRunYork(t, home, 0, "route", "create", "--date", "2026-03-18")

	stopResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "Recurring Ants",
		"--address", "200 Maple St, York PA",
		"--job-type", "general pest",
		"--pest", "ants",
		"--requires-photos=true",
	)
	jobID := nestedDataMap(t, stopResp, "job")["id"].(string)

	checkinResp := mustRunYork(t, home, 0, "field", "checkin",
		"--job", jobID,
		"--status", "on site",
		"--text", "Kitchen ants active. Exterior needs hit too.",
	)
	requireResponseCode(t, checkinResp, true, "field.checkin.ok")

	fieldPhoto := writeFixtureFile(t, fixtureDir, "field-ant.jpg", "field-photo")
	fieldPhotoResp := mustRunYork(t, home, 0, "field", "photo", "--job", jobID, "--file", fieldPhoto)
	requireResponseCode(t, fieldPhotoResp, true, "field.photo.ok")

	mustRunYork(t, home, 0, "closeout", "note",
		"--job", jobID,
		"--service-summary", "Interior crack and crevice plus exterior perimeter treatment complete.",
		"--treatment-notes", "Kitchen baiting added. Exterior foundation treated.",
	)
	closeoutPhoto := writeFixtureFile(t, fixtureDir, "closeout-ant.jpg", "closeout-photo")
	closeoutPhotoResp := mustRunYork(t, home, 0, "closeout", "photo", "--job", jobID, "--file", closeoutPhoto)
	requireResponseCode(t, closeoutPhotoResp, true, "closeout.photo.ok")
	closeoutPhotoArtifact := dataMap(t, closeoutPhotoResp)
	assertPathExists(t, filepath.Join(home, closeoutPhotoArtifact["relative_path"].(string)))

	evaluateResp := mustRunYork(t, home, 0, "closeout", "evaluate", "--job", jobID)
	requireResponseCode(t, evaluateResp, true, "closeout.evaluate.ok")
	if dataMap(t, evaluateResp)["closeout_state"].(string) != "complete" {
		t.Fatalf("expected complete closeout, got %#v", evaluateResp.Data)
	}

	draftResp := mustRunYork(t, home, 0, "billing", "draft", "--job", jobID)
	requireResponseCode(t, draftResp, true, "billing.draft.ok")
	if dataMap(t, draftResp)["state"].(string) != "draft" {
		t.Fatalf("expected draft invoice, got %#v", draftResp.Data)
	}

	statusResp := mustRunYork(t, home, 0, "closeout", "status", "--job", jobID)
	requireResponseCode(t, statusResp, true, "closeout.status.ok")
	statusData := dataMap(t, statusResp)
	packet := anyToMap(t, statusData["document_packet"])
	if packet["completeness_status"].(string) != "complete" {
		t.Fatalf("expected complete document packet, got %#v", packet)
	}
	if len(asStringSliceFromAny(t, packet["media_ids"])) == 0 {
		t.Fatalf("expected document packet media ids, got %#v", packet["media_ids"])
	}

	jobEventsResp := mustRunYork(t, home, 0, "report", "events", "--entity-type", "job", "--entity-id", jobID)
	requireResponseCode(t, jobEventsResp, true, "report.events.ok")
	jobEventNames := eventNames(t, jobEventsResp.Data)
	requireContains(t, jobEventNames, "inspection.completed")
	requireContains(t, jobEventNames, "treatment.completed")
	requireContains(t, jobEventNames, "job.closed")
}

func TestVAL004SameDayUrgentInsert(t *testing.T) {
	home := t.TempDir()
	mustRunYork(t, home, 0, "init")

	routeResp := mustRunYork(t, home, 0, "route", "create", "--date", "2026-03-18")
	routeID := dataMap(t, routeResp)["id"].(string)

	generalResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--route-id", routeID,
		"--customer", "Smith Residence",
		"--address", "101 Queen St, York PA",
		"--job-type", "general pest",
		"--pest", "ants",
	)
	_ = generalResp

	termiteResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--route-id", routeID,
		"--customer", "Keystone Realty",
		"--address", "455 Market St, Harrisburg PA",
		"--job-type", "termite inspection",
		"--pest", "termite",
		"--priority", "high",
		"--paperwork-required=true",
	)
	termiteJobID := nestedDataMap(t, termiteResp, "job")["id"].(string)

	mustRunYork(t, home, 0, "route", "insert-urgent",
		"--route-id", routeID,
		"--customer", "Brown Residence",
		"--address", "22 Hill Rd, Lancaster PA",
		"--job-type", "stinging insect response",
		"--pest", "yellowjacket",
		"--priority", "high",
		"--position", "1",
	)

	summaryResp := mustRunYork(t, home, 0, "route", "summary", "--date", "2026-03-18")
	requireResponseCode(t, summaryResp, true, "route.summary.ok")
	stops := nestedDataSlice(t, summaryResp, "stops")
	if len(stops) != 3 {
		t.Fatalf("expected 3 route stops, got %d", len(stops))
	}
	if stops[0]["pest_target"].(string) != "yellowjacket" {
		t.Fatalf("expected urgent yellowjacket stop first, got %#v", stops[0])
	}
	if stops[2]["job_id"].(string) != termiteJobID {
		t.Fatalf("expected termite job to remain visible in route summary, got %#v", stops[2])
	}

	routeRiskResp := mustRunYork(t, home, 0, "report", "route-risk", "--date", "2026-03-18")
	requireResponseCode(t, routeRiskResp, true, "report.route_risk.ok")
}

func TestVAL005VoiceMemoFallback(t *testing.T) {
	home := t.TempDir()
	fixtureDir := t.TempDir()
	mustRunYork(t, home, 0, "init")
	mustRunYork(t, home, 0, "route", "create", "--date", "2026-03-18")

	stopResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "Ant Customer",
		"--address", "10 Pine St, York PA",
		"--job-type", "general pest",
		"--pest", "ants",
	)
	jobID := nestedDataMap(t, stopResp, "job")["id"].(string)

	audioPath := writeFixtureFile(t, fixtureDir, "ants-update.m4a", "voice-data")
	voiceResp := mustRunYork(t, home, 2, "field", "voice",
		"--job", jobID,
		"--file", audioPath,
		"--summary", "Ants still active in kitchen.",
		"--confidence", "0.50",
	)
	requireResponseCode(t, voiceResp, false, "field.voice.ok")
	voiceData := dataMap(t, voiceResp)
	voiceMemoID := voiceData["voice_memo_id"].(string)
	artifact := anyToMap(t, voiceData["artifact"])
	assertPathExists(t, filepath.Join(home, artifact["relative_path"].(string)))

	reviewResp := mustRunYork(t, home, 0, "field", "list-review")
	requireResponseCode(t, reviewResp, true, "field.review.ok")
	tasks := asSliceOfMapsFromAny(t, reviewResp.Data)
	if len(tasks) < 2 {
		t.Fatalf("expected at least two review tasks, got %d", len(tasks))
	}
	taskReasons := collectField(tasks, "reason")
	requireContains(t, taskReasons, "transcription_unavailable")
	requireContains(t, taskReasons, "job_link_unclear")

	eventResp := mustRunYork(t, home, 0, "report", "events", "--entity-type", "voice_memo", "--entity-id", voiceMemoID)
	requireResponseCode(t, eventResp, true, "report.events.ok")
	eventNames := eventNames(t, eventResp.Data)
	requireContains(t, eventNames, "voice_memo.received")
	requireNotContains(t, eventNames, "voice_memo.transcribed")
}

func TestVAL006TermiteWDIBlocker(t *testing.T) {
	home := t.TempDir()
	mustRunYork(t, home, 0, "init")

	stopResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "Termite Customer",
		"--address", "100 Oak St, York PA",
		"--job-type", "termite inspection",
		"--pest", "termite",
		"--paperwork-required=true",
	)
	jobID := nestedDataMap(t, stopResp, "job")["id"].(string)

	mustRunYork(t, home, 0, "closeout", "note",
		"--job", jobID,
		"--service-summary", "Termite evidence found in sill plate.",
		"--inspection-notes", "Mud tubes at rear foundation.",
	)
	evaluateResp := mustRunYork(t, home, 2, "closeout", "evaluate", "--job", jobID)
	requireResponseCode(t, evaluateResp, false, "closeout.evaluate.ok")
	requireContains(t, asStringSliceFromAny(t, dataMap(t, evaluateResp)["missing_items"]), "paperwork_ready")

	statusResp := mustRunYork(t, home, 0, "closeout", "status", "--job", jobID)
	packet := anyToMap(t, dataMap(t, statusResp)["document_packet"])
	if packet["completeness_status"].(string) != "blocked" {
		t.Fatalf("expected blocked document packet, got %#v", packet)
	}

	draftResp := mustRunYork(t, home, 2, "billing", "draft", "--job", jobID)
	requireResponseCode(t, draftResp, false, "billing.draft.ok")
	requireEqualString(t, dataMap(t, draftResp)["billing_hold_reason"], "closeout_incomplete")

	blockedResp := mustRunYork(t, home, 0, "report", "blocked-closeouts")
	requireResponseCode(t, blockedResp, true, "report.blocked_closeouts.ok")
	if int(dataMap(t, blockedResp)["blocked_closeouts"].(float64)) != 1 {
		t.Fatalf("expected 1 blocked closeout, got %#v", blockedResp.Data)
	}

	eventResp := mustRunYork(t, home, 0, "report", "events", "--entity-type", "job", "--entity-id", jobID)
	eventNames := eventNames(t, eventResp.Data)
	requireContains(t, eventNames, "inspection.completed")
	requireContains(t, eventNames, "job.closeout.blocked")
}

func TestVAL007BedBugPrepAndFollowUpPath(t *testing.T) {
	home := t.TempDir()
	mustRunYork(t, home, 0, "init")

	stopResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "Miller Apartment",
		"--address", "88 Locust St, York PA",
		"--job-type", "bed bug treatment",
		"--pest", "bed bugs",
		"--priority", "high",
		"--requires-prep=true",
		"--requires-follow-up=true",
	)
	jobID := nestedDataMap(t, stopResp, "job")["id"].(string)

	prepRequiredResp := mustRunYork(t, home, 0, "closeout", "prep",
		"--job", jobID,
		"--required=true",
		"--completed=false",
		"--follow-up-plan", "Return after prep is completed.",
	)
	requireResponseCode(t, prepRequiredResp, true, "closeout.prep.ok")

	mustRunYork(t, home, 0, "closeout", "note",
		"--job", jobID,
		"--service-summary", "Prep incomplete on arrival.",
	)
	evaluateBlocked := mustRunYork(t, home, 2, "closeout", "evaluate", "--job", jobID)
	requireResponseCode(t, evaluateBlocked, false, "closeout.evaluate.ok")
	missing := asStringSliceFromAny(t, dataMap(t, evaluateBlocked)["missing_items"])
	requireContains(t, missing, "treatment_notes")
	requireContains(t, missing, "prep_complete")

	prepClearedResp := mustRunYork(t, home, 0, "closeout", "prep",
		"--job", jobID,
		"--required=true",
		"--completed=true",
		"--follow-up-plan", "Reinspect in 7 days.",
	)
	requireResponseCode(t, prepClearedResp, true, "closeout.prep.ok")

	statusResp := mustRunYork(t, home, 0, "closeout", "status", "--job", jobID)
	packet := anyToMap(t, dataMap(t, statusResp)["document_packet"])
	if !packet["prep_required"].(bool) || !packet["prep_complete"].(bool) {
		t.Fatalf("expected prep flags to be true, got %#v", packet)
	}
	requireEqualString(t, packet["follow_up_plan"], "Reinspect in 7 days.")

	eventResp := mustRunYork(t, home, 0, "report", "events", "--entity-type", "job", "--entity-id", jobID)
	eventNames := eventNames(t, eventResp.Data)
	requireContains(t, eventNames, "prep_notice.required")
	requireContains(t, eventNames, "prep_notice.cleared")
	requireContains(t, eventNames, "job.closeout.blocked")
}

func TestVAL008CallbackBillingHold(t *testing.T) {
	home := t.TempDir()
	mustRunYork(t, home, 0, "init")

	originResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "Smith Residence",
		"--address", "101 Queen St, York PA",
		"--job-type", "general pest",
		"--pest", "ants",
	)
	originJobID := nestedDataMap(t, originResp, "job")["id"].(string)

	callbackResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "Smith Residence Callback",
		"--address", "101 Queen St, York PA",
		"--job-type", "general pest callback",
		"--pest", "ants",
		"--callback-of", originJobID,
	)
	callbackJobID := nestedDataMap(t, callbackResp, "job")["id"].(string)

	mustRunYork(t, home, 0, "closeout", "note",
		"--job", callbackJobID,
		"--service-summary", "Returned to inspect persistent activity.",
		"--treatment-notes", "Confirmed persistent activity near kitchen sink.",
	)
	evaluateResp := mustRunYork(t, home, 2, "closeout", "evaluate", "--job", callbackJobID)
	requireResponseCode(t, evaluateResp, false, "closeout.evaluate.ok")
	followUps := asStringSliceFromAny(t, dataMap(t, evaluateResp)["follow_up_required"])
	requireContains(t, followUps, "callback_under_review")

	draftResp := mustRunYork(t, home, 2, "billing", "draft", "--job", callbackJobID)
	requireResponseCode(t, draftResp, false, "billing.draft.ok")
	requireEqualString(t, dataMap(t, draftResp)["billing_hold_reason"], "callback_under_review")

	holdsResp := mustRunYork(t, home, 0, "billing", "list-holds")
	holds := asSliceOfMapsFromAny(t, holdsResp.Data)
	if len(holds) == 0 {
		t.Fatalf("expected at least one billing hold")
	}

	originEventResp := mustRunYork(t, home, 0, "report", "events", "--entity-type", "job", "--entity-id", originJobID)
	originEvents := eventNames(t, originEventResp.Data)
	requireContains(t, originEvents, "callback.requested")

	callbackEventResp := mustRunYork(t, home, 0, "report", "events", "--entity-type", "job", "--entity-id", callbackJobID)
	callbackEvents := eventNames(t, callbackEventResp.Data)
	requireContains(t, callbackEvents, "callback.scheduled")
}

func TestVAL009MissedCheckinRecovery(t *testing.T) {
	home := t.TempDir()
	mustRunYork(t, home, 0, "init")

	mustRunYork(t, home, 0, "route", "create", "--date", "2026-03-18")
	stopResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "No Checkin Stop",
		"--address", "500 Duke St, York PA",
		"--job-type", "general pest",
		"--pest", "spiders",
	)
	stopID := nestedDataMap(t, stopResp, "stop")["id"].(string)

	updateResp := mustRunYork(t, home, 0, "route", "update-stop",
		"--stop-id", stopID,
		"--status", "unresolved",
		"--unresolved-reason", "missed check-in",
	)
	requireResponseCode(t, updateResp, true, "route.update.ok")

	endDayResp := mustRunYork(t, home, 0, "report", "end-day", "--date", "2026-03-18")
	requireResponseCode(t, endDayResp, true, "report.end_day.ok")
	endDayData := dataMap(t, endDayResp)
	unresolvedStops := asSliceOfMapsFromAny(t, endDayData["unresolved_stops"])
	if len(unresolvedStops) != 1 {
		t.Fatalf("expected 1 unresolved stop, got %#v", unresolvedStops)
	}
	requireEqualString(t, unresolvedStops[0]["unresolved_reason"], "missed check-in")

	eventResp := mustRunYork(t, home, 0, "report", "events", "--entity-type", "route_stop", "--entity-id", stopID)
	eventNames := eventNames(t, eventResp.Data)
	requireContains(t, eventNames, "field_checkin.received")
}

func TestVAL010EndDayAndBackupTrust(t *testing.T) {
	home := t.TempDir()
	fixtureDir := t.TempDir()
	mustRunYork(t, home, 0, "init")
	mustRunYork(t, home, 0, "route", "create", "--date", "2026-03-18")

	recurringResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "Recurring Ants",
		"--address", "200 Maple St, York PA",
		"--job-type", "general pest",
		"--pest", "ants",
		"--requires-photos=true",
	)
	recurringJobID := nestedDataMap(t, recurringResp, "job")["id"].(string)
	mustRunYork(t, home, 0, "closeout", "note",
		"--job", recurringJobID,
		"--service-summary", "Interior crack and crevice plus exterior perimeter treatment complete.",
		"--treatment-notes", "Kitchen baiting added. Exterior foundation treated.",
	)
	recurringPhoto := writeFixtureFile(t, fixtureDir, "recurring-photo.jpg", "recurring-photo")
	mustRunYork(t, home, 0, "closeout", "photo", "--job", recurringJobID, "--file", recurringPhoto)
	mustRunYork(t, home, 0, "closeout", "evaluate", "--job", recurringJobID)
	mustRunYork(t, home, 0, "billing", "draft", "--job", recurringJobID)

	termiteResp := mustRunYork(t, home, 0, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "Blocked Termite",
		"--address", "900 Walnut St, York PA",
		"--job-type", "termite inspection",
		"--pest", "termite",
		"--paperwork-required=true",
	)
	termiteJobID := nestedDataMap(t, termiteResp, "job")["id"].(string)
	mustRunYork(t, home, 0, "closeout", "note",
		"--job", termiteJobID,
		"--service-summary", "Evidence found under sill plate.",
		"--inspection-notes", "Activity at rear joist bay.",
	)
	mustRunYork(t, home, 2, "closeout", "evaluate", "--job", termiteJobID)

	audioPath := writeFixtureFile(t, fixtureDir, "backup-audio.m4a", "backup-audio")
	mustRunYork(t, home, 2, "field", "voice", "--file", audioPath, "--summary", "backup path test")

	endDayResp := mustRunYork(t, home, 0, "report", "end-day", "--date", "2026-03-18")
	requireResponseCode(t, endDayResp, true, "report.end_day.ok")
	endDayData := dataMap(t, endDayResp)
	if int(endDayData["blocked_closeouts"].(float64)) != 1 {
		t.Fatalf("expected 1 blocked closeout, got %#v", endDayData["blocked_closeouts"])
	}
	if int(endDayData["invoice_ready"].(float64)) != 1 {
		t.Fatalf("expected 1 invoice-ready job, got %#v", endDayData["invoice_ready"])
	}

	backupResp := mustRunYork(t, home, 0, "backup", "create")
	requireResponseCode(t, backupResp, true, "backup.create.ok")
	backupPath := dataMap(t, backupResp)["path"].(string)
	assertPathExists(t, backupPath)

	verifyResp := mustRunYork(t, home, 0, "backup", "verify", "--file", backupPath)
	requireResponseCode(t, verifyResp, true, "backup.verify.ok")
	verifyData := dataMap(t, verifyResp)
	files := asStringSliceFromAny(t, verifyData["files"])
	requireContains(t, files, "state/york.db")
	requireContainsPrefix(t, files, "artifacts/audio/")
	requireContainsPrefix(t, files, "artifacts/photos/")
}

func mustRunYork(t *testing.T, home string, expectedExit int, args ...string) cliResponse {
	t.Helper()

	command := exec.CommandContext(context.Background(), builtBinaryPath, append([]string{"--json", "--home", home}, args...)...)
	command.Dir = home
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("run york command %v: %v", args, err)
		}
	}

	if exitCode != expectedExit {
		t.Fatalf("expected exit code %d for %v, got %d\nstdout=%s\nstderr=%s", expectedExit, args, exitCode, stdout.String(), stderr.String())
	}

	raw := strings.TrimSpace(stdout.String())
	if raw == "" {
		raw = strings.TrimSpace(stderr.String())
	}
	if raw == "" {
		t.Fatalf("expected CLI JSON output for %v", args)
	}

	var response cliResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		t.Fatalf("decode CLI response for %v: %v\nraw=%s\nstderr=%s", args, err, raw, stderr.String())
	}

	return response
}

func requireResponseCode(t *testing.T, response cliResponse, expectedOK bool, expectedCode string) {
	t.Helper()
	if response.OK != expectedOK {
		t.Fatalf("expected OK=%v, got %#v", expectedOK, response)
	}
	if response.Code != expectedCode {
		t.Fatalf("expected response code %s, got %#v", expectedCode, response)
	}
}

func buildYorkBinary(root string, binaryPath string) error {
	command := exec.Command("go", "build", "-o", binaryPath, "./cmd/york")
	command.Dir = root
	command.Env = goBuildEnv()
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

func goBuildEnv() []string {
	env := os.Environ()
	tempBase := filepath.Join(os.TempDir(), "york-go-e2e")
	appendIfMissing := func(key string, value string) {
		prefix := key + "="
		for _, item := range env {
			if strings.HasPrefix(item, prefix) {
				return
			}
		}
		env = append(env, prefix+value)
	}

	appendIfMissing("GOCACHE", filepath.Join(tempBase, "cache"))
	appendIfMissing("GOPATH", filepath.Join(tempBase, "gopath"))
	appendIfMissing("GOMODCACHE", filepath.Join(tempBase, "gopath", "pkg", "mod"))
	return env
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	current := wd
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("go.mod not found from %s", wd)
		}
		current = parent
	}
}

func writeFixtureFile(t *testing.T, dir string, name string, contents string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}
	return path
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path %s to exist: %v", path, err)
	}
}

func dataMap(t *testing.T, response cliResponse) map[string]any {
	t.Helper()
	return anyToMap(t, response.Data)
}

func nestedDataMap(t *testing.T, response cliResponse, key string) map[string]any {
	t.Helper()
	return anyToMap(t, dataMap(t, response)[key])
}

func nestedDataSlice(t *testing.T, response cliResponse, key string) []map[string]any {
	t.Helper()
	return asSliceOfMapsFromAny(t, dataMap(t, response)[key])
}

func anyToMap(t *testing.T, raw any) map[string]any {
	t.Helper()
	value, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("expected map data, got %#v", raw)
	}
	return value
}

func asSliceOfMapsFromAny(t *testing.T, raw any) []map[string]any {
	t.Helper()
	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected slice data, got %#v", raw)
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, anyToMap(t, item))
	}
	return out
}

func asStringSliceFromAny(t *testing.T, raw any) []string {
	t.Helper()
	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected []any, got %#v", raw)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			t.Fatalf("expected string item, got %#v", item)
		}
		out = append(out, value)
	}
	return out
}

func eventNames(t *testing.T, raw any) []string {
	t.Helper()
	events := asSliceOfMapsFromAny(t, raw)
	names := make([]string, 0, len(events))
	for _, event := range events {
		names = append(names, event["event_name"].(string))
	}
	return names
}

func collectField(items []map[string]any, key string) []string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		if value, ok := item[key].(string); ok {
			values = append(values, value)
		}
	}
	return values
}

func requireContains(t *testing.T, values []string, expected string) {
	t.Helper()
	for _, value := range values {
		if value == expected {
			return
		}
	}
	t.Fatalf("expected %q in %v", expected, values)
}

func requireNotContains(t *testing.T, values []string, forbidden string) {
	t.Helper()
	for _, value := range values {
		if value == forbidden {
			t.Fatalf("did not expect %q in %v", forbidden, values)
		}
	}
}

func requireContainsPrefix(t *testing.T, values []string, prefix string) {
	t.Helper()
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return
		}
	}
	t.Fatalf("expected prefix %q in %v", prefix, values)
}

func requireEqualString(t *testing.T, raw any, expected string) {
	t.Helper()
	value, ok := raw.(string)
	if !ok {
		t.Fatalf("expected string value, got %#v", raw)
	}
	if value != expected {
		t.Fatalf("expected %q, got %q", expected, value)
	}
}
