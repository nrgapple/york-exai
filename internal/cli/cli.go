package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"yorkexai/internal/app"
	"yorkexai/internal/store"
)

const (
	exitSuccess = 0
	exitUsage   = 1
	exitReview  = 2
	exitStorage = 3
	exitExec    = 4
)

func Run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) int {
	global, remaining, err := parseGlobals(args)
	if err != nil {
		_ = writeResponse(stderr, global.JSON, Response{
			OK:      false,
			Code:    "usage.invalid",
			Message: err.Error(),
			Errors:  []Issue{{Code: "usage.invalid", Detail: err.Error()}},
		})
		return exitUsage
	}
	if len(remaining) == 0 {
		_ = writeResponse(stderr, global.JSON, Response{
			OK:      false,
			Code:    "usage.missing_command",
			Message: "Command is required.",
			Errors:  []Issue{{Code: "usage.missing_command", Detail: "expected a top-level command"}},
		})
		return exitUsage
	}

	command := remaining[0]
	commandArgs := remaining[1:]

	switch command {
	case "init":
		return runInit(ctx, global, commandArgs, stdout, stderr)
	case "doctor":
		return runDoctor(ctx, global, commandArgs, stdout, stderr)
	case "route":
		return runRoute(ctx, global, commandArgs, stdout, stderr)
	case "field":
		return runField(ctx, global, commandArgs, stdout, stderr)
	case "closeout":
		return runCloseout(ctx, global, commandArgs, stdout, stderr)
	case "billing":
		return runBilling(ctx, global, commandArgs, stdout, stderr)
	case "report":
		return runReport(ctx, global, commandArgs, stdout, stderr)
	case "backup":
		return runBackup(ctx, global, commandArgs, stdout, stderr)
	default:
		_ = writeResponse(stderr, global.JSON, Response{
			OK:      false,
			Code:    "usage.unknown_command",
			Message: fmt.Sprintf("Unknown command: %s", command),
			Errors:  []Issue{{Code: "usage.unknown_command", Detail: command}},
		})
		return exitUsage
	}
}

type globalFlags struct {
	JSON bool
	Home string
}

func parseGlobals(args []string) (globalFlags, []string, error) {
	flags := globalFlags{}
	remaining := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--json":
			flags.JSON = true
		case "--home":
			if i+1 >= len(args) {
				return flags, nil, errors.New("--home requires a value")
			}
			i++
			flags.Home = args[i]
		default:
			if strings.HasPrefix(arg, "--home=") {
				flags.Home = strings.TrimPrefix(arg, "--home=")
				continue
			}
			remaining = append(remaining, args[i:]...)
			return flags, remaining, nil
		}
	}

	return flags, remaining, nil
}

