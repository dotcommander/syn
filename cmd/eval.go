package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dotcommander/syn/internal/app"
	"github.com/dotcommander/syn/internal/eval"
)

const formatJSON = "json" // goconst: shared by flag default check and viper override

var ( //nolint:gochecknoglobals // cobra flag bindings require package-level vars
	evalDatasetPath    string
	evalOutputPath     string
	evalFormat         string
	evalModelFilterCSV string
	evalCaseLimit      int
	evalRecallMin      float64
	evalHistoryPath    string
	evalLeaderboardOut string
	evalLeaderboardTop int
	evalNoHistory      bool
	evalResponsesDir   string
)

var evalModelDenylist = map[string]struct{}{ //nolint:gochecknoglobals // static config
	"hf:deepseek-ai/DeepSeek-R1-0528":       {},
	"hf:deepseek-ai/DeepSeek-V3":            {},
	"hf:deepseek-ai/DeepSeek-V3-0324":       {},
	"hf:MiniMaxAI/MiniMax-M2":               {},
	"hf:Qwen/Qwen3-235B-A22B-Thinking-2507": {},
	"hf:zai-org/GLM-4.6":                    {},
	"hf:moonshotai/Kimi-K2-Instruct-0905":   {},
	"hf:meta-llama/Llama-3.3-70B-Instruct":  {},
}

