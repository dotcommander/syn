package api_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAPIIndexExists(t *testing.T) {
	path := "README.md"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("API index does not exist: %s", path)
	}
}

func TestAPIIndexListsAllPackages(t *testing.T) {
	path := "README.md"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read API index: %v", err)
	}

	contentStr := string(content)

	// Check for all public packages
	expectedPackages := []string{
		"app",
		"cmd",
		"config",
	}

	for _, pkg := range expectedPackages {
		if !strings.Contains(contentStr, pkg) {
			t.Errorf("Package %q not listed in API index", pkg)
		}
	}
}

func TestAPIIndexHasDescriptions(t *testing.T) {
	content, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("Failed to read API index: %v", err)
	}

	contentStr := string(content)

	packageDescriptions := map[string]string{
		"./app.md":    "Core client library",
		"./cmd.md":    "CLI commands",
		"./config.md": "Configuration management",
	}

	for link, desc := range packageDescriptions {
		assertLinkHasDescription(t, contentStr, link, desc)
	}
}

func assertLinkHasDescription(t *testing.T, content, link, description string) {
	t.Helper()

	if !strings.Contains(content, link) {
		return // link not present, other tests cover this
	}
	if !strings.Contains(content, description) {
		t.Errorf("link %s missing description %q", link, description)
	}
}

func TestAPIIndexLinksToDocumentation(t *testing.T) {
	indexPath := "README.md"
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("Failed to read API index: %v", err)
	}

	contentStr := string(content)

	// Check that each package links to its documentation
	expectedLinks := map[string]string{
		"app":    "./app.md",
		"cmd":    "./cmd.md",
		"config": "./config.md",
	}

	for pkg, link := range expectedLinks {
		if !strings.Contains(contentStr, link) {
			t.Errorf("Package %q does not link to %s", pkg, link)
		}

		// Verify the linked file exists
		docPath := filepath.Base(link)
		if _, err := os.Stat(docPath); os.IsNotExist(err) {
			t.Errorf("Documentation file does not exist: %s", docPath)
		}
	}
}

func TestPackageDocumentationFilesExist(t *testing.T) {
	expectedFiles := []string{
		"app.md",
		"cmd.md",
		"config.md",
	}

	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Documentation file does not exist: %s", file)
		}
	}
}
