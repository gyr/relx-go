package gitea

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gyr/relx-go/pkg/command"
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/core"
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

// GetPullRequests executes the git-obs command to fetch pull requests and unmarshals the JSON output.
func (c *Client) GetPullRequests(ctx context.Context, owner string) ([]core.PullRequest, error) {
	apiPath, err := buildGiteaURL(owner)
	if err != nil {
		return nil, fmt.Errorf("gitea: failed to build API URL: %w", err)
	}

	timeout := time.Duration(c.cfg.OperationTimeoutSeconds) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	output, err := c.runner.Run(timeoutCtx, "" /* workDir */, "git-obs", "api", apiPath)
	if err != nil {
		return nil, fmt.Errorf("gitea: 'git-obs api' failed: %w", err)
	}

	var response struct {
		Issues []core.PullRequest `json:"issues"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("gitea: failed to parse JSON output: %w\nOutput received: %s", err, string(output))
	}

	return response.Issues, nil
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

func buildGiteaURL(owner string) (string, error) {
	u, err := url.Parse("/repos/issues/search")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("type", "pulls")
	q.Set("owner", owner)
	q.Set("state", "open")
	q.Set("limit", "50")
	u.RawQuery = q.Encode()
	return u.String(), nil
}
