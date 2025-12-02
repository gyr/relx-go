package app

import (
	"fmt"

	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/obs"
)

// HandleBuildStatus initializes the OBS client, fetches build status, and prints the results.
// This function encapsulates the business logic for the 'status' command.
func HandleBuildStatus(cfg *config.Config, project, pkg string) error {
	cfg.Logger.Debugf("Handling build status for project=%s, package=%s", project, pkg)

	// Initialize the specific OBS client
	// If the OBS client needed configuration from cfg, it would be passed here.
	obsClient := obs.NewClient()

	// The client handles the os/exec command and XML parsing internally.
	results, err := obsClient.GetBuildStatus(project, pkg)
	if err != nil {
		return fmt.Errorf("obs status error: %w", err)
	}

	if _, err := fmt.Fprintf(cfg.OutputWriter, "\n--- OBS Build Results for %s/%s ---\n", project, pkg); err != nil {
		return err
	}
	for _, res := range results {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "Project: %s, Package: %s, Repo: %s, Status: %s\n", res.Project, res.Package, res.Repository, res.Status); err != nil {
			return err
		}
	}
	if len(results) == 0 {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "No build results found.\n"); err != nil {
			return err
		}
	}
	return nil
}
