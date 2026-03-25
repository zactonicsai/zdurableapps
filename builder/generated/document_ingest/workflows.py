from __future__ import annotations

from datetime import timedelta

from temporalio import workflow
from temporalio.common import RetryPolicy

with workflow.unsafe.imports_passed_through():
    import activities
    from models import *


@workflow.defn
class DocumentIngestWorkflow:
    @workflow.run
    async def run(self, input_data: DocumentIngestInput) -> DocumentIngestResult:
        file_type_result = await workflow.execute_activity(
            activities.determine_file_type,
            DetermineFileTypeInput(file_path=input_data.file_path),
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(maximum_attempts=3),
        )

        convert_result = await workflow.execute_activity(
            activities.convert_to_text,
            ConvertToTextInput(file_path=input_data.file_path, file_type=file_type_result.file_type),
            start_to_close_timeout=timedelta(seconds=60),
            retry_policy=RetryPolicy(maximum_attempts=3),
        )

        save_result = await workflow.execute_activity(
            activities.save_converted_text,
            SaveConvertedTextInput(file_path=input_data.file_path, storage_type=input_data.storage_type, text=convert_result.text),
            start_to_close_timeout=timedelta(seconds=60),
            retry_policy=RetryPolicy(maximum_attempts=3),
        )

        return DocumentIngestResult(file_type=file_type_result.file_type, mime_type=file_type_result.mime_type, text_preview=convert_result.text[:120], saved_to=save_result.saved_to, record_id=save_result.record_id)


