package main

import (
	"log"
	"os"

	"github.com/workflow/temporal-worker-go/activities"
	"github.com/workflow/temporal-worker-go/models"
	wf "github.com/workflow/temporal-worker-go/workflow"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultTarget    = "localhost:7233"
	defaultNamespace = "default"
)

func main() {
	target := envOr("TEMPORAL_TARGET", defaultTarget)
	namespace := envOr("TEMPORAL_NAMESPACE", defaultNamespace)
	taskQueue := envOr("TEMPORAL_TASK_QUEUE", models.DefaultTaskQueue)

	log.Printf("┌──────────────────────────────────────────────┐")
	log.Printf("│  Temporal Go Worker                          │")
	log.Printf("├──────────────────────────────────────────────┤")
	log.Printf("│  Target    : %-30s │", target)
	log.Printf("│  Namespace : %-30s │", namespace)
	log.Printf("│  TaskQueue : %-30s │", taskQueue)
	log.Printf("└──────────────────────────────────────────────┘")

	// ── connect to Temporal ─────────────────────────────
	c, err := client.Dial(client.Options{
		HostPort:  target,
		Namespace: namespace,
	})
	if err != nil {
		log.Fatalf("Unable to connect to Temporal at %s: %v", target, err)
	}
	defer c.Close()

	// ── create worker ───────────────────────────────────
	w := worker.New(c, taskQueue, worker.Options{})

	// ── register workflow ───────────────────────────────
	// Name MUST match the Java @WorkflowInterface type name
	w.RegisterWorkflowWithOptions(wf.GenericWorkflow, workflow.RegisterOptions{
		Name: models.WorkflowTypeName,
	})
	log.Printf("  ✓ Registered workflow: %s", models.WorkflowTypeName)

	// ── register activities ─────────────────────────────
	// Name strings here MUST match the string names used in
	// workflow.ExecuteActivity(ctx, "ReviewRequestAgent", ...)
	// inside GenericWorkflow. Both pull from models.Activity* constants.

	registerActivity(w, activities.ReviewRequestAgent, models.ActivityReviewRequest)
	registerActivity(w, activities.AddMeaningAgent, models.ActivityAddMeaning)
	registerActivity(w, activities.LogActivityAgent, models.ActivityLogActivity)
	registerActivity(w, activities.AIAnswerAgent, models.ActivityAIAnswer)

	// ── run ─────────────────────────────────────────────
	log.Printf("Worker polling task queue: %s", taskQueue)
	log.Println("Starting worker… (Ctrl-C to stop)")
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("Worker exited with error: %v", err)
	}
}

// registerActivity registers a function with an explicit name and logs it.
func registerActivity(w worker.Worker, fn any, name string) {
	w.RegisterActivityWithOptions(fn, activity.RegisterOptions{
		Name: name,
	})
	log.Printf("  ✓ Registered activity: %s", name)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
