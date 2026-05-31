# Contributing

Contributions are welcome, and they are greatly appreciated! Every little bit helps, and credit will always be given.

## Bug Reports

When [reporting a bug](https://github.com/andreoliwa/logseq-doctor/issues), please include:

- Your operating system name and version
- Any details about your local setup that might be helpful in troubleshooting
- Detailed steps to reproduce the bug

## Documentation Improvements

Logseq Doctor could always use more documentation, whether as part of the official docs, in docstrings, or even on the web in blog posts, articles, and such.

## Feature Requests and Feedback

The best way to send feedback is to [file an issue](https://github.com/andreoliwa/logseq-doctor/issues).

If you are proposing a feature:

- Explain in detail how it would work
- Keep the scope as narrow as possible, to make it easier to implement
- Remember that this is a volunteer-driven project, and that code contributions are welcome :)

## Development

To set up `logseq-doctor` for local development:

1. Fork [logseq-doctor](https://github.com/andreoliwa/logseq-doctor) (look for the "Fork" button)

2. Clone your fork locally:

   ```bash
   git clone git@github.com:YOURGITHUBNAME/logseq-doctor.git
   cd logseq-doctor
   ```

3. Set up your local development environment:

   ```bash
   make setup
   ```

   This will:
   - Download Go module dependencies
   - Install development tools

4. Create a branch for local development:

   ```bash
   git checkout -b name-of-your-bugfix-or-feature
   ```

   Now you can make your changes locally.

5. When you're done making changes, run all the checks:

   ```bash
   make test
   golangci-lint run
   ```

6. Commit your changes and push your branch to GitHub:

   ```bash
   git add .
   git commit -m "Your detailed description of your changes."
   git push origin name-of-your-bugfix-or-feature
   ```

7. Submit a pull request through the GitHub website

## Pull Request Guidelines

If you need some code review or feedback while you're developing the code, just make the pull request.

For merging, you should:

1. Include passing tests (run `make test`)
2. Update documentation when there's new API, functionality, etc.
3. Add a note to the changelog about the changes

## Development Tips

### Running a Subset of Tests

```bash
go test -v -run TestMyFeature ./...
```

### Running Tests in Parallel

```bash
go test -v -parallel 4 ./...
```

### Code Style

- We use `gofmt` for code formatting
- We use `golangci-lint` for linting
- Run `golangci-lint run` to check code style

### Building Documentation Locally

To build and preview the documentation locally:

```bash
# Install MkDocs and dependencies
pip install -r docs/requirements.txt

# Serve the documentation locally
mkdocs serve
```

Then open [http://127.0.0.1:8000](http://127.0.0.1:8000) in your browser.

### Project Structure

```
logseq-doctor/
├── cmd/              # Go CLI commands
├── internal/         # Go internal packages
├── pkg/              # Go public packages
├── docs/             # Documentation (MkDocs)
└── go.mod            # Go module definition
```

## Code of Conduct

Please note that this project is released with a Contributor Code of Conduct. By participating in this project you agree to abide by its terms.

## Questions?

Feel free to [open an issue](https://github.com/andreoliwa/logseq-doctor/issues) if you have any questions about contributing!
