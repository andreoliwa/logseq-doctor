use assert_fs::prelude::*;
use assert_fs::TempDir;
use chrono::NaiveDate;
use logseq::{Journal, Page};
use std::fs;
use std::path::{Path, PathBuf};

fn fixture(name: &str) -> String {
    let path = PathBuf::from("tests/fixtures").join(name);
    // Logseq doesn't save a line break at the end of the file, so we need to trim it off to simulate the same behavior
    fs::read_to_string(path).unwrap().trim_end().to_string()
}

fn compare_to_fixture(fixture_relative_path: &str, markdown_path: &Path) -> anyhow::Result<()> {
    let expected = fixture(fixture_relative_path);
    let actual = fs::read_to_string(markdown_path).unwrap();
    assert_eq!(actual, expected);
    Ok(())
}

struct FakeJournal {
    // We need to keep the TempDir around to prevent it from being cleaned up before the Journal is done with it
    _temp_path: TempDir,
    journal: Journal,
    content_to_add: String,
}

impl FakeJournal {
    fn new(existing_fixture_path: &str) -> Self {
        let temp = TempDir::new().unwrap();
        let journal = Journal::new(
            temp.path().to_path_buf(),
            Some(
                NaiveDate::from_ymd_opt(1913, 12, 23)
                    .ok_or_else(|| anyhow::anyhow!("Invalid date"))
                    .unwrap(),
            ),
        );
        fs::create_dir(journal.as_path().parent().unwrap()).unwrap();
        let content_to_add = fixture("journal/add-this.md");

        fs::write(&journal.as_path(), fixture(existing_fixture_path)).unwrap();

        FakeJournal {
            _temp_path: temp,
            journal,
            content_to_add,
        }
    }

    fn assert_journal(&self, expected_fixture_path: &str) -> anyhow::Result<()> {
        compare_to_fixture(expected_fixture_path, self.journal.as_path().as_path())
    }
}

#[test]
fn test_journal_has_content_append() -> anyhow::Result<()> {
    let fake = FakeJournal::new("journal/has-content/existing.md");
    fake.journal.append(fake.content_to_add.clone())?;
    fake.assert_journal("journal/has-content/expected-append.md")?;
    Ok(())
}

#[test]
fn test_journal_has_content_prepend() -> anyhow::Result<()> {
    let fake = FakeJournal::new("journal/has-content/existing.md");
    fake.journal.prepend(fake.content_to_add.clone())?;
    fake.assert_journal("journal/has-content/expected-prepend.md")?;
    Ok(())
}

#[test]
fn test_journal_empty_append() -> anyhow::Result<()> {
    let fake = FakeJournal::new("journal/empty/existing.md");
    fake.journal.append(fake.content_to_add.clone())?;
    fake.assert_journal("journal/add-this.md")?;
    Ok(())
}

#[test]
fn test_journal_empty_prepend() -> anyhow::Result<()> {
    let fake = FakeJournal::new("journal/empty/existing.md");
    fake.journal.prepend(fake.content_to_add.clone())?;
    fake.assert_journal("journal/add-this.md")?;
    Ok(())
}

struct FakePage {
    // We need to keep the TempDir around to prevent it from being cleaned up before the struct is done with it
    _temp_path: TempDir,
    page: Page,
}

impl FakePage {
    fn new(existing_fixture_path: &str) -> Self {
        let temp = TempDir::new().unwrap();
        let child = temp.child("page.md");
        let page = Page::new(child.path());
        fs::write(child.path(), fixture(existing_fixture_path)).unwrap();

        FakePage {
            _temp_path: temp,
            page,
        }
    }

    fn assert_page(&self, expected_fixture_path: &str) -> anyhow::Result<()> {
        compare_to_fixture(expected_fixture_path, &self.page.path)
    }
}

#[test]
fn test_page_with_unneeded_brackets() -> anyhow::Result<()> {
    let fake = FakePage::new("page/unneeded-brackets-before.md");
    fake.page.tidy_up()?;
    fake.assert_page("page/unneeded-brackets-after.md")?;
    Ok(())
}
