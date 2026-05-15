from __future__ import annotations

import typer
from rich.table import Table
from rich.console import Console
from sqlalchemy.exc import IntegrityError

from daily_tasks.db import session_scope
from daily_tasks.services import projects as projects_svc

project_app = typer.Typer(help="Create and list projects.")


@project_app.command("add")
def project_add(
    ctx: typer.Context,
    name: str = typer.Argument(..., help="Project name (unique)."),
    description: str = typer.Option(
        "",
        "--description",
        "-d",
        help="Fixed task description for this project (copied on each log).",
    ),
    subtask: list[str] = typer.Option(
        [],
        "--subtask",
        "-s",
        help="Default subtask line for this project (repeatable).",
    ),
) -> None:
    engine = ctx.obj["engine"]
    try:
        with session_scope(engine) as session:
            _ = projects_svc.add_project(
                session, name=name, default_subtask_titles=subtask, description=description
            )
    except IntegrityError:
        typer.secho(f"Project name already exists: {name!r}", fg=typer.colors.RED, err=True)
        raise typer.Exit(code=1) from None
    typer.secho(f"Created project {name!r}.", fg=typer.colors.GREEN)


@project_app.command("list")
def project_list(ctx: typer.Context) -> None:
    engine = ctx.obj["engine"]
    with session_scope(engine) as session:
        rows = projects_svc.list_projects(session)
    table = Table("id", "name", "description", "created_at")
    for p in rows:
        desc = (p.description or "").strip()
        if len(desc) > 48:
            desc = desc[:45] + "..."
        table.add_row(str(p.id), p.name, desc, p.created_at.isoformat(timespec="seconds"))
    Console().print(table)


@project_app.command("set-description")
def project_set_description(
    ctx: typer.Context,
    name: str = typer.Argument(..., help="Existing project name (exact match)."),
    description: str = typer.Option(
        ...,
        "--description",
        "-d",
        help="Fixed task description for this project (used automatically when logging).",
    ),
) -> None:
    engine = ctx.obj["engine"]
    try:
        with session_scope(engine) as session:
            projects_svc.set_project_description(session, name, description)
    except KeyError:
        typer.secho(f"Unknown project: {name!r}", fg=typer.colors.RED, err=True)
        raise typer.Exit(code=1) from None
    typer.secho(f"Updated description for {name!r}.", fg=typer.colors.GREEN)


@project_app.command("set-default-subtasks")
def project_set_default_subtasks(
    ctx: typer.Context,
    name: str = typer.Argument(..., help="Existing project name (exact match)."),
    subtask: list[str] = typer.Option(
        [],
        "--subtask",
        "-s",
        help="Default subtask line (repeatable). Replaces current defaults; pass none to clear.",
    ),
) -> None:
    engine = ctx.obj["engine"]
    try:
        with session_scope(engine) as session:
            projects_svc.set_default_subtasks(session, name, subtask)
    except KeyError:
        typer.secho(f"Unknown project: {name!r}", fg=typer.colors.RED, err=True)
        raise typer.Exit(code=1) from None
    typer.secho(f"Updated default subtasks for {name!r}.", fg=typer.colors.GREEN)
