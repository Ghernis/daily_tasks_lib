from __future__ import annotations

from datetime import date as DateType
from datetime import datetime
from typing import Optional

from sqlalchemy import Column, Date, Float, Integer, String, UniqueConstraint
from sqlmodel import Field, SQLModel


class Project(SQLModel, table=True):
    __tablename__ = "project"

    id: Optional[int] = Field(default=None, primary_key=True)
    name: str = Field(sa_column=Column(String(255), unique=True, nullable=False))
    description: str = Field(default="", sa_column=Column(String(4000), nullable=False))
    created_at: datetime = Field(default_factory=datetime.now)


class AuthorProfile(SQLModel, table=True):
    """Single global author row (``id`` is always 1) for ticket export."""

    __tablename__ = "author_profile"

    id: int = Field(default=1, primary_key=True)
    name: str = Field(sa_column=Column(String(255), nullable=False))
    email: str = Field(default="", sa_column=Column(String(255), nullable=False))
    sn_id: str = Field(default="", sa_column=Column(String(64), nullable=False))


class ProjectDefaultSubtask(SQLModel, table=True):
    __tablename__ = "project_default_subtask"

    id: Optional[int] = Field(default=None, primary_key=True)
    project_id: int = Field(foreign_key="project.id", nullable=False)
    title: str = Field(sa_column=Column(String(500), nullable=False))
    sort_order: int = Field(default=0, sa_column=Column(Integer, nullable=False))


class DailyTicket(SQLModel, table=True):
    """One row per calendar day (the container for that day's tasks)."""

    __tablename__ = "ticket_daily"
    __table_args__ = (UniqueConstraint("date", name="uq_ticket_daily_date"),)

    id: Optional[int] = Field(default=None, primary_key=True)
    entry_date: DateType = Field(sa_column=Column("date", Date(), nullable=False))
    title: str = Field(sa_column=Column(String(500), nullable=False))
    description: str = Field(default="", sa_column=Column(String(4000), nullable=False))
    ritm_number: Optional[str] = Field(default=None, sa_column=Column(String(128), nullable=True))
    created_at: datetime = Field(default_factory=datetime.now)


class DailyTicketTask(SQLModel, table=True):
    __tablename__ = "ticket_daily_task"
    __table_args__ = (UniqueConstraint("ticket_daily_id", "project_id", name="uq_ticket_task_project"),)

    id: Optional[int] = Field(default=None, primary_key=True)
    ticket_daily_id: int = Field(foreign_key="ticket_daily.id", nullable=False)
    project_id: int = Field(foreign_key="project.id", nullable=False)
    hours: float = Field(sa_column=Column(Float, nullable=False))
    task_description: str = Field(default="", sa_column=Column(String(4000), nullable=False))
    sort_order: int = Field(default=0, sa_column=Column(Integer, nullable=False))


class DailyTicketTaskSubtask(SQLModel, table=True):
    __tablename__ = "ticket_daily_task_subtask"

    id: Optional[int] = Field(default=None, primary_key=True)
    ticket_daily_task_id: int = Field(foreign_key="ticket_daily_task.id", nullable=False)
    name: str = Field(sa_column=Column(String(500), nullable=False))
    sort_order: int = Field(default=0, sa_column=Column(Integer, nullable=False))
