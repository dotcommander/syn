package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type goldFile struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	KeyInsights []string `json:"key_insights"`
}

// LoadDataset loads source_*.txt and gold_*.json pairs from a directory.
func LoadDataset(dir string) ([]Case, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dataset dir: %w", err)
	}

	bySuffix := map[string]Case{}
	sourceSuffixes := map[string]struct{}{}
	goldSuffixes := map[string]struct{}{}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if err := classifyEntry(name, dir, bySuffix, sourceSuffixes, goldSuffixes); err != nil {
			return nil, err
		}
	}

	if err := validatePairing(sourceSuffixes, goldSuffixes); err != nil {
		return nil, err
	}

	return collectCases(bySuffix, dir)
}

// classifyEntry routes a single directory entry into source or gold maps.
func classifyEntry(name, dir string, bySuffix map[string]Case, sourceSuffixes, goldSuffixes map[string]struct{}) error {
	if strings.HasPrefix(name, "source_") && strings.HasSuffix(name, ".txt") {
		return loadSourceFile(name, dir, bySuffix, sourceSuffixes)
	}
	if strings.HasPrefix(name, "gold_") && strings.HasSuffix(name, ".json") {
		return loadGoldFile(name, dir, bySuffix, goldSuffixes)
	}
	return nil
}

func loadSourceFile(name, dir string, bySuffix map[string]Case, sourceSuffixes map[string]struct{}) error {
	suffix := strings.TrimSuffix(strings.TrimPrefix(name, "source_"), ".txt")
	sourceSuffixes[suffix] = struct{}{}

	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return fmt.Errorf("read %s: %w", name, err)
	}

	c := bySuffix[suffix]
	c.ID = suffix
	c.Source = strings.TrimSpace(string(b))
	bySuffix[suffix] = c
	return nil
}

func loadGoldFile(name, dir string, bySuffix map[string]Case, goldSuffixes map[string]struct{}) error {
	suffix := strings.TrimSuffix(strings.TrimPrefix(name, "gold_"), ".json")
	goldSuffixes[suffix] = struct{}{}

	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return fmt.Errorf("read %s: %w", name, err)
	}

	var g goldFile
	if err := json.Unmarshal(b, &g); err != nil {
		return fmt.Errorf("parse %s: %w", name, err)
	}

	c := bySuffix[suffix]
	c.ID = caseID(g.ID, suffix)
	c.Title = g.Title
	c.GoldInsights = normalizeLines(g.KeyInsights)
	bySuffix[suffix] = c
	return nil
}

func caseID(goldID, suffix string) string {
	if goldID != "" {
		return goldID
	}
	return suffix
}

// validatePairing ensures every source has a gold and vice versa.
func validatePairing(sourceSuffixes, goldSuffixes map[string]struct{}) error {
	missingGold := missingKeys(sourceSuffixes, goldSuffixes)
	missingSource := missingKeys(goldSuffixes, sourceSuffixes)

	if len(missingGold) == 0 && len(missingSource) == 0 {
		return nil
	}

	return fmt.Errorf(
		"dataset has unmatched files (missing gold for: %s; missing source for: %s)",
		joinOrNone(missingGold),
		joinOrNone(missingSource),
	)
}

// missingKeys returns sorted keys in 'have' that are absent from 'want'.
func missingKeys(have, want map[string]struct{}) []string {
	var missing []string
	for k := range have {
		if _, ok := want[k]; !ok {
			missing = append(missing, k)
		}
	}
	sort.Strings(missing)
	return missing
}

func joinOrNone(ss []string) string {
	if len(ss) == 0 {
		return "none"
	}
	return strings.Join(ss, ", ")
}

func collectCases(bySuffix map[string]Case, dir string) ([]Case, error) {
	cases := make([]Case, 0, len(bySuffix))
	for _, c := range bySuffix {
		if c.Source == "" || len(c.GoldInsights) == 0 {
			continue
		}
		cases = append(cases, c)
	}

	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	if len(cases) == 0 {
		return nil, fmt.Errorf("no valid source_/gold_ pairs found in %s", dir)
	}

	return cases, nil
}
