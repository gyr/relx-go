package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/gitutils"
)

// HandleBugownerByPackage fetches and displays the bug owners for a given package.
func HandleBugownerByPackage(cfg *config.Config, pkg string) {
	cfg.Logger.Infof("Handling bug owner request for package %s", pkg)
	fmt.Printf("Fetching bug owners for package: %s\n", pkg)

	// Clone or update the repository
	localPath, err := gitutils.ManageRepo(cfg)
	if err != nil {
		cfg.Logger.Fatalf("Error cloning/updating repository: %v", err)
	}
	cfg.Logger.Infof("Repository available at: %s", localPath)

	maintainers, err := loadMaintainershipData(localPath)
	if err != nil {
		cfg.Logger.Fatalf("Error loading maintainership data: %v", err)
	}

	cfg.Logger.Debugf("Loaded maintainership data: %+v", maintainers)

	// TODO: Use the maintainers map to find bug owners for the package.
	// For now, we'll just print a placeholder message.
}

// HandlePackagesByMaintainer lists the packages maintained by a given user.
func HandlePackagesByMaintainer(cfg *config.Config, maintainer string) {
	// TODO: Implement the logic to fetch packages for the maintainer.
	// This will likely involve interacting with a system that maps maintainers to packages.
	// For now, we'll just print a placeholder message.
	fmt.Printf("Listing packages for maintainer: %s\n", maintainer)
	cfg.Logger.Infof("Handling packages by maintainer request for %s", maintainer)
}

// loadMaintainershipData reads and unmarshals the _maintainership.json file.
func loadMaintainershipData(localPath string) (map[string][]string, error) {
	maintainershipFilename := "_maintainership.json"
	maintainershipFile := filepath.Join(localPath, maintainershipFilename)

	data, err := os.ReadFile(maintainershipFile)
	if err != nil {
		return nil, fmt.Errorf("error reading maintainership file %s: %w", maintainershipFile, err)
	}

	var maintainers map[string][]string
	if err := json.Unmarshal(data, &maintainers); err != nil {
		return nil, fmt.Errorf("error unmarshaling maintainership JSON from %s: %w", maintainershipFile, err)
	}

	return maintainers, nil
}
