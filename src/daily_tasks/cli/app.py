from __future__ import annotations

from datetime import date
from pathlib import Path
from typing import Annotated, Optional

import typer

from daily_tasks.cli import export as export_cli
from daily_tasks.cli import log as log_cli
from daily_tasks.cli.config import config_app
from daily_tasks.cli.project import project_app
from daily_tasks.cli.ticket import ticket_app
from daily_tasks.db import create_engine_for_url, default_sqlite_url, init_db, sqlite_url_from_path

app = typer.Typer(help="Log daily time per project with subtasks and JSON export.")
app.add_typer(project_app, name="project")
app.add_typer(ticket_app, name="ticket")
app.add_typer(config_app, name="config")


@app.callback()
def main_callback(
    ctx: typer.Context,
    db: Annotated[
        Optional[Path],
        typer.Option(
            "--db",
            help=(
                "SQLite database file. Default: %%LOCALAPPDATA%%\\daily_tasks\\data.db on Windows, "
                "else ~/.daily_tasks/data.db."
            ),
        ),
    ] = None,
) -> None:
    if db is None:
        url = default_sqlite_url()
    else:
        url = sqlite_url_from_path(db)
    engine = create_engine_for_url(url)
    init_db(engine)
    ctx.ensure_object(dict)
    ctx.obj["engine"] = engine


def _parse_iso_date(name: str, raw: str | None) -> date | None:
    if raw is None:
        return None
    try:
        return date.fromisoformat(raw)
    except ValueError as exc:
        raise typer.BadParameter(f"{name} must be ISO YYYY-MM-DD (got {raw!r}).") from exc


@app.command("log")
def log_command(
    ctx: typer.Context,
    entry_date: Annotated[
        Optional[str],
        typer.Option("--date", help="Log date as YYYY-MM-DD (default: today, UTC)."),
    ] = None,
) -> None:
    log_cli.run_log_wizard(ctx.obj["engine"], _parse_iso_date("--date", entry_date))


@app.command("export")
def export_command(
    ctx: typer.Context,
    output: Annotated[
        Optional[Path],
        typer.Option("--output", "-o", help="Output JSON file. Default: export-<today>.json in cwd."),
    ] = None,
    date_from: Annotated[
        Optional[str],
        typer.Option("--from", help="Include entries on or after this date (YYYY-MM-DD, inclusive)."),
    ] = None,
    date_to: Annotated[
        Optional[str],
        typer.Option("--to", help="Include entries on or before this date (YYYY-MM-DD, inclusive)."),
    ] = None,
) -> None:
    export_cli.run_export(
        ctx.obj["engine"],
        output=output,
        date_from=_parse_iso_date("--from", date_from),
        date_to=_parse_iso_date("--to", date_to),
    )