func runtimeFor(ctx context.Context, global globalFlags, autoInit bool) (*app.Runtime, error) {
	return app.LoadOrCreateRuntime(ctx, global.Home, autoInit)
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func usageError(stderr io.Writer, asJSON bool, err error) int {
	_ = writeResponse(stderr, asJSON, Response{
		OK:      false,
		Code:    "usage.invalid",
		Message: err.Error(),
		Errors:  []Issue{{Code: "usage.invalid", Detail: err.Error()}},
	})
	return exitUsage
}

func runtimeError(stderr io.Writer, asJSON bool, err error) int {
	code := exitExec
	if strings.Contains(err.Error(), "config missing") || strings.Contains(err.Error(), "database") {
		code = exitStorage
	}
	_ = writeResponse(stderr, asJSON, Response{
		OK:      false,
		Code:    "runtime.error",
		Message: err.Error(),
		Errors:  []Issue{{Code: "runtime.error", Detail: err.Error()}},
	})
	return code
}

func parseOptionalBool(value string) (*bool, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, fmt.Errorf("invalid boolean value %q", value)
	}
	return &parsed, nil
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func runInit(ctx context.Context, global globalFlags, args []string, stdout io.Writer, stderr io.Writer) int {
	fs := newFlagSet("init")
	if err := fs.Parse(args); err != nil {
		return usageError(stderr, global.JSON, err)
	}

	rt, err := runtimeFor(ctx, global, true)
	if err != nil {
		return runtimeError(stderr, global.JSON, err)
	}
	defer rt.Store.Close()

	resp := Response{
		OK:      true,
		Code:    "init.ok",
		Message: "York runtime initialized.",
		Data: map[string]any{
			"home":        rt.Paths.Home,
			"db_path":     rt.Paths.DBPath,
			"config_path": rt.Paths.ConfigPath,
		},
	}
	if err := writeResponse(stdout, global.JSON, resp); err != nil {
		return runtimeError(stderr, global.JSON, err)
	}
	return exitSuccess
}

func runDoctor(ctx context.Context, global globalFlags, args []string, stdout io.Writer, stderr io.Writer) int {
	fs := newFlagSet("doctor")
	if err := fs.Parse(args); err != nil {
		return usageError(stderr, global.JSON, err)
	}

	rt, err := runtimeFor(ctx, global, true)
	if err != nil {
		return runtimeError(stderr, global.JSON, err)
	}
	defer rt.Store.Close()

	schemaVersion, err := rt.Store.SchemaVersion(ctx)
	if err != nil {
		return runtimeError(stderr, global.JSON, err)
	}

	integrations := map[string]map[string]string{}
	for name, cfg := range rt.Config.Integrations {
		integrations[name] = map[string]string{
			"status": cfg.Status,
			"mode":   cfg.Mode,
			"notes":  cfg.Notes,
		}
	}

	resp := Response{
		OK:      true,
		Code:    "doctor.ok",
		Message: "Runtime health check complete.",
		Data: map[string]any{
			"home":           rt.Paths.Home,
			"config_path":    rt.Paths.ConfigPath,
			"db_path":        rt.Paths.DBPath,
			"schema_version": schemaVersion,
			"artifacts_dir":  rt.Paths.ArtifactsDir,
			"backups_dir":    rt.Paths.BackupsDir,
			"integrations":   integrations,
		},
	}
	if err := writeResponse(stdout, global.JSON, resp); err != nil {
		return runtimeError(stderr, global.JSON, err)
	}
	return exitSuccess
}

func runRoute(ctx context.Context, global globalFlags, args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, global.JSON, errors.New("route subcommand is required"))
	}

	rt, err := runtimeFor(ctx, global, true)
	if err != nil {
		return runtimeError(stderr, global.JSON, err)
	}
	defer rt.Store.Close()

	switch args[0] {
	case "create":
		fs := newFlagSet("route create")
		routeDate := fs.String("date", "", "Route date in YYYY-MM-DD format.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *routeDate == "" {
			return usageError(stderr, global.JSON, errors.New("--date is required"))
		}
		route, err := rt.Store.CreateRouteDay(ctx, store.CreateRouteDayInput{RouteDate: *routeDate})
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: true, Code: "route.create.ok", Message: "Route day ready.", Data: route})
		return exitSuccess
	case "add-stop", "insert-urgent":
		fs := newFlagSet(args[0])
		routeDate := fs.String("date", "", "Route date in YYYY-MM-DD format.")
		routeID := fs.String("route-id", "", "Route day ID.")
		jobID := fs.String("job", "", "Existing job ID.")
		customer := fs.String("customer", "", "Customer name.")
		address := fs.String("address", "", "Property address.")
		jobType := fs.String("job-type", "", "Job type.")
		pest := fs.String("pest", "", "Pest target.")
		priority := fs.String("priority", "medium", "Job priority.")
		window := fs.String("window", "unspecified", "Scheduled window.")
		callbackOf := fs.String("callback-of", "", "Origin job ID if this stop is a callback.")
		position := fs.Int("position", 0, "Insert position for urgent work.")
		paperworkRequired := fs.Bool("paperwork-required", false, "Whether paperwork is required.")
		requiresPhotos := fs.Bool("requires-photos", false, "Whether photos are required.")
		requiresPrep := fs.Bool("requires-prep", false, "Whether prep is required.")
		requiresFollowUp := fs.Bool("requires-follow-up", false, "Whether follow-up is required.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *routeDate == "" && *routeID == "" {
			return usageError(stderr, global.JSON, errors.New("--date or --route-id is required"))
		}
		if *jobID == "" && (*customer == "" || *address == "" || *jobType == "" || *pest == "") {
			return usageError(stderr, global.JSON, errors.New("new stops require --customer, --address, --job-type, and --pest"))
		}

		stop, job, err := rt.Store.AddStop(ctx, store.AddStopInput{
			RouteDayID:        *routeID,
			RouteDate:         *routeDate,
			JobID:             *jobID,
			CustomerName:      *customer,
			PropertyAddress:   *address,
			JobType:           *jobType,
			PestTarget:        *pest,
			Priority:          *priority,
			ScheduledWindow:   *window,
			CallbackOfJobID:   *callbackOf,
			PaperworkRequired: *paperworkRequired,
			RequiresPhotos:    *requiresPhotos,
			RequiresPrep:      *requiresPrep,
			RequiresFollowUp:  *requiresFollowUp,
		}, args[0] == "insert-urgent", *position)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{
			OK:      true,
			Code:    "route.stop.ok",
			Message: "Route stop saved.",
			Data:    map[string]any{"stop": stop, "job": job},
		})
		return exitSuccess
	case "update-stop":
		fs := newFlagSet("route update-stop")
		stopID := fs.String("stop-id", "", "Route stop ID.")
		status := fs.String("status", "", "Stop status.")
		blocker := fs.String("blocker", "", "Blocker note.")
		unresolved := fs.String("unresolved-reason", "", "Unresolved reason.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *stopID == "" || *status == "" {
			return usageError(stderr, global.JSON, errors.New("--stop-id and --status are required"))
		}
		stop, err := rt.Store.UpdateStop(ctx, store.UpdateStopInput{
			RouteStopID:      *stopID,
			Status:           *status,
			Blocker:          *blocker,
			UnresolvedReason: *unresolved,
		})
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: true, Code: "route.update.ok", Message: "Route stop updated.", Data: stop})
		return exitSuccess
	case "summary":
		fs := newFlagSet("route summary")
		routeDate := fs.String("date", "", "Route date in YYYY-MM-DD format.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *routeDate == "" {
			return usageError(stderr, global.JSON, errors.New("--date is required"))
		}
		route, stops, err := rt.Store.RouteSummary(ctx, *routeDate)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{
			OK: true, Code: "route.summary.ok", Message: "Route summary ready.",
			Data: map[string]any{"route": route, "stops": stops},
		})
		return exitSuccess
	default:
		return usageError(stderr, global.JSON, fmt.Errorf("unknown route subcommand: %s", args[0]))
	}
}

