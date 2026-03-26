from __future__ import annotations

from dataclasses import dataclass
from enum import Enum


class FileStorageType(str, Enum):
    SHARED_FS = "shared_fs"
    DATABASE = "database"
    S3 = "s3"


@dataclass
class DocumentIngestInput:
    file_path: str
    storage_type: FileStorageType


@dataclass
class DetermineFileTypeInput:
    file_path: str


@dataclass
class DetermineFileTypeOutput:
    file_type: str
    mime_type: str


@dataclass
class ConvertToTextInput:
    file_path: str
    file_type: str


@dataclass
class ConvertToTextOutput:
    text: str
    page_count: int


@dataclass
class SaveConvertedTextInput:
    file_path: str
    storage_type: FileStorageType
    text: str


@dataclass
class SaveConvertedTextOutput:
    saved_to: str
    record_id: str


@dataclass
class DocumentIngestResult:
    file_type: str
    mime_type: str
    text_preview: str
    saved_to: str
    record_id: str
