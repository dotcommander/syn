package eval

import (
	"fmt"
	"sort"
	"strings"
)

// SortByRecallDesc sorts models from best recall to worst.
func SortByRecallDesc(models []ModelResult) {
	sort.Slice(models, func(i, j int) bool {
		return models[i].Summary.AverageRecall > models[j].Summary.AverageRecall
	})
}

// RenderMarkdown returns a concise markdown report.
func RenderMarkdown(r Report) string {
	var b strings.Builder
	b.WriteString("# syn eval report\n\n")
	b.WriteString(fmt.Sprintf("- Generated: %s\n", r.GeneratedAt.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("- Dataset: `%s`\n", r.DatasetPath))
	b.WriteString("- Scoring: disabled (manual review workflow)\n\n")

	b.WriteString("| Model | Parsed | Errors | Elapsed (s) | Tokens | Tok/s | TTFT (ms) |\n")
	b.WriteString("|---|---:|---:|---:|---:|---:|---:|\n")
	for _, m := range r.Models {
		parsed, errs := caseStats(m.Cases)
		b.WriteString(fmt.Sprintf(
			"| `%s` | %d | %d | %.2f | %d | %.1f | %d |\n",
			m.ModelID,
			parsed,
			errs,
			float64(m.ElapsedMS)/1000,
			m.CompletionTokens,
			m.TokensPerSec,
			m.AvgTTFMS,
		))
	}

	b.WriteString("\n")
	return b.String()
}

func caseStats(cases []CaseResult) (parsed int, errs int) {
	for _, c := range cases {
		if strings.TrimSpace(c.Error) != "" {
			errs++
			continue
		}
		parsed++
	}
	return parsed, errs
}
