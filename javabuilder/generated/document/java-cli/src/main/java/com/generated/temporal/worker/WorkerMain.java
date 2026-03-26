package com.generated.temporal.worker;

import io.temporal.client.WorkflowClient;
import io.temporal.serviceclient.WorkflowServiceStubs;
import io.temporal.worker.Worker;
import io.temporal.worker.WorkerFactory;
import com.generated.temporal.activities.GeneratedActivitiesImpl;
import com.generated.temporal.workflow.GeneratedWorkflowImpl;

public class WorkerMain {
    public static void main(String[] args) {
        WorkflowServiceStubs service = WorkflowServiceStubs.newLocalServiceStubs();
        WorkflowClient client = WorkflowClient.newInstance(service);
        WorkerFactory factory = WorkerFactory.newInstance(client);
        Worker worker = factory.newWorker("document-ingest-queue");
        worker.registerWorkflowImplementationTypes(GeneratedWorkflowImpl.class);
        worker.registerActivitiesImplementations(new GeneratedActivitiesImpl());
        factory.start();
        System.out.println("Java 21 worker started on task queue: document-ingest-queue");
    }
}
