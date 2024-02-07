//! # Handle [Logseq](https://logseq.com/) Markdown files

use chrono::{Local, NaiveDate};
use regex::Regex;
use std::fs;
use std::fs::{File, OpenOptions};
use std::io::Write;
use std::path::PathBuf;

/// Remove consecutive spaces on lines that begin with a dash, keeping leading spaces
///
/// # Arguments
///
/// * `file_contents`: Contents of a file as a string
///
/// returns: Result<String, ()>
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

/// A Logseq journal file
pub struct Journal {
    graph: PathBuf,
    date: NaiveDate,
}

impl Journal {
    /// Constructs a new Journal for the given date, or uses the current date if None is provided.
    pub fn new(graph: PathBuf, date: Option<NaiveDate>) -> Self {
        let final_date = date.unwrap_or_else(|| Local::now().date_naive());
        Journal {
            graph,
            date: final_date,
        }
    }

    /// Returns the full path to the journal file
    pub fn as_path(&self) -> PathBuf {
        let journal_file_name = format!("journals/{}.md", self.date.format("%Y_%m_%d"));
        self.graph.join(journal_file_name)
    }

    /// Appends the given Markdown content to the journal file
    pub fn append(&self, markdown: String) -> anyhow::Result<()> {
        let path = self.as_path();
        eprint!("Journal {}: ", path.to_string_lossy());

        // if no markdown content, print an error and return
        if markdown.is_empty() {
            eprintln!("no content provided");
            return Ok(());
        }

        let empty: bool;
        if let Ok(content) = fs::read_to_string(&path) {
            let trimmed_content = content.trim_end().trim_start_matches('-').trim_start();
            empty = trimmed_content.is_empty();
        } else {
            empty = true;
        }

        let mut file: File;
        if empty {
            // We need to overwrite the file because Logseq adds an empty bullet (dash) to empty pages
            file = OpenOptions::new()
                .write(true)
                .create(true)
                .truncate(true)
                .open(&path)?;
            eprintln!("new/recreated file");
        } else {
            file = OpenOptions::new().append(true).open(&path)?;
            eprintln!("appending");

            println!(); // Output all content to stdout
            file.write_all(b"\n")?;
        }

        print!("{}", markdown);
        file.write_all(markdown.as_bytes())?;
        file.flush()?;
        Ok(())
    }
}
