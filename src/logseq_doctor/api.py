"""Logseq API client."""
from dataclasses import dataclass
from pathlib import Path
from textwrap import dedent, indent
from typing import List, Optional, TextIO
from uuid import UUID

import requests

DASH = "-"
LINE_BREAK = "\n"
SPACE = " "


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

    @staticmethod
    def indent(text: str, level: int = 0, *, nl: bool = False) -> str:
        """Indent text by the desired level."""
        spaces = SPACE * (level * 2)
        return indent(dedent(text).strip(), spaces) + (LINE_BREAK if nl else "")


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

    def _open(self) -> TextIO:
        mode = "r+" if self.path.exists() else "w"
        return self.path.open(mode)

    def append(self, text: str, *, level: int = 0) -> None:
        """Append text to the end of page."""
        with self._open() as file:
            file.seek(0, 2)
            file.write(Block.indent(text, level) + LINE_BREAK)

    def insert(self, text: str, start: int, *, level: int = 0) -> int:
        """Insert text at the desired offset. Return the next insert position."""
        with self._open() as file:
            file.seek(start)
            remaining_content = file.read()
            file.seek(start)
            new_text = Block.indent(text, level) + LINE_BREAK
            file.write(new_text)
            file.write(remaining_content)
            return start + len(new_text)

    def replace(self, new_text: str, start: int, end: int, *, level: int = 0) -> None:
        """Replace text at the desired offset."""
        with self._open() as file:
            file.seek(end)
            remaining_content = file.read()

            file.seek(start)
            file.write(Block.indent(new_text, level) + LINE_BREAK)
            file.write(remaining_content)

    def find_slice(  # noqa: C901 # FIXME[AA]: refactor this method
        self,
        search_string: str,
        *,
        start: int = 0,
        end: int = None,
        level: int = None,
    ) -> Optional[Slice]:
        """Find a slice of Markdown blocks in a Logseq page."""
        if not self.path.exists():
            return None

        # TODO: The right way would be to navigate the Markdown AST.
        #  This is a hacky/buggy solution that works for now, for most cases.
        content = self.path.read_text()
        relative_content = content[start:] if end is None else content[start:end]
        pos_search_string = relative_content.find(search_string)
        if pos_search_string == -1:
            return None

        if level is None:
            bullet = DASH + SPACE
            previous_dash = relative_content[:pos_search_string].rfind(bullet)
            if previous_dash == -1:
                return None  # TODO: test: no outline before search string

            previous_line_break = relative_content[:previous_dash].rfind(LINE_BREAK)
            # There are no line breaks before on the first line of the file
            slice_start = 0 if previous_line_break == -1 else previous_line_break + 1
        else:
            bullet = LINE_BREAK + (SPACE * (level * 2)) + DASH + SPACE
            previous_line_break = relative_content[:pos_search_string].rfind(bullet)
            if previous_line_break == -1:
                return None  # TODO: test this case
            slice_start = previous_line_break + 1

        column = relative_content[slice_start:].find(DASH)
        if column == -1:
            column = 0  # TODO: test this case

        spaces = LINE_BREAK + (SPACE * column)
        spaces_with_dash = spaces + DASH

        pos_last_line = relative_content[slice_start:].find(spaces_with_dash)
        if pos_last_line == -1:
            pos_last_line = 0
            while True:
                pos_next = relative_content[slice_start + pos_last_line :].find(spaces)
                if pos_next == -1:
                    break
                pos_last_line += pos_next + 1

        pos_next_line_break = relative_content[slice_start + pos_last_line :].find(LINE_BREAK)
        # TODO: test: file without line break at the end
        slice_end = (
            len(relative_content)
            if pos_next_line_break == -1
            else slice_start + pos_last_line + pos_next_line_break + 1
        )

        if end is not None and slice_end > end:
            return None

        return Slice(
            content=str(content[slice_start + start : slice_end + start]),
            start_index=slice_start + start,
            end_index=slice_end + start,
        )
