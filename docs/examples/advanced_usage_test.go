package examples_test

import (
	"os"
	"strings"
	"testing"
)

func TestAdvancedUsageExists(t *testing.T) {
	path := "advanced-usage.md"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Advanced usage example does not exist: %s", path)
	}
}

func TestAdvancedUsageDemonstratesThreePlusFeatures(t *testing.T) {
	path := "advanced-usage.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read advanced usage: %v", err)
	}

	contentStr := string(content)

	// Check for at least 3 advanced feature examples with proper headings
	advancedFeatures := []string{
		"Custom Retry Configuration",
		"Conversational Context Management",
		"File Analysis with Vision",
		"Temperature Tuning",
		"Batch Embeddings",
		"Context-Aware Chat",
		"Configuration File Override",
	}

	foundCount := 0
	var foundFeatures []string

	for _, feature := range advancedFeatures {
		if strings.Contains(contentStr, "## "+feature) {
			foundCount++
			foundFeatures = append(foundFeatures, feature)
		}
	}

	if foundCount < 3 {
		t.Errorf("Expected at least 3 advanced features, found %d: %v", foundCount, foundFeatures)
	}
}

func TestAdvancedUsageIncludesErrorHandlingPatterns(t *testing.T) {
	path := "advanced-usage.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read advanced usage: %v", err)
	}

	contentStr := string(content)

	// Check for error handling patterns
	errorPatterns := []string{
		"if err != nil",
		"APIError",
		"os.IsNotExist",
		"panic",
		"log.Fatal",
	}

	foundPatterns := 0
	for _, pattern := range errorPatterns {
		if strings.Contains(contentStr, pattern) {
			foundPatterns++
		}
	}

	if foundPatterns < 2 {
		t.Errorf("Expected error handling patterns (at least 2), found %d", foundPatterns)
	}

	// Specifically check for structured error handling examples
	if !strings.Contains(contentStr, "apiErr.StatusCode") {
		t.Error("Missing structured error handling with status code checking")
	}

	// Check for file error handling
	if !strings.Contains(contentStr, "os.IsNotExist") {
		t.Error("Missing file existence check error handling")
	}
}

func TestAdvancedUsageDemonstratesConfigurationCustomization(t *testing.T) {
	path := "advanced-usage.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read advanced usage: %v", err)
	}

	contentStr := string(content)

	// Check for configuration customization examples
	configPatterns := []struct {
		pattern     string
		description string
	}{
		{"RetryConfig", "Retry configuration customization"},
		{"Temperature", "Temperature parameter configuration"},
		{"config.yaml", "Configuration file demonstration"},
		{"ClientConfig", "Client configuration struct"},
	}

	foundConfig := 0
	var missingConfigs []string

	for _, cp := range configPatterns {
		if strings.Contains(contentStr, cp.pattern) {
			foundConfig++
		} else {
			missingConfigs = append(missingConfigs, cp.description)
		}
	}

	if foundConfig < 2 {
		t.Errorf("Expected configuration customization examples (at least 2), found %d. Missing: %v",
			foundConfig, missingConfigs)
	}

	// Verify there's a config.yaml example with actual settings
	if !strings.Contains(contentStr, "yaml") && !strings.Contains(contentStr, ".yml") {
		t.Error("Missing YAML configuration file example")
	}

	// Check for specific config parameters being customized
	configParams := []string{"max_attempts", "initial_backoff", "temperature", "max_tokens"}
	foundParams := 0
	for _, param := range configParams {
		if strings.Contains(contentStr, param) {
			foundParams++
		}
	}

	if foundParams < 2 {
		t.Errorf("Expected specific config parameters in examples (at least 2), found %d", foundParams)
	}
}

func TestAdvancedUsageCodeExamplesAreExecutable(t *testing.T) {
	path := "advanced-usage.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read advanced usage: %v", err)
	}

	contentStr := string(content)

	// Check that code examples are in Go and have proper structure
	if !strings.Contains(contentStr, "```go") {
		t.Error("Missing Go code blocks in examples")
	}

	// Check for package declarations
	if !strings.Contains(contentStr, "package main") {
		t.Error("Missing proper package declarations in code examples")
	}

	// Check for main functions (indicates complete, runnable examples)
	mainCount := strings.Count(contentStr, "func main()")
	if mainCount < 3 {
		t.Errorf("Expected at least 3 complete code examples with main(), found %d", mainCount)
	}
}
