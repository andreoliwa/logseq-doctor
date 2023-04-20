from unittest.mock import patch

import pytest
from typer.testing import CliRunner

from logseq_doctor.api import Logseq
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
