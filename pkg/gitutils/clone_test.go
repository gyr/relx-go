package gitutils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
)

// Helper to create a dummy .git directory to simulate an existing repo
func createDummyGitRepo(t *testing.T, path string) {
	err := os.MkdirAll(filepath.Join(path, ".git"), 0755)
	if err != nil {
		t.Fatalf("Failed to create dummy .git repo at %s: %v", path, err)
	}
}

func TestCloneRepo(t *testing.T) {
	originalExecCommand := execCommand // Save original
	defer func() {
		execCommand = originalExecCommand // Restore original
	}()

	// Setup a temporary directory for the cache
	tempCacheDir := t.TempDir()

	mockCfg := &config.Config{
		RepoURL:  "https://example.com/test.git",
		RepoName: "test-repo",
		CacheDir: tempCacheDir,
		Logger:   logging.NewLogger(logging.LevelDebug),
	}
	expectedRepoPath := filepath.Join(tempCacheDir, mockCfg.RepoName)

	t.Run("InitialClone", func(t *testing.T) {
		// Mock execCommand for initial clone
		execCommand = func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) > 0 && args[0] == "clone" {
				if args[len(args)-1] != expectedRepoPath {
					t.Errorf("git clone path mismatch: got %s, want %s", args[len(args)-1], expectedRepoPath)
				}
				// Simulate successful clone by creating the directory and .git inside
				createDummyGitRepo(t, expectedRepoPath)
				return []byte("clone success"), nil
			}
			return nil, errors.New("unexpected exec command in initial clone: " + name + " " + strings.Join(args, " "))
		}

		path, err := CloneRepo(mockCfg)
		if err != nil {
			t.Fatalf("Initial clone failed: %v", err)
		}
		if path != expectedRepoPath {
			t.Errorf("Initial clone returned wrong path: got %s, want %s", path, expectedRepoPath)
		}
	})

	t.Run("UpdateExistingRepo", func(t *testing.T) {
		// Ensure the dummy repo exists from previous test run or create if not
		if _, err := os.Stat(filepath.Join(expectedRepoPath, ".git")); os.IsNotExist(err) {
			createDummyGitRepo(t, expectedRepoPath)
		}

		// Mock execCommand for update (fetch and pull)
		var fetchCalled, pullCalled bool
		execCommand = func(name string, args ...string) ([]byte, error) {
			if name == "git" {
				if len(args) > 0 && args[0] == "fetch" {
					fetchCalled = true
					return []byte("fetch success"), nil // Simulate successful fetch
				}
				if len(args) > 0 && args[0] == "pull" {
					pullCalled = true
					return []byte("pull success"), nil // Simulate successful pull
				}
			}
			return nil, errors.New("unexpected exec command in update: " + name + " " + strings.Join(args, " "))
		}

		path, err := CloneRepo(mockCfg)
		if err != nil {
			t.Fatalf("Update existing repo failed: %v", err)
		}
		if path != expectedRepoPath {
			t.Errorf("Update returned wrong path: got %s, want %s", path, expectedRepoPath)
		}
		if !fetchCalled {
			t.Error("git fetch was not called during update")
		}
		if !pullCalled {
			t.Error("git pull was not called during update")
		}
	})

	t.Run("CloneFailure", func(t *testing.T) {
		// Clean up the dummy repo if it exists, so clone attempts again
		if err := os.RemoveAll(expectedRepoPath); err != nil {
			t.Fatalf("Failed to remove dummy repo: %v", err)
		}

		mockError := errors.New("mock git clone failure")
		execCommand = func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) > 0 && args[0] == "clone" {
				return []byte("clone failed output"), mockError
			}
			return nil, errors.New("unexpected exec command in clone failure: " + name + " " + strings.Join(args, " "))
		}

		_, err := CloneRepo(mockCfg)
		if err == nil {
			t.Fatal("Expected error during clone failure, got nil")
		}
		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Error message missing expected substring %q: %v", mockError.Error(), err)
		}
	})

	t.Run("UpdateFetchFailure", func(t *testing.T) {
		// Ensure dummy repo exists to trigger update path
		createDummyGitRepo(t, expectedRepoPath)

		mockError := errors.New("mock git fetch failure")
		execCommand = func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) > 0 && args[0] == "fetch" {
				return []byte("fetch failed output"), mockError
			}
			if name == "git" && len(args) > 0 && args[0] == "pull" {
				// Should not be called if fetch fails
				t.Fatal("git pull was called after fetch failure")
			}
			return nil, errors.New("unexpected exec command in fetch failure: " + name + " " + strings.Join(args, " "))
		}

		_, err := CloneRepo(mockCfg)
		if err == nil {
			t.Fatal("Expected error during fetch failure, got nil")
		}
		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Error message missing expected substring %q: %v", mockError.Error(), err)
		}
	})

	t.Run("UpdatePullFailure", func(t *testing.T) {
		// Ensure dummy repo exists to trigger update path
		createDummyGitRepo(t, expectedRepoPath)

		mockError := errors.New("mock git pull failure")
		execCommand = func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) > 0 && args[0] == "fetch" {
				return []byte("fetch success"), nil
			}
			if name == "git" && len(args) > 0 && args[0] == "pull" {
				return []byte("pull failed output"), mockError
			}
			return nil, errors.New("unexpected exec command in pull failure: " + name + " " + strings.Join(args, " "))
		}

		_, err := CloneRepo(mockCfg)
		if err == nil {
			t.Fatal("Expected error during pull failure, got nil")
		}
		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Error message missing expected substring %q: %v", mockError.Error(), err)
		}
	})
}
