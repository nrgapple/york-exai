package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"yorkexai/internal/app"
	"yorkexai/internal/cli"
	"yorkexai/internal/store"
)

func TestInitCreatesCleanRuntime(t *testing.T) {
	home := t.TempDir()

	exitCode, resp := runCLI(t, home, "init")
	if exitCode != 0 {
		t.Fatalf("expected init exit code 0, got %d", exitCode)
	}
	if !resp.OK {
		t.Fatalf("expected init success response, got %+v", resp)
	}

	assertPathExists(t, filepath.Join(home, "config.json"))
	assertPathExists(t, filepath.Join(home, "state", "york.db"))
	assertPathExists(t, filepath.Join(home, "artifacts", "audio"))
	assertPathExists(t, filepath.Join(home, "backups"))
}

func TestRouteUrgentInsertPreservesSpecialtyVisibility(t *testing.T) {
	home := t.TempDir()
	runCLI(t, home, "init")

	_, routeResp := runCLI(t, home, "route", "create", "--date", "2026-03-18")
	routeID := dataMap(t, routeResp)["id"].(string)

	_, termiteResp := runCLI(t, home, "route", "add-stop",
		"--route-id", routeID,
		"--customer", "Termite Customer",
		"--address", "1 Main St",
		"--job-type", "termite inspection",
		"--pest", "termite",
		"--priority", "high",
		"--paperwork-required=true",
	)
	termiteJobID := nestedDataMap(t, termiteResp, "job")["id"].(string)

	_, _ = runCLI(t, home, "route", "insert-urgent",
		"--route-id", routeID,
		"--customer", "Urgent Wasp",
		"--address", "2 Main St",
		"--job-type", "stinging insect response",
		"--pest", "yellowjacket",
		"--priority", "high",
		"--position", "1",
	)

	_, summaryResp := runCLI(t, home, "route", "summary", "--date", "2026-03-18")
	stops := nestedDataSlice(t, summaryResp, "stops")
	if len(stops) != 2 {
		t.Fatalf("expected 2 route stops, got %d", len(stops))
	}
	if int(stops[0]["position"].(float64)) != 1 || stops[0]["pest_target"].(string) != "yellowjacket" {
		t.Fatalf("expected urgent yellowjacket stop first, got %+v", stops[0])
	}
	if int(stops[1]["position"].(float64)) != 2 || stops[1]["job_id"].(string) != termiteJobID {
		t.Fatalf("expected termite stop to remain visible at position 2, got %+v", stops[1])
	}
}

func TestVoiceMemoWithoutTranscriptQueuesReview(t *testing.T) {
	home := t.TempDir()
	runCLI(t, home, "init")
	runCLI(t, home, "route", "create", "--date", "2026-03-18")
	_, stopResp := runCLI(t, home, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "Ant Customer",
		"--address", "10 Pine St",
		"--job-type", "general pest",
		"--pest", "ants",
	)
	jobID := nestedDataMap(t, stopResp, "job")["id"].(string)

	audioPath := writeTempFile(t, home, "memo.m4a", "voice-data")
	exitCode, voiceResp := runCLI(t, home, "field", "voice",
		"--job", jobID,
		"--file", audioPath,
		"--summary", "ants active in kitchen",
		"--confidence", "0.50",
	)
	if exitCode != 2 {
		t.Fatalf("expected review exit code 2, got %d", exitCode)
	}
	if voiceResp.OK {
		t.Fatalf("expected response to indicate review required, got %+v", voiceResp)
	}

	artifact := nestedDataMap(t, voiceResp, "artifact")
	assertPathExists(t, filepath.Join(home, artifact["relative_path"].(string)))

	_, reviewResp := runCLI(t, home, "field", "list-review")
	tasks := asSliceOfMaps(t, reviewResp.Data)
	if len(tasks) < 2 {
		t.Fatalf("expected at least two review tasks for missing transcript and weak job match, got %d", len(tasks))
	}
}

