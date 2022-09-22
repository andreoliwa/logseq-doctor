from __future__ import annotations

import mistletoe
from mistletoe import block_token
from mistletoe import span_token
from mistletoe.base_renderer import BaseRenderer

__version__ = '0.1.1'

DASH = "-"


class LogseqRenderer(BaseRenderer):
    """Render Markdown as an outline with bullets, like Logseq expects."""

    def __init__(self, *extras):
        super().__init__(*extras)
        self.current_level = 0
        self.bullet = "-"

    def outline(self, indent: int, text: str, *, nl=True) -> str:
        leading_spaces = '  ' * indent
        new_line_at_the_end = "\n" if nl else ""
        return f"{leading_spaces}{self.bullet} {text}{new_line_at_the_end}"

    def render_heading(self, token: block_token.Heading | block_token.SetextHeading):
        """Setext headings: https://spec.commonmark.org/0.30/#setext-headings."""
        if isinstance(token, block_token.SetextHeading):
            # For now, only dealing with level 2 setext headers (dashes)
            return self.render_inner(token) + f"\n{DASH * 3}\n"

        self.current_level = token.level
        hashes = '#' * token.level
        inner = self.render_inner(token)
        return self.outline(token.level - 1, f"{hashes} {inner}")

    def render_line_break(self, token: span_token.LineBreak) -> str:
        return token.content + '\n'

    def render_paragraph(self, token):
        input_lines = self.render_inner(token).strip().splitlines()
        output_lines: list[str] = [self.outline(self.current_level, line, nl=False) for line in input_lines]
        return '\n'.join(output_lines) + "\n"

    def render_link(self, token: span_token.Link):
        text = self.render_inner(token)
        url = token.target
        return f"[{text}]({url})"

    def render_list_item(self, token):
        if len(token.children) <= 1:
            return self.render_inner(token)

        self.current_level += 1

        inner = self.render_inner(token)
        headless_parent_with_children = inner.lstrip(f"{self.bullet} ")
        rv = self.outline(self.current_level - 1, headless_parent_with_children, nl=False)

        self.current_level -= 1
        return rv

    def render_thematic_break(self, token: block_token.ThematicBreak) -> str:
        return f'{DASH * 3}\n'

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
