package com.workflow.model;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Size;
import java.io.Serializable;
import java.util.Objects;

/**
 * Payload passed into every Temporal workflow execution.
 */
@Schema(description = "Data payload for a workflow execution")
public class WorkflowData implements Serializable {

    private static final long serialVersionUID = 1L;

    @Schema(description = "Logical name of the workflow run", example = "order-processing-42")
    @NotBlank(message = "name is required")
    private String name;

    @Schema(description = "Numeric value associated with the workflow", example = "99.95")
    @NotNull(message = "value is required")
    private Double value;

    @Schema(description = "Current status", example = "READY")
    @NotNull(message = "status is required")
    private Status status;

    @Schema(description = "Free-form text (max 1024 chars)", example = "Process this order ASAP")
    @Size(max = 1024, message = "text must not exceed 1024 characters")
    private String text;

    // ── constructors ────────────────────────────────────────

    public WorkflowData() {
    }

    public WorkflowData(String name, Double value, Status status, String text) {
        this.name = name;
        this.value = value;
        this.status = status;
        this.text = text;
    }

    // ── getters / setters ───────────────────────────────────

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Double getValue() {
        return value;
    }

    public void setValue(Double value) {
        this.value = value;
    }

    public Status getStatus() {
        return status;
    }

    public void setStatus(Status status) {
        this.status = status;
    }

    public String getText() {
        return text;
    }

    public void setText(String text) {
        this.text = text;
    }

    // ── equals / hashCode / toString ────────────────────────

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        WorkflowData that = (WorkflowData) o;
        return Objects.equals(name, that.name)
                && Objects.equals(value, that.value)
                && status == that.status
                && Objects.equals(text, that.text);
    }

    @Override
    public int hashCode() {
        return Objects.hash(name, value, status, text);
    }

    @Override
    public String toString() {
        return "WorkflowData{name='%s', value=%s, status=%s, text='%s'}"
                .formatted(name, value, status,
                        text == null ? null
                                : text.length() > 60 ? text.substring(0, 60) + "…" : text);
    }
}
