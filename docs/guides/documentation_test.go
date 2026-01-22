package guides_test

import (
	"os"
	"strings"
	"testing"
)

func TestDocumentationGuideExists(t *testing.T) {
	path := "documentation.md"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Documentation guide does not exist: %s", path)
	}
}

func TestDocumentationGuideExplainsSyntax(t *testing.T) {
	path := "documentation.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read documentation guide: %v", err)
	}

	contentStr := string(content)

	// Check for Go doc comment syntax explanation
	expectedTerms := []string{
		"//",
		"doc comment",
		"exported",
		"unexported",
		"PascalCase",
		"camelCase",
	}

	for _, term := range expectedTerms {
		if !strings.Contains(contentStr, term) {
			t.Errorf("Documentation guide missing explanation of: %s", term)
		}
	}
}

func TestDocumentationGuideHasGoodVsBadExamples(t *testing.T) {
	path := "documentation.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read documentation guide: %v", err)
	}

	contentStr := string(content)

	// Check for good vs bad examples sections
	hasGoodSection := strings.Contains(contentStr, "// ✅ GOOD")
	hasBadSection := strings.Contains(contentStr, "// ❌ BAD")

	if !hasGoodSection {
		t.Error("Documentation guide missing good examples")
	}

	if !hasBadSection {
		t.Error("Documentation guide missing bad examples")
	}
}

func TestDocumentationGuideCoversAllRequiredSections(t *testing.T) {
	path := "documentation.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read documentation guide: %v", err)
	}

	contentStr := string(content)

	// Check for required sections based on acceptance criteria
	requiredSections := map[string]string{
		"summary line":     "Summary line",
		"description":      "Description",
		"parameters":       "Param name:",
		"returns":          "Returns:",
		"examples":         "Example:",
	}

	for sectionKey, sectionMarker := range requiredSections {
		if !strings.Contains(contentStr, sectionMarker) {
			t.Errorf("Documentation guide missing section: %s (looked for marker: %s)", sectionKey, sectionMarker)
		}
	}
}

func TestDocumentationGuideHasFunctionExamples(t *testing.T) {
	path := "documentation.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read documentation guide: %v", err)
	}

	contentStr := string(content)

	// Check for function documentation examples
	hasFuncExample := strings.Contains(contentStr, "func ") &&
		strings.Contains(contentStr, "Param ") &&
		strings.Contains(contentStr, "Returns:")

	if !hasFuncExample {
		t.Error("Documentation guide missing function documentation examples with parameters and returns")
	}
}

func TestDocumentationGuideHasTypeExamples(t *testing.T) {
	path := "documentation.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read documentation guide: %v", err)
	}

	contentStr := string(content)

	// Check for type/struct documentation examples
	hasTypeExample := strings.Contains(contentStr, "type ") &&
		strings.Contains(contentStr, "struct")

	if !hasTypeExample {
		t.Error("Documentation guide missing type/struct documentation examples")
	}
}
