# Java Generator Fix

## What changed

The Go generator (`cmd/tidsgen/main.go`) previously only generated Python code.
Java files in `java-cli/` were hand-written and not produced by the generator.

### Changes made

1. **Added full Java code generation** to `Generate()` — the generator now emits
   a complete `java-cli/` Maven project alongside the Python files:
   - `pom.xml` with Temporal SDK, SLF4J, and Jackson dependencies
   - Java 21 records for all schema struct types
   - Java enums for all schema enum types
   - `GeneratedActivities` interface with `@ActivityInterface`/`@ActivityMethod`
   - `GeneratedActivitiesImpl` with static stub returns
   - `GeneratedWorkflow` interface with `@WorkflowInterface`/`@WorkflowMethod`
   - `GeneratedWorkflowImpl` with proper step variable wiring
   - `WorkerMain` that registers workflow + activities on the schema task queue
   - Client main class with CLI arg parsing for workflow input fields
   - `.gitignore`, `README.md`

2. **Fixed the hardcoded variable name bug** — the workflow impl now builds a
   `stepVars` map (`activity_name -> output_var`) from the schema steps, and
   passes it through `javaWorkflowStepInput`, `javaFieldExpr`, and
   `javaWorkflowOutputExpr`. Field resolution works generically:
   - Fields matching workflow input fields -> `input.fieldName()`
   - Fields matching a prior activity's output -> `stepVar.fieldName()`
   - `text_preview`/`excerpt` fields -> auto-truncate from the text source

3. **Updated `docker-compose.yml`** rendering to include `java-worker` and
   `java-client` services alongside the Python ones.

4. **Updated schema** to list `java` in the `language` array.

## Symptom (before fix)

Running `go run ./cmd/tidsgen` would only produce Python files. The Java
`java-cli/` directory had to be maintained by hand, and if the schema changed,
the Java code would break with errors like:

```
cannot find symbol: variable fileTypeResult
```

## Regenerate

```bash
go mod tidy
go run ./cmd/tidsgen -schema ./examples/document_ingest_schema.yaml -out ./generated/document_ingest
cd ./generated/document_ingest/java-cli
mvn clean compile
```

## Expected generated workflow (Java)

```java
var file_type_result = activities.determineFileType(new DetermineFileTypeInput(input.filePath()));
var convert_result = activities.convertToText(new ConvertToTextInput(input.filePath(), file_type_result.fileType()));
var save_result = activities.saveConvertedText(new SaveConvertedTextInput(input.filePath(), input.storageType(), convert_result.text()));
return new DocumentIngestResult(
    file_type_result.fileType(),
    file_type_result.mimeType(),
    convert_result.text().substring(0, Math.min(120, convert_result.text().length())),
    save_result.savedTo(),
    save_result.recordId()
);
```

Variable names (`file_type_result`, `convert_result`, `save_result`) come from
the schema `output_var` fields, not hardcoded strings.
