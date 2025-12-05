package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gyr/relx-go/pkg/command"
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/gitea"
)

// HandleReview initializes the Gitea client, fetches PRs, and prints the results.
// This function encapsulates the business logic for the 'review' command.
func HandleReview(ctx context.Context, cfg *config.Config, runner command.Runner, branch, repository string) error {
	cfg.Logger.Debugf("Handling review for branch=%s, repository=%s", branch, repository)

	if cfg.PRReviewer == "" {
		return fmt.Errorf("missing 'pr_reviewer' configuration")
	}

	giteaClient := gitea.NewClient(runner, cfg)

	prIDs, err := giteaClient.GetOpenPullRequests(ctx, cfg.PRReviewer, branch, repository)
	if err != nil {
		return fmt.Errorf("failed to get open pull requests: %w", err)
	}

	if len(prIDs) == 0 {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "No open pull requests found for reviewer '%s' on branch '%s' in repository '%s'.\n", cfg.PRReviewer, branch, repository); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintf(cfg.OutputWriter, "\n--- Open Pull Requests for Review ---\n"); err != nil {
		return err
	}
	for _, id := range prIDs {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "PR ID: %s\n", id); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(cfg.OutputWriter, "Do you want to review these pull requests? (y/n): "); err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}
	response = strings.ToLower(strings.TrimSpace(response))

	if response == "y" || response == "yes" {
		// Placeholder for future implementation
		if _, err := fmt.Fprintf(cfg.OutputWriter, "Proceeding with review (future implementation).\n"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "Exiting without review.\n"); err != nil {
			return err
		}
	}

	return nil
}
