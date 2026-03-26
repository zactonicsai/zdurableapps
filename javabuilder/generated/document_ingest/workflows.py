from __future__ import annotations

from datetime import timedelta

from temporalio import workflow
from temporalio.common import RetryPolicy

with workflow.unsafe.imports_passed_through():
    import activities
    from logging_config import get_logger
    from models import *

logger = get_logger(__name__)


@workflow.defn
class DocumentIngestWorkflow:
    @workflow.run
    async def run(self, input_data: DocumentIngestInput) -> DocumentIngestResult:
        logger.info("workflow.start", workflow="DocumentIngestWorkflow")
        file_type_result = await workflow.execute_activity(
            activities.determine_file_type,
            DetermineFileTypeInput(file_path=input_data.file_path),
            start_to_close_timeout=timedelta(seconds=30),
            schedule_to_close_timeout=timedelta(seconds=60),
            retry_policy=RetryPolicy(initial_interval=timedelta(seconds=1), backoff_coefficient=2, maximum_interval=timedelta(seconds=30), maximum_attempts=3, non_retryable_error_types=["ValueError"]),
        )

        convert_result = await workflow.execute_activity(
            activities.convert_to_text,
            ConvertToTextInput(file_path=input_data.file_path, file_type=file_type_result.file_type),
            start_to_close_timeout=timedelta(seconds=60),
            schedule_to_close_timeout=timedelta(seconds=120),
            retry_policy=RetryPolicy(initial_interval=timedelta(seconds=1), backoff_coefficient=2, maximum_interval=timedelta(seconds=30), maximum_attempts=3, non_retryable_error_types=["ValueError"]),
        )

        save_result = await workflow.execute_activity(
            activities.save_converted_text,
            SaveConvertedTextInput(file_path=input_data.file_path, storage_type=input_data.storage_type, text=convert_result.text),
            start_to_close_timeout=timedelta(seconds=60),
            schedule_to_close_timeout=timedelta(seconds=120),
            retry_policy=RetryPolicy(initial_interval=timedelta(seconds=1), backoff_coefficient=2, maximum_interval=timedelta(seconds=30), maximum_attempts=3, non_retryable_error_types=["ValueError"]),
        )

        result = DocumentIngestResult(file_type=file_type_result.file_type, mime_type=file_type_result.mime_type, text_preview=convert_result.text[:120], saved_to=save_result.saved_to, record_id=save_result.record_id)
        logger.info("workflow.complete", workflow="DocumentIngestWorkflow")
        return result


