"""Logseq API client."""
from dataclasses import dataclass
from pathlib import Path
from textwrap import dedent, indent
from typing import IO, List, Optional
from uuid import UUID

import requests

OUTLINE_DASH = "-"


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


@dataclass(frozen=True)
class Slice:
    """Slice of Markdown blocks in a Logseq page."""

    content: str
    start_index: int
    end_index: int


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

    def find_slice(self, search_string: str) -> Optional[Slice]:
        """Find a slice of Markdown blocks in a Logseq page."""
        # TODO: The right way would be to navigate the Markdown AST.
        #  This is a hacky/buggy solution that works for now, for most cases.
        content = self.path.read_text()
        pos_search_string = content.find(search_string)
        if pos_search_string == -1:
            return None

        # FIXME[AA]: try with a regex
        # import re
        # pattern = r"\s*- "
        # matches = re.findall(pattern, content[:pos_search_string])
        #
        # if matches:
        #     last_match = matches[-1]
        #     print("Last match found:", last_match)
        # else:
        #     print("No match found.")

        previous_dash = content[:pos_search_string].rfind(f"{OUTLINE_DASH} ")
        if previous_dash == -1:
            return None  # TODO: test: no outline before search string

        previous_line_break = content[:previous_dash].rfind("\n")
        # There are no line breaks before on the first line of the file
        pos_start = 0 if previous_line_break == -1 else previous_line_break + 1

        column = content[pos_start:].find(OUTLINE_DASH)
        if column == -1:
            column = 0  # TODO: test this case

        spaces = "\n" + (" " * column)
        spaces_with_dash = spaces + OUTLINE_DASH

        pos_last_line = content[pos_start:].find(spaces_with_dash)
        if pos_last_line == -1:
            pos_last_line = 0
            while True:
                pos_next = content[pos_start + pos_last_line :].find(spaces)
                if pos_next == -1:
                    break
                pos_last_line += pos_next + 1

        pos_next_line_break = content[pos_start + pos_last_line :].find("\n")
        # TODO: test: file without line break at the end
        pos_end = len(content) if pos_next_line_break == -1 else pos_start + pos_last_line + pos_next_line_break + 1

        return Slice(content=str(content[pos_start:pos_end]), start_index=pos_start, end_index=pos_end)