func runField(ctx context.Context, global globalFlags, args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, global.JSON, errors.New("field subcommand is required"))
	}

	rt, err := runtimeFor(ctx, global, true)
	if err != nil {
		return runtimeError(stderr, global.JSON, err)
	}
	defer rt.Store.Close()

	switch args[0] {
	case "checkin":
		fs := newFlagSet("field checkin")
		source := fs.String("source", "imessage", "Source channel.")
		jobID := fs.String("job", "", "Linked job ID.")
		routeID := fs.String("route-id", "", "Linked route day ID.")
		status := fs.String("status", "update", "Short status.")
		text := fs.String("text", "", "Raw field text.")
		blockers := fs.String("blockers", "", "Comma-separated blockers.")
		confidence := fs.Float64("confidence", 1.0, "Job match confidence.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *text == "" {
			return usageError(stderr, global.JSON, errors.New("--text is required"))
		}
		checkInID, tasks, err := rt.Store.CreateFieldCheckIn(ctx, store.FieldCheckInInput{
			SourceChannel: *source,
			RouteDayID:    *routeID,
			JobID:         *jobID,
			ShortStatus:   *status,
			RawText:       *text,
			Blockers:      splitCSV(*blockers),
			Confidence:    *confidence,
		})
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		exitCode := exitSuccess
		if len(tasks) > 0 {
			exitCode = exitReview
		}
		_ = writeResponse(stdout, global.JSON, Response{
			OK:      len(tasks) == 0,
			Code:    "field.checkin.ok",
			Message: "Field check-in captured.",
			Data:    map[string]any{"field_checkin_id": checkInID, "review_tasks": tasks},
		})
		return exitCode
	case "photo":
		fs := newFlagSet("field photo")
		source := fs.String("source", "imessage", "Source channel.")
		jobID := fs.String("job", "", "Linked job ID.")
		filePath := fs.String("file", "", "Path to the photo file.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *filePath == "" {
			return usageError(stderr, global.JSON, errors.New("--file is required"))
		}
		ownerID := *jobID
		if ownerID == "" {
			ownerID = "unlinked"
		}
		persisted, err := rt.PersistArtifact(*filePath, "photo", ownerID)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		artifact, err := rt.Store.RecordArtifact(ctx, "photo", "job", *jobID, persisted.RelativePath, persisted.OriginalName, persisted.SHA256)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		checkInID, err := rt.Store.RecordFieldPhoto(ctx, *jobID, *source, artifact)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		exitCode := exitSuccess
		ok := true
		if *jobID == "" {
			exitCode = exitReview
			ok = false
		}
		_ = writeResponse(stdout, global.JSON, Response{
			OK:      ok,
			Code:    "field.photo.ok",
			Message: "Field photo captured.",
			Data:    map[string]any{"field_checkin_id": checkInID, "artifact": artifact},
		})
		return exitCode
	case "voice":
		fs := newFlagSet("field voice")
		source := fs.String("source", "imessage", "Source channel.")
		jobID := fs.String("job", "", "Linked job ID.")
		filePath := fs.String("file", "", "Path to the audio file.")
		transcript := fs.String("transcript", "", "Transcript text.")
		summary := fs.String("summary", "", "Short summary.")
		pestFacts := fs.String("pest-facts", "", "Comma-separated pest facts.")
		treatmentFacts := fs.String("treatment-facts", "", "Comma-separated treatment facts.")
		followUps := fs.String("follow-up-needs", "", "Comma-separated follow-up needs.")
		billingChanges := fs.String("billing-changes", "", "Comma-separated billing changes.")
		contentIdeas := fs.String("content-ideas", "", "Comma-separated content ideas.")
		confidence := fs.Float64("confidence", 1.0, "Job match confidence.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *filePath == "" {
			return usageError(stderr, global.JSON, errors.New("--file is required"))
		}
		ownerID := *jobID
		if ownerID == "" {
			ownerID = "unlinked"
		}
		persisted, err := rt.PersistArtifact(*filePath, "audio", ownerID)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		artifact, err := rt.Store.RecordArtifact(ctx, "audio", "voice_memo", ownerID, persisted.RelativePath, persisted.OriginalName, persisted.SHA256)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		voiceMemoID, tasks, err := rt.Store.RecordVoiceMemo(ctx, store.VoiceMemoInput{
			SourceChannel:   *source,
			JobID:           *jobID,
			AudioSourcePath: *filePath,
			Transcript:      *transcript,
			Summary:         *summary,
			PestFacts:       splitCSV(*pestFacts),
			TreatmentFacts:  splitCSV(*treatmentFacts),
			FollowUpNeeds:   splitCSV(*followUps),
			BillingChanges:  splitCSV(*billingChanges),
			ContentIdeas:    splitCSV(*contentIdeas),
			Confidence:      *confidence,
		}, artifact)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		exitCode := exitSuccess
		ok := true
		if len(tasks) > 0 {
			ok = false
			exitCode = exitReview
		}
		_ = writeResponse(stdout, global.JSON, Response{
			OK:      ok,
			Code:    "field.voice.ok",
			Message: "Voice memo captured.",
			Data:    map[string]any{"voice_memo_id": voiceMemoID, "artifact": artifact, "review_tasks": tasks},
		})
		return exitCode
	case "list-review":
		fs := newFlagSet("field list-review")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		tasks, err := rt.Store.ListOpenReviewTasks(ctx)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: true, Code: "field.review.ok", Message: "Open review tasks ready.", Data: tasks})
		return exitSuccess
	default:
		return usageError(stderr, global.JSON, fmt.Errorf("unknown field subcommand: %s", args[0]))
	}
}

