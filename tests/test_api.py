import json
from pathlib import Path
from textwrap import dedent, indent
from uuid import UUID

import pytest
import responses

from logseq_doctor.api import Block, Logseq, Page, Slice
from logseq_doctor.cli import KANBAN_BOARD_SEARCH_STRING


@pytest.fixture()
def logseq():
    return Logseq("http://localhost:1234", "token", "my-notes")


def test_block_url():
    block = Block(
        block_id=UUID("d5cfa844-82d7-439b-b512-fbdea5564cff"),
        journal_iso_date=20230419,
        page_title="Wednesday, 19.04.2023",
        content="Bla bla",
        marker="",
    )
    assert (
        block.url("my-personal-notes")
        == "logseq://graph/my-personal-notes?block-id=d5cfa844-82d7-439b-b512-fbdea5564cff"
    )


@responses.activate
def test_query(logseq, datadir: Path):
    responses.post("http://localhost:1234/api", json=json.loads((datadir / "valid-todo-tasks.json").read_text()))
    assert logseq.query("doesn't matter, the response is mocked anyway") == [
        Block(
            block_id=UUID("644069fc-ecd3-4ac0-9363-4fd63cdb18b3"),
            journal_iso_date=20230419,
            page_title="Wednesday, 19.04.2023",
            content="TODO Write a [[CLI]] script to parse #Logseq tasks: [some link](https://example.com/path/to/file.html)",
            marker="TODO",
        ),
        Block(
            block_id=UUID("644069fc-022a-4d64-af27-c62d92fba9e6"),
            journal_iso_date=20230419,
            page_title="Wednesday, 19.04.2023",
            content="TODO Complete this tutorial: [Getting started](https://tutorials.net/index.html)",
            marker="TODO",
        ),
        Block(
            block_id=UUID("644069fc-6f99-49f5-8499-fc795d1209b4"),
            journal_iso_date=20230420,
            page_title="Thursday, 20.04.2023",
            content="TODO Parse the CSV file",
            marker="TODO",
        ),
    ]


def test_append_to_non_existing_page(datadir: Path):
    path = datadir / "non-existing-page.md"
    assert not path.exists()
    page = Page(path, overwrite=True)
    page.append("- new item")
    page.close()
    assert path.read_text() == "- new item\n"


def test_append_to_existing_page(datadir: Path):
    before = datadir / "page-before.md"
    assert before.exists()
    page = Page(before)
    page.append("- new item")
    page.close()

    after = datadir / "page-after.md"
    assert before.read_text() == after.read_text()


@pytest.fixture()
def existing_kanban(shared_datadir: Path):
    return Page(shared_datadir / "existing-kanban.md")


def test_non_existing_text(existing_kanban: Page):
    assert existing_kanban.find_slice("doesn't exist") is None


def test_find_by_outline_content(existing_kanban: Page):
    expected = """
      - Second page: Another card in the same column
        kanban-list:: Preparing
        collapsed:: true
        - ((b9f4b406-2033-4f0a-996d-16a5537cc8b8))
    """
    assert existing_kanban.find_slice("Another card in the same column") == Slice(
        content=indent(dedent(expected).strip() + "\n", " " * 2),
        start_index=503,
        end_index=648,
    )


@pytest.fixture()
def nested_kanban(datadir):
    return Page(datadir / "nested-kanban.md")


def test_find_kanban_header(nested_kanban):
    assert nested_kanban.find_slice(KANBAN_BOARD_SEARCH_STRING) == Slice(
        content="        - {{renderer :kboard, be7f0de9-4e88-42f9-911d-9b7fc51a654e, kanban-list}}\n",
        start_index=104,
        end_index=186,
    )


def test_find_first_line(nested_kanban):
    assert nested_kanban.find_slice("first line") == Slice(
        content="- Item on the first line of the file\n",
        start_index=0,
        end_index=37,
    )


def test_find_last_line(nested_kanban):
    assert nested_kanban.find_slice("The last in line") == Slice(
        content="  - Sub-item c - The last in line\n",
        start_index=708,
        end_index=742,
    )


def test_find_block_by_property(nested_kanban):
    expected = """
        - My board
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
    """
    assert nested_kanban.find_slice("id:: be7f0de9-4e88-42f9-911d-9b7fc51a654e") == Slice(
        content=indent(dedent(expected).strip() + "\n", " " * 8),
        start_index=186,
        end_index=628,
    )
