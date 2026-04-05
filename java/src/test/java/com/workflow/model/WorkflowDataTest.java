package com.workflow.model;

import jakarta.validation.ConstraintViolation;
import jakarta.validation.Validation;
import jakarta.validation.Validator;
import jakarta.validation.ValidatorFactory;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.util.Set;

import static org.assertj.core.api.Assertions.assertThat;

class WorkflowDataTest {

    private static Validator validator;

    @BeforeAll
    static void setUpValidator() {
        try (ValidatorFactory factory = Validation.buildDefaultValidatorFactory()) {
            validator = factory.getValidator();
        }
    }

    @Test
    @DisplayName("Valid WorkflowData passes validation")
    void valid() {
        WorkflowData data = new WorkflowData("order-1", 42.0, Status.READY, "some text");
        Set<ConstraintViolation<WorkflowData>> violations = validator.validate(data);
        assertThat(violations).isEmpty();
    }

    @Test
    @DisplayName("Blank name triggers violation")
    void blankName() {
        WorkflowData data = new WorkflowData("", 1.0, Status.SET, null);
        Set<ConstraintViolation<WorkflowData>> violations = validator.validate(data);
        assertThat(violations).anyMatch(v -> v.getPropertyPath().toString().equals("name"));
    }

    @Test
    @DisplayName("Null value triggers violation")
    void nullValue() {
        WorkflowData data = new WorkflowData("ok", null, Status.GO, null);
        Set<ConstraintViolation<WorkflowData>> violations = validator.validate(data);
        assertThat(violations).anyMatch(v -> v.getPropertyPath().toString().equals("value"));
    }

    @Test
    @DisplayName("Null status triggers violation")
    void nullStatus() {
        WorkflowData data = new WorkflowData("ok", 1.0, null, null);
        Set<ConstraintViolation<WorkflowData>> violations = validator.validate(data);
        assertThat(violations).anyMatch(v -> v.getPropertyPath().toString().equals("status"));
    }

    @Test
    @DisplayName("Text exceeding 1024 chars triggers violation")
    void textTooLong() {
        String longText = "a".repeat(1025);
        WorkflowData data = new WorkflowData("ok", 1.0, Status.READY, longText);
        Set<ConstraintViolation<WorkflowData>> violations = validator.validate(data);
        assertThat(violations).anyMatch(v -> v.getPropertyPath().toString().equals("text"));
    }

    @Test
    @DisplayName("Text of exactly 1024 chars is valid")
    void textExactly1024() {
        String maxText = "b".repeat(1024);
        WorkflowData data = new WorkflowData("ok", 1.0, Status.GO, maxText);
        Set<ConstraintViolation<WorkflowData>> violations = validator.validate(data);
        assertThat(violations).isEmpty();
    }

    @Test
    @DisplayName("Null text is valid (field is optional)")
    void nullText() {
        WorkflowData data = new WorkflowData("ok", 1.0, Status.SET, null);
        Set<ConstraintViolation<WorkflowData>> violations = validator.validate(data);
        assertThat(violations).isEmpty();
    }

    @Test
    @DisplayName("equals and hashCode contract")
    void equalsHashCode() {
        WorkflowData a = new WorkflowData("x", 1.0, Status.READY, "t");
        WorkflowData b = new WorkflowData("x", 1.0, Status.READY, "t");
        assertThat(a).isEqualTo(b);
        assertThat(a.hashCode()).isEqualTo(b.hashCode());
    }

    @Test
    @DisplayName("toString truncates long text")
    void toStringTruncation() {
        String longText = "z".repeat(100);
        WorkflowData data = new WorkflowData("n", 1.0, Status.GO, longText);
        assertThat(data.toString()).contains("…");
    }
}
