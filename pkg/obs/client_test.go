package obs

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/gyr/relx-go/pkg/command/commandtest"
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
)

func TestListArtifacts(t *testing.T) {
	mockCfg := &config.Config{
		OBSAPIURL: "https://api.suse.de",
		Logger:    logging.NewLogger(logging.LevelDebug),
		PackageFilterPatterns: []config.PackageFilter{
			{Pattern: "000product*", Repository: "repo1"},
			{Pattern: "SLES_transactional:*", Repository: "repo2"},
		},
		BinaryFilterPatterns: []string{
			"*.report", // only include report files
		},
		OperationTimeoutSeconds: 5,
	}
	const project = "SUSE:SLE-15-SP4:Update"

	t.Run("Success with filtering", func(t *testing.T) {
		// This is the sample output from `osc ls`
		oscLsOutput := `
000productcompose:sles_product
SLES_transactional:self-install
some-other-package
000product-another
unrelated-package
`
		// Sample outputs for `osc ls -b <package>`
		binariesForProductCompose := `
 SLE-INSTALLER-16.1-aarch64-Build8.1.report
 SLE-INSTALLER-16.1-aarch64-Build8.1.json
`
		binariesForTransactional := `
 SLE-INSTALLER-16.1-aarch64-Build9.1.report
`
		binariesForAnother := `
 SLE-INSTALLER-16.1-x86_64-Build10.1.report
 SLE-INSTALLER-16.1-x86_64-Build10.1.iso
`

		mockRunner := &commandtest.MockRunner{}
		mockRunner.RunFunc = func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
			if name != "osc" {
				return nil, fmt.Errorf("unexpected command: %s %v", name, args)
			}
			cmd := strings.Join(args, " ")
			if strings.Contains(cmd, "ls -b") {
				switch {
				case strings.Contains(cmd, "000productcompose:sles_product") && strings.Contains(cmd, "-r repo1"):
					return []byte(binariesForProductCompose), nil
				case strings.Contains(cmd, "SLES_transactional:self-install") && strings.Contains(cmd, "-r repo2"):
					return []byte(binariesForTransactional), nil
				case strings.Contains(cmd, "000product-another") && strings.Contains(cmd, "-r repo1"):
					return []byte(binariesForAnother), nil
				default:
					return nil, fmt.Errorf("unexpected ls -b command for package in: %s", cmd)
				}
			}
			expectedCmdPart := fmt.Sprintf("ls %s", project)
			if strings.Contains(cmd, expectedCmdPart) {
				return []byte(oscLsOutput), nil
			}
			return nil, fmt.Errorf("unexpected command: %s %v", name, args)
		}

		client := NewClient(mockRunner, mockCfg)
		artifacts, err := client.ListArtifacts(context.Background(), project)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expectedArtifacts := []string{
			"SLE-INSTALLER-16.1-aarch64-Build8.1.report",
			"SLE-INSTALLER-16.1-aarch64-Build9.1.report",
			"SLE-INSTALLER-16.1-x86_64-Build10.1.report",
		}

		// Sort the expected slice for stable comparison
		sort.Strings(expectedArtifacts)

		if !reflect.DeepEqual(artifacts, expectedArtifacts) {
			t.Errorf("Filtered list mismatch:\nGot:  %v\nWant: %v", artifacts, expectedArtifacts)
		}
	})

	t.Run("listPackages fails", func(t *testing.T) {
		mockError := errors.New("osc command failed")
		mockRunner := &commandtest.MockRunner{}
		mockRunner.RunFunc = func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
			// Fail only on the project listing command
			if strings.Contains(strings.Join(args, " "), "ls "+project) {
				return nil, mockError
			}
			return nil, nil // Should not be reached in this test
		}

		client := NewClient(mockRunner, mockCfg)
		_, err := client.ListArtifacts(context.Background(), project)

		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}
		expectedErrStr := "failed to list packages for project"
		if !strings.Contains(err.Error(), expectedErrStr) {
			t.Errorf("Expected error to contain '%s', but got '%s'", expectedErrStr, err.Error())
		}
	})

	t.Run("listBinariesForPackage fails concurrently", func(t *testing.T) {
		oscLsOutput := `
000productcompose:sles_product
SLES_transactional:self-install
`
		mockError := errors.New("specific binary fetch failed")
		mockRunner := &commandtest.MockRunner{}
		mockRunner.RunFunc = func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
			cmd := strings.Join(args, " ")
			if strings.Contains(cmd, "ls -b") {
				// Fail for one specific package
				if strings.Contains(cmd, "SLES_transactional:self-install") {
					return nil, mockError
				}
				return []byte("some-binary"), nil
			}
			if strings.Contains(cmd, "ls "+project) {
				return []byte(oscLsOutput), nil
			}
			return nil, fmt.Errorf("unexpected command: %s %v", name, args)
		}

		client := NewClient(mockRunner, mockCfg)
		_, err := client.ListArtifacts(context.Background(), project)

		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}

		expectedErrStr := "multiple errors occurred"
		if !strings.Contains(err.Error(), expectedErrStr) {
			t.Errorf("Expected error message to contain '%s', got '%s'", expectedErrStr, err.Error())
		}
		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Expected error message to contain the specific mock error '%s', got '%s'", mockError.Error(), err.Error())
		}
	})
}

func TestListBinariesForPackage(t *testing.T) {
	mockCfg := &config.Config{
		OBSAPIURL:               "https://api.suse.de",
		Logger:                  logging.NewLogger(logging.LevelDebug),
		OperationTimeoutSeconds: 5,
	}
	const project = "SUSE:SLE-15-SP4:Update"
	const pkg = "my-package"

	t.Run("Success with filtering", func(t *testing.T) {
		oscLsBOutput := `
standard/aarch64
product/x86_64
 SLE-INSTALLER-16.1-aarch64-Build8.1.report
 _buildenv
 SLE-INSTALLER-16.1-x86_64-Build10.1.report
`
		mockRunner := &commandtest.MockRunner{
			RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
				return []byte(oscLsBOutput), nil
			},
		}

		client := NewClient(mockRunner, mockCfg)
		binaries, err := client.listBinariesForPackage(context.Background(), project, pkg, "")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expectedBinaries := []string{
			"SLE-INSTALLER-16.1-aarch64-Build8.1.report",
			"SLE-INSTALLER-16.1-x86_64-Build10.1.report",
		}
		sort.Strings(binaries)
		sort.Strings(expectedBinaries)

		if !reflect.DeepEqual(binaries, expectedBinaries) {
			t.Errorf("Binary list mismatch:\nGot:  %v\nWant: %v", binaries, expectedBinaries)
		}
	})

	t.Run("Command fails", func(t *testing.T) {
		mockError := errors.New("osc ls -b failed")
		mockRunner := &commandtest.MockRunner{
			RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
				return nil, mockError
			},
		}

		client := NewClient(mockRunner, mockCfg)
		_, err := client.listBinariesForPackage(context.Background(), project, pkg, "")

		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}

		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Expected error to contain '%v', but got '%v'", mockError, err)
		}
	})
}
