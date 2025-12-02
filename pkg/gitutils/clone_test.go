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
		RepoURL:    "https://example.com/test.git",
		RepoBranch: "main",
		CacheDir:   tempCacheDir,
		Logger:     logging.NewLogger(logging.LevelDebug),
	}
	expectedRepoPath := filepath.Join(tempCacheDir, "test")

	t.Run("InitialClone", func(t *testing.T) {
		// Mock execCommand for initial clone
		execCommand = func(dir, name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) > 0 && args[0] == "clone" {
				if dir != "" {
					t.Errorf("Expected empty dir for clone, got %s", dir)
				}
				if args[1] != "--branch" || args[2] != "main" {
					t.Errorf("git clone branch mismatch: got args %v", args)
				}
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

		// Mock execCommand for update (switch, fetch and pull)
		var switchCalled, fetchCalled, pullCalled bool
		execCommand = func(dir, name string, args ...string) ([]byte, error) {
			if dir != expectedRepoPath {
				t.Errorf("Expected dir %s, got %s", expectedRepoPath, dir)
			}
			if name == "git" {
				if len(args) > 0 && args[0] == "switch" {
					if args[1] != "main" {
						t.Errorf("git switch branch mismatch: got %s, want main", args[1])
					}
					switchCalled = true
					return []byte("switch success"), nil // Simulate successful switch
				}
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
		if !switchCalled {
			t.Error("git switch was not called during update")
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
		execCommand = func(dir, name string, args ...string) ([]byte, error) {
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
		execCommand = func(dir, name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) > 0 && args[0] == "switch" {
				return []byte("switch success"), nil // Simulate successful switch
			}
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
		execCommand = func(dir, name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) > 0 && args[0] == "switch" {
				return []byte("switch success"), nil // Simulate successful switch
			}
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

	t.Run("RepoURLMissing", func(t *testing.T) {
		cfg := &config.Config{
			RepoURL:    "", // Empty RepoURL
			RepoBranch: "main",
			CacheDir:   tempCacheDir,
			Logger:     logging.NewLogger(logging.LevelDebug),
		}

		_, err := CloneRepo(cfg)
		if err == nil {
			t.Fatal("Expected an error for missing RepoURL, but got nil")
		}
		expectedErrorMsg := "gitutils: repository URL (RepoURL) cannot be empty in the configuration"
		if !strings.Contains(err.Error(), expectedErrorMsg) {
			t.Errorf("Error message missing expected substring %q: %v", expectedErrorMsg, err)
		}
	})

	t.Run("RepoBranchMissing", func(t *testing.T) {
		cfg := &config.Config{
			RepoURL:    "https://example.com/test.git",
			RepoBranch: "", // Empty RepoBranch
			CacheDir:   tempCacheDir,
			Logger:     logging.NewLogger(logging.LevelDebug),
		}

		_, err := CloneRepo(cfg)
		if err == nil {
			t.Fatal("Expected an error for missing RepoBranch, but got nil")
		}
		expectedErrorMsg := "gitutils: repository branch (RepoBranch) cannot be empty in the configuration"
		if !strings.Contains(err.Error(), expectedErrorMsg) {
			t.Errorf("Error message missing expected substring %q: %v", expectedErrorMsg, err)
		}
	})

	t.Run("InvalidRepoURL", func(t *testing.T) {
		cfg := &config.Config{
			RepoURL:    "https://", // Invalid URL that results in empty repo name
			RepoBranch: "main",
			CacheDir:   tempCacheDir,
			Logger:     logging.NewLogger(logging.LevelDebug),
		}

		_, err := CloneRepo(cfg)
		if err == nil {
			t.Fatal("Expected an error for invalid RepoURL, but got nil")
		}
		expectedErrorMsg := "gitutils: could not derive a valid repository name from URL"
		if !strings.Contains(err.Error(), expectedErrorMsg) {
			t.Errorf("Error message missing expected substring %q: %v", expectedErrorMsg, err)
		}
	})
}
