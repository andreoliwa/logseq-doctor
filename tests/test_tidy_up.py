from pathlib import Path

from typer.testing import CliRunner

from logseq_doctor.cli import app


def test_remove_empty_bullets_from_multiple_files(datadir):
    actual1: Path = datadir / "empty-bullets-1.md"
    expected1: Path = datadir / "empty-bullets-1-clean.md"
    actual2: Path = datadir / "empty-bullets-2.md"
    expected2: Path = datadir / "empty-bullets-2-clean.md"

    result = CliRunner().invoke(app, ["tidy-up", str(actual1), str(actual2)])
    assert result.output == f"removed empty bullets from {actual1}\nremoved empty bullets from {actual2}\n"
    assert result.exit_code == 0
    assert actual1.read_text() == expected1.read_text().strip()
    assert actual2.read_text() == expected2.read_text().strip()
