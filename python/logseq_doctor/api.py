"""Logseq API client."""

from __future__ import annotations

import os
import urllib.parse
from dataclasses import dataclass, field
from io import SEEK_END
from pathlib import Path  # noqa: TCH003 Typer needs this import to infer the type of the argument
from textwrap import dedent, indent
from typing import TextIO
from uuid import UUID, uuid4

import requests

from logseq_doctor.constants import (
    BEGINNING,
    CHAR_DASH,
    CHAR_NBSP,
    CHAR_SPACE,
    CHAR_TAB,
    KANBAN_BOARD_SEARCH_STRING,
    KANBAN_BOARD_TITLE,
    KANBAN_LIST,
    KANBAN_UNKNOWN_COLUMN,
)


@dataclass(frozen=True)
class Block:
    """Logseq block."""

    block_id: UUID
    journal_iso_date: int
    page_title: str
    raw_content: str
    marker: str

    @property
    def pretty_content(self) -> str:
        """Return the block content without the marker."""
        if self.raw_content.startswith(self.marker):
            len_marker = len(self.marker)
            return self.raw_content[len_marker:].strip()
        return self.raw_content

    @property
    def embed(self) -> str:
        """Return the block content as an embed."""
        return f"{{{{embed (({self.block_id}))}}}}"

    def url(self, graph_name: str) -> str:
        """Build a Logseq block URL."""
        return f"logseq://graph/{graph_name}?block-id={self.block_id}"

    @staticmethod
    def indent(text: str, *, level: int = 0, nl: bool = False) -> str:
        """Indent text by the desired level, preserving tabs."""
        spaces = CHAR_SPACE * (level * 2)
        return indent(dedent(text.replace(CHAR_TAB, CHAR_NBSP)).strip().replace(CHAR_NBSP, CHAR_TAB), spaces) + (
            os.linesep if nl else ""
        )

    @staticmethod
    def sort_by_date(blocks: list) -> list:
        """Sort a list of blocks by date."""
        return sorted(blocks, key=lambda row: (row.journal_iso_date, row.raw_content))


@dataclass(frozen=True)
class Logseq:
    """Logseq API client."""

    url: str
    token: str
    graph_path: Path

    @property
    def graph_name(self) -> str:
        """Return the graph name from the path."""
        return self.graph_path.stem

    def query(self, query: str) -> list[Block]:
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

        rows: list[Block] = []
        for obj in resp.json():
            page = obj.get("page", {})
            block_id = obj.get("uuid")
            rows.append(
                Block(
                    block_id=UUID(block_id),
                    journal_iso_date=page.get("journalDay", 0),
                    page_title=page.get("originalName"),
                    raw_content=obj.get("content").splitlines()[0],
                    marker=obj.get("marker"),
                ),
            )
        return rows

    def page_from_name(self, name: str) -> Page:
        """Return a Page object from a page name."""
        return Page(self.graph_path.expanduser() / "pages" / f"{name}.md")


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

    def url(self, graph_name: str) -> str:
        """Build a Logseq page URL."""
        page = urllib.parse.quote(self.path.stem)
        return f"logseq://graph/{graph_name}?page={page}"

    def _open(self, mode: str = "") -> TextIO:
        return self.path.open(mode or ("r+" if self.path.exists() else "w"))

    def add_line_break(self) -> bool:
        """Return True if a line break was added to the end of the file."""
        if not self.path.exists():
            return False
        len_line_break = len(os.linesep)
        with self._open("rb+") as file:
            file.seek(-1 * len_line_break, 2)
            last_chars = file.read(len_line_break)
            if last_chars != os.linesep.encode():
                file.write(os.linesep.encode())
                return True
        return False

    def remove_line_break(self) -> bool:
        """Return True if a line break was removed from the end of the file."""
        if not self.path.exists():
            return False
        content = self.path.read_text()
        len_line_break = len(os.linesep)
        self.path.write_text(content[: -1 * len_line_break])
        return True

    def append(self, text: str, *, level: int = 0) -> None:
        """Append text to the end of page."""
        with self._open() as file:
            file.seek(0, SEEK_END)
            file.write(Block.indent(text, level=level) + os.linesep)

    def insert(self, text: str, start: int, *, level: int = 0) -> int:
        """Insert text at the desired offset. Return the next insert position."""
        with self._open() as file:
            file.seek(start)
            remaining_content = file.read()

            # A dirty hack to adjust the start position of a file:
            # search for the nearest new line and move the pointer if found.
            # For some reasons, in one file the start was off by 2;
            # maybe it's the BOM in the beginning of the file?
            # TODO: this whole logic of manipulating blocks should be done with Markdown AST...
            pos_nearest_new_line = remaining_content[:BEGINNING].find(os.linesep)
            if pos_nearest_new_line != -1:
                start += pos_nearest_new_line + 1
                remaining_content = remaining_content[pos_nearest_new_line + 1 :]

            file.seek(start)
            new_text = Block.indent(text, level=level) + os.linesep
            file.write(new_text)
            pos_after_writing = file.tell()
            file.write(remaining_content)
            return pos_after_writing

    def replace(self, new_text: str, start: int, end: int, *, level: int = 0) -> None:
        """Replace text at the desired offset."""
        with self._open() as file:
            file.seek(end)
            remaining_content = file.read()

            file.seek(start)
            file.write(Block.indent(new_text, level=level) + os.linesep)
            file.write(remaining_content)

    def find_slice(
        self,
        search_string: str,
        *,
        start: int = 0,
        end: int | None = None,
        level: int | None = None,
    ) -> Slice | None:
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

            column = relative_content[slice_start:].find(CHAR_DASH)
            if column == -1:
                column = 0  # TODO: test this case

            spaces = os.linesep + (CHAR_SPACE * column)
            spaces_with_dash = spaces + CHAR_DASH

            pos_last_line = relative_content[slice_start:].find(spaces_with_dash)
            if pos_last_line == -1:
                pos_last_line = 0
                while True:
                    pos_next = relative_content[slice_start + pos_last_line :].find(spaces)
                    if pos_next == -1:
                        break
                    pos_last_line += pos_next + 1

            pos_next_line_break = relative_content[slice_start + pos_last_line :].find(os.linesep)
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
    def _find_previous_line_break(relative_content: str, pos_search_string: int, level: int | None = None) -> int:
        if level is None:
            bullet = CHAR_DASH + CHAR_SPACE
            previous_dash = relative_content[:pos_search_string].rfind(bullet)
            if previous_dash == -1:
                msg = "No bullet found before search string"
                raise ValueError(msg)

            return relative_content[:previous_dash].rfind(os.linesep)
            # There are no line breaks before on the first line of the file
        bullet = os.linesep + (CHAR_SPACE * (level * 2)) + CHAR_DASH + CHAR_SPACE
        return relative_content[:pos_search_string].rfind(bullet)


