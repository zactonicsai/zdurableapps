# Generated Java 21 Temporal CLI

## Build
```bash
mvn clean compile
```

## Run worker
```bash
mvn exec:java -Dexec.mainClass=com.generated.temporal.worker.WorkerMain
```

## Run client
```bash
mvn exec:java -Dexec.mainClass=com.generated.temporal.client.DocumentIngestClientMain -Dexec.args="./sample-input/demo.pdf s3"
```

Task queue: `document-ingest-queue`

The generated code returns static values for each activity. Add your real processor logic in `GeneratedActivitiesImpl`.
