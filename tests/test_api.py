import json
import os
from pathlib import Path
from typing import Optional
from uuid import UUID

import pytest
import responses

from logseq_doctor.api import Block, Logseq, Page, Slice
from logseq_doctor.constants import KANBAN_BOARD_SEARCH_STRING


@pytest.fixture()
def logseq() -> Logseq:
    return Logseq("http://localhost:1234", "token", "my-notes")


def test_block_url() -> None:
    block = Block(
        block_id=UUID("d5cfa844-82d7-439b-b512-fbdea5564cff"),
        journal_iso_date=20230419,
        page_title="Wednesday, 19.04.2023",
        raw_content="Bla bla",
        marker="",
    )
    assert (
        block.url("my-personal-notes")
        == "logseq://graph/my-personal-notes?block-id=d5cfa844-82d7-439b-b512-fbdea5564cff"
    )


@responses.activate
def test_query(logseq: Logseq, datadir: Path) -> None:
    responses.post("http://localhost:1234/api", json=json.loads((datadir / "valid-todo-tasks.json").read_text()))
    assert logseq.query("doesn't matter, the response is mocked anyway") == [
        Block(
            block_id=UUID("644069fc-ecd3-4ac0-9363-4fd63cdb18b3"),
            journal_iso_date=20230419,
            page_title="Wednesday, 19.04.2023",
            raw_content="TODO Write a [[CLI]] script to parse #Logseq tasks: [some link](https://example.com/path/to/file.html)",
            marker="TODO",
        ),
        Block(
            block_id=UUID("644069fc-022a-4d64-af27-c62d92fba9e6"),
            journal_iso_date=20230419,
            page_title="Wednesday, 19.04.2023",
            raw_content="TODO Complete this tutorial: [Getting started](https://tutorials.net/index.html)",
            marker="TODO",
        ),
        Block(
            block_id=UUID("644069fc-6f99-49f5-8499-fc795d1209b4"),
            journal_iso_date=20230420,
            page_title="Thursday, 20.04.2023",
            raw_content="TODO Parse the CSV file",
            marker="TODO",
        ),
    ]


def test_append_to_non_existing_page(datadir: Path) -> None:
    path = datadir / "non-existing-page.md"
    assert not path.exists()
    page = Page(path)
    assert not page.add_line_break()
    page.append("- new item")
    assert path.read_text() == f"- new item{os.linesep}"


def test_append_to_existing_page(datadir: Path) -> None:
    before = datadir / "page-before.md"
    assert before.exists()
    page = Page(before)
    assert not page.add_line_break()
    page.append("- new item")
    assert before.read_text() == (datadir / "page-append.md").read_text()


def test_append_to_existing_page_without_line_break(datadir: Path) -> None:
    before = datadir / "page-before.md"

    page = Page(before)
    assert page.remove_line_break()
    assert page.add_line_break()

    page.append("- new item")
    assert before.read_text() == (datadir / "page-append.md").read_text()


def test_insert_text_into_existing_page(datadir: Path) -> None:
    before = datadir / "page-before.md"

    page = Page(before)
    assert page.remove_line_break()
    assert page.add_line_break()

    assert page.insert("- a new line after position 18", start=18) == 49  # noqa: PLR2004
    assert (
        page.insert(
            """
            - nested item
              with:: property
            """,
            level=1,
            start=49,
        )
        == 85  # noqa: PLR2004
    )
    assert before.read_text() == (datadir / "page-insert.md").read_text()


def test_replace_slice_from_existing_page(datadir: Path) -> None:
    before = datadir / "page-before.md"
    assert before.exists()
    page = Page(before)
    new_item = """
        - removed text from positions 9 to 18, add this new item
          id: 123
          - nested
            key:: value
    """
    page.replace(new_item, 9, 18, level=1)
    assert before.read_text() == (datadir / "page-replace.md").read_text()


@pytest.fixture()
def existing_kanban(shared_datadir: Path) -> Page:
    return Page(shared_datadir / "existing-kanban.md")


def test_find_non_existing_block(existing_kanban: Page) -> None:
    assert existing_kanban.find_slice("doesn't exist") is None


def test_find_block_by_outline_content(existing_kanban: Page) -> None:
    assert existing_kanban.find_slice("Another card in the same column") == Slice(
        content=Block.indent(
            """
              - Second page: Another card in the same column
                kanban-list:: Preparing
                collapsed:: true
                - ((b9f4b406-2033-4f0a-996d-16a5537cc8b8))
            """,
            level=1,
            nl=True,
        ),
        start_index=500,
        end_index=645,
    )


@pytest.mark.parametrize(
    ("search_string", "expected"),
    [
        (
            "key:: value",
            Slice(
                content=Block.indent(
                    """
                    - Item 3
                      key:: value
                    """,
                    nl=True,
                ),
                start_index=78,
                end_index=101,
            ),
        ),
        ("outside search area", None),
    ],
)
def test_find_block_within_start_end_offsets(
    datadir: Path,
    search_string: str,
    expected: Optional[Slice],
) -> None:
    page = Page(datadir / "within-start-end.md")
    assert page.find_slice(search_string, start=43, end=161) == expected


@pytest.fixture()
def nested_kanban(datadir: Path) -> Page:
    return Page(datadir / "nested-kanban.md")


def test_find_kanban_header(nested_kanban: Page) -> None:
    assert nested_kanban.find_slice(KANBAN_BOARD_SEARCH_STRING) == Slice(
        content="        - {{renderer :kboard, be7f0de9-4e88-42f9-911d-9b7fc51a654e, kanban-list}}" + os.linesep,
        start_index=104,
        end_index=186,
    )


def test_find_first_line(nested_kanban: Page) -> None:
    assert nested_kanban.find_slice("first line") == Slice(
        content=f"- Item on the first line of the file{os.linesep}",
        start_index=0,
        end_index=37,
    )


def test_find_last_line(nested_kanban: Page) -> None:
    assert nested_kanban.find_slice("The last in line") == Slice(
        content=f"  - Sub-item c - The last in line{os.linesep}",
        start_index=705,
        end_index=739,
    )


def test_find_block_by_property(nested_kanban: Page) -> None:
    assert nested_kanban.find_slice("id:: be7f0de9-4e88-42f9-911d-9b7fc51a654e") == Slice(
        content=Block.indent(
            """
            - Tasks
              id:: be7f0de9-4e88-42f9-911d-9b7fc51a654e
              collapsed:: true
              - placeholder #.kboard-placeholder
                kanban-list:: TODO
              - Card 1
                kanban-list:: TODO
                collapsed:: true
                - ((c7d26a17-f430-47b7-ad3b-f1b43099a1d5))
              - Card 2
                kanban-list:: TODO
                collapsed:: true
                - ((b9f4b406-2033-4f0a-996d-16a5537cc8b8))
            """,
            level=4,
            nl=True,
        ),
        start_index=186,
        end_index=625,
    )
    assert nested_kanban.find_slice("kanban-list:: TODO") == Slice(
        content=Block.indent(
            """
            - placeholder #.kboard-placeholder
              kanban-list:: TODO
            """,
            level=5,
            nl=True,
        ),
        start_index=281,
        end_index=357,
    )
