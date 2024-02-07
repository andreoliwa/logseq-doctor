"""Logseq Doctor: heal your Markdown files."""

from __future__ import annotations

import os

import mistletoe
from mistletoe import block_token, span_token, token
from mistletoe.base_renderer import BaseRenderer

from logseq_doctor.constants import CHAR_DASH

__version__ = "0.3.0"


class LogseqRenderer(BaseRenderer):
    """Render Markdown as an outline with bullets, like Logseq expects."""

    def __init__(self, *extras: token.Token) -> None:
        super().__init__(*extras)
        self.current_level = 0
        self.bullet = "-"

    def outline(self, indent: int, text: str, *, nl: bool = True) -> str:
        """Render a line of text with the correct indentation."""
        leading_spaces = "  " * indent
        new_line_at_the_end = os.linesep if nl else ""
        return f"{leading_spaces}{self.bullet} {text}{new_line_at_the_end}"

    def render_heading(self, token: block_token.Heading | block_token.SetextHeading) -> str:
        """Setext headings: https://spec.commonmark.org/0.30/#setext-headings."""
        if isinstance(token, block_token.SetextHeading):
            # For now, only dealing with level 2 setext headers (dashes)
            return self.render_inner(token) + f"{os.linesep}{CHAR_DASH * 3}{os.linesep}"

        self.current_level = token.level
        hashes = "#" * token.level
        inner = self.render_inner(token)
        return self.outline(token.level - 1, f"{hashes} {inner}")

    def render_line_break(self, token: span_token.LineBreak) -> str:
        """Render a line break."""
        return token.content + os.linesep

    def render_paragraph(self, token: block_token.Paragraph) -> str:
        """Render a paragraph with the correct indentation."""
        input_lines = self.render_inner(token).strip().splitlines()
        output_lines: list[str] = [self.outline(self.current_level, line, nl=False) for line in input_lines]
        return os.linesep.join(output_lines) + os.linesep

    def render_link(self, token: span_token.Link) -> str:
        """Render a link as a Markdown link."""
        text = self.render_inner(token)
        url = token.target
        return f"[{text}]({url})"

    def render_list_item(self, token: block_token.ListItem) -> str:
        """Render a list item with the correct indentation."""
        if len(token.children) <= 1:
            return self.render_inner(token)

        self.current_level += 1

        inner = self.render_inner(token)
        headless_parent_with_children = inner.lstrip(f"{self.bullet} ")
        value_before_changing_level = self.outline(self.current_level - 1, headless_parent_with_children, nl=False)

        self.current_level -= 1
        return value_before_changing_level

    def render_thematic_break(self, token: block_token.ThematicBreak) -> str:  # noqa: ARG002
        """Render a horizontal rule as a line of dashes."""
        return f"{CHAR_DASH * 3}{os.linesep}"

    # TODO: refactor: the methods below are placeholders taken from BaseRenderer.render_map.
    #  - Uncomment them to use them during debugging.
    #  - Remove them when there will be enough test coverage for all the different elements below
    # def render_strong(self, token):
    #     return self.render_inner(token)
    #
    # def render_emphasis(self, token):
    #     return self.render_inner(token)
    #
    # def render_inline_code(self, token):
    #     return self.render_inner(token)
    #
    # def render_raw_text(self, token):
    #     return self.render_inner(token)
    #
    # def render_strikethrough(self, token):
    #     return self.render_inner(token)
    #
    # def render_image(self, token):
    #     return self.render_inner(token)
    #
    # def render_auto_link(self, token):
    #     return self.render_inner(token)
    #
    # def render_escape_sequence(self, token):
    #     return self.render_inner(token)
    #
    # def render_quote(self, token):
    #     return self.render_inner(token)
    #
    # def render_block_code(self, token):
    #     return self.render_inner(token)
    #
    # def render_list(self, token):
    #     return self.render_inner(token)
    #
    # def render_table(self, token):
    #     return self.render_inner(token)
    #
    # def render_table_row(self, token):
    #     return self.render_inner(token)
    #
    # def render_table_cell(self, token):
    #     return self.render_inner(token)
    #
    # def render_document(self, token):
    #     return self.render_inner(token)


def flat_markdown_to_outline(markdown_contents: str) -> str:
    """Convert flat Markdown to an outline."""
    return mistletoe.markdown(markdown_contents, LogseqRenderer)
