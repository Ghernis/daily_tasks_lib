"""Ticket JSON models used by export and downstream ticket creation.

Prefer importing from the package when the project is installed::

    from daily_tasks.ticket_schema import TicketDaily, Task, Subtask, Author

This file re-exports the same symbols so existing ``from models_needed import ...`` paths keep working
when the repo root is on ``PYTHONPATH`` or you run from the project directory after ``pip install -e .``.
"""

from __future__ import annotations

try:
    from daily_tasks.ticket_schema import Author, Subtask, Task, TicketDaily
except ModuleNotFoundError:  # pragma: no cover - local script without install
    import sys
    from pathlib import Path

    _src = Path(__file__).resolve().parent / "src"
    if _src.is_dir():
        sys.path.insert(0, str(_src))
    from daily_tasks.ticket_schema import Author, Subtask, Task, TicketDaily

__all__ = ["Author", "Subtask", "Task", "TicketDaily"]
