from __future__ import annotations

from textwrap import dedent
from unittest.mock import Mock, patch
from uuid import UUID

import pytest
from logseq_doctor.api import Block, Logseq
from logseq_doctor.cli import app
from typer.testing import CliRunner


@pytest.fixture()
def mock_logseq_query() -> Mock:
    with patch.object(Logseq, "query") as mocked_method:
        yield mocked_method


@pytest.fixture()
def unsorted_blocks() -> list[Block]:
    return [
        Block(
            block_id=UUID("6bcebf82-c557-4f58-84d0-3b91c7e59e93"),
            journal_iso_date=20230421,
            page_title="Another page",
            raw_content="DOING Focus on tasks",
            marker="DOING",
        ),
        Block(
            block_id=UUID("0d62459e-8a6b-4635-8c2e-41fbef6da6f8"),
            journal_iso_date=20230421,
            page_title="Title",
            raw_content="TODO Do yet another thing",
            marker="TODO",
        ),
        Block(
            block_id=UUID("c3824880-7527-4ba6-a465-719f01f327de"),
            journal_iso_date=20230420,
            page_title="Page",
            raw_content="TODO Do something",
            marker="TODO",
        ),
        Block(
            block_id=UUID("be4116f7-2c74-42af-af62-f38352631d11"),
            journal_iso_date=0,
            page_title="Third page",
            raw_content="Item without marker",
            marker="",
        ),
    ]


@pytest.fixture()
def blocks_sorted_by_date_content(unsorted_blocks: list[Block]) -> list[Block]:
    return [unsorted_blocks[3], unsorted_blocks[2], unsorted_blocks[0], unsorted_blocks[1]]


@pytest.mark.parametrize(
    ("tags", "expected_query"),
    [
        ([], "(and (task TODO DOING WAITING NOW LATER))"),
        (["tag1"], "(and [[tag1]] (task TODO DOING WAITING NOW LATER))"),
        (["tag1", "page2"], "(and (or [[tag1]] [[page2]]) (task TODO DOING WAITING NOW LATER))"),
    ],
)
def test_search_with_tags(mock_logseq_query: Mock, tags: list[str], expected_query: str) -> None:
    result = CliRunner().invoke(app, ["tasks", *tags])
    assert result.exit_code == 0
    assert mock_logseq_query.call_count == 1
    mock_logseq_query.assert_called_once_with(expected_query)


def test_simple_text_output(
    mock_logseq_query: Mock,
    unsorted_blocks: list[Block],
) -> None:
    mock_logseq_query.return_value = unsorted_blocks
    result = CliRunner().invoke(app, ["tasks"])
    assert result.exit_code == 0
    expected = """
        Third page§logseq://graph/my-notes?block-id=be4116f7-2c74-42af-af62-f38352631d11§Item without marker
        Page§logseq://graph/my-notes?block-id=c3824880-7527-4ba6-a465-719f01f327de§TODO Do something
        Another page§logseq://graph/my-notes?block-id=6bcebf82-c557-4f58-84d0-3b91c7e59e93§DOING Focus on tasks
        Title§logseq://graph/my-notes?block-id=0d62459e-8a6b-4635-8c2e-41fbef6da6f8§TODO Do yet another thing
    """
    assert result.stdout == dedent(expected).lstrip()


def test_blocks_sorted_by_date(
    unsorted_blocks: list[Block],
    blocks_sorted_by_date_content: list[Block],
) -> None:
    assert Block.sort_by_date(unsorted_blocks) == blocks_sorted_by_date_content
