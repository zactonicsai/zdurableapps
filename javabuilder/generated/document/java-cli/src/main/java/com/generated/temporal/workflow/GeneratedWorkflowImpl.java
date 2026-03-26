package com.generated.temporal.workflow;

import java.time.Duration;
import io.temporal.activity.ActivityOptions;
import io.temporal.common.RetryOptions;
import io.temporal.workflow.Workflow;
import com.generated.temporal.activities.GeneratedActivities;
import com.generated.temporal.model.*;

public class GeneratedWorkflowImpl implements GeneratedWorkflow {
    private final GeneratedActivities activities = Workflow.newActivityStub(
        GeneratedActivities.class,
        ActivityOptions.newBuilder()
            .setStartToCloseTimeout(Duration.ofSeconds(30))
            .setRetryOptions(RetryOptions.newBuilder().setMaximumAttempts(3).build())
            .build()
    );

    @Override
    public DocumentIngestResult run(DocumentIngestInput input) {
        var file_type_result = activities.determineFileType(new DetermineFileTypeInput(input.filePath()));
        var convert_result = activities.convertToText(new ConvertToTextInput(input.filePath(), file_type_result.fileType()));
        var save_result = activities.saveConvertedText(new SaveConvertedTextInput(input.filePath(), input.storageType(), convert_result.text()));
        return new DocumentIngestResult(file_type_result.fileType(),
            file_type_result.mimeType(),
            convert_result.text().substring(0, Math.min(120, convert_result.text().length())),
            save_result.savedTo(),
            save_result.recordId()
        );
    }
}
