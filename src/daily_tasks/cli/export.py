from __future__ import annotations

from datetime import date
from pathlib import Path
from typing import Optional

import typer

from daily_tasks.db import session_scope
from daily_tasks.services.export import build_export


def run_export(
    engine,
    *,
    output: Optional[Path],
    date_from: Optional[date],
    date_to: Optional[date],
) -> None:
    with session_scope(engine) as session:
        doc = build_export(session, date_from=date_from, date_to=date_to)
    text = doc.model_dump_json(indent=2)
    if output is None:
        default_name = f"export-{doc.exported_at.date().isoformat()}.json"
        out_path = Path(default_name)
        out_path.write_text(text, encoding="utf-8")
        typer.secho(f"Wrote {out_path.resolve()}", fg=typer.colors.GREEN)
    else:
        output.parent.mkdir(parents=True, exist_ok=True)
        output.write_text(text, encoding="utf-8")
        typer.secho(f"Wrote {output.resolve()}", fg=typer.colors.GREEN)
