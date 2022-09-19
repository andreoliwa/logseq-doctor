from textwrap import dedent

from click.testing import CliRunner

from logseq_doctor import flat_markdown_to_outline
from logseq_doctor.cli import main

NBSP = "\\u00A0"


def test_main():
    runner = CliRunner()
    result = runner.invoke(main, [])

    assert (
        result.output
        == dedent(
            """
        Usage: main [OPTIONS] COMMAND [ARGS]...

          Logseq Doctor: heal your flat old Markdown files before importing them.

        Options:
          --help  Show this message and exit.

        Commands:
          outline  Convert flat Markdown to outline.
        """
        ).lstrip()
    )
    assert result.exit_code == 0


def assert_markdown(flat_md: str, outlined_md: str):
    """Assert flat Markdown is converted to outline.

    Use non-breaking spaces to trick dedent() into keeping leading spaces on output.
    """
    output_without_nbsp = dedent(outlined_md).lstrip().replace(NBSP, " ")
    assert flat_markdown_to_outline(dedent(flat_md).lstrip()) == output_without_nbsp


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


def test_single_nested_lists():
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
