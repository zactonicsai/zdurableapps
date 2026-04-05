package workflow_test

import (
	"encoding/json"
	"testing"

	"github.com/workflow/temporal-worker-go/activities"
	"github.com/workflow/temporal-worker-go/models"
	wf "github.com/workflow/temporal-worker-go/workflow"
	"go.temporal.io/sdk/testsuite"
)

func TestGenericWorkflow_FullPipeline(t *testing.T) {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()

	// Register activities
	act := &activities.Activities{}
	env.RegisterActivity(act.ReviewRequestAgent)
	env.RegisterActivity(act.AddMeaningAgent)
	env.RegisterActivity(act.LogActivityAgent)
	env.RegisterActivity(act.AIAnswerAgent)

	// Input matching what the Java client sends
	input := models.WorkflowData{
		Name:   "integration-test",
		Value:  99.95,
		Status: models.StatusReady,
		Text:   "Full pipeline test",
	}

	// Execute
	env.ExecuteWorkflow(wf.GenericWorkflow, input)

	if !env.IsWorkflowCompleted() {
		t.Fatal("workflow did not complete")
	}
	if err := env.GetWorkflowError(); err != nil {
		t.Fatalf("workflow returned error: %v", err)
	}

	// Parse result
	var resultJSON string
	if err := env.GetWorkflowResult(&resultJSON); err != nil {
		t.Fatalf("failed to get result: %v", err)
	}

	var result models.WorkflowResult
	if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Verify pipeline output
	if result.Answer == "" {
		t.Error("expected non-empty answer")
	}
	if result.Confidence <= 0 {
		t.Error("expected positive confidence")
	}
	if result.Model == "" {
		t.Error("expected non-empty model")
	}
	if result.LogID == "" {
		t.Error("expected non-empty logId")
	}
	if result.Status != models.StatusGo {
		t.Errorf("expected final status GO, got %s", result.Status)
	}
	if result.Data.Name != "integration-test" {
		t.Errorf("expected data.name=integration-test, got %s", result.Data.Name)
	}

	t.Logf("Workflow result: %s", resultJSON)
}

func TestGenericWorkflow_RejectedRequest(t *testing.T) {
	suite := &testsuite.WorkflowTestSuite{}
	env := suite.NewTestWorkflowEnvironment()

	act := &activities.Activities{}
	env.RegisterActivity(act.ReviewRequestAgent)
	env.RegisterActivity(act.AddMeaningAgent)
	env.RegisterActivity(act.LogActivityAgent)
	env.RegisterActivity(act.AIAnswerAgent)

	// Empty name triggers rejection in ReviewRequestAgent
	input := models.WorkflowData{
		Name:   "",
		Value:  10.0,
		Status: models.StatusReady,
		Text:   "Should be rejected",
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
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if result.Confidence != 0.0 {
		t.Errorf("expected confidence=0.0 for rejected request, got %f", result.Confidence)
	}
	if result.Status != models.StatusReady {
		t.Errorf("expected status READY for rejected request, got %s", result.Status)
	}

	t.Logf("Rejected result: %s", resultJSON)
}
