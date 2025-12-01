package app

import (
	"fmt"

	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/gitea"
)

// HandlePullRequest initializes the Gitea client, fetches PRs, and prints the results.
// This function encapsulates the business logic for the 'pr' command.
func HandlePullRequest(cfg *config.Config, owner, repo string) {
	cfg.Logger.Debugf("Handling pull request for owner=%s, repo=%s", owner, repo)

	// Initialize the specific Gitea client
	giteaClient := gitea.NewClient(cfg.CacheDir)

	// The client handles the os/exec command and JSON parsing internally.
	prs, err := giteaClient.GetPullRequests(owner, repo)
	if err != nil {
		cfg.Logger.Fatalf("Gitea PR Error: %v", err)
	}

	fmt.Printf("\n--- Open Pull Requests in %s/%s ---\n", owner, repo)
	for _, pr := range prs {
		fmt.Printf("[%d] %s (State: %s, URL: %s)\n", pr.ID, pr.Title, pr.State, pr.URL)
	}
	if len(prs) == 0 {
		fmt.Println("No open pull requests found.")
	}
}
