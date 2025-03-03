package internal

import (
	"encoding/json"
	"fmt"
	"github.com/andreoliwa/lsd/pkg/utils"
	"time"
)

type TaskJSON struct {
	UUID      string   `json:"uuid"`
	Marker    string   `json:"marker"`
	Content   string   `json:"content"`
	Page      pageJSON `json:"page"`
	Deadline  int      `json:"deadline"`
	Scheduled int      `json:"scheduled"`
}

type pageJSON struct {
	JournalDay int `json:"journalDay"`
}

func ExtractTasksFromJSON(jsonStr string) ([]TaskJSON, error) {
	var tasks []TaskJSON
	if err := json.Unmarshal([]byte(jsonStr), &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return tasks, nil
}

// Overdue checks if the task is overdue based on deadline or scheduled date.
func (t TaskJSON) Overdue() bool {
	currentDate := DateYYYYMMDD(time.Now())

	return (t.Deadline > 0 && t.Deadline <= currentDate) || (t.Scheduled > 0 && t.Scheduled <= currentDate)
}

type CategorizedTasks struct {
	All     *utils.Set[string]
	Overdue *utils.Set[string]
}

func NewCategorizedTasks() CategorizedTasks {
	return CategorizedTasks{
		All:     utils.NewSet[string](),
		Overdue: utils.NewSet[string](),
	}
}
