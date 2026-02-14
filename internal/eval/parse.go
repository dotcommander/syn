package eval

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseOutput parses model output into ParsedOutput.
func ParseOutput(raw string) (ParsedOutput, error) {
	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	if idx := strings.Index(clean, "{"); idx >= 0 {
		clean = clean[idx:]
	}
	if idx := strings.LastIndex(clean, "}"); idx >= 0 {
		clean = clean[:idx+1]
	}

	type payload struct {
		TLDR           string   `json:"tldr"`
		KeyInsights    []string `json:"key_insights"`
		EvidenceQuotes []string `json:"evidence_quotes"`
	}

	var p payload
	if err := json.Unmarshal([]byte(clean), &p); err != nil {
		return ParsedOutput{}, fmt.Errorf("invalid JSON output: %w", err)
	}

	return ParsedOutput{
		TLDR:           strings.TrimSpace(p.TLDR),
		KeyInsights:    normalizeLines(p.KeyInsights),
		EvidenceQuotes: normalizeLines(p.EvidenceQuotes),
	}, nil
}

func normalizeLines(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
