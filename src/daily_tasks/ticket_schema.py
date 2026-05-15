"""Pydantic shapes for JSON export / downstream ticket creation."""

from __future__ import annotations

from typing import List

from pydantic import BaseModel, Field


class Author(BaseModel):
    name: str
    email: str = ""
    sn_id: str = ""


class Subtask(BaseModel):
    name: str = ""


class Task(BaseModel):
    name: str
    description: str = ""
    subtasks: List[Subtask] = Field(default_factory=list)
    time: float = 0.0  # hours (supports fractional, e.g. 0.5)


class TicketDaily(BaseModel):
    title: str
    date: str
    description: str = ""
    author: List[Author] = Field(default_factory=list)
    tasks: List[Task] = Field(default_factory=list)
    ritm_number: str | None = None
