//! Handle [Logseq](https://logseq.com/) Markdown files in Rust

use chrono::{Local, NaiveDate};
use regex::Regex;
use std::fs;
use std::fs::{File, OpenOptions};
use std::io::Write;
use std::path::{Path, PathBuf};

/// Remove consecutive spaces on lines that begin with a dash, keeping leading spaces
///
/// # Arguments
///
/// * `file_contents`: Contents of a file as a string
///
/// returns: Result<String, Error>
///
/// # Examples
///
/// ```
/// use logseq::remove_consecutive_spaces;
/// assert_eq!(remove_consecutive_spaces("    abc   123     def  ".to_string()).unwrap(), "    abc   123     def  ".to_string());
/// assert_eq!(remove_consecutive_spaces("\n  - abc  123\n    - def   4  5 ".to_string()).unwrap(), "\n  - abc 123\n    - def 4 5 ".to_string());
/// assert_eq!(remove_consecutive_spaces(
///     "   -This   is   a  test\n   Another  test\n-  Dash  line  here".to_string()).unwrap(),
///     "   -This is a test\n   Another  test\n- Dash line here".to_string());
/// assert_eq!(remove_consecutive_spaces(
///     "    -   This   is   a  test\n   Another  test\n-  Dash  line  here   with   extra  spaces".to_string()).unwrap(),
///     "    - This is a test\n   Another  test\n- Dash line here with extra spaces".to_string());
///
/// let ends_with_linebreak = "- Root\n  - Child\n";
/// assert_eq!(remove_consecutive_spaces(ends_with_linebreak.to_string()).unwrap(), ends_with_linebreak);
/// ```
pub fn remove_consecutive_spaces(file_contents: String) -> anyhow::Result<String> {
    let space_re = Regex::new(r" {2,}").unwrap();
    let ends_with_linebreak = file_contents.ends_with('\n');

    let result = file_contents
        .lines()
        .map(|line| {
            if line.trim_start().starts_with('-') {
                // Replace multiple spaces with a single space, except for leading spaces
                let first_non_space = line.find('-').unwrap_or(0);
                let (leading_spaces, rest) = line.split_at(first_non_space);
                format!("{}{}", leading_spaces, space_re.replace_all(rest, " "))
            } else {
                // Leave line unchanged
                line.to_string()
            }
        })
        .collect::<Vec<_>>()
        .join("\n");

    // Append a line break if the original string ended with one
    let final_result = if ends_with_linebreak {
        format!("{}\n", result)
    } else {
        result
    };

    Ok(final_result)
}

/// Remove unnecessary brackets from tags
///
/// # Arguments
///
/// * `file_contents`: Contents of a file as a string
///
/// returns: Result<String, Error>
///
/// # Examples
///
/// ```
/// use logseq::remove_unnecessary_brackets_from_tags;
/// assert_eq!(remove_unnecessary_brackets_from_tags(&"#[[tag]]".to_string()).unwrap(), "#tag".to_string());
/// assert_eq!(remove_unnecessary_brackets_from_tags(&"#[[tag with spaces]]".to_string()).unwrap(), "#[[tag with spaces]]".to_string());
/// assert_eq!(remove_unnecessary_brackets_from_tags(&"text before #[[some-tag]] then after".to_string()).unwrap(), "text before #some-tag then after".to_string());
/// ```
pub fn remove_unnecessary_brackets_from_tags(file_contents: &str) -> anyhow::Result<String> {
    let tag_re = Regex::new(r"#\[\[([^ ]*?)\]\]").unwrap();
    let result = tag_re.replace_all(file_contents, "#$1");
    Ok(result.to_string())
}

/// Subdirectory for Logseq pages
pub const SUBDIR_PAGES: &str = "pages";
/// Subdirectory for Logseq journals
pub const SUBDIR_JOURNALS: &str = "journals";

/// Represents a Logseq graph
/// Placeholder for future functionality (API client, global and graph configuration, etc.)
pub struct Logseq {
    /// The path to the Logseq graph
    pub graph_path: PathBuf,
}

