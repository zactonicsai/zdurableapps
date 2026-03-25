from __future__ import annotations

import asyncio

from temporalio.client import Client
from temporalio.worker import Worker

import activities
from workflows import *


async def main() -> None:
    client = await Client.connect("localhost:7233", namespace="default")
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
    )
    await worker.run()


if __name__ == "__main__":
    asyncio.run(main())
