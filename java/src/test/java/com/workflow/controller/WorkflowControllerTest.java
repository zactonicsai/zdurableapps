package com.workflow.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.workflow.model.StartWorkflowRequest;
import com.workflow.model.Status;
import com.workflow.model.WorkflowData;
import com.workflow.model.WorkflowStartResponse;
import com.workflow.service.WorkflowStarterService;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.test.context.bean.override.mockito.MockitoBean;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(WorkflowController.class)
class WorkflowControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private ObjectMapper objectMapper;

    @MockitoBean
    private WorkflowStarterService starterService;

    @Test
    @DisplayName("POST /api/v1/workflows – 202 with valid request")
    void startWorkflow_returns202() throws Exception {

        WorkflowStartResponse response =
                new WorkflowStartResponse("wf-123", "run-456", "GenericWorkflow");

        when(starterService.startWorkflow(any(StartWorkflowRequest.class)))
                .thenReturn(response);

        WorkflowData data = new WorkflowData("test", 9.99, Status.SET, "hello");
        StartWorkflowRequest request = new StartWorkflowRequest("tq", "wf-123", data);

        mockMvc.perform(post("/api/v1/workflows")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isAccepted())
                .andExpect(jsonPath("$.workflowId").value("wf-123"))
                .andExpect(jsonPath("$.runId").value("run-456"))
                .andExpect(jsonPath("$.workflowType").value("GenericWorkflow"));
    }

    @Test
    @DisplayName("POST /api/v1/workflows – 400 when name is blank")
    void startWorkflow_returns400_whenNameBlank() throws Exception {

        WorkflowData data = new WorkflowData("", 1.0, Status.READY, null);
        StartWorkflowRequest request = new StartWorkflowRequest("tq", null, data);

        mockMvc.perform(post("/api/v1/workflows")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isBadRequest());
    }

    @Test
    @DisplayName("POST /api/v1/workflows – 400 when data is null")
    void startWorkflow_returns400_whenDataNull() throws Exception {

        String json = """
                { "taskQueue": "tq", "data": null }
                """;

        mockMvc.perform(post("/api/v1/workflows")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(json))
                .andExpect(status().isBadRequest());
    }

    @Test
    @DisplayName("POST /api/v1/workflows – 400 when text exceeds 1024 chars")
    void startWorkflow_returns400_whenTextTooLong() throws Exception {

        String longText = "x".repeat(1025);
        WorkflowData data = new WorkflowData("name", 1.0, Status.GO, longText);
        StartWorkflowRequest request = new StartWorkflowRequest("tq", null, data);

        mockMvc.perform(post("/api/v1/workflows")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isBadRequest());
    }
}
