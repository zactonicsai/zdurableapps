package com.workflow;

import com.workflow.model.StartWorkflowRequest;
import com.workflow.model.Status;
import com.workflow.model.WorkflowData;
import com.workflow.model.WorkflowStartResponse;
import org.springframework.web.client.RestClient;

/**
 * Standalone CLI that POSTs a workflow start request to the running Spring Boot app.
 * <p>
 * Usage:
 * <pre>
 *   # 1) Start the app:   mvn spring-boot:run
 *   # 2) Run the CLI:     mvn exec:java -Dexec.mainClass="com.workflow.WorkflowCli"
 *   #    or:              java -cp target/classes com.workflow.WorkflowCli [baseUrl]
 * </pre>
 */
public class WorkflowCli {

    private static final String DEFAULT_BASE_URL = "http://localhost:8080";

    public static void main(String[] args) {
        String baseUrl = args.length > 0 ? args[0] : DEFAULT_BASE_URL;

        System.out.println("╔══════════════════════════════════════════════╗");
        System.out.println("║  Temporal Workflow CLI                       ║");
        System.out.println("╚══════════════════════════════════════════════╝");
        System.out.println("Target: " + baseUrl);
        System.out.println();

        // Build the payload
        WorkflowData data = new WorkflowData(
                "cli-demo-order",
                42.99,
                Status.READY,
                "Submitted from the CLI runner"
        );

        StartWorkflowRequest request = new StartWorkflowRequest(
                "default-task-queue",
                null,   // auto-generate workflow ID
                data
        );

        // POST via Spring 6's RestClient (zero extra deps)
        RestClient client = RestClient.builder()
                .baseUrl(baseUrl)
                .build();

        try {
            WorkflowStartResponse response = client.post()
                    .uri("/api/v1/workflows")
                    .header("Content-Type", "application/json")
                    .body(request)
                    .retrieve()
                    .body(WorkflowStartResponse.class);

            System.out.println("✓ Workflow started successfully!");
            System.out.println("  Workflow ID   : " + response.workflowId());
            System.out.println("  Run ID        : " + response.runId());
            System.out.println("  Workflow Type : " + response.workflowType());

        } catch (Exception ex) {
            System.err.println("✗ Failed to start workflow: " + ex.getMessage());
            System.exit(1);
        }
    }
}
