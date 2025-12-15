package app

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gyr/relx-go/pkg/command"
	"github.com/gyr/relx-go/pkg/command/commandtest"
	"github.com/gyr/relx-go/pkg/config"
	"github.com/gyr/relx-go/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func TestHandleArtifacts(t *testing.T) {
	tests := []struct {
		name                 string
		project              string
		runner               command.Runner
		cfg                  *config.Config
		expectedOutput       string
		expectError          bool
		expectedErrorMessage string
	}{
		{
			name:    "success with artifacts",
			project: "test-project",
			runner: &commandtest.MockRunner{
				RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
					if strings.Contains(strings.Join(args, " "), "ls -b") {
						return []byte(" artifact1.iso\n artifact2.qcow2"), nil
					}
					// This is for `osc ls <project>`
					return []byte("test-package"), nil
				},
			},
			cfg: &config.Config{
				PackageFilterPatterns: []config.PackageFilter{
					{Pattern: "test-package", Repository: "test-repo"},
				},
				BinaryFilterPatterns: []string{"*.iso", "*.qcow2"},
			},
			expectedOutput: "Artifacts for project 'test-project':\nartifact1.iso\nartifact2.qcow2\n",
			expectError:    false,
		},
		{
			name:    "success with no artifacts",
			project: "test-project",
			runner: &commandtest.MockRunner{
				RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
					if strings.Contains(strings.Join(args, " "), "ls -b") {
						return []byte(""), nil // No binaries
					}
					return []byte("test-package"), nil
				},
			},
			cfg: &config.Config{
				PackageFilterPatterns: []config.PackageFilter{
					{Pattern: "test-package", Repository: "test-repo"},
				},
			},
			expectedOutput: "No artifacts found for project 'test-project'.\n",
			expectError:    false,
		},
		{
			name:    "error listing packages",
			project: "error-project",
			runner: &commandtest.MockRunner{
				RunFunc: func(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
					return nil, errors.New("osc command failed")
				},
			},
			cfg:                  &config.Config{},
			expectError:          true,
			expectedErrorMessage: "failed to list artifacts for project error-project: failed to list packages for project error-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var out bytes.Buffer

			// All tests need a logger and an output writer
			tt.cfg.Logger = logging.NewLogger(logging.LevelDebug)
			tt.cfg.OutputWriter = &out

			err := HandleArtifacts(ctx, tt.cfg, tt.runner, tt.project)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, out.String())
			}
		})
	}
}
