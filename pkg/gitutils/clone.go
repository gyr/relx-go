package gitutils

import (
	"fmt"

	"os/exec"

	"github.com/gyr/relx-go/pkg/cache"
	"github.com/gyr/relx-go/pkg/config"
)

var execCommand = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// CloneRepo clones a Git repository into a specified cache directory.
// It skips submodules and returns the local path to the cloned repository.
func CloneRepo(cfg *config.Config) (string, error) {

	cache, err := cache.New(cfg.CacheDir)
	if err != nil {
		return "", err
	}

	localPath := cache.GetPath(cfg.RepoName)

	// Check if the repository already exists
	if cache.Has(cfg.RepoName, ".git") {
		// Repository exists, update it
		cfg.Logger.Infof("Repository %s already exists at %s. Updating...", cfg.RepoURL, localPath)

		// Change to the repository directory and fetch
		output, err := execCommand("git", "fetch", "--prune", "--all")
		if err != nil {
			return "", fmt.Errorf("gitutils: git fetch failed for %s. Output:\n%s\nError: %w", cfg.RepoURL, string(output), err)
		}
		cfg.Logger.Debugf("Git fetch output:\n%s", string(output))

		// Pull with rebase
		output, err = execCommand("git", "pull", "--rebase")
		if err != nil {
			return "", fmt.Errorf("gitutils: git pull --rebase failed for %s. Output:\n%s\nError: %w", cfg.RepoURL, string(output), err)
		}
		cfg.Logger.Debugf("Git pull --rebase output:\n%s", string(output))

	} else { // Repository does not exist, clone it
		cfg.Logger.Infof("Cloning repository %s to %s...", cfg.RepoURL, localPath)
		output, err := execCommand("git", "clone", "--recurse-submodules=no", cfg.RepoURL, localPath)
		if err != nil {
			return "", fmt.Errorf("gitutils: git clone failed for %s. Output:\n%s\nError: %w", cfg.RepoURL, string(output), err)
		}
		cfg.Logger.Debugf("Git clone output:\n%s", string(output))
	}

	return localPath, nil
}
