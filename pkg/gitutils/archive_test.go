package gitutils

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gyr/relx-go/pkg/command/commandtest"
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
)

func TestFetchRemoteFile(t *testing.T) {
	mockCfg := &config.Config{
		RepoURL:                 "https://example.com/test.git",
		RepoBranch:              "main",
		Logger:                  logging.NewLogger(logging.LevelDebug),
		OperationTimeoutSeconds: 5,
	}
	const filePath = "_maintainership.json"

	t.Run("SuccessfulFetch", func(t *testing.T) {
		expectedContent := `{"pkg1": ["userA"]}`
		mockRunner := &commandtest.MockRunner{}
		mockRunner.RunFunc = func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
			// Check that the correct command is being run
			expectedCmdPart := "git archive --remote=https://example.com/test.git main _maintainership.json"
			if name != "bash" || !strings.Contains(args[1], expectedCmdPart) {
				t.Errorf("Expected command containing %q, got %s %s", expectedCmdPart, name, args[1])
			}
			return []byte(expectedContent), nil
		}

		content, err := FetchRemoteFile(context.Background(), mockCfg, mockRunner, filePath)
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		if string(content) != expectedContent {
			t.Errorf("Expected content %q, but got %q", expectedContent, string(content))
		}
	})

	t.Run("FailedFetch", func(t *testing.T) {
		mockError := errors.New("git command failed")
		mockRunner := &commandtest.MockRunner{}
		mockRunner.RunFunc = func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
			return nil, mockError
		}

		_, err := FetchRemoteFile(context.Background(), mockCfg, mockRunner, filePath)
		if err == nil {
			t.Fatal("Expected an error, but got nil")
		}

		if !strings.Contains(err.Error(), mockError.Error()) {
			t.Errorf("Expected error message to contain %q, but got: %v", mockError.Error(), err)
		}
	})

	t.Run("ContextTimeout", func(t *testing.T) {
		// Create a context that is already cancelled
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		mockRunner := &commandtest.MockRunner{}
		mockRunner.RunFunc = func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
			// The real runner would fail because the context is cancelled.
			// We can simulate this behavior.
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, nil
		}

		_, err := FetchRemoteFile(ctx, mockCfg, mockRunner, filePath)
		if err == nil {
			t.Fatal("Expected a context cancellation error, but got nil")
		}

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected error to be context.Canceled, but got: %v", err)
		}
	})
}
