"""Logseq API client."""
from dataclasses import dataclass
from pathlib import Path
from textwrap import dedent, indent
from typing import IO, List, Optional
from uuid import UUID

import requests
import typer


@dataclass(frozen=True)
class Block:
    """Logseq block."""

    block_id: UUID
    journal_iso_date: int
    name: str
    url: str
    content: str
    marker: str


@dataclass(frozen=True)
class Logseq:
    """Logseq API client."""

    url: str
    token: str
    graph: str

    def build_block_url(self, block_id: UUID) -> str:
        """Build a Logseq block URL."""
        return f"logseq://graph/{self.graph}?block-id={block_id}"

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
                    name=page.get("originalName"),
                    url=self.build_block_url(block_id),
                    content=obj.get("content").splitlines()[0],
                    marker=obj.get("marker"),
                )
            )
        return rows


@dataclass
class Page:
    """Logseq page."""

    path: Path
    _handle: Optional[IO] = None

    def __post_init__(self) -> None:
        """Open file handle if path is provided."""
        if self.path:
            self._handle = self.path.open("w")

    def append(self, markdown: str, *, level: int = 0) -> None:
        """Append markdown to page."""
        content = indent(dedent(markdown).strip(), " " * (level * 2))
        typer.echo(content, self._handle)
        # TODO: if the path is empty, store in string stream and print to stdout at the end

    def close(self) -> None:
        """Close file handle."""
        if self._handle:
            self._handle.close()
