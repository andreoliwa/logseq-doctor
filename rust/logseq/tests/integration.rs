use assert_fs::TempDir;
use chrono::NaiveDate;
use logseq::Journal;
use std::fs;
use std::path::PathBuf;

fn fixture(name: &str) -> String {
    let path = PathBuf::from("tests/fixtures/journal").join(name);
    // Logseq doesn't save a line break at the end of the file, so we need to trim it off to simulate the same behavior
    fs::read_to_string(path).unwrap().trim_end().to_string()
}

struct FakeJournal {
    // We need to keep the TempDir around to prevent it from being cleaned up before the Journal is done with it
    _graph: TempDir,
    journal: Journal,
    content_to_add: String,
}

impl FakeJournal {
    fn new(existing_fixture_path: &str) -> Self {
        let graph = TempDir::new().unwrap();
        let journal = Journal::new(
            graph.path().to_path_buf(),
            Some(
                NaiveDate::from_ymd_opt(1913, 12, 23)
                    .ok_or_else(|| anyhow::anyhow!("Invalid date"))
                    .unwrap(),
            ),
        );
        fs::create_dir_all(journal.as_path().parent().unwrap()).unwrap();
        let content_to_add = fixture("add-this.md");

        fs::write(&journal.as_path(), fixture(existing_fixture_path)).unwrap();

        FakeJournal {
            _graph: graph,
            journal,
            content_to_add,
        }
    }

    fn assert_journal_content(&self, expected_fixture_path: &str) -> anyhow::Result<()> {
        let expected = fixture(expected_fixture_path);
        let actual = fs::read_to_string(self.journal.as_path()).unwrap();
        assert_eq!(actual, expected);
        Ok(())
    }
}

#[test]
fn test_journal_has_content_append() -> anyhow::Result<()> {
    let fake = FakeJournal::new("has-content/existing.md");
    fake.journal.append(fake.content_to_add.clone())?;
    fake.assert_journal_content("has-content/expected-append.md")?;
    Ok(())
}

#[test]
fn test_journal_has_content_prepend() -> anyhow::Result<()> {
    let fake = FakeJournal::new("has-content/existing.md");
    fake.journal.prepend(fake.content_to_add.clone())?;
    fake.assert_journal_content("has-content/expected-prepend.md")?;
    Ok(())
}

#[test]
fn test_journal_empty_append() -> anyhow::Result<()> {
    let fake = FakeJournal::new("empty/existing.md");
    fake.journal.append(fake.content_to_add.clone())?;
    fake.assert_journal_content("add-this.md")?;
    Ok(())
}

#[test]
fn test_journal_empty_prepend() -> anyhow::Result<()> {
    let fake = FakeJournal::new("empty/existing.md");
    fake.journal.prepend(fake.content_to_add.clone())?;
    fake.assert_journal_content("add-this.md")?;
    Ok(())
}
