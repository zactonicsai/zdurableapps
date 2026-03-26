package com.generated.temporal.workflow;

import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;
import com.generated.temporal.model.*;

@WorkflowInterface
public interface GeneratedWorkflow {
    @WorkflowMethod(name = "DocumentIngestWorkflow")
    DocumentIngestResult run(DocumentIngestInput input);
}
