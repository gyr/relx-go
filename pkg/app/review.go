package app

import (
	"fmt"

	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/gitea"
)

// HandleReview initializes the Gitea client, fetches PRs, and prints the results.
// This function encapsulates the business logic for the 'review' command.
func HandleReview(cfg *config.Config, owner, repo string) error {
	cfg.Logger.Debugf("Handling review for owner=%s, repo=%s", owner, repo)

	if cfg.PRReviewer == "" {
		return fmt.Errorf("missing 'pr_reviewer' configuration")
	}

	// Initialize the specific Gitea client
	giteaClient := gitea.NewClient(cfg.CacheDir)

	// The client handles the os/exec command and JSON parsing internally.
	prs, err := giteaClient.GetPullRequests(owner, repo)
	if err != nil {
		return fmt.Errorf("gitea PR error: %w", err)
	}

	if _, err := fmt.Fprintf(cfg.OutputWriter, "\n--- Open Pull Requests in %s/%s ---\n", owner, repo); err != nil {
		return err
	}
	for _, pr := range prs {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "[%d] %s (State: %s, URL: %s)\n", pr.ID, pr.Title, pr.State, pr.URL); err != nil {
			return err
		}
	}
	if len(prs) == 0 {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "No open pull requests found.\n"); err != nil {
			return err
		}
	}
	return nil
}
