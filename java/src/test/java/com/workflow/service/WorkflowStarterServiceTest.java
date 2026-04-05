package com.workflow.service;

import com.workflow.model.StartWorkflowRequest;
import com.workflow.model.Status;
import com.workflow.model.WorkflowData;
import com.workflow.model.WorkflowStartResponse;
import com.workflow.workflow.GenericWorkflow;
import io.temporal.client.WorkflowClient;
import io.temporal.client.WorkflowOptions;
import io.temporal.client.WorkflowStub;
import io.temporal.api.common.v1.WorkflowExecution;
import io.temporal.workflow.Functions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentMatchers;
import org.mockito.Mock;
import org.mockito.MockedStatic;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class WorkflowStarterServiceTest {

    @Mock
    private WorkflowClient workflowClient;

    @Mock
    private GenericWorkflow workflowStub;

    @Mock
    private WorkflowStub untypedStub;

    private WorkflowStarterService service;

    @BeforeEach
    void setUp() {
        service = new WorkflowStarterService(workflowClient);
    }

    @Test
    @DisplayName("startWorkflow – generates ID when not provided and returns response")
    void startWorkflow_generatesId() {
        // Arrange
        WorkflowData data = new WorkflowData("order-test", 19.99, Status.READY, "unit test");
        StartWorkflowRequest request = new StartWorkflowRequest("test-queue", null, data);

        when(workflowClient.newWorkflowStub(eq(GenericWorkflow.class), any(WorkflowOptions.class)))
                .thenReturn(workflowStub);

        WorkflowExecution execution = WorkflowExecution.newBuilder()
                .setWorkflowId("order-test-abc12345")
                .setRunId("run-id-999")
                .build();

        try (MockedStatic<WorkflowClient> wcStatic = mockStatic(WorkflowClient.class);
             MockedStatic<WorkflowStub> wsStatic = mockStatic(WorkflowStub.class)) {

            // Cast to Func1 to resolve overload ambiguity – execute() returns String
            wcStatic.when(() -> WorkflowClient.start(
                    ArgumentMatchers.<Functions.Func1<WorkflowData, String>>any(),
                    any(WorkflowData.class)
            )).thenReturn(null);

            wsStatic.when(() -> WorkflowStub.fromTyped(workflowStub))
                    .thenReturn(untypedStub);

            when(untypedStub.getExecution()).thenReturn(execution);

            // Act
            WorkflowStartResponse response = service.startWorkflow(request);

            // Assert
            assertThat(response).isNotNull();
            assertThat(response.workflowId()).startsWith("order-test-");
            assertThat(response.runId()).isEqualTo("run-id-999");
            assertThat(response.workflowType()).isEqualTo("GenericWorkflow");
        }
    }

    @Test
    @DisplayName("startWorkflow – uses provided workflow ID when present")
    void startWorkflow_usesProvidedId() {
        // Arrange
        WorkflowData data = new WorkflowData("my-flow", 100.0, Status.GO, "go!");
        StartWorkflowRequest request = new StartWorkflowRequest("q", "custom-id-42", data);

        when(workflowClient.newWorkflowStub(eq(GenericWorkflow.class), any(WorkflowOptions.class)))
                .thenReturn(workflowStub);

        WorkflowExecution execution = WorkflowExecution.newBuilder()
                .setWorkflowId("custom-id-42")
                .setRunId("run-abc")
                .build();

        try (MockedStatic<WorkflowClient> wcStatic = mockStatic(WorkflowClient.class);
             MockedStatic<WorkflowStub> wsStatic = mockStatic(WorkflowStub.class)) {

            wcStatic.when(() -> WorkflowClient.start(
                    ArgumentMatchers.<Functions.Func1<WorkflowData, String>>any(),
                    any(WorkflowData.class)
            )).thenReturn(null);

            wsStatic.when(() -> WorkflowStub.fromTyped(workflowStub))
                    .thenReturn(untypedStub);
            when(untypedStub.getExecution()).thenReturn(execution);

            // Act
            WorkflowStartResponse response = service.startWorkflow(request);

            // Assert
            assertThat(response.workflowId()).isEqualTo("custom-id-42");
            assertThat(response.runId()).isEqualTo("run-abc");
        }
    }

    @Test
    @DisplayName("startWorkflow – blank workflowId is treated as auto-generate")
    void startWorkflow_blankIdAutoGenerates() {
        // Arrange
        WorkflowData data = new WorkflowData("auto-gen", 0.0, Status.SET, null);
        StartWorkflowRequest request = new StartWorkflowRequest("tq", "   ", data);

        when(workflowClient.newWorkflowStub(eq(GenericWorkflow.class), any(WorkflowOptions.class)))
                .thenReturn(workflowStub);

        WorkflowExecution execution = WorkflowExecution.newBuilder()
                .setWorkflowId("auto-gen-12345678")
                .setRunId("run-xyz")
                .build();

        try (MockedStatic<WorkflowClient> wcStatic = mockStatic(WorkflowClient.class);
             MockedStatic<WorkflowStub> wsStatic = mockStatic(WorkflowStub.class)) {

            wcStatic.when(() -> WorkflowClient.start(
                    ArgumentMatchers.<Functions.Func1<WorkflowData, String>>any(),
                    any(WorkflowData.class)
            )).thenReturn(null);

            wsStatic.when(() -> WorkflowStub.fromTyped(workflowStub))
                    .thenReturn(untypedStub);
            when(untypedStub.getExecution()).thenReturn(execution);

            // Act
            WorkflowStartResponse response = service.startWorkflow(request);

            // Assert – blank was treated as null, so ID should be auto-generated
            assertThat(response.workflowId()).startsWith("auto-gen-");
        }
    }
}
