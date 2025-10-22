package gitutils

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

var userCurrent = user.Current
var osMkdirAll = os.MkdirAll
var execCommand = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// CloneRepo clones a Git repository into a specified cache directory.
// It skips submodules and returns the local path to the cloned repository.
func CloneRepo(repoURL string, cacheDir string) (string, error) {
	// If cacheDir is empty, use the default behavior
	if cacheDir == "" {
		currentUser, err := userCurrent()
		if err != nil {
			return "", fmt.Errorf("gitutils: could not get current user: %w", err)
		}
		cacheDir = filepath.Join(currentUser.HomeDir, ".cache", "grxs")
	}

	if err := osMkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("gitutils: error creating cache directory %s: %w", cacheDir, err)
	}

	repoName := filepath.Base(repoURL)
	localPath := filepath.Join(cacheDir, repoName)

	output, err := execCommand("git", "clone", "--recurse-submodules=no", repoURL, localPath)
	if err != nil {
		return "", fmt.Errorf("gitutils: git clone failed for %s. Output:\n%s\nError: %w", repoURL, string(output), err)
	}

	return localPath, nil
}
