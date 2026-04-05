package workflow_test

import (
	"encoding/json"
	"testing"

	"github.com/workflow/temporal-worker-go/activities"
	"github.com/workflow/temporal-worker-go/models"
	wf "github.com/workflow/temporal-worker-go/workflow"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

// registerAllActivities registers each activity function with the same
// string name the workflow uses in ExecuteActivity calls.
// This mirrors cmd/worker/main.go registration exactly.
func registerAllActivities(env *testsuite.TestWorkflowEnvironment) {
	env.RegisterActivityWithOptions(activities.ReviewRequestAgent, activity.RegisterOptions{
		Name: models.ActivityReviewRequest,
	})
	env.RegisterActivityWithOptions(activities.AddMeaningAgent, activity.RegisterOptions{
		Name: models.ActivityAddMeaning,
	})
	env.RegisterActivityWithOptions(activities.LogActivityAgent, activity.RegisterOptions{
		Name: models.ActivityLogActivity,
	})
	env.RegisterActivityWithOptions(activities.AIAnswerAgent, activity.RegisterOptions{
		Name: models.ActivityAIAnswer,
	})
}

func TestGenericWorkflow_FullPipeline(t *testing.T) {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()
	registerAllActivities(env)

	// Input matching what the Java client sends
	input := models.WorkflowData{
		Name:   "integration-test",
		Value:  99.95,
		Status: models.StatusReady,
		Text:   "Full pipeline test through all 4 agents",
	}

	env.ExecuteWorkflow(wf.GenericWorkflow, input)

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow returned error: %v", err)
	}

	var resultJSON string
	if err := env.GetWorkflowResult(&resultJSON); err != nil {
		t.Fatalf("failed to get result: %v", err)
	}

	var result models.WorkflowResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		t.Fatalf("failed to unmarshal result JSON: %v\nraw: %s", err, resultJSON)
	}

	// Verify all 4 activities ran
	if result.Answer == "" {
		t.Error("expected non-empty answer from AIAnswerAgent (Step 4)")
	}
	if result.Confidence <= 0 {
		t.Error("expected positive confidence from AIAnswerAgent (Step 4)")
	}
	if result.Model == "" {
		t.Error("expected non-empty model from AIAnswerAgent (Step 4)")
	}
	if result.LogID == "" {
		t.Error("expected non-empty logId from LogActivityAgent (Step 3)")
	}
	if result.Status != models.StatusGo {
		t.Errorf("expected final status GO, got %s", result.Status)
	}
	if result.Data.Name != "integration-test" {
		t.Errorf("expected data.name=integration-test, got %s", result.Data.Name)
	}
	if result.Data.Status != models.StatusGo {
		t.Errorf("expected data.status=GO from AIAnswerAgent, got %s", result.Data.Status)
	}

	t.Logf("✓ All 4 activities executed successfully")
	t.Logf("  Answer    : %.80s…", result.Answer)
	t.Logf("  Confidence: %.2f", result.Confidence)
	t.Logf("  Model     : %s", result.Model)
	t.Logf("  LogID     : %s", result.LogID)
	t.Logf("  Status    : %s", result.Status)
}

func TestGenericWorkflow_RejectedRequest(t *testing.T) {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()
	registerAllActivities(env)

	// Empty name triggers rejection in ReviewRequestAgent
	input := models.WorkflowData{
		Name:   "",
		Value:  10.0,
		Status: models.StatusReady,
		Text:   "Should be rejected at Step 1",
	}

	env.ExecuteWorkflow(wf.GenericWorkflow, input)

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow returned error: %v", err)
	}

	var resultJSON string
	if err := env.GetWorkflowResult(&resultJSON); err != nil {
		t.Fatalf("failed to get result: %v", err)
	}

	var result models.WorkflowResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Only Step 1 ran – verify early exit
	if result.Confidence != 0.0 {
		t.Errorf("expected confidence=0.0 for rejected request, got %f", result.Confidence)
	}
	if result.Status != models.StatusReady {
		t.Errorf("expected status READY for rejected request, got %s", result.Status)
	}
	if result.LogID != "" {
		t.Errorf("expected empty logId (Step 3 should not have run), got %s", result.LogID)
	}

	t.Logf("✓ Rejected correctly – only ReviewRequestAgent ran")
	t.Logf("  Answer: %s", result.Answer)
}

func TestGenericWorkflow_HighValuePositiveSentiment(t *testing.T) {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()
	registerAllActivities(env)

	// Value > 100 triggers positive sentiment + higher confidence
	input := models.WorkflowData{
		Name:   "premium-order",
		Value:  250.00,
		Status: models.StatusReady,
		Text:   "High-value order for premium path",
	}

	env.ExecuteWorkflow(wf.GenericWorkflow, input)

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow error: %v", err)
	}

	var resultJSON string
	if err := env.GetWorkflowResult(&resultJSON); err != nil {
		t.Fatalf("failed to get result: %v", err)
	}

	var result models.WorkflowResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if result.Confidence != 0.97 {
		t.Errorf("expected confidence=0.97 for positive sentiment, got %f", result.Confidence)
	}

	t.Logf("✓ High-value path: confidence=%.2f", result.Confidence)
}
