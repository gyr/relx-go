package gitutils

import (
	"context" // Import context for timeout management
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time" // Import time for duration

	"github.com/gyr/relx-go/pkg/cache"
	"github.com/gyr/relx-go/pkg/command" // Import the new command package
	"github.com/gyr/relx-go/pkg/config"
)

// deriveRepoName extracts the repository name from a URL.
func deriveRepoName(repoURL string) (string, error) {
	var path string

	if strings.HasPrefix(repoURL, "http://") || strings.HasPrefix(repoURL, "https://") {
		u, err := url.Parse(repoURL)
		if err != nil {
			return "", fmt.Errorf("gitutils: failed to parse URL %s: %w", repoURL, err)
		}
		path = u.Path
	} else {
		// Assume scp-like syntax, e.g., git@example.com:user/repo.git
		path = repoURL
		if atIndex := strings.LastIndex(path, "@"); atIndex != -1 {
			path = path[atIndex+1:]
		}
		if colonIndex := strings.Index(path, ":"); colonIndex != -1 {
			path = path[colonIndex+1:]
		}
	}

	// Get the base name (last component) of the URL path
	base := filepath.Base(path)
	// Remove the ".git" suffix if it exists
	repoName := strings.TrimSuffix(base, ".git")

	if repoName == "" || repoName == "." || repoName == "/" {
		return "", fmt.Errorf("gitutils: could not derive a valid repository name from URL: %s", repoURL)
	}
	return repoName, nil
}

// CloneRepo clones a Git repository into a specified cache directory.
// It skips submodules and returns the local path to the cloned repository.
// ManageRepo manages a Git repository by cloning or updating it.
//
// It accepts a parent context to enable cancellation of the entire operation from the caller.
// For the git commands it executes, it creates a new derived context with a timeout
// based on the `OperationTimeoutSeconds` in the configuration. This ensures that the git
// operations don't hang indefinitely, while still respecting cancellation from the parent context.
//
// This function relies on a command.Runner for executing external commands, which allows
// for mocking during tests.
func ManageRepo(ctx context.Context, cfg *config.Config, runner command.Runner) (string, error) {
	if cfg.RepoURL == "" {
		return "", fmt.Errorf("gitutils: repository URL (RepoURL) cannot be empty in the configuration")
	}
	if cfg.RepoBranch == "" {
		return "", fmt.Errorf("gitutils: repository branch (RepoBranch) cannot be empty in the configuration")
	}

	// Create a context with a timeout for the git operations.
	// This context is derived from the parent `ctx`, so if the parent context
	// is cancelled, this derived context will be cancelled as well.
	timeout := time.Duration(cfg.OperationTimeoutSeconds) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel() // This is crucial to release resources associated with the context.

	cache, err := cache.New(cfg.CacheDir)
	if err != nil {
		return "", err
	}

	repoName, err := deriveRepoName(cfg.RepoURL)
	if err != nil {
		return "", err
	}
	localPath := cache.GetPath(repoName)

	var output []byte
	var cmdErr error // Use a distinct error variable for git commands

	// Check if the repository already exists
	if cache.Has(repoName, ".git") {
		// Repository exists, update it
		cfg.Logger.Infof("Repository %s already exists at %s. Updating...", cfg.RepoURL, localPath)

		// Ensure we are on the correct branch before pulling.
		// We use the new runner, passing the timeout-enabled context.
		output, cmdErr = runner.Run(timeoutCtx, localPath, "git", "switch", cfg.RepoBranch)
		if cmdErr != nil {
			return "", fmt.Errorf("gitutils: git switch to branch %s failed for %s. Output:\n%s\nError: %w", cfg.RepoBranch, cfg.RepoURL, string(output), cmdErr)
		}
		cfg.Logger.Debugf("Git switch output:\n%s", string(output))

		// Fetch changes from the remote.
		output, cmdErr = runner.Run(timeoutCtx, localPath, "git", "fetch", "--prune", "--all")
		if cmdErr != nil {
			return "", fmt.Errorf("gitutils: git fetch failed for %s. Output:\n%s\nError: %w", cfg.RepoURL, string(output), cmdErr)
		}
		cfg.Logger.Debugf("Git fetch output:\n%s", string(output))

		// Pull with rebase to keep a clean history.
		output, cmdErr = runner.Run(timeoutCtx, localPath, "git", "pull", "--rebase")
		if cmdErr != nil {
			return "", fmt.Errorf("gitutils: git pull --rebase failed for %s. Output:\n%s\nError: %w", cfg.RepoURL, string(output), cmdErr)
		}
		cfg.Logger.Debugf("Git pull --rebase output:\n%s", string(output))

	} else { // Repository does not exist, clone it
		cfg.Logger.Infof("Cloning repository %s (branch %s) to %s...", cfg.RepoURL, cfg.RepoBranch, localPath)
		output, cmdErr = runner.Run(timeoutCtx, "", "git", "clone", "--branch", cfg.RepoBranch, "--recurse-submodules=no", cfg.RepoURL, localPath)
		if cmdErr != nil {
			return "", fmt.Errorf("gitutils: git clone failed for %s. Output:\n%s\nError: %w", cfg.RepoURL, string(output), cmdErr)
		}
		cfg.Logger.Debugf("Git clone output:\n%s", string(output))
	}

	return localPath, nil
}
