package com.workflow.workflow;

import com.workflow.model.WorkflowData;
import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;

/**
 * Generic Temporal workflow contract.
 * <p>
 * Any concrete workflow implementation that should be startable through
 * the library's REST layer must implement this interface.
 */
@WorkflowInterface
public interface GenericWorkflow {

    /**
     * Entry-point executed by Temporal when the workflow is started.
     *
     * @param data the payload supplied by the caller
     * @return a human-readable result string
     */
    @WorkflowMethod
    String execute(WorkflowData data);
}
