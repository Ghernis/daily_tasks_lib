from __future__ import annotations

import os
from contextlib import contextmanager
from pathlib import Path
from typing import Iterator

from sqlalchemy import event, inspect, text
from sqlalchemy.engine import Engine
from sqlmodel import Session, SQLModel, create_engine


def default_sqlite_url() -> str:
    local = os.environ.get("LOCALAPPDATA")
    if local:
        path = Path(local) / "daily_tasks" / "data.db"
    else:
        path = Path.home() / ".daily_tasks" / "data.db"
    path.parent.mkdir(parents=True, exist_ok=True)
    return sqlite_url_from_path(path)


def sqlite_url_from_path(path: Path) -> str:
    resolved = path.expanduser().resolve()
    resolved.parent.mkdir(parents=True, exist_ok=True)
    return f"sqlite:///{resolved.as_posix()}"


def create_engine_for_url(database_url: str) -> Engine:
    connect_args = {"check_same_thread": False} if database_url.startswith("sqlite") else {}
    engine = create_engine(database_url, echo=False, connect_args=connect_args)

    if database_url.startswith("sqlite"):

        @event.listens_for(engine, "connect")
        def _sqlite_pragma(dbapi_conn, _connection_record) -> None:  # type: ignore[no-untyped-def]
            cur = dbapi_conn.cursor()
            cur.execute("PRAGMA foreign_keys=ON")
            cur.close()

    return engine


def _migrate_sqlite(engine: Engine) -> None:
    if not str(engine.url).startswith("sqlite"):
        return
    insp = inspect(engine)
    tables = set(insp.get_table_names())
    if "project" in tables:
        cols = {c["name"] for c in insp.get_columns("project")}
        if "description" not in cols:
            with engine.begin() as conn:
                conn.execute(
                    text(
                        "ALTER TABLE project ADD COLUMN description VARCHAR(4000) "
                        "NOT NULL DEFAULT ''"
                    )
                )


def init_db(engine: Engine) -> None:
    SQLModel.metadata.create_all(engine)
    _migrate_sqlite(engine)


@contextmanager
def session_scope(engine: Engine) -> Iterator[Session]:
    with Session(engine) as session:
        yield session
