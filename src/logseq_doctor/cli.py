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
import uuid
from enum import Enum
from pathlib import Path
from textwrap import dedent
from textwrap import indent
from typing import List

import typer

from logseq_doctor import flat_markdown_to_outline
from logseq_doctor.api import LogseqApi

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
    kanban = 'kanban'


@app.command()
def tasks(
    tag_or_page: List[str] = typer.Argument(None, metavar="TAG", help="Tags or pages to query"),
    format: TaskFormat = typer.Option(
        TaskFormat.text, '--format', '--pretty', '-f', help="Output format", case_sensitive=False
    ),
    logseq_host_url=typer.Option(..., '--host', '-h', help="Logseq host", envvar="LOGSEQ_HOST_URL"),
    logseq_api_token=typer.Option(..., '--token', '-t', help="Logseq API token", envvar="LOGSEQ_API_TOKEN"),
    logseq_graph=typer.Option(..., '--graph', '-g', help="Logseq graph", envvar="LOGSEQ_GRAPH"),
):
    """List tasks in Logseq."""
    condition = ""
    if tag_or_page:
        if len(tag_or_page) == 1:
            condition = f" [[{tag_or_page[0]}]]"
        else:
            pages = " ".join([f"[[{tp}]]" for tp in tag_or_page])
            condition = f" (or {pages})"
    api = LogseqApi(logseq_host_url, logseq_api_token, logseq_graph)
    query = f"(and{condition} (task TODO DOING WAITING NOW LATER))"
    rows = sorted(api.query(query), key=lambda row: (row.journal_iso_date, row.content))
    if format == TaskFormat.text:
        _output_text(rows)
    elif format == TaskFormat.kanban:
        _output_kanban(rows)


def _output_text(rows):
    for row in rows:
        typer.secho(f"{row.name}: ", fg=typer.colors.GREEN, nl=False)
        typer.secho(row.url, fg=typer.colors.BLUE, nl=False)
        typer.echo(f" {row.content}")


def _output_kanban(rows):
    block_id = uuid.uuid4()
    renderer = "{{renderer :kboard, %s, kanban-list}}" % block_id
    title = "My board"
    header = f"""
    - {renderer}
    - {title}
      id:: {block_id}
      collapsed:: true
      - placeholder #.kboard-placeholder
        kanban-list:: Triage
      - placeholder #.kboard-placeholder
        kanban-list:: TODO
      - placeholder #.kboard-placeholder
        kanban-list:: DOING
      - placeholder #.kboard-placeholder
        kanban-list:: DONE
    """
    typer.echo(dedent(header).strip())
    for row in rows:
        list_item = f"""
        - {row.content}
          kanban-list:: Triage
          collapsed:: true
          - (({row.block_id}))
        """
        typer.echo(indent(dedent(list_item).strip(), " " * 2))