func TestCloseoutAndBillingRules(t *testing.T) {
	home := t.TempDir()
	runCLI(t, home, "init")

	_, termiteResp := runCLI(t, home, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "Termite Customer",
		"--address", "100 Oak St",
		"--job-type", "termite inspection",
		"--pest", "termite",
		"--paperwork-required=true",
	)
	termiteJobID := nestedDataMap(t, termiteResp, "job")["id"].(string)

	_, _ = runCLI(t, home, "closeout", "note",
		"--job", termiteJobID,
		"--service-summary", "Termite evidence found in sill plate.",
		"--inspection-notes", "Mud tubes at rear foundation.",
	)

	exitCode, closeoutResp := runCLI(t, home, "closeout", "evaluate", "--job", termiteJobID)
	if exitCode != 2 {
		t.Fatalf("expected termite closeout review exit code 2, got %d", exitCode)
	}
	result := dataMap(t, closeoutResp)
	if result["closeout_state"].(string) != "blocked" {
		t.Fatalf("expected blocked closeout, got %+v", result)
	}
	if !containsStringFromAny(result["missing_items"], "paperwork_ready") {
		t.Fatalf("expected paperwork_ready blocker, got %+v", result["missing_items"])
	}

	_, recurringResp := runCLI(t, home, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "Recurring Ants",
		"--address", "200 Maple St",
		"--job-type", "general pest",
		"--pest", "ants",
		"--requires-photos=true",
	)
	recurringJobID := nestedDataMap(t, recurringResp, "job")["id"].(string)

	_, _ = runCLI(t, home, "closeout", "note",
		"--job", recurringJobID,
		"--service-summary", "Interior crack and crevice treatment complete.",
		"--treatment-notes", "Baited kitchen and treated exterior foundation.",
	)
	photoPath := writeTempFile(t, home, "photo.jpg", "photo-data")
	_, _ = runCLI(t, home, "closeout", "photo", "--job", recurringJobID, "--file", photoPath)
	_, recurringCloseout := runCLI(t, home, "closeout", "evaluate", "--job", recurringJobID)
	if dataMap(t, recurringCloseout)["closeout_state"].(string) != "complete" {
		t.Fatalf("expected recurring job to be closeout complete")
	}

	exitCode, draftResp := runCLI(t, home, "billing", "draft", "--job", recurringJobID)
	if exitCode != 0 {
		t.Fatalf("expected recurring billing draft success, got %d", exitCode)
	}
	if dataMap(t, draftResp)["state"].(string) != "draft" {
		t.Fatalf("expected draft invoice, got %+v", draftResp.Data)
	}

	_, callbackResp := runCLI(t, home, "route", "add-stop",
		"--date", "2026-03-18",
		"--customer", "Callback Ants",
		"--address", "200 Maple St",
		"--job-type", "general pest callback",
		"--pest", "ants",
		"--callback-of", recurringJobID,
	)
	callbackJobID := nestedDataMap(t, callbackResp, "job")["id"].(string)
	_, _ = runCLI(t, home, "closeout", "note",
		"--job", callbackJobID,
		"--service-summary", "Returned to inspect persistent activity.",
	)
	_, _ = runCLI(t, home, "closeout", "evaluate", "--job", callbackJobID)
	exitCode, holdResp := runCLI(t, home, "billing", "draft", "--job", callbackJobID)
	if exitCode != 2 {
		t.Fatalf("expected callback billing hold exit code 2, got %d", exitCode)
	}
	if dataMap(t, holdResp)["billing_hold_reason"].(string) != "callback_under_review" {
		t.Fatalf("expected callback hold, got %+v", holdResp.Data)
	}
}

