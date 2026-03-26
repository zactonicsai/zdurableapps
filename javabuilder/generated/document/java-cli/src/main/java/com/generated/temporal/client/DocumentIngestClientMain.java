package com.generated.temporal.client;

import java.util.UUID;
import io.temporal.client.WorkflowClient;
import io.temporal.client.WorkflowOptions;
import io.temporal.serviceclient.WorkflowServiceStubs;
import com.generated.temporal.model.*;
import com.generated.temporal.workflow.GeneratedWorkflow;

public class DocumentIngestClientMain {
    public static void main(String[] args) {
        String filePath = args.length > 0 ? args[0] : "./sample-input/demo.pdf";
        String storageTypeValue = args.length > 1 ? args[1] : "s3";
        WorkflowServiceStubs service = WorkflowServiceStubs.newLocalServiceStubs();
        WorkflowClient client = WorkflowClient.newInstance(service);
        WorkflowOptions options = WorkflowOptions.newBuilder()
            .setTaskQueue("document-ingest-queue")
            .setWorkflowId("documentingest-workflow-" + UUID.randomUUID())
            .build();
        GeneratedWorkflow workflow = client.newWorkflowStub(GeneratedWorkflow.class, options);
        DocumentIngestInput input = new DocumentIngestInput(filePath, FileStorageType.fromValue(storageTypeValue));
        var result = workflow.run(input);
        System.out.println(result);
    }
}