impl Logseq {
    /// Constructs a new Logseq graph for the given path
    pub fn new(graph_path: &Path) -> Self {
        Logseq {
            graph_path: graph_path.to_path_buf(),
        }
    }
    /// Constructs a new Logseq graph from any file in the graph dir, be it a page or a journal
    pub fn guess_graph_from_path<P: AsRef<Path>>(path: P) -> Self {
        let mut graph_path = PathBuf::new();
        for component in path.as_ref().components() {
            if component.as_os_str() == SUBDIR_PAGES || component.as_os_str() == SUBDIR_JOURNALS {
                break;
            }
            graph_path.push(component);
        }

        if graph_path.as_os_str().is_empty() {
            panic!(
                "The file {} doesn't belong to a Logseq graph.",
                path.as_ref().display()
            );
        }

        Logseq { graph_path }
    }
}

/// A Logseq page
pub struct Page {
    /// The full path to the page
    pub path: PathBuf,
}

impl Page {
    /// Constructs a new Page for the given path
    pub fn new(path: &Path) -> Self {
        Page {
            path: path.to_path_buf(),
        }
    }

    /// Tidy up the page by calling several clean-up functions.
    /// E.g. removing consecutive spaces, unnecessary brackets from tags, etc.
    /// Returns true if the file was modified, false otherwise.
    pub fn tidy_up(&self) -> anyhow::Result<bool> {
        let path = self.path.clone();

        let original_contents = fs::read_to_string(&path)?;
        let no_brackets = remove_unnecessary_brackets_from_tags(&original_contents)?;
        if no_brackets == original_contents {
            return Ok(false);
        }

        fs::write(&path, no_brackets.as_bytes())?;
        Ok(true)
    }
}

/// A Logseq journal file
pub struct Journal {
    /// The path to the Logseq graph
    // TODO: move this to Logseq struct
    pub graph_path: PathBuf,
    /// The date of the journal
    pub date: NaiveDate,
}

impl Journal {
    /// Constructs a new Journal for the given date, or uses the current date if None is provided.
    pub fn new(graph_path: PathBuf, date: Option<NaiveDate>) -> Self {
        let final_date = date.unwrap_or_else(|| Local::now().date_naive());
        Journal {
            graph_path,
            date: final_date,
        }
    }

    /// Returns the full path to the journal file
    pub fn as_path(&self) -> PathBuf {
        let journal_file_name = format!("journals/{}.md", self.date.format("%Y_%m_%d"));
        self.graph_path.join(journal_file_name)
    }

    fn _append_or_prepend(&self, markdown: String, append: bool) -> anyhow::Result<()> {
        let prepend: bool = !append;
        let path = self.as_path();
        eprint!("Journal {}: ", path.to_string_lossy());

        // if no markdown content, print an error and return
        if markdown.is_empty() {
            eprintln!("no content provided");
            return Ok(());
        }

        let empty: bool;
        let content;
        if let Ok(valid_content) = fs::read_to_string(&path) {
            content = valid_content.clone();
            let trimmed_content = valid_content
                .trim_end()
                .trim_start_matches('-')
                .trim_start();
            empty = trimmed_content.is_empty();
            if empty {
                eprintln!("truncated file");
            }
        } else {
            empty = true;
            content = String::new();
            eprintln!("new file");
        }

        let mut file: File;
        if empty || prepend {
            // We need to overwrite the file because Logseq adds an empty bullet (dash) to empty pages
            file = OpenOptions::new()
                .write(true)
                .create(true)
                .truncate(true)
                .open(&path)?;
        } else {
            file = OpenOptions::new().append(true).open(&path)?;
            eprintln!("appending");

            println!(); // Output all content to stdout
            file.write_all(b"\n")?;
        }

        print!("{}", markdown);
        file.write_all(markdown.as_bytes())?;
        if prepend && !empty {
            file.write_all(b"\n")?;
            file.write_all(content.as_bytes())?;
        }
        file.flush()?;
        Ok(())
    }

    /// Appends the given Markdown content to the journal file at the end of the file
    pub fn append(&self, markdown: String) -> anyhow::Result<()> {
        self._append_or_prepend(markdown, true)
    }

    /// Prepends the given Markdown content to the journal file at the beginning of the file
    pub fn prepend(&self, markdown: String) -> anyhow::Result<()> {
        self._append_or_prepend(markdown, false)
    }
}
