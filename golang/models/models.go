package models

// ─────────────────────────────────────────────────────────
// Status enum – mirrors Java: com.workflow.model.Status
// ─────────────────────────────────────────────────────────

type Status string

const (
	StatusReady Status = "READY"
	StatusSet   Status = "SET"
	StatusGo    Status = "GO"
)

// ─────────────────────────────────────────────────────────
// WorkflowData – mirrors Java: com.workflow.model.WorkflowData
// JSON tags match the Java field names exactly so Temporal
// serialisation is compatible across the two languages.
// ─────────────────────────────────────────────────────────

type WorkflowData struct {
	Name   string  `json:"name"`
	Value  float64 `json:"value"`
	Status Status  `json:"status"`
	Text   string  `json:"text"`
}

// ─────────────────────────────────────────────────────────
// Per-activity generic input / output types
// Each activity gets its own pair so they can evolve
// independently while keeping the pipeline typed.
// ─────────────────────────────────────────────────────────

// ── 1. Review Request Agent ─────────────────────────────

type ReviewRequestInput struct {
	OriginalData WorkflowData `json:"originalData"`
}

type ReviewRequestOutput struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason"`
	Reviewed WorkflowData `json:"reviewed"`
}

// ── 2. Add Meaning Agent (NLP + context) ────────────────

type AddMeaningInput struct {
	ReviewedData WorkflowData `json:"reviewedData"`
	Approved     bool         `json:"approved"`
}

type AddMeaningOutput struct {
	Intent     string            `json:"intent"`
	Entities   []string          `json:"entities"`
	Sentiment  string            `json:"sentiment"`
	Context    map[string]string `json:"context"`
	Enriched   WorkflowData      `json:"enriched"`
}

// ── 3. Log Activity Agent ───────────────────────────────

type LogActivityInput struct {
	Stage       string       `json:"stage"`
	WorkflowID  string       `json:"workflowId"`
	Data        WorkflowData `json:"data"`
	NLPIntent   string       `json:"nlpIntent"`
	NLPSentiment string     `json:"nlpSentiment"`
}

type LogActivityOutput struct {
	LogID     string `json:"logId"`
	Timestamp string `json:"timestamp"`
	Logged    bool   `json:"logged"`
}

// ── 4. AI Answer Agent ──────────────────────────────────

type AIAnswerInput struct {
	OriginalData WorkflowData      `json:"originalData"`
	Intent       string            `json:"intent"`
	Entities     []string          `json:"entities"`
	Sentiment    string            `json:"sentiment"`
	Context      map[string]string `json:"context"`
}

type AIAnswerOutput struct {
	Answer       string  `json:"answer"`
	Confidence   float64 `json:"confidence"`
	Model        string  `json:"model"`
	ResponseData WorkflowData `json:"responseData"`
}

// ─────────────────────────────────────────────────────────
// WorkflowResult – final output returned to the Java caller.
// The Java side receives this as the workflow return value
// (a String in GenericWorkflow.execute), so we serialise
// everything into a single JSON-friendly struct.
// ─────────────────────────────────────────────────────────

type WorkflowResult struct {
	Answer     string       `json:"answer"`
	Confidence float64      `json:"confidence"`
	Model      string       `json:"model"`
	LogID      string       `json:"logId"`
	Status     Status       `json:"status"`
	Data       WorkflowData `json:"data"`
}
