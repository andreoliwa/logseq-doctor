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
from typing import TextIO

import click

from logseq_doctor import flat_markdown_to_outline


@click.group()
def main():
    """Logseq Doctor: heal your flat old Markdown files before importing them."""
    pass


@main.command()
@click.argument('file', type=click.File())
def outline(file: TextIO):
    """Convert flat Markdown to outline."""
    print(flat_markdown_to_outline(file.read()))
