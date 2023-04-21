import json
from pathlib import Path
from uuid import UUID

import pytest
import responses

from logseq_doctor.api import Block, Logseq, Page


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
