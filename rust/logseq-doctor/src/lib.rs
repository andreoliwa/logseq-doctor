//! Logseq Doctor: heal your Markdown files
//!
//! Python extension written in Rust, until the whole project is ported to Rust.
use pyo3::prelude::*;

#[pymodule]
fn _logseq_doctor(_python: Python, module: &PyModule) -> PyResult<()> {
    module.add_function(wrap_pyfunction!(rust_remove_consecutive_spaces, module)?)?;
    Ok(())
}

#[pyfunction]
fn rust_remove_consecutive_spaces(file_contents: String) -> PyResult<String> {
    Ok(logseq::remove_consecutive_spaces(file_contents).unwrap())
}
