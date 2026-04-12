package pocketbase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	pbPollInterval = 200 * time.Millisecond
	pbPollTimeout  = 500 * time.Millisecond
)

// ErrWaitTimeout is returned when WaitForReady times out.
var ErrWaitTimeout = errors.New("timed out waiting for service to be ready")

// ErrPocketBaseNotFound is returned when the pocketbase binary cannot be located.
var ErrPocketBaseNotFound = errors.New("pocketbase executable not found in $PATH or ~/.local/bin")

// StartPocketBase starts pocketbase serve as a managed subprocess from workDir.
// Stdout and stderr are inherited so PocketBase logs appear in the same terminal.
// The caller is responsible for calling cmd.Process.Kill() on shutdown.
func StartPocketBase(workDir string) (*exec.Cmd, error) {
	pbPath, err := findPocketBase()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, pbPath, "serve", "--dir="+workDir+"/pb_data")
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("start pocketbase: %w", err)
	}

	return cmd, nil
}

// findPocketBase resolves the pocketbase executable path.
// It checks $PATH first, then falls back to ~/.local/bin/pocketbase.
func findPocketBase() (string, error) {
	path, lookErr := exec.LookPath("pocketbase")
	if lookErr == nil {
		return path, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find pocketbase: get home dir: %w", err)
	}

	localBin := homeDir + "/.local/bin/pocketbase"

	_, statErr := os.Stat(localBin)
	if statErr == nil {
		return localBin, nil
	}

	return "", ErrPocketBaseNotFound
}

// WaitForReady polls healthURL until it returns 200 OK or timeout elapses.
func WaitForReady(healthURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{ //nolint:exhaustruct // only Timeout needed for short-lived polling
		Timeout: pbPollTimeout,
	}

	for time.Now().Before(deadline) {
		resp, err := client.Get(healthURL) //nolint:noctx // short-lived polling, context not needed
		if err == nil {
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		time.Sleep(pbPollInterval)
	}

	return fmt.Errorf("%w: %s", ErrWaitTimeout, healthURL)
}
