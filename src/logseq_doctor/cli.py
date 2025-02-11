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

import os
import subprocess
from dataclasses import dataclass
from enum import Enum
from pathlib import Path  # Typer needs this import to infer the type of the argument
from typing import cast

import typer

from logseq_doctor import flat_markdown_to_outline
from logseq_doctor.api import Block, Logseq

app = typer.Typer(no_args_is_help=True)


@dataclass
class GlobalOptions:
    """Global options for every sub-command."""

    logseq_graph_path: Path


@app.callback()
def lsd(
    ctx: typer.Context,
    logseq_graph_path: Path = typer.Option(  # noqa: B008
        ...,
        "--graph",
        "-g",
        help="Logseq graph",
        envvar="LOGSEQ_GRAPH_PATH",
        dir_okay=True,
        file_okay=False,
    ),
) -> None:
    """Logseq Doctor: heal your flat old Markdown files before importing them."""
    ctx.obj = GlobalOptions(logseq_graph_path)


@app.command(no_args_is_help=True)
def outline(text_file: typer.FileText) -> None:
    """Convert flat Markdown to outline."""
    typer.echo(flat_markdown_to_outline(text_file.read()))


def _call_golang_exe(command: str, markdown_file: Path) -> int:
    display_errors = not bool(os.environ.get("LOGSEQ_GO_IGNORE_ERRORS", False))

    exe_str = os.environ.get("LOGSEQ_GO_EXE_PATH", "~/go/bin/lsdg")
    if not exe_str:
        if display_errors:
            typer.secho("The environment variable 'LOGSEQ_GO_EXE_PATH' is not set.", fg=typer.colors.RED, bold=True)
            return 1
        return 0

    # TODO: install the Go executable when the Python package is installed
    exe_path = Path(exe_str).expanduser()
    if not exe_path.exists():
        if display_errors:
            typer.secho(f"The executable '{exe_path}' does not exist.", fg=typer.colors.RED, bold=True)
        return 2
    if not os.access(exe_path, os.X_OK):
        if display_errors:
            typer.secho(f"The file '{exe_path}' is not executable.", fg=typer.colors.RED, bold=True)
        return 3
    try:
        result = subprocess.run([exe_path, command, str(markdown_file)], check=False)  # noqa: S603
    except subprocess.CalledProcessError as err:
        if display_errors:
            typer.secho(f"An error occurred while running the executable: {err}")
        return 4
    else:
        return result.returncode


class TaskFormat(str, Enum):
    """Task format."""

    text = "text"


@app.command()
def tasks(
    ctx: typer.Context,
    tag_or_page: list[str] = typer.Argument(None, metavar="TAG", help="Tags or pages to query"),  # noqa: B008
    logseq_host_url: str = typer.Option(..., "--host", "-h", help="Logseq host", envvar="LOGSEQ_HOST_URL"),
    logseq_api_token: str = typer.Option(..., "--token", "-t", help="Logseq API token", envvar="LOGSEQ_API_TOKEN"),
    json_: bool = typer.Option(False, "--json", help="Output in JSON format"),
) -> None:
    """List tasks in Logseq."""
    logseq = Logseq(logseq_host_url, logseq_api_token, cast(GlobalOptions, ctx.obj).logseq_graph_path)
    condition = ""
    if tag_or_page:
        if len(tag_or_page) == 1:
            condition = f" [[{tag_or_page[0]}]]"
        else:
            pages = " ".join([f"[[{tp}]]" for tp in tag_or_page])
            condition = f" (or {pages})"
    query = f"(and{condition} (task TODO DOING WAITING NOW LATER))"

    if json_:
        typer.echo(logseq.query_json(query))
        return

    blocks_sorted_by_date = Block.sort_by_date(logseq.query_blocks(query))
    for block in blocks_sorted_by_date:
        typer.secho(f"{block.page_title}§", fg=typer.colors.GREEN, nl=False)
        typer.secho(block.url(logseq.graph_name), fg=typer.colors.BLUE, bold=True, nl=False)
        typer.echo(f"§{block.raw_content}")
