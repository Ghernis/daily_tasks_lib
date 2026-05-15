from __future__ import annotations

import json
from datetime import date

import pytest
from sqlmodel import select

from daily_tasks.db import create_engine_for_url, init_db, session_scope
from daily_tasks.models import DailyTicket
from daily_tasks.services.export import build_export
from daily_tasks.services.profile import set_author_profile
from daily_tasks.services.projects import (
    add_project,
    list_default_subtask_titles,
    list_past_subtask_names_for_project,
    set_default_subtasks,
    set_project_description,
)
from daily_tasks.services.tickets import save_daily_ticket, set_ritm_for_date


@pytest.fixture
def engine(tmp_path):
    url = f"sqlite:///{(tmp_path / 'test.db').as_posix()}"
    eng = create_engine_for_url(url)
    init_db(eng)
    return eng


def test_add_project_with_default_subtasks(engine) -> None:
    with session_scope(engine) as session:
        p = add_project(session, "Alpha", ["t1", "t2"])
        pid = p.id
    assert pid is not None
    with session_scope(engine) as session:
        titles = list_default_subtask_titles(session, int(pid))
    assert titles == ["t1", "t2"]


def test_past_subtask_names_exclude_same_day_and_order_by_recency(engine) -> None:
    with session_scope(engine) as session:
        p = add_project(session, "PastProj", [])
        pid = int(p.id)
    d1 = date(2026, 5, 1)
    d2 = date(2026, 5, 2)
    d3 = date(2026, 5, 5)
    with session_scope(engine) as session:
        save_daily_ticket(
            session,
            entry_date=d1,
            title="a",
            description="",
            tasks=[(pid, 1.0, "", ["alpha", "beta"])],
        )
        save_daily_ticket(
            session,
            entry_date=d2,
            title="b",
            description="",
            tasks=[(pid, 0.5, "", ["gamma"])],
        )
    with session_scope(engine) as session:
        before_d1 = list_past_subtask_names_for_project(session, pid, before_date=d1)
        before_d2 = list_past_subtask_names_for_project(session, pid, before_date=d2)
        before_d3 = list_past_subtask_names_for_project(session, pid, before_date=d3)
    assert before_d1 == []
    assert before_d2 == ["alpha", "beta"]
    assert before_d3 == ["gamma", "alpha", "beta"]


def test_set_default_subtasks_replaces(engine) -> None:
    with session_scope(engine) as session:
        p = add_project(session, "Zed", ["old"])
        pid = int(p.id)
    with session_scope(engine) as session:
        set_default_subtasks(session, "Zed", ["n1", "n2"])
    with session_scope(engine) as session:
        assert list_default_subtask_titles(session, pid) == ["n1", "n2"]


def test_export_includes_author_and_project_description(engine) -> None:
    with session_scope(engine) as session:
        set_author_profile(
            session,
            name="Hernan Gomez",
            email="hernan@example.com",
            sn_id="SE40634",
        )
        p = add_project(session, "Beta", [], description="Standing project scope")
        pid = int(p.id)
    with session_scope(engine) as session:
        save_daily_ticket(
            session,
            entry_date=date(2026, 5, 10),
            title="Day",
            description="ticket body",
            tasks=[(pid, 1.0, "Standing project scope", ["x"])],
        )
    with session_scope(engine) as session:
        doc = build_export(session)
    assert len(doc.tickets) == 1
    td = doc.tickets[0]
    assert len(td.author) == 1
    assert td.author[0].name == "Hernan Gomez"
    assert td.author[0].email == "hernan@example.com"
    assert td.author[0].sn_id == "SE40634"
    assert td.tasks[0].description == "Standing project scope"


def test_set_project_description(engine) -> None:
    with session_scope(engine) as session:
        add_project(session, "DescProj", [])
    with session_scope(engine) as session:
        set_project_description(session, "DescProj", "My fixed blurb")
    with session_scope(engine) as session:
        from daily_tasks.services.projects import get_project_by_name

        p = get_project_by_name(session, "DescProj")
    assert p is not None
    assert p.description == "My fixed blurb"


def test_export_ticket_daily_shape(engine) -> None:
    with session_scope(engine) as session:
        p = add_project(session, "Beta", ["d1"])
        pid = int(p.id)
    with session_scope(engine) as session:
        save_daily_ticket(
            session,
            entry_date=date(2026, 5, 1),
            title="Day one",
            description="desc",
            tasks=[(pid, 1.5, "task note", ["x", "y"])],
        )
    with session_scope(engine) as session:
        doc = build_export(session)
    payload = json.loads(doc.model_dump_json())
    assert "exported_at" in payload
    assert "tickets" in payload
    assert len(payload["tickets"]) == 1
    td = payload["tickets"][0]
    assert td["date"] == "2026-05-01"
    assert td["title"] == "Day one"
    assert td["description"] == "desc"
    assert len(td["tasks"]) == 1
    task = td["tasks"][0]
    assert task["name"] == "Beta"
    assert task["description"] == "task note"
    assert task["time"] == 1.5
    assert [s["name"] for s in task["subtasks"]] == ["x", "y"]


def test_daily_ticket_upsert_same_day_preserves_ritm(engine) -> None:
    with session_scope(engine) as session:
        p = add_project(session, "Gamma", [])
        pid = int(p.id)
    d = date(2026, 5, 2)
    with session_scope(engine) as session:
        save_daily_ticket(
            session,
            entry_date=d,
            title="T",
            description="",
            tasks=[(pid, 1.0, "", ["a"])],
        )
    with session_scope(engine) as session:
        set_ritm_for_date(session, entry_date=d, ritm_number="RITM001")
    with session_scope(engine) as session:
        save_daily_ticket(
            session,
            entry_date=d,
            title="T2",
            description="",
            tasks=[(pid, 2.0, "", ["b", "c"])],
        )
    with session_scope(engine) as session:
        rows = list(session.exec(select(DailyTicket).where(DailyTicket.entry_date == d)).all())
    assert len(rows) == 1
    assert rows[0].ritm_number == "RITM001"
    assert rows[0].title == "T2"
    with session_scope(engine) as session:
        doc = build_export(session, date_from=d, date_to=d)
    assert len(doc.tickets) == 1
    assert doc.tickets[0].ritm_number == "RITM001"
    assert [s.name for s in doc.tickets[0].tasks[0].subtasks] == ["b", "c"]
    assert doc.tickets[0].tasks[0].time == 2.0
