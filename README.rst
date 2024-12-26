========
Overview
========

.. start-badges

.. list-table::
    :stub-columns: 1

    * - docs
      - |docs|
    * - tests
      - | |github-actions|
        | |codecov|
    * - package
      - | |version| |wheel| |supported-versions| |supported-implementations|
        | |commits-since|
.. |docs| image:: https://readthedocs.org/projects/logseq-doctor/badge/?style=flat
    :target: https://logseq-doctor.readthedocs.io/
    :alt: Documentation Status

.. |github-actions| image:: https://github.com/andreoliwa/logseq-doctor/actions/workflows/github-actions.yml/badge.svg
    :alt: GitHub Actions Build Status
    :target: https://github.com/andreoliwa/logseq-doctor/actions

.. |codecov| image:: https://codecov.io/gh/andreoliwa/logseq-doctor/branch/master/graphs/badge.svg?branch=master
    :alt: Coverage Status
    :target: https://codecov.io/github/andreoliwa/logseq-doctor

.. |version| image:: https://img.shields.io/pypi/v/logseq-doctor.svg
    :alt: PyPI Package latest release
    :target: https://pypi.org/project/logseq-doctor

.. |wheel| image:: https://img.shields.io/pypi/wheel/logseq-doctor.svg
    :alt: PyPI Wheel
    :target: https://pypi.org/project/logseq-doctor

.. |supported-versions| image:: https://img.shields.io/pypi/pyversions/logseq-doctor.svg
    :alt: Supported versions
    :target: https://pypi.org/project/logseq-doctor

.. |supported-implementations| image:: https://img.shields.io/pypi/implementation/logseq-doctor.svg
    :alt: Supported implementations
    :target: https://pypi.org/project/logseq-doctor

.. |commits-since| image:: https://img.shields.io/github/commits-since/andreoliwa/logseq-doctor/v0.3.0.svg
    :alt: Commits since latest release
    :target: https://github.com/andreoliwa/logseq-doctor/compare/v0.3.0...master



.. end-badges

Logseq Doctor: heal your flat old Markdown files before importing them.

.. warning::

    This project is still alpha, so it's very rough on the edges (documentation and feature-wise).

    At the moment, it has a Python package shipped with a Rust module, plus an external Go executable with recent additions.

    The long-term plan is to convert it to Go and slowly remove Rust and Python.
    New features will be added to the Go executable only.

Installation
============

The recommended way is to install ``logseq-doctor`` globally with `pipx <https://github.com/pypa/pipx>`_::

    pipx install logseq-doctor

You can also install the development version with::

    pipx install git+https://github.com/andreoliwa/logseq-doctor

You will then have the ``lsd`` command available globally in your system.

If you want to use the ``lsd tidy-up`` command, (for now) you will need to manually install the Go executable from the latest ``master`` branch::

    go install github.com/andreoliwa/logseq-doctor@f7c0f0f  # use the latest commit hash after the @

``lsdg`` is the expected name for the Go executable, so you need to rename it::

    mv $(go env GOPATH)/bin/logseq-doctor $(go env GOPATH)/bin/lsdg

Confirm if it's in your path::

    ls -l $(go env GOPATH)/bin/lsdg

Quick start
===========

Type ``lsd`` without arguments to check the current commands and options::

     Usage: lsd [OPTIONS] COMMAND [ARGS]...

     Logseq Doctor: heal your flat old Markdown files before importing them.

    ╭─ Options ────────────────────────────────────────────────────────────────────╮
    │ --install-completion          Install completion for the current shell.      │
    │ --show-completion             Show completion for the current shell, to copy │
    │                               it or customize the installation.              │
    │ --help                        Show this message and exit.                    │
    ╰──────────────────────────────────────────────────────────────────────────────╯
    ╭─ Commands ───────────────────────────────────────────────────────────────────╮
    │ outline  Convert flat Markdown to outline.                                   │
    │ tasks    List tasks in Logseq.                                               │
    │ tidy-up  Tidy up your Markdown files by removing empty bullets in any block. │
    ╰──────────────────────────────────────────────────────────────────────────────╯

Development
===========

To run all the tests run::

    tox

Note, to combine the coverage data from all the tox environments run:

.. list-table::
    :widths: 10 90
    :stub-columns: 1

    - - Windows
      - ::

            set PYTEST_ADDOPTS=--cov-append
            tox

    - - Other
      - ::

            PYTEST_ADDOPTS=--cov-append tox
