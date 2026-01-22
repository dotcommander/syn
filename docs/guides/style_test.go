package guides_test

import (
	"os"
	"strings"
	"testing"
)

func TestStyleGuideExists(t *testing.T) {
	path := "style-guide.md"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Style guide does not exist: %s", path)
	}
}

func TestStyleGuideDefinesVoiceAndTone(t *testing.T) {
	path := "style-guide.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read style guide: %v", err)
	}

	contentStr := string(content)

	// Check for voice/tone section with active voice and "you" addressing
	hasVoiceSection := strings.Contains(contentStr, "Voice and Tone") ||
		strings.Contains(contentStr, "## Voice and Tone")

	hasActiveVoice := strings.Contains(contentStr, "active voice")

	hasYouAddressing := strings.Contains(contentStr, `"you"`) ||
		strings.Contains(contentStr, "`you`") ||
		strings.Contains(contentStr, "as \"you\"") ||
		strings.Contains(contentStr, "as `you`")

	if !hasVoiceSection {
		t.Error("Style guide missing Voice and Tone section")
	}

	if !hasActiveVoice {
		t.Error("Style guide missing explanation of active voice")
	}

	if !hasYouAddressing {
		t.Error("Style guide missing instruction to address reader as 'you'")
	}
}

func TestStyleGuideSpecifiesCodeBlockLanguages(t *testing.T) {
	path := "style-guide.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read style guide: %v", err)
	}

	contentStr := string(content)

	// Check for code block language annotation rules
	hasCodeBlockSection := strings.Contains(contentStr, "Code Blocks") ||
		strings.Contains(contentStr, "## Code Blocks")

	hasLanguageAnnotation := strings.Contains(contentStr, "language") &&
		strings.Contains(contentStr, "annotation")

	hasFencedCodeSyntax := strings.Contains(contentStr, "```")

	if !hasCodeBlockSection {
		t.Error("Style guide missing Code Blocks section")
	}

	if !hasLanguageAnnotation {
		t.Error("Style guide missing explanation of language annotations")
	}

	if !hasFencedCodeSyntax {
		t.Error("Style guide missing fenced code syntax examples")
	}
}

func TestStyleGuideDocumentsHeadingHierarchy(t *testing.T) {
	path := "style-guide.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read style guide: %v", err)
	}

	contentStr := string(content)

	// Check for heading hierarchy rules
	hasHeadingSection := strings.Contains(contentStr, "Heading") ||
		strings.Contains(contentStr, "## Heading")

	hasHierarchyRules := strings.Contains(contentStr, "Hierarchy") ||
		strings.Contains(contentStr, "hierarchy")

	hasH1Rule := strings.Contains(contentStr, "H1") ||
		strings.Contains(contentStr, "#")

	hasH2Rule := strings.Contains(contentStr, "H2") ||
		strings.Contains(contentStr, "##")

	if !hasHeadingSection {
		t.Error("Style guide missing Heading section")
	}

	if !hasHierarchyRules {
		t.Error("Style guide missing heading hierarchy documentation")
	}

	if !hasH1Rule {
		t.Error("Style guide missing H1/title rules")
	}

	if !hasH2Rule {
		t.Error("Style guide missing H2/section rules")
	}
}
