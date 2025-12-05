package gitea

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gyr/relx-go/pkg/command"
	"github.com/gyr/relx-go/pkg/config"
)

// Client handles interaction with the Gitea API via the 'git-obs api' command.
type Client struct {
	runner command.Runner
	cfg    *config.Config
}

// NewClient creates a new Gitea client instance.
func NewClient(runner command.Runner, cfg *config.Config) *Client {
	return &Client{
		runner: runner,
		cfg:    cfg,
	}
}

// ShowPullRequest executes the `git obs pr show` command and pipes its output to `delta` to display the content and diff of a pull request.
func (c *Client) ShowPullRequest(ctx context.Context, repository, prID string) error {
	timeout := time.Duration(c.cfg.OperationTimeoutSeconds) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	gitObsCmd := []string{
		"git-obs",
		"pr",
		"show",
		"--timeline",
		"--patch",
		fmt.Sprintf("%s#%s", repository, prID),
	}
	deltaCmd := []string{"delta"}

	return c.runner.RunPipeline(timeoutCtx, "" /* workDir */, gitObsCmd, deltaCmd)
}

// GetOpenPullRequests executes the `git obs pr list` command to get the list of open pull requests.
func (c *Client) GetOpenPullRequests(ctx context.Context, prReviewer, branch, repository string) ([]string, error) {
	timeout := time.Duration(c.cfg.OperationTimeoutSeconds) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := []string{
		"pr",
		"list",
		"--state", "open",
		"--reviewer", prReviewer,
		"--review-state", "REQUEST_REVIEW",
		"--no-draft",
		"--target-branch", branch,
		repository,
	}

	output, err := c.runner.Run(timeoutCtx, "" /* workDir */, "git-obs", args...)
	if err != nil {
		return nil, fmt.Errorf("gitea: 'git-obs pr list' failed: %w", err)
	}

	var prIDs []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID") {
			parts := strings.Split(line, "#")
			if len(parts) == 2 {
				prIDs = append(prIDs, strings.TrimSpace(parts[1]))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("gitea: failed to read command output: %w", err)
	}

	return prIDs, nil
}
