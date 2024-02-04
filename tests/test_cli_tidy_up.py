import os
from pathlib import Path

from logseq_doctor.cli import app
from typer.testing import CliRunner


def test_remove_empty_bullets_from_multiple_files(datadir: Path) -> None:
    actual1: Path = datadir / "empty-bullets-1.md"
    expected1: Path = datadir / "empty-bullets-1-clean.md"
    actual2: Path = datadir / "empty-bullets-2.md"
    expected2: Path = datadir / "empty-bullets-2-clean.md"

    result = CliRunner().invoke(app, ["tidy-up", str(actual1), str(actual2)])
    assert (
        result.output == f"removed empty bullets and double spaces from {actual1}{os.linesep}"
        f"removed empty bullets and double spaces from {actual2}{os.linesep}"
    )
    assert result.exit_code == 0
    assert actual1.read_text() == expected1.read_text().strip()
    assert actual2.read_text() == expected2.read_text().strip()


def test_nothing_to_remove(datadir: Path) -> None:
    file: Path = datadir / "empty-bullets-1-clean.md"
    content_before = file.read_text()
    result = CliRunner().invoke(app, ["tidy-up", str(file)])
    assert not result.output
    assert result.exit_code == 0
    assert file.read_text() == content_before
