import mistletoe
from mistletoe import block_token
from mistletoe import span_token
from mistletoe.base_renderer import BaseRenderer

__version__ = '0.1.1'


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

    def render_heading(self, token: block_token.Heading):
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

    # TODO: refactor: the methods below are placeholders taken from BaseRenderer.render_map.
    #  - Uncomment them to use them during debugging.
    #  - Remove them when there will be enough test coverage for all the different elements below
    # def render_strong(self, token):
    #     pass
    #
    # def render_emphasis(self, token):
    #     pass
    #
    # def render_inline_code(self, token):
    #     pass
    #
    # def render_raw_text(self, token):
    #     pass
    #
    # def render_strikethrough(self, token):
    #     pass
    #
    # def render_image(self, token):
    #     pass
    #
    # def render_auto_link(self, token):
    #     pass
    #
    # def render_escape_sequence(self, token):
    #     pass
    #
    # def render_quote(self, token):
    #     pass
    #
    # def render_block_code(self, token):
    #     pass
    #
    # def render_list(self, token):
    #     pass
    #
    # def render_list_item(self, token):
    #     pass
    #
    # def render_table(self, token):
    #     pass
    #
    # def render_table_row(self, token):
    #     pass
    #
    # def render_table_cell(self, token):
    #     pass
    #
    # def render_thematic_break(self, token):
    #     pass
    #
    # def render_document(self, token):
    #     pass


def flat_markdown_to_outline(markdown_contents: str) -> str:
    """Convert flat Markdown to an outline."""
    return mistletoe.markdown(markdown_contents, LogseqRenderer)
