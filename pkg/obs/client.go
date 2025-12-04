package obs

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gyr/relx-go/pkg/command"
	"github.com/gyr/relx-go/pkg/config"
)

const maxConcurrentOscCalls = 10

// Client handles interaction with the OBS API via the 'osc' command-line tool.
// It encapsulates all the logic for forming and executing osc commands.
type Client struct {
	runner command.Runner
	cfg    *config.Config
}

// NewClient creates a new OBS client instance.
// It requires a command.Runner for executing external commands and the application config.
func NewClient(runner command.Runner, cfg *config.Config) *Client {
	return &Client{
		runner: runner,
		cfg:    cfg,
	}
}

// ListArtifacts is the high-level method to get a final list of artifacts.
// It encapsulates the entire workflow of listing packages, filtering them,
// and eventually finding and filtering their binaries.
func (c *Client) ListArtifacts(ctx context.Context, project string) ([]string, error) {
	c.cfg.Logger.Infof("Starting artifact search for project: %s", project)

	// Step 1: Get the list of all packages in the project.
	packages, err := c.listPackages(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages for project %s: %w", project, err)
	}
	c.cfg.Logger.Debugf("Found %d packages in project %s.", len(packages), project)

	// Step 2: Filter the package list based on configured patterns and associate them with a repository.
	// A map is used to store the package -> repository mapping, which also de-duplicates packages.
	filteredPackages := make(map[string]string)
	if len(c.cfg.PackageFilterPatterns) > 0 {
		for _, pkg := range packages {
			for _, filter := range c.cfg.PackageFilterPatterns {
				matched, err := filepath.Match(filter.Pattern, pkg)
				if err != nil {
					c.cfg.Logger.Warnf("Invalid pattern '%s' in config: %v", filter.Pattern, err)
					continue
				}
				if matched {
					filteredPackages[pkg] = filter.Repository
					break // Match found, no need to check other patterns for this package
				}
			}
		}
	} else {
		// If no package filters are defined, we cannot proceed because we don't know which
		// repositories to target. The user must be explicit.
		c.cfg.Logger.Infof("No package_filter_patterns defined in config. No packages to process.")
	}

	c.cfg.Logger.Infof("Found %d packages matching filter patterns.", len(filteredPackages))
	c.cfg.Logger.Debugf("Filtered packages and their repositories: %v", filteredPackages)

	if len(filteredPackages) == 0 {
		return []string{}, nil
	}

	// Step 3: Concurrently get binaries for each filtered package.
	var wg sync.WaitGroup
	errCh := make(chan error, len(filteredPackages))
	resultsCh := make(chan []string, len(filteredPackages))
	sem := make(chan struct{}, maxConcurrentOscCalls)

	for pkg, repo := range filteredPackages {
		sem <- struct{}{}
		wg.Add(1)
		go func(pkgName, repoName string) {
			defer wg.Done()
			defer func() { <-sem }()

			binaries, err := c.listBinariesForPackage(ctx, project, pkgName, repoName)
			if err != nil {
				errCh <- fmt.Errorf("failed to list binaries for package '%s': %w", pkgName, err)
				return
			}
			resultsCh <- binaries
		}(pkg, repo)
	}

	// Start a separate goroutine to wait for all other goroutines to complete.
	// This allows the main flow to proceed to collecting results and errors
	// while waiting for the channels to be closed.
	go func() {
		wg.Wait()
		// Once all goroutines are done, close the channels to signal that no more
		// data will be sent.
		close(errCh)
		close(resultsCh)
	}()

	// Collect all errors from the error channel.
	// This loop will run until the channel is closed.
	var allErrors []error
	for err := range errCh {
		allErrors = append(allErrors, err)
	}

	// If any errors were collected, return them.
	if len(allErrors) > 0 {
		return nil, fmt.Errorf("multiple errors occurred while listing binaries: %v", allErrors)
	}

	// Collect all the results from the results channel.
	// This loop will also run until the channel is closed.
	var allBinaries []string
	for binaries := range resultsCh {
		allBinaries = append(allBinaries, binaries...)
	}

	// Step 4: Filter the final binary list based on configured patterns.
	if len(c.cfg.BinaryFilterPatterns) == 0 {
		sort.Strings(allBinaries)
		return allBinaries, nil // No filter, return everything
	}

	var filteredBinaries []string
	for _, binary := range allBinaries {
		for _, pattern := range c.cfg.BinaryFilterPatterns {
			matched, err := filepath.Match(pattern, binary)
			if err != nil {
				c.cfg.Logger.Warnf("Invalid binary pattern '%s' in config: %v", pattern, err)
				continue
			}
			if matched {
				filteredBinaries = append(filteredBinaries, binary)
				break // Match found, no need to check other patterns
			}
		}
	}

	c.cfg.Logger.Infof("Found %d binaries matching filter patterns.", len(filteredBinaries))
	c.cfg.Logger.Debugf("Filtered binaries: %v", filteredBinaries)

	sort.Strings(filteredBinaries)
	return filteredBinaries, nil
}

// listPackages runs `osc ls` to get a list of all packages in a project.
func (c *Client) listPackages(ctx context.Context, project string) ([]string, error) {
	timeout := time.Duration(c.cfg.OperationTimeoutSeconds) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	c.cfg.Logger.Debugf("Executing 'osc ls' for project: %s", project)

	var output []byte
	var err error

	if c.cfg.OBSAPIURL != "" {
		output, err = c.runner.Run(timeoutCtx, "" /* workDir */, "osc", "-A", c.cfg.OBSAPIURL, "ls", project)
	} else {
		output, err = c.runner.Run(timeoutCtx, "" /* workDir */, "osc", "ls", project)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to run 'osc ls' for project '%s': %w. Output: %s", project, err, string(output))
	}

	packages := strings.Split(string(output), "\n")

	var cleanedPackages []string
	for _, pkg := range packages {
		if strings.TrimSpace(pkg) != "" {
			cleanedPackages = append(cleanedPackages, strings.TrimSpace(pkg))
		}
	}

	return cleanedPackages, nil
}

// listBinariesForPackage runs `osc ls -b` for a single package and optional repository.
func (c *Client) listBinariesForPackage(ctx context.Context, project, pkg, repository string) ([]string, error) {
	timeout := time.Duration(c.cfg.OperationTimeoutSeconds) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	c.cfg.Logger.Debugf("Executing 'osc ls -b' for package: %s, repository: %s", pkg, repository)

	args := []string{"ls", "-b", project, pkg}
	if c.cfg.OBSAPIURL != "" {
		// Prepend -A <api_url> to the argument list
		args = append([]string{"-A", c.cfg.OBSAPIURL}, args...)
	}
	if repository != "" {
		args = append(args, "-r", repository)
	}

	output, err := c.runner.Run(timeoutCtx, "" /* workDir */, "osc", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to run 'osc ls -b' for package '%s': %w. Output: %s", pkg, err, string(output))
	}

	binaries := strings.Split(string(output), "\n")

	// Use a map to automatically handle duplicates from osc's output.
	cleanedBinariesMap := make(map[string]struct{})
	for _, b := range binaries {
		if strings.HasPrefix(b, " ") && !strings.HasPrefix(b, " _") {
			cleanedBinariesMap[strings.TrimSpace(b)] = struct{}{}
		}
	}

	// Convert map keys back to a slice.
	var cleanedBinaries []string
	for b := range cleanedBinariesMap {
		cleanedBinaries = append(cleanedBinaries, b)
	}

	return cleanedBinaries, nil
}