func TestConcurrentRouteAddsStayConsistentAndBackupVerifies(t *testing.T) {
	home := t.TempDir()
	runCLI(t, home, "init")

	rt, err := app.LoadOrCreateRuntime(context.Background(), home, true)
	if err != nil {
		t.Fatalf("load runtime: %v", err)
	}
	defer rt.Store.Close()

	route, err := rt.Store.CreateRouteDay(context.Background(), store.CreateRouteDayInput{RouteDate: "2026-03-18"})
	if err != nil {
		t.Fatalf("create route: %v", err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for idx := 0; idx < 2; idx++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _, err := rt.Store.AddStop(context.Background(), store.AddStopInput{
				RouteDayID:      route.ID,
				CustomerName:    "Concurrent Customer",
				PropertyAddress: "300 Cedar St",
				JobType:         "general pest",
				PestTarget:      "spiders",
				Priority:        "medium",
				ScheduledWindow: "am",
			}, false, 0)
			errs <- err
		}(idx)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent add stop error: %v", err)
		}
	}

	_, stops, err := rt.Store.RouteSummary(context.Background(), "2026-03-18")
	if err != nil {
		t.Fatalf("route summary: %v", err)
	}
	if len(stops) != 2 {
		t.Fatalf("expected 2 stops after concurrent adds, got %d", len(stops))
	}
	if stops[0].Position == stops[1].Position {
		t.Fatalf("expected unique route positions, got %+v", stops)
	}

	audioPath := writeTempFile(t, home, "backup-audio.m4a", "backup-audio")
	_, _ = runCLI(t, home, "field", "voice", "--file", audioPath, "--summary", "backup path test")

	exitCode, backupResp := runCLI(t, home, "backup", "create")
	if exitCode != 0 {
		t.Fatalf("expected backup create success, got %d", exitCode)
	}
	backupPath := dataMap(t, backupResp)["path"].(string)

	exitCode, verifyResp := runCLI(t, home, "backup", "verify", "--file", backupPath)
	if exitCode != 0 {
		t.Fatalf("expected backup verify success, got %d", exitCode)
	}
	files := asStringSlice(t, dataMap(t, verifyResp)["files"])
	if !containsString(files, "state/york.db") {
		t.Fatalf("expected backup to include state/york.db, got %v", files)
	}
	if !containsPrefix(files, "artifacts/audio/") {
		t.Fatalf("expected backup to include audio artifact, got %v", files)
	}
}

func runCLI(t *testing.T, home string, args ...string) (int, cli.Response) {
	t.Helper()

	fullArgs := append([]string{"--json", "--home", home}, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := cli.Run(context.Background(), fullArgs, &stdout, &stderr)
	raw := strings.TrimSpace(stdout.String())
	if raw == "" {
		raw = strings.TrimSpace(stderr.String())
	}
	if raw == "" {
		t.Fatalf("expected CLI output, got none")
	}

	var resp cli.Response
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("unmarshal CLI response: %v\nraw=%s", err, raw)
	}
	return exitCode, resp
}

func writeTempFile(t *testing.T, dir string, name string, contents string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path %s to exist: %v", path, err)
	}
}

func dataMap(t *testing.T, resp cli.Response) map[string]any {
	t.Helper()
	return anyToMap(t, resp.Data)
}

func nestedDataMap(t *testing.T, resp cli.Response, key string) map[string]any {
	t.Helper()
	data := dataMap(t, resp)
	return anyToMap(t, data[key])
}

func nestedDataSlice(t *testing.T, resp cli.Response, key string) []map[string]any {
	t.Helper()
	data := dataMap(t, resp)
	return asSliceOfMapsFromAny(t, data[key])
}

func anyToMap(t *testing.T, raw any) map[string]any {
	t.Helper()
	value, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("expected map data, got %#v", raw)
	}
	return value
}

func asSliceOfMaps(t *testing.T, raw any) []map[string]any {
	t.Helper()
	return asSliceOfMapsFromAny(t, raw)
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

func containsStringFromAny(raw any, needle string) bool {
	items, ok := raw.([]any)
	if !ok {
		return false
	}
	for _, item := range items {
		if value, ok := item.(string); ok && value == needle {
			return true
		}
	}
	return false
}

func asStringSlice(t *testing.T, raw any) []string {
	t.Helper()
	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected []any, got %#v", raw)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.(string))
	}
	return out
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func containsPrefix(values []string, prefix string) bool {
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
