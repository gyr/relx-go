package commandtest

import (
	"context"
	"fmt"

	"github.com/gyr/relx-go/pkg/command"
)

// MockRunner is a mock implementation of the command.Runner for testing.
// It is shared across different packages to test components that depend on command.Runner.
type MockRunner struct {
	RunFunc func(ctx context.Context, workDir, name string, args ...string) ([]byte, error)
}

// This line is a compile-time check to ensure MockRunner implements command.Runner.
var _ command.Runner = (*MockRunner)(nil)

// Run executes the mock command.
func (m *MockRunner) Run(ctx context.Context, workDir, name string, args ...string) ([]byte, error) {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, workDir, name, args...)
	}
	return nil, fmt.Errorf("RunFunc not defined for mock runner")
}