@dataclass
class Kanban:
    """Create/update the Kanban board used by the https://github.com/sethyuan/logseq-plugin-kanban-board plugin."""

    page: Page
    blocks: list[Block]

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
    def render_column(column: str) -> tuple[str, str]:
        """Render a column for the Kanban board."""
        key = f"{KANBAN_LIST}:: {column}"
        card = Block.indent(
            f"""
            {CHAR_TAB}- placeholder #.kboard-placeholder
            {CHAR_TAB}  {key}
            """,
        )
        return key, card

    @classmethod
    def render_card(cls, column: str, block: Block) -> str:
        """Render a collapsed card for the Kanban board."""
        if block.journal_iso_date:
            content = block.pretty_content
        else:
            content = f"{block.page_title}: {block.pretty_content} #[[{block.page_title}]]"
        return Block.indent(
            f"""
            {CHAR_TAB}- {content}
            {CHAR_TAB}  {KANBAN_LIST}:: {column}
            {CHAR_TAB}  collapsed:: true
            {CHAR_TAB}{CHAR_TAB}- {block.embed}
            """,
        )

    def add(self) -> None:
        """Add a Kanban board to the page."""
        if not self.page.path.exists():
            msg = f"Page {self.page.path} does not exist"
            raise FileNotFoundError(msg)

        self.page.append(self.render_header(self._generate_kanban_id(), KANBAN_BOARD_TITLE))

        columns = set()
        for block in self.blocks:
            column = block.marker or KANBAN_UNKNOWN_COLUMN
            if column not in columns:
                columns.add(column)
                _, card = self.render_column(column)
                self.page.append(card)

            self.page.append(self.render_card(column, block))

    def update(self) -> None:
        """Update an existing Kanban board."""
        kanban_id = UUID(self._renderer.content.split(",")[1].strip())
        board_slice = self.page.find_slice(f"id:: {kanban_id}", start=self._renderer.end_index)
        board_start = board_slice.start_index
        board_end = board_slice.end_index

        # Insert the next card at the beginning of the board
        first_child = self.page.find_slice(f"{KANBAN_LIST}:: ", start=board_start, end=board_end)
        if not first_child:
            return
        pos_next_insert = first_child.start_index

        columns = set()
        for block in self.blocks:
            if self.page.find_slice(str(block.block_id), start=board_start, end=board_end):
                continue

            column = block.marker or KANBAN_UNKNOWN_COLUMN
            if column not in columns:
                columns.add(column)
                key, card = self.render_column(column)
                if not self.page.find_slice(key, start=board_start, end=board_end):
                    pos = self.page.insert(card, start=pos_next_insert)
                    if pos_next_insert < pos:
                        pos_next_insert = pos
                    board_end += len(card) + len(os.linesep)

            card = self.render_card(column, block)
            pos = self.page.insert(card, start=pos_next_insert)
            if pos_next_insert < pos:
                pos_next_insert = pos
            board_end += len(card) + len(os.linesep)
