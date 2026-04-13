package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/backlog"
	"github.com/andreoliwa/logseq-doctor/internal/dashboard"
	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/andreoliwa/logseq-doctor/internal/pocketbase"
	"github.com/andreoliwa/logseq-doctor/internal/serve"
)

const (
	defaultServePort     = 8091
	defaultPocketBaseURL = "http://127.0.0.1:8090"
	pbReadyTimeout       = 10 * time.Second
	shutdownTimeout      = 5 * time.Second
	readHeaderTimeout    = 5 * time.Second
)

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(dashboardCmd)
	dashboardCmd.Flags().IntP("port", "p", defaultServePort, "HTTP server port (also LQD_SERVE_PORT env var)")
}

// DashboardAliases returns the Cobra aliases for the dashboard command.
// Exported so tests can verify the alias without accessing the unexported global.
func DashboardAliases() []string { return []string{"dash"} }

var dashboardCmd = &cobra.Command{ //nolint:exhaustruct,gochecknoglobals
	Use:     "dashboard",
	Aliases: DashboardAliases(),
	Short:   "Start PocketBase and the backlog web UI",
	Long: `Starts PocketBase as a managed subprocess, waits for it to be ready,
then serves the backlog dashboard at http://localhost:8091 (configurable).

Environment variables:
  POCKETBASE_URL       PocketBase URL (default http://127.0.0.1:8090)
  POCKETBASE_USERNAME  PocketBase admin email
  POCKETBASE_PASSWORD  PocketBase admin password
  LOGSEQ_GRAPH_PATH    Path to Logseq graph (required for write-back)
  LQD_SERVE_PORT       HTTP server port (default 8091)`,
	RunE: runDashboard,
}

func runDashboard(cmd *cobra.Command, _ []string) error {
	port := ResolvePort(cmd)
	pbURL := ResolveEnvWithDefault("POCKETBASE_URL", defaultPocketBaseURL)
	pbUser := os.Getenv("POCKETBASE_USERNAME")
	pbPass := os.Getenv("POCKETBASE_PASSWORD")
	graphPath := os.Getenv("LOGSEQ_GRAPH_PATH")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	pbCmd, err := pocketbase.StartPocketBase(homeDir)
	if err != nil {
		return fmt.Errorf("start pocketbase: %w", err)
	}

	defer func() {
		if pbCmd.Process != nil {
			_ = pbCmd.Process.Kill()
		}
	}()

	healthURL := pbURL + "/api/health"

	err = pocketbase.WaitForReady(healthURL, pbReadyTimeout)
	if err != nil {
		return fmt.Errorf("pocketbase not ready: %w", err)
	}

	fmt.Fprintf(os.Stderr, "PocketBase ready at %s\n", pbURL)

	token, err := authenticate(pbURL, pbUser, pbPass)
	if err != nil {
		return err
	}

	mux := BuildHTTPMux(pbURL, token, graphPath)
	uiURL := fmt.Sprintf("http://localhost:%d", port)

	fmt.Fprintf(os.Stderr, "Backlog UI ready at %s\n", uiURL)

	return startHTTPServer(cmd.Context(), port, mux)
}

// ResolvePort returns the effective port: flag > env var > default.
func ResolvePort(cmd *cobra.Command) int {
	port, _ := cmd.Flags().GetInt("port")

	if envPort := os.Getenv("LQD_SERVE_PORT"); envPort != "" {
		p, convErr := strconv.Atoi(envPort)
		if convErr == nil {
			port = p
		}
	}

	return port
}

// ResolveEnvWithDefault returns the env var value or fallback if unset.
func ResolveEnvWithDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return fallback
}

// authenticate obtains a PocketBase token. Returns "" and nil when credentials are absent.
func authenticate(pbURL, pbUser, pbPass string) (string, error) {
	if pbUser == "" || pbPass == "" {
		return "", nil
	}

	pb, err := pocketbase.NewClient(pbURL, pbUser, pbPass)
	if err != nil {
		return "", fmt.Errorf("pocketbase auth: %w", err)
	}

	return pb.Token(), nil
}

// BuildHTTPMux creates the HTTP mux with all routes registered.
func BuildHTTPMux(pbURL, token, graphPath string) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write(backlogHTML)
	})

	mux.Handle("GET /api/", serve.NewProxy(pbURL, token))
	mux.Handle("POST /api/", serve.NewProxy(pbURL, token))
	mux.Handle("PATCH /api/", serve.NewProxy(pbURL, token))
	mux.Handle("DELETE /api/", serve.NewProxy(pbURL, token))

	mux.HandleFunc("GET /backlog.css", func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = writer.Write(backlogCSS)
	})

	mux.HandleFunc("GET /internal/config", func(writer http.ResponseWriter, _ *http.Request) {
		handleConfig(writer, graphPath) //nolint:contextcheck // logseq-go graph API has no context support
	})

	mux.HandleFunc("POST /internal/move-to-unranked", func(writer http.ResponseWriter, req *http.Request) {
		//nolint:contextcheck // logseq-go graph API has no context support
		handleMoveToUnranked(writer, req, graphPath, pbURL, token)
	})

	return mux
}