func runCloseout(ctx context.Context, global globalFlags, args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, global.JSON, errors.New("closeout subcommand is required"))
	}

	rt, err := runtimeFor(ctx, global, true)
	if err != nil {
		return runtimeError(stderr, global.JSON, err)
	}
	defer rt.Store.Close()

	switch args[0] {
	case "note":
		fs := newFlagSet("closeout note")
		jobID := fs.String("job", "", "Job ID.")
		serviceSummary := fs.String("service-summary", "", "Service summary.")
		inspectionNotes := fs.String("inspection-notes", "", "Inspection notes.")
		treatmentNotes := fs.String("treatment-notes", "", "Treatment notes.")
		paperworkRequired := fs.String("paperwork-required", "", "Optional boolean.")
		paperworkReady := fs.String("paperwork-ready", "", "Optional boolean.")
		requiresPhotos := fs.String("requires-photos", "", "Optional boolean.")
		requiresPrep := fs.String("requires-prep", "", "Optional boolean.")
		requiresFollowUp := fs.String("requires-follow-up", "", "Optional boolean.")
		followUpPlan := fs.String("follow-up-plan", "", "Follow-up instructions.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *jobID == "" {
			return usageError(stderr, global.JSON, errors.New("--job is required"))
		}

		paperworkRequiredPtr, err := parseOptionalBool(*paperworkRequired)
		if err != nil {
			return usageError(stderr, global.JSON, err)
		}
		paperworkReadyPtr, err := parseOptionalBool(*paperworkReady)
		if err != nil {
			return usageError(stderr, global.JSON, err)
		}
		requiresPhotosPtr, err := parseOptionalBool(*requiresPhotos)
		if err != nil {
			return usageError(stderr, global.JSON, err)
		}
		requiresPrepPtr, err := parseOptionalBool(*requiresPrep)
		if err != nil {
			return usageError(stderr, global.JSON, err)
		}
		requiresFollowUpPtr, err := parseOptionalBool(*requiresFollowUp)
		if err != nil {
			return usageError(stderr, global.JSON, err)
		}

		job, err := rt.Store.UpdateCloseoutNotes(ctx, store.CloseoutNoteInput{
			JobID:             *jobID,
			ServiceSummary:    *serviceSummary,
			InspectionNotes:   *inspectionNotes,
			TreatmentNotes:    *treatmentNotes,
			PaperworkRequired: paperworkRequiredPtr,
			PaperworkReady:    paperworkReadyPtr,
			RequiresPhotos:    requiresPhotosPtr,
			RequiresPrep:      requiresPrepPtr,
			RequiresFollowUp:  requiresFollowUpPtr,
			FollowUpPlan:      *followUpPlan,
		})
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: true, Code: "closeout.note.ok", Message: "Closeout notes updated.", Data: job})
		return exitSuccess
	case "photo":
		fs := newFlagSet("closeout photo")
		jobID := fs.String("job", "", "Job ID.")
		filePath := fs.String("file", "", "Path to the photo file.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *jobID == "" || *filePath == "" {
			return usageError(stderr, global.JSON, errors.New("--job and --file are required"))
		}
		persisted, err := rt.PersistArtifact(*filePath, "photo", *jobID)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		artifact, err := rt.Store.RecordArtifact(ctx, "photo", "job", *jobID, persisted.RelativePath, persisted.OriginalName, persisted.SHA256)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		if err := rt.Store.AttachDocumentPhoto(ctx, *jobID, artifact); err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: true, Code: "closeout.photo.ok", Message: "Closeout photo attached.", Data: artifact})
		return exitSuccess
	case "prep":
		fs := newFlagSet("closeout prep")
		jobID := fs.String("job", "", "Job ID.")
		required := fs.Bool("required", true, "Whether prep is required.")
		completed := fs.Bool("completed", false, "Whether prep is completed.")
		followUpPlan := fs.String("follow-up-plan", "", "Follow-up plan.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *jobID == "" {
			return usageError(stderr, global.JSON, errors.New("--job is required"))
		}
		if err := rt.Store.SetPrepState(ctx, store.PrepInput{
			JobID:        *jobID,
			Required:     *required,
			Completed:    *completed,
			FollowUpPlan: *followUpPlan,
		}); err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: true, Code: "closeout.prep.ok", Message: "Prep state updated."})
		return exitSuccess
	case "evaluate":
		fs := newFlagSet("closeout evaluate")
		jobID := fs.String("job", "", "Job ID.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *jobID == "" {
			return usageError(stderr, global.JSON, errors.New("--job is required"))
		}
		result, err := rt.Store.EvaluateCloseout(ctx, *jobID)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		exitCode := exitSuccess
		ok := true
		if result.CloseoutState != "complete" {
			exitCode = exitReview
			ok = false
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: ok, Code: "closeout.evaluate.ok", Message: "Closeout evaluation complete.", Data: result})
		return exitCode
	default:
		return usageError(stderr, global.JSON, fmt.Errorf("unknown closeout subcommand: %s", args[0]))
	}
}

