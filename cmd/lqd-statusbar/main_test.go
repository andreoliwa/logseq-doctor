//go:build darwin

package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollDoingCountPublishesInitialCount(t *testing.T) {
	counts := make(chan int)
	quit := make(chan struct{})

	go pollDoingCount("/missing-graph", counts, quit)

	t.Cleanup(func() { close(quit) })

	select {
	case count := <-counts:
		assert.Zero(t, count)
	case <-time.After(time.Second):
		require.Fail(t, "poller did not publish its initial count")
	}
}
