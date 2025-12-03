package gitutils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gyr/relx-go/pkg/command"
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
)

// MockRunner is a mock implementation of the command.Runner for testing.
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
	return nil, fmt.Errorf("RunFunc not defined for mock runner")
}

// Helper to create a dummy .git directory to simulate an existing repo
func createDummyGitRepo(t *testing.T, path string) {
	t.Helper()
	err := os.MkdirAll(filepath.Join(path, ".git"), 0755)
	if err != nil {
		t.Fatalf("Failed to create dummy .git repo at %s: %v", path, err)
	}
}

func TestManageRepo(t *testing.T) {
	// Setup a temporary directory for the cache
	tempCacheDir := t.TempDir()

	mockCfg := &config.Config{
		RepoURL:                 "https://example.com/test.git",
		RepoBranch:              "main",
		CacheDir:                tempCacheDir,
		Logger:                  logging.NewLogger(logging.LevelDebug),
		OperationTimeoutSeconds: 5, // A short timeout for tests
	}
	expectedRepoPath := filepath.Join(tempCacheDir, "test")

	t.Run("InitialClone", func(t *testing.T) {
		mockRunner := &MockRunner{}
		mockRunner.RunFunc = func(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
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
				createDummyGitRepo(t, expectedRepoPath)
				return []byte("clone success"), nil
			}
			return nil, fmt.Errorf("unexpected command in initial clone: %s %s", name, strings.Join(args, " "))
		}

		path, err := ManageRepo(context.Background(), mockCfg, mockRunner)
		if err != nil {
			t.Fatalf("Initial clone failed: %v", err)
		}
		if path != expectedRepoPath {
			t.Errorf("Initial clone returned wrong path: got %s, want %s", path, expectedRepoPath)
		}
	})

	t.Run("UpdateExistingRepo", func(t *testing.T) {
		createDummyGitRepo(t, expectedRepoPath)

		var switchCalled, fetchCalled, pullCalled bool
		mockRunner := &MockRunner{}
		mockRunner.RunFunc = func(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
			if dir != expectedRepoPath {
				t.Errorf("Expected dir %s, got %s", expectedRepoPath, dir)
			}
			if name == "git" {
				switch args[0] {
				case "switch":
					if args[1] != "main" {
						t.Errorf("git switch branch mismatch: got %s, want main", args[1])
					}
					switchCalled = true
					return []byte("switch success"), nil
				case "fetch":
					fetchCalled = true
					return []byte("fetch success"), nil
				case "pull":
					pullCalled = true
					return []byte("pull success"), nil
				}
			}
			return nil, fmt.Errorf("unexpected command in update: %s %s", name, strings.Join(args, " "))
		}

		path, err := ManageRepo(context.Background(), mockCfg, mockRunner)
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
		if err := os.RemoveAll(expectedRepoPath); err != nil {
			t.Fatalf("Failed to remove dummy repo: %v", err)
		}

		mockError := errors.New("mock git clone failure")
		mockRunner := &MockRunner{}
		mockRunner.RunFunc = func(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
			if name == "git" && args[0] == "clone" {
				return []byte("clone failed output"), mockError
			}
			return nil, fmt.Errorf("unexpected command in clone failure: %s %s", name, strings.Join(args, " "))
		}

		_, err := ManageRepo(context.Background(), mockCfg, mockRunner)
		if err == nil {
			t.Fatal("Expected error during clone failure, got nil")
		}
		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Error message missing expected substring %q: %v", mockError.Error(), err)
		}
	})

	t.Run("UpdateFetchFailure", func(t *testing.T) {
		createDummyGitRepo(t, expectedRepoPath)

		mockError := errors.New("mock git fetch failure")
		mockRunner := &MockRunner{}
		mockRunner.RunFunc = func(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
			if name == "git" && args[0] == "switch" {
				return []byte("switch success"), nil
			}
			if name == "git" && args[0] == "fetch" {
				return []byte("fetch failed output"), mockError
			}
			return nil, nil // Other commands succeed
		}

		_, err := ManageRepo(context.Background(), mockCfg, mockRunner)
		if err == nil {
			t.Fatal("Expected error during fetch failure, got nil")
		}
		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Error message missing expected substring %q: %v", mockError.Error(), err)
		}
	})

	t.Run("UpdatePullFailure", func(t *testing.T) {
		createDummyGitRepo(t, expectedRepoPath)

		mockError := errors.New("mock git pull failure")
		mockRunner := &MockRunner{}
		mockRunner.RunFunc = func(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
			if name == "git" && args[0] == "pull" {
				return []byte("pull failed output"), mockError
			}
			return []byte("success"), nil // Other commands succeed
		}

		_, err := ManageRepo(context.Background(), mockCfg, mockRunner)
		if err == nil {
			t.Fatal("Expected error during pull failure, got nil")
		}
		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Error message missing expected substring %q: %v", mockError.Error(), err)
		}
	})

	t.Run("RepoURLMissing", func(t *testing.T) {
		cfg := &config.Config{RepoBranch: "main", CacheDir: tempCacheDir}
		_, err := ManageRepo(context.Background(), cfg, &MockRunner{})
		if err == nil {
			t.Fatal("Expected an error for missing RepoURL, but got nil")
		}
		if !strings.Contains(err.Error(), "cannot be empty") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("RepoBranchMissing", func(t *testing.T) {
		cfg := &config.Config{RepoURL: "https://example.com/test.git", CacheDir: tempCacheDir}
		_, err := ManageRepo(context.Background(), cfg, &MockRunner{})
		if err == nil {
			t.Fatal("Expected an error for missing RepoBranch, but got nil")
		}
		if !strings.Contains(err.Error(), "cannot be empty") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("InvalidRepoURL", func(t *testing.T) {
		cfg := &config.Config{RepoURL: "https://", RepoBranch: "main", CacheDir: tempCacheDir}
		_, err := ManageRepo(context.Background(), cfg, &MockRunner{})
		if err == nil {
			t.Fatal("Expected an error for invalid RepoURL, but got nil")
		}
		if !strings.Contains(err.Error(), "could not derive a valid repository name") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}

func TestDeriveRepoName(t *testing.T) {
	testCases := []struct {
		name     string
		repoURL  string
		expected string
		err      bool
	}{
		{"Standard HTTPS", "https://example.com/user/repo.git", "repo", false},
		{"HTTPS without .git", "https://example.com/user/repo", "repo", false},
		{"Standard SSH", "git@example.com:user/repo.git", "repo", false},
		{"SSH without .git", "git@example.com:user/repo", "repo", false},
		{"Generic SSH", "gitea@example.com:user/another-repo.git", "another-repo", false},
		{"GitLab SSH with slash path", "gitlab@my-instance.com/user/repo.git", "repo", false},
		{"No name", "https://example.com/", "", true},
		{"Invalid URL", "https://", "", true},
		{"Empty URL", "", "", true},
		{"Just .git", ".git", "", true},
		{"Just a dot", ".", "", true},
		{"Colon only", ":", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoName, err := deriveRepoName(tc.repoURL)
			if (err != nil) != tc.err {
				t.Fatalf("Expected error: %v, got: %v", tc.err, err)
			}
			if repoName != tc.expected {
				t.Errorf("Expected repo name: %q, got: %q", tc.expected, repoName)
			}
		})
	}
}
