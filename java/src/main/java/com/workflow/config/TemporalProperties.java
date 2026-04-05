package com.workflow.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

/**
 * Externalized Temporal connection settings.
 * <p>
 * Bind with prefix {@code temporal} in {@code application.yml}:
 * <pre>
 * temporal:
 *   target: localhost:7233
 *   namespace: default
 *   enable-https: false
 * </pre>
 */
@ConfigurationProperties(prefix = "temporal")
public class TemporalProperties {

    /** gRPC target for the Temporal front-end service (host:port). */
    private String target = "localhost:7233";

    /** Temporal namespace to operate in. */
    private String namespace = "default";

    /** Whether to use TLS for the gRPC connection. */
    private boolean enableHttps = false;

    // ── getters / setters ───────────────────────────────────

    public String getTarget() {
        return target;
    }

    public void setTarget(String target) {
        this.target = target;
    }

    public String getNamespace() {
        return namespace;
    }

    public void setNamespace(String namespace) {
        this.namespace = namespace;
    }

    public boolean isEnableHttps() {
        return enableHttps;
    }

    public void setEnableHttps(boolean enableHttps) {
        this.enableHttps = enableHttps;
    }
}
