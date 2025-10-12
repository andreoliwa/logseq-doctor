"""Generate the code reference pages and navigation for Python API documentation."""

from pathlib import Path

import mkdocs_gen_files

nav = mkdocs_gen_files.Nav()

# Path to the Python source code
src_root = Path("src")
package_name = "logseq_doctor"

for path in sorted((src_root / package_name).rglob("*.py")):
    module_path = path.relative_to(src_root).with_suffix("")
    doc_path = path.relative_to(src_root / package_name).with_suffix(".md")
    full_doc_path = Path("reference", "python", doc_path)

    parts = tuple(module_path.parts)

    # Skip __pycache__ and test files
    if "__pycache__" in parts or "test_" in path.name:
        continue

    # Add to navigation
    nav[parts] = doc_path.as_posix()

    with mkdocs_gen_files.open(full_doc_path, "w") as fd:
        # For __init__.py files, use the package name without __init__
        ident = ".".join(parts[:-1]) if path.name == "__init__.py" else ".".join(parts)
        fd.write(f"# {ident}\n\n")
        fd.write(f"::: {ident}\n")

    mkdocs_gen_files.set_edit_path(full_doc_path, path)

# Write the navigation file
with mkdocs_gen_files.open("reference/python/SUMMARY.md", "w") as nav_file:
    nav_file.writelines(nav.build_literate_nav())
