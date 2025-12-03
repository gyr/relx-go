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

	// Step 2: Filter the package list based on configured patterns.
	var filteredPackages []string
	if len(c.cfg.PackageFilterPatterns) > 0 {
		for _, pkg := range packages {
			for _, pattern := range c.cfg.PackageFilterPatterns {
				matched, err := filepath.Match(pattern, pkg)
				if err != nil {
					c.cfg.Logger.Warnf("Invalid pattern '%s' in config: %v", pattern, err)
					continue
				}
				if matched {
					filteredPackages = append(filteredPackages, pkg)
					break // Match found, no need to check other patterns for this package
				}
			}
		}
	} else {
		// If no patterns are defined, include all packages.
		filteredPackages = packages
	}

	c.cfg.Logger.Infof("Found %d packages matching filter patterns.", len(filteredPackages))
	c.cfg.Logger.Debugf("Filtered packages: %v", filteredPackages)

	// Step 3: Concurrently get binaries for each filtered package.
	// Use a WaitGroup to wait for all goroutines to finish.
	var wg sync.WaitGroup
	// Create a buffered channel to collect errors from goroutines.
	// The buffer size is the number of packages, so each goroutine can send an error without blocking.
	errCh := make(chan error, len(filteredPackages))
	// Create a buffered channel to collect the results (lists of binaries) from goroutines.
	resultsCh := make(chan []string, len(filteredPackages))
	// Create a semaphore to limit the number of concurrent 'osc' commands.
	// This prevents overwhelming the system with too many processes.
	sem := make(chan struct{}, maxConcurrentOscCalls)

	// Loop over each filtered package and start a goroutine to process it.
	for _, pkg := range filteredPackages {
		// Acquire a slot from the semaphore. This will block if the maximum number of
		// concurrent goroutines is already running.
		sem <- struct{}{}
		// Increment the WaitGroup counter.
		wg.Add(1)
		// Launch a new goroutine.
		go func(pkgName string) {
			// Decrement the WaitGroup counter when the goroutine finishes.
			defer wg.Done()
			// Release the slot back to the semaphore.
			defer func() { <-sem }()

			// Call the function to get the binaries for the package.
			binaries, err := c.listBinariesForPackage(ctx, project, pkgName)
			if err != nil {
				// If there's an error, send it to the error channel.
				errCh <- fmt.Errorf("failed to list binaries for package '%s': %w", pkgName, err)
				return
			}
			// If successful, send the list of binaries to the results channel.
			resultsCh <- binaries
		}(pkg)
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

// listBinariesForPackage runs `osc ls -b` for a single package.
func (c *Client) listBinariesForPackage(ctx context.Context, project, pkg string) ([]string, error) {
	timeout := time.Duration(c.cfg.OperationTimeoutSeconds) * time.Second
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	c.cfg.Logger.Debugf("Executing 'osc ls -b' for package: %s", pkg)

	var output []byte
	var err error

	if c.cfg.OBSAPIURL != "" {
		output, err = c.runner.Run(timeoutCtx, "" /* workDir */, "osc", "-A", c.cfg.OBSAPIURL, "ls", "-b", project, pkg)
	} else {
		output, err = c.runner.Run(timeoutCtx, "" /* workDir */, "osc", "ls", "-b", project, pkg)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to run 'osc ls -b' for package '%s': %w. Output: %s", pkg, err, string(output))
	}

	binaries := strings.Split(string(output), "\n")

	var cleanedBinaries []string
	for _, b := range binaries {
		// Filter lines that start with a space but not with " _"
		if strings.HasPrefix(b, " ") && !strings.HasPrefix(b, " _") {
			cleanedBinaries = append(cleanedBinaries, strings.TrimSpace(b))
		}
	}

	return cleanedBinaries, nil
}
