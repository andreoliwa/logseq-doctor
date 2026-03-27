// Package logseqext provides generic extensions to the logseq-go library.
// Functions here are candidates for upstreaming into logseq-go once that
// library is restructured.
//
// This package must remain self-contained: it may only import logseq-go,
// the Go standard library, and golang.org/x packages. It must never import
// logseq-doctor internal packages (doing so would also create a cyclic import).
package logseqext
