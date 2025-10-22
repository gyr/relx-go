package app

import (
	"fmt"
	"log"

	"github.com/gyr/grxs/pkg/config"
	"github.com/gyr/grxs/pkg/obs"
)

// HandleBuildStatus initializes the OBS client, fetches build status, and prints the results.
// This function encapsulates the business logic for the 'status' command.
func HandleBuildStatus(cfg *config.Config, project, pkg string) {
	if cfg.Debug {
		log.Printf("Debug: Handling build status for project=%s, package=%s", project, pkg)
	}

	// Initialize the specific OBS client
	// If the OBS client needed configuration from cfg, it would be passed here.
	obsClient := obs.NewClient()

	// The client handles the os/exec command and XML parsing internally.
	results, err := obsClient.GetBuildStatus(project, pkg)
	if err != nil {
		log.Fatalf("OBS Status Error: %v", err)
	}

	fmt.Printf("\n--- OBS Build Results for %s/%s ---\n", project, pkg)
	for _, res := range results {
		fmt.Printf("Project: %s, Package: %s, Repo: %s, Status: %s\n", res.Project, res.Package, res.Repository, res.Status)
	}
	if len(results) == 0 {
		fmt.Println("No build results found.")
	}
}