var evalCmd = &cobra.Command{ //nolint:gochecknoglobals // cobra command registration
	Use:   "eval",
	Short: "Evaluate model insight extraction",
	Long: `Run a lightweight evaluation across models for lossless key-insight extraction.

Examples:
  syn eval
  syn eval --dataset testdata/eval/walter_lewin --format json
  syn eval --models "hf:deepseek-ai/DeepSeek-V3.2,hf:moonshotai/Kimi-K2-Thinking"
  syn eval --out analysis-results/eval-report.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetBool("json") {
			evalFormat = formatJSON
		}
		return runEval(cmd.Context())
	},
}

func init() { //nolint:gochecknoinits // cobra command registration
	rootCmd.AddCommand(evalCmd)
	evalCmd.Flags().StringVar(&evalDatasetPath, "dataset", "testdata/eval/walter_lewin", "dataset directory containing source_*.txt and gold_*.json")
	evalCmd.Flags().StringVar(&evalOutputPath, "out", "", "write report to file (optional)")
	evalCmd.Flags().StringVar(&evalFormat, "format", "md", "output format: md or json")
	evalCmd.Flags().StringVar(&evalModelFilterCSV, "models", "", "comma-separated model IDs to evaluate (default: all from syn model list)")
	evalCmd.Flags().IntVar(&evalCaseLimit, "limit", 0, "max dataset cases to evaluate (0 = all)")
	evalCmd.Flags().Float64Var(&evalRecallMin, "recall-threshold", 0.90, "minimum recall required for pass")
	evalCmd.Flags().StringVar(&evalHistoryPath, "history", "analysis-results/eval-history.jsonl", "jsonl file for appending model run scores")
	evalCmd.Flags().StringVar(&evalLeaderboardOut, "leaderboard-out", "analysis-results/eval-leaderboard.md", "path to write leaderboard markdown (empty disables write)")
	evalCmd.Flags().IntVar(&evalLeaderboardTop, "leaderboard-top", 10, "number of leaderboard rows to print")
	evalCmd.Flags().BoolVar(&evalNoHistory, "no-history", false, "disable history append and leaderboard updates")
	evalCmd.Flags().StringVar(&evalResponsesDir, "responses-dir", "analysis-results/eval-responses", "base directory to save per-run raw model responses and scores")
}

func runEval(parent context.Context) error {
	if evalFormat != "md" && evalFormat != formatJSON {
		return fmt.Errorf("invalid --format %q (expected md or json)", evalFormat)
	}
	humanOutput := evalFormat == "md"

	client := newClient()
	cases, err := eval.LoadDataset(evalDatasetPath)
	if err != nil {
		return fmt.Errorf("failed to load dataset: %w", err)
	}
	if evalCaseLimit > 0 && evalCaseLimit < len(cases) {
		cases = cases[:evalCaseLimit]
	}

	selected, err := fetchAndSelectModels(parent, client)
	if err != nil {
		return err
	}

	if humanOutput {
		printEvalBanner(len(selected), len(cases))
	}

	report := buildEvalReport(parent, client, selected, cases, humanOutput)

	return finalizeEvalReport(report, humanOutput)
}

func fetchAndSelectModels(parent context.Context, client *app.Client) ([]app.Model, error) {
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	models, err := client.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	selected := selectModels(models, evalModelFilterCSV)
	if len(selected) == 0 {
		return nil, fmt.Errorf("no models selected")
	}
	return selected, nil
}

func printEvalBanner(modelCount, caseCount int) {
	fmt.Println()
	fmt.Println(theme.Section.Render(fmt.Sprintf("Running eval (%d models, %d cases)", modelCount, caseCount)))
	fmt.Println(theme.Divider.Render(strings.Repeat("-", 60)))
}

func buildEvalReport(parent context.Context, client *app.Client, selected []app.Model, cases []eval.Case, humanOutput bool) eval.Report {
	report := eval.Report{
		GeneratedAt:     time.Now(),
		DatasetPath:     evalDatasetPath,
		RecallThreshold: evalRecallMin,
		Models:          make([]eval.ModelResult, 0, len(selected)),
	}

	for _, m := range selected {
		result := evalModel(parent, client, m.ID, cases)
		report.Models = append(report.Models, result)
		if humanOutput {
			parsed, errs := modelCaseStats(result)
			fmt.Printf("  %s parsed=%d errors=%d elapsed=%.2fs tok/s=%.1f ttft=%dms\n", theme.Command.Render(m.ID), parsed, errs, float64(result.ElapsedMS)/1000, result.TokensPerSec, result.AvgTTFMS)
		}
	}
	return report
}

func finalizeEvalReport(report eval.Report, humanOutput bool) error {
	out, renderErr := renderReport(report, evalFormat)
	if renderErr != nil {
		return renderErr
	}

	if writeErr := writeReportFile(out); writeErr != nil {
		return writeErr
	}

	if humanOutput {
		fmt.Println()
	}
	fmt.Println(out)
	if humanOutput && evalOutputPath != "" {
		fmt.Printf("Saved report to %s\n", evalOutputPath)
	}

	responsesPath, artifactErr := writeResponseArtifacts(report, evalResponsesDir)
	if artifactErr != nil {
		return artifactErr
	}
	if humanOutput && responsesPath != "" {
		fmt.Printf("Saved responses to %s\n", responsesPath)
	}

	printErrorCount(report, humanOutput)

	return maybeWriteLeaderboard(report, responsesPath, humanOutput)
}

func writeReportFile(out string) error {
	if evalOutputPath == "" {
		return nil
	}
	if mkdirErr := os.MkdirAll(filepath.Dir(evalOutputPath), 0o755); mkdirErr != nil {
		return fmt.Errorf("failed to prepare output dir: %w", mkdirErr)
	}
	if writeErr := os.WriteFile(evalOutputPath, []byte(out), 0o600); writeErr != nil {
		return fmt.Errorf("failed to write report: %w", writeErr)
	}
	return nil
}

func printErrorCount(report eval.Report, humanOutput bool) {
	if !humanOutput {
		return
	}
	errorCount := countCaseErrors(report)
	fmt.Printf("Case errors: %d\n", errorCount)
}

func maybeWriteLeaderboard(report eval.Report, responsesPath string, humanOutput bool) error {
	if strings.TrimSpace(evalLeaderboardOut) == "" {
		return nil
	}

	created, err := ensureManualLeaderboard(report, responsesPath, evalLeaderboardOut)
	if err != nil {
		return err
	}

	if !humanOutput {
		return nil
	}
	if created {
		fmt.Printf("Created manual leaderboard template at %s\n", evalLeaderboardOut)
	} else {
		fmt.Printf("Left existing manual leaderboard unchanged at %s\n", evalLeaderboardOut)
	}
	return nil
}

var nonFileRe = regexp.MustCompile(`[^a-zA-Z0-9._-]+`) //nolint:gochecknoglobals // compiled regex

func sanitizeFilePart(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unknown"
	}
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, ":", "_")
	s = nonFileRe.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if s == "" {
		return "unknown"
	}
	return s
}

func writeResponseArtifacts(report eval.Report, baseDir string) (string, error) {
	if strings.TrimSpace(baseDir) == "" {
		return "", nil
	}
	runDir := filepath.Join(baseDir, report.GeneratedAt.Format("20060102-150405"))
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create responses dir: %w", err)
	}

	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report json: %w", err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "report.json"), reportJSON, 0o600); err != nil {
		return "", fmt.Errorf("failed to write report json: %w", err)
	}

	for _, model := range report.Models {
		if err := writeModelCases(runDir, model); err != nil {
			return "", err
		}
	}

	return runDir, nil
}

func writeModelCases(runDir string, model eval.ModelResult) error {
	modelDir := filepath.Join(runDir, sanitizeFilePart(model.ModelID))
	if err := os.MkdirAll(modelDir, 0o755); err != nil {
		return fmt.Errorf("failed to create model dir: %w", err)
	}
	for _, c := range model.Cases {
		b, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal case result: %w", err)
		}
		name := fmt.Sprintf("case_%s.json", sanitizeFilePart(c.CaseID))
		if err := os.WriteFile(filepath.Join(modelDir, name), b, 0o600); err != nil {
			return fmt.Errorf("failed to write case result: %w", err)
		}
	}
	return nil
}

func countCaseErrors(report eval.Report) int {
	count := 0
	for _, m := range report.Models {
		for _, c := range m.Cases {
			if strings.TrimSpace(c.Error) != "" {
				count++
			}
		}
	}
	return count
}

func modelCaseStats(res eval.ModelResult) (parsed int, errors int) {
	for _, c := range res.Cases {
		if strings.TrimSpace(c.Error) != "" {
			errors++
			continue
		}
		parsed++
	}
	return parsed, errors
}

func ensureManualLeaderboard(report eval.Report, responsesPath, outPath string) (bool, error) {
	if shouldKeepExisting(outPath) {
		return false, nil
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return false, fmt.Errorf("failed to prepare leaderboard dir: %w", err)
	}

	content := buildLeaderboardContent(report, responsesPath)
	if err := os.WriteFile(outPath, []byte(content), 0o600); err != nil {
		return false, fmt.Errorf("failed to write manual leaderboard: %w", err)
	}

	return true, nil
}

// shouldKeepExisting returns true when an existing leaderboard file should not be overwritten.
func shouldKeepExisting(outPath string) bool {
	info, err := os.Stat(outPath)
	if err != nil || info.IsDir() {
		return false
	}
	existing, readErr := os.ReadFile(outPath)
	if readErr != nil {
		return true // can't read, don't overwrite
	}
	text := string(existing)
	return strings.Contains(text, "# syn eval manual leaderboard") ||
		!strings.Contains(text, "# syn eval leaderboard")
}

func buildLeaderboardContent(report eval.Report, responsesPath string) string {
	var b strings.Builder
	b.WriteString("# syn eval manual leaderboard\n\n")
	b.WriteString("Update ranks manually after reviewing model outputs.\n\n")
	fmt.Fprintf(&b, "- generated_at: %s\n", report.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "- dataset: `%s`\n", report.DatasetPath)
	fmt.Fprintf(&b, "- responses_path: `%s`\n\n", responsesPath)
	b.WriteString("| Rank | Model | Notes |\n")
	b.WriteString("|---:|---|---|\n")
	for _, m := range report.Models {
		fmt.Fprintf(&b, "|  | `%s` |  |\n", m.ModelID)
	}
	return b.String()
}

func evalModel(parent context.Context, client *app.Client, modelID string, cases []eval.Case) eval.ModelResult {
	res := eval.ModelResult{ModelID: modelID, Cases: make([]eval.CaseResult, 0, len(cases))}
	started := time.Now()

	totalCompletionTokens := 0
	var totalTTFMS int64
	ttfCount := 0

	for _, c := range cases {
		prompt := eval.BuildPrompt(c.Source)
		opts := app.ChatOptions{
			Model: modelID,
			TopP:  app.Float64Ptr(1.0),
		}

		ctx, cancel := context.WithTimeout(parent, 2*time.Minute)
		sr, chatErr := client.ChatStream(ctx, prompt, opts)
		cancel()

		totalCompletionTokens += sr.Usage.CompletionTokens

		caseResult := eval.CaseResult{CaseID: c.ID, RawOutput: sr.Content, TTFMS: sr.TTFMS}
		if chatErr != nil {
			caseResult.Error = chatErr.Error()
			res.Cases = append(res.Cases, caseResult)
			continue
		}

		if sr.TTFMS > 0 {
			totalTTFMS += sr.TTFMS
			ttfCount++
		}

		parsed, parseErr := eval.ParseOutput(sr.Content)
		if parseErr != nil {
			caseResult.Error = parseErr.Error()
			res.Cases = append(res.Cases, caseResult)
			continue
		}

		caseResult.Parsed = parsed
		res.Cases = append(res.Cases, caseResult)
	}

	res.Summary = eval.ModelSummary{}
	res.ElapsedMS = time.Since(started).Milliseconds()
	res.CompletionTokens = totalCompletionTokens
	if res.ElapsedMS > 0 {
		res.TokensPerSec = float64(totalCompletionTokens) / (float64(res.ElapsedMS) / 1000)
	}
	if ttfCount > 0 {
		res.AvgTTFMS = totalTTFMS / int64(ttfCount)
	}
	return res
}

func renderReport(r eval.Report, format string) (string, error) {
	if format == formatJSON {
		b, err := json.MarshalIndent(r, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal report: %w", err)
		}
		return string(b), nil
	}
	return eval.RenderMarkdown(r), nil
}

func selectModels(models []app.Model, csv string) []app.Model {
	isDenied := func(modelID string) bool {
		_, blocked := evalModelDenylist[modelID]
		return blocked
	}

	if strings.TrimSpace(csv) == "" {
		selected := make([]app.Model, 0, len(models))
		for _, m := range models {
			if isDenied(m.ID) {
				continue
			}
			selected = append(selected, m)
		}
		return selected
	}

	allow := map[string]struct{}{}
	for m := range strings.SplitSeq(csv, ",") {
		m = strings.TrimSpace(m)
		if m != "" {
			allow[m] = struct{}{}
		}
	}

	selected := make([]app.Model, 0, len(allow))
	for _, m := range models {
		if isDenied(m.ID) {
			continue
		}
		if _, ok := allow[m.ID]; ok {
			selected = append(selected, m)
		}
	}
	return selected
}
