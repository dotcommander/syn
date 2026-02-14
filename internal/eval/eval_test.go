package eval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseOutput(t *testing.T) {
	raw := "```json\n{\"tldr\":\"hi\",\"key_insights\":[\"a\"],\"evidence_quotes\":[\"b\"]}\n```"
	out, err := ParseOutput(raw)
	if err != nil {
		t.Fatalf("ParseOutput() error = %v", err)
	}
	if out.TLDR != "hi" || len(out.KeyInsights) != 1 || len(out.EvidenceQuotes) != 1 {
		t.Fatalf("unexpected parsed output: %+v", out)
	}
}

func TestParseOutputErrors(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "truncated-json", raw: `{"tldr":"x","key_insights":["a"]`},
		{name: "html-response", raw: `<html>bad gateway</html>`},
		{name: "not-json", raw: `just text`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseOutput(tc.raw)
			if err == nil {
				t.Fatalf("expected parse error")
			}
		})
	}
}

func TestScoreCase(t *testing.T) {
	c := Case{
		ID:     "01",
		Source: "Physics requires assumptions, units, and experiments.",
		GoldInsights: []string{
			"Assumptions must be explicit.",
			"Check units.",
		},
	}
	out := ParsedOutput{
		TLDR:           "Use disciplined physics reasoning.",
		KeyInsights:    []string{"Make assumptions explicit.", "Check units in calculations."},
		EvidenceQuotes: []string{"assumptions, units, and experiments"},
	}

	s := ScoreCase(c, out, 0.9)
	if s.Recall < 0.9 {
		t.Fatalf("expected high recall, got %.2f", s.Recall)
	}
	if !s.FormatCompliant {
		t.Fatalf("expected format compliant")
	}
}

func TestScoreCaseContradiction(t *testing.T) {
	c := Case{
		ID:     "01",
		Source: "Physics requires assumptions.",
		GoldInsights: []string{
			"Physics requires assumptions.",
		},
	}
	out := ParsedOutput{
		TLDR:           "summary",
		KeyInsights:    []string{"Physics requires no assumptions."},
		EvidenceQuotes: []string{"Physics requires assumptions."},
	}

	s := ScoreCase(c, out, 0.9)
	if s.Contradictions == 0 {
		t.Fatalf("expected contradiction to be detected")
	}
}

func TestLoadDataset(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "source_01.txt"), []byte("source text"), 0o600); err != nil {
		t.Fatal(err)
	}
	gold := `{"id":"01","title":"x","key_insights":["a","b"]}`
	if err := os.WriteFile(filepath.Join(dir, "gold_01.json"), []byte(gold), 0o600); err != nil {
		t.Fatal(err)
	}

	cases, err := LoadDataset(dir)
	if err != nil {
		t.Fatalf("LoadDataset() error = %v", err)
	}
	if len(cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(cases))
	}
}

func TestLoadDatasetErrors(t *testing.T) {
	t.Run("no-pairs", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "source_01.txt"), []byte("source only"), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := LoadDataset(dir)
		if err == nil {
			t.Fatalf("expected error for missing gold file")
		}
		if !strings.Contains(err.Error(), "unmatched files") {
			t.Fatalf("expected unmatched files error, got %v", err)
		}
	})

	t.Run("missing-source-for-gold", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "gold_01.json"), []byte(`{"id":"01","title":"x","key_insights":["a"]}`), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := LoadDataset(dir)
		if err == nil {
			t.Fatalf("expected error for missing source file")
		}
		if !strings.Contains(err.Error(), "missing source") {
			t.Fatalf("expected missing source error, got %v", err)
		}
	})

	t.Run("bad-gold-json", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "source_01.txt"), []byte("source text"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "gold_01.json"), []byte(`{"id":"01",`), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := LoadDataset(dir)
		if err == nil {
			t.Fatalf("expected parse error for malformed gold json")
		}
	})
}

func TestBuildModelSummary(t *testing.T) {
	cases := []CaseResult{
		{Score: Score{Recall: 1.0, QuoteCoverage: 1.0, Contradictions: 0, FormatCompliant: true, Pass: true}},
		{Score: Score{Recall: 0.5, QuoteCoverage: 0.5, Contradictions: 1, FormatCompliant: false, Pass: false}},
	}

	s := BuildModelSummary(cases, 0.9)
	if s.AverageRecall != 0.75 {
		t.Fatalf("expected avg recall 0.75, got %.2f", s.AverageRecall)
	}
	if s.TotalContradictions != 1 {
		t.Fatalf("expected 1 contradiction, got %d", s.TotalContradictions)
	}
	if s.FormatPassRate != 0.5 {
		t.Fatalf("expected format pass rate 0.5, got %.2f", s.FormatPassRate)
	}
	if s.OverallPass {
		t.Fatalf("expected overall pass false")
	}
}

func TestRenderReportAndSort(t *testing.T) {
	r := Report{
		GeneratedAt:     time.Date(2026, 2, 6, 17, 0, 0, 0, time.UTC),
		DatasetPath:     "testdata/eval/walter_lewin",
		RecallThreshold: 0.9,
		Models: []ModelResult{
			{ModelID: "m-low", Cases: []CaseResult{{CaseID: "01", Error: "parse failed"}}},
			{ModelID: "m-high", Cases: []CaseResult{{CaseID: "01"}}},
		},
	}

	md := RenderMarkdown(r)
	if !strings.Contains(md, "# syn eval report") {
		t.Fatalf("missing report title")
	}
	if !strings.Contains(md, "| Model | Parsed | Errors | Elapsed (s) | Tokens | Tok/s | TTFT (ms) |") {
		t.Fatalf("missing markdown table header")
	}
	if !strings.Contains(md, "| `m-low` | 0 | 1 | 0.00 | 0 | 0.0 | 0 |") {
		t.Fatalf("missing m-low case stats")
	}
	if !strings.Contains(md, "| `m-high` | 1 | 0 | 0.00 | 0 | 0.0 | 0 |") {
		t.Fatalf("missing m-high case stats")
	}
}

func TestBuildPrompt(t *testing.T) {
	source := "sample source"
	p := BuildPrompt(source)
	if !strings.Contains(p, "Return JSON only") {
		t.Fatalf("prompt missing JSON instruction")
	}
	if !strings.Contains(p, "\"key_insights\"") {
		t.Fatalf("prompt missing schema keys")
	}
	if !strings.Contains(p, source) {
		t.Fatalf("prompt missing source content")
	}
}
