package internal

import "errors"

var ErrFailedOpenGraph = errors.New("failed to open graph")
var ErrMissingConfig = errors.New("LOGSEQ_API_TOKEN and LOGSEQ_HOST_URL must be set")
var ErrInvalidResponseStatus = errors.New("invalid response status")
