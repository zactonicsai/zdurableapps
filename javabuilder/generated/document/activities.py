from __future__ import annotations

from pathlib import Path

from temporalio import activity

from logging_config import get_logger
from models import *

logger = get_logger(__name__)


@activity.defn
async def determine_file_type(input_data: DetermineFileTypeInput) -> DetermineFileTypeOutput:
    logger.info(
        "activity.start",
        activity="determine_file_type",
        input_type=type(input_data).__name__,
    )
    # TODO: Replace this static stub with the real processor implementation.
    # Add your file parser, OCR, persistence, or external service code here.
    suffix = Path(input_data.file_path).suffix.lower()
    file_type = "text"
    mime_type = "text/plain"
    if suffix == ".pdf":
        file_type = "pdf"
        mime_type = "application/pdf"
    elif suffix in {".png", ".jpg", ".jpeg", ".tif", ".tiff", ".bmp"}:
        file_type = "image"
        mime_type = "image/" + suffix.lstrip(".")
    elif suffix in {".doc", ".docx"}:
        file_type = "word"
        mime_type = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
    result = DetermineFileTypeOutput(file_type=file_type, mime_type=mime_type)
    logger.info("activity.complete", activity="determine_file_type", result=result)
    return result

@activity.defn
async def convert_to_text(input_data: ConvertToTextInput) -> ConvertToTextOutput:
    logger.info(
        "activity.start",
        activity="convert_to_text",
        input_type=type(input_data).__name__,
    )
    # TODO: Replace this static stub with the real processor implementation.
    # Add your file parser, OCR, persistence, or external service code here.
    text = f"static extracted text for {input_data.file_path} as {input_data.file_type}"
    result = ConvertToTextOutput(text=text, page_count=1)
    logger.info("activity.complete", activity="convert_to_text", page_count=result.page_count)
    return result

@activity.defn
async def save_converted_text(input_data: SaveConvertedTextInput) -> SaveConvertedTextOutput:
    logger.info(
        "activity.start",
        activity="save_converted_text",
        input_type=type(input_data).__name__,
    )
    # TODO: Replace this static stub with the real processor implementation.
    # Add your file parser, OCR, persistence, or external service code here.
    storage = getattr(input_data.storage_type, "value", str(input_data.storage_type))
    if storage == "s3":
        saved_to = "s3://demo-bucket/converted/output.txt"
    elif storage == "database":
        saved_to = "database://document_ingest/records/record-123"
    else:
        saved_to = "/shared/output/converted-output.txt"
    result = SaveConvertedTextOutput(saved_to=saved_to, record_id="record-123")
    logger.info("activity.complete", activity="save_converted_text", saved_to=result.saved_to)
    return result

