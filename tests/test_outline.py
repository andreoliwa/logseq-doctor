from textwrap import dedent

import mistletoe
from mistletoe.ast_renderer import ASTRenderer
from typer.testing import CliRunner

from logseq_doctor import flat_markdown_to_outline
from logseq_doctor.cli import app, lsd, outline
from logseq_doctor.constants import NBSP


def test_cli_help():
    """The Typer output is a colourful rich text, so let's only assert the presence of commands."""
    # FIXME: this test is useless, it's testing Typer actually
    runner = CliRunner()
    result = runner.invoke(app, [])
    for expected_text in (lsd.__doc__, outline.__doc__):
        assert expected_text in result.output
    assert result.exit_code == 0


def assert_markdown(flat_md: str, outlined_md: str, *, ast=False):
    """Assert flat Markdown is converted to outline.

    Use non-breaking spaces to trick dedent() into keeping leading spaces on output.
    """
    output_without_nbsp = dedent(outlined_md).lstrip().replace(NBSP, " ")
    stripped_md = dedent(flat_md).lstrip()

    # For debugging purposes
    if ast:
        print("\nASTRenderer:\n" + mistletoe.markdown(stripped_md, ASTRenderer))

    assert flat_markdown_to_outline(stripped_md) == output_without_nbsp


def test_header_hierarchy_preserved_and_whitespace_removed():
    assert_markdown(
        """
        #  Header 1


        -  Item 1

        -  Item 2

        ## Header 2

        - Item 3
        ###  Header 3
        -  Item 4
        """,
        """
        - # Header 1
          - Item 1
          - Item 2
          - ## Header 2
            - Item 3
            - ### Header 3
              - Item 4
        """,
    )


def test_links():
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


def test_flat_paragraphs_without_header():
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


def test_flat_paragraphs_with_header():
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


def test_flat_paragraphs_with_deeper_headers():
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


def test_nested_lists_single_level():
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


def test_nested_lists_multiple_levels():
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


def test_thematic_break_setext_heading():
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
