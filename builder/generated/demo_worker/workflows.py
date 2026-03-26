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
class my_workflow_project_main_workflow:
    @workflow.run
    async def run(self, input_data: DocumentRef) -> Document:
        logger.info("workflow.start", workflow="my_workflow_project_main_workflow")
        review_document_result = await workflow.execute_activity(
            activities.review_document,
            input_data,
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(initial_interval=timedelta(seconds=1), backoff_coefficient=2, maximum_interval=timedelta(seconds=30), maximum_attempts=3),
        )

        fill_template_result = await workflow.execute_activity(
            activities.fill_template,
            input_data,
            start_to_close_timeout=timedelta(seconds=30),
            retry_policy=RetryPolicy(initial_interval=timedelta(seconds=1), backoff_coefficient=2, maximum_interval=timedelta(seconds=30), maximum_attempts=3),
        )

        result = Document(data="static-string")
        logger.info("workflow.complete", workflow="my_workflow_project_main_workflow")
        return result


