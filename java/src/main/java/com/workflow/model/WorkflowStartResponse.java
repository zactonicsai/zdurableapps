package com.workflow.model;

import io.swagger.v3.oas.annotations.media.Schema;

/**
 * Response returned after a workflow is successfully started.
 */
@Schema(description = "Result of starting a workflow")
public record WorkflowStartResponse(

        @Schema(description = "Temporal workflow ID", example = "order-processing-42-a1b2c3")
        String workflowId,

        @Schema(description = "Temporal run ID", example = "d4e5f6a7-b8c9-0123-4567-89abcdef0123")
        String runId,

        @Schema(description = "Workflow type that was started", example = "GenericWorkflow")
        String workflowType
) {
}
