# Temporal Interface Definition Schema (TIDS) — v1.0.0

## Overview

TIDS is a declarative YAML schema for defining the complete interface of a
Temporal.io application: activities, workflows, signals, queries, updates,
pipelines, workers, and clients. A code generator reads a conforming YAML file
and produces fully-typed, production-ready Temporal worker and client code in
the target language.

## Files in This Package

| File | Purpose |
|---|---|
| `temporal-interface-schema.yaml` | **The schema itself** — documents every key, its type, and constraints. Treat this as the spec a generator validates against. |
| `temporal-interface-example.yaml` | **Concrete example** — an order-processing system that exercises every section of the schema (sync/async activities, saga compensation, cron workflows, fan-out pipelines, multiple workers & clients). |

---

## Key Design Decisions

### Sync vs Async Activities

The `mode` field on each activity tells the generator what wiring to produce:

- **`sync`** — Standard blocking activity. The generator produces a simple
  function signature with the configured timeouts and retry policy.
- **`async`** — The generator wires up heartbeat infrastructure and, where the
  SDK supports it, an async-completion token flow. This is critical for
  long-running operations like payment capture or warehouse fulfillment where
  the activity must periodically report liveness.

### Sync vs Async vs Cron Workflows

- **`sync`** — The generated client stub calls `execute_workflow` and blocks
  until a result is returned. Best for lightweight, fast lookups.
- **`async`** — The generated client stub calls `start_workflow` and returns a
  handle. The caller can later query, signal, or await the result.
- **`cron`** — Same as `async` but the generator adds the `cron_schedule` to
  the start options. The generated worker registers the workflow as usual; the
  cron semantics are handled by the Temporal server.

### Pipeline Configuration

Pipelines describe **multi-workflow orchestration** above the level of a single
workflow. Each pipeline has ordered stages that can run sequentially, in
parallel, or as fan-out/fan-in over a dynamic collection. This lets the
generator produce an orchestrator workflow (or a thin dispatch layer) that
coordinates child workflows with bounded concurrency, stage-level error
handling, and optional compensation rollback.

### Declarative Step Graph

Workflow `steps` use a declarative DAG approach rather than imperative code.
Each step declares its `kind`, inputs, outputs, dependencies (`depends_on`),
and error handling. The generator is responsible for translating this graph into
deterministic workflow code in the target language, respecting the dependency
order and wiring up saga-style compensation when `on_error.strategy: fallback`
is used.

---

## How a Code Generator Should Consume This

### 1. Parse & Validate

```
Load YAML → resolve type references → validate against schema
```

- Ensure all `type` references in activities/workflows resolve to entries in
  `types` or to built-in primitives (`string`, `int`, `bool`, `float`,
  `datetime`, `duration`, `null`, `any`, `list<T>`, `map<K,V>`).
- Ensure all `activity`, `workflow`, `signal`, `query`, `update`, and
  `retry_policy` references point to defined entries.
- Validate `depends_on` edges form a DAG (no cycles).

### 2. Generate Types

For each entry in `types`, produce:
- A data class / struct / interface in the target language.
- Serialization/deserialization helpers compatible with the chosen
  `data_converter.encoding`.

### 3. Generate Activities

For each activity, produce:
- An abstract activity function signature (input → output) for the developer to
  implement.
- A registration helper that wires the function into the worker with the
  correct timeouts, retry policy, and task queue.
- If `mode: async`: heartbeat scaffolding and async-completion wiring.

### 4. Generate Workflows

For each workflow, produce:
- A workflow class/function with the declared steps translated into
  deterministic SDK calls.
- Signal handler registration.
- Query handler registration.
- Update handler registration (with optional validator).
- Compensation logic derived from `on_error.strategy: fallback` chains.

### 5. Generate Workers

For each worker, produce:
- A runnable entry point that creates a Temporal worker, registers the listed
  activities and workflows, applies concurrency settings, attaches
  interceptors, and starts polling.
- Optionally, a Dockerfile / Kubernetes manifest using the `runtime` hints.

### 6. Generate Clients

For each client, produce:
- A typed client class with methods for every `allowed_workflows`,
  `allowed_signals`, `allowed_queries`, and `allowed_updates`.
- Connection setup with TLS, data converter, and interceptor wiring.
- Two method variants per workflow when `default_mode` is overridable:
  `start_<workflow>` (async handle) and `execute_<workflow>` (sync await).

### 7. Generate Pipelines (Optional)

For each pipeline, produce:
- An orchestrator workflow (often itself a Temporal workflow) that executes
  stages in order, manages fan-out concurrency, and runs compensation on
  failure.
- A trigger adapter (webhook handler, cron starter, signal listener) based on
  `trigger.kind`.

---

## Type Expression Syntax

Type expressions used in `input.type`, `output.type`, and `fields[].type`
follow this grammar:

```
type_expr  = primitive | reference | generic
primitive  = "string" | "int" | "float" | "bool" | "datetime" | "duration" | "null" | "any"
reference  = PascalCaseName          (refers to an entry in `types`)
generic    = "list<" type_expr ">"
           | "map<" type_expr "," type_expr ">"
           | "optional<" type_expr ">"
```

---

## Duration Format

All `duration` values use Go-style shorthand: `500ms`, `5s`, `1m`, `2h`, `24h`.

---

## Extension Points

- **`annotations`** on activities, workflows, workers, clients, and pipelines
  allow generator plugins to receive arbitrary configuration without changing
  the core schema.
- **`labels`** on metadata propagate to generated code as constants or tags.
- **`interceptors`** are name-referenced; the generator is expected to have a
  registry of known interceptor implementations per target language.
