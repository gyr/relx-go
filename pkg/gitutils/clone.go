package gitutils

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	"github.com/gyr/grxs/pkg/config"
)

var userCurrent = user.Current
var osMkdirAll = os.MkdirAll
var execCommand = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// CloneRepo clones a Git repository into a specified cache directory.
// It skips submodules and returns the local path to the cloned repository.
func CloneRepo(cfg *config.Config) (string, error) {
	// If cacheDir is empty, use the default behavior
	if cfg.CacheDir == "" {
		currentUser, err := userCurrent()
		if err != nil {
			return "", fmt.Errorf("gitutils: could not get current user: %w", err)
		}
		cfg.CacheDir = filepath.Join(currentUser.HomeDir, ".cache", "grxs")
	}

	if err := osMkdirAll(cfg.CacheDir, 0755); err != nil {
		return "", fmt.Errorf("gitutils: error creating cache directory %s: %w", cfg.CacheDir, err)
	}

	localPath := filepath.Join(cfg.CacheDir, cfg.RepoName)

	// Check if the repository already exists
	if _, err := os.Stat(localPath); err == nil {
		// Repository exists, update it
		cfg.Logger.Infof("Repository %s already exists at %s. Updating...", cfg.RepoURL, localPath)

		// Change to the repository directory and fetch
		cmd := exec.Command("git", "fetch", "--prune", "--all")
		cmd.Dir = localPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("gitutils: git fetch failed for %s. Output:\n%s\nError: %w", cfg.RepoURL, string(output), err)
		}
		cfg.Logger.Debugf("Git fetch output:\n%s", string(output))

		// Pull with rebase
		cmd = exec.Command("git", "pull", "--rebase")
		cmd.Dir = localPath
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("gitutils: git pull --rebase failed for %s. Output:\n%s\nError: %w", cfg.RepoURL, string(output), err)
		}
		cfg.Logger.Debugf("Git pull --rebase output:\n%s", string(output))

	} else if os.IsNotExist(err) {
		// Repository does not exist, clone it
		cfg.Logger.Infof("Cloning repository %s to %s...", cfg.RepoURL, localPath)
		output, err := execCommand("git", "clone", "--recurse-submodules=no", cfg.RepoURL, localPath)
		if err != nil {
			return "", fmt.Errorf("gitutils: git clone failed for %s. Output:\n%s\nError: %w", cfg.RepoURL, string(output), err)
		}
		cfg.Logger.Debugf("Git clone output:\n%s", string(output))
	} else {
		// Other error checking localPath
		return "", fmt.Errorf("gitutils: error checking repository existence at %s: %w", localPath, err)
	}

	return localPath, nil
}
