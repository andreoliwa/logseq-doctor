"""
Module that contains the command line app.

Why does this file exist, and why not put this in __main__?

  You might be tempted to import things from __main__ later, but that will cause
  problems: the code will get executed twice:

  - When you run `python -mlogseq_doctor` python will execute
    ``__main__.py`` as a script. That means there won't be any
    ``logseq_doctor.__main__`` in ``sys.modules``.
  - When you import __main__ it will get executed again (as a module) because
    there's no ``logseq_doctor.__main__`` in ``sys.modules``.

  Also see (1) from http://click.pocoo.org/5/setuptools/#setuptools-integration
"""
import re
from dataclasses import dataclass
from enum import Enum
from pathlib import Path
from typing import List

import requests
import typer

from logseq_doctor import flat_markdown_to_outline

app = typer.Typer(no_args_is_help=True)


@app.callback()
def lsd():
    """Logseq Doctor: heal your flat old Markdown files before importing them."""


@app.command(no_args_is_help=True)
def outline(text_file: typer.FileText):
    """Convert flat Markdown to outline."""
    print(flat_markdown_to_outline(text_file.read()))


@app.command()
def tidy_up(
    markdown_file: List[Path] = typer.Argument(
        ..., help="Markdown files to tidy up", exists=True, file_okay=True, dir_okay=False, writable=True
    ),
):
    """Tidy up your Markdown files by removing empty bullets in any block."""
    for each_file in markdown_file:
        old_contents = each_file.read_text()
        new_contents = re.sub(r"(\n\s*-\s*$)", "", old_contents, flags=re.MULTILINE)
        if old_contents != new_contents:
            typer.echo(f'removed empty bullets from {each_file}')
            each_file.write_text(new_contents)


class TaskFormat(str, Enum):
    """Task format."""

    text = 'text'


@dataclass
class OneLineRow:
    journal: int
    name: str
    url: str
    content: str

    def __lt__(self, other: "OneLineRow") -> bool:
        if not self.journal or not other.journal:
            return False
        return self.journal < other.journal and self.content < other.content


@app.command()
def tasks(
    tag_or_page: List[str] = typer.Argument(None, metavar="TAG", help="Tags or pages to query"),
    format: TaskFormat = typer.Option(
        TaskFormat.text, '--format', '--pretty', '-f', help="Output format", case_sensitive=False
    ),
    logseq_host_url=typer.Option(..., '--host', '-h', help="Logseq host", envvar="LOGSEQ_HOST_URL"),
    logseq_api_token=typer.Option(..., '--token', '-t', help="Logseq API token", envvar="LOGSEQ_API_TOKEN"),
):
    """List tasks in Logseq."""
    condition = ""
    if tag_or_page:
        if len(tag_or_page) == 1:
            condition = f" [[{tag_or_page[0]}]]"
        else:
            pages = " ".join([f"[[{tp}]]" for tp in tag_or_page])
            condition = f" (or {pages})"
    req_query = f"(and{condition} (task TODO NOW DOING))"
    # FIXME[AA]: move to class LogseqApi
    session = requests.Session()
    session.headers.update(
        {
            "Authorization": f"Bearer {logseq_api_token}",
            "Content-Type": "application/json",
        }
    )
    resp = session.post(f"{logseq_host_url}/api", json={"method": "logseq.db.q", "args": [req_query]})
    resp.raise_for_status()

    block_url = "logseq://graph/captains-log?block-id="
    rows: List[OneLineRow] = []
    for obj in resp.json():
        uuid = obj.get("uuid")
        content = obj.get("content").splitlines()[0]
        page = obj.get("page", {})
        journal = page.get("journalDay")
        name = page.get("originalName")
        if format == TaskFormat.text:
            rows.append(OneLineRow(journal, name, f"{block_url}{uuid}", content))

    for row in sorted(rows):
        typer.secho(f"{row.name}: ", fg=typer.colors.GREEN, nl=False)
        typer.secho(row.url, fg=typer.colors.BLUE, nl=False)
        typer.echo(f" {row.content}")
