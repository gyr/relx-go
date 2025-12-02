package app

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
)

func TestHandleBugownerByPackage(t *testing.T) {
	// Save and restore original ManageRepo
	originalManageRepo := ManageRepo
	defer func() {
		ManageRepo = originalManageRepo
	}()

	tempDir := t.TempDir()

	// Setup mock maintainership file
	maintainershipContent := `{"pkg1": ["userA", "userB"], "pkg2": ["userC"]}`
	err := os.WriteFile(filepath.Join(tempDir, "_maintainership.json"), []byte(maintainershipContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock maintainership file: %v", err)
	}

	t.Run("PackageFound", func(t *testing.T) {
		var out bytes.Buffer // Create a buffer to capture output
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out, // Inject the buffer as the output writer
		}

		ManageRepo = func(cfg *config.Config) (string, error) {
			return tempDir, nil
		}

		err := HandleBugownerByPackage(cfg, "pkg1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		output := out.String() // Get captured output
		if !strings.Contains(output, "Maintainers for package pkg1:") {
			t.Errorf("Output missing expected header. Got: %s", output)
		}
		if !strings.Contains(output, "userA") || !strings.Contains(output, "userB") {
			t.Errorf("Output missing expected maintainers. Got: %s", output)
		}
	})

	t.Run("PackageNotFound", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
		}

		ManageRepo = func(cfg *config.Config) (string, error) {
			return tempDir, nil
		}

		err := HandleBugownerByPackage(cfg, "nonexistent")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		output := out.String()
		expectedMsg := "Package 'nonexistent' not found"
		if !strings.Contains(output, expectedMsg) {
			t.Errorf("Expected message %q, but it was not found in output: %s", expectedMsg, output)
		}
	})

	t.Run("ManageRepoFailure", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out, // Still inject buffer, though no output expected
		}

		mockError := errors.New("git failed")
		ManageRepo = func(cfg *config.Config) (string, error) {
			return "", mockError
		}

		err := HandleBugownerByPackage(cfg, "pkg1")
		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Error message missing expected substring %q: %v", mockError.Error(), err)
		}
	})
}

func TestHandlePackagesByMaintainer(t *testing.T) {
	// Save and restore original ManageRepo
	originalManageRepo := ManageRepo
	defer func() {
		ManageRepo = originalManageRepo
	}()

	tempDir := t.TempDir()

	// Setup mock maintainership file
	maintainershipContent := `{"pkg1": ["userA", "userB"], "pkg2": ["userC"], "pkg3": ["userA"]}`
	err := os.WriteFile(filepath.Join(tempDir, "_maintainership.json"), []byte(maintainershipContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock maintainership file: %v", err)
	}

	t.Run("MaintainerFound", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
		}

		ManageRepo = func(cfg *config.Config) (string, error) {
			return tempDir, nil
		}

		err := HandlePackagesByMaintainer(cfg, "userA")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		output := out.String()
		if !strings.Contains(output, "Packages maintained by userA:") {
			t.Errorf("Output missing expected header. Got: %s", output)
		}
		if !strings.Contains(output, "pkg1") || !strings.Contains(output, "pkg3") {
			t.Errorf("Output missing expected packages. Got: %s", output)
		}
	})

	t.Run("MaintainerNotFound", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
		}

		ManageRepo = func(cfg *config.Config) (string, error) {
			return tempDir, nil
		}

		err := HandlePackagesByMaintainer(cfg, "nonexistent")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		output := out.String()
		expectedMsg := "No packages found for maintainer 'nonexistent'"
		if !strings.Contains(output, expectedMsg) {
			t.Errorf("Expected message %q, but it was not found in output: %s", expectedMsg, output)
		}
	})

	t.Run("MaintainershipFileMissing", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
		}

		ManageRepo = func(cfg *config.Config) (string, error) {
			// Return a path where _maintainership.json doesn't exist
			return t.TempDir(), nil
		}

		err := HandlePackagesByMaintainer(cfg, "userA")
		if err == nil {
			t.Fatal("Expected an error for missing maintainership file, got nil")
		}
		expectedErrSubstring := "error reading maintainership file"
		if !strings.Contains(err.Error(), expectedErrSubstring) {
			t.Errorf("Expected error message to contain %q, but got %v", expectedErrSubstring, err)
		}
	})

	t.Run("MaintainershipFileMalformed", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
		}
		malformedTempDir := t.TempDir()
		err := os.WriteFile(filepath.Join(malformedTempDir, "_maintainership.json"), []byte("this is not json"), 0644)
		if err != nil {
			t.Fatalf("Failed to create malformed maintainership file: %v", err)
		}

		ManageRepo = func(cfg *config.Config) (string, error) {
			return malformedTempDir, nil
		}

		err = HandlePackagesByMaintainer(cfg, "userA")
		if err == nil {
			t.Fatal("Expected an error for malformed maintainership file, got nil")
		}
		expectedErrSubstring := "error unmarshaling maintainership JSON"
		if !strings.Contains(err.Error(), expectedErrSubstring) {
			t.Errorf("Expected error message to contain %q, but got %v", expectedErrSubstring, err)
		}
	})
}
