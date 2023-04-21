from pathlib import Path
from textwrap import dedent
from unittest.mock import patch
from uuid import UUID

import pytest
from typer.testing import CliRunner

from logseq_doctor.api import Block, Logseq
from logseq_doctor.cli import app


@pytest.fixture()
def mock_logseq_query():
    with patch.object(Logseq, "query") as mocked_method:
        yield mocked_method


@pytest.fixture()
def some_blocks():
    return [
        Block(
            block_id=UUID("6bcebf82-c557-4f58-84d0-3b91c7e59e93"),
            journal_iso_date=20230421,
            page_title="Another page",
            content="DOING Focus on tasks",
            marker="DOING",
        ),
        Block(
            block_id=UUID("0d62459e-8a6b-4635-8c2e-41fbef6da6f8"),
            journal_iso_date=20230421,
            page_title="Title",
            content="TODO Do yet another thing",
            marker="TODO",
        ),
        Block(
            block_id=UUID("c3824880-7527-4ba6-a465-719f01f327de"),
            journal_iso_date=20230420,
            page_title="Page",
            content="TODO Do something",
            marker="TODO",
        ),
        Block(
            block_id=UUID("be4116f7-2c74-42af-af62-f38352631d11"),
            journal_iso_date=20230417,
            page_title="Third page",
            content="Item without marker",
            marker="",
        ),
    ]


@pytest.mark.parametrize(
    ("tags", "expected_query"),
    [
        ([], "(and (task TODO DOING WAITING NOW LATER))"),
        (["tag1"], "(and [[tag1]] (task TODO DOING WAITING NOW LATER))"),
        (["tag1", "page2"], "(and (or [[tag1]] [[page2]]) (task TODO DOING WAITING NOW LATER))"),
    ],
)
def test_search_with_tags(mock_logseq_query, tags, expected_query):
    result = CliRunner().invoke(app, ["tasks", *tags])
    assert result.exit_code == 0
    assert mock_logseq_query.call_count == 1
    mock_logseq_query.assert_called_once_with(expected_query)


def test_simple_text_output(mock_logseq_query, some_blocks):
    mock_logseq_query.return_value = some_blocks
    result = CliRunner().invoke(app, ["tasks"])
    assert result.exit_code == 0
    expected = """
        Third page: logseq://graph/my-notes?block-id=be4116f7-2c74-42af-af62-f38352631d11 Item without marker
        Page: logseq://graph/my-notes?block-id=c3824880-7527-4ba6-a465-719f01f327de TODO Do something
        Another page: logseq://graph/my-notes?block-id=6bcebf82-c557-4f58-84d0-3b91c7e59e93 DOING Focus on tasks
        Title: logseq://graph/my-notes?block-id=0d62459e-8a6b-4635-8c2e-41fbef6da6f8 TODO Do yet another thing
    """
    assert result.stdout == dedent(expected).lstrip()


def test_kanban_needs_output_path():
    result = CliRunner().invoke(app, ["tasks", "--format", "kanban"])
    assert result.exit_code == 1
    assert "Kanban format requires an output path" in result.stdout


@patch("logseq_doctor.cli._get_kanban_id")
def test_override_kanban_file(mock_get_kanban_id, mock_logseq_query, datadir, some_blocks):
    mock_get_kanban_id.return_value = UUID("7991f73d-628a-4f98-af7a-901e2f51caa6")
    mock_logseq_query.return_value = some_blocks
    file: Path = datadir / "kanban.md"
    assert not file.exists()
    result = CliRunner().invoke(app, ["tasks", "--format", "kanban", "--output", str(file)])
    assert result.exit_code == 0
    assert result.stdout == f"Overriding {file} with Kanban board\n✨ Done.\n"
    assert file.read_text() == (datadir / "expected-kanban.md").read_text()
