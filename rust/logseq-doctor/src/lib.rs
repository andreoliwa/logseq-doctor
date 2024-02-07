//! Logseq Doctor: heal your Markdown files
//!
//! Python extension written in Rust, until the whole project is ported to Rust.
use chrono::NaiveDate;
use pyo3::prelude::*;
use pyo3::types::PyDate;
use std::path::PathBuf;

#[pymodule]
fn rust_ext(_python: Python, module: &PyModule) -> PyResult<()> {
    module.add_function(wrap_pyfunction!(remove_consecutive_spaces, module)?)?;
    module.add_function(wrap_pyfunction!(add_content, module)?)?;
    Ok(())
}

#[pyfunction]
fn remove_consecutive_spaces(file_contents: String) -> PyResult<String> {
    Ok(logseq::remove_consecutive_spaces(file_contents).unwrap())
}

#[pyfunction]
fn add_content(
    graph_path: PathBuf,
    markdown: String,
    prepend: bool,
    parsed_date: Option<&PyDate>,
) -> PyResult<()> {
    let naive_date = match parsed_date {
        None => None,
        Some(pydate) => pydate_to_naivedate(pydate)
            .unwrap_or_else(|err| panic!("Error {} while parsing date {:?}", err, pydate)),
    };
    let journal = logseq::Journal::new(graph_path, naive_date);
    if prepend {
        journal
            .prepend(markdown)
            .expect("Failed to prepend content to journal");
        return Ok(());
    }
    journal
        .append(markdown)
        .expect("Failed to append content to journal");
    Ok(())
}

fn pydate_to_naivedate(pydate: &PyDate) -> PyResult<Option<NaiveDate>> {
    let year = pydate.getattr("year")?.extract::<i32>()?;
    let month = pydate.getattr("month")?.extract::<u32>()?;
    let day = pydate.getattr("day")?.extract::<u32>()?;

    Ok(NaiveDate::from_ymd_opt(year, month, day))
}
