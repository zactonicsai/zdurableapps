# Generated Java 21 Temporal CLI

## Build
```bash
mvn clean package
```

## Run worker

Fat JAR (default main class is the worker):
```bash
java -jar target/temporal-generated-java-cli-1.0.0.jar
```

Or via Maven:
```bash
mvn exec:java -Dexec.mainClass=com.generated.temporal.worker.WorkerMain
```

## Run client

Build the client fat JAR:
```bash
mvn clean package -Pclient
java -jar target/temporal-generated-java-cli-1.0.0.jar ./sample-input/demo.pdf s3
```

Or via Maven (no rebuild needed):
```bash
mvn exec:java -Dexec.mainClass=com.generated.temporal.client.DocumentIngestClientMain -Dexec.args="./sample-input/demo.pdf s3"
```

Or run the client class directly from the worker JAR:
```bash
java -cp target/temporal-generated-java-cli-1.0.0.jar com.generated.temporal.client.DocumentIngestClientMain ./sample-input/demo.pdf s3
```

## Maven profiles

| Profile   | Main class | Usage |
|-----------|-----------|-------|
| (default) | `com.generated.temporal.worker.WorkerMain` | `mvn clean package` |
| worker    | `com.generated.temporal.worker.WorkerMain` | `mvn clean package -Pworker` |
| client    | `com.generated.temporal.client.DocumentIngestClientMain` | `mvn clean package -Pclient` |

Task queue: `document-ingest-queue`

The generated code returns static values for each activity. Add your real
processor logic in `GeneratedActivitiesImpl`.
