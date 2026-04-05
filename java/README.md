# Temporal Workflow Library

A Spring Boot 3.5.13 library for starting and managing [Temporal](https://temporal.io/) workflows via REST, built with Temporal SDK 1.33.

---

## Tech Stack

| Component          | Version |
|--------------------|---------|
| Java               | 21      |
| Spring Boot        | 3.5.13  |
| Temporal SDK       | 1.33.0  |
| SpringDoc (Swagger)| 2.8.6   |
| JUnit 5 + Mockito  | (managed by Spring Boot BOM) |

---

## Project Structure

```
temporal-workflow-lib/
├── pom.xml
└── src/
    ├── main/
    │   ├── java/com/workflow/
    │   │   ├── TemporalWorkflowApplication.java   # Spring Boot entry-point
    │   │   ├── WorkflowCli.java                   # CLI runner
    │   │   ├── config/
    │   │   │   ├── OpenApiConfig.java              # Swagger metadata
    │   │   │   ├── TemporalConfig.java             # WorkflowServiceStubs & WorkflowClient beans
    │   │   │   └── TemporalProperties.java         # Externalized temporal.* config
    │   │   ├── controller/
    │   │   │   ├── GlobalExceptionHandler.java     # Centralised error handling
    │   │   │   └── WorkflowController.java         # POST /api/v1/workflows
    │   │   ├── model/
    │   │   │   ├── StartWorkflowRequest.java       # Request DTO
    │   │   │   ├── Status.java                     # READY | SET | GO
    │   │   │   ├── WorkflowData.java               # Payload (name, value, status, text)
    │   │   │   └── WorkflowStartResponse.java      # Response DTO
    │   │   ├── service/
    │   │   │   └── WorkflowStarterService.java     # Generic workflow start logic
    │   │   └── workflow/
    │   │       ├── GenericWorkflow.java             # @WorkflowInterface
    │   │       └── GenericWorkflowImpl.java         # Sample implementation
    │   └── resources/
    │       └── application.yml
    └── test/
        ├── java/com/workflow/
        │   ├── config/TemporalConfigTest.java
        │   ├── controller/WorkflowControllerTest.java
        │   ├── model/WorkflowDataTest.java
        │   └── service/WorkflowStarterServiceTest.java
        └── resources/
            └── application.yml
```

---

## Quick Start

### 1. Prerequisites

- Java 21+
- Maven 3.9+
- A running Temporal server (e.g. `temporal server start-dev`)

### 2. Build

```bash
mvn clean install
```

### 3. Run the Spring Boot app

```bash
mvn spring-boot:run
```

The app starts on `http://localhost:8080`.

### 4. Open Swagger UI

Navigate to: [http://localhost:8080/swagger-ui.html](http://localhost:8080/swagger-ui.html)

### 5. Start a workflow via cURL

```bash
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "taskQueue": "default-task-queue",
    "workflowId": "order-42",
    "data": {
      "name": "order-processing",
      "value": 99.95,
      "status": "READY",
      "text": "Process this order"
    }
  }'
```

Response (HTTP 202):
```json
{
  "workflowId": "order-42",
  "runId": "d4e5f6a7-b8c9-...",
  "workflowType": "GenericWorkflow"
}
```

### 6. Run the CLI

```bash
java -cp target/classes:target/dependency/* com.workflow.WorkflowCli
```

Or with a custom base URL:
```bash
java -cp target/classes:target/dependency/* com.workflow.WorkflowCli http://my-server:8080
```

---

## Configuration

All Temporal settings are in `application.yml` under the `temporal` prefix:

```yaml
temporal:
  target: localhost:7233       # gRPC endpoint
  namespace: default           # Temporal namespace
  enable-https: false          # TLS for gRPC
```

Override via environment variables:
```bash
TEMPORAL_TARGET=temporal.prod.internal:7233 \
TEMPORAL_NAMESPACE=production \
TEMPORAL_ENABLE_HTTPS=true \
mvn spring-boot:run
```

---

## Running Tests

```bash
mvn test
```

Tests use Mockito mocks for Temporal SDK classes — no running Temporal server required.

---

## Registering the Worker

The library ships the workflow interface + REST layer. To actually execute workflows you need a worker:

```java
@Bean
public WorkerFactory workerFactory(WorkflowClient client) {
    WorkerFactory factory = WorkerFactory.newInstance(client);
    Worker worker = factory.newWorker("default-task-queue");
    worker.registerWorkflowImplementationTypes(GenericWorkflowImpl.class);
    factory.start();
    return factory;
}
```

---

## License

MIT
