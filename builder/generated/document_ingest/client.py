from __future__ import annotations

import asyncio
import sys
from uuid import uuid4

from temporalio.client import Client

from models import *
from workflows import DocumentIngestWorkflow


def build_input() -> DocumentIngestInput:
    file_path = sys.argv[1] if len(sys.argv) > 1 else "./sample-input/demo.pdf"
    storage_arg = sys.argv[2] if len(sys.argv) > 2 else "s3"
    return DocumentIngestInput(file_path=file_path, storage_type=FileStorageType(storage_arg))


async def main() -> None:
    client = await Client.connect("localhost:7233", namespace="default")
    result = await client.execute_workflow(
        DocumentIngestWorkflow.run,
        build_input(),
        id=f"document-ingest-workflow-{uuid4()}",
        task_queue="document-ingest-queue",
    )
    print(result)


if __name__ == "__main__":
    asyncio.run(main())
