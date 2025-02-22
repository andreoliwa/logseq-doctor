package cmd

import "errors"

var ErrFailedOpenGraph = errors.New("failed to open graph")
var ErrMissingConfig = errors.New("LOGSEQ_API_TOKEN and LOGSEQ_HOST_URL must be set")
var ErrPageNotFound = errors.New("page not found")
var ErrQueryLogseqAPI = errors.New("failed to query Logseq API")
