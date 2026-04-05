# Temporal Go Worker

A Go-based Temporal worker with four stubbed agent activities, designed to pair with the [Java Spring Boot workflow client](../temporal-workflow-lib).

---

## Activity Pipeline

```
WorkflowData (from Java client)
       │
       ▼
┌─────────────────────┐
│  ReviewRequestAgent  │  Validate & approve/reject
└─────────┬───────────┘
          │ approved?
          ▼
┌─────────────────────┐
│  AddMeaningAgent     │  NLP: intent, entities, sentiment, context
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  LogActivityAgent    │  Audit log with workflow + NLP metadata
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│  AIAnswerAgent       │  Generate answer from enriched context
└─────────┬───────────┘
          │
          ▼
    WorkflowResult (JSON string → Java client)
```

## Data Model Compatibility

| Go struct       | Java class                      | Role                  |
|-----------------|---------------------------------|-----------------------|
| `WorkflowData`  | `com.workflow.model.WorkflowData` | Shared input payload |
| `Status`        | `com.workflow.model.Status`       | READY / SET / GO     |
| `WorkflowResult`| *(returned as JSON string)*       | Final output         |

The workflow is registered as `"GenericWorkflow"` which matches the Java `@WorkflowInterface` type name exactly.

---

## Quick Start

### 1. Prerequisites

- Go 1.22+
- A running Temporal server (`temporal server start-dev`)

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Run the worker

```bash
go run ./cmd/worker
```

With custom config:

```bash
TEMPORAL_TARGET=localhost:7233 \
TEMPORAL_NAMESPACE=default \
TEMPORAL_TASK_QUEUE=default-task-queue \
go run ./cmd/worker
```

### 4. Trigger from the Java client

```bash
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "taskQueue": "default-task-queue",
    "data": {
      "name": "demo-order",
      "value": 99.95,
      "status": "READY",
      "text": "Process this order through all agents"
    }
  }'
```

### 5. Run tests

```bash
# All tests (no Temporal server needed)
go test ./...

# Verbose
go test -v ./...

# Activities only
go test -v ./activities/

# Workflow integration test
go test -v ./workflow/
```

---

## Project Structure

```
temporal-worker-go/
├── go.mod
├── README.md
├── models/
│   └── models.go              # All data types (matches Java client)
├── activities/
│   ├── activities.go          # 4 stubbed agent activities
│   └── activities_test.go     # Unit tests for each activity
├── workflow/
│   ├── workflow.go            # Workflow chaining all 4 activities
│   └── workflow_test.go       # Integration tests (Temporal test env)
└── cmd/
    └── worker/
        └── main.go            # Worker entry point
```

---

## Environment Variables

| Variable              | Default            | Description                     |
|-----------------------|--------------------|---------------------------------|
| `TEMPORAL_TARGET`     | `localhost:7233`   | Temporal gRPC endpoint          |
| `TEMPORAL_NAMESPACE`  | `default`          | Temporal namespace              |
| `TEMPORAL_TASK_QUEUE` | `default-task-queue` | Task queue the worker polls   |

---

## End-to-End Flow

```
1. Java client  ─POST─▶  WorkflowController
2. Controller   ─start─▶  Temporal Server
3. Temporal      ─task──▶  Go Worker (this project)
4. Go Worker     ─runs──▶  ReviewRequest → AddMeaning → LogActivity → AIAnswer
5. Result        ─back──▶  Temporal Server → Java client (via WorkflowStub.getResult)
```
