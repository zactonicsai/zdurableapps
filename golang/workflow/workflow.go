package workflow

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/workflow/temporal-worker-go/models"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GenericWorkflow is the Temporal workflow function.
//
// Registered as "GenericWorkflow" to match the Java @WorkflowInterface name.
// Chains four activities sequentially, passing enriched data between each step.
//
// CRITICAL: activities are dispatched by STRING name (not Go function reference)
// so the names match exactly what the worker registered via RegisterActivityWithOptions.
func GenericWorkflow(ctx workflow.Context, data models.WorkflowData) (string, error) {

	logger := workflow.GetLogger(ctx)
	wfInfo := workflow.GetInfo(ctx)

	logger.Info(">>> GenericWorkflow STARTED",
		"name", data.Name,
		"value", data.Value,
		"status", data.Status,
		"taskQueue", wfInfo.TaskQueueName,
		"workflowId", wfInfo.WorkflowExecution.ID,
	)

	// ── activity options (shared by all four steps) ─────
	// TaskQueue MUST be set explicitly so Temporal schedules
	// activity tasks on the same queue the worker polls.
	// Without this, cross-language dispatch can silently drop tasks.
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		TaskQueue:            wfInfo.TaskQueueName, // ← same queue as workflow
		StartToCloseTimeout:  30 * time.Second,
		ScheduleToStartTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	})

	// ═════════════════════════════════════════════════════
	// Step 1 – Review Request Agent
	// ═════════════════════════════════════════════════════

	logger.Info(">>> Step 1: Dispatching ReviewRequestAgent",
		"activityType", models.ActivityReviewRequest,
		"taskQueue", wfInfo.TaskQueueName,
	)

	reviewInput := models.ReviewRequestInput{
		OriginalData: data,
	}

	var reviewOutput models.ReviewRequestOutput
	err := workflow.ExecuteActivity(
		ctx,
		models.ActivityReviewRequest, // STRING name – matches worker registration
		reviewInput,
	).Get(ctx, &reviewOutput)
	if err != nil {
		return "", fmt.Errorf("Step 1 %s failed: %w", models.ActivityReviewRequest, err)
	}

	logger.Info(">>> Step 1 COMPLETE",
		"approved", reviewOutput.Approved,
		"reason", reviewOutput.Reason,
	)

	// Early exit if rejected
	if !reviewOutput.Approved {
		logger.Info(">>> Workflow exiting early – request rejected")
		result := models.WorkflowResult{
			Answer:     "Request rejected: " + reviewOutput.Reason,
			Confidence: 0.0,
			Model:      "n/a",
			Status:     models.StatusReady,
			Data:       data,
		}
		return toJSON(result), nil
	}

	// ═════════════════════════════════════════════════════
	// Step 2 – Add Meaning Agent (NLP + context)
	// ═════════════════════════════════════════════════════

	logger.Info(">>> Step 2: Dispatching AddMeaningAgent",
		"activityType", models.ActivityAddMeaning,
		"taskQueue", wfInfo.TaskQueueName,
	)

	meaningInput := models.AddMeaningInput{
		ReviewedData: reviewOutput.Reviewed,
		Approved:     reviewOutput.Approved,
	}

	var meaningOutput models.AddMeaningOutput
	err = workflow.ExecuteActivity(
		ctx,
		models.ActivityAddMeaning, // STRING name
		meaningInput,
	).Get(ctx, &meaningOutput)
	if err != nil {
		return "", fmt.Errorf("Step 2 %s failed: %w", models.ActivityAddMeaning, err)
	}

	logger.Info(">>> Step 2 COMPLETE",
		"intent", meaningOutput.Intent,
		"sentiment", meaningOutput.Sentiment,
		"entityCount", len(meaningOutput.Entities),
	)

	// ═════════════════════════════════════════════════════
	// Step 3 – Log Activity Agent
	// ═════════════════════════════════════════════════════

	logger.Info(">>> Step 3: Dispatching LogActivityAgent",
		"activityType", models.ActivityLogActivity,
		"taskQueue", wfInfo.TaskQueueName,
	)

	logInput := models.LogActivityInput{
		Stage:        "post-nlp-enrichment",
		WorkflowID:   wfInfo.WorkflowExecution.ID,
		Data:         meaningOutput.Enriched,
		NLPIntent:    meaningOutput.Intent,
		NLPSentiment: meaningOutput.Sentiment,
	}

	var logOutput models.LogActivityOutput
	err = workflow.ExecuteActivity(
		ctx,
		models.ActivityLogActivity, // STRING name
		logInput,
	).Get(ctx, &logOutput)
	if err != nil {
		return "", fmt.Errorf("Step 3 %s failed: %w", models.ActivityLogActivity, err)
	}

	logger.Info(">>> Step 3 COMPLETE",
		"logId", logOutput.LogID,
		"timestamp", logOutput.Timestamp,
		"logged", logOutput.Logged,
	)

	// ═════════════════════════════════════════════════════
	// Step 4 – AI Answer Agent
	// ═════════════════════════════════════════════════════

	logger.Info(">>> Step 4: Dispatching AIAnswerAgent",
		"activityType", models.ActivityAIAnswer,
		"taskQueue", wfInfo.TaskQueueName,
	)

	answerInput := models.AIAnswerInput{
		OriginalData: data,
		Intent:       meaningOutput.Intent,
		Entities:     meaningOutput.Entities,
		Sentiment:    meaningOutput.Sentiment,
		Context:      meaningOutput.Context,
	}

	var answerOutput models.AIAnswerOutput
	err = workflow.ExecuteActivity(
		ctx,
		models.ActivityAIAnswer, // STRING name
		answerInput,
	).Get(ctx, &answerOutput)
	if err != nil {
		return "", fmt.Errorf("Step 4 %s failed: %w", models.ActivityAIAnswer, err)
	}

	logger.Info(">>> Step 4 COMPLETE",
		"confidence", answerOutput.Confidence,
		"model", answerOutput.Model,
		"answerLen", len(answerOutput.Answer),
	)

	// ═════════════════════════════════════════════════════
	// Assemble final result
	// ═════════════════════════════════════════════════════

	result := models.WorkflowResult{
		Answer:     answerOutput.Answer,
		Confidence: answerOutput.Confidence,
		Model:      answerOutput.Model,
		LogID:      logOutput.LogID,
		Status:     models.StatusGo,
		Data:       answerOutput.ResponseData,
	}

	resultJSON := toJSON(result)

	logger.Info(">>> GenericWorkflow COMPLETED – all 4 activities executed",
		"finalStatus", result.Status,
		"resultLength", len(resultJSON),
	)

	return resultJSON, nil
}

// toJSON marshals v to an indented JSON string.
func toJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error":"%s"}`, err.Error())
	}
	return string(b)
}
