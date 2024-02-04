use regex::Regex;

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
pub fn remove_consecutive_spaces(file_contents: String) -> Result<String, ()> {
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
