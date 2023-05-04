import os
from pathlib import Path


def remove_last_chars(path: Path) -> None:
    assert path.exists(), path
    content = path.read_text()
    len_sep = len(os.linesep)
    path.write_text(content[: -1 * len_sep])
