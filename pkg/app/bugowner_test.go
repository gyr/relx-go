package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gyr/relx-go/pkg/command"
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
)

// MockRunner is a mock implementation of the command.Runner for testing.
// It allows us to simulate the success or failure of external commands.
type MockRunner struct {
	RunFunc func(ctx context.Context, workDir, name string, args ...string) ([]byte, error)
}

// This line is a compile-time check to ensure MockRunner implements command.Runner.
var _ command.Runner = (*MockRunner)(nil)

// Run executes the mock command.
func (m *MockRunner) Run(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, workDir, name, args...)
	}
	// Default behavior: success with no output
	return nil, nil
}

// setupTestRepo creates a mock repository structure in a temporary directory
// and writes a _maintainership.json file with the given content.
func setupTestRepo(t *testing.T, cacheDir, repoName, content string) {
	t.Helper()
	repoPath := filepath.Join(cacheDir, repoName)
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		t.Fatalf("Failed to create mock repo dir: %v", err)
	}
	// Also create a .git dir to satisfy the `cache.Has` check in ManageRepo
	if err := os.MkdirAll(filepath.Join(repoPath, ".git"), 0755); err != nil {
		t.Fatalf("Failed to create mock .git dir: %v", err)
	}
	err := os.WriteFile(filepath.Join(repoPath, "_maintainership.json"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock maintainership file: %v", err)
	}
}

func TestHandleBugownerByPackage(t *testing.T) {
	const repoURL = "https://example.com/test/repo.git" // Gives repo name "repo"
	const repoName = "repo"
	maintainershipContent := `{"pkg1": ["userA", "userB"], "pkg2": ["userC"]}`

	t.Run("PackageFound", func(t *testing.T) {
		tempDir := t.TempDir()
		setupTestRepo(t, tempDir, repoName, maintainershipContent)

		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			CacheDir:     tempDir,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}
		mockRunner := &MockRunner{} // Default mock simulates success

		err := HandleBugownerByPackage(context.Background(), cfg, mockRunner, "pkg1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		output := out.String()
		if !strings.Contains(output, "Maintainers for package pkg1:") {
			t.Errorf("Output missing expected header. Got: %s", output)
		}
		if !strings.Contains(output, "userA") || !strings.Contains(output, "userB") {
			t.Errorf("Output missing expected maintainers. Got: %s", output)
		}
	})

	t.Run("PackageNotFound", func(t *testing.T) {
		tempDir := t.TempDir()
		setupTestRepo(t, tempDir, repoName, maintainershipContent)

		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			CacheDir:     tempDir,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}
		mockRunner := &MockRunner{}

		err := HandleBugownerByPackage(context.Background(), cfg, mockRunner, "nonexistent")
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
		tempDir := t.TempDir()
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			CacheDir:     tempDir,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}

		mockError := errors.New("git failed")
		mockRunner := &MockRunner{
			RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
				return nil, mockError
			},
		}

		err := HandleBugownerByPackage(context.Background(), cfg, mockRunner, "pkg1")
		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		// The error is now wrapped, so we check for our specific error message
		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Error message missing expected substring %q: %v", mockError.Error(), err)
		}
	})
}

func TestHandlePackagesByMaintainer(t *testing.T) {
	const repoURL = "https://example.com/test/repo.git"
	const repoName = "repo"
	maintainershipContent := `{"pkg1": ["userA", "userB"], "pkg2": ["userC"], "pkg3": ["userA"]}`

	t.Run("MaintainerFound", func(t *testing.T) {
		tempDir := t.TempDir()
		setupTestRepo(t, tempDir, repoName, maintainershipContent)

		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			CacheDir:     tempDir,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}
		mockRunner := &MockRunner{}

		err := HandlePackagesByMaintainer(context.Background(), cfg, mockRunner, "userA")
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
		tempDir := t.TempDir()
		setupTestRepo(t, tempDir, repoName, maintainershipContent)

		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			CacheDir:     tempDir,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}
		mockRunner := &MockRunner{}

		err := HandlePackagesByMaintainer(context.Background(), cfg, mockRunner, "nonexistent")
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
		tempDir := t.TempDir() // An empty temp dir
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			CacheDir:     tempDir,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}
		// We still need to create the repo dir for ManageRepo to succeed
		if err := os.MkdirAll(filepath.Join(tempDir, repoName), 0755); err != nil {
			t.Fatalf("Failed to create mock repo dir: %v", err)
		}

		mockRunner := &MockRunner{}

		err := HandlePackagesByMaintainer(context.Background(), cfg, mockRunner, "userA")
		if err == nil {
			t.Fatal("Expected an error for missing maintainership file, got nil")
		}
		// This error comes from `loadMaintainershipData`, which is what we want to test
		expectedErrSubstring := "error reading maintainership file"
		if !strings.Contains(err.Error(), expectedErrSubstring) {
			t.Errorf("Expected error message to contain %q, but got %v", expectedErrSubstring, err)
		}
	})

	t.Run("MaintainershipFileMalformed", func(t *testing.T) {
		tempDir := t.TempDir()
		setupTestRepo(t, tempDir, repoName, "this is not json")

		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			CacheDir:     tempDir,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}
		mockRunner := &MockRunner{}

		err := HandlePackagesByMaintainer(context.Background(), cfg, mockRunner, "userA")
		if err == nil {
			t.Fatal("Expected an error for malformed maintainership file, got nil")
		}
		expectedErrSubstring := "error unmarshaling maintainership JSON"
		if !strings.Contains(err.Error(), expectedErrSubstring) {
			t.Errorf("Expected error message to contain %q, but got %v", expectedErrSubstring, err)
		}
	})
}
