package gitutils

import (
	"context"
	"fmt"
	"time"

	"github.com/gyr/relx-go/pkg/command"
	"github.com/gyr/relx-go/pkg/config"
)

// FetchRemoteFile uses 'git archive' to fetch a single file from a remote repository
// without cloning the entire repository. This is much more efficient than a full clone.
func FetchRemoteFile(ctx context.Context, cfg *config.Config, runner command.Runner, filePath string) ([]byte, error) {
	// Create a context with the configured timeout.
	timeout := time.Duration(cfg.OperationTimeoutSeconds) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Construct the command. Note that we are using `bash -c` to handle the pipe.
	// The `git archive` command creates a tarball of the requested file, which we then
	// pipe to `tar -xO` to extract the raw file content to standard output.
	archiveCmd := fmt.Sprintf(
		"git archive --remote=%s %s %s | tar -xO",
		cfg.RepoURL,
		cfg.RepoBranch,
		filePath,
	)

	cfg.Logger.Infof("Fetching remote file: %s from %s (branch: %s)", filePath, cfg.RepoURL, cfg.RepoBranch)

	// Execute the command using the injected runner.
	output, err := runner.Run(timeoutCtx, "" /* workDir */, "bash", "-c", archiveCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote file '%s' with command '%s': %w. Output: %s", filePath, archiveCmd, err, string(output))
	}

	cfg.Logger.Debugf("Successfully fetched remote file '%s'.", filePath)

	return output, nil
}
