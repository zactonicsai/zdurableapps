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
    logger.info("worker.start", task_queue="document-ingest-queue", target=target, namespace=namespace)
    worker = Worker(
        client,
        task_queue="document-ingest-queue",
        workflows=[
            DocumentIngestWorkflow,
        ],
        activities=[
            activities.determine_file_type,
            activities.convert_to_text,
            activities.save_converted_text,
        ],
        max_concurrent_activities=100,
        max_concurrent_workflow_tasks=100,
    )
    await worker.run()


if __name__ == "__main__":
    asyncio.run(main())
