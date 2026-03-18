package store

type RouteDay struct {
	ID        string `json:"id"`
	RouteDate string `json:"route_date"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Job struct {
	ID                string `json:"id"`
	CustomerName      string `json:"customer_name"`
	PropertyAddress   string `json:"property_address"`
	JobType           string `json:"job_type"`
	PestTarget        string `json:"pest_target"`
	Priority          string `json:"priority"`
	ScheduledWindow   string `json:"scheduled_window"`
	Status            string `json:"status"`
	CloseoutState     string `json:"closeout_state"`
	BillingHoldReason string `json:"billing_hold_reason"`
	ServiceSummary    string `json:"service_summary"`
	InspectionNotes   string `json:"inspection_notes"`
	TreatmentNotes    string `json:"treatment_notes"`
	PaperworkRequired bool   `json:"paperwork_required"`
	PaperworkReady    bool   `json:"paperwork_ready"`
	RequiresPhotos    bool   `json:"requires_photos"`
	RequiresPrep      bool   `json:"requires_prep"`
	RequiresFollowUp  bool   `json:"requires_follow_up"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

type RouteStop struct {
	ID           string `json:"id"`
	RouteDayID   string `json:"route_day_id"`
	JobID        string `json:"job_id"`
	Position     int    `json:"position"`
	Status       string `json:"status"`
	IsUrgent     bool   `json:"is_urgent"`
	Blocker      string `json:"blocker"`
	Unresolved   string `json:"unresolved_reason"`
	CustomerName string `json:"customer_name"`
	JobType      string `json:"job_type"`
	PestTarget   string `json:"pest_target"`
	Priority     string `json:"priority"`
}

type Artifact struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	LinkedEntityType string `json:"linked_entity_type"`
	LinkedEntityID   string `json:"linked_entity_id"`
	RelativePath     string `json:"relative_path"`
	OriginalName     string `json:"original_name"`
	SHA256           string `json:"sha256"`
}

type DocumentPacket struct {
	ID                   string   `json:"id"`
	JobID                string   `json:"job_id"`
	Notes                string   `json:"notes"`
	MediaIDs             []string `json:"media_ids"`
	FormIDs              []string `json:"form_ids"`
	FollowUpInstructions string   `json:"follow_up_instructions"`
	CompletenessStatus   string   `json:"completeness_status"`
	MissingItems         []string `json:"missing_items"`
	PrepRequired         bool     `json:"prep_required"`
	PrepComplete         bool     `json:"prep_complete"`
	FollowUpPlan         string   `json:"follow_up_plan"`
	UpdatedAt            string   `json:"updated_at"`
}

type EventRecord struct {
	ID         string         `json:"id"`
	EventName  string         `json:"event_name"`
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	OccurredAt string         `json:"occurred_at"`
	Payload    map[string]any `json:"payload"`
}

type Task struct {
	ID               string `json:"id"`
	Owner            string `json:"owner"`
	Reason           string `json:"reason"`
	ResolutionState  string `json:"resolution_state"`
	LinkedEntityType string `json:"linked_entity_type"`
	LinkedEntityID   string `json:"linked_entity_id"`
	Details          string `json:"details"`
	DueDate          string `json:"due_date"`
}

type CallbackSummary struct {
	Reason  string `json:"reason"`
	Urgency string `json:"urgency"`
	Count   int    `json:"count"`
}

type ReportSummary struct {
	RouteDate          string            `json:"route_date,omitempty"`
	StopCounts         map[string]int    `json:"stop_counts,omitempty"`
	BlockedCloseouts   int               `json:"blocked_closeouts,omitempty"`
	FollowUpNeeded     int               `json:"follow_up_needed,omitempty"`
	InvoiceReady       int               `json:"invoice_ready,omitempty"`
	InvoiceHolds       int               `json:"invoice_holds,omitempty"`
	OverdueCallbacks   int               `json:"overdue_callbacks,omitempty"`
	UnresolvedStops    []RouteStop       `json:"unresolved_stops,omitempty"`
	AtRiskStops        []RouteStop       `json:"at_risk_stops,omitempty"`
	BlockedJobs        []Job             `json:"blocked_jobs,omitempty"`
	CallbackBreakdown  []CallbackSummary `json:"callback_breakdown,omitempty"`
	ReviewTaskCount    int               `json:"review_task_count,omitempty"`
	MissingRecordsJobs []Job             `json:"missing_records_jobs,omitempty"`
}

type CreateRouteDayInput struct {
	RouteDate string
}

type AddStopInput struct {
	RouteDayID        string
	RouteDate         string
	JobID             string
	CustomerName      string
	PropertyAddress   string
	JobType           string
	PestTarget        string
	Priority          string
	ScheduledWindow   string
	CallbackOfJobID   string
	PaperworkRequired bool
	RequiresPhotos    bool
	RequiresPrep      bool
	RequiresFollowUp  bool
}

type UpdateStopInput struct {
	RouteStopID      string
	Status           string
	Blocker          string
	UnresolvedReason string
}

type FieldCheckInInput struct {
	SourceChannel string
	RouteDayID    string
	JobID         string
	ShortStatus   string
	RawText       string
	Blockers      []string
	Confidence    float64
}

type VoiceMemoInput struct {
	SourceChannel   string
	JobID           string
	AudioSourcePath string
	Transcript      string
	Summary         string
	PestFacts       []string
	TreatmentFacts  []string
	FollowUpNeeds   []string
	BillingChanges  []string
	ContentIdeas    []string
	Confidence      float64
}

type PhotoInput struct {
	JobID            string
	SourcePath       string
	LinkedEntityType string
	LinkedEntityID   string
}

type CloseoutNoteInput struct {
	JobID             string
	ServiceSummary    string
	InspectionNotes   string
	TreatmentNotes    string
	PaperworkRequired *bool
	PaperworkReady    *bool
	RequiresPhotos    *bool
	RequiresPrep      *bool
	RequiresFollowUp  *bool
	FollowUpPlan      string
}

type PrepInput struct {
	JobID        string
	Required     bool
	Completed    bool
	FollowUpPlan string
}

type CloseoutResult struct {
	JobID            string   `json:"job_id"`
	CloseoutState    string   `json:"closeout_state"`
	MissingItems     []string `json:"missing_items"`
	FollowUpRequired []string `json:"follow_up_required"`
	CallbackOpen     bool     `json:"callback_open"`
}

type CloseoutStatus struct {
	Job            Job            `json:"job"`
	DocumentPacket DocumentPacket `json:"document_packet"`
}

type InvoiceDraftResult struct {
	InvoiceDraftID    string   `json:"invoice_draft_id"`
	JobID             string   `json:"job_id"`
	State             string   `json:"state"`
	BillingHoldReason string   `json:"billing_hold_reason"`
	Reasons           []string `json:"reasons"`
}
