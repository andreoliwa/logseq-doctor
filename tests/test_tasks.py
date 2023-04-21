from pathlib import Path
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


@patch("logseq_doctor.cli._get_kanban_id")
def test_override_kanban_file(mock_get_kanban_id, mock_logseq_query, datadir):
    mock_get_kanban_id.return_value = UUID("7991f73d-628a-4f98-af7a-901e2f51caa6")
    mock_logseq_query.return_value = [
        Block(
            UUID("6bcebf82-c557-4f58-84d0-3b91c7e59e93"),
            20230421,
            "Another page",
            "",
            "DOING Focus on tasks",
            "DOING",
        ),
        Block(
            UUID("0d62459e-8a6b-4635-8c2e-41fbef6da6f8"),
            20230421,
            "Title",
            "",
            "TODO Do yet another thing",
            "TODO",
        ),
        Block(
            UUID("c3824880-7527-4ba6-a465-719f01f327de"),
            20230420,
            "Page",
            "",
            "TODO Do something",
            "TODO",
        ),
        Block(
            UUID("be4116f7-2c74-42af-af62-f38352631d11"),
            20230417,
            "Third page",
            "",
            "Item without marker",
            "",
        ),
    ]
    file: Path = datadir / "kanban.md"
    assert not file.exists()
    result = CliRunner().invoke(app, ["tasks", "--format", "kanban", "--output", str(file)])
    assert result.exit_code == 0
    assert result.stdout == f"Overriding {file} with Kanban board\n✨ Done.\n"
    assert file.read_text() == (datadir / "expected-kanban.md").read_text()
