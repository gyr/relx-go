package obs

import (
	"encoding/xml"
	"fmt"
	"os/exec"

	"github.com/gyr/grxs/pkg/core" // Import core types
)

// execCommand is a function that runs a command and returns its output.
// It's a variable so it can be replaced by a mock in tests.
var execCommand = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// Client handles interaction with the OBS API via the 'osc api' command.
type Client struct{}

// NewClient creates a new OBS client instance.
func NewClient() *Client {
	return &Client{}
}

// GetBuildStatus executes the osc command to check a build status and unmarshals the XML output.
func (o *Client) GetBuildStatus(project, pkg string) ([]core.BuildStatus, error) {
	apiPath := fmt.Sprintf("/build/%s/%s/_result", project, pkg)

	output, err := execCommand("osc", "api", apiPath)
	if err != nil {
		return nil, fmt.Errorf("obs: 'osc api' failed: %w", err)
	}

	var resultList struct {
		Results []core.BuildStatus `xml:"result"`
	}

	if err := xml.Unmarshal(output, &resultList); err != nil {
		return nil, fmt.Errorf("obs: failed to parse XML output: %w\nOutput received: %s", err, string(output))
	}

	return resultList.Results, nil
}