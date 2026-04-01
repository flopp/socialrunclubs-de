package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	// Test config JSON
	configJSON := `{
		"IsRemoteTarget": true,
		"OutputDir": "output",
		"CacheDir": "cache",
		"ImageDir": "images",
		"Google": {
			"APIKey": "test-api-key",
			"SheetId": "test-sheet-id",
			"SubmitUrl": "https://example.com/submit",
			"ReportUrl": "https://example.com/report"
		},
		"AHrefs": {
			"IndexNow": "test-indexnow-key"
		},
		"Umami": {
			"WebsiteId": "test-website-id"
		}
	}`

	err := os.WriteFile(configFile, []byte(configJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading the config
	var config Config
	err = LoadConfig(configFile, &config)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify the config was loaded correctly
	if !config.IsRemoteTarget {
		t.Error("Expected IsRemoteTarget to be true")
	}

	if config.CacheDir != "cache" {
		t.Errorf("Expected CacheDir to be 'cache', got '%s'", config.CacheDir)
	}

	if config.ImageDir != "images" {
		t.Errorf("Expected ImageDir to be 'images', got '%s'", config.ImageDir)
	}

	if config.Google.APIKey != "test-api-key" {
		t.Errorf("Expected Google.APIKey to be 'test-api-key', got '%s'", config.Google.APIKey)
	}

	if config.Google.SheetId != "test-sheet-id" {
		t.Errorf("Expected Google.SheetId to be 'test-sheet-id', got '%s'", config.Google.SheetId)
	}

	if config.Google.SubmitUrl != "https://example.com/submit" {
		t.Errorf("Expected Google.SubmitUrl to be 'https://example.com/submit', got '%s'", config.Google.SubmitUrl)
	}

	if config.Google.ReportUrl != "https://example.com/report" {
		t.Errorf("Expected Google.ReportUrl to be 'https://example.com/report', got '%s'", config.Google.ReportUrl)
	}

	if config.AHrefs.IndexNow != "test-indexnow-key" {
		t.Errorf("Expected AHrefs.IndexNow to be 'test-indexnow-key', got '%s'", config.AHrefs.IndexNow)
	}

	if config.Umami.WebsiteId != "test-website-id" {
		t.Errorf("Expected Umami.WebsiteId to be 'test-website-id', got '%s'", config.Umami.WebsiteId)
	}

	// Test that OutputDir was converted to absolute path
	if !filepath.IsAbs(config.OutputDir) {
		t.Errorf("Expected OutputDir to be absolute path, got '%s'", config.OutputDir)
	}

	if !strings.HasSuffix(config.OutputDir, "output") {
		t.Errorf("Expected OutputDir to end with 'output', got '%s'", config.OutputDir)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid_config.json")

	invalidJSON := `{
		"IsRemoteTarget": true,
		"OutputDir": "output",
		"CacheDir": "cache",
		"ImageDir": "images"
		// Missing closing brace
	`

	err := os.WriteFile(configFile, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading the invalid config
	var config Config
	err = LoadConfig(configFile, &config)
	if err == nil {
		t.Error("Expected LoadConfig to fail with invalid JSON")
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	// Test loading a non-existent file
	var config Config
	err := LoadConfig("non_existent_file.json", &config)
	if err == nil {
		t.Error("Expected LoadConfig to fail with non-existent file")
	}
}

func TestLoadConfig_RelativeOutputDir(t *testing.T) {
	// Create a temporary config file with relative OutputDir
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.json")

	configJSON := `{
		"IsRemoteTarget": false,
		"OutputDir": "relative/output/path",
		"CacheDir": "cache",
		"ImageDir": "images"
	}`

	err := os.WriteFile(configFile, []byte(configJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading the config
	var config Config
	err = LoadConfig(configFile, &config)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify OutputDir was converted to absolute path
	if !filepath.IsAbs(config.OutputDir) {
		t.Errorf("Expected OutputDir to be absolute path, got '%s'", config.OutputDir)
	}

	// Verify it ends with the relative path we specified
	if !strings.HasSuffix(config.OutputDir, "relative/output/path") {
		t.Errorf("Expected OutputDir to end with 'relative/output/path', got '%s'", config.OutputDir)
	}
}
