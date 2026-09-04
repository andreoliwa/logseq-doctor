//go:build darwin

// lqd-statusbar shows DOING task count in the macOS status bar,
// polling the Logseq graph directory every few seconds.
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/systray"
	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
)

const defaultDashboardPort = "8091"

const (
	tickerIntervalSecs = 2
	iconIdle           = "🛑"
	iconActive         = "🟢"
)

// countDoing syncs the DOING page and returns the number of tasks marked DOING.
// Returns 0 on any error so the icon degrades gracefully.
func countDoing(graphPath string) int {
	files, count, err := doingFiles(graphPath)
	if err != nil {
		return 0
	}

	graph, err := logseq.Open(context.Background(), graphPath)
	if err != nil {
		return 0
	}
	defer graph.Close()

	_, syncErr := logseqext.SyncDoingTasksInFiles(graph, files)
	if syncErr != nil {
		return 0
	}

	return count
}

func doingFiles(graphPath string) ([]string, int, error) {
	//nolint:gosec // graph path is configured by the user
	command := exec.CommandContext(context.Background(), "rg", "--files-with-matches", "--glob", "*.md", "--",
		"- DOING", graphPath)

	output, err := command.Output()
	if err != nil {
		return nil, 0, fmt.Errorf("find DOING task files: %w", err)
	}

	files := make([]string, 0)
	for file := range strings.FieldsSeq(string(output)) {
		files = append(files, file)
	}

	//nolint:gosec // graph path is configured by the user
	command = exec.CommandContext(context.Background(), "rg", "--no-filename", "--no-line-number", "--",
		"- DOING", graphPath)

	output, err = command.Output()
	if err != nil {
		return nil, 0, fmt.Errorf("count DOING tasks: %w", err)
	}

	return files, bytes.Count(output, []byte("\n")), nil
}

// doingIcon maps a DOING task count to a Unicode status-bar label.
func doingIcon(count int) string {
	if count <= 0 {
		return iconIdle
	}

	return fmt.Sprintf("%s %d", iconActive, count)
}

func main() {
	graphPath := os.Getenv("LOGSEQ_GRAPH_PATH")
	if graphPath == "" {
		fmt.Fprintln(os.Stderr, "LOGSEQ_GRAPH_PATH is not set")
		os.Exit(1)
	}

	graphName := filepath.Base(graphPath)

	port := os.Getenv("LQD_SERVE_PORT")
	if port == "" {
		port = defaultDashboardPort
	}

	dashboardURL := "http://localhost:" + port

	systray.Run(func() { onReady(graphPath, graphName, dashboardURL) }, onExit)
}

func openURL(url string) {
	err := exec.CommandContext(context.Background(), "open", url).Start() //nolint:gosec
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open %s: %v\n", url, err)
	}
}

func onReady(graphPath, graphName, dashboardURL string) {
	systray.SetTitle(iconIdle)

	mOpen := systray.AddMenuItem(fmt.Sprintf("Open %s in Logseq", content.TaskStringDoing), "")
	mDash := systray.AddMenuItem("Open Backlog UI", "")
	mQuit := systray.AddMenuItem("Quit", "")

	counts := make(chan int)
	quit := make(chan struct{})

	go pollDoingCount(graphPath, counts, quit)

	go func() {
		for {
			select {
			case count := <-counts:
				systray.SetTitle(doingIcon(count))

			case <-mOpen.ClickedCh:
				logseqext.OpenPageInApp(graphName, content.TaskStringDoing)

			case <-mDash.ClickedCh:
				openURL(dashboardURL)

			case <-mQuit.ClickedCh:
				close(quit)
				systray.Quit()

				return
			}
		}
	}()
}

func pollDoingCount(graphPath string, counts chan<- int, quit <-chan struct{}) {
	ticker := time.NewTicker(tickerIntervalSecs * time.Second)
	defer ticker.Stop()

	for {
		count := countDoing(graphPath)

		select {
		case counts <- count:
		case <-quit:
			return
		}

		select {
		case <-ticker.C:
		case <-quit:
			return
		}
	}
}

func onExit() {}
