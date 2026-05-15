from __future__ import annotations

from sqlmodel import Session

from daily_tasks.models import AuthorProfile
from daily_tasks.ticket_schema import Author

_AUTHOR_ID = 1


def get_author_profile(session: Session) -> AuthorProfile | None:
    return session.get(AuthorProfile, _AUTHOR_ID)


def get_author_for_export(session: Session) -> list[Author]:
    row = get_author_profile(session)
    if row is None or not row.name.strip():
        return []
    return [Author(name=row.name.strip(), email=row.email or "", sn_id=row.sn_id or "")]


def set_author_profile(
    session: Session,
    *,
    name: str,
    email: str = "",
    sn_id: str = "",
) -> AuthorProfile:
    key = name.strip()
    if not key:
        raise ValueError("Author name is required.")
    row = session.get(AuthorProfile, _AUTHOR_ID)
    if row is None:
        row = AuthorProfile(id=_AUTHOR_ID, name=key, email=email.strip(), sn_id=sn_id.strip())
        session.add(row)
    else:
        row.name = key
        row.email = email.strip()
        row.sn_id = sn_id.strip()
        session.add(row)
    session.commit()
    session.refresh(row)
    return row


def default_ticket_description(session: Session, entry_date_iso: str) -> str:
    row = get_author_profile(session)
    if row is None or not row.name.strip():
        return f"Tareas realizadas el {entry_date_iso}"
    if row.sn_id.strip():
        return f"Tareas realizadas el {entry_date_iso} por {row.name.strip()}({row.sn_id.strip()})"
    return f"Tareas realizadas el {entry_date_iso} por {row.name.strip()}"
