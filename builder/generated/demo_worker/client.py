from __future__ import annotations

import asyncio
import os

from dotenv import load_dotenv
from temporalio.client import Client

from logging_config import configure_logging, get_logger
from models import *
from workflows import my_workflow_project_main_workflow

load_dotenv()
configure_logging()
logger = get_logger(__name__)


async def main() -> None:
    target = os.getenv("TEMPORAL_TARGET", "localhost:7233")
    namespace = os.getenv("TEMPORAL_NAMESPACE", "default")
    client = await Client.connect(target, namespace=namespace)
    request = DocumentRef(data="static-string")
    handle = await client.start_workflow(
        my_workflow_project_main_workflow.run,
        request,
        id="document-ingest-sample",
        task_queue="my-workflow-project-queue",
    )
    logger.info("client.started", workflow_id=handle.id, run_id=handle.result_run_id)
    result = await handle.result()
    logger.info("client.result", result=result)
    print(result)


if __name__ == "__main__":
    asyncio.run(main())
