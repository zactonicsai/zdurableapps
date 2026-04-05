package main

import (
	"log"
	"os"

	"github.com/workflow/temporal-worker-go/activities"
	wf "github.com/workflow/temporal-worker-go/workflow"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultTaskQueue = "default-task-queue"
	defaultTarget    = "localhost:7233"
	defaultNamespace = "default"
)

func main() {
	target := envOr("TEMPORAL_TARGET", defaultTarget)
	namespace := envOr("TEMPORAL_NAMESPACE", defaultNamespace)
	taskQueue := envOr("TEMPORAL_TASK_QUEUE", defaultTaskQueue)

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

	// Register the workflow with the EXACT name the Java client uses.
	// Java derives the type name from the interface: "GenericWorkflow".
	w.RegisterWorkflowWithOptions(wf.GenericWorkflow, workflow.RegisterOptions{
		Name: "GenericWorkflow",
	})

	// Register all four activity methods.
	w.RegisterActivity(&activities.Activities{})

	// ── run ─────────────────────────────────────────────
	log.Println("Starting worker… (Ctrl-C to stop)")
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("Worker exited with error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
