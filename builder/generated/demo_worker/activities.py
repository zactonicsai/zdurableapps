from __future__ import annotations

from pathlib import Path

from temporalio import activity

from logging_config import get_logger
from models import *

logger = get_logger(__name__)


@activity.defn
async def review_document(input_data: DocumentRef) -> ReviewResult:
    logger.info(
        "activity.start",
        activity="review_document",
        input_type=type(input_data).__name__,
    )
    # TODO: Replace this static stub with the real processor implementation.
    # Add your file parser, OCR, persistence, or external service code here.
    return ReviewResult(data="static-string")

@activity.defn
async def fill_template(input_data: TemplateData) -> Document:
    logger.info(
        "activity.start",
        activity="fill_template",
        input_type=type(input_data).__name__,
    )
    # TODO: Replace this static stub with the real processor implementation.
    # Add your file parser, OCR, persistence, or external service code here.
    return Document(data="static-string")

