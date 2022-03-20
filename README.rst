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

.. |commits-since| image:: https://img.shields.io/github/commits-since/andreoliwa/logseq-doctor/v0.1.0.svg
    :alt: Commits since latest release
    :target: https://github.com/andreoliwa/logseq-doctor/compare/v0.1.0...master



.. end-badges

Heal your old Markdown files to use them with Logseq

* Free software: MIT license

Installation
============

::

    pip install logseq-doctor

You can also install the in-development version with::

    pip install https://github.com/andreoliwa/logseq-doctor/archive/master.zip


Documentation
=============


https://logseq-doctor.readthedocs.io/


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
