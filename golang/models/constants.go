package models

// ─────────────────────────────────────────────────────────
// Shared constants – single source of truth for names that
// must match between worker registration and workflow dispatch.
// ─────────────────────────────────────────────────────────

const (
	// DefaultTaskQueue is the task queue both the worker and
	// the workflow activities use. Must match the Java client's
	// taskQueue field in StartWorkflowRequest.
	DefaultTaskQueue = "default-task-queue"

	// Workflow type name – must match Java @WorkflowInterface name
	WorkflowTypeName = "GenericWorkflow"

	// Activity type names – used for BOTH registration and dispatch
	ActivityReviewRequest = "ReviewRequestAgent"
	ActivityAddMeaning    = "AddMeaningAgent"
	ActivityLogActivity   = "LogActivityAgent"
	ActivityAIAnswer      = "AIAnswerAgent"
)
