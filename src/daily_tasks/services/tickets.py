from __future__ import annotations

from datetime import date as DateType

from sqlmodel import Session, select

from daily_tasks.models import DailyTicket, DailyTicketTask, DailyTicketTaskSubtask, Project


def save_daily_ticket(
    session: Session,
    *,
    entry_date: DateType,
    title: str,
    description: str,
    tasks: list[tuple[int, float, str, list[str]]],
) -> DailyTicket:
    """Replace all tasks for the given day with a single ticket snapshot.

    Each task tuple is: (project_id, hours, task_description, subtask_names).
    Preserves ``ritm_number`` if the ticket row already exists.
    """
    existing = session.exec(select(DailyTicket).where(DailyTicket.entry_date == entry_date)).first()

    if existing:
        ticket = existing
        ticket.title = title.strip() or ticket.title
        ticket.description = description.strip()
        session.add(ticket)
        session.flush()
        for row in session.exec(
            select(DailyTicketTask).where(DailyTicketTask.ticket_daily_id == ticket.id)
        ).all():
            for sub in session.exec(
                select(DailyTicketTaskSubtask).where(
                    DailyTicketTaskSubtask.ticket_daily_task_id == row.id,
                )
            ).all():
                session.delete(sub)
            session.flush()
            session.delete(row)
        session.flush()
    else:
        ticket = DailyTicket(
            entry_date=entry_date,
            title=title.strip() or f"Daily work {entry_date.isoformat()}",
            description=description.strip(),
        )
        session.add(ticket)
        session.flush()

    for order, (project_id, hours, task_description, subtask_names) in enumerate(tasks):
        if session.get(Project, project_id) is None:
            continue
        task_row = DailyTicketTask(
            ticket_daily_id=ticket.id,
            project_id=project_id,
            hours=hours,
            task_description=task_description.strip(),
            sort_order=order,
        )
        session.add(task_row)
        session.flush()
        for i, raw in enumerate(subtask_names):
            name = raw.strip()
            if not name:
                continue
            session.add(
                DailyTicketTaskSubtask(
                    ticket_daily_task_id=task_row.id,
                    name=name,
                    sort_order=i,
                )
            )

    session.commit()
    session.refresh(ticket)
    return ticket


def set_ritm_for_date(session: Session, *, entry_date: DateType, ritm_number: str) -> DailyTicket | None:
    ticket = session.exec(select(DailyTicket).where(DailyTicket.entry_date == entry_date)).first()
    if ticket is None:
        return None
    ticket.ritm_number = ritm_number.strip() or None
    session.add(ticket)
    session.commit()
    session.refresh(ticket)
    return ticket