func runBilling(ctx context.Context, global globalFlags, args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, global.JSON, errors.New("billing subcommand is required"))
	}

	rt, err := runtimeFor(ctx, global, true)
	if err != nil {
		return runtimeError(stderr, global.JSON, err)
	}
	defer rt.Store.Close()

	switch args[0] {
	case "draft":
		fs := newFlagSet("billing draft")
		jobID := fs.String("job", "", "Job ID.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *jobID == "" {
			return usageError(stderr, global.JSON, errors.New("--job is required"))
		}
		result, err := rt.Store.DraftInvoice(ctx, *jobID)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		exitCode := exitSuccess
		ok := true
		if result.State == "hold" {
			ok = false
			exitCode = exitReview
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: ok, Code: "billing.draft.ok", Message: "Billing handoff evaluated.", Data: result})
		return exitCode
	case "list-holds":
		fs := newFlagSet("billing list-holds")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		holds, err := rt.Store.ListBillingHolds(ctx)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: true, Code: "billing.holds.ok", Message: "Billing holds ready.", Data: holds})
		return exitSuccess
	default:
		return usageError(stderr, global.JSON, fmt.Errorf("unknown billing subcommand: %s", args[0]))
	}
}

func runReport(ctx context.Context, global globalFlags, args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, global.JSON, errors.New("report subcommand is required"))
	}

	rt, err := runtimeFor(ctx, global, true)
	if err != nil {
		return runtimeError(stderr, global.JSON, err)
	}
	defer rt.Store.Close()

	switch args[0] {
	case "morning":
		fs := newFlagSet("report morning")
		routeDate := fs.String("date", "", "Route date in YYYY-MM-DD format.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		report, err := rt.Store.MorningReport(ctx, *routeDate)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: true, Code: "report.morning.ok", Message: "Morning brief ready.", Data: report})
		return exitSuccess
	case "route-risk":
		fs := newFlagSet("report route-risk")
		routeDate := fs.String("date", "", "Route date in YYYY-MM-DD format.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *routeDate == "" {
			return usageError(stderr, global.JSON, errors.New("--date is required"))
		}
		report, err := rt.Store.RouteRiskReport(ctx, *routeDate)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: true, Code: "report.route_risk.ok", Message: "Route risk report ready.", Data: report})
		return exitSuccess
	case "blocked-closeouts":
		fs := newFlagSet("report blocked-closeouts")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		report, err := rt.Store.BlockedCloseoutReport(ctx)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: true, Code: "report.blocked_closeouts.ok", Message: "Blocked closeouts ready.", Data: report})
		return exitSuccess
	case "callback-pressure":
		fs := newFlagSet("report callback-pressure")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		report, err := rt.Store.CallbackPressureReport(ctx)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: true, Code: "report.callback_pressure.ok", Message: "Callback pressure ready.", Data: report})
		return exitSuccess
	case "end-day":
		fs := newFlagSet("report end-day")
		routeDate := fs.String("date", "", "Route date in YYYY-MM-DD format.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *routeDate == "" {
			return usageError(stderr, global.JSON, errors.New("--date is required"))
		}
		report, err := rt.Store.EndDayReport(ctx, *routeDate)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{OK: true, Code: "report.end_day.ok", Message: "End-of-day wrap ready.", Data: report})
		return exitSuccess
	default:
		return usageError(stderr, global.JSON, fmt.Errorf("unknown report subcommand: %s", args[0]))
	}
}

