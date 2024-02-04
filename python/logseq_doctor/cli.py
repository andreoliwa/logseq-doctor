"""Module that contains the command line app.

Why does this file exist, and why not put this in __main__?

  You might be tempted to import things from __main__ later, but that will cause
  problems: the code will get executed twice:

  - When you run `python -m logseq_doctor` python will execute
    ``__main__.py`` as a script. That means there won't be any
    ``logseq_doctor.__main__`` in ``sys.modules``.
  - When you import __main__ it will get executed again (as a module) because
    there's no ``logseq_doctor.__main__`` in ``sys.modules``.

  Also see (1) from https://click.pocoo.org/5/setuptools/#setuptools-integration
"""

from __future__ import annotations

import re
from enum import Enum
from pathlib import Path  # noqa: TCH003 Typer needs this import to infer the type of the argument
from typing import List

import typer

from logseq_doctor import flat_markdown_to_outline, rust_ext
from logseq_doctor.api import Block, Logseq

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
    """Tidy up your Markdown files by removing empty bullets and double spaces in any block."""
    for each_file in markdown_file:
        old_contents = each_file.read_text()
        rm_empty_bullets = re.sub(r"(\n\s*-\s*$)", "", old_contents, flags=re.MULTILINE)
        rm_double_spaces = rust_ext.remove_consecutive_spaces(rm_empty_bullets)
        if old_contents != rm_double_spaces:
            typer.echo(f"removed empty bullets and double spaces from {each_file}")
            each_file.write_text(rm_double_spaces)


class TaskFormat(str, Enum):
    """Task format."""

    text = "text"


@app.command()
def tasks(
    tag_or_page: List[str] = typer.Argument(None, metavar="TAG", help="Tags or pages to query"),
    logseq_host_url: str = typer.Option(..., "--host", "-h", help="Logseq host", envvar="LOGSEQ_HOST_URL"),
    logseq_api_token: str = typer.Option(..., "--token", "-t", help="Logseq API token", envvar="LOGSEQ_API_TOKEN"),
    logseq_graph_path: Path = typer.Option(
        ...,
        "--graph",
        "-g",
        help="Logseq graph",
        envvar="LOGSEQ_GRAPH_PATH",
        dir_okay=True,
        file_okay=False,
    ),
) -> None:
    """List tasks in Logseq."""
    logseq = Logseq(logseq_host_url, logseq_api_token, logseq_graph_path)
    condition = ""
    if tag_or_page:
        if len(tag_or_page) == 1:
            condition = f" [[{tag_or_page[0]}]]"
        else:
            pages = " ".join([f"[[{tp}]]" for tp in tag_or_page])
            condition = f" (or {pages})"
    query = f"(and{condition} (task TODO DOING WAITING NOW LATER))"

    blocks_sorted_by_date = Block.sort_by_date(logseq.query(query))
    for block in blocks_sorted_by_date:
        typer.secho(f"{block.page_title}: ", fg=typer.colors.GREEN, nl=False)
        typer.secho(block.url(logseq.graph_name), fg=typer.colors.BLUE, bold=True, nl=False)
        typer.echo(f" {block.raw_content}")
