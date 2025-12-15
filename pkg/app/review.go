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
func HandleReview(ctx context.Context, cfg *config.Config, runner command.Runner, branch, repository, user string) error {
	cfg.Logger.Debugf("Handling review for branch=%s, repository=%s", branch, repository)

	var reviewer string
	if user != "" {
		reviewer = user
	} else {
		reviewer = cfg.PRReviewer
	}

	if reviewer == "" {
		return fmt.Errorf("missing 'pr_reviewer' configuration and no user specified with -u/--user")
	}

	giteaClient := gitea.NewClient(runner, cfg)

	prIDs, err := giteaClient.GetOpenPullRequests(ctx, reviewer, branch, repository)
	if err != nil {
		return fmt.Errorf("failed to get open pull requests: %w", err)
	}

	if len(prIDs) == 0 {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "No open pull requests found for reviewer '%s' on branch '%s' in repository '%s'.\n", reviewer, branch, repository); err != nil {
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
		for _, id := range prIDs {
			if err := giteaClient.ShowPullRequest(ctx, repository, id); err != nil {
				return fmt.Errorf("failed to show pull request %s: %w", id, err)
			}

			if _, err := fmt.Fprintf(cfg.OutputWriter, "Approve, skip, or exit? (a/s/e): "); err != nil {
				return err
			}

			actionResponse, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read user input: %w", err)
			}
			actionResponse = strings.ToLower(strings.TrimSpace(actionResponse))

			switch actionResponse {
			case "a", "approve":
				if err := giteaClient.ApprovePullRequest(ctx, repository, id, reviewer); err != nil {
					return fmt.Errorf("failed to approve pull request %s: %w", id, err)
				}
				if _, err := fmt.Fprintf(cfg.OutputWriter, "PR %s approved.\n", id); err != nil {
					return err
				}
			case "s", "skip":
				if _, err := fmt.Fprintf(cfg.OutputWriter, "Skipping PR %s.\n", id); err != nil {
					return err
				}
				continue
			case "e", "exit":
				if _, err := fmt.Fprintf(cfg.OutputWriter, "Exiting review process.\n"); err != nil {
					return err
				}
				return nil
			default:
				if _, err := fmt.Fprintf(cfg.OutputWriter, "Invalid option. Skipping PR %s.\n", id); err != nil {
					return err
				}
			}
		}
	} else {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "Exiting without review.\n"); err != nil {
			return err
		}
	}

	return nil
}
