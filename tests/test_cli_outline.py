import os
from pathlib import Path
from textwrap import dedent

import mistletoe
from logseq_doctor import flat_markdown_to_outline
from logseq_doctor.cli import app
from logseq_doctor.constants import NBSP
from mistletoe.ast_renderer import ASTRenderer
from typer.testing import CliRunner


def assert_markdown(flat_md: str, outlined_md: str, *, ast: bool = False) -> None:
    """Assert flat Markdown is converted to outline.

    Use non-breaking spaces to trick dedent() into keeping leading spaces on output.
    """
    output_without_nbsp = dedent(outlined_md).lstrip().replace(NBSP, " ")
    stripped_md = dedent(flat_md).lstrip()

    # For debugging purposes
    if ast:  # pragma: no cover
        print(f"{os.linesep}ASTRenderer:{os.linesep}" + mistletoe.markdown(stripped_md, ASTRenderer))

    assert flat_markdown_to_outline(stripped_md) == output_without_nbsp


def test_header_hierarchy_preserved_and_whitespace_removed(datadir: Path) -> None:
    result = CliRunner().invoke(app, ["outline", str(datadir / "dirty.md")])
    assert result.exit_code == 0
    assert result.stdout == (datadir / "clean.md").read_text() + os.linesep


def test_links() -> None:
    assert_markdown(
        """
        #  Header

        -  [Link only](https://example.com)
        -   Text before, then [a link](https://link.com), then text after
        - Only text before, [link a the end](https://endlink.com)
        """,
        """
        - # Header
          - [Link only](https://example.com)
          - Text before, then [a link](https://link.com), then text after
          - Only text before, [link a the end](https://endlink.com)
        """,
    )


def test_flat_paragraphs_without_header() -> None:
    assert_markdown(
        """
        Some flat paragraph.

        [Link only](https://example.com).
        Text before, then [a link](https://link.com), then text after.

        Only text before, [link a the end](https://endlink.com).
        """,
        """
        - Some flat paragraph.
        - [Link only](https://example.com).
        - Text before, then [a link](https://link.com), then text after.
        - Only text before, [link a the end](https://endlink.com).
        """,
    )


def test_flat_paragraphs_with_header() -> None:
    assert_markdown(
        """
        # Some sneaky header
        Some flat paragraph.

        [Link only](https://example.com).
        Text before, then [a link](https://link.com), then text after.

        Only text before, [link a the end](https://endlink.com).
        """,
        """
        - # Some sneaky header
          - Some flat paragraph.
          - [Link only](https://example.com).
          - Text before, then [a link](https://link.com), then text after.
          - Only text before, [link a the end](https://endlink.com).
        """,
    )


def test_flat_paragraphs_with_deeper_headers() -> None:
    assert_markdown(
        """
        ## Some sneaky h2 without h1
        Some flat paragraph.

        [Link only](https://example.com).
        Text before, then [a link](https://link.com), then text after.

        Only text before, [link a the end](https://endlink.com).
        """,
        f"""
        {NBSP * 2}- ## Some sneaky h2 without h1
            - Some flat paragraph.
            - [Link only](https://example.com).
            - Text before, then [a link](https://link.com), then text after.
            - Only text before, [link a the end](https://endlink.com).
        """,
    )


def test_nested_lists_single_level() -> None:
    assert_markdown(
        """
        # Header

        - Parent
          - Child 1
          - Child 2
        """,
        """
        - # Header
          - Parent
            - Child 1
            - Child 2
        """,
    )


def test_nested_lists_multiple_levels() -> None:
    assert_markdown(
        """
        # Header

        - Parent
          - Child 1
            - Grand child 1.1
            - Grand child 1.2
            - Grand child 1.3
          - Child 2
            - Grand child 2.1
              - ABC
        """,
        """
        - # Header
          - Parent
            - Child 1
              - Grand child 1.1
              - Grand child 1.2
              - Grand child 1.3
            - Child 2
              - Grand child 2.1
                - ABC
        """,
    )


def test_thematic_break_setext_heading() -> None:
    assert_markdown(
        """
        ---
        date: 2021-10-29T09:41:12.490Z
        dateCreated: 2021-10-14T20:48:58.837Z
        ---

        # Some title

        Line1
        Line2
        """,
        """
        ---
        date: 2021-10-29T09:41:12.490Z
        dateCreated: 2021-10-14T20:48:58.837Z
        ---
        - # Some title
          - Line1
          - Line2
        """,
    )
