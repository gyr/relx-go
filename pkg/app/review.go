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
func HandleReview(ctx context.Context, cfg *config.Config, runner command.Runner, branch string, prIDs []string, repository, user string) error {
	cfg.Logger.Debugf("Handling review for branch=%s, prIDs=%v, repository=%s", branch, prIDs, repository)

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

	// Always get PRs by branch first, as branch is now mandatory.
	fetchedPRs, err := giteaClient.GetOpenPullRequests(ctx, reviewer, branch, repository)
	if err != nil {
		return fmt.Errorf("failed to get open pull requests for branch '%s': %w", branch, err)
	}

	var prsToReview []string
	if len(prIDs) > 0 {
		// User provided specific PR IDs, so filter the fetched PRs.
		fetchedPRsMap := make(map[string]struct{}, len(fetchedPRs))
		for _, id := range fetchedPRs {
			fetchedPRsMap[id] = struct{}{}
		}

		for _, providedID := range prIDs {
			if _, exists := fetchedPRsMap[providedID]; exists {
				prsToReview = append(prsToReview, providedID)
			} else {
				if _, err := fmt.Fprintf(cfg.OutputWriter, "Info: PR #%s (provided with -p) was not found pending review on branch '%s'.\n", providedID, branch); err != nil {
					return err
				}
			}
		}
	} else {
		// No specific PR IDs provided, review all fetched PRs for the branch.
		prsToReview = fetchedPRs
	}

	if len(prsToReview) == 0 {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "No open pull requests found for review.\n"); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintf(cfg.OutputWriter, "\n--- Open Pull Requests for Review ---\n"); err != nil {
		return err
	}
	for _, id := range prsToReview {
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
		for _, id := range prsToReview {
			if err := giteaClient.ShowPullRequest(ctx, repository, id); err != nil {
				cfg.Logger.Warnf("Failed to show pull request %s: %v. Skipping.", id, err)
				continue
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
					cfg.Logger.Warnf("Failed to approve pull request %s: %v.", id, err)
				} else {
					if _, err := fmt.Fprintf(cfg.OutputWriter, "PR %s approved.\n", id); err != nil {
						return err
					}
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
