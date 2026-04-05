package com.workflow.workflow;

import com.workflow.model.WorkflowData;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Sample implementation of {@link GenericWorkflow}.
 * <p>
 * Register this with a Temporal worker that polls the desired task queue.
 */
public class GenericWorkflowImpl implements GenericWorkflow {

    private static final Logger log = LoggerFactory.getLogger(GenericWorkflowImpl.class);

    @Override
    public String execute(WorkflowData data) {
        log.info("Workflow executing: {}", data);

        // ── place your activity calls / saga logic here ─────
        // e.g.:
        //   activities.validate(data);
        //   activities.process(data);
        //   activities.notify(data);

        return "Workflow completed – name=%s, status=%s, value=%s"
                .formatted(data.getName(), data.getStatus(), data.getValue());
    }
}
