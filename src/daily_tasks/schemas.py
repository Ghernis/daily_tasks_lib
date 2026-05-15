from __future__ import annotations

from datetime import datetime
from typing import List

from pydantic import BaseModel, ConfigDict

from daily_tasks.ticket_schema import TicketDaily


class ExportDocument(BaseModel):
    model_config = ConfigDict(frozen=True)

    exported_at: datetime
    tickets: List[TicketDaily]
