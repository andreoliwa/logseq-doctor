"""Logseq API client."""
from dataclasses import dataclass
from pathlib import Path
from textwrap import dedent, indent
from typing import IO, List, Optional
from uuid import UUID

import requests


@dataclass(frozen=True)
class Block:
    """Logseq block."""

    block_id: UUID
    journal_iso_date: int
    page_title: str
    content: str
    marker: str

    def url(self, graph: str) -> str:
        """Build a Logseq block URL."""
        return f"logseq://graph/{graph}?block-id={self.block_id}"


@dataclass(frozen=True)
class Logseq:
    """Logseq API client."""

    url: str
    token: str
    graph: str

    def query(self, query: str) -> List[Block]:
        """Query Logseq API."""
        session = requests.Session()
        session.headers.update(
            {
                "Authorization": f"Bearer {self.token}",
                "Content-Type": "application/json",
            }
        )
        resp = session.post(f"{self.url}/api", json={"method": "logseq.db.q", "args": [query]})
        resp.raise_for_status()

        rows: List[Block] = []
        for obj in resp.json():
            page = obj.get("page", {})
            block_id = obj.get("uuid")
            rows.append(
                Block(
                    block_id=UUID(block_id),
                    journal_iso_date=page.get("journalDay", 0),
                    page_title=page.get("originalName"),
                    content=obj.get("content").splitlines()[0],
                    marker=obj.get("marker"),
                )
            )
        return rows


@dataclass
class Page:
    """Logseq page."""

    path: Path
    overwrite: bool = False

    _handle: Optional[IO] = None

    def __post_init__(self) -> None:
        """Open file handle if path is provided."""
        self._handle = self.path.open("w" if self.overwrite else "a")

    def append(self, markdown: str, *, level: int = 0) -> None:
        """Append markdown to page."""
        content = indent(dedent(markdown).strip(), " " * (level * 2))
        self._handle.write(content + "\n")

    def close(self) -> None:
        """Close file handle."""
        self._handle.close()
