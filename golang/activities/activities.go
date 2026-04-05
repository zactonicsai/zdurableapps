package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/workflow/temporal-worker-go/models"
	"go.temporal.io/sdk/activity"
)

// Activities groups all activity methods so they can be registered
// on the worker in one call.
type Activities struct{}

// ─────────────────────────────────────────────────────────
// 1. ReviewRequestAgent
//    Validates the incoming request and decides whether to
//    approve it for downstream processing.
// ─────────────────────────────────────────────────────────

func (a *Activities) ReviewRequestAgent(
	ctx context.Context,
	input models.ReviewRequestInput,
) (*models.ReviewRequestOutput, error) {

	logger := activity.GetLogger(ctx)
	logger.Info("ReviewRequestAgent: evaluating request",
		"name", input.OriginalData.Name,
		"status", input.OriginalData.Status,
	)

	// ── stubbed logic ───────────────────────────────────
	approved := input.OriginalData.Name != "" &&
		input.OriginalData.Value >= 0

	reason := "Request meets all validation criteria"
	if !approved {
		reason = "Request rejected: missing name or negative value"
	}

	reviewed := input.OriginalData
	reviewed.Status = models.StatusSet // advance status

	return &models.ReviewRequestOutput{
		Approved: approved,
		Reason:   reason,
		Reviewed: reviewed,
	}, nil
}

// ─────────────────────────────────────────────────────────
// 2. AddMeaningAgent
//    NLP enrichment – extracts intent, entities, sentiment,
//    and builds a context map for downstream agents.
// ─────────────────────────────────────────────────────────

func (a *Activities) AddMeaningAgent(
	ctx context.Context,
	input models.AddMeaningInput,
) (*models.AddMeaningOutput, error) {

	logger := activity.GetLogger(ctx)
	logger.Info("AddMeaningAgent: enriching with NLP",
		"name", input.ReviewedData.Name,
		"textLen", len(input.ReviewedData.Text),
	)

	// ── stubbed NLP results ─────────────────────────────
	intent := "information_request"
	if input.ReviewedData.Value > 50 {
		intent = "high_value_transaction"
	}

	entities := []string{
		input.ReviewedData.Name,
		fmt.Sprintf("value:%.2f", input.ReviewedData.Value),
	}

	sentiment := "neutral"
	if input.ReviewedData.Value > 100 {
		sentiment = "positive"
	}

	contextMap := map[string]string{
		"source":        "temporal-worker-go",
		"intent":        intent,
		"original_text": truncate(input.ReviewedData.Text, 256),
		"processed_at":  time.Now().UTC().Format(time.RFC3339),
		"approved":      fmt.Sprintf("%t", input.Approved),
	}

	enriched := input.ReviewedData
	enriched.Text = fmt.Sprintf("[NLP enriched] %s | intent=%s sentiment=%s",
		enriched.Text, intent, sentiment)

	return &models.AddMeaningOutput{
		Intent:    intent,
		Entities:  entities,
		Sentiment: sentiment,
		Context:   contextMap,
		Enriched:  enriched,
	}, nil
}

// ─────────────────────────────────────────────────────────
// 3. LogActivityAgent
//    Persists an audit log entry for the current pipeline
//    stage (stubbed – returns a static log ID).
// ─────────────────────────────────────────────────────────

func (a *Activities) LogActivityAgent(
	ctx context.Context,
	input models.LogActivityInput,
) (*models.LogActivityOutput, error) {

	logger := activity.GetLogger(ctx)
	logger.Info("LogActivityAgent: recording audit entry",
		"stage", input.Stage,
		"workflowId", input.WorkflowID,
		"intent", input.NLPIntent,
	)

	// ── stubbed persistence ─────────────────────────────
	now := time.Now().UTC()
	logID := fmt.Sprintf("LOG-%s-%d", input.WorkflowID, now.UnixMilli())

	return &models.LogActivityOutput{
		LogID:     logID,
		Timestamp: now.Format(time.RFC3339),
		Logged:    true,
	}, nil
}

// ─────────────────────────────────────────────────────────
// 4. AIAnswerAgent
//    Takes the enriched context and produces an AI-generated
//    answer (stubbed with static response data).
// ─────────────────────────────────────────────────────────

func (a *Activities) AIAnswerAgent(
	ctx context.Context,
	input models.AIAnswerInput,
) (*models.AIAnswerOutput, error) {

	logger := activity.GetLogger(ctx)
	logger.Info("AIAnswerAgent: generating answer",
		"intent", input.Intent,
		"sentiment", input.Sentiment,
		"entities", input.Entities,
	)

	// ── stubbed AI inference ────────────────────────────
	answer := fmt.Sprintf(
		"Based on the %s intent with %s sentiment for '%s' (value=%.2f): "+
			"The request has been processed successfully. "+
			"Analysis indicates %d entities detected across the enriched context. "+
			"Recommendation: proceed with standard workflow execution.",
		input.Intent,
		input.Sentiment,
		input.OriginalData.Name,
		input.OriginalData.Value,
		len(input.Entities),
	)

	confidence := 0.92
	if input.Sentiment == "positive" {
		confidence = 0.97
	}

	responseData := input.OriginalData
	responseData.Status = models.StatusGo
	responseData.Text = answer

	return &models.AIAnswerOutput{
		Answer:       answer,
		Confidence:   confidence,
		Model:        "stubbed-gpt-4o-agent-v1",
		ResponseData: responseData,
	}, nil
}

// ─────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
