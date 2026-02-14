package eval

import (
	"regexp"
	"strings"
)

var nonWordRe = regexp.MustCompile(`[^a-z0-9\s]+`)

// ScoreCase computes simple extraction quality metrics.
func ScoreCase(c Case, out ParsedOutput, recallThreshold float64) Score {
	gold := normalizeLines(c.GoldInsights)
	matched := 0
	for _, g := range gold {
		if hasInsightMatch(g, out.KeyInsights) {
			matched++
		}
	}

	recall := 0.0
	if len(gold) > 0 {
		recall = float64(matched) / float64(len(gold))
	}

	quoteHits := 0
	for _, q := range out.EvidenceQuotes {
		if strings.Contains(strings.ToLower(c.Source), strings.ToLower(strings.TrimSpace(q))) {
			quoteHits++
		}
	}

	coverage := 0.0
	if len(out.EvidenceQuotes) > 0 {
		coverage = float64(quoteHits) / float64(len(out.EvidenceQuotes))
	}

	contradictions := estimateContradictions(gold, out.KeyInsights)
	formatOK := out.TLDR != "" && len(out.KeyInsights) > 0 && len(out.EvidenceQuotes) > 0

	return Score{
		Recall:           recall,
		MissingInsights:  missingCount(len(gold), matched),
		Contradictions:   contradictions,
		QuoteCoverage:    coverage,
		FormatCompliant:  formatOK,
		Pass:             recall >= recallThreshold && contradictions == 0 && formatOK,
		MatchedGoldCount: matched,
	}
}

func missingCount(total, matched int) int {
	v := total - matched
	if v < 0 {
		return 0
	}
	return v
}

// BuildModelSummary aggregates results from a model's cases.
func BuildModelSummary(cases []CaseResult, recallThreshold float64) ModelSummary {
	if len(cases) == 0 {
		return ModelSummary{}
	}

	var totalRecall float64
	var totalCoverage float64
	var totalContradictions int
	var formatPasses int
	var passCount int

	for _, c := range cases {
		totalRecall += c.Score.Recall
		totalCoverage += c.Score.QuoteCoverage
		totalContradictions += c.Score.Contradictions
		if c.Score.FormatCompliant {
			formatPasses++
		}
		if c.Score.Pass {
			passCount++
		}
	}

	caseCount := float64(len(cases))
	avgRecall := totalRecall / caseCount

	return ModelSummary{
		AverageRecall:       avgRecall,
		AverageCoverage:     totalCoverage / caseCount,
		TotalContradictions: totalContradictions,
		FormatPassRate:      float64(formatPasses) / caseCount,
		OverallPass:         avgRecall >= recallThreshold && totalContradictions == 0 && passCount == len(cases),
	}
}

func hasInsightMatch(gold string, predicted []string) bool {
	g := normalizeText(gold)
	for _, p := range predicted {
		np := normalizeText(p)
		if np == "" {
			continue
		}
		if strings.Contains(np, g) || strings.Contains(g, np) {
			return true
		}
		if tokenOverlap(g, np) >= 0.55 {
			return true
		}
	}
	return false
}

func estimateContradictions(gold []string, predicted []string) int {
	count := 0
	for _, p := range predicted {
		np := normalizeText(p)
		if np == "" {
			continue
		}
		if !containsNegation(np) {
			continue
		}

		closestHasNegation := false
		best := 0.0
		for _, g := range gold {
			ng := normalizeText(g)
			s := tokenOverlap(np, ng)
			if s > best {
				best = s
				closestHasNegation = containsNegation(ng)
			}
		}
		if best >= 0.45 && !closestHasNegation {
			count++
		}
	}
	return count
}

func containsNegation(s string) bool {
	for _, n := range []string{" not ", " no ", " never ", " cannot ", " can't ", " without "} {
		if strings.Contains(" "+s+" ", n) {
			return true
		}
	}
	return false
}

func tokenOverlap(a, b string) float64 {
	aSet := tokenSet(a)
	bSet := tokenSet(b)
	if len(aSet) == 0 || len(bSet) == 0 {
		return 0
	}
	inter := 0
	for t := range aSet {
		if _, ok := bSet[t]; ok {
			inter++
		}
	}
	den := len(aSet)
	if len(bSet) > den {
		den = len(bSet)
	}
	return float64(inter) / float64(den)
}

func tokenSet(s string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, t := range strings.Fields(normalizeText(s)) {
		if len(t) >= 3 {
			out[t] = struct{}{}
		}
	}
	return out
}

func normalizeText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonWordRe.ReplaceAllString(s, " ")
	return strings.Join(strings.Fields(s), " ")
}
