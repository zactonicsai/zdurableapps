package com.workflow.service;

import com.workflow.model.StartWorkflowRequest;
import com.workflow.model.WorkflowStartResponse;
import com.workflow.workflow.GenericWorkflow;
import io.temporal.client.WorkflowClient;
import io.temporal.client.WorkflowOptions;
import io.temporal.client.WorkflowStub;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.time.Duration;
import java.util.UUID;

/**
 * Service that starts a Temporal workflow in a generic, reusable way.
 * <p>
 * Callers supply a {@link StartWorkflowRequest} – the service builds the
 * typed workflow stub, fires-and-forgets the execution, and returns the
 * resulting workflow/run IDs.
 */
@Service
public class WorkflowStarterService {

    private static final Logger log = LoggerFactory.getLogger(WorkflowStarterService.class);

    private final WorkflowClient workflowClient;

    public WorkflowStarterService(WorkflowClient workflowClient) {
        this.workflowClient = workflowClient;
    }

    /**
     * Start a new {@link GenericWorkflow} execution.
     *
     * @param request contains the task queue, optional workflow ID, and payload
     * @return metadata about the started execution
     */
    public WorkflowStartResponse startWorkflow(StartWorkflowRequest request) {

        String workflowId = (request.workflowId() == null || request.workflowId().isBlank())
                ? request.data().getName() + "-" + UUID.randomUUID().toString().substring(0, 8)
                : request.workflowId();

        WorkflowOptions options = WorkflowOptions.newBuilder()
                .setTaskQueue(request.taskQueue())
                .setWorkflowId(workflowId)
                .setWorkflowExecutionTimeout(Duration.ofMinutes(10))
                .build();

        GenericWorkflow workflow =
                workflowClient.newWorkflowStub(GenericWorkflow.class, options);

        // Fire-and-forget – returns immediately
        WorkflowClient.start(workflow::execute, request.data());

        // Retrieve the run ID from the untyped stub
        WorkflowStub untypedStub = WorkflowStub.fromTyped(workflow);
        String runId = untypedStub.getExecution().getRunId();

        log.info("Started workflow workflowId={}, runId={}, taskQueue={}",
                workflowId, runId, request.taskQueue());

        return new WorkflowStartResponse(
                workflowId,
                runId,
                GenericWorkflow.class.getSimpleName()
        );
    }
}
