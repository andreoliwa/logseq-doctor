package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	logseqapi "github.com/andreoliwa/logseq-doctor/internal/api"
	"github.com/andreoliwa/logseq-doctor/internal/backlog"
	"github.com/andreoliwa/logseq-doctor/internal/logseqext"
	"github.com/andreoliwa/logseq-doctor/internal/pocketbase"
	lqdsync "github.com/andreoliwa/logseq-doctor/internal/sync"
	"github.com/andreoliwa/logseq-go"
	"github.com/spf13/cobra"
)

// SyncDependencies holds all injectable dependencies for the sync command.
// This enables unit testing without connecting to PocketBase or Logseq.
type SyncDependencies struct {
	TimeNow func() time.Time
}

// NewSyncCmd creates a new sync command with the specified dependencies.
// If deps is nil, it uses default (production) implementations.
func NewSyncCmd(deps *SyncDependencies) *cobra.Command {
	if deps == nil {
		deps = &SyncDependencies{
			TimeNow: time.Now,
		}
	}

	var initFlag bool

	cmd := &cobra.Command{ //nolint:exhaustruct
		Use:   "sync",
		Short: "Sync Logseq tasks to PocketBase",
		Long:  "Reads backlog config and tasks from Logseq, calculates ranks, and upserts to PocketBase.",
		Run: func(_ *cobra.Command, _ []string) {
			runSyncWith(deps.TimeNow, initFlag)
		},
	}

	cmd.Flags().BoolVar(&initFlag, "init", false, "Drop and recreate lqd_tasks collection before syncing")

	return cmd
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(NewSyncCmd(nil))
}

