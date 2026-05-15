from __future__ import annotations

from datetime import date as DateType

from sqlalchemy import func
from sqlmodel import Session, select

from daily_tasks.models import DailyTicket, DailyTicketTask, DailyTicketTaskSubtask, Project, ProjectDefaultSubtask


def add_project(
    session: Session,
    name: str,
    default_subtask_titles: list[str],
    *,
    description: str = "",
) -> Project:
    project = Project(name=name.strip(), description=description.strip())
    session.add(project)
    session.flush()
    for i, title in enumerate(default_subtask_titles):
        t = title.strip()
        if not t:
            continue
        session.add(ProjectDefaultSubtask(project_id=project.id, title=t, sort_order=i))
    session.commit()
    session.refresh(project)
    return project


def list_projects(session: Session) -> list[Project]:
    return list(session.exec(select(Project).order_by(Project.name)))


def get_project_by_id(session: Session, project_id: int) -> Project | None:
    return session.get(Project, project_id)


def list_default_subtask_titles(session: Session, project_id: int) -> list[str]:
    rows = session.exec(
        select(ProjectDefaultSubtask)
        .where(ProjectDefaultSubtask.project_id == project_id)
        .order_by(ProjectDefaultSubtask.sort_order, ProjectDefaultSubtask.id)
    ).all()
    return [r.title for r in rows]


def list_past_subtask_names_for_project(
    session: Session,
    project_id: int,
    *,
    before_date: DateType,
) -> list[str]:
    """Distinct subtask names from tickets strictly before ``before_date``, newest day first."""
    stmt = (
        select(DailyTicketTaskSubtask.name, func.max(DailyTicket.entry_date).label("last_day"))
        .select_from(DailyTicketTaskSubtask)
        .join(DailyTicketTask, DailyTicketTask.id == DailyTicketTaskSubtask.ticket_daily_task_id)
        .join(DailyTicket, DailyTicket.id == DailyTicketTask.ticket_daily_id)
        .where(
            DailyTicketTask.project_id == project_id,
            DailyTicket.entry_date < before_date,
        )
        .group_by(DailyTicketTaskSubtask.name)
        .order_by(func.max(DailyTicket.entry_date).desc(), DailyTicketTaskSubtask.name)
    )
    rows = session.exec(stmt).all()
    return [str(row[0]) for row in rows]


def get_project_by_name(session: Session, name: str) -> Project | None:
    key = name.strip()
    return session.exec(select(Project).where(Project.name == key)).first()


def set_default_subtasks(session: Session, project_name: str, titles: list[str]) -> Project:
    proj = get_project_by_name(session, project_name)
    if proj is None:
        raise KeyError(project_name)
    for row in session.exec(
        select(ProjectDefaultSubtask).where(ProjectDefaultSubtask.project_id == proj.id)
    ).all():
        session.delete(row)
    session.flush()
    order = 0
    for raw in titles:
        t = raw.strip()
        if not t:
            continue
        session.add(ProjectDefaultSubtask(project_id=proj.id, title=t, sort_order=order))
        order += 1
    session.commit()
    session.refresh(proj)
    return proj


def set_project_description(session: Session, project_name: str, description: str) -> Project:
    proj = get_project_by_name(session, project_name)
    if proj is None:
        raise KeyError(project_name)
    proj.description = description.strip()
    session.add(proj)
    session.commit()
    session.refresh(proj)
    return proj
