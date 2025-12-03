package app

import (
	"context" // Import context for cancellation and timeouts
	"encoding/json"
	"fmt"

	"github.com/gyr/relx-go/pkg/command" // Import the new command runner interface
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/gitutils"
)

// prepareMaintainershipData fetches the _maintainership.json file from the remote git
// repository and unmarshals it. It uses the efficient 'git archive' method
// to avoid cloning the entire repository.
func prepareMaintainershipData(ctx context.Context, cfg *config.Config, runner command.Runner) (map[string][]string, error) {
	const maintainershipFilename = "_maintainership.json"

	// Fetch the remote file content using the new, efficient git archive method.
	fileContent, err := gitutils.FetchRemoteFile(ctx, cfg, runner, maintainershipFilename)
	if err != nil {
		return nil, fmt.Errorf("error fetching maintainership data: %w", err)
	}

	// Unmarshal the JSON data directly.
	var maintainers map[string][]string
	if err := json.Unmarshal(fileContent, &maintainers); err != nil {
		return nil, fmt.Errorf("error unmarshaling maintainership JSON: %w", err)
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
