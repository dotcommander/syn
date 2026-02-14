package eval

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RunRecord stores one model result from one eval run.
type RunRecord struct {
	GeneratedAt     time.Time `json:"generated_at"`
	DatasetPath     string    `json:"dataset_path"`
	RecallThreshold float64   `json:"recall_threshold"`
	ModelID         string    `json:"model_id"`
	CaseCount       int       `json:"case_count"`
	AverageRecall   float64   `json:"average_recall"`
	AverageCoverage float64   `json:"average_quote_coverage"`
	Contradictions  int       `json:"total_contradictions"`
	FormatPassRate  float64   `json:"format_pass_rate"`
	OverallPass     bool      `json:"overall_pass"`
}

// LeaderboardRow aggregates scores across run history.
type LeaderboardRow struct {
	ModelID             string
	Runs                int
	AverageRecall       float64
	BestRecall          float64
	AverageCoverage     float64
	TotalContradictions int
	OverallPassRate     float64
	LastSeen            time.Time
}

// AppendHistory appends one jsonl record per model in report.
func AppendHistory(path string, report Report) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open history file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, m := range report.Models {
		rec := RunRecord{
			GeneratedAt:     report.GeneratedAt,
			DatasetPath:     report.DatasetPath,
			RecallThreshold: report.RecallThreshold,
			ModelID:         m.ModelID,
			CaseCount:       len(m.Cases),
			AverageRecall:   m.Summary.AverageRecall,
			AverageCoverage: m.Summary.AverageCoverage,
			Contradictions:  m.Summary.TotalContradictions,
			FormatPassRate:  m.Summary.FormatPassRate,
			OverallPass:     m.Summary.OverallPass,
		}
		if err := enc.Encode(rec); err != nil {
			return fmt.Errorf("append history record: %w", err)
		}
	}

	return nil
}

// LoadHistory reads jsonl run records.
func LoadHistory(path string) ([]RunRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("open history file: %w", err)
	}
	defer f.Close()

	rows := make([]RunRecord, 0, 128)
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		var rec RunRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			continue
		}
		rows = append(rows, rec)
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("scan history file: %w", err)
	}

	return rows, nil
}

// FilterHistory keeps records matching dataset and recall threshold.
func FilterHistory(records []RunRecord, datasetPath string, recallThreshold float64) []RunRecord {
	if len(records) == 0 {
		return nil
	}

	targetDataset := filepath.Clean(strings.TrimSpace(datasetPath))
	const epsilon = 1e-9

	filtered := make([]RunRecord, 0, len(records))
	for _, r := range records {
		if filepath.Clean(strings.TrimSpace(r.DatasetPath)) != targetDataset {
			continue
		}
		if math.Abs(r.RecallThreshold-recallThreshold) > epsilon {
			continue
		}
		filtered = append(filtered, r)
	}

	return filtered
}

// BuildLeaderboard builds per-model aggregates sorted by performance.
func BuildLeaderboard(records []RunRecord) []LeaderboardRow {
	type agg struct {
		runs           int
		recall         float64
		bestRecall     float64
		coverage       float64
		contradictions int
		passes         int
		lastSeen       time.Time
	}

	byModel := map[string]*agg{}
	for _, r := range records {
		a, ok := byModel[r.ModelID]
		if !ok {
			a = &agg{}
			byModel[r.ModelID] = a
		}
		a.runs++
		a.recall += r.AverageRecall
		a.coverage += r.AverageCoverage
		a.contradictions += r.Contradictions
		if r.AverageRecall > a.bestRecall {
			a.bestRecall = r.AverageRecall
		}
		if r.OverallPass {
			a.passes++
		}
		if r.GeneratedAt.After(a.lastSeen) {
			a.lastSeen = r.GeneratedAt
		}
	}

	rows := make([]LeaderboardRow, 0, len(byModel))
	for model, a := range byModel {
		runs := float64(a.runs)
		rows = append(rows, LeaderboardRow{
			ModelID:             model,
			Runs:                a.runs,
			AverageRecall:       a.recall / runs,
			BestRecall:          a.bestRecall,
			AverageCoverage:     a.coverage / runs,
			TotalContradictions: a.contradictions,
			OverallPassRate:     float64(a.passes) / runs,
			LastSeen:            a.lastSeen,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		if rows[i].AverageRecall != rows[j].AverageRecall {
			return rows[i].AverageRecall > rows[j].AverageRecall
		}
		if rows[i].BestRecall != rows[j].BestRecall {
			return rows[i].BestRecall > rows[j].BestRecall
		}
		return rows[i].Runs > rows[j].Runs
	})

	return rows
}

// RenderLeaderboardMarkdown renders leaderboard as a non-table list to avoid
// truncation in narrow terminals/renderers.
func RenderLeaderboardMarkdown(rows []LeaderboardRow) string {
	var b strings.Builder
	b.WriteString("# syn eval leaderboard\n\n")
	b.WriteString("fields: rank, model, runs, average_recall, best_recall, average_coverage, total_contradictions, pass_rate, last_seen\n\n")
	for i, r := range rows {
		b.WriteString(fmt.Sprintf(
			"%d) `%s`\n- runs: %d\n- average_recall: %.2f\n- best_recall: %.2f\n- average_coverage: %.2f\n- total_contradictions: %d\n- pass_rate: %.2f\n- last_seen: %s\n\n",
			i+1,
			r.ModelID,
			r.Runs,
			r.AverageRecall,
			r.BestRecall,
			r.AverageCoverage,
			r.TotalContradictions,
			r.OverallPassRate,
			r.LastSeen.Format(time.RFC3339),
		))
	}
	return b.String()
}
