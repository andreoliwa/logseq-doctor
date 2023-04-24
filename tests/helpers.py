from pathlib import Path


def remove_last_char(path: Path) -> None:
    assert path.exists()
    content = path.read_text()
    path.write_text(content[:-1])
