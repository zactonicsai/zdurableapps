package com.workflow.config;

import io.temporal.client.WorkflowClient;
import io.temporal.client.WorkflowClientOptions;
import io.temporal.serviceclient.WorkflowServiceStubs;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.MockedStatic;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class TemporalConfigTest {

    @Test
    @DisplayName("workflowClient bean is created with correct namespace")
    void workflowClient_usesConfiguredNamespace() {
        TemporalProperties props = new TemporalProperties();
        props.setTarget("localhost:7233");
        props.setNamespace("test-ns");
        props.setEnableHttps(false);

        assertThat(props.getTarget()).isEqualTo("localhost:7233");
        assertThat(props.getNamespace()).isEqualTo("test-ns");
        assertThat(props.isEnableHttps()).isFalse();
    }

    @Test
    @DisplayName("workflowClient bean delegates to WorkflowClient.newInstance with correct namespace")
    void workflowClient_wrapsStub() {
        WorkflowServiceStubs mockStubs = mock(WorkflowServiceStubs.class);
        WorkflowClient mockClient = mock(WorkflowClient.class);

        TemporalProperties props = new TemporalProperties();
        props.setNamespace("my-ns");

        TemporalConfig config = new TemporalConfig();

        // Mock the static factory – the real one needs a live gRPC channel
        try (MockedStatic<WorkflowClient> wcStatic = mockStatic(WorkflowClient.class)) {
            wcStatic.when(() -> WorkflowClient.newInstance(eq(mockStubs), any(WorkflowClientOptions.class)))
                    .thenReturn(mockClient);

            WorkflowClient client = config.workflowClient(mockStubs, props);

            assertThat(client).isSameAs(mockClient);
            wcStatic.verify(() ->
                    WorkflowClient.newInstance(eq(mockStubs), any(WorkflowClientOptions.class)));
        }
    }

    @Test
    @DisplayName("TemporalProperties defaults are sane")
    void defaults() {
        TemporalProperties props = new TemporalProperties();
        assertThat(props.getTarget()).isEqualTo("localhost:7233");
        assertThat(props.getNamespace()).isEqualTo("default");
        assertThat(props.isEnableHttps()).isFalse();
    }

    @Test
    @DisplayName("TemporalProperties setters work")
    void setters() {
        TemporalProperties props = new TemporalProperties();
        props.setTarget("temporal.prod:7233");
        props.setNamespace("production");
        props.setEnableHttps(true);

        assertThat(props.getTarget()).isEqualTo("temporal.prod:7233");
        assertThat(props.getNamespace()).isEqualTo("production");
        assertThat(props.isEnableHttps()).isTrue();
    }
}
