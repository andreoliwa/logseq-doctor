//! Logseq Doctor: heal your Markdown files
//!
//! Python extension written in Rust, until the whole project is ported to Rust.
use pyo3::prelude::*;
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
fn add_content(graph_path: PathBuf, markdown: String) -> PyResult<()> {
    let journal = logseq::Journal::new(graph_path, None);
    journal
        .append(markdown)
        .expect("Failed to append content to journal");
    Ok(())
}
