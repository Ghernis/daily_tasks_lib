from __future__ import annotations

from datetime import date, datetime, timezone

import questionary
import typer
from questionary import Choice

from daily_tasks.db import session_scope
from daily_tasks.services import projects as projects_svc
from daily_tasks.services import tickets as tickets_svc
from daily_tasks.services.profile import default_ticket_description


def _parse_hours(raw: str) -> float:
    text = raw.strip().replace(",", ".")
    if not text:
        raise ValueError("empty")
    return float(text)


def _parse_subtask_lines(raw: str) -> list[str]:
    parts: list[str] = []
    for line in raw.replace(",", "\n").splitlines():
        t = line.strip()
        if t:
            parts.append(t)
    return parts


def _merge_subtask_checklist_rows(defaults: list[str], from_past_days: list[str]) -> tuple[list[str], set[str]]:
    """Build ordered checkbox rows: project defaults first, then prior-day-only names (recency order).

    Returns (ordered_labels, labels_that_should_start_checked) — only project defaults are pre-checked.
    """
    default_set = set(defaults)
    ordered: list[str] = []
    seen: set[str] = set()
    for d in defaults:
        if d not in seen:
            seen.add(d)
            ordered.append(d)
    for p in from_past_days:
        if p not in seen:
            seen.add(p)
            ordered.append(p)
    return ordered, default_set


def _pick_subtasks_for_project(
    proj_name: str,
    defaults: list[str],
    from_past_days: list[str],
) -> list[str] | None:
    """Return ordered subtask names for this day, or None if the user aborted."""
    rows, default_set = _merge_subtask_checklist_rows(defaults, from_past_days)

    if rows:
        n_def = len([x for x in rows if x in default_set])
        n_past = len(rows) - n_def
        hint = f"{n_def} default(s)"
        if n_past:
            hint += f", {n_past} from earlier day(s) (unchecked — tick any you need again)"
        choices = [
            Choice(title=label, value=label, checked=(label in default_set)) for label in rows
        ]
        picked = questionary.checkbox(
            f"Subtasks for {proj_name!r} — {hint}. Toggle lines for today.",
            choices=choices,
        ).ask()
        if picked is None:
            return None
        selected_set = set(picked)
        ordered = [label for label in rows if label in selected_set]

        extra_raw = questionary.text(
            "Extra subtasks for today only (comma-separated or one per line; empty to skip).",
            default="",
        ).ask()
        if extra_raw is None:
            return None
        extras = _parse_subtask_lines(extra_raw)
        seen2 = set(ordered)
        for line in extras:
            if line not in seen2:
                seen2.add(line)
                ordered.append(line)
        return ordered

    custom = questionary.text(
        f"No saved defaults or prior-day subtasks for {proj_name!r}. Enter subtasks "
        "(comma-separated or one per line). Empty to skip.",
        default="",
    ).ask()
    if custom is None:
        return None
    return _parse_subtask_lines(custom)


def run_log_wizard(engine, entry_date: date | None) -> None:
    log_date = entry_date or datetime.now(timezone.utc).date()

    with session_scope(engine) as session:
        projects = projects_svc.list_projects(session)

    if not projects:
        typer.secho("No projects yet. Use: daily-tasks project add ...", fg=typer.colors.YELLOW, err=True)
        raise typer.Exit(code=1)

    choices = [Choice(title=p.name, value=p.id) for p in projects]
    selected_ids = questionary.checkbox(
        "Select one or more projects for this day (one ticket will be built for all of them)",
        choices=choices,
    ).ask()

    if selected_ids is None:
        raise typer.Abort()
    if not selected_ids:
        typer.secho("No projects selected.", fg=typer.colors.YELLOW)
        raise typer.Exit(code=0)

    default_title = f"Tareas diarias {log_date.isoformat()}"
    title_raw = questionary.text(
        "Ticket title for this day",
        default=default_title,
        validate=lambda t: True if t.strip() else "Title cannot be empty.",
    ).ask()
    if title_raw is None:
        raise typer.Abort()
    ticket_title = title_raw.strip() or default_title

    with session_scope(engine) as session:
        ticket_desc_default = default_ticket_description(session, log_date.isoformat())
    desc_raw = questionary.text(
        "Ticket description (optional)",
        default=ticket_desc_default,
    ).ask()
    if desc_raw is None:
        raise typer.Abort()
    ticket_description = (desc_raw or "").strip()

    tasks_payload: list[tuple[int, float, str, list[str]]] = []

    for pid in selected_ids:
        proj = None
        defaults: list[str] = []
        from_past: list[str] = []
        with session_scope(engine) as session:
            proj = projects_svc.get_project_by_id(session, int(pid))
            if proj is not None:
                defaults = projects_svc.list_default_subtask_titles(session, int(pid))
                from_past = projects_svc.list_past_subtask_names_for_project(
                    session, int(pid), before_date=log_date
                )
        if proj is None:
            continue

        def _valid_hours(t: str) -> bool | str:
            if not t.strip():
                return "Enter hours (e.g. 1, 0.5, 2.25)."
            try:
                v = _parse_hours(t)
            except ValueError:
                return "Must be a number (hours)."
            if v < 0:
                return "Hours cannot be negative."
            return True

        hours_raw = questionary.text(
            f"Hours spent on {proj.name!r} on {log_date.isoformat()}",
            validate=_valid_hours,
        ).ask()
        if hours_raw is None:
            raise typer.Abort()
        hours = _parse_hours(hours_raw)

        task_description = (proj.description or "").strip()
        if task_description:
            typer.secho(f"  Task description: {task_description[:80]}{'…' if len(task_description) > 80 else ''}", fg=typer.colors.BRIGHT_BLACK)

        subtask_names = _pick_subtasks_for_project(proj.name, defaults, from_past)
        if subtask_names is None:
            raise typer.Abort()
        rows, _ = _merge_subtask_checklist_rows(defaults, from_past)
        if not rows and not subtask_names:
            typer.secho(
                f"No defaults or prior subtasks for {proj.name!r}; none typed — empty list.",
                fg=typer.colors.YELLOW,
            )

        tasks_payload.append((int(pid), hours, task_description, subtask_names))
        typer.secho(
            f"Queued {proj.name!r} ({hours} h, {len(subtask_names)} subtasks).",
            fg=typer.colors.CYAN,
        )

    with session_scope(engine) as session:
        tickets_svc.save_daily_ticket(
            session,
            entry_date=log_date,
            title=ticket_title,
            description=ticket_description,
            tasks=tasks_payload,
        )

    typer.secho(
        f"Saved daily ticket for {log_date.isoformat()} with {len(tasks_payload)} task(s).",
        fg=typer.colors.GREEN,
    )