// runSyncWith is the testable core of runSync.
func runSyncWith(currentTime func() time.Time, initFlag bool) {
	path := os.Getenv("LOGSEQ_GRAPH_PATH")
	logseqAPI := logseqapi.NewLogseqAPI(path, os.Getenv("LOGSEQ_HOST_URL"), os.Getenv("LOGSEQ_API_TOKEN"))
	graph := logseqapi.OpenGraphFromPath(path)

	pbURL := os.Getenv("POCKETBASE_URL")
	if pbURL == "" {
		pbURL = "http://127.0.0.1:8090"
	}

	pbClient, err := pocketbase.NewClient(pbURL, os.Getenv("POCKETBASE_USERNAME"), os.Getenv("POCKETBASE_PASSWORD"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if initFlag {
		err = initCollection(pbClient)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		exists, existsErr := pbClient.CollectionExists("lqd_tasks")
		if existsErr != nil {
			fmt.Println(existsErr)
			os.Exit(1)
		}

		if !exists {
			fmt.Println("Collection 'lqd_tasks' not found. Run 'lqd sync --init' to create it.")
			os.Exit(1)
		}
	}

	err = runSyncPipeline(graph, logseqAPI, pbClient, currentTime)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initCollection(client *pocketbase.Client) error {
	exists, err := client.CollectionExists("lqd_tasks")
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}

	if exists {
		fmt.Println("Dropping existing lqd_tasks collection...")

		err = client.DeleteCollection("lqd_tasks")
		if err != nil {
			return fmt.Errorf("failed to delete collection: %w", err)
		}
	}

	fmt.Println("Creating lqd_tasks collection...")

	err = client.CreateCollection(pocketbase.LqdTasksSchema())
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	return nil
}

func runSyncPipeline(
	graph *logseq.Graph, logseqAPI logseqapi.LogseqAPI, pbClient *pocketbase.Client, currentTime func() time.Time,
) error {
	reader := backlog.NewPageConfigReader(graph, "backlog")

	config, err := reader.ReadConfig()
	if err != nil {
		return fmt.Errorf("failed to read backlog config: %w", err)
	}

	ranks, backlogOrder := collectBacklogRefs(graph, config)
	fmt.Printf("Calculated ranks for %d unique tasks across %d backlogs\n", len(ranks), len(backlogOrder))

	tasks, err := fetchLogseqTasks(logseqAPI)
	if err != nil {
		return err
	}

	fmt.Println("Building tag lookup table...")

	refLookup := logseqapi.BuildRefLookup(tasks)
	fmt.Printf("Resolved %d unique ref IDs\n", len(refLookup))

	fmt.Println("Enriching tasks with ancestor tags...")

	tagsByUUID := logseqapi.EnrichTasksWithAncestorTags(tasks, refLookup)
	desired := buildDesiredRecords(tasks, ranks, tagsByUUID, currentTime)

	return applyChanges(pbClient, desired)
}

// unrankedSectionTexts lists the header texts that mark the start of an unranked
// section on a backlog page. Any block ref that is a child of one of these headers
// backlogUnrankedSectionTexts returns the header texts that mark the start of an
// unranked section on a backlog page. Any block ref that is a child of one of
// these headers is assigned SectionUnranked during sync.
func backlogUnrankedSectionTexts() []string {
	return []string{
		backlog.HeaderOverdue.Label,
		backlog.HeaderNewTasks.Label,
		backlog.HeaderTriaged.Label,
		backlog.HeaderScheduled.Label,
		backlog.HeaderUnranked.Label,
	}
}

// collectBacklogRefs scans the Focus page and all configured backlog pages,
// returning a rank map (uuid → []RankInfo) and the ordered list of backlog names.
// Each ref is classified as SectionRanked or SectionUnranked based on which
// section header it lives under on the page.
func collectBacklogRefs(
	graph *logseq.Graph, config *backlog.Config,
) (map[string][]lqdsync.RankInfo, []string) {
	ranks := make(map[string][]lqdsync.RankInfo)
	backlogOrder := make([]string, 0, len(config.Backlogs)+1)

	collectFocusRefs(graph, config.FocusPage, ranks, &backlogOrder)

	for _, bc := range config.Backlogs {
		collectPageRefs(graph, bc.BacklogPage, ranks, &backlogOrder)
	}

	fmt.Println()

	return ranks, backlogOrder
}

func collectFocusRefs(
	graph *logseq.Graph, focusPagePath string,
	ranks map[string][]lqdsync.RankInfo, backlogOrder *[]string,
) {
	focusName := filepath.Base(focusPagePath)
	focusPage := logseqapi.OpenPage(graph, focusPagePath)
	focusUUIDs := logseqext.ExtractBlockRefUUIDs(focusPage)

	if len(focusUUIDs) == 0 {
		return
	}

	*backlogOrder = append(*backlogOrder, focusName)
	focusIdx := len(*backlogOrder)

	for pos, uuid := range focusUUIDs {
		ranks[uuid] = append(ranks[uuid], lqdsync.RankInfo{
			BacklogName:  focusName,
			BacklogIndex: focusIdx,
			Section:      backlog.SectionRanked,
			Rank:         pos + 1,
		})
	}

	fmt.Printf("%s=%d ", focusName, len(focusUUIDs))
}

func collectPageRefs(
	graph *logseq.Graph, backlogPagePath string,
	ranks map[string][]lqdsync.RankInfo, backlogOrder *[]string,
) {
	pageName := filepath.Base(backlogPagePath)
	page := logseqapi.OpenPage(graph, backlogPagePath)

	sectioned := logseqext.ExtractSectionedBlockRefUUIDs(page, backlogUnrankedSectionTexts())
	if len(sectioned) == 0 {
		return
	}

	*backlogOrder = append(*backlogOrder, pageName)
	backlogIdx := len(*backlogOrder)

	rankedPos, unrankedPos := 0, 0

	for _, ref := range sectioned {
		if hasRankInBacklog(ranks[ref.UUID], pageName) {
			continue
		}

		if ref.Ranked {
			rankedPos++
			ranks[ref.UUID] = append(ranks[ref.UUID], lqdsync.RankInfo{
				BacklogName:  pageName,
				BacklogIndex: backlogIdx,
				Section:      backlog.SectionRanked,
				Rank:         rankedPos,
			})
		} else {
			unrankedPos++
			ranks[ref.UUID] = append(ranks[ref.UUID], lqdsync.RankInfo{
				BacklogName:  pageName,
				BacklogIndex: backlogIdx,
				Section:      backlog.SectionUnranked,
				Rank:         unrankedPos,
			})
		}
	}

	fmt.Printf("%s=%d ", pageName, len(sectioned))
}

func fetchLogseqTasks(logseqAPI logseqapi.LogseqAPI) ([]logseqapi.TaskJSON, error) {
	query := "(and (task TODO DOING WAITING NOW LATER))"

	jsonStr, err := logseqAPI.PostQuery(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}

	tasks, err := logseqapi.ExtractTasksFromJSON(jsonStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}

	fmt.Printf("Found %d tasks from Logseq\n", len(tasks))

	return tasks, nil
}

func buildDesiredRecords(
	tasks []logseqapi.TaskJSON, ranks map[string][]lqdsync.RankInfo, tagsByUUID map[string]string,
	currentTime func() time.Time,
) []map[string]any {
	desired := make([]map[string]any, 0, len(tasks))

	for _, task := range tasks {
		if task.UUID == "" {
			continue
		}

		enrichedTags := tagsByUUID[task.UUID]
		rankInfos := ranks[task.UUID]

		if len(rankInfos) == 0 {
			// Task not in any backlog: one record with no backlog metadata.
			desired = append(desired, lqdsync.TaskToRecord(task, nil, enrichedTags, currentTime))
		} else {
			// One PocketBase record per (task, backlog) pair.
			for i := range rankInfos {
				desired = append(desired, lqdsync.TaskToRecord(task, &rankInfos[i], enrichedTags, currentTime))
			}
		}
	}

	return desired
}

// hasRankInBacklog reports whether the given backlog name already has an entry in the rank slice.
func hasRankInBacklog(rankInfos []lqdsync.RankInfo, backlogName string) bool {
	for _, r := range rankInfos {
		if r.BacklogName == backlogName {
			return true
		}
	}

	return false
}

func applyChanges(pbClient *pocketbase.Client, desired []map[string]any) error {
	existing, err := pbClient.FetchRecords("lqd_tasks", "", "")
	if err != nil {
		return fmt.Errorf("failed to fetch existing records: %w", err)
	}

	fmt.Printf("Found %d existing records in PocketBase\n", len(existing))

	toCreate, toUpdate, toDelete := lqdsync.DiffRecords(existing, desired)

	for _, record := range toCreate {
		createErr := pbClient.CreateRecord("lqd_tasks", record)
		if createErr != nil {
			fmt.Printf("Warning: failed to create %s: %v\n", record["id"], createErr)
		}
	}

	for _, record := range toUpdate {
		id, _ := record["id"].(string)

		updateErr := pbClient.UpdateRecord("lqd_tasks", id, record)
		if updateErr != nil {
			fmt.Printf("Warning: failed to update %s: %v\n", id, updateErr)
		}
	}

	for _, id := range toDelete {
		deleteErr := pbClient.DeleteRecord("lqd_tasks", id)
		if deleteErr != nil {
			fmt.Printf("Warning: failed to delete %s: %v\n", id, deleteErr)
		}
	}

	fmt.Printf("\nSync complete! Created=%d Updated=%d Deleted=%d\n",
		len(toCreate), len(toUpdate), len(toDelete))

	return nil
}
