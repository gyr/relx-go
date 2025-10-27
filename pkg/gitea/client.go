package gitea

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"

	"github.com/gyr/grxs/pkg/core"
)

// execCommand is a function that runs a command and returns its output.
// It's a variable so it can be replaced by a mock in tests.
var execCommand = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// Client handles interaction with the Gitea API via the 'git-obs api' command.
type Client struct {
	CacheDir string
}

// NewClient creates a new Gitea client instance.
func NewClient(cacheDir string) *Client {
	return &Client{
		CacheDir: cacheDir,
	}
}

// GetPullRequests executes the git-obs command to fetch pull requests and unmarshals the JSON output.
func (c *Client) GetPullRequests(owner, repo string) ([]core.PullRequest, error) {
	apiPath, err := buildGiteaURL(owner)
	if err != nil {
		return nil, fmt.Errorf("gitea: failed to build API URL: %w", err)
	}

	output, err := execCommand("git-obs", "api", apiPath)
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
