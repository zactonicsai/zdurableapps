package workflow

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/workflow/temporal-worker-go/activities"
	"github.com/workflow/temporal-worker-go/models"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GenericWorkflow is the Temporal workflow function.
// Its name and signature match the Java @WorkflowInterface:
//
//	@WorkflowMethod
//	String execute(WorkflowData data);
//
// Temporal routes by workflow type name "GenericWorkflow" which is
// registered in the worker with workflow.RegisterOptions{Name: "GenericWorkflow"}.
func GenericWorkflow(ctx workflow.Context, data models.WorkflowData) (string, error) {

	logger := workflow.GetLogger(ctx)
	logger.Info("GenericWorkflow started",
		"name", data.Name,
		"value", data.Value,
		"status", data.Status,
	)

	// ── activity options (shared by all four steps) ─────
	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, actOpts)

	var act *activities.Activities // nil receiver; Temporal resolves methods by name

	// ─────────────────────────────────────────────────────
	// Step 1 – Review Request Agent
	// ─────────────────────────────────────────────────────

	reviewInput := models.ReviewRequestInput{
		OriginalData: data,
	}

	var reviewOutput models.ReviewRequestOutput
	if err := workflow.ExecuteActivity(ctx, act.ReviewRequestAgent, reviewInput).
		Get(ctx, &reviewOutput); err != nil {
		return "", fmt.Errorf("ReviewRequestAgent failed: %w", err)
	}

	logger.Info("ReviewRequestAgent completed",
		"approved", reviewOutput.Approved,
		"reason", reviewOutput.Reason,
	)

	if !reviewOutput.Approved {
		result := models.WorkflowResult{
			Answer:     "Request rejected: " + reviewOutput.Reason,
			Confidence: 0.0,
			Model:      "n/a",
			Status:     models.StatusReady,
			Data:       data,
		}
		return toJSON(result), nil
	}

	// ─────────────────────────────────────────────────────
	// Step 2 – Add Meaning Agent (NLP + context)
	// ─────────────────────────────────────────────────────

	meaningInput := models.AddMeaningInput{
		ReviewedData: reviewOutput.Reviewed,
		Approved:     reviewOutput.Approved,
	}

	var meaningOutput models.AddMeaningOutput
	if err := workflow.ExecuteActivity(ctx, act.AddMeaningAgent, meaningInput).
		Get(ctx, &meaningOutput); err != nil {
		return "", fmt.Errorf("AddMeaningAgent failed: %w", err)
	}

	logger.Info("AddMeaningAgent completed",
		"intent", meaningOutput.Intent,
		"sentiment", meaningOutput.Sentiment,
		"entities", meaningOutput.Entities,
	)

	// ─────────────────────────────────────────────────────
	// Step 3 – Log Activity Agent
	// ─────────────────────────────────────────────────────

	info := workflow.GetInfo(ctx)
	logInput := models.LogActivityInput{
		Stage:        "post-nlp-enrichment",
		WorkflowID:   info.WorkflowExecution.ID,
		Data:         meaningOutput.Enriched,
		NLPIntent:    meaningOutput.Intent,
		NLPSentiment: meaningOutput.Sentiment,
	}

	var logOutput models.LogActivityOutput
	if err := workflow.ExecuteActivity(ctx, act.LogActivityAgent, logInput).
		Get(ctx, &logOutput); err != nil {
		return "", fmt.Errorf("LogActivityAgent failed: %w", err)
	}

	logger.Info("LogActivityAgent completed",
		"logId", logOutput.LogID,
		"timestamp", logOutput.Timestamp,
	)

	// ─────────────────────────────────────────────────────
	// Step 4 – AI Answer Agent
	// ─────────────────────────────────────────────────────

	answerInput := models.AIAnswerInput{
		OriginalData: data,
		Intent:       meaningOutput.Intent,
		Entities:     meaningOutput.Entities,
		Sentiment:    meaningOutput.Sentiment,
		Context:      meaningOutput.Context,
	}

	var answerOutput models.AIAnswerOutput
	if err := workflow.ExecuteActivity(ctx, act.AIAnswerAgent, answerInput).
		Get(ctx, &answerOutput); err != nil {
		return "", fmt.Errorf("AIAnswerAgent failed: %w", err)
	}

	logger.Info("AIAnswerAgent completed",
		"confidence", answerOutput.Confidence,
		"model", answerOutput.Model,
	)

	// ─────────────────────────────────────────────────────
	// Assemble final result
	// ─────────────────────────────────────────────────────

	result := models.WorkflowResult{
		Answer:     answerOutput.Answer,
		Confidence: answerOutput.Confidence,
		Model:      answerOutput.Model,
		LogID:      logOutput.LogID,
		Status:     models.StatusGo,
		Data:       answerOutput.ResponseData,
	}

	resultJSON := toJSON(result)

	logger.Info("GenericWorkflow completed", "resultLength", len(resultJSON))
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
