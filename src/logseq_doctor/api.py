"""Logseq API client."""
from dataclasses import dataclass, field
from pathlib import Path
from textwrap import dedent, indent
from typing import List, Optional, TextIO, Tuple
from uuid import UUID, uuid4

import requests

from logseq_doctor.constants import (
    DASH,
    KANBAN_BOARD_SEARCH_STRING,
    KANBAN_LIST,
    KANBAN_UNKNOWN_COLUMN,
    LINE_BREAK,
    SPACE,
)


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

    @staticmethod
    def sort_by_date(blocks: List) -> List:
        """Sort a list blocks by date."""
        return sorted(blocks, key=lambda row: (row.journal_iso_date, row.content))


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
            },
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
                ),
            )
        return rows


@dataclass(frozen=True)
class Slice:
    """Slice of Markdown blocks in a Logseq page."""

    content: str
    start_index: int
    end_index: int


@dataclass(frozen=True)
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

    def find_slice(
        self,
        search_string: str,
        *,
        start: int = 0,
        end: int = None,
        level: int = None,
    ) -> Optional[Slice]:
        """Find a slice of Markdown blocks in a Logseq page."""
        try:
            # TODO: The right way would be to navigate the Markdown AST.
            #  This is a hacky/buggy solution that works for now, for most cases.
            content = self.path.read_text()
            relative_content = content[start:] if end is None else content[start:end]
            pos_search_string = relative_content.find(search_string)
            if pos_search_string == -1:
                return None

            previous_line_break = self._find_previous_line_break(relative_content, pos_search_string, level)
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
        except (FileNotFoundError, ValueError):
            return None

    @staticmethod
    def _find_previous_line_break(relative_content: str, pos_search_string: int, level: int = None) -> int:
        if level is None:
            bullet = DASH + SPACE
            previous_dash = relative_content[:pos_search_string].rfind(bullet)
            if previous_dash == -1:
                msg = "No bullet found before search string"
                raise ValueError(msg)

            return relative_content[:previous_dash].rfind(LINE_BREAK)
            # There are no line breaks before on the first line of the file
        bullet = LINE_BREAK + (SPACE * (level * 2)) + DASH + SPACE
        return relative_content[:pos_search_string].rfind(bullet)


@dataclass
class Kanban:
    """Create/update the Kanban board used by the https://github.com/sethyuan/logseq-plugin-kanban-board plugin."""

    page: Page
    blocks: List[Block]

    _renderer: Slice = field(init=False)

    def find(self) -> Slice:
        """Find a Kanban board in the page."""
        self._renderer = self.page.find_slice(KANBAN_BOARD_SEARCH_STRING)
        return self._renderer

    @staticmethod
    def _generate_kanban_id() -> UUID:  # pragma: no cover
        """Generate a random UUID for the Kanban board. Mocked on tests."""
        return uuid4()

    @staticmethod
    def render_header(kanban_id: UUID, title: str) -> str:
        """Render the header of the Kanban board."""
        return Block.indent(
            f"""
            - {{{{{KANBAN_BOARD_SEARCH_STRING} {kanban_id}, {KANBAN_LIST}}}}}
            - {title}
              id:: {kanban_id}
              collapsed:: true
            """,
        )

    @staticmethod
    def render_column(column: str) -> Tuple[str, str]:
        """Render a column for the Kanban board."""
        key = f"{KANBAN_LIST}:: {column}"
        card = Block.indent(
            f"""
            - placeholder #.kboard-placeholder
              {key}
            """,
        )
        return key, card

    @classmethod
    def render_card(cls, column: str, block: Block) -> str:
        """Render a card for the Kanban board."""
        if block.journal_iso_date:
            content = block.content
        else:
            content = f"{block.page_title}: {block.content} #[[{block.page_title}]]"
        key, _ = cls.render_column(column)
        return Block.indent(
            f"""
            - {content}
              {KANBAN_LIST}:: {column}
              collapsed:: true
              - (({block.block_id}))
            """,
        )

    def add(self) -> None:
        """Add a Kanban board to the page."""
        self.page.append(self.render_header(self._generate_kanban_id(), "My board"))

        columns = set()
        for block in self.blocks:
            column = block.marker or KANBAN_UNKNOWN_COLUMN
            if column not in columns:
                columns.add(column)
                _, card = self.render_column(column)
                self.page.append(card, level=1)

            self.page.append(self.render_card(column, block), level=1)

    def update(self) -> None:
        """Update an existing Kanban board."""
        kanban_id = UUID(self._renderer.content.split(",")[1].strip())
        board_slice = self.page.find_slice(f"id:: {kanban_id}", start=self._renderer.end_index)
        board_start = board_slice.start_index
        pos_next_insert = board_end = board_slice.end_index

        columns = set()
        for block in self.blocks:
            if self.page.find_slice(str(block.block_id), start=board_start, end=board_end):
                continue

            column = block.marker or KANBAN_UNKNOWN_COLUMN
            if column not in columns:
                columns.add(column)
                key, card = self.render_column(column)
                if not self.page.find_slice(key, start=board_start, end=board_end):
                    pos = self.page.insert(card, start=pos_next_insert, level=1)
                    if pos_next_insert < pos:
                        pos_next_insert = pos

            pos = self.page.insert(self.render_card(column, block), start=pos_next_insert, level=1)
            if pos_next_insert < pos:
                pos_next_insert = pos
