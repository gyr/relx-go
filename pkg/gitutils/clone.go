package gitutils

import (
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gyr/relx-go/pkg/cache"
	"github.com/gyr/relx-go/pkg/config"
)

var execCommand = func(dir, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	return cmd.CombinedOutput()
}

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
func CloneRepo(cfg *config.Config) (string, error) {
	if cfg.RepoURL == "" {
		return "", fmt.Errorf("gitutils: repository URL (RepoURL) cannot be empty in the configuration")
	}
	if cfg.RepoBranch == "" {
		return "", fmt.Errorf("gitutils: repository branch (RepoBranch) cannot be empty in the configuration")
	}

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
	var cloneErr error // Use a distinct error variable for git commands

	// Check if the repository already exists
	if cache.Has(repoName, ".git") {
		// Repository exists, update it
		cfg.Logger.Infof("Repository %s already exists at %s. Updating...", cfg.RepoURL, localPath)

		// Change to the repository directory
		// Ensure we are on the correct branch before pulling
		output, cloneErr = execCommand(localPath, "git", "switch", cfg.RepoBranch)
		if cloneErr != nil {
			return "", fmt.Errorf("gitutils: git switch to branch %s failed for %s. Output:\n%s\nError: %w", cfg.RepoBranch, cfg.RepoURL, string(output), cloneErr)
		}
		cfg.Logger.Debugf("Git switch output:\n%s", string(output))

		// Change to the repository directory and fetch
		output, cloneErr = execCommand(localPath, "git", "fetch", "--prune", "--all")
		if cloneErr != nil {
			return "", fmt.Errorf("gitutils: git fetch failed for %s. Output:\n%s\nError: %w", cfg.RepoURL, string(output), cloneErr)
		}
		cfg.Logger.Debugf("Git fetch output:\n%s", string(output))

		// Pull with rebase
		output, cloneErr = execCommand(localPath, "git", "pull", "--rebase")
		if cloneErr != nil {
			return "", fmt.Errorf("gitutils: git pull --rebase failed for %s. Output:\n%s\nError: %w", cfg.RepoURL, string(output), cloneErr)
		}
		cfg.Logger.Debugf("Git pull --rebase output:\n%s", string(output))

	} else { // Repository does not exist, clone it
		cfg.Logger.Infof("Cloning repository %s (branch %s) to %s...", cfg.RepoURL, cfg.RepoBranch, localPath)
		output, cloneErr = execCommand("", "git", "clone", "--branch", cfg.RepoBranch, "--recurse-submodules=no", cfg.RepoURL, localPath)
		if cloneErr != nil {
			return "", fmt.Errorf("gitutils: git clone failed for %s. Output:\n%s\nError: %w", cfg.RepoURL, string(output), cloneErr)
		}
		cfg.Logger.Debugf("Git clone output:\n%s", string(output))
	}

	return localPath, nil
}
