from __future__ import annotations

import typer
from rich.console import Console
from rich.table import Table

from daily_tasks.db import session_scope
from daily_tasks.services.profile import get_author_profile, set_author_profile

config_app = typer.Typer(help="Global settings (author profile for exports).")


@config_app.command("set-author")
def config_set_author(
    ctx: typer.Context,
    name: str = typer.Option(..., "--name", "-n", help="Your display name on tickets."),
    email: str = typer.Option("", "--email", "-e", help="Email address."),
    sn_id: str = typer.Option("", "--sn-id", help="ServiceNow / corporate ID (e.g. SE40634)."),
) -> None:
    engine = ctx.obj["engine"]
    try:
        with session_scope(engine) as session:
            set_author_profile(session, name=name, email=email, sn_id=sn_id)
    except ValueError as exc:
        typer.secho(str(exc), fg=typer.colors.RED, err=True)
        raise typer.Exit(code=1) from None
    typer.secho("Author profile saved.", fg=typer.colors.GREEN)


@config_app.command("show-author")
def config_show_author(ctx: typer.Context) -> None:
    engine = ctx.obj["engine"]
    with session_scope(engine) as session:
        row = get_author_profile(session)
    if row is None:
        typer.secho(
            "No author configured. Run: daily-tasks config set-author --name \"...\" [--email ...] [--sn-id ...]",
            fg=typer.colors.YELLOW,
        )
        raise typer.Exit(code=1)
    table = Table("field", "value")
    table.add_row("name", row.name)
    table.add_row("email", row.email or "")
    table.add_row("sn_id", row.sn_id or "")
    Console().print(table)
