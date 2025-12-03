package app

import (
	"context" // Import context for cancellation and timeouts
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gyr/relx-go/pkg/command" // Import the new command runner interface
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/gitutils"
)

// prepareMaintainershipData ensures the repository is cloned/updated and loads the maintainership data from it.
// It now accepts a context and a command.Runner, making it more testable and enabling
// proper cancellation/timeout propagation.
// This is an example of Dependency Injection: the 'runner' dependency is passed in,
// rather than being created internally, improving modularity and testability.
func prepareMaintainershipData(ctx context.Context, cfg *config.Config, runner command.Runner) (map[string][]string, error) {
	// Clone or update the repository using the injected runner.
	localPath, err := gitutils.ManageRepo(ctx, cfg, runner)
	if err != nil {
		return nil, fmt.Errorf("error cloning/updating repository: %w", err)
	}
	cfg.Logger.Infof("Repository available at: %s", localPath)

	maintainers, err := loadMaintainershipData(localPath)
	if err != nil {
		return nil, fmt.Errorf("error loading maintainership data: %w", err)
	}

	return maintainers, nil
}

// HandleBugownerByPackage fetches and displays the bug owners for a given package.
// It now accepts a context and a command.Runner, demonstrating Dependency Injection
// for improved testability and operational control.
func HandleBugownerByPackage(ctx context.Context, cfg *config.Config, runner command.Runner, pkg string) error {
	cfg.Logger.Infof("Handling bug owner request for package %s", pkg)

	maintainers, err := prepareMaintainershipData(ctx, cfg, runner)
	if err != nil {
		return err
	}

	if pkgMaintainers, found := maintainers[pkg]; found {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "Maintainers for package %s:\n", pkg); err != nil {
			return err
		}
		for _, m := range pkgMaintainers {
			if _, err := fmt.Fprintf(cfg.OutputWriter, "  - %s\n", m); err != nil {
				return err
			}
		}
	} else {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "Package '%s' not found in maintainership data.\n", pkg); err != nil {
			return err
		}
	}
	return nil
}

// HandlePackagesByMaintainer lists the packages maintained by a given user.
// It now accepts a context and a command.Runner, demonstrating Dependency Injection
// for improved testability and operational control.
func HandlePackagesByMaintainer(ctx context.Context, cfg *config.Config, runner command.Runner, maintainer string) error {
	cfg.Logger.Infof("Handling packages by maintainer request for %s", maintainer)

	maintainers, err := prepareMaintainershipData(ctx, cfg, runner)
	if err != nil {
		return err
	}

	var foundPackages []string
	for pkg, maintainerList := range maintainers {
		for _, m := range maintainerList {
			if m == maintainer {
				foundPackages = append(foundPackages, pkg)
				break // Move to the next package once a match is found
			}
		}
	}

	if len(foundPackages) > 0 {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "Packages maintained by %s:\n", maintainer); err != nil {
			return err
		}
		for _, pkg := range foundPackages {
			if _, err := fmt.Fprintf(cfg.OutputWriter, "  - %s\n", pkg); err != nil {
				return err
			}
		}
	} else {
		if _, err := fmt.Fprintf(cfg.OutputWriter, "No packages found for maintainer '%s'.\n", maintainer); err != nil {
			return err
		}
	}
	return nil
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
