package com.workflow.config;

import io.temporal.client.WorkflowClient;
import io.temporal.client.WorkflowClientOptions;
import io.temporal.serviceclient.WorkflowServiceStubs;
import io.temporal.serviceclient.WorkflowServiceStubsOptions;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * Spring configuration that wires the Temporal SDK beans.
 * <p>
 * Creates:
 * <ul>
 *     <li>{@link WorkflowServiceStubs} – the gRPC channel to Temporal server</li>
 *     <li>{@link WorkflowClient}       – used to start / signal / query workflows</li>
 * </ul>
 */
@Configuration
@EnableConfigurationProperties(TemporalProperties.class)
public class TemporalConfig {

    private static final Logger log = LoggerFactory.getLogger(TemporalConfig.class);

    /**
     * gRPC stub connected to the Temporal front-end service.
     */
    @Bean(destroyMethod = "shutdown")
    public WorkflowServiceStubs workflowServiceStubs(TemporalProperties props) {
        log.info("Connecting to Temporal at {} (namespace={}, tls={})",
                props.getTarget(), props.getNamespace(), props.isEnableHttps());

        WorkflowServiceStubsOptions.Builder builder =
                WorkflowServiceStubsOptions.newBuilder()
                        .setTarget(props.getTarget());

        if (props.isEnableHttps()) {
            builder.setEnableHttps(true);
        }

        return WorkflowServiceStubs.newServiceStubs(builder.build());
    }

    /**
     * Workflow client configured for the target namespace.
     */
    @Bean
    public WorkflowClient workflowClient(WorkflowServiceStubs stubs,
                                         TemporalProperties props) {
        WorkflowClientOptions options =
                WorkflowClientOptions.newBuilder()
                        .setNamespace(props.getNamespace())
                        .build();

        return WorkflowClient.newInstance(stubs, options);
    }
}
