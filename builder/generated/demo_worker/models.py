from __future__ import annotations

from dataclasses import dataclass
from enum import Enum
from typing import Optional


@dataclass
class DocumentRef:
    data: str


@dataclass
class ReviewResult:
    data: str


@dataclass
class TemplateData:
    data: str


@dataclass
class Document:
    data: str


