package cmd

// web.go lives in package cmd so that the //go:embed directives can reference
// paths relative to this file's directory (cmd/web/). The Go embed spec
// requires the embedded path to be a descendant of the directory containing
// the source file, so moving the assets elsewhere would require moving this
// file (and its package) too. If cmd/web/ grows (multiple pages, JS bundles),
// consider extracting a dedicated internal/webui package that exports the
// embedded []byte variables and owns the assets alongside them.

import _ "embed"

//go:embed web/backlog.html
var backlogHTML []byte

//go:embed web/backlog.css
var backlogCSS []byte