func runBackup(ctx context.Context, global globalFlags, args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		return usageError(stderr, global.JSON, errors.New("backup subcommand is required"))
	}

	rt, err := runtimeFor(ctx, global, true)
	if err != nil {
		return runtimeError(stderr, global.JSON, err)
	}
	defer rt.Store.Close()

	switch args[0] {
	case "create":
		fs := newFlagSet("backup create")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		path, manifest, err := rt.CreateBackup(ctx)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{
			OK: true, Code: "backup.create.ok", Message: "Backup created.",
			Data: map[string]any{"path": path, "manifest": manifest},
		})
		return exitSuccess
	case "verify":
		fs := newFlagSet("backup verify")
		path := fs.String("file", "", "Path to the backup archive.")
		if err := fs.Parse(args[1:]); err != nil {
			return usageError(stderr, global.JSON, err)
		}
		if *path == "" {
			return usageError(stderr, global.JSON, errors.New("--file is required"))
		}
		manifest, files, err := app.VerifyBackup(*path)
		if err != nil {
			return runtimeError(stderr, global.JSON, err)
		}
		_ = writeResponse(stdout, global.JSON, Response{
			OK: true, Code: "backup.verify.ok", Message: "Backup verified.",
			Data: map[string]any{"manifest": manifest, "files": files},
		})
		return exitSuccess
	default:
		return usageError(stderr, global.JSON, fmt.Errorf("unknown backup subcommand: %s", args[0]))
	}
}
