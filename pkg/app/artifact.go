package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/gyr/relx-go/pkg/command" // Needed to pass to obs.NewClient
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/obs" // Import the new OBS client
)

// HandleArtifacts is the handler for the 'artifact' subcommand.
// It orchestrates the fetching and display of artifacts for a given project.
func HandleArtifacts(ctx context.Context, cfg *config.Config, runner command.Runner, project string) error {
	cfg.Logger.Infof("Handling artifact request for project: %s", project)

	// Create an OBS client instance
	obsClient := obs.NewClient(runner, cfg)

	// Use the OBS client to list artifacts
	artifacts, err := obsClient.ListArtifacts(ctx, project)
	if err != nil {
		return fmt.Errorf("failed to list artifacts for project %s: %w", project, err)
	}

	if len(artifacts) > 0 {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "Artifacts for project '%s':\n%s\n", project, strings.Join(artifacts, "\n")); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "No artifacts found for project '%s'.\n", project); err != nil {
			return err
		}
	}

	return nil
}
