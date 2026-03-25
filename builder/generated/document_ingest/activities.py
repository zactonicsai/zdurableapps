from __future__ import annotations

from temporalio import activity

from models import *


@activity.defn
async def determine_file_type(input_data: DetermineFileTypeInput) -> DetermineFileTypeOutput:
    # TODO: Replace the static return below with the real processor implementation.
    # Example real logic: sniff MIME type, convert to text, save to DB/shared FS/S3, etc.
    activity.logger.info("running activity", extra={"activity": "determine_file_type"})
    return DetermineFileTypeOutput(file_type="pdf", mime_type="application/pdf")


@activity.defn
async def convert_to_text(input_data: ConvertToTextInput) -> ConvertToTextOutput:
    # TODO: Replace the static return below with the real processor implementation.
    # Example real logic: sniff MIME type, convert to text, save to DB/shared FS/S3, etc.
    activity.logger.info("running activity", extra={"activity": "convert_to_text"})
    return ConvertToTextOutput(text="static extracted text", page_count=1)


@activity.defn
async def save_converted_text(input_data: SaveConvertedTextInput) -> SaveConvertedTextOutput:
    # TODO: Replace the static return below with the real processor implementation.
    # Example real logic: sniff MIME type, convert to text, save to DB/shared FS/S3, etc.
    activity.logger.info("running activity", extra={"activity": "save_converted_text"})
    return SaveConvertedTextOutput(saved_to="s3://demo-bucket/static-output.txt", record_id="record-123")


