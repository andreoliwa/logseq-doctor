//go:build darwin

// lqd-statusbar shows DOING task count in the macOS status bar,
// polling the Logseq graph directory every few seconds via rg.
package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/systray"
	"github.com/andreoliwa/logseq-go/content"

	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
)

const defaultDashboardPort = "8091"

const (
	tickerIntervalSecs = 2
	iconIdle           = "🛑"
	iconActive         = "🟢"
)

// countDoing counts tasks marked DOING in the graph by running rg.
// Returns 0 on any error so the icon degrades gracefully.
func countDoing(graphPath string) int {
	//nolint:gosec
	cmd := exec.CommandContext(context.Background(), "rg",
		"--no-filename", "--no-line-number", "--", `- DOING`, graphPath)

	out, err := cmd.Output()
	if err != nil {
		return 0
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))

	count := 0

	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) != "" {
			count++
		}
	}

	return count
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
	systray.SetTitle(doingIcon(countDoing(graphPath)))

	mOpen := systray.AddMenuItem(fmt.Sprintf("Open %s in Logseq", content.TaskStringDoing), "")
	mDash := systray.AddMenuItem("Open Backlog UI", "")
	mQuit := systray.AddMenuItem("Quit", "")

	ticker := time.NewTicker(tickerIntervalSecs * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				systray.SetTitle(doingIcon(countDoing(graphPath)))

			case <-mOpen.ClickedCh:
				logseqext.OpenPageInApp(graphName, content.TaskStringDoing)

			case <-mDash.ClickedCh:
				openURL(dashboardURL)

			case <-mQuit.ClickedCh:
				ticker.Stop()
				systray.Quit()

				return
			}
		}
	}()
}

func onExit() {}
