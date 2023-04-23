"""Module that contains the command line app.

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
from enum import Enum
from pathlib import Path
from typing import List

import typer

from logseq_doctor import flat_markdown_to_outline
from logseq_doctor.api import Block, Kanban, Logseq, Page

app = typer.Typer(no_args_is_help=True)


@app.callback()
def lsd() -> None:
    """Logseq Doctor: heal your flat old Markdown files before importing them."""


@app.command(no_args_is_help=True)
def outline(text_file: typer.FileText) -> None:
    """Convert flat Markdown to outline."""
    typer.echo(flat_markdown_to_outline(text_file.read()))


@app.command()
def tidy_up(
    markdown_file: List[Path] = typer.Argument(
        ...,
        help="Markdown files to tidy up",
        exists=True,
        file_okay=True,
        dir_okay=False,
        writable=True,
    ),
) -> None:
    """Tidy up your Markdown files by removing empty bullets in any block."""
    for each_file in markdown_file:
        old_contents = each_file.read_text()
        new_contents = re.sub(r"(\n\s*-\s*$)", "", old_contents, flags=re.MULTILINE)
        if old_contents != new_contents:
            typer.echo(f"removed empty bullets from {each_file}")
            each_file.write_text(new_contents)


class TaskFormat(str, Enum):
    """Task format."""

    text = "text"
    kanban = "kanban"


@app.command()
def tasks(
    tag_or_page: List[str] = typer.Argument(None, metavar="TAG", help="Tags or pages to query"),
    logseq_host_url: str = typer.Option(..., "--host", "-h", help="Logseq host", envvar="LOGSEQ_HOST_URL"),
    logseq_api_token: str = typer.Option(..., "--token", "-t", help="Logseq API token", envvar="LOGSEQ_API_TOKEN"),
    logseq_graph: str = typer.Option(..., "--graph", "-g", help="Logseq graph", envvar="LOGSEQ_GRAPH"),
    format_: TaskFormat = typer.Option(
        TaskFormat.text,
        "--format",
        "--pretty",
        "-f",
        help="Output format",
        case_sensitive=False,
    ),
    output_path: Path = typer.Option(
        None,
        "--output",
        "-o",
        help="Output path for the Kanban",
        file_okay=True,
        dir_okay=False,
        writable=True,
        resolve_path=True,
    ),
) -> None:
    """List tasks in Logseq."""
    if format_ == TaskFormat.kanban and not output_path:
        typer.secho("Kanban format requires an output path", fg=typer.colors.BRIGHT_RED)
        raise typer.Exit(1)

    condition = ""
    if tag_or_page:
        if len(tag_or_page) == 1:
            condition = f" [[{tag_or_page[0]}]]"
        else:
            pages = " ".join([f"[[{tp}]]" for tp in tag_or_page])
            condition = f" (or {pages})"
    logseq = Logseq(logseq_host_url, logseq_api_token, logseq_graph)
    query = f"(and{condition} (task TODO DOING WAITING NOW LATER))"

    blocks_by_date = Block.sort_by_date(logseq.query(query))

    if format_ == TaskFormat.kanban:
        page = Page(output_path)
        kanban = Kanban(page, blocks_by_date)
        if kanban.find():
            typer.echo(f"Kanban board being updated at {output_path}")
            kanban.update()
        else:
            typer.echo(f"Kanban board being added to {output_path}")
            try:
                kanban.add()
            except FileNotFoundError as err:
                typer.secho(str(err), fg=typer.colors.BRIGHT_RED)
                typer.secho("Add some content to the page and try again: ", fg=typer.colors.BRIGHT_RED, nl=False)
                typer.secho(page.url(logseq.graph), fg=typer.colors.BLUE)
                raise typer.Exit(1) from err

        typer.secho("✨ Done.", fg=typer.colors.BRIGHT_WHITE, bold=True)
        return

    for block in blocks_by_date:
        typer.secho(f"{block.page_title}: ", fg=typer.colors.GREEN, nl=False)
        typer.secho(block.url(logseq.graph), fg=typer.colors.BLUE, nl=False)
        typer.echo(f" {block.raw_content}")