// handleConfig returns UI configuration derived from the server environment.
func handleConfig(writer http.ResponseWriter, graphPath string) {
	graphName := ""
	if graphPath != "" {
		graphName = filepath.Base(graphPath)
	}

	// Build a map from short backlog name → full page title (e.g. "self" → "backlog/self").
	backlogPages := map[string]string{}

	if graphPath != "" {
		graph := logseqapi.OpenGraphFromPath(graphPath)
		reader := backlog.NewPageConfigReader(graph, "backlog")

		cfg, readErr := reader.ReadConfig()
		if readErr == nil {
			// Focus is a special page stored in cfg.FocusPage, not in cfg.Backlogs.
			// Add it explicitly so the JS can build a correct deep link for it.
			if cfg.FocusPage != "" {
				shortName := filepath.Base(cfg.FocusPage)
				backlogPages[shortName] = cfg.FocusPage
			}

			for _, backlogCfg := range cfg.Backlogs {
				// Key is the last path component (the short name stored in PocketBase).
				shortName := filepath.Base(backlogCfg.BacklogPage)
				backlogPages[shortName] = backlogCfg.BacklogPage
			}
		}
	}

	// Read the journal page-title format from the graph's logseq/config.edn.
	// This is the JS-style date format string (e.g. "EEEE, dd.MM.yyyy") that
	// Logseq uses to name journal pages. We pass it to the JS so it can build
	// correct deep links to journal pages without hardcoding the format.
	journalTitleFormat := logseqext.ReadJournalTitleFormat(graphPath)

	type configResponse struct {
		GraphName          string            `json:"graphName"`
		BacklogPages       map[string]string `json:"backlogPages"`
		JournalTitleFormat string            `json:"journalTitleFormat"`
	}

	payload, err := json.Marshal(configResponse{
		GraphName:          graphName,
		BacklogPages:       backlogPages,
		JournalTitleFormat: journalTitleFormat,
	})
	if err != nil {
		http.Error(writer, "marshal config: "+err.Error(), http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, _ = writer.Write(payload)
}

// resolveBacklogPage maps a short backlog name (e.g. "self") to its full page title
// (e.g. "Backlogs/self") by reading the backlog config page from the graph.
// Falls back to the short name if the config cannot be read or the name is not found.
func resolveBacklogPage(graphPath, shortName string) string {
	graph := logseqapi.OpenGraphFromPath(graphPath)
	reader := backlog.NewPageConfigReader(graph, "backlog")

	cfg, err := reader.ReadConfig()
	if err != nil {
		return shortName
	}

	if full := cfg.FindBacklogPageTitle(shortName); full != "" {
		return full
	}

	return shortName
}

// startHTTPServer runs the server until a signal arrives or a listen error occurs.
func startHTTPServer(ctx context.Context, port int, mux http.Handler) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{ //nolint:exhaustruct
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	errCh := make(chan error, 1)

	go func() {
		listenErr := srv.ListenAndServe()
		if listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			errCh <- listenErr
		}
	}()

	select {
	case listenErr := <-errCh:
		return listenErr
	case <-ctx.Done():
		log.Println("shutting down...")

		return gracefulShutdown(srv) //nolint:contextcheck // parent ctx cancelled; need fresh ctx for shutdown
	}
}

// gracefulShutdown shuts the server down with a fresh timeout context.
// The parent context is already cancelled at this point, so a new one is needed.
func gracefulShutdown(srv *http.Server) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	return nil
}

// handleMoveToUnranked handles POST /internal/move-to-unranked.
func handleMoveToUnranked(writer http.ResponseWriter, req *http.Request, graphPath, pbURL, token string) {
	var body struct {
		BacklogPage string   `json:"backlogPage"`
		UUIDs       []string `json:"uuids"`
	}

	rawBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(writer, "read body: "+err.Error(), http.StatusBadRequest)

		return
	}

	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		http.Error(writer, "decode body: "+err.Error(), http.StatusBadRequest)

		return
	}

	if graphPath == "" {
		http.Error(writer, "LOGSEQ_GRAPH_PATH not set", http.StatusInternalServerError)

		return
	}

	pageTitle := resolveBacklogPage(graphPath, body.BacklogPage)

	// body.UUIDs contains composite PocketBase record IDs (uuid_backlogname).
	// MoveToUnranked needs bare block UUIDs to match ((uuid)) refs in the .md file.
	suffix := "_" + strings.ToLower(body.BacklogPage)
	bareUUIDs := make([]string, 0, len(body.UUIDs))

	for _, id := range body.UUIDs {
		bareUUIDs = append(bareUUIDs, strings.TrimSuffix(id, suffix))
	}

	err = dashboard.MoveToUnranked(graphPath, pageTitle, bareUUIDs)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	// Update section in PocketBase so these tasks immediately appear as unranked
	// in the dashboard without requiring a full lqd sync.
	// body.UUIDs are already composite record IDs (uuid_backlogname).
	pb := pocketbase.NewClientWithToken(pbURL, token)

	for _, recordID := range body.UUIDs {
		_ = pb.UpdateRecord("lqd_tasks", recordID, map[string]any{"section": backlog.SectionUnranked})
	}

	writer.WriteHeader(http.StatusNoContent)
}
