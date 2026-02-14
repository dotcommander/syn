package eval

import "time"

// Case is a single evaluation sample.
type Case struct {
	ID           string
	Title        string
	Source       string
	GoldInsights []string
}

// ParsedOutput is the normalized model output for scoring.
type ParsedOutput struct {
	TLDR           string   `json:"tldr"`
	KeyInsights    []string `json:"key_insights"`
	EvidenceQuotes []string `json:"evidence_quotes"`
}

// Score contains scoring metrics for one case.
type Score struct {
	Recall           float64 `json:"recall"`
	MissingInsights  int     `json:"missing_insights"`
	Contradictions   int     `json:"contradictions"`
	QuoteCoverage    float64 `json:"quote_coverage"`
	FormatCompliant  bool    `json:"format_compliant"`
	Pass             bool    `json:"pass"`
	MatchedGoldCount int     `json:"matched_gold_count"`
}

// CaseResult is one model response + score for one case.
type CaseResult struct {
	CaseID    string       `json:"case_id"`
	RawOutput string       `json:"raw_output"`
	Parsed    ParsedOutput `json:"parsed"`
	Score     Score        `json:"score"`
	TTFMS     int64        `json:"ttf_ms,omitempty"`
	Error     string       `json:"error,omitempty"`
}

// ModelSummary aggregates case-level scores for a model.
type ModelSummary struct {
	AverageRecall       float64 `json:"average_recall"`
	AverageCoverage     float64 `json:"average_quote_coverage"`
	TotalContradictions int     `json:"total_contradictions"`
	FormatPassRate      float64 `json:"format_pass_rate"`
	OverallPass         bool    `json:"overall_pass"`
}

// ModelResult includes all cases for one model.
type ModelResult struct {
	ModelID          string       `json:"model_id"`
	Cases            []CaseResult `json:"cases"`
	Summary          ModelSummary `json:"summary"`
	ElapsedMS        int64        `json:"elapsed_ms"`
	CompletionTokens int          `json:"completion_tokens"`
	TokensPerSec     float64      `json:"tokens_per_sec"`
	AvgTTFMS         int64        `json:"avg_ttf_ms"`
}

// Report is the top-level evaluation artifact.
type Report struct {
	GeneratedAt     time.Time     `json:"generated_at"`
	DatasetPath     string        `json:"dataset_path"`
	RecallThreshold float64       `json:"recall_threshold"`
	Models          []ModelResult `json:"models"`
}
