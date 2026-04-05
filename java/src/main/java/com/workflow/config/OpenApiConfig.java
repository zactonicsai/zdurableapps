package com.workflow.config;

import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Contact;
import io.swagger.v3.oas.models.info.Info;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * OpenAPI 3 metadata shown in the Swagger UI.
 */
@Configuration
public class OpenApiConfig {

    @Bean
    public OpenAPI workflowOpenApi() {
        return new OpenAPI()
                .info(new Info()
                        .title("Temporal Workflow API")
                        .version("1.0.0")
                        .description("REST API for starting and managing "
                                + "Temporal workflows via Spring Boot 3.5")
                        .contact(new Contact()
                                .name("Workflow Team")
                                .email("workflow-team@example.com")));
    }
}
