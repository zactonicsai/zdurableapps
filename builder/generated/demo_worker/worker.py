from __future__ import annotations

import asyncio
import os

from dotenv import load_dotenv
from temporalio.client import Client
from temporalio.worker import Worker

import activities
from logging_config import configure_logging, get_logger
from workflows import *

load_dotenv()
configure_logging()
logger = get_logger(__name__)


async def main() -> None:
    target = os.getenv("TEMPORAL_TARGET", "localhost:7233")
    namespace = os.getenv("TEMPORAL_NAMESPACE", "default")
    client = await Client.connect(target, namespace=namespace)
    logger.info("worker.start", task_queue="my-workflow-project-queue", target=target, namespace=namespace)
    worker = Worker(
        client,
        task_queue="my-workflow-project-queue",
        workflows=[
            my_workflow_project_main_workflow,
        ],
        activities=[
            activities.review_document,
            activities.fill_template,
        ],
        max_concurrent_activities=200,
        max_concurrent_workflow_tasks=200,
    )
    await worker.run()


if __name__ == "__main__":
    asyncio.run(main())
