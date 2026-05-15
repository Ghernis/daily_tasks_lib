from __future__ import annotations

from datetime import date as DateType
from datetime import datetime, timezone

from sqlmodel import Session, select

from daily_tasks.models import DailyTicket, DailyTicketTask, DailyTicketTaskSubtask, Project
from daily_tasks.schemas import ExportDocument
from daily_tasks.services.profile import get_author_for_export
from daily_tasks.ticket_schema import Subtask, Task, TicketDaily


def build_export(
    session: Session,
    *,
    date_from: DateType | None = None,
    date_to: DateType | None = None,
) -> ExportDocument:
    stmt = select(DailyTicket).order_by(DailyTicket.entry_date)
    if date_from is not None:
        stmt = stmt.where(DailyTicket.entry_date >= date_from)
    if date_to is not None:
        stmt = stmt.where(DailyTicket.entry_date <= date_to)

    authors = get_author_for_export(session)
    tickets_out: list[TicketDaily] = []
    for ticket in session.exec(stmt).all():
        task_rows = session.exec(
            select(DailyTicketTask)
            .where(DailyTicketTask.ticket_daily_id == ticket.id)
            .order_by(DailyTicketTask.sort_order, DailyTicketTask.id)
        ).all()
        tasks: list[Task] = []
        for tr in task_rows:
            proj = session.get(Project, tr.project_id)
            if proj is None:
                continue
            subs = session.exec(
                select(DailyTicketTaskSubtask)
                .where(DailyTicketTaskSubtask.ticket_daily_task_id == tr.id)
                .order_by(DailyTicketTaskSubtask.sort_order, DailyTicketTaskSubtask.id)
            ).all()
            desc = tr.task_description.strip() or (proj.description or "").strip()
            tasks.append(
                Task(
                    name=proj.name,
                    description=desc,
                    subtasks=[Subtask(name=s.name) for s in subs],
                    time=float(tr.hours),
                )
            )
        tickets_out.append(
            TicketDaily(
                title=ticket.title,
                date=ticket.entry_date.isoformat(),
                description=ticket.description,
                author=list(authors),
                tasks=tasks,
                ritm_number=ticket.ritm_number,
            )
        )

    return ExportDocument(exported_at=datetime.now(timezone.utc), tickets=tickets_out)
