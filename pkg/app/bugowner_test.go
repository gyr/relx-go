package app

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gyr/relx-go/pkg/command/commandtest" // Import the shared mock runner
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
)

func TestHandleBugownerByPackage(t *testing.T) {
	const repoURL = "https://example.com/test/repo.git"
	const maintainershipContent = `{"pkg1": ["userA", "userB"], "pkg2": ["userC"]}`

	// This is the mock for a successful git archive call.
	successfulRunner := &commandtest.MockRunner{
		RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
			// Check if the command is what we expect
			if name == "bash" && strings.Contains(args[1], "git archive") {
				return []byte(maintainershipContent), nil
			}
			return nil, nil
		},
	}

	t.Run("PackageFound", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}

		err := HandleBugownerByPackage(context.Background(), cfg, successfulRunner, "pkg1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		output := out.String()
		if !strings.Contains(output, "Maintainers for package pkg1:") {
			t.Errorf("Output missing expected header. Got: %s", output)
		}
		if !strings.Contains(output, "userA") || !strings.Contains(output, "userB") {
			t.Errorf("Output missing expected maintainers. Got: %s", output)
		}
	})

	t.Run("PackageNotFound", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}

		err := HandleBugownerByPackage(context.Background(), cfg, successfulRunner, "nonexistent")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		output := out.String()
		expectedMsg := "Package 'nonexistent' not found"
		if !strings.Contains(output, expectedMsg) {
			t.Errorf("Expected message %q, but it was not found in output: %s", expectedMsg, output)
		}
	})

	t.Run("FetchFileFailure", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}

		mockError := errors.New("git archive failed")
		failedRunner := &commandtest.MockRunner{
			RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
				return nil, mockError
			},
		}

		err := HandleBugownerByPackage(context.Background(), cfg, failedRunner, "pkg1")
		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Error message missing expected substring %q: %v", mockError.Error(), err)
		}
	})
}

func TestHandlePackagesByMaintainer(t *testing.T) {
	const repoURL = "https://example.com/test/repo.git"
	const maintainershipContent = `{"pkg1": ["userA", "userB"], "pkg2": ["userC"], "pkg3": ["userA"]}`

	successfulRunner := &commandtest.MockRunner{
		RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
			if name == "bash" && strings.Contains(args[1], "git archive") {
				return []byte(maintainershipContent), nil
			}
			return nil, nil
		},
	}

	t.Run("MaintainerFound", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}

		err := HandlePackagesByMaintainer(context.Background(), cfg, successfulRunner, "userA")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		output := out.String()
		if !strings.Contains(output, "Packages maintained by userA:") {
			t.Errorf("Output missing expected header. Got: %s", output)
		}
		if !strings.Contains(output, "pkg1") || !strings.Contains(output, "pkg3") {
			t.Errorf("Output missing expected packages. Got: %s", output)
		}
	})

	t.Run("MaintainerNotFound", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}

		err := HandlePackagesByMaintainer(context.Background(), cfg, successfulRunner, "nonexistent")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		output := out.String()
		expectedMsg := "No packages found for maintainer 'nonexistent'"
		if !strings.Contains(output, expectedMsg) {
			t.Errorf("Expected message %q, but it was not found in output: %s", expectedMsg, output)
		}
	})

	t.Run("JSONMalformed", func(t *testing.T) {
		var out bytes.Buffer
		cfg := &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			OutputWriter: &out,
			RepoURL:      repoURL,
			RepoBranch:   "main",
		}

		malformedRunner := &commandtest.MockRunner{
			RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
				return []byte("this is not json"), nil
			},
		}

		err := HandlePackagesByMaintainer(context.Background(), cfg, malformedRunner, "userA")
		if err == nil {
			t.Fatal("Expected an error for malformed JSON, got nil")
		}
		expectedErrSubstring := "error unmarshaling maintainership JSON"
		if !strings.Contains(err.Error(), expectedErrSubstring) {
			t.Errorf("Expected error message to contain %q, but got %v", expectedErrSubstring, err)
		}
	})
}
