package internal

import "github.com/fatih/color"

// PageColor is a color function for page names.
var PageColor = color.New(color.FgHiWhite).SprintfFunc() //nolint:gochecknoglobals
