from __future__ import annotations

from datetime import date

import typer

from daily_tasks.db import session_scope
from daily_tasks.services import tickets as tickets_svc

ticket_app = typer.Typer(help="After a ticket is created, store identifiers on the daily ticket.")


def _parse_date_arg(raw: str) -> date:
    try:
        return date.fromisoformat(raw.strip())
    except ValueError as exc:
        raise typer.BadParameter(f"DATE must be YYYY-MM-DD (got {raw!r}).") from exc


@ticket_app.command("set-ritm")
def ticket_set_ritm(
    ctx: typer.Context,
    entry_date: str = typer.Argument(..., metavar="DATE", help="Ticket day as YYYY-MM-DD."),
    ritm: str = typer.Argument(..., help="RITM / reference number returned by the ticketing system."),
) -> None:
    engine = ctx.obj["engine"]
    d = _parse_date_arg(entry_date)
    with session_scope(engine) as session:
        updated = tickets_svc.set_ritm_for_date(session, entry_date=d, ritm_number=ritm)
    if updated is None:
        typer.secho(f"No ticket found for {d.isoformat()}. Run `log` for that day first.", fg=typer.colors.RED, err=True)
        raise typer.Exit(code=1)
    typer.secho(f"Stored RITM {ritm!r} on {d.isoformat()}.", fg=typer.colors.GREEN)
