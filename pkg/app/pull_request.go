package app

import (
	"fmt"
	"log"

	"github.com/gyr/grxs/pkg/config"
	"github.com/gyr/grxs/pkg/gitea"
)

// HandlePullRequest initializes the Gitea client, fetches PRs, and prints the results.
// This function encapsulates the business logic for the 'pr' command.
func HandlePullRequest(cfg *config.Config, owner, repo string) {
	if cfg.Debug {
		log.Printf("Debug: Handling pull request for owner=%s, repo=%s", owner, repo)
	}

	// Initialize the specific Gitea client
	giteaClient := gitea.NewClient(cfg.CacheDir)

	// The client handles the os/exec command and JSON parsing internally.
	prs, err := giteaClient.GetPullRequests(owner, repo)
	if err != nil {
		log.Fatalf("Gitea PR Error: %v", err)
	}

	fmt.Printf("\n--- Open Pull Requests in %s/%s ---\n", owner, repo)
	for _, pr := range prs {
		fmt.Printf("[%d] %s (State: %s, URL: %s)\n", pr.ID, pr.Title, pr.State, pr.URL)
	}
	if len(prs) == 0 {
		fmt.Println("No open pull requests found.")
	}
}