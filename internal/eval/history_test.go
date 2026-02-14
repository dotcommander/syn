package eval

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendLoadAndLeaderboard(t *testing.T) {
	history := filepath.Join(t.TempDir(), "eval-history.jsonl")

	r1 := Report{
		GeneratedAt:     time.Date(2026, 2, 6, 10, 0, 0, 0, time.UTC),
		DatasetPath:     "testdata/eval/walter_lewin",
		RecallThreshold: 0.90,
		Models: []ModelResult{
			{ModelID: "m1", Cases: []CaseResult{{CaseID: "01"}}, Summary: ModelSummary{AverageRecall: 0.6, AverageCoverage: 1.0, TotalContradictions: 0, FormatPassRate: 1.0, OverallPass: false}},
			{ModelID: "m2", Cases: []CaseResult{{CaseID: "01"}}, Summary: ModelSummary{AverageRecall: 0.8, AverageCoverage: 1.0, TotalContradictions: 0, FormatPassRate: 1.0, OverallPass: false}},
		},
	}
	r2 := Report{
		GeneratedAt:     time.Date(2026, 2, 7, 10, 0, 0, 0, time.UTC),
		DatasetPath:     "testdata/eval/walter_lewin",
		RecallThreshold: 0.90,
		Models: []ModelResult{
			{ModelID: "m1", Cases: []CaseResult{{CaseID: "01"}}, Summary: ModelSummary{AverageRecall: 0.9, AverageCoverage: 1.0, TotalContradictions: 0, FormatPassRate: 1.0, OverallPass: true}},
		},
	}

	if err := AppendHistory(history, r1); err != nil {
		t.Fatalf("AppendHistory(r1) error = %v", err)
	}
	if err := AppendHistory(history, r2); err != nil {
		t.Fatalf("AppendHistory(r2) error = %v", err)
	}

	records, err := LoadHistory(history)
	if err != nil {
		t.Fatalf("LoadHistory() error = %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}

	rows := BuildLeaderboard(records)
	if len(rows) != 2 {
		t.Fatalf("expected 2 models in leaderboard, got %d", len(rows))
	}

	if rows[0].ModelID != "m2" {
		t.Fatalf("expected m2 first by avg recall, got %s", rows[0].ModelID)
	}
	if rows[1].ModelID != "m1" {
		t.Fatalf("expected m1 second, got %s", rows[1].ModelID)
	}
	if rows[1].Runs != 2 {
		t.Fatalf("expected m1 runs=2, got %d", rows[1].Runs)
	}
	if rows[1].BestRecall < 0.9 {
		t.Fatalf("expected m1 best recall >= 0.9, got %.2f", rows[1].BestRecall)
	}
}

func TestRenderLeaderboardMarkdown(t *testing.T) {
	rows := []LeaderboardRow{
		{
			ModelID:             "m1",
			Runs:                2,
			AverageRecall:       0.75,
			BestRecall:          0.9,
			AverageCoverage:     1.0,
			TotalContradictions: 1,
			OverallPassRate:     0.5,
			LastSeen:            time.Date(2026, 2, 7, 10, 0, 0, 0, time.UTC),
		},
	}

	md := RenderLeaderboardMarkdown(rows)
	if !strings.Contains(md, "fields: rank, model") {
		t.Fatalf("missing field legend")
	}
	if !strings.Contains(md, "`m1`") {
		t.Fatalf("missing model id")
	}
	if !strings.Contains(md, "average_recall: 0.75") {
		t.Fatalf("missing average_recall value")
	}
}

func TestFilterHistory(t *testing.T) {
	records := []RunRecord{
		{ModelID: "m1", DatasetPath: "testdata/eval/walter_lewin", RecallThreshold: 0.90},
		{ModelID: "m2", DatasetPath: "testdata/eval/walter_lewin", RecallThreshold: 0.85},
		{ModelID: "m3", DatasetPath: "testdata/eval/other", RecallThreshold: 0.90},
	}

	filtered := FilterHistory(records, "testdata/eval/walter_lewin", 0.90)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered record, got %d", len(filtered))
	}
	if filtered[0].ModelID != "m1" {
		t.Fatalf("expected m1, got %s", filtered[0].ModelID)
	}
}
