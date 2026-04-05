package com.workflow.model;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

/**
 * Request body for the start-workflow endpoint.
 */
@Schema(description = "Request to start a new Temporal workflow")
public record StartWorkflowRequest(

        @Schema(description = "Temporal task queue the worker polls",
                example = "default-task-queue")
        @NotBlank(message = "taskQueue is required")
        String taskQueue,

        @Schema(description = "Optional custom workflow ID (auto-generated if blank)",
                example = "order-42")
        String workflowId,

        @Schema(description = "Workflow payload")
        @NotNull(message = "data is required")
        @Valid
        WorkflowData data
) {
}
