package com.generated.temporal.activities;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import com.generated.temporal.model.*;

public class GeneratedActivitiesImpl implements GeneratedActivities {
    private static final Logger log = LoggerFactory.getLogger(GeneratedActivitiesImpl.class);

    @Override
    public DetermineFileTypeOutput determineFileType(DetermineFileTypeInput input) {
        log.info("Running activity: determine_file_type");
        // TODO: add real processor implementation here.
        return new DetermineFileTypeOutput("static-file_type", "static-mime_type");
    }

    @Override
    public ConvertToTextOutput convertToText(ConvertToTextInput input) {
        log.info("Running activity: convert_to_text");
        // TODO: add real processor implementation here.
        return new ConvertToTextOutput("static-text", 1);
    }

    @Override
    public SaveConvertedTextOutput saveConvertedText(SaveConvertedTextInput input) {
        log.info("Running activity: save_converted_text");
        // TODO: add real processor implementation here.
        return new SaveConvertedTextOutput("static-saved_to", "static-record_id");
    }

}
