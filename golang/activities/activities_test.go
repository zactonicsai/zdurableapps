package activities_test

import (
	"context"
	"strings"
	"testing"

	"github.com/workflow/temporal-worker-go/activities"
	"github.com/workflow/temporal-worker-go/models"
)

func sampleData() models.WorkflowData {
	return models.WorkflowData{
		Name:   "test-order",
		Value:  75.50,
		Status: models.StatusReady,
		Text:   "Process this test order",
	}
}

// ── ReviewRequestAgent ──────────────────────────────────

func TestReviewRequestAgent_Approved(t *testing.T) {
	act := &activities.Activities{}
	input := models.ReviewRequestInput{OriginalData: sampleData()}

	out, err := act.ReviewRequestAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Approved {
		t.Errorf("expected approved=true, got false; reason=%s", out.Reason)
	}
	if out.Reviewed.Status != models.StatusSet {
		t.Errorf("expected status SET, got %s", out.Reviewed.Status)
	}
}

func TestReviewRequestAgent_Rejected_EmptyName(t *testing.T) {
	act := &activities.Activities{}
	data := sampleData()
	data.Name = ""
	input := models.ReviewRequestInput{OriginalData: data}

	out, err := act.ReviewRequestAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Approved {
		t.Error("expected approved=false for empty name")
	}
}

func TestReviewRequestAgent_Rejected_NegativeValue(t *testing.T) {
	act := &activities.Activities{}
	data := sampleData()
	data.Value = -1.0
	input := models.ReviewRequestInput{OriginalData: data}

	out, err := act.ReviewRequestAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Approved {
		t.Error("expected approved=false for negative value")
	}
}

// ── AddMeaningAgent ─────────────────────────────────────

func TestAddMeaningAgent_StandardIntent(t *testing.T) {
	act := &activities.Activities{}
	input := models.AddMeaningInput{
		ReviewedData: sampleData(),
		Approved:     true,
	}

	out, err := act.AddMeaningAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Intent != "high_value_transaction" {
		t.Errorf("expected intent=high_value_transaction for value>50, got %s", out.Intent)
	}
	if out.Sentiment != "neutral" {
		t.Errorf("expected sentiment=neutral, got %s", out.Sentiment)
	}
	if len(out.Entities) == 0 {
		t.Error("expected at least one entity")
	}
	if _, ok := out.Context["intent"]; !ok {
		t.Error("expected context map to contain 'intent' key")
	}
}

func TestAddMeaningAgent_HighValuePositive(t *testing.T) {
	act := &activities.Activities{}
	data := sampleData()
	data.Value = 200.0
	input := models.AddMeaningInput{ReviewedData: data, Approved: true}

	out, err := act.AddMeaningAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Sentiment != "positive" {
		t.Errorf("expected sentiment=positive for value>100, got %s", out.Sentiment)
	}
}

// ── LogActivityAgent ────────────────────────────────────

func TestLogActivityAgent(t *testing.T) {
	act := &activities.Activities{}
	input := models.LogActivityInput{
		Stage:        "test-stage",
		WorkflowID:   "wf-123",
		Data:         sampleData(),
		NLPIntent:    "information_request",
		NLPSentiment: "neutral",
	}

	out, err := act.LogActivityAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.Logged {
		t.Error("expected logged=true")
	}
	if !strings.HasPrefix(out.LogID, "LOG-wf-123-") {
		t.Errorf("expected logId prefix LOG-wf-123-, got %s", out.LogID)
	}
	if out.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

// ── AIAnswerAgent ───────────────────────────────────────

func TestAIAnswerAgent(t *testing.T) {
	act := &activities.Activities{}
	input := models.AIAnswerInput{
		OriginalData: sampleData(),
		Intent:       "information_request",
		Entities:     []string{"test-order", "value:75.50"},
		Sentiment:    "neutral",
		Context:      map[string]string{"source": "test"},
	}

	out, err := act.AIAnswerAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Answer == "" {
		t.Error("expected non-empty answer")
	}
	if out.Confidence < 0 || out.Confidence > 1 {
		t.Errorf("confidence out of range: %f", out.Confidence)
	}
	if out.Model == "" {
		t.Error("expected non-empty model")
	}
	if out.ResponseData.Status != models.StatusGo {
		t.Errorf("expected response status GO, got %s", out.ResponseData.Status)
	}
}

func TestAIAnswerAgent_PositiveSentiment_HigherConfidence(t *testing.T) {
	act := &activities.Activities{}
	input := models.AIAnswerInput{
		OriginalData: sampleData(),
		Intent:       "high_value_transaction",
		Entities:     []string{"entity1"},
		Sentiment:    "positive",
		Context:      map[string]string{},
	}

	out, err := act.AIAnswerAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Confidence != 0.97 {
		t.Errorf("expected confidence=0.97 for positive sentiment, got %f", out.Confidence)
	}
}
