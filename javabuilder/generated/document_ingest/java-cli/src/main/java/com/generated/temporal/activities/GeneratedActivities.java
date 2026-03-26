package com.generated.temporal.activities;

import io.temporal.activity.ActivityInterface;
import io.temporal.activity.ActivityMethod;
import com.generated.temporal.model.*;

@ActivityInterface
public interface GeneratedActivities {
    @ActivityMethod(name = "determine_file_type")
    DetermineFileTypeOutput determineFileType(DetermineFileTypeInput input);

    @ActivityMethod(name = "convert_to_text")
    ConvertToTextOutput convertToText(ConvertToTextInput input);

    @ActivityMethod(name = "save_converted_text")
    SaveConvertedTextOutput saveConvertedText(SaveConvertedTextInput input);

}
