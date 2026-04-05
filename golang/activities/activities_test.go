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
	input := models.ReviewRequestInput{OriginalData: sampleData()}

	out, err := activities.ReviewRequestAgent(context.Background(), input)
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
	data := sampleData()
	data.Name = ""
	input := models.ReviewRequestInput{OriginalData: data}

	out, err := activities.ReviewRequestAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Approved {
		t.Error("expected approved=false for empty name")
	}
}

func TestReviewRequestAgent_Rejected_NegativeValue(t *testing.T) {
	data := sampleData()
	data.Value = -1.0
	input := models.ReviewRequestInput{OriginalData: data}

	out, err := activities.ReviewRequestAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Approved {
		t.Error("expected approved=false for negative value")
	}
}

// ── AddMeaningAgent ─────────────────────────────────────

func TestAddMeaningAgent_HighValueTransaction(t *testing.T) {
	input := models.AddMeaningInput{
		ReviewedData: sampleData(), // value=75.50 > 50
		Approved:     true,
	}

	out, err := activities.AddMeaningAgent(context.Background(), input)
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
	if !strings.Contains(out.Enriched.Text, "[NLP enriched]") {
		t.Error("expected enriched text to contain [NLP enriched] prefix")
	}
}

func TestAddMeaningAgent_PositiveSentiment(t *testing.T) {
	data := sampleData()
	data.Value = 200.0
	input := models.AddMeaningInput{ReviewedData: data, Approved: true}

	out, err := activities.AddMeaningAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Sentiment != "positive" {
		t.Errorf("expected sentiment=positive for value>100, got %s", out.Sentiment)
	}
}

func TestAddMeaningAgent_LowValueInformationRequest(t *testing.T) {
	data := sampleData()
	data.Value = 10.0
	input := models.AddMeaningInput{ReviewedData: data, Approved: true}

	out, err := activities.AddMeaningAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Intent != "information_request" {
		t.Errorf("expected intent=information_request for value<=50, got %s", out.Intent)
	}
}

// ── LogActivityAgent ────────────────────────────────────

func TestLogActivityAgent(t *testing.T) {
	input := models.LogActivityInput{
		Stage:        "test-stage",
		WorkflowID:   "wf-123",
		Data:         sampleData(),
		NLPIntent:    "information_request",
		NLPSentiment: "neutral",
	}

	out, err := activities.LogActivityAgent(context.Background(), input)
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
	input := models.AIAnswerInput{
		OriginalData: sampleData(),
		Intent:       "information_request",
		Entities:     []string{"test-order", "value:75.50"},
		Sentiment:    "neutral",
		Context:      map[string]string{"source": "test"},
	}

	out, err := activities.AIAnswerAgent(context.Background(), input)
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
	if !strings.Contains(out.Answer, "test-order") {
		t.Error("expected answer to reference the original name")
	}
}

func TestAIAnswerAgent_PositiveSentiment_HigherConfidence(t *testing.T) {
	input := models.AIAnswerInput{
		OriginalData: sampleData(),
		Intent:       "high_value_transaction",
		Entities:     []string{"entity1"},
		Sentiment:    "positive",
		Context:      map[string]string{},
	}

	out, err := activities.AIAnswerAgent(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Confidence != 0.97 {
		t.Errorf("expected confidence=0.97 for positive sentiment, got %f", out.Confidence)
	}
}

// ── Full chain sanity (no Temporal, just functions) ─────

func TestFullChain_AllFourActivities(t *testing.T) {
	ctx := context.Background()
	data := sampleData()

	// Step 1
	reviewOut, err := activities.ReviewRequestAgent(ctx, models.ReviewRequestInput{OriginalData: data})
	if err != nil {
		t.Fatalf("Step 1 failed: %v", err)
	}
	if !reviewOut.Approved {
		t.Fatalf("Step 1 rejected unexpectedly: %s", reviewOut.Reason)
	}
	t.Logf("Step 1 OK: approved=%t status=%s", reviewOut.Approved, reviewOut.Reviewed.Status)

	// Step 2
	meaningOut, err := activities.AddMeaningAgent(ctx, models.AddMeaningInput{
		ReviewedData: reviewOut.Reviewed,
		Approved:     reviewOut.Approved,
	})
	if err != nil {
		t.Fatalf("Step 2 failed: %v", err)
	}
	t.Logf("Step 2 OK: intent=%s sentiment=%s entities=%v", meaningOut.Intent, meaningOut.Sentiment, meaningOut.Entities)

	// Step 3
	logOut, err := activities.LogActivityAgent(ctx, models.LogActivityInput{
		Stage:        "test-chain",
		WorkflowID:   "chain-test-001",
		Data:         meaningOut.Enriched,
		NLPIntent:    meaningOut.Intent,
		NLPSentiment: meaningOut.Sentiment,
	})
	if err != nil {
		t.Fatalf("Step 3 failed: %v", err)
	}
	t.Logf("Step 3 OK: logId=%s", logOut.LogID)

	// Step 4
	answerOut, err := activities.AIAnswerAgent(ctx, models.AIAnswerInput{
		OriginalData: data,
		Intent:       meaningOut.Intent,
		Entities:     meaningOut.Entities,
		Sentiment:    meaningOut.Sentiment,
		Context:      meaningOut.Context,
	})
	if err != nil {
		t.Fatalf("Step 4 failed: %v", err)
	}
	t.Logf("Step 4 OK: confidence=%.2f model=%s", answerOut.Confidence, answerOut.Model)
	t.Logf("Final answer: %s", answerOut.Answer)

	// Verify the full chain produced the expected final state
	if answerOut.ResponseData.Status != models.StatusGo {
		t.Errorf("expected final status GO, got %s", answerOut.ResponseData.Status)
	}
}
