package gitea

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gyr/relx-go/pkg/command/commandtest"
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
)

func TestGetOpenPullRequests(t *testing.T) {
	mockCfg := &config.Config{
		Logger:                  logging.NewLogger(logging.LevelDebug),
		OperationTimeoutSeconds: 5,
	}
	const prReviewer = "test_reviewer"
	const branch = "master"
	const repository = "test_repo"

	t.Run("Success", func(t *testing.T) {
		mockOutput := `
ID          : products/SLES#499
URL         : https://src.suse.de/products/SLES/pulls/499
Title       : PackageHub release spec file fixes
State       : open
ID          : products/SLES#496
URL         : https://src.suse.de/products/SLES/pulls/496
Title       : Adding development-tools-obs to build for Backports
State       : open
`
		mockRunner := &commandtest.MockRunner{}
		mockRunner.RunFunc = func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
			if name != "git-obs" {
				return nil, fmt.Errorf("unexpected command: %s %v", name, args)
			}
			return []byte(mockOutput), nil
		}

		client := NewClient(mockRunner, mockCfg)
		prIDs, err := client.GetOpenPullRequests(context.Background(), prReviewer, branch, repository)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expectedPRs := []string{"499", "496"}
		if !reflect.DeepEqual(prIDs, expectedPRs) {
			t.Errorf("Pull request list mismatch:\nGot:  %v\nWant: %v", prIDs, expectedPRs)
		}
	})

	t.Run("Command fails", func(t *testing.T) {
		mockError := errors.New("git-obs command failed")
		mockRunner := &commandtest.MockRunner{
			RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
				return nil, mockError
			},
		}

		client := NewClient(mockRunner, mockCfg)
		_, err := client.GetOpenPullRequests(context.Background(), prReviewer, branch, repository)

		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}

		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Expected error to contain '%v', but got '%v'", mockError, err)
		}
	})
}
