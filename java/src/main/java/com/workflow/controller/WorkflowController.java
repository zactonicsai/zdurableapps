package com.workflow.controller;

import com.workflow.model.StartWorkflowRequest;
import com.workflow.model.WorkflowStartResponse;
import com.workflow.service.WorkflowStarterService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.responses.ApiResponses;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * REST endpoint for starting Temporal workflows.
 */
@RestController
@RequestMapping("/api/v1/workflows")
@Tag(name = "Workflow", description = "Temporal workflow operations")
public class WorkflowController {

    private static final Logger log = LoggerFactory.getLogger(WorkflowController.class);

    private final WorkflowStarterService starterService;

    public WorkflowController(WorkflowStarterService starterService) {
        this.starterService = starterService;
    }

    @Operation(
            summary = "Start a new workflow",
            description = "Submits a fire-and-forget workflow execution to Temporal "
                    + "and returns the workflow & run IDs."
    )
    @ApiResponses({
            @ApiResponse(
                    responseCode = "202",
                    description = "Workflow accepted and started",
                    content = @Content(
                            mediaType = MediaType.APPLICATION_JSON_VALUE,
                            schema = @Schema(implementation = WorkflowStartResponse.class)
                    )
            ),
            @ApiResponse(responseCode = "400", description = "Invalid request payload"),
            @ApiResponse(responseCode = "500", description = "Temporal service unavailable")
    })
    @PostMapping(
            consumes = MediaType.APPLICATION_JSON_VALUE,
            produces = MediaType.APPLICATION_JSON_VALUE
    )
    public ResponseEntity<WorkflowStartResponse> startWorkflow(
            @Valid @RequestBody StartWorkflowRequest request) {

        log.info("POST /api/v1/workflows – starting workflow for {}", request.data().getName());
        WorkflowStartResponse response = starterService.startWorkflow(request);
        return ResponseEntity.status(HttpStatus.ACCEPTED).body(response);
    }
}
