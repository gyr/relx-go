package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gyr/relx-go/pkg/command/commandtest"
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
)

// mockStdin is a helper function to simulate user input for interactive tests.
func mockStdin(t *testing.T, input string) func() {
	t.Helper()

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	os.Stdin = r

	go func() {
		defer func() {
			if err := w.Close(); err != nil {
				t.Errorf("failed to close pipe writer: %v", err)
			}
		}()
		_, err := w.WriteString(input)
		if err != nil {
			t.Errorf("failed to write to pipe: %v", err)
		}
	}()

	return func() {
		os.Stdin = oldStdin
	}
}
func TestHandleReview(t *testing.T) {
	// Common test setup
	const branch = "test-branch"
	const repository = "test-repo"
	const prID = "123"
	const reviewer = "test-reviewer"

	baseConfig := func() *config.Config {
		return &config.Config{
			Logger:       logging.NewLogger(logging.LevelDebug),
			PRReviewer:   reviewer,
			OutputWriter: &bytes.Buffer{},
		}
	}

	testCases := []struct {
		name           string
		userInput      string
		userFlag       string
		configReviewer string
		prIDs          []string // New field for passing PR IDs
		runner         *commandtest.MockRunner
		wantErr        string
		wantOutput     []string
	}{
		{
			name:           "Success_Approve_PR_From_Branch",
			userInput:      "y\na\n",
			userFlag:       "",
			configReviewer: reviewer,
			prIDs:          []string{}, // Empty slice to test branch logic
			runner: &commandtest.MockRunner{
				RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
					if name == "git-obs" && args[0] == "pr" && args[1] == "list" {
						return []byte(fmt.Sprintf("ID: #%s", prID)), nil
					}
					if name == "git-obs" && args[0] == "pr" && args[1] == "comment" {
						if !strings.Contains(args[4], fmt.Sprintf("@%s: approve", reviewer)) {
							return nil, fmt.Errorf("unexpected approval message: %s", args[4])
						}
						return nil, nil
					}
					return nil, nil // Default for other commands
				},
				RunPipelineFunc: func(ctx context.Context, workDir string, cmd1, cmd2 []string) error {
					if cmd1[0] == "git-obs" && cmd1[1] == "pr" && cmd1[2] == "show" {
						return nil
					}
					return fmt.Errorf("unexpected pipeline command: %v", cmd1)
				},
			},
			wantErr:    "",
			wantOutput: []string{"PR ID: 123", "Approve, skip, or exit?", "PR 123 approved."},
		},
		{
			name:           "Success_Approve_With_PR_IDs",
			userInput:      "y\na\na\n",
			userFlag:       "",
			configReviewer: reviewer,
			prIDs:          []string{"123", "456"}, // Provide PR IDs directly
			runner: &commandtest.MockRunner{
				// No RunFunc needed for 'pr list' as it should be skipped
				RunPipelineFunc: func(ctx context.Context, workDir string, cmd1, cmd2 []string) error {
					// This will be called for 'pr show'
					return nil
				},
				RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
					// This will be called for 'pr comment' (approval)
					if name == "git-obs" && args[0] == "pr" && args[1] == "comment" {
						return nil, nil
					}
					return nil, nil
				},
			},
			wantErr:    "",
			wantOutput: []string{"PR ID: 123", "PR ID: 456", "PR 123 approved.", "PR 456 approved."},
		},
		{
			name:           "No PRs found",
			userInput:      "",
			userFlag:       "",
			configReviewer: reviewer,
			prIDs:          []string{},
			runner: &commandtest.MockRunner{
				RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
					if name == "git-obs" && args[0] == "pr" && args[1] == "list" {
						return []byte(""), nil // Empty output
					}
					return nil, nil
				},
			},
			wantErr:    "",
			wantOutput: []string{"No open pull requests found"},
		},
		{
			name:           "User provided via flag",
			userInput:      "y\na\n",
			userFlag:       "cli-user",
			configReviewer: reviewer,
			prIDs:          []string{},
			runner: &commandtest.MockRunner{
				RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
					if name == "git-obs" && args[0] == "pr" && args[1] == "list" {
						return []byte(fmt.Sprintf("ID: #%s", prID)), nil
					}
					if name == "git-obs" && args[0] == "pr" && args[1] == "comment" {
						if !strings.Contains(args[4], "@cli-user: approve") {
							return nil, fmt.Errorf("unexpected approval message: %s", args[4])
						}
						return nil, nil
					}
					return nil, nil
				},
				RunPipelineFunc: func(ctx context.Context, workDir string, cmd1, cmd2 []string) error {
					return nil
				},
			},
			wantErr:    "",
			wantOutput: []string{"PR 123 approved."},
		},
		{
			name:           "No reviewer configured",
			userInput:      "",
			userFlag:       "",
			configReviewer: "", // No reviewer in config
			prIDs:          []string{},
			runner:         &commandtest.MockRunner{},
			wantErr:        "missing 'pr_reviewer' configuration",
		},
		{
			name:           "GetOpenPullRequests fails",
			userInput:      "",
			userFlag:       "",
			configReviewer: reviewer,
			prIDs:          []string{},
			runner: &commandtest.MockRunner{
				RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
					return nil, errors.New("gitea is down")
				},
			},
			wantErr: "failed to get open pull requests",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.userInput != "" {
				restoreStdin := mockStdin(t, tc.userInput)
				defer restoreStdin()
			}

			cfg := baseConfig()
			cfg.PRReviewer = tc.configReviewer

			err := HandleReview(context.Background(), cfg, tc.runner, branch, tc.prIDs, repository, tc.userFlag)

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("Expected error containing %q, but got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("Expected error message to contain %q, but got %q", tc.wantErr, err.Error())
				}
				return // Test ends here if an error is expected
			}

			if err != nil {
				t.Fatalf("Expected no error, but got: %v", err)
			}

			out := cfg.OutputWriter.(*bytes.Buffer).String()
			for _, expected := range tc.wantOutput {
				if !strings.Contains(out, expected) {
					t.Errorf("Output missing expected string %q. Full output:\n%s", expected, out)
				}
			}
		})
	}
}
